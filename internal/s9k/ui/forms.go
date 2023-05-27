package ui

import "github.com/rivo/tview"

func CreateInstallForm(functionName string, deployables []string, selected func(string, int), quit func(), deploy func()) *tview.Form {
	form := tview.NewForm().
		AddDropDown("Deployable", deployables, 0, selected).
		AddButton("Deploy", deploy).
		AddButton("Cancel", quit)

	return form
}
