package ui

import (
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/rivo/tview"
)

type TablePage interface {
	Render(accountData *data.AccountData)
	Name() string
	Table() *tview.Table
}
