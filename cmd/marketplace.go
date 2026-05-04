package cmd

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium-plugin-sdk/types"
	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
)

// compareVersions compares two semantic versions (e.g., "1.2.3").
// Returns -1 if a < b, 0 if a == b, 1 if a > b.
func compareVersions(a, b string) int {
	parseV := func(v string) (major, minor, patch int) {
		parts := strings.Split(v, ".")
		if len(parts) >= 1 {
			major, _ = strconv.Atoi(parts[0])
		}
		if len(parts) >= 2 {
			minor, _ = strconv.Atoi(parts[1])
		}
		if len(parts) >= 3 {
			patch, _ = strconv.Atoi(parts[2])
		}
		return
	}

	ma, mi, pa := parseV(a)
	mb, mib, pb := parseV(b)

	if ma != mb {
		if ma < mb {
			return -1
		}
		return 1
	}
	if mi != mib {
		if mi < mib {
			return -1
		}
		return 1
	}
	if pa != pb {
		if pa < pb {
			return -1
		}
		return 1
	}
	return 0
}

// MarketplacePanel represents the plugin marketplace UI
type MarketplacePanel struct {
	*tview.Flex
	table           *tview.Table
	details         *tview.TextView
	plugins         []types.RegistryPlugin
	filteredPlugins []types.RegistryPlugin
	manager         *plugins.PluginManager
	client          *plugins.RegistryClient
	ui              *UIOrchestrator
	searchField     *tview.InputField
	returnFocus     tview.Primitive
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

	m.details.SetBorder(true).SetTitle(" Plugin Details ").SetTitleColor(ui.Colors.Foreground)
	m.details.SetBorderColor(ui.Colors.Border)
	m.details.SetBackgroundColor(ui.Colors.Background)
	m.details.SetTextColor(ui.Colors.Foreground)

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
		AddItem(m.table, 0, 2, true).
		AddItem(m.details, 0, 1, false)
	innerFlex.SetBackgroundColor(ui.Colors.Background)

	m.AddItem(innerFlex, 0, 1, false)

	marketplaceHost := "petitorium.dev"
	if config.C.Plugins.RegistryURL != "" {
		if u, err := url.Parse(config.C.Plugins.RegistryURL); err == nil && u.Host != "" {
			marketplaceHost = u.Host
		}
	}
	m.SetBorder(true).SetTitle(fmt.Sprintf(" Plugin Marketplace (%s) ", marketplaceHost))
	m.SetBorderColor(ui.Colors.BorderFocus)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("marketplace")
			if m.returnFocus != nil {
				m.ui.App.SetFocus(m.returnFocus)
			} else {
				m.ui.App.SetFocus(m.ui.CollectionsTreeView)
			}
			if m.ui.ExitModal != nil {
				m.ui.ExitModal()
			}
			return nil
		}
		// Only close with 'q' if NOT in search field
		if event.Rune() == 'q' && m.ui.App.GetFocus() != m.searchField {
			m.ui.Pages.RemovePage("marketplace")
			if m.returnFocus != nil {
				m.ui.App.SetFocus(m.returnFocus)
			} else {
				m.ui.App.SetFocus(m.ui.CollectionsTreeView)
			}
			if m.ui.ExitModal != nil {
				m.ui.ExitModal()
			}
			return nil
		}
		// Uninstall with 'u' key
		if event.Rune() == 'u' && m.ui.App.GetFocus() != m.searchField {
			row, _ := m.table.GetSelection()
			if row > 0 && row-1 < len(m.filteredPlugins) {
				m.handlePluginUninstall(m.filteredPlugins[row-1])
			}
			return nil
		}
		return event
	})

	return m
}

func (m *MarketplacePanel) filterPlugins(query string) {
	m.table.Clear()
	m.filteredPlugins = []types.RegistryPlugin{}
	query = strings.ToLower(query)

	// Set headers
	headers := []string{"Name", "Version", "Author", "Status", "Downloads"}
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
			statusText, _ := m.getPluginStatusInfo(p)
			downloadCount := strconv.FormatInt(p.DownloadCount, 10)

			m.table.SetCell(row, 0, tview.NewTableCell(p.Name).SetExpansion(2).SetTextColor(m.ui.Colors.Foreground))
			m.table.SetCell(row, 1, tview.NewTableCell(p.Version).SetAlign(tview.AlignCenter).SetTextColor(m.ui.Colors.Foreground))
			m.table.SetCell(row, 2, tview.NewTableCell(p.Author).SetAlign(tview.AlignCenter).SetTextColor(m.ui.Colors.Foreground))
			m.table.SetCell(row, 3, tview.NewTableCell(statusText).SetAlign(tview.AlignCenter).SetTextColor(m.ui.Colors.Foreground))
			m.table.SetCell(row, 4, tview.NewTableCell(downloadCount).SetAlign(tview.AlignCenter).SetTextColor(m.ui.Colors.Foreground))
			row++
		}
	}

	if len(m.filteredPlugins) > 0 {
		m.updateDetails(m.filteredPlugins[0])
	}
}

