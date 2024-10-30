package ecs

import (
	"github.com/bsek/s9k/internal/logs"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func showLogs(logGroupName string) {
	// log.Info().Msgf("Looking for logs for log group name: %s and prefix: %s", logGroupName, container.LogStreamPrefix)

	// logStreams, err := aws.FetchLogStreams(logGroupName, nil, nil)
	// if err != nil {
	// 	log.Error().Err(err).Msgf("Failed to load log streams for service %s and container %s", serviceName, container.LogStreamPrefix)
	// 	return
	// }

	logGroupArn, err := logs.ConstructLogGroupArn(logGroupName)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to construct log group arn for log group: %s", logGroupName)
	}
	logPage := logs.NewLogPage(*logGroupArn)

	ui.App.RegisterContent(logPage)
	ui.App.ShowPage(logPage)
}
