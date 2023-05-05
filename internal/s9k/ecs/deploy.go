package ecs

import (
	"fmt"

	"github.com/bsek/s9k/internal/s9k/github"
	"github.com/rivo/tview"
)

func deploy(clusterName, serviceName, version string, pages *tview.Pages) {
	modal := tview.NewModal().
		SetText(fmt.Sprintf("Do you want to deploy version %s?", version)).
		AddButtons([]string{"Deploy", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Deploy" {
				github.CallGithubAction(clusterName, serviceName, version)
				pages.RemovePage("modal")
			}
			if buttonLabel == "Cancel" {
				pages.RemovePage("modal")
			}
		})
	pages.AddAndSwitchToPage("modal", modal, false)
}
