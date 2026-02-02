package cmd

import (
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/workspace"
)

// PanelIndices defines the indices for main UI panels
type PanelIndices struct {
	Workspace   int
	Environment int
	Collections int
	URLBar      int
	Request     int
	Response    int
}

type Workspace struct {
	WorkspaceSelector int
	WorkspaceMenu     int
}

type Environment struct {
	EnvironmentSelector int
	EnvironmentMenu     int
}

type Collections struct {
	TreeView int
}

type URLBar struct {
	MethodDropdown int
	URLInput       int
	SendButton     int
	CurlButton     int
	AnotherItem    int
}

type BodyTab struct {
	ContentTypeSelector int
	JSONEditor          int
	MultipartFields     int
	NoBody              int
}

type Request struct {
	BodyTab    BodyTab
	AuthTab    int
	QueryTab   int
	HeadersTab int
}

type Response struct {
	PreviewTab  int
	HeadersTab  int
	CookiesTab  int
	TimelineTab int
}

type NavIndices struct {
	Workspace   Workspace
	Environment Environment
	Collections Collections
	URLBar      URLBar
	Request     Request
	Response    Response
}

// UIOrchestrator holds all UI components and state
type UIOrchestrator struct {
	App                      *tview.Application
	WorkspaceData            *workspace.Workspace
	DataManager              *DataManager
	EnvironmentsData         *[]workspace.Environment
	Colors                   *ColorManager
	PluginManager            *plugins.PluginManager
	RootNode                 *tview.TreeNode
	MethodURLBar             *tview.Flex
	MethodDropdown           *tview.DropDown
	ContentTypeDropdown      *tview.DropDown
	URLInput                 *URLVariableInput
	SendButton               *CustomButton
	CurlButton               *CustomButton
	BodyViewPanel            *tview.TextView
	BodyEditPanel            *tview.TextArea
	BodyContainer            *tview.Flex
	MultipartFieldsTab       *tview.Flex
	MultipartAddButton       *CustomButton
	MultipartDeleteAllButton *CustomButton
	AddHeaderButton          *CustomButton
	DeleteAllHeadersButton   *CustomButton
	Response                 *tview.Flex
	Footer                   *tview.Flex
	FooterLeft               *tview.TextView
	FooterRight              *tview.TextView
	CollectionsTreeView      *tview.TreeView
	TreeSelectionHandler     func(*tview.TreeNode)
	TreeHighlightHandler     func(*tview.TreeNode)
	ResponsePages            *tview.Pages
	ResponseTabHeader        *tview.Flex
	ResponseInfoBar          *tview.Flex
	ResponseTimeText         *tview.TextView
	ResponsePreviewPanel     *tview.TextView
	ResponseHeadersPanel     tview.Primitive
	ResponseCookiesPanel     *tview.TextView
	ResponseTimelinePanel    *tview.TextView
	EnvironmentPanel         *tview.Flex
	EnvDropdown              *tview.DropDown
	EnvConfigButton          *CustomButton
	WorkspacePanel           *tview.Flex
	WorkspaceSelector        *tview.DropDown
	WorkspaceConfigButton    *CustomButton
	Pages                    *tview.Pages
	Grid                     *tview.Grid
	KeyManager               *KeyBindingManager
	LastResponse             *HTTPResponse
	LastResponseTime         *time.Time

	// State variables
	CurrentSelectedNode                 *tview.TreeNode
	CurrentRequest                      *workspace.Request
	Navigating                          bool
	ProgrammaticallyUpdatingMethod      bool
	ProgrammaticallyUpdatingURL         bool
	ProgrammaticallyUpdatingContentType bool
	ProgrammaticallyUpdatingEnv         bool
	ProgrammaticallyUpdatingWorkspace   bool
	RequestInProgress                   bool
	TabPages                            *tview.Pages
	TabHeader                           *tview.Flex
	CurrentTabIndex                     int
	dropdownAdded                       bool
	CurrentResponseTabIndex             int
	PanelIndices                        PanelIndices
	NavIndices                          NavIndices
	NavCurrentContainer                 int
	NavCurrentChild                     int
	NavCurrentSubchild                  int
	NavCurrentMultipartElement          int // For navigation within multipart fields (0: Add Field, 1: Delete All, 2+: field rows)
	NavCurrentFieldRowElement           int // For navigation within a field row (0: Name, 1: Type, 2: Value, 3: Browse, 4: X)
	NavCurrentHeaderRowElement          int // For navigation within headers (0: Add Header, 1: Delete All, 2+: header rows)
	NavCurrentHeaderElement             int // For navigation within a header row (0: Key input, 1: Value input, 2: Delete button)
	NavPreviousContainer                int
	NavRequestInTabHeaders              bool // True when in Request panel tab headers
	NavResponseInTabHeaders             bool // True when in Response panel tab headers
	RequestDataTabs                     *tview.Flex
	MainCycle                           *MainCycle
	HeadersCycle                        *HeadersCycle
	URLBarCycle                         *URLBarCycle
	EnvironmentsCycle                   *EnvironmentsCycle
	WorkspaceCycle                      *WorkspaceCycle
	LastSelectedRequestNode             *tview.TreeNode
	BodyEditMode                        bool
	CurrentBodyContent                  string
	LastJSONBodyContent                 string // Store last JSON body when switching to No Body
	JSONBodyContent                     string // Store JSON body when switching away from JSON
	MultipartBodyContent                string // Store multipart fields when switching away from Multipart
	RequestPanel                        *tview.Flex
	RightSide                           *tview.Flex
	LeftSide                            *tview.Flex
	CurrentFocus                        int
	SetPanelFocus                       func(int, bool)
	SetActiveBorder                     func(element tview.Primitive)
	SetInactiveBorder                   func(element tview.Primitive)
	SyncBodyContent                     func(content string)
	SwitchBodyMode                      func()
	SwitchWorkspace                     func(string)
	RefreshMultipartFieldsUI            func()
	UpdateFooter                        func()
	CopyResponse                        func()
	Suspend                             func(func()) bool
	WorkspaceSelectorIndex              int
	WorkspaceConfigButtonIndex          int
	EnvironmentSelectorIndex            int
	EnvironmentConfigButtonIndex        int
	URLBarSelectorIndex                 int
	URLBarInputIndex                    int
	URLBarSendButtonIndex               int
	URLBarCurlButtonIndex               int
	RPBodyTabIndex                      int
	RPAuthTabIndex                      int
	RPQueryTabIndex                     int
	RPHeadersTabIndex                   int
}

