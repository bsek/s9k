package data

import (
	"time"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"

	"github.com/bsek/s9k/internal/aws"
)

// Stores information about AWS services and functions
type ServiceData struct {
	Service    *ecsTypes.Service
	Containers []Container
}

type ECSClusterData struct {
	Services []ServiceData
}

type AccountData struct {
	Functions   []Function
	ClusterData *ECSClusterData
	ClusterName string
	Refreshed   time.Time
	AccountId   string
	Apis        []aws.ApiGateway
}

type Container struct {
	Name            string
	Image           string
	LogStreamPrefix string
	LogGroupName    string
}

type Function struct {
	Tags map[string]string
	lambdaTypes.FunctionConfiguration
}
