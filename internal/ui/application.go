package ui

import (
	"github.com/bsek/s9k/internal/data"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var App *Application

func NewApplication(accountData *data.AccountData) *Application {
	return &Application{
		TviewApp:    tview.NewApplication(),
		Layout:      tview.NewFlex(),
		Content:     tview.NewPages(),
		ContentMap:  make(map[int32]ContentPage, 0),
		HeaderBar:   NewHeader(accountData.AccountId, accountData.ClusterName),
		AccountData: accountData,
	}
}

func (a *Application) Run() error {
	return a.TviewApp.Run()
}

// Select a table page with a single key shortcut
func (a *Application) setContentByKey(key int32) bool {
	if page, found := a.ContentMap[key]; found {
		a.ShowPage(page)
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
	a.getCurrentlyDisplayedTablePage().Render(a.AccountData)
	a.HeaderBar.UpdateRefreshTime(a.AccountData.Refreshed)
	a.HeaderBar.Render(a.ContentMap)
}

// Handle a user input event
func (a *Application) handleAppInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		key := event.Rune()

		// change listpage by shortcut
		if a.setContentByKey(key) {
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

func (a *Application) getCurrentlyDisplayedTablePage() ContentPage {
	name, _ := a.Content.GetFrontPage()

	var p ContentPage
	for _, page := range a.ContentMap {
		if page.Name() == name {
			p = page
		}
	}

	return p
}

func (a *Application) RegisterContent(page ContentPage) {
	a.Content.AddPage(page.Name(), page.View(), true, false)
	a.ContentMap[int32(page.Shortcut())] = page
}

func (a *Application) RemoveContent(page ContentPage) {
	key := int32(page.Shortcut())

	delete(a.ContentMap, key)
	a.Content.RemovePage(page.Name())

	if key > 0 {
		a.setContentByKey(key - 1)
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

	if len(a.ContentMap) > 0 {
		a.setContentByKey('1')
	}
}
