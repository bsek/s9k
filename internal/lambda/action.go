package lambda

import (
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/rivo/tview"

	"github.com/bsek/s9k/internal/ui"
)

func createActionForm(functionName string, arch lambdatypes.Architecture) {
	const ACTION_FORM = "action_form"

	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Show logs", "Restart service", "Deploy version", "Close"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Show logs" {
				showLogs(functionName)
			}
			if buttonLabel == "Restart service" {
				restart(functionName)
			}
			if buttonLabel == "Deploy version" {
				deploy(functionName, arch)
			}
			if buttonLabel == "Close" {
				ui.App.Content.RemovePage(ACTION_FORM)
			}
		})

	ui.App.Content.AddAndSwitchToPage(ACTION_FORM, modal, true)
}
