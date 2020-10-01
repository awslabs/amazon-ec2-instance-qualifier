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

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"testing"

	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	configFilesPath = "../../test/static/UserConfigFiles"
)

var outputStream = os.Stdout

// Helpers

func resetTestFixture() {
	testFixture.RunId = ""
	testFixture.TestSuiteName = ""
	testFixture.CompressedTestSuiteName = ""
	testFixture.BucketName = ""
	testFixture.BucketRootDir = ""
	testFixture.CpuThreshold = 0
	testFixture.MemThreshold = 0
	testFixture.Timeout = 0
	testFixture.CfnStackName = ""
	testFixture.FinalResultFilename = ""
	testFixture.UserConfigFilename = ""
	testFixture.CfnTemplateFilename = ""
	testFixture.AmiId = ""
}

func resetFlagsForTest() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}
}

// Tests

func TestPopulateTestFixtureForNewRun(t *testing.T) {
	resetTestFixture()
	// For new run, bucketName should already be set before calling PopulateTestFixture
	testFixture.BucketName = "BUCKET_NAME"
	userConfig = UserConfig{
		InstanceTypes: "INSTANCE_TYPES",
		TestSuiteName: "TEST_SUITE_NAME",
		CpuThreshold:  50,
		MemThreshold:  50,
		Timeout:       12345,
		Region:        "REGION",
	}

	err := PopulateTestFixture(userConfig, "RUN_ID", "AMI_ID")
	h.Ok(t, err)

	cwd, err := os.Getwd()
	h.Assert(t, err == nil, "Error getting the working directory")
	// Assert all the values were set
	h.Equals(t, "RUN_ID", testFixture.RunId)
	h.Equals(t, cwd+"/TEST_SUITE_NAME", testFixture.TestSuiteName)
	h.Equals(t, cwd+"/TEST_SUITE_NAME.tar.gz", testFixture.CompressedTestSuiteName)
	h.Equals(t, "BUCKET_NAME", testFixture.BucketName)
	h.Equals(t, "Instance-Qualifier-Run-RUN_ID", testFixture.BucketRootDir)
	h.Equals(t, 50, testFixture.CpuThreshold)
	h.Equals(t, 50, testFixture.MemThreshold)
	h.Equals(t, 12345, testFixture.Timeout)
	h.Equals(t, "qualifier-stack-RUN_ID", testFixture.CfnStackName)
	h.Equals(t, "final-results-RUN_ID.json", testFixture.FinalResultFilename)
	h.Equals(t, "instance-qualifier-RUN_ID.config", testFixture.UserConfigFilename)
	h.Equals(t, "qualifier-cfn-template-RUN_ID.json", testFixture.CfnTemplateFilename)
	h.Equals(t, "AMI_ID", testFixture.AmiId)
}

func TestPopulateTestFixtureForResumedRun(t *testing.T) {
	prevTestFixture := testFixture
	prevTestFixByte, err := json.Marshal(prevTestFixture)
	h.Ok(t, err)

	resetTestFixture()
	h.Equals(t, "", testFixture.TestSuiteName)
	h.Equals(t, "", testFixture.CompressedTestSuiteName)

	err = RestoreTestFixture(prevTestFixByte)
	h.Ok(t, err)

	cwd, err := os.Getwd()
	h.Assert(t, err == nil, "Error getting the working directory")
	// Assert all the values were set
	h.Equals(t, "RUN_ID", testFixture.RunId)
	h.Equals(t, cwd+"/TEST_SUITE_NAME", testFixture.TestSuiteName)
	h.Equals(t, cwd+"/TEST_SUITE_NAME.tar.gz", testFixture.CompressedTestSuiteName)
	h.Equals(t, "BUCKET_NAME", testFixture.BucketName)
	h.Equals(t, "Instance-Qualifier-Run-RUN_ID", testFixture.BucketRootDir)
	h.Equals(t, 50, testFixture.CpuThreshold)
	h.Equals(t, 50, testFixture.MemThreshold)
	h.Equals(t, 12345, testFixture.Timeout)
	h.Equals(t, "qualifier-stack-RUN_ID", testFixture.CfnStackName)
	h.Equals(t, "final-results-RUN_ID.json", testFixture.FinalResultFilename)
	h.Equals(t, "instance-qualifier-RUN_ID.config", testFixture.UserConfigFilename)
	h.Equals(t, "qualifier-cfn-template-RUN_ID.json", testFixture.CfnTemplateFilename)
	h.Equals(t, "AMI_ID", testFixture.AmiId)
}

func TestParseCliArgsEnvSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Setenv("AWS_REGION", "us-weast-1")
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=30",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	h.Equals(t, "us-weast-1", userConfig.Region)
}

func TestParseCliArgsSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=30",
		"--vpc=VPC",
		"--subnet=SUBNET",
		"--ami=AMI",
		"--timeout=12345",
		"--persist=true",
		"--profile=PROFILE",
		"--region=REGION",
		"--bucket=BUCKET",
		"--custom-script=/path/to/script",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	// Assert all the values were set
	h.Equals(t, "INSTANCE_TYPES", userConfig.InstanceTypes)
	h.Equals(t, "TEST_SUITE", userConfig.TestSuiteName)
	h.Equals(t, 30, userConfig.CpuThreshold)
	h.Equals(t, 30, userConfig.MemThreshold)
	h.Equals(t, "VPC", userConfig.VpcId)
	h.Equals(t, "SUBNET", userConfig.SubnetId)
	h.Equals(t, "AMI", userConfig.AmiId)
	h.Equals(t, 12345, userConfig.Timeout)
	h.Equals(t, true, userConfig.Persist)
	h.Equals(t, "PROFILE", userConfig.Profile)
	h.Equals(t, "REGION", userConfig.Region)
	h.Equals(t, "BUCKET", userConfig.Bucket)
	h.Equals(t, "/path/to/script", userConfig.CustomScriptPath)
}

func TestParseCliArgsPriority(t *testing.T) {
	// Priority:
	// 1. CLI Args
	// 2. Env Vars
	// 3. Config File
	resetFlagsForTest()
	os.Setenv("AWS_REGION", "us-weast-1")
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES_Args",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=30",
		"--config-file=../../test/static/UserConfigFiles/valid.config",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	h.Equals(t, "INSTANCE_TYPES_Args", userConfig.InstanceTypes)                             //Args
	h.Equals(t, 30, userConfig.CpuThreshold)												  //Args
	h.Equals(t, 30, userConfig.MemThreshold)                                                 //Args
	h.Equals(t, "../../test/static/UserConfigFiles/valid.config", userConfig.ConfigFilePath) //Args
	h.Equals(t, "us-weast-1", userConfig.Region)                                             //Env
	h.Equals(t, 12345, userConfig.Timeout)                                                   //Config File
}

func TestParseCliArgsOnlyBucketSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--bucket=BUCKET",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	h.Equals(t, "BUCKET", userConfig.Bucket)
}

func TestParseCliArgsOverrides(t *testing.T) {
	resetFlagsForTest()
	os.Setenv("AWS_DEFAULT_REGION", "AWS_DEFAULT_REGION")
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=25",
		"--vpc=VPC",
		"--subnet=SUBNET",
		"--ami=AMI",
		"--timeout=12345",
		"--persist=true",
		"--profile=PROFILE",
		"--region=REGION",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	// Assert all the values were set
	h.Equals(t, "INSTANCE_TYPES", userConfig.InstanceTypes)
	h.Equals(t, "TEST_SUITE", userConfig.TestSuiteName)
	h.Equals(t, 30, userConfig.CpuThreshold)
	h.Equals(t, 25, userConfig.MemThreshold)
	h.Equals(t, "VPC", userConfig.VpcId)
	h.Equals(t, "SUBNET", userConfig.SubnetId)
	h.Equals(t, "AMI", userConfig.AmiId)
	h.Equals(t, 12345, userConfig.Timeout)
	h.Equals(t, true, userConfig.Persist)
	h.Equals(t, "PROFILE", userConfig.Profile)
	h.Equals(t, "REGION", userConfig.Region)
}

func TestParseCliArgsMissingInstanceTypesFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=25",
		"--vpc=VPC",
		"--subnet=SUBNET",
		"--timeout=12345",
		"--persist=true",
		"--region=REGION",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when instance-types not provided")
}

func TestParseCliArgsMissingTestSuiteFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--cpu-threshold=30",
		"--mem-threshold=25",
		"--subnet=SUBNET",
		"--ami=AMI",
		"--timeout=12345",
		"--profile=PROFILE",
		"--region=REGION",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when test-suite not provided")
}

func TestParseCliArgsMissingThresholdsFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--vpc=VPC",
		"--ami=AMI",
		"--timeout=12345",
		"--persist=true",
		"--profile=PROFILE",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when benchmark thresholds not provided")
}

func TestParseCliArgsNonPositiveThresholdFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=-25",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when non-positive benchmark threshold provided")
}

func TestParseCliArgsNonPositiveTimeoutFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--cpu-threshold=30",
		"--mem-threshold=25",
		"--timeout=0",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when non-positive Timeout provided")
}

func TestWriteUserConfigSuccess(t *testing.T) {
	actualConfigFile := "actual.config"
	defer os.Remove(actualConfigFile)
	userConfig = UserConfig{
		InstanceTypes: "INSTANCE_TYPES",
		TestSuiteName: "TEST_SUITE",
		CpuThreshold:  50,
		MemThreshold:  25,
		VpcId:         "VPC_ID",
		Timeout:       12345,
		Region:        "us-east-2",
	}

	err := WriteUserConfig(actualConfigFile)
	h.Ok(t, err)

	// Assert files contents are same
	expected, err := ioutil.ReadFile(configFilesPath + "/valid.config")
	h.Assert(t, err == nil, "Error reading config file valid.config")
	actual, err := ioutil.ReadFile(actualConfigFile)
	h.Assert(t, err == nil, "Error reading config file "+actualConfigFile)
	h.Equals(t, string(expected), string(actual))
}

func TestWriteUserConfigNonExistentFilePathFailure(t *testing.T) {
	userConfig = UserConfig{
		InstanceTypes: "INSTANCE_TYPES",
		TestSuiteName: "TEST_SUITE",
		CpuThreshold:  50,
		MemThreshold:  25,
		Timeout:       12345,
	}

	err := WriteUserConfig("non-existent-folder/CONFIG.config")
	h.Assert(t, err != nil, "Failed to return error when file path doesn't exist")
}

func TestReadUserConfigSuccess(t *testing.T) {
	expected := UserConfig{
		InstanceTypes: "INSTANCE_TYPES",
		TestSuiteName: "TEST_SUITE",
		CpuThreshold:  50,
		MemThreshold:  25,
		VpcId:         "VPC_ID",
		Timeout:       12345,
		Region:        "us-east-2",
	}

	actual, err := ReadUserConfig(configFilesPath + "/valid.config")
	h.Ok(t, err)

	// Assert all fields are same
	h.Equals(t, expected.InstanceTypes, actual.InstanceTypes)
	h.Equals(t, expected.TestSuiteName, actual.TestSuiteName)
	h.Equals(t, expected.CpuThreshold, actual.CpuThreshold)
	h.Equals(t, expected.MemThreshold, actual.MemThreshold)
	h.Equals(t, expected.VpcId, actual.VpcId)
	h.Equals(t, expected.SubnetId, actual.SubnetId)
	h.Equals(t, expected.AmiId, actual.AmiId)
	h.Equals(t, expected.Timeout, actual.Timeout)
	h.Equals(t, expected.Persist, actual.Persist)
	h.Equals(t, expected.Profile, actual.Profile)
	h.Equals(t, expected.Region, actual.Region)
	h.Equals(t, expected.Bucket, actual.Bucket)
}

func TestReadUserConfigNonExistentFileFailure(t *testing.T) {
	_, err := ReadUserConfig("non-existent-file")
	h.Assert(t, err != nil, "Failed to return error when file doesn't exist")
}

func TestReadUserConfigInvalidKeysFailure(t *testing.T) {
	_, err := ReadUserConfig(configFilesPath + "/invalid-keys.config")
	h.Assert(t, err != nil, "Failed to return error when user config file contains invalid keys")
}

func TestReadUserConfigInvalidFormatFailure(t *testing.T) {
	_, err := ReadUserConfig(configFilesPath + "/invalid-format.config")
	h.Assert(t, err != nil, "Failed to return error when user config file has invalid format")
}
