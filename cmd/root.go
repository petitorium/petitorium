package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/hbarral/petitorium/config"
)

var rootCmd = &cobra.Command{
	Use:   "petitorium",
	Short: "A powerful TUI for API interaction and testing.",
	Run:   runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := config.LoadConfig(); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	})
}

func hexToColor(hexStr string) tcell.Color {
	hexStr = strings.TrimPrefix(hexStr, "#")
	if colorValue, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
		return tcell.NewHexColor(int32(colorValue))
	}

	return tcell.ColorWhite
}

func strToRune(s string) rune {
	for _, r := range s {
		return r
	}
	return ' '
}

func runTUI(cmd *cobra.Command, args []string) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	borderColor := hexToColor(theme.BorderColor)
	borderFocusColor := hexToColor(theme.BorderFocusColor)
	titleColor := hexToColor(theme.TitleColor)

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

	newPanel := func(title string) *tview.TextView {
		tv := tview.NewTextView()
		tv.SetBorder(true)
		tv.SetTitle(title)
		tv.SetBackgroundColor(backgroundColor)
		tv.SetBorderColor(borderColor)
		tv.SetTitleColor(titleColor)
		tv.SetTextColor(foregroundColor)
		return tv
	}

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

	app := tview.NewApplication().
		EnableMouse(true)

	header := newPanel(" Petitorium ")
	collections := newPanel(" Collections ")
	request := newPanel(" Request ")
	response := newPanel(" Response ")
	footer := newPanel("")
	footer.SetText(" [Tab] Cycle Focus | [Enter] Select | [Q] Quit ")

	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(request, 0, 1, false).
		AddItem(response, 0, 1, false)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(collections, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	currentFocus := 0
	panels := []*tview.TextView{collections, request, response}

	setFocusStyle(panels[currentFocus], true)

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

		return event
	})

	if err := app.SetRoot(grid, true).SetFocus(collections).Run(); err != nil {
		panic(err)
	}
}
