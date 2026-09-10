package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/workspace"
)

func isBinaryContentType(contentType string) bool {
	if contentType == "" {
		return false
	}
	binaryPrefixes := []string{
		"image/",
		"audio/",
		"video/",
		"application/octet-stream",
		"application/pdf",
		"application/zip",
		"application/x-rar-compressed",
		"application/x-tar",
		"application/gzip",
	}
	for _, prefix := range binaryPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

func entriesToStringMap(entries map[string]workspace.Entry) map[string]string {
	result := make(map[string]string)
	for key, entry := range entries {
		result[key] = entry.Value
	}
	return result
}

func stringMapToEntries(stringMap map[string]string) map[string]workspace.Entry {
	result := make(map[string]workspace.Entry)
	for key, value := range stringMap {
		result[key] = workspace.Entry{Value: value, Enabled: true}
	}
	return result
}

func refreshCollectionsTree(ui *UIOrchestrator) {
	ui.RootNode.ClearChildren()

	addWorkspaceToTree(ui.WorkspaceData, ui.RootNode)

	ui.CollectionsTreeView.SetRoot(ui.RootNode)

	// Try to restore selection from workspace data
	var nodeToSelect *tview.TreeNode
	if len(ui.WorkspaceData.SelectedRequest) > 0 {
		nodeToSelect = findNodeByPath(ui.RootNode, ui.WorkspaceData.SelectedRequest)
	}

	if nodeToSelect != nil {
		ui.CollectionsTreeView.SetCurrentNode(nodeToSelect)
	} else {
		ui.CollectionsTreeView.SetCurrentNode(ui.RootNode)
		// If there are children, select the first one instead of the root
		if len(ui.RootNode.GetChildren()) > 0 {
			ui.CollectionsTreeView.SetCurrentNode(ui.RootNode.GetChildren()[0])
		}
	}
}

// isFocusOnHeaderInputField checks if focus is on any header input field
// including HeaderValueInput components in edit mode
func isFocusOnHeaderInputField(currentFocusedElement tview.Primitive, headerRows []*HeaderRow) bool {
	for _, row := range headerRows {
		if currentFocusedElement == row.KeyInput {
			return true
		}
		// Check if focused on HeaderValueInput or its children when in edit mode
		if row.ValueInput != nil && row.ValueInput.HasFocus() && row.ValueInput.IsEditMode() {
			return true
		}
		// Also check if focused directly on HeaderValueInput component
		if hvi, ok := currentFocusedElement.(*HeaderValueInput); ok && currentFocusedElement == row.ValueInput {
			if hvi.IsEditMode() {
				return true
			}
		}
	}
	return false
}

// isFocusOnQueryParamInputField checks if focus is on any query param input field
func isFocusOnQueryParamInputField(currentFocusedElement tview.Primitive, queryParamRows []*QueryParamRow) bool {
	for _, row := range queryParamRows {
		if currentFocusedElement == row.KeyInput {
			return true
		}
		if row.ValueInput != nil && row.ValueInput.HasFocus() && row.ValueInput.IsEditMode() {
			return true
		}
		if hvi, ok := currentFocusedElement.(*HeaderValueInput); ok && currentFocusedElement == row.ValueInput {
			if hvi.IsEditMode() {
				return true
			}
		}
	}
	return false
}

// handleTabSwitch handles tab switching for both number keys (1-4) and arrow keys
// Returns true if the event was handled (tab was switched), false otherwise
func handleTabSwitch(ui *UIOrchestrator, event *tcell.EventKey, headerRows []*HeaderRow) bool {
	// Check if we should handle tab switching
	// For number keys: check CurrentFocus (which tracks keyboard focus)
	// For arrow keys: check MainCycle.current and also check !BodyEditMode for request tabs
	isNumberKey := event.Rune() == '1' || event.Rune() == '2' || event.Rune() == '3' || event.Rune() == '4'
	isArrowKey := event.Key() == tcell.KeyLeft || event.Key() == tcell.KeyRight

	// Check if we're in request panel
	isRequestPanel := ui.MainCycle.current == ui.PanelIndices.Request
	// Check if we're in response panel
	isResponsePanel := ui.MainCycle.current == ui.PanelIndices.Response

	if isNumberKey {
		// Number keys work when request or response panel has focus
		if ui.CurrentFocus != ui.PanelIndices.Request && ui.CurrentFocus != ui.PanelIndices.Response {
			return false
		}
	} else if isArrowKey {
		// Arrow keys work when request panel is active AND we're not editing body
		// OR when response panel is active
		if !((isRequestPanel && !ui.BodyEditMode) || isResponsePanel) {
			return false
		}
	} else {
		// Not a tab switching key
		return false
	}

	// Check if focus is on an input field (don't switch tabs if typing)
	// Both number keys and arrow keys should respect this for header and query param input fields
	currentFocusedElement := ui.App.GetFocus()
	if isFocusOnHeaderInputField(currentFocusedElement, headerRows) {
		return false
	}
	if isFocusOnQueryParamInputField(currentFocusedElement, currentQueryRows) {
		return false
	}

	// Determine which tab to switch to based on key
	var targetTabIndex int = -1

	switch {
	case event.Rune() == '1':
		targetTabIndex = 0
	case event.Rune() == '2':
		targetTabIndex = 1
	case event.Rune() == '3':
		targetTabIndex = 2
	case event.Rune() == '4':
		targetTabIndex = 3
	case event.Key() == tcell.KeyLeft:
		// Wrap around from left
		if isRequestPanel {
			targetTabIndex = (ui.CurrentTabIndex - 1 + len(RequestTabInternalNames)) % len(RequestTabInternalNames)
		} else if isResponsePanel {
			targetTabIndex = (ui.CurrentResponseTabIndex - 1 + 4) % 4
		}
	case event.Key() == tcell.KeyRight:
		// Wrap around from right
		if isRequestPanel {
			targetTabIndex = (ui.CurrentTabIndex + 1) % len(RequestTabInternalNames)
		} else if isResponsePanel {
			targetTabIndex = (ui.CurrentResponseTabIndex + 1) % 4
		}
	default:
		return false
	}

	// Perform the tab switch
	if isRequestPanel {
		ui.TabPages.SwitchToPage(RequestTabInternalNames[targetTabIndex])
		requestTabs := RequestTabDisplayNames
		updateTabHeader(requestTabs, ui.TabHeader, targetTabIndex, ui.Colors)
		ui.CurrentTabIndex = targetTabIndex

		// Update navigation state
		if ui.NavCurrentContainer == 4 && ui.NavRequestInTabHeaders {
			ui.NavCurrentChild = targetTabIndex
			// Update focus and footer for navigation
			setFocusForCoordinates(ui)
		}

		// Update content type dropdown visibility
		if targetTabIndex == 0 {
			if !ui.dropdownAdded {
				ui.RequestDataTabs.RemoveItem(ui.TabPages)
				ui.RequestDataTabs.AddItem(ui.ContentTypeDropdown, 1, 0, false)
				ui.RequestDataTabs.AddItem(ui.TabPages, 0, 1, false)
				ui.dropdownAdded = true
			}
		} else {
			if ui.dropdownAdded {
				ui.RequestDataTabs.RemoveItem(ui.ContentTypeDropdown)
				ui.dropdownAdded = false
			}
		}

		// Focus the appropriate tab content
		// But not if we're in tab headers mode
		if !(ui.NavCurrentContainer == 4 && ui.NavRequestInTabHeaders) {
			switch targetTabIndex {
			case 0: // Body tab
				if ui.BodyEditMode {
					ui.App.SetFocus(ui.BodyEditPanel)
				} else {
					ui.App.SetFocus(ui.BodyViewPanel)
				}
			case 2: // Query tab
				if len(currentQueryRows) > 0 && currentQueryRows[0].KeyInput != nil {
					ui.App.SetFocus(currentQueryRows[0].KeyInput)
				} else {
					ui.App.SetFocus(ui.RequestDataTabs)
				}
			case 3: // Headers tab
				if len(headerRows) > 0 && headerRows[0].KeyInput != nil {
					ui.App.SetFocus(headerRows[0].KeyInput)
				} else {
					ui.App.SetFocus(ui.RequestDataTabs)
				}
			default:
				ui.App.SetFocus(ui.RequestDataTabs)
			}
		}
	} else if isResponsePanel {
		responseTabNames := []string{"preview", "headers", "cookies", "timeline"}
		ui.ResponsePages.SwitchToPage(responseTabNames[targetTabIndex])
		updateResponseTabHeader(ui.ResponseTabHeader, targetTabIndex, ui.Colors)
		ui.CurrentResponseTabIndex = targetTabIndex

		// Update navigation state
		if ui.NavCurrentContainer == 5 && ui.NavResponseInTabHeaders {
			ui.NavCurrentChild = targetTabIndex
			// Update focus and footer for navigation
			setFocusForCoordinates(ui)
		}
	}

	// Update footer after tab switch
	ui.UpdateFooter()

	return true
}

// savePluginEnvironmentChanges saves plugin-modified environment variables back
// to the current environment.  It skips keys whose original workspace value
// contains a {{...}} tag but whose resolved value does not, so that dynamic
// tags (command-runner and cross-references) are not overwritten with static
// text after each request.
func savePluginEnvironmentChanges(pluginEnv map[string]string, currentEnvIndex int, environmentsData *[]workspace.Environment) {
	if pluginEnv == nil || len(pluginEnv) == 0 || environmentsData == nil {
		return
	}

	// Find the current environment
	var currentEnv *workspace.Environment
	if currentEnvIndex == 0 {
		// Base Environment
		for i, env := range *environmentsData {
			if env.Name == "Base" {
				currentEnv = &(*environmentsData)[i]
				break
			}
		}
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(*environmentsData) {
		currentEnv = &(*environmentsData)[currentEnvIndex-1]
	}

	if currentEnv == nil {
		return
	}

	// Initialize Variables map if nil
	if currentEnv.Variables == nil {
		currentEnv.Variables = make(map[string]string)
	}

	// Only persist keys that were genuinely modified by plugins.
	// If the original value contained a {{...}} tag and the resolved value
	// does not, we assume it was mechanically resolved and skip it.
	for key, resolvedValue := range pluginEnv {
		originalValue, exists := currentEnv.Variables[key]
		if exists && strings.Contains(originalValue, "{{") && !strings.Contains(resolvedValue, "{{") {
			continue
		}
		currentEnv.Variables[key] = resolvedValue
	}
}

var collectionSearchDebounceTimer *time.Timer

var collectionSearchResults []RequestSearchResult

type CollectionSearchModal struct {
	*tview.Flex
	table       *tview.Table
	searchField *tview.InputField
	ui          *UIOrchestrator
	results     []RequestSearchResult
	returnFocus tview.Primitive
}

func NewCollectionSearchModal(ui *UIOrchestrator) *CollectionSearchModal {
	m := &CollectionSearchModal{
		Flex:        tview.NewFlex().SetDirection(tview.FlexRow),
		table:       tview.NewTable().SetSelectable(true, false).SetFixed(1, 0),
		ui:          ui,
		returnFocus: ui.CollectionsTreeView,
	}

	m.SetBackgroundColor(ui.Colors.Background)

	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.Foreground))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.searchField = tview.NewInputField().
		SetLabel(" / ").
		SetLabelColor(ui.Colors.BorderFocus).
		SetPlaceholder("Search requests...").
		SetPlaceholderTextColor(ui.Colors.Placeholder).
		SetFieldBackgroundColor(ui.Colors.InputBackground).
		SetFieldTextColor(ui.Colors.Foreground)
	m.searchField.SetBackgroundColor(ui.Colors.Background)

	m.searchField.SetChangedFunc(func(text string) {
		if collectionSearchDebounceTimer != nil {
			collectionSearchDebounceTimer.Stop()
		}
		collectionSearchDebounceTimer = time.AfterFunc(300*time.Millisecond, func() {
			ui.App.QueueUpdateDraw(func() {
				m.filterRequests(text)
			})
		})
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeRequestSearchModal(m)
			return nil
		}
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			if m.table.GetRowCount() > 1 {
				ui.App.SetFocus(m.table)
			}
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			if len(m.results) > 0 {
				m.table.Select(1, 0)
				m.selectResult()
			}
			return nil
		}
		return event
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := m.table.GetSelection()
		if event.Key() == tcell.KeyUp && row == 1 {
			ui.App.SetFocus(m.searchField)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			ui.App.SetFocus(m.searchField)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			m.selectResult()
			return nil
		}
		if event.Rune() == '/' {
			closeRequestSearchModal(m)
			return nil
		}
		return event
	})

	m.table.SetSelectedFunc(func(row, column int) {
		if row > 0 {
			m.selectResult()
		}
	})

	m.AddItem(m.searchField, 1, 0, true)
	m.AddItem(m.table, 0, 1, false)

	m.SetBorder(true).SetTitle(" Search Requests ")
	m.SetBorderColor(ui.Colors.BorderFocus)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeRequestSearchModal(m)
			return nil
		}
		if event.Rune() == '/' {
			closeRequestSearchModal(m)
			return nil
		}
		return event
	})

	return m
}

func (m *CollectionSearchModal) filterRequests(query string) {
	m.table.Clear()
	m.results = nil

	if query == "" {
		return
	}

	results := searchRequests(query, m.ui.WorkspaceData.Collections)
	if len(results) == 0 {
		return
	}

	headers := []string{"Collection", "Name", "Method", "URL"}
	for i, h := range headers {
		m.table.SetCell(0, i, tview.NewTableCell(" "+h+" ").
			SetTextColor(m.ui.Colors.Title).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignCenter))
	}
	m.table.GetCell(0, 0).SetAlign(tview.AlignLeft)

	m.results = results
	row := 1
	for _, result := range results {
		pathStr := strings.Join(result.CollectionPath, " > ")
		m.table.SetCell(row, 0, tview.NewTableCell(" "+pathStr+" ").SetExpansion(1).SetTextColor(m.ui.Colors.Foreground).SetAlign(tview.AlignLeft))
		m.table.SetCell(row, 1, tview.NewTableCell(" "+result.Request.Name+" ").SetExpansion(2).SetTextColor(m.ui.Colors.Foreground).SetAlign(tview.AlignLeft))
		m.table.SetCell(row, 2, tview.NewTableCell(" "+result.Request.Method+" ").SetExpansion(0).SetTextColor(m.ui.Colors.Foreground).SetAlign(tview.AlignCenter))
		urlDisplay := result.Request.URL
		if len(urlDisplay) > 40 {
			urlDisplay = urlDisplay[:37] + "..."
		}
		m.table.SetCell(row, 3, tview.NewTableCell(" "+urlDisplay+" ").SetExpansion(1).SetTextColor(m.ui.Colors.Placeholder).SetAlign(tview.AlignLeft))
		row++
	}
}

