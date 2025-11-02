package cmd

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

// KeyBinding represents a keyboard shortcut and its associated action
type KeyBinding struct {
	Key         tcell.Key
	Rune        rune
	Modifiers   tcell.ModMask
	Action      func(*UIOrchestrator, *tcell.EventKey) *tcell.EventKey
	Description string
	Context     string // "global", "body_view", "tree_view", "body_edit"
}

// KeyBindingManager manages all keybindings for the application
type KeyBindingManager struct {
	globalBindings   []KeyBinding
	bodyViewBindings []KeyBinding
	treeViewBindings []KeyBinding
	bodyEditBindings []KeyBinding
}

// NewKeyBindingManager creates a new keybinding manager with all default bindings
func NewKeyBindingManager() *KeyBindingManager {
	manager := &KeyBindingManager{}

	// Global keybindings (main app level)
	manager.globalBindings = []KeyBinding{
		{
			Rune:        'q',
			Action:      quitApp,
			Description: "Quit application",
			Context:     "global",
		},
		{
			Rune:        'Q',
			Action:      quitApp,
			Description: "Quit application",
			Context:     "global",
		},
		{
			Key:         tcell.KeyTab,
			Action:      handleTabNavigationAction,
			Description: "Cycle focus forward",
			Context:     "global",
		},
		{
			Key:         tcell.KeyBacktab,
			Action:      handleBacktabNavigationAction,
			Description: "Cycle focus backward",
			Context:     "global",
		},
		{
			Rune:        'n',
			Action:      newCollection,
			Description: "Create new collection",
			Context:     "global",
		},
		{
			Rune:        'r',
			Action:      newRequest,
			Description: "Create new request",
			Context:     "global",
		},
		{
			Rune:        'R',
			Action:      renameItem,
			Description: "Rename collection/request",
			Context:     "global",
		},
		{
			Rune:        'm',
			Action:      moveItem,
			Description: "Move collection/request",
			Context:     "global",
		},
		{
			Rune:        'd',
			Action:      deleteItem,
			Description: "Delete collection/request",
			Context:     "global",
		},
		{
			Key:         tcell.KeyF4,
			Action:      openExternalEditor,
			Description: "Open body in external editor",
			Context:     "global",
		},
		{
			Rune:        '1',
			Action:      switchToBodyTab,
			Description: "Switch to Body tab",
			Context:     "global",
		},
		{
			Rune:        '2',
			Action:      switchToAuthTab,
			Description: "Switch to Auth tab",
			Context:     "global",
		},
		{
			Rune:        '3',
			Action:      switchToQueryTab,
			Description: "Switch to Query tab",
			Context:     "global",
		},
		{
			Rune:        '4',
			Action:      switchToHeadersTab,
			Description: "Switch to Headers tab",
			Context:     "global",
		},
		{
			Rune:        'i',
			Action:      enterInsertMode,
			Description: "Enter insert mode (edit body)",
			Context:     "global",
		},
		{
			Key:         tcell.KeyLeft,
			Action:      navigateTabLeft,
			Description: "Navigate to previous tab",
			Context:     "global",
		},
		{
			Key:         tcell.KeyRight,
			Action:      navigateTabRight,
			Description: "Navigate to next tab",
			Context:     "global",
		},
	}

	// Body view panel keybindings (vim-style navigation)
	manager.bodyViewBindings = []KeyBinding{
		{
			Rune:        'h',
			Action:      scrollBodyLeft,
			Description: "Scroll body left",
			Context:     "body_view",
		},
		{
			Rune:        'j',
			Action:      scrollBodyDown,
			Description: "Scroll body down",
			Context:     "body_view",
		},
		{
			Rune:        'k',
			Action:      scrollBodyUp,
			Description: "Scroll body up",
			Context:     "body_view",
		},
		{
			Rune:        'l',
			Action:      scrollBodyRight,
			Description: "Scroll body right",
			Context:     "body_view",
		},
		{
			Rune:        'g',
			Action:      scrollBodyToTop,
			Description: "Scroll body to top",
			Context:     "body_view",
		},
		{
			Rune:        'G',
			Action:      scrollBodyToBottom,
			Description: "Scroll body to bottom",
			Context:     "body_view",
		},
		{
			Rune:        'w',
			Action:      pageBodyDown,
			Description: "Page body down",
			Context:     "body_view",
		},
		{
			Rune:        'b',
			Action:      pageBodyUp,
			Description: "Page body up",
			Context:     "body_view",
		},
	}

	// Tree view keybindings (collection navigation)
	manager.treeViewBindings = []KeyBinding{
		{
			Rune:        'h',
			Action:      collapseOrMoveToParent,
			Description: "Collapse collection or move to parent",
			Context:     "tree_view",
		},
		{
			Rune:        'l',
			Action:      expandOrSelectRequest,
			Description: "Expand collection or select request",
			Context:     "tree_view",
		},
	}

	// Body edit panel keybindings
	manager.bodyEditBindings = []KeyBinding{
		{
			Key:         tcell.KeyEscape,
			Action:      exitInsertMode,
			Description: "Exit insert mode",
			Context:     "body_edit",
		},
		{
			Rune:        'h',
			Modifiers:   tcell.ModCtrl,
			Action:      moveCursorLeft,
			Description: "Move cursor left",
			Context:     "body_edit",
		},
		{
			Rune:        'j',
			Modifiers:   tcell.ModCtrl,
			Action:      moveCursorDown,
			Description: "Move cursor down",
			Context:     "body_edit",
		},
		{
			Rune:        'k',
			Modifiers:   tcell.ModCtrl,
			Action:      moveCursorUp,
			Description: "Move cursor up",
			Context:     "body_edit",
		},
		{
			Rune:        'l',
			Modifiers:   tcell.ModCtrl,
			Action:      moveCursorRight,
			Description: "Move cursor right",
			Context:     "body_edit",
		},
		{
			Rune:        's',
			Modifiers:   tcell.ModCtrl,
			Action:      saveBodyContent,
			Description: "Save body content",
			Context:     "body_edit",
		},
	}

	return manager
}

