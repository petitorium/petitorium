package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/workspace"
)

func refreshCollectionsTree(ui *UIOrchestrator) {
	ui.RootNode.ClearChildren()

	addWorkspaceToTree(ui.WorkspaceData, ui.RootNode)

	ui.CollectionsTreeView.SetRoot(ui.RootNode)
	ui.CollectionsTreeView.SetCurrentNode(ui.RootNode)
	// If there are children, select the first one instead of the root
	if len(ui.RootNode.GetChildren()) > 0 {
		ui.CollectionsTreeView.SetCurrentNode(ui.RootNode.GetChildren()[0])
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
	// Both number keys and arrow keys should respect this for header input fields
	currentFocusedElement := ui.App.GetFocus()
	if isFocusOnHeaderInputField(currentFocusedElement, headerRows) {
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
			targetTabIndex = (ui.CurrentTabIndex - 1 + 4) % 4
		} else if isResponsePanel {
			targetTabIndex = (ui.CurrentResponseTabIndex - 1 + 4) % 4
		}
	case event.Key() == tcell.KeyRight:
		// Wrap around from right
		if isRequestPanel {
			targetTabIndex = (ui.CurrentTabIndex + 1) % 4
		} else if isResponsePanel {
			targetTabIndex = (ui.CurrentResponseTabIndex + 1) % 4
		}
	default:
		return false
	}

	// Perform the tab switch
	if isRequestPanel {
		tabNames := requestTabInternalNames
		ui.TabPages.SwitchToPage(tabNames[targetTabIndex])
		requestTabs := requestTabDisplayNames
		updateTabHeader(requestTabs, ui.TabHeader, targetTabIndex, ui.Colors)
		ui.CurrentTabIndex = targetTabIndex

		// Update experimental navigation state if enabled
		if ui.ExperimentalNavigationEnabled && ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalRequestInTabHeaders {
			ui.ExperimentalCurrentChild = targetTabIndex
			// Update focus and footer for experimental navigation
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
		// But not if experimental navigation is enabled and we're in tab headers mode
		if !(ui.ExperimentalNavigationEnabled && ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalRequestInTabHeaders) {
			switch targetTabIndex {
			case 0: // Body tab
				if ui.BodyEditMode {
					ui.App.SetFocus(ui.BodyEditPanel)
				} else {
					ui.App.SetFocus(ui.BodyViewPanel)
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

		// Update experimental navigation state if enabled
		if ui.ExperimentalNavigationEnabled && ui.ExperimentalCurrentContainer == 5 && ui.ExperimentalResponseInTabHeaders {
			ui.ExperimentalCurrentChild = targetTabIndex
			// Update focus and footer for experimental navigation
			setFocusForCoordinates(ui)
		}
	}

	// Update footer after tab switch
	ui.UpdateFooter()

	return true
}

// savePluginEnvironmentChanges saves plugin-modified environment variables back to the current environment
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

	// Add or update variables in the current environment
	for key, value := range pluginEnv {
		currentEnv.Variables[key] = value
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
	workspaceNames, err := workspace.ListWorkspaces()
	if err != nil {
		workspaceNames = []string{"Default"}
	}

	switchWorkspace := func(text string) {
		if text == lastWorkspace {
			return
		}
		lastWorkspace = text

		if err := workspace.SwitchWorkspace(text); err != nil {
			// ui.FooterRight.SetText(fmt.Sprintf("Error switching workspace: %v", err))
			return
		}

		newWorkspace, err := workspace.LoadWorkspace()
		if err != nil {
			// ui.FooterRight.SetText(fmt.Sprintf("Error switching workspace: %v", err))
			return
		}

		ui.WorkspaceData = newWorkspace
		ui.EnvironmentsData = &newWorkspace.Environments
		ui.DataManager = NewDataManager(newWorkspace)

		// Update environment dropdown with new workspace's environments
		updateEnvironmentDropdown(ui.EnvDropdown, *ui.EnvironmentsData)

		// Set selected environment based on workspace
		selectedEnv := newWorkspace.SelectedEnvironment
		if selectedEnv == "" {
			ui.EnvDropdown.SetCurrentOption(0) // Default to "Base Environment"
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
				ui.EnvDropdown.SetCurrentOption(0) // Default to "Base Environment"
			}
		}

		refreshCollectionsTree(ui)

		ui.CurrentRequest = nil
		ui.CurrentSelectedNode = nil
		ui.LastSelectedRequestNode = nil

		ui.ProgrammaticallyUpdatingURL = true
		ui.URLInput.SetText("")
		ui.ProgrammaticallyUpdatingURL = false

		ui.SyncBodyContent("")
		setHeadersInUI(ui.Colors, nil, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)

		updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, nil)
	}

	ui.WorkspaceSelector.SetSelectedFunc(func(text string, index int) {
		// ui.FooterRight.SetText(fmt.Sprintf("Text %s, Index: %d, workspace name: %s", text, index, workspaceNames[index]))
		if index >= 0 && index < len(workspaceNames) {
			switchWorkspace(workspaceNames[index])
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
	ui.ResponsePages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		return ui.KeyManager.HandleKeyEvent(ui, event, "response_view")
	})

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
			ui.BodyEditPanel.SetBorderColor(tcell.ColorDefault)
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
					ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
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

	ui.MethodDropdown.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		if ui.ProgrammaticallyUpdatingMethod {
			return
		}

		if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
			index, _ := ui.MethodDropdown.GetCurrentOption()
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			if index >= 0 && index < len(methods) {
				newMethod := methods[index]
				if newMethod != ui.CurrentRequest.Method {
					ui.CurrentRequest.Method = newMethod

					ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)

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
						updateMultipartFieldsFromBody(ui.CurrentRequest.Body, ui.Colors, ui.App, ui.Pages)
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

			// Update the node's reference with the new request data
			ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)

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

			// Update the node's reference with the new request data
			ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)

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
			if reference := ui.LastSelectedRequestNode.GetReference(); reference != nil {
				if req, ok := reference.(workspace.Request); ok {
					// Only remove icon if the new selection is also a request
					if newReference := node.GetReference(); newReference != nil {
						if _, isRequest := newReference.(workspace.Request); isRequest {
							coloredMethod := getColoredMethod(req.Method)
							paddedName := padNameToMinLength(req.Name, 4)
							// Restore to reserved space (icon width + fixed space)
							iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon) // + 1
							spacePadding := strings.Repeat(" ", iconWidth)
							ui.LastSelectedRequestNode.SetText(fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName))
						}
					}
				}
			}
		}

		// Add icon to newly selected node if it's a request
		reference := node.GetReference()
		if req, ok := reference.(workspace.Request); ok {
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

			// Set current request for persistence
			ui.CurrentSelectedNode = node
			ui.LastSelectedRequestNode = node

			// Find the request pointer in collectionsData
			ui.CurrentRequest = ui.DataManager.FindRequestPtr(req)

			// Update node reference to current data
			if ui.CurrentRequest != nil {
				node.SetReference(*ui.CurrentRequest)
			}

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
					updateMultipartFieldsFromBody(ui.CurrentRequest.Body, ui.Colors, ui.App, ui.Pages)
				}

				// Show last response if available
				if len((*ui.CurrentRequest).ResponseHistory) > 0 {
					lastResponse := (*ui.CurrentRequest).ResponseHistory[len((*ui.CurrentRequest).ResponseHistory)-1]
					// Convert workspace.HTTPResponse to cmd.HTTPResponse for formatting
					cmdResp := &HTTPResponse{
						StatusCode: lastResponse.StatusCode,
						Status:     lastResponse.Status,
						Headers:    lastResponse.Headers,
						Cookies:    nil, // No cookies in stored history
						Body:       lastResponse.Body,
						Duration:   lastResponse.Duration,
						Timestamp:  lastResponse.Timestamp,
						BodySize:   len(lastResponse.Body),
					}

					// For historical responses, show when that specific request was made
					updateResponseTabs(cmdResp, &lastResponse.Timestamp, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, nil)
				} else {
					// Clear response if no history
					updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, nil)
				}
			}
		} else if _, ok := reference.(workspace.Collection); ok {
			// Save current request headers before clearing
			if ui.CurrentRequest != nil {
				ui.CurrentRequest.Headers = getHeadersFromUI()
			}

			// Clear current request when a collection is selected
			ui.CurrentRequest = nil
			ui.CurrentSelectedNode = nil

			// Track the last selected request node
			if reference := node.GetReference(); reference != nil {
				if _, ok := reference.(workspace.Request); ok {
					ui.LastSelectedRequestNode = node
				}
			}
		}
	}

	// Set the tree selection handler
	ui.TreeSelectionHandler = handleTreeSelection

	// Set the tree highlight handler for navigation
	ui.TreeHighlightHandler = highlightTreeNode

	// Set up collections tree view changed function (called on current node change)
	ui.CollectionsTreeView.SetChangedFunc(func(node *tview.TreeNode) {
		// Do nothing on current node change - loading only on explicit selection
	})

	// Set up collections tree view selected function
	ui.CollectionsTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		handleTreeSelection(node)
	})

	// Set the initial selected style for the current node
	currentNode := ui.CollectionsTreeView.GetCurrentNode()
	if currentNode != nil {
		currentNode.SetSelectedTextStyle(tcell.StyleDefault.Background(ui.Colors.TreeSelection).Foreground(ui.Colors.Foreground))
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
		var selectedEnvName string
		if index == 0 {
			selectedEnvName = "Base"
		} else if index > 0 && index <= len(*ui.EnvironmentsData) {
			selectedEnvName = (*ui.EnvironmentsData)[index-1].Name
		}

		// Save selected environment to workspace
		ui.WorkspaceData.SelectedEnvironment = selectedEnvName
		if err := workspace.SaveWorkspace(ui.WorkspaceData); err != nil {
			// Handle error silently for now
		}
	}

	ui.EnvDropdown.SetSelectedFunc(func(text string, index int) {
		saveEnvironmentSelection(index)
	})

	ui.EnvDropdown.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		index, _ := ui.EnvDropdown.GetCurrentOption()
		saveEnvironmentSelection(index)
	})

	// Add send button functionality
	ui.SendButton.SetSelectedFunc(func() {
		// Prevent multiple requests
		if ui.RequestInProgress {
			return
		}

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

		// For multipart requests, collect the current UI state
		if contentType == "Multipart" {
			body = collectMultipartFieldsFromUI()
			// If no fields are filled, don't send as multipart
			if body == "" {
				contentType = ""
			} else {
				// Remove any manual Content-Type header to allow automatic multipart setting
				delete(headers, "Content-Type")
				delete(headers, "content-type")
			}
		}

		// Determine collection and request name
		collection := ""
		requestName := ""
		if ui.CurrentRequest != nil {
			requestName = ui.CurrentRequest.Name
			// TODO: find collection name
		}

		// Get current environment variables
		var envVars map[string]string
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		if currentEnvIndex == 0 {
			// Base Environment selected - find and use the "Base" environment
			for _, env := range *ui.EnvironmentsData {
				if env.Name == "Base" {
					envVars = env.GetEffectiveVariables(*ui.EnvironmentsData)
					break
				}
			}
		} else {
			// Specific environment selected
			if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
				env := &(*ui.EnvironmentsData)[currentEnvIndex-1]
				envVars = env.GetEffectiveVariables(*ui.EnvironmentsData)
			}
		}

		// Substitute environment variables in URL, body, and headers
		url = substituteVariables(url, envVars)
		body = substituteVariables(body, envVars)
		headers = substituteVariablesInHeaders(headers, envVars)

		requestData := &plugins.RequestData{
			Method:      method,
			URL:         url,
			Headers:     headers,
			Body:        body,
			Collection:  collection,
			RequestName: requestName,
		}

		context := &plugins.HookContext{
			Request:     requestData,
			Environment: envVars,
			Config:      config.C.Plugins.Config,
		}

		// Ensure config is available for all hooks
		if context.Config == nil {
			context.Config = config.C.Plugins.Config
		}

		if ui.PluginManager != nil {
			ui.PluginManager.ExecuteHooks(plugins.PreSend, context)
		}

		// Update headers from context (plugins may have modified them)
		headers = context.Request.Headers

		// Also update the request data in context to reflect the final headers for logging
		context.Request.Headers = headers

		// Mark request as in progress and update UI
		ui.RequestInProgress = true
		ui.SendButton.SetText("Sending").SetSending(true)

		// Send the request in a goroutine
		go func() {
			resp, err := SendRequest(method, url, body, contentType, headers)

			// Use QueueUpdateDraw to handle the response on the main thread
			ui.App.QueueUpdateDraw(func() {
				defer func() {
					// Reset UI state when done
					ui.RequestInProgress = false
					ui.SendButton.SetText(" Send ").SetSending(false)
				}()

				if err != nil {
					if ui.PluginManager != nil {
						ui.PluginManager.ExecuteHooks(plugins.OnError, context)
					}
					updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, nil)
					return
				}

				context.Response = resp

				// Ensure config is available for PostReceive hook
				if context.Config == nil {
					context.Config = config.C.Plugins.Config
				}

				// Run plugin hooks that might modify data but not UI
				if ui.PluginManager != nil {
					ui.PluginManager.ExecuteHooks(plugins.PostReceive, context)
					ui.PluginManager.ExecuteHooks(plugins.ResponseValidation, context)
					ui.PluginManager.ExecuteHooks(plugins.ResponseTransform, context)
				}

				// Save any plugin-modified environment variables back to the current environment
				if context.Environment != nil && len(context.Environment) > 0 {
					savePluginEnvironmentChanges(context.Environment, currentEnvIndex, ui.EnvironmentsData)
				}

				// Store the response in the current request's history
				if ui.CurrentRequest != nil {
					workspaceResp := workspace.HTTPResponse{
						StatusCode: resp.StatusCode,
						Status:     resp.Status,
						Headers:    resp.Headers,
						Body:       resp.Body,
						Duration:   resp.Duration,
						Timestamp:  resp.Timestamp,
					}
					(*ui.CurrentRequest).ResponseHistory = append((*ui.CurrentRequest).ResponseHistory, workspaceResp)

					// Update the node's reference with the new response history
					if ui.CurrentSelectedNode != nil {
						ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
						saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
					}
				}

				if ui.PluginManager != nil {
					ui.PluginManager.ExecuteHooks(plugins.PostSave, context)
					ui.PluginManager.ExecuteHooks(plugins.PostRequest, context)
					ui.PluginManager.ExecuteHooks(plugins.PreUIUpdate, context)
					ui.PluginManager.ExecuteHooks(plugins.PostUIUpdate, context)
					ui.PluginManager.ExecuteHooks(plugins.PreSave, context)
				}

				// Update the response tabs with the new response
				now := time.Now()
				ui.LastResponse = resp
				updateResponseTabs(resp, &now, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors, ui.CopyResponse)
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

		// Get current environment variables
		var envVars map[string]string
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		if currentEnvIndex == 0 {
			// Base Environment selected - find and use the "Base" environment
			for _, env := range *ui.EnvironmentsData {
				if env.Name == "Base" {
					envVars = env.GetEffectiveVariables(*ui.EnvironmentsData)
					break
				}
			}
		} else {
			// Specific environment selected
			if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
				env := &(*ui.EnvironmentsData)[currentEnvIndex-1]
				envVars = env.GetEffectiveVariables(*ui.EnvironmentsData)
			}
		}

		// Substitute environment variables in URL, body, and headers
		url = substituteVariables(url, envVars)
		body = substituteVariables(body, envVars)
		headers = substituteVariablesInHeaders(headers, envVars)

		// Generate curl command
		curlCommand := generateCurlCommand(method, url, headers, body, contentType)

		// Copy to clipboard
		copyToClipboard(curlCommand)

		// Show a brief notification (could be improved with a proper toast notification)
		ui.FooterRight.SetText("cURL command copied to clipboard")
		go func() {
			time.Sleep(2 * time.Second)
			ui.App.QueueUpdateDraw(func() {
				ui.FooterRight.SetText("Petitorium ")
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
			if currentPage != "main" && currentPage != "envVariables" {
				// On modal, 'q' closes the modal
				ui.Pages.RemovePage(currentPage)
				ui.Pages.SwitchToPage("main")
				ui.App.SetFocus(ui.CollectionsTreeView)
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
		if ui.BodyEditMode && ui.App.GetFocus() == ui.BodyEditPanel {
			return event
		}

		// Allow URLVariableInput to handle its own Enter key events
		if event.Key() == tcell.KeyEnter && ui.App.GetFocus() == ui.URLInput {
			return event
		}

		// Allow URLVariableInput to handle its own 'i' key events
		if event.Rune() == 'i' && ui.App.GetFocus() == ui.URLInput {
			return event
		}

		// Allow dropdown lists to handle their own input when dropdown is open
		if isDropdownOpen(ui.EnvDropdown) || isDropdownOpen(ui.WorkspaceSelector) || isDropdownOpen(ui.MethodDropdown) {
			return event
		}

		// Allow dropdown lists to handle their own input
		if _, ok := ui.App.GetFocus().(*tview.List); ok {
			return event
		}

		// Skip global 'i' keybinding if focused on input fields
		if event.Rune() == 'i' {
			focus := ui.App.GetFocus()
			if _, ok := focus.(*tview.InputField); ok {
				return event
			}
			if _, ok := focus.(*tview.DropDown); ok {
				return event
			}
		}

		// First check if this is a global keybinding
		if result := ui.KeyManager.HandleKeyEvent(ui, event, "global"); result != event {
			return result
		}

		// Check if we're focused on a form input field
		focus := ui.App.GetFocus()

		// If we're in a form input field, don't handle collection shortcuts
		if _, isInput := focus.(*tview.InputField); isInput {
			return event // Let input fields handle their own keys
		}
		if _, isTextArea := focus.(*tview.TextArea); isTextArea {
			return event // Let text areas handle their own keys
		}

		// Also check if we're in a form by looking at the current page
		currentPage, _ := ui.Pages.GetFrontPage()
		if currentPage == "newCollection" || currentPage == "newRequest" || currentPage == "workspaceMenu" || currentPage == "envVariables" || currentPage == "deleteAllHeaders" {
			// We're in a popup form, check if focus is on the form itself
			if event.Rune() == 'n' || event.Rune() == 'r' {
				// Let the form handle these keys
				return event
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
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'n' && !isFormPopupActive {
			form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
			modal := createModal(form, 50, 12, tcell.ColorDefault)
			setFormPopupActive(true)
			ui.Pages.AddPage("newCollection", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}

		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'r' && !isFormPopupActive {
			// New request - check if a collection or request is selected
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				var selectedCollection *workspace.Collection

				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Collection is selected
					selectedCollection = &col
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Request is selected - find its parent collection
					selectedCollection = findParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.Name, req.Method, req.URL)
				}

				if selectedCollection != nil {
					form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
					modal := createModal(form, 60, 14, tcell.ColorDefault)
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

			// Handle 'n' and 'r' keys
			if event.Rune() == 'n' || event.Rune() == 'r' {
				if isFormPopupActive {
					// We're in a popup form, let the key pass through to the form input
					return event
				} else {
					// Normal operation - create new collection/request
					if event.Rune() == 'n' {
						form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
						modal := createModal(form, 50, 12, tcell.ColorDefault)
						setFormPopupActive(true)
						ui.Pages.AddPage("newCollection", modal, true, true)
						ui.App.SetFocus(form)
						return nil
					} else if event.Rune() == 'r' {
						// New request - check if a collection or request is selected
						node := ui.CollectionsTreeView.GetCurrentNode()
						if node != nil {
							var selectedCollection *workspace.Collection

							if col, ok := node.GetReference().(workspace.Collection); ok {
								// Collection is selected
								selectedCollection = &col
							} else if req, ok := node.GetReference().(workspace.Request); ok {
								// Request is selected - find its parent collection
								selectedCollection = findParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.Name, req.Method, req.URL)
							}

							if selectedCollection != nil {
								form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
								modal := createModal(form, 60, 14, tcell.ColorDefault)
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

		// Rename functionality (Shift+R)
		if ui.MainCycle.current == ui.PanelIndices.Collections && event.Rune() == 'R' {
			node := ui.CollectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Rename collection
					form := createRenameCollectionForm(ui.App, ui.Pages, &col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createModal(form, 25, 10, tcell.ColorDefault)
					ui.Pages.AddPage("renameCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Rename request - need to find parent collection
					form := createRenameRequestForm(ui.App, ui.Pages, &req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createModal(form, 47, 10, tcell.ColorDefault)
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
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Move collection
					form := createMoveCollectionForm(ui.App, ui.Pages, &col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createModal(form, 40, 12, tcell.ColorDefault)
					ui.Pages.AddPage("moveCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Move request
					form := createMoveRequestForm(ui.App, ui.Pages, &req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
					modal := createModal(form, 40, 10, tcell.ColorDefault)
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
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Delete collection with confirmation
					form := createDeleteCollectionConfirm(ui.App, ui.Pages, &col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createModal(form, 50, 8, tcell.ColorDefault)
					ui.Pages.AddPage("deleteCollection", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Delete request with confirmation
					form := createDeleteRequestConfirm(ui.App, ui.Pages, &req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
					modal := createModal(form, 50, 8, tcell.ColorDefault)
					ui.Pages.AddPage("deleteRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}

		// F4 to open body in external editor
		if ui.MainCycle.current == ui.PanelIndices.Request && event.Key() == tcell.KeyF4 {
			if ui.CurrentRequest != nil {
				// Suspend TUI to open external editor
				ui.App.Suspend(func() {
					modifiedContent, err := openInExternalEditor(ui.CurrentBodyContent)
					if err != nil {
						fmt.Printf("Error opening external editor: %v\n", err)
						fmt.Println("Press Enter to continue...")
						var dummy string
						fmt.Scanln(&dummy)
						return
					}

					// Update the body with modified content
					ui.SyncBodyContent(modifiedContent)
					if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
						ui.CurrentRequest.Body = modifiedContent
						ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
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
			// Instead of switching to inline editor, open external editor for better paste support
			if ui.CurrentRequest != nil {
				ui.App.Suspend(func() {
					modifiedContent, err := openInExternalEditor(ui.CurrentBodyContent)
					if err != nil {
						// Could show error but for now just continue with current content
						return
					}

					// Update the body with modified content
					ui.SyncBodyContent(modifiedContent)
					if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
						ui.CurrentRequest.Body = modifiedContent
						ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
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
		return 3 // BodyTab (0), AuthTab (1), QueryTab (2), HeadersTab (3)
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
	return container == 4 && child == 0 && !ui.ExperimentalRequestInTabHeaders
}

// getMaxSubchildForChild returns the maximum subchild index for a given container/child
func getMaxSubchildForChild(container, child int, ui *UIOrchestrator) int {
	if container == 4 && child == 0 { // Request BodyTab
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
	}
	return 0 // No subchildren by default
}

// getMaxFieldRowElement returns the maximum field row element index for a field row
func getMaxFieldRowElement(fieldRow *MultipartFieldRow) int {
	if fieldRow == nil {
		return 0
	}

	// With consistent layout, we always have:
	// Name (0), Type (1), Value (2), Browse/Empty (3), X button (4)
	return 4
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
	if ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalRequestInTabHeaders {
		// Request panel tab headers mode
		ui.ExperimentalCurrentChild = ui.CurrentTabIndex
	} else if ui.ExperimentalCurrentContainer == 5 && ui.ExperimentalResponseInTabHeaders {
		// Response panel tab headers mode
		ui.ExperimentalCurrentChild = ui.CurrentResponseTabIndex
	}
}

// syncMainCycleWithExperimental syncs MainCycle.current with ExperimentalCurrentContainer
func syncMainCycleWithExperimental(ui *UIOrchestrator) {
	if !ui.ExperimentalNavigationEnabled {
		return
	}

	// Map experimental container to MainCycle panel index
	switch ui.ExperimentalCurrentContainer {
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
	if ui.ExperimentalCurrentContainer < 0 || ui.ExperimentalCurrentContainer > 5 {
		ui.ExperimentalCurrentContainer = 0
	}
	maxChild := getMaxChildForContainer(ui.ExperimentalCurrentContainer)
	if ui.ExperimentalCurrentChild < 0 || ui.ExperimentalCurrentChild > maxChild {
		ui.ExperimentalCurrentChild = 0
	}

	switch ui.ExperimentalCurrentContainer {
	case 0: // Workspace panel
		switch ui.ExperimentalCurrentChild {
		case 0: // WorkspaceSelector
			ui.App.SetFocus(ui.WorkspaceSelector)
		case 1: // WorkspaceMenu (Config button)
			ui.App.SetFocus(ui.WorkspaceConfigButton)
		default:
			// Fallback to first child
			ui.ExperimentalCurrentChild = 0
			ui.App.SetFocus(ui.WorkspaceSelector)
		}

	case 1: // Environment panel
		switch ui.ExperimentalCurrentChild {
		case 0: // EnvironmentSelector
			ui.App.SetFocus(ui.EnvDropdown)
		case 1: // EnvironmentMenu (Config button)
			ui.App.SetFocus(ui.EnvConfigButton)
		default:
			// Fallback to first child
			ui.ExperimentalCurrentChild = 0
			ui.App.SetFocus(ui.EnvDropdown)
		}

	case 2: // Collections panel
		// Only child 0: CollectionsTreeView
		ui.App.SetFocus(ui.CollectionsTreeView)

	case 3: // URLBar panel
		switch ui.ExperimentalCurrentChild {
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
			ui.ExperimentalCurrentChild = 0
			ui.App.SetFocus(ui.MethodDropdown)
		}

	case 4: // Request panel
		if ui.ExperimentalRequestInTabHeaders {
			// Sync experimental child with current tab index
			ui.ExperimentalCurrentChild = ui.CurrentTabIndex
			// Focus the tab header
			ui.App.SetFocus(ui.TabHeader)
		} else {
			// In tab content mode
			switch ui.ExperimentalCurrentChild {
			case 0: // BodyTab (has subchildren)
				// Handle BodyTab subchildren with bounds checking
				maxSubchild := getMaxSubchildForChild(4, 0, ui)
				if ui.ExperimentalCurrentSubchild < 0 || ui.ExperimentalCurrentSubchild > maxSubchild {
					ui.ExperimentalCurrentSubchild = 0
				}

				switch ui.ExperimentalCurrentSubchild {
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
								if ui.ExperimentalCurrentMultipartElement < 0 || ui.ExperimentalCurrentMultipartElement >= maxMultipartElement {
									ui.ExperimentalCurrentMultipartElement = 0
									ui.ExperimentalCurrentFieldRowElement = 0
								}

								switch ui.ExperimentalCurrentMultipartElement {
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
									fieldRowIndex := ui.ExperimentalCurrentMultipartElement - 2
									if fieldRowIndex >= 0 && fieldRowIndex < len(currentMultipartFieldRows) {
										fieldRow := currentMultipartFieldRows[fieldRowIndex]
										if fieldRow != nil {
											// Get max field row element for this row
											maxFieldRowElement := getMaxFieldRowElement(fieldRow)

											// Bounds checking for field row element
											if ui.ExperimentalCurrentFieldRowElement < 0 || ui.ExperimentalCurrentFieldRowElement > maxFieldRowElement {
												ui.ExperimentalCurrentFieldRowElement = 0
											}

											// Focus the appropriate element in the row
											switch ui.ExperimentalCurrentFieldRowElement {
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
													// For text type, case 3 is empty space (not focusable)
													// Skip to X button at case 4
													ui.ExperimentalCurrentFieldRowElement = 4
													if fieldRow.DeleteButton != nil {
														ui.App.SetFocus(fieldRow.DeleteButton)
													} else {
														ui.App.SetFocus(ui.MultipartFieldsTab)
													}
												}
											case 4: // X button (always at position 4)
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
				ui.App.SetFocus(ui.RequestDataTabs)

			case 3: // HeadersTab
				// Focus headers tab content
				ui.App.SetFocus(ui.RequestDataTabs)
			default:
				// Fallback to first child
				ui.ExperimentalCurrentChild = 0
				ui.ExperimentalCurrentSubchild = 0
				ui.App.SetFocus(ui.ContentTypeDropdown)
			}
		}

	case 5: // Response panel
		if ui.ExperimentalResponseInTabHeaders {
			// Sync experimental child with current response tab index
			ui.ExperimentalCurrentChild = ui.CurrentResponseTabIndex
			// Focus the response tab header
			ui.App.SetFocus(ui.ResponseTabHeader)
		} else {
			// In tab content mode
			switch ui.ExperimentalCurrentChild {
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
				ui.ExperimentalCurrentChild = 0
				ui.App.SetFocus(ui.ResponsePreviewPanel)
			}
		}

	default:
		// Should never happen due to bounds checking above, but just in case
		ui.ExperimentalCurrentContainer = 0
		ui.ExperimentalCurrentChild = 0
		ui.ExperimentalCurrentSubchild = 0
		ui.App.SetFocus(ui.WorkspaceSelector)
	}
}

func handleTabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Experimental navigation system (disabled by default)
	if ui.ExperimentalNavigationEnabled {
		// Check if a modal is open - if so, let the modal handle Tab
		if name, _ := ui.Pages.GetFrontPage(); name != "main" {
			return event
		}
		// Store previous container before updating
		previousContainer := ui.ExperimentalCurrentContainer

		// Special handling for Request panel (container 4)
		if ui.ExperimentalCurrentContainer == 4 {
			// Request panel navigation logic
			if ui.ExperimentalRequestInTabHeaders {
				// We're in tab headers mode
				// Tab should enter the current tab's content
				ui.ExperimentalRequestInTabHeaders = false
				// Reset subchild for tab content
				ui.ExperimentalCurrentSubchild = 0
				// For Body tab, start at content type selector (subchild 0)
				// For other tabs, no subchildren
			} else {
				// We're in tab content mode
				if ui.ExperimentalCurrentChild == 0 && hasSubchildren(4, 0, ui) {
					// Body tab with subchildren
					// Check if we're in MultipartFields and need to navigate within multipart elements
					if ui.ExperimentalCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
						// Check if we're in a field row (multipart element >= 2)
						if ui.ExperimentalCurrentMultipartElement >= 2 {
							// We're in a field row, navigate within field row elements
							fieldRowIndex := ui.ExperimentalCurrentMultipartElement - 2
							if fieldRowIndex >= 0 && fieldRowIndex < len(currentMultipartFieldRows) {
								fieldRow := currentMultipartFieldRows[fieldRowIndex]
								if fieldRow != nil {
									maxFieldRowElement := getMaxFieldRowElement(fieldRow)

									if ui.ExperimentalCurrentFieldRowElement < maxFieldRowElement {
										// Move to next element within the field row
										ui.ExperimentalCurrentFieldRowElement++
									} else {
										// At last element in field row, move to next multipart element
										ui.ExperimentalCurrentFieldRowElement = 0
										ui.ExperimentalCurrentMultipartElement++

										// Check if we're past the last multipart element
										maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
										if currentMultipartFieldRows != nil {
											maxMultipartElement = 2 + len(currentMultipartFieldRows)
										}

										if ui.ExperimentalCurrentMultipartElement >= maxMultipartElement {
											// Past the last multipart element, exit multipart fields and jump to response panel
											ui.ExperimentalCurrentMultipartElement = 0
											ui.ExperimentalCurrentFieldRowElement = 0
											ui.ExperimentalCurrentContainer = 5 // Response panel
											ui.ExperimentalCurrentChild = 0     // PreviewTab
											ui.ExperimentalCurrentSubchild = 0
											ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
										}
									}
								} else {
									// Field row is nil, move to next multipart element
									ui.ExperimentalCurrentFieldRowElement = 0
									ui.ExperimentalCurrentMultipartElement++

									// Check if we're past the last multipart element
									maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
									if currentMultipartFieldRows != nil {
										maxMultipartElement = 2 + len(currentMultipartFieldRows)
									}

									if ui.ExperimentalCurrentMultipartElement >= maxMultipartElement {
										// Past the last multipart element, exit multipart fields and jump to response panel
										ui.ExperimentalCurrentMultipartElement = 0
										ui.ExperimentalCurrentFieldRowElement = 0
										ui.ExperimentalCurrentContainer = 5 // Response panel
										ui.ExperimentalCurrentChild = 0     // PreviewTab
										ui.ExperimentalCurrentSubchild = 0
										ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
									}
								}
							} else {
								// Invalid field row index - we're past the last field row
								// Check if we're past the last multipart element
								maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
								if currentMultipartFieldRows != nil {
									maxMultipartElement = 2 + len(currentMultipartFieldRows)
								}

								if ui.ExperimentalCurrentMultipartElement >= maxMultipartElement {
									// Past the last multipart element, exit multipart fields and jump to response panel
									ui.ExperimentalCurrentMultipartElement = 0
									ui.ExperimentalCurrentFieldRowElement = 0
									ui.ExperimentalCurrentContainer = 5 // Response panel
									ui.ExperimentalCurrentChild = 0     // PreviewTab
									ui.ExperimentalCurrentSubchild = 0
									ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
								} else {
									// Not past the last multipart element, reset to first
									ui.ExperimentalCurrentFieldRowElement = 0
									ui.ExperimentalCurrentMultipartElement = 0
								}
							}
						} else {
							// We're at a button (Add Field or Delete All), move to next multipart element
							maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
							if currentMultipartFieldRows != nil {
								maxMultipartElement = 2 + len(currentMultipartFieldRows)
							}

							if ui.ExperimentalCurrentMultipartElement < maxMultipartElement {
								// Move to next multipart element
								ui.ExperimentalCurrentMultipartElement++
								// Reset field row element when moving to a new multipart element
								ui.ExperimentalCurrentFieldRowElement = 0
							} else {
								// At last multipart element, exit multipart fields and jump to response panel
								ui.ExperimentalCurrentMultipartElement = 0
								ui.ExperimentalCurrentFieldRowElement = 0
								ui.ExperimentalCurrentContainer = 5 // Response panel
								ui.ExperimentalCurrentChild = 0     // PreviewTab
								ui.ExperimentalCurrentSubchild = 0
								ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
							}
						}
					} else {
						// Not in multipart navigation mode, move to next valid subchild
						maxSubchild := getMaxSubchildForChild(4, 0, ui)

						// Special case: JSON with no content at ContentTypeSelector
						if ui.ExperimentalCurrentSubchild == 0 && getCurrentContentType(ui) == "JSON" {
							// Check if JSON has content
							hasJSONContent := false
							if ui.CurrentBodyContent != "" && ui.CurrentBodyContent != "{}" && ui.CurrentBodyContent != "[]" {
								hasJSONContent = true
							} else if ui.JSONBodyContent != "" && ui.JSONBodyContent != "{}" && ui.JSONBodyContent != "[]" {
								hasJSONContent = true
							}

							if !hasJSONContent {
								// JSON has no content, jump directly to response panel
								ui.ExperimentalCurrentContainer = 5 // Response panel
								ui.ExperimentalCurrentChild = 0     // PreviewTab
								// Don't reset subchild - keep it as 0 (ContentTypeSelector)
								ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
								// Set Request panel to tab headers mode for backtab navigation
								ui.ExperimentalRequestInTabHeaders = true
								// Reset multipart element
								ui.ExperimentalCurrentMultipartElement = 0
								ui.ExperimentalCurrentFieldRowElement = 0
								// Update borders and sync state
								if previousContainer != ui.ExperimentalCurrentContainer {
									// Deactivate border of previous container
									if previousContainer < len(ui.MainCycle.panels) {
										ui.SetInactiveBorder(ui.MainCycle.panels[previousContainer])
									}
									// Activate border of current container
									if ui.ExperimentalCurrentContainer < len(ui.MainCycle.panels) {
										ui.SetActiveBorder(ui.MainCycle.panels[ui.ExperimentalCurrentContainer])
									}
									// Update previous container tracking
									ui.ExperimentalPreviousContainer = previousContainer
								}
								// Sync MainCycle.current with experimental container
								syncMainCycleWithExperimental(ui)
								// Set focus based on new coordinates
								setFocusForCoordinates(ui)
								// Update footer with navigation info
								ui.UpdateFooter()
								return nil
							}
						}

						if ui.ExperimentalCurrentSubchild < maxSubchild {
							// Move to next valid subchild in BodyTab (skip invalid ones based on content type)
							ui.ExperimentalCurrentSubchild = getNextValidSubchild(ui.ExperimentalCurrentSubchild, ui)
							// Reset multipart element when leaving MultipartFields
							ui.ExperimentalCurrentMultipartElement = 0
							ui.ExperimentalCurrentFieldRowElement = 0
						} else {
							// At last BodyTab subchild, jump to response panel
							ui.ExperimentalCurrentContainer = 5 // Response panel
							ui.ExperimentalCurrentChild = 0     // PreviewTab
							// Don't reset subchild - keep it for returning to same position
							// ui.ExperimentalCurrentSubchild remains as is (1 for JSONEditor, etc.)
							ui.ExperimentalResponseInTabHeaders = true // Start in tab headers mode
							// Set Request panel to tab headers mode
							ui.ExperimentalRequestInTabHeaders = true
							// Reset multipart element
							ui.ExperimentalCurrentMultipartElement = 0
							ui.ExperimentalCurrentFieldRowElement = 0
						}
					}
				} else {
					// Other tabs (Auth, Query, Headers) or BodyTab without subchildren
					// For these tabs, tab should exit tab content mode and go back to tab headers
					ui.ExperimentalRequestInTabHeaders = true
					// Stay on current child (current tab header)
					ui.ExperimentalCurrentSubchild = 0
					ui.ExperimentalCurrentMultipartElement = 0
					ui.ExperimentalCurrentFieldRowElement = 0
				}
			}
		} else if ui.ExperimentalCurrentContainer == 5 {
			// Response panel navigation logic
			if ui.ExperimentalResponseInTabHeaders {
				// We're in tab headers mode
				// Tab should enter the current tab's content
				ui.ExperimentalResponseInTabHeaders = false
				// Reset subchild for tab content (Response panel doesn't have subchildren)
				ui.ExperimentalCurrentSubchild = 0
			} else {
				// We're in tab content mode
				// Tab should move to next container (Workspace panel)
				ui.ExperimentalCurrentContainer = 0 // Workspace panel
				ui.ExperimentalCurrentChild = 0     // WorkspaceSelector
				ui.ExperimentalCurrentSubchild = 0
				// Reset multipart and field row elements
				ui.ExperimentalCurrentMultipartElement = 0
				ui.ExperimentalCurrentFieldRowElement = 0
				// Reset response tab headers for next time
				ui.ExperimentalResponseInTabHeaders = true
			}
		} else {
			// Normal navigation for other containers
			// Check if current position has subchildren
			if hasSubchildren(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui) {
				// We're in a container/child that has subchildren (e.g., Request BodyTab)
				maxSubchild := getMaxSubchildForChild(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui)

				if ui.ExperimentalCurrentSubchild < maxSubchild {
					// Move to next subchild
					ui.ExperimentalCurrentSubchild++
					// If moving to MultipartFields, reset multipart element
					if ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalCurrentChild == 0 && ui.ExperimentalCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
						ui.ExperimentalCurrentMultipartElement = 0
						ui.ExperimentalCurrentFieldRowElement = 0
					} else {
						ui.ExperimentalCurrentMultipartElement = 0
						ui.ExperimentalCurrentFieldRowElement = 0
					}
				} else {
					// At last subchild, move to next child and reset subchild
					ui.ExperimentalCurrentSubchild = 0
					ui.ExperimentalCurrentMultipartElement = 0
					ui.ExperimentalCurrentFieldRowElement = 0
					maxChild := getMaxChildForContainer(ui.ExperimentalCurrentContainer)

					if ui.ExperimentalCurrentChild < maxChild {
						// Move to next child in same container
						ui.ExperimentalCurrentChild++
					} else {
						// At last child, move to next container and reset child
						ui.ExperimentalCurrentChild = 0
						ui.ExperimentalCurrentContainer = (ui.ExperimentalCurrentContainer + 1) % 6 // 6 containers total
					}
				}
			} else {
				// No subchildren at current position
				maxChild := getMaxChildForContainer(ui.ExperimentalCurrentContainer)

				if ui.ExperimentalCurrentChild < maxChild {
					// Move to next child in same container
					ui.ExperimentalCurrentChild++
				} else {
					// At last child, move to next container and reset child
					ui.ExperimentalCurrentChild = 0
					ui.ExperimentalCurrentContainer = (ui.ExperimentalCurrentContainer + 1) % 6 // 6 containers total
				}
				// Reset subchild when moving to a position without subchildren
				ui.ExperimentalCurrentSubchild = 0
			}
		}

		// Update borders if container changed
		if previousContainer != ui.ExperimentalCurrentContainer {
			// Deactivate border of previous container
			if previousContainer < len(ui.MainCycle.panels) {
				ui.SetInactiveBorder(ui.MainCycle.panels[previousContainer])
			}
			// Activate border of current container
			if ui.ExperimentalCurrentContainer < len(ui.MainCycle.panels) {
				ui.SetActiveBorder(ui.MainCycle.panels[ui.ExperimentalCurrentContainer])
			}
			// Update previous container tracking
			ui.ExperimentalPreviousContainer = previousContainer
		}

		// Sync MainCycle.current with experimental container
		syncMainCycleWithExperimental(ui)

		// Sync experimental child with current tab index before getting message
		syncExperimentalChildWithCurrentTab(ui)

		// Set focus to the appropriate UI element based on current coordinates
		setFocusForCoordinates(ui)

		// Return nil to prevent further navigation processing
		// This makes the experimental system take over tab navigation
		return nil
	}

	return event
}

// handleBacktabNavigation handles Backtab (Shift+Tab) key navigation
func handleBacktabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Experimental navigation system (disabled by default)
	if ui.ExperimentalNavigationEnabled {
		// Check if a modal is open - if so, let the modal handle Backtab
		if name, _ := ui.Pages.GetFrontPage(); name != "main" {
			return event
		}
		// Store previous container before updating
		previousContainer := ui.ExperimentalCurrentContainer

		// Special handling for Request panel (container 4)
		if ui.ExperimentalCurrentContainer == 4 {
			// Request panel backtab navigation logic
			if ui.ExperimentalRequestInTabHeaders {
				// We're in tab headers mode
				// Backtab should move to previous container (URLBar)
				ui.ExperimentalCurrentContainer = 3                      // URLBar panel
				ui.ExperimentalCurrentChild = getMaxChildForContainer(3) // Last child of URLBar
				ui.ExperimentalRequestInTabHeaders = true                // Reset for Request panel
				ui.ExperimentalCurrentSubchild = 0
			} else {
				// We're in tab content mode
				if ui.ExperimentalCurrentChild == 0 && ui.ExperimentalCurrentSubchild > 0 {
					// Body tab with subchildren, not at first subchild
					// Check if we're in MultipartFields and need to navigate within multipart elements
					if ui.ExperimentalCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
						// Check if we're in a field row (multipart element >= 2)
						if ui.ExperimentalCurrentMultipartElement >= 2 {
							// We're in a field row
							if ui.ExperimentalCurrentFieldRowElement > 0 {
								// Move to previous element within the field row
								ui.ExperimentalCurrentFieldRowElement--
							} else {
								// At first element in field row, move to previous multipart element
								ui.ExperimentalCurrentFieldRowElement = 0

								// Move to previous multipart element
								if ui.ExperimentalCurrentMultipartElement > 0 {
									ui.ExperimentalCurrentMultipartElement--

									// If moving to another field row, set field row element to last element
									if ui.ExperimentalCurrentMultipartElement >= 2 {
										fieldRowIndex := ui.ExperimentalCurrentMultipartElement - 2
										if fieldRowIndex >= 0 && fieldRowIndex < len(currentMultipartFieldRows) {
											fieldRow := currentMultipartFieldRows[fieldRowIndex]
											if fieldRow != nil {
												ui.ExperimentalCurrentFieldRowElement = getMaxFieldRowElement(fieldRow)
											}
										}
									}
								}
							}
						} else if ui.ExperimentalCurrentMultipartElement > 0 {
							// We're at a button (Delete All), move to previous multipart element
							ui.ExperimentalCurrentMultipartElement--
							ui.ExperimentalCurrentFieldRowElement = 0
						} else {
							// At Add Field button (multipart element 0), move to previous subchild
							ui.ExperimentalCurrentSubchild = getPrevValidSubchild(ui.ExperimentalCurrentSubchild, ui)
							// Reset multipart element when leaving MultipartFields
							ui.ExperimentalCurrentMultipartElement = 0
							ui.ExperimentalCurrentFieldRowElement = 0
						}
					} else {
						// Not in multipart navigation mode, move to previous subchild
						ui.ExperimentalCurrentSubchild = getPrevValidSubchild(ui.ExperimentalCurrentSubchild, ui)
						// Reset multipart element when leaving MultipartFields
						ui.ExperimentalCurrentMultipartElement = 0
						ui.ExperimentalCurrentFieldRowElement = 0
						ui.ExperimentalCurrentFieldRowElement = 0
						ui.ExperimentalCurrentFieldRowElement = 0
					}
				} else if ui.ExperimentalCurrentChild == 0 && ui.ExperimentalCurrentSubchild == 0 {
					// At first BodyTab subchild, enter tab headers mode
					ui.ExperimentalRequestInTabHeaders = true
					// Stay on child 0 (Body tab header)
					// Reset multipart element
					ui.ExperimentalCurrentMultipartElement = 0
					ui.ExperimentalCurrentFieldRowElement = 0
				} else if ui.ExperimentalCurrentChild > 0 {
					// Other tabs (Auth, Query, Headers)
					// For these tabs, backtab should exit tab content mode and go back to tab headers
					ui.ExperimentalRequestInTabHeaders = true
					// Stay on current child (current tab header)
					ui.ExperimentalCurrentSubchild = 0
					ui.ExperimentalCurrentMultipartElement = 0
					ui.ExperimentalCurrentFieldRowElement = 0
				}
			}
		} else if ui.ExperimentalCurrentContainer == 5 {
			// Response panel backtab navigation logic
			if ui.ExperimentalResponseInTabHeaders {
				// We're in tab headers mode
				// Backtab should move to previous container (Request panel)
				ui.ExperimentalCurrentContainer = 4 // Request panel

				// Check if Request panel is in tab headers mode
				if ui.ExperimentalRequestInTabHeaders {
					// Request panel is in tab headers mode, focus on current tab header
					ui.ExperimentalCurrentChild = ui.CurrentTabIndex
				} else {
					// Request panel is in tab content mode, focus on last child
					ui.ExperimentalCurrentChild = getMaxChildForContainer(4) // Last child of Request
				}

				ui.ExperimentalResponseInTabHeaders = true // Reset for Response panel
				ui.ExperimentalCurrentSubchild = 0
			} else {
				// We're in tab content mode
				// Backtab should exit tab content mode and go back to tab headers
				ui.ExperimentalResponseInTabHeaders = true
				// Stay on current child (current tab header)
				ui.ExperimentalCurrentSubchild = 0
			}
		} else {
			// Normal backtab navigation for other containers
			// Check if current position has subchildren
			if hasSubchildren(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui) {
				// We're in a container/child that has subchildren
				if ui.ExperimentalCurrentSubchild > 0 {
					// Move to previous subchild
					ui.ExperimentalCurrentSubchild--
				} else {
					// At first subchild, need to move to previous child
					if ui.ExperimentalCurrentChild > 0 {
						// Move to previous child in same container
						ui.ExperimentalCurrentChild--
						// Set subchild to max if new child has subchildren
						if hasSubchildren(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui) {
							ui.ExperimentalCurrentSubchild = getMaxSubchildForChild(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui)
							// If moving to MultipartFields, set multipart element to last element
							if ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalCurrentChild == 0 && ui.ExperimentalCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
								maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
								if currentMultipartFieldRows != nil {
									maxMultipartElement = 2 + len(currentMultipartFieldRows)
								}
								ui.ExperimentalCurrentMultipartElement = maxMultipartElement
							} else {
								ui.ExperimentalCurrentMultipartElement = 0
								ui.ExperimentalCurrentFieldRowElement = 0
							}
						} else {
							ui.ExperimentalCurrentSubchild = 0
							ui.ExperimentalCurrentMultipartElement = 0
							ui.ExperimentalCurrentFieldRowElement = 0
						}
					} else {
						// At first child, move to previous container
						prevContainer := (ui.ExperimentalCurrentContainer - 1 + 6) % 6
						prevMaxChild := getMaxChildForContainer(prevContainer)
						ui.ExperimentalCurrentContainer = prevContainer
						ui.ExperimentalCurrentChild = prevMaxChild
						// Check if new child has subchildren
						if hasSubchildren(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui) {
							ui.ExperimentalCurrentSubchild = getMaxSubchildForChild(ui.ExperimentalCurrentContainer, ui.ExperimentalCurrentChild, ui)
							// If moving to MultipartFields, set multipart element to last element
							if ui.ExperimentalCurrentContainer == 4 && ui.ExperimentalCurrentChild == 0 && ui.ExperimentalCurrentSubchild == 2 && getCurrentContentType(ui) == "Multipart" {
								maxMultipartElement := 2 // Start with 2 buttons (0: Add, 1: Delete All)
								if currentMultipartFieldRows != nil {
									maxMultipartElement = 2 + len(currentMultipartFieldRows)
								}
								ui.ExperimentalCurrentMultipartElement = maxMultipartElement
							} else {
								ui.ExperimentalCurrentMultipartElement = 0
								ui.ExperimentalCurrentFieldRowElement = 0
							}
						} else {
							ui.ExperimentalCurrentSubchild = 0
							ui.ExperimentalCurrentMultipartElement = 0
							ui.ExperimentalCurrentFieldRowElement = 0
						}
					}
				}
			} else {
				// No subchildren at current position
				if ui.ExperimentalCurrentChild > 0 {
					// Move to previous child in same container
					ui.ExperimentalCurrentChild--
					// Reset subchild
					ui.ExperimentalCurrentSubchild = 0
				} else {
					// At first child, move to previous container
					prevContainer := (ui.ExperimentalCurrentContainer - 1 + 6) % 6
					prevMaxChild := getMaxChildForContainer(prevContainer)
					ui.ExperimentalCurrentContainer = prevContainer
					ui.ExperimentalCurrentChild = prevMaxChild
					// Reset subchild
					ui.ExperimentalCurrentSubchild = 0
				}
			}
		}

		// Update borders if container changed
		if previousContainer != ui.ExperimentalCurrentContainer {
			// Deactivate border of previous container
			if previousContainer < len(ui.MainCycle.panels) {
				ui.SetInactiveBorder(ui.MainCycle.panels[previousContainer])
			}
			// Activate border of current container
			if ui.ExperimentalCurrentContainer < len(ui.MainCycle.panels) {
				ui.SetActiveBorder(ui.MainCycle.panels[ui.ExperimentalCurrentContainer])
			}
			// Update previous container tracking
			ui.ExperimentalPreviousContainer = previousContainer
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

	// If experimental navigation is not enabled, return event to let default handling
	return event
}
