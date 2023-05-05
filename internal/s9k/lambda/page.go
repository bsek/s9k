package lambda

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/samber/lo"
)

type LambdasPage struct {
	name      string
	tableInfo *ui.TableInfo
}

const name = "Lambda functions"

func NewLambdasPage(app *tview.Application, flex *tview.Flex) *LambdasPage {
	lambdasTable := tview.NewTable()
	lambdasTable.
		SetBorders(true).
		SetBorder(true).
		SetTitle(" λ Lambda functions ")

	lambdasTable.SetSelectable(true, false)

	lambdasTable.SetSelectedFunc(func(row, column int) {
		text := lambdasTable.GetCell(row, 1).Text
		inputCaptureFunction := app.GetInputCapture()

		var pages *tview.Pages

		closeFunction := func() {
			app.SetInputCapture(inputCaptureFunction)
			app.SetRoot(flex, true)
			app.SetFocus(flex.GetItem(0))
		}

		form := ui.CreateActionForm(text,
			func() {
				deploy(text, text, pages, closeFunction)
			},
			func() {
				restart(text, pages, closeFunction)
			},
			func() {
				closeFunction()
			})

		form.SetBorder(true).SetTitle("What do you want to do?").SetTitleAlign(tview.AlignLeft)

		pages = ui.CreateModalPage(form, flex, 60, 5, "ask")

		app.SetInputCapture(nil)
		app.SetRoot(pages, true)
	})

	tableInfo := &ui.TableInfo{
		Table:      lambdasTable,
		Alignment:  []int{ui.L, ui.L, ui.L, ui.L, ui.R, ui.R, ui.L, ui.R},
		Expansions: []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		Selectable: true,
	}

	ui.AddTableConfigData(tableInfo, 0, [][]string{
		{"#", "Name ▾", "Runtime", "Code size", "Memory size", "Timeout", "Architecture", "Last modified"},
	}, tcell.ColorYellow)

	return &LambdasPage{
		name:      name,
		tableInfo: tableInfo,
	}
}

func (l *LambdasPage) Render(accountData *data.AccountData) {
	ui.TruncTableRows(l.tableInfo.Table, 1)

	lambdaData := accountData.Functions

	if len(lambdaData) == 0 {
		return
	}

	data := lo.Map(lambdaData, func(function types.FunctionConfiguration, index int) []string {
		modified, _ := time.Parse(time.RFC3339, *function.LastModified)
		return []string{
			*function.FunctionName,
			string(function.Runtime),
			utils.FormatBytes(&function.CodeSize),
			fmt.Sprintf("%s b", utils.I32ToString(*function.MemorySize)),
			fmt.Sprintf("%s s", utils.I32ToString(*function.Timeout)),
			string(function.Architectures[0]),
			utils.FormatLocalDateTime(modified),
		}
	})

	data = ui.PrependRowNumColumn(data)

	ui.AddTableConfigData(l.tableInfo, 1, data, tcell.ColorWhite)

}

func (l *LambdasPage) Name() string {
	return l.name
}

func (l *LambdasPage) Table() *tview.Table {
	return l.tableInfo.Table
}
