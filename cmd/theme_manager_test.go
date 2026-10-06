package cmd

import (
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/petitorium/petitorium/config"
)

func TestTokyoNightNightTheme(t *testing.T) {
	tm := GetThemeManager()
	theme, err := tm.GetTheme("tokyonight-night")
	if err != nil {
		t.Fatalf("expected tokyonight-night theme to exist: %v", err)
	}

	if theme.SyntaxTheme != "tokyonight-night" {
		t.Errorf("syntax theme = %q, want %q", theme.SyntaxTheme, "tokyonight-night")
	}

	c := theme.UIColors

	assertColor(t, "Background", c.Background, "#1a1b26")
	assertColor(t, "Foreground", c.Foreground, "#a9b1d6")
	assertColor(t, "Border", c.Border, "#414868")
	assertColor(t, "BorderFocus", c.BorderFocus, "#7aa2f7")
	assertColor(t, "Title", c.Title, "#c0caf5")
	assertColor(t, "Selection", c.Selection, "#24283b")
	assertColor(t, "SelectedBackground", c.SelectedBackground, "#7aa2f7")
	assertColor(t, "SelectedForeground", c.SelectedForeground, "#1a1b26")
	assertColor(t, "ActiveTab", c.ActiveTab, "#7aa2f7")
	assertColor(t, "ButtonBackground", c.ButtonBackground, "#414868")
	assertColor(t, "ButtonSelected", c.ButtonSelected, "#bb9af7")
	assertColor(t, "DropdownFocused", c.DropdownFocused, "#24283b")
	assertColor(t, "Placeholder", c.Placeholder, "#565f89")
	assertColor(t, "InputBackground", c.InputBackground, "#24283b")
	assertColor(t, "Success", c.Success, "#9ece6a")
	assertColor(t, "Error", c.Error, "#f7768e")
	assertColor(t, "Warning", c.Warning, "#e0af68")
	assertColor(t, "SelectedRequestIcon", c.SelectedRequestIcon, "#9ece6a")
	assertColor(t, "LabelColor", c.LabelColor, "#bb9af7")
	assertColor(t, "ValueColor", c.ValueColor, "#c0caf5")

	assertColor(t, "GET method", c.MethodColors.GET, "#9ece6a")
	assertColor(t, "POST method", c.MethodColors.POST, "#7aa2f7")
	assertColor(t, "PUT method", c.MethodColors.PUT, "#e0af68")
	assertColor(t, "PATCH method", c.MethodColors.PATCH, "#ff9e64")
	assertColor(t, "DELETE method", c.MethodColors.DELETE, "#f7768e")
	assertColor(t, "OPTIONS method", c.MethodColors.OPTIONS, "#565f89")
	assertColor(t, "HEAD method", c.MethodColors.HEAD, "#73daca")

	assertColor(t, "Status Success", c.StatusColors.Success, "#9ece6a")
	assertColor(t, "Status SuccessText", c.StatusColors.SuccessText, "#1a1b26")
	assertColor(t, "Status Redirection", c.StatusColors.Redirection, "#e0af68")
	assertColor(t, "Status ClientError", c.StatusColors.ClientError, "#f7768e")
	assertColor(t, "Status Default", c.StatusColors.Default, "#565f89")
	assertColor(t, "Status DefaultText", c.StatusColors.DefaultText, "#c0caf5")
}

func TestTokyoNightStormTheme(t *testing.T) {
	tm := GetThemeManager()
	theme, err := tm.GetTheme("tokyonight-storm")
	if err != nil {
		t.Fatalf("expected tokyonight-storm theme to exist: %v", err)
	}

	c := theme.UIColors

	assertColor(t, "Background", c.Background, "#24283b")
	assertColor(t, "Foreground", c.Foreground, "#a9b1d6")
	assertColor(t, "Selection", c.Selection, "#292e42")
	assertColor(t, "InputBackground", c.InputBackground, "#292e42")
	assertColor(t, "Success", c.Success, "#9ece6a")
	assertColor(t, "ActiveTab", c.ActiveTab, "#7aa2f7")
}

func TestGenericExtractionStillBuildsTheme(t *testing.T) {
	tm := GetThemeManager()
	theme, err := tm.GetTheme("catppuccin-mocha")
	if err != nil {
		t.Fatalf("expected catppuccin-mocha theme to exist: %v", err)
	}

	if theme.UIColors.Background == "" {
		t.Error("expected non-empty background color")
	}
	if theme.UIColors.Foreground == "" {
		t.Error("expected non-empty foreground color")
	}
	if theme.UIColors.MethodColors.GET == "" {
		t.Error("expected non-empty GET method color")
	}
}

func TestColorManagerPlaceholderColor(t *testing.T) {
	oldPlaceholder := config.C.Theme.PlaceholderColor
	defer func() { config.C.Theme.PlaceholderColor = oldPlaceholder }()

	config.C.Theme.PlaceholderColor = "#565f89"
	cm := NewColorManager()
	if got := colorToHex(cm.Placeholder); !strings.EqualFold(got, "#565f89") {
		t.Errorf("Placeholder = %q, want theme placeholderColor #565f89", got)
	}

	// Configs predating the field keep the previous hardcoded default.
	config.C.Theme.PlaceholderColor = ""
	cm = NewColorManager()
	if got := colorToHex(cm.Placeholder); !strings.EqualFold(got, "#4A5053") {
		t.Errorf("Placeholder fallback = %q, want #4A5053", got)
	}
}

