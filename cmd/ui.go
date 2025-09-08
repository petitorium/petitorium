package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
)

// setupTheme configures the theme colors and borders
func setupTheme() (tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	borderColor := hexToColor(theme.BorderColor)
	borderFocusColor := hexToColor(theme.BorderFocusColor)
	titleColor := hexToColor(theme.TitleColor)
	selectionBackgroundColor := hexToColor(theme.SelectionBackground)

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

	return backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor
}

// createPanel creates a new text view panel with consistent styling
func createPanel(title string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(title)
	tv.SetBackgroundColor(backgroundColor)
	tv.SetBorderColor(borderColor)
	tv.SetTitleColor(titleColor)
	tv.SetTextColor(foregroundColor)
	tv.SetBorderPadding(0, 0, 0, 0)
	return tv
}

// setFocusStyle sets the border color based on focus state
func setFocusStyle(p tview.Primitive, focused bool, borderColor, borderFocusColor tcell.Color) {
	type borderStyler interface {
		SetBorderColor(tcell.Color) *tview.Box
	}
	styler, ok := p.(borderStyler)
	if !ok {
		return
	}
	if focused {
		styler.SetBorderColor(borderFocusColor)
	} else {
		styler.SetBorderColor(borderColor)
	}
}

// createModal creates a centered modal dialog
func createModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}
