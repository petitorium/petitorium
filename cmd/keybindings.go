package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// KeyBinding represents a keyboard shortcut and its associated action
type KeyBinding struct {
	Key         tcell.Key
	Rune        rune
	Modifiers   tcell.ModMask
	Action      func(*UIOrchestrator, *tcell.EventKey) *tcell.EventKey
	Description string
	Context     string // "global", "body_view", "tree_view", "body_edit", "response_view"
}

// KeyBindingManager manages all keybindings for the application
type KeyBindingManager struct {
	globalBindings       []KeyBinding
	bodyViewBindings     []KeyBinding
	treeViewBindings     []KeyBinding
	bodyEditBindings     []KeyBinding
	responseViewBindings []KeyBinding
	modalBindings        []KeyBinding
}

// Tree navigation functions
var (
	navigateTreeDown func(*UIOrchestrator, *tcell.EventKey) *tcell.EventKey
	navigateTreeUp   func(*UIOrchestrator, *tcell.EventKey) *tcell.EventKey
)

// getVisibleNodes collects all visible nodes in the tree
func getVisibleNodes(root *tview.TreeNode) []*tview.TreeNode {
	var nodes []*tview.TreeNode
	var collect func(*tview.TreeNode)
	collect = func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		nodes = append(nodes, node)
		if node.IsExpanded() {
			for _, child := range node.GetChildren() {
				collect(child)
			}
		}
	}
	if root != nil {
		collect(root)
	}
	return nodes
}

