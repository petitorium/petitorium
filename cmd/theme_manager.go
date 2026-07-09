package cmd

import (
	"fmt"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"

	"github.com/petitorium/petitorium/config"
)

// UnifiedTheme represents a complete theme with both syntax highlighting and UI colors
type UnifiedTheme struct {
	Name        string
	SyntaxTheme string
	UIColors    ThemeColors
}

// ThemeColors contains all UI color definitions
type ThemeColors struct {
	Background             string
	Foreground             string
	Border                 string
	BorderFocus            string
	Title                  string
	Selection              string
	TreeSelection          string
	SelectedBackground     string
	SelectedForeground     string
	ActiveTab              string
	ButtonBackground       string
	ButtonSelected         string
	DropdownFocused        string
	Placeholder            string
	InputBackground        string
	InputBackgroundLighter string
	Success                string
	Error                  string
	Warning                string
	SelectedRequestIcon    string
	LabelColor             string
	ValueColor             string
	MethodColors           MethodColors
	StatusColors           StatusColors
}

// MethodColors contains HTTP method-specific colors
type MethodColors struct {
	GET     string
	POST    string
	PUT     string
	PATCH   string
	DELETE  string
	OPTIONS string
	HEAD    string
	Default string
}

// StatusColors contains HTTP status code colors
type StatusColors struct {
	Success         string
	SuccessText     string
	Redirection     string
	RedirectionText string
	ClientError     string
	ClientErrorText string
	ServerError     string
	ServerErrorText string
	Default         string
	DefaultText     string
}

// ThemeManager manages unified themes and color extraction
type ThemeManager struct {
	themes       map[string]*UnifiedTheme
	currentTheme string
	colorManager *ColorManager
}

var themeManager *ThemeManager

// GetThemeManager returns the singleton theme manager instance
func GetThemeManager() *ThemeManager {
	if themeManager == nil {
		themeManager = &ThemeManager{
			themes:       make(map[string]*UnifiedTheme),
			currentTheme: config.C.SyntaxTheme,
		}
		themeManager.initializeThemes()
	}
	return themeManager
}

// initializeThemes sets up all predefined unified themes
func (tm *ThemeManager) initializeThemes() {
	// Hand-tuned themes use a canonical palette so the full color range is
	// represented in the UI instead of collapsing into a few derived tones.
	tm.themes["tokyonight-night"] = buildThemeFromPalette(tokyoNightNightPalette)
	tm.themes["tokyonight-storm"] = buildThemeFromPalette(tokyoNightStormPalette)

	// Other themes are extracted from their Chroma syntax definition.
	tm.themes["github-dark"] = tm.extractThemeFromChroma("github-dark", "#0d1117")
	tm.themes["dracula"] = tm.extractThemeFromChroma("dracula", "#282a36")
	tm.themes["monokai"] = tm.extractThemeFromChroma("monokai", "#272822")
	tm.themes["solarized-dark"] = tm.extractThemeFromChroma("solarized-dark", "#002b36")
	tm.themes["nord"] = tm.extractThemeFromChroma("nord", "#2e3440")
	tm.themes["one-dark"] = tm.extractThemeFromChroma("one-dark", "#282c34")
	tm.themes["vim"] = tm.extractThemeFromChroma("vim", "#1c1c1c")
	tm.themes["gruvbox"] = tm.extractThemeFromChroma("gruvbox", "#282828")
	tm.themes["catppuccin-mocha"] = tm.extractThemeFromChroma("catppuccin-mocha", "#1e1e2e")
	tm.themes["evergarden"] = tm.extractThemeFromChroma("evergarden", "#1a1b26")
	tm.themes["doom-one"] = tm.extractThemeFromChroma("doom-one", "#282c34")
	tm.themes["rose-pine-moon"] = tm.extractThemeFromChroma("rose-pine-moon", "#1f1d2e")
}

