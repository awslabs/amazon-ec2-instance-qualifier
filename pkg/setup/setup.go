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

package setup

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	agentBin           = "agent"
	monitorCpuFilename = "monitor-cpu.sh"
	monitorMemFilename = "monitor-mem.sh"
)

// DO NOT EDIT: these values are populated by the Makefile
var (
	encodedMonitorCpuScript string
	encodedMonitorMemScript string
)

// UserScript encapsulates the data required for creating user script that will be deployed to the instance(s)
type UserScript struct {
	InstanceType, VCpus, Memory, Os, Architecture, BucketName, Timeout, BucketRootDir, TargetUtil, CompressedTestSuiteName, TestSuiteName string
}

// SetTestSuite copies agent scripts to test suite, compresses test suite into a tarball, then removes agent
// scripts from test suite.
func SetTestSuite() error {
	testFixture := config.GetTestFixture()
	if err := copyAgentScriptsToTestSuite(testFixture.TestSuiteName()); err != nil {
		return err
	}
	if err := cmdutil.Compress(testFixture.TestSuiteName(), testFixture.CompressedTestSuiteName()); err != nil {
		return err
	}
	if err := removeAgentScriptsFromTestSuite(testFixture.TestSuiteName()); err != nil {
		// Failing to remove is a not fatal error
		log.Println(err)
	}

	return nil
}

// GetUserData returns the userdata script used for the launching of an instance.
func GetUserData(instance resources.Instance) (script string) {
	testFixture := config.GetTestFixture()
	testSuiteName := filepath.Base(testFixture.TestSuiteName())
	compressedTestSuiteName := filepath.Base(testFixture.CompressedTestSuiteName())
	filePrefix, _ := filepath.Abs("../templates/")
	t := template.Must(template.ParseFiles(filePrefix + "/user-data.template"))

	userScript := UserScript{
		InstanceType:            instance.InstanceType,
		VCpus:                   instance.VCpus,
		Memory:                  instance.Memory,
		Os:                      instance.Os,
		Architecture:            instance.Architecture,
		BucketName:              testFixture.BucketName(),
		Timeout:                 fmt.Sprint(testFixture.Timeout()),
		BucketRootDir:           testFixture.BucketRootDir(),
		TargetUtil:              fmt.Sprint(testFixture.TargetUtil()),
		CompressedTestSuiteName: compressedTestSuiteName,
		TestSuiteName:           testSuiteName,
	}
	var byteBuffer bytes.Buffer
	err := t.Execute(&byteBuffer, userScript)
	if err != nil {
		log.Println("There was an error generating user data script: ", err)
		return ""
	}

	script = byteBuffer.String()
	return script
}

// IsInstanceQualifierScript checks whether a file is an internal script file of the instance-qualifier.
func IsInstanceQualifierScript(filename string) bool {
	if filename == agentBin || filename == monitorCpuFilename || filename == monitorMemFilename {
		return true
	}
	return false
}

func copyAgentScriptsToTestSuite(testSuiteName string) error {
	monitorCpuScript, err := cmdutil.DecodeBase64(encodedMonitorCpuScript)
	if err != nil {
		return err
	}
	monitorMemScript, err := cmdutil.DecodeBase64(encodedMonitorMemScript)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(testSuiteName+"/"+monitorCpuFilename, []byte(monitorCpuScript), 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(testSuiteName+"/"+monitorMemFilename, []byte(monitorMemScript), 0644); err != nil {
		return err
	}

	src, err := os.Open(agentBin)
	if err != nil {
		return err
	}
	defer src.Close()
	dest, err := os.Create(testSuiteName + "/" + agentBin)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, src)
	if err != nil {
		return err
	}

	log.Printf("All required scripts successfully copied to %s\n", testSuiteName)

	return nil
}

func removeAgentScriptsFromTestSuite(testSuiteName string) error {
	files, err := ioutil.ReadDir(testSuiteName)
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		if IsInstanceQualifierScript(filename) {
			if err = os.Remove(testSuiteName + "/" + filename); err != nil {
				log.Println(err)
			}
		}
	}
	if err == nil {
		log.Printf("%s successfully cleand up\n", testSuiteName)
	}

	return nil
}
