package aws

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	awscloudwatchlogstypes "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/oslokommune/common-lib-go/aws/awsapigateway"
	"github.com/oslokommune/common-lib-go/aws/awsapigatewayv2"
	awscloudwatch "github.com/oslokommune/common-lib-go/aws/awscloudwatch"
	"github.com/oslokommune/common-lib-go/aws/awsecs"
	"github.com/oslokommune/common-lib-go/aws/awslambda"
	"github.com/oslokommune/common-lib-go/aws/awss3"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/utils"
)

// Struct describing a version
type Version struct {
	Name      string
	Message   string
	CreatedAt string
}

// Logitem, describing who installed which version
type LogItem struct {
	Updated time.Time `json:"updated,omitempty"`
	User    string    `json:"user,omitempty"`
	Version string    `json:"version,omitempty"`
}

const S3_BUCKET_VAR_NAME = "S3_DEPLOYMENT_BUCKET_NAME"

var (
	ecsClient            *ecs.Client
	s3Client             *s3.Client
	stsClient            *sts.Client
	lambdaClient         *lambda.Client
	apigatewayv2Client   *apigatewayv2.Client
	apigatewayClient     *apigateway.Client
	cloudwatchClient     *cloudwatch.Client
	cloudwatchLogsClient *cloudwatchlogs.Client
	s3_bucket_name       string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())

	value, found := os.LookupEnv(S3_BUCKET_VAR_NAME)
	if !found {
		fmt.Printf("The %s environment variable is not set.\nIt must be set to the parent directory where your lambda function zip files are stored.\nExiting...", S3_BUCKET_VAR_NAME)
		os.Exit(0)
	}

	s3_bucket_name = value

	// configures clients
	ecsClient = ecs.NewFromConfig(cfg)
	apigatewayv2Client = apigatewayv2.NewFromConfig(cfg)
	apigatewayClient = apigateway.NewFromConfig(cfg)
	lambdaClient = lambda.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	stsClient = sts.NewFromConfig(cfg)
	cloudwatchClient = cloudwatch.NewFromConfig(cfg)
	cloudwatchLogsClient = cloudwatchlogs.NewFromConfig(cfg)
}

// FetchApis reads apigateway apis
func FetchApis() []ApiGateway {
	values := make([]ApiGateway, 0, 20)

	domainNames, error := awsapigatewayv2.GetDomainNames(context.TODO(), apigatewayv2Client)
	if error != nil {
		log.Error().Err(error).Msg("Failed to read apigateway domain names")
		return nil
	}

	mapping := make(map[string]string, 0)
	for _, v := range domainNames {
		mappings, err := awsapigatewayv2.GetApiMappings(context.TODO(), apigatewayv2Client, *v.DomainName)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read api gateway domain name mappings for domain %s", *v.DomainName)
		} else {
			if len(mappings) > 0 {
				mapping[*mappings[0].ApiId] = *v.DomainName
			}
		}
	}

	httpApis, error := awsapigatewayv2.GetApiGatewayv2Apis(context.TODO(), apigatewayv2Client)
	if error != nil {
		log.Error().Err(error).Msg("Failed to read apigateway rest apis")
	} else {
		for _, api := range httpApis {
			description := ""
			if api.Description != nil {
				description = *api.Description
			}

			domainMapping := ""
			if val, ok := mapping[*api.ApiId]; ok {
				domainMapping = val
			}

			values = append(values, ApiGateway{
				Name:        *api.Name,
				Description: description,
				ApiId:       *api.ApiId,
				DomainName:  domainMapping,
				Type:        Http,
				CreatedDate: *api.CreatedDate,
			})
		}
	}

	restApis, error := awsapigateway.GetApiGatewayApis(context.TODO(), apigatewayClient)
	if error != nil {
		log.Error().Err(error).Msg("Failed to read apigateway http apis")
	} else {
		for _, api := range restApis {
			description := ""
			if api.Description != nil {
				description = *api.Description
			}

			domainMapping := ""
			if val, ok := mapping[*api.Id]; ok {
				domainMapping = val
			}

			values = append(values, ApiGateway{
				Name:        *api.Name,
				Description: description,
				ApiId:       *api.Id,
				DomainName:  domainMapping,
				Type:        Rest,
				CreatedDate: *api.CreatedDate,
			})
		}
	}

	return values
}

