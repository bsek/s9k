package ecs

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/github"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type ServiceDetailPage struct {
	Flex *tview.Flex
}

func NewServiceDetailsPage(app *tview.Application, inputData *data.ServiceData, deployFunc func(version string), restartFunc func(), openActions func(task *types.Task, container data.Container), closeFunc func()) *ServiceDetailPage {
	clusterName := utils.RemoveAllBeforeLastChar("/", *inputData.Service.ClusterArn)
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(createServiceDetailsTable(inputData.Service, clusterName), 7, 1, false).
		AddItem(createServiceTaskTable(inputData, clusterName, openActions), 0, 3, false).
		AddItem(createDeployablesTable(inputData.Service, deployFunc), 0, 3, false).
		AddItem(createActionsRow(restartFunc), 5, 1, false)

	flex.SetInputCapture(createInputHandler(flex, app, closeFunc))
	flex.SetBorder(true)

	return &ServiceDetailPage{
		Flex: flex,
	}
}

func createInputHandler(flex *tview.Flex, app *tview.Application, closeFunc func()) func(event *tcell.EventKey) *tcell.EventKey {
	function := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {

		}

		if event.Key() == tcell.KeyRune {
			key := event.Rune()

			if key == 'c' || key == 'C' {
				app.SetFocus(flex.GetItem(1))
			}

			if key == 'd' || key == 'D' {
				app.SetFocus(flex.GetItem(2))
			}

			if key == 'a' || key == 'A' {
				app.SetFocus(flex.GetItem(3))
			}

			if key == 'q' || key == 'Q' {
				closeFunc()
			}
		}

		return event
	}

	return function
}

func createActionsRow(restartFunc func()) *tview.Flex {
	flex := tview.NewFlex()

	flex.SetBorder(true).SetTitle(" 🔥 (A)ctions ")

	button := tview.NewButton("Restart service").SetSelectedFunc(func() {
		restartFunc()
	})
	button.SetBorder(true)
	flex.AddItem(button, 20, 0, true)

	return flex
}

func createServiceDetailsTable(service *types.Service, clusterName string) *tview.Table {
	detailsTable := tview.NewTable()

	detailsTable.
		SetBorder(true).
		SetTitle(fmt.Sprintf(" 📋 %s details ", *service.ServiceName))

	tableInfo := &ui.TableInfo{
		Table:      detailsTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.R, ui.R},
		Expansions: []int{1, 1, 1, 1, 1, 1, 1},
		Selectable: true,
	}

	ui.AddTableConfigData(tableInfo, 0, [][]string{
		{"Name", "Task Definition", "Status", "Deployed", "Tasks", "Cpu usage", "Memory usage"},
	}, tcell.ColorYellow)

	// fetch metrics
	meterWidth := 10
	cpuMeter := ""
	memoryMeter := ""
	cpuUsed, cpuReserved, memUsed, memReserved, err := aws.FetchCpuAndMemoryUsage(*service.ServiceName)
	if err == nil {
		cpuMeter = utils.BuildAsciiMeterCurrentTotal(cpuUsed, cpuReserved, meterWidth)
		memoryMeter = utils.BuildAsciiMeterCurrentTotal(memUsed, memReserved, meterWidth)
	}

	var tableData [][]string

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
		utils.RemoveAllBeforeLastChar("/", *service.TaskDefinition),
		utils.LowerTitle(*service.Status),
		deployTimeTxt,
		taskCount,
		cpuMeter,
		memoryMeter,
	})

	ui.AddTableConfigData(tableInfo, 1, tableData, tcell.ColorWhite)

	return detailsTable
}

