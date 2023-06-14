package aws

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	metrictypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/bsek/s9k/internal/utils"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var cloudwatchClient *cloudwatchlogs.Client
var cloudwatchMetricsClient *cloudwatch.Client

func init() {
	// Use the SDK's default configuration with region and custome endpoint resolver
	cfg, _ := config.LoadDefaultConfig(context.TODO())

	cloudwatchClient = cloudwatchlogs.NewFromConfig(cfg)
	cloudwatchMetricsClient = cloudwatch.NewFromConfig(cfg)
}

func listLogStreams(logGroupName string) ([]types.LogStream, error) {
	ctx := context.Background()

	input := cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroupName),
		OrderBy:      types.OrderByLastEventTime,
		Limit:        aws.Int32(10),
		Descending:   aws.Bool(true),
	}

	output, err := cloudwatchClient.DescribeLogStreams(ctx, &input)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to lookup logstreams for logGroupName %s", logGroupName)
		return nil, err
	}

	return output.LogStreams, nil
}

// FetchLambdaLogStreams reads available logstreams from the specified cloudwatch LogGroup and returns them as a slice
// The returned list is filter to only return streams which have been written to during the last three days
func FetchLambdaLogStreams(logGroupName string) ([]types.LogStream, error) {
	logStreams, err := listLogStreams(logGroupName)
	if err != nil {
		return nil, err
	}

	threeDaysAgo := time.Now().Add(time.Duration(-3) * time.Hour * 24).Unix()

	events := lo.Filter(logStreams, func(item types.LogStream, index int) bool {
		if *item.LastEventTimestamp > (threeDaysAgo * 1000) {
			return true
		}

		return false
	})

	log.Info().Msgf("Found %d logstreams", len(events))

	return events, nil
}

// FetchLogStreams reads available logstreams from the specified cloudwatch LogGroup and returns them as a slice.
// The list is filtered to only include streams for the specified container and ECS taskArn and only includes
// streams that have been written to during the last three days
func FetchLogStreams(logGroupName string, containerName, taskArn string) ([]types.LogStream, error) {
	logStreams, err := listLogStreams(logGroupName)
	if err != nil {
		return nil, err
	}

	threeDaysAgo := time.Now().Add(time.Duration(-3) * time.Hour * 24).Unix()

	log.Info().Msgf("Looking for logstreams with %s and %s in name", utils.RemoveAllBeforeLastChar("/", taskArn), containerName)

	events := lo.Filter(logStreams, func(item types.LogStream, index int) bool {
		if *item.LastEventTimestamp > (threeDaysAgo * 1000) {
			return strings.Contains(*item.LogStreamName, utils.RemoveAllBeforeLastChar("/", taskArn)) && strings.Contains(*item.LogStreamName, containerName)
		}

		return false
	})

	log.Info().Msgf("Found %d logstreams", len(events))

	return events, nil
}

// FetchCloudwatchLogs reads the content of a log stream and returns a two dimensional slice of log events
func FetchCloudwatchLogs(logGroupName, logStreamName string, nextForwardToken *string, interval time.Duration) (outputList [][]types.OutputLogEvent, nextToken *string, err error) {
	ctx := context.Background()

	startTime := time.Now().Add(-interval)

	log.Debug().Msgf("Looking for logs greater than: %s", utils.FormatLocalDateTime(startTime))

	logEventsInput := cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		StartFromHead: aws.Bool(false),
		StartTime:     aws.Int64(startTime.Unix() * 1000),
	}

	nextToken = nextForwardToken

	var logEventsOutput *cloudwatchlogs.GetLogEventsOutput

	for {
		if nextToken != nil {
			logEventsInput.NextToken = nextToken
		}

		logEventsOutput, err = cloudwatchClient.GetLogEvents(ctx, &logEventsInput)

		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch log events for logGroupName %s and logStreamName %s", logGroupName, logStreamName)
			return
		}

		if nextToken != nil && *logEventsOutput.NextForwardToken == *nextToken {
			break
		}

		nextToken = logEventsOutput.NextForwardToken
		outputList = append(outputList, logEventsOutput.Events)
	}

	return
}

