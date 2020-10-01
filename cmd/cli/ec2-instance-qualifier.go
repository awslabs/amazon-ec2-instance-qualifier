// Copyright 2020 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/data"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/template"
)

// Enums indicating what resources need to be deleted
const (
	deleteNothing = iota
	deleteCfnStack
	deleteAll           // delete bucket and CloudFormation stack
	testFixtureFileName = "test-fixture.json"
)

func main() {
	deleteState := deleteNothing
	inputStream := os.Stdin
	outputStream := os.Stdout
	rand.Seed(time.Now().UnixNano())

	userConfig, err := config.ParseCliArgs(outputStream)
	if err != nil {
		log.Fatal(err)
	}

	sess, err := newSession(userConfig)
	if err != nil {
		log.Fatal(err)
	}
	svc := resources.New(sess)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		terminate(sess, fmt.Errorf("interrupted"), deleteState)
	}()

	if userConfig.Bucket == "" {
		// For a new run, before tests begin on all instances, if the CLI is interrupted, all resources should be
		// deleted because it could be interpreted as "I don't want this run to be continued any more"
		deleteState = deleteAll

		vpcId, subnetId, err := svc.GetVpcAndSubnetIds(userConfig.VpcId, userConfig.SubnetId, inputStream, outputStream)
		if err != nil {
			log.Fatal(err)
		}

		runId := cmdutil.GetRandomString()
		fmt.Printf("Test Run ID: %s\n", runId)

		if err := svc.CreateBucket(runId, outputStream); err != nil {
			terminate(sess, err)
		}

		cfnTemplate, err := prepareForNewRun(sess, userConfig, runId, subnetId, inputStream, outputStream)
		if err != nil {
			terminate(sess, err)
		}

		if err := svc.CreateCfnStack(cfnTemplate, vpcId, subnetId, outputStream); err != nil {
			terminate(sess, err)
		}

		// After the tests begin, if the CLI is interrupted, we think the user may resume the session later to grab
		// the results, so nothing should be deleted
		deleteState = deleteNothing
		log.Println("The execution of test suite has been kicked off on all instances. You may quit now and later run the CLI again with the bucket name flag to get the result")
	} else {
		userConfig, err = prepareForResumedRun(sess, userConfig)
		if err != nil {
			terminate(sess, err)
		}
	}

	if err := data.PollForResults(sess); err != nil {
		terminate(sess, err)
	}

	instances, err := svc.GetInstancesInCfnStack()
	if err != nil {
		terminate(sess, err)
	}
	testFixture := config.GetTestFixture()
	cwResults, err := svc.GetCloudWatchData(instances, testFixture)
	if err != nil {
		terminate(sess, err)
	}

	if err := data.OutputAsTable(sess, outputStream, cwResults.MetricDataResults); err != nil {
		terminate(sess, err)
	}
	fmt.Println("User configuration and CloudFormation template are stored in the root directory of the bucket. You may check them if you want")
	// After outputting the final table, stack is no longer needed, but bucket should be kept for any deep dive
	deleteState = deleteCfnStack

	terminate(sess, nil, deleteState)
	fmt.Println("The process of cleaning up stack resources has started. You can quit now")
	if err := svc.WaitUntilCfnStackDeleteComplete(); err != nil {
		terminate(sess, err)
	}

	fmt.Println("Completed!")
}

// newSession returns a session with user provided config.
func newSession(userConfig config.UserConfig) (*session.Session, error) {
	sessOpts := session.Options{}
	region := userConfig.Region
	profile := userConfig.Profile
	if region != "" {
		sessOpts.Config.Region = &region
	}
	if profile != "" {
		sessOpts.Profile = profile
	}
	sess := session.Must(session.NewSessionWithOptions(sessOpts))
	log.Printf("Created session with region: %s\n", *sess.Config.Region)
	if sess.Config.Region != nil && *sess.Config.Region != "" {
		return sess, nil
	}

	return sess, fmt.Errorf("cannot start test run without a region; refer to Configuration for more information")
}

