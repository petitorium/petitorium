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
	CurlButton            *CustomButton
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
	RequestInProgress              bool
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
	URLBarCurlButtonIndex          int
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
	urlBarCurlButtonIndex := 3

	RPBodyTabIndex := 0
	RPAuthTabIndex := 1
	RPQueryTabIndex := 2
	RPHeadersTabIndex := 3

	// Additional UI variables
	var requestDataTabs *tview.Flex
	var bodyContainer *tview.Flex

	requestCycle = &RequestCycle{
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

	requestCycle.parent = mainCycle

	// Create main grid layout
	grid := tview.NewGrid().
		SetColumns(30, 0).
		SetBorders(false)

	grid.SetRows(0, 3)
	grid.AddItem(leftSide, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(footer, 1, 0, 1, 2, 0, 0, false)

	// Initial focus is on requestPanel (panels[1])
	currentFocus := collectionsIndex

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
	tabIndexSetter := func(tabIndex int) {
		currentTabIndex = tabIndex
		updateTabHeader([]string{"Body", "Auth", "Query", "Headers"}, tabHeader, currentTabIndex, colors)
	}

	requestDataTabs, tabPages, bodyContainer, tabHeader, _, _, _, _ =
		createRequestDataTabs(bodyViewPanel,
			bodyEditPanel,
			colors,
			func() { saveCurrentRequest(currentRequest, workspaceData) },
			func(p tview.Primitive) { app.SetFocus(p) },
			tabIndexSetter,
			nil,       // panelFocusSetter will be set later
			func() {}, // footerUpdater - will be replaced later
			app,
			pages,
		)

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
		URLInput:                       urlInput,
		SendButton:                     sendButton,
		CurlButton:                     curlButton,
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
		URLBarCurlButtonIndex:          urlBarCurlButtonIndex,
		RPBodyTabIndex:                 RPBodyTabIndex,
		RPAuthTabIndex:                 RPAuthTabIndex,
		RPQueryTabIndex:                RPQueryTabIndex,
		RPHeadersTabIndex:              RPHeadersTabIndex,
	}

	// Update tabIndexSetter to also update UIOrchestrator's CurrentTabIndex
	tabIndexSetter = func(tabIndex int) {
		currentTabIndex = tabIndex
		uiOrchestrator.CurrentTabIndex = tabIndex
		updateTabHeader([]string{"Body", "Auth", "Query", "Headers"}, tabHeader, currentTabIndex, colors)
		uiOrchestrator.UpdateFooter()
	}

	// Set up tree view expansion handling
	SetupTreeViewExpansionHandling(workspaceData, rootNode)

	// Set initial focus
	setPanelFocus(currentFocus, true)

	// Set initial footer right text
	uiOrchestrator.FooterRight.
		SetText("Petitorium ").
		SetTextAlign(tview.AlignRight)

	// Function to update footer based on current focus
	updateFooterFunc := func() {
		currentPage, _ := uiOrchestrator.Pages.GetFrontPage()
		if currentPage == "envVariables" {
			uiOrchestrator.FooterLeft.SetText(" (j/k) Navigate | (Enter) Select | (n) New Environment | (c) Clone Environment | (r/R) Rename Environment | (d) Delete Environment | (Tab) Switch Panel | (Esc/q) Close") // Environment Config
			return
		}
		switch uiOrchestrator.MainCycle.current {
		case uiOrchestrator.EnviromentIndex:
			uiOrchestrator.FooterLeft.SetText(" (Tab) Next Panel | (q) Quit") // Environment
		case uiOrchestrator.CollectionsIndex:
			uiOrchestrator.FooterLeft.SetText(" (n) New Collection | (r) New Request | (R) Rename | (m) Move | (d) Delete | (D) Duplicate Request | (Tab) Next Panel | (q) Quit") // Collections
		case uiOrchestrator.URLBarIndex:
			uiOrchestrator.FooterLeft.SetText(" (i) Edit URL | (Tab) Next Panel | (c) Export cURL | (q) Quit") // Request
		case uiOrchestrator.RequestIndex:
			switch uiOrchestrator.CurrentTabIndex {
			case uiOrchestrator.RPBodyTabIndex:
				if uiOrchestrator.BodyEditMode {
					uiOrchestrator.FooterLeft.SetText(" (Esc) Exit Edit | (F4) External Editor | (Tab) Next Panel | (q) Quit") // Request Body (Edit)
				} else {
					uiOrchestrator.FooterLeft.SetText(" (i) Edit | (F4) External Editor | (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Body
				}
			case uiOrchestrator.RPAuthTabIndex:
				uiOrchestrator.FooterLeft.SetText(" (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Auth
			case uiOrchestrator.RPQueryTabIndex:
				uiOrchestrator.FooterLeft.SetText(" (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Query
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
					uiOrchestrator.FooterLeft.SetText(" (Esc) Exit Edit | (Tab) Next Panel | (q) Quit") //  Request Headers (Edit)
				} else {
					uiOrchestrator.FooterLeft.SetText(" (i) Edit Key/Value | (n) New Header | (d) Delete Header | (D) Delete All | (F4) Bulk Edit | (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request Headers
				}
			default:
				uiOrchestrator.FooterLeft.SetText(" (1-4/←/→) Switch Tabs | (Tab) Next Panel | (q) Quit") // Request
			}
		case uiOrchestrator.ResponseIndex:
			uiOrchestrator.FooterLeft.SetText(" (1-4/←/→) Switch tabs | (j/k) Scroll up/down | (d/u) Half page scroll | (g/G) Scroll to top/bottom | (Tab) Next Panel | (q) Quit") // Response
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
				app.QueueUpdateDraw(func() {
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

	return uiOrchestrator, nil
}
