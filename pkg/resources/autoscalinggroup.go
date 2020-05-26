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
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)

// suspendHealthCheckProcess suspends HealthCheck process of the auto scaling group to avoid the automatic
// relaunching after the termination of instances.
func (itf Resources) suspendHealthCheckProcess(asg string) error {
	_, err := itf.AutoScaling.SuspendProcesses(&autoscaling.ScalingProcessQuery{
		AutoScalingGroupName: aws.String(asg),
		ScalingProcesses:     []*string{aws.String("HealthCheck")},
	})
	if err != nil {
		return err
	}
	log.Printf("Auto Scaling Group %s successfully suspended HealthCheck process\n", asg)

	return nil
}

// attachInstancesToAutoScalingGroup attaches instances to the auto scaling group.
// The instances must be in running state.
func (itf Resources) attachInstancesToAutoScalingGroup(asg string, instanceIds []*string) error {
	_, err := itf.AutoScaling.AttachInstances(&autoscaling.AttachInstancesInput{
		AutoScalingGroupName: aws.String(asg),
		InstanceIds:          instanceIds,
	})
	if err != nil {
		return err
	}
	log.Printf("Successfully attached instances to Auto Scaling Group %s\n", asg)

	return nil
}
