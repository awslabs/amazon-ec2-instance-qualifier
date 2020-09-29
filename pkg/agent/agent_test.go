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

	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

// Tests

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
