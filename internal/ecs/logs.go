package ecs

import (
	"fmt"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/logs"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func showLogs(taskArn, serviceName string, container data.Container) {
	logGroupName := fmt.Sprintf("/ecs/%s", serviceName)

	log.Info().Msgf("Using container: %v", container)
	log.Info().Msgf("Looking for logs for log group name: %s and prefix: %s", logGroupName, container.LogStreamPrefix)

	logStreams, err := aws.FetchLogStreams(logGroupName, &container.LogStreamPrefix, &taskArn)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load log streams for service %s and container %s", serviceName, container.LogStreamPrefix)
		return
	}

	logPage := logs.NewLogPage(logGroupName, logStreams)

	ui.App.RegisterContent(logPage)
	ui.App.ShowPage(logPage)
}
