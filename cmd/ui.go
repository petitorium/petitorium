package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
)

// setupTheme configures the theme colors and borders
func setupTheme() (tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color, tcell.Color) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	borderColor := hexToColor(theme.BorderColor)
	borderFocusColor := hexToColor(theme.BorderFocusColor)
	titleColor := hexToColor(theme.TitleColor)
	selectionBackgroundColor := hexToColor(theme.SelectionBackground)
	activeTabColor := hexToColor(theme.ActiveTabColor)
	buttonSelectedColor := hexToColor(theme.ButtonSelectedColor)
	dropdownFocusedBackgroundColor := hexToColor(theme.DropdownFocusedBackground)

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

	return backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor, activeTabColor, buttonSelectedColor, dropdownFocusedBackgroundColor
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
func createDropDown(title string,
	options []string,
	backgroundColor,
	borderColor,
	titleColor,
	foregroundColor,
	activeTabColor,
	listBackgroundColor,
	listSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) *tview.DropDown {
	dropdown := tview.NewDropDown()
	dropdown.SetBorder(true)
	dropdown.SetTitle(title)
	dropdown.SetBackgroundColor(backgroundColor)
	dropdown.SetBorderColor(borderColor)
	dropdown.SetTitleColor(titleColor)
	dropdown.SetFieldTextColor(activeTabColor)
	dropdown.SetFieldBackgroundColor(backgroundColor)
	dropdown.SetBorderPadding(0, 0, 0, 0)
	dropdown.SetOptions(options, nil)
	dropdown.SetFocusedStyle(tcell.StyleDefault.Background(dropdownFocusedBackgroundColor).Foreground(activeTabColor))

	// Set the dropdown list styles to match the theme
	unselectedStyle := tcell.StyleDefault.Background(listBackgroundColor).Foreground(foregroundColor)
	selectedStyle := tcell.StyleDefault.Background(listSelectedColor).Foreground(foregroundColor)
	dropdown.SetListStyles(unselectedStyle, selectedStyle)

	return dropdown
}

