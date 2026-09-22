package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/workspace"
)

const workspaceConfigPageName = "workspaceModal"

// WorkspaceConfigPanel presents workspace management using the same split
// layout as the plugin marketplace: a search field, a selectable table, and a
// details panel.
type WorkspaceConfigPanel struct {
	*tview.Flex
	ui               *UIOrchestrator
	searchField      *tview.InputField
	table            *tview.Table
	details          *tview.TextView
	workspaces       []workspace.WorkspaceMetadata
	filtered         []workspace.WorkspaceMetadata
	currentWorkspace string
	returnFocus      tview.Primitive
}

// NewWorkspaceConfigPanel creates a new workspace configuration panel.
func NewWorkspaceConfigPanel(ui *UIOrchestrator) *WorkspaceConfigPanel {
	m := &WorkspaceConfigPanel{
		Flex:    tview.NewFlex().SetDirection(tview.FlexRow),
		table:   tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		details: tview.NewTextView().SetDynamicColors(true).SetWrap(true),
		ui:      ui,
	}

	m.SetBackgroundColor(ui.Colors.Background)

	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.details.SetBorder(true).SetTitle(" Workspace Information ").SetTitleColor(ui.Colors.Foreground)
	m.details.SetBorderColor(ui.Colors.Border)
	m.details.SetBackgroundColor(ui.Colors.Background)
	m.details.SetTextColor(ui.Colors.Foreground)

	m.searchField = createSearchField(" Search Workspaces: ", "", ui.Colors)

	manager, err := workspace.LoadWorkspaceManager()
	if err == nil && manager != nil {
		m.currentWorkspace = manager.CurrentWorkspace
		m.workspaces = manager.Workspaces
	}

	m.searchField.SetChangedFunc(func(text string) {
		m.filterWorkspaces(text)
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			if m.table.GetRowCount() > 1 {
				m.ui.App.SetFocus(m.table)
			}
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
		if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		return event
	})

	m.table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(m.filtered) {
			m.updateDetails(m.filtered[row-1])
		}
	})

	m.table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row-1 < len(m.filtered) {
			m.selectWorkspace(m.filtered[row-1].Name)
		}
	})

	m.AddItem(m.searchField, 1, 0, true)

	innerFlex := tview.NewFlex().
		AddItem(m.table, 0, 2, true).
		AddItem(m.details, 0, 1, false)
	innerFlex.SetBackgroundColor(ui.Colors.Background)

	m.AddItem(innerFlex, 0, 1, false)

	m.SetBorder(true).SetTitle(" Workspace Configuration ")
	m.SetBorderColor(ui.Colors.BorderFocus)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.close()
			return nil
		}
		if m.ui.App.GetFocus() == m.searchField {
			return event
		}
		switch event.Rune() {
		case 'q', 'Q':
			m.close()
			return nil
		case 'N':
			m.createNewWorkspace()
			return nil
		case 'c', 'C':
			m.duplicateSelected()
			return nil
		case 'r':
			m.renameSelected()
			return nil
		case 'd':
			m.deleteSelected()
			return nil
		}
		return event
	})

	m.filterWorkspaces("")

	return m
}

func (m *WorkspaceConfigPanel) filterWorkspaces(query string) {
	m.table.Clear()
	m.filtered = nil
	query = strings.ToLower(strings.TrimSpace(query))

	headers := []string{"Name", "Cols", "Envs", "Created", "Updated", "Current"}
	for i, h := range headers {
		m.table.SetCell(0, i, tview.NewTableCell(" "+h+" ").
			SetTextColor(m.ui.Colors.Title).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignCenter))
	}
	m.table.GetCell(0, 0).SetAlign(tview.AlignLeft)

	for _, ws := range m.workspaces {
		if query != "" &&
			!strings.Contains(strings.ToLower(ws.Name), query) &&
			!strings.Contains(strings.ToLower(ws.Description), query) {
			continue
		}

		m.filtered = append(m.filtered, ws)
		row := m.table.GetRowCount()

		collections := -1
		environments := -1
		if full, err := workspace.LoadWorkspaceByName(ws.Name); err == nil && full != nil {
			collections = countCollectionsRecursive(full.Collections)
			environments = len(full.Environments)
		}

		m.table.SetCell(row, 0, tview.NewTableCell(" "+ws.Name+" ").
			SetExpansion(3).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignLeft))

		m.table.SetCell(row, 1, tview.NewTableCell(" "+formatCount(collections)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 2, tview.NewTableCell(" "+formatCount(environments)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 3, tview.NewTableCell(" "+formatDate(ws.CreatedAt)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 4, tview.NewTableCell(" "+formatDate(ws.UpdatedAt)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		currentText := ""
		if ws.Name == m.currentWorkspace {
			currentText = "*"
		}
		m.table.SetCell(row, 5, tview.NewTableCell(" "+currentText+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))
	}

	if len(m.filtered) > 0 {
		m.updateDetails(m.filtered[0])
	}
}

func (m *WorkspaceConfigPanel) updateDetails(ws workspace.WorkspaceMetadata) {
	m.details.Clear()

	labelR, labelG, labelB := m.ui.Colors.LabelColor.RGB()
	valueR, valueG, valueB := m.ui.Colors.Foreground.RGB()
	accentR, accentG, accentB := m.ui.Colors.Success.RGB()

	name := ws.Name
	if ws.Name == m.currentWorkspace {
		name += " (current)"
	}

	description := ws.Description
	collections := -1
	environments := -1
	created := ws.CreatedAt
	updated := ws.UpdatedAt
	if full, err := workspace.LoadWorkspaceByName(ws.Name); err == nil && full != nil {
		description = full.Description
		collections = countCollectionsRecursive(full.Collections)
		environments = len(full.Environments)
		created = full.CreatedAt
		updated = full.UpdatedAt
	}

	fmt.Fprintf(m.details, "[#%02x%02x%02x]Name:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, accentR, accentG, accentB, name)
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Description:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, valueR, valueG, valueB, description)
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Collections:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, valueR, valueG, valueB, formatCount(collections))
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Environments:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, valueR, valueG, valueB, formatCount(environments))
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Created:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, valueR, valueG, valueB, formatDate(created))
	fmt.Fprintf(m.details, "[#%02x%02x%02x]Updated:[-] [#%02x%02x%02x]%s[-]\n", labelR, labelG, labelB, valueR, valueG, valueB, formatDate(updated))
}