// HandleKeyEvent processes a key event for the given context
func (kbm *KeyBindingManager) HandleKeyEvent(ui *UIOrchestrator, event *tcell.EventKey, context string) *tcell.EventKey {
	var bindings []KeyBinding

	switch context {
	case "global":
		bindings = kbm.globalBindings
	case "body_view":
		bindings = kbm.bodyViewBindings
	case "tree_view":
		bindings = kbm.treeViewBindings
	case "body_edit":
		bindings = kbm.bodyEditBindings
	default:
		return event
	}

	for _, binding := range bindings {
		if binding.Matches(event) {
			return binding.Action(ui, event)
		}
	}

	return event
}

// Matches checks if a keybinding matches the given event
func (kb *KeyBinding) Matches(event *tcell.EventKey) bool {
	if kb.Key != 0 && event.Key() == kb.Key {
		return event.Modifiers() == kb.Modifiers
	}
	if kb.Rune != 0 && event.Rune() == kb.Rune {
		return event.Modifiers() == kb.Modifiers
	}
	return false
}

// GetAllKeyBindings returns all keybindings for documentation purposes
func (kbm *KeyBindingManager) GetAllKeyBindings() []KeyBinding {
	var all []KeyBinding
	all = append(all, kbm.globalBindings...)
	all = append(all, kbm.bodyViewBindings...)
	all = append(all, kbm.treeViewBindings...)
	all = append(all, kbm.bodyEditBindings...)
	return all
}

// Action function implementations

// Global actions
func quitApp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Save expansion state before quitting if in "remember" mode
	if config.C.UI.CollectionExpansion == "remember" {
		if err := workspace.SaveExpansionState(*ui.CollectionsData); err != nil {
			// Could log error but for now just continue
		}
	}
	ui.App.Stop()
	return nil
}

func handleTabNavigationAction(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return handleTabNavigation(ui, event)
}

func handleBacktabNavigationAction(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return handleBacktabNavigation(ui, event)
}

func newCollection(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.CollectionsIndex {
		form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView)
		modal := createModal(form, 50, 12, tcell.ColorDefault)
		ui.Pages.AddPage("newCollection", modal, true, true)
		ui.App.SetFocus(form)
		return nil
	}
	return event
}

func newRequest(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.CollectionsIndex {
		// New request - check if a collection or request is selected
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			var selectedCollection *workspace.Collection

			if col, ok := node.GetReference().(workspace.Collection); ok {
				// Collection is selected
				selectedCollection = &col
			} else if req, ok := node.GetReference().(workspace.Request); ok {
				// Request is selected - find its parent collection
				selectedCollection = findParentCollectionOfRequest(ui.CollectionsData, req.Name, req.Method, req.URL)
			}

			if selectedCollection != nil {
				form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView)
				modal := createModal(form, 60, 14, tcell.ColorDefault)
				ui.Pages.AddPage("newRequest", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			}
		}
	}
	return event
}

func renameItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.CollectionsIndex {
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			reference := node.GetReference()

			if col, ok := reference.(workspace.Collection); ok {
				// Rename collection
				form := createRenameCollectionForm(ui.App, ui.Pages, &col, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView, node)
				modal := createModal(form, 25, 10, tcell.ColorDefault)
				ui.Pages.AddPage("renameCollection", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			} else if req, ok := reference.(workspace.Request); ok {
				// Rename request - need to find parent collection
				form := createRenameRequestForm(ui.App, ui.Pages, &req, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView, node)
				modal := createModal(form, 47, 10, tcell.ColorDefault)
				ui.Pages.AddPage("renameRequest", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			}
		}
	} else if ui.MainCycle.current == ui.EnviromentIndex {
		// Rename environment
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
			// Get the selected environment
			env := &(*ui.EnvironmentsData)[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0

			// Create a simple rename form
			form := createRenameEnvironmentForm(ui.App, ui.Pages, env, ui.EnvironmentsData, ui.EnvDropdown)
			modal := createModal(form, 30, 8, tcell.ColorDefault)
			ui.Pages.AddPage("renameEnvironment", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}
	}
	return event
}

func moveItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.CollectionsIndex {
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			if col, ok := node.GetReference().(workspace.Collection); ok {
				// Move collection
				form := createMoveCollectionForm(ui.App, ui.Pages, &col, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView, node)
				modal := createModal(form, 40, 12, tcell.ColorDefault)
				ui.Pages.AddPage("moveCollection", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			} else if req, ok := node.GetReference().(workspace.Request); ok {
				// Move request
				form := createMoveRequestForm(ui.App, ui.Pages, &req, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView)
				modal := createModal(form, 40, 10, tcell.ColorDefault)
				ui.Pages.AddPage("moveRequest", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			}
		}
	}
	return event
}

func deleteItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.CollectionsIndex {
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			reference := node.GetReference()

			if col, ok := reference.(workspace.Collection); ok {
				// Delete collection with confirmation
				form := createDeleteCollectionConfirm(ui.App, ui.Pages, &col, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView, node)
				modal := createModal(form, 50, 8, tcell.ColorDefault)
				ui.Pages.AddPage("deleteCollection", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			} else if req, ok := reference.(workspace.Request); ok {
				// Delete request with confirmation
				form := createDeleteRequestConfirm(ui.App, ui.Pages, &req, ui.CollectionsData, ui.RootNode, ui.CollectionsTreeView, node)
				modal := createModal(form, 50, 8, tcell.ColorDefault)
				ui.Pages.AddPage("deleteRequest", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			}
		}
	}
	return event
}

func openExternalEditor(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.RequestIndex {
		if ui.CurrentRequest != nil {
			// Suspend TUI to open external editor
			ui.App.Suspend(func() {
				modifiedContent, err := openInExternalEditor(ui.CurrentBodyContent)
				if err != nil {
					// Could show error but for now just continue
					return
				}

				// Update the body with modified content
				ui.SyncBodyContent(modifiedContent)
				if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
					ui.CurrentRequest.Body = modifiedContent
					ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
					saveCurrentRequest(ui.CurrentRequest, *ui.CollectionsData)
				}
			})
		}
		return nil
	}
	return event
}

func switchToBodyTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
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
			if row.KeyInput == currentFocusedElement || row.ValueInput == currentFocusedElement {
				isOnInputField = true
				break
			}
		}

		// Only switch tabs if not focused on an input field
		if !isOnInputField {
			ui.TabPages.SwitchToPage("body")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 0, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
			ui.CurrentTabIndex = 0
			return nil
		}
	}
	return event
}

func switchToAuthTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.URLBarIndex {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput == currentFocusedElement {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("auth")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 1, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
			ui.CurrentTabIndex = 1
			return nil
		}
	}
	return event
}

func switchToQueryTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.URLBarIndex {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput == currentFocusedElement {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("query")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 2, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
			ui.CurrentTabIndex = 2
			return nil
		}
	}
	return event
}

func switchToHeadersTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.URLBarIndex {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput == currentFocusedElement {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("headers")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 3, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
			ui.CurrentTabIndex = 3
			return nil
		}
	}
	return event
}

func enterInsertMode(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.RequestIndex && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
		ui.SwitchBodyMode() // Switch to edit mode
		ui.App.SetFocus(ui.BodyEditPanel)
		return nil
	}
	return event
}

