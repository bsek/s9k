package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awsecs"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awslambda"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awss3"

	"github.com/bsek/s9k/internal/s9k/utils"
)

// Struct describing a version
type Version struct {
	Name      string
	Message   string
	CreatedAt string
}

// Logitem, describing who installed which version
type LogItem struct {
	User    string    `json:"user,omitempty"`
	Version string    `json:"version,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

var ecsClient *ecs.Client
var s3Client *s3.Client
var lambdaClient *lambda.Client

const S3_BUCKET_NAME = "skjema-qa-deployment-bucket"

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())

	ecsClient = ecs.NewFromConfig(cfg)
	lambdaClient = lambda.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
}

func RestartECSService(clusterName, name string) error {
	return awsecs.StopEcsService(context.Background(), ecsClient, name, clusterName)
}

func FetchAvailableVersions(functionName string) ([]Version, error) {
	files, err := awss3.ListBucketObjects(context.Background(), s3Client, S3_BUCKET_NAME, functionName)
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

// Returns a list of lambda functions
func ListLambdaFunctions() ([]lambdatypes.FunctionConfiguration, error) {
	output, err := awslambda.ListFunctions(context.Background(), lambdaClient)
	if err != nil {
		return nil, err
	}

	return output.Functions, err
}

// Return lambda function description
func GetFunctionDescription(functionName string) (*string, error) {
	return awslambda.GetFunctionDescription(context.Background(), lambdaClient, functionName)
}

// Update lambda function description
func UpdateLambdaFunctionDescription(functionName, newDescription string) error {
	return awslambda.UpdateFunctionDescription(context.Background(), lambdaClient, functionName, newDescription)
}

// Deploy lambda function
func DeployLambdaFunction(functionName, version string, arch lambdatypes.Architecture) error {
	_, err := awslambda.UpdateFunctionCode(context.Background(), lambdaClient, functionName, S3_BUCKET_NAME, version, arch)
	return err

}

// Return a slice of all clusters in the account
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

// Return a slice of the services in the given ECS cluster
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

// Return a slice of the tasks in the given ECS cluster
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

// Executes a command in an ECS container
func ExecuteCommand(taskArn, containerName, clusterName string) (*ecs.ExecuteCommandOutput, error) {
	input := ecs.ExecuteCommandInput{
		Command:     aws.String("/bin/bash"),
		Interactive: true,
		Task:        &taskArn,
		Cluster:     &clusterName,
		Container:   &containerName,
	}

	output, err := ecsClient.ExecuteCommand(context.TODO(), &input)
	if err != nil {
		return nil, err
	}

	return output, nil
}

// Return a slice of the task definitions in the given ECS tasks
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

// Return a short version of the task definition arn
func ShortenTaskDefArn(taskDefinitionArn *string) string {
	return utils.RemoveAllBeforeLastChar("/", *taskDefinitionArn)
}
