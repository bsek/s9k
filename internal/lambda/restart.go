package lambda

import (
	"fmt"
	"strings"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rs/zerolog/log"
)

func restart(functionName string) {
	const RESTART_DIALOG = "restart_dialog"

	text := fmt.Sprintf("Are you sure you want to restart the %s lambda function?", functionName)
	ui.CreateConfirmBox(text, doRestart(functionName), func() {})
}

func addDotToDescription(currentDescription *string) string {
	runes := []rune(*currentDescription)
	var newDescription string
	if strings.HasSuffix(*currentDescription, ".") {
		newDescription = string(runes[0 : len(runes)-1])
	} else {
		newDescription = fmt.Sprintf("%s.", *currentDescription)
	}
	return newDescription
}

func doRestart(functionName string) func() {
	return func() {

		currentDescription, err := aws.GetFunctionDescription(functionName)

		if err != nil {
			log.Error().Err(err).Msgf("Failed to read %s function description", functionName)
			ui.CreateMessageBox("Failed to restart function, see log for more information.")

			return
		}

		newDescription := addDotToDescription(currentDescription)

		err = aws.UpdateLambdaFunctionDescription(functionName, newDescription)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update %s function description", functionName)
			ui.CreateMessageBox("Failed to restart function, see log for more information.")
		}

		ui.CreateMessageBox(fmt.Sprintf("Lambda function %s restarted successfully", functionName))
	}
}
