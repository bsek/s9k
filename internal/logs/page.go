package logs

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ui"
)

var _ ui.ContentPage = (*LogPage)(nil)

type LogPage struct {
	Flex                *tview.Flex
	streams             []string
	logGroupName        string
	highlightedText     string
	logStreamPagesIndex map[string]*LogStreamPage
	logStreamPages      *tview.Pages
	highlightField      *tview.InputField
	currentStreamName   string
	closefunc           func()
}

func NewLogPage(logGroupName string, logStreams []types.LogStream) *LogPage {
	streams := lo.Map(logStreams, func(item types.LogStream, index int) string {
		return *item.LogStreamName
	})

	flex := tview.NewFlex()

	var currentStreamName string
	if len(streams) > 0 {
		currentStreamName = "F1"
	}

	logPage := &LogPage{
		Flex:              flex,
		logGroupName:      logGroupName,
		streams:           streams,
		currentStreamName: currentStreamName,
		highlightField:    tview.NewInputField(),
	}

	logPage.highlightField.SetChangedFunc(logPage.highlightTextChanged)

	flex.SetInputCapture(logPage.inputHandler)

	return logPage
}

func (l *LogPage) highlightTextChanged(text string) {
	logStreamPage := l.logStreamPagesIndex[l.currentStreamName]
	logStreamPage.HighlightText(&text)
}

func (l *LogPage) inputHandler(event *tcell.EventKey) *tcell.EventKey {
	if l.selectLogStreamPageByIndex(event.Key()) {
		return event
	}

	if event.Key() == tcell.KeyRune {

		key := event.Rune()

		if key == 't' || key == 'T' {
			logStreamPage := l.logStreamPagesIndex[l.currentStreamName]
			logStreamPage.SwitchFollow()
		}

		if key == 'w' || key == 'W' {
			logStreamPage := l.logStreamPagesIndex[l.currentStreamName]
			logStreamPage.SwitchWrap()
		}
	}

	return event
}

func (l *LogPage) selectLogStreamPageByIndex(key tcell.Key) bool {
	log.Debug().Msgf("Key pressed: %s", tcell.KeyNames[key])

	if key >= tcell.KeyF1 && key <= tcell.KeyF9 {
		// stop reading log records for current stream
		logStreamPage := l.logStreamPagesIndex[l.currentStreamName]
		logStreamPage.End()

		if value, exist := l.logStreamPagesIndex[tcell.KeyNames[key]]; exist {
			go value.LoadData()
			l.currentStreamName = tcell.KeyNames[key]
			l.logStreamPages.SwitchToPage(value.StreamName)
		}

		return true
	}

	return false
}

func (l *LogPage) buildUI() {
	l.Flex.Clear()

	l.logStreamPages = tview.NewPages()
	l.logStreamPagesIndex = make(map[string]*LogStreamPage, len(l.streams))

	if len(l.streams) == 0 {
		text := tview.NewTextView().SetText("No log streams found...")
		l.Flex.
			SetDirection(tview.FlexRow).
			AddItem(text, 0, 1, false)
	} else {
		for i, v := range l.streams {
			p := NewLogStreamPage(l.logGroupName, v, i == 0)
			l.logStreamPagesIndex[fmt.Sprintf("F%d", i+1)] = p
			l.logStreamPages.AddPage(v, p.View, true, i == 0)
		}

		l.Flex.
			SetDirection(tview.FlexRow).
			AddItem(l.logStreamPages, 0, 1, false).
			AddItem(l.highlightField, 2, 1, false)
	}
}

func buildContextMenu(streams []string) *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)

	streamBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)

	pageCommands := make([]string, 0)
	for index, stream := range streams {
		pageCommands = append(pageCommands, fmt.Sprintf(`[white::b]F%d ["%d"][darkcyan]%s[white][""]`, index+1, index+1, stream))
	}
	sort.Strings(pageCommands)

	footerPageText := strings.Join(pageCommands, "\n")

	fmt.Fprint(streamBar, footerPageText)

	configBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(true)

	bw := configBar.BatchWriter()
	defer bw.Close()

	fmt.Fprintln(bw, "[white::b]w [darkcyan::-]wrap")
	fmt.Fprintln(bw, "[white::b]t [darkcyan::-]tail")
	fmt.Fprintln(bw, "[white::b]p [darkcyan::-]parse json")
	fmt.Fprintln(bw, "")
	fmt.Fprintln(bw, "[white::b]0 [darkcyan::-]1m")
	fmt.Fprintln(bw, "[white::b]1 [darkcyan::-]5m")
	fmt.Fprintln(bw, "[white::b]2 [darkcyan::-]15m")
	fmt.Fprintln(bw, "[white::b]3 [darkcyan::-]30m")

	flex.AddItem(streamBar, 0, 4, false)
	flex.AddItem(configBar, 0, 1, false)

	return flex
}

func (l *LogPage) ContextView() tview.Primitive {
	return buildContextMenu(l.streams)
}

func (l *LogPage) Name() string {
	return "logs"
}

func (l *LogPage) Render(accountData *data.AccountData) {
	l.buildUI()
}

func (l *LogPage) SetFocus(app *tview.Application) {
	if l.Flex.GetItemCount() > 0 {
		app.SetFocus(l.Flex.GetItem(0))
	} else {
		app.SetFocus(l.Flex)
	}
}

func (l *LogPage) IsPersistent() bool {
	return false
}

func (l *LogPage) View() tview.Primitive {
	return l.Flex
}

func (l *LogPage) Close() {
	log.Debug().Msgf("Trying to close log stream...")
	logStreamPage := l.logStreamPagesIndex[l.currentStreamName]
	if logStreamPage != nil {
		logStreamPage.End()
	}
}
