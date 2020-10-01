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
	CpuThreshold     int    `json:"cpu-threshold"`
	MemThreshold     int    `json:"mem-threshold"`
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
	RunId                   string `json:"runId"`
	TestSuiteName           string `json:"test-suite"`
	CompressedTestSuiteName string `json:"compressed-test-suite"`
	BucketName              string `json:"bucket-name"`
	BucketRootDir           string `json:"bucket-root-dir"`
	CpuThreshold            int    `json:"cpu-threshold"`
	MemThreshold            int    `json:"mem-threshold"`
	Timeout                 int    `json:"timeout"`
	CfnStackName            string `json:"stack-name"`
	FinalResultFilename     string `json:"final-results"`
	UserConfigFilename      string `json:"user-config"`
	CfnTemplateFilename     string `json:"cfn-template"`
	AmiId                   string `json:"ami"`
	StartTime               string `json:"start-time"`
}

var testFixture TestFixture
var userConfig UserConfig

// SetUserConfig sets empty fields of UserConfig to reqConfig
func (UserConfig) SetUserConfig(reqConfig UserConfig) {
	if userConfig.InstanceTypes == "" {
		userConfig.InstanceTypes = reqConfig.InstanceTypes
	}
	if userConfig.TestSuiteName == "" {
		userConfig.TestSuiteName = reqConfig.TestSuiteName
	}
	if userConfig.CpuThreshold <= 0 {
		userConfig.CpuThreshold = reqConfig.CpuThreshold
	}
	if userConfig.MemThreshold <= 0 {
		userConfig.MemThreshold = reqConfig.MemThreshold
	}
	if userConfig.VpcId == "" {
		userConfig.VpcId = reqConfig.VpcId
	}
	if userConfig.SubnetId == "" {
		userConfig.SubnetId = reqConfig.SubnetId
	}
	if userConfig.AmiId == "" {
		userConfig.AmiId = reqConfig.AmiId
	}
	if userConfig.Timeout == defaultTimeout && reqConfig.Timeout > 0 {
		userConfig.Timeout = reqConfig.Timeout
	}
	if userConfig.Persist != true {
		userConfig.Persist = reqConfig.Persist
	}
	if userConfig.Profile == "" {
		userConfig.Profile = reqConfig.Profile
	}
	if userConfig.Region == "" {
		userConfig.Region = reqConfig.Region
	}
	if userConfig.Bucket == "" {
		userConfig.Bucket = reqConfig.Bucket
	}
	if userConfig.CustomScriptPath == "" {
		userConfig.CustomScriptPath = reqConfig.CustomScriptPath
	}
	if userConfig.ConfigFilePath == "" {
		userConfig.ConfigFilePath = reqConfig.ConfigFilePath
	}
}
