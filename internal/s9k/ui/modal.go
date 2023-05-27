package ui

import (
	"github.com/rivo/tview"
)

const MESSAGES_BOX = "message_box"

func CreateModalPage(content tview.Primitive, background tview.Primitive, width, height int, name string) *tview.Pages {
	modal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, true).
			AddItem(nil, 0, 1, false)
	}

	pages := tview.NewPages()

	if background != nil {
		pages.AddPage("background", background, true, true)
	}

	pages.AddPage(name, modal(content, width, height), true, true)

	return pages
}

func CreateMessageBox(content string, pages *tview.Pages) {
	modal := tview.NewModal().
		SetText(content).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(_ int, _ string) {
			pages.RemovePage("MESSAGE_BOX")
		})
	pages.AddAndSwitchToPage("MESSAGE_BOX", modal, true)
}

func CreateConfirmBox(content string, yesFunc, noFunc func()) *tview.Modal {
	modal := tview.NewModal().
		SetText(content).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				yesFunc()
			}
			if buttonLabel == "No" {
				noFunc()
			}
		})

	return modal
}
