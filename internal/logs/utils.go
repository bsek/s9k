package logs

import (
	"fmt"

	"github.com/bsek/s9k/internal/aws"
)

func ConstructLogGroupArn(logGroupName string) (*string, error) {
	accountID, region, err := aws.GetAccountInformation()
	if err != nil {
		return nil, err
	}

	arn := fmt.Sprintf("arn:aws:logs:%s:%s:log-group:%s", *region, *accountID, logGroupName)

	return &arn, nil
}
