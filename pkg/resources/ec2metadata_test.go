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

package resources_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	getInstanceIdentityDocument = "GetInstanceIdentityDocument"
)

// Mocking Helpers

type mockedEC2Metadata struct {
	resources.EC2MetadataAPI
	GetInstanceIdentityDocumentResp ec2metadata.EC2InstanceIdentityDocument
	GetInstanceIdentityDocumentErr  error
}

func (m mockedEC2Metadata) GetInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error) {
	return m.GetInstanceIdentityDocumentResp, m.GetInstanceIdentityDocumentErr
}

func setupMockedEC2Metadata(t *testing.T, api string, file string) mockedEC2Metadata {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+mockFilename)
	if api == getInstanceIdentityDocument {
		eiid := ec2metadata.EC2InstanceIdentityDocument{}
		err = json.Unmarshal(mockFile, &eiid)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		return mockedEC2Metadata{
			GetInstanceIdentityDocumentResp: eiid,
		}
	} else {
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedEC2Metadata{}
}

// Tests

func TestCreateInstanceSuccess(t *testing.T) {
	expected := resources.Instance{
		InstanceId:   "i-0df3ef636ba12ee2a",
		InstanceType: "m4.large",
		VCpus:        "2",
		Memory:       "8192",
		Os:           "Linux/UNIX",
		Architecture: "x86_64",
		Results:      make([]resources.Result, 0),
	}
	ec2MetadataMock := setupMockedEC2Metadata(t, getInstanceIdentityDocument, "m4_large.json")
	itf := resources.Resources{
		EC2Metadata: ec2MetadataMock,
	}

	actual, err := itf.CreateInstance("m4.large", "2", "8192", "Linux/UNIX", "x86_64")
	h.Ok(t, err)
	h.Equals(t, expected, actual)
}
