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
	testFixture.runId = ""
	testFixture.testSuiteName = ""
	testFixture.compressedTestSuiteName = ""
	testFixture.bucketName = ""
	testFixture.bucketRootDir = ""
	testFixture.targetUtil = 0
	testFixture.timeout = 0
	testFixture.cfnStackName = ""
	testFixture.finalResultFilename = ""
	testFixture.userConfigFilename = ""
	testFixture.cfnTemplateFilename = ""
	testFixture.amiId = ""
}

func resetFlagsForTest() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	os.Args = []string{"cmd"}
}

// Tests

func TestPopulateTestFixtureForNewRun(t *testing.T) {
	resetTestFixture()
	// For new run, bucketName should already be set before calling PopulateTestFixture
	testFixture.bucketName = "BUCKET_NAME"
	userConfig := UserConfig{
		instanceTypes: "INSTANCE_TYPES",
		testSuiteName: "TEST_SUITE_NAME",
		targetUtil:    50,
		timeout:       12345,
		region:        "REGION",
	}

	err := PopulateTestFixture(userConfig, "RUN_ID", "AMI_ID")
	h.Ok(t, err)

	cwd, err := os.Getwd()
	h.Assert(t, err == nil, "Error getting the working directory")
	// Assert all the values were set
	h.Equals(t, "RUN_ID", testFixture.RunId())
	h.Equals(t, cwd+"/TEST_SUITE_NAME", testFixture.TestSuiteName())
	h.Equals(t, cwd+"/TEST_SUITE_NAME.tar.gz", testFixture.CompressedTestSuiteName())
	h.Equals(t, "BUCKET_NAME", testFixture.BucketName())
	h.Equals(t, "Instance-Qualifier-Run-RUN_ID", testFixture.BucketRootDir())
	h.Equals(t, 50, testFixture.TargetUtil())
	h.Equals(t, 12345, testFixture.Timeout())
	h.Equals(t, "qualifier-stack-RUN_ID", testFixture.CfnStackName())
	h.Equals(t, "final-results-RUN_ID.json", testFixture.FinalResultFilename())
	h.Equals(t, "instance-qualifier-RUN_ID.config", testFixture.UserConfigFilename())
	h.Equals(t, "qualifier-cfn-template-RUN_ID.json", testFixture.CfnTemplateFilename())
	h.Equals(t, "AMI_ID", testFixture.AmiId())
}

func TestPopulateTestFixtureForResumedRun(t *testing.T) {
	resetTestFixture()
	userConfig := UserConfig{
		bucket:  "BUCKET_NAME",
		timeout: 3600,
	}

	err := PopulateTestFixture(userConfig, "RUN_ID")
	h.Ok(t, err)

	// Assert all the values were set
	h.Equals(t, "RUN_ID", testFixture.RunId())
	h.Equals(t, "", testFixture.TestSuiteName())
	h.Equals(t, "", testFixture.CompressedTestSuiteName())
	h.Equals(t, "BUCKET_NAME", testFixture.BucketName())
	h.Equals(t, "Instance-Qualifier-Run-RUN_ID", testFixture.BucketRootDir())
	h.Equals(t, 0, testFixture.TargetUtil())
	h.Equals(t, 0, testFixture.Timeout())
	h.Equals(t, "qualifier-stack-RUN_ID", testFixture.CfnStackName())
	h.Equals(t, "final-results-RUN_ID.json", testFixture.FinalResultFilename())
	h.Equals(t, "instance-qualifier-RUN_ID.config", testFixture.UserConfigFilename())
	h.Equals(t, "qualifier-cfn-template-RUN_ID.json", testFixture.CfnTemplateFilename())
	h.Equals(t, "", testFixture.AmiId())
}

func TestParseCliArgsEnvSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Setenv("AWS_DEFAULT_REGION", "AWS_DEFAULT_REGION")
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--target-utilization=30",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	// Assert the value was set
	h.Equals(t, "AWS_DEFAULT_REGION", userConfig.Region())

	// Check the env var was set
	value, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	h.Equals(t, true, ok)
	h.Equals(t, "AWS_DEFAULT_REGION", value)
}

func TestParseCliArgsSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--target-utilization=30",
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
	h.Equals(t, "INSTANCE_TYPES", userConfig.InstanceTypes())
	h.Equals(t, "TEST_SUITE", userConfig.TestSuiteName())
	h.Equals(t, 30, userConfig.TargetUtil())
	h.Equals(t, "VPC", userConfig.VpcId())
	h.Equals(t, "SUBNET", userConfig.SubnetId())
	h.Equals(t, "AMI", userConfig.AmiId())
	h.Equals(t, 12345, userConfig.Timeout())
	h.Equals(t, true, userConfig.Persist())
	h.Equals(t, "PROFILE", userConfig.Profile())
	h.Equals(t, "REGION", userConfig.Region())
	h.Equals(t, "BUCKET", userConfig.Bucket())
	h.Equals(t, "/path/to/script", userConfig.CustomScriptPath())
}

