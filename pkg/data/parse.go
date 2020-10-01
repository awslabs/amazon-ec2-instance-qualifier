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
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	cpuMetric     = "cpu_usage_active"
	memMetric     = "mem_used_percent"
	resultFail    = "fail"
	statusSuccess = "SUCCESS"
	statusFail    = "FAIL"
)

// finalResultToArray parses the final result json file, populates and returns the instance results array.
func finalResultToArray(finalResult string) ([]resources.Instance, error) {
	finalResultJsonData, err := ioutil.ReadFile(resultsDir + "/" + finalResult)
	if err != nil {
		return nil, err
	}

	var v []resources.Instance
	if err := json.Unmarshal(finalResultJsonData, &v); err != nil {
		return nil, err
	}

	return v, nil
}

// parseInstanceResultToRow parses the instance result, populates and returns the row data which is used to
// generate the final output table.
func parseInstanceResultToRow(instanceResult resources.Instance) (row []string, err error) {
	maxCPU := 0.0
	maxMem := 0.0
	totalExecutionTime := 0.0
	success := true
	allTestsPass := true

	for _, result := range instanceResult.Results {
		if result.Status == resultFail {
			allTestsPass = false
		}

		if executionTime, err := strconv.ParseFloat(result.ExecutionTime, 64); err == nil {
			totalExecutionTime += executionTime
		} else {
			return nil, err
		}

		for _, metric := range result.Metrics {
			if metric.MetricUsed == cpuMetric {
				maxCPU = metric.Value
			} else if metric.MetricUsed == memMetric {
				maxMem = metric.Value
			}
			if metric.Value >= metric.Threshold {
				success = false
			}
		}
	}

	row = append(row, instanceResult.InstanceType)
	if success {
		row = append(row, statusSuccess)
	} else {
		row = append(row, statusFail)
	}
	row = append(row, fmt.Sprintf("%.3f", maxCPU))
	row = append(row, fmt.Sprintf("%.3f", maxMem))
	if instanceResult.IsTimeout {
		allTestsPass = false
	}
	row = append(row, strconv.FormatBool(allTestsPass))
	row = append(row, fmt.Sprintf("%.3f", totalExecutionTime))

	return row, nil
}
