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
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/agent"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	instanceResultSuffix = "-test-results.json"
	testResultSuffix     = "-result.json"
	bucketTestsDir       = "Tests"
)

// The agent runs all the tests in the test suite, populates the result json files, and uploads them to the
// S3 bucket.
func main() {
	outputStream := os.Stdout
	errStream := os.Stderr

	instanceType := os.Args[1]
	vCpus := os.Args[2]
	memory := os.Args[3]
	osVersion := os.Args[4]
	architecture := os.Args[5]
	bucketName := os.Args[6]
	timeout := os.Args[7]
	bucketRootDir := os.Args[8]
	targetUtil := os.Args[9]
	region := os.Args[10]

	// Warm-up time for 2 purposes:
	// 1. During booting, the CPU load is not stable, so wait for some time before starting all tests
	// 2. Ensure the instance is not terminated before being added to the auto scaling group
	time.Sleep(1 * time.Minute)

	sess, err := newAgentSession(region)
	if err != nil {
		agent.TerminateInstance()
	}
	svc := resources.New(sess)

	instance, err := svc.CreateInstance(instanceType, vCpus, memory, osVersion, architecture)
	if err != nil {
		agent.TerminateInstance()
	}

	agentFixture, err := createAgentFixture(instance, bucketName, timeout, bucketRootDir)
	if err != nil {
		agent.TerminateInstance()
	}

	done := make(chan bool, 1)
	go func() {
		select {
		case <-done:
			return
		case <-time.After(time.Second * time.Duration(agentFixture.Timeout)):
			fmt.Printf("\n======================================================================================================\n")
			fmt.Printf("💀 Timeout! One or more tests were not executed\n")
			fmt.Printf("======================================================================================================\n")

			instance.IsTimeout = true
			err = marshalAndUploadToBucketTestsDir(sess, instance, agentFixture.InstanceResultFilename, agentFixture)
			agent.Fatal(sess, agentFixture, err)
		}
	}()

	if err := agent.PopulateThresholds(instance, targetUtil); err != nil {
		agent.Fatal(sess, agentFixture, err)
	}

	// Upload first, in case that timeout occurs before getting any result
	if err := marshalAndUploadToBucketTestsDir(sess, instance, agentFixture.InstanceResultFilename, agentFixture); err != nil {
		agent.Fatal(sess, agentFixture, err)
	}

	testFileList, err := agent.GetTestFileList(agentFixture.ScriptPath)
	if err != nil {
		agent.Fatal(sess, agentFixture, err)
	}

	for _, testFile := range testFileList {
		testResult := agent.PopulateResult(testFile, agentFixture, outputStream, errStream)
		instance.Results = append(instance.Results, testResult)
		testResultFilename := testFile + testResultSuffix

		if err := marshalAndUploadToBucketTestsDir(sess, testResult, testResultFilename, agentFixture); err != nil {
			// Failing on one test result shouldn't terminate the whole program
			log.Println(err)
			continue
		}

		if err := marshalAndUploadToBucketTestsDir(sess, instance, agentFixture.InstanceResultFilename, agentFixture); err != nil {
			log.Println(err)
		}
	}

	remoteFinalInstanceResultFilename := agentFixture.BucketDir + "/" + filepath.Base(agentFixture.InstanceResultFilename)
	if err := svc.UploadToBucket(agentFixture.BucketName, agentFixture.InstanceResultFilename, remoteFinalInstanceResultFilename); err != nil {
		agent.Fatal(sess, agentFixture, err)
	}

	done <- true
	fmt.Printf("\n======================================================================================================\n")
	fmt.Printf("🎉 All test files finish execution\n")
	fmt.Printf("======================================================================================================\n")
	agent.Fatal(sess, agentFixture, nil)
}

// newAgentSession returns a session with region config.
func newAgentSession(region string) (*session.Session, error) {
	sessOpts := session.Options{}
	if region != "" {
		sessOpts.Config.Region = &region
	}
	sess := session.Must(session.NewSessionWithOptions(sessOpts))
	if sess.Config.Region != nil && *sess.Config.Region != "" {
		return sess, nil
	}
	errorMsg := "Unable to set a region for new agent session: \n"
	return sess, fmt.Errorf(errorMsg)
}

// createAgentFixture populates the AgentFixture struct.
func createAgentFixture(instance resources.Instance, bucketName string, timeout string, bucketRootDir string) (agentFixture agent.AgentFixture, err error) {
	agentFixture.BucketName = bucketName
	agentFixture.Timeout, err = strconv.Atoi(timeout)
	if err != nil {
		return agentFixture, err
	}
	agentFixture.ScriptPath, err = os.Getwd()
	if err != nil {
		return agentFixture, err
	}

	agentFixture.BucketDir = bucketRootDir + "/" + instance.InstanceType + "/" + instance.InstanceId
	agentFixture.InstanceResultFilename = agentFixture.ScriptPath + "/" + instance.InstanceId + instanceResultSuffix
	agentFixture.LogFilename = agentFixture.ScriptPath + "/" + instance.InstanceType + ".log"

	return agentFixture, nil
}

// marshalAndUploadToBucketTestsDir marshals an object to json string, writes it to a file, and uploads the
// file to the bucket tests directory.
func marshalAndUploadToBucketTestsDir(sess *session.Session, v interface{}, filename string, agentFixture agent.AgentFixture) error {
	svc := resources.New(sess)

	if err := cmdutil.MarshalToFile(v, filename); err != nil {
		return err
	}

	remoteFilename := agentFixture.BucketDir + "/" + bucketTestsDir + "/" + filepath.Base(filename)
	if err := svc.UploadToBucket(agentFixture.BucketName, filename, remoteFilename); err != nil {
		return err
	}

	return nil
}
