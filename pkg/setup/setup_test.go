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

package setup_test

import (
	"io/ioutil"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	userDataFile = "../../test/static/UserData/user_data.sh"
)

// Tests

func TestGetUserData(t *testing.T) {
	expected, err := ioutil.ReadFile(userDataFile)
	h.Assert(t, err == nil, "Error reading the user data file")

	actual := setup.GetUserData(resources.Instance{
		InstanceType: "m4.large",
		VCpus:        "2",
		Memory:       "8192",
		Os:           "Linux/UNIX",
		Architecture: "x86_64",
	})
	h.Equals(t, string(expected), actual)
}
