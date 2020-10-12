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
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

var globalInstanceResult = resources.Instance{
	InstanceId:   "i-0ff4a2f594b270b54",
	InstanceType: "m4.large",
	VCpus:        "2",
	Memory:       "8192",
	Os:           "Linux/UNIX",
	Architecture: "x86_64",
	IsTimeout:    false,
	Results: []resources.Result{
		{
			Label:         "cpu-test.sh",
			Status:        "pass",
			ExecutionTime: "120.029",
			Metrics: []resources.Metric{
				{
					MetricUsed: "cpu_usage_active",
					Value:      35.8,
					Threshold:  40.0,
					Unit:       "Percent",
				},
				{
					MetricUsed: "mem_used_percent",
					Value:      1.48,
					Threshold:  40.0,
					Unit:       "Percent",
				},
			},
		},
		{
			Label:         "mem-test.sh",
			Status:        "pass",
			ExecutionTime: "10.725",
			Metrics: []resources.Metric{
				{
					MetricUsed: "cpu_usage_active",
					Value:      10.523333333,
					Threshold:  40.0,
					Unit:       "Percent",
				},
				{
					MetricUsed: "mem_used_percent",
					Value:      37.77,
					Threshold:  40.0,
					Unit:       "Percent",
				},
			},
		},
	},
}

// Helpers

func deepCopy(src resources.Instance, t *testing.T) (dest resources.Instance) {
	bytes, err := json.Marshal(src)
	h.Assert(t, err == nil, "Error marshalling src")
	err = json.Unmarshal(bytes, &dest)
	h.Assert(t, err == nil, "Error unmarshalling to dest")
	return dest
}

// Tests

func TestParseInstanceResultToRow_StatusSuccess_AllPass(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	expected := []string{"m4.large", "SUCCESS", "10.52", "40.00", "37.77", "40.00", "true", "130.75"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusFail_AllPass(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Metrics[0].Value = 41.623
	expected := []string{"m4.large", "FAIL", "41.62", "40.00", "37.77", "40.00", "true", "130.75"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusSuccess_NotAllPass(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Status = "fail"
	expected := []string{"m4.large", "SUCCESS", "10.52", "40.00", "37.77", "40.00", "false", "130.75"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusFail_Timeout(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Metrics[1].Value = 45.456
	instanceResult.IsTimeout = true
	expected := []string{"m4.large", "FAIL", "10.52", "40.00", "45.46", "40.00", "false", "130.75"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRowInvalidExecutionTimeFailure(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].ExecutionTime = "EXECUTION_TIME"

	_, err := parseInstanceResultToRow(instanceResult)
	h.Assert(t, err != nil, "Failed to return error when ExecutionTime is invalid")
}
