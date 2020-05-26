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
	"os"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/agent"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

var instance = resources.Instance{
	InstanceId:   "i-0df3ef636ba12ee2a",
	InstanceType: "m4.large",
	VCpus:        "2",
	Memory:       "8192",
	Os:           "Linux/UNIX",
	Architecture: "x86_64",
	Results:      make([]resources.Result, 0),
}

// Tests

func TestCreateAgentFixtureSuccess(t *testing.T) {
	cwd, err := os.Getwd()
	h.Assert(t, err == nil, "Error getting the working directory")
	expected := agent.AgentFixture{
		BucketName:             "qualifier-bucket-12345",
		Timeout:                3600,
		BucketDir:              "Instance-Qualifier-Run-12345/m4.large/i-0df3ef636ba12ee2a",
		ScriptPath:             cwd,
		InstanceResultFilename: cwd + "/i-0df3ef636ba12ee2a-test-results.json",
		LogFilename:            cwd + "/m4.large.log",
	}

	actual, err := createAgentFixture(instance, "qualifier-bucket-12345", "3600", "Instance-Qualifier-Run-12345")
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}

func TestCreateAgentFixtureInvalidTimeoutFailure(t *testing.T) {
	_, err := createAgentFixture(instance, "qualifier-bucket-12345", "TIMEOUT", "Instance-Qualifier-Run-12345")
	h.Assert(t, err != nil, "Failed to return error when timeout is invalid")
}
