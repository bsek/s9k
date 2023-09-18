package data

import (
	"sort"
	"strings"
	"time"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/utils"
)

const MAX_IMAGE_WIDTH = 50

func NewAccountData(clusterName string, account string) *AccountData {
	return &AccountData{
		Functions:   nil,
		ClusterName: clusterName,
		ClusterData: nil,
		Refreshed:   time.Now(),
		AccountId:   account,
		Apis:        nil,
	}
}

func (d *AccountData) Refresh() {
	ecsClusterData := loadECSClusterData(d.ClusterName)
	lambdaFunctions := loadLambdaFunctions()
	apis := loadApis()
	d.Functions = lambdaFunctions
	d.ClusterData = ecsClusterData
	d.Apis = apis
	d.Refreshed = time.Now()
}

func loadApis() []aws.ApiGateway {
	apis := aws.FetchApis()
	log.Debug().Msgf("Found gateways: %v", apis)
	return apis
}

func loadLambdaFunctions() []Function {
	functionsResult, err := aws.ListLambdaFunctions()
	if err != nil {
		log.Fatal().Err(err).Msg("An issue occurred when loading lambda functions")
	}

	sort.SliceStable(functionsResult, func(i int, j int) bool {
		return 0 > strings.Compare(*functionsResult[i].FunctionName, *functionsResult[j].FunctionName)
	})

	functions := lo.Map(functionsResult, func(item lambdaTypes.FunctionConfiguration, _ int) Function {
		output, err := aws.GetLambdaFunction(*item.FunctionArn)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to read lambda function details for function %s", *item.FunctionName)
		}
		return Function{
			output.Tags,
			item,
		}
	})

	return functions
}

func loadECSClusterData(clusterName string) *ECSClusterData {
	// Read all services
	services, err := aws.DescribeClusterServices(clusterName)
	if err != nil {
		log.Fatal().Err(err).Msg("An issue occurred when loading ECS cluster data")
	}

	sort.SliceStable(services, func(i, j int) bool {
		return 0 > strings.Compare(*services[i].ServiceName, *services[j].ServiceName)
	})

	// Read all task definitions
	taskDefinitionArns := make([]string, 0, len(services))
	for _, service := range services {
		taskDefinitionArns = append(taskDefinitionArns, *service.TaskDefinition)
	}

	taskDefinitions, err := aws.GetTaskDefinitions(taskDefinitionArns)
	if err != nil {
		log.Fatal().Err(err).Msg("An issue occurred when loading ECS task definitions")
	}

	taskDefinitionArnLookup := make(map[string]ecsTypes.TaskDefinition)
	for _, taskDef := range taskDefinitions {
		taskDefinitionArnLookup[*taskDef.TaskDefinitionArn] = taskDef
	}

	serviceDataList := make([]ServiceData, 0, len(services))

	for i := range services {
		service := services[i]
		taskDefintion := taskDefinitionArnLookup[*service.TaskDefinition]

		containers := lo.Map(taskDefintion.ContainerDefinitions, func(value ecsTypes.ContainerDefinition, _ int) Container {
			container := Container{
				Name:            *value.Name,
				Image:           utils.TakeLeft(utils.RemoveAllBeforeLastChar("/", *value.Image), MAX_IMAGE_WIDTH),
				LogStreamPrefix: value.LogConfiguration.Options["awslogs-stream-prefix"],
			}

			return container
		})

		data := ServiceData{
			Service:    &service,
			Containers: containers,
		}

		serviceDataList = append(serviceDataList, data)
	}

	data := &ECSClusterData{
		Services: serviceDataList,
	}

	return data
}
