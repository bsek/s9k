package lambda

import (
	"fmt"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/logs"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func showLogs(functionName string) {
	logGroupName := fmt.Sprintf("/aws/lambda/%s", functionName)

	log.Debug().Msgf("Looking for logs for log group name: %s", logGroupName)

	logStreams, err := aws.FetchLogStreams(logGroupName, nil, nil)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to load logStreams for function: %s", functionName)
		ui.CreateMessageBox("Failed to read log records, see log for more information.")
		return
	}

	logPage := logs.NewLogPage(logGroupName, logStreams)

	ui.App.RegisterContent(logPage)
	ui.App.ShowPage(logPage)
}
