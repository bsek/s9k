package ecs

import (
	"fmt"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/logs"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

func showLogs(taskArn, serviceName string, container data.Container, pages *tview.Pages) {
	logGroupName := fmt.Sprintf("/ecs/%s", serviceName)

	log.Info().Msgf("Using container: %v", container)
	log.Info().Msgf("Looking for logs for log group name: %s and prefix: %s", logGroupName, container.LogStreamPrefix)

	closeFunc := func() {
		pages.RemovePage("logs")
	}

	logStreams, err := aws.FetchLogStreams(logGroupName, container.LogStreamPrefix, taskArn)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load log streams for service %s and container %s", serviceName, container.LogStreamPrefix)
		return
	}

	logPage := logs.NewLogPage(logGroupName, logStreams, closeFunc)

	pages.AddAndSwitchToPage("logs", logPage.Flex, true)
}
