package lambda

import (
	"fmt"
	"strings"

	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

const DEPLOY_DIALOG = "deploy_dialog"

func retrieveListOfDeployables(functionName string) []string {
	versions, err := aws.FetchAvailableVersions(functionName)
	if err != nil {
		log.Error().Msgf("Failed to read versions from s3 %v", err)
		return []string{}
	}

	list := make([]string, 0, len(versions))
	for _, v := range versions {
		pos := strings.Index(v.Name, "/")
		list = append(list, v.Name[pos+1:])
	}

	return list
}

func deploy(functionName string, arch lambdatypes.Architecture, pages *tview.Pages) {
	var selectedOption string

	form := ui.CreateInstallForm(functionName, retrieveListOfDeployables(functionName),
		func(option string, optionIndex int) {
			selectedOption = option
		},
		func() {
			pages.RemovePage(DEPLOY_DIALOG)
		},
		func() {
			log.Info().Msgf("Deployed version %s", selectedOption)

			err := aws.DeployLambdaFunction(functionName, selectedOption, arch)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to update %s to version %s", functionName, selectedOption)
				ui.CreateMessageBox(fmt.Sprintf("Failed to update %s to version %s", functionName, selectedOption), pages)
			} else {
				ui.CreateMessageBox(fmt.Sprintf("%s updated to version %s sucessfully", functionName, selectedOption), pages)
			}

			pages.RemovePage(DEPLOY_DIALOG)
		})

	form.SetBorder(true).SetTitle("Which version do you want to deploy?").SetTitleAlign(tview.AlignLeft)

	modalPage := ui.CreateModalPage(form, nil, 60, 10, DEPLOY_DIALOG)

	pages.AddPage(DEPLOY_DIALOG, modalPage, true, true)
}