// FetchCpuAndMemoryUsage reads memory and cpu usage from cloudwatch metrics for a given ECS service. Service and cluster name must be provided.
func FetchCpuAndMemoryUsage(serviceName, clusterName string) (memoryUtilized uint32, memoryReserved uint32, cpuUtilized uint32, cpuReserved uint32, err error) {
	memoryUtilized, memoryReserved, cpuUtilized, cpuReserved, err = awscloudwatch.FetchCpuAndMemoryUsage(context.TODO(), serviceName, clusterName, cloudwatchClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to retrieve metrics for ecs service")
	}
	return
}

// FetchLogStreams fetches log streams for a given log group. If container and taskArn is provided, it is used to filter the returned result.
func FetchLogStreams(logGroupName string, container, taskArn *string) ([]awscloudwatchlogstypes.LogStream, error) {
	task := new(string)

	if taskArn != nil {
		shortTask := utils.RemoveAllBeforeLastChar("/", *taskArn)
		task = &shortTask
	}

	output, err := awscloudwatch.FetchLogStreams(context.Background(), logGroupName, container, task, cloudwatchLogsClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read logstreams from cloudwatch")
		return nil, err
	}

	return output, err
}

// FetchCloudwatchLogs fetches log records from cloudwatch. If a nextForwardToken is provided, it will be used to in the query to cloudwatch, if not, it starts from scratch
func FetchCloudwatchLogs(logGroupName, logStreamName string, nextForwardToken *string, interval time.Duration) (outputList [][]awscloudwatchlogstypes.OutputLogEvent, nextToken *string, err error) {
	outputList, nextToken, err = awscloudwatch.FetchCloudwatchLogs(context.TODO(), logGroupName, logStreamName, nextForwardToken, interval, cloudwatchLogsClient)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read log records from cloudwatch")
		return
	}

	return
}

// GetAccountInformation returns information about the logged in account
func GetAccountInformation() (*string, error) {
	input := &sts.GetCallerIdentityInput{}
	output, err := stsClient.GetCallerIdentity(context.TODO(), input)
	if err != nil {
		return nil, err
	}
	return output.Account, nil
}

// RestartECSService restarts an ECS service by using updateing the service and setting the
// ForceNewDeployment flag, but not changing anything else. This effectively forces the ECS service
// to restart.
func RestartECSService(clusterName, name string) error {
	return awsecs.StopEcsService(context.Background(), ecsClient, name, clusterName)
}

// FetchAvailableVersions reads all files in the s3 bucket and returns the name and created timestamp.
// The Message field is not filled
func FetchAvailableVersions(functionName string) ([]Version, error) {
	files, err := awss3.ListBucketObjects(context.Background(), s3Client, s3_bucket_name, functionName)
	if err != nil {
		return nil, err
	}

	versions := make([]Version, 0, len(files))
	for _, v := range files {
		versions = append(versions, Version{
			Name:      v.Name,
			Message:   "",
			CreatedAt: v.CreatedAt.Format(time.DateTime),
		})
	}
	return versions, nil
}

func GetLambdaFunction(functionName string) (*lambda.GetFunctionOutput, error) {
	return awslambda.GetFunction(context.Background(), lambdaClient, functionName)
}

// ListLambdaFunctions lists all lambda functions the account and returns them as a slice of FunctionConfigurations
func ListLambdaFunctions() ([]lambdatypes.FunctionConfiguration, error) {
	output, err := awslambda.ListFunctions(context.Background(), lambdaClient)
	if err != nil {
		return nil, err
	}

	return output.Functions, err
}

// GetFunctionDescription reads the lambda function description for the specified function and returns it
func GetFunctionDescription(functionName string) (*string, error) {
	return awslambda.GetFunctionDescription(context.Background(), lambdaClient, functionName)
}

// UpdateLambdaFunctionDescription updates the lambda function description to the specified string. This is
// commonly used to force a cold start and makes the lambda function "restart"
func UpdateLambdaFunctionDescription(functionName, newDescription string) error {
	return awslambda.UpdateFunctionDescription(context.Background(), lambdaClient, functionName, newDescription)
}

