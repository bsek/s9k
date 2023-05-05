package logs

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/s9k/aws"
)

var logStreamPagesIndex map[string]*LogStreamPage
var logStreamPages *tview.Pages

type LogPage struct {
	Flex *tview.Flex
}

func NewLogPage(logGroupName, taskArn, containerName string, closeFunc func()) *LogPage {
	output, err := aws.FetchLogStreams(logGroupName, containerName, taskArn)
	if err != nil {
		log.Printf("Failed to read log streams %v\n", err)
		return nil
	}

	streams := lo.Map(output, func(item types.LogStream, index int) string {
		return *item.LogStreamName
	})

	inputHandler := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			key := event.Rune()

			// close
			if key == 'q' || key == 'Q' {
				closeFunc()
			}
		} else {
			if selectLogStreamPageByIndex(event.Key()) {
				return event
			}
		}

		return event
	}

	flex := buildUI(logGroupName, streams)
	flex.SetInputCapture(inputHandler)

	return &LogPage{
		Flex: flex,
	}
}

func showLogStreamPage(page *LogStreamPage) {
	page.LoadData()
	logStreamPages.SwitchToPage(page.StreamName)
}

func selectLogStreamPageByIndex(key tcell.Key) bool {
	if key >= tcell.KeyF1 && key <= tcell.KeyF9 {
		if value, exist := logStreamPagesIndex[tcell.KeyNames[key]]; exist == true {
			showLogStreamPage(value)
		}
		return true
	}
	return false
}

func buildUI(logGroupName string, streams []string) *tview.Flex {
	flex := tview.NewFlex()
	logStreamPages = tview.NewPages()
	logStreamPagesIndex = make(map[string]*LogStreamPage, len(streams))

	for i, v := range streams {
		p := NewLogStreamPage(logGroupName, v, i == 0)
		logStreamPagesIndex[fmt.Sprintf("F%d", i+1)] = p
		logStreamPages.AddPage(v, p.View, true, i == 0)
	}

	commandFooterBar := buildCommandFooterBar(streams)

	footer := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(commandFooterBar, 0, 1, false)

	flex.
		SetDirection(tview.FlexRow).
		AddItem(logStreamPages, 0, 1, false).
		AddItem(footer, 3, 1, false)

	return flex
}

func buildCommandFooterBar(streams []string) *tview.TextView {
	footerBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)

	pageCommands := make([]string, 0)
	for index, stream := range streams {
		pageCommands = append(pageCommands, fmt.Sprintf(`[bold]F%d ["%d"][darkcyan]%s[white][""]`, index+1, index+1, stream))
	}
	sort.Strings(pageCommands)

	footerPageText := strings.Join(pageCommands, " ")

	fmt.Fprint(footerBar, footerPageText)

	return footerBar
}
