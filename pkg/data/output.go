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
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	finalOutputTableHeader = "INSTANCE TYPE,STATUS,CPU_USAGE_ACTIVE,CPU_THRESHOLD,MEM_USED_PERCENT,MEM_THRESHOLD,ALL TESTS PASS?,TOTAL EXECUTION TIME (sec)"
	notApplicable          = "N/A"
	instanceIdRegex        = "i-[0-9a-z]{17}"
)

// OutputAsTable parses the final result json file and outputs in table format.
func OutputAsTable(sess *session.Session, outputStream *os.File, results []*cloudwatch.MetricDataResult) error {
	svc := resources.New(sess)
	testFixture := config.GetTestFixture()
	finalResult, err := updateResults(results, testFixture)
	if err != nil {
		return err
	}

	log.Println("Updating local and remote results files after merging CloudWatch data")
	localPath := resultsDir + "/" + testFixture.FinalResultFilename
	remotePath := testFixture.BucketRootDir + "/" + testFixture.FinalResultFilename
	if err := cmdutil.MarshalToFile(finalResult, localPath); err != nil {
		log.Printf("There was an error saving updated results locally. final result: %v\n", finalResult)
	}
	if err := svc.UploadToBucket(testFixture.BucketName, localPath, remotePath); err != nil {
		log.Println("There was an error uploading updated results to S3")
	}

	var tableData [][]string
	for _, instanceResult := range finalResult {
		row, err := parseInstanceResultToRow(instanceResult)
		if err != nil {
			return err
		}
		tableData = append(tableData, row)
	}

	instances, err := svc.GetInstancesInCfnStack()
	if err != nil {
		return err
	}
	for _, instance := range instances {
		isFound := false
		for _, instanceResult := range finalResult {
			if instance.InstanceType == instanceResult.InstanceType {
				isFound = true
				break
			}
		}
		if !isFound {
			var row []string
			row = append(row, instance.InstanceType, notApplicable, notApplicable, notApplicable, notApplicable, notApplicable, notApplicable, notApplicable)
			tableData = append(tableData, row)
		}
	}
	cmdutil.RenderTable(tableData, strings.Split(finalOutputTableHeader, ","), outputStream)
	fmt.Fprintf(outputStream, "\nDetailed test results can be found in s3://%s/%s\n", testFixture.BucketName, testFixture.BucketRootDir)
	return nil
}

// updateResults updates the FinalResult of a test fixture with corresponding CloudWatch data
// then saves updates results file locally and remotely
func updateResults(results []*cloudwatch.MetricDataResult, testFixture config.TestFixture) ([]resources.Instance, error) {
	cwMetrics := make(map[string][]resources.Metric)
	metricThresholds := map[string]int{
		"cpu_usage_active": testFixture.CpuThreshold,
		"mem_used_percent": testFixture.MemThreshold,
	}
	for _, metricData := range results {
		if metricData.Values != nil {
			splitLabel := strings.Split(*metricData.Label, " ")
			var instanceId string
			for i, tag := range splitLabel {
				matched, err := regexp.MatchString(instanceIdRegex, tag)
				if err != nil {
					log.Println("Could not extract instanceId from MetricDataResult")
					return nil, err
				}
				if matched {
					instanceId = splitLabel[i]
					break
				}
			}
			if instanceId != "" {
				metricName := splitLabel[len(splitLabel)-1] //name is always last in label
				metricValue := *metricData.Values[0]
				thresholdValue := metricThresholds[metricName]
				metric := resources.Metric{
					MetricUsed: metricName,
					Value:      metricValue,
					Threshold:  float64(thresholdValue),
					Unit:       "Percent", //UserConfig
				}
				cwMetrics[instanceId] = append(cwMetrics[instanceId], metric)
			}
		}
	}

	finalResult, err := finalResultToArray(testFixture.FinalResultFilename)
	if err != nil {
		return finalResult, err
	}
	for _, instanceResult := range finalResult {
		oldRes := instanceResult.Results
		for i := range oldRes {
			oldRes[i].Metrics = cwMetrics[instanceResult.InstanceId]
		}
	}
	return finalResult, nil
}
