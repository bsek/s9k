package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/bsek/s9k/internal/utils"
	"github.com/rivo/tview"
)

const logo = ` ▄▄▄▄▄▄▄ ▄▄▄▄▄▄▄ ▄▄▄   ▄ 
█       █  ▄    █   █ █ █
█  ▄▄▄▄▄█ █ █   █   █▄█ █
█ █▄▄▄▄▄█ █▄█   █      ▄█
█▄▄▄▄▄  █▄▄▄    █     █▄ 
 ▄▄▄▄▄█ █   █   █    ▄  █
█▄▄▄▄▄▄▄█   █▄▄▄█▄▄▄█ █▄█`

func NewHeader(accountId, clusterName string) *Header {
	layout := tview.NewFlex().
		SetDirection(tview.FlexColumn)

	return &Header{
		Logo:        logo,
		Context:     nil,
		RefreshTime: time.Time{},
		Layout:      layout,
		AccountId:   accountId,
		ClusterName: clusterName,
	}
}

func (h *Header) SetContextView(view tview.Primitive) {
	h.Context = view
}

func (h *Header) UpdateRefreshTime(when time.Time) {
	h.RefreshTime = when
}

func (h *Header) Render(pagesMap map[int32]ContentPage) {
	h.Layout.Clear()

	logoView := tview.NewTextView().
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	fmt.Fprint(logoView, h.Logo)

	// Build the command bar with detail page shortcuts that appears in the header
	shortcuts := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	pageCommands := make([]string, 0)
	for key, page := range pagesMap {
		pageCommands = append(pageCommands, fmt.Sprintf(`[white::b]%c ["%c"][darkcyan::]View %s[""]`, key, key, page.Name()))
	}
	sort.Strings(pageCommands)

	fmt.Fprintln(shortcuts, strings.Join(pageCommands, "\n"))

	accountInfo := tview.NewTextView().SetDynamicColors(true).SetRegions(true).SetWrap(false)

	aw := accountInfo.BatchWriter()

	fmt.Fprintln(aw, fmt.Sprintf("[darkolivegreen::b]Account id: [-::]%s", h.AccountId))
	fmt.Fprintln(aw, fmt.Sprintf("[darkolivegreen::b]Cluster name: [-::]%s", h.ClusterName))
	fmt.Fprintln(aw, "")
	fmt.Fprintln(aw, "[white::b]u[darkcyan::-] Update data")
	fmt.Fprintln(aw, "[white::b]q[darkcyan::-] Quit application")
	fmt.Fprintln(aw, "")
	fmt.Fprintf(aw, "[darkolivegreen::b]Refreshed at %s", utils.FormatLocalTime(h.RefreshTime))

	aw.Close()

	h.Layout.AddItem(accountInfo, 0, 1, false)
	h.Layout.AddItem(shortcuts, 0, 1, false)

	if h.Context != nil {
		h.Layout.AddItem(h.Context, 0, 2, false)
	}

	h.Layout.AddItem(logoView, 0, 1, false)
	h.Layout.SetBorder(false)
}
