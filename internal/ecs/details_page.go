package ecs

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/github"
	"github.com/bsek/s9k/internal/ui"
	"github.com/bsek/s9k/internal/utils"
)

var _ ui.ContentPage = (*ServiceDetailPage)(nil)
var region = os.Getenv("AWS_REGION")

type ServiceDetailPage struct {
	Flex        *tview.Flex
	CurrentItem int
}

func NewServiceDetailsPage(inputData *data.ServiceData, deployFunc func(version string), restartFunc func(), openActions func(task *types.Task, container data.Container)) *ServiceDetailPage {
	clusterName := utils.RemoveAllBeforeLastChar("/", inputData.Service.ClusterArn)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(createServiceDetailsTable(inputData.Service, clusterName), 6, 1, false).
		AddItem(createServiceTaskTable(inputData, clusterName, openActions), 0, 3, false).
		AddItem(createDeployablesTable(inputData.Service, clusterName, deployFunc), 0, 3, false)

	page := &ServiceDetailPage{
		Flex:        flex,
		CurrentItem: 1,
	}

	handler := page.createInputHandler(restartFunc)
	flex.SetInputCapture(handler)
	flex.SetBorder(true)

	return page
}

func (s *ServiceDetailPage) createInputHandler(restartFunc func()) func(event *tcell.EventKey) *tcell.EventKey {
	function := func(event *tcell.EventKey) *tcell.EventKey {
		app := ui.App

		if event.Key() == tcell.KeyTab {
			s.CurrentItem = (s.CurrentItem % (s.Flex.GetItemCount() - 1)) + 1
			app.TviewApp.SetFocus(s.Flex.GetItem(s.CurrentItem))
		}

		if event.Key() == tcell.KeyRune {
			key := event.Rune()

			if key == 'r' || key == 'R' {
				restartFunc()
			}
		}

		return event
	}

	return function
}

func createServiceDetailsTable(service *types.Service, clusterName string) *tview.Table {
	detailsTable := tview.NewTable()

	detailsTable.
		SetBorder(true).
		SetTitle(fmt.Sprintf(" 📋 %s details ", *service.ServiceName))

	// fetch metrics
	meterWidth := 10
	cpuMeter := ""
	memoryMeter := ""
	cpuUsed, cpuReserved, memUsed, memReserved, err := aws.FetchCpuAndMemoryUsage(*service.ServiceName, clusterName)
	if err == nil {
		cpuMeter = utils.BuildAsciiMeterCurrentTotal(cpuUsed, cpuReserved, meterWidth)
		memoryMeter = utils.BuildAsciiMeterCurrentTotal(memUsed, memReserved, meterWidth)
	}

	tableData := [][]string{}

	deployTimeTxt := "n/a"
	if len(service.Deployments) > 0 {
		deployTimeTxt = utils.FormatLocalDateTime(*service.Deployments[0].CreatedAt)
	}

	taskCount := utils.I32ToString(service.RunningCount)
	if service.PendingCount > 0 {
		taskCount = fmt.Sprintf("%s (%d pending)", taskCount, service.PendingCount)
	}
	if service.DesiredCount != service.RunningCount {
		taskCount = fmt.Sprintf("%s (%d desired)", taskCount, service.DesiredCount)
	}

	tableData = append(tableData, []string{
		*service.ServiceName,
		utils.RemoveAllBeforeLastChar("/", service.TaskDefinition),
		utils.LowerTitle(*service.Status),
		deployTimeTxt,
		taskCount,
		cpuMeter,
		memoryMeter,
	})

	headers := []string{"Name", "Task Definition", "Status", "Deployed", "Tasks", "Cpu usage", "Memory usage"}
	alignments := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignRight, tview.AlignRight}
	expansions := []int{1, 1, 1, 1, 1, 1, 1}

	ui.AddTableData(detailsTable, headers, tableData, alignments, expansions, tcell.ColorGreenYellow, true)

	return detailsTable
}

