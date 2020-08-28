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
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	templatesPath                = "../templates"
	userDataScriptSampleTemplate = templatesPath + "/user-data_sample.template"
	masterSampleTemplate         = templatesPath + "/master_sample.template"
)

var instances = []resources.Instance{
	{
		InstanceType: "m4.large",
		VCpus:        "2",
		Memory:       "8192",
		Os:           "Linux/UNIX",
		Architecture: "x86_64",
	},
	{
		InstanceType: "m4.xlarge",
		VCpus:        "4",
		Memory:       "16384",
		Os:           "Linux/UNIX",
		Architecture: "x86_64",
	},
}

var inputStream = os.Stdin
var outputStream = os.Stdout

// Helpers

func encodeTemplate(filename string, t *testing.T) string {
	template, err := ioutil.ReadFile(templatesPath + "/" + filename)
	h.Assert(t, err == nil, "Error reading "+filename)
	return base64.StdEncoding.EncodeToString(template)
}

func setEncodedTemplates(t *testing.T) {
	encodedMasterTemplate = encodeTemplate("master.template", t)
	encodedLaunchTemplateTemplate = encodeTemplate("launch-template.template", t)
	encodedInstanceTemplate = encodeTemplate("instance.template", t)
	encodedAutoScalingGroupTemplate = encodeTemplate("auto-scaling-group.template", t)
}

// getTimesWithBuffer returns an array of times incremented by buffer
func getTimesWithBuffer(buffer int, timeout int) []string {
	var results []string
	for i := 0; i <= buffer; i++ {
		startTime := time.Now().UTC().Add(time.Second * time.Duration(timeout+timeBuffer+i))
		results = append(results, startTime.Format(time.RFC3339))
	}
	return results
}

func removeStartTimeFromTemplate(template string) string {
	// $startTime in the auto scaling group template is decided by the current time, so remove it to make comparison feasible
	rfc3339Regex := regexp.MustCompile("[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z")
	return rfc3339Regex.ReplaceAllString(template, "")
}

func prepareInput(input string) (*os.File, error) {
	tempFile, err := ioutil.TempFile("", "temp-file")
	if err != nil {
		return nil, err
	}
	if _, err := tempFile.WriteString(input); err != nil {
		return nil, err
	}
	if _, err := tempFile.Seek(0, 0); err != nil {
		return nil, err
	}

	return tempFile, nil
}

// Tests

func TestPopulateInstanceTemplate(t *testing.T) {
	setEncodedTemplates(t)
	numInstances := 2
	expectedVals := []string{"instance0", "instance1", "launchTemplate0", "launchTemplate1"}
	actual, err := populateInstanceTemplate(numInstances)
	h.Assert(t, err == nil, "Error calling populateInstanceTemplate")
	for _, ev := range expectedVals {
		h.Assert(t, strings.Contains(actual, ev), "Error: could not find "+ev+" in instance template")
	}
}

func TestPopulateLaunchTemplate(t *testing.T) {
	setEncodedTemplates(t)
	allInstanceTypes := "m4.large,m4.xlarge"
	userDataScript, _ := ioutil.ReadFile(userDataScriptSampleTemplate)
	userScriptForTemplate := processRawUserData(string(userDataScript))
	expectedVals := []string{"launchTemplate0", "launchTemplate1", "m4.large", "m4.xlarge", userScriptForTemplate}
	actual, err := populateLaunchTemplateTemplate(instances, allInstanceTypes, "", inputStream, outputStream)
	h.Assert(t, err == nil, "Error calling populateLaunchTemplateTemplate")
	for _, ev := range expectedVals {
		h.Assert(t, strings.Contains(actual, ev), "Error: could not find "+ev+" in launch template")
	}
}

