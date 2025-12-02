package cmd

import (
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
)

type PanelOptions struct {
	HasBorder   *bool
	BorderColor *tcell.Color
	Scrollable  *bool
	Wrap        *bool
}

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xclip", "-selection", "clipboard")
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

// createPanel creates a new text view panel with consistent styling and 16m color support
func createPanel(title string, colors *ColorManager, opts *PanelOptions) *tview.TextView {
	border := true
	scrollable := true
	wrap := true
	borderColor := colors.Border

	if opts != nil {
		if opts.HasBorder != nil {
			border = *opts.HasBorder
		}
		if opts.Scrollable != nil {
			scrollable = *opts.Scrollable
		}
		if opts.Wrap != nil {
			wrap = *opts.Wrap
		}
		if opts.BorderColor != nil {
			borderColor = tcell.Color(*opts.BorderColor)
		}
	}

	tv := tview.NewTextView()
	tv.SetBorder(border)
	tv.SetTitle(title)
	tv.SetBackgroundColor(colors.Background)
	tv.SetBorderColor(borderColor)
	tv.SetTitleColor(colors.Title)
	tv.SetTextColor(colors.Foreground)
	tv.SetBorderPadding(0, 0, 0, 0)
	tv.SetDynamicColors(true)    // Enable color interpretation with hex support
	tv.SetRegions(true)          // Enable regions for better color handling
	tv.SetWordWrap(true)         // Enable word wrap for better formatting
	tv.SetScrollable(scrollable) // Enable scrolling for long content
	tv.SetWrap(wrap)

	return tv
}

// createInputField creates a new input field with consistent styling
func createInputField(title string, colors *ColorManager) *tview.InputField {
	input := tview.NewInputField()
	input.SetBorder(true)
	input.SetTitle(title)
	input.SetBackgroundColor(colors.Background)
	input.SetBorderColor(colors.Border)
	input.SetTitleColor(colors.Title)
	input.SetFieldTextColor(colors.Foreground)
	input.SetFieldBackgroundColor(colors.Background)
	input.SetBorderPadding(0, 0, 0, 0)
	return input
}

// createDropDown creates a dropdown with consistent styling
func createDropDown(title string,
	options []string,
	colors *ColorManager,
) *tview.DropDown {
	return createDropDownWithOpenOnFocus(title, options, colors, true)
}

func createDropDownWithOpenOnFocus(title string,
	options []string,
	colors *ColorManager,
	openOnFocus bool,
) *tview.DropDown {
	dropdown := tview.NewDropDown()
	dropdown.SetBorder(true)
	dropdown.SetTitle(title)
	dropdown.SetBackgroundColor(colors.Background)
	dropdown.SetBorderColor(colors.Border)
	dropdown.SetTitleColor(colors.Title)
	dropdown.SetFieldTextColor(colors.ActiveTab)
	dropdown.SetFieldBackgroundColor(colors.Background)
	dropdown.SetBorderPadding(0, 0, 0, 0)
	dropdown.SetOptions(options, nil)
	dropdown.SetFocusedStyle(tcell.StyleDefault.Background(colors.DropdownFocus).Foreground(colors.ActiveTab))

	// Set the dropdown list styles to match the theme
	unselectedStyle := tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground)
	selectedStyle := tcell.StyleDefault.Background(colors.Selection).Foreground(colors.Foreground)
	dropdown.SetListStyles(unselectedStyle, selectedStyle)

	// Use reflection to set openOnFocus
	if !openOnFocus {
		v := reflect.ValueOf(dropdown).Elem()
		field := v.FieldByName("openOnFocus")
		if field.IsValid() && field.CanSet() {
			field.SetBool(false)
		}
	}

	return dropdown
}

func isDropdownOpen(dropdown *tview.DropDown) bool {
	v := reflect.ValueOf(dropdown).Elem()
	field := v.FieldByName("open")
	if field.IsValid() && field.Kind() == reflect.Bool {
		return field.Bool()
	}
	return false
}

func createButton(text string, colors *ColorManager) *tview.Button {
	button := tview.NewButton(text)
	button.SetBorder(false)
	button.SetBackgroundColor(colors.Background)
	button.SetLabelColor(colors.Foreground)
	button.SetBackgroundColorActivated(colors.ButtonSelect)
	button.SetLabelColorActivated(colors.Background)
	button.SetBorderPadding(0, 0, 0, 0)
	button.SetBorderColor(colors.Border)
	return button
}

// CustomButton is a custom button primitive with full control over styling
type CustomButton struct {
	*tview.Box
	text                string
	textAlignment       string // "left", "center", "right"
	onSelected          func()
	backgroundColor     tcell.Color
	activatedColor      tcell.Color
	labelColor          tcell.Color
	labelActivatedColor tcell.Color
	isActivated         bool
}

// NewCustomButton creates a new custom button
func NewCustomButton(text string) *CustomButton {
	box := tview.NewBox()
	box.SetBorder(false)

	cb := &CustomButton{
		Box:                 box,
		text:                text,
		textAlignment:       "center",
		backgroundColor:     tcell.ColorDefault,
		activatedColor:      tcell.ColorDefault,
		labelColor:          tcell.ColorDefault,
		labelActivatedColor: tcell.ColorDefault,
		isActivated:         false,
	}

	// Set mouse capture for click handling
	// box.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
	// 	if action == tview.MouseLeftClick && cb.onSelected != nil {
	// 		cb.isActivated = true
	// 		cb.updateBackground()
	// 		cb.onSelected()
	// 		// Reset activation after a short delay
	// 		go func() {
	// 			// Small delay to show activated state
	// 			time.Sleep(100 * time.Millisecond)
	// 			cb.isActivated = false
	// 			cb.updateBackground()
	// 		}()
	// 	}
	// 	return action, event
	// })

	return cb
}

