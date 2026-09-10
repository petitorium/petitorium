package cmd

import "fmt"

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
		{ID: "file.newCollection", Label: "New Collection", Category: "File", Description: "Create a new collection or folder", Handler: dummyCommandHandler("New Collection")},
		{ID: "file.newRequest", Label: "New Request", Category: "File", Description: "Create a new request", Handler: dummyCommandHandler("New Request")},
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
		{ID: "plugins.openMarketplace", Label: "Open Plugin Marketplace", Category: "Plugins", Description: "Browse and manage plugins", Handler: dummyCommandHandler("Open Plugin Marketplace")},

		// Application
		{ID: "app.quit", Label: "Quit Petitorium", Category: "Application", Description: "Exit the application", Handler: dummyCommandHandler("Quit Petitorium")},
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