func createButton(text string, backgroundColor, borderColor, titleColor, foregroundColor, selectedColor tcell.Color) *tview.Button {
	button := tview.NewButton(text)
	button.SetBorder(false)
	button.SetBackgroundColor(backgroundColor)
	button.SetLabelColor(foregroundColor)
	button.SetBackgroundColorActivated(selectedColor)
	button.SetLabelColorActivated(backgroundColor)
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

// createMethodURLBar creates a unified method+URL+Send input component
func createMethodURLBar(
	title string,
	backgroundColor,
	borderColor,
	titleColor,
	foregroundColor,
	selectionBackgroundColor,
	activeTabColor,
	buttonSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) (*tview.Flex, *tview.DropDown, *tview.InputField, *tview.Button) {
	// Create the components without borders
	methodDropdown := createDropDown(
		"",
		[]string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		backgroundColor,
		backgroundColor,
		backgroundColor,
		foregroundColor,
		activeTabColor,
		backgroundColor,
		selectionBackgroundColor,
		dropdownFocusedBackgroundColor,
	)
	methodDropdown.SetBorder(false)

	urlInput := createInputField("", backgroundColor, backgroundColor, foregroundColor, foregroundColor)
	urlInput.SetBorder(false)

	sendButton := createButton(" Send ", borderColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
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
func createTabHeader(tabs []string, backgroundColor,
	borderColor,
	titleColor,
	foregroundColor,
	activeTabColor,
	selectionBackgroundColor tcell.Color,
	onTabClick func(int),
) *tview.Flex {
	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tabHeader.SetBackgroundColor(backgroundColor)
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
			tab.SetTextColor(activeTabColor)
			tab.SetBackgroundColor(selectionBackgroundColor)
		} else {
			tab.SetText(tabTitle)
			tab.SetBackgroundColor(backgroundColor)
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
func updateTabHeader(
	tabs []string,
	tabHeader *tview.Flex,
	activeTabIndex int,
	backgroundColor,
	foregroundColor,
	activeTabColor,
	selectionBackgroundColor tcell.Color,
) {

	// Update each tab's appearance based on whether it's active
	for i := 0; i < len(tabs); i++ {
		// Find the tab TextView (skip separators)
		tabIndex := i * 2 // Every other item is a tab (alternating with separators)
		if tabIndex < tabHeader.GetItemCount() {
			if tab, ok := tabHeader.GetItem(tabIndex).(*tview.TextView); ok {
				if i == activeTabIndex {
					// Active tab - use active tab color and background
					tab.SetText(tabs[i])
					tab.SetTextColor(activeTabColor)
					tab.SetBackgroundColor(selectionBackgroundColor)
				} else {
					// Inactive tab - use regular foreground color and background
					tab.SetText(fmt.Sprintf(" %s ", tabs[i]))
					tab.SetTextColor(foregroundColor)
					tab.SetBackgroundColor(backgroundColor)
				}
			}
		}
	}
}

// createAuthTab creates placeholder authentication tab content
func createAuthTab(backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	// Auth panel
	// borderColor
	authPanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	authPanel.SetText("Authentication settings will go here.\n\n• Bearer Token\n• Basic Auth\n• API Key\n• OAuth (future)")
	return authPanel
}

// createQueryTab creates placeholder query parameters tab content
func createQueryTab(backgroundColor, borderColor, titleColor, foregroundColor tcell.Color) *tview.TextView {
	// Query Parameters panel
	// borderColor
	queryPanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	queryPanel.SetText("URL Query Parameters editor will go here.\n\n• Key-Value pairs\n• Add/Remove parameters\n• Bulk import\n• Templates (future)")
	return queryPanel
}

// Global variables for headers management
var currentHeadersTab *tview.Flex

var currentHeaderRows []*HeaderRow

var currentHeadersList *tview.Flex

var rowHeight int = 1

// HeaderRow represents a single header key-value pair in the UI
type HeaderRow struct {
	KeyInput     *tview.InputField
	ValueInput   *tview.InputField
	DeleteButton *tview.Button
	Row          *tview.Flex
}

// createHeadersTabWithData creates headers tab with initial data
func createHeadersTabWithData(backgroundColor,
	borderColor,
	titleColor,
	foregroundColor,
	buttonSelectedColor tcell.Color,
	initialHeaders map[string]string,
	saveCallback func(),
	focusSetter func(tview.Primitive),
) *tview.Flex {
	headersContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	headersContainer.SetBackgroundColor(backgroundColor)
	headersContainer.SetBorder(true)
	headersContainer.SetBorderColor(backgroundColor) // borderColor
	headersContainer.SetTitleColor(titleColor)
	headersContainer.SetBackgroundColor(backgroundColor)

	// Scrollable area for header entries
	headersList := tview.NewFlex().SetDirection(tview.FlexRow)
	headersList.SetBackgroundColor(backgroundColor)

	// Store references for global access
	currentHeadersTab = headersContainer
	currentHeadersList = headersList

	// Initialize global header rows
	currentHeaderRows = []*HeaderRow{}

	// Function to refresh the UI
	var refreshHeadersUI func()
	refreshHeadersUI = func() {
		headersList.Clear()
		for _, row := range currentHeaderRows {
			headersList.AddItem(row.Row, rowHeight, 0, false)
		}
		// Always keep at least one empty row
		if len(currentHeaderRows) == 0 {
			addHeaderRow(headersList, backgroundColor, foregroundColor, borderColor, refreshHeadersUI, saveCallback, focusSetter)
		}
	}

	// Add initial rows based on data
	if initialHeaders != nil && len(initialHeaders) > 0 {
		for key, value := range initialHeaders {
			addHeaderRowWithData(headersList, backgroundColor, foregroundColor, borderColor, key, value, refreshHeadersUI, saveCallback, focusSetter)
		}
	}
	// Always add at least one empty row
	addHeaderRow(headersList, backgroundColor, foregroundColor, borderColor, refreshHeadersUI, saveCallback, focusSetter)

	// Add button row at the top
	buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonRow.SetBackgroundColor(backgroundColor)

	addButton := createButton(" Add Header ", borderColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
	addButton.SetBackgroundColorActivated(foregroundColor)
	addButton.SetLabelColor(backgroundColor)
	addButton.SetSelectedFunc(func() {
		addHeaderRow(headersList, backgroundColor, foregroundColor, borderColor, refreshHeadersUI, saveCallback, focusSetter)
		refreshHeadersUI()
		if saveCallback != nil {
			saveCallback()
		}
	})

	// Delete all button
	deleteAllButton := createButton(" Delete All ", borderColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
	deleteAllButton.SetBorder(false)
	deleteAllButton.SetStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	deleteAllButton.SetSelectedFunc(func() {
		// Clear all header rows
		currentHeaderRows = []*HeaderRow{}
		refreshHeadersUI()
		if saveCallback != nil {
			saveCallback()
		}
	})

	buttonRow.AddItem(addButton, 15, 0, false)
	buttonRow.AddItem(deleteAllButton, 15, 0, false)
	buttonRow.AddItem(nil, 0, 1, false)

	headersContainer.AddItem(buttonRow, 1, 0, false)

	// Add visual spacing between buttons and headers
	spacer := tview.NewBox().SetBackgroundColor(backgroundColor)
	headersContainer.AddItem(spacer, 1, 0, false)

	headersContainer.AddItem(headersList, 0, 1, false)

	return headersContainer
}

// addHeaderRow adds a new key-value header input row to the headers list
func addHeaderRow(headersList *tview.Flex,
	backgroundColor,
	foregroundColor,
	borderColor tcell.Color,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(backgroundColor)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(backgroundColor)
	keyInput.SetFieldBackgroundColor(backgroundColor)
	keyInput.SetFieldTextColor(foregroundColor)
	keyInput.SetLabelColor(foregroundColor)
	keyInput.SetPlaceholder("Header name")
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(hexToColor("#4A5053")))
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	keyInput.SetFieldBackgroundColor(backgroundColor)
	keyInput.SetChangedFunc(func(text string) {
		if saveCallback != nil {
			saveCallback()
		}
	})
	keyInput.SetBlurFunc(func() {
		if saveCallback != nil {
			saveCallback()
		}
	})

	valueInput := tview.NewInputField()
	valueInput.SetBackgroundColor(backgroundColor)
	valueInput.SetFieldBackgroundColor(backgroundColor)
	valueInput.SetFieldTextColor(foregroundColor)
	valueInput.SetLabelColor(foregroundColor)
	valueInput.SetPlaceholder("Header value")
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(hexToColor("#4A5053")))
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	valueInput.SetFieldBackgroundColor(backgroundColor)
	valueInput.SetChangedFunc(func(text string) {
		if saveCallback != nil {
			saveCallback()
		}
	})
	valueInput.SetBlurFunc(func() {
		if saveCallback != nil {
			saveCallback()
		}
	})

	removeButton := tview.NewButton(config.C.UI.HeaderRemoveIcon)
	removeButton.SetBackgroundColor(backgroundColor)
	removeButton.SetLabelColor(foregroundColor)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))

	headerRow := &HeaderRow{
		KeyInput:     keyInput,
		ValueInput:   valueInput,
		DeleteButton: removeButton,
		Row:          row,
	}

	removeButton.SetSelectedFunc(func() {
		// Find and remove this row
		for i, r := range currentHeaderRows {
			if r == headerRow {
				currentHeaderRows = append(currentHeaderRows[:i], currentHeaderRows[i+1:]...)
				refreshUI()
				if saveCallback != nil {
					saveCallback()
				}
				break
			}
		}
	})

	// Handle Tab navigation for delete button
	removeButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Tab from delete button to next row's key input
			// Simple implementation: cycle to first row's key input
			if len(currentHeaderRows) > 0 {
				firstRow := currentHeaderRows[0]
				focusSetter(firstRow.KeyInput)
			}
			return nil
		}
		return event
	})

	// Handle Tab to add new header row when on the last value input
	valueInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Check if this is the last value input
			if len(currentHeaderRows) > 0 && currentHeaderRows[len(currentHeaderRows)-1] == headerRow {
				// Cycle back to first key input
				firstRow := currentHeaderRows[0]
				focusSetter(firstRow.KeyInput)
				return nil
			}
		}
		return event
	})

	// Wrap button in a container to match row height
	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonContainer.SetBackgroundColor(backgroundColor)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentHeaderRows = append(currentHeaderRows, headerRow)
	headersList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	separator := tview.NewBox().SetBackgroundColor(backgroundColor)
	separator.SetBorder(false)
	headersList.AddItem(separator, 1, 0, false)

	// Update headers cycle
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// getHeadersFromUI extracts headers from the current UI state
func getHeadersFromUI() map[string]string {
	headers := make(map[string]string)
	for _, row := range currentHeaderRows {
		key := strings.TrimSpace(row.KeyInput.GetText())
		value := strings.TrimSpace(row.ValueInput.GetText())
		if key != "" {
			headers[key] = value
		}
	}
	return headers
}