// switchBodyContent switches the body container content based on content type
func (ui *UIOrchestrator) switchBodyContent(newContentType, oldContentType string) {
	// Save current body content before switching
	if ui.CurrentRequest != nil {
		if oldContentType == newContentType {
			// Initial load or no change - initialize saved content
			ui.CurrentBodyContent = ui.CurrentRequest.Body
			if newContentType == "JSON" {
				ui.JSONBodyContent = ui.CurrentBodyContent
			} else if newContentType == "Multipart" {
				ui.MultipartBodyContent = ui.CurrentBodyContent
			}
		} else if oldContentType == "Multipart" && newContentType == "JSON" {
			// Switching FROM Multipart TO JSON - save current multipart and restore saved JSON body
			// First save current multipart fields
			multipartBody := collectMultipartFieldsFromUI()
			ui.CurrentRequest.Body = multipartBody
			ui.CurrentBodyContent = ui.CurrentRequest.Body
			ui.MultipartBodyContent = ui.CurrentBodyContent

			// Now restore saved JSON body only if we have some
			// Otherwise keep the current multipart body (might be empty)
			if ui.JSONBodyContent != "" {
				ui.CurrentBodyContent = ui.JSONBodyContent
				ui.CurrentRequest.Body = ui.JSONBodyContent
			}
		} else if oldContentType == "JSON" && newContentType == "Multipart" {
			// Switching FROM JSON TO Multipart - save current JSON and restore saved multipart fields
			// First save current JSON body content
			if ui.BodyEditMode {
				// In edit mode, get from edit panel
				ui.CurrentRequest.Body = ui.BodyEditPanel.GetText()
			} else {
				// In view mode, CurrentBodyContent should have the current JSON
				ui.CurrentRequest.Body = ui.CurrentBodyContent
			}
			ui.CurrentBodyContent = ui.CurrentRequest.Body
			ui.JSONBodyContent = ui.CurrentBodyContent

			// Now restore saved multipart fields only if we have some
			// Otherwise keep the current JSON body
			if ui.MultipartBodyContent != "" {
				ui.CurrentBodyContent = ui.MultipartBodyContent
				ui.CurrentRequest.Body = ui.MultipartBodyContent
			}
		} else if oldContentType == "No Body" && newContentType == "JSON" {
			// Switching FROM No Body TO JSON - restore last JSON body content
			ui.CurrentBodyContent = ui.LastJSONBodyContent
			ui.CurrentRequest.Body = ui.LastJSONBodyContent
		} else if oldContentType == "Multipart" && newContentType != "Multipart" {
			// Switching FROM multipart - collect fields into body text
			multipartBody := collectMultipartFieldsFromUI()
			ui.CurrentRequest.Body = multipartBody
			ui.CurrentBodyContent = ui.CurrentRequest.Body
			// Save multipart fields for later restoration
			ui.MultipartBodyContent = ui.CurrentBodyContent
		} else if oldContentType == "JSON" && newContentType != "JSON" {
			// Switching FROM JSON - save current body content
			// Get the latest body content from the appropriate source
			if ui.BodyEditMode {
				// In edit mode, get from edit panel
				ui.CurrentRequest.Body = ui.BodyEditPanel.GetText()
			} else {
				// In view mode, CurrentBodyContent should have the current JSON
				// But to be safe, we'll use CurrentBodyContent which should be synced
				ui.CurrentRequest.Body = ui.CurrentBodyContent
			}
			// CurrentBodyContent should match CurrentRequest.Body
			ui.CurrentBodyContent = ui.CurrentRequest.Body
			// Save JSON body content when switching away from JSON
			ui.JSONBodyContent = ui.CurrentBodyContent
			if newContentType == "No Body" {
				ui.LastJSONBodyContent = ui.CurrentBodyContent
			}
		}
	}

	// Exit edit mode if switching to No Body or Multipart
	if newContentType == "No Body" || newContentType == "Multipart" {
		ui.BodyEditMode = false
	}

	ui.BodyContainer.Clear()

	switch newContentType {
	case "JSON":
		// Use the standard body view/edit panels
		ui.BodyContainer.SetTitle("")
		if ui.BodyEditMode {
			ui.BodyContainer.AddItem(ui.BodyEditPanel, 0, 1, false)
			ui.BodyEditPanel.SetText(ui.CurrentBodyContent, false)
		} else {
			ui.BodyContainer.AddItem(ui.BodyViewPanel, 0, 1, false)
			ui.SyncBodyContent(ui.CurrentBodyContent)
		}
	case "No Body":
		// Show empty state for no body
		ui.BodyContainer.SetTitle("")
		ui.BodyContainer.AddItem(ui.BodyViewPanel, 0, 1, false)
		ui.BodyViewPanel.SetText("(No body for this request)")
		ui.BodyViewPanel.SetTextAlign(tview.AlignCenter)
	case "Multipart":
		// Use the multipart fields UI
		ui.BodyContainer.SetTitle(" Multipart Fields ")
		ui.BodyContainer.AddItem(ui.MultipartFieldsTab, 0, 1, false)
	}
}

