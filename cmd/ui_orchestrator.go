package cmd

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

// UIOrchestrator holds all UI components and state
type UIOrchestrator struct {
	App                   *tview.Application
	WorkspaceData         *workspace.Workspace
	DataManager           *DataManager
	EnvironmentsData      *[]workspace.Environment
	Colors                *ColorManager
	RootNode              *tview.TreeNode
	MethodURLBar          *tview.Flex
	MethodDropdown        *tview.DropDown
	URLInput              *URLVariableInput
	SendButton            *tview.Button
	BodyViewPanel         *tview.TextView
	BodyEditPanel         *tview.TextArea
	Response              *tview.Flex
	Footer                *tview.TextView
	CollectionsTreeView   *tview.TreeView
	ResponsePages         *tview.Pages
	ResponseTabHeader     *tview.Flex
	ResponseInfoBar       *tview.Flex
	ResponseTimeText      *tview.TextView
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
	LastResponseTime      *time.Time

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
	EnvironmentsCycle              *EnvironmentsCycle
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
	UpdateFooter                   func()
}

// SetupUI initializes all UI components and layout
func SetupUI(workspaceData *workspace.Workspace, dataManager *DataManager, environmentsData *[]workspace.Environment) (*UIOrchestrator, error) {
	// Initialize centralized color management
	colors := NewColorManager()

	app := tview.NewApplication().EnableMouse(true)

	// Create all UI components
	ui := setupUIComponents(colors)

	header := ui.Header
	rootNode := ui.RootNode
	methodURLBar := ui.MethodURLBar
	methodDropdown := ui.MethodDropdown
	urlInput := ui.URLInput
	sendButton := ui.SendButton
	bodyViewPanel := ui.BodyViewPanel
	bodyEditPanel := ui.BodyEditPanel
	responsePanel := ui.Response
	footer := ui.Footer
	collectionsTreeView := ui.CollectionsTreeView
	responsePages := ui.ResponsePages
	responseTabHeader := ui.ResponseTabHeader
	responseInfoBar := ui.ResponseInfoBar
	responseTimeText := ui.ResponseTimeText
	responsePreviewPanel := ui.ResponsePreviewPanel
	responseHeadersPanel := ui.ResponseHeadersPanel
	responseCookiesPanel := ui.ResponseCookiesPanel
	responseTimelinePanel := ui.ResponseTimelinePanel
	environmentPanel := ui.EnvironmentPanel
	envDropdown := ui.EnvDropdown
	envConfigButton := ui.EnvConfigButton

	// Update environment dropdown with loaded environments
	updateEnvironmentDropdown(envDropdown, *environmentsData)

	// Set initial environment selection based on config
	if config.C.SelectedEnvironment == "Base" {
		envDropdown.SetCurrentOption(0)
	} else {
		// Find the environment by name
		for i, env := range *environmentsData {
			if env.Name == config.C.SelectedEnvironment {
				envDropdown.SetCurrentOption(i + 1) // +1 because 0 is "Base"
				break
			}
		}
	}

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
			colors,
			func() { saveCurrentRequest(currentRequest, workspaceData) },
			func(p tview.Primitive) { app.SetFocus(p) },
			func(tabIndex int) { currentTabIndex = tabIndex },
			nil, // panelFocusSetter will be set later
		)

	// Create unified Request panel containing method+URL+send and tabs
	requestPanel := setupRequestPanel(methodURLBar, requestDataTabs, colors)
	rightSide := setupRightSide(requestPanel, responsePanel)

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
	mainPanels := []tview.Primitive{environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, responsePanel}

	mainCycle = &MainCycle{
		panels:  mainPanels,
		current: 1,
	}

	headersCycle = &HeadersCycle{
		inputs:   []tview.Primitive{},
		current:  0,
		parent:   mainCycle,
		children: nil,
	}

	environmentsCycle = &EnvironmentsCycle{
		inputs: []tview.Primitive{
			envDropdown,
			envConfigButton,
		},
		current:  0,
		parent:   mainCycle,
		children: nil,
	}

	requestCycle.parent = mainCycle

	// Create main grid layout
	grid := tview.NewGrid().
		SetColumns(30, 0).
		SetBorders(false)

	if header != nil {
		grid.SetRows(3, 0, 3)
		grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
		grid.AddItem(leftSide, 1, 0, 1, 1, 0, 0, false)
		grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)
		grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	} else {
		grid.SetRows(0, 3)
		grid.AddItem(leftSide, 0, 0, 1, 1, 0, 0, false)
		grid.AddItem(rightSide, 0, 1, 1, 1, 0, 0, false)
		grid.AddItem(footer, 1, 0, 1, 2, 0, 0, false)
	}

	// Initial focus is on requestPanel (panels[1])
	currentFocus := collectionsIndex

	// Custom focus handler that knows about special containers
	setPanelFocus := func(panelIndex int, focused bool) {
		setFocusStyle(mainPanels[panelIndex], focused, colors.Border, colors.BorderFocus)
	}

	setActiveBorder := func(element tview.Primitive) {
		setFocusStyle(element, true, colors.Border, colors.BorderFocus)
	}

	setInactiveBorder := func(element tview.Primitive) {
		setFocusStyle(element, false, colors.Border, colors.BorderFocus)
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
	SetupTreeViewExpansionHandling(workspaceData, rootNode)

	// Set initial focus
	setPanelFocus(currentFocus, true)

	uiOrchestrator := &UIOrchestrator{
		App:                            app,
		WorkspaceData:                  workspaceData,
		DataManager:                    dataManager,
		EnvironmentsData:               environmentsData,
		Colors:                         colors,
		RootNode:                       rootNode,
		MethodURLBar:                   methodURLBar,
		MethodDropdown:                 methodDropdown,
		URLInput:                       urlInput,
		SendButton:                     sendButton,
		BodyViewPanel:                  bodyViewPanel,
		BodyEditPanel:                  bodyEditPanel,
		Response:                       responsePanel,
		Footer:                         footer,
		CollectionsTreeView:            collectionsTreeView,
		ResponsePages:                  responsePages,
		ResponseTabHeader:              responseTabHeader,
		ResponseInfoBar:                responseInfoBar,
		ResponseTimeText:               responseTimeText,
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
		EnvironmentsCycle:              environmentsCycle,
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
		UpdateFooter:                   func() {}, // Will be set below
		KeyManager:                     NewKeyBindingManager(),
		LastResponseTime:               nil,
	}

	// Function to update footer based on current focus
	updateFooterFunc := func() {
		currentPage, _ := uiOrchestrator.Pages.GetFrontPage()
		if currentPage == "envVariables" {
			uiOrchestrator.Footer.SetText(" Environment Config: (j/k) Navigate | (Enter) Select | (n) New Environment | (r/R) Rename Environment | (d) Delete Environment | (Tab) Switch Panel | (Esc/q) Close")
			return
		}
		switch uiOrchestrator.MainCycle.current {
		case uiOrchestrator.EnviromentIndex:
			uiOrchestrator.Footer.SetText(" Environment: (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.CollectionsIndex:
			uiOrchestrator.Footer.SetText(" Collections: (n) New Collection | (r) New Request | (R) Rename | (m) Move | (d) Delete | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.URLBarIndex:
			uiOrchestrator.Footer.SetText(" Request: (1-4) Switch Tabs | (i) Edit Body | (F4) External Editor | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.RequestIndex:
			uiOrchestrator.Footer.SetText(" Request: (1-4) Switch Tabs | (i) Edit Body | (F4) External Editor | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.ResponseIndex:
			uiOrchestrator.Footer.SetText(" Response: (Tab) Next Panel | (q) Quit")
		default:
			uiOrchestrator.Footer.SetText(" (Tab) Cycle Focus | (q) Quit")
		}
	}
	uiOrchestrator.UpdateFooter = updateFooterFunc

	// Set initial footer content
	uiOrchestrator.UpdateFooter()

	// Start clock goroutine to update time every second
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if uiOrchestrator.ResponseTimeText != nil {
					app.QueueUpdateDraw(func() {
						if uiOrchestrator.LastResponseTime != nil {
							text := humanize.Time(*uiOrchestrator.LastResponseTime)
							uiOrchestrator.ResponseTimeText.SetText(fmt.Sprintf(" %s", text))
						} else {
							uiOrchestrator.ResponseTimeText.SetText(" -")
						}
					})
				}
			}
		}
	}()

	return uiOrchestrator, nil
}
