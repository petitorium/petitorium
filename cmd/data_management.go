package cmd

import (
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// DataManager handles data operations for workspace and requests
type DataManager struct {
	workspaceData *workspace.Workspace
}

// NewDataManager creates a new data manager
func NewDataManager(workspaceData *workspace.Workspace) *DataManager {
	return &DataManager{
		workspaceData: workspaceData,
	}
}

// SaveWorkspace saves the workspace data to disk
func (dm *DataManager) SaveWorkspace() error {
	return workspace.SaveWorkspace(dm.workspaceData)
}

// UpdateWorkspaceData updates the workspace data
func (dm *DataManager) UpdateWorkspaceData(newData *workspace.Workspace) {
	dm.workspaceData = newData
}

// GetWorkspaceData returns the current workspace data
func (dm *DataManager) GetWorkspaceData() *workspace.Workspace {
	return dm.workspaceData
}

// SetupTreeViewExpansionHandling sets up the initial tree view expansion state based on configuration
func SetupTreeViewExpansionHandling(workspaceData *workspace.Workspace, rootNode *tview.TreeNode) {
	// For "closed" mode, we need to ensure all collections start collapsed
	// For "remember" mode, the expansion state is handled within addWorkspaceToTree
	if config.C.UI.CollectionExpansion == "closed" {
		// Temporarily set config to "closed" for tree building, then restore
		originalExpansion := config.C.UI.CollectionExpansion
		config.C.UI.CollectionExpansion = "closed"
		addWorkspaceToTree(workspaceData, rootNode)
		config.C.UI.CollectionExpansion = originalExpansion
	} else {
		addWorkspaceToTree(workspaceData, rootNode)
	}
}