// SetupUI initializes all UI components and layout
func SetupUI(workspaceData *workspace.Workspace, dataManager *DataManager, environmentsData *[]workspace.Environment) (*UIOrchestrator, error) {
	// Initialize centralized color management with theme manager
	tm := GetThemeManager()
	colors := tm.GetColorManager()

	app := tview.NewApplication().EnableMouse(true)
	globalAppPtr = app

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
	curlButton := ui.CurlButton
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

	// Panel indices
	panelIndices := PanelIndices{
		Workspace:   0,
		Environment: 1,
		Collections: 2,
		URLBar:      3,
		Request:     4,
		Response:    5,
	}

	workspace := Workspace{
		WorkspaceSelector: 0,
		WorkspaceMenu:     1,
	}

	environment := Environment{
		EnvironmentSelector: 0,
		EnvironmentMenu:     1,
	}

	collections := Collections{
		TreeView: 0,
	}

	urlBar := URLBar{
		MethodDropdown: 0,
		URLInput:       1,
		SendButton:     2,
		CurlButton:     3,
		AnotherItem:    4,
	}

	bodyTab := BodyTab{
		ContentTypeSelector: 0,
		JSONEditor:          1,
		MultipartFields:     2,
		NoBody:              3,
	}

	request := Request{
		BodyTab:    bodyTab,
		AuthTab:    1,
		QueryTab:   2,
		HeadersTab: 3,
	}

	response := Response{
		PreviewTab:  0,
		HeadersTab:  1,
		CookiesTab:  2,
		TimelineTab: 3,
	}

	experimental := NavIndices{
		Workspace:   workspace,
		Environment: environment,
		Collections: collections,
		URLBar:      urlBar,
		Request:     request,
		Response:    response,
	}

	workspaceSelectorIndex := 0
	workspaceConfigButtonIndex := 1

	environmentSelectorIndex := 0
	environmentConfigButtonIndex := 1

	urlBarSelectorIndex := 0
	urlBarInputIndex := 1
	urlBarSendButtonIndex := 2
	urlBarCurlButtonIndex := 3

	RPBodyTabIndex := 0
	RPAuthTabIndex := 1
	RPQueryTabIndex := 2
	RPHeadersTabIndex := 3

	// Additional UI variables
	var requestDataTabs *tview.Flex
	var bodyContainer *tview.Flex

	urlBarCycle = &URLBarCycle{
		elements: []tview.Primitive{methodDropdown, urlInput, sendButton, curlButton},
		current:  0,
		parent:   nil, // Will be set later
	}

	// Create left side layout
	leftSide := tview.NewFlex().SetDirection(tview.FlexRow)
	leftSide.AddItem(workspacePanel, 3, 1, false)
	leftSide.AddItem(environmentPanel, 3, 1, false)
	leftSide.AddItem(collectionsTreeView, 0, 1, false)

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

	urlBarCycle.parent = mainCycle

	// Create main grid layout
	grid := tview.NewGrid().
		SetColumns(30, 0).
		SetBorders(false)

	grid.SetRows(0, 3)
	grid.AddItem(leftSide, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(footer, 1, 0, 1, 2, 0, 0, false)

	// Initial focus is on requestPanel (panels[1])
	currentFocus := panelIndices.Collections

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

	// Create the tabbed interface for request data (Body, Auth, Query, Headers)
	// We'll define tabIndexSetter and saveCallbackProxy after uiOrchestrator is created
	var tabIndexSetter func(int)
	var saveCallbackProxy func()

	requestDataTabs, tabPages, bodyContainer, tabHeader, _, _, _, _, contentTypeDropdown, multipartFieldsTab, refreshMultipartFieldsUI :=
		createRequestDataTabs(bodyViewPanel,
			bodyEditPanel,
			colors,
			func() {
				if saveCallbackProxy != nil {
					saveCallbackProxy()
				}
			},
			func(p tview.Primitive) { app.SetFocus(p) },
			func(tabIndex int) {
				// Call tabIndexSetter if it's been defined
				if tabIndexSetter != nil {
					tabIndexSetter(tabIndex)
				}
			},
			nil,       // panelFocusSetter will be set later
			func() {}, // footerUpdater - will be replaced later
			app,
			pages,
			currentRequest,
		)

	// Track if content type dropdown is added (now in uiOrchestrator.dropdownAdded)

	// Create main panels
	mainPanels := []tview.Primitive{workspacePanel, environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, responsePanel}

	mainCycle = &MainCycle{
		panels:  mainPanels,
		current: 2,
	}

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

	// Create unified Request panel containing method+URL+send and tabs
	requestPanel := setupRequestPanel(methodURLBar, requestDataTabs, colors)
	rightSide := setupRightSide(requestPanel, responsePanel)

	grid.AddItem(rightSide, 0, 1, 1, 1, 0, 0, false)

	uiOrchestrator := &UIOrchestrator{
		App:                            app,
		WorkspaceData:                  workspaceData,
		DataManager:                    dataManager,
		EnvironmentsData:               environmentsData,
		Colors:                         colors,
		RootNode:                       rootNode,
		MethodURLBar:                   methodURLBar,
		MethodDropdown:                 methodDropdown,
		ContentTypeDropdown:            contentTypeDropdown,
		URLInput:                       urlInput,
		SendButton:                     sendButton,
		CurlButton:                     curlButton,
		BodyViewPanel:                  bodyViewPanel,
		BodyEditPanel:                  bodyEditPanel,
		BodyContainer:                  bodyContainer,
		MultipartFieldsTab:             multipartFieldsTab,
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
		PanelIndices:                   panelIndices,
		NavIndices:                     experimental,
		NavCurrentContainer:            2,
		NavCurrentChild:                0,
		NavCurrentSubchild:             0,
		NavCurrentMultipartElement:     0,
		NavCurrentFieldRowElement:      0,
		NavCurrentHeaderRowElement:     0,
		NavCurrentHeaderElement:        0,
		NavPreviousContainer:           2,
		NavRequestInTabHeaders:         true,  // Start in tab headers when in Request panel
		NavResponseInTabHeaders:        false, // Start in tab content when in Response panel
		RequestDataTabs:                requestDataTabs,
		MainCycle:                      mainCycle,
		HeadersCycle:                   headersCycle,
		URLBarCycle:                    urlBarCycle,
		EnvironmentsCycle:              environmentsCycle,
		WorkspaceCycle:                 workspaceCycle,
		LastSelectedRequestNode:        nil,
		TreeHighlightHandler:           nil,
		BodyEditMode:                   false,
		CurrentBodyContent:             "",
		LastJSONBodyContent:            "",
		JSONBodyContent:                "",
		MultipartBodyContent:           "",
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
		URLBarCurlButtonIndex:          urlBarCurlButtonIndex,
		RPBodyTabIndex:                 RPBodyTabIndex,
		RPAuthTabIndex:                 RPAuthTabIndex,
		RPQueryTabIndex:                RPQueryTabIndex,
		RPHeadersTabIndex:              RPHeadersTabIndex,
		RefreshMultipartFieldsUI:       refreshMultipartFieldsUI,
		AddHeaderButton:                currentAddHeaderButton,
		DeleteAllHeadersButton:         currentDeleteAllHeadersButton,
		MultipartAddButton:             currentMultipartAddButton,
		MultipartDeleteAllButton:       currentMultipartDeleteAllButton,
		Suspend:                        app.Suspend,
	}

	// Define tabIndexSetter now that we have all the variables
	tabIndexSetter = func(tabIndex int) {
		// Update content type dropdown visibility
		if tabIndex == 0 {
			if !uiOrchestrator.dropdownAdded {
				requestDataTabs.RemoveItem(tabPages)
				requestDataTabs.AddItem(contentTypeDropdown, 1, 0, false)
				requestDataTabs.AddItem(tabPages, 0, 1, false)
				uiOrchestrator.dropdownAdded = true
			}
		} else {
			if uiOrchestrator.dropdownAdded {
				requestDataTabs.RemoveItem(contentTypeDropdown)
				uiOrchestrator.dropdownAdded = false
			}
		}
		currentTabIndex = tabIndex
		uiOrchestrator.CurrentTabIndex = tabIndex
		updateTabHeader(requestTabDisplayNames, tabHeader, currentTabIndex, colors)
		uiOrchestrator.UpdateFooter()

		// Switch to the selected tab page
		if tabIndex >= 0 && tabIndex < len(requestTabInternalNames) {
			tabPages.SwitchToPage(requestTabInternalNames[tabIndex])
		}
	}

	// Define saveCallbackProxy to use the current request from uiOrchestrator
	saveCallbackProxy = func() {
		saveCurrentRequest(uiOrchestrator.CurrentRequest, workspaceData)
	}

	// Set up tree view expansion handling
	SetupTreeViewExpansionHandling(workspaceData, rootNode)

	// Set initial focus
	setPanelFocus(currentFocus, true)

	// Set initial border for navigation
	if uiOrchestrator.NavCurrentContainer < len(mainPanels) {
		uiOrchestrator.SetActiveBorder(mainPanels[uiOrchestrator.NavCurrentContainer])
		// Sync MainCycle.current with experimental container
		// Map experimental container (0-5) to MainCycle panel indices
		switch uiOrchestrator.NavCurrentContainer {
		case 0:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.Workspace
		case 1:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.Environment
		case 2:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.Collections
		case 3:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.URLBar
		case 4:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.Request
		case 5:
			uiOrchestrator.MainCycle.current = uiOrchestrator.PanelIndices.Response
		}
		// Also update CurrentFocus for compatibility with old code
		uiOrchestrator.CurrentFocus = uiOrchestrator.MainCycle.current
		// Initial position is [2,0,0] - Collections panel, CollectionsTreeView
		uiOrchestrator.App.SetFocus(uiOrchestrator.CollectionsTreeView)
	}

	// Set initial footer right text
	uiOrchestrator.FooterRight.
		SetText("Petitorium ").
		SetTextAlign(tview.AlignRight)

	// Function to update footer based on current focus
	updateFooterFunc := func() {
		currentPage, _ := uiOrchestrator.Pages.GetFrontPage()
		if currentPage == "envVariables" {
			uiOrchestrator.FooterLeft.SetText(" (j/k) Navigate | (Enter) Select | (N) New Environment | (c) Clone Environment | (r) Rename Environment | (d) Delete Environment | (Tab) Switch Panel | (Esc/q) Close") // Environment Config
			return
		}

		// Check if experimental navigation is enabled
		var expPrefix string = ""

		switch uiOrchestrator.MainCycle.current {
		case uiOrchestrator.PanelIndices.Workspace:
			uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Select Workspace | (Tab) Next Panel | (q) Quit") // Workspace
		case uiOrchestrator.PanelIndices.Environment:
			uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Select Environment | (Tab) Next Panel | (q) Quit") // Environment
		case uiOrchestrator.PanelIndices.Collections:
			uiOrchestrator.FooterLeft.SetText(expPrefix + "(N) New Collection | (n) New Request | (r) Rename | (m) Move | (d) Delete | (D) Duplicate Request | (Tab) Next Panel | (q) Quit") // Collections
		case uiOrchestrator.PanelIndices.URLBar:
			focused := uiOrchestrator.App.GetFocus()
			if focused == uiOrchestrator.MethodDropdown {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Select the method | (Tab) Next | (q) Quit")
			} else if focused == uiOrchestrator.URLInput.viewMode {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(i) Edit URL | (Tab) Next | (q) Quit")
			} else if focused == uiOrchestrator.URLInput.editMode {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Save | (Esc) Cancel")
			} else if focused == uiOrchestrator.SendButton {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Send the request | (Tab) Next | (q) Quit")
			} else if focused == uiOrchestrator.CurlButton {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Export to cURL | (Tab) Next | (q) Quit")
			} else {
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(Enter) Send Request | (i) Edit URL | (c) Export cURL | (Tab) Next Panel | (q) Quit")
			}
		case uiOrchestrator.PanelIndices.Request:
			switch uiOrchestrator.CurrentTabIndex {
			case uiOrchestrator.RPBodyTabIndex:
				if uiOrchestrator.BodyEditMode {
					uiOrchestrator.FooterLeft.SetText(expPrefix + "(Esc) Exit Edit | (F4) External Editor | (Tab) Next Panel | (q) Quit") // Request Body (Edit)
				} else {
					uiOrchestrator.FooterLeft.SetText(expPrefix + "(i) Edit | (F4) External Editor | (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Body
				}
			case uiOrchestrator.RPAuthTabIndex:
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Auth
			case uiOrchestrator.RPQueryTabIndex:
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Query
			case uiOrchestrator.RPHeadersTabIndex:
				// Check if any header is in edit mode
				headerInEditMode := false
				for _, row := range currentHeaderRows {
					if (row.KeyInput != nil && row.KeyInput.IsEditMode()) ||
						(row.ValueInput != nil && row.ValueInput.IsEditMode()) {
						headerInEditMode = true
						break
					}
				}

				if headerInEditMode {
					uiOrchestrator.FooterLeft.SetText(expPrefix + "(Esc) Exit Edit | (Tab) Next Panel | (q) Quit") //  Request Headers (Edit)
				} else {
					uiOrchestrator.FooterLeft.SetText(expPrefix + "(i) Edit Key/Value | (n) New Header | (d) Delete Header | (D) Delete All | (F4) Bulk Edit | (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Headers
				}
			default:
				uiOrchestrator.FooterLeft.SetText(expPrefix + "(1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request
			}
		case uiOrchestrator.PanelIndices.Response:
			uiOrchestrator.FooterLeft.SetText(expPrefix + "(1-4/←/→) Switch tabs | (j/k) Scroll up/down | (d/u) Half page scroll | (g/G) Scroll to top/bottom | (f) Open in fx | (Tab) Next Panel | (q) Quit") // Response
		default:
			uiOrchestrator.FooterLeft.SetText(expPrefix + "(Tab) Cycle Focus | (q) Quit")
		}
	}

	// CopyResponse copies the current response body to clipboard
	uiOrchestrator.CopyResponse = func() {
		if uiOrchestrator.LastResponse != nil {
			copyToClipboard(uiOrchestrator.LastResponse.Body)
		}
	}

	uiOrchestrator.UpdateFooter = updateFooterFunc

	// Set onModeChange for URLInput to update footer when switching between view/edit modes
	uiOrchestrator.URLInput.onModeChange = uiOrchestrator.UpdateFooter

	// Set initial footer content
	uiOrchestrator.UpdateFooter()

	// Start clock goroutine to update time every second
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		lastFocus := app.GetFocus()

		for {
			select {
			case <-ticker.C:
				app.QueueUpdateDraw(func() {
					// Check for focus change to update footer
					currentFocus := app.GetFocus()
					if currentFocus != lastFocus {
						lastFocus = currentFocus
						updateFooterFunc()
					}

					// Update time only once per second (approx)
					// We check if it's been about a second since last time update
					// But for simplicity, we can just update it every 200ms too, it's not expensive
					if uiOrchestrator.ResponseTimeText != nil {
						if uiOrchestrator.LastResponseTime != nil {
							text := humanize.Time(*uiOrchestrator.LastResponseTime)
							uiOrchestrator.ResponseTimeText.SetText(fmt.Sprintf(" %s", text))
						} else {
							uiOrchestrator.ResponseTimeText.SetText(" -")
						}
					}
				})
			}
		}
	}()

	// Initialize response tab header with preview tab active
	updateResponseTabHeader(uiOrchestrator.ResponseTabHeader, uiOrchestrator.CurrentResponseTabIndex, uiOrchestrator.Colors)

	// Initialize request tab header with body tab active
	updateTabHeader(requestTabDisplayNames, uiOrchestrator.TabHeader, uiOrchestrator.CurrentTabIndex, uiOrchestrator.Colors)

	// Set initial tab to Body (0) to show content type dropdown
	tabIndexSetter(0)

	return uiOrchestrator, nil
}