// prepareForNewRun does the preparation work for a new instance-qualifier run, including populating TestFixture,
// finding supported instance types, uploading the user configuration file, uploading the compressed test
// suite, and uploading the final CloudFormation template.
func prepareForNewRun(sess *session.Session, userConfig config.UserConfig, runId string, subnetId string, inputStream *os.File, outputStream *os.File) (cfnTemplate string, err error) {
	svc := resources.New(sess)

	amiId, err := svc.GetAmiId(userConfig.AmiId, inputStream, outputStream)
	if err != nil {
		return "", err
	}
	if err := config.PopulateTestFixture(userConfig, runId, amiId); err != nil {
		return "", err
	}
	testFixture := config.GetTestFixture()

	availabilityZone, instanceTypes, err := svc.FindBestAvailabilityZone(userConfig.InstanceTypes, subnetId)
	if err != nil {
		return "", err
	}
	instances, err := svc.GetSupportedInstances(instanceTypes, amiId, subnetId)
	if err != nil {
		return "", err
	}

	if err := config.WriteUserConfig(testFixture.UserConfigFilename); err != nil {
		return "", err
	}
	if err := uploadAndRemoveFile(sess, testFixture.BucketName, testFixture.UserConfigFilename, testFixture.UserConfigFilename); err != nil {
		return "", err
	}

	if err := setup.SetTestSuite(); err != nil {
		return "", err
	}
	if err := uploadAndRemoveFile(sess, testFixture.BucketName, testFixture.CompressedTestSuiteName, filepath.Base(testFixture.CompressedTestSuiteName)); err != nil {
		return "", err
	}
	// persist test fixture
	tfByte, err := json.Marshal(testFixture)
	if err != nil {
		return "", err
	}
	tfReader := bytes.NewReader(tfByte)
	if err := svc.UploadToS3(testFixture.BucketName, tfReader, testFixtureFileName); err != nil {
		return "", err
	}

	cfnTemplate, err = template.GenerateCfnTemplate(instances, userConfig.InstanceTypes, availabilityZone, inputStream, outputStream)
	if err != nil {
		return "", err
	}
	if err := ioutil.WriteFile(testFixture.CfnTemplateFilename, []byte(cfnTemplate), 0644); err != nil {
		return "", err
	}
	if err := uploadAndRemoveFile(sess, testFixture.BucketName, testFixture.CfnTemplateFilename, testFixture.CfnTemplateFilename); err != nil {
		return "", err
	}

	return cfnTemplate, nil
}

// prepareForResumedRun populates TestFixture, finds the types of all instances running in the stack, and populates
// UserConfig struct with the configuration in the previous session.
func prepareForResumedRun(sess *session.Session, userConfig config.UserConfig) (config.UserConfig, error) {
	svc := resources.New(sess)

	runId := resources.RemoveBucketNamePrefix(userConfig.Bucket)
	log.Printf("Test Run ID: %s\n", runId)
	log.Printf("Bucket Used: %s\n", userConfig.Bucket)

	// rehydrate test fixture
	tfByte, err := svc.DownloadFromS3(userConfig.Bucket, testFixtureFileName)
	if err != nil {
		return userConfig, err
	}
	if tfByte == nil {
		return userConfig, fmt.Errorf("downloaded TestFixture is nil")
	}
	if err := config.RestoreTestFixture(tfByte); err != nil {
		return userConfig, err
	}

	testFixture := config.GetTestFixture()

	if err := svc.DownloadFromBucket(testFixture.BucketName, testFixture.UserConfigFilename, testFixture.UserConfigFilename); err != nil {
		return userConfig, err
	}

	userConfig, err = config.ReadUserConfig(testFixture.UserConfigFilename)
	if err != nil {
		return userConfig, err
	}

	if err := os.Remove(testFixture.UserConfigFilename); err != nil {
		log.Println(err)
	}

	return userConfig, nil
}

func uploadAndRemoveFile(sess *session.Session, bucketName string, localPath string, remotePath string) error {
	svc := resources.New(sess)

	if err := svc.UploadToBucket(bucketName, localPath, remotePath); err != nil {
		return err
	}

	if err := os.Remove(localPath); err != nil {
		log.Println(err)
	}

	return nil
}

// terminate outputs error, deletes resources and exits conditionally.
func terminate(sess *session.Session, err error, deleteState ...int) {
	svc := resources.New(sess)

	state := deleteAll
	if len(deleteState) > 0 {
		state = deleteState[0]
	}

	if err != nil {
		log.Println(err)
	}

	if state != deleteNothing {
		if state == deleteAll {
			svc.DeleteBucket()
		}
		svc.DeleteCfnStack()
	}

	if err != nil {
		os.Exit(1)
	}
}
