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
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

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
	testFixture.RunId = runId
	testFixture.BucketRootDir = bucketRootDirPrefix + testFixture.RunId
	testFixture.CfnStackName = cfnStackNamePrefix + testFixture.RunId
	testFixture.FinalResultFilename = finalResultPrefix + testFixture.RunId + ".json"
	testFixture.UserConfigFilename = userConfigFilePrefix + testFixture.RunId + ".config"
	testFixture.CfnTemplateFilename = cfnTemplateFilePrefix + testFixture.RunId + ".json"
	testFixture.AmiId = amiId[0]
	testFixture.TestSuiteName, err = filepath.Abs(userConfig.TestSuiteName)
	if err != nil {
		return err
	}
	testFixture.CompressedTestSuiteName = testFixture.TestSuiteName + compressSuffix
	testFixture.CpuThreshold = userConfig.CpuThreshold
	testFixture.MemThreshold = userConfig.MemThreshold
	testFixture.Timeout = userConfig.Timeout
	testFixture.StartTime = time.Now().Format(time.RFC3339)

	return nil
}

// RestoreTestFixture populates the test fixture from a previous state
func RestoreTestFixture(data []byte) (err error) {
	if err := json.Unmarshal(data, &testFixture); err != nil {
		return err
	}
	log.Printf("Restored test fixture to: %v\n", testFixture)
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
	testFixture.BucketName = bucketName
}

// ParseCliArgs parses CLI arguments and uses environment variables as fallback values for some flags.
func ParseCliArgs(outputStream *os.File) (UserConfig, error) {
	// Customize usage message
	flag.Usage = func() {
		longUsage := fmt.Sprintf(`%s is a CLI tool that automates testing on a range of EC2 instance types.
Provided with a test suite and a list of EC2 instance types, %s will then
run the input on all designated types, test against multiple metrics, and output the results
in a user friendly format`, binName, binName)
		examples := fmt.Sprintf(`./%s --instance-types=m4.large,c5.large,m4.xlarge --test-suite=path/to/test-folder --cpu-threshold=30 --mem-threshold=30 --vpc=vpc-294b9542 --subnet=subnet-4879bf23 --timeout=2400
./%s --instance-types=m4.xlarge,c1.large,c5.large --test-suite=path/to/test-folder --cpu-threshold=30 --mem-threshold=30 --profile=default
./%s --bucket=qualifier-Bucket-123456789abcdef`, binName, binName, binName)
		fmt.Fprintf(outputStream,
			longUsage+"\n\n"+
				"Usage:\n"+
				"  "+binName+" [flags]\n\n"+
				"Examples:\n"+examples+"\n\n"+
				"Flags:\n",
		)
		flag.PrintDefaults()
	}

	flag.StringVar(&userConfig.InstanceTypes, "instance-types", "", "[REQUIRED] comma-separated list of instance-types to test")
	flag.StringVar(&userConfig.TestSuiteName, "test-suite", "", "[REQUIRED] folder containing test files to execute")
	flag.IntVar(&userConfig.CpuThreshold, "cpu-threshold", 0, "[REQUIRED] % cpu utilization that should not be exceeded measured by cpu_usage_active. ex: 30 means instances using 30% or less CPU SUCCEED")
	flag.IntVar(&userConfig.MemThreshold, "mem-threshold", 0, "[REQUIRED] % of memory used that should not be exceeded measured by mem_used_percent. ex: 30 means instances using 30% or less MEM SUCCEED")
	flag.StringVar(&userConfig.ConfigFilePath, "config-file", "", "[OPTIONAL] path to config file for cli input parameters in JSON")
	flag.StringVar(&userConfig.CustomScriptPath, "custom-script", "", "[OPTIONAL] path to Bash script to be executed on instance-types BEFORE agent runs test-suite and monitoring")
	flag.StringVar(&userConfig.VpcId, "vpc", "", "[OPTIONAL] vpc id")
	flag.StringVar(&userConfig.SubnetId, "subnet", "", "[OPTIONAL] subnet id")
	flag.StringVar(&userConfig.AmiId, "ami", "", "[OPTIONAL] ami id")
	flag.IntVar(&userConfig.Timeout, "timeout", defaultTimeout, "[OPTIONAL] max seconds for test-suite execution on instances") // default value will be automatically appended
	flag.BoolVar(&userConfig.Persist, "persist", false, "[OPTIONAL] set to true if you'd like the tool to keep the CloudFormation stack after the run. Default is deleting the stack")
	flag.StringVar(&userConfig.Profile, "profile", "", "[OPTIONAL] AWS CLI Profile to use for credentials and config")
	flag.StringVar(&userConfig.Region, "region", "", "[OPTIONAL] AWS Region to use for API requests")
	flag.StringVar(&userConfig.Bucket, "bucket", "", "[OPTIONAL] the name of the Bucket created in the last run. When provided with this flag, the CLI won't create new resources, but try to grab test results from the Bucket. If you provide this flag, you don't need to specify any required flags")
	flag.BoolVar(&userConfig.IsDemo, "demo", false, "[OPTIONAL] set to true if you'd like the tool to execute against the demo app, app-for-e2e")

	// Apply config with precedence: cli args, env vars, config file
	flag.Parse()

	if userConfig.IsDemo {
		userConfig.SetDemoConfig()
		return userConfig, nil
	}

	setUserConfigRegion()
	var configFile string
	if userConfig.ConfigFilePath != "" {
		configFile = userConfig.ConfigFilePath
		tmpConfig, err := ReadUserConfig(userConfig.ConfigFilePath)
		userConfig.SetUserConfig(tmpConfig)
		if err != nil {
			return userConfig, err
		}
	}

	if userConfig.Region == "" {
		errorMsg := "Failed to determine region from the following sources: \n"
		errorMsg = errorMsg + "\t - --region flag\n"
		if userConfig.Profile != "" {
			errorMsg = errorMsg + fmt.Sprintf("\t - profile region in %s\n", awsConfigFile)
		}
		errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", awsRegionEnvVar)
		errorMsg = errorMsg + fmt.Sprintf("\t - default profile region in %s\n", awsConfigFile)
		errorMsg = errorMsg + fmt.Sprintf("\t - %s environment variable\n", defaultRegionEnvVar)
		return userConfig, fmt.Errorf(errorMsg)
	}

	// preserve this data if config file defines "config-file" as nil
	if configFile != "" {
		userConfig.ConfigFilePath = configFile
	}

	// Validation
	if userConfig.Bucket == "" {
		if userConfig.InstanceTypes == "" {
			return userConfig, errors.New("you must provide a comma-separated list of instance-types")
		}
		if userConfig.CpuThreshold <= 0 {
			return userConfig, errors.New("you must provide a cpu-threshold greater than 0")
		}
		if userConfig.MemThreshold <= 0 {
			return userConfig, errors.New("you must provide a mem-threshold greater than 0")
		}
		if userConfig.TestSuiteName == "" {
			return userConfig, errors.New("you must provide a folder containing test files to execute")
		}
	}
	if userConfig.Timeout <= 0 {
		return userConfig, errors.New("you must provide a timeout greater than 0")
	}
	log.Printf("Starting Instance-Qualifier with User Config: %s\n", userConfig.String())
	return userConfig, nil
}