func TestParseCliArgsOnlyBucketSuccess(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--bucket=BUCKET",
	}
	userConfig, err := ParseCliArgs(outputStream)
	h.Ok(t, err)

	h.Equals(t, "BUCKET", userConfig.Bucket())
}

func TestParseCliArgsOverrides(t *testing.T) {
	resetFlagsForTest()
	os.Setenv("AWS_DEFAULT_REGION", "AWS_DEFAULT_REGION")
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--target-utilization=30",
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
	h.Equals(t, "INSTANCE_TYPES", userConfig.InstanceTypes())
	h.Equals(t, "TEST_SUITE", userConfig.TestSuiteName())
	h.Equals(t, 30, userConfig.TargetUtil())
	h.Equals(t, "VPC", userConfig.VpcId())
	h.Equals(t, "SUBNET", userConfig.SubnetId())
	h.Equals(t, "AMI", userConfig.AmiId())
	h.Equals(t, 12345, userConfig.Timeout())
	h.Equals(t, true, userConfig.Persist())
	h.Equals(t, "PROFILE", userConfig.Profile())
	h.Equals(t, "REGION", userConfig.Region())

	// Check that env var was set
	value, ok := os.LookupEnv("AWS_DEFAULT_REGION")
	h.Equals(t, true, ok)
	h.Equals(t, "AWS_DEFAULT_REGION", value)
}

func TestParseCliArgsMissingInstanceTypesFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--test-suite=TEST_SUITE",
		"--target-utilization=30",
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
		"--target-utilization=30",
		"--subnet=SUBNET",
		"--ami=AMI",
		"--timeout=12345",
		"--profile=PROFILE",
		"--region=REGION",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when test-suite not provided")
}

func TestParseCliArgsMissingTargetUtilizationFailure(t *testing.T) {
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
	h.Assert(t, err != nil, "Failed to return error when target-utilization not provided")
}

func TestParseCliArgsNonPositiveTargetUtilizationFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--target-utilization=-1",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when non-positive target utilization provided")
}

func TestParseCliArgsNonPositiveTimeoutFailure(t *testing.T) {
	resetFlagsForTest()
	os.Args = []string{
		"cmd",
		"--instance-types=INSTANCE_TYPES",
		"--test-suite=TEST_SUITE",
		"--target-utilization=123",
		"--timeout=0",
	}
	_, err := ParseCliArgs(outputStream)
	h.Assert(t, err != nil, "Failed to return error when non-positive timeout provided")
}

func TestWriteUserConfigSuccess(t *testing.T) {
	actualConfigFile := "actual.config"
	defer os.Remove(actualConfigFile)
	userConfig := UserConfig{
		instanceTypes: "INSTANCE_TYPES",
		testSuiteName: "TEST_SUITE",
		targetUtil:    50,
		vpcId:         "VPC_ID",
		timeout:       12345,
		region:        "us-east-2",
	}

	err := WriteUserConfig(userConfig, actualConfigFile)
	h.Ok(t, err)

	// Assert files contents are same
	expected, err := ioutil.ReadFile(configFilesPath + "/valid.config")
	h.Assert(t, err == nil, "Error reading config file valid.config")
	actual, err := ioutil.ReadFile(actualConfigFile)
	h.Assert(t, err == nil, "Error reading config file "+actualConfigFile)
	h.Equals(t, string(expected), string(actual))
}

func TestWriteUserConfigNonExistentFilePathFailure(t *testing.T) {
	userConfig := UserConfig{
		instanceTypes: "INSTANCE_TYPES",
		testSuiteName: "TEST_SUITE",
		targetUtil:    50,
		timeout:       12345,
	}

	err := WriteUserConfig(userConfig, "non-existent-folder/CONFIG.config")
	h.Assert(t, err != nil, "Failed to return error when file path doesn't exist")
}

func TestReadUserConfigSuccess(t *testing.T) {
	expected := UserConfig{
		instanceTypes: "INSTANCE_TYPES",
		testSuiteName: "TEST_SUITE",
		targetUtil:    50,
		vpcId:         "VPC_ID",
		timeout:       12345,
		region:        "us-east-2",
	}

	actual, err := ReadUserConfig(configFilesPath + "/valid.config")
	h.Ok(t, err)

	// Assert all fields are same
	h.Equals(t, expected.InstanceTypes(), actual.InstanceTypes())
	h.Equals(t, expected.TestSuiteName(), actual.TestSuiteName())
	h.Equals(t, expected.TargetUtil(), actual.TargetUtil())
	h.Equals(t, expected.VpcId(), actual.VpcId())
	h.Equals(t, expected.SubnetId(), actual.SubnetId())
	h.Equals(t, expected.AmiId(), actual.AmiId())
	h.Equals(t, expected.Timeout(), actual.Timeout())
	h.Equals(t, expected.Persist(), actual.Persist())
	h.Equals(t, expected.Profile(), actual.Profile())
	h.Equals(t, expected.Region(), actual.Region())
	h.Equals(t, expected.Bucket(), actual.Bucket())
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
