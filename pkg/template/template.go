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

package template

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
)

const (
	timeBuffer   = 600 // 10 min
	resourcesKey = "Resources"
)

// DO NOT EDIT: these values are populated by the Makefile
var (
	encodedMasterTemplate           string
	encodedLaunchTemplateTemplate   string
	encodedAutoScalingGroupTemplate string
	encodedInstanceTemplate         string
	encodedUserData                 string
)

// UserScript encapsulates the data required for creating user script that will be deployed to the instance(s)
type UserScript struct {
	InstanceType, VCpus, Memory, Os, Architecture, BucketName, Timeout, BucketRootDir, CompressedTestSuiteName, TestSuiteName, CustomScript, Region string
}

// GenerateCfnTemplate returns the CloudFormation template used to create resources for instance-qualifier.
func GenerateCfnTemplate(instances []resources.Instance, allInstanceTypes string, availabilityZone string, inputStream *os.File, outputStream *os.File) (template string, err error) {
	testFixture := config.GetTestFixture()
	template, err = populateMasterTemplate(availabilityZone)
	if err != nil {
		return "", err
	}

	launchTemplateTemplate, err := populateLaunchTemplateTemplate(instances, allInstanceTypes, testFixture.AmiId, inputStream, outputStream)
	if err != nil {
		return "", err
	}
	template = appendTemplate(template, launchTemplateTemplate)

	autoScalingGroupTemplate, err := populateAutoScalingGroupTemplate(len(instances), testFixture.Timeout)
	if err != nil {
		return "", err
	}
	template = appendTemplate(template, autoScalingGroupTemplate)

	instanceTemplate, err := populateInstanceTemplate(len(instances))
	if err != nil {
		return "", err
	}
	template = appendTemplate(template, instanceTemplate)

	log.Println("Successfully generated the final CloudFormation template")

	return template, nil
}

// appendTemplate appends the "Resources" object of one template to the "Resources" object of the other.
func appendTemplate(existingTemplate string, templateToAppend string) (newTemplate string) {
	if existingTemplate == "" {
		return templateToAppend
	}

	oldResources := extractResourcesFromTemplate(existingTemplate)
	appendedResources := extractResourcesFromTemplate(templateToAppend)
	newResources := oldResources + ",\n" + appendedResources
	newTemplate = strings.Replace(existingTemplate, oldResources, newResources, 1)

	return newTemplate
}

// extractResourcesFromTemplate extracts the "Resources" object from the template.
func extractResourcesFromTemplate(template string) string {
	keyIdx := strings.Index(template, resourcesKey)

	var startIdx int
	for i := keyIdx; i < len(template); i++ {
		if template[i] == '{' {
			startIdx = i
			break
		}
	}
	raw := template[startIdx:]

	unclosedBracket := 0
	var endIdx int
	for i, c := range raw {
		if c == '{' {
			unclosedBracket++
		} else if c == '}' {
			unclosedBracket--
			if unclosedBracket == 0 {
				endIdx = i
				break
			}
		}
	}

	// Use [2:] to remove the leading "{\n"
	resourcesString := strings.TrimSpace(raw[:endIdx])[2:]
	return resourcesString
}

// processRawUserData converts some escape characters to literals.
func processRawUserData(rawUserData string) string {
	var userData strings.Builder
	for _, c := range rawUserData {
		if c == '"' {
			userData.WriteString("\\\"")
		} else if c == '\n' {
			userData.WriteString("\\n")
		} else if c == '\t' {
			userData.WriteString("\\t")
		} else {
			userData.WriteRune(c)
		}
	}

	return userData.String()
}

// populateMasterTemplate populates the Master template with the correct value, and returns it.
func populateMasterTemplate(availabilityZone string) (template string, err error) {
	rawTemplate, err := cmdutil.DecodeBase64(encodedMasterTemplate)
	if err != nil {
		return "", err
	}

	if availabilityZone != "" {
		template = strings.ReplaceAll(rawTemplate, "$availabilityZone", availabilityZone)
	} else {
		// An existent VPC infrastructure will be used, so no need to populate the availabilityZone value
		template = rawTemplate
	}
	log.Println("Successfully populated the Master template")

	return template, nil
}

