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

package cmdutil_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/cmdutil"
	h "github.com/awslabs/amazon-ec2-instance-qualifier/pkg/test"
)

const (
	compressTestFolder = "../../test/static/Compress/Folder"
)

var inputStream = os.Stdin
var outputStream = os.Stdout

// Helpers

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

func TestGetRandomString(t *testing.T) {
	randomStringRegex := regexp.MustCompile("^[a-z0-9]{15}$")

	randomString := cmdutil.GetRandomString()
	h.Equals(t, true, randomStringRegex.MatchString(randomString))
}

func TestCompressSuccess(t *testing.T) {
	folderName := filepath.Base(compressTestFolder)
	compressedFolder := folderName + ".tar.gz"
	defer os.Remove(compressedFolder)
	err := cmdutil.Compress(compressTestFolder, compressedFolder)
	h.Ok(t, err)

	// Assert the extracted archive has the expected structure
	expected := []string{"Folder", "Folder/0", "Folder/0/0", "Folder/0/1", "Folder/1", "Folder/2"}
	cmd := exec.Command("tar", "xf", compressedFolder)
	err = cmd.Run()
	defer os.RemoveAll(folderName)
	h.Assert(t, err == nil, "Error extracting "+compressedFolder)
	actual := make([]string, 0)
	err = filepath.Walk(folderName, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		actual = append(actual, path)
		return nil
	})
	h.Assert(t, err == nil, "Error walking through "+folderName)
	h.Equals(t, expected, actual)
}

func TestCompressNonExistentSrcFileFailure(t *testing.T) {
	err := cmdutil.Compress("non-existent-src-file", "COMPRESSED.tar.gz")
	h.Assert(t, err != nil, "Failed to return error when source file doesn't exist")
}

func TestCompressNonExistentSrcFilePathFailure(t *testing.T) {
	err := cmdutil.Compress("non-existent-folder/folder", "COMPRESSED.tar.gz")
	h.Assert(t, err != nil, "Failed to return error when source file path doesn't exist")
}

func TestCompressNonExistentDestFilePathFailure(t *testing.T) {
	err := cmdutil.Compress(compressTestFolder, "non-existent-folder/COMPRESSED.tar.gz")
	h.Assert(t, err != nil, "Failed to return error when dest file path doesn't exist")
}

func TestBoolPromptSuccess(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("invalid_answer\ny\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	// Prepare output
	outputStream, err := ioutil.TempFile("", "temp-file")
	defer os.Remove(outputStream.Name())
	h.Assert(t, err == nil, "Error preparing the output stream")

	answer, err := cmdutil.BoolPrompt("PROMPT_Y", inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, true, answer)

	// Assert prompts were as expected
	outputStream.Sync()
	actual, err := ioutil.ReadFile(outputStream.Name())
	h.Assert(t, err == nil, "Error reading outputs from output stream")
	// The first line asks the user to answer; the second line tells the user that the previous answer is invalid, please re-answer
	expected := `PROMPT_Y y/N
Invalid input. Please type y/N
`
	h.Equals(t, expected, string(actual))
}

func TestBoolPromptEOFFailure(t *testing.T) {
	_, err := cmdutil.BoolPrompt("PROMPT", inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when there is no input")
}

func TestOptionPromptSuccess(t *testing.T) {
	// Prepare input
	inputStream, err := prepareInput("invalid_answer\n5\n1\n")
	defer os.Remove(inputStream.Name())
	h.Assert(t, err == nil, "Error preparing the input stream")

	// Prepare output
	outputStream, err := ioutil.TempFile("", "temp-file")
	defer os.Remove(outputStream.Name())
	h.Assert(t, err == nil, "Error preparing the output stream")

	answer, err := cmdutil.OptionPrompt("PROMPT", 3, inputStream, outputStream)
	h.Ok(t, err)
	h.Equals(t, 1, answer)

	// Assert prompts were as expected
	outputStream.Sync()
	actual, err := ioutil.ReadFile(outputStream.Name())
	h.Assert(t, err == nil, "Error reading outputs from output stream")
	// The first line is the prompt; the second line asks the user to re-choose because the previous answer is invalid (not a number);
	// the third line asks the user to re-choose again because the previous answer is out of range
	expected := `PROMPT
Invalid option. Please choose again:
Invalid option. Please choose again:
`
	h.Equals(t, expected, string(actual))
}

func TestOptionPromptEOFFailure(t *testing.T) {
	_, err := cmdutil.OptionPrompt("PROMPT", 3, inputStream, outputStream)
	h.Assert(t, err != nil, "Failed to return error when there is no input")
}
