package ui

import "github.com/rivo/tview"

func RemoveFrontPage(pages *tview.Pages) {
	name, _ := pages.GetFrontPage()
	pages.RemovePage(name)
}
