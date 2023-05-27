package ecs

import (
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/shell"
	"github.com/rivo/tview"
)

func action(taskArn, serviceName string, container data.Container, pages *tview.Pages, app *tview.Application) {
	modal := tview.NewModal().
		SetText("What do you want to do?").
		AddButtons([]string{"Show logs", "Open shell", "Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Show logs" {
				showLogs(taskArn, serviceName, container, pages)
				pages.RemovePage("modal")
			}
			if buttonLabel == "Open shell" {
				shellPage, _ := shell.NewShellPage(taskArn, container.Name, "qa", app)
				pages.RemovePage("modal")
				pages.AddAndSwitchToPage("shell", shellPage, true)
			}
			if buttonLabel == "Close" {
				pages.RemovePage("modal")
			}
		})
	pages.AddAndSwitchToPage("modal", modal, false)
}
