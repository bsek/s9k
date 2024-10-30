package lambda

import (
	"github.com/bsek/s9k/internal/logs"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func showLogs(logGroupName string) {
	// log.Debug().Msgf("Looking for logs for log group name: %s", logGroupName)
	// logStreams, err := aws.FetchLogStreams(logGroupName, nil, nil)
	// if err != nil {
	// 	log.Error().Err(err).Msgf("Failed to load logStreams for function: %s", functionName)
	// 	ui.CreateMessageBox("Failed to read log records, see log for more information.")
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