// NewKeyBindingManager creates a new keybinding manager with all default bindings
func NewKeyBindingManager() *KeyBindingManager {
	manager := &KeyBindingManager{}

	// Modal keybindings (for modal dialogs)
	manager.modalBindings = []KeyBinding{
		{
			Key:         tcell.KeyEscape,
			Action:      closeModal,
			Description: "Close modal dialog",
			Context:     "modal",
		},
		{
			Rune:        'q',
			Action:      closeModal,
			Description: "Close modal dialog",
			Context:     "modal",
		},
		{
			Rune:        'Q',
			Action:      closeModal,
			Description: "Close modal dialog",
			Context:     "modal",
		},
	}

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
			Rune:        'N',
			Action:      newCollection,
			Description: "Create new collection",
			Context:     "global",
		},
		{
			Rune:        'n',
			Action:      newRequest,
			Description: "Create new request",
			Context:     "global",
		},
		{
			Rune:        'D',
			Action:      duplicateRequest,
			Description: "Duplicate request",
			Context:     "tree_view",
		},
		{
			Rune:        'r',
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
		// {
		// 	Key:         tcell.KeyDelete,
		// 	Action:      deleteItem,
		// 	Description: "Delete collection/request",
		// 	Context:     "global",
		// },
		{
			Rune:        'w',
			Modifiers:   tcell.ModCtrl,
			Action:      showWorkspaceMenu,
			Description: "Show workspace menu",
			Context:     "global",
		},
		{
			Key:         tcell.KeyF4,
			Action:      openExternalEditor,
			Description: "Open body in external editor",
			Context:     "global",
		},
		{
			Rune:        'i',
			Action:      enterInsertMode,
			Description: "Enter insert mode (edit body)",
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

	// Response view panel keybindings (vim-style navigation)
	manager.responseViewBindings = []KeyBinding{
		{
			Rune:        'j',
			Action:      scrollResponseDown,
			Description: "Scroll response down",
			Context:     "response_view",
		},
		{
			Rune:        'k',
			Action:      scrollResponseUp,
			Description: "Scroll response up",
			Context:     "response_view",
		},
		{
			Rune:        'g',
			Action:      scrollResponseToTop,
			Description: "Scroll response to top",
			Context:     "response_view",
		},
		{
			Rune:        'G',
			Action:      scrollResponseToBottom,
			Description: "Scroll response to bottom",
			Context:     "response_view",
		},
		{
			Rune:        'd',
			Action:      scrollResponseHalfPageDown,
			Description: "Scroll response down by half page",
			Context:     "response_view",
		},
		{
			Rune:        'u',
			Action:      scrollResponseHalfPageUp,
			Description: "Scroll response up by half page",
			Context:     "response_view",
		},
		{
			Rune:        'f',
			Action:      openResponseInFx,
			Description: "Open response in fx",
			Context:     "response_view",
		},
	}

	// Define tree navigation functions
	navigateTreeDown = func(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
		root := ui.CollectionsTreeView.GetRoot()
		if root == nil {
			return nil
		}
		visibleNodes := getVisibleNodes(root)
		current := ui.CollectionsTreeView.GetCurrentNode()
		for i, node := range visibleNodes {
			if node == current && i < len(visibleNodes)-1 {
				next := visibleNodes[i+1]
				ui.CollectionsTreeView.SetCurrentNode(next)
				if ui.TreeHighlightHandler != nil {
					ui.TreeHighlightHandler(next)
				}
				break
			}
		}
		return nil
	}

	navigateTreeUp = func(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
		root := ui.CollectionsTreeView.GetRoot()
		if root == nil {
			return nil
		}
		visibleNodes := getVisibleNodes(root)
		current := ui.CollectionsTreeView.GetCurrentNode()
		for i, node := range visibleNodes {
			if node == current && i > 0 {
				prev := visibleNodes[i-1]
				ui.CollectionsTreeView.SetCurrentNode(prev)
				if ui.TreeHighlightHandler != nil {
					ui.TreeHighlightHandler(prev)
				}
				break
			}
		}
		return nil
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
			Rune:        'j',
			Action:      navigateTreeDown,
			Description: "Navigate down in tree",
			Context:     "tree_view",
		},
		{
			Rune:        'k',
			Action:      navigateTreeUp,
			Description: "Navigate up in tree",
			Context:     "tree_view",
		},
		{
			Rune:        'l',
			Action:      expandOrSelectRequest,
			Description: "Expand collection or select request",
			Context:     "tree_view",
		},
		{
			Rune:        'g',
			Action:      navigateTreeToTop,
			Description: "Go to top of tree",
			Context:     "tree_view",
		},
		{
			Rune:        'G',
			Action:      navigateTreeToBottom,
			Description: "Go to bottom of tree",
			Context:     "tree_view",
		},
		// {
		// 	Rune:        'd',
		// 	Action:      navigateTreeHalfPageDown,
		// 	Description: "Scroll down half page",
		// 	Context:     "tree_view",
		// },
		// {
		// 	Rune:        'u',
		// 	Action:      navigateTreeHalfPageUp,
		// 	Description: "Scroll up half page",
		// 	Context:     "tree_view",
		// },
		{
			Rune:        'D',
			Action:      duplicateRequest,
			Description: "Duplicate selected request",
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
	case "response_view":
		bindings = kbm.responseViewBindings
	case "tree_view":
		bindings = kbm.treeViewBindings
	case "body_edit":
		bindings = kbm.bodyEditBindings
	case "modal":
		bindings = kbm.modalBindings
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
	all = append(all, kbm.modalBindings...)
	return all
}

// Action function implementations

// Global actions
func quitApp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Save expansion state before quitting if in "remember" mode
	if config.C.UI.CollectionExpansion == "remember" {
		if err := workspace.SaveExpansionState(&ui.WorkspaceData.Collections); err != nil {
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
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
		form := createCollectionFormWithLocation(ui.App, ui.Pages, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
		modal := createModal(form, 50, 12, tcell.ColorDefault)
		ui.Pages.AddPage("newCollection", modal, true, true)
		ui.App.SetFocus(form)
		return nil
	}
	return event
}

func newRequest(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
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
	return event
}

func duplicateRequest(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
		// Duplicate request - check if a request is selected
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			if req, ok := node.GetReference().(workspace.Request); ok {
				// Request is selected - find its parent collection
				selectedCollection := findParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.Name, req.Method, req.URL)

				if selectedCollection != nil {
					form := createDuplicateRequestForm(ui.App, ui.Pages, &req, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
					modal := createModal(form, 60, 15, tcell.ColorDefault)
					ui.Pages.AddPage("duplicateRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}
	}
	return event
}

func renameItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
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
	} else if ui.MainCycle.current == ui.PanelIndices.Environment {
		// Rename environment
		currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
		if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
			// Get the selected environment
			env := &(*ui.EnvironmentsData)[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0

			// Create a simple rename form
			form := createRenameEnvironmentForm(ui.App, ui.Pages, env, ui.EnvironmentsData, ui.WorkspaceData, ui.EnvDropdown, ui.EnvConfigButton, ui.Colors, ui.EnvConfigButton)
			modal := createModal(form, 30, 8, tcell.ColorDefault)
			ui.Pages.AddPage("renameEnvironment", modal, true, true)
			ui.App.SetFocus(form)
			return nil
		}
	}
	return event
}

// isInFormPopup checks if the current page is a form popup
func isInFormPopup(ui *UIOrchestrator) bool {
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
		"duplicateRequest",
		"cloneEnvironment",
		"createWorkspace",
		"deleteEnvironment",
		"deleteWorkspace",
		"duplicateWorkspace",
		"renameEnvironment",
		"renameWorkspace",
		"workspaceModal",
	}

	for _, popup := range formPopups {
		if currentPage == popup {
			return true
		}
	}
	return false
}

func moveItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
		node := ui.CollectionsTreeView.GetCurrentNode()
		if node != nil {
			if col, ok := node.GetReference().(workspace.Collection); ok {
				// Move collection
				form := createMoveCollectionForm(ui.App, ui.Pages, &col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
				modal := createModal(form, 40, 12, tcell.ColorDefault)
				ui.Pages.AddPage("moveCollection", modal, true, true)
				ui.App.SetFocus(form)
				return nil
			} else if _, ok := node.GetReference().(workspace.Request); ok {
				// Move request - use current request data instead of stale node reference
				if ui.CurrentRequest != nil {
					form := createMoveRequestForm(ui.App, ui.Pages, ui.CurrentRequest, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
					modal := createModal(form, 40, 10, tcell.ColorDefault)
					ui.Pages.AddPage("moveRequest", modal, true, true)
					ui.App.SetFocus(form)
					return nil
				}
			}
		}
	}
	return event
}

func deleteItem(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if isInFormPopup(ui) {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Collections {
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
	return event
}

func openExternalEditor(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.PanelIndices.Request {
		if ui.CurrentTabIndex == ui.RPBodyTabIndex {
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
						// Also save to JSONBodyContent if we're in JSON mode
						if ui.CurrentRequest.ContentType == "JSON" {
							ui.JSONBodyContent = modifiedContent
						}
						ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
						saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
					}
				})
			}
			return nil
		} else if ui.CurrentTabIndex == ui.RPHeadersTabIndex {
			// Bulk edit headers
			ui.App.Suspend(func() {
				// Get current headers
				headers := getHeadersFromUI()

				// Format headers as "Key: Value" one per line
				var headerLines []string
				for key, value := range headers {
					headerLines = append(headerLines, fmt.Sprintf("%s: %s", key, value))
				}
				headerContent := strings.Join(headerLines, "\n")

				// Open external editor
				modifiedContent, err := openInExternalEditor(headerContent)
				if err != nil {
					// Could show error but for now just continue
					return
				}

				// Parse the modified content back into headers
				lines := strings.Split(strings.TrimSpace(modifiedContent), "\n")
				newHeaders := make(map[string]string)
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					// Split on first colon
					parts := strings.SplitN(line, ":", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(parts[1])
						if key != "" {
							newHeaders[key] = value
						}
					}
				}

				// Update headers in UI
				setHeadersInUI(ui.Colors, newHeaders, func() {
					if ui.CurrentRequest != nil {
						ui.CurrentRequest.Headers = newHeaders
						if ui.CurrentSelectedNode != nil {
							ui.CurrentSelectedNode.SetReference(*ui.CurrentRequest)
							saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
						}
					}
				}, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter)
			})
			return nil
		}
	}
	return event
}

func switchToBodyTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.PanelIndices.Request {
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
			if row.KeyInput == currentFocusedElement || row.ValueInput.HasFocus() {
				isOnInputField = true
				break
			}
		}

		// Only switch tabs if not focused on an input field
		if !isOnInputField {
			ui.TabPages.SwitchToPage("body")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 0, ui.Colors)
			ui.CurrentTabIndex = 0
			ui.UpdateFooter()
			return nil
		}
	}
	return event
}

func switchToAuthTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.PanelIndices.Request {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput.HasFocus() {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("auth")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 1, ui.Colors)
			ui.CurrentTabIndex = 1
			ui.UpdateFooter()
			return nil
		}
	}
	return event
}

func switchToQueryTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.PanelIndices.Request {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput.HasFocus() {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("query")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 2, ui.Colors)
			ui.CurrentTabIndex = 2
			ui.UpdateFooter()
			return nil
		}
	}
	return event
}

func switchToHeadersTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.CurrentFocus == ui.PanelIndices.Request {
		currentFocusedElement := ui.App.GetFocus()
		isOnInputField := false

		if currentFocusedElement == ui.MethodDropdown || currentFocusedElement == ui.URLInput || currentFocusedElement == ui.BodyEditPanel {
			isOnInputField = true
		}
		for _, row := range currentHeaderRows {
			if row.KeyInput == currentFocusedElement || row.ValueInput.HasFocus() {
				isOnInputField = true
				break
			}
		}

		if !isOnInputField {
			ui.TabPages.SwitchToPage("headers")
			requestTabs := []string{"Body", "Auth", "Query", "Headers"}
			updateTabHeader(requestTabs, ui.TabHeader, 3, ui.Colors)
			ui.CurrentTabIndex = 3
			ui.UpdateFooter()
			return nil
		}
	}
	return event
}

