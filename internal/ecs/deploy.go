package ecs

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
)

func deploy(clusterName, serviceName, version string) {
	const DEPLOY_DIALOG = "deploy_dialog"

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to deploy version %s?", version)).
		AddButtons([]string{"Deploy", "Cancel"}).
		SetDoneFunc(func(_ int, buttonLabel string) {
			if buttonLabel == "Deploy" {
				log.Debug().Msgf("Deploying version: [%s] for service [%s]", version, serviceName)
				err := aws.UpdateECSImage(version, serviceName, clusterName)
				if err != nil {
					ui.CreateMessageBox("Failed to deploy version, check log file")
				}
				ui.CreateMessageBox("Deploy request successfully sent")
				ui.App.Content.RemovePage(DEPLOY_DIALOG)
			}
			if buttonLabel == "Cancel" {
				ui.App.Content.RemovePage(DEPLOY_DIALOG)
			}
		})

	ui.App.Content.AddAndSwitchToPage(DEPLOY_DIALOG, modal, false)
}