// WriteUserConfig writes user config to config file.
func WriteUserConfig(filename string) error {
	configJson, err := json.MarshalIndent(userConfig, "", "\t")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, configJson, 0644)
	if err != nil {
		return err
	}
	return nil
}

// ReadUserConfig reads user config from config file.
func ReadUserConfig(filename string) (UserConfig, error) {
	result := UserConfig{}
	configByteData, err := ioutil.ReadFile(filename)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(configByteData, &result)
	if err != nil {
		fmt.Println("error while processing config file: ", err)
		return result, err
	}
	return result, nil
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
		return "", fmt.Errorf("Warning: unable to load aws config file for Profile at path: %s", awsConfigPath)
	}
	section, err := awsConfigIni.GetSection(profileName)
	if err != nil {
		return "", fmt.Errorf("Warning: there is no configuration for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	regionConfig, err := section.GetKey("Region")
	if err != nil || regionConfig.String() == "" {
		return "", fmt.Errorf("Warning: there is no region configured for the specified aws profile %s at %s", profileName, awsConfigPath)
	}
	return regionConfig.String(), nil
}

func setUserConfigRegion() {
	if userConfig.Region == "" {
		if userConfig.Profile != "" {
			if profileRegion, err := getProfileRegion(userConfig.Profile); err == nil {
				userConfig.Region = profileRegion
			}
		} else if envRegion, ok := os.LookupEnv(awsRegionEnvVar); ok && envRegion != "" {
			userConfig.Region = envRegion
		} else if defaultProfileRegion, err := getProfileRegion(defaultProfile); err == nil {
			userConfig.Region = defaultProfileRegion
		} else if defaultRegion, ok := os.LookupEnv(defaultRegionEnvVar); ok && defaultRegion != "" {
			userConfig.Region = defaultRegion
		}
	}
}
