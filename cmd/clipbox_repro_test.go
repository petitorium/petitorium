package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/workspace"
)

// assertNoDefaultBackground draws the primitive and fails if any cell renders
// with tcell.ColorDefault or tcell.ColorBlack (which would appear as a black
// box instead of the themed background).
func assertNoDefaultBackground(t *testing.T, p tview.Primitive, w, h int) {
	t.Helper()
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(w, h)
	p.SetRect(0, 0, w, h)
	p.Draw(screen)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			_, _, style, _ := screen.GetContent(x, y)
			_, bg, _ := style.Decompose()
			if bg == tcell.ColorBlack || bg == tcell.ColorDefault {
				t.Errorf("found default/black background at (%d,%d)", x, y)
			}
		}
	}
}

func TestHeadersAndQueryTabsNoBlackBackground(t *testing.T) {
	colors := NewColorManager()
	app := tview.NewApplication()
	pages := tview.NewPages()

	headers := createHeadersTabWithData(colors, map[string]workspace.Entry{}, nil, nil, app, pages, nil)
	query := createQueryParamsTabWithData(colors, map[string]workspace.Entry{}, nil, nil, app, pages, nil)

	t.Run("headers empty", func(t *testing.T) {
		assertNoDefaultBackground(t, headers, 80, 40)
	})
	t.Run("query empty", func(t *testing.T) {
		assertNoDefaultBackground(t, query, 80, 40)
	})
}
