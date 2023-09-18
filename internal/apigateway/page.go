package apigateway

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/samber/lo"

	"github.com/bsek/s9k/internal/aws"
	"github.com/bsek/s9k/internal/data"
	"github.com/bsek/s9k/internal/ui"
)

var _ ui.ContentPage = (*ApiGatewayPage)(nil)

type ApiGatewayPage struct {
	table *tview.Table
}

func NewApiGatewayPage() *ApiGatewayPage {
	page := &ApiGatewayPage{
		table: tview.NewTable(),
	}

	page.table.
		SetBorder(true).
		SetTitle(" 📋 Api Gateway apis ")

	page.table.SetSelectable(true, false)

	page.table.SetSelectedFunc(func(row, _ int) {
		cell := page.table.GetCell(row, 1)
		api := cell.Reference.(aws.ApiGateway)

		createActionForm(api.Name)
	})

	return page
}

func (a *ApiGatewayPage) Close() {
}

func (a *ApiGatewayPage) ContextView() tview.Primitive {
	tw := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWrap(false)

	bw := tw.BatchWriter()
	defer bw.Close()

	fmt.Fprintln(bw, "[white::b]Enter [darkcyan::-]Select")

	return tw
}

func (a *ApiGatewayPage) IsPersistent() bool {
	return true
}

func (a *ApiGatewayPage) Name() string {
	return "Apigateway"
}

func (a *ApiGatewayPage) View() tview.Primitive {
	return a.table
}

func (a *ApiGatewayPage) Render(accountData *data.AccountData) {
	apis := accountData.Apis

	if len(apis) == 0 {
		return
	}

	data := lo.Map(accountData.Apis, func(api aws.ApiGateway, index int) []string {
		return []string{
			api.Name,
			api.ApiId,
			api.DomainName,
			api.Type.String(),
			api.Description,
			api.CreatedDate.Format("2006-01-02T15:04:05.9Z0700"),
		}
	})

	data = ui.PrependRowNumColumn(data)

	alignment := []int{tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignLeft, tview.AlignRight, tview.AlignRight, tview.AlignRight, tview.AlignLeft, tview.AlignRight}
	expansions := []int{1, 1, 1, 1, 1, 2, 1}
	headers := []string{"#", "Name ▾", "Id", "Domain name", "Protocol", "Description", "Created"}

	ui.AddTableData(a.table, headers, data, alignment, expansions, tview.Styles.PrimaryTextColor, true)

	// set reference to api
	for i := 1; i < len(apis)+1; i++ {
		cell := a.table.GetCell(i, 1)
		cell.SetReference(apis[i-1])
	}
}

func (a *ApiGatewayPage) SetFocus(app *tview.Application) {
	app.SetFocus(a.table)
}