// FetchCpuAndMemoryUsage queries cloudwatch for ECS service metrics and returns the following metrics:
// - MemoryUtilized
// - MemoryReserved
// - CpuUtilized
// - CpuReserved
func FetchCpuAndMemoryUsage(name, clusterName string) (memoryUtilized uint32, memoryReserved uint32, cpuUtilized uint32, cpuReserved uint32, err error) {
	startTime := time.Now().Add(-10 * time.Minute)
	endTime := time.Now()
	input := &cloudwatch.GetMetricDataInput{
		MetricDataQueries: []metrictypes.MetricDataQuery{
			{
				Id: aws.String("mem_used"),
				MetricStat: &metrictypes.MetricStat{
					Metric: &metrictypes.Metric{
						Namespace:  aws.String("ECS/ContainerInsights"),
						MetricName: aws.String("MemoryUtilized"),
						Dimensions: []metrictypes.Dimension{
							{
								Name:  aws.String("TaskDefinitionFamily"),
								Value: aws.String(name),
							},
							{
								Name:  aws.String("ClusterName"),
								Value: aws.String(clusterName),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("mem_reserved"),
				MetricStat: &metrictypes.MetricStat{
					Metric: &metrictypes.Metric{
						Namespace:  aws.String("ECS/ContainerInsights"),
						MetricName: aws.String("MemoryReserved"),
						Dimensions: []metrictypes.Dimension{
							{
								Name:  aws.String("TaskDefinitionFamily"),
								Value: aws.String(name),
							},
							{
								Name:  aws.String("ClusterName"),
								Value: aws.String(clusterName),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("cpu_used"),
				MetricStat: &metrictypes.MetricStat{
					Metric: &metrictypes.Metric{
						Namespace:  aws.String("ECS/ContainerInsights"),
						MetricName: aws.String("CpuUtilized"),
						Dimensions: []metrictypes.Dimension{
							{
								Name:  aws.String("TaskDefinitionFamily"),
								Value: aws.String(name),
							},
							{
								Name:  aws.String("ClusterName"),
								Value: aws.String(clusterName),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Average"),
				},
			},
			{
				Id: aws.String("cpu_reserved"),
				MetricStat: &metrictypes.MetricStat{
					Metric: &metrictypes.Metric{
						Namespace:  aws.String("ECS/ContainerInsights"),
						MetricName: aws.String("CpuReserved"),
						Dimensions: []metrictypes.Dimension{
							{
								Name:  aws.String("TaskDefinitionFamily"),
								Value: aws.String(name),
							},
							{
								Name:  aws.String("ClusterName"),
								Value: aws.String(clusterName),
							},
						},
					},
					Period: aws.Int32(60),
					Stat:   aws.String("Average"),
				},
			},
		},
		StartTime: &startTime,
		EndTime:   &endTime,
	}

	// Call CloudWatch's GetMetricData API to retrieve metric data
	result, err := cloudwatchMetricsClient.GetMetricData(context.Background(), input)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching CloudWatch metrics")
		return 0, 0, 0, 0, err
	}

	datapoints := uint32(0)

	// Process the returned data
	for _, metricDataResult := range result.MetricDataResults {
		for _, value := range metricDataResult.Values {
			switch *metricDataResult.Id {
			case "mem_used":
				memoryUtilized += uint32(value)
			case "mem_reserved":
				memoryReserved += uint32(value)
			case "cpu_used":
				cpuUtilized += uint32(value)
			case "cpu_reserved":
				cpuReserved += uint32(value)
				datapoints++
			}
		}
	}

	if datapoints == 0 {
		return 0, 0, 0, 0, nil
	}

	return (cpuUtilized / datapoints), (cpuReserved / datapoints), (memoryUtilized / datapoints), (memoryReserved / datapoints), nil
}
