package ui

import (
	"github.com/rivo/tview"
)

// CreateModalPage creates a modal window by creating a pages view and add a flex view with the specified width and
// height. The page is named the same as the name parameter and if the background primitive is set, it is added as a background
// to the modal
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

// CreateMessageBox creates a modal message box using tview's native modal element. Only one button is added (close) and
// no callback is possible. The modal is added to the provided pages view and removed when the button is pressed
func CreateMessageBox(content string) {
	const MESSAGE_BOX = "message_box"

	modal := tview.NewModal().
		SetText(content).
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(_ int, _ string) {
			App.Content.RemovePage(MESSAGE_BOX)
		})

	RemoveFrontPage(App.Content)
	App.Content.AddAndSwitchToPage(MESSAGE_BOX, modal, true)
}

// CreateConfirmBox creates a modal message box using tview's native modal element. Two buttons are added (yes and no)
// for which the yesFunc and noFunc callback functiosn will be called when pressed. The modal view is returned, so it
// is up to the callee to add and remove it from the view
func CreateConfirmBox(content string, yesFunc, noFunc func()) {
	const CONFIRM_BOX = "confirm_box"
	modal := tview.NewModal().
		SetText(content).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				App.Content.RemovePage(CONFIRM_BOX)
				yesFunc()
			}
			if buttonLabel == "No" {
				App.Content.RemovePage(CONFIRM_BOX)
				noFunc()
			}
		})

	App.Content.AddAndSwitchToPage(CONFIRM_BOX, modal, true)
}
