package data

import (
	"sort"
	"strings"
	"time"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/rs/zerolog/log"
)

// Stores information about an AWS services and functions
type ServiceData struct {
	Service ecsTypes.Service
	Tasks   []string
}

type ECSClusterData struct {
	Services []ServiceData
}

type AccountData struct {
	Functions []lambdaTypes.FunctionConfiguration
	Cluster   ecsTypes.Cluster
	*ECSClusterData
	Refreshed time.Time
}

const maxImageWidth = 50

func NewAccountData(cluster ecsTypes.Cluster) *AccountData {
	ecsClusterData := loadECSClusterData(*cluster.ClusterName)
	lambdaFunctions := loadLambdaFunctions()

	return &AccountData{
		Functions:      lambdaFunctions,
		ECSClusterData: ecsClusterData,
		Cluster:        cluster,
		Refreshed:      time.Now(),
	}
}

func (d *AccountData) Refresh() {
	ecsClusterData := loadECSClusterData(*d.Cluster.ClusterName)
	lambdaFunctions := loadLambdaFunctions()

	d.Functions = lambdaFunctions
	d.ECSClusterData = ecsClusterData
	d.Refreshed = time.Now()
}

// Got an AWS error? Print an error message and suggest correctly configuring their AWS profile
func fatalAwsError(err error) {
	if err != nil {
		log.Fatal().Err(err).Msg("An issue occurred calling the AWS SDK. Please confirm you have the right AWS credentials in place.")
	}
}

func loadLambdaFunctions() []lambdaTypes.FunctionConfiguration {
	functionsResult, err := aws.ListLambdaFunctions()
	fatalAwsError(err)

	sort.SliceStable(functionsResult, func(i int, j int) bool {
		return 0 > strings.Compare(*functionsResult[i].FunctionName, *functionsResult[j].FunctionName)
	})

	return functionsResult
}

func loadECSClusterData(clusterName string) *ECSClusterData {
	// Read all services
	services, err := aws.DescribeClusterServices(clusterName)
	fatalAwsError(err)

	sort.SliceStable(services, func(i, j int) bool {
		return 0 > strings.Compare(*services[i].ServiceName, *services[j].ServiceName)
	})

	// Read all task definitions
	taskDefinitionArns := make([]string, 0, len(services))
	for _, service := range services {
		taskDefinitionArns = append(taskDefinitionArns, *service.TaskDefinition)
	}

	taskDefinitions, err := aws.GetTaskDefinitions(taskDefinitionArns)
	fatalAwsError(err)

	taskDefinitionArnLookup := make(map[string]ecsTypes.TaskDefinition)
	for _, taskDef := range taskDefinitions {
		taskDefinitionArnLookup[*taskDef.TaskDefinitionArn] = taskDef
	}

	serviceDataList := make([]ServiceData, 0, len(services))

	for _, v := range services {
		taskDefintion := taskDefinitionArnLookup[*v.TaskDefinition]
		tasks := lo.Map(taskDefintion.ContainerDefinitions, func(value ecsTypes.ContainerDefinition, index int) string {
			return utils.TakeLeft(utils.RemoveAllRegex(`.*/`, *value.Image), maxImageWidth)
		})

		data := ServiceData{
			Service: v,
			Tasks:   tasks,
		}

		serviceDataList = append(serviceDataList, data)
	}

	data := &ECSClusterData{
		Services: serviceDataList,
	}

	return data
}