// setHeadersInUI populates the UI with the given headers
func setHeadersInUI(headers map[string]string, saveCallback func(), focusSetter func(tview.Primitive)) {
	// Clear existing rows
	currentHeaderRows = []*HeaderRow{}

	if currentHeadersList != nil {
		currentHeadersList.Clear()

		// Get current theme colors
		theme := config.C.Theme
		backgroundColor := hexToColor(theme.BackgroundColor)
		foregroundColor := hexToColor(theme.ForegroundColor)
		borderColor := hexToColor(theme.BorderColor)

		// Add rows for each header (sorted by key for consistent ordering)
		var keys []string
		for key := range headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := headers[key]
			addHeaderRowWithData(currentHeadersList, backgroundColor, foregroundColor, borderColor, key, value, func() {
				// Refresh function - for now just rebuild the list
				setHeadersInUI(getHeadersFromUI(), saveCallback, focusSetter)
			}, saveCallback, focusSetter)
		}

		// Always add one empty row
		addHeaderRow(currentHeadersList, backgroundColor, foregroundColor, borderColor, func() {
			setHeadersInUI(getHeadersFromUI(), saveCallback, focusSetter)
		}, saveCallback, focusSetter)
	}

	// Update headers cycle
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// addHeaderRowWithData adds a header row with pre-filled data
func addHeaderRowWithData(
	headersList *tview.Flex,
	backgroundColor, foregroundColor, borderColor tcell.Color,
	key, value string,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(backgroundColor)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(backgroundColor)
	keyInput.SetFieldBackgroundColor(backgroundColor)
	keyInput.SetFieldTextColor(foregroundColor)
	keyInput.SetLabelColor(foregroundColor)
	keyInput.SetPlaceholder("Header name")
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	keyInput.SetText(key)
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	if saveCallback != nil {
		keyInput.SetChangedFunc(func(text string) {
			saveCallback()
		})
		keyInput.SetBlurFunc(func() {
			saveCallback()
		})
	}

	valueInput := tview.NewInputField()
	valueInput.SetBackgroundColor(backgroundColor)
	valueInput.SetFieldBackgroundColor(backgroundColor)
	valueInput.SetFieldTextColor(foregroundColor)
	valueInput.SetLabelColor(foregroundColor)
	valueInput.SetPlaceholder("Header value")
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	valueInput.SetText(value)
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))
	if saveCallback != nil {
		valueInput.SetChangedFunc(func(text string) {
			saveCallback()
		})
		valueInput.SetBlurFunc(func() {
			saveCallback()
		})
	}

	removeButton := tview.NewButton(config.C.UI.HeaderRemoveIcon)
	removeButton.SetBackgroundColor(backgroundColor)
	removeButton.SetLabelColor(foregroundColor)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(backgroundColor).Foreground(foregroundColor))

	headerRow := &HeaderRow{
		KeyInput:     keyInput,
		ValueInput:   valueInput,
		DeleteButton: removeButton,
		Row:          row,
	}

	removeButton.SetSelectedFunc(func() {
		// Find and remove this row
		for i, r := range currentHeaderRows {
			if r == headerRow {
				currentHeaderRows = append(currentHeaderRows[:i], currentHeaderRows[i+1:]...)
				refreshUI()
				if saveCallback != nil {
					saveCallback()
				}
				break
			}
		}
	})

	// Handle Tab navigation for delete button
	removeButton.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Tab from delete button to next row's key input
			// Simple implementation: cycle to first row's key input
			if len(currentHeaderRows) > 0 {
				firstRow := currentHeaderRows[0]
				focusSetter(firstRow.KeyInput)
			}
			return nil
		}
		return event
	})

	// Handle Shift+Tab navigation between key/value inputs
	keyInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBacktab {
			// Shift+Tab: Find current row index and move to previous row's value input
			currentIndex := -1
			for i, row := range currentHeaderRows {
				if row == headerRow {
					currentIndex = i
					break
				}
			}

			if currentIndex > 0 {
				// Move to previous row's value input
				prevRow := currentHeaderRows[currentIndex-1]
				focusSetter(prevRow.ValueInput)
			} else {
				// First row, cycle to last row's value input
				lastRow := currentHeaderRows[len(currentHeaderRows)-1]
				focusSetter(lastRow.ValueInput)
			}
			return nil
		}
		return event
	})

	valueInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyBacktab {
			// Shift+Tab: Move to key input of same row
			focusSetter(keyInput)
			return nil
		}
		return event
	})

	// Wrap button in a container to match row height
	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonContainer.SetBackgroundColor(backgroundColor)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentHeaderRows = append(currentHeaderRows, headerRow)
	headersList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	// separator := tview.NewBox().SetBackgroundColor(backgroundColor)
	// separator.SetBorder(true)
	// headersList.AddItem(separator, 1, 0, false)

	// Update headers cycle
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// createRequestDataTabs creates the main tabbed interface for request data
func createRequestDataTabs(
	bodyViewPanel *tview.TextView,
	bodyEditPanel *tview.TextArea,
	backgroundColor, borderColor, borderFocusColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor, buttonSelectedColor tcell.Color,
	saveCallback func(),
	focusSetter func(tview.Primitive),
	tabIndexSetter func(int),
	panelFocusSetter func(int, bool),
) (*tview.Flex, *tview.Pages, *tview.Flex, *tview.Flex, *tview.TextView, *tview.TextView, *tview.Flex, func(int)) {
	// Create the tab content pages
	tabPages := tview.NewPages()

	// Create placeholder tabs
	authTab := createAuthTab(backgroundColor, borderColor, titleColor, foregroundColor)
	queryTab := createQueryTab(backgroundColor, borderColor, titleColor, foregroundColor)
	headersTab := createHeadersTabWithData(backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor, nil, saveCallback, focusSetter)

	// Body tab uses the existing dual-mode container
	bodyContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
	bodyContainer.SetBorderColor(backgroundColor)

	// Add all tabs to pages
	tabPages.AddPage("body", bodyContainer, true, true)
	tabPages.AddPage("auth", authTab, true, false)
	tabPages.AddPage("query", queryTab, true, false)
	tabPages.AddPage("headers", headersTab, true, false)
	tabPages.SetBackgroundColor(backgroundColor)

	// Create tab header first
	var tabHeader *tview.Flex
	// var switchToTab func(int)

	// Create tab switching function
	switchToTab := func(tabIndex int) {
		tabNames := []string{"body", "auth", "query", "headers"}
		if tabIndex >= 0 && tabIndex < len(tabNames) {
			tabPages.SwitchToPage(tabNames[tabIndex])
			// Update the tab header to show the new active tab
			if tabHeader != nil {
				tabs := []string{"Body", "Auth", "Query", "Headers"}
				updateTabHeader(tabs, tabHeader, tabIndex, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
			}

			// Ensure only the request panel is focused
			// Unfocus all panels that have borders
			if panelFocusSetter != nil {
				// Unfocus all panels: collections, request panel, response
				for i := 0; i < 3; i++ {
					if i != 1 { // Don't unfocus request panel itself
						panelFocusSetter(i, false)
					}
				}
			}

			// Explicitly focus the request panel
			if panelFocusSetter != nil {
				panelFocusSetter(1, true)
			}
		}
	}

	// Create tab header
	tabs := []string{"Body", "Auth", "Query", "Headers"}
	tabHeader = createTabHeader(tabs, backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor, switchToTab)

	// Create main container with header and content
	tabContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tabContainer.AddItem(tabHeader, 1, 0, false)
	tabContainer.AddItem(tabPages, 0, 1, false)
	tabContainer.SetBorder(true)
	tabContainer.SetBorderColor(borderColor)
	tabContainer.SetTitle(" Request ")
	tabContainer.SetTitleColor(titleColor)
	tabContainer.SetBackgroundColor(backgroundColor)

	return tabContainer, tabPages, bodyContainer, tabHeader, authTab, queryTab, headersTab, switchToTab
}

// getStatusCodeColors returns appropriate background and foreground colors for HTTP status codes
func getStatusCodeColors(statusCode int, defaultBg, defaultFg tcell.Color) (tcell.Color, tcell.Color) {
	var bgColorHex, fgColorHex string

	switch {
	case statusCode >= 200 && statusCode < 300:
		// 2xx Success
		bgColorHex = config.C.StatusColors.Success
		fgColorHex = config.C.StatusColors.SuccessText
	case statusCode >= 300 && statusCode < 400:
		// 3xx Redirection
		bgColorHex = config.C.StatusColors.Redirection
		fgColorHex = config.C.StatusColors.RedirectionText
	case statusCode >= 400 && statusCode < 500:
		// 4xx Client Error
		bgColorHex = config.C.StatusColors.ClientError
		fgColorHex = config.C.StatusColors.ClientErrorText
	case statusCode >= 500:
		// 5xx Server Error
		bgColorHex = config.C.StatusColors.ServerError
		fgColorHex = config.C.StatusColors.ServerErrorText
	default:
		// Unknown/other - Use default from config or fallback to default colors
		if config.C.StatusColors.Default != "" {
			return hexToColor(config.C.StatusColors.Default), hexToColor(config.C.StatusColors.DefaultText)
		}
		return defaultBg, defaultFg
	}

	// Convert hex colors
	return hexToColor(bgColorHex), hexToColor(fgColorHex)
}

// createResponseInfoBar creates the info bar showing status, time, bytes, and last request time
func createResponseInfoBar(
	backgroundColor, foregroundColor, titleColor tcell.Color,
	response *HTTPResponse,
	lastRequestTime *time.Time,
) *tview.Flex {
	infoBar := tview.NewFlex().SetDirection(tview.FlexColumn)
	infoBar.SetBackgroundColor(backgroundColor)

	// Status
	statusText := " --- "
	statusViewWidth := 5
	statusBgColor := backgroundColor
	statusFgColor := foregroundColor
	if response != nil {
		statusText = fmt.Sprintf(" %d ", response.StatusCode)
		statusBgColor, statusFgColor = getStatusCodeColors(response.StatusCode, backgroundColor, foregroundColor)
	}
	// Create status container with fixed width
	statusContainer := tview.NewFlex().SetDirection(tview.FlexColumn)
	statusContainer.SetBackgroundColor(statusBgColor)

	statusView := tview.NewTextView()
	statusView.SetText(statusText)
	statusView.SetTextColor(statusFgColor)
	statusView.SetBackgroundColor(statusBgColor)
	statusView.SetBorder(false)

	statusContainer.AddItem(statusView, statusViewWidth, 0, false) // Fixed width for status code

	// Duration
	durationText := "Time: -"
	if response != nil {
		durationText = fmt.Sprintf("Time: %v", response.Duration.Round(time.Millisecond))
	}
	durationView := tview.NewTextView()
	durationView.SetText(durationText)
	durationView.SetTextColor(foregroundColor)
	durationView.SetBackgroundColor(backgroundColor)
	durationView.SetBorder(false)

	// Size
	sizeText := "Size: -"
	if response != nil {
		sizeText = fmt.Sprintf("Size: %d bytes", response.BodySize)
	}
	sizeView := tview.NewTextView()
	sizeView.SetText(sizeText)
	sizeView.SetTextColor(foregroundColor)
	sizeView.SetBackgroundColor(backgroundColor)
	sizeView.SetBorder(false)

	// Last request time
	lastTimeText := "Last: never"
	if lastRequestTime != nil {
		lastTimeText = fmt.Sprintf("Last: %s ago", timeSinceHuman(*lastRequestTime))
	}
	lastView := tview.NewTextView()
	lastView.SetText(lastTimeText)
	lastView.SetTextColor(foregroundColor)
	lastView.SetBackgroundColor(backgroundColor)
	lastView.SetBorder(false)

	infoBar.AddItem(statusContainer, statusViewWidth, 0, false)
	infoBar.AddItem(durationView, 0, 1, false)
	infoBar.AddItem(sizeView, 0, 1, false)
	infoBar.AddItem(lastView, 0, 1, false)

	return infoBar
}

// timeSinceHuman returns a human-readable time since the given time
func timeSinceHuman(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	} else {
		return fmt.Sprintf("%dd", int(duration.Hours()/24))
	}
}

