package ecs

import (
	"github.com/bsek/s9k/internal/ui"
	"github.com/rivo/tview"
)

func restart(serviceName string) {
	inputHandler := ui.App.TviewApp.GetInputCapture()

	modal := tview.NewModal().
		SetText("Do you really want to restart the service?").
		AddButtons([]string{"Restart", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Restart" {

			} else if buttonLabel == "Cancel" {
				ui.App.Content.RemovePage("modal")
				ui.App.TviewApp.SetInputCapture(inputHandler)
			}
		})

	ui.App.Content.AddAndSwitchToPage("modal", modal, false)
}
