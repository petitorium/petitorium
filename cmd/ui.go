package cmd

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
)

// setupTheme configures the theme colors and borders
func setupTheme() (tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	borderColor := hexToColor(theme.BorderColor)
	borderFocusColor := hexToColor(theme.BorderFocusColor)
	titleColor := hexToColor(theme.TitleColor)
	selectionBackgroundColor := hexToColor(theme.SelectionBackground)
	activeTabColor := hexToColor(theme.ActiveTabColor)

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

	return backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor, activeTabColor
}

// createPanel creates a new text view panel with consistent styling and 16m color support
func createPanel(title string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	tv := tview.NewTextView()
	tv.SetBorder(true)
	tv.SetTitle(title)
	tv.SetBackgroundColor(backgroundColor)
	tv.SetBorderColor(borderColor)
	tv.SetTitleColor(titleColor)
	tv.SetTextColor(foregroundColor)
	tv.SetBorderPadding(0, 0, 0, 0)
	tv.SetDynamicColors(true) // Enable color interpretation with hex support
	tv.SetRegions(true)       // Enable regions for better color handling
	tv.SetWordWrap(true)      // Enable word wrap for better formatting
	tv.SetScrollable(true)    // Enable scrolling for long content
	tv.SetWrap(true)
	return tv
}

