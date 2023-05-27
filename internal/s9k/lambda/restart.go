package lambda

import (
	"fmt"
	"strings"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const RESTART_DIALOG = "restart_dialog"

func restart(functionName string, pages *tview.Pages) {
	text := fmt.Sprintf("Are you sure you want to restart the %s lambda function?", functionName)
	confirmBox := ui.CreateConfirmBox(text, doRestart(functionName, pages), func() {
		pages.RemovePage(RESTART_DIALOG)
	})

	pages.AddAndSwitchToPage(RESTART_DIALOG, confirmBox, true)
}

func doRestart(functionName string, pages *tview.Pages) func() {
	return func() {
		currentDescription, err := aws.GetFunctionDescription(functionName)

		if err != nil {
			log.Error().Err(err).Msgf("Failed to read %s function description", functionName)

			name, _ := pages.GetFrontPage()
			pages.RemovePage(name)
			ui.CreateMessageBox("Failed to restart function, see log for more information.", pages)
			return
		}

		runes := []rune(*currentDescription)
		var newDescription string
		if strings.HasSuffix(*currentDescription, ".") {
			newDescription = string(runes[0 : len(runes)-1])
		} else {
			newDescription = fmt.Sprintf("%s.", *currentDescription)
		}

		err = aws.UpdateLambdaFunctionDescription(functionName, newDescription)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to update %s function description", functionName)

			name, _ := pages.GetFrontPage()
			pages.RemovePage(name)
			ui.CreateMessageBox("Failed to restart function, see log for more information.", pages)
		}

		name, _ := pages.GetFrontPage()
		pages.RemovePage(name)
		ui.CreateMessageBox(fmt.Sprintf("Lambda function %s restarted successfully", functionName), pages)
	}
}
