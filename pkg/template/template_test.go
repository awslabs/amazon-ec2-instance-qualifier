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
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	templatesPath = "../../test/static/Templates"
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

func readFileInTemplatesDir(filename string, t *testing.T) string {
	template, err := ioutil.ReadFile(templatesPath + "/" + filename)
	h.Assert(t, err == nil, "Error reading "+filename)
	return string(template)
}

func setEncodedTemplates(t *testing.T) {
	encodedMasterTemplate = readFileInTemplatesDir("encoded_master", t)
	encodedLaunchTemplateTemplate = readFileInTemplatesDir("encoded_launch_template", t)
	encodedInstanceTemplate = readFileInTemplatesDir("encoded_instance", t)
	encodedAutoScalingGroupTemplate = readFileInTemplatesDir("encoded_auto_scaling_group", t)
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

func TestGenerateCfnTemplateAllInstanceTypesSupported(t *testing.T) {
	setEncodedTemplates(t)
	allInstanceTypes := "m4.large,m4.xlarge"

	expected := readFileInTemplatesDir("existing_vpc.template", t)
	actual, err := GenerateCfnTemplate(instances, allInstanceTypes, "", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, expected, removeStartTimeFromTemplate(actual))
}

func TestGenerateCfnTemplateUnsupportedInstanceTypes_Proceed(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("y\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	setEncodedTemplates(t)
	allInstanceTypes := "m4.large,m4.xlarge,a1.large"

	expected := readFileInTemplatesDir("new_vpc.template", t)
	actual, err := GenerateCfnTemplate(instances, allInstanceTypes, "us-east-2a", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, expected, removeStartTimeFromTemplate(actual))
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
	existingTemplate := readFileInTemplatesDir("old.template", t)
	templateToAppend := readFileInTemplatesDir("template_to_append.template", t)
	expected := readFileInTemplatesDir("new.template", t)

	actual := appendTemplate(existingTemplate, templateToAppend)
	h.Equals(t, expected, actual)
}

func TestAppendTemplateToEmptyTemplate(t *testing.T) {
	templateToAppend := readFileInTemplatesDir("template_to_append.template", t)

	actual := appendTemplate("", templateToAppend)
	h.Equals(t, templateToAppend, actual)
}

func TestExtractResourcesFromTemplate(t *testing.T) {
	existingTemplate := readFileInTemplatesDir("old.template", t)
	expected := readFileInTemplatesDir("extracted_resources", t)

	actual := extractResourcesFromTemplate(existingTemplate)
	h.Equals(t, expected, actual)
}
