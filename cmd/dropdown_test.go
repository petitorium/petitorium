package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TestFormDropDownKeepsAppBackground verifies that a dropdown placed inside a
// tview.Form keeps its unified field background. A form re-applies its own field
// style to every item on each draw, which would otherwise paint the dropdown
// field with the form's field background (the lighter input shade, or tview's
// default blue) instead of the app background.
func TestFormDropDownKeepsAppBackground(t *testing.T) {
	appBg := tcell.NewHexColor(0x00ff00)
	formFieldBg := tcell.NewHexColor(0xff0000)

	colors := &ColorManager{
		Background:         appBg,
		ActiveTab:          tcell.NewHexColor(0x7aa2f7),
		DropdownFocus:      tcell.NewHexColor(0x24283b),
		ButtonBackground:   tcell.NewHexColor(0x414868),
		SelectedBackground: tcell.NewHexColor(0x7aa2f7),
		SelectedForeground: tcell.NewHexColor(0x1a1b26),
		Border:             tcell.NewHexColor(0x414868),
	}

	form := tview.NewForm()
	form.SetBackgroundColor(tcell.NewHexColor(0x000000))
	form.SetFieldBackgroundColor(formFieldBg)

	dropdown := tview.NewDropDown().
		SetLabel("Move to: ").
		SetOptions([]string{"Root", "Alpha"}, nil)
	form.AddFormItem(newFormDropDown(dropdown, colors))

	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(40, 5)
	form.SetRect(0, 0, 40, 5)
	form.Draw(screen)

	sawAppBg := false
	for y := 0; y < 5; y++ {
		for x := 0; x < 40; x++ {
			_, _, style, _ := screen.GetContent(x, y)
			if _, bg, _ := style.Decompose(); bg == formFieldBg {
				t.Fatalf("dropdown field was overridden by the form field background at (%d, %d)", x, y)
			} else if bg == appBg {
				sawAppBg = true
			}
		}
	}
	if !sawAppBg {
		t.Error("expected the dropdown field to be drawn with the app background")
	}
}
