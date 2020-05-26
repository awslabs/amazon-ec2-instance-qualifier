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
					MetricUsed: "cpu-load",
					Value:      1.22717948,
					Threshold:  1.6,
					Unit:       "n",
				},
				{
					MetricUsed: "mem-used",
					Value:      91.1025641025,
					Threshold:  6553.6,
					Unit:       "MiB",
				},
			},
		},
		{
			Label:         "mem-test.sh",
			Status:        "pass",
			ExecutionTime: "10.725",
			Metrics: []resources.Metric{
				{
					MetricUsed: "cpu-load",
					Value:      1.523333333,
					Threshold:  1.6,
					Unit:       "n",
				},
				{
					MetricUsed: "mem-used",
					Value:      3195,
					Threshold:  6553.6,
					Unit:       "MiB",
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
	expected := []string{"m4.large", "SUCCESS", "1.523", "76", "3195", "39", "true", "130.754"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusFail_AllPass(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Metrics[0].Value = 1.623
	expected := []string{"m4.large", "FAIL", "1.623", "81", "3195", "39", "true", "130.754"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusSuccess_NotAllPass(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Status = "fail"
	expected := []string{"m4.large", "SUCCESS", "1.523", "76", "3195", "39", "false", "130.754"}

	actual, err := parseInstanceResultToRow(instanceResult)
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestParseInstanceResultToRow_StatusFail_Timeout(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Results[1].Metrics[1].Value = 7000
	instanceResult.IsTimeout = true
	expected := []string{"m4.large", "FAIL", "1.523", "76", "7000", "85", "false", "130.754"}

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

func TestParseInstanceResultToRowInvalidVCpusFailure(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.VCpus = "VCPUS"

	_, err := parseInstanceResultToRow(instanceResult)
	h.Assert(t, err != nil, "Failed to return error when VCpus is invalid")
}

func TestParseInstanceResultToRowInvalidMemoryFailure(t *testing.T) {
	instanceResult := deepCopy(globalInstanceResult, t)
	instanceResult.Memory = "MEMORY"

	_, err := parseInstanceResultToRow(instanceResult)
	h.Assert(t, err != nil, "Failed to return error when Memory is invalid")
}
