package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

func TestApplyThemeLiveMutatesExistingColorManager(t *testing.T) {
	saved := config.C
	defer func() { config.C = saved }()

	colors := NewColorManager()
	ui := &UIOrchestrator{
		App:               tview.NewApplication(),
		Colors:            colors,
		WorkspaceData:     &workspace.Workspace{},
		BodyViewPanel:     tview.NewTextView(),
		TabHeader:         tview.NewFlex(),
		ResponseTabHeader: tview.NewFlex(),
		UpdateFooter:      func() {},
	}

	originalPointer := ui.Colors
	if err := applyThemeLive(ui, "dracula"); err != nil {
		t.Fatalf("applyThemeLive returned error: %v", err)
	}

	if ui.Colors != originalPointer {
		t.Error("applyThemeLive must mutate the existing ColorManager pointer in place")
	}
	if got, want := ui.Colors.Background, GetThemeManager().GetColorManager().Background; got != want {
		t.Errorf("ui.Colors.Background = %v, want %v", got, want)
	}
	if config.C.SyntaxTheme != "dracula" {
		t.Errorf("config.C.SyntaxTheme = %q, want %q", config.C.SyntaxTheme, "dracula")
	}
}

func TestThemePickerThemesSorted(t *testing.T) {
	names := themePickerThemes()
	if len(names) == 0 {
		t.Fatal("themePickerThemes returned no themes")
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] >= names[i] {
			t.Fatalf("themePickerThemes not sorted at %d: %q >= %q", i, names[i-1], names[i])
		}
	}
}

func TestThemePickerModalFiltering(t *testing.T) {
	ui := &UIOrchestrator{
		App:          tview.NewApplication(),
		Pages:        tview.NewPages(),
		Colors:       &ColorManager{Placeholder: tcell.ColorTeal},
		UpdateFooter: func() {},
	}

	m := NewThemePickerModal(ui)

	if got := len(m.filtered); got != len(m.themes) {
		t.Errorf("empty query should show all %d themes, got %d", len(m.themes), got)
	}

	m.filterThemes("tokyonight")
	if got := len(m.filtered); got != 2 {
		t.Fatalf("expected 2 themes matching 'tokyonight', got %d", got)
	}

	m.filterThemes("zzz-nothing")
	if got := len(m.rowThemes); got != 0 {
		t.Errorf("expected no rows for a non-matching query, got %d", got)
	}
}

func TestSelectThemeCommandRegistered(t *testing.T) {
	found := false
	for _, cmd := range getCommands() {
		if cmd.ID == "view.selectTheme" {
			found = true
			if cmd.Handler == nil {
				t.Error("view.selectTheme command must have a handler")
			}
		}
	}
	if !found {
		t.Error("getCommands() must include the view.selectTheme command")
	}
}
