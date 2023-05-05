package logs

import (
	"fmt"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type LogStreamPage struct {
	StreamName   string
	LogGroupName string
	View         *tview.TextView
}

func NewLogStreamPage(logGroupName, logStreamName string, load bool) *LogStreamPage {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetRegions(true).
		SetChangedFunc(func() {})

	textView.SetBorder(true)

	page := LogStreamPage{
		StreamName:   logStreamName,
		LogGroupName: logGroupName,
		View:         textView,
	}

	if load {
		page.LoadData()
	}

	return &page
}

func (p *LogStreamPage) LoadData() {
	p.View.SetTitle(fmt.Sprintf(" %s (Loading...) ", p.StreamName))

	output, err := aws.FetchCloudwatchLogs(p.LogGroupName, p.StreamName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load log data")
		return
	}

	p.View.SetTitle(fmt.Sprintf(" %s (%d rows) ", p.StreamName, len(output)))

	for _, v := range output {
		for _, v2 := range v {
			fmt.Fprintf(p.View, "%s\n", *v2.Message)
		}
	}
}
