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
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"gopkg.in/ini.v1"
)

const (
	bucketRootDirPrefix   = "Instance-Qualifier-Run-"
	compressSuffix        = ".tar.gz"
	cfnStackNamePrefix    = "qualifier-stack-"
	finalResultPrefix     = "final-results-"
	userConfigFilePrefix  = "instance-qualifier-"
	cfnTemplateFilePrefix = "qualifier-cfn-template-"
	binName               = "ec2-instance-qualifier"
	defaultTimeout        = 3600
	defaultProfile        = "default"
	awsConfigFile         = "~/.aws/config"
	awsRegionEnvVar       = "AWS_REGION"
	defaultRegionEnvVar   = "AWS_DEFAULT_REGION"
)

// PopulateTestFixture populates the test fixture which contains constant information for the entire run.
func PopulateTestFixture(userConfig UserConfig, runId string, amiId ...string) (err error) {
	testFixture.runId = runId
	testFixture.bucketRootDir = bucketRootDirPrefix + testFixture.runId
	testFixture.cfnStackName = cfnStackNamePrefix + testFixture.runId
	testFixture.finalResultFilename = finalResultPrefix + testFixture.runId + ".json"
	testFixture.userConfigFilename = userConfigFilePrefix + testFixture.runId + ".config"
	testFixture.cfnTemplateFilename = cfnTemplateFilePrefix + testFixture.runId + ".json"

	if userConfig.bucket == "" { // new run
		testFixture.amiId = amiId[0]
		testFixture.testSuiteName, err = filepath.Abs(userConfig.testSuiteName)
		if err != nil {
			return err
		}
		testFixture.compressedTestSuiteName = testFixture.testSuiteName + compressSuffix
		testFixture.targetUtil = userConfig.targetUtil
		testFixture.timeout = userConfig.timeout
	} else { // resumed run
		testFixture.bucketName = userConfig.bucket
	}

	return nil
}

// GetTestFixture returns testFixture.
func GetTestFixture() TestFixture {
	return testFixture
}

// GetUserConfig returns userConfig.
func GetUserConfig() UserConfig {
	return userConfig
}

// SetTestFixtureBucketName sets bucketName of testFixture.
func SetTestFixtureBucketName(bucketName string) {
	testFixture.bucketName = bucketName
}

// ParseCliArgs parses CLI arguments and uses environment variables as fallback values for some flags.
func ParseCliArgs(outputStream *os.File) (UserConfig, error) {
	// Customize usage message
	flag.Usage = func() {
		longUsage := fmt.Sprintf(`%s is a CLI tool that automates testing on a range of EC2 instance types.
Provided with a test suite and a list of EC2 instance types, %s will then
run the input on all designated types, test against multiple metrics, and output the results
in a user friendly format`, binName, binName)
		examples := fmt.Sprintf(`./%s --instance-types=m4.large,c5.large,m4.xlarge --test-suite=path/to/test-folder --target-utilization=30 --vpc=vpc-294b9542 --subnet=subnet-4879bf23 --timeout=2400
./%s --instance-types=m4.xlarge,c1.large,c5.large --test-suite=path/to/test-folder --target-utilization=50 --profile=default
./%s --bucket=qualifier-bucket-123456789abcdef`, binName, binName, binName)
		fmt.Fprintf(outputStream,
			longUsage+"\n\n"+
				"Usage:\n"+
				"  "+binName+" [flags]\n\n"+
				"Examples:\n"+examples+"\n\n"+
				"Flags:\n",
		)
		flag.PrintDefaults()
	}

	flag.StringVar(&userConfig.instanceTypes, "instance-types", "", "[REQUIRED] comma-separated list of instance-types to test")
	flag.StringVar(&userConfig.testSuiteName, "test-suite", "", "[REQUIRED] folder containing test files to execute")
	flag.IntVar(&userConfig.targetUtil, "target-utilization", 0, "[REQUIRED] % of total resource used (CPU, Mem) benchmark (must be an integer). ex: 30 means instances using less than 30% CPU and Mem SUCCEED")
	flag.StringVar(&userConfig.customScriptPath, "custom-script", "", "[OPTIONAL] path to Bash script to be executed on instance-types BEFORE agent runs test-suite and monitoring")
	flag.StringVar(&userConfig.vpcId, "vpc", "", "[OPTIONAL] vpc id")
	flag.StringVar(&userConfig.subnetId, "subnet", "", "[OPTIONAL] subnet id")
	flag.StringVar(&userConfig.amiId, "ami", "", "[OPTIONAL] ami id")
	flag.IntVar(&userConfig.timeout, "timeout", defaultTimeout, "[OPTIONAL] max seconds for test-suite execution on instances") // default value will be automatically appended
	flag.BoolVar(&userConfig.persist, "persist", false, "[OPTIONAL] set to true if you'd like the tool to keep the CloudFormation stack after the run. Default is deleting the stack")
	flag.StringVar(&userConfig.profile, "profile", "", "[OPTIONAL] AWS CLI profile to use for credentials and config")
	flag.StringVar(&userConfig.region, "region", "", "[OPTIONAL] AWS Region to use for API requests")
	flag.StringVar(&userConfig.bucket, "bucket", "", "[OPTIONAL] the name of the bucket created in the last run. When provided with this flag, the CLI won't create new resources, but try to grab test results from the bucket. If you provide this flag, you don't need to specify any required flags")

	flag.Parse()

	// Validation
	if userConfig.bucket == "" {
		if userConfig.instanceTypes == "" {
			return userConfig, errors.New("you must provide a comma-separated list of instance-types")
		}
		if userConfig.targetUtil <= 0 {
			return userConfig, errors.New("you must provide a target-utilization greater than 0")
		}
		if userConfig.testSuiteName == "" {
			return userConfig, errors.New("you must provide a folder containing test files to execute")
		}
	}
	if userConfig.timeout <= 0 {
		return userConfig, errors.New("you must provide a timeout greater than 0")
	}
	setUserConfigRegion()

	fmt.Fprintf(outputStream, "Configuration used: %v\n", userConfig)

	return userConfig, nil
}

