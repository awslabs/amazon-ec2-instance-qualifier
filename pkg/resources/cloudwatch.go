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
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/awslabs/amazon-ec2-instance-qualifier/pkg/config"
)

// GetCloudWatchData retrieves instance metric data from CloudWatch
func (itf Resources) GetCloudWatchData(instances []Instance, testFixture config.TestFixture) (*cloudwatch.GetMetricDataOutput, error) {
	startTime, _ := time.Parse(time.RFC3339, testFixture.StartTime)
	duration := time.Duration(testFixture.Timeout) * time.Second
	endTime := startTime.Add(duration)

	var queries []*cloudwatch.MetricDataQuery
	for i, instance := range instances {
		metricId := "cpu_metric" + strconv.Itoa(i)
		queries = append(queries, createCPUUtilQuery(instance, metricId, testFixture.Timeout))
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
	fmt.Printf("Requesting metrics with GetMetricDataInput: %v\n", input)
	resp, err := itf.CloudWatch.GetMetricData(input)

	if err != nil {
		fmt.Println("error getting metric data ", err)
		fmt.Printf("Resp: %v\n", resp)
		return nil, err
	}

	fmt.Printf("metric data from CW: %v\n", resp)
	return resp, nil
}

func createCPUUtilQuery(instance Instance, metricId string, metricPeriod int) *cloudwatch.MetricDataQuery {
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
	period := int64(metricPeriod)
	stat := "Maximum" //p100
	return &cloudwatch.MetricDataQuery{
		Id:         &metricid,
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
	period := int64(metricPeriod)
	stat := "Maximum"
	return &cloudwatch.MetricDataQuery{
		Id:         &metricid,
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