func (m *MarketplacePanel) getPluginStatusInfo(p types.RegistryPlugin) (string, tcell.Color) {
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		cmp := compareVersions(info.Version, p.Version)
		if cmp == 0 {
			return "Installed", tcell.ColorGreen
		}
		if cmp < 0 {
			return "Update", tcell.ColorYellow
		}
		return "Outdated", m.ui.Colors.Foreground
	}
	return "Available", m.ui.Colors.Foreground
}

func (m *MarketplacePanel) handlePluginAction(p types.RegistryPlugin) {
	needsReload := false
	if m.manager.IsPluginInstalled(p.Name) {
		info, _ := m.manager.GetInstalledInfo(p.Name)
		if compareVersions(info.Version, p.Version) >= 0 {
			return
		}
		needsReload = true
	}

	// Install/Update
	_, stopProgress := showProgressModal(m.ui.App, m.ui.Pages, " Installing Plugin ", fmt.Sprintf("Downloading %s...", p.Name), m.ui.Colors)

	go func() {
		// Unload the plugin if it's currently loaded to avoid "text file busy" error
		if needsReload {
			m.manager.UnloadPlugin(p.Name)
		}

		err := m.manager.InstallPlugin(p)
		if err == nil {
			m.manager.EnablePlugin(p.Name)
			config.SaveConfig(&config.C)
			// Reload the updated plugin
			_ = m.manager.LoadPlugin(p.Name)
			// Also reload other plugins that may have been unloaded by UnloadPlugin
			for _, name := range m.manager.GetEnabledPlugins() {
				if name != p.Name {
					_ = m.manager.LoadPlugin(name)
				}
			}
		}
		stopProgress()
		m.ui.App.QueueUpdateDraw(func() {
			m.ui.Pages.RemovePage("progress")
			if err != nil {
				showErrorModalWithFocus(m.ui.App, m.ui.Pages, fmt.Sprintf("Failed to install plugin: %v", err), m.table, m.ui.Colors)
				return
			}
			m.filterPlugins(m.searchField.GetText())
			m.ui.App.SetFocus(m.table)
		})
	}()
}

func (m *MarketplacePanel) handlePluginUninstall(p types.RegistryPlugin) {
	if !m.manager.IsPluginInstalled(p.Name) {
		return
	}

	form := showConfirmModal(
		m.ui.Pages,
		fmt.Sprintf(" Uninstall Plugin "),
		fmt.Sprintf("Uninstall [yellow]%s[-]? This will remove the plugin file.", p.Name),
		[]string{"Cancel", "Uninstall"},
		m.ui.Colors,
		func(buttonIndex int) {
			if buttonIndex == 1 {
				m.performUninstall(p.Name)
			} else {
				m.ui.App.SetFocus(m.table)
			}
		},
	)

	// Attach input capture to handle Esc key and restore focus to the table
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("confirm")
			m.ui.App.SetFocus(m.table)
			return nil
		}
		return event
	})

	// Explicitly set focus to the form so it captures key events
	m.ui.App.SetFocus(form)
}

func (m *MarketplacePanel) performUninstall(name string) {
	_, stopProgress := showProgressModal(m.ui.App, m.ui.Pages, " Uninstalling Plugin ", fmt.Sprintf("Removing %s...", name), m.ui.Colors)

	go func() {
		err := m.manager.UninstallPlugin(name)
		if err == nil {
			config.SaveConfig(&config.C)
		}
		stopProgress()
		m.ui.App.QueueUpdateDraw(func() {
			m.ui.Pages.RemovePage("progress")
			if err != nil {
				showErrorModalWithFocus(m.ui.App, m.ui.Pages, fmt.Sprintf("Failed to uninstall plugin: %v", err), m.table, m.ui.Colors)
				return
			}
			m.filterPlugins(m.searchField.GetText())
			m.ui.App.SetFocus(m.table)
		})
	}()
}

func (m *MarketplacePanel) updateDetails(p types.RegistryPlugin) {
	m.details.Clear()

	versionR, versionG, versionB := m.ui.Colors.Error.RGB()
	authorR, authorG, authorB := m.ui.Colors.Success.RGB()
	repoR, repoG, repoB := m.ui.Colors.LabelColor.RGB()

	fmt.Fprintf(m.details, "[#%02x%02x%02x]Version:[-] %s\n", versionR, versionG, versionB, p.Version)
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Author:[-] %s\n", authorR, authorG, authorB, p.Author)
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Repository:[-] %s\n\n", repoR, repoG, repoB, p.Repository)
	fmt.Fprintf(m.details, "%s\n", p.Description)
}

// ShowMarketplace displays the marketplace modal
func (ui *UIOrchestrator) ShowMarketplace() {
	currentFocus := ui.App.GetFocus()
	m := NewMarketplacePanel(ui)
	m.returnFocus = currentFocus
	ui.EnterModal()

	// Fetch plugins in background
	go func() {
		pluginsList, err := m.client.ListPlugins()
		ui.App.QueueUpdateDraw(func() {
			if err != nil {
				showErrorModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Failed to fetch plugins: %v", err), m.searchField, ui.Colors)
				return
			}
			m.plugins = pluginsList
			m.filterPlugins("")
		})
	}()

	ui.Pages.AddPage("marketplace", createModal(m, 130, 30, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
}
