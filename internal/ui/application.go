package ui

import (
	"strings"

	"github.com/bsek/s9k/internal/data"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

var App *Application

func NewApplication(accountData *data.AccountData) *Application {
	return &Application{
		TviewApp:    tview.NewApplication(),
		Layout:      tview.NewFlex(),
		Content:     tview.NewPages(),
		ContentMap:  make(map[string]ContentPage, 0),
		HeaderBar:   NewHeader(accountData.AccountId, accountData.ClusterName),
		AccountData: accountData,
	}
}

func (a *Application) Run() error {
	return a.TviewApp.Run()
}

// Select a content page with a single key shortcut
func (a *Application) setContentByKey(key string) bool {
	currentPage := a.getCurrentDisplayedContentPage()

	if page, found := a.ContentMap[key]; found {
		// If same page, do nothing
		if page.Name() != currentPage.Name() {
			if !currentPage.IsPersistent() {
				a.RemoveContent(currentPage)
			}
			a.ShowPage(page)
		}
		return true

	}
	return false
}

func (a *Application) ShowPage(selectedPage ContentPage) {
	selectedPage.Render(a.AccountData)
	a.Content.SwitchToPage(selectedPage.Name())

	// Update header view
	a.HeaderBar.UpdateRefreshTime(a.AccountData.Refreshed)
	a.HeaderBar.SetContextView(selectedPage.ContextView())
	a.HeaderBar.Render(a.ContentMap)

	selectedPage.SetFocus(a.TviewApp)
}

func (a *Application) updateData() {
	a.AccountData.Refresh()
	a.getCurrentDisplayedContentPage().Render(a.AccountData)
	a.HeaderBar.UpdateRefreshTime(a.AccountData.Refreshed)
	a.HeaderBar.Render(a.ContentMap)
}

// Handle a user input event
func (a *Application) handleAppInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		key := event.Rune()

		// change listpage by shortcut
		if a.setContentByKey(string(key)) {
			return event
		}

		// update data
		if key == 'u' || key == 'U' {
			go App.TviewApp.QueueUpdateDraw(a.updateData)
			return event
		}

		// quit application
		if key == 'q' || key == 'Q' {
			a.TviewApp.Stop()
			return event
		}
	}

	// pass handling to primitive which has focus
	return event
}

func (a *Application) getCurrentDisplayedContentPage() ContentPage {
	name, _ := a.Content.GetFrontPage()

	log.Debug().Msgf("Looking for page with name %s", name)

	var p ContentPage
	for _, page := range a.ContentMap {
		if page.Name() == name {
			p = page
		}
	}

	log.Debug().Msgf("Found page %v", p)

	return p
}

func (a *Application) RegisterContent(page ContentPage) {
	log.Debug().Msgf("Registering page %s", page.Name())
	a.Content.AddPage(page.Name(), page.View(), true, false)
	a.ContentMap[strings.ToLower(page.Name()[0:1])] = page
}

func (a *Application) RemoveContent(page ContentPage) {
	log.Debug().Msgf("Removing page %s", page.Name())
	if !page.IsPersistent() {
		page.Close()
		delete(a.ContentMap, strings.ToLower(page.Name()[0:1]))
		a.Content.RemovePage(page.Name())
	}
}

// Build the UI elements and configures the application
func (a *Application) BuildApplicationUI() {
	a.TviewApp = tview.NewApplication()

	a.Layout.
		SetDirection(tview.FlexRow).
		AddItem(a.HeaderBar.Layout, 8, 0, false).
		AddItem(a.Content, 0, 1, true)

	a.TviewApp.
		SetRoot(a.Layout, true).
		SetInputCapture(a.handleAppInput).
		EnableMouse(true)
}
