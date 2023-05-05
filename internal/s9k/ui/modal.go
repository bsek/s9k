package ui

import (
	"github.com/rivo/tview"
)

func CreateModalPage(content tview.Primitive, background *tview.Flex, width, height int, name string) *tview.Pages {
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

	pages.
		//AddPage("background", background, true, true).
		AddPage(name, modal(content, width, height), true, true)

	return pages
}
