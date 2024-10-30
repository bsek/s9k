package entrypoint

import (
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/apigateway"
	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ecs"
	"github.com/bsek/s9k/internal/lambda"
	"github.com/bsek/s9k/internal/ui"
)

func LoadData() *data.AccountData {
	account, _, err := aws.GetAccountInformation()
	if err != nil {
		fmt.Println("Failed to read account information, make sure you are logged in to AWS")
		log.Fatal().Err(err).Msg("Failed to read account information, make sure you are logged in to AWS")
	}

	clusters, err := aws.ListECSClusters()
	if err != nil || len(clusters) < 1 {
		fmt.Println("Failed to read ecs clusters or no clusters found")
		log.Fatal().Err(err).Msg("Failed to read ecs clusters or no clusters found")
	}

	clusterName := clusters[0].ClusterName
	accountData := data.NewAccountData(*clusterName, *account)
	accountData.Refresh()

	return accountData
}

// Entrypoint for the application
func Entrypoint() {
	fmt.Println("Loading information...")

	accountData := LoadData()

	ui.App = ui.NewApplication(accountData)

	servicesPage := ecs.NewServicesPage()
	lambdasPage := lambda.NewLambdasPage()
	apigatewayPage := apigateway.NewApiGatewayPage()

	ui.App.BuildApplicationUI()

	ui.App.RegisterContent(servicesPage)
	ui.App.RegisterContent(lambdasPage)
	ui.App.RegisterContent(apigatewayPage)

	ui.App.ShowPage(servicesPage)

	if err := ui.App.Run(); err != nil {
		fmt.Println("Failed to start application")
		log.Fatal().Err(err).Msg("Failed to start application")
	}
}