// SetBackgroundColor sets the normal background color
func (cb *CustomButton) SetBackgroundColor(color tcell.Color) *CustomButton {
	cb.backgroundColor = color
	cb.updateBackground()
	return cb
}

// SetBackgroundColorActivated sets the activated background color
func (cb *CustomButton) SetBackgroundColorActivated(color tcell.Color) *CustomButton {
	cb.activatedColor = color
	cb.updateBackground()
	return cb
}

// SetLabelColor sets the normal label color
func (cb *CustomButton) SetLabelColor(color tcell.Color) *CustomButton {
	cb.labelColor = color
	return cb
}

// SetLabelColorActivated sets the activated label color
func (cb *CustomButton) SetLabelColorActivated(color tcell.Color) *CustomButton {
	cb.labelActivatedColor = color
	return cb
}

// SetSelectedFunc sets the function to call when the button is selected
func (cb *CustomButton) SetSelectedFunc(handler func()) *CustomButton {
	cb.onSelected = handler
	return cb
}

// SetText sets the button text
func (cb *CustomButton) SetText(text string) *CustomButton {
	cb.text = text
	return cb
}

// SetTextAlignment sets the text alignment ("left", "center", "right")
func (cb *CustomButton) SetTextAlignment(alignment string) *CustomButton {
	cb.textAlignment = alignment
	return cb
}

// updateBackground updates the background color based on activation state
func (cb *CustomButton) updateBackground() {
	if cb.isActivated && cb.activatedColor != tcell.ColorDefault {
		cb.Box.SetBackgroundColor(cb.activatedColor)
	} else if cb.backgroundColor != tcell.ColorDefault {
		cb.Box.SetBackgroundColor(cb.backgroundColor)
	}
}

// Draw implements the Primitive interface
func (cb *CustomButton) Draw(screen tcell.Screen) {
	cb.Box.Draw(screen)

	x, y, width, height := cb.Box.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Calculate text X position based on alignment
	textLen := utf8.RuneCountInString(cb.text)
	var textX int
	switch cb.textAlignment {
	case "left":
		textX = x
	case "right":
		textX = x + width - textLen
	default: // "center"
		textX = x + (width-textLen)/2
	}
	textY := y + height/2

	labelColor := cb.labelColor
	if cb.isActivated && cb.labelActivatedColor != tcell.ColorDefault {
		labelColor = cb.labelActivatedColor
	}

	// Draw each character of the text
	for i, ch := range cb.text {
		if textX+i >= x && textX+i < x+width && textY >= y && textY < y+height {
			screen.SetContent(textX+i, textY, ch, nil, tcell.StyleDefault.Background(cb.Box.GetBackgroundColor()).Foreground(labelColor))
		}
	}
}

// InputHandler implements the Primitive interface
func (cb *CustomButton) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event.Key() == tcell.KeyEnter && cb.onSelected != nil {
			cb.isActivated = true
			cb.updateBackground()
			cb.onSelected()
			// Reset activation after a short delay
			go func() {
				// Small delay to show activated state
				time.Sleep(100 * time.Millisecond)
				cb.isActivated = false
				cb.updateBackground()
			}()
		}
	}
}

// MouseHandler implements the Primitive interface for mouse events
func (cb *CustomButton) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
			x, y := event.Position()
			bx, by, width, height := cb.Box.GetRect()

			// Check if click is within button bounds
			if x >= bx && x < bx+width && y >= by && y < by+height {
				if cb.onSelected != nil {
					cb.isActivated = true
					cb.updateBackground()
					cb.onSelected()
					// Reset activation after a short delay
					go func() {
						// Small delay to show activated state
						time.Sleep(100 * time.Millisecond)
						cb.isActivated = false
						cb.updateBackground()
					}()
				}
				return true, nil
			}
		}
		return false, nil
	}
}

// createCustomButton creates a custom button with full styling control
func createCustomButton(text string, backgroundColor, activatedColor, labelColor, labelActivatedColor tcell.Color) *CustomButton {
	button := NewCustomButton(text)
	button.SetBackgroundColor(backgroundColor)
	button.SetBackgroundColorActivated(activatedColor)
	button.SetLabelColor(labelColor)
	button.SetLabelColorActivated(labelActivatedColor)
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
	colors *ColorManager,
	app *tview.Application,
) (*tview.Flex, *tview.DropDown, *URLVariableInput, *tview.Button) {
	// Create the components without borders
	methodDropdown := createDropDown(
		"",
		[]string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		colors,
	)
	methodDropdown.SetBorder(false)

	urlInput := NewURLVariableInput(colors, app)

	sendButton := createButton(" Send ", colors)
	sendButton.SetBackgroundColor(colors.Border)
	sendButton.SetLabelColor(colors.Background)

	spacer := tview.NewBox().SetBackgroundColor(colors.Background)

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
	container.SetBackgroundColor(colors.Background)
	container.SetBorderColor(colors.Border)
	container.SetTitleColor(colors.Title)
	container.SetBorderPadding(0, 0, 0, 0)

	return container, methodDropdown, urlInput, sendButton
}

