package application

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bsek/s9k/internal/s9k/aws"
	"github.com/bsek/s9k/internal/s9k/data"
	"github.com/bsek/s9k/internal/s9k/ecs"
	"github.com/bsek/s9k/internal/s9k/lambda"
	"github.com/bsek/s9k/internal/s9k/ui"
	"github.com/bsek/s9k/internal/s9k/utils"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/rs/zerolog/log"
)

type Application struct {
	tviewApp          *tview.Application
	flex              *tview.Flex
	pages             *tview.Pages
	tablePagesMap     map[int32]ui.TablePage
	commandFooterBar  *tview.TextView
	accountData       *data.AccountData
	progressFooterBar *tview.TextView
}

func NewApplication() *Application {
	accountData := loadData()

	return &Application{
		tviewApp:          tview.NewApplication(),
		flex:              tview.NewFlex(),
		pages:             tview.NewPages(),
		tablePagesMap:     make(map[int32]ui.TablePage, 0),
		commandFooterBar:  tview.NewTextView(),
		progressFooterBar: tview.NewTextView(),
		accountData:       accountData,
	}
}

// Entrypoint for the application
func Entrypoint() {
	log.Info().Msg("Loading information about your AWS ECS clusters and Lambda functions...")

	app := NewApplication()
	app.buildApplicationUI()

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func loadData() *data.AccountData {
	clusters, err := aws.ListECSClusters()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to read ecs clusters")
	}
	return data.NewAccountData(clusters[0])
}

func (a Application) Run() error {
	return a.tviewApp.Run()
}

// Select a table page with a single key shortcut
func (a *Application) setTablePageByKey(key int32) bool {
	if page, found := a.tablePagesMap[key]; found {
		a.showPage(page, key)
		return true
	}
	return false
}

// Re-render currently displayed table page
func (a *Application) renderCurrentTablePage() {
	a.getCurrentlyDisplayedTablePage().Render(a.accountData)
}

// Show the selected cluster detail page
func (a *Application) showPage(selectedPage ui.TablePage, key int32) {
	_, frontPageView := a.pages.GetFrontPage()

	a.commandFooterBar.Highlight(string(key)).ScrollToHighlight()
	selectedPage.Render(a.accountData)
	a.pages.SwitchToPage(selectedPage.Name())
	a.showRefreshTime(*a.accountData.Cluster.ClusterName, a.accountData.Refreshed)

	// If the page about to be hidden has focus, switch focus to the new page
	if frontPageView != nil && frontPageView.HasFocus() {
		a.tviewApp.SetFocus(selectedPage.Table())
	}
}

// Handle a user input event
func (a *Application) handleAppInput(event *tcell.EventKey) *tcell.EventKey {
	if event.Key() == tcell.KeyRune {
		key := event.Rune()

		if a.setTablePageByKey(key) {
			return event
		}

		// update data
		if key == 'r' || key == 'R' {
			a.accountData.Refresh()
			a.renderCurrentTablePage()
		}

		if key == 'q' || key == 'Q' {
			a.tviewApp.Stop()
		}
	}

	return event
}

func (a *Application) getCurrentlyDisplayedTablePage() ui.TablePage {
	name, _ := a.pages.GetFrontPage()

	var p ui.TablePage
	for _, page := range a.tablePagesMap {
		if page.Name() == name {
			p = page
		}
	}

	return p
}

func (a *Application) showRefreshTime(what string, when time.Time) {
	a.progressFooterBar.Clear()
	fmt.Fprintf(a.progressFooterBar, "%s refreshed at %s", what, utils.FormatLocalTime(when))
}

// Build the UI elements and configures the application
func (a *Application) buildApplicationUI() {
	a.tviewApp = tview.NewApplication()
	a.pages = tview.NewPages()
	a.flex = tview.NewFlex()

	// Build the cluster detail pages and add their view shortcuts
	a.tablePagesMap['1'] = ecs.NewServicesPage(a.tviewApp, a.flex)
	a.tablePagesMap['2'] = lambda.NewLambdasPage(a.tviewApp, a.flex)

	for _, page := range a.tablePagesMap {
		page.Table().SetBorderColor(tcell.ColorGoldenrod)
		a.pages.AddPage(page.Name(), page.Table(), true, false)
	}

	a.commandFooterBar = a.buildCommandFooterBar()

	a.progressFooterBar = tview.NewTextView().
		SetDynamicColors(true).
		SetWrap(false).
		SetTextAlign(ui.R)
	a.progressFooterBar.SetBorderPadding(0, 0, 1, 2)

	footer := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.commandFooterBar, 0, 6, false).
		AddItem(a.progressFooterBar, 0, 4, false)

	a.flex.
		SetDirection(tview.FlexRow).
		AddItem(a.pages, 0, 1, false).
		AddItem(footer, 1, 1, false)

	a.tviewApp.SetRoot(a.flex, true).SetInputCapture(a.handleAppInput).EnableMouse(true)
	a.showPage(a.tablePagesMap['1'], 1)

	a.tviewApp.SetFocus(a.pages)
}

// Build the command bar with detail page shortcuts that appears in the footer
func (a *Application) buildCommandFooterBar() *tview.TextView {

	footerBar := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	pageCommands := make([]string, 0)
	for key, page := range a.tablePagesMap {
		pageCommands = append(pageCommands, fmt.Sprintf(`[bold]%c ["%c"][darkcyan]%s[white][""]`, key, key, page.Name()))
	}
	sort.Strings(pageCommands)

	footerPageText := strings.Join(pageCommands, " ")
	footerPageText = fmt.Sprintf(`%s %c [white::b]R[darkcyan::-] Refresh data`, footerPageText, tcell.RuneVLine)
	footerPageText = fmt.Sprintf(`%s [white::b]Q[darkcyan::-] Quit application`, footerPageText)

	fmt.Fprint(footerBar, footerPageText)

	return footerBar
}
