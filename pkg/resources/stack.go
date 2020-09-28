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
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
)

const (
	parameterVpc     = "providedVpc"
	parameterSubnet  = "providedSubnet"
	tagKey           = "instance-qualifier:id"
	errorTableHeader = "LOGICAL ID,TYPE,FAILURE REASON"
)

// CreateCfnStack creates the CloudFormation stack for the instance-qualifier run.
func (itf Resources) CreateCfnStack(cfnTemplate string, vpcId string, subnetId string, outputStream *os.File) error {
	testFixture := config.GetTestFixture()

	output, err := itf.CloudFormation.CreateStack(&cloudformation.CreateStackInput{
		StackName:    aws.String(testFixture.CfnStackName),
		TemplateBody: aws.String(cfnTemplate),
		Capabilities: []*string{aws.String("CAPABILITY_NAMED_IAM")},
		Parameters: []*cloudformation.Parameter{
			{
				ParameterKey:   aws.String(parameterVpc),
				ParameterValue: aws.String(vpcId),
			},
			{
				ParameterKey:   aws.String(parameterSubnet),
				ParameterValue: aws.String(subnetId),
			},
		},
		Tags: []*cloudformation.Tag{
			{
				Key:   aws.String(tagKey),
				Value: aws.String(testFixture.RunId),
			},
		},
	})
	if err != nil {
		return err
	}

	stackId := *output.StackId
	log.Printf("Waiting for stack %s to be created...\n", testFixture.CfnStackName)
	if err := itf.CloudFormation.WaitUntilStackCreateComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackId),
	}); err != nil {
		fmt.Fprintf(outputStream, "CloudFormation Stack %s failed to be created. Error: %v\n",testFixture.CfnStackName, err)
		fmt.Fprintf(outputStream, "Navigate to the AWS Console for additional information\n")
		fmt.Fprintln(outputStream, "Use the following command to delete the stack with AWS CLI:")
		fmt.Fprintf(outputStream, "aws cloudformation delete-stack --stack-name %s\n", testFixture.CfnStackName)
		return err
	}
	fmt.Fprintf(outputStream, "Stack Created: %s\n", testFixture.CfnStackName)

	// Do work post CloudFormation stack completion
	if err := itf.decorateComputeResources(); err != nil {
		return err
	}

	return nil
}

// DeleteCfnStack starts the async deletion of instance-qualifier CloudFormation stack.
func (itf Resources) DeleteCfnStack() error {
	stackName := config.GetTestFixture().CfnStackName

	_, err := itf.CloudFormation.DeleteStack(&cloudformation.DeleteStackInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return err
	}
	log.Printf("Started the process of deleting stack %s\n", stackName)

	return nil
}

// WaitUntilCfnStackDeleteComplete waits until the stack deletion is complete.
func (itf Resources) WaitUntilCfnStackDeleteComplete() error {
	stackName := config.GetTestFixture().CfnStackName

	log.Printf("Waiting for stack %s to be deleted...\n", stackName)
	if err := itf.CloudFormation.WaitUntilStackDeleteComplete(&cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}); err != nil {
		return err
	}
	log.Printf("Stack %s successfully deleted\n", stackName)

	return nil
}

// decorateComputeResources does some work that can't be done using CloudFormation template, including
// suspending the HealthCheck process of the auto scaling group, attaching instances to the auto scaling group,
// and adding tags to launch templates.
func (itf Resources) decorateComputeResources() error {
	testFixture := config.GetTestFixture()

	output, err := itf.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(testFixture.CfnStackName),
	})
	if err != nil {
		return err
	}

	var asgName string
	var instanceIds []*string
	var launchTemplateIds []*string
	for _, resource := range output.StackResources {
		switch *resource.ResourceType {
		case "AWS::AutoScaling::AutoScalingGroup":
			asgName = *resource.PhysicalResourceId
		case "AWS::EC2::Instance":
			instanceIds = append(instanceIds, resource.PhysicalResourceId)
		case "AWS::EC2::LaunchTemplate":
			launchTemplateIds = append(launchTemplateIds, resource.PhysicalResourceId)
		}
	}

	if err := itf.suspendHealthCheckProcess(asgName); err != nil {
		return err
	}
	if err := itf.attachInstancesToAutoScalingGroup(asgName, instanceIds); err != nil {
		return err
	}
	if err := itf.addTagsToEc2Resources(launchTemplateIds, testFixture.RunId); err != nil {
		return err
	}

	return nil
}

func (itf Resources) getInstanceIds(stackName string) (instanceIds []*string, err error) {
	output, err := itf.CloudFormation.DescribeStackResources(&cloudformation.DescribeStackResourcesInput{
		StackName: aws.String(stackName),
	})
	if err != nil {
		return nil, err
	}

	for _, resource := range output.StackResources {
		switch *resource.ResourceType {
		case "AWS::EC2::Instance":
			instanceIds = append(instanceIds, resource.PhysicalResourceId)
		}
	}

	return instanceIds, nil
}
