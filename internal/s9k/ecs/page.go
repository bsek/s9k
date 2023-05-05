package ecs

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/github"
	"github.com/bsek/s9k/internal/s9k/logs"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

type ServicePage struct {
	name              string
	servicesTableInfo *ui.TableInfo
}

// Returns a page that displays the services in a cluster
func NewServicesPage(app *tview.Application, flex *tview.Flex) *ServicePage {
	servicesTable := tview.NewTable()

	var pages *tview.Pages

	servicesTable.
		SetBorders(true).
		SetBorder(true).
		SetTitle(" 📋 ECS Services ")

	servicesTable.SetSelectable(true, false)

	servicesTable.SetSelectedFunc(func(row, column int) {
		cell := servicesTable.GetCell(row, 1)
		serviceName := cell.Text
		service := cell.Reference.(types.Service)
		inputCaptureFunction := app.GetInputCapture()

		closeFunction := func() {
			app.SetInputCapture(inputCaptureFunction)
			app.SetRoot(flex, true)
			app.SetFocus(flex.GetItem(0))
		}

		deployFunction := func(version string) {
			deploy(utils.RemoveAllBeforeLastChar("/", *service.ClusterArn), serviceName, version, pages)
		}

		restartFunction := func() {
			restart(serviceName, pages)
		}

		actionsFunc := func(task *types.Task, containerName string) {
			action(*task.TaskArn, serviceName, containerName, pages, app)
		}

		detailsPage := NewServiceDetailsPage(app, &service, deployFunction, restartFunction, actionsFunc, closeFunction)

		pages = ui.CreateModalPage(detailsPage.Flex, flex, 150, 40, "details")

		app.SetInputCapture(nil)
		app.SetRoot(pages, true)
	})

	servicesTableInfo := &ui.TableInfo{
		Table:      servicesTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.R},
		Expansions: []int{1, 1, 2, 1, 1, 1, 1, 1},
		Selectable: true,
	}

	ui.AddTableConfigData(servicesTableInfo, 0, [][]string{
		{"#", "Name ▾", "TaskDef", "Images", "Status", "Deployed", "Deployment status", "Tasks"},
	}, tcell.ColorYellow)

	return &ServicePage{
		name:              "services",
		servicesTableInfo: servicesTableInfo,
	}
}

func retrieveListOfECSDeployables(serviceName string) []string {
	packages, _ := github.FetchPackagesfromGhcr(fmt.Sprintf("skjema-%s", serviceName))
	list := make([]string, 0, len(packages))
	for _, v := range packages {
		list = append(list, v.Image)
	}
	return list
}

func getLogGroupName(pageName, selection string) string {
	var prefix string

	switch pageName {
	case "Lambda functions":
		prefix = "/aws/lambda"
	case "Services":
		prefix = "/ecs"
	}

	return fmt.Sprintf("%s/%s", prefix, selection)
}

func showLogs(taskArn, serviceName, containerName string, pages *tview.Pages) {
	logGroupName := getLogGroupName("Services", serviceName)

	log.Info().Msgf("Found log group name: %s", logGroupName)

	closeFunc := func() {
		pages.RemovePage("logs")
	}

	logPage := logs.NewLogPage(logGroupName, taskArn, containerName, closeFunc)

	pages.AddAndSwitchToPage("logs", logPage.Flex, true)
}

func (p *ServicePage) Table() *tview.Table {
	return p.servicesTableInfo.Table
}

func (p *ServicePage) Name() string {
	return p.name
}

func (p *ServicePage) Render(accountData *data.AccountData) {
	var clusterData = accountData.ECSClusterData

	ui.TruncTableRows(p.servicesTableInfo.Table, 1)

	if len(clusterData.Services) == 0 {
		return
	}

	serviceData := lo.Map(clusterData.Services, func(service data.ServiceData, index int) []string {

		serviceImages := service.Tasks

		deployTimeTxt := "n/a"
		deployStatus := ""
		if len(service.Service.Deployments) > 0 {
			deployTimeTxt = utils.FormatLocalDateTime(*service.Service.Deployments[0].CreatedAt)
			for _, v := range service.Service.Deployments {
				if *v.Status == "PRIMARY" {
					deployStatus = *v.RolloutStateReason
				}
			}
		}

		taskCount := utils.I32ToString(service.Service.RunningCount)
		if service.Service.PendingCount > 0 {
			taskCount = fmt.Sprintf("%s (%d pending)", taskCount, service.Service.PendingCount)
		}
		if service.Service.DesiredCount != service.Service.RunningCount {
			taskCount = fmt.Sprintf("%s (%d desired)", taskCount, service.Service.DesiredCount)
		}

		return []string{
			*service.Service.ServiceName,
			utils.RemoveAllRegex(`.*/`, *service.Service.TaskDefinition),
			strings.Join(serviceImages, ","),
			utils.LowerTitle(*service.Service.Status),
			deployTimeTxt,
			deployStatus,
			taskCount,
		}
	})

	tableData := ui.PrependRowNumColumn(serviceData)

	ui.AddTableConfigData(p.servicesTableInfo, 1, tableData, tcell.ColorWhite)

	// set reference to service
	for i := 1; i < len(clusterData.Services)+1; i++ {
		cell := p.servicesTableInfo.Table.GetCell(i, 1)
		cell.SetReference(clusterData.Services[i-1].Service)
	}
}
