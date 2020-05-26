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
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
)

const (
	defaultAmi = "amzn2-ami-hvm-2.?.????????.?-x86_64-gp2"
)

// GetAmiId returns the AMI ID. If the user provides a valid AMI ID, return it; otherwise, return the ID of the
// latest AMI of Amazon Linux 2 in the current region.
func (itf Resources) GetAmiId(amiId string, inputStream *os.File, outputStream *os.File) (string, error) {
	if amiId != "" {
		_, err := itf.EC2.DescribeImages(&ec2.DescribeImagesInput{
			ImageIds: []*string{aws.String(amiId)},
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok {
				switch aerr.Code() {
				case "InvalidAMIID.NotFound", "InvalidAMIID.Malformed", "InvalidAMIID.Unavailable":
					prompt := fmt.Sprintf("The specified AMI %s doesn't exist/is no longer available. Use the default Amazon Linux 2 image?", amiId)
					answer, err := cmdutil.BoolPrompt(prompt, inputStream, outputStream)
					if err != nil {
						return "", err
					}
					if answer {
						return itf.GetAmiId("", inputStream, outputStream)
					}
					return "", fmt.Errorf("failed to use a valid ami")
				}
			}
			return "", err
		}
		return amiId, nil
	}

	output, err := itf.EC2.DescribeImages(&ec2.DescribeImagesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("name"),
				Values: []*string{aws.String(defaultAmi)},
			},
			{
				Name:   aws.String("state"),
				Values: []*string{aws.String("available")},
			},
		},
	})
	if err != nil {
		return "", err
	}
	if len(output.Images) == 0 {
		return "", fmt.Errorf("no available image is found")
	}

	// Sort the images to get the latest one
	sort.SliceStable(output.Images, func(i, j int) bool {
		return *output.Images[i].CreationDate > *output.Images[j].CreationDate
	})
	amiId = *output.Images[0].ImageId
	log.Printf("AMI id for Amazon Linux 2 is %s (created at %s)\n", amiId, *output.Images[0].CreationDate)

	return amiId, nil
}
