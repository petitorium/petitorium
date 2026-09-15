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
		{ID: "file.duplicateRequest", Label: "Duplicate Request", Category: "File", Description: "Duplicate the selected request", Handler: dummyCommandHandler("Duplicate Request")},

		// Edit
		{ID: "edit.renameItem", Label: "Rename Item", Category: "Edit", Description: "Rename the selected collection, folder or request", Handler: dummyCommandHandler("Rename Item")},
		{ID: "edit.moveItem", Label: "Move Item", Category: "Edit", Description: "Move the selected collection, folder or request", Handler: dummyCommandHandler("Move Item")},
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

// dummyCommandHandler returns a placeholder handler that confirms execution.
// It exists to test the palette mechanism; real handlers replace these
// command by command over time.
func dummyCommandHandler(label string) func(*UIOrchestrator) {
	return func(ui *UIOrchestrator) {
		showSuccessModalWithFocus(ui.App, ui.Pages, fmt.Sprintf("Executed: %s", label), ui.App.GetFocus(), ui.Colors)
	}
}