// createTabHeader creates a clickable tab header bar
func createTabHeader(tabs []string, colors *ColorManager, onTabClick func(int)) *tview.Flex {
	tabHeader := tview.NewFlex().SetDirection(tview.FlexColumn)
	tabHeader.SetBackgroundColor(colors.Background)

	// Tab titles and their active states
	activeTab := 0 // Default to first tab

	// Create tab buttons
	for i, tabTitle := range tabs {
		tabIndex := i // Capture for closure

		// Create individual tab as TextView (clickable)
		tab := tview.NewTextView()
		tab.SetBackgroundColor(colors.Background)
		tab.SetTextColor(colors.Foreground)
		tab.SetTextAlign(tview.AlignCenter)
		tab.SetBorder(false)

		// Set initial text based on whether it's active
		if i == activeTab {
			tab.SetText(tabTitle)
			tab.SetTextColor(colors.Title)
			tab.SetBackgroundColor(colors.Background)
		} else {
			tab.SetText(tabTitle)
			tab.SetBackgroundColor(colors.Background)
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
			separator.SetBackgroundColor(colors.Background)
			separator.SetTextColor(colors.Foreground)
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
	colors *ColorManager,
) {
	// Tab titles
	tabs = []string{"Body", "Auth", "Query", "Headers"}

	// Update each tab's appearance based on whether it's active
	for i := 0; i < len(tabs); i++ {
		// Find the tab TextView (skip separators)
		tabIndex := i * 2 // Every other item is a tab (alternating with separators)
		if tabIndex < tabHeader.GetItemCount() {
			if tab, ok := tabHeader.GetItem(tabIndex).(*tview.TextView); ok {
				if i == activeTabIndex {
					// Active tab - use active tab color and background
					tab.SetText(tabs[i])
					tab.SetTextColor(colors.ActiveTab)
					tab.SetBackgroundColor(colors.Selection)
				} else {
					// Inactive tab - use regular foreground color and background
					tab.SetText(fmt.Sprintf(" %s ", tabs[i]))
					tab.SetTextColor(colors.Foreground)
					tab.SetBackgroundColor(colors.Background)
				}
			}
		}
	}
}

// updateResponseTabHeader updates the active tab indicator in the response tab header
func updateResponseTabHeader(
	tabHeader *tview.Flex,
	activeTabIndex int,
	colors *ColorManager,
) {
	tabs := []string{"Preview", "Headers", "Cookies", "Timeline"}

	// Update each tab's appearance based on whether it's active
	for i := 0; i < len(tabs); i++ {
		// Find the tab TextView (skip separators)
		tabIndex := i * 2 // Every other item is a tab (alternating with separators)
		if tabIndex < tabHeader.GetItemCount() {
			if tab, ok := tabHeader.GetItem(tabIndex).(*tview.TextView); ok {
				if i == activeTabIndex {
					// Active tab - use active tab color and background
					tab.SetText(tabs[i])
					tab.SetTextColor(colors.ActiveTab)
					tab.SetBackgroundColor(colors.Selection)
				} else {
					// Inactive tab - use regular foreground color and background
					tab.SetText(fmt.Sprintf(" %s ", tabs[i]))
					tab.SetTextColor(colors.Foreground)
					tab.SetBackgroundColor(colors.Background)
				}
			}
		}
	}
}

// createAuthTab creates placeholder authentication tab content
func createAuthTab(colors *ColorManager) *tview.TextView {
	// Auth panel
	authPanel := createPanel("", colors, &PanelOptions{HasBorder: &[]bool{true}[0], BorderColor: &colors.Background})
	authPanel.SetText("Authentication settings will go here.\n\n• Bearer Token\n• Basic Auth\n• API Key\n• OAuth (future)")
	return authPanel
}

// createQueryTab creates placeholder query parameters tab content
func createQueryTab(colors *ColorManager) *tview.TextView {
	// Query Parameters panel
	queryPanel := createPanel("", colors, &PanelOptions{HasBorder: &[]bool{true}[0], BorderColor: &colors.Background})
	queryPanel.SetText("URL Query Parameters editor will go here.\n\n• Key-Value pairs\n• Add/Remove parameters\n• Bulk import\n• Templates (future)")
	return queryPanel
}

// Global variables for headers management
var currentHeadersTab *tview.Flex

var currentHeaderRows []*HeaderRow

var currentHeadersList *tview.Flex

// Global variables for environment variables management
var currentEnvRows []*EnvVarRow

var currentEnvVarsList *tview.Flex

var rowHeight int = 1

// HeaderRow represents a single header key-value pair in the UI
type HeaderRow struct {
	KeyInput     *tview.InputField
	ValueInput   *tview.InputField
	DeleteButton *tview.Button
	Row          *tview.Flex
}

// EnvVarRow represents a single environment variable key-value pair in the UI
type EnvVarRow struct {
	KeyInput     *tview.InputField
	ValueInput   *tview.InputField
	DeleteButton *tview.Button
	Row          *tview.Flex
}

// createHeadersTabWithData creates headers tab with initial data
func createHeadersTabWithData(colors *ColorManager,
	initialHeaders map[string]string,
	saveCallback func(),
	focusSetter func(tview.Primitive),
) *tview.Flex {
	headersContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	headersContainer.SetBackgroundColor(colors.Background)
	headersContainer.SetBorder(true)
	headersContainer.SetBorderColor(colors.Background)
	headersContainer.SetTitleColor(colors.Title)
	headersContainer.SetBackgroundColor(colors.Background)

	// Scrollable area for header entries
	headersList := tview.NewFlex().SetDirection(tview.FlexRow)
	headersList.SetBackgroundColor(colors.Background)

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
			addHeaderRow(headersList, colors, refreshHeadersUI, saveCallback, focusSetter)
		}
	}

	// Add initial rows based on data
	if initialHeaders != nil && len(initialHeaders) > 0 {
		for key, value := range initialHeaders {
			addHeaderRowWithData(headersList, colors, key, value, refreshHeadersUI, saveCallback, focusSetter)
		}
	}
	// Always add at least one empty row
	addHeaderRow(headersList, colors, refreshHeadersUI, saveCallback, focusSetter)

	// Add button row at the top
	buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonRow.SetBackgroundColor(colors.Background)

	addButton := createButton(" Add Header ", colors)
	addButton.SetBackgroundColorActivated(colors.Foreground)
	addButton.SetLabelColor(colors.Background)
	addButton.SetSelectedFunc(func() {
		addHeaderRow(headersList, colors, refreshHeadersUI, saveCallback, focusSetter)
	})

	// Delete all button
	deleteAllButton := createButton(" Delete All ", colors)
	deleteAllButton.SetBorder(false)
	deleteAllButton.SetStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
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
	spacer := tview.NewBox().SetBackgroundColor(colors.Background)
	headersContainer.AddItem(spacer, 1, 0, false)

	headersContainer.AddItem(headersList, 0, 1, false)

	return headersContainer
}

