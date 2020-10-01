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

package data

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	resultsDir           = "results"
	instanceResultSuffix = "-test-results.json"
	bucketTestsDir       = "Tests"
	pollingPeriod        = 5 * time.Second
)

// PollForResults polls for all instance results from the bucket in parallel.
// After each successful polling, the instance result is appended to the final result and the new final result
// is uploaded to the bucket. Upon returning, final result json file is ready to be parsed.
func PollForResults(sess *session.Session) error {
	svc := resources.New(sess)
	testFixture := config.GetTestFixture()
	instances, err := svc.GetInstancesInCfnStack()
	if err != nil {
		return err
	}

	results := make(chan string, len(instances))
	errChan := make(chan error, 1)

	// This goroutine will receive instance results from the channel. Each time it gets a new instance result, it will append it to the final result, then upload the new final result to the bucket
	go func() {
		if err := os.MkdirAll(resultsDir, os.ModePerm); err != nil {
			errChan <- err
			return
		}

		localFinalResult := resultsDir + "/" + testFixture.FinalResultFilename
		remoteFinalResult := testFixture.BucketRootDir + "/" + testFixture.FinalResultFilename
		if err := ioutil.WriteFile(localFinalResult, []byte("[]"), 0644); err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case instanceResult, ok := <-results:
				if ok {
					if err := appendResultAndUpload(sess, testFixture.BucketName, localFinalResult, remoteFinalResult, instanceResult); err != nil {
						// Failing to append one instance result should not terminate the whole program
						log.Println(err)
					}
				} else {
					if err := svc.DownloadFromBucket(testFixture.BucketName, localFinalResult, remoteFinalResult); err != nil {
						// Can use the local version
						log.Println(err)
					}
					errChan <- nil
					return
				}
			}
		}
	}()

	var wg sync.WaitGroup
	for _, instance := range instances {
		wg.Add(1)
		go func(instance resources.Instance) {
			defer wg.Done()

			instanceId := instance.InstanceId
			instanceType := instance.InstanceType
			filename := instanceId + instanceResultSuffix
			localInstanceResult := resultsDir + "/" + filename
			remoteInstanceResult := testFixture.BucketRootDir + "/" + instanceType + "/" + instanceId + "/" + filename
			// If the instance doesn't finish the execution of all test files before timeout, fetch this partial instance result
			remoteFallbackInstanceResult := testFixture.BucketRootDir + "/" + instanceType + "/" + instanceId + "/" + bucketTestsDir + "/" + filename

			if err := pollForResult(sess, testFixture.BucketName, instanceId, localInstanceResult, remoteInstanceResult, remoteFallbackInstanceResult); err == nil {
				instanceResult, err := ioutil.ReadFile(localInstanceResult)
				if err != nil {
					// Failing to read the instance result from the file should terminate the current goroutine
					log.Println(err)
					return
				}
				results <- string(instanceResult)
				if err := os.Remove(localInstanceResult); err != nil {
					// Failing to delete the file is acceptable
					log.Println(err)
				}
			} else {
				// Failing to poll for the result of one instance should not terminate the whole program
				log.Println(err)
			}
		}(instance)
	}
	wg.Wait()
	close(results)

	return <-errChan
}

// pollForResult polls for one instance result periodically until the instance is not running.
func pollForResult(sess *session.Session, bucket string, instanceId string, localPath string, remotePath string, fallbackPath string) error {
	svc := resources.New(sess)
	ticker := time.NewTicker(pollingPeriod)
	filename := filepath.Base(localPath)

	log.Printf("Polling for %s...\n", filename)
	for {
		select {
		case <-ticker.C:
			if err := svc.DownloadFromBucket(bucket, localPath, remotePath); err == nil {
				ticker.Stop()
				log.Printf("Polling for %s succeeded\n", filename)
				return nil
			}
			isRunning, err := svc.IsInstanceRunning(instanceId)
			if err != nil {
				return err
			}
			if !isRunning {
				ticker.Stop()
				if err := svc.DownloadFromBucket(bucket, localPath, fallbackPath); err != nil {
					return err
				}
				log.Printf("Polling for %s timeout, downloaded from %s\n", filename, fallbackPath)
				return nil
			}
		}
	}
}

// appendResultAndUpload appends the instance result to the final result and uploads the new final result to
// the bucket.
func appendResultAndUpload(sess *session.Session, bucket string, localPath string, remotePath string, instanceResult string) error {
	svc := resources.New(sess)

	allResults, err := finalResultToArray(filepath.Base(localPath))
	if err != nil {
		return err
	}

	var newResult resources.Instance
	if err := json.Unmarshal([]byte(instanceResult), &newResult); err != nil {
		return err
	}

	allResults = append(allResults, newResult)
	if err := cmdutil.MarshalToFile(allResults, localPath); err != nil {
		return err
	}

	if err := svc.UploadToBucket(bucket, localPath, remotePath); err != nil {
		return err
	}

	return nil
}
