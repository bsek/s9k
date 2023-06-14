package logs

import (
	"fmt"
	"regexp"
	"time"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type LogStreamPage struct {
	StreamName   string
	LogGroupName string
	View         *tview.TextView
	NextToken    *string
	wrap         bool
	follow       bool
	ticker       *time.Ticker
	interval     time.Duration
	endChan      chan bool
}

const duration = 2 * time.Second

func NewLogStreamPage(logGroupName, logStreamName string, load bool) *LogStreamPage {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetRegions(true).
		SetMaxLines(400)

	textView.SetBorder(true)

	page := LogStreamPage{
		StreamName:   logStreamName,
		LogGroupName: logGroupName,
		View:         textView,
		wrap:         false,
		follow:       true,
		ticker:       time.NewTicker(duration),
		interval:     30 * time.Minute,
		endChan:      make(chan bool),
	}

	if load {
		go page.LoadData()
	}

	return &page
}

func (p *LogStreamPage) End() {
	p.endChan <- true
}

func (p *LogStreamPage) SwitchWrap() {
	p.wrap = !p.wrap
	p.View.SetWrap(p.wrap)

	p.View.SetTitle(p.createTitle(p.View.GetOriginalLineCount()))
}

func (p *LogStreamPage) SwitchFollow() {
	p.follow = !p.follow
	if p.follow {
		p.ticker.Reset(duration)
	} else {
		p.ticker.Stop()
	}

	p.View.SetTitle(p.createTitle(p.View.GetOriginalLineCount()))
}

func (p *LogStreamPage) SetInterval(d time.Duration) {
	p.interval = d
}

func (p *LogStreamPage) HighlightText(text *string) {
	p.View.Highlight(*text)
}

func (p *LogStreamPage) LoadData() {
	ui.App.TviewApp.QueueUpdateDraw(p.fetchLogItems)

	for {
		select {
		case <-p.ticker.C:
			log.Debug().Msgf("Loading log data from %s", p.StreamName)
			ui.App.TviewApp.QueueUpdateDraw(p.fetchLogItems)
		case <-p.endChan:
			log.Debug().Msgf("Stopping reading log records from %s", p.StreamName)
			return
		}
	}
}

func (p *LogStreamPage) createTitle(length int) string {
	title := fmt.Sprintf(" %s (%d rows", p.StreamName, length)

	if p.follow {
		title = fmt.Sprintf(`%s, follow`, title)
	}
	if p.wrap {
		title = fmt.Sprintf(`%s, wrap`, title)
	}

	title = fmt.Sprintf("%s, %s) ", title, p.interval)

	return title
}

func (p *LogStreamPage) fetchLogItems() {
	log.Debug().Msgf("Loading log data from %s", p.StreamName)
	title := fmt.Sprintf(" %s (Loading...) ", p.StreamName)

	p.View.SetTitle(title)

	output, nextToken, err := aws.FetchCloudwatchLogs(p.LogGroupName, p.StreamName, p.NextToken, p.interval)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load log data")
		return
	}

	p.NextToken = nextToken

	length := p.View.GetOriginalLineCount()

	for _, v := range output {
		length = length + len(v)
	}

	log.Debug().Msgf("Found %d rows", length)

	p.View.SetTitle(p.createTitle(length))

	bw := p.View.BatchWriter()
	for _, v := range output {
		for _, v2 := range v {
			fmt.Fprintln(bw, highlightDateTime(*v2.Message))
		}
	}
	bw.Close()

	p.View.ScrollToEnd()
}

func highlightDateTime(input string) string {
	re := regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z)`)
	return re.ReplaceAllString(input, "[lightgreen::b]${1}[white::-]")
}
