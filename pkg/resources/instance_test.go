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
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

// Helpers

func isInstanceTypeAvailable(instanceType string) bool {
	return instanceType == "m4.large" || instanceType == "m4.xlarge" || instanceType == "c4.large" || instanceType == "a1.large"
}

// Tests

func TestFindBestAvailabilityZoneNewSubnet(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeAvailabilityZonesResp:             setupMockedEC2(t, describeAvailabilityZones, "availability_zones.json").DescribeAvailabilityZonesResp,
		DescribeInstanceTypeOfferingsRespUsEast2a: setupMockedEC2(t, describeInstanceTypeOfferings, "us_east_2a.json").DescribeInstanceTypeOfferingsRespUsEast2a,
		DescribeInstanceTypeOfferingsRespUsEast2b: setupMockedEC2(t, describeInstanceTypeOfferings, "us_east_2b.json").DescribeInstanceTypeOfferingsRespUsEast2b,
		DescribeInstanceTypeOfferingsRespUsEast2c: setupMockedEC2(t, describeInstanceTypeOfferings, "us_east_2c.json").DescribeInstanceTypeOfferingsRespUsEast2c,
	}
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	// c5a.12xlarge is not available in us-east-2a, a1.4xlarge is not available in us-east-2c
	availabilityZone, instanceTypes, err := itf.FindBestAvailabilityZone("m4.large,m4.xlarge,c4.large,c5a.12xlarge,a1.4xlarge", "NONE")
	h.Ok(t, err)
	h.Equals(t, "us-east-2b", availabilityZone)
	h.Equals(t, []string{"m4.large", "m4.xlarge", "c4.large", "c5a.12xlarge", "a1.4xlarge"}, instanceTypes)
}

func TestFindBestAvailabilityZoneExistingSubnet(t *testing.T) {
	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	availabilityZone, instanceTypes, err := itf.FindBestAvailabilityZone("m4.large,m4.xlarge,c4.large,c5a.12xlarge,a1.4xlarge", "subnet-12345")
	h.Ok(t, err)
	h.Equals(t, "", availabilityZone)
	h.Equals(t, []string{"m4.large", "m4.xlarge", "c4.large", "c5a.12xlarge", "a1.4xlarge"}, instanceTypes)
}

func TestGetSupportedInstancesNewSubnet(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesRespC5a12xlarge: setupMockedEC2(t, describeInstanceTypes, "c5a_12xlarge.json").DescribeInstanceTypesRespC5a12xlarge,
		DescribeImagesResp:                   setupMockedEC2(t, describeImages, "valid_ami_id.json").DescribeImagesResp,
	}
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	instances, err := itf.GetSupportedInstances([]string{"c5a.12xlarge"}, "VALID_AMI_ID", "NONE")
	h.Ok(t, err)
	expected := []resources.Instance{
		{
			InstanceType: "c5a.12xlarge",
			VCpus:        "48",
			Memory:       "98304",
			Os:           "Linux/UNIX",
			Architecture: "x86_64",
		},
	}
	h.Equals(t, expected, instances)
}

func TestGetSupportedInstancesExistingSubnet(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesRespM4Large:  setupMockedEC2(t, describeInstanceTypes, "m4_large.json").DescribeInstanceTypesRespM4Large,
		DescribeInstanceTypesRespM4Xlarge: setupMockedEC2(t, describeInstanceTypes, "m4_xlarge.json").DescribeInstanceTypesRespM4Xlarge,
		DescribeInstanceTypesRespA1Large:  setupMockedEC2(t, describeInstanceTypes, "a1_large.json").DescribeInstanceTypesRespA1Large,
		DescribeImagesResp:                setupMockedEC2(t, describeImages, "valid_ami_id.json").DescribeImagesResp,
		DescribeSubnetsResp:               setupMockedEC2(t, describeSubnets, "subnet_in_us_east_2a.json").DescribeSubnetsResp,
	}
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	instances, err := itf.GetSupportedInstances([]string{"m4.large", "m4.xlarge", "a1.large", "c5a.12xlarge"}, "VALID_AMI_ID", "subnet-123456")
	h.Ok(t, err)
	expected := []resources.Instance{
		{
			InstanceType: "m4.large",
			VCpus:        "2",
			Memory:       "8192",
			Os:           "Linux/UNIX",
			Architecture: "x86_64",
		},
		{
			InstanceType: "m4.xlarge",
			VCpus:        "4",
			Memory:       "16384",
			Os:           "Linux/UNIX",
			Architecture: "x86_64",
		},
	}
	h.Equals(t, expected, instances)
}

func TestGetSupportedInstancesNoSupportedInstanceTypeFailure(t *testing.T) {
	ec2Mock := mockedEC2{
		DescribeInstanceTypesRespA1Large: setupMockedEC2(t, describeInstanceTypes, "a1_large.json").DescribeInstanceTypesRespA1Large,
		DescribeImagesResp:               setupMockedEC2(t, describeImages, "valid_ami_id.json").DescribeImagesResp,
		DescribeSubnetsResp:              setupMockedEC2(t, describeSubnets, "subnet_in_us_east_2a.json").DescribeSubnetsResp,
	}
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	_, err := itf.GetSupportedInstances([]string{"a1.large", "c5a.12xlarge"}, "VALID_AMI_ID", "subnet-123456")
	h.Assert(t, err != nil, "Failed to return error when there is no supported instance type")
}
