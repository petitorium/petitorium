package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/joho/godotenv"
	"github.com/rivo/tview"
)

// loadColorFromEnv loads a color from environment variable in #AABBCC format
// Falls back to defaultValue if not found or invalid
func loadColorFromEnv(key string, defaultValue int32) tcell.Color {
	hexStr := os.Getenv(key)
	if hexStr == "" {
		return tcell.NewHexColor(defaultValue)
	}

	// Remove '#' if present
	hexStr = strings.TrimPrefix(hexStr, "#")

	// Parse hex string to int64
	if colorValue, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
		return tcell.NewHexColor(int32(colorValue))
	}

	// Return default if parsing failed
	return tcell.NewHexColor(defaultValue)
}

var (
	backgroundColor  tcell.Color
	foregroundColor  tcell.Color
	borderColor      tcell.Color
	borderFocusColor tcell.Color
	titleColor       tcell.Color
)

// loadRuneFromEnv loads a rune from an environment variable.
// Falls back to defaultValue if not found or empty.
func loadRuneFromEnv(key string, defaultValue rune) rune {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	// Return the first rune of the string
	for _, r := range strValue {
		return r
	}
	return defaultValue
}

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using defaults")
	}

	backgroundColor = loadColorFromEnv("BACKGROUND_COLOR", 0x000000)
	foregroundColor = loadColorFromEnv("FOREGROUND_COLOR", 0xFFFFFF)
	borderColor = loadColorFromEnv("BORDER_COLOR", 0x888888)
	borderFocusColor = loadColorFromEnv("BORDER_FOCUS_COLOR", 0xFFFFFF)
	titleColor = loadColorFromEnv("TITLE_COLOR", 0xFFFFFF)

	// Load border characters from .env
	tview.Borders.TopLeft = loadRuneFromEnv("BORDER_TOP_LEFT", '┌')
	tview.Borders.TopRight = loadRuneFromEnv("BORDER_TOP_RIGHT", '┐')
	tview.Borders.BottomLeft = loadRuneFromEnv("BORDER_BOTTOM_LEFT", '└')
	tview.Borders.BottomRight = loadRuneFromEnv("BORDER_BOTTOM_RIGHT", '┘')
	tview.Borders.Horizontal = loadRuneFromEnv("BORDER_HORIZONTAL", '─')
	tview.Borders.Vertical = loadRuneFromEnv("BORDER_VERTICAL", '│')

	tview.Borders.TopLeftFocus = loadRuneFromEnv("BORDER_TOP_LEFT_FOCUS", '┌')
	tview.Borders.TopRightFocus = loadRuneFromEnv("BORDER_TOP_RIGHT_FOCUS", '┐')
	tview.Borders.BottomLeftFocus = loadRuneFromEnv("BORDER_BOTTOM_LEFT_FOCUS", '└')
	tview.Borders.BottomRightFocus = loadRuneFromEnv("BORDER_BOTTOM_RIGHT_FOCUS", '┘')
	tview.Borders.HorizontalFocus = loadRuneFromEnv("BORDER_HORIZONTAL_FOCUS", '─')
	tview.Borders.VerticalFocus = loadRuneFromEnv("BORDER_VERTICAL_FOCUS", '│')
}

func newPanel(title string) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(title)

	tv.SetBackgroundColor(backgroundColor)
	tv.SetBorderColor(borderColor)
	tv.SetTitleColor(titleColor)
	tv.SetTextColor(foregroundColor)
	return tv
}

func main() {
	unfocusedBorderStyle := tcell.StyleDefault.Foreground(borderColor)
	focusedBorderStyle := tcell.StyleDefault.Foreground(borderFocusColor)

	setFocusStyle := func(p *tview.TextView, focused bool) {
		if focused {
			p.SetBorderStyle(focusedBorderStyle)
			p.SetBackgroundColor(backgroundColor)
		} else {
			p.SetBorderStyle(unfocusedBorderStyle)
			p.SetBackgroundColor(backgroundColor)
		}
	}

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
	grid.AddItem(collections, 1, 0, 1, 1, 0, 0, true)

	// Add the right-side Flex layout (request/response) to the grid.
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	// --- Focus and Navigation ---

	// A slice of the panels that can be focused.
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
	if err := app.SetRoot(grid, true).SetFocus(collections).Run(); err != nil {
		panic(err)
	}
}