// addHeaderRow adds a new key-value header input row to the headers list
func addHeaderRow(headersList *tview.Flex,
	colors *ColorManager,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(colors.Background)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(colors.Background)
	keyInput.SetFieldBackgroundColor(colors.Background)
	keyInput.SetFieldTextColor(colors.Foreground)
	keyInput.SetLabelColor(colors.Foreground)
	keyInput.SetPlaceholder("Header name")
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	keyInput.SetFieldBackgroundColor(colors.Background)
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
	valueInput.SetBackgroundColor(colors.Background)
	valueInput.SetFieldBackgroundColor(colors.Background)
	valueInput.SetFieldTextColor(colors.Foreground)
	valueInput.SetLabelColor(colors.Foreground)
	valueInput.SetPlaceholder("Header value")
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	valueInput.SetFieldBackgroundColor(colors.Background)
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
	removeButton.SetBackgroundColor(colors.Background)
	removeButton.SetLabelColor(colors.Foreground)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))

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
	buttonContainer.SetBackgroundColor(colors.Background)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentHeaderRows = append(currentHeaderRows, headerRow)
	headersList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	separator := tview.NewBox().SetBackgroundColor(colors.Background)
	separator.SetBorder(false)
	headersList.AddItem(separator, 1, 0, false)

	// Update cycles if needed
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// addHeaderRowWithData adds a header row with pre-filled data
func addHeaderRowWithData(headersList *tview.Flex,
	colors *ColorManager,
	key, value string,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(colors.Background)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(colors.Background)
	keyInput.SetFieldBackgroundColor(colors.Background)
	keyInput.SetFieldTextColor(colors.Foreground)
	keyInput.SetLabelColor(colors.Foreground)
	keyInput.SetPlaceholder("Header name")
	keyInput.SetText(key)
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	keyInput.SetFieldBackgroundColor(colors.Background)
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
	valueInput.SetBackgroundColor(colors.Background)
	valueInput.SetFieldBackgroundColor(colors.Background)
	valueInput.SetFieldTextColor(colors.Foreground)
	valueInput.SetLabelColor(colors.Foreground)
	valueInput.SetPlaceholder("Header value")
	valueInput.SetText(value)
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	valueInput.SetFieldBackgroundColor(colors.Background)
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
	removeButton.SetBackgroundColor(colors.Background)
	removeButton.SetLabelColor(colors.Foreground)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))

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
	buttonContainer.SetBackgroundColor(colors.Background)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentHeaderRows = append(currentHeaderRows, headerRow)
	headersList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	separator := tview.NewBox().SetBackgroundColor(colors.Background)
	separator.SetBorder(false)
	headersList.AddItem(separator, 1, 0, false)

	// Update cycles if needed
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// addEnvVarRowWithData adds an environment variable row with pre-filled data
func addEnvVarRowWithData(
	variablesList *tview.Flex,
	colors *ColorManager,
	key, value string,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(colors.Background)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(colors.Background)
	keyInput.SetFieldBackgroundColor(colors.Background)
	keyInput.SetFieldTextColor(colors.Foreground)
	keyInput.SetLabelColor(colors.Foreground)
	keyInput.SetPlaceholder("Variable name")
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	keyInput.SetText(key)
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	if saveCallback != nil {
		keyInput.SetChangedFunc(func(text string) {
			saveCallback()
		})
		keyInput.SetBlurFunc(func() {
			saveCallback()
		})
	}

	valueInput := tview.NewInputField()
	valueInput.SetBackgroundColor(colors.Background)
	valueInput.SetFieldBackgroundColor(colors.Background)
	valueInput.SetFieldTextColor(colors.Foreground)
	valueInput.SetLabelColor(colors.Foreground)
	valueInput.SetPlaceholder("Variable value")
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	valueInput.SetText(value)
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	if saveCallback != nil {
		valueInput.SetChangedFunc(func(text string) {
			saveCallback()
		})
		valueInput.SetBlurFunc(func() {
			saveCallback()
		})
	}

	removeButton := tview.NewButton(config.C.UI.HeaderRemoveIcon)
	removeButton.SetBackgroundColor(colors.Background)
	removeButton.SetLabelColor(colors.Foreground)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))

	envVarRow := &EnvVarRow{
		KeyInput:     keyInput,
		ValueInput:   valueInput,
		DeleteButton: removeButton,
		Row:          row,
	}

	removeButton.SetSelectedFunc(func() {
		// Find and remove this row
		for i, r := range currentEnvRows {
			if r == envVarRow {
				currentEnvRows = append(currentEnvRows[:i], currentEnvRows[i+1:]...)
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
			if len(currentEnvRows) > 0 {
				firstRow := currentEnvRows[0]
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
			for i, row := range currentEnvRows {
				if row == envVarRow {
					currentIndex = i
					break
				}
			}

			if currentIndex > 0 {
				// Move to previous row's value input
				prevRow := currentEnvRows[currentIndex-1]
				focusSetter(prevRow.ValueInput)
			} else {
				// First row, cycle to last row's value input
				lastRow := currentEnvRows[len(currentEnvRows)-1]
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
	buttonContainer.SetBackgroundColor(colors.Background)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentEnvRows = append(currentEnvRows, envVarRow)
	variablesList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	separator := tview.NewBox().SetBackgroundColor(colors.Background)
	separator.SetBorder(false)
	variablesList.AddItem(separator, 1, 0, false)

	// Update cycles if needed
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// addEnvVarRow adds a new environment variable row to the variables list
func addEnvVarRow(variablesList *tview.Flex,
	colors *ColorManager,
	refreshUI func(),
	saveCallback func(),
	focusSetter func(tview.Primitive),
) {
	row := tview.NewFlex().SetDirection(tview.FlexColumn)
	row.SetBackgroundColor(colors.Background)

	keyInput := tview.NewInputField()
	keyInput.SetBackgroundColor(colors.Background)
	keyInput.SetFieldBackgroundColor(colors.Background)
	keyInput.SetFieldTextColor(colors.Foreground)
	keyInput.SetLabelColor(colors.Foreground)
	keyInput.SetPlaceholder("Variable name")
	keyInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	keyInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	keyInput.SetFieldBackgroundColor(colors.Background)
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
	valueInput.SetBackgroundColor(colors.Background)
	valueInput.SetFieldBackgroundColor(colors.Background)
	valueInput.SetFieldTextColor(colors.Foreground)
	valueInput.SetLabelColor(colors.Foreground)
	valueInput.SetPlaceholder("Variable value")
	valueInput.SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Background).Foreground(hexToColor("#4A5053")))
	valueInput.SetFieldStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))
	valueInput.SetFieldBackgroundColor(colors.Background)
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
	removeButton.SetBackgroundColor(colors.Background)
	removeButton.SetLabelColor(colors.Foreground)
	removeButton.SetBorder(false)
	removeButton.SetStyle(tcell.StyleDefault.Background(colors.Background).Foreground(colors.Foreground))

	envVarRow := &EnvVarRow{
		KeyInput:     keyInput,
		ValueInput:   valueInput,
		DeleteButton: removeButton,
		Row:          row,
	}

	removeButton.SetSelectedFunc(func() {
		// Find and remove this row
		for i, r := range currentEnvRows {
			if r == envVarRow {
				currentEnvRows = append(currentEnvRows[:i], currentEnvRows[i+1:]...)
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
			if len(currentEnvRows) > 0 {
				firstRow := currentEnvRows[0]
				focusSetter(firstRow.KeyInput)
			}
			return nil
		}
		return event
	})

	// Handle Tab to add new variable row when on the last value input
	valueInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			// Check if this is the last value input
			if len(currentEnvRows) > 0 && currentEnvRows[len(currentEnvRows)-1] == envVarRow {
				// Cycle back to first key input
				firstRow := currentEnvRows[0]
				focusSetter(firstRow.KeyInput)
				return nil
			}
		}
		return event
	})

	// Wrap button in a container to match row height
	buttonContainer := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonContainer.SetBackgroundColor(colors.Background)
	buttonContainer.AddItem(nil, 0, 1, false)
	buttonContainer.AddItem(removeButton, 1, 0, false)
	buttonContainer.AddItem(nil, 0, 1, false)

	row.AddItem(keyInput, 0, 1, false)
	row.AddItem(valueInput, 0, 1, false)
	row.AddItem(buttonContainer, 4, 0, false)

	currentEnvRows = append(currentEnvRows, envVarRow)
	variablesList.AddItem(row, rowHeight, 0, false)

	// Add separator line between rows
	separator := tview.NewBox().SetBackgroundColor(colors.Background)
	separator.SetBorder(false)
	variablesList.AddItem(separator, 1, 0, false)

	// Update cycles if needed
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// setHeadersInUI populates the UI with the given headers
func setHeadersInUI(colors *ColorManager, headers map[string]string, saveCallback func(), focusSetter func(tview.Primitive)) {
	// Clear existing rows
	currentHeaderRows = []*HeaderRow{}

	if currentHeadersList != nil {
		currentHeadersList.Clear()

		// Add rows for each header (sorted by key for consistent ordering)
		var keys []string
		for key := range headers {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := headers[key]
			addHeaderRowWithData(currentHeadersList, colors, key, value, func() {
				// Refresh function - for now just rebuild the list
				setHeadersInUI(colors, getHeadersFromUI(), saveCallback, focusSetter)
			}, saveCallback, focusSetter)
		}

		// Always add one empty row
		addHeaderRow(currentHeadersList, colors, func() {
			setHeadersInUI(colors, getHeadersFromUI(), saveCallback, focusSetter)
		}, saveCallback, focusSetter)
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

// getEnvVarsFromUI extracts environment variables from the current UI state
func getEnvVarsFromUI() map[string]string {
	variables := make(map[string]string)
	for _, row := range currentEnvRows {
		key := strings.TrimSpace(row.KeyInput.GetText())
		value := strings.TrimSpace(row.ValueInput.GetText())
		if key != "" {
			variables[key] = value
		}
	}
	return variables
}

// setEnvVarsInUI populates the UI with the given environment variables
func setEnvVarsInUI(colors *ColorManager, variables map[string]string, saveCallback func(), focusSetter func(tview.Primitive)) {
	// Clear existing rows
	currentEnvRows = []*EnvVarRow{}

	if currentEnvVarsList != nil {
		currentEnvVarsList.Clear()

		// Add rows for each variable (sorted by key for consistent ordering)
		var keys []string
		for key := range variables {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			value := variables[key]
			addEnvVarRowWithData(currentEnvVarsList, colors, key, value, func() {
				// Refresh function - for now just rebuild the list
				setEnvVarsInUI(colors, getEnvVarsFromUI(), saveCallback, focusSetter)
			}, saveCallback, focusSetter)
		}

		// Always add one empty row
		addEnvVarRow(currentEnvVarsList, colors, func() {
			setEnvVarsInUI(colors, getEnvVarsFromUI(), saveCallback, focusSetter)
		}, saveCallback, focusSetter)
	}

	// Update cycles if needed
	if headersCycle != nil {
		headersCycle.UpdateInputs()
	}
}

// createResponseInfoBar creates the response information bar
func createResponseInfoBar(colors *ColorManager, resp *HTTPResponse, lastTime *time.Time, copyCallback func()) (*tview.Flex, *tview.TextView, int) {
	infoBar := tview.NewFlex().SetDirection(tview.FlexColumn)
	infoBar.SetBackgroundColor(colors.Background)

	var timeText *tview.TextView

	if resp == nil {
		// No response yet
		noResponseText := tview.NewTextView()
		noResponseText.SetBackgroundColor(colors.Background)
		noResponseText.SetTextColor(colors.Foreground)
		noResponseText.SetText("No response")
		infoBar.AddItem(noResponseText, 0, 1, false)
		return infoBar, nil, 12 // "No response" is 11 chars, plus some padding
	}

	// Determine status background color
	var statusBgColor tcell.Color
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		statusBgColor = colors.Success
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		statusBgColor = colors.Warning
	} else if resp.StatusCode >= 400 {
		statusBgColor = colors.Error
	} else {
		statusBgColor = colors.Background // For 1xx or unknown
	}

	// Copy button
	if copyCallback != nil {
		copyButton := tview.NewButton("📋")
		copyButton.SetBackgroundColor(colors.Background)
		copyButton.SetLabelColor(colors.Foreground)
		copyButton.SetSelectedFunc(copyCallback)
		infoBar.AddItem(copyButton, 3, 0, false)
	}

	// Status code and status
	statusText := tview.NewTextView()
	statusText.SetBackgroundColor(statusBgColor)
	statusText.SetTextColor(colors.Title)
	statusText.SetText(fmt.Sprintf("%d", resp.StatusCode))
	statusText.SetTextAlign(tview.AlignCenter)
	infoBar.AddItem(statusText, 5, 0, false)

	// Size
	minSizeWidth := 9
	sizeText := tview.NewTextView()
	sizeText.SetBackgroundColor(colors.Border)
	sizeText.SetTextColor(colors.Foreground)
	sizeStr := fmt.Sprintf(" %s", humanize.IBytes(uint64(resp.BodySize)))
	sizeText.SetText(sizeStr)
	sizeText.SetTextAlign(tview.AlignCenter)
	sizeWidth := utf8.RuneCountInString(sizeStr)
	if sizeWidth < minSizeWidth {
		sizeWidth = minSizeWidth
	}
	infoBar.AddItem(sizeText, sizeWidth, 0, false)

	// Duration
	minDurationWidth := 7
	durationText := tview.NewTextView()
	durationText.SetBackgroundColor(colors.Border)
	durationText.SetTextColor(colors.Foreground)
	durationStr := " -"
	if resp.Duration > 0 {
		durationStr = fmt.Sprintf(" %v", resp.Duration.Round(time.Millisecond))
	}
	durationText.SetText(durationStr)
	durationText.SetTextAlign(tview.AlignCenter)
	durationWidth := utf8.RuneCountInString(durationStr)
	if durationWidth < minDurationWidth {
		durationWidth = minDurationWidth
	}
	infoBar.AddItem(durationText, durationWidth, 0, false)

	// Time
	minWidth := 16
	timeText = tview.NewTextView()
	timeText.SetBackgroundColor(colors.Border)
	timeText.SetTextColor(colors.Foreground)
	timeStr := " -"
	if lastTime != nil {
		timeStr = fmt.Sprintf(" %s", humanize.Time(*lastTime))
	}
	timeText.SetText(timeStr)
	timeText.SetTextAlign(tview.AlignLeft)
	timeWidth := utf8.RuneCountInString(timeStr)
	if timeWidth < minWidth {
		timeWidth = minWidth
	}
	infoBar.AddItem(timeText, timeWidth, 0, false)

	totalWidth := 5 + sizeWidth + durationWidth + timeWidth // status 5, size dynamic, duration dynamic, time dynamic
	return infoBar, timeText, totalWidth
}

// createRequestDataTabs creates the request data tabs interface
func createRequestDataTabs(bodyViewPanel *tview.TextView, bodyEditPanel *tview.TextArea, colors *ColorManager, saveCallback func(), focusSetter func(tview.Primitive), tabIndexSetter func(int), panelFocusSetter func(tview.Primitive)) (*tview.Flex, *tview.Pages, *tview.Flex, *tview.Flex, *tview.TextView, *tview.TextView, *tview.TextView, *tview.Flex) {
	// Create main request data container
	requestDataTabs := tview.NewFlex().SetDirection(tview.FlexRow)
	requestDataTabs.SetBackgroundColor(colors.Background)
	requestDataTabs.SetBorder(true)
	requestDataTabs.SetBorderColor(colors.Border)
	requestDataTabs.SetTitle(" Request ")
	requestDataTabs.SetTitleColor(colors.Title)

	// Create tab header
	tabHeader := createTabHeader([]string{"Body", "Auth", "Query", "Headers"}, colors, func(index int) {
		if tabIndexSetter != nil {
			tabIndexSetter(index)
		}
	})

	// Create tab pages
	tabPages := tview.NewPages()
	tabPages.SetBackgroundColor(colors.Background)

	// Create body container (switch between view and edit)
	bodyContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	bodyContainer.SetBackgroundColor(colors.Background)
	bodyContainer.AddItem(bodyViewPanel, 0, 1, false)

	// Create auth tab
	authTab := createAuthTab(colors)

	// Create query tab
	queryTab := createQueryTab(colors)

	// Create headers tab
	headersTab := createHeadersTabWithData(colors, nil, saveCallback, focusSetter)

	// Add pages
	tabPages.AddPage("body", bodyContainer, true, true)
	tabPages.AddPage("auth", authTab, true, false)
	tabPages.AddPage("query", queryTab, true, false)
	tabPages.AddPage("headers", headersTab, true, false)

	// Add to main container
	requestDataTabs.AddItem(tabHeader, 1, 0, false)
	requestDataTabs.AddItem(tabPages, 0, 1, false)

	return requestDataTabs, tabPages, bodyContainer, tabHeader, bodyViewPanel, authTab, queryTab, headersTab
}

// createResponseTabs creates the response tabs interface
func createResponseTabs(colors *ColorManager, resp *HTTPResponse, lastTime *time.Time, tabCallback func(int), copyCallback func()) (*tview.Flex, *tview.Pages, *tview.Flex, *tview.Flex, *tview.Flex, *tview.TextView, tview.Primitive, *tview.TextView, *tview.TextView, *tview.TextView) {
	// Create main response container
	response := tview.NewFlex().SetDirection(tview.FlexRow)
	response.SetBackgroundColor(colors.Background)
	response.SetBorder(true)
	response.SetBorderColor(colors.Border)
	response.SetTitle(" Response ")
	response.SetTitleColor(colors.Title)

	// Create tab pages first
	responsePages := tview.NewPages()
	responsePages.SetBackgroundColor(colors.Background)

	// Create tab header
	responseTabHeader := createTabHeader([]string{"Preview", "Headers", "Cookies", "Timeline"}, colors, func(index int) {
		// Switch to the selected response tab
		pageNames := []string{"preview", "headers", "cookies", "timeline"}
		if index >= 0 && index < len(pageNames) {
			responsePages.SwitchToPage(pageNames[index])
		}
		if tabCallback != nil {
			tabCallback(index)
		}
	})

	// Create info bar
	responseInfoBar, responseTimeText, infoBarWidth := createResponseInfoBar(colors, resp, lastTime, copyCallback)

	// Create top row with tab header and info bar
	topRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	topRow.SetBackgroundColor(colors.Background)
	topRow.AddItem(responseTabHeader, 0, 1, false)
	spacer := tview.NewBox().SetBackgroundColor(colors.Background)
	topRow.AddItem(spacer, 0, 1, false)
	topRow.AddItem(responseInfoBar, infoBarWidth, 0, false)

	// Create individual tab panels
	responsePreviewPanel := createPanel(" Response Body ", colors, &PanelOptions{
		HasBorder: &[]bool{false}[0],
	})

	responseHeadersPanel := tview.NewTable()
	responseHeadersPanel.SetBorders(false)
	responseHeadersPanel.SetBackgroundColor(colors.Background)
	responseHeadersPanel.SetBorderColor(colors.Background)
	responseHeadersPanel.SetFixed(1, 0)
	responseHeadersPanel.SetSelectable(true, false)
	responseHeadersPanel.SetBorderPadding(1, 0, 1, 0)
	responseHeadersPanel.SetTitleColor(colors.Title)

	responseCookiesPanel := tview.NewTextView()
	responseCookiesPanel.SetBackgroundColor(colors.Background)
	responseCookiesPanel.SetTextColor(colors.Foreground)
	responseCookiesPanel.SetBorder(true)
	responseCookiesPanel.SetTitleColor(colors.Title)
	responseCookiesPanel.SetBorderColor(colors.Background)

	responseTimelinePanel := tview.NewTextView()
	responseTimelinePanel.SetBackgroundColor(colors.Background)
	responseTimelinePanel.SetTextColor(colors.Foreground)
	responseTimelinePanel.SetBorder(true)
	responseTimelinePanel.SetTitleColor(colors.Title)
	responseTimelinePanel.SetBorderColor(colors.Background)

	// Add pages
	responsePages.AddPage("preview", responsePreviewPanel, true, true)
	responsePages.AddPage("headers", responseHeadersPanel, true, false)
	responsePages.AddPage("cookies", responseCookiesPanel, true, false)
	responsePages.AddPage("timeline", responseTimelinePanel, true, false)

	// Add to main container
	response.AddItem(topRow, 1, 0, false)
	response.AddItem(responsePages, 0, 1, true)

	// Initialize response tabs with initial data
	updateResponseTabs(resp, lastTime, response, responseTabHeader, &responseInfoBar, &responseTimeText, &lastTime, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel, colors, nil)

	return response, responsePages, responseTabHeader, responseInfoBar, responseInfoBar, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel, responseTimeText
}

// createModal creates a centered modal dialog
func createModal(p tview.Primitive, width, height int, backgroundColor tcell.Color) tview.Primitive {
	// Create a background box for the content area to ensure solid coverage
	contentBackground := tview.NewBox().SetBackgroundColor(backgroundColor)

	// Stack the content on top of the background box
	contentArea := tview.NewFlex().SetDirection(tview.FlexRow)
	contentArea.AddItem(contentBackground, 0, 1, false)
	contentArea.AddItem(p, height, 1, true)

	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(contentArea, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
	modal.SetBackgroundColor(backgroundColor)
	return modal
}

// URLVariableInput is a dual-mode input component for URLs with environment variables
type URLVariableInput struct {
	*tview.Pages
	viewMode       *tview.TextView
	editMode       *tview.InputField
	currentMode    string // "view" or "edit"
	rawText        string // The actual {{variable}} text
	variables      map[string]string
	onChanged      func(string)
	onEnterPressed func()
	colors         *ColorManager
	variableRegex  *regexp.Regexp
	app            *tview.Application
}

// NewURLVariableInput creates a new dual-mode URL input component
func NewURLVariableInput(colors *ColorManager, app *tview.Application) *URLVariableInput {
	variableRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)

	// Create view mode component (TextView)
	viewMode := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(false)

	viewMode.SetBackgroundColor(colors.Background)
	viewMode.SetTextColor(colors.Foreground)
	viewMode.SetBorderPadding(0, 0, 1, 1)

	// Create edit mode component (InputField)
	editMode := tview.NewInputField()
	editMode.SetBackgroundColor(colors.Background)
	editMode.SetFieldBackgroundColor(colors.Background)
	editMode.SetFieldTextColor(colors.Foreground)
	editMode.SetBorder(false)

	// Ensure edit mode is properly focusable
	editMode.SetFocusFunc(func() {
		editMode.SetFieldBackgroundColor(colors.Selection)
	})

	editMode.SetBlurFunc(func() {
		editMode.SetFieldBackgroundColor(colors.Background)
	})

	// Create Pages container
	pages := tview.NewPages()
	pages.SetBackgroundColor(colors.Background)
	pages.AddPage("view", viewMode, true, true)
	pages.AddPage("edit", editMode, true, false)

	input := &URLVariableInput{
		Pages:         pages,
		viewMode:      viewMode,
		editMode:      editMode,
		currentMode:   "view",
		rawText:       "",
		variables:     make(map[string]string),
		colors:        colors,
		variableRegex: variableRegex,
		app:           app,
	}

	// Set up event handlers
	editMode.SetChangedFunc(func(text string) {
		input.rawText = text
		if input.onChanged != nil {
			input.onChanged(text)
		}
	})

	editMode.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter && input.onEnterPressed != nil {
			input.onEnterPressed()
		} else if key == tcell.KeyEsc {
			input.switchToViewMode()
		}
	})

	// Set up view mode click to enter edit mode
	viewMode.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter {
			input.switchToEditMode()
			return nil
		}
		return event
	})

	// Make view mode focusable and handle focus properly
	viewMode.SetFocusFunc(func() {
		// When view mode gets focus, ensure it's properly highlighted
		viewMode.SetBackgroundColor(colors.Selection)
	})

	viewMode.SetBlurFunc(func() {
		// When view mode loses focus, reset background
		viewMode.SetBackgroundColor(colors.Background)
	})

	return input
}

