package ecs

import (
	"github.com/rivo/tview"

	"github.com/bsek/s9k/internal/ui"
)

func restart(_ string) {
	inputHandler := ui.App.TviewApp.GetInputCapture()

	modal := tview.NewModal().
		SetText("Do you really want to restart the service?").
		AddButtons([]string{"Restart", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			switch buttonLabel {
			case "Restart":
			case "Cancel":
				ui.App.Content.RemovePage("modal")
				ui.App.TviewApp.SetInputCapture(inputHandler)
			}
		})

	ui.App.Content.AddAndSwitchToPage("modal", modal, false)
}
