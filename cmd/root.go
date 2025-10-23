// Package cmd provides the root command for the Petitorium CLI application.
// It includes the main TUI setup, UI components, and event handling.
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
	backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor, activeTabColor, buttonSelectedColor, dropdownFocusedBackgroundColor := setupTheme()

	collectionsData, err := workspace.LoadCollections()
	if err != nil {
		panic(fmt.Sprintf("Failed to load collections: %v", err))
	}

	app := tview.NewApplication().
		EnableMouse(true)

	// Create all UI components
	header,
		rootNode,
		methodURLBar,
		methodDropdown,
		urlInput,
		sendButton,
		bodyViewPanel,
		bodyEditPanel,
		response,
		footer,
		collectionsTreeView :=
		setupUIComponents(backgroundColor,
			foregroundColor,
			borderColor,
			borderFocusColor,
			titleColor,
			selectionBackgroundColor,
			activeTabColor,
			buttonSelectedColor,
			dropdownFocusedBackgroundColor,
		)

	// Variable declarations
	var currentSelectedNode *tview.TreeNode
	var currentRequest *workspace.Request
	var programmaticallyUpdatingMethod bool // Track programmatic updates
	var programmaticallyUpdatingURL bool    // Track programmatic URL updates
	var tabPages *tview.Pages
	var tabHeader *tview.Flex

	// Track current tab index (0=body, 1=auth, 2=query, 3=headers)
	currentTabIndex := 0

	// Additional UI variables
	var requestDataTabs *tview.Flex
	var bodyContainer *tview.Flex

	// Set up vim-style navigation for body view panel (TextView)
	bodyViewPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'h':
			// Move left (scroll horizontally left)
			row, col := bodyViewPanel.GetScrollOffset()
			if col > 0 {
				bodyViewPanel.ScrollTo(row, col-1)
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

	// Helper function to save current request changes
	saveCurrentRequest := func() {
		if currentRequest != nil {
			// Sync headers from UI before saving
			currentRequest.Headers = getHeadersFromUI()
			if err := workspace.SaveCollections(collectionsData); err != nil {
				// Handle error (could show in status or log)
				return
			}
		}
	}

	// Placeholder for panel focus setter - will be updated after setPanelFocus is defined
	var panelFocusSetter func(int, bool)
	// var currentFocusSetter func(int)

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
			saveCurrentRequest,
			func(p tview.Primitive) { app.SetFocus(p) },
			func(tabIndex int) { currentTabIndex = tabIndex },
			panelFocusSetter,
		)

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

	// Track body editing mode and current body content
	bodyEditMode := false
	currentBodyContent := ""

	// Create unified Request panel containing method+URL+send and tabs
	var requestPanel *tview.Flex
	var rightSide *tview.Flex

	// Create main panels
	mainPanels := []tview.Primitive{collectionsTreeView, methodURLBar, requestDataTabs, response}

	// Create request panel and right side layout
	requestPanel = setupRequestPanel(methodURLBar, requestDataTabs, backgroundColor)
	rightSide = setupRightSide(requestPanel, response)

	// Create main grid layout
	grid := setupLayout(header, footer, collectionsTreeView, rightSide)

	// Initialize cycles
	mainCycle = &MainCycle{
		panels:   mainPanels,
		current:  0, // start with collections
		children: nil,
	}

	requestCycle = &RequestCycle{
		elements: []tview.Primitive{methodDropdown, urlInput, sendButton},
		current:  0,
		parent:   mainCycle,
	}

	headersCycle = &HeadersCycle{
		inputs:   []tview.Primitive{},
		current:  0,
		parent:   mainCycle,
		children: nil,
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
			setHeadersInUI(req.Headers, saveCurrentRequest, func(p tview.Primitive) { app.SetFocus(p) })

			// Set current request for persistence
			currentSelectedNode = node
			lastSelectedRequestNode = node

			// Find the request pointer in collectionsData
			currentRequest = findRequestPtr(collectionsData, req)

			// Set method in dropdown AFTER currentRequest is set
			if currentRequest != nil {
				syncMethodDropdown()
			}
		} else if col, ok := reference.(workspace.Collection); ok {
			// Save current request headers before clearing
			if currentRequest != nil {
				currentRequest.Headers = getHeadersFromUI()
			}

			// Clear current request when a collection is selected
			currentRequest = nil
			currentSelectedNode = nil

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

	// Initial focus is on requestPanel (panels[1])
	currentFocus := 0

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

	// Update the panel focus setter now that setPanelFocus is defined
	panelFocusSetter = setPanelFocus

	// Update the original switchBodyMode with proper focus handling
	switchBodyMode = func() {
		bodyEditMode = !bodyEditMode
		bodyContainer.Clear()

		if bodyEditMode {
			// Switch to edit mode
			bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
			bodyEditPanel.SetText(currentBodyContent, false)
			bodyEditPanel.SetBorderColor(backgroundColor)
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

	// Now that switchBodyMode is defined, set up the input capture for bodyEditPanel
	bodyEditPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Esc key to exit insert mode back to view mode (only when in body tab)
		if event.Key() == tcell.KeyEscape && currentTabIndex == 0 {
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
			case 's':
				// Save without leaving edit mode (Ctrl+s)
				if currentRequest != nil && currentSelectedNode != nil {
					currentBodyContent = bodyEditPanel.GetText()
					currentRequest.Body = currentBodyContent
					currentSelectedNode.SetReference(*currentRequest)
					saveCurrentRequest()
					// Show save confirmation
					// footer.SetText(" ✓ Saved - Changes saved while staying in edit mode")
				}
				return nil
			}
		}
		return event
	})

	setPanelFocus(currentFocus, true)

	// Create pages for modals
	pages := tview.NewPages()

	// Add main page
	pages.AddPage("main", grid, true, true)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're on a modal page (not main)
		currentPage, _ := pages.GetFrontPage()
		if currentPage != "main" {
			// Let the form handle its own input
			return event
		}

		// When in body edit mode and focused on bodyEditPanel, pass all input through to allow pasting
		if bodyEditMode && app.GetFocus() == bodyEditPanel {
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

			// collections panel
			if mainCycle.current == 0 {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				// It can advance to the next panel
				nextElement := mainCycle.Next()
				requestCycle.current = 0
				setActiveBorder(nextElement)

				app.SetFocus(methodDropdown)
				currentFocus = mainCycle.current

				return nil
			}

			// urlbar panel & dropdown
			if mainCycle.current == 1 && requestCycle.current == 0 {
				next := requestCycle.Next()
				app.SetFocus(next)
				currentFocus = mainCycle.current

				return nil
			}

			// urlbar panel & url input
			if mainCycle.current == 1 && requestCycle.current == 1 {
				next := requestCycle.Next()
				app.SetFocus(next)
				currentFocus = mainCycle.current

				return nil
			}

			// urlbar panel & send button
			if mainCycle.current == 1 && requestCycle.current == 2 {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				nextElement := mainCycle.Next()
				setActiveBorder(nextElement)
				app.SetFocus(bodyViewPanel)
				currentFocus = mainCycle.current

				return nil
			}

			// requests editor/viewer panel
			if mainCycle.current == 2 && currentTabIndex == 0 && !bodyEditMode {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				nextElement := mainCycle.Next()
				setActiveBorder(nextElement)
				app.SetFocus(nextElement)
				currentFocus = mainCycle.current

				return nil
			}

			// requests headers panel - cycle through header inputs
			if mainCycle.current == 2 && currentTabIndex == 3 {
				// Cycle through header key/value/delete inputs, then jump to next panel
				currentFocusedElement := app.GetFocus()

				// Find current focused header input
				found := false
				for i, row := range currentHeaderRows {
					if row.KeyInput == currentFocusedElement {
						// Currently on key input, move to value input of same row
						app.SetFocus(row.ValueInput)
						found = true
						break
					} else if row.ValueInput == currentFocusedElement {
						// Currently on value input, move to delete button of same row
						app.SetFocus(row.DeleteButton)
						found = true
						break
					} else if row.DeleteButton == currentFocusedElement {
						// Currently on delete button, move to next row's key input or next panel
						if i < len(currentHeaderRows)-1 {
							// Move to next row's key input
							app.SetFocus(currentHeaderRows[i+1].KeyInput)
						} else {
							// Last delete button, move to next main panel (response)
							setInactiveBorder(mainCycle.panels[mainCycle.current])
							nextElement := mainCycle.Next()
							setActiveBorder(nextElement)
							app.SetFocus(nextElement)
						}
						found = true
						break
					}
				}

				// If no header input was focused, focus the first key input
				if !found && len(currentHeaderRows) > 0 {
					app.SetFocus(currentHeaderRows[0].KeyInput)
				}

				return nil
			}

			// response panel
			if mainCycle.current == 3 {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				nextElement := mainCycle.Next()
				setActiveBorder(nextElement)
				app.SetFocus(nextElement)
				currentFocus = mainCycle.current

				return nil
			}

			return nil
		}

		// Shift+KeyTab (Backtab) for backward navigation
		if event.Key() == tcell.KeyBacktab {
			// collections panel
			if mainCycle.current == 0 {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				prevElement := mainCycle.Prev()
				setActiveBorder(prevElement)
				app.SetFocus(prevElement)
				currentFocus = mainCycle.current
				return nil
			}

			// urlbar panel & dropdown
			if mainCycle.current == 1 && requestCycle.current == 0 {
				prev := requestCycle.Prev()
				if prev != nil {
					app.SetFocus(prev)
				} else {
					// Wrap to previous main panel
					setInactiveBorder(mainCycle.panels[mainCycle.current])
					prevElement := mainCycle.Prev()
					setActiveBorder(prevElement)
					app.SetFocus(prevElement)
					currentFocus = mainCycle.current
				}
				return nil
			}

			// urlbar panel & url input
			if mainCycle.current == 1 && requestCycle.current == 1 {
				prev := requestCycle.Prev()
				app.SetFocus(prev)
				currentFocus = mainCycle.current
				return nil
			}

			// urlbar panel & send button
			if mainCycle.current == 1 && requestCycle.current == 2 {
				prev := requestCycle.Prev()
				app.SetFocus(prev)
				currentFocus = mainCycle.current
				return nil
			}

			// requests editor/viewer panel
			if mainCycle.current == 2 && currentTabIndex == 0 && !bodyEditMode {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				prevElement := mainCycle.Prev()
				setActiveBorder(prevElement)
				app.SetFocus(sendButton)
				requestCycle.current = 2
				currentFocus = mainCycle.current
				return nil
			}

			// requests headers panel - cycle backward through header inputs
			if mainCycle.current == 2 && currentTabIndex == 3 {
				// Cycle backward through header key/value/delete inputs
				currentFocusedElement := app.GetFocus()

				// Find current focused header input
				found := false
				for i, row := range currentHeaderRows {
					if row.KeyInput == currentFocusedElement {
						// Currently on key input, move to previous row's delete button or previous panel
						if i > 0 {
							// Move to previous row's delete button
							app.SetFocus(currentHeaderRows[i-1].DeleteButton)
						} else {
							// First key input, move to previous main panel (request panel)
							setInactiveBorder(mainCycle.panels[mainCycle.current])
							prevElement := mainCycle.Prev()
							setActiveBorder(prevElement)
							app.SetFocus(sendButton)
							requestCycle.current = 2
						}
						found = true
						break
					} else if row.ValueInput == currentFocusedElement {
						// Currently on value input, move to key input of same row
						app.SetFocus(row.KeyInput)
						found = true
						break
					} else if row.DeleteButton == currentFocusedElement {
						// Currently on delete button, move to value input of same row
						app.SetFocus(row.ValueInput)
						found = true
						break
					}
				}

				// If no header input was focused, focus the last delete button
				if !found && len(currentHeaderRows) > 0 {
					app.SetFocus(currentHeaderRows[len(currentHeaderRows)-1].DeleteButton)
				}

				return nil
			}

			// response panel
			if mainCycle.current == 3 {
				setInactiveBorder(mainCycle.panels[mainCycle.current])
				prevElement := mainCycle.Prev()
				setActiveBorder(prevElement)
				app.SetFocus(bodyViewPanel)
				currentFocus = mainCycle.current
				return nil
			}

			return nil
		}

		if mainCycle.current == 0 && event.Rune() == 'n' {
			form := createCollectionFormWithLocation(app, pages, &collectionsData, rootNode, collectionsTreeView)
			modal := createModal(form, 50, 12)
			pages.AddPage("newCollection", modal, true, true)
			app.SetFocus(form)
			return nil
		}

		if mainCycle.current == 0 && event.Rune() == 'r' {
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
		if mainCycle.current == 0 && event.Rune() == 'R' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Rename collection
					form := createRenameCollectionForm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 25, 10)
					pages.AddPage("renameCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Rename request - need to find parent collection
					form := createRenameRequestForm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 47, 10)
					pages.AddPage("renameRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Move collection/request functionality (M)
		if mainCycle.current == 0 && event.Rune() == 'm' {
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
		if mainCycle.current == 0 && event.Rune() == 'd' {
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
		if mainCycle.current == 2 && event.Key() == tcell.KeyF4 {
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
		// Request data tabs focused and not on an input field
		if currentFocus == 2 {
			// Check if focus is on an input field (don't switch tabs if typing)
			currentFocusedElement := app.GetFocus()
			isOnInputField := false

			// Check if focused on method dropdown
			if currentFocusedElement == methodDropdown {
				isOnInputField = true
			}
			// Check if focused on URL input
			if currentFocusedElement == urlInput {
				isOnInputField = true
			}
			// Check if focused on body edit panel
			if currentFocusedElement == bodyEditPanel {
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
					tabPages.SwitchToPage("body")
					updateTabHeader(tabHeader, 0, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
					currentTabIndex = 0
					return nil
				case '2':
					tabPages.SwitchToPage("auth")
					updateTabHeader(tabHeader, 1, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
					currentTabIndex = 1
					return nil
				case '3':
					tabPages.SwitchToPage("query")
					updateTabHeader(tabHeader, 2, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
					currentTabIndex = 2
					return nil
				case '4':
					tabPages.SwitchToPage("headers")
					updateTabHeader(tabHeader, 3, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
					currentTabIndex = 3
					return nil
				}
			}
		}

		// Vim-style modal editing: 'i' to enter insert mode (only when request panel is active, tabs are focused, body tab is selected, and in view mode)
		// if event.Rune() == 'i' && currentFocus == 1 && currentRequestSubFocus == 3 && currentTabIndex == 0 && !bodyEditMode {
		if event.Rune() == 'i' && mainCycle.current == 2 && currentTabIndex == 0 && !bodyEditMode {
			switchBodyMode() // Switch to edit mode
			app.SetFocus(bodyEditPanel)
			return nil
		}

		// Arrow key navigation for tabs when request panel tabs are focused and not in body edit mode
		if mainCycle.current == 2 && !bodyEditMode {
			if event.Key() == tcell.KeyLeft {
				currentTabIndex = (currentTabIndex - 1 + 4) % 4
				tabNames := []string{"body", "auth", "query", "headers"}
				tabPages.SwitchToPage(tabNames[currentTabIndex])
				updateTabHeader(tabHeader, currentTabIndex, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
				// Focus the appropriate tab content
				switch currentTabIndex {
				case 0: // Body tab
					if bodyEditMode {
						app.SetFocus(bodyEditPanel)
					} else {
						app.SetFocus(bodyViewPanel)
					}
				case 3: // Headers tab
					if len(currentHeaderRows) > 0 && currentHeaderRows[0].KeyInput != nil {
						app.SetFocus(currentHeaderRows[0].KeyInput)
					} else {
						app.SetFocus(requestDataTabs)
					}
				default:
					app.SetFocus(requestDataTabs)
				}
				return nil
			} else if event.Key() == tcell.KeyRight {
				currentTabIndex = (currentTabIndex + 1) % 4
				tabNames := []string{"body", "auth", "query", "headers"}
				tabPages.SwitchToPage(tabNames[currentTabIndex])
				updateTabHeader(tabHeader, currentTabIndex, backgroundColor, foregroundColor, activeTabColor, selectionBackgroundColor)
				// Focus the appropriate tab content
				switch currentTabIndex {
				case 0: // Body tab
					if bodyEditMode {
						app.SetFocus(bodyEditPanel)
					} else {
						app.SetFocus(bodyViewPanel)
					}
				case 3: // Headers tab
					if len(currentHeaderRows) > 0 && currentHeaderRows[0].KeyInput != nil {
						app.SetFocus(currentHeaderRows[0].KeyInput)
					} else {
						app.SetFocus(requestDataTabs)
					}
				default:
					app.SetFocus(requestDataTabs)
				}
				return nil
			}
		}

		return event
	})

	// Add send button functionality
	sendButton.SetSelectedFunc(func() {
		// Get current request data from UI
		_, method := methodDropdown.GetCurrentOption()
		url := urlInput.GetText()
		body := ""
		if currentRequest != nil {
			body = currentRequest.Body
		}
		headers := getHeadersFromUI()

		// Validate URL
		if url == "" {
			response.SetText("Error: URL is required")
			return
		}

		// Send the request
		response.SetText("Sending request...")
		resp, err := SendRequest(method, url, body, headers)
		if err != nil {
			response.SetText(fmt.Sprintf("Error: %v", err))
			return
		}

		// Format and display the response
		formattedResponse := FormatResponse(resp)
		response.SetText(formattedResponse)
	})

	if err := app.
		SetRoot(pages, true).
		SetFocus(collectionsTreeView).
		Run(); err != nil {
		panic(err)
	}
}