func createServiceTaskTable(service *data.ServiceData, clusterName string, openActions func(task *types.Task, container data.Container)) *tview.Table {
	log.Info().Msgf("Creating detailspage with data: %v", service)

	containerTable := tview.NewTable().SetSelectable(true, false)

	containerTable.SetBorder(true).
		SetTitle(" 🐳 Containers ")

	// fetch tasks
	tasks, err := aws.DescribeClusterTasks(&clusterName, service.Service.ServiceName)
	if err != nil {
		log.Error().Err(err).Msg("Klarte ikke hente cluster tasks")
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return 0 > strings.Compare(
			utils.RemoveAllBeforeLastChar("/", tasks[i].TaskDefinitionArn),
			utils.RemoveAllBeforeLastChar("/", tasks[j].TaskDefinitionArn))
	})

	tableData := [][]string{{}}

	for _, task := range tasks {
		for _, container := range task.Containers {
			var memory, cpu string

			if container.Memory == nil {
				memory = ""
			} else {
				memory = *container.Memory
			}

			if container.Cpu == nil {
				cpu = ""
			} else {
				cpu = *container.Cpu
			}

			tableData = append(tableData, []string{
				*container.Name,
				utils.RemoveAllBeforeLastChar("/", task.TaskArn),
				utils.RemoveAllBeforeLastChar("/", task.TaskDefinitionArn),
				utils.RemoveAllBeforeLastChar("/", container.Image),
				*container.LastStatus,
				string(container.HealthStatus),
				memory,
				cpu,
			})
		}

		expansions := []int{1, 1, 1, 2, 1, 1, 1, 1}
		headers := []string{"Name", "Task", "Task definition", "Image", "Status", "Health", "Memory", "CPU"}
		alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft}

		ui.AddTableData(containerTable, headers, tableData, alignment, expansions, tcell.ColorLightBlue, true)

		containerTable.SetSelectedFunc(func(row, _ int) {
			cell := containerTable.GetCell(row, 0)
			container, found := containerTable.GetCell(row, 0).Reference.(data.Container)
			if found {
				log.Info().Msgf("Using cell: %s and found container data: %v", cell.Text, container)
				openActions(&task, container)
			}
		})

		// set reference to container
		for i := 0; i < len(task.Containers); i++ {
			cell := containerTable.GetCell(i+2, 0)
			for _, v := range service.Containers {
				log.Info().Msgf("Comparing %s with %s", v.Name, cell.Text)
				if v.Name == cell.Text {
					cell.SetReference(v)
					log.Info().Msgf("Setting container %s to reference %v", cell.Text, v)
				}
			}
		}
	}

	return containerTable
}

func createDeployablesTable(service *types.Service, clusterName string, deployFunc func(version string)) *tview.Table {
	deployTable := tview.NewTable()

	deployTable.
		SetBorder(true).
		SetTitle(" 📦 Deployables ")

	deployTable.SetSelectable(true, false)

	var wg sync.WaitGroup

	var commits []github.Commit
	var packages []aws.Package

	wg.Add(2)
	go func() {
		c := fetchCommits(*service.ServiceName)
		commits = append(commits, c...)
		wg.Done()
	}()

	go func() {
		p := fetchPackages(clusterName, *service.ServiceName)
		packages = append(packages, p...)
		wg.Done()
	}()

	wg.Wait()

	data := [][]string{{}}

	sort.SliceStable(packages, func(i, j int) bool {
		return packages[i].Created.After(packages[j].Created)
	})

	for _, p := range packages {
		msg := ""
		for _, c := range commits {
			if p.Sha[4:] == c.Sha {
				msg = c.Message
			}
		}

		data = append(data, []string{
			utils.FormatLocalDateTime(p.Created),
			p.Sha,
			msg,
		})
	}

	headers := []string{"Created", "Image", "Commit Message"}
	expansions := []int{1, 2, 2}
	alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft}

	ui.AddTableData(deployTable, headers, data, alignment, expansions, tcell.ColorBlue, true)

	deployTable.SetSelectedFunc(func(row, _ int) {
		version := deployTable.GetCell(row, 1).Text
		for _, v := range packages {
			if v.Sha == version {
				deployFunc(fmt.Sprintf("%s.dkr.ecr.%s.amazonaws.com/%s:%s", v.RegistryID, region, v.RepositoryName, version))
			}
		}
	})

	return deployTable
}

func fetchPackages(cluster, name string) []aws.Package {
	ecrRepo := fmt.Sprintf("%s-%s", cluster, name)
	p, err := aws.FetchPackagesFromECR(ecrRepo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read packages from ecr")
	}

	return p
}

func fetchCommits(name string) []github.Commit {
	c, err := github.FetchCommits("fasit-" + name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read commits from github")
	}

	log.Debug().Msgf("Found commits: (%v)", c)

	return c
}

func (*ServiceDetailPage) Name() string {
	return "details page"
}

func (*ServiceDetailPage) Render(accountData *data.AccountData) {
}

func (s *ServiceDetailPage) View() tview.Primitive {
	return s.Flex
}

func (s *ServiceDetailPage) Close() {
}

func (s *ServiceDetailPage) IsPersistent() bool {
	return false
}

func (s *ServiceDetailPage) SetFocus(app *tview.Application) {
	app.SetFocus(s.Flex.GetItem(s.CurrentItem))
}

func (*ServiceDetailPage) ContextView() tview.Primitive {
	tw := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWrap(false)

	bw := tw.BatchWriter()
	defer bw.Close()

	fmt.Fprintln(bw, "[white::b]Tab [darkcyan::-]Select view")
	fmt.Fprintln(bw, "")
	fmt.Fprintln(bw, "[white::b]r [darkcyan::-]Restart service")
	fmt.Fprintln(bw, "")
	fmt.Fprintln(bw, "[white::b]Enter [darkcyan::-]Select")

	return tw
}