// extractThemeFromChroma extracts colors from a Chroma syntax theme and creates a unified theme
func (tm *ThemeManager) extractThemeFromChroma(themeName, fallbackBackground string) *UnifiedTheme {
	style := styles.Get(themeName)
	if style == nil {
		// Return a default theme if the syntax theme doesn't exist
		return tm.createDefaultTheme(themeName, fallbackBackground)
	}

	// Extract colors from the syntax theme
	colors := tm.extractColorsFromStyle(style)

	return &UnifiedTheme{
		Name:        themeName,
		SyntaxTheme: themeName,
		UIColors:    colors,
	}
}

// extractColorsFromStyle extracts UI colors from a Chroma style by translating
// the style tokens into a ThemePalette and then using the shared palette builder.
func (tm *ThemeManager) extractColorsFromStyle(style *chroma.Style) ThemeColors {
	palette := tm.paletteFromStyle(style)
	return buildThemeFromPalette(palette).UIColors
}

// paletteFromStyle converts a Chroma syntax style into a ThemePalette.
// It extracts the most common tokens and fills the remaining palette slots with
// sensible fallbacks so every theme benefits from the full UI color range.
func (tm *ThemeManager) paletteFromStyle(style *chroma.Style) ThemePalette {
	background := tm.colorToHex(style.Get(chroma.Background).Background)
	if background == "" {
		background = "#1a1b26"
	}

	foreground := tm.colorToHex(style.Get(chroma.Text).Colour)
	if foreground == "" {
		foreground = "#e4e4e4"
	}

	keyword := tm.colorToHex(style.Get(chroma.Keyword).Colour)
	if keyword == "" {
		keyword = "#7aa2f7"
	}

	stringColor := tm.colorToHex(style.Get(chroma.String).Colour)
	if stringColor == "" {
		stringColor = "#9ece6a"
	}

	commentColor := tm.colorToHex(style.Get(chroma.Comment).Colour)
	if commentColor == "" {
		commentColor = "#565f89"
	}

	numberColor := tm.colorToHex(style.Get(chroma.Number).Colour)
	if numberColor == "" {
		numberColor = "#ff9e64"
	}

	errorEntry := style.Get(chroma.GenericError)
	if errorEntry.Colour == 0 {
		errorEntry = style.Get(chroma.Error)
	}
	errorColor := tm.colorToHex(errorEntry.Colour)
	if errorColor == "" {
		errorColor = "#fb4f49"
	}

	// Optional richer tokens; when absent we fall back to the main accents.
	functionColor := tm.colorToHex(style.Get(chroma.NameFunction).Colour)
	classColor := tm.colorToHex(style.Get(chroma.NameClass).Colour)
	operatorColor := tm.colorToHex(style.Get(chroma.Operator).Colour)
	attributeColor := tm.colorToHex(style.Get(chroma.NameAttribute).Colour)
	literalColor := tm.colorToHex(style.Get(chroma.Literal).Colour)

	return ThemePalette{
		Name:        style.Name,
		SyntaxTheme: style.Name,
		Background:  background,
		Foreground:  foreground,
		SecondaryBg: tm.adjustBrightness(background, 1.2),
		Black:       commentColor,
		Red:         errorColor,
		Orange:      coalesce(classColor, numberColor),
		Yellow:      numberColor,
		Green:       stringColor,
		Teal:        coalesce(operatorColor, stringColor),
		Cyan:        coalesce(literalColor, keyword),
		LightBlue:   coalesce(attributeColor, keyword),
		Blue:        coalesce(functionColor, keyword),
		Magenta:     keyword,
		White:       foreground,
		Comment:     commentColor,
	}
}

// coalesce returns the first non-empty color string.
func coalesce(colors ...string) string {
	for _, c := range colors {
		if c != "" {
			return c
		}
	}
	return ""
}