func navigateTabLeft(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.RequestIndex && !ui.BodyEditMode {
		ui.CurrentTabIndex = (ui.CurrentTabIndex - 1 + 4) % 4
		tabNames := []string{"body", "auth", "query", "headers"}
		ui.TabPages.SwitchToPage(tabNames[ui.CurrentTabIndex])
		requestTabs := []string{"Body", "Auth", "Query", "Headers"}
		updateTabHeader(requestTabs, ui.TabHeader, ui.CurrentTabIndex, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
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
	return event
}

func navigateTabRight(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.RequestIndex && !ui.BodyEditMode {
		ui.CurrentTabIndex = (ui.CurrentTabIndex + 1) % 4
		tabNames := []string{"body", "auth", "query", "headers"}
		ui.TabPages.SwitchToPage(tabNames[ui.CurrentTabIndex])
		requestTabs := []string{"Body", "Auth", "Query", "Headers"}
		updateTabHeader(requestTabs, ui.TabHeader, ui.CurrentTabIndex, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault, tcell.ColorDefault)
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
	return event
}

// Body view actions
func scrollBodyLeft(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	row, col := ui.BodyViewPanel.GetScrollOffset()
	if col > 0 {
		ui.BodyViewPanel.ScrollTo(row, col-1)
	}
	return nil
}

func scrollBodyDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	row, col := ui.BodyViewPanel.GetScrollOffset()
	ui.BodyViewPanel.ScrollTo(row+1, col)
	return nil
}

func scrollBodyUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	row, col := ui.BodyViewPanel.GetScrollOffset()
	if row > 0 {
		ui.BodyViewPanel.ScrollTo(row-1, col)
	}
	return nil
}

func scrollBodyRight(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	row, col := ui.BodyViewPanel.GetScrollOffset()
	ui.BodyViewPanel.ScrollTo(row, col+1)
	return nil
}

func scrollBodyToTop(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.BodyViewPanel.ScrollToBeginning()
	return nil
}

func scrollBodyToBottom(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.BodyViewPanel.ScrollToEnd()
	return nil
}

func pageBodyDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyPgDn, 0, tcell.ModNone)
}

func pageBodyUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyPgUp, 0, tcell.ModNone)
}

// Tree view actions
func collapseOrMoveToParent(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node != nil {
		if col, ok := node.GetReference().(workspace.Collection); ok {
			// Check if this is a nested collection (has a parent that is also a collection)
			parentNode := findParentNode(ui.RootNode, node)
			isNestedCollection := parentNode != nil && parentNode != ui.RootNode

			if node.IsExpanded() {
				// First: collapse the current collection if it's expanded
				node.SetExpanded(false)
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
				node.ClearChildren()
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(ui.CollectionsData, col.Name, false)
				}
				return nil
			} else if isNestedCollection {
				// Second: if current is already collapsed, move to parent collection
				ui.CollectionsTreeView.SetCurrentNode(parentNode)
				return nil
			}
			// If it's a root collection and already collapsed, do nothing
		} else if _, ok := node.GetReference().(workspace.Request); ok {
			// Handle request navigation - 'h' collapses parent collection
			parentNode := findParentNode(ui.RootNode, node)
			if parentNode != nil {
				if col, ok := parentNode.GetReference().(workspace.Collection); ok && parentNode.IsExpanded() {
					// Collapse parent collection
					parentNode.SetExpanded(false)
					parentNode.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
					parentNode.ClearChildren()
					if config.C.UI.CollectionExpansion == "remember" {
						updateCollectionExpansionState(ui.CollectionsData, col.Name, false)
					}
					// Move selection to the parent collection
					ui.CollectionsTreeView.SetCurrentNode(parentNode)
					return nil
				}
			}
		}
	}
	return event
}

func expandOrSelectRequest(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node != nil {
		if col, ok := node.GetReference().(workspace.Collection); ok {
			if !node.IsExpanded() {
				// Expand collection
				node.SetExpanded(true)
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
				if len(node.GetChildren()) == 0 {
					addChildrenToCollectionNode(node, col)
				}
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(ui.CollectionsData, col.Name, true)
				}
				return nil
			}
		} else if _, ok := node.GetReference().(workspace.Request); ok {
			// 'l' on a request opens/selects it (trigger the selection function)
			// Simulate pressing Enter on the request to open it
			return tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
		}
	}
	return event
}

// Body edit actions
func exitInsertMode(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentTabIndex == 0 {
		ui.SwitchBodyMode() // Switch back to view mode
		return nil
	}
	return event
}

func moveCursorLeft(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
}

func moveCursorDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
}

func moveCursorUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
}

func moveCursorRight(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
}

func saveBodyContent(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentRequest != nil && ui.CurrentSelectedNode != nil {
		ui.CurrentBodyContent = ui.BodyEditPanel.GetText()
		ui.CurrentRequest.Body = ui.CurrentBodyContent
		ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
		saveCurrentRequest(ui.CurrentRequest, *ui.CollectionsData)
		// Show save confirmation could be added here
	}
	return nil
}
