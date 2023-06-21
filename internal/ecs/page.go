package ecs

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/github"
	"github.com/bsek/s9k/internal/ui"
	"github.com/bsek/s9k/internal/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"
)

var _ ui.ContentPage = (*ServicePage)(nil)

const SERVICE_DETAIL_PAGE = "details"

type ServicePage struct {
	name          string
	servicesTable *tview.Table
	header        *tview.TextView
	inputCapture  func()
}

// Returns a page that displays the services in a cluster
func NewServicesPage() *ServicePage {
	servicesTable := tview.NewTable()

	servicesTable.
		SetBorder(true).
		SetTitle(" 📋 ECS Services ")

	servicesTable.SetSelectable(true, false)

	servicesTable.SetSelectedFunc(func(row, column int) {
		cell := servicesTable.GetCell(row, 1)
		service := cell.Reference.(data.ServiceData)

		deployFunction := func(version string) {
			deploy(utils.RemoveAllBeforeLastChar("/", *service.Service.ClusterArn), *service.Service.ServiceName, version)
		}

		restartFunction := func() {
			restart(*service.Service.ServiceName)
		}

		actionsFunc := func(task *types.Task, container data.Container) {
			action(*task.TaskArn, *service.Service.ServiceName, *service.Service.ClusterArn, container)
		}

		detailsPage := NewServiceDetailsPage(&service, deployFunction, restartFunction, actionsFunc)

		ui.App.RegisterContent(detailsPage)
		ui.App.ShowPage(detailsPage)
	})

	return &ServicePage{
		name:          "services",
		servicesTable: servicesTable,
		header:        createHelpText(),
	}
}

func createHelpText() *tview.TextView {
	tw := tview.NewTextView()
	tw.SetDynamicColors(true).SetWrap(false)

	fmt.Fprintln(tw, "[::b]Enter [darkcyan::-]action")

	return tw
}

func retrieveListOfECSDeployables(serviceName string) []string {
	packages, _ := github.FetchPackagesfromGhcr(fmt.Sprintf("skjema-%s", serviceName))
	list := make([]string, 0, len(packages))
	for _, v := range packages {
		list = append(list, v.Image)
	}

	return list
}

func (p *ServicePage) Table() *tview.Table {
	return p.servicesTable
}

func (p *ServicePage) Name() string {
	return p.name
}

func (p *ServicePage) ContextView() tview.Primitive {
	return p.header
}

func (p *ServicePage) HandleInputCapture(event *tcell.EventKey) *tcell.EventKey {
	return event
}

func (p *ServicePage) Render(accountData *data.AccountData) {
	clusterData := accountData.ClusterData

	if len(clusterData.Services) == 0 {
		return
	}

	serviceData := lo.Map(clusterData.Services, func(service data.ServiceData, index int) []string {
		containers := service.Containers

		deployTimeTxt := "n/a"
		deployStatus := ""
		if len(service.Service.Deployments) > 0 {
			deployTimeTxt = utils.FormatLocalDateTime(*service.Service.Deployments[0].CreatedAt)
			for _, v := range service.Service.Deployments {
				if *v.Status == "PRIMARY" {
					deployStatus = string(*(&v.RolloutState))
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

		var image string
		image = containers[0].Image

		return []string{
			*service.Service.ServiceName,
			utils.RemoveAllBeforeLastChar("/", *service.Service.TaskDefinition),
			image,
			deployTimeTxt,
			deployStatus,
			taskCount,
		}
	})

	alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignRight}
	expansions := []int{1, 1, 1, 2, 1, 1, 1, 1}
	headers := []string{"#", "Name ▾", "TaskDef", "Images", "Last Deployed", "Deployment status", "Tasks"}

	ui.PrependRowNumColumn(serviceData)
	ui.AddTableData(p.servicesTable, headers, serviceData, alignment, expansions, tcell.ColorWhite, true)

	// set reference to service
	for i := 1; i < len(clusterData.Services)+1; i++ {
		cell := p.servicesTable.GetCell(i, 1)
		cell.SetReference(clusterData.Services[i-1])
	}
}

func (s *ServicePage) SetFocus(app *tview.Application) {
	app.SetFocus(s.servicesTable)
}

func (s *ServicePage) View() tview.Primitive {
	return s.servicesTable
}

func (s *ServicePage) IsPersistent() bool {
	return true
}

func (s *ServicePage) Close() {
}
