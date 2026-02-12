package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
)

// MarketplacePanel represents the plugin marketplace UI
type MarketplacePanel struct {
	*tview.Flex
	table           *tview.Table
	details         *tview.TextView
	plugins         []plugins.RegistryPlugin
	filteredPlugins []plugins.RegistryPlugin
	manager         *plugins.PluginManager
	client          *plugins.RegistryClient
	ui              *UIOrchestrator
	searchField     *tview.InputField
}

// NewMarketplacePanel creates a new MarketplacePanel
func NewMarketplacePanel(ui *UIOrchestrator) *MarketplacePanel {
	m := &MarketplacePanel{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		table:   tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		details: tview.NewTextView().SetDynamicColors(true).SetWrap(true),
		manager: ui.PluginManager,
		client:  plugins.NewRegistryClient(config.C.Plugins.RegistryURL),
		ui:      ui,
	}

	m.Flex.SetBackgroundColor(ui.Colors.Background)

	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.details.SetBorder(true).SetTitle(" Plugin Details ")
	m.details.SetBorderColor(ui.Colors.Border)
	m.details.SetBackgroundColor(ui.Colors.Background)

	m.searchField = tview.NewInputField().
		SetLabel(" Search Plugins: ").
		SetLabelColor(ui.Colors.LabelColor).
		SetFieldBackgroundColor(ui.Colors.Selection).
		SetFieldTextColor(ui.Colors.Foreground)
	m.searchField.SetBackgroundColor(ui.Colors.Background)

	m.searchField.SetChangedFunc(func(text string) {
		m.filterPlugins(text)
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			m.ui.App.SetFocus(m.table)
			return nil
		}
		return event
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := m.table.GetSelection()
		if event.Key() == tcell.KeyUp && row == 1 {
			m.ui.App.SetFocus(m.searchField)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			m.ui.App.SetFocus(m.searchField)
			return nil
		}
		return event
	})

	m.table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(m.filteredPlugins) {
			m.updateDetails(m.filteredPlugins[row-1])
		}
	})

	m.table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row-1 < len(m.filteredPlugins) {
			m.handlePluginAction(m.filteredPlugins[row-1])
		}
	})

	m.AddItem(m.searchField, 1, 0, true)

	innerFlex := tview.NewFlex().
		AddItem(m.table, 0, 1, true).
		AddItem(m.details, 0, 1, false)
	innerFlex.SetBackgroundColor(ui.Colors.Background)

	m.AddItem(innerFlex, 0, 1, false)

	m.SetBorder(true).SetTitle(" Plugin Marketplace (petitorium.dev) ")
	m.SetBorderColor(ui.Colors.Border)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("marketplace")
			return nil
		}
		// Only close with 'q' if NOT in search field
		if event.Rune() == 'q' && m.ui.App.GetFocus() != m.searchField {
			m.ui.Pages.RemovePage("marketplace")
			return nil
		}
		return event
	})

	return m
}

func (m *MarketplacePanel) filterPlugins(query string) {
	m.table.Clear()
	m.filteredPlugins = []plugins.RegistryPlugin{}
	query = strings.ToLower(query)

	// Set headers
	headers := []string{"Name", "Version", "Official", "Status"}
	for i, h := range headers {
		m.table.SetCell(0, i, tview.NewTableCell(" "+h+" ").
			SetTextColor(m.ui.Colors.Title).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignCenter))
	}
	m.table.GetCell(0, 0).SetAlign(tview.AlignLeft)

	row := 1
	for _, p := range m.plugins {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Description), query) {
			m.filteredPlugins = append(m.filteredPlugins, p)
			statusText, statusColor := m.getPluginStatusInfo(p)
			official := ""
			officialColor := m.ui.Colors.Foreground
			if p.Official {
				official = "✔"
				officialColor = tcell.ColorGreen
			}

			m.table.SetCell(row, 0, tview.NewTableCell(p.Name).SetTextColor(m.ui.Colors.Foreground).SetExpansion(2))
			m.table.SetCell(row, 1, tview.NewTableCell(p.Version).SetTextColor(m.ui.Colors.Foreground).SetAlign(tview.AlignCenter))
			m.table.SetCell(row, 2, tview.NewTableCell(official).SetAlign(tview.AlignCenter).SetTextColor(officialColor))
			m.table.SetCell(row, 3, tview.NewTableCell(statusText).SetAlign(tview.AlignCenter).SetTextColor(statusColor))
			row++
		}
	}

	if len(m.filteredPlugins) > 0 {
		m.updateDetails(m.filteredPlugins[0])
	}
}

func (m *MarketplacePanel) getPluginStatusInfo(p plugins.RegistryPlugin) (string, tcell.Color) {
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		if info.Version != p.Version {
			return "Update", tcell.ColorYellow
		}
		return "Installed", tcell.ColorGreen
	}
	return "Available", m.ui.Colors.Foreground
}

func (m *MarketplacePanel) handlePluginAction(p plugins.RegistryPlugin) {
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		if info.Version == p.Version {
			// Already up to date, maybe uninstall?
			// For now, just return
			return
		}
	}

	// Install/Update
	_ = showProgressModal(m.ui.Pages, " Installing Plugin ", fmt.Sprintf("Downloading %s...", p.Name), m.ui.Colors.Background)

	go func() {
		err := m.manager.InstallPlugin(p)
		if err == nil {
			m.manager.EnablePlugin(p.Name)
			config.SaveConfig(&config.C)
			// Try to load it immediately
			_ = m.manager.LoadPlugin(p.Name)
		}
		m.ui.App.QueueUpdateDraw(func() {
			m.ui.Pages.RemovePage("progress")
			if err != nil {
				showErrorModalWithFocus(m.ui.App, m.ui.Pages, fmt.Sprintf("Failed to install plugin: %v", err), m.table)
				return
			}
			m.filterPlugins(m.searchField.GetText())
			m.ui.App.SetFocus(m.table)
		})
	}()
}

func (m *MarketplacePanel) updateDetails(p plugins.RegistryPlugin) {
	m.details.Clear()
	official := ""
	if p.Official {
		official = " [green](Official Plugin)[-]"
	}
	fmt.Fprintf(m.details, "[yellow]%s[-]%s\n", p.Name, official)
	fmt.Fprintf(m.details, "[green]Version:[-] %s\n", p.Version)
	fmt.Fprintf(m.details, "[green]Author:[-] %s\n", p.Author)
	fmt.Fprintf(m.details, "[blue]Repo:[-] %s\n\n", p.Repo)
	fmt.Fprintf(m.details, "%s\n", p.Description)
}

// ShowMarketplace displays the marketplace modal
func (ui *UIOrchestrator) ShowMarketplace() {
	m := NewMarketplacePanel(ui)

	// Fetch plugins in background
	go func() {
		pluginsList, err := m.client.ListPlugins()
		ui.App.QueueUpdateDraw(func() {
			if err != nil {
				showErrorModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Failed to fetch plugins: %v", err), m.searchField)
				return
			}
			m.plugins = pluginsList
			m.filterPlugins("")
		})
	}()

	ui.Pages.AddPage("marketplace", createModal(m, 100, 30, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
}
