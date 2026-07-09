package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/workspace"
)

const workspaceSearchPageName = "workspaceSearch"

type workspaceSearchItem struct {
	metadata     workspace.WorkspaceMetadata
	collections  int
	environments int
}

// WorkspaceSearchModal provides a searchable popup for quickly switching workspaces.
type WorkspaceSearchModal struct {
	*tview.Flex
	ui               *UIOrchestrator
	searchField      *tview.InputField
	table            *tview.Table
	workspaces       []workspaceSearchItem
	filtered         []workspaceSearchItem
	currentWorkspace string
	returnFocus      tview.Primitive
}

// NewWorkspaceSearchModal creates a new workspace search modal populated with workspace data.
func NewWorkspaceSearchModal(ui *UIOrchestrator) *WorkspaceSearchModal {
	m := &WorkspaceSearchModal{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		ui:          ui,
		table:       tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		returnFocus: ui.WorkspaceSelector,
	}

	m.SetBackgroundColor(ui.Colors.Background)

	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.searchField = tview.NewInputField().
		SetLabel(" Search Workspaces: ").
		SetLabelColor(ui.Colors.LabelColor).
		SetFieldBackgroundColor(ui.Colors.Selection).
		SetFieldTextColor(ui.Colors.Foreground)
	m.searchField.SetBackgroundColor(ui.Colors.Background)

	m.loadWorkspaces()

	m.searchField.SetChangedFunc(func(text string) {
		m.filterWorkspaces(text)
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.close()
			return nil
		}
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			if m.table.GetRowCount() > 1 {
				m.ui.App.SetFocus(m.table)
			}
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			if len(m.filtered) == 1 {
				m.selectWorkspace(m.filtered[0].metadata.Name)
			} else if m.table.GetRowCount() > 1 {
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
		if event.Key() == tcell.KeyEnter {
			m.selectCurrentRow()
			return nil
		}
		if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			m.close()
			return nil
		}
		return event
	})

	m.table.SetSelectedFunc(func(row, column int) {
		if row > 0 {
			m.selectCurrentRow()
		}
	})

	m.AddItem(m.searchField, 1, 0, true)
	m.AddItem(m.table, 0, 1, false)

	m.SetBorder(true).SetTitle(" Search Workspaces ")
	m.SetBorderColor(ui.Colors.BorderFocus)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.close()
			return nil
		}
		if event.Rune() == 'q' && m.ui.App.GetFocus() != m.searchField {
			m.close()
			return nil
		}
		return event
	})

	m.filterWorkspaces("")

	return m
}

func (m *WorkspaceSearchModal) loadWorkspaces() {
	manager, err := workspace.LoadWorkspaceManager()
	if err != nil {
		return
	}

	m.currentWorkspace = manager.CurrentWorkspace

	for _, metadata := range manager.Workspaces {
		item := workspaceSearchItem{
			metadata: metadata,
		}

		ws, err := workspace.LoadWorkspaceByName(metadata.Name)
		if err == nil && ws != nil {
			item.collections = countCollectionsRecursive(ws.Collections)
			item.environments = len(ws.Environments)
		} else {
			item.collections = -1
			item.environments = -1
		}

		m.workspaces = append(m.workspaces, item)
	}
}

func (m *WorkspaceSearchModal) filterWorkspaces(query string) {
	m.table.Clear()
	m.filtered = nil
	query = strings.ToLower(query)

	headers := []string{"Name", "Cols", "Envs", "Created", "Updated", "Current"}
	for i, h := range headers {
		m.table.SetCell(0, i, tview.NewTableCell(" "+h+" ").
			SetTextColor(m.ui.Colors.Title).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignCenter))
	}
	m.table.GetCell(0, 0).SetAlign(tview.AlignLeft)

	for _, item := range m.workspaces {
		if query != "" &&
			!strings.Contains(strings.ToLower(item.metadata.Name), query) &&
			!strings.Contains(strings.ToLower(item.metadata.Description), query) {
			continue
		}

		m.filtered = append(m.filtered, item)
		row := m.table.GetRowCount()

		m.table.SetCell(row, 0, tview.NewTableCell(" "+item.metadata.Name+" ").
			SetExpansion(3).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignLeft))

		m.table.SetCell(row, 1, tview.NewTableCell(" "+formatCount(item.collections)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 2, tview.NewTableCell(" "+formatCount(item.environments)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 3, tview.NewTableCell(" "+formatDate(item.metadata.CreatedAt)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		m.table.SetCell(row, 4, tview.NewTableCell(" "+formatDate(item.metadata.UpdatedAt)+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))

		currentText := ""
		if item.metadata.Name == m.currentWorkspace {
			currentText = "*"
		}
		m.table.SetCell(row, 5, tview.NewTableCell(" "+currentText+" ").
			SetExpansion(1).
			SetTextColor(m.ui.Colors.Foreground).
			SetAlign(tview.AlignCenter))
	}
}

func (m *WorkspaceSearchModal) selectCurrentRow() {
	row, _ := m.table.GetSelection()
	if row < 1 || row-1 >= len(m.filtered) {
		return
	}
	m.selectWorkspace(m.filtered[row-1].metadata.Name)
}

func (m *WorkspaceSearchModal) selectWorkspace(name string) {
	m.ui.SwitchWorkspace(name)
	m.close()
}

func (m *WorkspaceSearchModal) close() {
	m.ui.Pages.RemovePage(workspaceSearchPageName)
	m.ui.Pages.SwitchToPage("main")
	if m.returnFocus != nil {
		m.ui.App.SetFocus(m.returnFocus)
	}
	if m.ui.ExitModal != nil {
		m.ui.ExitModal()
	}
}

func showWorkspaceSearchModal(ui *UIOrchestrator) {
	ui.EnterModal()
	m := NewWorkspaceSearchModal(ui)
	ui.Pages.AddPage(workspaceSearchPageName, createModal(m, 110, 20, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
}

func countCollectionsRecursive(collections []workspace.Collection) int {
	n := len(collections)
	for i := range collections {
		n += countCollectionsRecursive(collections[i].Collections)
	}
	return n
}

func formatCount(n int) string {
	if n < 0 {
		return "?"
	}
	return fmt.Sprintf("%d", n)
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Format("2006-01-02")
}
