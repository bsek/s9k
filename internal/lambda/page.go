package lambda

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ui"
	"github.com/bsek/s9k/internal/utils"
)

var _ ui.ContentPage = (*LambdasPage)(nil)

type LambdasPage struct {
	table *tview.Table
	name  string
}

const name = "functions"

func NewLambdasPage() *LambdasPage {
	lambdasTable := tview.NewTable()
	lambdasTable.
		SetBorders(false).
		SetBorder(true).
		SetTitle(" λ Lambda functions ")

	lambdasTable.SetSelectable(true, false)

	lambdasTable.SetSelectedFunc(func(row, _ int) {
		ref := lambdasTable.GetCell(row, 1).Reference.(data.Function)

		createActionForm(*ref.FunctionName, ref.Architectures[0])
	})

	return &LambdasPage{
		name:  name,
		table: lambdasTable,
	}
}

func (l *LambdasPage) Render(accountData *data.AccountData) {
	lambdaData := accountData.Functions

	if len(lambdaData) == 0 {
		return
	}

	data := lo.Map(lambdaData, func(function data.Function, _ int) []string {
		modified, _ := time.Parse("2006-01-02T15:04:05.9Z0700", *function.LastModified)
		return []string{
			*function.FunctionName,
			string(function.Runtime),
			string(function.PackageType),
			function.Tags["LastDeployed"],
			utils.FormatBytes(function.CodeSize),
			utils.FormatBytes(((int64)(*function.MemorySize) * 1000000)),
			fmt.Sprintf("%s s", utils.I32ToString(*function.Timeout)),
			string(function.Architectures[0]),
			utils.FormatLocalDateTime(modified),
		}
	})

	data = ui.PrependRowNumColumn(data)

	alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignRight, tview.AlignRight, tview.AlignRight, tview.AlignLeft, tview.AlignRight}
	expansions := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	headers := []string{"#", "Name ▾", "Runtime", "Package type", "Version", "Code size", "Memory size", "Timeout", "Architecture", "Last modified"}

	ui.AddTableData(l.table, headers, data, alignment, expansions, tview.Styles.PrimaryTextColor, true)

	// set reference to service
	for i := 1; i < len(lambdaData)+1; i++ {
		cell := l.table.GetCell(i, 1)
		cell.SetReference(lambdaData[i-1])
	}
}

func (l *LambdasPage) Name() string {
	return l.name
}

func (l *LambdasPage) ContextView() tview.Primitive {
	tw := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWrap(false)

	bw := tw.BatchWriter()
	defer bw.Close()

	fmt.Fprintln(bw, "[white::b]Enter [darkcyan::-]Select")

	return tw
}

func (l *LambdasPage) SetFocus(app *tview.Application) {
	app.SetFocus(l.table)
}

func (l *LambdasPage) View() tview.Primitive {
	return l.table
}

func (l *LambdasPage) Close() {
}

func (l *LambdasPage) IsPersistent() bool {
	return true
}
