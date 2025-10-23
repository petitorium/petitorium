package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// setupUIComponents creates and configures all UI components
func setupUIComponents(
	backgroundColor,
	foregroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	selectionBackgroundColor,
	activeTabColor,
	buttonSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) (
	*tview.TextView, *tview.TreeNode, *tview.Flex, *tview.DropDown, *tview.InputField, *tview.Button,
	*tview.TextView, *tview.TextArea, *tview.TextView, *tview.TextView, *tview.TreeView,
) {
	// Create header panel
	header := createPanel(" Petitorium ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create root node for tree
	rootNode := tview.NewTreeNode("").SetSelectable(false)

	// Create unified method+URL+Send bar
	methodURLBar, methodDropdown, urlInput, sendButton := createMethodURLBar(
		"",
		backgroundColor,
		borderColor,
		titleColor,
		foregroundColor,
		selectionBackgroundColor,
		activeTabColor,
		buttonSelectedColor,
		dropdownFocusedBackgroundColor,
	)

	// Create both view and edit panels for body
	bodyViewPanel := createPanel(" Body (VIEW) ", backgroundColor, backgroundColor, titleColor, foregroundColor)
	bodyEditPanel := createTextArea(" Body (EDIT) ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create response panel
	response := createPanel(" Response ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create footer panel
	footer := createPanel("", backgroundColor, borderColor, titleColor, foregroundColor)
	footer.SetText(" (Tab) Cycle Focus | Body Tab: (i) Insert (Esc) Normal (hjkl) Nav | (F4) External Editor | (q) Quit | Collections: (n) New Collection | (r) New Request | (R) Rename | (m) Move Item | (d) Delete")

	// Create collections tree view
	collectionsTreeView := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	collectionsTreeView.
		SetGraphics(false).
		SetTopLevel(0)

	collectionsTreeView.
		SetBorder(true).
		SetTitle(" Collections ").
		SetBackgroundColor(backgroundColor).
		SetBorderColor(borderColor).
		SetTitleColor(titleColor).
		SetBorderPadding(0, 0, 0, 0)

	return header, rootNode, methodURLBar, methodDropdown, urlInput, sendButton, bodyViewPanel, bodyEditPanel, response, footer, collectionsTreeView
}

// setupRequestPanel creates the unified request panel
func setupRequestPanel(methodURLBar, requestDataTabs *tview.Flex, backgroundColor tcell.Color) *tview.Flex {
	requestPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(methodURLBar, 3, 0, false).
		AddItem(requestDataTabs, 0, 1, false)
	requestPanel.SetBorder(false)
	requestPanel.SetBackgroundColor(backgroundColor)

	return requestPanel
}

// setupRightSide creates the right side layout
func setupRightSide(requestPanel *tview.Flex, response *tview.TextView) *tview.Flex {
	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(requestPanel, 0, 1, false).
		AddItem(response, 0, 1, false)

	return rightSide
}

// setupLayout creates the main grid layout
func setupLayout(header, footer *tview.TextView, collectionsTreeView *tview.TreeView, rightSide *tview.Flex) *tview.Grid {
	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(collectionsTreeView, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	return grid
}

// setupPanels creates the panels slice for cycling
func setupPanels(collectionsTreeView *tview.TreeView, methodURLBar *tview.Flex, requestDataTabs *tview.Flex, response *tview.TextView) []tview.Primitive {
	return []tview.Primitive{collectionsTreeView, methodURLBar, requestDataTabs, response}
}

// setupCycles initializes the navigation cycles
func setupCycles(panels []tview.Primitive) {
	mainCycle = &MainCycle{
		panels:  panels,
		current: 0, // start with collections
	}

	headersCycle = &HeadersCycle{
		inputs:  []tview.Primitive{},
		current: 0,
		parent:  mainCycle,
	}
}

// syncBodyContent syncs body content between view and edit panels
func syncBodyContent(content string, bodyEditMode bool, bodyEditPanel *tview.TextArea, bodyViewPanel *tview.TextView) {
	if bodyEditMode {
		bodyEditPanel.SetText(content, false)
	} else {
		bodyViewPanel.Clear()
		if content != "" {
			// Format content with syntax highlighting
			formattedContent := formatBodyContent(content)

			// Set content with proper handling
			bodyViewPanel.SetText(formattedContent)
			bodyViewPanel.SetTextAlign(tview.AlignLeft)

			// Force proper rendering and scrolling
			go func() {
				// Small delay to ensure content is set before scrolling
				bodyViewPanel.ScrollToEnd()
				bodyViewPanel.ScrollToBeginning()
			}()
		} else {
			bodyViewPanel.SetText("")
			bodyViewPanel.SetTextAlign(tview.AlignLeft)
		}
	}
}

// switchBodyMode switches between view and edit modes for body content
func switchBodyMode(
	bodyEditMode *bool,
	bodyContainer *tview.Flex,
	bodyViewPanel *tview.TextView,
	bodyEditPanel *tview.TextArea,
	currentBodyContent string,
	currentRequest **workspace.Request,
	currentSelectedNode **tview.TreeNode,
	saveCurrentRequest func(),
) {
	*bodyEditMode = !*bodyEditMode
	bodyContainer.Clear()

	if *bodyEditMode {
		// Switch to edit mode
		bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
		bodyEditPanel.SetText(currentBodyContent, false)
	} else {
		// Switch to view mode
		bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
		// Update body content from edit panel if we were editing
		if *currentRequest != nil {
			currentBodyContent = bodyEditPanel.GetText()
			(*currentRequest).Body = currentBodyContent
			if *currentSelectedNode != nil {
				(*currentSelectedNode).SetReference(**currentRequest)
				saveCurrentRequest()
			}
		}
		syncBodyContent(currentBodyContent, *bodyEditMode, bodyEditPanel, bodyViewPanel)
	}
}
