package lambda

import (
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/rivo/tview"
)

const ACTION_FORM = "action_form"

func createActionForm(functionName string, arch lambdatypes.Architecture, pages *tview.Pages, closeFunc func()) {
	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Show logs", "Restart service", "Deploy version", "Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Show logs" {
				showLogs(functionName, pages)
			}
			if buttonLabel == "Restart service" {
				restart(functionName, pages)
			}
			if buttonLabel == "Deploy version" {
				deploy(functionName, arch, pages)
			}
			if buttonLabel == "Close" {
				closeFunc()
			}
		})

	pages.AddAndSwitchToPage(ACTION_FORM, modal, true)
}