func enterInsertMode(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Don't enter insert mode if content type is Multipart (user is editing fields)
	// or if content type is No Body (no body to edit)
	if ui.CurrentRequest != nil && (ui.CurrentRequest.ContentType == "Multipart" || ui.CurrentRequest.ContentType == "No Body") {
		return event
	}

	if ui.MainCycle.current == ui.PanelIndices.Request && ui.CurrentTabIndex == 0 && !ui.BodyEditMode {
		ui.SwitchBodyMode() // Switch to edit mode
		ui.App.SetFocus(ui.BodyEditPanel)
		return nil
	}
	return event
}

func navigateTabLeft(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.PanelIndices.Request && !ui.BodyEditMode {
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
		ui.UpdateFooter()
		return nil
	} else if ui.MainCycle.current == ui.PanelIndices.Response {
		// Navigate response tabs
		ui.CurrentResponseTabIndex = (ui.CurrentResponseTabIndex - 1 + 4) % 4
		responseTabNames := []string{"preview", "headers", "cookies", "timeline"}
		ui.ResponsePages.SwitchToPage(responseTabNames[ui.CurrentResponseTabIndex])
		updateResponseTabHeader(ui.ResponseTabHeader, ui.CurrentResponseTabIndex, ui.Colors)
		return nil
	}
	return event
}

func navigateTabRight(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.MainCycle.current == ui.PanelIndices.Request && !ui.BodyEditMode {
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
		ui.UpdateFooter()
		return nil
	} else if ui.MainCycle.current == ui.PanelIndices.Response {
		// Navigate response tabs
		ui.CurrentResponseTabIndex = (ui.CurrentResponseTabIndex + 1) % 4
		responseTabNames := []string{"preview", "headers", "cookies", "timeline"}
		ui.ResponsePages.SwitchToPage(responseTabNames[ui.CurrentResponseTabIndex])
		updateResponseTabHeader(ui.ResponseTabHeader, ui.CurrentResponseTabIndex, ui.Colors)
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

func scrollResponseDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Scroll the current response panel down
	return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
}

func scrollResponseUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Scroll the current response panel up
	return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
}

func scrollResponseToTop(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Get the current response page and scroll it to top
	currentPage, _ := ui.ResponsePages.GetFrontPage()
	switch currentPage {
	case "preview":
		ui.ResponsePreviewPanel.ScrollToBeginning()
	case "headers":
		if headersTable, ok := ui.ResponseHeadersPanel.(*tview.Table); ok {
			headersTable.ScrollToBeginning()
		}
	case "cookies":
		ui.ResponseCookiesPanel.ScrollToBeginning()
	case "timeline":
		ui.ResponseTimelinePanel.ScrollToBeginning()
	}
	return nil
}

