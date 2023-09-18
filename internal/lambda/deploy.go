package lambda

import (
	"fmt"
	"strings"

	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
)

func retrieveListOfDeployables(functionName string) []string {
	versions, err := aws.FetchAvailableVersions(functionName)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to read versions from s3")
		return []string{}
	}

	list := make([]string, 0, len(versions))
	for _, v := range versions {
		pos := strings.Index(v.Name, "/")
		list = append(list, v.Name[pos+1:])
	}

	return list
}

func createInstallForm(_ string, deployables []string, selected func(string, int), quit func(), deploy func()) *tview.Form {
	form := tview.NewForm().
		AddDropDown("Deployable", deployables, 0, selected).
		AddButton("Deploy", deploy).
		AddButton("Cancel", quit)

	return form
}

func deploy(functionName string, arch lambdatypes.Architecture) {
	const DEPLOY_DIALOG = "deploy_dialog"
	pages := ui.App.Content
	var selectedOption string

	form := createInstallForm(functionName, retrieveListOfDeployables(functionName),
		func(option string, _ int) {
			selectedOption = option
		},
		func() {
			pages.RemovePage(DEPLOY_DIALOG)
		},
		func() {
			log.Info().Msgf("Deployed version %s", selectedOption)

			arn, err := aws.DeployLambdaFunction(functionName, selectedOption, arch)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update %s to version %s", functionName, selectedOption)
				ui.CreateMessageBox(fmt.Sprintf("Failed to update %s to version %s", functionName, selectedOption))
			} else {
				err := aws.TagLambdaFunctionWithVersion(*arn, selectedOption)
				if err != nil {
					log.Error().Err(err).Msgf("Failed to tag %s with version %s", functionName, selectedOption)
				}
				ui.CreateMessageBox(fmt.Sprintf("%s successfully updated to version %s", functionName, selectedOption))
			}

			pages.RemovePage(DEPLOY_DIALOG)
		})

	form.SetBorder(true).SetTitle("Which version do you want to deploy?").SetTitleAlign(tview.AlignLeft)

	modalPage := ui.CreateModalPage(form, nil, 60, 10, DEPLOY_DIALOG)

	pages.AddPage(DEPLOY_DIALOG, modalPage, true, true)
}