// WriteUserConfig writes user config to config file.
func WriteUserConfig(userConfig UserConfig, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	configString := fmt.Sprintf(`instance-types: %s
test-suite: %s
target-utilization: %d
vpc: %s
subnet: %s
ami: %s
timeout: %d
persist: %v
profile: %s
region: %s
bucket: %s`, userConfig.instanceTypes, userConfig.testSuiteName, userConfig.targetUtil, userConfig.vpcId, userConfig.subnetId,
		userConfig.amiId, userConfig.timeout, userConfig.persist, userConfig.profile, userConfig.region, userConfig.bucket)
	_, err = file.WriteString(configString)
	if err != nil {
		return err
	}
	file.Sync()

	return nil
}

// ReadUserConfig reads user config from config file.
func ReadUserConfig(filename string) (userConfig UserConfig, err error) {
	configByteData, err := ioutil.ReadFile(filename)
	if err != nil {
		return userConfig, err
	}

	scanner := bufio.NewScanner(strings.NewReader(string(configByteData)))
	for scanner.Scan() {
		attr := strings.Split(scanner.Text(), ": ")
		if len(attr) != 2 { // Check whether the line is a key-value pair
			return userConfig, fmt.Errorf("invalid user config: %s", scanner.Text())
		}
		if err := populateUserConfig(&userConfig, attr[0], attr[1]); err != nil {
			return userConfig, err
		}
	}
	if err := scanner.Err(); err != nil {
		return userConfig, err
	}

	return userConfig, nil
}

// populateUserConfig populates one field of the UserConfig struct from a key-value pair.
func populateUserConfig(userConfig *UserConfig, key string, value string) (err error) {
	switch key {
	case "instance-types":
		userConfig.instanceTypes = value
	case "test-suite":
		userConfig.testSuiteName = value
	case "target-utilization":
		userConfig.targetUtil, err = strconv.Atoi(value)
	case "custom-script":
		userConfig.customScriptPath = value
	case "vpc":
		userConfig.vpcId = value
	case "subnet":
		userConfig.subnetId = value
	case "ami":
		userConfig.amiId = value
	case "timeout":
		userConfig.timeout, err = strconv.Atoi(value)
	case "persist":
		userConfig.persist, err = strconv.ParseBool(value)
	case "profile":
		userConfig.profile = value
	case "region":
		userConfig.region = value
	case "bucket":
		userConfig.bucket = value
	default:
		err = fmt.Errorf("unknown user config")
	}

	return err
}

func getProfileRegion(profileName string) (string, error) {
	if profileName != defaultProfile {
		profileName = fmt.Sprintf("profile %s", profileName)
	}
	awsConfigPath, err := homedir.Expand(awsConfigFile)
	if err != nil {
		return "", fmt.Errorf("Warning: unable to find home directory to parse aws config file")
	}
	awsConfigIni, err := ini.Load(awsConfigPath)
	if err != nil {
		return "", fmt.Errorf("Warning: unable to load aws config file for profile at path: %s", awsConfigPath)
	}
	section, err := awsConfigIni.GetSection(profileName)
	if err != nil {
		return "", fmt.Errorf("Warning: there is no configuration for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	regionConfig, err := section.GetKey("region")
	if err != nil || regionConfig.String() == "" {
		return "", fmt.Errorf("Warning: there is no region configured for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	return regionConfig.String(), nil
}

func setUserConfigRegion() {
	if userConfig.region == "" {
		if userConfig.profile != "" {
			if profileRegion, err := getProfileRegion(userConfig.profile); err == nil {
				userConfig.region = profileRegion
			}
		} else if envRegion, ok := os.LookupEnv(awsRegionEnvVar); ok && envRegion != "" {
			userConfig.region = envRegion
		} else if defaultProfileRegion, err := getProfileRegion(defaultProfile); err == nil {
			userConfig.region = defaultProfileRegion
		} else if defaultRegion, ok := os.LookupEnv(defaultRegionEnvVar); ok && defaultRegion != "" {
			userConfig.region = defaultRegion
		} else {
			errorMsg := "Failed to determine region from the following sources: \n"
			errorMsg = errorMsg + "\t - --region flag\n"
			if userConfig.profile != "" {
				errorMsg = errorMsg + fmt.Sprintf("\t - profile region in %s\n", awsConfigFile)
			}
			errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", awsRegionEnvVar)
			errorMsg = errorMsg + fmt.Sprintf("\t - default profile region in %s\n", awsConfigFile)
			errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", defaultRegionEnvVar)
			fmt.Println(errorMsg)
		}
	}
}
