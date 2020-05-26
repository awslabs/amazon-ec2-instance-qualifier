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
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
)

const (
	runningState = "16"
)

var osVersion string
var architecture string

// IsInstanceRunning returns true if an instance is in running state; false otherwise.
func (itf Resources) IsInstanceRunning(instanceId string) (bool, error) {
	output, err := itf.EC2.DescribeInstanceStatus(&ec2.DescribeInstanceStatusInput{
		InstanceIds: []*string{aws.String(instanceId)},
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-state-code"),
				Values: []*string{aws.String(runningState)},
			},
		},
	})
	if err != nil {
		return false, err
	}

	if len(output.InstanceStatuses) > 0 {
		return true, nil
	}
	return false, nil
}

// GetInstancesInCfnStack populates InstanceId and InstanceType fields of the Instance struct for all instances in the
// CloudFormation stack, and returns them.
func (itf Resources) GetInstancesInCfnStack() (instances []Instance, err error) {
	instanceIds, err := itf.getInstanceIds(config.GetTestFixture().CfnStackName())
	if err != nil {
		return nil, err
	}

	output, err := itf.EC2.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: instanceIds,
	})
	if err != nil {
		return nil, err
	}

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			instances = append(instances, Instance{
				InstanceType: *instance.InstanceType,
				InstanceId:   *instance.InstanceId,
			})
		}
	}

	return instances, nil
}

// FindBestAvailabilityZone finds the Availability Zone that can support the most instance types provided by the
// user. It returns the name of the best Availability Zone, and user-specified instance types that are available
// in that Availability Zone.
func (itf Resources) FindBestAvailabilityZone(instanceTypes string, subnetId string) (bestAvailabilityZone string, supportedInstanceTypes []string, err error) {
	instanceTypesList := strings.Split(instanceTypes, ",")
	if subnetId != none {
		// If the user provides a subnet, the CLI must use it
		return "", instanceTypesList, nil
	}

	azMap := make(map[string][]string)

	availabilityZones, err := itf.getAvailabilityZones()
	if err != nil {
		return "", nil, err
	}

	for _, availabilityZone := range availabilityZones {
		output, err := itf.EC2.DescribeInstanceTypeOfferings(&ec2.DescribeInstanceTypeOfferingsInput{
			LocationType: aws.String("availability-zone"),
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("location"),
					Values: []*string{aws.String(availabilityZone)},
				},
			},
		})
		if err != nil {
			return "", nil, err
		}

		azMap[availabilityZone] = make([]string, 0)
		for _, instanceType := range instanceTypesList {
			isFound := false
			for _, instanceTypeOffering := range output.InstanceTypeOfferings {
				if instanceType == *instanceTypeOffering.InstanceType {
					isFound = true
					break
				}
			}

			if isFound {
				azMap[availabilityZone] = append(azMap[availabilityZone], instanceType)
			}
		}
	}

	maxSupportedInstancesNum := 0
	for availabilityZone, supportedInstanceTypesList := range azMap {
		if len(supportedInstanceTypesList) > maxSupportedInstancesNum {
			maxSupportedInstancesNum = len(supportedInstanceTypesList)
			bestAvailabilityZone = availabilityZone
		}
	}
	supportedInstanceTypes = azMap[bestAvailabilityZone]
	log.Printf("Use %s to create subnet which supports %v\n", bestAvailabilityZone, supportedInstanceTypes)

	return bestAvailabilityZone, supportedInstanceTypes, nil
}

// GetSupportedInstances returns instances that are supported to be launched. It checks 2 things:
// 1. Whether the instance type is available in the Availability Zone.
// 2. Whether the instance type supports the AMI architecture.
// The metadata of returned instances is also populated.
func (itf Resources) GetSupportedInstances(instanceTypes []string, amiId string, subnetId string) (instances []Instance, err error) {
	for _, instanceType := range instanceTypes {
		if isAvailableInAZ, err := itf.isInstanceTypeAvailableInSubnet(instanceType, subnetId); err != nil || !isAvailableInAZ {
			if err != nil {
				log.Println(err)
			}
			continue
		}

		instance, err := itf.populateMetadata(instanceType, amiId)
		if err != nil {
			log.Println(err)
			continue
		}

		instances = append(instances, instance)
	}

	if len(instances) == 0 {
		return nil, fmt.Errorf("no instance type is supported")
	}

	return instances, nil
}

// isInstanceTypeAvailableInSubnet checks whether an instance type is available in a subnet.
func (itf Resources) isInstanceTypeAvailableInSubnet(instanceType string, subnetId string) (bool, error) {
	if subnetId == none {
		return true, nil
	}

	availabilityZone, err := itf.getSubnetAvailabilityZone(subnetId)
	if err != nil {
		return false, err
	}

	output, err := itf.EC2.DescribeInstanceTypeOfferings(&ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: aws.String("availability-zone"),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("location"),
				Values: []*string{aws.String(availabilityZone)},
			},
			{
				Name:   aws.String("instance-type"),
				Values: []*string{aws.String(instanceType)},
			},
		},
	})
	if err != nil {
		return false, err
	}

	if len(output.InstanceTypeOfferings) == 0 {
		log.Printf("%s is not available in %s\n", instanceType, availabilityZone)
		return false, nil
	}
	return true, nil
}

// populateMetadata populates the Instance struct with metadata of the instance, including instance type,
// number of vCPUs, memory size, OS info, and architecture. The function also checks whether the instance
// type supports the AMI architecture. If not, an error is returned.
func (itf Resources) populateMetadata(instanceType string, amiId string) (instance Instance, err error) {
	instance.InstanceType = instanceType

	instanceTypesOutput, err := itf.EC2.DescribeInstanceTypes(&ec2.DescribeInstanceTypesInput{
		InstanceTypes: []*string{aws.String(instanceType)},
	})
	if err != nil {
		return instance, err
	}

	instanceTypeInfo := instanceTypesOutput.InstanceTypes[0]
	instance.VCpus = strconv.Itoa(int(*instanceTypeInfo.VCpuInfo.DefaultVCpus))
	instance.Memory = strconv.Itoa(int(*instanceTypeInfo.MemoryInfo.SizeInMiB))

	if osVersion == "" {
		imagesOutput, err := itf.EC2.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(amiId)},
		})
		if err != nil {
			return instance, err
		}
		imageInfo := imagesOutput.Images[0]
		osVersion = *imageInfo.PlatformDetails
		architecture = *imageInfo.Architecture
	}

	isSupported := false
	for _, arch := range instanceTypeInfo.ProcessorInfo.SupportedArchitectures {
		if *arch == architecture {
			isSupported = true
			break
		}
	}
	if isSupported == false {
		return instance, fmt.Errorf("%s doesn't support %s (%s)", instanceType, amiId, architecture)
	}

	instance.Os = osVersion
	instance.Architecture = architecture

	return instance, nil
}