// switchToViewMode switches to view mode, showing rendered variables
func (u *URLVariableInput) switchToViewMode() {
	u.currentMode = "view"
	u.viewMode.SetBackgroundColor(u.colors.Background)
	u.Pages.SwitchToPage("view")
	u.updateViewMode()
}

// switchToEditMode switches to edit mode, showing raw text
func (u *URLVariableInput) switchToEditMode() {
	u.currentMode = "edit"
	u.editMode.SetBackgroundColor(u.colors.Background)
	u.editMode.SetFieldBackgroundColor(u.colors.Background)
	u.editMode.SetText(u.rawText)
	u.Pages.SwitchToPage("edit")
}

// Focus delegates focus to the appropriate child component
func (u *URLVariableInput) Focus(delegate func(p tview.Primitive)) {
	// Let the Pages container handle focus for the visible page
	u.Pages.Focus(delegate)
}

// HasFocus returns whether the component or its children have focus
func (u *URLVariableInput) HasFocus() bool {
	if u.currentMode == "edit" {
		return u.editMode.HasFocus()
	}
	return u.viewMode.HasFocus()
}

// InputHandler delegates to the Pages container
func (u *URLVariableInput) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return u.Pages.InputHandler()
}

// MouseHandler delegates to the Pages container
func (u *URLVariableInput) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return u.Pages.MouseHandler()
}

