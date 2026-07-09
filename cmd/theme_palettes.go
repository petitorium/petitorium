package cmd

// ThemePalette defines the canonical terminal/ansi-like colors for a theme.
// It is the single source of truth from which all UI colors are derived.
// Keeping palettes as editable Go values makes the theme configurable in code
// and lays the groundwork for loading palettes from configuration later.
type ThemePalette struct {
	Name        string
	SyntaxTheme string

	Background  string // editor background
	Foreground  string // editor foreground
	SecondaryBg string // lighter background used for selections and inputs

	Black     string // terminal black, used for subtle UI surfaces
	Red       string // terminal red
	Orange    string
	Yellow    string
	Green     string
	Teal      string
	Cyan      string
	LightBlue string
	Blue      string
	Magenta   string
	White     string // terminal white

	Comment string
}

// buildThemeFromPalette maps a canonical palette to the full ThemeColors
// used by the application. All themes (hand-tuned or extracted) flow through
// this function so the UI mapping is consistent and easy to tweak in one place.
func buildThemeFromPalette(p ThemePalette) *UnifiedTheme {
	background := p.Background
	if background == "" {
		background = "#1a1b26"
	}
	foreground := p.Foreground
	if foreground == "" {
		foreground = "#a9b1d6"
	}
	secondaryBg := p.SecondaryBg
	if secondaryBg == "" {
		secondaryBg = adjustBrightness(background, 1.2)
	}

	black := fallbackColor(p.Black, "#414868")
	red := fallbackColor(p.Red, "#f7768e")
	orange := fallbackColor(p.Orange, "#ff9e64")
	yellow := fallbackColor(p.Yellow, "#e0af68")
	green := fallbackColor(p.Green, "#9ece6a")
	teal := fallbackColor(p.Teal, "#73daca")
	blue := fallbackColor(p.Blue, "#7aa2f7")
	magenta := fallbackColor(p.Magenta, "#bb9af7")
	white := fallbackColor(p.White, "#c0caf5")
	comment := fallbackColor(p.Comment, "#565f89")

	return &UnifiedTheme{
		Name:        p.Name,
		SyntaxTheme: p.SyntaxTheme,
		UIColors: ThemeColors{
			Background:             background,
			Foreground:             foreground,
			Border:                 black,
			BorderFocus:            blue,
			Title:                  white,
			Selection:              secondaryBg,
			TreeSelection:          adjustBrightness(secondaryBg, 1.05),
			SelectedBackground:     blue,
			SelectedForeground:     background,
			ActiveTab:              blue,
			ButtonBackground:       black,
			ButtonSelected:         magenta,
			DropdownFocused:        secondaryBg,
			Placeholder:            comment,
			InputBackground:        secondaryBg,
			InputBackgroundLighter: adjustBrightness(secondaryBg, 1.1),
			Success:                green,
			Error:                  red,
			Warning:                yellow,
			SelectedRequestIcon:    green,
			LabelColor:             magenta,
			ValueColor:             white,
			MethodColors: MethodColors{
				GET:     green,
				POST:    blue,
				PUT:     yellow,
				PATCH:   orange,
				DELETE:  red,
				OPTIONS: comment,
				HEAD:    teal,
				Default: white,
			},
			StatusColors: StatusColors{
				Success:         green,
				SuccessText:     background,
				Redirection:     yellow,
				RedirectionText: background,
				ClientError:     red,
				ClientErrorText: background,
				ServerError:     adjustBrightness(red, 0.65),
				ServerErrorText: white,
				Default:         comment,
				DefaultText:     white,
			},
		},
	}
}

// fallbackColor returns color if non-empty, otherwise the supplied fallback.
func fallbackColor(color, fallback string) string {
	if color == "" {
		return fallback
	}
	return color
}

// tokyoNightNightPalette is the canonical Tokyo Night "Night" palette from
// https://github.com/tokyo-night/tokyo-night-vscode-theme
var tokyoNightNightPalette = ThemePalette{
	Name:        "tokyonight-night",
	SyntaxTheme: "tokyonight-night",
	Background:  "#1a1b26",
	Foreground:  "#a9b1d6",
	SecondaryBg: "#24283b",
	Black:       "#414868",
	Red:         "#f7768e",
	Orange:      "#ff9e64",
	Yellow:      "#e0af68",
	Green:       "#9ece6a",
	Teal:        "#73daca",
	Cyan:        "#b4f9f8",
	LightBlue:   "#7dcfff",
	Blue:        "#7aa2f7",
	Magenta:     "#bb9af7",
	White:       "#c0caf5",
	Comment:     "#565f89",
}

// tokyoNightStormPalette is the canonical Tokyo Night "Storm" palette.
// It shares the same accent colors as Night but uses the lighter Storm background.
var tokyoNightStormPalette = ThemePalette{
	Name:        "tokyonight-storm",
	SyntaxTheme: "tokyonight-storm",
	Background:  "#24283b",
	Foreground:  "#a9b1d6",
	SecondaryBg: "#292e42",
	Black:       "#414868",
	Red:         "#f7768e",
	Orange:      "#ff9e64",
	Yellow:      "#e0af68",
	Green:       "#9ece6a",
	Teal:        "#73daca",
	Cyan:        "#b4f9f8",
	LightBlue:   "#7dcfff",
	Blue:        "#7aa2f7",
	Magenta:     "#bb9af7",
	White:       "#c0caf5",
	Comment:     "#565f89",
}
