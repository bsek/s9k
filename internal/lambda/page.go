package lambda

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ui"
	"github.com/bsek/s9k/internal/utils"
	"github.com/rivo/tview"
	"github.com/samber/lo"
)

var _ ui.ContentPage = (*LambdasPage)(nil)

type LambdasPage struct {
	name   string
	table  *tview.Table
	header *tview.TextView
}

// Shortcut implements ui.ContentPage.
func (*LambdasPage) Shortcut() rune {
	return '2'
}

// SetFocus implements ui.ContextPage.
func (l *LambdasPage) SetFocus(app *tview.Application) {
	app.SetFocus(l.table)
}

// View implements ui.ContextPage.
func (l *LambdasPage) View() tview.Primitive {
	return l.table
}

const name = "Lambda functions"

func NewLambdasPage() *LambdasPage {
	lambdasTable := tview.NewTable()
	lambdasTable.
		SetBorders(false).
		SetBorder(true).
		SetTitle(" λ Lambda functions ")

	lambdasTable.SetSelectable(true, false)

	lambdasTable.SetSelectedFunc(func(row, column int) {
		ref := lambdasTable.GetCell(row, 1).Reference.(types.FunctionConfiguration)

		createActionForm(*ref.FunctionName, ref.Architectures[0])
	})

	return &LambdasPage{
		name:   name,
		table:  lambdasTable,
		header: tview.NewTextView(),
	}
}

func (l *LambdasPage) Render(accountData *data.AccountData) {
	lambdaData := accountData.Functions

	if len(lambdaData) == 0 {
		return
	}

	data := lo.Map(lambdaData, func(function types.FunctionConfiguration, index int) []string {
		modified, _ := time.Parse("2006-01-02T15:04:05.9Z0700", *function.LastModified)
		return []string{
			*function.FunctionName,
			string(function.Runtime),
			string(function.PackageType),
			utils.FormatBytes(function.CodeSize),
			utils.FormatBytes(((int64)(*function.MemorySize) * 1000000)),
			fmt.Sprintf("%s s", utils.I32ToString(*function.Timeout)),
			string(function.Architectures[0]),
			utils.FormatLocalDateTime(modified),
		}
	})

	data = ui.PrependRowNumColumn(data)

	alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignRight, tview.AlignRight, tview.AlignRight, tview.AlignLeft, tview.AlignRight}
	expansions := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	headers := []string{"#", "Name ▾", "Runtime", "Package type", "Code size", "Memory size", "Timeout", "Architecture", "Last modified"}

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
	return l.header
}
