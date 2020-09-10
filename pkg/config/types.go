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
	instanceTypes    string
	testSuiteName    string
	targetUtil       int
	vpcId            string
	subnetId         string
	amiId            string
	timeout          int
	persist          bool
	profile          string
	region           string
	bucket           string
	customScriptPath string
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

// InstanceTypes returns instanceTypes.
func (u UserConfig) InstanceTypes() string {
	return u.instanceTypes
}

// TestSuiteName returns testSuiteName.
func (u UserConfig) TestSuiteName() string {
	return u.testSuiteName
}

// TargetUtil returns targetUtil.
func (u UserConfig) TargetUtil() int {
	return u.targetUtil
}

// CustomScriptPath returns customScriptPath.
func (u UserConfig) CustomScriptPath() string {
	return u.customScriptPath
}

// VpcId returns vpcId.
func (u UserConfig) VpcId() string {
	return u.vpcId
}

// SubnetId returns subnetId.
func (u UserConfig) SubnetId() string {
	return u.subnetId
}

// AmiId returns amiId.
func (u UserConfig) AmiId() string {
	return u.amiId
}

// Timeout returns timeout.
func (u UserConfig) Timeout() int {
	return u.timeout
}

// Persist returns persist.
func (u UserConfig) Persist() bool {
	return u.persist
}

// Profile returns profile.
func (u UserConfig) Profile() string {
	return u.profile
}

// Region returns region.
func (u UserConfig) Region() string {
	return u.region
}

// Bucket returns bucket.
func (u UserConfig) Bucket() string {
	return u.bucket
}

// RunId returns runId.
func (t TestFixture) RunId() string {
	return t.runId
}

// TestSuiteName returns testSuiteName.
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

// TargetUtil returns targetUtil.
func (t TestFixture) TargetUtil() int {
	return t.targetUtil
}

// Timeout returns timeout.
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

// AmiId returns amiId.
func (t TestFixture) AmiId() string {
	return t.amiId
}
