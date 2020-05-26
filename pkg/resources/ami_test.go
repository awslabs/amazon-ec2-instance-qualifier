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
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	describeImages                = "DescribeImages"
	describeAvailabilityZones     = "DescribeAvailabilityZones"
	describeInstanceTypeOfferings = "DescribeInstanceTypeOfferings"
	describeInstanceTypes         = "DescribeInstanceTypes"
	describeSubnets               = "DescribeSubnets"
	describeVpcs                  = "DescribeVpcs"
	mockFilesPath                 = "../../test/static"
)

var inputStream = os.Stdin
var outputStream = os.Stdout

// Helpers

func prepareInput(input string) (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "temp-file")
	if err != nil {
		return nil, err
	}
	if _, err := tempFile.WriteString(input); err != nil {
		return nil, err
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return tempFile, nil
}

// Mocking helpers

type mockedEC2 struct {
	ec2iface.EC2API
	DescribeImagesResp                        ec2.DescribeImagesOutput
	DescribeImagesErr                         error
	DescribeAvailabilityZonesResp             ec2.DescribeAvailabilityZonesOutput
	DescribeAvailabilityZonesErr              error
	DescribeInstanceTypeOfferingsRespUsEast2a ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsRespUsEast2b ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsRespUsEast2c ec2.DescribeInstanceTypeOfferingsOutput
	DescribeInstanceTypeOfferingsErr          error
	DescribeInstanceTypesRespM4Large          ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesRespM4Xlarge         ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesRespA1Large          ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesRespC5a12xlarge      ec2.DescribeInstanceTypesOutput
	DescribeInstanceTypesErr                  error
	DescribeSubnetsResp                       ec2.DescribeSubnetsOutput
	DescribeSubnetsErr                        error
	DescribeVpcsResp                          ec2.DescribeVpcsOutput
	DescribeVpcsErr                           error
}

func (m mockedEC2) DescribeImages(input *ec2.DescribeImagesInput) (*ec2.DescribeImagesOutput, error) {
	if input.ImageIds != nil && *input.ImageIds[0] == "INVALID_AMI_ID" {
		return &ec2.DescribeImagesOutput{}, awserr.New("InvalidAMIID.NotFound", "INVALID AMI ID", nil)
	}
	return &m.DescribeImagesResp, m.DescribeImagesErr
}

func (m mockedEC2) DescribeAvailabilityZones(input *ec2.DescribeAvailabilityZonesInput) (*ec2.DescribeAvailabilityZonesOutput, error) {
	return &m.DescribeAvailabilityZonesResp, m.DescribeAvailabilityZonesErr
}

func (m mockedEC2) DescribeInstanceTypeOfferings(input *ec2.DescribeInstanceTypeOfferingsInput) (*ec2.DescribeInstanceTypeOfferingsOutput, error) {
	if len(input.Filters) == 1 { // filters don't contain instance-type
		value := *input.Filters[0].Values[0]
		switch value {
		case "us-east-2a":
			return &m.DescribeInstanceTypeOfferingsRespUsEast2a, m.DescribeInstanceTypeOfferingsErr
		case "us-east-2b":
			return &m.DescribeInstanceTypeOfferingsRespUsEast2b, m.DescribeInstanceTypeOfferingsErr
		case "us-east-2c":
			return &m.DescribeInstanceTypeOfferingsRespUsEast2c, m.DescribeInstanceTypeOfferingsErr
		}
	} else {
		var value string
		for _, filter := range input.Filters {
			if *filter.Name == "instance-type" {
				value = *filter.Values[0]
				break
			}
		}
		if isInstanceTypeAvailable(value) {
			return &ec2.DescribeInstanceTypeOfferingsOutput{
				InstanceTypeOfferings: []*ec2.InstanceTypeOffering{
					{
						InstanceType: aws.String(value),
						Location:     aws.String("us-east-2a"),
						LocationType: aws.String("availability-zone"),
					},
				},
			}, nil
		}
	}
	return &ec2.DescribeInstanceTypeOfferingsOutput{}, nil
}

func (m mockedEC2) DescribeInstanceTypes(input *ec2.DescribeInstanceTypesInput) (*ec2.DescribeInstanceTypesOutput, error) {
	switch *input.InstanceTypes[0] {
	case "m4.large":
		return &m.DescribeInstanceTypesRespM4Large, m.DescribeInstanceTypesErr
	case "m4.xlarge":
		return &m.DescribeInstanceTypesRespM4Xlarge, m.DescribeInstanceTypesErr
	case "a1.large":
		return &m.DescribeInstanceTypesRespA1Large, m.DescribeInstanceTypesErr
	case "c5a.12xlarge":
		return &m.DescribeInstanceTypesRespC5a12xlarge, m.DescribeInstanceTypesErr
	}
	return &ec2.DescribeInstanceTypesOutput{}, nil
}

