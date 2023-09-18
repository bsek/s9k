package ecs

import (
	"github.com/rivo/tview"

	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/shell"
	"github.com/bsek/s9k/internal/ui"
)

func action(taskArn, serviceName, clusterArn string, container data.Container) {
	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Show logs", "Open shell", "Close"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Show logs" {
				showLogs(taskArn, serviceName, container)
				ui.App.Content.RemovePage("modal")
			}
			if buttonLabel == "Open shell" {
				shell.NewShellPage(taskArn, container.Name, clusterArn)
				ui.App.Content.RemovePage("modal")
			}
			if buttonLabel == "Close" {
				ui.App.Content.RemovePage("modal")
			}
		})
	ui.App.Content.AddAndSwitchToPage("modal", modal, false)
}
