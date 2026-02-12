package cmd

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
)

// MarketplacePanel represents the plugin marketplace UI
type MarketplacePanel struct {
	*tview.Flex
	list        *tview.List
	details     *tview.TextView
	plugins     []plugins.RegistryPlugin
	manager     *plugins.PluginManager
	client      *plugins.RegistryClient
	ui          *UIOrchestrator
	searchField *tview.InputField
}

// NewMarketplacePanel creates a new MarketplacePanel
func NewMarketplacePanel(ui *UIOrchestrator) *MarketplacePanel {
	m := &MarketplacePanel{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		list:    tview.NewList(),
		details: tview.NewTextView().SetDynamicColors(true).SetWrap(true),
		manager: ui.PluginManager,
		client:  plugins.NewRegistryClient(config.C.Plugins.RegistryURL),
		ui:      ui,
	}

	m.list.SetSelectedFocusOnly(true)
	m.list.SetMainTextColor(ui.Colors.Foreground)
	m.list.SetSelectedBackgroundColor(ui.Colors.Selection)

	m.details.SetBorder(true).SetTitle(" Plugin Details ")
	m.details.SetBorderColor(ui.Colors.Border)

	m.searchField = tview.NewInputField().
		SetLabel(" Search Plugins: ").
		SetLabelColor(ui.Colors.LabelColor).
		SetFieldBackgroundColor(ui.Colors.Selection).
		SetFieldTextColor(ui.Colors.Foreground)

	m.searchField.SetChangedFunc(func(text string) {
		m.filterPlugins(text)
	})

	m.list.SetChangedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(m.plugins) {
			m.updateDetails(m.plugins[index])
		}
	})

	m.list.SetSelectedFunc(func(index int, mainText, secondaryText string, shortcut rune) {
		if index >= 0 && index < len(m.plugins) {
			m.handlePluginAction(m.plugins[index])
		}
	})

	m.AddItem(m.searchField, 1, 0, true)
	m.AddItem(tview.NewFlex().
		AddItem(m.list, 0, 1, true).
		AddItem(m.details, 0, 1, false), 0, 1, false)

	m.SetBorder(true).SetTitle(" Plugin Marketplace (petitorium.dev) ")
	m.SetBorderColor(ui.Colors.Border)

	return m
}

func (m *MarketplacePanel) filterPlugins(query string) {
	m.list.Clear()
	query = strings.ToLower(query)
	for _, p := range m.plugins {
		if query == "" || strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Description), query) {
			status := m.getPluginStatus(p)
			m.list.AddItem(p.Name, fmt.Sprintf("%s - %s", p.Version, status), 0, nil)
		}
	}
}

func (m *MarketplacePanel) getPluginStatus(p plugins.RegistryPlugin) string {
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		if info.Version != p.Version {
			return "[yellow]Update Available[-]"
		}
		return "[green]Installed[-]"
	}
	return "Available"
}

func (m *MarketplacePanel) handlePluginAction(p plugins.RegistryPlugin) {
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		if info.Version == p.Version {
			// Already up to date
			return
		}
	}

	// Install/Update
	_ = showProgressModal(m.ui.Pages, " Installing Plugin ", fmt.Sprintf("Downloading %s...", p.Name))

	go func() {
		err := m.manager.InstallPlugin(p)
		m.ui.App.QueueUpdateDraw(func() {
			m.ui.Pages.RemovePage("progress")
			if err != nil {
				showErrorModal(m.ui.Pages, fmt.Sprintf("Failed to install plugin: %v", err))
				return
			}
			m.filterPlugins(m.searchField.GetText())
		})
	}()
}

func (m *MarketplacePanel) updateDetails(p plugins.RegistryPlugin) {
	m.details.Clear()
	fmt.Fprintf(m.details, "[yellow]%s[-]\n", p.Name)
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
				showErrorModal(ui.Pages, fmt.Sprintf("Failed to fetch plugins: %v", err))
				return
			}
			m.plugins = pluginsList
			m.filterPlugins("")
		})
	}()

	ui.Pages.AddPage("marketplace", createModal(m, 100, 30, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
}