func (m *CollectionSearchModal) selectResult() {
	row, _ := m.table.GetSelection()
	if row < 1 || row-1 >= len(m.results) {
		return
	}

	result := m.results[row-1]
	collectionPath := result.CollectionPath
	request := result.Request

	node := findNodeByPath(m.ui.RootNode, collectionPath)
	if node == nil {
		closeRequestSearchModal(m)
		return
	}

	for _, pathPart := range collectionPath {
		ancestorNode := findNodeByPath(m.ui.RootNode, []string{pathPart})
		if ancestorNode != nil {
			m.ui.expandCollectionAndLoadChildren(ancestorNode)
		}
	}
	m.ui.expandCollectionAndLoadChildren(node)

	for _, child := range node.GetChildren() {
		if childReq := m.ui.requestFromNode(child); childReq != nil && childReq.ID == request.ID {
			child.Expand()
			m.ui.CollectionsTreeView.SetCurrentNode(child)
			if m.ui.TreeHighlightHandler != nil {
				m.ui.TreeHighlightHandler(child)
			}
			if m.ui.TreeSelectionHandler != nil {
				m.ui.TreeSelectionHandler(child)
			}
			break
		}
	}

	closeRequestSearchModal(m)
}

func closeRequestSearchModal(m *CollectionSearchModal) {
	m.ui.Pages.RemovePage("collectionSearch")

	// Restore focus
	if m.returnFocus != nil {
		m.ui.App.SetFocus(m.returnFocus)
	} else {
		m.ui.App.SetFocus(m.ui.CollectionsTreeView)
	}

	if m.ui.ExitModal != nil {
		m.ui.ExitModal()
	}
}