func TestPopulateASGTemplate(t *testing.T) {
	setEncodedTemplates(t)
	numberInstances := 2
	timeout := 3
	testBuffer := 5 //StartTime should be within 5sec
	expectedStartTimes := getTimesWithBuffer(testBuffer, timeout)
	actual, err := populateAutoScalingGroupTemplate(numberInstances, timeout)
	h.Assert(t, err == nil, "Error calling populateAutoScalingGroupTemplate")
	h.Assert(t, strings.Contains(actual, fmt.Sprint(numberInstances)), "Error: could not find instanceNum in ASG template")
	found := false
	for _, est := range expectedStartTimes {
		if strings.Contains(actual, est) {
			found = true
			break
		}
	}
	h.Assert(t, found, "Error: could not find valid StartTime in ASG template")
}

func TestGenerateCfnTemplateUnsupportedInstanceTypes_Proceed(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("y\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	setEncodedTemplates(t)
	allInstanceTypes := "m4.large,m4.xlarge,a1.large"

	expected, err := ioutil.ReadFile(masterSampleTemplate)
	h.Assert(t, err == nil, "Error reading "+masterSampleTemplate)

	actual, err := GenerateCfnTemplate(instances, allInstanceTypes, "us-east-2a", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, string(expected), removeStartTimeFromTemplate(actual))
}

func TestGenerateCfnTemplateUnsupportedInstanceTypes_NotProceedFailure(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("N\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	setEncodedTemplates(t)
	allInstanceTypes := "m4.large,m4.xlarge,a1.large"

	_, err = GenerateCfnTemplate(instances, allInstanceTypes, "", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when answering no to proceeding with the rest instance types")
}

func TestAppendTemplate(t *testing.T) {
	setEncodedTemplates(t)
	numInstances := 2
	timeout := 3

	template, err := populateAutoScalingGroupTemplate(numInstances, timeout)
	h.Assert(t, err == nil, "Error calling populateAutoScalingGroupTemplate")
	instanceTemplate, err := populateInstanceTemplate(numInstances)
	h.Assert(t, err == nil, "Error calling populateInstanceTemplate")
	template = appendTemplate(template, instanceTemplate)

	expectedVals := []string{"instance0", "instance1", "launchTemplate0", "launchTemplate1"}
	expectedStartTimes := getTimesWithBuffer(5, timeout)
	found := false

	for _, ev := range expectedVals {
		h.Assert(t, strings.Contains(template, ev), "Error: could not find "+ev+" in appended template")
	}
	h.Assert(t, strings.Contains(template, fmt.Sprint(numInstances)), "Error: could not find instanceNum in appended template")
	for _, est := range expectedStartTimes {
		if strings.Contains(template, est) {
			found = true
			break
		}
	}
	h.Assert(t, found, "Error: could not find valid StartTime in appended template")
}

func TestAppendTemplateToEmptyTemplate(t *testing.T) {
	setEncodedTemplates(t)
	numInstances := 2
	timeout := 3

	template, err := populateAutoScalingGroupTemplate(numInstances, timeout)
	h.Assert(t, err == nil, "Error calling populateAutoScalingGroupTemplate")

	actual := appendTemplate("", template)
	h.Equals(t, template, actual)
}

func TestExtractResourcesFromTemplate(t *testing.T) {
	numInstances := 2
	instanceTemplate, err := populateInstanceTemplate(numInstances)
	h.Assert(t, err == nil, "Error calling populateInstanceTemplate")
	expectedVals := []string{"instance0", "instance1", "launchTemplate0", "launchTemplate1"}
	actual := extractResourcesFromTemplate(instanceTemplate)

	for _, ev := range expectedVals {
		h.Assert(t, strings.Contains(actual, ev), "Error: could not find "+ev+" in extracted resources from instance template")
	}
	h.Assert(t, !strings.Contains(actual, "Resources"), "Error: extractResourcesFromTemplate should NOT contain 'Resources'")
}

func TestPopulateUserData(t *testing.T) {
	expected, err := ioutil.ReadFile(userDataScriptSampleTemplate)
	h.Assert(t, err == nil, "Error reading the user data file")

	actual := populateUserData(resources.Instance{
		InstanceType: "m4.large",
		VCpus:        "2",
		Memory:       "8192",
		Os:           "Linux/UNIX",
		Architecture: "x86_64",
	})
	h.Equals(t, string(expected), actual)
}