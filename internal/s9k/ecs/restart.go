package ecs

import "github.com/rivo/tview"

func restart(serviceName string, pages *tview.Pages) {
	modal := tview.NewModal().
		SetText("Do you really want to restart the service?").
		AddButtons([]string{"Restart", "Cancel"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Restart" {

			} else if buttonLabel == "Cancel" {
				pages.RemovePage("modal")
			}
		})
	pages.AddAndSwitchToPage("modal", modal, false)
}
