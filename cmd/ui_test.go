package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestCreateSearchFieldConsistentBackground(t *testing.T) {
	colors := &ColorManager{
		Background:      tcell.ColorBlack,
		Foreground:      tcell.ColorWhite,
		LabelColor:      tcell.ColorPurple,
		Placeholder:     tcell.ColorGray,
		InputBackground: tcell.ColorTeal,
	}

	field := createSearchField(" > ", "Type to filter commands...", colors)

	// tview renders the input area with the field style when text is present
	// and with the placeholder style when the field is empty. Both styles
	// must use the same background or the box changes color as soon as the
	// user types.
	_, fieldBg, _ := field.GetFieldStyle().Decompose()
	_, placeholderBg, _ := field.GetPlaceholderStyle().Decompose()
	if fieldBg != tcell.ColorTeal {
		t.Errorf("field style background = %v, want InputBackground (teal)", fieldBg)
	}
	if placeholderBg != tcell.ColorTeal {
		t.Errorf("placeholder style background = %v, want InputBackground (teal)", placeholderBg)
	}

	placeholderFg, _, _ := field.GetPlaceholderStyle().Decompose()
	if placeholderFg != tcell.ColorGray {
		t.Errorf("placeholder style foreground = %v, want Placeholder (gray)", placeholderFg)
	}

	// The search box itself (the area around the label and field) must use
	// the modal background, not the tview default.
	if got := field.GetBackgroundColor(); got != tcell.ColorBlack {
		t.Errorf("search field background = %v, want Background (black)", got)
	}
}
