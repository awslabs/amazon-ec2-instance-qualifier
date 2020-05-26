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

package setup

import (
	"io/ioutil"
	"os"
	"testing"

	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	testFolder = "../../test/static/TestSuite/test-folder"
)

// Helpers

func cleanup() {
	os.Remove(testFolder + "/agent")
	os.Remove(testFolder + "/monitor-cpu.sh")
	os.Remove(testFolder + "/monitor-mem.sh")
}

// Tests

func TestCopyAgentScriptsToTestSuiteSuccess(t *testing.T) {
	// Mock agent bin
	err := ioutil.WriteFile("agent", []byte("AGENT"), 0644)
	defer os.Remove("agent")
	h.Assert(t, err == nil, "Error writing agent file")

	err = copyAgentScriptsToTestSuite(testFolder)
	defer cleanup()
	h.Ok(t, err)

	// Assert all script files are in the test suite
	assertScriptFileInTestSuite := func(filename string, expectedContent string) {
		data, err := ioutil.ReadFile(testFolder + "/" + filename)
		h.Assert(t, err == nil, "Error reading script file "+filename)
		h.Equals(t, expectedContent, string(data))
	}
	assertScriptFileInTestSuite("agent", "AGENT")
	assertScriptFileInTestSuite("monitor-cpu.sh", "")
	assertScriptFileInTestSuite("monitor-mem.sh", "")
}

func TestCopyAgentScriptsToTestSuiteNonExistentTestSuiteFailure(t *testing.T) {
	// Mock agent bin
	err := ioutil.WriteFile("agent", []byte("AGENT"), 0644)
	defer os.Remove("agent")
	h.Assert(t, err == nil, "Error writing agent file")

	err = copyAgentScriptsToTestSuite("non-existent-folder")
	defer cleanup()
	h.Assert(t, err != nil, "Failed to return error when test suite doesn't exist")
}

func TestCopyAgentScriptsToTestSuiteNonExistentAgentBinFailure(t *testing.T) {
	err := copyAgentScriptsToTestSuite(testFolder)
	defer cleanup()
	h.Assert(t, err != nil, "Failed to return error when agent bin doesn't exist")
}

func TestRemoveAgentScriptsFromTestSuiteSuccess(t *testing.T) {
	oldFiles, err := ioutil.ReadDir(testFolder)
	h.Assert(t, err == nil, "Error reading the directory "+testFolder)

	// Prepare the test suite
	defer cleanup()
	createScriptFileInTestSuite := func(filename string) {
		err := ioutil.WriteFile(testFolder+"/"+filename, []byte(""), 0644)
		h.Assert(t, err == nil, "Error creating script file "+filename)
	}
	createScriptFileInTestSuite("agent")
	createScriptFileInTestSuite("monitor-cpu.sh")
	createScriptFileInTestSuite("monitor-mem.sh")

	err = removeAgentScriptsFromTestSuite(testFolder)
	h.Ok(t, err)
	newFiles, err := ioutil.ReadDir(testFolder)
	h.Assert(t, err == nil, "Error reading the directory "+testFolder)

	// Assert the test suite doesn't change
	h.Equals(t, oldFiles, newFiles)
}

func TestRemoveAgentScriptsFromTestSuiteNonExistentTestSuiteFailure(t *testing.T) {
	err := removeAgentScriptsFromTestSuite("non-existent-folder")
	h.Assert(t, err != nil, "Failed to return error when test suite doesn't exist")
}
