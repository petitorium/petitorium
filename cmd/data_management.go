package cmd

import (
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

// DataManager handles data operations for collections and requests
type DataManager struct {
	collectionsData []workspace.Collection
}

// NewDataManager creates a new data manager
func NewDataManager(collectionsData []workspace.Collection) *DataManager {
	return &DataManager{
		collectionsData: collectionsData,
	}
}

// FindRequestPtr finds the pointer to a request in the collections data
func (dm *DataManager) FindRequestPtr(targetReq workspace.Request) *workspace.Request {
	return findRequestPtr(dm.collectionsData, targetReq)
}

// SaveCollections saves the collections data to disk
func (dm *DataManager) SaveCollections() error {
	return workspace.SaveCollections(dm.collectionsData)
}

// UpdateCollectionsData updates the collections data
func (dm *DataManager) UpdateCollectionsData(newData []workspace.Collection) {
	dm.collectionsData = newData
}

// GetCollectionsData returns the current collections data
func (dm *DataManager) GetCollectionsData() []workspace.Collection {
	return dm.collectionsData
}

// SetupTreeViewExpansionHandling sets up the initial tree view expansion state based on configuration
func SetupTreeViewExpansionHandling(collectionsData []workspace.Collection, rootNode *tview.TreeNode) {
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
}
