package shell

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/ui"
	"github.com/creack/pty"
	"github.com/rs/zerolog/log"
	"golang.org/x/term"
)

func NewShellPage(taskArn, containerName, clusterArn string) {
	output, err := aws.ExecuteCommand(taskArn, containerName, clusterArn)
	if err != nil {
		log.Error().Err(err).Msg("Failed to open pty")
		return
	}

	v := struct {
		SessionID  string `json:"SessionId"`
		StreamURL  string `json:"StreamUrl"`
		TokenValue string `json:"TokenValue"`
	}{
		SessionID:  *output.Session.SessionId,
		StreamURL:  *output.Session.StreamUrl,
		TokenValue: *output.Session.TokenValue,
	}
	json, err := json.Marshal(v)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal input parameters")
		return
	}

	ui.App.TviewApp.Suspend(runSessionManangerPlugin(json))
}

func runSessionManangerPlugin(json []byte) func() {
	return func() {
		clearScreen()

		cmd := exec.Command("session-manager-plugin", string(json), "eu-north-1", "StartSession")

		ptmx, err := pty.Start(cmd)
		if err != nil {
			log.Error().Err(err).Msg("Failed to open pty")
			return
		}
		// Make sure to close the pty at the end.
		defer func() {
			_ = ptmx.Close()
		}() // Best effort.

		// Handle pty size.
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGWINCH)
		go func() {
			for range ch {
				if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
					log.Error().Err(err).Msg("Error resizing pty")
				}
			}
		}()

		ch <- syscall.SIGWINCH // Initial resize.
		defer func() {
			signal.Stop(ch)
			close(ch)
		}() // Cleanup signals when done.

		// Set stdin in raw mode.
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			log.Error().Err(err).Msg("Error setting pty in raw mode")
			return
		}

		defer func() {
			_ = term.Restore(int(os.Stdin.Fd()), oldState)
		}() // Best effort.
		defer clearScreen()

		// Copy stdin to the pty and the pty to stdout.
		// NOTE: The goroutine will keep reading until the next keystroke before returning.
		go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
		_, _ = io.Copy(os.Stdout, ptmx)

	}
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

const MaxBufferSize = 16

func CloseShell(p *os.File) {
	p.Close()
}

// func NewShellPage(taskArn, containerName, clusterName string, app *tview.Application) (*tview.TextView, *os.File) {
// 	textView := tview.NewTextView().
// 		SetDynamicColors(true).
// 		SetChangedFunc(func() {
// 			app.Draw()
// 		})
// 	textView.SetBorder(true).SetTitle("Stdin")
//
// 	output, err := aws.ExecuteCommand(taskArn, containerName, clusterName)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to open pty")
// 		return nil, nil
// 	}
//
// 	v := struct {
// 		SessionID  string `json:"SessionId"`
// 		StreamURL  string `json:"StreamUrl"`
// 		TokenValue string `json:"TokenValue"`
// 	}{
// 		SessionID:  *output.Session.SessionId,
// 		StreamURL:  *output.Session.StreamUrl,
// 		TokenValue: *output.Session.TokenValue,
// 	}
// 	json, err := json.Marshal(v)
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to marshal input parameters")
// 		return nil, nil
// 	}
//
// 	cmd := exec.Command("session-manager-plugin", string(json), "eu-north-1", "StartSession")
// 	p, err := pty.Start(cmd)
//
// 	if err != nil {
// 		log.Error().Err(err).Msg("Failed to open pty")
// 		return nil, nil
// 	}
//
// 	inputCapture := func(keyEvent *tcell.EventKey) *tcell.EventKey {
// 		log.Info().Msgf("Key pressed: %s", keyEvent.Name())
// 		if keyEvent.Key() == tcell.KeyBackspace2 || keyEvent.Key() == tcell.KeyBackspace {
// 			p.Write([]byte{'\b'})
// 		} else if keyEvent.Key() == tcell.KeyUp {
// 			p.Write([]byte{'\x27'})
// 		} else {
// 			_, _ = p.WriteString(string(keyEvent.Rune()))
// 		}
// 		return keyEvent
// 	}
//
// 	buffer := [][]rune{}
// 	textView.SetInputCapture(inputCapture)
// 	reader := bufio.NewReader(p)
//
// 	// Goroutine that reads from pty
// 	go func() {
// 		line := []rune{}
// 		buffer = append(buffer, line)
//
// 		for {
// 			// Current line we are editing
// 			line = buffer[len(buffer)-1]
// 			//logger.Info().Msgf("Line: %s:%d", string(line), len(line))
//
// 			r, err := readNextRune(reader)
// 			if err != nil {
// 				break
// 			}
//
// 			if r == '\b' {
// 				line = removeLast(line)
// 			} else if r == '\u001B' {
// 				line = handleEscapeSequence(r, buffer, line, reader)
// 			} else {
// 				line = append(line, r)
// 			}
//
// 			buffer[len(buffer)-1] = line
//
// 			if r == '\n' {
// 				if len(buffer) > MaxBufferSize { // If the buffer is at capacity...
// 					buffer = buffer[1:] // ...pop the first line in the buffer
// 				}
//
// 				line = []rune{}
// 				buffer = append(buffer, line)
// 			}
// 		}
// 	}()
//
// 	// Goroutine that writes to textView
// 	go func() {
// 		w := tview.ANSIWriter(textView)
//
// 		for {
// 			time.Sleep(100 * time.Millisecond)
//
// 			textView.Clear()
//
// 			for _, line := range buffer {
// 				w.Write([]byte(string(line)))
// 			}
// 		}
//
// 	}()
//
// 	return textView, p
// }
//
// func removeLast(s []rune) []rune {
// 	length := len(s)
// 	ret := make([]rune, 0)
// 	if length > 0 {
// 		ret = append(ret, s[:length-1]...)
// 	}
// 	return ret
// }
//
// func readNextRune(reader *bufio.Reader) (rune, error) {
// 	r, _, err := reader.ReadRune()
//
// 	log.Info().Msgf("Read: %s", string(r))
//
// 	if err != nil {
// 		if err == io.EOF {
// 			return r, err
// 		}
// 		os.Exit(0)
// 	}
//
// 	return r, nil
// }
//
// func handleEscapeSequence(s rune, buffer [][]rune, line []rune, reader *bufio.Reader) []rune {
// 	log.Info().Msg("ESCAPE")
// 	r, _ := readNextRune(reader)
//
// 	if r == '[' {
// 		n, _ := readNextRune(reader)
// 		if n == 'K' {
// 			// backspace
// 			line = removeLast(line)
// 		} else if n == '2' {
// 			// clear
// 			d, _ := readNextRune(reader)
// 			if d == 'J' {
// 				log.Info().Msg("Clear screen")
// 				// clear screen
// 				buffer = [][]rune{}
// 				line = []rune{}
// 			} else if d == 'K' {
// 				// clear line
// 			}
// 		} else if n == '3' {
// 			d, _ := readNextRune(reader)
// 			if d == 'J' {
// 				// erased saved lines
// 			}
// 		} else if n == 'H' {
// 			// new line
// 		} else {
// 			// other
// 			line = append(line, s, r, n)
// 		}
// 	} else {
// 		// possible?
// 		line = append(line, s, r)
// 	}
//
// 	return line
// }