// createResponseTabs creates the tabbed interface for response data
func createResponseTabs(
	backgroundColor, borderColor, borderFocusColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor tcell.Color,
	response *HTTPResponse,
	lastRequestTime *time.Time,
) (*tview.Flex, *tview.Pages, *tview.Flex, func(int)) {
	// Create info bar
	infoBar := createResponseInfoBar(backgroundColor, foregroundColor, titleColor, response, lastRequestTime)

	// Create the tab content pages
	tabPages := tview.NewPages()

	// Preview tab - shows formatted response body
	previewPanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	if response != nil && response.Body != "" {
		// Pretty-print JSON if it's valid JSON
		bodyToFormat := response.Body
		if strings.TrimSpace(response.Body)[0] == '{' || strings.TrimSpace(response.Body)[0] == '[' {
			var jsonData interface{}
			if err := json.Unmarshal([]byte(response.Body), &jsonData); err == nil {
				// It's valid JSON, pretty-print it
				prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
				if err == nil {
					bodyToFormat = string(prettyJSON)
				}
			}
		}
		formattedBody := formatBodyContent(bodyToFormat)
		previewPanel.SetText(formattedBody)
	} else {
		previewPanel.SetText("(empty response)")
	}

	// Headers tab
	headersPanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	if response != nil && len(response.Headers) > 0 {
		var headersText strings.Builder
		headersText.WriteString("Response Headers:\n\n")
		for key, values := range response.Headers {
			for _, value := range values {
				headersText.WriteString(fmt.Sprintf("%s: %s\n", key, value))
			}
		}
		headersPanel.SetText(headersText.String())
	} else {
		headersPanel.SetText("No response headers")
	}

	// Cookies tab
	cookiesPanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	if response != nil && len(response.Cookies) > 0 {
		var cookiesText strings.Builder
		cookiesText.WriteString("Response Cookies:\n\n")
		for _, cookie := range response.Cookies {
			cookiesText.WriteString(fmt.Sprintf("Name: %s\n", cookie.Name))
			cookiesText.WriteString(fmt.Sprintf("Value: %s\n", cookie.Value))
			if cookie.Domain != "" {
				cookiesText.WriteString(fmt.Sprintf("Domain: %s\n", cookie.Domain))
			}
			if cookie.Path != "" {
				cookiesText.WriteString(fmt.Sprintf("Path: %s\n", cookie.Path))
			}
			if !cookie.Expires.IsZero() {
				cookiesText.WriteString(fmt.Sprintf("Expires: %s\n", cookie.Expires.Format("2006-01-02 15:04:05")))
			}
			cookiesText.WriteString(fmt.Sprintf("Secure: %t\n", cookie.Secure))
			cookiesText.WriteString(fmt.Sprintf("HttpOnly: %t\n\n", cookie.HttpOnly))
		}
		cookiesPanel.SetText(cookiesText.String())
	} else {
		cookiesPanel.SetText("No response cookies")
	}

	// Timeline tab - timing information
	timelinePanel := createPanel("", backgroundColor, backgroundColor, titleColor, foregroundColor)
	if response != nil {
		var timelineText strings.Builder
		timelineText.WriteString("Request Timeline:\n\n")
		timelineText.WriteString(fmt.Sprintf("Request sent: %s\n", response.Timestamp.Format("2006-01-02 15:04:05")))
		timelineText.WriteString(fmt.Sprintf("Response received: %s\n", response.Timestamp.Add(response.Duration).Format("2006-01-02 15:04:05")))
		timelineText.WriteString(fmt.Sprintf("Total duration: %v\n", response.Duration.Round(time.Millisecond)))
		timelineText.WriteString(fmt.Sprintf("Response size: %d bytes\n", response.BodySize))
		timelinePanel.SetText(timelineText.String())
	} else {
		timelinePanel.SetText("No request timeline available")
	}

	// Add all tabs to pages
	tabPages.AddPage("preview", previewPanel, true, true)
	tabPages.AddPage("headers", headersPanel, true, false)
	tabPages.AddPage("cookies", cookiesPanel, true, false)
	tabPages.AddPage("timeline", timelinePanel, true, false)
	tabPages.SetBackgroundColor(backgroundColor)

	// Create tab header
	var tabHeader *tview.Flex

	// Create tab switching function
	switchToTab := func(tabIndex int) {
		tabNames := []string{"preview", "headers", "cookies", "timeline"}
		if tabIndex >= 0 && tabIndex < len(tabNames) {
			tabPages.SwitchToPage(tabNames[tabIndex])
			// Update the tab header to show the new active tab
			if tabHeader != nil {
				responseTabs := []string{"Preview", "Headers", "Cookies", "Timeline"}
				updateTabHeader(responseTabs, tabHeader, tabIndex, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
			}
		}
	}

	// Create tab header
	responseTabs := []string{"Preview", "Headers", "Cookies", "Timeline"}
	tabHeader = createTabHeader(responseTabs, backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor, switchToTab)

	// Create combined top row: tabs on left, spacer, info bar on right
	topRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	topRow.SetBackgroundColor(backgroundColor)
	topRow.AddItem(tabHeader, 0, 1, false) // Tabs take minimal space on left
	spacer := tview.NewBox().SetBackgroundColor(backgroundColor)
	topRow.AddItem(spacer, 0, 1, false)  // Spacer takes flexible space in middle
	topRow.AddItem(infoBar, 0, 1, false) // Info bar takes flexible space on right

	// Create main container with combined row and content
	tabContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	tabContainer.AddItem(topRow, 1, 0, false)
	tabContainer.AddItem(tabPages, 0, 1, false)
	tabContainer.SetBorder(true)
	tabContainer.SetBorderColor(borderColor)
	tabContainer.SetTitle(" Response ")
	tabContainer.SetTitleColor(titleColor)
	tabContainer.SetBackgroundColor(backgroundColor)

	return tabContainer, tabPages, tabHeader, switchToTab
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
