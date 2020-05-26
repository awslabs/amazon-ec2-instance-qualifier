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
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	finalOutputTableHeader = "INSTANCE TYPE,MEETS TARGET UTILIZATION?,MAX CPU (n),%,MAX MEM (MiB),%,ALL TESTS PASS?,TOTAL EXECUTION TIME (sec)"
	notApplicable          = "N/A"
)

// OutputAsTable parses the final result json file and outputs in table format.
func OutputAsTable(sess *session.Session, outputStream *os.File) error {
	svc := resources.New(sess)

	testFixture := config.GetTestFixture()
	finalResult, err := finalResultToArray(testFixture.FinalResultFilename())
	if err != nil {
		return err
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

	fmt.Fprintf(outputStream, "\nDetailed test results can be found in s3://%s/%s\n", testFixture.BucketName(), testFixture.BucketRootDir())

	return nil
}
