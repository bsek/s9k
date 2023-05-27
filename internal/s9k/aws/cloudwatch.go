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
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

var cloudwatchClient *cloudwatchlogs.Client
var cloudwatchMetricsClient *cloudwatch.Client

func init() {
	// Use the SDK's default configuration with region and custome endpoint resolver
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-north-1"))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

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

func FetchCloudwatchLogs(logGroupName, logStreamName string) ([][]types.OutputLogEvent, error) {
	ctx := context.Background()

	threeDaysAgo := time.Now().Add(time.Duration(-3) * time.Hour * 24)

	logEventsInput := cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(logGroupName),
		LogStreamName: aws.String(logStreamName),
		StartFromHead: aws.Bool(false),
		StartTime:     aws.Int64(threeDaysAgo.Unix() * 1000),
	}

	outputList := [][]types.OutputLogEvent{}
	var nextToken *string

	for {
		if nextToken != nil {
			logEventsInput.NextToken = nextToken
		}

		logEventsOutput, err := cloudwatchClient.GetLogEvents(ctx, &logEventsInput)

		if err != nil {
			log.Error().Err(err).Msgf("Failed to fetch log events for logGroupName %s and logStreamName %s", logGroupName, logStreamName)
			return nil, err
		}

		if nextToken != nil && *logEventsOutput.NextForwardToken == *nextToken {
			break
		}

		nextToken = logEventsOutput.NextForwardToken
		outputList = append(outputList, logEventsOutput.Events)
	}

	return outputList, nil
}

func FetchCpuAndMemoryUsage(name string) (uint32, uint32, uint32, uint32, error) {
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
								Value: aws.String("qa"),
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
								Value: aws.String("qa"),
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
								Value: aws.String("qa"),
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
								Value: aws.String("qa"),
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
		log.Error().Err(err).Msg("Error getting CloudWatch metrics:")
		return 0, 0, 0, 0, err
	}

	var cpuUsed uint32
	var cpuReserved uint32
	var memUsed uint32
	var memReserved uint32

	datapoints := uint32(0)

	// Process the returned data
	for _, metricDataResult := range result.MetricDataResults {
		for _, value := range metricDataResult.Values {
			switch *metricDataResult.Id {
			case "mem_used":
				memUsed += uint32(value)
			case "mem_reserved":
				memReserved += uint32(value)
			case "cpu_used":
				cpuUsed += uint32(value)
			case "cpu_reserved":
				cpuReserved += uint32(value)
				datapoints++
			}
		}
	}

	if datapoints == 0 {
		return 0, 0, 0, 0, nil
	}

	return (cpuUsed / datapoints), (cpuReserved / datapoints), (memUsed / datapoints), (memReserved / datapoints), nil
}
