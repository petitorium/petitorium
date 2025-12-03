package cmd

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/workspace"
)

// UIOrchestrator holds all UI components and state
type UIOrchestrator struct {
	App                   *tview.Application
	WorkspaceData         *workspace.Workspace
	DataManager           *DataManager
	EnvironmentsData      *[]workspace.Environment
	Colors                *ColorManager
	PluginManager         *plugins.PluginManager
	RootNode              *tview.TreeNode
	MethodURLBar          *tview.Flex
	MethodDropdown        *tview.DropDown
	URLInput              *URLVariableInput
	SendButton            *CustomButton
	BodyViewPanel         *tview.TextView
	BodyEditPanel         *tview.TextArea
	Response              *tview.Flex
	Footer                *tview.Flex
	FooterLeft            *tview.TextView
	FooterRight           *tview.TextView
	CollectionsTreeView   *tview.TreeView
	TreeSelectionHandler  func(*tview.TreeNode)
	TreeHighlightHandler  func(*tview.TreeNode)
	ResponsePages         *tview.Pages
	ResponseTabHeader     *tview.Flex
	ResponseInfoBar       *tview.Flex
	ResponseTimeText      *tview.TextView
	ResponsePreviewPanel  *tview.TextView
	ResponseHeadersPanel  tview.Primitive
	ResponseCookiesPanel  *tview.TextView
	ResponseTimelinePanel *tview.TextView
	EnvironmentPanel      *tview.Flex
	EnvDropdown           *tview.DropDown
	EnvConfigButton       *CustomButton
	WorkspacePanel        *tview.Flex
	WorkspaceSelector     *tview.DropDown
	WorkspaceConfigButton *CustomButton
	Pages                 *tview.Pages
	Grid                  *tview.Grid
	KeyManager            *KeyBindingManager
	LastResponse          *HTTPResponse
	LastResponseTime      *time.Time

	// State variables
	CurrentSelectedNode            *tview.TreeNode
	CurrentRequest                 *workspace.Request
	Navigating                     bool
	ProgrammaticallyUpdatingMethod bool
	ProgrammaticallyUpdatingURL    bool
	TabPages                       *tview.Pages
	TabHeader                      *tview.Flex
	CurrentTabIndex                int
	CurrentResponseTabIndex        int
	WorkspaceIndex                 int
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
	WorkspaceCycle                 *WorkspaceCycle
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
	CopyResponse                   func()
	WorkspaceSelectorIndex         int
	WorkspaceConfigButtonIndex     int
	EnvironmentSelectorIndex       int
	EnvironmentConfigButtonIndex   int
	URLBarSelectorIndex            int
	URLBarInputIndex               int
	URLBarSendButtonIndex          int
	RPBodyTabIndex                 int
	RPAuthTabIndex                 int
	RPQueryTabIndex                int
	RPHeadersTabIndex              int
}