// populateLaunchTemplateTemplate populates the CloudFormation template of launch templates with the correct
// values, merges all, and returns the generated template.
func populateLaunchTemplateTemplate(instances []resources.Instance, allInstanceTypes string, amiId string, inputStream *os.File, outputStream *os.File) (template string, err error) {
	rawTemplate, err := cmdutil.DecodeBase64(encodedLaunchTemplateTemplate)
	if err != nil {
		return "", err
	}

	for i, instance := range instances {
		processedTemplate := strings.ReplaceAll(rawTemplate, "$idx", strconv.Itoa(i))
		processedTemplate = strings.ReplaceAll(processedTemplate, "$amiId", amiId)
		processedTemplate = strings.ReplaceAll(processedTemplate, "$instanceType", instance.InstanceType)
		processedTemplate = strings.ReplaceAll(processedTemplate, "$userData", processRawUserData(populateUserData(instance)))

		template = appendTemplate(template, processedTemplate)
	}

	supportedInstanceTypes, unsupportedInstanceTypes := classifyInstanceTypes(instances, allInstanceTypes)
	if len(unsupportedInstanceTypes) > 0 {
		prompt := fmt.Sprintf("Instance types %v are not supported due to AMI or Availability Zone. Do you want to proceed with the rest instance types %v ?", unsupportedInstanceTypes, supportedInstanceTypes)
		answer, err := cmdutil.BoolPrompt(prompt, inputStream, outputStream)
		if err != nil {
			return "", err
		}
		if !answer {
			return "", fmt.Errorf("failed to proceed due to unsupported instance types")
		}
	}

	log.Println("Successfully generated the CloudFormation template of launch templates")

	return template, nil
}

// classifyInstanceTypes classifies instance types to supported and unsupported.
func classifyInstanceTypes(supportedInstances []resources.Instance, allInstanceTypes string) (supportedInstanceTypes []string, unsupportedInstanceTypes []string) {
	for _, instance := range supportedInstances {
		supportedInstanceTypes = append(supportedInstanceTypes, instance.InstanceType)
	}

	for _, instanceType := range strings.Split(allInstanceTypes, ",") {
		isFound := false
		for _, supportedInstanceType := range supportedInstanceTypes {
			if instanceType == supportedInstanceType {
				isFound = true
				break
			}
		}

		if !isFound {
			unsupportedInstanceTypes = append(unsupportedInstanceTypes, instanceType)
		}
	}

	return supportedInstanceTypes, unsupportedInstanceTypes
}

// populateAutoScalingGroupTemplate populates the CloudFormation template of the auto scaling group with the
// correct values, and returns it.
func populateAutoScalingGroupTemplate(instanceNum int, timeout int) (string, error) {
	rawTemplate, err := cmdutil.DecodeBase64(encodedAutoScalingGroupTemplate)
	if err != nil {
		return "", err
	}

	processedTemplate := strings.ReplaceAll(rawTemplate, "$instanceNum", strconv.Itoa(instanceNum))
	startTime := time.Now().UTC().Add(time.Second * time.Duration(timeout+timeBuffer))
	processedTemplate = strings.ReplaceAll(processedTemplate, "$startTime", startTime.Format(time.RFC3339))

	log.Println("Successfully generated the CloudFormation template of the auto scaling group")

	return processedTemplate, nil
}

// populateInstanceTemplate populates the CloudFormation template of instances with the correct values, merges
// all, and returns the generated template.
func populateInstanceTemplate(instanceNum int) (template string, err error) {
	rawTemplate, err := cmdutil.DecodeBase64(encodedInstanceTemplate)
	if err != nil {
		return "", err
	}

	for idx := 0; idx < instanceNum; idx++ {
		processedTemplate := strings.ReplaceAll(rawTemplate, "$idx", strconv.Itoa(idx))
		template = appendTemplate(template, processedTemplate)
	}

	log.Println("Successfully generated the CloudFormation template of instances")

	return template, nil
}

// populateUserData populates the userdata script template used for the launching of an instance.
func populateUserData(instance resources.Instance) string {
	testFixture := config.GetTestFixture()
	userConfig := config.GetUserConfig()
	testSuiteName := filepath.Base(testFixture.TestSuiteName)
	compressedTestSuiteName := filepath.Base(testFixture.CompressedTestSuiteName)
	userDataTemplate, err := cmdutil.DecodeBase64(encodedUserData)
	if err != nil {
		log.Println("Error decoding user data: ", err)
		return ""
	}
	var customScript []byte
	customScriptPath := userConfig.CustomScriptPath
	if customScriptPath != "" {
		customScript, err = ioutil.ReadFile(customScriptPath)
		if err != nil {
			log.Println("There was an error extracting custom script: ", err)
			return ""
		}
	} else {
		log.Println("no customScriptPath provided")
	}

	t := template.Must(template.New("").Parse(userDataTemplate))

	userScript := UserScript{
		InstanceType:            instance.InstanceType,
		VCpus:                   instance.VCpus,
		Memory:                  instance.Memory,
		Os:                      instance.Os,
		Architecture:            instance.Architecture,
		BucketName:              testFixture.BucketName,
		Timeout:                 fmt.Sprint(testFixture.Timeout),
		BucketRootDir:           testFixture.BucketRootDir,
		CompressedTestSuiteName: compressedTestSuiteName,
		TestSuiteName:           testSuiteName,
		CustomScript:            string(customScript),
		Region:                  userConfig.Region,
	}
	var byteBuffer bytes.Buffer
	err = t.Execute(&byteBuffer, userScript)
	if err != nil {
		log.Println("There was an error generating user data script: ", err)
		return ""
	}

	return byteBuffer.String()
}
