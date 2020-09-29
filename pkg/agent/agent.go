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

package agent

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup"
)

const (
	// If the test file finishes execution too fast, we should wait for some time before killing the monitoring script to ensure the output file has been created
	waitUntilFileExistTime = 1
	resultSuccess          = "pass"
	resultFail             = "fail"
)

// GetTestFileList returns the list of test files in the test suite.
func GetTestFileList(scriptPath string) (testFileList []string, err error) {
	files, err := ioutil.ReadDir(scriptPath)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if isValidTestFile(file) {
			testFileList = append(testFileList, scriptPath+"/"+file.Name())
		}
	}

	return testFileList, nil
}

// PopulateResult takes a test file, executes it, then persists the pass/fail result and execution time
func PopulateResult(filename string, agentFixture AgentFixture, outputStream *os.File, errStream *os.File) (testResult resources.Result) {
	testResult.Label = filepath.Base(filename)
	testResult.Metrics = make([]resources.Metric, 0)

	success, execTime := execute(filename, agentFixture.ScriptPath, outputStream, errStream)
	if success {
		testResult.Status = resultSuccess
		fmt.Fprintf(outputStream, "\n------------------------------------------------------------------------------------------------------\n")
		fmt.Fprintf(outputStream, "✅ %s passed!\n", filename)
		fmt.Fprintf(outputStream, "------------------------------------------------------------------------------------------------------\n\n")
	} else {
		testResult.Status = resultFail
		fmt.Fprintf(outputStream, "\n------------------------------------------------------------------------------------------------------\n")
		fmt.Fprintf(outputStream, "❌ %s failed!\n", filename)
		fmt.Fprintf(outputStream, "------------------------------------------------------------------------------------------------------\n\n")
	}

	testResult.ExecutionTime = fmt.Sprintf("%.3f", execTime)
	return testResult
}

// TerminateInstance terminates the instance.
func TerminateInstance() {
	cmd := exec.Command("shutdown", "-h", "now")
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}

// Fatal logs the fatal error, uploads the log to the bucket, then terminates the instance.
func Fatal(sess *session.Session, agentFixture AgentFixture, err error) {
	svc := resources.New(sess)

	if err != nil {
		log.Println(err)
	}
	remoteLogFilename := agentFixture.BucketDir + "/" + filepath.Base(agentFixture.LogFilename)
	if err := svc.UploadToBucket(agentFixture.BucketName, agentFixture.LogFilename, remoteLogFilename); err != nil {
		log.Println(err)
	}
	TerminateInstance()
}

func isValidTestFile(file os.FileInfo) bool {
	filename := file.Name()
	if file.IsDir() || setup.IsInstanceQualifierScript(filename) || strings.HasSuffix(filename, ".load") || strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".log") || strings.HasSuffix(filename, ".rpm") {
		return false
	}
	return true
}

// execute executes the test file, then returns the test result and execution time
func execute(filename string, scriptPath string, outputStream *os.File, errStream *os.File) (success bool, execTime float64) {
	fmt.Fprintf(outputStream, "\n======================================================================================================\n")
	fmt.Fprintf(outputStream, "🥑 Starting %s\n", filename)
	fmt.Fprintf(outputStream, "======================================================================================================\n")

	// Run the test file
	cmd := exec.Command(filename)
	cmd.Stdout = outputStream
	cmd.Stderr = errStream
	start := time.Now()
	err := cmd.Run()
	execTime = time.Since(start).Seconds()
	if err != nil {
		success = false
	} else {
		success = true
	}
	if execTime < waitUntilFileExistTime {
		time.Sleep(waitUntilFileExistTime * time.Second)
	}

	return success, execTime
}
