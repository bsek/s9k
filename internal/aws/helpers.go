package aws

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awsecs"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awslambda"
	"github.com/oslokommune/skjema-bibliotek-commons-go/aws/awss3"

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
	User    string    `json:"user,omitempty"`
	Version string    `json:"version,omitempty"`
	Updated time.Time `json:"updated,omitempty"`
}

const S3_BUCKET_VAR_NAME = "S3_DEPLOYMENT_BUCKET_NAME"

var (
	ecsClient      *ecs.Client
	s3Client       *s3.Client
	stsClient      *sts.Client
	lambdaClient   *lambda.Client
	s3_bucket_name string
)

func init() {
	cfg, _ := config.LoadDefaultConfig(context.TODO())

	value, found := os.LookupEnv(S3_BUCKET_VAR_NAME)
	if !found {
		fmt.Printf("The %s environment variable is not set.\nIt must be set to the parent directory where your lambda function zip files are stored.\nExiting...", S3_BUCKET_VAR_NAME)
		os.Exit(0)
	}

	s3_bucket_name = value

	ecsClient = ecs.NewFromConfig(cfg)
	lambdaClient = lambda.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
	stsClient = sts.NewFromConfig(cfg)
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
func DeployLambdaFunction(functionName, zipfile string, arch lambdatypes.Architecture) error {
	_, err := awslambda.UpdateFunctionCode(context.Background(), lambdaClient, functionName, s3_bucket_name, zipfile, arch)
	return err

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
