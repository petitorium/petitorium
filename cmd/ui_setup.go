package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// UIComponents holds all UI components for the application
type UIComponents struct {
	Header                *tview.TextView
	RootNode              *tview.TreeNode
	MethodURLBar          *tview.Flex
	MethodDropdown        *tview.DropDown
	URLInput              *tview.InputField
	SendButton            *tview.Button
	BodyViewPanel         *tview.TextView
	BodyEditPanel         *tview.TextArea
	Response              *tview.Flex
	Footer                *tview.TextView
	CollectionsTreeView   *tview.TreeView
	ResponsePages         *tview.Pages
	ResponseTabHeader     *tview.Flex
	ResponseInfoBar       *tview.Flex
	ResponsePreviewPanel  *tview.TextView
	ResponseHeadersPanel  *tview.TextView
	ResponseCookiesPanel  *tview.TextView
	ResponseTimelinePanel *tview.TextView
	EnvironmentPanel      *tview.Flex
	EnvDropdown           *tview.DropDown
	EnvConfigButton       *tview.Button
	EnvIndicatorButton    *CustomButton
}

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
) *UIComponents {
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

	// Create response tabs panel
	response, responsePages, responseTabHeader, _, responseInfoBar, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel := createResponseTabs(
		backgroundColor, borderColor, borderFocusColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor,
		nil, nil, // No initial response or last request time
	)

	// Create footer panel
	footer := createPanel("", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create environment panel with dropdown and config button
	environmentPanel, envDropdown, envConfigButton, envIndicatorButton := createEnvironmentPanel(
		backgroundColor,
		borderColor,
		borderFocusColor,
		titleColor,
		foregroundColor,
		selectionBackgroundColor,
		activeTabColor,
		buttonSelectedColor,
		dropdownFocusedBackgroundColor,
	)

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

	return &UIComponents{
		Header:                header,
		RootNode:              rootNode,
		MethodURLBar:          methodURLBar,
		MethodDropdown:        methodDropdown,
		URLInput:              urlInput,
		SendButton:            sendButton,
		BodyViewPanel:         bodyViewPanel,
		BodyEditPanel:         bodyEditPanel,
		Response:              response,
		Footer:                footer,
		CollectionsTreeView:   collectionsTreeView,
		ResponsePages:         responsePages,
		ResponseTabHeader:     responseTabHeader,
		ResponseInfoBar:       responseInfoBar,
		ResponsePreviewPanel:  responsePreviewPanel,
		ResponseHeadersPanel:  responseHeadersPanel,
		ResponseCookiesPanel:  responseCookiesPanel,
		ResponseTimelinePanel: responseTimelinePanel,
		EnvironmentPanel:      environmentPanel,
		EnvDropdown:           envDropdown,
		EnvConfigButton:       envConfigButton,
		EnvIndicatorButton:    envIndicatorButton,
	}
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
func setupRightSide(requestPanel *tview.Flex, response *tview.Flex) *tview.Flex {
	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(requestPanel, 0, 1, false).
		AddItem(response, 0, 1, false)

	return rightSide
}

// setupLayout creates the main grid layout
func setupLayout(header, footer *tview.TextView, collectionsTreeView *tview.TreeView, environmentPanel *tview.Flex, rightSide *tview.Flex) *tview.Grid {
	leftSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(environmentPanel, 3, 0, false).
		AddItem(collectionsTreeView, 0, 1, false)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(leftSide, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	return grid
}

// setupPanels creates the panels slice for cycling
func setupPanels(environmentPanel *tview.Flex, collectionsTreeView *tview.TreeView, methodURLBar *tview.Flex, requestDataTabs *tview.Flex, response *tview.TextView) []tview.Primitive {
	return []tview.Primitive{environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, response}
}

// setupCycles initializes the navigation cycles
func setupCycles(panels []tview.Primitive) {
	mainCycle = &MainCycle{
		panels:  panels,
		current: 1, // start with collections
	}

	headersCycle = &HeadersCycle{
		inputs:  []tview.Primitive{},
		current: 0,
		parent:  mainCycle,
	}
}
