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

package config

// UserConfig contains configuration provided by the user, which remains unchanged throughout the entire run.
type UserConfig struct {
	InstanceTypes    string `json:"instance-types"`
	TestSuiteName    string `json:"test-suite"`
	TargetUtil       int    `json:"target-utilization"`
	VpcId            string `json:"vpc"`
	SubnetId         string `json:"subnet"`
	AmiId            string `json:"ami"`
	Timeout          int    `json:"timeout"`
	Persist          bool   `json:"persist"`
	Profile          string `json:"profile"`
	Region           string `json:"region"`
	Bucket           string `json:"bucket"`
	CustomScriptPath string `json:"custom-script"`
	ConfigFilePath   string `json:"config-file"`
}

// TestFixture contains constant information for the entire run.
type TestFixture struct {
	runId                   string
	testSuiteName           string
	compressedTestSuiteName string
	bucketName              string
	bucketRootDir           string
	targetUtil              int
	timeout                 int
	cfnStackName            string
	finalResultFilename     string
	userConfigFilename      string
	cfnTemplateFilename     string
	amiId                   string
}

var testFixture TestFixture
var userConfig UserConfig

// RunId returns runId.
func (t TestFixture) RunId() string {
	return t.runId
}

// TestSuiteName returns TestSuiteName.
func (t TestFixture) TestSuiteName() string {
	return t.testSuiteName
}

// CompressedTestSuiteName returns compressedTestSuiteName.
func (t TestFixture) CompressedTestSuiteName() string {
	return t.compressedTestSuiteName
}

// BucketName returns bucketName.
func (t TestFixture) BucketName() string {
	return t.bucketName
}

// BucketRootDir returns bucketRootDir.
func (t TestFixture) BucketRootDir() string {
	return t.bucketRootDir
}

// TargetUtil returns TargetUtil.
func (t TestFixture) TargetUtil() int {
	return t.targetUtil
}

// Timeout returns Timeout.
func (t TestFixture) Timeout() int {
	return t.timeout
}

// CfnStackName returns cfnStackName.
func (t TestFixture) CfnStackName() string {
	return t.cfnStackName
}

// FinalResultFilename returns finalResultFilename.
func (t TestFixture) FinalResultFilename() string {
	return t.finalResultFilename
}

// UserConfigFilename returns userConfigFilename.
func (t TestFixture) UserConfigFilename() string {
	return t.userConfigFilename
}

// CfnTemplateFilename returns cfnTemplateFilename.
func (t TestFixture) CfnTemplateFilename() string {
	return t.cfnTemplateFilename
}

// AmiId returns AmiId.
func (t TestFixture) AmiId() string {
	return t.amiId
}