func (m *WorkspaceConfigPanel) selectedWorkspace() *workspace.WorkspaceMetadata {
	row, _ := m.table.GetSelection()
	if row < 1 || row-1 >= len(m.filtered) {
		return nil
	}
	return &m.filtered[row-1]
}

func (m *WorkspaceConfigPanel) selectWorkspace(name string) {
	m.ui.SwitchWorkspace(name)
	m.close()
}

func (m *WorkspaceConfigPanel) createNewWorkspace() {
	currentFocus := m.ui.App.GetFocus()
	form := createNewWorkspaceForm(m.ui.App, m.ui.Pages, m.ui, m.ui.RootNode, m.ui.CollectionsTreeView, m.ui.Colors)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("createWorkspace")
			m.ui.App.SetFocus(currentFocus)
			return nil
		}
		return event
	})
	modal := createSizedModal(form, modalSizeForm, m.ui.Colors.Background)
	m.ui.Pages.AddPage("createWorkspace", modal, true, true)
	m.ui.App.SetFocus(form)
}

func (m *WorkspaceConfigPanel) deleteSelected() {
	ws := m.selectedWorkspace()
	if ws == nil || ws.Name == m.currentWorkspace {
		return
	}
	currentFocus := m.ui.App.GetFocus()
	form := createDeleteWorkspaceForm(m.ui.App, m.ui.Pages, ws.Name, m.ui.WorkspaceConfigButton, m.ui.Colors)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("deleteWorkspace")
			m.ui.App.SetFocus(currentFocus)
			return nil
		}
		return event
	})
	modal := createSizedModal(form, modalSizeConfirm, m.ui.Colors.Background)
	m.ui.Pages.AddPage("deleteWorkspace", modal, true, true)
	m.ui.App.SetFocus(form)
}

func (m *WorkspaceConfigPanel) renameSelected() {
	ws := m.selectedWorkspace()
	if ws == nil {
		return
	}
	currentFocus := m.ui.App.GetFocus()
	form := createRenameWorkspaceForm(m.ui.App, m.ui.Pages, ws.Name, m.ui.WorkspaceConfigButton, m.ui.Colors)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("renameWorkspace")
			m.ui.App.SetFocus(currentFocus)
			return nil
		}
		return event
	})
	modal := createSizedModal(form, modalSizeForm, m.ui.Colors.Background)
	m.ui.Pages.AddPage("renameWorkspace", modal, true, true)
	m.ui.App.SetFocus(form)
}

func (m *WorkspaceConfigPanel) duplicateSelected() {
	ws := m.selectedWorkspace()
	if ws == nil {
		return
	}
	currentFocus := m.ui.App.GetFocus()
	form := createDuplicateWorkspaceForm(m.ui.App, m.ui.Pages, ws.Name, m.ui.WorkspaceConfigButton, m.ui.Colors)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			m.ui.Pages.RemovePage("duplicateWorkspace")
			m.ui.App.SetFocus(currentFocus)
			return nil
		}
		return event
	})
	modal := createSizedModal(form, modalSizeForm, m.ui.Colors.Background)
	m.ui.Pages.AddPage("duplicateWorkspace", modal, true, true)
	m.ui.App.SetFocus(form)
}

func (m *WorkspaceConfigPanel) close() {
	m.ui.Pages.RemovePage(workspaceConfigPageName)
	m.ui.Pages.SwitchToPage("main")
	if m.returnFocus != nil {
		m.ui.App.SetFocus(m.returnFocus)
	} else {
		m.ui.App.SetFocus(m.ui.WorkspaceConfigButton)
	}
	if m.ui.ExitModal != nil {
		m.ui.ExitModal()
	}
}

// showWorkspaceModal displays a modal for workspace configuration.
func showWorkspaceModal(ui *UIOrchestrator) {
	previousFocus := ui.App.GetFocus()
	m := NewWorkspaceConfigPanel(ui)
	m.returnFocus = previousFocus
	ui.EnterModal()

	ui.Pages.RemovePage(workspaceConfigPageName)
	ui.Pages.AddPage(workspaceConfigPageName, createSizedModal(m, modalSizeFullscreen, ui.Colors.Background), true, true)
	ui.UpdateFooter()
	ui.App.SetFocus(m.searchField)
}
