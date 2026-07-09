package cmd

import (
	"strings"
	"testing"
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

func assertColor(t *testing.T, name, got, want string) {
	t.Helper()
	if !strings.EqualFold(got, want) {
		t.Errorf("%s = %q, want %q", name, got, want)
	}
}
