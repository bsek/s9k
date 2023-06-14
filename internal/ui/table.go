package ui

import (
	"strconv"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// AddTableData adds data to the table. Lets you set the headers, texts, alignment, color, and if cells are selectable. Selectable is not per cell,
// but a global value for the whole table. Header row is not selectable
func AddTableData(table *tview.Table, headers []string, data [][]string, alignment []int, expansions []int, color tcell.Color, selectable bool) {

	// Set header text, not selectable
	for col, text := range headers {
		cell := tview.NewTableCell(text).
			SetAlign(alignment[col]).
			SetExpansion(expansions[col]).
			SetTextColor(tview.Styles.SecondaryTextColor).
			SetSelectable(false)
		table.SetCell(0, col, cell)
	}

	// Set data text
	for row, line := range data {
		for col, text := range line {
			cell := tview.NewTableCell(text).
				SetAlign(alignment[col]).
				SetExpansion(expansions[col]).
				SetTextColor(color).
				SetSelectable(selectable)
			table.SetCell(row+1, col, cell)
		}
	}
}

// PrependRowNumColumn prepends every slice in data with a row-number value
func PrependRowNumColumn(data [][]string) [][]string {
	for i := 0; i < len(data); i++ {
		row := strconv.Itoa(i + 1)
		data[i] = append([]string{row}, data[i]...)
	}
	return data
}