func createServiceTaskTable(service *data.ServiceData, clusterName string, openActions func(task *types.Task, container data.Container)) *tview.Frame {
	log.Info().Msgf("Creating detailspage with data: %v", service)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	frame := tview.NewFrame(flex)

	frame.SetBorder(true).
		SetTitle(" 🐳 (C)ontainers ")

	// fetch tasks
	tasks, err := aws.DescribeClusterTasks(&clusterName, service.Service.ServiceName)
	if err != nil {
		log.Error().Err(err).Msg("Klarte ikke hente cluster tasks")
	}

	sort.SliceStable(tasks, func(i, j int) bool {
		return 0 > strings.Compare(
			utils.RemoveAllBeforeLastChar("/", *tasks[i].TaskDefinitionArn),
			utils.RemoveAllBeforeLastChar("/", *tasks[j].TaskDefinitionArn))
	})

	for _, task := range tasks {
		containerTable := tview.NewTable().SetSelectable(true, false)

		var tableData [][]string

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
				utils.RemoveAllBeforeLastChar("/", *task.TaskArn),
				utils.RemoveAllBeforeLastChar("/", *task.TaskDefinitionArn),
				utils.RemoveAllBeforeLastChar("/", *container.Image),
				*container.LastStatus,
				string(container.HealthStatus),
				memory,
				cpu,
			})
		}

		tableInfo := &ui.TableInfo{
			Table:      containerTable,
			Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L, ui.L},
			Expansions: []int{1, 1, 1, 2, 1, 1, 1, 1},
			Selectable: true,
		}

		ui.AddTableConfigData(tableInfo, 0, [][]string{
			{"Name", "Task", "Task definition", "Image", "Status", "Health", "Memory", "CPU"},
		}, tcell.ColorYellow)

		ui.AddTableConfigData(tableInfo, 1, tableData, tcell.ColorWhite)

		tableInfo.Table.SetSelectedFunc(func(row, column int) {
			cell := tableInfo.Table.GetCell(row, 0)
			container := tableInfo.Table.GetCell(row, 0).Reference.(data.Container)
			log.Info().Msgf("Using cell: %s and found container data: %v", cell.Text, container)
			openActions(&task, container)
		})

		// set reference to container
		for i := 0; i < len(service.Containers); i++ {
			cell := tableInfo.Table.GetCell(i+1, 0)
			for _, v := range service.Containers {
				if v.Name == cell.Text {
					cell.SetReference(v)
					log.Info().Msgf("Setting container %s to reference %v", cell.Text, v)
				}
			}
		}

		flex.AddItem(tableInfo.Table, 0, 1, true)
	}

	return frame
}

func createDeployablesTable(service *types.Service, deployFunc func(version string)) *tview.Table {
	deployTable := tview.NewTable()

	deployTable.
		SetBorder(true).
		SetTitle(" 📦 (D)eployables ")

	deployTable.SetSelectable(true, false)

	tableInfo := &ui.TableInfo{
		Table:      deployTable,
		Alignment:  []int{ui.L, ui.L, ui.L},
		Expansions: []int{1, 2, 2},
		Selectable: true,
	}

	ui.AddTableConfigData(tableInfo, 0, [][]string{
		{"Created", "Image", "Commit Message"},
	}, tcell.ColorYellow)

	var wg sync.WaitGroup

	var commits []github.Commit
	var packages []github.Package

	wg.Add(2)
	go func() {
		c := fetchCommits(*service.ServiceName)
		commits = append(commits, c...)
		wg.Done()
	}()

	go func() {
		p := fetchPackages(*service.ServiceName)
		packages = append(packages, p...)
		wg.Done()
	}()

	wg.Wait()

	data := [][]string{{}}

	for _, p := range packages {
		msg := ""
		for _, c := range commits {
			if p.Sha == c.Sha[0:7] {
				msg = c.Message
			}
		}

		data = append(data, []string{
			utils.FormatLocalDateTime(p.Created),
			strings.Replace(p.Image, "ghcr.io/oslokommune/", "", 1),
			msg,
		})
	}

	ui.AddTableConfigData(tableInfo, 1, data, tcell.ColorWhite)

	tableInfo.Table.SetSelectedFunc(func(row, column int) {
		version := tableInfo.Table.GetCell(row, 1).Text
		deployFunc(version)
	})

	return deployTable
}

func fetchPackages(name string) []github.Package {
	p, err := github.FetchPackagesfromGhcr(fmt.Sprintf("skjema-%s", name))
	if err != nil {
		log.Error().Err(err).Msg("Failed to read packages from github")
	}

	return p
}

func fetchCommits(name string) []github.Commit {
	c, err := github.FetchCommits(fmt.Sprintf("skjema-app-%s", name))
	if err != nil {
		log.Error().Err(err).Msg("Failed to read commits from github")
	}

	return c
}
