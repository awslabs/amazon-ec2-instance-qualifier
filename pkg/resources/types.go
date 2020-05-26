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
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3/s3manager/s3manageriface"
)

// EC2MetadataAPI provides an interface to enable mocking the ec2metadata.EC2Metadata service client's APIs.
type EC2MetadataAPI interface {
	GetInstanceIdentityDocument() (ec2metadata.EC2InstanceIdentityDocument, error)
}

// Resources is used to store clients for AWS services.
type Resources struct {
	EC2                 ec2iface.EC2API
	AutoScaling         autoscalingiface.AutoScalingAPI
	S3                  s3iface.S3API
	S3ManagerUploader   s3manageriface.UploaderAPI
	S3ManagerDownloader s3manageriface.DownloaderAPI
	CloudFormation      cloudformationiface.CloudFormationAPI
	EC2Metadata         EC2MetadataAPI
}

// Metric represents the metric data.
type Metric struct {
	MetricUsed string  `json:"metric"`
	Value      float64 `json:"value"`
	Threshold  float64 `json:"threshold"`
	Unit       string  `json:"unit"`
}

// Result represents the result of one test file.
type Result struct {
	Label         string   `json:"label"`
	Status        string   `json:"status"`
	ExecutionTime string   `json:"execution-time"`
	Metrics       []Metric `json:"Metrics"`
}

// Instance contains the data of an instance.
type Instance struct {
	InstanceId   string   `json:"instance-id"`
	InstanceType string   `json:"instance-type"`
	VCpus        string   `json:"vCPUs"`
	Memory       string   `json:"memory"`
	Os           string   `json:"OS"`
	Architecture string   `json:"Architecture"`
	IsTimeout    bool     `json:"isTimeout"`
	Results      []Result `json:"results"`
}

// New creates an instance of Resources provided an AWS session.
func New(sess *session.Session) *Resources {
	return &Resources{
		EC2:                 ec2.New(sess),
		AutoScaling:         autoscaling.New(sess),
		S3:                  s3.New(sess),
		S3ManagerUploader:   s3manager.NewUploader(sess),
		S3ManagerDownloader: s3manager.NewDownloader(sess),
		CloudFormation:      cloudformation.New(sess),
		EC2Metadata:         ec2metadata.New(sess),
	}
}
