package logs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
)

type LogStreamPage struct {
	View        *tview.TextView
	NextToken   *string
	stream      *cloudwatchlogs.StartLiveTailEventStream
	LogGroupArn string
	ParseFields []string
	//logStreams   []types.LogStream
	wrap   bool
	follow bool
	Json   bool
}

const duration = 2 * time.Second

var (
	timestampRe = regexp.MustCompile(`(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z)`)
	newlineRe   = regexp.MustCompile(`\n`)
)

func NewLogStreamPage(logGroupArn string, load bool) *LogStreamPage {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetRegions(true).
		SetMaxLines(400)

	textView.SetBorder(true)

	stream := openEventStream(logGroupArn)

	page := LogStreamPage{
		LogGroupArn: logGroupArn,
		//	logStreams:   logStreams,
		View:   textView,
		wrap:   false,
		follow: true,
		stream: stream,
	}

	if load {
		go page.LoadData()
	}

	return &page
}

func (p *LogStreamPage) End() {
	if p.stream != nil {
		p.stream.Close()
	}
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
		p.stream.Close()
		//	p.ticker.Reset(duration)
	} else {
		p.stream = openEventStream(p.LogGroupArn)
		go p.LoadData()
		//	p.ticker.Stop()
	}

	p.View.SetTitle(p.createTitle(p.View.GetOriginalLineCount()))
}

func (p *LogStreamPage) HighlightText(text *string) {
	p.View.Highlight(*text)
}

func (p *LogStreamPage) LoadData() {
	eventsChan := p.stream.Events()
	for {
		event := <-eventsChan
		switch e := event.(type) {
		case *types.StartLiveTailResponseStreamMemberSessionStart:
			log.Info().Msg("Received SessionStart event")
		case *types.StartLiveTailResponseStreamMemberSessionUpdate:
			log.Info().Msg("Received tail response")
			ui.App.TviewApp.QueueUpdateDraw(func() {
				bw := p.View.BatchWriter()
				for _, logEvent := range e.Value.SessionResults {
					fmt.Fprintln(bw, stripNewLines(highlightDateTime(*logEvent.Message)))
				}
				bw.Close()
				p.View.ScrollToEnd()
			})
		default:
			// Handle on-stream exceptions
			if err := p.stream.Err(); err != nil {
				log.Fatal().Err(err).Msg("Error occured during streaming")
			} else if event == nil {
				log.Info().Msg("Stream is Closed")
				return
			} else {
				log.Error().Msgf("Unknown event type: %T", e)
			}
		}

		p.View.SetTitle(p.createTitle(p.View.GetOriginalLineCount()))
	}
}

func (p *LogStreamPage) createTitle(length int) string {
	title := fmt.Sprintf(" %s (%d rows", p.LogGroupArn, length)

	if p.follow {
		title = fmt.Sprintf(`%s, tail`, title)
	}
	if p.wrap {
		title = fmt.Sprintf(`%s, wrap`, title)
	}

	title = fmt.Sprintf("%s) ", title)

	return title
}

func openEventStream(logGroupArn string) *cloudwatchlogs.StartLiveTailEventStream {
	stream, err := aws.FetchTailLogsChannel(logGroupArn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load log data")
	}
	return stream
}

// func (p *LogStreamPage) fetchLogItems() {
// 	log.Debug().Msgf("Loading log data from %s", p.StreamName)
// 	title := fmt.Sprintf(" %s (Loading...) ", p.StreamName)

// 	p.View.SetTitle(title)

// 	output, nextToken, err := aws.FetchCloudwatchLogs(p.LogGroupName, p.StreamName, p.NextToken, p.Interval)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to load log data")
// 		return
// 	}

// 	p.NextToken = nextToken

// 	length := p.View.GetOriginalLineCount()

// 	for _, v := range output {
// 		length = length + len(v)
// 	}

// 	log.Debug().Msgf("Found %d rows", length)

// 	p.View.SetTitle(p.createTitle(length))

// 	bw := p.View.BatchWriter()
// 	for _, v := range output {
// 		for _, v2 := range v {
// 			fmt.Fprintln(bw, stripNewLines(highlightDateTime(*v2.Message)))
// 		}
// 	}
// 	bw.Close()

// 	p.View.ScrollToEnd()
// }

func stripNewLines(input string) string {
	return newlineRe.ReplaceAllLiteralString(input, "")
}

func highlightDateTime(input string) string {
	return timestampRe.ReplaceAllString(input, "[lightgreen::b]${1}[white::-]")
}
