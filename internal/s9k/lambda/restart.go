package lambda

import (
	"fmt"
	"strings"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/rivo/tview"
)

const restartDialog = "RestartDialog"

func restart(functionName string, pages *tview.Pages, closeFunction func()) {
	modal := tview.NewModal().
		SetText("Are you sure you want to restart the lambda function?").
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				currentDescription, err := aws.GetFunctionDescription(functionName)

				if err != nil {
					runes := []rune(*currentDescription)
					var newDescription string
					if strings.HasSuffix(*currentDescription, ".") {
						newDescription = string(runes[0 : len(runes)-1])
					} else {
						newDescription = fmt.Sprintf("%s.", *currentDescription)
					}

					aws.UpdateLambdaFunctionDescription(functionName, newDescription)
				}
				pages.RemovePage(restartDialog)
				closeFunction()
			}
			if buttonLabel == "No" {
				pages.RemovePage(restartDialog)
				closeFunction()
			}
		})

	pages.AddPage(restartDialog, modal, true, true)
}