func scrollResponseToBottom(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Get the current response page and scroll it to bottom
	currentPage, _ := ui.ResponsePages.GetFrontPage()
	switch currentPage {
	case "preview":
		ui.ResponsePreviewPanel.ScrollToEnd()
	case "headers":
		if headersTable, ok := ui.ResponseHeadersPanel.(*tview.Table); ok {
			headersTable.ScrollToEnd()
		}
	case "cookies":
		ui.ResponseCookiesPanel.ScrollToEnd()
	case "timeline":
		ui.ResponseTimelinePanel.ScrollToEnd()
	}
	return nil
}

func scrollResponseHalfPageDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Get the current response page and scroll it down by half page
	currentPage, _ := ui.ResponsePages.GetFrontPage()
	switch currentPage {
	case "preview":
		_, _, _, height := ui.ResponsePreviewPanel.GetRect()
		currentTop, _ := ui.ResponsePreviewPanel.GetScrollOffset()
		newTop := currentTop + height/2
		ui.ResponsePreviewPanel.ScrollTo(newTop, 0)
	case "headers":
		// For table, scroll to end as approximation for fast scroll
		if headersTable, ok := ui.ResponseHeadersPanel.(*tview.Table); ok {
			headersTable.ScrollToEnd()
		}
	case "cookies":
		_, _, _, height := ui.ResponseCookiesPanel.GetRect()
		currentTop, _ := ui.ResponseCookiesPanel.GetScrollOffset()
		newTop := currentTop + height/2
		ui.ResponseCookiesPanel.ScrollTo(newTop, 0)
	case "timeline":
		_, _, _, height := ui.ResponseTimelinePanel.GetRect()
		currentTop, _ := ui.ResponseTimelinePanel.GetScrollOffset()
		newTop := currentTop + height/2
		ui.ResponseTimelinePanel.ScrollTo(newTop, 0)
	}
	return nil
}

func scrollResponseHalfPageUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Get the current response page and scroll it up by half page
	currentPage, _ := ui.ResponsePages.GetFrontPage()
	switch currentPage {
	case "preview":
		_, _, _, height := ui.ResponsePreviewPanel.GetRect()
		currentTop, _ := ui.ResponsePreviewPanel.GetScrollOffset()
		newTop := currentTop - height/2
		if newTop < 0 {
			newTop = 0
		}
		ui.ResponsePreviewPanel.ScrollTo(newTop, 0)
	case "headers":
		// For table, scroll to beginning as approximation for fast scroll
		if headersTable, ok := ui.ResponseHeadersPanel.(*tview.Table); ok {
			headersTable.ScrollToBeginning()
		}
	case "cookies":
		_, _, _, height := ui.ResponseCookiesPanel.GetRect()
		currentTop, _ := ui.ResponseCookiesPanel.GetScrollOffset()
		newTop := currentTop - height/2
		if newTop < 0 {
			newTop = 0
		}
		ui.ResponseCookiesPanel.ScrollTo(newTop, 0)
	case "timeline":
		_, _, _, height := ui.ResponseTimelinePanel.GetRect()
		currentTop, _ := ui.ResponseTimelinePanel.GetScrollOffset()
		newTop := currentTop - height/2
		if newTop < 0 {
			newTop = 0
		}
		ui.ResponseTimelinePanel.ScrollTo(newTop, 0)
	}
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
					updateCollectionExpansionState(&ui.WorkspaceData.Collections, col.Name, false)
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
						updateCollectionExpansionState(&ui.WorkspaceData.Collections, col.Name, false)
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
					updateCollectionExpansionState(&ui.WorkspaceData.Collections, col.Name, true)
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