func (m mockedEC2) DescribeSubnets(input *ec2.DescribeSubnetsInput) (*ec2.DescribeSubnetsOutput, error) {
	if input.SubnetIds != nil && *input.SubnetIds[0] == "INVALID_SUBNET_ID" {
		return &ec2.DescribeSubnetsOutput{}, awserr.New("InvalidSubnetID.NotFound", "INVALID SUBNET ID", nil)
	}
	return &m.DescribeSubnetsResp, m.DescribeSubnetsErr
}

func (m mockedEC2) DescribeVpcs(input *ec2.DescribeVpcsInput) (*ec2.DescribeVpcsOutput, error) {
	if *input.VpcIds[0] == "INVALID_VPC_ID" {
		return &ec2.DescribeVpcsOutput{}, awserr.New("InvalidVpcID.NotFound", "INVALID VPC ID", nil)
	}
	return &m.DescribeVpcsResp, m.DescribeVpcsErr
}

func setupMockedEC2(t *testing.T, api string, file string) mockedEC2 {
	mockFilename := fmt.Sprintf("%s/%s/%s", mockFilesPath, api, file)
	mockFile, err := ioutil.ReadFile(mockFilename)
	h.Assert(t, err == nil, "Error reading mock file "+mockFilename)
	switch api {
	case describeImages:
		dio := ec2.DescribeImagesOutput{}
		err = json.Unmarshal(mockFile, &dio)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		return mockedEC2{
			DescribeImagesResp: dio,
		}
	case describeAvailabilityZones:
		dazo := ec2.DescribeAvailabilityZonesOutput{}
		err = json.Unmarshal(mockFile, &dazo)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		return mockedEC2{
			DescribeAvailabilityZonesResp: dazo,
		}
	case describeInstanceTypeOfferings:
		ditoo := ec2.DescribeInstanceTypeOfferingsOutput{}
		err = json.Unmarshal(mockFile, &ditoo)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		switch file {
		case "us_east_2a.json":
			return mockedEC2{
				DescribeInstanceTypeOfferingsRespUsEast2a: ditoo,
			}
		case "us_east_2b.json":
			return mockedEC2{
				DescribeInstanceTypeOfferingsRespUsEast2b: ditoo,
			}
		case "us_east_2c.json":
			return mockedEC2{
				DescribeInstanceTypeOfferingsRespUsEast2c: ditoo,
			}
		}
	case describeInstanceTypes:
		dito := ec2.DescribeInstanceTypesOutput{}
		err = json.Unmarshal(mockFile, &dito)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		switch file {
		case "m4_large.json":
			return mockedEC2{
				DescribeInstanceTypesRespM4Large: dito,
			}
		case "m4_xlarge.json":
			return mockedEC2{
				DescribeInstanceTypesRespM4Xlarge: dito,
			}
		case "a1_large.json":
			return mockedEC2{
				DescribeInstanceTypesRespA1Large: dito,
			}
		case "c5a_12xlarge.json":
			return mockedEC2{
				DescribeInstanceTypesRespC5a12xlarge: dito,
			}
		}
	case describeSubnets:
		dso := ec2.DescribeSubnetsOutput{}
		err = json.Unmarshal(mockFile, &dso)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		return mockedEC2{
			DescribeSubnetsResp: dso,
		}
	case describeVpcs:
		dvo := ec2.DescribeVpcsOutput{}
		err = json.Unmarshal(mockFile, &dvo)
		h.Assert(t, err == nil, "Error parsing mock json file contents "+mockFilename)
		return mockedEC2{
			DescribeVpcsResp: dvo,
		}
	default:
		h.Assert(t, false, "Unable to mock the provided API type "+api)
	}
	return mockedEC2{}
}

// Tests

func TestGetAmiIdValidId(t *testing.T) {
	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	amiId, err := itf.GetAmiId("VALID_AMI_ID", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, "VALID_AMI_ID", amiId)
}

func TestGetAmiIdInvalidId_UseDefaultAmi(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("y\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	ec2Mock := setupMockedEC2(t, describeImages, "no_image_id.json")
	itf := resources.Resources{
		EC2: ec2Mock,
	}
	amiId, err := itf.GetAmiId("INVALID_AMI_ID", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, "ami-016b213e65284e9c9", amiId)
}

func TestGetAmiIdInvalidId_NotProceedFailure(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("N\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	_, err = itf.GetAmiId("INVALID_AMI_ID", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when answering no to using default AMI")
}

func TestGetAmiIdEmptyId_NoAvailableAmiFailure(t *testing.T) {
	itf := resources.Resources{
		EC2: mockedEC2{},
	}
	_, err := itf.GetAmiId("", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when there is no available AMI")
}
