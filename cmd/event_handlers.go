package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

func refreshCollectionsTree(ui *UIOrchestrator) {
	ui.RootNode.ClearChildren()

	addWorkspaceToTree(ui.WorkspaceData, ui.RootNode)

	ui.CollectionsTreeView.SetRoot(ui.RootNode)
	ui.CollectionsTreeView.SetCurrentNode(ui.RootNode)
}

// SetupEventHandlers configures all event handlers for the UI
func SetupEventHandlers(ui *UIOrchestrator) {
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
		ui.DataManager = NewDataManager(newWorkspace)

		refreshCollectionsTree(ui)

		ui.CurrentRequest = nil
		ui.CurrentSelectedNode = nil
		ui.LastSelectedRequestNode = nil

		ui.ProgrammaticallyUpdatingURL = true
		ui.URLInput.SetText("")
		ui.ProgrammaticallyUpdatingURL = false

		ui.SyncBodyContent("")
		setHeadersInUI(ui.Colors, nil, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) })

		updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors)

		// ui.FooterRight.SetText(fmt.Sprintf("Switched to workspace: %s (%d collections)", text, len(newWorkspace.Collections)))

		go func() {
			time.Sleep(50 * time.Millisecond)
			ui.App.QueueUpdateDraw(func() {
				ui.App.SetFocus(ui.CollectionsTreeView)
			})
		}()
	}

	ui.WorkspaceSelector.SetSelectedFunc(func(text string, index int) {
		// ui.FooterRight.SetText(fmt.Sprintf("Text %s, Index: %d, workspace name: %s", text, index, workspaceNames[index]))
		if index >= 0 && index < len(workspaceNames) {
			switchWorkspace(workspaceNames[index])
		}
	})

	// ui.WorkspaceSelector.SetDoneFunc(func(key tcell.Key) {
	// if key != tcell.KeyEnter {
	// 	return
	// }
	// index, _ := ui.WorkspaceSelector.GetCurrentOption()
	// if index >= 0 && index < len(workspaceNames) {
	// 	ui.FooterRight.SetText(fmt.Sprintf("Workspace: %s", index))
	// 	// switchWorkspace(workspaceNames[index])
	// }
	// })

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
				// Format content with syntax highlighting
				formattedContent := formatBodyContent(content)

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

	// Set up collections tree view selected function
	ui.CollectionsTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(tcell.ColorDefault).Foreground(tcell.ColorDefault))

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
			setHeadersInUI(ui.Colors, req.Headers, func() { saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData) }, func(p tview.Primitive) { ui.App.SetFocus(p) })

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
					updateResponseTabs(cmdResp, &lastResponse.Timestamp, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors)
				} else {
					// Clear response if no history
					updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors)
				}
			}
		} else if col, ok := reference.(workspace.Collection); ok {
			// Save current request headers before clearing
			if ui.CurrentRequest != nil {
				ui.CurrentRequest.Headers = getHeadersFromUI()
			}

			// Clear current request when a collection is selected
			ui.CurrentRequest = nil
			ui.CurrentSelectedNode = nil

			// Handle collection expansion
			expanded := !node.IsExpanded()
			node.SetExpanded(expanded)

			if expanded {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
				if len(node.GetChildren()) == 0 {
					addChildrenToCollectionNode(node, col)
				}
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&ui.WorkspaceData.Collections, col.Name, true)
				}
			} else {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
				node.ClearChildren()
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&ui.WorkspaceData.Collections, col.Name, false)
				}
			}
		}

		// Track the last selected request node
		if reference := node.GetReference(); reference != nil {
			if _, ok := reference.(workspace.Request); ok {
				ui.LastSelectedRequestNode = node
			}
		}
	})

	// Set up environment config button click handler
	ui.EnvConfigButton.SetSelectedFunc(func() {
		showEnvironmentModal(ui)
	})

	// Set up workspace config button click handler
	ui.WorkspaceConfigButton.SetSelectedFunc(func() {
		// Create workspace management modal
		form := createWorkspaceManagementForm(ui.App, ui.Pages, ui.WorkspaceData, ui.WorkspaceSelector, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
		modal := createModal(form, 60, 15, tcell.ColorDefault)
		ui.Pages.AddPage("workspaceMenu", modal, true, true)
		ui.App.SetFocus(form)
	})

	ui.EnvDropdown.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		index, _ := ui.EnvDropdown.GetCurrentOption()
		if index == 0 {
			config.C.SelectedEnvironment = "Base"
		} else if index > 0 && index <= len(*ui.EnvironmentsData) {
			config.C.SelectedEnvironment = (*ui.EnvironmentsData)[index-1].Name
		}
		if err := config.SaveConfig(&config.C); err != nil {
		}
	})

	// Add send button functionality
	ui.SendButton.SetSelectedFunc(func() {
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
		} else if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
			// Other environment selected
			env := &(*ui.EnvironmentsData)[currentEnvIndex-1]
			envVars = env.GetEffectiveVariables(*ui.EnvironmentsData)
		}

		// Substitute variables in URL, body, and headers
		url = substituteVariables(url, envVars)
		body = substituteVariables(body, envVars)
		headers = substituteVariablesInHeaders(headers, envVars)

		// Validate URL
		if url == "" {
			updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors)
			return
		}

		// Send the request
		updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors) // Show loading state
		resp, err := SendRequest(method, url, body, headers)
		if err != nil {
			updateResponseTabs(nil, nil, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors) // Show error state
			return
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

		// Update the response tabs with the new response
		now := time.Now()
		updateResponseTabs(resp, &now, ui.Response, ui.ResponseTabHeader, &ui.ResponseInfoBar, &ui.ResponseTimeText, &ui.LastResponseTime, ui.ResponsePreviewPanel, ui.ResponseHeadersPanel, ui.ResponseCookiesPanel, ui.ResponseTimelinePanel, ui.Colors)
	})

	// Set up main application input capture for navigation and shortcuts
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're on a modal page (not main)
		currentPage, _ := ui.Pages.GetFrontPage()
		if currentPage != "main" {
			// On modal, 'q' closes the modal
			if event.Rune() == 'q' || event.Rune() == 'Q' {
				ui.Pages.RemovePage(currentPage)
				ui.Pages.SwitchToPage("main")
				ui.App.SetFocus(ui.CollectionsTreeView)
				return nil
			}
			// Let the form handle its own input
			return event
		}

		// On main page, 'q' quits
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			// Save expansion state before quitting if in "remember" mode
			if config.C.UI.CollectionExpansion == "remember" {
				if err := workspace.SaveExpansionState(&ui.WorkspaceData.Collections); err != nil {
					fmt.Printf("Warning: Failed to save expansion state: %v\n", err)
				}
			}
			ui.App.Stop()
			return nil
		}

		// When in body edit mode and focused on bodyEditPanel, pass all input through to allow pasting
		if ui.BodyEditMode && ui.App.GetFocus() == ui.BodyEditPanel {
			return event
		}

		// Allow URLVariableInput to handle its own Enter key events
		if event.Key() == tcell.KeyEnter && ui.App.GetFocus() == ui.URLInput {
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

		if event.Key() == tcell.KeyTab {
			// Handle tab navigation logic here
			return handleTabNavigation(ui, event)
		}

		if event.Key() == tcell.KeyBacktab {
			// Handle backtab navigation logic here
			return handleBacktabNavigation(ui, event)
		}

		// Collection shortcuts
		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'n' {
			form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
			modal := createModal(form, 50, 12, tcell.ColorDefault)
			ui.Pages.AddPage("newCollection", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}

		if ui.MainCycle.current == ui.CollectionsIndex && event.Rune() == 'r' {
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
					ui.Pages.AddPage("newRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
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
		if ui.CurrentFocus == ui.URLBarIndex {
			// Check if focus is on an input field (don't switch tabs if typing)
			currentFocusedElement := ui.App.GetFocus()
			isOnInputField := false

			// Check if focused on method dropdown
			if currentFocusedElement == ui.MethodDropdown {
				isOnInputField = true
			}
			// Check if focused on URL input
			if currentFocusedElement == ui.URLInput {
				isOnInputField = true
			}
			// Check if focused on body edit panel
			if currentFocusedElement == ui.BodyEditPanel {
				isOnInputField = true
			}
			// Check if focused on any header input fields
			for _, row := range currentHeaderRows {
				if currentFocusedElement == row.KeyInput || currentFocusedElement == row.ValueInput {
					isOnInputField = true
					break
				}
			}

			// Only switch tabs if not focused on an input field
			if !isOnInputField {
				switch event.Rune() {
				case '1':
					ui.TabPages.SwitchToPage("body")
					requestTabs := []string{"Body", "Auth", "Query", "Headers"}
					updateTabHeader(requestTabs, ui.TabHeader, 0, ui.Colors)
					ui.CurrentTabIndex = 0
					return nil
				case '2':
					ui.TabPages.SwitchToPage("auth")
					requestTabs := []string{"Body", "Auth", "Query", "Headers"}
					updateTabHeader(requestTabs, ui.TabHeader, 1, ui.Colors)
					ui.CurrentTabIndex = 1
					return nil
				case '3':
					ui.TabPages.SwitchToPage("query")
					requestTabs := []string{"Body", "Auth", "Query", "Headers"}
					updateTabHeader(requestTabs, ui.TabHeader, 2, ui.Colors)
					ui.CurrentTabIndex = 2
					return nil
				case '4':
					ui.TabPages.SwitchToPage("headers")
					requestTabs := []string{"Body", "Auth", "Query", "Headers"}
					updateTabHeader(requestTabs, ui.TabHeader, 3, ui.Colors)
					ui.CurrentTabIndex = 3
					return nil
				}
			}
		}

		// Vim-style modal editing: 'i' to enter insert mode
		if event.Rune() == 'i' && ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
			ui.SwitchBodyMode() // Switch to edit mode
			ui.App.SetFocus(ui.BodyEditPanel)
			return nil
		}

		// Arrow key navigation for tabs when request panel tabs are focused and not in body edit mode
		if ui.MainCycle.current == ui.RequestIndex && !ui.BodyEditMode {
			if event.Key() == tcell.KeyLeft {
				ui.CurrentTabIndex = (ui.CurrentTabIndex - 1 + 4) % 4
				tabNames := []string{"body", "auth", "query", "headers"}
				ui.TabPages.SwitchToPage(tabNames[ui.CurrentTabIndex])
				requestTabs := []string{"Body", "Auth", "Query", "Headers"}
				updateTabHeader(requestTabs, ui.TabHeader, ui.CurrentTabIndex, ui.Colors)
				// Focus the appropriate tab content
				switch ui.CurrentTabIndex {
				case 0: // Body tab
					if ui.BodyEditMode {
						ui.App.SetFocus(ui.BodyEditPanel)
					} else {
						ui.App.SetFocus(ui.BodyViewPanel)
					}
				case 3: // Headers tab
					if len(currentHeaderRows) > 0 && currentHeaderRows[0].KeyInput != nil {
						ui.App.SetFocus(currentHeaderRows[0].KeyInput)
					} else {
						ui.App.SetFocus(ui.RequestDataTabs)
					}
				default:
					ui.App.SetFocus(ui.RequestDataTabs)
				}
				return nil
			} else if event.Key() == tcell.KeyRight {
				ui.CurrentTabIndex = (ui.CurrentTabIndex + 1) % 4
				tabNames := []string{"body", "auth", "query", "headers"}
				ui.TabPages.SwitchToPage(tabNames[ui.CurrentTabIndex])
				requestTabs := []string{"Body", "Auth", "Query", "Headers"}
				updateTabHeader(requestTabs, ui.TabHeader, ui.CurrentTabIndex, ui.Colors)
				// Focus the appropriate tab content
				switch ui.CurrentTabIndex {
				case 0: // Body tab
					if ui.BodyEditMode {
						ui.App.SetFocus(ui.BodyEditPanel)
					} else {
						ui.App.SetFocus(ui.BodyViewPanel)
					}
				case 3: // Headers tab
					if len(currentHeaderRows) > 0 && currentHeaderRows[0].KeyInput != nil {
						ui.App.SetFocus(currentHeaderRows[0].KeyInput)
					} else {
						ui.App.SetFocus(ui.RequestDataTabs)
					}
				default:
					ui.App.SetFocus(ui.RequestDataTabs)
				}
				return nil
			}
		}

		return event
	})
}

