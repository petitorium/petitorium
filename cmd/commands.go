package cmd

import (
	"fmt"

	"github.com/gdamore/tcell/v2"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// Command represents a single entry in the command palette.
type Command struct {
	// ID is a stable identifier for the command (e.g. "file.newRequest").
	ID string
	// Label is the display name shown in the palette.
	Label string
	// Category groups related commands in the palette (e.g. "File", "View").
	Category string
	// Description is a short explanation of what the command does.
	Description string
	// Handler is executed when the command is selected.
	Handler func(*UIOrchestrator)
}

// getCommands returns the commands available in the command palette.
// Commands are grouped by category: the palette preserves this order and
// renders one section per category, so entries of the same category must
// stay contiguous.
func getCommands() []Command {
	return []Command{
		// File
		{ID: "file.newCollection", Label: "New Collection", Category: "File", Description: "Create a new collection or folder", Handler: openNewCollectionForm},
		{ID: "file.newRequest", Label: "New Request", Category: "File", Description: "Create a new request", Handler: newRequestCommand},
		{ID: "file.duplicateRequest", Label: "Duplicate Request", Category: "File", Description: "Duplicate the selected request", Handler: duplicateRequestCommand},

		// Edit
		{ID: "edit.renameItem", Label: "Rename Item", Category: "Edit", Description: "Rename the selected collection, folder or request", Handler: renameItemCommand},
		{ID: "edit.moveItem", Label: "Move Item", Category: "Edit", Description: "Move the selected collection, folder or request", Handler: moveItemCommand},
		{ID: "edit.deleteItem", Label: "Delete Item", Category: "Edit", Description: "Delete the selected collection, folder or request", Handler: dummyCommandHandler("Delete Item")},

		// View
		{ID: "view.jumpToWorkspace", Label: "Jump to Workspace", Category: "View", Description: "Focus the workspace panel", Handler: dummyCommandHandler("Jump to Workspace")},
		{ID: "view.jumpToEnvironment", Label: "Jump to Environment", Category: "View", Description: "Focus the environment panel", Handler: dummyCommandHandler("Jump to Environment")},
		{ID: "view.jumpToCollections", Label: "Jump to Collections", Category: "View", Description: "Focus the collections tree", Handler: dummyCommandHandler("Jump to Collections")},
		{ID: "view.jumpToURLBar", Label: "Jump to URL Bar", Category: "View", Description: "Focus the URL bar", Handler: dummyCommandHandler("Jump to URL Bar")},
		{ID: "view.jumpToRequest", Label: "Jump to Request", Category: "View", Description: "Focus the request panel", Handler: dummyCommandHandler("Jump to Request")},
		{ID: "view.jumpToResponse", Label: "Jump to Response", Category: "View", Description: "Focus the response panel", Handler: dummyCommandHandler("Jump to Response")},

		// Navigate
		{ID: "navigate.searchRequests", Label: "Search Requests", Category: "Navigate", Description: "Quick-search collections and requests", Handler: dummyCommandHandler("Search Requests")},
		{ID: "navigate.searchWorkspaces", Label: "Search Workspaces", Category: "Navigate", Description: "Quick-search and switch workspaces", Handler: dummyCommandHandler("Search Workspaces")},

		// Plugins
		{ID: "plugins.openMarketplace", Label: "Open Plugin Marketplace", Category: "Plugins", Description: "Browse and manage plugins", Handler: openMarketplace},

		// Application
		{ID: "app.quit", Label: "Quit Petitorium", Category: "Application", Description: "Exit the application", Handler: quitApplication},
	}
}

// quitApplication saves the collection expansion state (when "remember" mode
// is enabled) and stops the application. Shared by the 'q' keybinding and the
// command palette "Quit" command.
func quitApplication(ui *UIOrchestrator) {
	if config.C.UI.CollectionExpansion == "remember" {
		if err := workspace.SaveExpansionState(&ui.WorkspaceData.Collections); err != nil {
			// Could log error but for now just continue
		}
	}
	ui.App.Stop()
}

// openMarketplace opens the plugin marketplace modal.
func openMarketplace(ui *UIOrchestrator) {
	ui.ShowMarketplace()
}

// openNewCollectionForm opens the "New Collection" form as a modal. Shared by
// the 'N' keybinding and the command palette "New Collection" command.
func openNewCollectionForm(ui *UIOrchestrator) {
	form := createCollectionFormWithLocation(ui)
	modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
	ui.Pages.AddPage("newCollection", modal, true, true)
	ui.App.SetFocus(form)
}

// openNewRequestForm opens the "New Request" form for the collection implied
// by the current tree selection (a selected collection, or the parent
// collection of a selected request). It returns false when no target
// collection can be determined, so callers can choose how to respond.
func openNewRequestForm(ui *UIOrchestrator) bool {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node == nil {
		return false
	}

	var selectedCollection *workspace.Collection
	if col := ui.collectionFromNode(node); col != nil {
		selectedCollection = col
	} else if req := ui.requestFromNode(node); req != nil {
		selectedCollection = workspace.FindParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.ID)
	}
	if selectedCollection == nil {
		return false
	}

	form := createRequestForm(ui.App, ui.Pages, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
	modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
	ui.Pages.AddPage("newRequest", modal, true, true)
	ui.App.SetFocus(form)
	return true
}

// newRequestCommand is the command palette handler for "New Request". It
// reports an error when no collection can be determined from the current
// selection.
func newRequestCommand(ui *UIOrchestrator) {
	if !openNewRequestForm(ui) {
		showErrorModalWithFocus(ui.App, ui.Pages, "Select a collection to create the request in", ui.App.GetFocus(), ui.Colors)
	}
}

// openDuplicateRequestForm opens the "Duplicate Request" form for the request
// selected in the tree. It returns false when no request is selected or its
// parent collection cannot be determined.
func openDuplicateRequestForm(ui *UIOrchestrator) bool {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node == nil {
		return false
	}

	req := ui.requestFromNode(node)
	if req == nil {
		return false
	}

	selectedCollection := workspace.FindParentCollectionOfRequest(&ui.WorkspaceData.Collections, req.ID)
	if selectedCollection == nil {
		return false
	}

	form := createDuplicateRequestForm(ui.App, ui.Pages, req, selectedCollection, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
	modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
	ui.Pages.AddPage("duplicateRequest", modal, true, true)
	ui.App.SetFocus(form)
	return true
}

// duplicateRequestCommand is the command palette handler for "Duplicate
// Request". It reports an error when no request is selected.
func duplicateRequestCommand(ui *UIOrchestrator) {
	if !openDuplicateRequestForm(ui) {
		showErrorModalWithFocus(ui.App, ui.Pages, "Select a request to duplicate", ui.App.GetFocus(), ui.Colors)
	}
}

// openRenameItemForm opens the appropriate rename form for the collection,
// folder or request selected in the tree. It returns false when the current
// node is neither a collection nor a request.
func openRenameItemForm(ui *UIOrchestrator) bool {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node == nil {
		return false
	}

	if col := ui.collectionFromNode(node); col != nil {
		form := createRenameCollectionForm(ui.App, ui.Pages, col, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
		modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
		ui.Pages.AddPage("renameCollection", modal, true, true)
		ui.App.SetFocus(form)
		return true
	}

	if req := ui.requestFromNode(node); req != nil {
		form := createRenameRequestForm(ui.App, ui.Pages, req, ui.WorkspaceData, ui.RootNode, ui.CollectionsTreeView, node, ui.Colors)
		modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
		ui.Pages.AddPage("renameRequest", modal, true, true)
		ui.App.SetFocus(form)
		return true
	}

	return false
}

// renameItemCommand is the command palette handler for "Rename Item". It
// reports an error when no collection or request is selected.
func renameItemCommand(ui *UIOrchestrator) {
	if !openRenameItemForm(ui) {
		showErrorModalWithFocus(ui.App, ui.Pages, "Select a collection, folder or request to rename", ui.App.GetFocus(), ui.Colors)
	}
}

// openMoveItemForm opens the appropriate move form for the collection, folder
// or request selected in the tree. It returns false when the current node is
// neither a collection nor a request.
func openMoveItemForm(ui *UIOrchestrator) bool {
	node := ui.CollectionsTreeView.GetCurrentNode()
	if node == nil {
		return false
	}

	if col := ui.collectionFromNode(node); col != nil {
		form := createMoveCollectionForm(ui, col)
		modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
		ui.Pages.AddPage("moveCollection", modal, true, true)
		ui.App.SetFocus(form)
		return true
	}

	if req := ui.requestFromNode(node); req != nil {
		form := createMoveRequestForm(ui, req)
		modal := createSizedModal(form, modalSizeEditor, tcell.ColorDefault)
		ui.Pages.AddPage("moveRequest", modal, true, true)
		ui.App.SetFocus(form)
		return true
	}

	return false
}

// moveItemCommand is the command palette handler for "Move Item". It reports
// an error when no collection or request is selected.
func moveItemCommand(ui *UIOrchestrator) {
	if !openMoveItemForm(ui) {
		showErrorModalWithFocus(ui.App, ui.Pages, "Select a collection, folder or request to move", ui.App.GetFocus(), ui.Colors)
	}
}

// dummyCommandHandler returns a placeholder handler that confirms execution.
// It exists to test the palette mechanism; real handlers replace these
// command by command over time.
func dummyCommandHandler(label string) func(*UIOrchestrator) {
	return func(ui *UIOrchestrator) {
		showSuccessModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Executed: %s", label), ui.App.GetFocus(), ui.Colors)
	}
}
