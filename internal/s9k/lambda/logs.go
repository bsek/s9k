package lambda

import (
	"fmt"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/logs"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

func showLogs(functionName string, pages *tview.Pages) {
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)

	log.Info().Msgf("Looking for logs for log group name: %s", logGroupName)

	closeFunc := func() {
		pages.RemovePage("logs")
	}

	logStreams, err := aws.FetchLambdaLogStreams(logGroupName)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load logStreams for function: %s", functionName)
		return
	}

	logPage := logs.NewLogPage(logGroupName, logStreams, closeFunc)

	pages.AddAndSwitchToPage("logs", logPage.Flex, true)
}
