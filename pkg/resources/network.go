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

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
)

const (
	none               = "NONE"
	subnetsTableHeader = "OPTION,SUBNET,AVAILABILITY ZONE,CIDR BLOCK"
)

// GetVpcAndSubnetIds returns VPC and subnet IDs.
//
// If the user doesn't specify VPC nor subnet, the function returns "NONE" for both VPC and subnet IDs, meaning
// that a new VPC infrastructure needs to be created.
//
// If the user only specifies VPC, the function asks the user to choose any one of subnets in that VPC. If the
// VPC ID is invalid, the user will be asked whether to create a new VPC. If there is no subnet in the VPC, an
// error will be returned.
//
// If the user only specifies subnet, the function finds the corresponding VPC and returns the IDs of them. If
// the subnet ID is invalid, the user will be asked whether to create a new VPC.
//
// If the user specifies both VPC and subnet, the function validates them. If valid, returns them; otherwise,
// returns error.
func (itf Resources) GetVpcAndSubnetIds(vpcId string, subnetId string, inputStream *os.File, outputStream *os.File) (string, string, error) {
	if vpcId == "" && subnetId == "" {
		return none, none, nil
	} else if vpcId != "" && subnetId == "" {
		return itf.onlyVpcProvided(vpcId, inputStream, outputStream)
	} else if vpcId == "" && subnetId != "" {
		return itf.onlySubnetProvided(subnetId, inputStream, outputStream)
	}
	return itf.bothVpcAndSubnetProvided(vpcId, subnetId)
}

func (itf Resources) onlyVpcProvided(vpcId string, inputStream *os.File, outputStream *os.File) (string, string, error) {
	isValid, err := itf.isVpcValid(vpcId)
	if err != nil {
		return "", "", err
	}
	if !isValid {
		prompt := fmt.Sprintf("The specified VPC %s doesn't exist/we were unable to locate that VPC. Create a new VPC and proceed?", vpcId)
		answer, err := cmdutil.BoolPrompt(prompt, inputStream, outputStream)
		if err != nil {
			return "", "", err
		}
		if answer {
			return none, none, nil
		}
		return "", "", fmt.Errorf("failed to use a valid vpc")
	}

	subnets, err := itf.getSubnets(vpcId)
	if err != nil {
		return "", "", err
	}
	if len(subnets) == 0 {
		return "", "", fmt.Errorf("there is no subnet in the VPC %s", vpcId)
	}

	var tableData [][]string
	for i, subnet := range subnets {
		var row []string
		row = append(row, strconv.Itoa(i))
		row = append(row, *subnet.SubnetId)
		row = append(row, *subnet.AvailabilityZone)
		row = append(row, *subnet.CidrBlock)
		tableData = append(tableData, row)
	}
	cmdutil.RenderTable(tableData, strings.Split(subnetsTableHeader, ","), outputStream)

	prompt := fmt.Sprintf("Please select a subnet. Your option:")
	option, err := cmdutil.OptionPrompt(prompt, len(subnets), inputStream, outputStream)
	if err != nil {
		return "", "", err
	}

	return vpcId, *subnets[option].SubnetId, nil
}

func (itf Resources) onlySubnetProvided(subnetId string, inputStream *os.File, outputStream *os.File) (string, string, error) {
	isValid, vpcId, err := itf.isSubnetValid(subnetId)
	if err != nil {
		return "", "", err
	}
	if !isValid {
		prompt := fmt.Sprintf("The specified subnet %s doesn't exist/we were unable to locate that subnet. Create a new VPC and proceed?", subnetId)
		answer, err := cmdutil.BoolPrompt(prompt, inputStream, outputStream)
		if err != nil {
			return "", "", err
		}
		if answer {
			return none, none, nil
		}
		return "", "", fmt.Errorf("failed to use a valid vpc")
	}

	return vpcId, subnetId, nil
}

func (itf Resources) bothVpcAndSubnetProvided(vpcId string, subnetId string) (string, string, error) {
	isValid, realVpcId, err := itf.isSubnetValid(subnetId)
	if err != nil {
		return "", "", err
	}
	if !isValid || vpcId != realVpcId {
		return "", "", fmt.Errorf("invalid vpc/subnet")
	}

	return vpcId, subnetId, nil
}

func (itf Resources) isVpcValid(vpcId string) (bool, error) {
	_, err := itf.EC2.DescribeVpcs(&ec2.DescribeVpcsInput{
		VpcIds: []*string{aws.String(vpcId)},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidVpcID.NotFound", "InvalidVpcID.Malformed":
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (itf Resources) isSubnetValid(subnetId string) (isExist bool, vpcId string, err error) {
	output, err := itf.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(subnetId)},
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case "InvalidSubnetID.NotFound", "InvalidSubnetID.Malformed":
				return false, "", nil
			}
		}
		return false, "", err
	}
	vpcId = *output.Subnets[0].VpcId
	return true, vpcId, nil
}

func (itf Resources) getSubnets(vpcId string) (subnets []*ec2.Subnet, err error) {
	subnetsOutput, err := itf.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcId)},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return subnetsOutput.Subnets, nil
}

func (itf Resources) getSubnetAvailabilityZone(subnetId string) (string, error) {
	output, err := itf.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: []*string{aws.String(subnetId)},
	})
	if err != nil {
		return "", err
	}

	return *output.Subnets[0].AvailabilityZone, nil
}

func (itf Resources) getAvailabilityZones() (availabilityZones []string, err error) {
	output, err := itf.EC2.DescribeAvailabilityZones(&ec2.DescribeAvailabilityZonesInput{})
	if err != nil {
		return nil, err
	}

	for _, availabilityZone := range output.AvailabilityZones {
		availabilityZones = append(availabilityZones, *availabilityZone.ZoneName)
	}

	return availabilityZones, nil
}