// createDefaultTheme creates a default theme when extraction fails
func (tm *ThemeManager) createDefaultTheme(themeName, background string) *UnifiedTheme {
	return &UnifiedTheme{
		Name:        themeName,
		SyntaxTheme: themeName,
		UIColors: ThemeColors{
			Background:             background,
			Foreground:             "#e4e4e4",
			Border:                 "#95ceda",
			BorderFocus:            "#ff9f77",
			Title:                  "#ebebeb",
			Selection:              "#1b4248",
			TreeSelection:          "#7aa2f7",
			SelectedBackground:     "#28a745",
			SelectedForeground:     background,
			ActiveTab:              "#ff9f77",
			ButtonBackground:       "#95ceda",
			ButtonSelected:         "#ffd700",
			DropdownFocused:        "#636da6",
			Placeholder:            "#4a5053",
			InputBackground:        "#2e3440",
			InputBackgroundLighter: "#3b4252",
			Success:                "#28a745",
			Error:                  "#dc3545",
			Warning:                "#fd7e14",
			SelectedRequestIcon:    "#28a745", // tm.getSelectedRequestIconColor(themeName, "#28a745"),
			MethodColors: MethodColors{
				GET:     "#6ea5a0",
				POST:    "#ff00ff",
				PUT:     "#ff9f77",
				PATCH:   "#ff9f77",
				DELETE:  "#fb4f49",
				OPTIONS: "#ffa500",
				HEAD:    "#800080",
				Default: "#888888",
			},
			StatusColors: StatusColors{
				Success:         "#28a745",
				SuccessText:     "#ffffff",
				Redirection:     "#fd7e14",
				RedirectionText: "#ffffff",
				ClientError:     "#dc3545",
				ClientErrorText: "#ffffff",
				ServerError:     "#8b0000",
				ServerErrorText: "#ffffff",
				Default:         "#636da6",
				DefaultText:     "#ffffff",
			},
		},
	}
}

// colorToHex converts a chroma color to hex string
func (tm *ThemeManager) colorToHex(color chroma.Colour) string {
	if color == 0 {
		return ""
	}
	return fmt.Sprintf("#%06x", int(color))
}

