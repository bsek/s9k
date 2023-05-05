package ui

import "github.com/rivo/tview"

func CreateInstallForm(functionName string, deployables []string, selected func(string, int), quit func(), deploy func()) *tview.Form {
	form := tview.NewForm().
		AddDropDown("Deployable", deployables, 0, selected).
		AddButton("Deploy", deploy).
		AddButton("Cancel", quit)

	return form
}

func CreateActionForm(functionName string, deploy func(), restart func(), quit func()) *tview.Form {
	form := tview.NewForm().
		AddButton("Deploy specific version", deploy).
		AddButton("Restart service", restart).
		AddButton("Close", quit)
	return form
}

func CreateTaskForm(logs func(), shell func(), quit func()) *tview.Form {
	form := tview.NewForm().
		AddButton("Show logs", logs).
		AddButton("Open shell", shell).
		AddButton("Close", quit)
	return form
}