// SetupUI initializes all UI components and layout
func SetupUI(workspaceData *workspace.Workspace, dataManager *DataManager, environmentsData *[]workspace.Environment) (*UIOrchestrator, error) {
	// Initialize centralized color management with theme manager
	tm := GetThemeManager()
	colors := tm.GetColorManager()

	app := tview.NewApplication().EnableMouse(true)

	// Create all UI components
	ui := setupUIComponents(colors, app)

	workspacePanel := ui.WorkspacePanel
	workspaceSelector := ui.WorkspaceSelector
	workspaceConfigButton := ui.WorkspaceConfigButton
	rootNode := ui.RootNode
	methodURLBar := ui.MethodURLBar
	methodDropdown := ui.MethodDropdown
	urlInput := ui.URLInput
	sendButton := ui.SendButton
	bodyViewPanel := ui.BodyViewPanel
	bodyEditPanel := ui.BodyEditPanel
	responsePanel := ui.Response
	footer := ui.Footer
	footerLeft := ui.FooterLeft
	footerRight := ui.FooterRight
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

	// Set initial environment selection based on workspace
	selectedEnv := workspaceData.SelectedEnvironment
	if selectedEnv == "" {
		// Default to first environment or "Base" if no environments exist
		if len(*environmentsData) > 0 {
			envDropdown.SetCurrentOption(0)
		} else {
			envDropdown.SetCurrentOption(0) // "Base Environment"
		}
	} else {
		// Find the environment by name
		found := false
		for i, env := range *environmentsData {
			if env.Name == selectedEnv {
				envDropdown.SetCurrentOption(i + 1) // +1 because 0 is "Base Environment"
				found = true
				break
			}
		}
		if !found {
			// Selected environment not found, default to first option
			envDropdown.SetCurrentOption(0)
		}
	}

	// Initialize workspace selector
	workspaceNames, err := workspace.ListWorkspaces()
	if err != nil {
		// If there's an error, just use default
		workspaceNames = []string{"Default"}
	}
	workspaceSelector.SetOptions(workspaceNames, nil)

	// Set current workspace
	manager, err := workspace.LoadWorkspaceManager()
	if err == nil && manager.CurrentWorkspace != "" {
		for i, name := range workspaceNames {
			if name == manager.CurrentWorkspace {
				workspaceSelector.SetCurrentOption(i)
				break
			}
		}
	} else {
		workspaceSelector.SetCurrentOption(0)
	}

	// Set initial workspace name in right footer
	// currentWorkspaceIndex, _ := workspaceSelector.GetCurrentOption()
	// if currentWorkspaceIndex >= 0 && currentWorkspaceIndex < len(workspaceNames) {
	// 	footerRight.SetText(fmt.Sprintf("the Workspace: %s", workspaceNames[currentWorkspaceIndex]))
	// }

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
	workspaceIndex := 0
	enviromentIndex := 1
	collectionsIndex := 2
	urlBarIndex := 3
	requestIndex := 4
	responseIndex := 5

	workspaceSelectorIndex := 0
	workspaceConfigButtonIndex := 1

	environmentSelectorIndex := 0
	environmentConfigButtonIndex := 1

	urlBarSelectorIndex := 0
	urlBarInputIndex := 1
	urlBarSendButtonIndex := 2

	RPBodyTabIndex := 0
	RPAuthTabIndex := 1
	RPQueryTabIndex := 2
	RPHeadersTabIndex := 3

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
			func(tabIndex int) {
				currentTabIndex = tabIndex
				updateTabHeader([]string{"Body", "Auth", "Query", "Headers"}, tabHeader, currentTabIndex, colors)
			},
			nil, // panelFocusSetter will be set later
		)

	// Initialize tab header visual state
	updateTabHeader([]string{"Body", "Auth", "Query", "Headers"}, tabHeader, currentTabIndex, colors)

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
	leftSide.AddItem(workspacePanel, 3, 1, false)
	leftSide.AddItem(environmentPanel, 3, 1, false)
	leftSide.AddItem(collectionsTreeView, 0, 1, false)

	// Create main panels
	mainPanels := []tview.Primitive{workspacePanel, environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, responsePanel}

	mainCycle = &MainCycle{
		panels:  mainPanels,
		current: 2,
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

	workspaceCycle = &WorkspaceCycle{
		inputs: []tview.Primitive{
			workspaceSelector,
			workspaceConfigButton,
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

	grid.SetRows(0, 3)
	grid.AddItem(leftSide, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(rightSide, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(footer, 1, 0, 1, 2, 0, 0, false)

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
		FooterLeft:                     footerLeft,
		FooterRight:                    footerRight,
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
		WorkspacePanel:                 workspacePanel,
		WorkspaceSelector:              workspaceSelector,
		WorkspaceConfigButton:          workspaceConfigButton,
		Pages:                          pages,
		Grid:                           grid,
		CurrentSelectedNode:            currentSelectedNode,
		CurrentRequest:                 currentRequest,
		Navigating:                     false,
		ProgrammaticallyUpdatingMethod: programmaticallyUpdatingMethod,
		ProgrammaticallyUpdatingURL:    programmaticallyUpdatingURL,
		TabPages:                       tabPages,
		TabHeader:                      tabHeader,
		CurrentTabIndex:                currentTabIndex,
		CurrentResponseTabIndex:        0, // Start with preview tab
		WorkspaceIndex:                 workspaceIndex,
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
		WorkspaceCycle:                 workspaceCycle,
		LastSelectedRequestNode:        nil,
		TreeHighlightHandler:           nil,
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
		LastResponse:                   nil,
		LastResponseTime:               nil,
		WorkspaceSelectorIndex:         workspaceSelectorIndex,
		WorkspaceConfigButtonIndex:     workspaceConfigButtonIndex,
		EnvironmentSelectorIndex:       environmentSelectorIndex,
		EnvironmentConfigButtonIndex:   environmentConfigButtonIndex,
		URLBarSelectorIndex:            urlBarSelectorIndex,
		URLBarInputIndex:               urlBarInputIndex,
		URLBarSendButtonIndex:          urlBarSendButtonIndex,
		RPBodyTabIndex:                 RPBodyTabIndex,
		RPAuthTabIndex:                 RPAuthTabIndex,
		RPQueryTabIndex:                RPQueryTabIndex,
		RPHeadersTabIndex:              RPHeadersTabIndex,
	}

	// Set initial footer right text
	uiOrchestrator.FooterRight.
		SetText("Petitorium ").
		SetTextAlign(tview.AlignRight)

	// Function to update footer based on current focus
	updateFooterFunc := func() {
		currentPage, _ := uiOrchestrator.Pages.GetFrontPage()
		if currentPage == "envVariables" {
			uiOrchestrator.FooterLeft.SetText(" Environment Config: (j/k) Navigate | (Enter) Select | (n) New Environment | (r/R) Rename Environment | (d) Delete Environment | (Tab) Switch Panel | (Esc/q) Close")
			return
		}
		switch uiOrchestrator.MainCycle.current {
		case uiOrchestrator.EnviromentIndex:
			uiOrchestrator.FooterLeft.SetText(" Environment: (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.CollectionsIndex:
			uiOrchestrator.FooterLeft.SetText(" Collections: (n) New Collection | (r) New Request | (R) Rename | (m) Move | (d) Delete | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.URLBarIndex:
			uiOrchestrator.FooterLeft.SetText(" Request: (Enter) Edit URL | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.RequestIndex:
			uiOrchestrator.FooterLeft.SetText(" Request: (1-4/←/→) Switch Tabs | (i) Edit Body | (F4) External Editor | (Tab) Next Panel | (q) Quit")
		case uiOrchestrator.ResponseIndex:
			uiOrchestrator.FooterLeft.SetText(" Response: (1-4/←/→) Switch tabs | (j/k) Scroll up/down | (g/G) Scroll to top/bottom | (Tab) Next Panel | (q) Quit")
		default:
			uiOrchestrator.FooterLeft.SetText(" (Tab) Cycle Focus | (q) Quit")
		}
	}

	// CopyResponse copies the current response body to clipboard
	uiOrchestrator.CopyResponse = func() {
		if uiOrchestrator.LastResponse != nil {
			copyToClipboard(uiOrchestrator.LastResponse.Body)
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

	// Initialize response tab header with preview tab active
	updateResponseTabHeader(uiOrchestrator.ResponseTabHeader, uiOrchestrator.CurrentResponseTabIndex, uiOrchestrator.Colors)

	return uiOrchestrator, nil
}
