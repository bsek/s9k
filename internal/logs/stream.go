package logs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
)

type LogStreamPage struct {
	StreamName   string
	LogGroupName string
	View         *tview.TextView
	NextToken    *string
	wrap         bool
	follow       bool
	Json         bool
	ParseFields  []string
	Interval     time.Duration
	ticker       *time.Ticker
	endChan      chan bool
}

const duration = 2 * time.Second

var (
	timestampRe = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z)`)
	newlineRe   = regexp.MustCompile(`\n`)
)

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
		Interval:     30 * time.Minute,
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

func (p *LogStreamPage) ParseFirstMessage() {
	text := p.View.GetText(false)
	lines := strings.Split(text, "\n")

	if len(lines) > 0 {
		var fieldsMap map[string]any
		err := json.Unmarshal([]byte(lines[0]), &fieldsMap)
		if err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal log text")
			return
		}
		p.ParseFields = lo.Keys(fieldsMap)
	}
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
		title = fmt.Sprintf(`%s, tail`, title)
	}
	if p.wrap {
		title = fmt.Sprintf(`%s, wrap`, title)
	}

	title = fmt.Sprintf("%s, %s) ", title, p.Interval)

	return title
}

func (p *LogStreamPage) fetchLogItems() {
	log.Debug().Msgf("Loading log data from %s", p.StreamName)
	title := fmt.Sprintf(" %s (Loading...) ", p.StreamName)

	p.View.SetTitle(title)

	output, nextToken, err := aws.FetchCloudwatchLogs(p.LogGroupName, p.StreamName, p.NextToken, p.Interval)
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
			fmt.Fprintln(bw, stripNewLines(highlightDateTime(*v2.Message)))
		}
	}
	bw.Close()

	p.View.ScrollToEnd()
}

func stripNewLines(input string) string {
	return newlineRe.ReplaceAllLiteralString(input, "")
}

func highlightDateTime(input string) string {
	return timestampRe.ReplaceAllString(input, "[lightgreen::b]${1}[white::-]")
}
