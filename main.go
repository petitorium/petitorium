package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// selected line
// 0x1B4248

// background
// 0x102529

// border
// 0x95CEDA

// focus border
// 0xFB4F49
// 0xFF9F77

// title
// 0xEBEBEB

// Add these color definitions after your import statements.
var (
	//
	terafoxBg     = tcell.NewHexColor(0x102529)
	terafoxFg     = tcell.NewHexColor(0xe4e4e4)
	terafoxBorder = tcell.NewHexColor(0x95CEDA)
	terafoxTitle  = tcell.NewHexColor(0xEBEBEB)
	terafoxFocus  = tcell.NewHexColor(0xFF9F77)
)

// newPanel now creates a panel with a title on its border, but with empty content.
func newPanel(title string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(title)
	// Apply terafox theme
	tv.SetBackgroundColor(terafoxBg)
	tv.SetBorderColor(terafoxBorder)
	tv.SetTitleColor(terafoxTitle)
	tv.SetTextColor(terafoxFg)
	return tv
}

func main() {

	// Define styles for focused and unfocused borders, without extra attributes.
	unfocusedBorderStyle := tcell.StyleDefault.Foreground(terafoxBorder)
	focusedBorderStyle := tcell.StyleDefault.Foreground(terafoxFocus)

	// Define styles that explicitly remove the bold attribute.
	// unfocusedBorderStyle := tcell.StyleDefault.
	// 	Foreground(terafoxBorder).
	// 	Attributes(tcell.AttrBold)
	// focusedBorderStyle := tcell.StyleDefault.
	// 	Foreground(terafoxFocus).
	// 	Attributes(tcell.AttrBold)

	setFocusStyle := func(p *tview.TextView, focused bool) {
		if focused {
			p.SetBorderStyle(focusedBorderStyle)
			// p.SetTitleColor(terafoxFocus)
		} else {
			p.SetBorderStyle(unfocusedBorderStyle)
			// p.SetTitleColor(terafoxTitle)
		}
	}

	tview.Borders.TopLeft = '┌'
	tview.Borders.TopRight = '┐'
	tview.Borders.BottomLeft = '└'
	tview.Borders.BottomRight = '┘'
	tview.Borders.Horizontal = '─'
	tview.Borders.Vertical = '│'

	tview.Borders.TopLeftFocus = '┌'
	tview.Borders.TopRightFocus = '┐'
	tview.Borders.BottomLeftFocus = '└'
	tview.Borders.BottomRightFocus = '┘'
	tview.Borders.HorizontalFocus = '─'
	tview.Borders.VerticalFocus = '│'

	// Use a tview.Application.
	app := tview.NewApplication()

	// Create the main panels of the layout.
	header := newPanel(" Petitorium ")
	collections := newPanel(" Collections ")
	request := newPanel(" Request ")
	response := newPanel(" Response ")
	footer := newPanel("")
	footer.SetText(" [Tab] Cycle Focus | [Enter] Select | [Q] Quit ")

	// Create a Flex layout for the right-hand side (Request and Response).
	// This will arrange them vertically.
	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(request, 0, 1, false). // 0 means proportional, 1 is the proportion
		AddItem(response, 0, 1, false) // false means it's not focused initially

	// Create the main Grid layout.
	grid := tview.NewGrid().
		// Set the size of the rows and columns.
		// Rows:
		// - 3 lines for the header
		// - 0 means the main content takes up the rest of the space
		// - 3 lines for the footer
		SetRows(3, 0, 3).
		// Columns:
		// - 30 columns for the collections panel
		// - 0 means the right side takes up the rest of the space
		SetColumns(30, 0).
		SetBorders(false)

	// Add the widgets to the grid.
	// grid.AddItem(item, row, col, rowSpan, colSpan, minGridHeight, minGridWidth, focus)

	// Header spans all 2 columns.
	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)

	// Footer spans all 2 columns.
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)

	// Add collections panel on the left.
	grid.AddItem(collections, 1, 0, 1, 1, 0, 0, true) // Initially focused

	// Add the right-side Flex layout (request/response) to the grid.
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	// --- Focus and Navigation ---

	// A slice of the panels that can be focused.
	// panels := []tview.Primitive{collections, request, response}
	// currentPanel := 0

	currentFocus := 0
	panels := []*tview.TextView{collections, request, response}

	// Set initial focus style.
	setFocusStyle(panels[currentFocus], true)

	// Set a function to be called on every input event.
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Handle 'q' to quit.
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			app.Stop()
			return nil
		}

		// Handle 'Tab' to cycle focus.
		if event.Key() == tcell.KeyTab {
			setFocusStyle(panels[currentFocus], false)
			currentFocus = (currentFocus + 1) % len(panels)
			app.SetFocus(panels[currentFocus])
			setFocusStyle(panels[currentFocus], true)
			return nil
		}

		// Handle 'Shift+Tab' to cycle focus backwards.
		if event.Key() == tcell.KeyBacktab {
			setFocusStyle(panels[currentFocus], false)
			currentFocus = (currentFocus - 1 + len(panels)) % len(panels)
			app.SetFocus(panels[currentFocus])
			setFocusStyle(panels[currentFocus], true)
			return nil
		}

		// If we don't handle the key, return it to be processed by the focused widget.
		return event
	})

	// Set the grid as the root widget and run the application.
	grid.SetBackgroundColor(terafoxBg)

	// Set the grid as the root widget and run the application.
	if err := app.SetRoot(grid, true).SetFocus(collections).Run(); err != nil {
		panic(err)
	}
}
