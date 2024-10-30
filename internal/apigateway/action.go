package apigateway

import (
	"github.com/rivo/tview"

	"github.com/bsek/s9k/internal/ui"
)

func createActionForm(logGroupArn string) {
	const ACTION_FORM = "action_form"

	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Show access logs", "Close"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Show access logs" {
				showLogs(logGroupArn)
			}
			if buttonLabel == "Close" {
				ui.App.Content.RemovePage(ACTION_FORM)
			}
		})

	ui.App.Content.AddAndSwitchToPage(ACTION_FORM, modal, true)
}
