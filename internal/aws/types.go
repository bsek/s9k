package aws

import (
	"fmt"
	"time"
)

type ApiType int

const (
	Http ApiType = iota + 1
	Rest
)

func (a ApiType) String() string {
	apiTypes := [...]string{"Http", "Rest"}
	if a < Http || a > Rest {
		return fmt.Sprintf("ApiType(%d)", int(a))
	}
	return apiTypes[a-1]
}

type ApiGateway struct {
	CreatedDate time.Time
	Name        string
	Description string
	DomainName  string
	ApiId       string
	Type        ApiType
	LogGropuArn string
}

type Package struct {
	Sha            string
	Image          string
	Created        time.Time
	RegistryID     string
	RepositoryName string
}
