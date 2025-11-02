package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
)

// ColorManager centralizes all color management for the application
type ColorManager struct {
	// Theme colors from configuration
	Background    tcell.Color
	Foreground    tcell.Color
	Border        tcell.Color
	BorderFocus   tcell.Color
	Title         tcell.Color
	Selection     tcell.Color
	ActiveTab     tcell.Color
	ButtonSelect  tcell.Color
	DropdownFocus tcell.Color

	// Semantic colors for consistent usage across the app
	Placeholder tcell.Color // Placeholder text color
	Success     tcell.Color // Success status color
	Error       tcell.Color // Error status color
	Warning     tcell.Color // Warning status color
}

// NewColorManager creates a new ColorManager with colors from the current theme
func NewColorManager() *ColorManager {
	theme := config.C.Theme

	// Set up tview borders from theme configuration
	setupBorders(theme)

	return &ColorManager{
		Background:    hexToColor(theme.BackgroundColor),
		Foreground:    hexToColor(theme.ForegroundColor),
		Border:        hexToColor(theme.BorderColor),
		BorderFocus:   hexToColor(theme.BorderFocusColor),
		Title:         hexToColor(theme.TitleColor),
		Selection:     hexToColor(theme.SelectionBackground),
		ActiveTab:     hexToColor(theme.ActiveTabColor),
		ButtonSelect:  hexToColor(theme.ButtonSelectedColor),
		DropdownFocus: hexToColor(theme.DropdownFocusedBackground),
		Placeholder:   hexToColor("#4A5053"),
		Success:       hexToColor("#28a745"),
		Error:         hexToColor("#dc3545"),
		Warning:       hexToColor("#fd7e14"),
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
