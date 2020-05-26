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

package resources

// GetRegion returns the AWS region.
func (itf Resources) GetRegion() (region string, err error) {
	identityDoc, err := itf.EC2Metadata.GetInstanceIdentityDocument()
	if err != nil {
		return "", err
	}
	return identityDoc.Region, nil
}

// CreateInstance populates the Instance struct with metadata.
func (itf Resources) CreateInstance(instanceType string, vCpus string, memory string, osVersion string, architecture string) (instance Instance, err error) {
	instance.InstanceType = instanceType
	instance.VCpus = vCpus
	instance.Memory = memory
	instance.Os = osVersion
	instance.Architecture = architecture
	instance.Results = make([]Result, 0)

	instance.InstanceId, err = itf.getInstanceId()
	if err != nil {
		return instance, err
	}

	return instance, nil
}

func (itf Resources) getInstanceId() (instanceId string, err error) {
	identityDoc, err := itf.EC2Metadata.GetInstanceIdentityDocument()
	if err != nil {
		return "", err
	}
	return identityDoc.InstanceID, nil
}
