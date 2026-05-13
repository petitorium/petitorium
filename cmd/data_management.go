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

// FindRequestPtr finds the pointer to a request in the workspace data
func (dm *DataManager) FindRequestPtr(targetReq workspace.Request) *workspace.Request {
	return dm.findRequestPtr(dm.workspaceData, targetReq)
}

// findRequestPtr finds the request pointer in workspace collections
func (dm *DataManager) findRequestPtr(data *workspace.Workspace, req workspace.Request) *workspace.Request {
	for i := range data.Collections {
		for j := range data.Collections[i].Requests {
			if data.Collections[i].Requests[j].Name == req.Name &&
				data.Collections[i].Requests[j].Method == req.Method &&
				data.Collections[i].Requests[j].URL == req.URL {
				return &data.Collections[i].Requests[j]
			}
		}
		if ptr := dm.findRequestPtrInCollections(data.Collections[i].Collections, req); ptr != nil {
			return ptr
		}
	}
	return nil
}

// Helper function to find in nested collections
func (dm *DataManager) findRequestPtrInCollections(data []workspace.Collection, req workspace.Request) *workspace.Request {
	for i := range data {
		for j := range data[i].Requests {
			if data[i].Requests[j].Name == req.Name &&
				data[i].Requests[j].Method == req.Method &&
				data[i].Requests[j].URL == req.URL {
				return &data[i].Requests[j]
			}
		}
		if ptr := dm.findRequestPtrInCollections(data[i].Collections, req); ptr != nil {
			return ptr
		}
	}
	return nil
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
