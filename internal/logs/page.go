package logs

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"

	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ui"
)

var _ ui.ContentPage = (*LogPage)(nil)

type LogPage struct {
	Flex            *tview.Flex
	logStreamPage   *LogStreamPage
	highlightField  *tview.InputField
	closefunc       func()
	logGroupArn     string
	logStreams      []types.LogStream
	highlightedText string
}

func NewLogPage(logGroupArn string) *LogPage {
	flex := tview.NewFlex()

	logPage := &LogPage{
		Flex:           flex,
		logGroupArn:    logGroupArn,
		highlightField: tview.NewInputField(),
	}

	logPage.highlightField.SetChangedFunc(logPage.highlightTextChanged)

	flex.SetInputCapture(logPage.inputHandler)

	return logPage
}

func (l *LogPage) highlightTextChanged(text string) {
	l.logStreamPage.HighlightText(&text)
}

func (l *LogPage) inputHandler(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {

		key := event.Rune()

		if key == 't' || key == 'T' {
			l.logStreamPage.SwitchFollow()
		}

		if key == 'w' || key == 'W' {
			l.logStreamPage.SwitchWrap()
		}
	}

	return event
}

func (l *LogPage) buildUI() {
	l.Flex.Clear()

	// if len(l.logStreams) == 0 {
	// 	text := tview.NewTextView().SetText(fmt.Sprintf("S9K could not find any log streams in log group [%s] with data for the last 30 minutes.", l.logGroupName))
	// 	l.Flex.
	// 		SetDirection(tview.FlexRow).
	// 		AddItem(text, 0, 1, false)
	// } else {
	p := NewLogStreamPage(l.logGroupArn, true)
	l.logStreamPage = p

	l.Flex.
		SetDirection(tview.FlexRow).
		AddItem(l.logStreamPage.View, 0, 1, false).
		AddItem(l.highlightField, 2, 1, false)
	// }
}

func buildContextMenu() *tview.Flex {
	flex := tview.NewFlex().SetDirection(tview.FlexColumn)

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

	flex.AddItem(configBar, 0, 1, false)

	return flex
}

func (l *LogPage) ContextView() tview.Primitive {
	return buildContextMenu()
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
	if l.logStreamPage != nil {
		l.logStreamPage.End()
	}
}