// handleTabNavigation handles Tab key navigation
func handleTabNavigation(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// workspace panel - cycle through elements
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == 0 {
		next := ui.WorkspaceCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// workspace panel - move to next main panel
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == 1 {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		ui.EnvironmentsCycle.current = 0
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)

		ui.App.SetFocus(ui.EnvironmentsCycle.inputs[0])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel new - el 1
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == 0 {
		next := ui.EnvironmentsCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// environment panel new - el 2
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == 1 {
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
		// It can advance to the next panel
		nextElement := ui.MainCycle.Next()
		ui.RequestCycle.current = 0
		ui.SetActiveBorder(nextElement)

		ui.App.SetFocus(ui.MethodDropdown)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// urlbar panel & dropdown
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 0 {
		next := ui.RequestCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & url input
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 1 {
		next := ui.RequestCycle.Next()
		ui.App.SetFocus(next)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// urlbar panel & send button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 2 {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(ui.BodyViewPanel)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests editor/viewer panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		nextElement := ui.MainCycle.Next()
		ui.SetActiveBorder(nextElement)
		ui.App.SetFocus(nextElement)
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// requests headers panel - cycle through header inputs
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 3 {
		// Cycle through header key/value/delete inputs, then jump to next panel
		currentFocusedElement := ui.App.GetFocus()

		// Find current focused header input
		found := false
		for i, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement {
				// Currently on key input, move to value input of same row
				ui.App.SetFocus(row.ValueInput)
				found = true
				break
			} else if row.ValueInput == currentFocusedElement {
				// Currently on value input, move to delete button of same row
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
		ui.WorkspaceCycle.current = 0
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
	// workspace panel - cycle backward through elements
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == 1 {
		// Currently on config button, move to dropdown
		prev := ui.WorkspaceCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current

		return nil
	}

	// workspace panel - move to previous main panel
	if ui.MainCycle.current == ui.WorkspaceIndex && ui.WorkspaceCycle.current == 0 {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.WorkspaceCycle.inputs[0])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// environment panel - cycle backward through elements
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == 1 {
		// Currently on config button, move to dropdown
		prev := ui.EnvironmentsCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current
		return nil
	}

	// environment panel - move to previous main panel
	if ui.MainCycle.current == ui.EnviromentIndex && ui.EnvironmentsCycle.current == 0 {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.WorkspaceCycle.inputs[1])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()
		return nil
	}

	// collections panel
	if ui.MainCycle.current == ui.CollectionsIndex {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.EnvironmentsCycle.current = 1
		ui.App.SetFocus(ui.EnvironmentsCycle.inputs[1])
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()

		return nil
	}

	// urlbar panel & dropdown
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 0 {
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
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 1 {
		prev := ui.RequestCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current
		return nil
	}

	// urlbar panel & send button
	if ui.MainCycle.current == ui.URLBarIndex && ui.RequestCycle.current == 2 {
		prev := ui.RequestCycle.Prev()
		ui.App.SetFocus(prev)
		ui.CurrentFocus = ui.MainCycle.current
		return nil
	}

	// requests editor/viewer panel
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
		ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
		prevElement := ui.MainCycle.Prev()
		ui.SetActiveBorder(prevElement)
		ui.App.SetFocus(ui.SendButton)
		ui.RequestCycle.current = 2
		ui.CurrentFocus = ui.MainCycle.current
		ui.UpdateFooter()
		return nil
	}

	// requests headers panel - cycle backward through header inputs
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 3 {
		// Cycle backward through header key/value/delete inputs
		currentFocusedElement := ui.App.GetFocus()

		// Find current focused header input
		found := false
		for i, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement {
				// Currently on key input, move to previous row's delete button or previous panel
				if i > 0 {
					// Move to previous row's delete button
					ui.App.SetFocus(currentHeaderRows[i-1].DeleteButton)
				} else {
					// First key input, move to previous main panel (request panel)
					ui.SetInactiveBorder(ui.MainCycle.panels[ui.MainCycle.current])
					prevElement := ui.MainCycle.Prev()
					ui.SetActiveBorder(prevElement)
					ui.App.SetFocus(ui.SendButton)
					ui.RequestCycle.current = 2
					ui.UpdateFooter()
				}
				found = true
				break
			} else if row.ValueInput == currentFocusedElement {
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
