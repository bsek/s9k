package ecs

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
)

func restart(serviceName, clusterName string) {
	inputHandler := ui.App.TviewApp.GetInputCapture()

	modal := tview.NewModal().
		SetText("Do you really want to restart the service?").
		AddButtons([]string{"Restart", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			switch buttonLabel {
			case "Restart":
				doRestart(serviceName, clusterName)
				ui.App.Content.RemovePage("modal")
				ui.App.TviewApp.SetInputCapture(inputHandler)
			case "Cancel":
				ui.App.Content.RemovePage("modal")
				ui.App.TviewApp.SetInputCapture(inputHandler)
			}
		})

	ui.App.Content.AddAndSwitchToPage("modal", modal, false)
}

func doRestart(serviceName, clusterName string) {
	err := aws.RestartECSService(clusterName, serviceName)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to restart ECS service %s", serviceName)
		ui.CreateMessageBox("Failed to restart service, see log for more information.")
	}

	ui.CreateMessageBox(fmt.Sprintf("ECS service %s restarted successfully", serviceName))
}
