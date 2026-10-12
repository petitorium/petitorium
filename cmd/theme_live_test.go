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

func TestRecolorPrimitiveDropdownUsesAppBackground(t *testing.T) {
	c := &ColorManager{
		Background:      tcell.NewHexColor(0x1a1b26),
		InputBackground: tcell.NewHexColor(0x24283b),
		Foreground:      tcell.NewHexColor(0xffffff),
		LabelColor:      tcell.NewHexColor(0xbb9af7),
		ActiveTab:       tcell.NewHexColor(0x7aa2f7),
	}

	dropdown := tview.NewDropDown().
		SetOptions([]string{"option-one", "option-two"}, nil).
		SetCurrentOption(0)
	recolorPrimitive(dropdown, c)

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(30, 3)
	dropdown.SetRect(0, 0, 30, 1)
	dropdown.Draw(screen)

	// Without a label the field starts at the top-left cell. Its resting
	// background must match the app background, not the input background.
	_, _, style, _ := screen.GetContent(0, 0)
	if _, bg, _ := style.Decompose(); bg != c.Background {
		t.Errorf("dropdown field background = %v, want app background %v", bg.Hex(), c.Background.Hex())
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
