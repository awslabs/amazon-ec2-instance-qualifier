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
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/resources"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/setup"
)

const (
	// How often to record a data point in seconds
	monitorPeriod = "3"
	// If the test file finishes execution too fast, we should wait for some time before killing the monitoring script to ensure the output file has been created
	waitUntilFileExistTime = 1
	resultSuccess          = "pass"
	resultFail             = "fail"
)

// Enum corresponding to the indices of metricInfos
const (
	cpu = iota
	mem
)

var metricInfos = [...]metricInfo{
	{"cpu-load", "n", "monitor-cpu.sh", "cpu.load", 0},
	{"mem-used", "MiB", "monitor-mem.sh", "mem.load", 0},
}

// PopulateThresholds populates the threshold values of all metrics.
func PopulateThresholds(instance resources.Instance, rawTargetUtil string) error {
	targetUtil, err := strconv.Atoi(rawTargetUtil)
	if err != nil {
		return err
	}
	vCpus, err := strconv.Atoi(instance.VCpus)
	if err != nil {
		return err
	}
	memory, err := strconv.Atoi(instance.Memory)
	if err != nil {
		return err
	}
	metricInfos[cpu].threshold = float64(vCpus*targetUtil) / 100.0
	metricInfos[mem].threshold = float64(memory*targetUtil) / 100.0

	return nil
}

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

// PopulateResult runs the given test file, populates and returns the result.
// The file will be benchmarked against CPU load and Memory used. During the execution, monitoring scripts
// will periodically record metric values. Upon finishing, for each metric, the average value of all data
// points will be used as the final value.
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

	for _, metricInfo := range metricInfos {
		var metric resources.Metric
		metric.MetricUsed = metricInfo.name
		if value, err := calculateAvgValue(agentFixture.ScriptPath + "/" + metricInfo.output); err != nil {
			// Failing to populate one metric shouldn't stop populating the other metrics
			log.Println(err)
			continue
		} else {
			metric.Value = value
		}
		metric.Threshold = metricInfo.threshold
		metric.Unit = metricInfo.unit

		testResult.Metrics = append(testResult.Metrics, metric)
	}

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
	if file.IsDir() || setup.IsInstanceQualifierScript(filename) || strings.HasSuffix(filename, ".load") || strings.HasSuffix(filename, ".json") || strings.HasSuffix(filename, ".log") {
		return false
	}
	return true
}

// execute executes the test file, runs the monitoring scripts in the background, and returns the test result
// and execution time.
func execute(filename string, scriptPath string, outputStream *os.File, errStream *os.File) (success bool, execTime float64) {
	fmt.Fprintf(outputStream, "\n======================================================================================================\n")
	fmt.Fprintf(outputStream, "🥑 Starting %s\n", filename)
	fmt.Fprintf(outputStream, "======================================================================================================\n")

	// Run the monitoring scripts
	var monitorProcesses []*os.Process
	for _, metricInfo := range metricInfos {
		cmd := exec.Command(scriptPath+"/"+metricInfo.monitorScript, metricInfo.output, monitorPeriod)
		cmd.Stdout = outputStream
		cmd.Stderr = errStream
		if err := cmd.Start(); err != nil {
			// Failing to start one monitoring script shouldn't prevent the test file from running
			log.Println(err)
			continue
		}
		monitorProcesses = append(monitorProcesses, cmd.Process)
	}

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

	for _, process := range monitorProcesses {
		process.Kill()
	}

	return success, execTime
}

// calculateAvgValue reads all recorded data points from the load output file and calculates the average value.
func calculateAvgValue(filename string) (value float64, err error) {
	defer os.Remove(filename)

	sum := 0.0
	dataPointNum := 0
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		dataPoint, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return 0, err
		}
		sum += dataPoint
		dataPointNum++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}

	value = sum / float64(dataPointNum)
	return value, nil
}