func TestColorManagerInputBackgroundFromConfig(t *testing.T) {
	oldInput := config.C.Theme.InputBackgroundColor
	oldSelection := config.C.Theme.SelectionBackground
	oldTreeSelection := config.C.Theme.TreeSelectionBackground
	defer func() {
		config.C.Theme.InputBackgroundColor = oldInput
		config.C.Theme.SelectionBackground = oldSelection
		config.C.Theme.TreeSelectionBackground = oldTreeSelection
	}()

	// The InputBackground token must honor the inputBackgroundColor key, not
	// treeSelectionBackground.
	config.C.Theme.InputBackgroundColor = "#1d1f2c"
	config.C.Theme.TreeSelectionBackground = "#4a4165"
	cm := NewColorManager()
	if got := colorToHex(cm.InputBackground); !strings.EqualFold(got, "#1d1f2c") {
		t.Errorf("InputBackground = %q, want inputBackgroundColor #1d1f2c", got)
	}

	// Fallback to the selection background when inputBackgroundColor is unset.
	config.C.Theme.InputBackgroundColor = ""
	config.C.Theme.SelectionBackground = "#1f202e"
	cm = NewColorManager()
	if got := colorToHex(cm.InputBackground); !strings.EqualFold(got, "#1f202e") {
		t.Errorf("InputBackground fallback = %q, want selectionBackground #1f202e", got)
	}
}

func TestApplyThemePropagatesPlaceholderColor(t *testing.T) {
	tm := GetThemeManager()

	oldSyntaxTheme := config.C.SyntaxTheme
	oldPlaceholder := config.C.Theme.PlaceholderColor
	defer func() {
		config.C.SyntaxTheme = oldSyntaxTheme
		config.C.Theme.PlaceholderColor = oldPlaceholder
	}()

	if err := tm.ApplyTheme("tokyonight-night"); err != nil {
		t.Fatalf("failed to apply tokyonight-night: %v", err)
	}

	theme, err := tm.GetTheme("tokyonight-night")
	if err != nil {
		t.Fatalf("failed to get theme: %v", err)
	}
	if got, want := config.C.Theme.PlaceholderColor, theme.UIColors.Placeholder; !strings.EqualFold(got, want) {
		t.Errorf("config placeholderColor = %q, want %q", got, want)
	}
}

func assertColor(t *testing.T, name, got, want string) {
	t.Helper()
	if !strings.EqualFold(got, want) {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}

// TestExtractedColorsAccountForChromaOffset guards against the regression where
// Chroma colours (stored with a +1 offset) were converted without subtracting
// it. White (#ffffff) overflowed to the invalid "#1000000", which downstream
// colour parsers masked back to black, producing black text on dark
// backgrounds for themes such as vim, xcode-dark and modus-vivendi.
func TestExtractedColorsAccountForChromaOffset(t *testing.T) {
	tm := GetThemeManager()

	cases := []struct {
		theme      string
		background string
		foreground string
	}{
		{"modus-vivendi", "#000000", "#ffffff"},
		{"xcode-dark", "#1f1f24", "#ffffff"},
		{"vim", "#000000", "#cccccc"},
	}

	for _, tc := range cases {
		theme, err := tm.GetTheme(tc.theme)
		if err != nil {
			t.Fatalf("expected theme %q: %v", tc.theme, err)
		}
		assertColor(t, tc.theme+" Background", theme.UIColors.Background, tc.background)
		assertColor(t, tc.theme+" Foreground", theme.UIColors.Foreground, tc.foreground)
	}
}

// TestSupportedThemesAreReadable ensures every theme offered by the picker
// emits valid #rrggbb colours and keeps text legible against its background.
func TestSupportedThemesAreReadable(t *testing.T) {
	tm := GetThemeManager()
	hexColor := regexp.MustCompile(`^#[0-9a-f]{6}$`)

	for _, name := range getSupportedUnifiedThemes() {
		theme, err := tm.GetTheme(name)
		if err != nil {
			t.Fatalf("expected theme %q: %v", name, err)
		}
		c := theme.UIColors

		for field, color := range map[string]string{
			"Background":      c.Background,
			"Foreground":      c.Foreground,
			"Border":          c.Border,
			"Title":           c.Title,
			"Selection":       c.Selection,
			"InputBackground": c.InputBackground,
		} {
			if !hexColor.MatchString(strings.ToLower(color)) {
				t.Errorf("theme %q %s = %q, want #rrggbb", name, field, color)
			}
		}

		if ratio := contrastRatio(c.Foreground, c.Background); ratio < 3.0 {
			t.Errorf("theme %q foreground/background contrast %.2f is too low", name, ratio)
		}
		if strings.EqualFold(c.Selection, c.Background) {
			t.Errorf("theme %q selection colour matches its background, selections would be invisible", name)
		}
	}
}

// contrastRatio returns the WCAG contrast ratio between two #rrggbb colours.
func contrastRatio(a, b string) float64 {
	la, lb := relativeLuminance(a), relativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}

// relativeLuminance returns the relative luminance of a #rrggbb colour.
func relativeLuminance(hexColor string) float64 {
	r, g, b := hexToRGB(hexColor)
	channel := func(v int) float64 {
		s := float64(v) / 255
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}
