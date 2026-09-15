package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestNewCommandPaletteModalRendering(t *testing.T) {
	ui := &UIOrchestrator{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		Colors:       &ColorManager{Placeholder: tcell.ColorTeal},
		UpdateFooter: func() {},
	}

	m := NewCommandPaletteModal(ui)

	commands := getCommands()
	categories := 0
	last := ""
	for _, cmd := range commands {
		if cmd.Category != last {
			categories++
			last = cmd.Category
		}
	}

	if got := m.table.GetRowCount(); got != len(commands)+categories {
		t.Errorf("expected %d table rows (commands + category sections), got %d", len(commands)+categories, got)
	}
	if len(m.rowCommands) != len(commands) {
		t.Errorf("expected %d mapped command rows, got %d", len(commands), len(m.rowCommands))
	}
	if m.firstCommandRow != 1 {
		t.Errorf("expected first command row at index 1, got %d", m.firstCommandRow)
	}

	// Row 0 is the first category section header: not selectable, not mapped
	// to a command, and rendered in the subdued placeholder color so section
	// headers don't compete with commands for attention. NewTableCell parks
	// text colors in the cell Style, so compare against a reference cell.
	if c := m.table.GetCell(0, 0); !c.NotSelectable {
		t.Error("category section row should not be selectable")
	}
	if want, got := tview.NewTableCell("").SetTextColor(tcell.ColorTeal).Style, m.table.GetCell(0, 0).Style; got != want {
		t.Errorf("category section row should use the subdued placeholder color, got style %v, want %v", got, want)
	}
	if _, ok := m.rowCommands[0]; ok {
		t.Error("category section row must not map to a command")
	}
}

func TestCommandPaletteFilterCommands(t *testing.T) {
	ui := &UIOrchestrator{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		Colors:       &ColorManager{},
		UpdateFooter: func() {},
	}

	m := NewCommandPaletteModal(ui)
	commands := getCommands()

	if got := len(m.filtered); got != len(commands) {
		t.Errorf("empty query should show all %d commands, got %d", len(commands), got)
	}
	if got := len(m.rowCommands); got != len(commands) {
		t.Errorf("empty query should map all %d command rows, got %d", len(commands), got)
	}

	// Filtering is case-insensitive and matches labels.
	m.filterCommands("RENAME")
	if got := len(m.filtered); got != 1 {
		t.Fatalf("expected 1 command matching 'RENAME', got %d", got)
	}
	if m.filtered[0].ID != "edit.renameItem" {
		t.Errorf("expected edit.renameItem, got %s", m.filtered[0].ID)
	}
	if row, _ := m.table.GetSelection(); row != m.firstCommandRow {
		t.Errorf("expected selection on first command row %d, got %d", m.firstCommandRow, row)
	}

	// Queries also match against descriptions and categories.
	m.filterCommands("workspace")
	if got := len(m.filtered); got != 2 {
		t.Errorf("expected 2 commands matching 'workspace', got %d", got)
	}
	m.filterCommands("file")
	if got := len(m.filtered); got != 3 {
		t.Errorf("expected 3 commands in the File category, got %d", got)
	}

	// A query with no matches clears the command rows and shows the
	// placeholder row instead.
	m.filterCommands("zzz-nothing-matches")
	if got := len(m.rowCommands); got != 0 {
		t.Errorf("expected no mapped command rows for a non-matching query, got %d", got)
	}
	if m.firstCommandRow != -1 {
		t.Errorf("expected firstCommandRow -1 for a non-matching query, got %d", m.firstCommandRow)
	}
	if c := m.table.GetCell(0, 0); c.Text != " No matching commands " || !c.NotSelectable {
		t.Errorf("expected non-selectable placeholder row, got %q (NotSelectable=%v)", c.Text, c.NotSelectable)
	}

	// Clearing the query restores the full list.
	m.filterCommands("")
	if got := len(m.filtered); got != len(commands) {
		t.Errorf("expected full list of %d commands after clearing the query, got %d", len(commands), got)
	}
}

func TestCommandPaletteExecuteCommand(t *testing.T) {
	app := tview.NewApplication()
	pages := tview.NewPages()
	ui := &UIOrchestrator{
		App:          app,
		Pages:        pages,
		Colors:       &ColorManager{},
		UpdateFooter: func() {},
	}

	m := NewCommandPaletteModal(ui)
	pages.AddPage(commandPalettePageName, m, true, true)
	if got := pages.GetPageCount(); got != 1 {
		t.Fatalf("expected 1 page after opening the palette, got %d", got)
	}

	executed := false
	cmd := Command{ID: "test", Label: "Test", Category: "Test", Description: "", Handler: func(ui *UIOrchestrator) {
		executed = true
	}}
	m.executeCommand(&cmd)

	if !executed {
		t.Error("command handler was not executed")
	}
	if got := pages.GetPageCount(); got != 0 {
		t.Errorf("palette page should be removed after executing a command, got %d pages", got)
	}
}
