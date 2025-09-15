package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

var rootCmd = &cobra.Command{
	Use:   "petitorium",
	Short: "A powerful TUI for API interaction and testing.",
	Run:   runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := config.LoadConfig(); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	})
}

func runTUI(cmd *cobra.Command, args []string) {
	backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor, activeTabColor := setupTheme()

	collectionsData, err := workspace.LoadCollections()
	if err != nil {
		panic(fmt.Sprintf("Failed to load collections: %v", err))
	}

	app := tview.NewApplication().
		EnableMouse(true)

	header := createPanel(" Petitorium ", backgroundColor, borderColor, titleColor, foregroundColor)
	// Change the root node's text to an empty string. It's cleaner.
	rootNode := tview.NewTreeNode("").SetSelectable(false)

	// Create unified method+URL+Send bar
	methodUrlBar, methodDropdown, urlInput, sendButton := createMethodUrlBar(" Request ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Add placeholder functionality to Send button (no logic yet)
	sendButton.SetSelectedFunc(func() {
		// TODO: Implement request sending logic
	})

	// Create both view and edit panels for body
	bodyViewPanel := createPanel(" Body [VIEW] ", backgroundColor, borderColor, titleColor, foregroundColor)
	bodyEditPanel := createTextArea(" Body [EDIT] ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Set up vim-style navigation for body view panel (TextView)
	bodyViewPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			// Move left (scroll horizontally left)
			_, col := bodyViewPanel.GetScrollOffset()
			if col > 0 {
				bodyViewPanel.ScrollTo(0, col-1)
			}
			return nil
		case 'j':
			// Move down (scroll down one line)
			row, col := bodyViewPanel.GetScrollOffset()
			bodyViewPanel.ScrollTo(row+1, col)
			return nil
		case 'k':
			// Move up (scroll up one line)
			row, col := bodyViewPanel.GetScrollOffset()
			if row > 0 {
				bodyViewPanel.ScrollTo(row-1, col)
			}
			return nil
		case 'l':
			// Move right (scroll horizontally right)
			row, col := bodyViewPanel.GetScrollOffset()
			bodyViewPanel.ScrollTo(row, col+1)
			return nil
		case 'g':
			// Go to top (gg in vim, but simplified to single g)
			bodyViewPanel.ScrollToBeginning()
			return nil
		case 'G':
			// Go to bottom
			bodyViewPanel.ScrollToEnd()
			return nil
		case 'w':
			// Page down (like Ctrl+F in vim)
			return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
		case 'b':
			// Page up (like Ctrl+B in vim)
			return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
		}
		return event
	})

	// Create the tabbed interface for request data (Body, Auth, Query, Headers)
	requestDataTabs, tabPages, bodyContainer, tabHeader := createRequestDataTabs(bodyViewPanel, bodyEditPanel, backgroundColor, borderColor, titleColor, foregroundColor, activeTabColor)
	response := createPanel(" Response ", backgroundColor, borderColor, titleColor, foregroundColor)
	footer := createPanel("", backgroundColor, borderColor, titleColor, foregroundColor)
	footer.SetText(" [Tab] Cycle Focus | Tabs: [1234] Switch | Body: [i] Insert [Esc] Normal [hjkl] Nav | [F4] External Editor | [q] Quit | [n] New Collection | [r] New Request | [R] Rename | [m] Move Item | [d] Delete")

	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(methodUrlBar, 3, 0, false).
		AddItem(requestDataTabs, 0, 1, false).
		AddItem(response, 0, 1, false)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

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

	// Set up TreeView-specific input capture for h/l navigation
	collectionsTreeView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Vim-style collection navigation: 'h' to collapse, 'l' to expand
		if event.Rune() == 'h' || event.Rune() == 'l' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Check if this is a nested collection (has a parent that is also a collection)
					parentNode := findParentNode(rootNode, node)
					isNestedCollection := parentNode != nil && parentNode != rootNode

					if event.Rune() == 'h' {
						// Hierarchical collapse behavior: First collapse current, then move to parent
						if node.IsExpanded() {
							// First: collapse the current collection if it's expanded
							node.SetExpanded(false)
							node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
							node.ClearChildren()
							if config.C.UI.CollectionExpansion == "remember" {
								updateCollectionExpansionState(&collectionsData, col.Name, false)
							}
							return nil
						} else if isNestedCollection {
							// Second: if current is already collapsed, move to parent collection
							collectionsTreeView.SetCurrentNode(parentNode)
							return nil
						}
						// If it's a root collection and already collapsed, do nothing
					} else if event.Rune() == 'l' && !node.IsExpanded() {
						// Expand collection (works for both root and nested collections)
						node.SetExpanded(true)
						node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
						if len(node.GetChildren()) == 0 {
							addChildrenToCollectionNode(node, col)
						}
						if config.C.UI.CollectionExpansion == "remember" {
							updateCollectionExpansionState(&collectionsData, col.Name, true)
						}
						return nil
					}
				} else if _, ok := node.GetReference().(workspace.Request); ok {
					// Handle request navigation - 'h' collapses parent collection, 'l' opens/selects request
					if event.Rune() == 'h' {
						parentNode := findParentNode(rootNode, node)
						if parentNode != nil {
							if col, ok := parentNode.GetReference().(workspace.Collection); ok && parentNode.IsExpanded() {
								// Collapse parent collection
								parentNode.SetExpanded(false)
								parentNode.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
								parentNode.ClearChildren()
								if config.C.UI.CollectionExpansion == "remember" {
									updateCollectionExpansionState(&collectionsData, col.Name, false)
								}
								// Move selection to the parent collection
								collectionsTreeView.SetCurrentNode(parentNode)
								return nil
							}
						}
					} else if event.Rune() == 'l' {
						// 'l' on a request opens/selects it (trigger the selection function)
						// Simulate pressing Enter on the request to open it
						return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
					}
				}
			}
		}
		return event
	})

	var lastSelectedRequestNode *tview.TreeNode
	var currentSelectedNode *tview.TreeNode
	var currentRequest *workspace.Request
	var originalRequestSnapshot workspace.Request // Store original state for matching
	var currentRequestCollectionIndex int
	var currentRequestIndex int
	var programmaticallyUpdatingMethod bool // Track programmatic updates
	var programmaticallyUpdatingURL bool    // Track programmatic URL updates

	// Track body editing mode and current body content
	var bodyEditMode bool = false
	var currentBodyContent string = ""

	// Helper function to save current request changes
	saveCurrentRequest := func() {
		if currentRequest != nil && currentRequestCollectionIndex >= 0 && currentRequestIndex >= 0 {
			if err := workspace.SaveCollections(collectionsData); err != nil {
				// Handle error (could show in status or log)
				return
			}
		}
	}

	// Helper function to sync body content between view and edit panels
	syncBodyContent := func(content string) {
		currentBodyContent = content
		if bodyEditMode {
			bodyEditPanel.SetText(content, false)
		} else {
			bodyViewPanel.Clear()
			if content != "" {
				// Format content with syntax highlighting
				formattedContent := formatBodyContent(content)

				// Set content with proper handling
				bodyViewPanel.SetText(formattedContent)
				bodyViewPanel.SetTextAlign(tview.AlignLeft)

				// Force proper rendering and scrolling
				go func() {
					// Small delay to ensure content is set before scrolling
					bodyViewPanel.ScrollToEnd()
					bodyViewPanel.ScrollToBeginning()
				}()
			} else {
				bodyViewPanel.SetText("")
				bodyViewPanel.SetTextAlign(tview.AlignLeft)
			}
		}
	}

	// Helper function to switch between view and edit modes (will be redefined later with focus handling)
	var switchBodyMode func()

	switchBodyMode = func() {
		bodyEditMode = !bodyEditMode
		bodyContainer.Clear()

		if bodyEditMode {
			// Switch to edit mode
			bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
			bodyEditPanel.SetText(currentBodyContent, false)
		} else {
			// Switch to view mode
			bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
			// Update body content from edit panel if we were editing
			if currentRequest != nil {
				currentBodyContent = bodyEditPanel.GetText()
				currentRequest.Body = currentBodyContent
				if currentSelectedNode != nil {
					currentSelectedNode.SetReference(*currentRequest)
					saveCurrentRequest()
				}
			}
			syncBodyContent(currentBodyContent)
		}
	}

	// Helper function to sync method dropdown with current request
	syncMethodDropdown := func() {
		if currentRequest != nil {
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			for i, method := range methods {
				if method == currentRequest.Method {
					// Set flag to prevent the SetSelectedFunc from firing
					programmaticallyUpdatingMethod = true
					methodDropdown.SetCurrentOption(i)
					programmaticallyUpdatingMethod = false
					break
				}
			}
		}
	}

	// Add change handler for method dropdown
	methodDropdown.SetSelectedFunc(func(text string, index int) {
		// Skip if we're programmatically updating from tree selection
		if programmaticallyUpdatingMethod {
			return
		}

		if currentRequest != nil && currentSelectedNode != nil {
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
			if index >= 0 && index < len(methods) {
				newMethod := methods[index]
				if newMethod != currentRequest.Method {
					currentRequest.Method = newMethod

					// Update the node's reference with the new request data
					currentSelectedNode.SetReference(*currentRequest)

					// Update the tree node text to reflect the new method
					coloredMethod := getColoredMethod(currentRequest.Method)
					paddedName := padNameToMinLength(currentRequest.Name, 4)
					iconColor := config.C.UI.SelectedRequestIconColor
					coloredIcon := fmt.Sprintf("[%s]%s[-:-:-]", iconColor, config.C.UI.SelectedRequestIcon)
					currentSelectedNode.SetText(fmt.Sprintf("%s%s%s", coloredIcon, coloredMethod, paddedName))

					saveCurrentRequest()
				}
			}
		}
	})

	// Add change handler for URL input
	urlInput.SetChangedFunc(func(text string) {
		// Skip if we're programmatically updating from tree selection
		if programmaticallyUpdatingURL {
			return
		}

		if currentRequest != nil && currentSelectedNode != nil {
			currentRequest.URL = text

			// Update the node's reference with the new request data
			currentSelectedNode.SetReference(*currentRequest)

			saveCurrentRequest()
		}
	})

	// Add change handler for body text area (only for edit mode)
	bodyEditPanel.SetChangedFunc(func() {
		if bodyEditMode && currentRequest != nil && currentSelectedNode != nil {
			currentRequest.Body = bodyEditPanel.GetText()
			currentBodyContent = currentRequest.Body

			// Update the node's reference with the new request data
			currentSelectedNode.SetReference(*currentRequest)

			saveCurrentRequest()
		}
	})

	collectionsTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(selectionBackgroundColor).Foreground(foregroundColor))

		// Remove icon from previously selected request only when another request is selected
		if lastSelectedRequestNode != nil && lastSelectedRequestNode != node {
			if reference := lastSelectedRequestNode.GetReference(); reference != nil {
				if req, ok := reference.(workspace.Request); ok {
					// Only remove icon if the new selection is also a request
					if newReference := node.GetReference(); newReference != nil {
						if _, isRequest := newReference.(workspace.Request); isRequest {
							coloredMethod := getColoredMethod(req.Method)
							paddedName := padNameToMinLength(req.Name, 4)
							// Restore to reserved space (icon width + fixed space)
							iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon) // + 1
							spacePadding := strings.Repeat(" ", iconWidth)
							lastSelectedRequestNode.SetText(fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName))
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
			programmaticallyUpdatingURL = true
			urlInput.SetText(req.URL)
			programmaticallyUpdatingURL = false

			syncBodyContent(req.Body)

			// Set current request for persistence
			currentSelectedNode = node
			lastSelectedRequestNode = node

			// Store original request snapshot for accurate matching
			originalRequestSnapshot = req

			// Find the request in collectionsData to get indices for saving
			// Match by all original fields (Name, Method, URL, Body)
			currentRequestCollectionIndex = -1
			currentRequestIndex = -1
			for i, collection := range collectionsData {
				for j, collectionReq := range collection.Requests {
					// Match by all original request fields for uniqueness
					if collectionReq.Name == originalRequestSnapshot.Name &&
						collectionReq.Method == originalRequestSnapshot.Method &&
						collectionReq.URL == originalRequestSnapshot.URL &&
						collectionReq.Body == originalRequestSnapshot.Body {
						currentRequestCollectionIndex = i
						currentRequestIndex = j
						// Update the reference to point to the actual data
						currentRequest = &collectionsData[i].Requests[j]
						break
					}
				}
				if currentRequestCollectionIndex >= 0 {
					break
				}
			}

			// Set method in dropdown AFTER currentRequest is set
			syncMethodDropdown()
		} else if col, ok := reference.(workspace.Collection); ok {
			// Clear current request when a collection is selected
			currentRequest = nil
			currentSelectedNode = nil
			currentRequestCollectionIndex = -1
			currentRequestIndex = -1

			// Handle collection expansion
			expanded := !node.IsExpanded()
			node.SetExpanded(expanded)

			if expanded {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
				if len(node.GetChildren()) == 0 {
					addChildrenToCollectionNode(node, col)
				}
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&collectionsData, col.Name, true)
				}
			} else {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
				node.ClearChildren()
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&collectionsData, col.Name, false)
				}
			}
		}

		// Track the last selected request node
		if reference := node.GetReference(); reference != nil {
			if _, ok := reference.(workspace.Request); ok {
				lastSelectedRequestNode = node
			}
		}
	})

	// For "closed" mode, we need to ensure all collections start collapsed
	// For "remember" mode, the expansion state is handled within addCollectionsToTree
	if config.C.UI.CollectionExpansion == "closed" {
		// Temporarily set config to "closed" for tree building, then restore
		originalExpansion := config.C.UI.CollectionExpansion
		config.C.UI.CollectionExpansion = "closed"
		addCollectionsToTree(collectionsData, rootNode)
		config.C.UI.CollectionExpansion = originalExpansion
	} else {
		addCollectionsToTree(collectionsData, rootNode)
	}

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(collectionsTreeView, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	currentFocus := 0
	panels := []tview.Primitive{collectionsTreeView, methodUrlBar, requestDataTabs, response}

	// Custom focus handler that knows about special containers
	setPanelFocus := func(panelIndex int, focused bool) {
		if panelIndex == 1 { // Method+URL bar index
			setMethodUrlBarFocusStyle(methodUrlBar, focused, borderColor, borderFocusColor)
		} else if panelIndex == 2 { // Request data tabs index (Body/Auth/Query/Headers)
			setBodyContainerFocusStyle(bodyViewPanel, bodyEditPanel, bodyEditMode, focused, borderColor, borderFocusColor)
		} else {
			setFocusStyle(panels[panelIndex], focused, borderColor, borderFocusColor)
		}
	}

	// Update the original switchBodyMode with proper focus handling
	switchBodyMode = func() {
		bodyEditMode = !bodyEditMode
		bodyContainer.Clear()

		if bodyEditMode {
			// Switch to edit mode
			bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
			bodyEditPanel.SetText(currentBodyContent, false)
		} else {
			// Switch to view mode
			bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
			// Update body content from edit panel if we were editing
			if currentRequest != nil {
				currentBodyContent = bodyEditPanel.GetText()
				currentRequest.Body = currentBodyContent
				if currentSelectedNode != nil {
					currentSelectedNode.SetReference(*currentRequest)
					saveCurrentRequest()
				}
			}
			syncBodyContent(currentBodyContent)
		}

		// Update focus styling for the current mode if body container is focused
		if currentFocus == 2 { // Body container index (now at index 2 instead of 3)
			setPanelFocus(currentFocus, true)
			// Also switch the actual focus to the newly active panel
			if bodyEditMode {
				app.SetFocus(bodyEditPanel)
			} else {
				app.SetFocus(bodyViewPanel)
			}
		}
	}

	// Now that switchBodyMode is defined, set up the input capture for bodyEditPanel
	bodyEditPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Esc key to exit insert mode back to view mode
		if event.Key() == tcell.KeyEscape {
			switchBodyMode() // Switch back to view mode
			return nil
		}

		// Use Ctrl+hjkl for vim navigation in edit mode to avoid interfering with typing
		if event.Modifiers()&tcell.ModCtrl != 0 {
			switch event.Rune() {
			case 'h':
				// Move cursor left
				return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
			case 'j':
				// Move cursor down
				return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
			case 'k':
				// Move cursor up
				return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
			case 'l':
				// Move cursor right
				return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
			}
		}
		return event
	})

	setPanelFocus(currentFocus, true)

	// Create pages for modals
	pages := tview.NewPages()

	// Add main page
	pages.AddPage("main", grid, true, true)

	// Helper function to center modals
	createModal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, false).
			AddItem(nil, 0, 1, false)
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're on a modal page (not main)
		currentPage, _ := pages.GetFrontPage()
		if currentPage != "main" {
			// Let the form handle its own input
			return event
		}

		if event.Rune() == 'q' || event.Rune() == 'Q' {
			// Save expansion state before quitting if in "remember" mode
			if config.C.UI.CollectionExpansion == "remember" {
				if err := workspace.SaveExpansionState(collectionsData); err != nil {
					fmt.Printf("Warning: Failed to save expansion state: %v\n", err)
				}
			}
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			setPanelFocus(currentFocus, false)
			currentFocus = (currentFocus + 1) % len(panels)
			if currentFocus == 1 { // Method+URL bar
				app.SetFocus(urlInput) // Focus the URL input by default
			} else if currentFocus == 2 { // Request data tabs (Body/Auth/Query/Headers)
				// Focus the active panel within the body tab (default tab)
				if bodyEditMode {
					app.SetFocus(bodyEditPanel)
				} else {
					app.SetFocus(bodyViewPanel)
				}
			} else {
				app.SetFocus(panels[currentFocus])
			}
			setPanelFocus(currentFocus, true)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			setPanelFocus(currentFocus, false)
			currentFocus = (currentFocus - 1 + len(panels)) % len(panels)
			if currentFocus == 1 { // Method+URL bar
				app.SetFocus(urlInput) // Focus the URL input by default
			} else if currentFocus == 2 { // Request data tabs (Body/Auth/Query/Headers)
				// Focus the active panel within the body tab (default tab)
				if bodyEditMode {
					app.SetFocus(bodyEditPanel)
				} else {
					app.SetFocus(bodyViewPanel)
				}
			} else {
				app.SetFocus(panels[currentFocus])
			}
			setPanelFocus(currentFocus, true)
			return nil
		}

		if event.Rune() == 'n' {
			// New collection - check if we're adding to root or nested
			node := collectionsTreeView.GetCurrentNode()
			var parentCollection *workspace.Collection = nil
			var parentNode *tview.TreeNode = rootNode

			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Find the actual collection in the data to get a proper pointer
					parentCollection = findCollectionByName(&collectionsData, col.Name)
					parentNode = node
				}
			}

			form := createCollectionForm(app, pages, &collectionsData, rootNode, parentNode, collectionsTreeView, parentCollection)
			modal := createModal(form, 40, 10)
			pages.AddPage("newCollection", modal, true, true)
			app.SetFocus(form)
			return nil
		}

		if event.Rune() == 'r' {
			// New request - check if a collection or request is selected
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				var selectedCollection *workspace.Collection

				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Collection is selected
					selectedCollection = &col
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Request is selected - find its parent collection
					selectedCollection = findParentCollectionOfRequest(&collectionsData, req.Name, req.Method, req.URL)
				}

				if selectedCollection != nil {
					form := createRequestForm(app, pages, selectedCollection, &collectionsData, rootNode, collectionsTreeView)
					modal := createModal(form, 60, 14)
					pages.AddPage("newRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Rename functionality (Shift+R)
		if event.Rune() == 'R' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Rename collection
					form := createRenameCollectionForm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 10)
					pages.AddPage("renameCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Rename request - need to find parent collection
					form := createRenameRequestForm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 10)
					pages.AddPage("renameRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Move collection/request functionality (M)
		if event.Rune() == 'm' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Move collection
					form := createMoveCollectionForm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 12)
					pages.AddPage("moveCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Move request
					form := createMoveRequestForm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView)
					modal := createModal(form, 40, 10)
					pages.AddPage("moveRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Delete functionality (d)
		if event.Rune() == 'd' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Delete collection with confirmation
					form := createDeleteCollectionConfirm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 50, 8)
					pages.AddPage("deleteCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Delete request with confirmation
					form := createDeleteRequestConfirm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 50, 8)
					pages.AddPage("deleteRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// F4 to open body in external editor
		if event.Key() == tcell.KeyF4 {
			if currentRequest != nil {
				// Suspend TUI to open external editor
				app.Suspend(func() {
					modifiedContent, err := openInExternalEditor(currentBodyContent)
					if err != nil {
						fmt.Printf("Error opening external editor: %v\n", err)
						fmt.Println("Press Enter to continue...")
						var dummy string
						fmt.Scanln(&dummy)
						return
					}

					// Update the body with modified content
					syncBodyContent(modifiedContent)
					if currentRequest != nil && currentSelectedNode != nil {
						currentRequest.Body = modifiedContent
						currentSelectedNode.SetReference(*currentRequest)
						saveCurrentRequest()
					}
				})
			}
			return nil
		}

		// Tab switching with number keys (1-4) when request data tabs are focused
		if currentFocus == 2 { // Request data tabs focused
			switch event.Rune() {
			case '1':
				tabPages.SwitchToPage("body")
				updateTabHeader(tabHeader, 0, backgroundColor, foregroundColor, activeTabColor)
				return nil
			case '2':
				tabPages.SwitchToPage("auth")
				updateTabHeader(tabHeader, 1, backgroundColor, foregroundColor, activeTabColor)
				return nil
			case '3':
				tabPages.SwitchToPage("query")
				updateTabHeader(tabHeader, 2, backgroundColor, foregroundColor, activeTabColor)
				return nil
			case '4':
				tabPages.SwitchToPage("headers")
				updateTabHeader(tabHeader, 3, backgroundColor, foregroundColor, activeTabColor)
				return nil
			}
		}

		// Vim-style modal editing: 'i' to enter insert mode (only when body is focused and in view mode)
		if event.Rune() == 'i' && currentFocus == 2 && !bodyEditMode {
			switchBodyMode() // Switch to edit mode
			return nil
		}

		return event
	})

	if err := app.SetRoot(pages, true).SetFocus(collectionsTreeView).Run(); err != nil {
		panic(err)
	}
}
