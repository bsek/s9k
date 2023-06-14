package data

import (
	"time"

	ecsTypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	lambdaTypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
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
	Functions   []lambdaTypes.FunctionConfiguration
	ClusterData *ECSClusterData
	ClusterName string
	Refreshed   time.Time
	AccountId   string
}

type Container struct {
	Name            string
	Image           string
	LogStreamPrefix string
}
