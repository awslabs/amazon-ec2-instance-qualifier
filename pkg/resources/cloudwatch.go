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

package resources

import (
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
)

const (
	retryAttempts = 3
	retryPeriod   = 1 //minutes
)

// GetCloudWatchData retrieves instance metric data from CloudWatch
func (itf Resources) GetCloudWatchData(instances []Instance, testFixture config.TestFixture) (resp *cloudwatch.GetMetricDataOutput, err error) {
	startTime, _ := time.Parse(time.RFC3339, testFixture.StartTime)
	duration := time.Duration(testFixture.Timeout) * time.Second
	endTime := startTime.Add(duration)

	var queries []*cloudwatch.MetricDataQuery
	for i, instance := range instances {
		metricId := "cpu_metric" + strconv.Itoa(i)
		queries = append(queries, createCpuActiveQuery(instance, metricId, testFixture.Timeout))
	}
	for i, instance := range instances {
		metricId := "mem_metric" + strconv.Itoa(i)
		queries = append(queries, createMemUsageQuery(instance, metricId, testFixture.Timeout))
	}

	input := &cloudwatch.GetMetricDataInput{
		EndTime:           &endTime,
		StartTime:         &startTime,
		MetricDataQueries: queries,
	}

	log.Printf("Requesting metrics with GetMetricDataInput: %v\n", input)
	// CloudWatch data is not always immediately available for retrieval; therefore, best-effort retry
	for i := 0; i < retryAttempts; i++ {
		log.Printf("GetMetricData attempt: %s\n", strconv.Itoa(i))
		resp, err = itf.CloudWatch.GetMetricData(input)
		if err != nil {
			log.Println("error getting metric data ", err)
			log.Printf("Resp: %v\n", resp)
			return nil, err
		}
		for _, result := range resp.MetricDataResults {
			if result.Values == nil {
				time.Sleep(retryPeriod * time.Minute)
				continue
			}
		}
	}
	return resp, nil
}

func createCpuActiveQuery(instance Instance, metricId string, metricPeriod int) *cloudwatch.MetricDataQuery {
	namespace := "CWAgent"
	metricname := "cpu_usage_active"
	metricid := metricId
	returndata := true
	metricDim1Name := "InstanceId"
	metricDim1Value := instance.InstanceId
	metricDim2Name := "InstanceType"
	metricDim2Value := instance.InstanceType
	metricDim3Name := "cpu"
	metricDim3Value := "cpu-total"
	label := namespace + " " + instance.InstanceId + " " + instance.InstanceType + " " + metricname
	period := int64(convertToMultiple(metricPeriod, 60)) // MetricStat.Period must be a multiple of 60
	stat := "Maximum"                                    //p100
	return &cloudwatch.MetricDataQuery{
		Id:         &metricid,
		Label:      &label,
		ReturnData: &returndata,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &namespace,
				MetricName: &metricname,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  &metricDim1Name,
						Value: &metricDim1Value,
					},
					{
						Name:  &metricDim2Name,
						Value: &metricDim2Value,
					},
					{
						Name:  &metricDim3Name,
						Value: &metricDim3Value,
					},
				},
			},
			Period: &period,
			Stat:   &stat,
		},
	}
}

func createMemUsageQuery(instance Instance, metricId string, metricPeriod int) *cloudwatch.MetricDataQuery {
	namespace := "CWAgent"
	metricname := "mem_used_percent"
	metricid := metricId
	returnData := true
	metricDim1Name := "InstanceId"
	metricDim1Value := instance.InstanceId
	metricDim2Name := "InstanceType"
	metricDim2Value := instance.InstanceType
	label := namespace + " " + instance.InstanceId + " " + instance.InstanceType + " " + metricname
	period := int64(convertToMultiple(metricPeriod, 60))
	stat := "Maximum"
	return &cloudwatch.MetricDataQuery{
		Id:         &metricid,
		Label:      &label,
		ReturnData: &returnData,
		MetricStat: &cloudwatch.MetricStat{
			Metric: &cloudwatch.Metric{
				Namespace:  &namespace,
				MetricName: &metricname,
				Dimensions: []*cloudwatch.Dimension{
					{
						Name:  &metricDim1Name,
						Value: &metricDim1Value,
					},
					{
						Name:  &metricDim2Name,
						Value: &metricDim2Value,
					},
				},
			},
			Period: &period,
			Stat:   &stat,
		},
	}
}

// convertToMultiple takes a value and returns the nearest multiple, rounding up
func convertToMultiple(value, multiple int) int {
	if value <= multiple {
		return multiple
	}
	if value%multiple == 0 {
		return value
	}
	resultPlusBuff := value/multiple + 1
	return resultPlusBuff * multiple
}
