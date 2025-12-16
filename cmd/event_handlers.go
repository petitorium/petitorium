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
	isRequestPanel := ui.MainCycle.current == ui.RequestIndex
	// Check if we're in response panel
	isResponsePanel := ui.MainCycle.current == ui.ResponseIndex

	if isNumberKey {
		// Number keys work when request or response panel has focus
		if ui.CurrentFocus != ui.RequestIndex && ui.CurrentFocus != ui.ResponseIndex {
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
		tabNames := []string{"body", "auth", "query", "headers"}
		ui.TabPages.SwitchToPage(tabNames[targetTabIndex])
		requestTabs := []string{"Body", "Auth", "Query", "Headers"}
		updateTabHeader(requestTabs, ui.TabHeader, targetTabIndex, ui.Colors)
		ui.CurrentTabIndex = targetTabIndex

		// Focus the appropriate tab content
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
	} else if isResponsePanel {
		responseTabNames := []string{"preview", "headers", "cookies", "timeline"}
		ui.ResponsePages.SwitchToPage(responseTabNames[targetTabIndex])
		updateResponseTabHeader(ui.ResponseTabHeader, targetTabIndex, ui.Colors)
		ui.CurrentResponseTabIndex = targetTabIndex
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

	// Initialize switchBodyMode function
	ui.SwitchBodyMode = func() {
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
				if ui.CurrentSelectedNode != nil {
					ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
					saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
				}
			}
			ui.SyncBodyContent(ui.CurrentBodyContent)
		}
		ui.UpdateFooter()
	}

	// Update the original switchBodyMode with proper focus handling
	ui.SwitchBodyMode = func() {
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

			// Set method in dropdown AFTER currentRequest is set
			if ui.CurrentRequest != nil {
				syncMethodDropdown(ui.CurrentRequest, ui.MethodDropdown, &ui.ProgrammaticallyUpdatingMethod)

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
		if ui.CurrentRequest != nil {
			body = ui.CurrentRequest.Body
		}
		headers := getHeadersFromUI()

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
			resp, err := SendRequest(method, url, body, headers)

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
		if ui.CurrentRequest != nil {
			body = ui.CurrentRequest.Body
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
		curlCommand := generateCurlCommand(method, url, headers, body)

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
		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'n' && !isFormPopupActive {
			form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
			modal := createModal(form, 50, 12, tcell.ColorDefault)
			setFormPopupActive(true)
			ui.Pages.AddPage("newCollection", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}

		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'r' && !isFormPopupActive {
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
		if ui.MainCycle.current == ui.CollectionsIndex {
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
		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'R' {
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
		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'm' {
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
		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'd' {
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
		if ui.MainCycle.current == ui.RequestIndex && event.Key() == tcell.KeyF4 {
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
		if event.Rune() == 'i' && ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
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
func handleTabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// If a modal is open, let it handle Tab navigation
	if name, _ := ui.Pages.GetFrontPage(); name != "main" {
		return event
	}

	// workspace panel - cycle through elements
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == ui.WorkspaceSelectorIndex {
		next := ui.WorkspaceCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// workspace panel - move to next main panel
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == ui.WorkspaceConfigButtonIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		ui.EnvironmentsCycle.current = ui.EnvironmentSelectorIndex
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)

		ui.App.SetFocus(ui.EnvironmentsCycle.inputs[0])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel new - el 0
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == ui.EnvironmentSelectorIndex {
		next := ui.EnvironmentsCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel new - el 1
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == ui.EnvironmentConfigButtonIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)

		ui.App.SetFocus(nextElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// collections panel
	if ui.MainCycle.current == ui.CollectionsIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.RequestCycle.current = ui.URLBarSelectorIndex
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(ui.MethodDropdown)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// urlbar panel & dropdown
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarSelectorIndex {
		next := ui.RequestCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & url input
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarInputIndex {
		next := ui.RequestCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & send button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarSendButtonIndex {
		next := ui.RequestCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & curl button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarCurlButtonIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(ui.BodyViewPanel)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests editor/viewer panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPBodyTabIndex && !ui.BodyEditMode {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(nextElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests auth panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPAuthTabIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(nextElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests query panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPQueryTabIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(nextElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests headers panel - cycle through header inputs
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPHeadersTabIndex {
		// Cycle through header key/value/delete inputs, then jump to next panel
		currentFocusedElement := ui.App.GetFocus()

		// Find current focused header input
		found := false
		for i, row := range currentHeaderRows {
			if row.KeyInput.HasFocus() {
				// Currently on key input, move to value input of same row
				ui.App.SetFocus(row.ValueInput)
				found = true
				break
			} else if row.ValueInput.HasFocus() {
				// Currently on value input (HeaderValueInput), move to delete button of same row
				ui.App.SetFocus(row.DeleteButton)
				found = true
				break
			} else if row.DeleteButton == currentFocusedElement {
				// Currently on delete button, move to next row's key input or next panel
				if i < len(currentHeaderRows)-1 {
					// Move to next row's key input
					ui.App.SetFocus(currentHeaderRows[i+1].KeyInput)
				} else {
					// Last delete button, move to next main panel (response)
					ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
					nextElement := ui.MainCycle.Next()
					ui.SetActiveBorder(nextElement)
					ui.App.SetFocus(nextElement)
					ui.CurrentFocus = ui.MainCycle.current
					ui.UpdateFooter()
				}
				found = true
				break
			}
		}

		// If no header input was focused, focus the first key input
		if !found && len(currentHeaderRows) > 0 {
			ui.App.SetFocus(currentHeaderRows[0].KeyInput)
		}

		return nil
	}

	// response panel
	if ui.MainCycle.current == ui.ResponseIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		ui.WorkspaceCycle.current = ui.WorkspaceSelectorIndex
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		// Set focus to the first element in the workspace cycle (dropdown)
		ui.App.SetFocus(ui.WorkspaceCycle.inputs[0])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	return nil
}

// handleBacktabNavigation handles Backtab (Shift+Tab) key navigation
func handleBacktabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// If a modal is open, let it handle Backtab navigation
	if name, _ := ui.Pages.GetFrontPage(); name != "main" {
		return event
	}

	// workspace panel - move to previous main panel
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == ui.WorkspaceSelectorIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// workspace panel - cycle backward through elements
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == ui.WorkspaceConfigButtonIndex {
		prev := ui.WorkspaceCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel - cycle backward through elements
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == ui.EnvironmentSelectorIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.WorkspaceCycle.current = 1
		ui.App.SetFocus(ui.WorkspaceCycle.inputs[1])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel - move to previous main panel
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == ui.EnvironmentConfigButtonIndex {
		ui.App.SetFocus(ui.EnvironmentsCycle.inputs[0])
		ui.EnvironmentsCycle.current = ui.EnvironmentSelectorIndex
		ui.UpdateFooter()

		return nil
	}

	// collections panel
	if ui.MainCycle.current == ui.CollectionsIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.EnvironmentsCycle.current = ui.EnvironmentConfigButtonIndex
		ui.App.SetFocus(ui.EnvironmentsCycle.inputs[1])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// urlbar panel & dropdown
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarSelectorIndex {
		prev := ui.RequestCycle.Prev()
		if prev != nil {
			ui.App.SetFocus(prev)
		} else {
			// Wrap to previous main panel
			ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
			prevElement := ui.MainCycle.Prev()
			ui.SetActiveBorder(prevElement)
			ui.App.SetFocus(prevElement)
			ui.CurrentFocus = ui.MainCycle.current
			ui.UpdateFooter()
		}

		return nil
	}

	// urlbar panel & url input
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarInputIndex {
		prev := ui.RequestCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & send button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarSendButtonIndex {
		prev := ui.RequestCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & curl button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == ui.URLBarCurlButtonIndex {
		prev := ui.RequestCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// requests editor/viewer panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPBodyTabIndex && !ui.BodyEditMode {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.CurlButton)
		ui.RequestCycle.current = ui.URLBarCurlButtonIndex
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests auth panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPAuthTabIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.CurlButton)
		ui.RequestCycle.current = ui.URLBarCurlButtonIndex
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests query panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPQueryTabIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.CurlButton)
		ui.RequestCycle.current = ui.URLBarCurlButtonIndex
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests headers panel - cycle backward through header inputs
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == ui.RPHeadersTabIndex {
		// Cycle backward through header key/value/delete inputs
		currentFocusedElement := ui.App.GetFocus()

		// Find current focused header input
		found := false
		for i, row := range currentHeaderRows {
			if row.KeyInput.HasFocus() {
				// Currently on key input, move to previous row's delete button or previous panel
				if i > 0 {
					// Move to previous row's delete button
					ui.App.SetFocus(currentHeaderRows[i-1].DeleteButton)
				} else {
					// First key input, move to previous main panel (request panel)
					ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
					prevElement := ui.MainCycle.Prev()
					ui.SetActiveBorder(prevElement)
					ui.App.SetFocus(ui.CurlButton)
					ui.RequestCycle.current = ui.URLBarCurlButtonIndex
					ui.CurrentFocus = ui.MainCycle.current
					ui.UpdateFooter()
				}
				found = true
				break
			} else if row.ValueInput.HasFocus() {
				// Currently on value input, move to key input of same row
				ui.App.SetFocus(row.KeyInput)
				found = true
				break
			} else if row.DeleteButton == currentFocusedElement {
				// Currently on delete button, move to value input of same row
				ui.App.SetFocus(row.ValueInput)
				found = true
				break
			}
		}

		// If no header input was focused, focus the last delete button
		if !found && len(currentHeaderRows) > 0 {
			ui.App.SetFocus(currentHeaderRows[len(currentHeaderRows)-1].DeleteButton)
		}

		return nil
	}

	// response panel
	if ui.MainCycle.current == ui.ResponseIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.BodyViewPanel)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()
		return nil
	}

	return nil
}
