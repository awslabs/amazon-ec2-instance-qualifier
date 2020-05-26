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

package agent

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

// Tests

func TestPopulateThresholdsSuccess(t *testing.T) {
	instance := resources.Instance{
		VCpus:  "2",
		Memory: "8192",
	}

	err := PopulateThresholds(instance, "50")
	h.Ok(t, err)
	h.Equals(t, 1.0, metricInfos[cpu].threshold)
	h.Equals(t, 4096.0, metricInfos[mem].threshold)
}

func TestPopulateThresholdsInvalidTargetUtilFailure(t *testing.T) {
	instance := resources.Instance{
		VCpus:  "2",
		Memory: "8192",
	}

	err := PopulateThresholds(instance, "TARGET_UTIL")
	h.Assert(t, err != nil, "Failed to return error when target utilization is invalid")
}

func TestPopulateThresholdsInvalidVCpusFailure(t *testing.T) {
	instance := resources.Instance{
		VCpus:  "VCPUS",
		Memory: "8192",
	}

	err := PopulateThresholds(instance, "50")
	h.Assert(t, err != nil, "Failed to return error when VCpus is invalid")
}

func TestPopulateThresholdsInvalidMemoryFailure(t *testing.T) {
	instance := resources.Instance{
		VCpus:  "2",
		Memory: "MEMORY",
	}

	err := PopulateThresholds(instance, "50")
	h.Assert(t, err != nil, "Failed to return error when Memory is invalid")
}

func TestGetTestFileListSuccess(t *testing.T) {
	// Prepare the test directory
	err := os.Mkdir("temp-dir", 0755)
	defer os.RemoveAll("temp-dir")
	h.Assert(t, err == nil, "Error creating the temporary directory")
	createEmptyFile := func(filename string) {
		err := ioutil.WriteFile("temp-dir/"+filename, []byte(""), 0644)
		h.Assert(t, err == nil, "Error creating the file "+filename)
	}
	createEmptyFile("cpu-test.sh")
	createEmptyFile("mem-test.sh")
	createEmptyFile("monitor-cpu.sh")
	createEmptyFile("monitor-mem.sh")
	createEmptyFile("agent")
	createEmptyFile("cpu.load")
	createEmptyFile("mem.load")
	createEmptyFile("m4.large.log")
	createEmptyFile("cpu-test.sh-result.json")
	createEmptyFile("mem-test.sh-result.json")
	createEmptyFile("i-123456-test-results.json")

	fileList, err := GetTestFileList("temp-dir")
	h.Ok(t, err)
	h.Equals(t, []string{"temp-dir/cpu-test.sh", "temp-dir/mem-test.sh"}, fileList)
}

func TestGetTestFileListNonExistentDirFailure(t *testing.T) {
	_, err := GetTestFileList("non-existent-dir")
	h.Assert(t, err != nil, "Failed to return error when the directory doesn't exist")
}

func TestCalculateAvgValueSuccess(t *testing.T) {
	values := "123\n456\n789\n123"
	tempFile, err := ioutil.TempFile("", "temp-load")
	defer os.Remove(tempFile.Name())
	h.Assert(t, err == nil, "Error creating the temporary load file")
	err = ioutil.WriteFile(tempFile.Name(), []byte(values), 0644)
	h.Assert(t, err == nil, "Error writing the temporary load file")

	avg, err := calculateAvgValue(tempFile.Name())
	h.Ok(t, err)
	h.Equals(t, 372.75, avg)
	// Assert the load file is deleted
	_, err = os.Stat(tempFile.Name())
	h.Assert(t, os.IsNotExist(err), "Failed to delete the load file")
}

func TestCalculateAvgValueNonExistentFileFailure(t *testing.T) {
	_, err := calculateAvgValue("non-existent-file")
	h.Assert(t, err != nil, "Failed to return error when the load file doesn't exist")
}

func TestCalculateAvgValueInvalidDataFailure(t *testing.T) {
	values := "123\n \nabc\n321"
	tempFile, err := ioutil.TempFile("", "temp-load")
	defer os.Remove(tempFile.Name())
	h.Assert(t, err == nil, "Error creating the temporary load file")
	err = ioutil.WriteFile(tempFile.Name(), []byte(values), 0644)
	h.Assert(t, err == nil, "Error writing the temporary load file")

	_, err = calculateAvgValue(tempFile.Name())
	h.Assert(t, err != nil, "Failed to return error when data points are invalid")
	// Assert the load file is deleted
	_, err = os.Stat(tempFile.Name())
	h.Assert(t, os.IsNotExist(err), "Failed to delete the load file")
}
