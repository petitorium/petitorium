package cmd

import (
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// UIOrchestrator holds all UI components and state
type UIOrchestrator struct {
	App                   *tview.Application
	CollectionsData       *[]workspace.Collection
	DataManager           *DataManager
	EnvironmentsData      *[]workspace.Environment
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
	Header                *tview.TextView
	Pages                 *tview.Pages
	Grid                  *tview.Grid
	KeyManager            *KeyBindingManager

	// State variables
	CurrentSelectedNode            *tview.TreeNode
	CurrentRequest                 *workspace.Request
	ProgrammaticallyUpdatingMethod bool
	ProgrammaticallyUpdatingURL    bool
	TabPages                       *tview.Pages
	TabHeader                      *tview.Flex
	CurrentTabIndex                int
	EnviromentIndex                int
	CollectionsIndex               int
	URLBarIndex                    int
	RequestIndex                   int
	ResponseIndex                  int
	RequestDataTabs                *tview.Flex
	BodyContainer                  *tview.Flex
	MainCycle                      *MainCycle
	HeadersCycle                   *HeadersCycle
	RequestCycle                   *RequestCycle
	LastSelectedRequestNode        *tview.TreeNode
	BodyEditMode                   bool
	CurrentBodyContent             string
	RequestPanel                   *tview.Flex
	RightSide                      *tview.Flex
	LeftSide                       *tview.Flex
	CurrentFocus                   int
	SetPanelFocus                  func(int, bool)
	SetActiveBorder                func(element tview.Primitive)
	SetInactiveBorder              func(element tview.Primitive)
	SyncBodyContent                func(content string)
	SwitchBodyMode                 func()
}

// SetupUI initializes all UI components and layout
func SetupUI(collectionsData *[]workspace.Collection, dataManager *DataManager, environmentsData *[]workspace.Environment) (*UIOrchestrator, error) {
	backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor, activeTabColor, buttonSelectedColor, dropdownFocusedBackgroundColor := setupTheme()

	app := tview.NewApplication().EnableMouse(true)

	// Create all UI components
	ui := setupUIComponents(backgroundColor,
		foregroundColor,
		borderColor,
		borderFocusColor,
		titleColor,
		selectionBackgroundColor,
		activeTabColor,
		buttonSelectedColor,
		dropdownFocusedBackgroundColor,
	)

	header := ui.Header
	rootNode := ui.RootNode
	methodURLBar := ui.MethodURLBar
	methodDropdown := ui.MethodDropdown
	urlInput := ui.URLInput
	sendButton := ui.SendButton
	bodyViewPanel := ui.BodyViewPanel
	bodyEditPanel := ui.BodyEditPanel
	response := ui.Response
	footer := ui.Footer
	collectionsTreeView := ui.CollectionsTreeView
	responsePages := ui.ResponsePages
	responseTabHeader := ui.ResponseTabHeader
	responseInfoBar := ui.ResponseInfoBar
	responsePreviewPanel := ui.ResponsePreviewPanel
	responseHeadersPanel := ui.ResponseHeadersPanel
	responseCookiesPanel := ui.ResponseCookiesPanel
	responseTimelinePanel := ui.ResponseTimelinePanel
	environmentPanel := ui.EnvironmentPanel
	envDropdown := ui.EnvDropdown
	envConfigButton := ui.EnvConfigButton

	// Update environment dropdown with loaded environments
	updateEnvironmentDropdown(envDropdown, *environmentsData)

	// Variable declarations
	var currentSelectedNode *tview.TreeNode
	var currentRequest *workspace.Request
	var programmaticallyUpdatingMethod bool // Track programmatic updates
	var programmaticallyUpdatingURL bool    // Track programmatic URL updates
	var tabPages *tview.Pages
	var tabHeader *tview.Flex

	// Track current tab index (0=body, 1=auth, 2=query, 3=headers)
	currentTabIndex := 0

	// Tab indices
	enviromentIndex := 0
	collectionsIndex := 1
	urlBarIndex := 2
	requestIndex := 3
	responseIndex := 4

	// Additional UI variables
	var requestDataTabs *tview.Flex
	var bodyContainer *tview.Flex

	// Create the tabbed interface for request data (Body, Auth, Query, Headers)
	requestDataTabs, tabPages, bodyContainer, tabHeader, _, _, _, _ =
		createRequestDataTabs(bodyViewPanel,
			bodyEditPanel,
			backgroundColor,
			borderColor,
			borderFocusColor,
			titleColor,
			foregroundColor,
			activeTabColor,
			selectionBackgroundColor,
			buttonSelectedColor,
			func() { saveCurrentRequest(currentRequest, *collectionsData) },
			func(p tview.Primitive) { app.SetFocus(p) },
			func(tabIndex int) { currentTabIndex = tabIndex },
			nil, // panelFocusSetter will be set later
		)

	// Create unified Request panel containing method+URL+send and tabs
	requestPanel := setupRequestPanel(methodURLBar, requestDataTabs, backgroundColor)
	rightSide := setupRightSide(requestPanel, response)

	requestCycle = &RequestCycle{
		elements: []tview.Primitive{methodDropdown, urlInput, sendButton},
		current:  0,
		parent:   nil, // Will be set later
	}

	// Create left side layout
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	leftSide.AddItem(environmentPanel, 3, 1, false)
	leftSide.AddItem(collectionsTreeView, 0, 1, false)

	// Create main panels
	mainPanels := []tview.Primitive{environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, response}

	mainCycle = &MainCycle{
		panels:  mainPanels,
		current: 0,
	}

	headersCycle = &HeadersCycle{
		inputs:   []tview.Primitive{},
		current:  0,
		parent:   mainCycle,
		children: nil,
	}

	requestCycle.parent = mainCycle

	// Create main grid layout
	grid := tview.NewGrid().
		SetRows(1, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(leftSide, 1, 0, 1, 1, 0, 0, false)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)

	// Initial focus is on requestPanel (panels[1])
	currentFocus := collectionsIndex

	// Custom focus handler that knows about special containers
	setPanelFocus := func(panelIndex int, focused bool) {
		setFocusStyle(mainPanels[panelIndex], focused, borderColor, borderFocusColor)
	}

	setActiveBorder := func(element tview.Primitive) {
		setFocusStyle(element, true, borderColor, borderFocusColor)
	}

	setInactiveBorder := func(element tview.Primitive) {
		setFocusStyle(element, false, borderColor, borderFocusColor)
	}

	// Helper function to sync body content between view and edit panels
	syncBodyContent := func(content string) {
		// This will be implemented in event handlers
	}

	// Helper function to switch between view and edit modes
	switchBodyMode := func() {
		// This will be implemented in event handlers
	}

	// Create pages for modals
	pages := tview.NewPages()
	pages.AddPage("main", grid, true, true)

	// Set up tree view expansion handling
	SetupTreeViewExpansionHandling(*collectionsData, rootNode)

	// Set initial focus
	setPanelFocus(currentFocus, true)

	uiOrchestrator := &UIOrchestrator{
		App:                            app,
		CollectionsData:                collectionsData,
		DataManager:                    dataManager,
		EnvironmentsData:               environmentsData,
		RootNode:                       rootNode,
		MethodURLBar:                   methodURLBar,
		MethodDropdown:                 methodDropdown,
		URLInput:                       urlInput,
		SendButton:                     sendButton,
		BodyViewPanel:                  bodyViewPanel,
		BodyEditPanel:                  bodyEditPanel,
		Response:                       response,
		Footer:                         footer,
		CollectionsTreeView:            collectionsTreeView,
		ResponsePages:                  responsePages,
		ResponseTabHeader:              responseTabHeader,
		ResponseInfoBar:                responseInfoBar,
		ResponsePreviewPanel:           responsePreviewPanel,
		ResponseHeadersPanel:           responseHeadersPanel,
		ResponseCookiesPanel:           responseCookiesPanel,
		ResponseTimelinePanel:          responseTimelinePanel,
		EnvironmentPanel:               environmentPanel,
		EnvDropdown:                    envDropdown,
		EnvConfigButton:                envConfigButton,
		Header:                         header,
		Pages:                          pages,
		Grid:                           grid,
		CurrentSelectedNode:            currentSelectedNode,
		CurrentRequest:                 currentRequest,
		ProgrammaticallyUpdatingMethod: programmaticallyUpdatingMethod,
		ProgrammaticallyUpdatingURL:    programmaticallyUpdatingURL,
		TabPages:                       tabPages,
		TabHeader:                      tabHeader,
		CurrentTabIndex:                currentTabIndex,
		EnviromentIndex:                enviromentIndex,
		CollectionsIndex:               collectionsIndex,
		URLBarIndex:                    urlBarIndex,
		RequestIndex:                   requestIndex,
		ResponseIndex:                  responseIndex,
		RequestDataTabs:                requestDataTabs,
		BodyContainer:                  bodyContainer,
		MainCycle:                      mainCycle,
		HeadersCycle:                   headersCycle,
		RequestCycle:                   requestCycle,
		LastSelectedRequestNode:        nil,
		BodyEditMode:                   false,
		CurrentBodyContent:             "",
		RequestPanel:                   requestPanel,
		RightSide:                      rightSide,
		LeftSide:                       leftSide,
		CurrentFocus:                   currentFocus,
		SetPanelFocus:                  setPanelFocus,
		SetActiveBorder:                setActiveBorder,
		SetInactiveBorder:              setInactiveBorder,
		SyncBodyContent:                syncBodyContent,
		SwitchBodyMode:                 switchBodyMode,
		KeyManager:                     NewKeyBindingManager(),
	}

	return uiOrchestrator, nil
}
