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
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
)

const (
	bucketNamePrefix = "qualifier-bucket-"
)

// CreateBucket creates a bucket and blocks all public access.
func (itf Resources) CreateBucket(runId string, outputStream *os.File) error {
	bucket := bucketNamePrefix + runId
	config.SetTestFixtureBucketName(bucket)

	// Create
	_, err := itf.S3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			// proceed if bucket already exists or owned
			case s3.ErrCodeBucketAlreadyExists:
				fmt.Println(s3.ErrCodeBucketAlreadyExists, aerr.Error())
			case s3.ErrCodeBucketAlreadyOwnedByYou:
				fmt.Println(s3.ErrCodeBucketAlreadyOwnedByYou, aerr.Error())
			default:
				fmt.Println(aerr.Error())
				return err
			}
		}
	}

	// Wait until exists
	log.Printf("Waiting for bucket %s to be created...\n", bucket)
	if err := itf.S3.WaitUntilBucketExists(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		return err
	}
	log.Printf("Bucket %s successfully created\n", bucket)

	// Block all public access
	_, err = itf.S3.PutPublicAccessBlock(&s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucket),
		PublicAccessBlockConfiguration: &s3.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	})
	if err != nil {
		return err
	}
	log.Printf("Bucket %s has blocked all public access\n", bucket)
	fmt.Fprintf(outputStream, "Bucket Created: %s\n", bucket)

	return nil
}

// UploadToBucket uploads a local file to the specified location of the bucket.
func (itf Resources) UploadToBucket(bucket string, localPath string, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = itf.S3ManagerUploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(remotePath),
		Body:   file,
	})
	if err != nil {
		return err
	}
	log.Printf("%s successfully uploaded to s3://%s/%s\n", filepath.Base(localPath), bucket, remotePath)

	return nil
}

// UploadToS3 is similar to UploadToBucket, but provide data directly instead of a local file
func (itf Resources) UploadToS3(bucketName string, data io.Reader, remotePath string) error {

	_, err := itf.S3ManagerUploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(remotePath),
		Body:   data,
	})
	if err != nil {
		return err
	}
	log.Printf("data successfully uploaded to s3://%s/%s\n", bucketName, remotePath)
	return nil
}

// DownloadFromBucket downloads a file from the bucket to the specified local location.
func (itf Resources) DownloadFromBucket(bucket string, localPath string, remotePath string) error {
	// Should first check whether the file exists in the bucket, otherwise we may create useless empty local file
	_, err := itf.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		return err
	}

	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = itf.S3ManagerDownloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(remotePath),
		})
	if err != nil {
		return err
	}
	log.Printf("%s successfully downloaded from s3://%s/%s\n", filepath.Base(localPath), bucket, remotePath)

	return nil
}

// DownloadFromS3 is similar to DownloadFromBucket, but returns bytes directly instead of saving to a local file
func (itf Resources) DownloadFromS3(bucketName string, remotePath string) ([]byte, error) {
	_, err := itf.S3.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(remotePath),
	})
	if err != nil {
		return nil, err
	}

	buf := aws.NewWriteAtBuffer([]byte{})

	_, err = itf.S3ManagerDownloader.Download(buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(remotePath),
		})
	if err != nil {
		return buf.Bytes(), err
	}
	log.Printf("data successfully downloaded from s3://%s/%s\n", bucketName, remotePath)
	return buf.Bytes(), nil
}

// DeleteBucket empties and deletes the instance-qualifier bucket.
func (itf Resources) DeleteBucket() error {
	bucket := config.GetTestFixture().BucketName

	// First delete all objects
	iter := s3manager.NewDeleteListIterator(itf.S3, &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	if err := s3manager.NewBatchDeleteWithClient(itf.S3).Delete(aws.BackgroundContext(), iter); err != nil {
		return err
	}

	// Delete the bucket
	_, err := itf.S3.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		return err
	}
	log.Printf("Bucket %s successfully deleted\n", bucket)

	return nil
}

// RemoveBucketNamePrefix removes the prefix from the bucket name and returns the test run ID.
func RemoveBucketNamePrefix(bucket string) string {
	return strings.Replace(bucket, bucketNamePrefix, "", 1)
}
