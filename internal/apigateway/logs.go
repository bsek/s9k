package apigateway

import (
	"github.com/bsek/s9k/internal/logs"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func showLogs(logGroupArn string) {

	log.Debug().Msgf("Looking for logs for log group name: %s", logGroupArn)

	// logStreams, err := aws.FetchLogStreams(logGroupName, nil, nil)
	// if err != nil {
	// 	log.Error().Err(err).Msgf("Failed to load logStreams for function: %s", apiName)
	// 	ui.CreateMessageBox("Failed to read log records, see log for more information.")
	// 	return
	// }

	logPage := logs.NewLogPage(logGroupArn)

	ui.App.RegisterContent(logPage)
	ui.App.ShowPage(logPage)
}