func navigateTreeToTop(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	root := ui.CollectionsTreeView.GetRoot()
	if root == nil {
		return nil
	}
	visibleNodes := getVisibleNodes(root)
	if len(visibleNodes) > 0 {
		ui.CollectionsTreeView.SetCurrentNode(visibleNodes[0])
		if ui.TreeHighlightHandler != nil {
			ui.TreeHighlightHandler(visibleNodes[0])
		}
	}
	return nil
}

func navigateTreeToBottom(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	root := ui.CollectionsTreeView.GetRoot()
	if root == nil {
		return nil
	}
	visibleNodes := getVisibleNodes(root)
	if len(visibleNodes) > 0 {
		lastNode := visibleNodes[len(visibleNodes)-1]
		ui.CollectionsTreeView.SetCurrentNode(lastNode)
		if ui.TreeHighlightHandler != nil {
			ui.TreeHighlightHandler(lastNode)
		}
	}
	return nil
}

func navigateTreeHalfPageDown(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	root := ui.CollectionsTreeView.GetRoot()
	if root == nil {
		return nil
	}
	visibleNodes := getVisibleNodes(root)
	current := ui.CollectionsTreeView.GetCurrentNode()
	_, _, _, height := ui.CollectionsTreeView.GetRect()
	offset := height / 2
	if offset == 0 {
		offset = 1
	}

	for i, node := range visibleNodes {
		if node == current {
			newIndex := i + offset
			if newIndex >= len(visibleNodes) {
				newIndex = len(visibleNodes) - 1
			}
			next := visibleNodes[newIndex]
			ui.CollectionsTreeView.SetCurrentNode(next)
			if ui.TreeHighlightHandler != nil {
				ui.TreeHighlightHandler(next)
			}
			break
		}
	}
	return nil
}

func navigateTreeHalfPageUp(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	root := ui.CollectionsTreeView.GetRoot()
	if root == nil {
		return nil
	}
	visibleNodes := getVisibleNodes(root)
	current := ui.CollectionsTreeView.GetCurrentNode()
	_, _, _, height := ui.CollectionsTreeView.GetRect()
	offset := height / 2
	if offset == 0 {
		offset = 1
	}

	for i, node := range visibleNodes {
		if node == current {
			newIndex := i - offset
			if newIndex < 0 {
				newIndex = 0
			}
			prev := visibleNodes[newIndex]
			ui.CollectionsTreeView.SetCurrentNode(prev)
			if ui.TreeHighlightHandler != nil {
				ui.TreeHighlightHandler(prev)
			}
			break
		}
	}
	return nil
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
		saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
		// Show save confirmation could be added here
	}
	return nil
}

// Modal actions
func closeModal(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// This function is called when modal keybindings are triggered
	// The actual modal closing logic is handled in the modal's SetInputCapture
	// We return a special marker to indicate the modal should be closed
	return nil
}

func switchToResponsePreviewTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.ResponsePages.SwitchToPage("preview")
	updateResponseTabHeader(ui.ResponseTabHeader, 0, ui.Colors)
	ui.CurrentResponseTabIndex = 0
	return nil
}

func switchToResponseHeadersTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.ResponsePages.SwitchToPage("headers")
	updateResponseTabHeader(ui.ResponseTabHeader, 1, ui.Colors)
	ui.CurrentResponseTabIndex = 1
	return nil
}

func switchToResponseCookiesTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.ResponsePages.SwitchToPage("cookies")
	updateResponseTabHeader(ui.ResponseTabHeader, 2, ui.Colors)
	ui.CurrentResponseTabIndex = 2
	return nil
}

func switchToResponseTimelineTab(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	ui.ResponsePages.SwitchToPage("timeline")
	updateResponseTabHeader(ui.ResponseTabHeader, 3, ui.Colors)
	ui.CurrentResponseTabIndex = 3
	return nil
}

func openResponseInFx(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	if ui.LastResponse != nil && ui.LastResponse.Body != "" {
		ui.Suspend(func() {
			openInFxFunc(ui.LastResponse.Body)
		})
	}
	return nil
}

func showWorkspaceMenu(ui *UIOrchestrator, event *tcell.EventKey) *tcell.EventKey {
	// Show workspace configuration modal
	showWorkspaceModal(ui)
	return nil
}