// updateViewMode renders the text with variables highlighted in view mode
func (u *URLVariableInput) updateViewMode() {
	if u.rawText == "" {
		u.viewMode.SetText("")
		return
	}

	// Find all variable positions
	matches := u.variableRegex.FindAllStringIndex(u.rawText, -1)
	if len(matches) == 0 {
		u.viewMode.SetText(u.rawText)
		return
	}

	// Build result with proper spacing
	var result strings.Builder
	lastEnd := 0

	for i, match := range matches {
		start, end := match[0], match[1]

		// Add text before this variable
		result.WriteString(u.rawText[lastEnd:start])

		// Extract variable name (remove {{ and }})
		varName := u.rawText[start+2 : end-2]

		// Render variable with background color
		result.WriteString(fmt.Sprintf("[%s:%s:-]%s[-:-:-]",
			config.C.Theme.DropdownFocusedBackground,
			config.C.Theme.BorderFocusColor,
			varName))

		// Add space only if next character is another variable (no text between)
		if i < len(matches)-1 && end == matches[i+1][0] {
			result.WriteString(" ")
		}

		lastEnd = end
	}

	// Add remaining text after last variable
	result.WriteString(u.rawText[lastEnd:])

	u.viewMode.SetText(result.String())
}

// SetText sets the raw text and updates both modes
func (u *URLVariableInput) SetText(text string) {
	u.rawText = text
	if u.currentMode == "view" {
		u.updateViewMode()
	} else {
		u.editMode.SetText(text)
	}
	if u.onChanged != nil {
		u.onChanged(text)
	}
}

// GetText returns the current raw text
func (u *URLVariableInput) GetText() string {
	return u.rawText
}

// SetChangedFunc sets the callback for when text changes
func (u *URLVariableInput) SetChangedFunc(callback func(string)) {
	u.onChanged = callback
}

// SetDoneFunc sets the callback for when Enter is pressed
func (u *URLVariableInput) SetDoneFunc(callback func()) {
	u.onEnterPressed = callback
}
