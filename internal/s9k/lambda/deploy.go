package lambda

import (
	"log"
	"strings"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/rivo/tview"
)

const deployDialog = "Deploy dialog"

func retrieveListOfDeployables(name, functionName string) []string {
	versions, err := aws.FetchAvailableVersions(name, functionName)
	if err != nil {
		log.Printf("Failed to read versions from s3 %v\n", err)
	}
	list := make([]string, 0, len(versions))
	for _, v := range versions {
		pos := strings.Index(v.Name, "/")
		list = append(list, v.Name[pos+1:])
	}
	return list
}

func deploy(name, functionName string, pages *tview.Pages, closeFunction func()) {
	var selectedOption string
	form := ui.CreateInstallForm(functionName, retrieveListOfDeployables(name, functionName),
		func(option string, optionIndex int) {
			selectedOption = option
		},
		func() {
			pages.RemovePage(deployDialog)
			closeFunction()
		}, func() {
			log.Printf("%s", selectedOption)
			pages.RemovePage(deployDialog)
			closeFunction()
		})

	form.SetBorder(true).SetTitle("Which version do you want to deploy?").SetTitleAlign(tview.AlignLeft)

	modalPage := ui.CreateModalPage(form, nil, 60, 10, deployDialog)

	pages.AddPage(deployDialog, modalPage, true, true)
}
