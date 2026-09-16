package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

const commandPalettePageName = "commandPalette"

// CommandPaletteModal provides a searchable menu of application commands,
// opened with Ctrl+P. Commands are rendered as one section per category.
type CommandPaletteModal struct {
	*tview.Flex
	ui              *UIOrchestrator
	searchField     *tview.InputField
	table           *tview.Table
	commands        []Command
	filtered        []Command
	rowCommands     map[int]*Command
	firstCommandRow int
	returnFocus     tview.Primitive
}

// NewCommandPaletteModal creates a new command palette modal populated with
// the command registry. The current focus is captured so it can be restored
// when the palette closes.
func NewCommandPaletteModal(ui *UIOrchestrator) *CommandPaletteModal {
	m := &CommandPaletteModal{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		ui:          ui,
		table:       tview.NewTable().SetSelectable(true, false),
		commands:    getCommands(),
		returnFocus: ui.App.GetFocus(),
	}

	m.SetBackgroundColor(ui.Colors.Background)

	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.searchField = createSearchField(" > ", "Type to filter commands...", ui.Colors)

	m.searchField.SetChangedFunc(func(text string) {
		m.filterCommands(text)
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.close()
			return nil
		}
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			m.focusTable()
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			if len(m.rowCommands) > 0 {
				m.table.Select(m.firstCommandRow, 0)
				m.selectCurrentRow()
			}
			return nil
		}
		return event
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := m.table.GetSelection()
		if event.Key() == tcell.KeyUp && row <= m.firstCommandRow {
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
		if event.Key() == tcell.KeyEscape {
			m.close()
			return nil
		}
		return event
	})

	m.table.SetSelectedFunc(func(row, column int) {
		m.selectCurrentRow()
	})

	m.AddItem(m.searchField, 1, 0, true)
	m.AddItem(m.table, 0, 1, false)

	m.SetBorder(true).SetTitle(" Command Palette ")
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

	m.filterCommands("")

	return m
}

// filterCommands narrows the visible commands to those whose label,
// description or category contain the query (case-insensitive) and re-renders
// the table. An empty query shows every command.
func (m *CommandPaletteModal) filterCommands(query string) {
	query = strings.ToLower(query)
	m.filtered = m.filtered[:0]
	for i := range m.commands {
		cmd := m.commands[i]
		if query == "" ||
			strings.Contains(strings.ToLower(cmd.Label), query) ||
			strings.Contains(strings.ToLower(cmd.Description), query) ||
			strings.Contains(strings.ToLower(cmd.Category), query) {
			m.filtered = append(m.filtered, cmd)
		}
	}
	m.renderCommands()
}

// renderCommands populates the table with one non-selectable section row per
// category, followed by the commands of that category. When the filter
// matches nothing, a placeholder row is shown instead. The first command row
// is pre-selected so Enter runs the top match.
func (m *CommandPaletteModal) renderCommands() {
	m.table.Clear()
	m.rowCommands = make(map[int]*Command)
	m.firstCommandRow = -1

	if len(m.filtered) == 0 {
		m.table.SetCell(0, 0, tview.NewTableCell(" No matching commands ").
			SetTextColor(m.ui.Colors.Placeholder).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignLeft))
		return
	}

	lastCategory := ""
	for i := range m.filtered {
		cmd := &m.filtered[i]

		if cmd.Category != lastCategory {
			lastCategory = cmd.Category
			row := m.table.GetRowCount()
			m.table.SetCell(row, 0, tview.NewTableCell(" "+cmd.Category+" ").
				SetTextColor(m.ui.Colors.Placeholder).
				SetSelectable(false).
				SetExpansion(1).
				SetAlign(tview.AlignLeft))
			m.table.SetCell(row, 1, tview.NewTableCell("").
				SetSelectable(false).
				SetExpansion(1))
		}

		row := m.table.GetRowCount()
		m.rowCommands[row] = cmd
		if m.firstCommandRow < 0 {
			m.firstCommandRow = row
		}

		m.table.SetCell(row, 0, tview.NewTableCell(" "+cmd.Label+" ").
			SetTextColor(m.ui.Colors.Foreground).
			SetExpansion(2).
			SetAlign(tview.AlignLeft))
		m.table.SetCell(row, 1, tview.NewTableCell(" "+cmd.Description+" ").
			SetTextColor(m.ui.Colors.Foreground).
			SetExpansion(3).
			SetAlign(tview.AlignLeft))
	}

	m.table.Select(m.firstCommandRow, 0)
}

// focusTable moves focus to the results table if it has any command rows.
func (m *CommandPaletteModal) focusTable() {
	if len(m.rowCommands) > 0 {
		m.ui.App.SetFocus(m.table)
	}
}

// selectCurrentRow executes the command selected in the table.
func (m *CommandPaletteModal) selectCurrentRow() {
	row, _ := m.table.GetSelection()
	cmd, ok := m.rowCommands[row]
	if !ok {
		return
	}
	m.executeCommand(cmd)
}

// executeCommand closes the palette and then runs the command handler, so
// handlers start from a clean UI state.
func (m *CommandPaletteModal) executeCommand(cmd *Command) {
	m.close()
	if cmd.Handler != nil {
		cmd.Handler(m.ui)
	}
}

// close removes the palette and restores the focus captured on open.
func (m *CommandPaletteModal) close() {
	m.ui.Pages.RemovePage(commandPalettePageName)
	m.ui.Pages.SwitchToPage("main")
	if m.returnFocus != nil {
		m.ui.App.SetFocus(m.returnFocus)
	}
	if m.ui.ExitModal != nil {
		m.ui.ExitModal()
	}
	m.ui.UpdateFooter()
}

// showCommandPaletteModal opens the command palette.
func showCommandPaletteModal(ui *UIOrchestrator) {
	ui.EnterModal()
	m := NewCommandPaletteModal(ui)
	ui.Pages.AddPage(commandPalettePageName, createSizedModal(m, modalSizeSearch, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
	ui.UpdateFooter()
}
