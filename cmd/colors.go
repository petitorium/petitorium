package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
)

// ColorManager centralizes all color management for the application
type ColorManager struct {
	// Theme colors from configuration
	Background       tcell.Color
	Foreground       tcell.Color
	Border           tcell.Color
	BorderFocus      tcell.Color
	Title            tcell.Color
	Selection        tcell.Color
	TreeSelection    tcell.Color
	ActiveTab        tcell.Color
	ButtonBackground tcell.Color // Button background color (normal state)
	ButtonSelect     tcell.Color
	DropdownFocus    tcell.Color
	InputBackground  tcell.Color // Input field background color (subtle contrast)

	// Semantic colors for consistent usage across the app
	Placeholder tcell.Color // Placeholder text color
	Success     tcell.Color // Success status color
	Error       tcell.Color // Error status color
	Warning     tcell.Color // Warning status color
	LabelColor  tcell.Color // Color for form labels (e.g., "Name:", "Type:", "Value:")
	ValueColor  tcell.Color // Color for form values/input text

	// Status colors for response panel
	StatusSuccessBg     tcell.Color
	StatusSuccessFg     tcell.Color
	StatusRedirectBg    tcell.Color
	StatusRedirectFg    tcell.Color
	StatusClientErrorBg tcell.Color
	StatusClientErrorFg tcell.Color
	StatusServerErrorBg tcell.Color
	StatusServerErrorFg tcell.Color
	StatusDefaultBg     tcell.Color
	StatusDefaultFg     tcell.Color
}

// NewColorManager creates a new ColorManager with colors from the current theme
func NewColorManager() *ColorManager {
	theme := config.C.Theme
	status := config.C.StatusColors

	// Set up tview borders from theme configuration
	setupBorders(theme)

	// Set input background color with fallback
	inputBackground := theme.InputBackgroundColor
	if inputBackground == "" {
		inputBackground = theme.BackgroundColor // Default to background (no contrast)
	}

	// Set label and value colors with fallbacks
	labelColor := theme.LabelColor
	if labelColor == "" {
		labelColor = "#4A5053" // Default placeholder color
	}
	valueColor := theme.ValueColor
	if valueColor == "" {
		valueColor = theme.ForegroundColor // Default to foreground color
	}

	return &ColorManager{
		Background:          hexToColor(theme.BackgroundColor),
		Foreground:          hexToColor(theme.ForegroundColor),
		Border:              hexToColor(theme.BorderColor),
		BorderFocus:         hexToColor(theme.BorderFocusColor),
		Title:               hexToColor(theme.TitleColor),
		Selection:           hexToColor(theme.SelectionBackground),
		TreeSelection:       hexToColor(theme.TreeSelectionBackground),
		ActiveTab:           hexToColor(theme.ActiveTabColor),
		ButtonBackground:    hexToColor(theme.ButtonBackgroundColor),
		ButtonSelect:        hexToColor(theme.ButtonSelectedColor),
		DropdownFocus:       hexToColor(theme.DropdownFocusedBackground),
		InputBackground:     hexToColor(inputBackground),
		Placeholder:         hexToColor("#4A5053"),
		Success:             hexToColor(status.Success),
		Error:               hexToColor(status.ClientError),
		Warning:             hexToColor(status.Redirection),
		LabelColor:          hexToColor(labelColor),
		ValueColor:          hexToColor(valueColor),
		StatusSuccessBg:     hexToColor(status.Success),
		StatusSuccessFg:     hexToColor(status.SuccessText),
		StatusRedirectBg:    hexToColor(status.Redirection),
		StatusRedirectFg:    hexToColor(status.RedirectionText),
		StatusClientErrorBg: hexToColor(status.ClientError),
		StatusClientErrorFg: hexToColor(status.ClientErrorText),
		StatusServerErrorBg: hexToColor(status.ServerError),
		StatusServerErrorFg: hexToColor(status.ServerErrorText),
		StatusDefaultBg:     hexToColor(status.Default),
		StatusDefaultFg:     hexToColor(status.DefaultText),
	}
}

// setupBorders configures tview border characters from theme
func setupBorders(theme config.ThemeConfig) {
	tview.Borders.TopLeft = strToRune(theme.Borders.TopLeft)
	tview.Borders.TopRight = strToRune(theme.Borders.TopRight)
	tview.Borders.BottomLeft = strToRune(theme.Borders.BottomLeft)
	tview.Borders.BottomRight = strToRune(theme.Borders.BottomRight)
	tview.Borders.Horizontal = strToRune(theme.Borders.Horizontal)
	tview.Borders.Vertical = strToRune(theme.Borders.Vertical)

	tview.Borders.TopLeftFocus = strToRune(theme.BordersFocus.TopLeft)
	tview.Borders.TopRightFocus = strToRune(theme.BordersFocus.TopRight)
	tview.Borders.BottomLeftFocus = strToRune(theme.BordersFocus.BottomLeft)
	tview.Borders.BottomRightFocus = strToRune(theme.BordersFocus.BottomRight)
	tview.Borders.HorizontalFocus = strToRune(theme.BordersFocus.Horizontal)
	tview.Borders.VerticalFocus = strToRune(theme.BordersFocus.Vertical)
}