// adjustBrightness adjusts the brightness of a hex color
func (tm *ThemeManager) adjustBrightness(hexColor string, factor float64) string {
	if hexColor == "" {
		return ""
	}

	// Remove # if present
	hexColor = strings.TrimPrefix(hexColor, "#")

	// Parse hex color
	var r, g, b int
	fmt.Sscanf(hexColor, "%02x%02x%02x", &r, &g, &b)

	// Adjust brightness
	r = int(float64(r) * factor)
	g = int(float64(g) * factor)
	b = int(float64(b) * factor)

	// Clamp values
	if r > 255 {
		r = 255
	}
	if g > 255 {
		g = 255
	}
	if b > 255 {
		b = 255
	}
	if r < 0 {
		r = 0
	}
	if g < 0 {
		g = 0
	}
	if b < 0 {
		b = 0
	}

	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// getSelectedRequestIconColor returns the appropriate color for the selected request icon based on theme
// func (tm *ThemeManager) getSelectedRequestIconColor(themeName, defaultColor string) string {
// 	switch themeName {
// 	case "tokyonight-night":
// 		return "#8DE328"
// 	default:
// 	return defaultColor // Use the string color as default
// 	}
// }

// GetTheme returns a unified theme by name
func (tm *ThemeManager) GetTheme(themeName string) (*UnifiedTheme, error) {
	if theme, exists := tm.themes[themeName]; exists {
		return theme, nil
	}

	// Try to create a theme dynamically if it doesn't exist
	if styles.Get(themeName) != nil {
		theme := tm.extractThemeFromChroma(themeName, "#1a1b26")
		tm.themes[themeName] = theme
		return theme, nil
	}

	return nil, fmt.Errorf("theme '%s' not found", themeName)
}

// GetAvailableThemes returns all available unified themes
func (tm *ThemeManager) GetAvailableThemes() []string {
	themes := make([]string, 0, len(tm.themes))
	for name := range tm.themes {
		themes = append(themes, name)
	}
	return themes
}

// ApplyTheme applies a unified theme to the application
func (tm *ThemeManager) ApplyTheme(themeName string) error {
	theme, err := tm.GetTheme(themeName)
	if err != nil {
		return err
	}

	// Update current theme
	tm.currentTheme = themeName

	// Update config syntax theme only
	config.C.SyntaxTheme = themeName

	// Update base theme colors
	config.C.Theme.BackgroundColor = theme.UIColors.Background
	config.C.Theme.ForegroundColor = theme.UIColors.Foreground
	config.C.Theme.BorderColor = theme.UIColors.Border
	config.C.Theme.BorderFocusColor = theme.UIColors.BorderFocus
	config.C.Theme.TitleColor = theme.UIColors.Title
	config.C.Theme.SelectionBackground = theme.UIColors.Selection
	config.C.Theme.TreeSelectionBackground = theme.UIColors.TreeSelection
	config.C.Theme.SelectedBackground = theme.UIColors.SelectedBackground
	config.C.Theme.SelectedForeground = theme.UIColors.SelectedForeground
	config.C.Theme.ActiveTabColor = theme.UIColors.ActiveTab
	config.C.Theme.ButtonBackgroundColor = theme.UIColors.ButtonBackground
	config.C.Theme.ButtonSelectedColor = theme.UIColors.ButtonSelected
	config.C.Theme.DropdownFocusedBackground = theme.UIColors.DropdownFocused
	config.C.Theme.InputBackgroundColor = theme.UIColors.InputBackground
	config.C.Theme.InputBackgroundLighterColor = theme.UIColors.InputBackgroundLighter
	config.C.Theme.LabelColor = theme.UIColors.LabelColor
	config.C.Theme.ValueColor = theme.UIColors.ValueColor
	config.C.UI.SelectedRequestIconColor = theme.UIColors.SelectedRequestIcon

	// Update method colors
	config.C.MethodColors.GET = theme.UIColors.MethodColors.GET
	config.C.MethodColors.POST = theme.UIColors.MethodColors.POST
	config.C.MethodColors.PUT = theme.UIColors.MethodColors.PUT
	config.C.MethodColors.PATCH = theme.UIColors.MethodColors.PATCH
	config.C.MethodColors.DELETE = theme.UIColors.MethodColors.DELETE
	config.C.MethodColors.OPTIONS = theme.UIColors.MethodColors.OPTIONS
	config.C.MethodColors.HEAD = theme.UIColors.MethodColors.HEAD
	config.C.MethodColors.Default = theme.UIColors.MethodColors.Default

	// Update status colors
	config.C.StatusColors.Success = theme.UIColors.StatusColors.Success
	config.C.StatusColors.SuccessText = theme.UIColors.StatusColors.SuccessText
	config.C.StatusColors.Redirection = theme.UIColors.StatusColors.Redirection
	config.C.StatusColors.RedirectionText = theme.UIColors.StatusColors.RedirectionText
	config.C.StatusColors.ClientError = theme.UIColors.StatusColors.ClientError
	config.C.StatusColors.ClientErrorText = theme.UIColors.StatusColors.ClientErrorText
	config.C.StatusColors.ServerError = theme.UIColors.StatusColors.ServerError
	config.C.StatusColors.ServerErrorText = theme.UIColors.StatusColors.ServerErrorText
	config.C.StatusColors.Default = theme.UIColors.StatusColors.Default
	config.C.StatusColors.DefaultText = theme.UIColors.StatusColors.DefaultText

	// Recreate color manager with new colors
	tm.colorManager = NewColorManager()

	return nil
}

// GetCurrentTheme returns the currently active theme
func (tm *ThemeManager) GetCurrentTheme() string {
	return tm.currentTheme
}

// GetColorManager returns the current color manager
func (tm *ThemeManager) GetColorManager() *ColorManager {
	if tm.colorManager == nil {
		tm.colorManager = NewColorManager()
	}
	return tm.colorManager
}