// SetupEventHandlers configures all event handlers for the UI
func SetupEventHandlers(ui *UIOrchestrator) {
	// Populate the collections tree initially
	refreshCollectionsTree(ui)

	manager, _ := workspace.LoadWorkspaceManager()
	lastWorkspace := ""
	if manager != nil {
		lastWorkspace = manager.CurrentWorkspace
	}

	// Store workspace names for lookup
	var err error
	ui.WorkspaceNames, err = workspace.ListWorkspaces()
	if err != nil {
		ui.WorkspaceNames = []string{"Default"}
	}

	ui.SwitchWorkspace = func(text string) {
		if text == lastWorkspace || text == "" {
			return
		}

		if ui.ProgrammaticallyUpdatingWorkspace {
			return
		}

		// 1. Save CURRENT workspace (Project A) before switching
		if ui.WorkspaceData != nil {
			// Ensure the current environment selection is captured
			currentIndex, _ := ui.EnvDropdown.GetCurrentOption()
			if currentIndex == 0 {
				ui.WorkspaceData.SelectedEnvironment = ""
			} else if currentIndex > 0 && currentIndex <= len(*ui.EnvironmentsData) {
				ui.WorkspaceData.SelectedEnvironment = (*ui.EnvironmentsData)[currentIndex-1].Name
			}
			workspace.SaveWorkspace(ui.WorkspaceData)
		}

		lastWorkspace = text

		// 2. Perform the switch in the manager
		if err := workspace.SwitchWorkspace(text); err != nil {
			return
		}

		// 3. Load the NEW workspace
		newWorkspace, err := workspace.LoadWorkspace()
		if err != nil {
			return
		}

		// 4. Update UI orchestrator state
		ui.WorkspaceData = newWorkspace
		ui.EnvironmentsData = &newWorkspace.Environments
		ui.DataManager = NewDataManager(newWorkspace)

		// 5. Update UI components
		ui.ProgrammaticallyUpdatingEnv = true
		updateEnvironmentDropdown(ui.EnvDropdown, *ui.EnvironmentsData)

		// Restore the last used environment
		selectedEnv := newWorkspace.SelectedEnvironment
		if selectedEnv == "" {
			ui.EnvDropdown.SetCurrentOption(0) // "Base Environment"
		} else {
			found := false
			for i, env := range *ui.EnvironmentsData {
				if env.Name == selectedEnv {
					ui.EnvDropdown.SetCurrentOption(i + 1) // +1 because 0 is "Base Environment"
					found = true
					break
				}
			}
			if !found {
				ui.EnvDropdown.SetCurrentOption(0)
			}
		}
		ui.ProgrammaticallyUpdatingEnv = false

		refreshCollectionsTree(ui)

		ui.CurrentRequest = nil
		ui.CurrentRequestID = ""
		ui.CurrentSelectedNode = nil
		ui.LastSelectedRequestNode = nil

		ui.ProgrammaticallyUpdatingURL = true
		ui.URLInput.SetText("")
		ui.ProgrammaticallyUpdatingURL = false

		ui.SyncBodyContent("")
		setHeadersInUI(ui.Colors, nil, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)
		setQueryParamsInUI(ui.Colors, nil, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)

		ui.LastResponse = nil
		updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)

		// Restore selection state for the new workspace
		currentNode := ui.CollectionsTreeView.GetCurrentNode()
		if currentNode != nil && ui.TreeSelectionHandler != nil {
			ui.TreeSelectionHandler(currentNode)
		}

		// Update workspace selector dropdown to show the current workspace
		ui.ProgrammaticallyUpdatingWorkspace = true
		ui.WorkspaceNames, _ = workspace.ListWorkspaces()
		for i, name := range ui.WorkspaceNames {
			if name == text {
				ui.WorkspaceSelector.SetCurrentOption(i)
				break
			}
		}
		ui.ProgrammaticallyUpdatingWorkspace = false
	}

	ui.WorkspaceSelector.SetSelectedFunc(func(text string, index int) {
		if text != "" {
			ui.SwitchWorkspace(text)
		}
	})

	// Set up vim-style navigation for body view panel (TextView)
	ui.BodyViewPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return ui.KeyManager.HandleKeyEvent(ui, event, "body_view")
	})

	// Set up TreeView-specific input capture for h/l navigation
	ui.CollectionsTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return ui.KeyManager.HandleKeyEvent(ui, event, "tree_view")
	})

	// Set up vim-style navigation for response panels
	responseViewCapture := func(event *tcell.EventKey) *tcell.EventKey {
		return ui.KeyManager.HandleKeyEvent(ui, event, "response_view")
	}
	ui.ResponsePages.SetInputCapture(responseViewCapture)
	ui.ResponsePreviewPanel.SetInputCapture(responseViewCapture)
	if ui.ResponseHeadersPanel != nil {
		if table, ok := ui.ResponseHeadersPanel.(*tview.Table); ok {
			table.SetInputCapture(responseViewCapture)
		}
	}
	ui.ResponseCookiesPanel.SetInputCapture(responseViewCapture)
	ui.ResponseTimelinePanel.SetInputCapture(responseViewCapture)

	// Initialize syncBodyContent function
	ui.SyncBodyContent = func(content string) {
		ui.CurrentBodyContent = content
		if ui.BodyEditMode {
			ui.BodyEditPanel.SetText(content, false)
		} else {
			ui.BodyViewPanel.Clear()
			if content != "" {
				// Format content with syntax highlighting and variable highlighting
				formattedContent := FormatBodyContentWithVariables(content)

				// Set content with proper handling
				ui.BodyViewPanel.SetText(formattedContent)
				ui.BodyViewPanel.SetTextAlign(tview.AlignLeft)

				// Force proper rendering and scrolling
				go func() {
					// Small delay to ensure content is set before scrolling
					ui.BodyViewPanel.ScrollToEnd()
					ui.BodyViewPanel.ScrollToBeginning()
				}()
			} else {
				ui.BodyViewPanel.SetText("")
				ui.BodyViewPanel.SetTextAlign(tview.AlignLeft)
			}
		}
	}

	// Initialize switchBodyMode function with proper focus handling
	ui.SwitchBodyMode = func() {
		// Don't allow switching to edit mode if content type is No Body or Multipart
		if ui.CurrentRequest != nil && (ui.CurrentRequest.ContentType == "No Body" || ui.CurrentRequest.ContentType == "Multipart") {
			return
		}

		ui.BodyEditMode = !ui.BodyEditMode
		ui.BodyContainer.Clear()

		if ui.BodyEditMode {
			// Switch to edit mode
			ui.BodyContainer.AddItem(ui.BodyEditPanel, 0, 1, false)
			ui.BodyEditPanel.SetText(ui.CurrentBodyContent, false)
			ui.BodyEditPanel.SetBorderColor(ui.Colors.Background)
		} else {
			// Switch to view mode
			ui.BodyContainer.AddItem(ui.BodyViewPanel, 0, 1, false)
			// Update body content from edit panel if we were editing
			if ui.CurrentRequest != nil {
				ui.CurrentBodyContent = ui.BodyEditPanel.GetText()
				ui.CurrentRequest.Body = ui.CurrentBodyContent
				// Also save to JSONBodyContent if we're in JSON mode
				if ui.CurrentRequest.ContentType == "JSON" {
					ui.JSONBodyContent = ui.CurrentBodyContent
				}
				if ui.CurrentSelectedNode != nil {
					saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
				}
			}
			ui.SyncBodyContent(ui.CurrentBodyContent)
		}
		ui.UpdateFooter()
	}

	// Set up input capture for bodyEditPanel
	ui.BodyEditPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return ui.KeyManager.HandleKeyEvent(ui, event, "body_edit")
	})

	ui.MethodDropdown.SetSelectedFunc(func(text string, index int) {
		if ui.ProgrammaticallyUpdatingMethod {
			return
		}

		if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
			if index >= 0 && index < len(workspace.HTTPMethods) {
				newMethod := workspace.HTTPMethods[index]
				if newMethod != ui.CurrentRequest.Method {
					ui.CurrentRequest.Method = newMethod

					coloredMethod := getColoredMethod(ui.CurrentRequest.Method)
					paddedName := padNameToMinLength(ui.CurrentRequest.Name, 4)
					iconColor := config.C.UI.SelectedRequestIconColor
					coloredIcon := fmt.Sprintf("[%s]%s[-:-:-]", iconColor, config.C.UI.SelectedRequestIcon)
					ui.CurrentSelectedNode.SetText(fmt.Sprintf("%s%s%s", coloredIcon, coloredMethod, paddedName))

					saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
				}
			}
		}
	})

	ui.ContentTypeDropdown.SetSelectedFunc(func(text string, index int) {
		if ui.ProgrammaticallyUpdatingContentType {
			return
		}

		if ui.CurrentRequest != nil {
			contentTypes := []string{"JSON", "Multipart", "No Body"}
			if index >= 0 && index < len(contentTypes) {
				newContentType := contentTypes[index]
				if newContentType != ui.CurrentRequest.ContentType {
					oldContentType := ui.CurrentRequest.ContentType
					ui.CurrentRequest.ContentType = newContentType

					// Switch the body UI based on content type
					ui.switchBodyContent(newContentType, oldContentType)

					// If switching TO multipart, parse the current body text into fields
					// This must be called AFTER switchBodyContent which restores the saved multipart content
					if newContentType == "Multipart" && oldContentType != "Multipart" {
						updateMultipartFieldsFromBody(ui.CurrentRequest.Body, ui)
					}
					saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
				}
			}
		}
	})

	// Add change handler for URL input
	ui.URLInput.SetChangedFunc(func(text string) {
		// Skip if we're programmatically updating from tree selection
		if ui.ProgrammaticallyUpdatingURL {
			return
		}

		if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
			ui.CurrentRequest.URL = text

			saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
		}
	})

	// Add change handler for body text area (only for edit mode)
	ui.BodyEditPanel.SetChangedFunc(func() {
		if ui.BodyEditMode && ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
			ui.CurrentRequest.Body = ui.BodyEditPanel.GetText()
			ui.CurrentBodyContent = ui.CurrentRequest.Body

			// Also save to JSONBodyContent if we're in JSON mode
			if ui.CurrentRequest.ContentType == "JSON" {
				ui.JSONBodyContent = ui.CurrentBodyContent
			}

			saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
		}
	})

	// Handle tree highlighting (for navigation)
	highlightTreeNode := func(node *tview.TreeNode) {
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(ui.Colors.TreeSelection).Foreground(ui.Colors.Foreground))
	}

	// Handle tree selection
	handleTreeSelection := func(node *tview.TreeNode) {
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(ui.Colors.TreeSelection).Foreground(ui.Colors.Foreground))

		// Remove icon from previously selected request only when another request is selected
		if ui.LastSelectedRequestNode != nil && ui.LastSelectedRequestNode != node {
			if prevReq := ui.requestFromNode(ui.LastSelectedRequestNode); prevReq != nil {
				// Only remove icon if the new selection is also a request
				if nodeIsRequest(node) {
					coloredMethod := getColoredMethod(prevReq.Method)
					paddedName := padNameToMinLength(prevReq.Name, 4)
					// Restore to reserved space (icon width + fixed space)
					iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon) // + 1
					spacePadding := strings.Repeat(" ", iconWidth)
					ui.LastSelectedRequestNode.SetText(fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName))
				}
			}
		}

		// Add icon to newly selected node if it's a request
		if req := ui.requestFromNode(node); req != nil {
			coloredMethod := getColoredMethod(req.Method)
			paddedName := padNameToMinLength(req.Name, 4)
			// Create colored icon with configured color, always add space after icon
			iconColor := config.C.UI.SelectedRequestIconColor
			coloredIcon := fmt.Sprintf("[%s]%s[-:-:-]", iconColor, config.C.UI.SelectedRequestIcon)
			node.SetText(fmt.Sprintf("%s%s%s", coloredIcon, coloredMethod, paddedName))

			// Set flag to prevent the SetChangedFunc from firing
			ui.ProgrammaticallyUpdatingURL = true
			ui.URLInput.SetText(req.URL)
			ui.ProgrammaticallyUpdatingURL = false

			ui.SyncBodyContent(req.Body)
			// Initialize saved body content based on content type
			if req.ContentType == "JSON" {
				ui.JSONBodyContent = req.Body
				ui.LastJSONBodyContent = req.Body
			} else if req.ContentType == "Multipart" {
				ui.MultipartBodyContent = req.Body
			}
			setHeadersInUI(ui.Colors, req.Headers, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)
			setQueryParamsInUI(ui.Colors, req.QueryParams, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)

			// Set current request for persistence
			ui.CurrentSelectedNode = node
			ui.LastSelectedRequestNode = node

			// Save selected request path
			path := findNodePath(ui.RootNode, node)
			if path != nil {
				ui.WorkspaceData.SelectedRequest = path
				// Save workspace to persist selection
				workspace.SaveWorkspace(ui.WorkspaceData)
			}

			// Resolve the live request pointer by stable ID. The node holds an
			// immutable NodeRef{ID}, so this always returns current data even
			// after the tree was rebuilt or the underlying slice was shifted.
			ui.CurrentRequest = req
			ui.CurrentRequestID = req.ID

			// Set method and content type in dropdowns AFTER currentRequest is set
			if ui.CurrentRequest != nil {
				syncMethodDropdown(ui.CurrentRequest, ui.MethodDropdown, &ui.ProgrammaticallyUpdatingMethod)
				syncContentTypeDropdown(ui.CurrentRequest, ui.ContentTypeDropdown, &ui.ProgrammaticallyUpdatingContentType)

				// Switch body UI based on content type
				// When loading, old content type is the same as new (no transition)
				ui.switchBodyContent(ui.CurrentRequest.ContentType, ui.CurrentRequest.ContentType)

				// Update multipart fields if this is a multipart request
				// This must be called AFTER switchBodyContent which sets up the container
				if ui.CurrentRequest.ContentType == "Multipart" {
					updateMultipartFieldsFromBody(ui.CurrentRequest.Body, ui)
				}

				// Show last response if available
				if len((*ui.CurrentRequest).ResponseHistory) > 0 {
					lastResponse := (*ui.CurrentRequest).ResponseHistory[len((*ui.CurrentRequest).ResponseHistory)-1]
					// Convert workspace.HTTPResponse to cmd.HTTPResponse for formatting
					cmdResp := &HTTPResponse{
						StatusCode: lastResponse.StatusCode,
						Status:     lastResponse.Status,
						Headers:    lastResponse.Headers,
						Cookies:    convertCookies(lastResponse.Cookies),
						Body:       lastResponse.Body,
						BodyBytes:  []byte(lastResponse.Body),
						Duration:   lastResponse.Duration,
						Timestamp:  lastResponse.Timestamp,
						BodySize:   len(lastResponse.Body),
					}

					// For historical responses, show when that specific request was made
					ui.LastResponse = cmdResp
					updateResponseTabs(cmdResp, &lastResponse.Timestamp, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)
				} else {
					// Clear response if no history
					ui.LastResponse = nil
					updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)
				}
			}
		} else if nodeIsCollection(node) {
			// Save current request headers before clearing
			if ui.CurrentRequest != nil {
				ui.CurrentRequest.Headers = getHeadersFromUI()
			}

			// Clear current request when a collection is selected
			ui.CurrentRequest = nil
			ui.CurrentRequestID = ""
			ui.CurrentSelectedNode = nil
			ui.LastResponse = nil
			updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)

			// Track the last selected request node
			if nodeIsRequest(node) {
				ui.LastSelectedRequestNode = node
			}
		}
	}

	// Set the tree selection handler
	ui.TreeSelectionHandler = handleTreeSelection

	// Set the tree highlight handler for navigation
	ui.TreeHighlightHandler = highlightTreeNode

	// Add focus handlers to update NavCurrentContainer when panels are clicked/focused
	ui.WorkspaceSelector.SetFocusFunc(func() {
		ui.NavCurrentContainer = 0
		ui.NavCurrentChild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.WorkspaceConfigButton.SetFocusFunc(func() {
		ui.NavCurrentContainer = 0
		ui.NavCurrentChild = 1
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
		ui.WorkspaceConfigButton.Flash()
	})
	ui.EnvDropdown.SetFocusFunc(func() {
		ui.NavCurrentContainer = 1
		ui.NavCurrentChild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.EnvConfigButton.SetFocusFunc(func() {
		ui.NavCurrentContainer = 1
		ui.NavCurrentChild = 1
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
		ui.EnvConfigButton.Flash()
	})
	ui.CollectionsTreeView.SetFocusFunc(func() {
		ui.NavCurrentContainer = 2
		ui.NavCurrentChild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.MethodDropdown.SetFocusFunc(func() {
		ui.NavCurrentContainer = 3
		ui.NavCurrentChild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.URLInput.SetFocusFunc(func() {
		ui.NavCurrentContainer = 3
		ui.NavCurrentChild = 1
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.SendButton.SetFocusFunc(func() {
		ui.NavCurrentContainer = 3
		ui.NavCurrentChild = 2
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
		ui.SendButton.Flash()
	})
	ui.CurlButton.SetFocusFunc(func() {
		ui.NavCurrentContainer = 3
		ui.NavCurrentChild = 3
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
		ui.CurlButton.Flash()
	})
	ui.ContentTypeDropdown.SetFocusFunc(func() {
		ui.NavCurrentContainer = 4
		ui.NavRequestInTabHeaders = false
		ui.NavCurrentChild = 0
		ui.NavCurrentSubchild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.TabHeader.SetFocusFunc(func() {
		ui.NavCurrentContainer = 4
		ui.NavRequestInTabHeaders = true
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.BodyViewPanel.SetFocusFunc(func() {
		ui.NavCurrentContainer = 4
		ui.NavRequestInTabHeaders = false
		ui.NavCurrentChild = 0 // Body tab
		// Use appropriate subchild index based on content type
		if getCurrentContentType(ui) == "No Body" {
			ui.NavCurrentSubchild = 3 // No Body panel
		} else {
			ui.NavCurrentSubchild = 1 // View panel
		}
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.BodyEditPanel.SetFocusFunc(func() {
		ui.NavCurrentContainer = 4
		ui.NavRequestInTabHeaders = false
		ui.NavCurrentChild = 0    // Body tab
		ui.NavCurrentSubchild = 1 // Edit panel
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.ResponseTabHeader.SetFocusFunc(func() {
		ui.NavCurrentContainer = 5
		ui.NavResponseInTabHeaders = true
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	ui.ResponsePreviewPanel.SetFocusFunc(func() {
		ui.NavCurrentContainer = 5
		ui.NavResponseInTabHeaders = false
		ui.NavCurrentChild = 0
		syncMainCycleWithExperimental(ui)
		ui.UpdateFooter()
	})
	if ui.MultipartAddButton != nil {
		ui.MultipartAddButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 0
			ui.NavCurrentSubchild = 2
			ui.NavCurrentMultipartElement = 0
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.MultipartAddButton.Flash()
		})
	}
	if ui.MultipartDeleteAllButton != nil {
		ui.MultipartDeleteAllButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 0
			ui.NavCurrentSubchild = 2
			ui.NavCurrentMultipartElement = 1
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.MultipartDeleteAllButton.Flash()
		})
	}
	if ui.AddHeaderButton != nil {
		ui.AddHeaderButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 3
			ui.NavCurrentHeaderRowElement = 0
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.AddHeaderButton.Flash()
		})
	}
	if ui.DeleteAllHeadersButton != nil {
		ui.DeleteAllHeadersButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 3
			ui.NavCurrentHeaderRowElement = 1
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.DeleteAllHeadersButton.Flash()
		})
	}
	if ui.AddQueryParamButton != nil {
		ui.AddQueryParamButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 2
			ui.NavCurrentQueryParamRowElement = 0
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.AddQueryParamButton.Flash()
		})
	}
	if ui.DeleteAllQueryParamsButton != nil {
		ui.DeleteAllQueryParamsButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 2
			ui.NavCurrentQueryParamRowElement = 1
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.DeleteAllQueryParamsButton.Flash()
		})
	}
	if ui.AddCookieButton != nil {
		ui.AddCookieButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 4
			ui.NavCurrentCookieRowElement = 0
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.AddCookieButton.Flash()
		})
	}
	if ui.DeleteAllCookiesButton != nil {
		ui.DeleteAllCookiesButton.SetFocusFunc(func() {
			ui.NavCurrentContainer = 4
			ui.NavRequestInTabHeaders = false
			ui.NavCurrentChild = 4
			ui.NavCurrentCookieRowElement = 1
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
			ui.DeleteAllCookiesButton.Flash()
		})
	}
	if ui.ResponseHeadersPanel != nil {
		// ResponseHeadersPanel is tview.Primitive, which has SetFocusFunc
		// But we need to use a type assertion to a concrete type if we want to call SetFocusFunc?
		// No, Primitive interface doesn't have SetFocusFunc.
		// Most tview components have it.
		if component, ok := ui.ResponseHeadersPanel.(interface{ SetFocusFunc(func()) *tview.Box }); ok {
			component.SetFocusFunc(func() {
				ui.NavCurrentContainer = 5
				ui.NavResponseInTabHeaders = false
				ui.NavCurrentChild = 1
				syncMainCycleWithExperimental(ui)
				ui.UpdateFooter()
			})
		}
	}
	if ui.ResponseCookiesPanel != nil {
		ui.ResponseCookiesPanel.SetFocusFunc(func() {
			ui.NavCurrentContainer = 5
			ui.NavResponseInTabHeaders = false
			ui.NavCurrentChild = 2
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
		})
	}
	if ui.ResponseTimelinePanel != nil {
		ui.ResponseTimelinePanel.SetFocusFunc(func() {
			ui.NavCurrentContainer = 5
			ui.NavResponseInTabHeaders = false
			ui.NavCurrentChild = 3
			syncMainCycleWithExperimental(ui)
			ui.UpdateFooter()
		})
	}

	// Set up collections tree view changed function (called on current node change)
	ui.CollectionsTreeView.SetChangedFunc(func(node *tview.TreeNode) {
		// Do nothing on current node change - loading only on explicit selection
	})

	// Set up collections tree view selected function
	ui.CollectionsTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		handleTreeSelection(node)
	})

	// Set the initial selected style for the current node and load it
	currentNode := ui.CollectionsTreeView.GetCurrentNode()
	if currentNode != nil {
		currentNode.SetSelectedTextStyle(tcell.StyleDefault.Background(ui.Colors.TreeSelection).Foreground(ui.Colors.Foreground))
		handleTreeSelection(currentNode)
	}

	// Set up environment config button click handler
	ui.EnvConfigButton.SetSelectedFunc(func() {
		showEnvironmentModal(ui)
	})

	// Set up workspace config button click handler
	ui.WorkspaceConfigButton.SetSelectedFunc(func() {
		// Show workspace configuration modal
		showWorkspaceModal(ui)
	})

	// Save environment selection function
	saveEnvironmentSelection := func(index int) {
		if ui.ProgrammaticallyUpdatingEnv || index < 0 {
			return
		}
		var selectedEnvName string
		if index == 0 {
			selectedEnvName = "" // Empty represents "Base Environment" (index 0)
		} else if index > 0 && index <= len(*ui.EnvironmentsData) {
			selectedEnvName = (*ui.EnvironmentsData)[index-1].Name
		}

		// Update in-memory state
		ui.WorkspaceData.SelectedEnvironment = selectedEnvName

		// Save selected environment to workspace file immediately
		if err := workspace.SaveWorkspace(ui.WorkspaceData); err != nil {
			// Handle error silently for now
		}
	}

	ui.EnvDropdown.SetSelectedFunc(func(text string, index int) {
		saveEnvironmentSelection(index)
	})

	// Add send button functionality
	ui.SendButton.SetSelectedFunc(func() {
		// Prevent multiple requests
		if ui.RequestInProgress {
			return
		}

		sendStartTime := time.Now()
		var pluginWarnings []string

		// Get current request data from UI
		_, method := ui.MethodDropdown.GetCurrentOption()
		url := ui.URLInput.GetText()
		body := ""
		contentType := ""
		if ui.CurrentRequest != nil {
			body = ui.CurrentRequest.Body
			contentType = ui.CurrentRequest.ContentType
		}
		headers := getHeadersFromUI()
		queryParams := getQueryParamsFromUI()

		// For multipart requests, collect the current UI state
		if contentType == "Multipart" {
			body = collectMultipartFieldsForSend()
			// If no fields are filled, don't send as multipart
			if body == "" {
				contentType = ""
			} else {
				// Remove any manual Content-Type header to allow automatic multipart setting
				if _, ok := headers["Content-Type"]; ok {
					delete(headers, "Content-Type")
				}
				if _, ok := headers["content-type"]; ok {
					delete(headers, "content-type")
				}
			}
		}

		// Determine collection and request name
		collection := ""
		requestName := ""
		if ui.CurrentRequest != nil {
			requestName = ui.CurrentRequest.Name
			if ui.WorkspaceData != nil {
				if parent := workspace.FindParentCollectionOfRequest(&ui.WorkspaceData.Collections, ui.CurrentRequest.ID); parent != nil {
					collection = parent.Name
				}
			}
		}

		workspaceName := "Default"
		if ui.WorkspaceData != nil {
			workspaceName = ui.WorkspaceData.Name
		}

		// Get current environment variables
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		envVars := resolveEnvVarsFromIndex(currentEnvIndex, *ui.EnvironmentsData, ui.PluginManager, workspaceName, config.C.Plugins.Config)

		// Filter out disabled headers before hooks and substitution
		filteredHeaders := make(map[string]workspace.Entry)
		for key, entry := range headers {
			if entry.Enabled {
				filteredHeaders[key] = entry
			}
		}
		headersStr := entriesToStringMap(filteredHeaders)

		// Filter out disabled query params before hooks and substitution
		filteredParams := make(map[string]workspace.Entry)
		for key, entry := range queryParams {
			if entry.Enabled {
				filteredParams[key] = entry
			}
		}
		queryParamsStr := entriesToStringMap(filteredParams)

		// Pre-variable-substitution hook: plugins can resolve custom tags before env vars are substituted
		preContext := &plugins.HookContext{
			Request: &plugins.RequestData{
				Method:      method,
				URL:         url,
				Headers:     headersStr,
				Body:        body,
				Collection:  collection,
				RequestName: requestName,
			},
			Environment: envVars,
			Config:      config.C.Plugins.Config,
			Workspace:   workspaceName,
		}
		if preContext.Config == nil {
			preContext.Config = config.C.Plugins.Config
		}
		if ui.PluginManager != nil {
			if err := ui.PluginManager.ExecuteHooks(plugins.PreVariableSubstitution, preContext); err != nil {
				pluginWarnings = append(pluginWarnings, err.Error())
			}
		}

		// Update local variables from hook modifications
		url = preContext.Request.URL
		body = preContext.Request.Body
		headersStr = preContext.Request.Headers

		// Substitute environment variables in URL, body, headers, and query params
		url = substituteVariables(url, envVars)
		body = substituteVariables(body, envVars)
		headersStr = substituteVariablesInHeaders(headersStr, envVars)
		queryParamsStr = substituteVariablesInHeaders(queryParamsStr, envVars)

		// Skip built-in command-runner resolution when the external plugin is
		// loaded to avoid double execution (DRY).
		hasExternalCommandRunner := false
		if ui.PluginManager != nil {
			_, hasExternalCommandRunner = ui.PluginManager.GetPlugin("command-runner")
		}
		if !hasExternalCommandRunner {
			// Execute any command-runner tags that were substituted into the request
			// fields or written directly in the body / URL / headers.
			url = processCommandRunnerTags(url, envVars)
			body = processCommandRunnerTags(body, envVars)
			for k, v := range headersStr {
				headersStr[k] = processCommandRunnerTags(v, envVars)
			}
			for k, v := range queryParamsStr {
				queryParamsStr[k] = processCommandRunnerTags(v, envVars)
			}
		}

		requestData := &plugins.RequestData{
			Method:      method,
			URL:         url,
			Headers:     headersStr,
			Body:        body,
			Collection:  collection,
			RequestName: requestName,
		}

		context := &plugins.HookContext{
			Request:     requestData,
			Environment: envVars,
			Config:      config.C.Plugins.Config,
			Workspace:   workspaceName,
		}

		// Ensure config is available for all hooks
		if context.Config == nil {
			context.Config = config.C.Plugins.Config
		}

		if ui.PluginManager != nil {
			if err := ui.PluginManager.ExecuteHooks(plugins.PreSend, context); err != nil {
				pluginWarnings = append(pluginWarnings, err.Error())
			}
		}

		// Update request fields from context (plugins may have modified them)
		url = context.Request.URL
		body = context.Request.Body
		headersStr = context.Request.Headers

		// Also update the request data in context to reflect the final fields for logging
		context.Request.URL = url
		context.Request.Body = body
		context.Request.Headers = headersStr

		// Mark request as in progress and update UI
		ui.RequestInProgress = true
		ui.SendButton.SetSending(true)

		syncCookiesFromUI(ui.WorkspaceData)

		// Send the request in a goroutine
		go func() {
			resp, err := SendRequest(method, url, body, contentType, headersStr, queryParamsStr, &ui.WorkspaceData.CookieJar)

			// Use QueueUpdateDraw to handle the response on the main thread
			ui.App.QueueUpdateDraw(func() {
				defer func() {
					// Reset UI state when done
					ui.RequestInProgress = false
					ui.SendButton.SetSending(false)
				}()

				if err != nil {
					if ui.PluginManager != nil {
						if hookErr := ui.PluginManager.ExecuteHooks(plugins.OnError, context); hookErr != nil {
							pluginWarnings = append(pluginWarnings, hookErr.Error())
						}
					}
					updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)
					errText := fmt.Sprintf("[red]Error: %v[-]", err)
					if len(pluginWarnings) > 0 {
						errText = "[yellow]⚠ Plugin warnings:\n" + strings.Join(pluginWarnings, "\n") + "[-]\n\n" + errText
					}
					ui.ResponsePreviewPanel.SetText(errText)
					return
				}

				resp.TotalDuration = time.Since(sendStartTime)

				context.Response = &plugins.ResponseData{
					StatusCode: resp.StatusCode,
					Status:     resp.Status,
					Headers:    resp.Headers,
					Body:       resp.Body,
					Duration:   resp.Duration.Milliseconds(),
				}

				// Ensure config is available for PostReceive hook
				if context.Config == nil {
					context.Config = config.C.Plugins.Config
				}

				// Run plugin hooks that might modify data but not UI
				if ui.PluginManager != nil {
					if err := ui.PluginManager.ExecuteHooks(plugins.PostReceive, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.ResponseValidation, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.ResponseTransform, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
				}

				// Check if response is binary and needs download
				contentType := ""
				if resp.Headers != nil {
					if ct, ok := resp.Headers["Content-Type"]; ok && len(ct) > 0 {
						contentType = ct[0]
					}
				}
				if isBinaryContentType(contentType) {
					// For binary responses, open file dialog to save the file
					suggestedName := "response"
					if headers, ok := resp.Headers["Content-Disposition"]; ok && len(headers) > 0 {
						parts := strings.Split(headers[0], "filename=")
						if len(parts) > 1 {
							suggestedName = strings.Trim(strings.Split(parts[1], ";")[0], "\" ")
						}
					}
					if suggestedName == "response" {
						switch {
						case strings.HasPrefix(contentType, "image/png"):
							suggestedName = "response.png"
						case strings.HasPrefix(contentType, "image/jpeg"):
							suggestedName = "response.jpg"
						case strings.HasPrefix(contentType, "image/gif"):
							suggestedName = "response.gif"
						case strings.HasPrefix(contentType, "image/webp"):
							suggestedName = "response.webp"
						case strings.HasPrefix(contentType, "application/pdf"):
							suggestedName = "response.pdf"
						case strings.HasPrefix(contentType, "application/zip"):
							suggestedName = "response.zip"
						case strings.HasPrefix(contentType, "video/"):
							suggestedName = "response.video"
						case strings.HasPrefix(contentType, "audio/"):
							suggestedName = "response.audio"
						}
					}
					suggestedPath := filepath.Join(os.Getenv("HOME"), suggestedName)

					currentFocus := ui.App.GetFocus()
					onSave := func(path string) {
						err := os.WriteFile(path, resp.BodyBytes, 0644)
						if err != nil {
							showErrorModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Failed to save file: %v", err), currentFocus, ui.Colors)
						} else {
							showSuccessModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Downloaded to: %s", path), currentFocus, ui.Colors)
						}
					}
					openSaveFileModal(ui.App, ui.Pages, ui.Colors, suggestedPath, onSave, nil)

					// Update response with download message instead of body
					resp.Body = fmt.Sprintf("[Binary response downloaded - %d bytes]", len(resp.BodyBytes))
				}

				// Save any plugin-modified environment variables back to the current environment
				if context.Environment != nil && len(context.Environment) > 0 {
					savePluginEnvironmentChanges(context.Environment, currentEnvIndex, ui.EnvironmentsData)
				}

				// Store the response in the current request's history. Re-resolve the
				// request by ID first: the request was sent from a goroutine and the
				// Requests backing array may have shifted (e.g. the user moved another
				// request) while the response was in flight.
				ui.refreshCurrentRequest()
				if ui.CurrentRequest != nil {
					body := resp.Body
					respContentType := ""
					if resp.Headers != nil {
						if ct, ok := resp.Headers["Content-Type"]; ok && len(ct) > 0 {
							respContentType = ct[0]
						}
					}
					if len(body) > 1024*1024 || isBinaryContentType(respContentType) {
						body = fmt.Sprintf("[Response body skipped - %d bytes]", len(body))
					}

					workspaceResp := workspace.HTTPResponse{
						StatusCode: resp.StatusCode,
						Status:     resp.Status,
						Headers:    resp.Headers,
						Cookies:    httpCookieToWorkspaceCookie(resp.Cookies),
						Body:       body,
						Duration:   resp.Duration,
						Timestamp:  resp.Timestamp,
					}

					maxHistory := config.C.MaxResponseHistory
					if maxHistory > 0 && len((*ui.CurrentRequest).ResponseHistory) >= maxHistory {
						(*ui.CurrentRequest).ResponseHistory = (*ui.CurrentRequest).ResponseHistory[1:]
					}
					(*ui.CurrentRequest).ResponseHistory = append((*ui.CurrentRequest).ResponseHistory, workspaceResp)

					if ui.CurrentSelectedNode != nil {
						saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
					}
				}

				if ui.PluginManager != nil {
					if err := ui.PluginManager.ExecuteHooks(plugins.PostSave, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.PostRequest, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.PreUIUpdate, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.PostUIUpdate, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
					if err := ui.PluginManager.ExecuteHooks(plugins.PreSave, context); err != nil {
						pluginWarnings = append(pluginWarnings, err.Error())
					}
				}

				// Prepend plugin timeout warnings to the response body preview
				if len(pluginWarnings) > 0 {
					warningText := "[yellow]⚠ Plugin warnings:\n" + strings.Join(pluginWarnings, "\n") + "[-]\n\n"
					resp.Body = warningText + resp.Body
				}

				// Update the response tabs with the new response
				now := time.Now()
				ui.LastResponse = resp
				updateResponseTabs(resp, &now, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse, ui.SaveResponse)
				RefreshCookiesTab(ui.WorkspaceData.CookieJar.Cookies, ui.Colors, ui.App, ui.Pages)
			})
		}()
	})

	// Add curl export button functionality
	ui.CurlButton.SetSelectedFunc(func() {
		// Get current request data from UI
		_, method := ui.MethodDropdown.GetCurrentOption()
		url := ui.URLInput.GetText()
		body := ""
		contentType := ""
		if ui.CurrentRequest != nil {
			body = ui.CurrentRequest.Body
			contentType = ui.CurrentRequest.ContentType
		}
		headers := getHeadersFromUI()
		queryParams := getQueryParamsFromUI()

		workspaceName := "Default"
		if ui.WorkspaceData != nil {
			workspaceName = ui.WorkspaceData.Name
		}

		// Get current environment variables
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		envVars := resolveEnvVarsFromIndex(currentEnvIndex, *ui.EnvironmentsData, ui.PluginManager, workspaceName, config.C.Plugins.Config)

		// Filter out disabled headers before hooks and substitution
		filteredHeaders := make(map[string]workspace.Entry)
		for key, entry := range headers {
			if entry.Enabled {
				filteredHeaders[key] = entry
			}
		}
		headersStr := entriesToStringMap(filteredHeaders)

		// Filter out disabled query params before hooks and substitution
		filteredParams := make(map[string]workspace.Entry)
		for key, entry := range queryParams {
			if entry.Enabled {
				filteredParams[key] = entry
			}
		}
		queryParamsStr := entriesToStringMap(filteredParams)

		// Pre-variable-substitution hook
		preContext := &plugins.HookContext{
			Request: &plugins.RequestData{
				Method:      method,
				URL:         url,
				Headers:     headersStr,
				Body:        body,
				Collection:  "",
				RequestName: "",
			},
			Environment: envVars,
			Config:      config.C.Plugins.Config,
			Workspace:   workspaceName,
		}
		if preContext.Config == nil {
			preContext.Config = config.C.Plugins.Config
		}
		if ui.PluginManager != nil {
			_ = ui.PluginManager.ExecuteHooks(plugins.PreVariableSubstitution, preContext)
		}

		// Update local variables from hook modifications
		url = preContext.Request.URL
		body = preContext.Request.Body
		headersStr = preContext.Request.Headers

		// Substitute environment variables in URL, body, headers, and query params
		url = substituteVariables(url, envVars)
		body = substituteVariables(body, envVars)
		headersStr = substituteVariablesInHeaders(headersStr, envVars)
		queryParamsStr = substituteVariablesInHeaders(queryParamsStr, envVars)

		// Skip built-in command-runner resolution when the external plugin is loaded.
		hasExternalCommandRunnerCurl := false
		if ui.PluginManager != nil {
			_, hasExternalCommandRunnerCurl = ui.PluginManager.GetPlugin("command-runner")
		}
		if !hasExternalCommandRunnerCurl {
			// Execute any command-runner tags that were substituted into the request
			// fields or written directly in the body / URL / headers.
			url = processCommandRunnerTags(url, envVars)
			body = processCommandRunnerTags(body, envVars)
			for k, v := range headersStr {
				headersStr[k] = processCommandRunnerTags(v, envVars)
			}
			for k, v := range queryParamsStr {
				queryParamsStr[k] = processCommandRunnerTags(v, envVars)
			}
		}

		curlCommand := generateCurlCommand(method, url, headersStr, body, contentType, queryParamsStr)

		copyToClipboard(curlCommand)

		// Show a brief notification (could be improved with a proper toast notification)
		ui.FooterRight.SetText(footerStatusText(ui.Colors, footerStatusCurlCopied))
		go func() {
			time.Sleep(2 * time.Second)
			ui.App.QueueUpdateDraw(func() {
				ui.FooterRight.SetText(footerBrandText(ui.Colors) + " ")
			})
		}()
	})

	// Track popup/form state
	var isFormPopupActive bool = false

	// Helper function to set popup state
	setFormPopupActive := func(active bool) {
		isFormPopupActive = active
	}

	// Set up main application input capture for navigation and shortcuts
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// When focus is on the collections tree, let Enter pass through to the
		// TreeView-specific input capture so request selection works.  Some
		// global shortcuts (e.g. ctrl+m) are indistinguishable from Enter in
		// terminals, so we must pass Enter through explicitly here.
		if (event.Key() == tcell.KeyEnter || event.Key() == tcell.KeyLF) &&
			ui.App.GetFocus() == ui.CollectionsTreeView {
			currentPage, _ := ui.Pages.GetFrontPage()
			if currentPage == "main" {
				return event
			}
		}
		// Handle 'q' to quit or close modals, but disable when focused on an input
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			focus := ui.App.GetFocus()
			if _, ok := focus.(*tview.InputField); ok {
				return event
			}
			if _, ok := focus.(*tview.TextArea); ok {
				return event
			}

			currentPage, _ := ui.Pages.GetFrontPage()
			if currentPage == "workspaceSearch" {
				ui.Pages.RemovePage(currentPage)
				ui.Pages.SwitchToPage("main")
				ui.App.SetFocus(ui.WorkspaceSelector)
				ui.UpdateFooter()
				return nil
			}
			if currentPage != "main" && currentPage != "envVariables" {
				// On modal, 'q' closes the modal
				ui.Pages.RemovePage(currentPage)
				ui.Pages.SwitchToPage("main")
				ui.App.SetFocus(ui.CollectionsTreeView)
				ui.UpdateFooter()
				return nil
			} else if currentPage == "main" {
				// On main page, 'q' quits
				// Save expansion state before quitting if in "remember" mode
				if config.C.UI.CollectionExpansion == "remember" {
					if err := workspace.SaveExpansionState(&ui.WorkspaceData.Collections); err != nil {
						fmt.Printf("Warning: Failed to save expansion state: %v\n", err)
					}
				}
				ui.App.Stop()
				return nil
			}
		}

		// When in body edit mode and focused on bodyEditPanel, pass all input through to allow pasting
		// except for the command runner shortcut
		if ui.BodyEditMode && ui.App.GetFocus() == ui.BodyEditPanel {
			if !isCommandRunnerEvent(event) {
				return event
			}
		}

		// Allow URLVariableInput to handle its own Enter key events
		if event.Key() == tcell.KeyEnter && ui.App.GetFocus() == ui.URLInput {
			return event
		}

		// Allow URLVariableInput to handle its own 'i' key events
		if event.Rune() == 'i' && ui.App.GetFocus() == ui.URLInput {
			return event
		}

		// Open workspace quick-search when '/' is pressed in the workspace dropdown context
		if event.Rune() == '/' {
			currentFocus := ui.App.GetFocus()
			if currentFocus == ui.WorkspaceSelector || isDropdownOpen(ui.WorkspaceSelector) {
				showWorkspaceSearchModal(ui)
				return nil
			}
		}

		// Allow dropdown lists to handle their own input when dropdown is open
		if isDropdownOpen(ui.EnvDropdown) || isDropdownOpen(ui.WorkspaceSelector) || isDropdownOpen(ui.MethodDropdown) {
			return event
		}

		// Allow dropdown lists to handle their own input
		if _, ok := ui.App.GetFocus().(*tview.List); ok {
			return event
		}

		// Check if we're focused on a form input field - do this BEFORE checking global keybindings
		focus := ui.App.GetFocus()

		// If we're in a form input field, don't handle collection shortcuts
		// BUT allow Tab, Backtab, the command runner shortcut and the command
		// palette (Ctrl+P) to pass through
		if _, isInput := focus.(*tview.InputField); isInput {
			if event.Key() != tcell.KeyTab && event.Key() != tcell.KeyBacktab && !isCommandRunnerEvent(event) && event.Key() != tcell.KeyCtrlP {
				return event // Let input fields handle their own keys
			}
		}
		if _, isTextArea := focus.(*tview.TextArea); isTextArea {
			if !isCommandRunnerEvent(event) && event.Key() != tcell.KeyCtrlP {
				return event // Let text areas handle their own keys
			}
		}

		// Check if we're in a form popup - do this BEFORE checking global keybindings
		currentPage, _ := ui.Pages.GetFrontPage()
		formPopups := []string{
			"newCollection",
			"newRequest",
			"workspaceMenu",
			"envVariables",
			"moveCollection",
			"moveRequest",
			"renameCollection",
			"renameRequest",
			"deleteCollection",
			"deleteRequest",
			"deleteAllHeaders",
			"deleteAllQueryParams",
			"duplicateRequest",
			"cloneEnvironment",
			"createWorkspace",
			"deleteEnvironment",
			"deleteWorkspace",
			"duplicateWorkspace",
			"renameEnvironment",
			"renameWorkspace",
			"workspaceModal",
			"workspaceSearch",
			"commandPalette",
		}

		isFormPopup := false
		for _, popup := range formPopups {
			if currentPage == popup {
				isFormPopup = true
				break
			}
		}

		// If we're in a form popup, disable certain keybindings
		if isFormPopup {
			// Check for keys that should work as normal characters in forms
			problematicKeys := map[rune]bool{
				'N': true, // new collection
				'n': true, // new request
				'r': true, // rename item
				'm': true, // move item
				'd': true, // delete item
				'D': true, // duplicate request
				'i': true, // enter insert mode
			}

			if problematicKeys[event.Rune()] {
				return event // Let the form handle these keys as normal characters
			}
		}

		// First check if this is a global keybinding
		if result := ui.KeyManager.HandleKeyEvent(ui, event, "global"); result != event {
			return result
		}

		// Open command-runner tag editor when pressing 'e' while focused on a view-mode
		// input that contains a command-runner tag.  Don't intercept in editable fields
		// (InputField / TextArea) where 'e' is a normal character.
		if event.Rune() == 'e' { // TODO: Change to another key
			focus := ui.App.GetFocus()
			if _, isInput := focus.(*tview.InputField); isInput {
				return event
			}
			if _, isTextArea := focus.(*tview.TextArea); isTextArea {
				return event
			}
			target := findActiveTextInput(ui)
			if target != nil && hasEditableTag(target.getText()) {
				openTagEditorForField(ui)
				return nil
			}
		}

		// If we're in any modal, don't handle tab navigation
		if currentPage != "main" {
			return event
		}

		if event.Key() == tcell.KeyTab {
			// Handle tab navigation logic here
			return handleTabNavigation(ui, event)
		}

		if event.Key() == tcell.KeyBacktab {
			// Handle backtab navigation logic here
			return handleBacktabNavigation(ui, event)
		}

		// Collection shortcuts (only when not in input fields and no form popup is active)
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'N' && !isFormPopupActive {
			form := createCollectionFormWithLocation(ui)
			modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
			setFormPopupActive(true)
			ui.Pages.AddPage("newCollection", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}

		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'n' && !isFormPopupActive {
			// New request - check if a collection or request is selected
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				var selectedCollection *workspace.Collection

				if col := ui.collectionFromNode(node); col != nil {
					// Collection is selected
					selectedCollection = col
				} else if req := ui.requestFromNode(node); req != nil {
					// Request is selected - find its parent collection by ID
					selectedCollection = workspace.FindParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.ID)
				}

				if selectedCollection != nil {
					form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
					modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
					setFormPopupActive(true)
					ui.Pages.AddPage("newRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}

		// Collection shortcuts - handle both normal operation and popup forms
		if ui.MainCycle.current == ui.PanelIndices.Collections {
			// Handle Esc to close popups
			if event.Key() == tcell.KeyEsc && isFormPopupActive {
				return event // Let Esc pass through to close the popup
			}

			// Handle 'N' and 'n' keys
			if event.Rune() == 'N' || event.Rune() == 'n' {
				if isFormPopupActive {
					// We're in a popup form, let the key pass through to the form input
					return event
				} else {
					// Normal operation - create new collection/request
					if event.Rune() == 'N' {
						form := createCollectionFormWithLocation(ui)
						modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
						setFormPopupActive(true)
						ui.Pages.AddPage("newCollection", modal, true, true)
						ui.App.SetFocus(form)
						return nil
					} else if event.Rune() == 'n' {
						// New request - check if a collection or request is selected
						node := ui.CollectionsTreeView.GetCurrentNode()
						if node != nil {
							var selectedCollection *workspace.Collection

							if col := ui.collectionFromNode(node); col != nil {
								// Collection is selected
								selectedCollection = col
							} else if req := ui.requestFromNode(node); req != nil {
								// Request is selected - find its parent collection by ID
								selectedCollection = workspace.FindParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.ID)
							}

							if selectedCollection != nil {
								form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
								modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
								setFormPopupActive(true)
								ui.Pages.AddPage("newRequest", modal, true, true)
								ui.App.SetFocus(form)
								return nil
							}
						}
					}
				}
			}
		}

		// Rename functionality (r)
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'r' {
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				if col := ui.collectionFromNode(node); col != nil {
					// Rename collection
					form := createRenameCollectionForm(ui.App, ui.Pages, col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
					ui.Pages.AddPage("renameCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req := ui.requestFromNode(node); req != nil {
					// Rename request - need to find parent collection
					form := createRenameRequestForm(ui.App, ui.Pages, req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
					ui.Pages.AddPage("renameRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}

		// Move collection/request functionality (m)
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'm' {
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				if col := ui.collectionFromNode(node); col != nil {
					form := createMoveCollectionForm(ui, col)
					modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
					ui.Pages.AddPage("moveCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req := ui.requestFromNode(node); req != nil {
					form := createMoveRequestForm(ui, req)
					modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
					ui.Pages.AddPage("moveRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}

		// Delete functionality (d)
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'd' {
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				if col := ui.collectionFromNode(node); col != nil {
					form := createDeleteCollectionConfirm(ui.App, ui.Pages, col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors, ui.DataManager)
					modal := createSizedModal(form, modalSizeConfirm, tcell.ColorDefault)
					ui.Pages.AddPage("deleteCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req := ui.requestFromNode(node); req != nil {
					form := createDeleteRequestConfirm(ui.App, ui.Pages, req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors, ui.DataManager)
					modal := createSizedModal(form, modalSizeConfirm, tcell.ColorDefault)
					ui.Pages.AddPage("deleteRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}

		// Open body in external editor
		if ui.MainCycle.current == ui.PanelIndices.Request && event.Key() == tcell.KeyF4 {
			if ui.CurrentRequest != nil {
				ui.App.Suspend(func() {
					modifiedContent, err := openInExternalEditor(ui.CurrentBodyContent, "json")
					if err != nil {
						fmt.Printf("Error opening external editor: %v\n", err)
						fmt.Println("Press Enter to continue...")
						var dummy string
						fmt.Scanln(&dummy)
						return
					}

					ui.SyncBodyContent(modifiedContent)
					if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
						ui.CurrentRequest.Body = modifiedContent
						saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
					}
				})
			}
			return nil
		}

		// Tab switching with number keys (1-4) when request data tabs are focused
		if handleTabSwitch(ui, event, currentHeaderRows) {
			return nil
		}

		// Vim-style modal editing: 'i' to enter insert mode
		if event.Rune() == 'i' && ui.MainCycle.current == ui.PanelIndices.Request && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
			// For Multipart content type the body tab hosts dual-mode inputs (Name, Value)
			// that handle 'i' themselves to switch those fields to edit mode. Falling through
			// here would suspend the TUI and open the external editor instead, so skip it.
			if getCurrentContentType(ui) == "Multipart" {
				return event
			}
			if ui.CurrentRequest != nil {
				ui.App.Suspend(func() {
					modifiedContent, err := openInExternalEditor(ui.CurrentBodyContent, "json")
					if err != nil {
						return
					}

					// Update the body with modified content
					ui.SyncBodyContent(modifiedContent)
					if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
						ui.CurrentRequest.Body = modifiedContent
						saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
					}
				})
			}
			return nil
		}

		// Arrow key navigation for tabs when request panel tabs are focused and not in body edit mode
		if handleTabSwitch(ui, event, currentHeaderRows) {
			return nil
		}

		return event
	})
}

// handleTabNavigation handles Tab key navigation
// getMaxChildForContainer returns the maximum child index for a given container
func getMaxChildForContainer(container int) int {
	switch container {
	case 0: // Workspace
		return 1 // WorkspaceSelector (0), WorkspaceMenu (1)
	case 1: // Environment
		return 1 // EnvironmentSelector (0), EnvironmentMenu (1)
	case 2: // Collections
		return 0 // Only TreeView (0)
	case 3: // URLBar
		return 3 // MethodDropdown (0), URLInput (1), SendButton (2), CurlButton (3)
	case 4: // Request
		return 4 // BodyTab (0), AuthTab (1), QueryTab (2), HeadersTab (3), CookiesTab (4)
	case 5: // Response
		return 3 // PreviewTab (0), HeadersTab (1), CookiesTab (2), TimelineTab (3)
	default:
		return 0
	}
}

// getContainerName returns the name of a container by index
func getContainerName(container int) string {
	switch container {
	case 0:
		return "Workspace"
	case 1:
		return "Environment"
	case 2:
		return "Collections"
	case 3:
		return "URLBar"
	case 4:
		return "Request"
	case 5:
		return "Response"
	default:
		return "Unknown"
	}
}

// hasSubchildren returns true if a container/child combination has subchildren
func hasSubchildren(container, child int, ui *UIOrchestrator) bool {
	// Request BodyTab (container 4, child 0) has subchildren when not in tab headers mode
	if container == 4 {
		// All request tabs can be considered to have subchildren in the new navigation model
		// because we want to enter them (except maybe empty ones, but for now allow entry)
		return !ui.NavRequestInTabHeaders
	}
	return false
}

// getMaxSubchildForChild returns the maximum subchild index for a given container/child
func getMaxSubchildForChild(container, child int, ui *UIOrchestrator) int {
	if container == 4 { // Request Panel
		switch child {
		case 0: // Body Tab
			// Check current content type to determine which subchildren are relevant
			contentType := getCurrentContentType(ui)
			if contentType != "" {
				switch contentType {
				case "JSON":
					return 1 // ContentTypeSelector (0), JSONEditor (1) - skip MultipartFields and NoBody
				case "Multipart":
					return 2 // ContentTypeSelector (0), MultipartFields (2) - skip JSONEditor and NoBody
				case "No Body":
					return 3 // ContentTypeSelector (0), NoBody (3) - skip JSONEditor and MultipartFields
				default:
					return 3 // Default to all subchildren
				}
			}
			return 3 // Default to all subchildren if no current request
		case 3: // Headers Tab
			// Headers Tab uses complex internal navigation, but we define 0 as the entry point
			return 0
		case 2: // Query Tab
			// Query Tab uses complex internal navigation like Headers
			return 0
		default: // Auth (1), Query (2)
			// Assuming single content area for now
			return 0
		}
	}
	return 0 // No subchildren by default
}

// getMaxFieldRowElement returns the maximum field row element index for a multipart field row.
// Layout order: Name (0), Type (1), Value (2), Browse/Empty (3), Checkbox (4), X button (5).
func getMaxFieldRowElement(fieldRow *MultipartFieldRow) int {
	if fieldRow == nil {
		return 0
	}
	return 5
}

func getMaxHeaderRowElement(headerRow *HeaderRow) int {
	if headerRow == nil {
		return 0
	}

	// Header rows have: Key input (0), Value input (1), Checkbox (2), Delete button (3)
	return 3
}

func getMaxQueryParamRowElement(queryParamRow *QueryParamRow) int {
	if queryParamRow == nil {
		return 0
	}

	// Query param rows have: Key input (0), Value input (1), Checkbox (2), Delete button (3)
	return 3
}

func getMaxCookieRowElement(cookieRow *CookieRow) int {
	if cookieRow == nil {
		return 0
	}

	// Cookie rows have: Domain (0), Name (1), Value (2), Path (3), Secure (4), HttpOnly (5), Checkbox (6), Delete button (7)
	return 7
}

// getCurrentContentType returns the current content type from the dropdown
func getCurrentContentType(ui *UIOrchestrator) string {
	if ui.ContentTypeDropdown != nil {
		_, contentType := ui.ContentTypeDropdown.GetCurrentOption()
		return contentType
	}
	// Fallback to current request content type
	if ui.CurrentRequest != nil {
		return ui.CurrentRequest.ContentType
	}
	return "" // Unknown
}

// getNextValidSubchild returns the next valid subchild index based on current content type
func getNextValidSubchild(currentSubchild int, ui *UIOrchestrator) int {
	contentType := getCurrentContentType(ui)
	if contentType == "" {
		// Unknown content type, use default behavior
		if currentSubchild < 3 {
			return currentSubchild + 1
		}
		return 3
	}

	// For BodyTab (container 4, child 0)
	// Determine which subchildren are valid for current content type
	switch contentType {
	case "JSON":
		// Valid subchildren: 0 (ContentTypeSelector), 1 (JSONEditor)
		if currentSubchild < 1 {
			return currentSubchild + 1
		}
		return 1 // Already at max
	case "Multipart":
		// Valid subchildren: 0 (ContentTypeSelector), 2 (MultipartFields)
		if currentSubchild == 0 {
			return 2 // Skip 1 (JSONEditor)
		}
		return 2 // Already at max
	case "No Body":
		// Valid subchildren: 0 (ContentTypeSelector), 3 (NoBody)
		if currentSubchild == 0 {
			return 3 // Skip 1 (JSONEditor) and 2 (MultipartFields)
		}
		return 3 // Already at max
	default:
		// Default: all subchildren are valid
		if currentSubchild < 3 {
			return currentSubchild + 1
		}
		return 3
	}
}

// getPrevValidSubchild returns the previous valid subchild index based on current content type
func getPrevValidSubchild(currentSubchild int, ui *UIOrchestrator) int {
	contentType := getCurrentContentType(ui)
	if contentType == "" {
		// Unknown content type, use default behavior
		if currentSubchild > 0 {
			return currentSubchild - 1
		}
		return 0
	}

	// For BodyTab (container 4, child 0)
	// Determine which subchildren are valid for current content type
	switch contentType {
	case "JSON":
		// Valid subchildren: 0 (ContentTypeSelector), 1 (JSONEditor)
		if currentSubchild == 1 {
			return 0
		}
		return 0 // Already at min
	case "Multipart":
		// Valid subchildren: 0 (ContentTypeSelector), 2 (MultipartFields)
		if currentSubchild == 2 {
			return 0
		}
		return 0 // Already at min
	case "No Body":
		// Valid subchildren: 0 (ContentTypeSelector), 3 (NoBody)
		if currentSubchild == 3 {
			return 0
		}
		return 0 // Already at min
	default:
		// Default: all subchildren are valid
		if currentSubchild > 0 {
			return currentSubchild - 1
		}
		return 0
	}
}

// syncExperimentalChildWithCurrentTab syncs experimental child with current tab index when in tab headers mode
func syncExperimentalChildWithCurrentTab(ui *UIOrchestrator) {
	if ui.NavCurrentContainer == 4 && ui.NavRequestInTabHeaders {
		// Request panel tab headers mode
		ui.NavCurrentChild = ui.CurrentTabIndex
	} else if ui.NavCurrentContainer == 5 && ui.NavResponseInTabHeaders {
		// Response panel tab headers mode
		ui.NavCurrentChild = ui.CurrentResponseTabIndex
	}
}

// syncMainCycleWithExperimental syncs MainCycle.current with NavCurrentContainer
func syncMainCycleWithExperimental(ui *UIOrchestrator) {
	// Always sync as this is now the main navigation system

	// Map experimental container to MainCycle panel index
	switch ui.NavCurrentContainer {
	case 0:
		ui.MainCycle.current = ui.PanelIndices.Workspace
	case 1:
		ui.MainCycle.current = ui.PanelIndices.Environment
	case 2:
		ui.MainCycle.current = ui.PanelIndices.Collections
	case 3:
		ui.MainCycle.current = ui.PanelIndices.URLBar
	case 4:
		ui.MainCycle.current = ui.PanelIndices.Request
	case 5:
		ui.MainCycle.current = ui.PanelIndices.Response
	}

	// Also update CurrentFocus for compatibility with old code
	ui.CurrentFocus = ui.MainCycle.current
}

// setFocusForCoordinates sets focus to the appropriate UI element based on current coordinates
func setFocusForCoordinates(ui *UIOrchestrator) {
	// Sync MainCycle.current with experimental container
	syncMainCycleWithExperimental(ui)

	// Sync experimental child with current tab index when in tab headers mode
	syncExperimentalChildWithCurrentTab(ui)

	// Bounds checking
	if ui.NavCurrentContainer < 0 || ui.NavCurrentContainer > 5 {
		ui.NavCurrentContainer = 0
	}
	maxChild := getMaxChildForContainer(ui.NavCurrentContainer)
	if ui.NavCurrentChild < 0 || ui.NavCurrentChild > maxChild {
		ui.NavCurrentChild = 0
	}

	switch ui.NavCurrentContainer {
	case 0: // Workspace panel
		switch ui.NavCurrentChild {
		case 0: // WorkspaceSelector
			ui.App.SetFocus(ui.WorkspaceSelector)
		case 1: // WorkspaceMenu (Config button)
			ui.App.SetFocus(ui.WorkspaceConfigButton)
		default:
			// Fallback to first child
			ui.NavCurrentChild = 0
			ui.App.SetFocus(ui.WorkspaceSelector)
		}

	case 1: // Environment panel
		switch ui.NavCurrentChild {
		case 0: // EnvironmentSelector
			ui.App.SetFocus(ui.EnvDropdown)
		case 1: // EnvironmentMenu (Config button)
			ui.App.SetFocus(ui.EnvConfigButton)
		default:
			// Fallback to first child
			ui.NavCurrentChild = 0
			ui.App.SetFocus(ui.EnvDropdown)
		}

	case 2: // Collections panel
		// Only child 0: CollectionsTreeView
		ui.App.SetFocus(ui.CollectionsTreeView)

	case 3: // URLBar panel
		switch ui.NavCurrentChild {
		case 0: // MethodDropdown
			ui.App.SetFocus(ui.MethodDropdown)
		case 1: // URLInput
			ui.App.SetFocus(ui.URLInput)
		case 2: // SendButton
			ui.App.SetFocus(ui.SendButton)
		case 3: // CurlButton
			ui.App.SetFocus(ui.CurlButton)
		default:
			// Fallback to first child
			ui.NavCurrentChild = 0
			ui.App.SetFocus(ui.MethodDropdown)
		}

	case 4: // Request panel
		if ui.NavRequestInTabHeaders {
			// Sync experimental child with current tab index
			ui.NavCurrentChild = ui.CurrentTabIndex
			// Focus the tab header
			ui.App.SetFocus(ui.TabHeader)
		} else {
			// In tab content mode
			switch ui.NavCurrentChild {
			case 0: // BodyTab (has subchildren)
				// Handle BodyTab subchildren with bounds checking
				maxSubchild := getMaxSubchildForChild(4, 0, ui)
				if ui.NavCurrentSubchild < 0 || ui.NavCurrentSubchild > maxSubchild {
					ui.NavCurrentSubchild = 0
				}

				switch ui.NavCurrentSubchild {
				case 0: // ContentTypeSelector
					ui.App.SetFocus(ui.ContentTypeDropdown)
				case 1: // JSONEditor
					if ui.BodyEditMode {
						ui.App.SetFocus(ui.BodyEditPanel)
					} else {
						ui.App.SetFocus(ui.BodyViewPanel)
					}
				case 2: // MultipartFields
					// Handle multipart elements navigation
					if ui.MultipartFieldsTab != nil {
						// Get the button row (first child of MultipartFieldsTab)
						buttonRow := ui.MultipartFieldsTab.GetItem(0)
						if buttonRow != nil {
							buttonRowFlex, ok := buttonRow.(*tview.Flex)
							if ok && buttonRowFlex != nil {
								// Get max multipart element (buttons + field rows)
								maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)

								// Add field rows if available
								if currentMultipartFieldRows != nil {
									maxMultipartElement = 2 + len(currentMultipartFieldRows) // 0: Add, 1: Delete All, 2+: field rows
								}

								// Bounds checking for multipart element
								if ui.NavCurrentMultipartElement < 0 || ui.NavCurrentMultipartElement >= maxMultipartElement {
									ui.NavCurrentMultipartElement = 0
									ui.NavCurrentFieldRowElement = 0
								}

								switch ui.NavCurrentMultipartElement {
								case 0: // Add Field button
									if buttonRowFlex.GetItemCount() > 0 {
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(ui.MultipartFieldsTab)
										}
									} else {
										ui.App.SetFocus(ui.MultipartFieldsTab)
									}
								case 1: // Delete All button
									if buttonRowFlex.GetItemCount() > 1 {
										deleteAllButton := buttonRowFlex.GetItem(1)
										if deleteAllButton != nil {
											ui.App.SetFocus(deleteAllButton)
										} else {
											ui.App.SetFocus(ui.MultipartFieldsTab)
										}
									} else {
										ui.App.SetFocus(ui.MultipartFieldsTab)
									}
								default: // Field rows (starting from index 2)
									fieldRowIndex := ui.NavCurrentMultipartElement - 2
									if fieldRowIndex >= 0 && fieldRowIndex < len(currentMultipartFieldRows) {
										fieldRow := currentMultipartFieldRows[fieldRowIndex]
										if fieldRow != nil {
											// Get max field row element for this row
											maxFieldRowElement := getMaxFieldRowElement(fieldRow)

											// Bounds checking for field row element
											if ui.NavCurrentFieldRowElement < 0 || ui.NavCurrentFieldRowElement > maxFieldRowElement {
												ui.NavCurrentFieldRowElement = 0
											}

											// Focus the appropriate element in the row
											switch ui.NavCurrentFieldRowElement {
											case 0: // Name input
												if fieldRow.NameInput != nil {
													ui.App.SetFocus(fieldRow.NameInput)
												} else {
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											case 1: // Type dropdown
												if fieldRow.TypeDropdown != nil {
													ui.App.SetFocus(fieldRow.TypeDropdown)
												} else {
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											case 2: // Value input
												if fieldRow.ValueInput != nil {
													ui.App.SetFocus(fieldRow.ValueInput)
												} else {
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											case 3: // Browse button (for file type) or empty space (for text type)
												selectedType, _ := fieldRow.TypeDropdown.GetCurrentOption()
												if selectedType == 2 { // "file" type
													if fieldRow.FilePickerButton != nil {
														ui.App.SetFocus(fieldRow.FilePickerButton)
													} else {
														ui.App.SetFocus(ui.MultipartFieldsTab)
													}
												} else {
													// For non-file type, index 3 is not focusable.
													// Navigation logic should have skipped this, but as fallback focus the row container or the whole tab
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											case 4: // Checkbox (enable/disable)
												if fieldRow.Checkbox != nil {
													ui.App.SetFocus(fieldRow.Checkbox)
												} else {
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											case 5: // X button (delete)
												if fieldRow.DeleteButton != nil {
													ui.App.SetFocus(fieldRow.DeleteButton)
												} else {
													ui.App.SetFocus(ui.MultipartFieldsTab)
												}
											default:
												ui.App.SetFocus(ui.MultipartFieldsTab)
											}
										} else {
											ui.App.SetFocus(ui.MultipartFieldsTab)
										}
									} else {
										ui.App.SetFocus(ui.MultipartFieldsTab)
									}
								}
							} else {
								ui.App.SetFocus(ui.MultipartFieldsTab)
							}
						} else {
							ui.App.SetFocus(ui.MultipartFieldsTab)
						}
					} else {
						ui.App.SetFocus(ui.BodyContainer)
					}
				case 3: // NoBody
					ui.App.SetFocus(ui.BodyViewPanel)
				default:
					ui.App.SetFocus(ui.BodyContainer)
				}

			case 1: // AuthTab
				// Focus auth tab content (implementation depends on auth UI)
				// For now, focus the request data tabs container
				ui.App.SetFocus(ui.RequestDataTabs)

			case 2: // QueryTab
				// Focus query tab content
				// Use NavCurrentQueryParamRowElement to determine which element to focus on:
				// 0: Add Param button
				// 1: Delete All button
				// 2+: Query param rows
				if currentQueryParamsTab != nil {
					// Get the button row (first child of query params container)
					buttonRow := currentQueryParamsTab.GetItem(0)
					if buttonRow != nil {
						buttonRowFlex, ok := buttonRow.(*tview.Flex)
						if ok && buttonRowFlex != nil && buttonRowFlex.GetItemCount() > 1 {
							// Determine which element to focus based on NavCurrentQueryParamRowElement
							if ui.NavCurrentQueryParamRowElement < 2 {
								// We're on a button (Add Param or Delete All)
								switch ui.NavCurrentQueryParamRowElement {
								case 0: // Add Param button
									addButton := buttonRowFlex.GetItem(0)
									if addButton != nil {
										ui.App.SetFocus(addButton)
									} else {
										ui.App.SetFocus(currentQueryParamsTab)
									}
								case 1: // Delete All button
									deleteAllButton := buttonRowFlex.GetItem(2)
									if deleteAllButton != nil {
										ui.App.SetFocus(deleteAllButton)
									} else {
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(currentQueryParamsTab)
										}
									}
								}
							} else {
								// We're on a query param row (NavCurrentQueryParamRowElement >= 2)
								queryParamRowIndex := ui.NavCurrentQueryParamRowElement - 2
								if queryParamRowIndex >= 0 && queryParamRowIndex < len(currentQueryRows) {
									queryParamRow := currentQueryRows[queryParamRowIndex]
									if queryParamRow != nil {
										switch ui.NavCurrentQueryParamElement {
										case 0: // Key input
											if queryParamRow.KeyInput != nil {
												ui.App.SetFocus(queryParamRow.KeyInput)
											} else {
												if queryParamRow.ValueInput != nil {
													ui.App.SetFocus(queryParamRow.ValueInput)
												} else {
													ui.App.SetFocus(currentQueryParamsTab)
												}
											}
										case 1: // Value input
											if queryParamRow.ValueInput != nil {
												ui.App.SetFocus(queryParamRow.ValueInput)
											} else {
												if queryParamRow.Checkbox != nil {
													ui.App.SetFocus(queryParamRow.Checkbox)
												} else {
													ui.App.SetFocus(currentQueryParamsTab)
												}
											}
										case 2: // Checkbox
											if queryParamRow.Checkbox != nil {
												ui.App.SetFocus(queryParamRow.Checkbox)
											} else {
												if queryParamRow.DeleteButton != nil {
													ui.App.SetFocus(queryParamRow.DeleteButton)
												} else {
													ui.App.SetFocus(currentQueryParamsTab)
												}
											}
										case 3: // Delete button
											if queryParamRow.DeleteButton != nil {
												ui.App.SetFocus(queryParamRow.DeleteButton)
											} else {
												if queryParamRow.KeyInput != nil {
													ui.App.SetFocus(queryParamRow.KeyInput)
												} else {
													ui.App.SetFocus(currentQueryParamsTab)
												}
											}
										default:
											ui.NavCurrentQueryParamElement = 0
											if queryParamRow.KeyInput != nil {
												ui.App.SetFocus(queryParamRow.KeyInput)
											} else {
												ui.App.SetFocus(currentQueryParamsTab)
											}
										}
									} else {
										ui.App.SetFocus(currentQueryParamsTab)
									}
								} else {
									ui.App.SetFocus(currentQueryParamsTab)
								}
							}
						} else {
							ui.App.SetFocus(currentQueryParamsTab)
						}
					} else {
						ui.App.SetFocus(currentQueryParamsTab)
					}
				} else {
					ui.App.SetFocus(ui.RequestDataTabs)
				}

			case 3: // HeadersTab
				// Focus headers tab content
				// Use NavCurrentHeaderRowElement to determine which element to focus on:
				// 0: Add Header button
				// 1: Delete All button
				// 2+: Header rows
				if currentHeadersTab != nil {
					// Get the button row (first child of headers container)
					// Headers container structure: [buttonRow, spacer, headersList]
					buttonRow := currentHeadersTab.GetItem(0)
					if buttonRow != nil {
						buttonRowFlex, ok := buttonRow.(*tview.Flex)
						if ok && buttonRowFlex != nil && buttonRowFlex.GetItemCount() > 2 {
							// Determine which element to focus based on NavCurrentHeaderRowElement
							if ui.NavCurrentHeaderRowElement < 2 {
								// We're on a button (Add Header or Delete All)
								switch ui.NavCurrentHeaderRowElement {
								case 0: // Add Header button
									addButton := buttonRowFlex.GetItem(0)
									if addButton != nil {
										ui.App.SetFocus(addButton)
									} else {
										ui.App.SetFocus(currentHeadersTab)
									}
								case 1: // Delete All button
									deleteAllButton := buttonRowFlex.GetItem(2)
									if deleteAllButton != nil {
										ui.App.SetFocus(deleteAllButton)
									} else {
										// Fallback to Add Header button
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(currentHeadersTab)
										}
									}
								}
							} else {
								// We're on a header row (NavCurrentHeaderRowElement >= 2)
								headerRowIndex := ui.NavCurrentHeaderRowElement - 2
								if headerRowIndex >= 0 && headerRowIndex < len(currentHeaderRows) {
									headerRow := currentHeaderRows[headerRowIndex]
									if headerRow != nil {
										// Determine which element within the header row to focus
										switch ui.NavCurrentHeaderElement {
										case 0: // Key input
											if headerRow.KeyInput != nil {
												ui.App.SetFocus(headerRow.KeyInput)
											} else {
												if headerRow.ValueInput != nil {
													ui.App.SetFocus(headerRow.ValueInput)
												} else {
													ui.App.SetFocus(currentHeadersTab)
												}
											}
										case 1: // Value input
											if headerRow.ValueInput != nil {
												ui.App.SetFocus(headerRow.ValueInput)
											} else {
												if headerRow.Checkbox != nil {
													ui.App.SetFocus(headerRow.Checkbox)
												} else {
													ui.App.SetFocus(currentHeadersTab)
												}
											}
										case 2: // Checkbox
											if headerRow.Checkbox != nil {
												ui.App.SetFocus(headerRow.Checkbox)
											} else {
												if headerRow.DeleteButton != nil {
													ui.App.SetFocus(headerRow.DeleteButton)
												} else {
													ui.App.SetFocus(currentHeadersTab)
												}
											}
										case 3: // Delete button
											if headerRow.DeleteButton != nil {
												ui.App.SetFocus(headerRow.DeleteButton)
											} else {
												if headerRow.KeyInput != nil {
													ui.App.SetFocus(headerRow.KeyInput)
												} else {
													ui.App.SetFocus(currentHeadersTab)
												}
											}
										default:
											ui.NavCurrentHeaderElement = 0
											if headerRow.KeyInput != nil {
												ui.App.SetFocus(headerRow.KeyInput)
											} else {
												ui.App.SetFocus(currentHeadersTab)
											}
										}
									} else {
										// Header row is nil, focus on Add Header button
										ui.NavCurrentHeaderRowElement = 0
										ui.NavCurrentHeaderElement = 0
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(currentHeadersTab)
										}
									}
								} else {
									// Invalid header row index, focus on Add Header button
									ui.NavCurrentHeaderRowElement = 0
									ui.NavCurrentHeaderElement = 0
									addButton := buttonRowFlex.GetItem(0)
									if addButton != nil {
										ui.App.SetFocus(addButton)
									} else {
										ui.App.SetFocus(currentHeadersTab)
									}
								}
							}
						} else {
							ui.App.SetFocus(currentHeadersTab)
						}
					} else {
						ui.App.SetFocus(currentHeadersTab)
					}
				} else {
					// Fallback to focusing on the tab container
					ui.App.SetFocus(ui.RequestDataTabs)
				}
			case 4: // CookiesTab
				// Focus cookies tab content
				// Use NavCurrentCookieRowElement to determine which element to focus on:
				// 0: Add Cookie button
				// 1: Delete All button
				// 2+: Cookie rows
				// Note: Cookies button row layout is [addButton, nil-spacer, clearAllButton, nil-spacer]
				//       so Delete All is at index 2, not 1 like Headers/Query.
				if currentCookiesTab != nil {
					// Get the button row (first child of cookies container)
					buttonRow := currentCookiesTab.GetItem(0)
					if buttonRow != nil {
						buttonRowFlex, ok := buttonRow.(*tview.Flex)
						if ok && buttonRowFlex != nil && buttonRowFlex.GetItemCount() > 2 {
							// Determine which element to focus based on NavCurrentCookieRowElement
							if ui.NavCurrentCookieRowElement < 2 {
								// We're on a button (Add Cookie or Delete All)
								switch ui.NavCurrentCookieRowElement {
								case 0: // Add Cookie button
									addButton := buttonRowFlex.GetItem(0)
									if addButton != nil {
										ui.App.SetFocus(addButton)
									} else {
										ui.App.SetFocus(currentCookiesTab)
									}
								case 1: // Delete All button
									deleteAllButton := buttonRowFlex.GetItem(2)
									if deleteAllButton != nil {
										ui.App.SetFocus(deleteAllButton)
									} else {
										// Fallback to Add Cookie button
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(currentCookiesTab)
										}
									}
								}
							} else {
								// We're on a cookie row (NavCurrentCookieRowElement >= 2)
								cookieRowIndex := ui.NavCurrentCookieRowElement - 2
								if cookieRowIndex >= 0 && cookieRowIndex < len(currentCookieRows) {
									cookieRow := currentCookieRows[cookieRowIndex]
									if cookieRow != nil {
										// Determine which element within the cookie row to focus
										switch ui.NavCurrentCookieElement {
										case 0: // Domain input
											if cookieRow.DomainInput != nil {
												ui.App.SetFocus(cookieRow.DomainInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 1: // Name input
											if cookieRow.NameInput != nil {
												ui.App.SetFocus(cookieRow.NameInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 2: // Value input
											if cookieRow.ValueInput != nil {
												ui.App.SetFocus(cookieRow.ValueInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 3: // Path input
											if cookieRow.PathInput != nil {
												ui.App.SetFocus(cookieRow.PathInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 4: // Secure input
											if cookieRow.SecureInput != nil {
												ui.App.SetFocus(cookieRow.SecureInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 5: // HttpOnly input
											if cookieRow.HttpOnlyInput != nil {
												ui.App.SetFocus(cookieRow.HttpOnlyInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 6: // Checkbox
											if cookieRow.Checkbox != nil {
												ui.App.SetFocus(cookieRow.Checkbox)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										case 7: // Delete button
											if cookieRow.DeleteButton != nil {
												ui.App.SetFocus(cookieRow.DeleteButton)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										default:
											ui.NavCurrentCookieElement = 0
											if cookieRow.DomainInput != nil {
												ui.App.SetFocus(cookieRow.DomainInput)
											} else {
												ui.App.SetFocus(currentCookiesTab)
											}
										}
									} else {
										// Cookie row is nil, focus on Add Cookie button
										ui.NavCurrentCookieRowElement = 0
										ui.NavCurrentCookieElement = 0
										addButton := buttonRowFlex.GetItem(0)
										if addButton != nil {
											ui.App.SetFocus(addButton)
										} else {
											ui.App.SetFocus(currentCookiesTab)
										}
									}
								} else {
									// Invalid cookie row index, focus on Add Cookie button
									ui.NavCurrentCookieRowElement = 0
									ui.NavCurrentCookieElement = 0
									addButton := buttonRowFlex.GetItem(0)
									if addButton != nil {
										ui.App.SetFocus(addButton)
									} else {
										ui.App.SetFocus(currentCookiesTab)
									}
								}
							}
						} else {
							ui.App.SetFocus(currentCookiesTab)
						}
					} else {
						ui.App.SetFocus(currentCookiesTab)
					}
				} else {
					// Fallback to focusing on the tab container
					ui.App.SetFocus(ui.RequestDataTabs)
				}
			default:
				// Fallback to first child
				ui.NavCurrentChild = 0
				ui.NavCurrentSubchild = 0
				ui.App.SetFocus(ui.ContentTypeDropdown)
			}
		}

	case 5: // Response panel
		if ui.NavResponseInTabHeaders {
			// Sync experimental child with current response tab index
			ui.NavCurrentChild = ui.CurrentResponseTabIndex
			// Focus the response tab header
			ui.App.SetFocus(ui.ResponseTabHeader)
		} else {
			// In tab content mode
			switch ui.NavCurrentChild {
			case 0: // PreviewTab
				ui.App.SetFocus(ui.ResponsePreviewPanel)
			case 1: // HeadersTab
				ui.App.SetFocus(ui.ResponseHeadersPanel)
			case 2: // CookiesTab
				ui.App.SetFocus(ui.ResponseCookiesPanel)
			case 3: // TimelineTab
				ui.App.SetFocus(ui.ResponseTimelinePanel)
			default:
				// Fallback to first child
				ui.NavCurrentChild = 0
				ui.App.SetFocus(ui.ResponsePreviewPanel)
			}
		}

	default:
		// Should never happen due to bounds checking above, but just in case
		ui.NavCurrentContainer = 0
		ui.NavCurrentChild = 0
		ui.NavCurrentSubchild = 0
		ui.App.SetFocus(ui.WorkspaceSelector)
	}

	// Update footer whenever focus coordinates change
	ui.UpdateFooter()
}

func handleTabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Check if a modal is open - if so, let the modal handle Tab
	if name, _ := ui.Pages.GetFrontPage(); name != "main" {
		return event
	}

	// Ensure Request panel starts at tab headers when entered from elsewhere
	if ui.NavCurrentContainer != 4 {
		ui.NavRequestInTabHeaders = true
	}

	// Store previous container before updating
	previousContainer := ui.NavCurrentContainer

	// Special handling for Request panel (container 4)
	if ui.NavCurrentContainer == 4 {
		// Request panel navigation logic (New Active Tab Logic)

		// 1. Sync NavCurrentChild with the actual active tab
		// This ensures we are navigating within the visible tab
		ui.NavCurrentChild = ui.CurrentTabIndex

		if ui.NavRequestInTabHeaders {
			// We're in tab headers mode
			// Tab should enter the current tab's content
			ui.NavRequestInTabHeaders = false
			// Reset subchild for tab content
			ui.NavCurrentSubchild = 0
			// Reset complex counters
			ui.NavCurrentMultipartElement = 0
			ui.NavCurrentFieldRowElement = 0
			ui.NavCurrentHeaderRowElement = 0
			ui.NavCurrentHeaderElement = 0
			ui.NavCurrentCookieRowElement = 0
			ui.NavCurrentCookieElement = 0
		} else {
			// We're in tab content mode
			// Navigation depends on which tab is active
			switch ui.NavCurrentChild {
			case 0: // Body Tab
				// Check if we're in MultipartFields and need to navigate within multipart elements
				if ui.NavCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
					// We're in a field row (multipart element >= 2)
					if ui.NavCurrentMultipartElement >= 2 {
						fieldRowIndex := ui.NavCurrentMultipartElement - 2

						fieldRowValid := false
						if fieldRowIndex >= 0 && fieldRowIndex < len(currentMultipartFieldRows) {
							if currentMultipartFieldRows[fieldRowIndex] != nil {
								fieldRowValid = true
							}
						}

						maxFieldRowElement := 0
						if fieldRowValid {
							maxFieldRowElement = getMaxFieldRowElement(currentMultipartFieldRows[fieldRowIndex])
						}

						if fieldRowValid && ui.NavCurrentFieldRowElement < maxFieldRowElement {
							ui.NavCurrentFieldRowElement++

							// Skip Browse button (3) if not file type
							if ui.NavCurrentFieldRowElement == 3 {
								fieldRow := currentMultipartFieldRows[fieldRowIndex]
								selectedType, _ := fieldRow.TypeDropdown.GetCurrentOption()
								if selectedType != 2 { // Not file
									// Skip to Checkbox (4). Checkbox is always focusable, so this
									// simply jumps over the empty filler slot where the Browse
									// button would normally live for "file" type rows.
									ui.NavCurrentFieldRowElement = 4
								}
							}
						} else {
							// Next multipart element
							ui.NavCurrentFieldRowElement = 0
							ui.NavCurrentMultipartElement++

							// Check exit condition
							maxMultipartElement := 2
							if currentMultipartFieldRows != nil {
								maxMultipartElement = 2 + len(currentMultipartFieldRows)
							}

							if ui.NavCurrentMultipartElement >= maxMultipartElement {
								// Exit to Response Panel
								ui.NavCurrentContainer = 5
								ui.NavResponseInTabHeaders = false
								ui.NavRequestInTabHeaders = true // Reset for next entry
							}
						}
					} else {
						// We're at a button
						// Move to next multipart element
						ui.NavCurrentMultipartElement++
						ui.NavCurrentFieldRowElement = 0

						// Check exit
						maxMultipartElement := 2
						if currentMultipartFieldRows != nil {
							maxMultipartElement = 2 + len(currentMultipartFieldRows)
						}
						if ui.NavCurrentMultipartElement >= maxMultipartElement {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				} else {
					// Standard Body Tab Navigation
					maxSubchild := getMaxSubchildForChild(4, 0, ui)

					if ui.NavCurrentSubchild < maxSubchild {
						ui.NavCurrentSubchild = getNextValidSubchild(ui.NavCurrentSubchild, ui)
						// Reset multipart/header state just in case
						ui.NavCurrentMultipartElement = 0
						ui.NavCurrentFieldRowElement = 0
					} else {
						// End of Tab -> Go to Response Panel
						ui.NavCurrentContainer = 5
						ui.NavResponseInTabHeaders = false
						ui.NavRequestInTabHeaders = true // Reset for next entry
					}
				}

			case 3: // Headers Tab
				// ... Headers internal navigation ...
				// 0: Add, 1: Delete All, 2+: Rows
				// Check if we're in a header row
				if ui.NavCurrentHeaderRowElement >= 2 {
					headerRowIndex := ui.NavCurrentHeaderRowElement - 2
					headerRowValid := false
					if headerRowIndex >= 0 && headerRowIndex < len(currentHeaderRows) {
						if currentHeaderRows[headerRowIndex] != nil {
							headerRowValid = true
						}
					}

					maxHeaderElement := 0
					if headerRowValid {
						maxHeaderElement = getMaxHeaderRowElement(currentHeaderRows[headerRowIndex])
					}

					if headerRowValid && ui.NavCurrentHeaderElement < maxHeaderElement {
						ui.NavCurrentHeaderElement++
					} else {
						// Next row
						ui.NavCurrentHeaderElement = 0
						ui.NavCurrentHeaderRowElement++

						// Check exit
						maxHeaderRowElement := 2
						if currentHeaderRows != nil {
							maxHeaderRowElement = 2 + len(currentHeaderRows)
						}
						if ui.NavCurrentHeaderRowElement >= maxHeaderRowElement {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				} else {
					// Buttons
					ui.NavCurrentHeaderRowElement++ // 0 -> 1 or 1 -> 2
					ui.NavCurrentHeaderElement = 0

					// Check exit (if no headers)
					if ui.NavCurrentHeaderRowElement == 2 {
						if currentHeaderRows == nil || len(currentHeaderRows) == 0 {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				}

			case 2: // Query Tab
				// Query params internal navigation ...
				// 0: Add, 1: Delete All, 2+: Rows
				// Check if we're in a query param row
				if ui.NavCurrentQueryParamRowElement >= 2 {
					queryParamRowIndex := ui.NavCurrentQueryParamRowElement - 2
					queryParamRowValid := false
					if queryParamRowIndex >= 0 && queryParamRowIndex < len(currentQueryRows) {
						if currentQueryRows[queryParamRowIndex] != nil {
							queryParamRowValid = true
						}
					}

					maxQueryParamElement := 0
					if queryParamRowValid {
						maxQueryParamElement = getMaxQueryParamRowElement(currentQueryRows[queryParamRowIndex])
					}

					if queryParamRowValid && ui.NavCurrentQueryParamElement < maxQueryParamElement {
						ui.NavCurrentQueryParamElement++
					} else {
						// Next row
						ui.NavCurrentQueryParamElement = 0
						ui.NavCurrentQueryParamRowElement++

						// Check exit
						maxQueryParamRowElement := 2
						if currentQueryRows != nil {
							maxQueryParamRowElement = 2 + len(currentQueryRows)
						}
						if ui.NavCurrentQueryParamRowElement >= maxQueryParamRowElement {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				} else {
					// Buttons
					ui.NavCurrentQueryParamRowElement++ // 0 -> 1 or 1 -> 2
					ui.NavCurrentQueryParamElement = 0

					// Check exit (if no query params)
					if ui.NavCurrentQueryParamRowElement == 2 {
						if currentQueryRows == nil || len(currentQueryRows) == 0 {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				}

			case 4: // Cookies Tab
				// Cookies internal navigation
				// 0: Add, 1: Delete All, 2+: Rows
				if ui.NavCurrentCookieRowElement >= 2 {
					cookieRowIndex := ui.NavCurrentCookieRowElement - 2
					cookieRowValid := false
					if cookieRowIndex >= 0 && cookieRowIndex < len(currentCookieRows) {
						if currentCookieRows[cookieRowIndex] != nil {
							cookieRowValid = true
						}
					}

					maxCookieElement := 0
					if cookieRowValid {
						maxCookieElement = getMaxCookieRowElement(currentCookieRows[cookieRowIndex])
					}

					if cookieRowValid && ui.NavCurrentCookieElement < maxCookieElement {
						ui.NavCurrentCookieElement++
					} else {
						// Next row
						ui.NavCurrentCookieElement = 0
						ui.NavCurrentCookieRowElement++

						// Check exit
						maxCookieRowElement := 2
						if currentCookieRows != nil {
							maxCookieRowElement = 2 + len(currentCookieRows)
						}
						if ui.NavCurrentCookieRowElement >= maxCookieRowElement {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				} else {
					// Buttons
					ui.NavCurrentCookieRowElement++ // 0 -> 1 or 1 -> 2
					ui.NavCurrentCookieElement = 0

					// Check exit (if no cookies)
					if ui.NavCurrentCookieRowElement == 2 {
						if currentCookieRows == nil || len(currentCookieRows) == 0 {
							// Exit to Response Panel
							ui.NavCurrentContainer = 5
							ui.NavResponseInTabHeaders = false
							ui.NavRequestInTabHeaders = true // Reset for next entry
						}
					}
				}

			default: // Auth (1) or others
				// Simple navigation: if in content, Tab goes to Response Panel
				ui.NavCurrentContainer = 5
				ui.NavResponseInTabHeaders = false
				ui.NavRequestInTabHeaders = true // Reset for next entry
			}
		}

	} else if ui.NavCurrentContainer == 5 {
		// Response panel navigation logic
		if ui.NavResponseInTabHeaders {
			// If we somehow got into tab headers (e.g. via mouse), Tab enters content
			ui.NavResponseInTabHeaders = false
			ui.NavCurrentSubchild = 0
		} else {
			// Tab should always move to next container (Workspace panel)
			// since we skip tab headers for faster navigation
			ui.NavCurrentContainer = 0 // Workspace panel
			ui.NavCurrentChild = 0     // WorkspaceSelector
			ui.NavCurrentSubchild = 0
			// Reset multipart and field row elements
			ui.NavCurrentMultipartElement = 0
			ui.NavCurrentFieldRowElement = 0
			// Ensure response tab headers is false for next time
			ui.NavResponseInTabHeaders = false
		}
	} else {
		// Normal navigation for other containers
		// Check if current position has subchildren
		if hasSubchildren(ui.NavCurrentContainer, ui.NavCurrentChild, ui) {
			maxChild := getMaxChildForContainer(ui.NavCurrentContainer)

			if ui.NavCurrentChild < maxChild {
				// Move to next child in same container
				ui.NavCurrentChild++
			} else {
				// At last child, move to next container and reset child
				ui.NavCurrentChild = 0
				ui.NavCurrentContainer = (ui.NavCurrentContainer + 1) % 6 // 6 containers total
				// If moving to container 5, ensure we skip headers
				if ui.NavCurrentContainer == 5 {
					ui.NavResponseInTabHeaders = false
				}
			}
			// Reset subchild
			ui.NavCurrentSubchild = 0
		} else {
			// No subchildren at current position (Containers 0, 1, 2, 3)
			maxChild := getMaxChildForContainer(ui.NavCurrentContainer)

			if ui.NavCurrentChild < maxChild {
				// Move to next child in same container
				ui.NavCurrentChild++
			} else {
				// At last child, move to next container and reset child
				ui.NavCurrentChild = 0
				ui.NavCurrentContainer = (ui.NavCurrentContainer + 1) % 6 // 6 containers total
				// If moving to container 5, ensure we skip headers
				if ui.NavCurrentContainer == 5 {
					ui.NavResponseInTabHeaders = false
				}
			}
			// Reset subchild when moving to a position without subchildren
			ui.NavCurrentSubchild = 0
		}
	}

	// Update borders if container changed
	if previousContainer != ui.NavCurrentContainer {
		// Deactivate border of previous container
		if previousContainer < len(ui.MainCycle.panels) {
			ui.SetInactiveBorder(ui.MainCycle.panels[previousContainer])
		}
		// Activate border of current container
		if ui.NavCurrentContainer < len(ui.MainCycle.panels) {
			ui.SetActiveBorder(ui.MainCycle.panels[ui.NavCurrentContainer])
		}
		// Update previous container tracking
		ui.NavPreviousContainer = previousContainer
	}

	// Sync MainCycle.current with experimental container
	syncMainCycleWithExperimental(ui)

	// Sync experimental child with current tab index before getting message
	syncExperimentalChildWithCurrentTab(ui)

	// Set focus to the appropriate UI element based on current coordinates
	setFocusForCoordinates(ui)

	// Return nil to prevent further navigation processing
	return nil
}

// handleBacktabNavigation handles Backtab (Shift+Tab) key navigation
func handleBacktabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Check if a modal is open - if so, let the modal handle Backtab
	if name, _ := ui.Pages.GetFrontPage(); name != "main" {
		return event
	}

	// Ensure Request panel starts at tab headers when entered from elsewhere
	if ui.NavCurrentContainer != 4 {
		ui.NavRequestInTabHeaders = true
	}

	// Store previous container before updating
	previousContainer := ui.NavCurrentContainer

	// Special handling for Request panel (container 4)
	if ui.NavCurrentContainer == 4 {
		// Request panel backtab navigation logic
		// Sync with current tab
		ui.NavCurrentChild = ui.CurrentTabIndex

		if ui.NavRequestInTabHeaders {
			// We're in tab headers mode
			// Backtab should move to previous container (URLBar)
			ui.NavCurrentContainer = 3                      // URLBar panel
			ui.NavCurrentChild = getMaxChildForContainer(3) // Last child of URLBar
			ui.NavRequestInTabHeaders = true                // Reset for Request panel
			ui.NavCurrentSubchild = 0
		} else {
			// We're in tab content mode
			// Backtab should check if we are at the start of content

			// Check Tab Type
			switch ui.NavCurrentChild {
			case 0: // Body
				// ... Body back navigation ...
				if ui.NavCurrentSubchild > 0 {
					// Multipart special case
					if ui.NavCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
						// Check multipart internal
						if ui.NavCurrentMultipartElement > 0 {
							// If in field row
							if ui.NavCurrentMultipartElement >= 2 {
								if ui.NavCurrentFieldRowElement > 0 {
									ui.NavCurrentFieldRowElement--

									// Skip Browse button (3) if not file type
									if ui.NavCurrentFieldRowElement == 3 {
										idx := ui.NavCurrentMultipartElement - 2
										if idx >= 0 && idx < len(currentMultipartFieldRows) {
											fieldRow := currentMultipartFieldRows[idx]
											selectedType, _ := fieldRow.TypeDropdown.GetCurrentOption()
											if selectedType != 2 { // Not file
												ui.NavCurrentFieldRowElement = 2 // Skip to Value input
											}
										}
									}
								} else {
									// Prev element
									ui.NavCurrentMultipartElement--
									// Set to end of prev
									if ui.NavCurrentMultipartElement >= 2 {
										// Is field row
										// Set to max element of this row
										idx := ui.NavCurrentMultipartElement - 2
										if idx >= 0 && idx < len(currentMultipartFieldRows) {
											ui.NavCurrentFieldRowElement = getMaxFieldRowElement(currentMultipartFieldRows[idx])
										}
									} else {
										ui.NavCurrentFieldRowElement = 0
									}
								}
							} else {
								// Buttons
								ui.NavCurrentMultipartElement--
								ui.NavCurrentFieldRowElement = 0
							}
						} else {
							// At start of Multipart (Add Button), go to prev subchild
							ui.NavCurrentSubchild = getPrevValidSubchild(ui.NavCurrentSubchild, ui)
							ui.NavCurrentMultipartElement = 0
							ui.NavCurrentFieldRowElement = 0
						}
					} else {
						// Standard Body
						ui.NavCurrentSubchild = getPrevValidSubchild(ui.NavCurrentSubchild, ui)
					}
				} else {
					// At start of content -> Go to Tab Headers
					ui.NavRequestInTabHeaders = true
				}
			case 2: // Query
				// Query params back navigation
				if ui.NavCurrentQueryParamRowElement > 0 {
					if ui.NavCurrentQueryParamRowElement >= 2 {
						// In row
						if ui.NavCurrentQueryParamElement > 0 {
							ui.NavCurrentQueryParamElement--
						} else {
							ui.NavCurrentQueryParamRowElement--
							// Check if prev is row
							if ui.NavCurrentQueryParamRowElement >= 2 {
								idx := ui.NavCurrentQueryParamRowElement - 2
								if idx >= 0 && idx < len(currentQueryRows) {
									ui.NavCurrentQueryParamElement = getMaxQueryParamRowElement(currentQueryRows[idx])
								}
							} else {
								ui.NavCurrentQueryParamElement = 0
							}
						}
					} else {
						// Buttons
						ui.NavCurrentQueryParamRowElement--
						ui.NavCurrentQueryParamElement = 0
					}
				} else {
					// At start of Query -> Go to Tab Headers
					ui.NavRequestInTabHeaders = true
				}
			case 3: // Headers
				// ... Headers back navigation ...
				if ui.NavCurrentHeaderRowElement > 0 {
					if ui.NavCurrentHeaderRowElement >= 2 {
						// In row
						if ui.NavCurrentHeaderElement > 0 {
							ui.NavCurrentHeaderElement--
						} else {
							ui.NavCurrentHeaderRowElement--
							// Check if prev is row
							if ui.NavCurrentHeaderRowElement >= 2 {
								idx := ui.NavCurrentHeaderRowElement - 2
								if idx >= 0 && idx < len(currentHeaderRows) {
									ui.NavCurrentHeaderElement = getMaxHeaderRowElement(currentHeaderRows[idx])
								}
							} else {
								ui.NavCurrentHeaderElement = 0
							}
						}
					} else {
						// Buttons
						ui.NavCurrentHeaderRowElement--
						ui.NavCurrentHeaderElement = 0
					}
				} else {
					// At start of Headers -> Go to Tab Headers
					ui.NavRequestInTabHeaders = true
				}
			case 4: // Cookies
				// Cookies back navigation
				if ui.NavCurrentCookieRowElement > 0 {
					if ui.NavCurrentCookieRowElement >= 2 {
						// In row
						if ui.NavCurrentCookieElement > 0 {
							ui.NavCurrentCookieElement--
						} else {
							ui.NavCurrentCookieRowElement--
							// Check if prev is row
							if ui.NavCurrentCookieRowElement >= 2 {
								idx := ui.NavCurrentCookieRowElement - 2
								if idx >= 0 && idx < len(currentCookieRows) {
									ui.NavCurrentCookieElement = getMaxCookieRowElement(currentCookieRows[idx])
								}
							} else {
								ui.NavCurrentCookieElement = 0
							}
						}
					} else {
						// Buttons
						ui.NavCurrentCookieRowElement--
						ui.NavCurrentCookieElement = 0
					}
				} else {
					// At start of Cookies -> Go to Tab Headers
					ui.NavRequestInTabHeaders = true
				}
			default: // Auth/Query
				// Go to Tab Headers
				ui.NavRequestInTabHeaders = true
			}
		}
	} else if ui.NavCurrentContainer == 5 {
		// Response panel backtab navigation logic
		// Skip tab headers and go directly to previous container (Request panel)
		// AND TARGET THE ACTIVE TAB's LAST ELEMENT

		ui.NavCurrentContainer = 4              // Request panel
		ui.NavCurrentChild = ui.CurrentTabIndex // Active Tab

		// We want to enter content mode, at the end
		ui.NavRequestInTabHeaders = false

		// Set state to end of tab
		switch ui.NavCurrentChild {
		case 0: // Body
			// Set to last subchild
			ui.NavCurrentSubchild = getMaxSubchildForChild(4, 0, ui) // e.g. 3 (NoBody) or 2 (Multipart)

			// If multipart, set to last element
			if ui.NavCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
				maxMultipartElement := 2
				if currentMultipartFieldRows != nil {
					maxMultipartElement = 2 + len(currentMultipartFieldRows)
				}
				if maxMultipartElement > 0 {
					ui.NavCurrentMultipartElement = maxMultipartElement - 1
				} else {
					ui.NavCurrentMultipartElement = 0
				}

				// If in row
				if ui.NavCurrentMultipartElement >= 2 {
					idx := ui.NavCurrentMultipartElement - 2
					if idx >= 0 && idx < len(currentMultipartFieldRows) {
						ui.NavCurrentFieldRowElement = getMaxFieldRowElement(currentMultipartFieldRows[idx])
					}
				} else {
					ui.NavCurrentFieldRowElement = 0
				}
			}
		case 2: // Query
			// Set to last element
			maxQueryParamRowElement := 2
			if currentQueryRows != nil {
				maxQueryParamRowElement = 2 + len(currentQueryRows)
			}
			if maxQueryParamRowElement > 0 {
				ui.NavCurrentQueryParamRowElement = maxQueryParamRowElement - 1
			} else {
				ui.NavCurrentQueryParamRowElement = 0
			}

			if ui.NavCurrentQueryParamRowElement >= 2 {
				idx := ui.NavCurrentQueryParamRowElement - 2
				if idx >= 0 && idx < len(currentQueryRows) {
					ui.NavCurrentQueryParamElement = getMaxQueryParamRowElement(currentQueryRows[idx])
				}
			} else {
				ui.NavCurrentQueryParamElement = 0
			}
		case 3: // Headers
			// Set to last element
			maxHeaderRowElement := 2
			if currentHeaderRows != nil {
				maxHeaderRowElement = 2 + len(currentHeaderRows)
			}
			if maxHeaderRowElement > 0 {
				ui.NavCurrentHeaderRowElement = maxHeaderRowElement - 1
			} else {
				ui.NavCurrentHeaderRowElement = 0
			}

			if ui.NavCurrentHeaderRowElement >= 2 {
				idx := ui.NavCurrentHeaderRowElement - 2
				if idx >= 0 && idx < len(currentHeaderRows) {
					ui.NavCurrentHeaderElement = getMaxHeaderRowElement(currentHeaderRows[idx])
				}
			} else {
				ui.NavCurrentHeaderElement = 0
			}
		case 4: // Cookies
			// Set to last element
			maxCookieRowElement := 2
			if currentCookieRows != nil {
				maxCookieRowElement = 2 + len(currentCookieRows)
			}
			if maxCookieRowElement > 0 {
				ui.NavCurrentCookieRowElement = maxCookieRowElement - 1
			} else {
				ui.NavCurrentCookieRowElement = 0
			}

			if ui.NavCurrentCookieRowElement >= 2 {
				idx := ui.NavCurrentCookieRowElement - 2
				if idx >= 0 && idx < len(currentCookieRows) {
					ui.NavCurrentCookieElement = getMaxCookieRowElement(currentCookieRows[idx])
				}
			} else {
				ui.NavCurrentCookieElement = 0
			}
		default:
			// Simple tabs, just enter content
			ui.NavCurrentSubchild = 0
		}
		// Ensure response tab headers is false for next time
		ui.NavResponseInTabHeaders = false

	} else {
		// Normal backtab navigation for other containers
		if ui.NavCurrentChild > 0 {
			// Move to previous child in same container
			ui.NavCurrentChild--
			// Reset subchild
			ui.NavCurrentSubchild = 0
		} else {
			// At first child, move to previous container
			prevContainer := (ui.NavCurrentContainer - 1 + 6) % 6
			ui.NavCurrentContainer = prevContainer

			if ui.NavCurrentContainer == 5 {
				// Entering Response from Workspace
				ui.NavCurrentChild = ui.CurrentResponseTabIndex
				ui.NavResponseInTabHeaders = false // Enter content
			} else {
				prevMaxChild := getMaxChildForContainer(prevContainer)
				ui.NavCurrentChild = prevMaxChild
			}
			// Reset subchild
			ui.NavCurrentSubchild = 0
		}
	}

	// Update borders if container changed
	if previousContainer != ui.NavCurrentContainer {
		// Deactivate border of previous container
		if previousContainer < len(ui.MainCycle.panels) {
			ui.SetInactiveBorder(ui.MainCycle.panels[previousContainer])
		}
		// Activate border of current container
		if ui.NavCurrentContainer < len(ui.MainCycle.panels) {
			ui.SetActiveBorder(ui.MainCycle.panels[ui.NavCurrentContainer])
		}
		// Update previous container tracking
		ui.NavPreviousContainer = previousContainer
	}

	// Sync MainCycle.current with experimental container
	syncMainCycleWithExperimental(ui)

	// Sync experimental child with current tab index before getting message
	syncExperimentalChildWithCurrentTab(ui)

	// Set focus to the appropriate UI element based on current coordinates
	setFocusForCoordinates(ui)

	// Return nil to prevent further navigation processing
	return nil
}