// createInputField creates a new input field with consistent styling
func createInputField(title string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.InputField {
	input := tview.NewInputField()
	input.SetBorder(true)
	input.SetTitle(title)
	input.SetBackgroundColor(backgroundColor)
	input.SetBorderColor(borderColor)
	input.SetTitleColor(titleColor)
	input.SetFieldTextColor(foregroundColor)
	input.SetFieldBackgroundColor(backgroundColor)
	input.SetBorderPadding(0, 0, 0, 0)
	return input
}

// createDropDown creates a dropdown with consistent styling
func createDropDown(title string, options []string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.DropDown {
	dropdown := tview.NewDropDown()
	dropdown.SetBorder(true)
	dropdown.SetTitle(title)
	dropdown.SetBackgroundColor(backgroundColor)
	dropdown.SetBorderColor(borderColor)
	dropdown.SetTitleColor(titleColor)
	dropdown.SetFieldTextColor(foregroundColor)
	dropdown.SetFieldBackgroundColor(backgroundColor)
	dropdown.SetBorderPadding(0, 0, 0, 0)
	dropdown.SetOptions(options, nil)
	return dropdown
}

func createButton(text string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.Button {
	button := tview.NewButton(text)
	button.SetBorder(false)
	button.SetBackgroundColor(backgroundColor)
	button.SetLabelColor(foregroundColor)
	button.SetBorderPadding(0, 0, 0, 0)
	button.SetBorderColor(borderColor)
	return button
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

// setBodyContainerFocusStyle handles focus styling for the body container's active panel
func setBodyContainerFocusStyle(bodyViewPanel *tview.TextView, bodyEditPanel *tview.TextArea, bodyEditMode bool, focused bool, borderColor, borderFocusColor tcell.Color) {
	if bodyEditMode {
		// Focus the edit panel
		setFocusStyle(bodyEditPanel, focused, borderColor, borderFocusColor)
		setFocusStyle(bodyViewPanel, false, borderColor, borderFocusColor)
	} else {
		// Focus the view panel
		setFocusStyle(bodyViewPanel, focused, borderColor, borderFocusColor)
		setFocusStyle(bodyEditPanel, false, borderColor, borderFocusColor)
	}
}

// setMethodUrlBarFocusStyle handles focus styling for the method+URL bar container
func setMethodUrlBarFocusStyle(methodUrlBar *tview.Flex, focused bool, borderColor, borderFocusColor tcell.Color) {
	setFocusStyle(methodUrlBar, focused, borderColor, borderFocusColor)
}

// createTextArea creates a new text area with consistent styling for body editing
func createTextArea(title string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextArea {
	textArea := tview.NewTextArea()
	textArea.SetBorder(true)
	textArea.SetTitle(title)
	textArea.SetBackgroundColor(backgroundColor)
	textArea.SetBorderColor(borderColor)
	textArea.SetTitleColor(titleColor)
	textArea.SetTextStyle(tcell.StyleDefault.Foreground(foregroundColor).Background(backgroundColor))
	textArea.SetBorderPadding(0, 0, 0, 0)
	textArea.SetWordWrap(true)
	return textArea
}

// createMethodUrlBar creates a unified method+URL+Send input component
func createMethodUrlBar(title string, backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) (*tview.Flex, *tview.DropDown, *tview.InputField, *tview.Button) {
	// Create the components without borders
	methodDropdown := createDropDown("", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}, backgroundColor, backgroundColor, foregroundColor, foregroundColor)
	methodDropdown.SetBorder(false)

	urlInput := createInputField("", backgroundColor, backgroundColor, foregroundColor, foregroundColor)
	urlInput.SetBorder(false)

	sendButton := createButton(" Send ", borderColor, borderColor, titleColor, foregroundColor)
	sendButton.SetBackgroundColor(borderColor)
	sendButton.SetLabelColor(backgroundColor)

	spacer := tview.NewBox().SetBackgroundColor(backgroundColor)

	// Create container with unified border
	container := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(spacer, 1, 0, false).
		AddItem(methodDropdown, 8, 0, false).
		AddItem(urlInput, 0, 1, false).
		AddItem(sendButton, 8, 0, false).
		AddItem(spacer, 1, 0, false)

	container.SetBorder(true)
	container.SetTitle(title)
	container.SetBackgroundColor(backgroundColor)
	container.SetBorderColor(borderColor)
	container.SetTitleColor(titleColor)
	container.SetBorderPadding(0, 0, 0, 0)

	return container, methodDropdown, urlInput, sendButton
}

// createTabHeader creates a clickable tab header bar
func createTabHeader(backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor tcell.Color, onTabClick func(int)) *tview.Flex {
	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tabHeader.SetBackgroundColor(backgroundColor)

	// Tab titles and their active states
	tabs := []string{"Body", "Auth", "Query", "Headers"}
	activeTab := 0 // Default to first tab

	// Create tab buttons
	for i, tabTitle := range tabs {
		tabIndex := i // Capture for closure

		// Create individual tab as TextView (clickable)
		tab := tview.NewTextView()
		tab.SetBackgroundColor(backgroundColor)
		tab.SetTextColor(foregroundColor)
		tab.SetTextAlign(tview.AlignCenter)
		tab.SetBorder(false)

		// Set initial text based on whether it's active
		if i == activeTab {
			tab.SetText(tabTitle)
			tab.SetTextColor(hexToColor(activeTabColor.String()))
		} else {
			tab.SetText(tabTitle)
		}

		// Make tab clickable
		tab.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
			if action == tview.MouseLeftClick {
				onTabClick(tabIndex)
			}
			return action, event
		})

		tabHeader.AddItem(tab, 0, 1, false)

		// Add separator except for last tab
		if i < len(tabs)-1 {
			separator := tview.NewTextView()
			separator.SetBackgroundColor(backgroundColor)
			separator.SetTextColor(foregroundColor)
			separator.SetText("|")
			tabHeader.AddItem(separator, 1, 0, false)
		}
	}

	return tabHeader
}

// updateTabHeader updates the active tab indicator in the tab header
func updateTabHeader(tabHeader *tview.Flex, activeTabIndex int, backgroundColor, foregroundColor, activeTabColor tcell.Color) {
	// Tab titles
	tabs := []string{"Body", "Auth", "Query", "Headers"}

	// Update each tab's appearance based on whether it's active
	for i := 0; i < len(tabs); i++ {
		// Find the tab TextView (skip separators)
		tabIndex := i * 2 // Every other item is a tab (alternating with separators)
		if tabIndex < tabHeader.GetItemCount() {
			if tab, ok := tabHeader.GetItem(tabIndex).(*tview.TextView); ok {
				if i == activeTabIndex {
					// Active tab - use active tab color
					// tab.SetText(fmt.Sprintf("[%s][ %s ][-]", activeTabColor.String(), tabs[i]))
					tab.SetText(tabs[i])
					tab.SetTextColor(activeTabColor)
				} else {
					// Inactive tab - use regular foreground color
					tab.SetText(fmt.Sprintf(" %s ", tabs[i]))
					tab.SetTextColor(foregroundColor)
				}
			}
		}
	}
}

// createAuthTab creates placeholder authentication tab content
func createAuthTab(backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	authPanel := createPanel(" Auth ", backgroundColor, borderColor, titleColor, foregroundColor)
	authPanel.SetText("Authentication settings will go here.\n\n• Bearer Token\n• Basic Auth\n• API Key\n• OAuth (future)")
	return authPanel
}

// createQueryTab creates placeholder query parameters tab content
func createQueryTab(backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	queryPanel := createPanel(" Query Parameters ", backgroundColor, borderColor, titleColor, foregroundColor)
	queryPanel.SetText("URL Query Parameters editor will go here.\n\n• Key-Value pairs\n• Add/Remove parameters\n• Bulk import\n• Templates (future)")
	return queryPanel
}

// createHeadersTab creates placeholder headers tab content
func createHeadersTab(backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	headersPanel := createPanel(" Headers ", backgroundColor, borderColor, titleColor, foregroundColor)
	headersPanel.SetText("HTTP Headers editor will go here.\n\n• Key-Value pairs\n• Common headers presets\n• Add/Remove headers\n• Auto-complete (future)")
	return headersPanel
}

// createRequestDataTabs creates the main tabbed interface for request data
func createRequestDataTabs(bodyViewPanel *tview.TextView, bodyEditPanel *tview.TextArea, backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor tcell.Color) (*tview.Flex, *tview.Pages, *tview.Flex, *tview.Flex) {
	// Create the tab content pages
	tabPages := tview.NewPages()

	// Create placeholder tabs
	authTab := createAuthTab(backgroundColor, borderColor, titleColor, foregroundColor)
	queryTab := createQueryTab(backgroundColor, borderColor, titleColor, foregroundColor)
	headersTab := createHeadersTab(backgroundColor, borderColor, titleColor, foregroundColor)

	// Body tab uses the existing dual-mode container
	bodyContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	bodyContainer.AddItem(bodyViewPanel, 0, 1, false)

	// Add all tabs to pages
	tabPages.AddPage("body", bodyContainer, true, true)
	tabPages.AddPage("auth", authTab, true, false)
	tabPages.AddPage("query", queryTab, true, false)
	tabPages.AddPage("headers", headersTab, true, false)

	// Create tab header first
	var tabHeader *tview.Flex
	var switchToTab func(int)

	// Create tab switching function
	switchToTab = func(tabIndex int) {
		tabNames := []string{"body", "auth", "query", "headers"}
		if tabIndex >= 0 && tabIndex < len(tabNames) {
			tabPages.SwitchToPage(tabNames[tabIndex])
			// Update the tab header to show the new active tab
			if tabHeader != nil {
				updateTabHeader(tabHeader, tabIndex, backgroundColor, foregroundColor, activeTabColor)
			}
		}
	}

	// Create tab header
	tabHeader = createTabHeader(backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor, switchToTab)

	// Create main container with header and content
	tabContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tabContainer.AddItem(tabHeader, 1, 0, false)
	tabContainer.AddItem(tabPages, 0, 1, false)

	return tabContainer, tabPages, bodyContainer, tabHeader
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
