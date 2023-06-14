package ui

import (
	"time"

	"github.com/bsek/s9k/internal/data"
	"github.com/rivo/tview"
)

type Application struct {
	TviewApp    *tview.Application
	Layout      *tview.Flex
	Content     *tview.Pages
	ContentMap  map[int32]ContentPage
	HeaderBar   *Header
	AccountData *data.AccountData
}

type Header struct {
	Logo        string
	Context     tview.Primitive
	RefreshTime time.Time
	Layout      *tview.Flex
	AccountId   string
	ClusterName string
}

type ContentPage interface {
	Render(accountData *data.AccountData)
	Name() string
	View() tview.Primitive
	ContextView() tview.Primitive
	Shortcut() rune
	SetFocus(app *tview.Application)
}
