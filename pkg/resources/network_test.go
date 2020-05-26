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
	"os"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	validVpcId    = "vpc-3513b45e"
	validSubnetId = "subnet-80ce00eb"
)

// Tests

func TestGetVpcAndSubnetIdsNoneProvided(t *testing.T) {
	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds("", "", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, "NONE", vpcId)
	h.Equals(t, "NONE", subnetId)
}

func TestGetVpcAndSubnetIdsOnlyVpcProvided_ValidVpc(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("1\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	ec2Mock := mockedEC2{
		DescribeVpcsResp:    setupMockedEC2(t, describeVpcs, "valid_vpc.json").DescribeVpcsResp,
		DescribeSubnetsResp: setupMockedEC2(t, describeSubnets, "subnets_in_vpc.json").DescribeSubnetsResp,
	}
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds(validVpcId, "", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, validVpcId, vpcId)
	h.Equals(t, "subnet-50f39f1c", subnetId)
}

func TestGetVpcAndSubnetIdsOnlyVpcProvided_ValidVpc_NoSubnetFailure(t *testing.T) {
	ec2Mock := setupMockedEC2(t, describeVpcs, "valid_vpc.json")
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	_, _, err := itf.GetVpcAndSubnetIds(validVpcId, "", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when there is no subnet in the VPC")
}

func TestGetVpcAndSubnetIdsOnlyVpcProvided_InvalidVpc_CreateNewVpc(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("y\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds("INVALID_VPC_ID", "", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, "NONE", vpcId)
	h.Equals(t, "NONE", subnetId)
}

func TestGetVpcAndSubnetIdsOnlyVpcProvided_InvalidVpc_NotProceedFailure(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("N\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	_, _, err = itf.GetVpcAndSubnetIds("INVALID_VPC_ID", "", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when answering no to creating a new VPC")
}

func TestGetVpcAndSubnetIdsOnlySubnetProvided_ValidSubnet(t *testing.T) {
	ec2Mock := setupMockedEC2(t, describeSubnets, "subnet_in_us_east_2a.json")
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds("", validSubnetId, inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, validVpcId, vpcId)
	h.Equals(t, validSubnetId, subnetId)
}

func TestGetVpcAndSubnetIdsOnlySubnetProvided_InvalidSubnet_CreateNewVpc(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("y\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds("", "INVALID_SUBNET_ID", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, "NONE", vpcId)
	h.Equals(t, "NONE", subnetId)
}

func TestGetVpcAndSubnetIdsOnlySubnetProvided_InvalidSubnet_NotProceedFailure(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("N\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	_, _, err = itf.GetVpcAndSubnetIds("", "INVALID_SUBNET_ID", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when answering no to creating a new VPC")
}

func TestGetVpcAndSubnetIdsBothProvided_ValidInput(t *testing.T) {
	ec2Mock := setupMockedEC2(t, describeSubnets, "subnet_in_us_east_2a.json")
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	vpcId, subnetId, err := itf.GetVpcAndSubnetIds(validVpcId, validSubnetId, inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, validVpcId, vpcId)
	h.Equals(t, validSubnetId, subnetId)
}

func TestGetVpcAndSubnetIdsBothProvided_InvalidVpcFailure(t *testing.T) {
	ec2Mock := setupMockedEC2(t, describeSubnets, "subnet_in_us_east_2a.json")
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	_, _, err := itf.GetVpcAndSubnetIds("INVALID_VPC_ID", validSubnetId, inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when VPC is invalid")
}

func TestGetVpcAndSubnetIdsBothProvided_InvalidSubnetFailure(t *testing.T) {
	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	_, _, err := itf.GetVpcAndSubnetIds(validVpcId, "INVALID_SUBNET_ID", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when subnet is invalid")
}
