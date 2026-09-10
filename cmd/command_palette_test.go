package cmd

import (
	"testing"

	"github.com/rivo/tview"
)

func TestNewCommandPaletteModalRendering(t *testing.T) {
	ui := &UIOrchestrator{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		Colors:       &ColorManager{},
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

	// Row 0 is the first category section header: not selectable and not
	// mapped to a command.
	if c := m.table.GetCell(0, 0); !c.NotSelectable {
		t.Error("category section row should not be selectable")
	}
	if _, ok := m.rowCommands[0]; ok {
		t.Error("category section row must not map to a command")
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