// DeployLambdaFunction lets you update the function code for a specified lambda function by
// setting the code to execute point to a zip file in a lambda bucket. Arch must match the architecture the
// code is compiled for
func DeployLambdaFunction(functionName, zipfile string, arch lambdatypes.Architecture) (*string, error) {
	output, err := awslambda.UpdateFunctionCode(context.Background(), lambdaClient, functionName, s3_bucket_name, zipfile, arch)
	if err != nil {
		return nil, err
	}
	return output.FunctionArn, nil
}

// TagLambdaFunctionWithVersion adds LastDeployed tag to a lambda function
func TagLambdaFunctionWithVersion(functionName, version string) error {
	return awslambda.TagLambdaFunction(context.Background(), lambdaClient, functionName, map[string]string{"LastDeployed": version})
}

// ListECSClusters returns a slice of all ECS clusters found in the account
func ListECSClusters() ([]types.Cluster, error) {
	list, err := awsecs.ListClusters(context.Background(), ecsClient)
	if err != nil {
		return nil, err
	}

	// If STATISTICS is specified, the task and service count is included, separated by launch type.
	output, err := awsecs.DescribeClusters(context.Background(), ecsClient, list.ClusterArns, []types.ClusterField{types.ClusterFieldStatistics})
	if err != nil {
		return nil, err
	}

	return output.Clusters, err
}

// DescribeClusterServices return a slice of the service descriptions found in the given ECS cluster
func DescribeClusterServices(clusterName string) ([]types.Service, error) {
	output, err := awsecs.ListServices(context.Background(), ecsClient, clusterName)
	if err != nil {
		return nil, err
	}

	serviceLen := len(output.ServiceArns)

	services := make([]types.Service, 0)

	for i := 0; i < serviceLen; i += 10 {
		max := i + 10
		if max > serviceLen {
			max = serviceLen
		}

		list := output.ServiceArns[i:max]

		servicesOutput, err := awsecs.DescribeServices(context.Background(), ecsClient, list, clusterName)
		if err != nil {
			return nil, err
		}

		services = append(services, servicesOutput.Services...)
	}

	return services, err
}

// DescribeClusterTasks lists all tasks in a cluster or service (if serivceName != nil) and
// returns a slice of the tasks found
func DescribeClusterTasks(clusterName, serviceName *string) ([]types.Task, error) {
	listTasksOutput, err := awsecs.ListTasks(context.Background(), ecsClient, *clusterName, *serviceName)
	if err != nil {
		return nil, err
	}

	describeTasksOutput, err := awsecs.DescribeTasks(context.Background(), ecsClient, listTasksOutput.TaskArns, *clusterName)
	if err != nil {
		return nil, err
	}

	return describeTasksOutput.Tasks, nil
}

// ExecuteCommand executes /bin/sh command in an ECS container and returns the output
func ExecuteCommand(taskArn, containerName, clusterArn string) (*ecs.ExecuteCommandOutput, error) {
	input := ecs.ExecuteCommandInput{
		Command:     aws.String("/bin/sh"),
		Interactive: true,
		Task:        &taskArn,
		Cluster:     &clusterArn,
		Container:   &containerName,
	}

	output, err := ecsClient.ExecuteCommand(context.Background(), &input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// GetTaskDefinitions returns a slice of the task definitions in the given ECS tasks identified by Task definition arns
func GetTaskDefinitions(taskDefinitionArns []string) ([]types.TaskDefinition, error) {
	var taskDefinitions []types.TaskDefinition
	for _, v := range taskDefinitionArns {
		output, err := awsecs.DescribeTaskDefinition(context.Background(), ecsClient, v)
		if err != nil {
			return nil, err
		}
		taskDefinitions = append(taskDefinitions, *output.TaskDefinition)
	}

	return taskDefinitions, nil
}

// ShortTaskDefArn returns a short version of the task definition arn
func ShortenTaskDefArn(taskDefinitionArn *string) string {
	return utils.RemoveAllBeforeLastChar("/", *taskDefinitionArn)
}
