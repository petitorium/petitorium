package cmd

import (
	"sync"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// DataManager handles data operations for workspace and requests.
//
// It owns a sync.RWMutex that guards the workspace data and its persisted form.
// The tview event loop is single-threaded, so data mutations rarely race, but
// HTTP responses are appended from goroutines queued via App.QueueUpdateDraw;
// the mutex makes that safe defensively. Lookups go through workspace.FindXByID
// (O(n), always correct) rather than a cached pointer index, because appending
// to a Requests/Collections slice can reallocate the backing array and dangle
// any cached pointers.
type DataManager struct {
	mu            sync.RWMutex
	workspaceData *workspace.Workspace
}

// NewDataManager creates a new data manager
func NewDataManager(workspaceData *workspace.Workspace) *DataManager {
	return &DataManager{
		workspaceData: workspaceData,
	}
}

// GetRequestByID returns the live *Request with the given ID, or nil.
func (dm *DataManager) GetRequestByID(id string) *workspace.Request {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return workspace.FindRequestByID(&dm.workspaceData.Collections, id)
}

// GetCollectionByID returns the live *Collection with the given ID, or nil.
func (dm *DataManager) GetCollectionByID(id string) *workspace.Collection {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return workspace.FindCollectionByID(&dm.workspaceData.Collections, id)
}

// FindParentCollectionOfRequest returns the collection that directly contains
// the request with the given ID, or nil.
func (dm *DataManager) FindParentCollectionOfRequest(requestID string) *workspace.Collection {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
	return workspace.FindParentCollectionOfRequest(&dm.workspaceData.Collections, requestID)
}

// SaveWorkspace saves the workspace data to disk atomically, under the data
// manager's lock.
func (dm *DataManager) SaveWorkspace() error {
	dm.mu.Lock()
	defer dm.mu.Unlock()
	return workspace.SaveWorkspace(dm.workspaceData)
}

// UpdateWorkspaceData updates the workspace data pointer (e.g. after a
// structural mutation or a workspace switch).
func (dm *DataManager) UpdateWorkspaceData(newData *workspace.Workspace) {
	dm.mu.Lock()
	dm.workspaceData = newData
	dm.mu.Unlock()
}

// GetWorkspaceData returns the current workspace data
func (dm *DataManager) GetWorkspaceData() *workspace.Workspace {
	dm.mu.RLock()
	defer dm.mu.RUnlock()
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
