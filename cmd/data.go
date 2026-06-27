package cmd

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// updateCollectionExpansionState updates the expansion state of a collection in the collections data
func updateCollectionExpansionState(collectionsData *[]workspace.Collection, collectionID string, expanded bool) {
	var updateExpansion func(collections *[]workspace.Collection) bool
	updateExpansion = func(collections *[]workspace.Collection) bool {
		for i := range *collections {
			if (*collections)[i].ID == collectionID {
				(*collections)[i].Expanded = expanded
				return true
			}
			if len((*collections)[i].Collections) > 0 {
				if updateExpansion(&(*collections)[i].Collections) {
					return true
				}
			}
		}
		return false
	}

	updateExpansion(collectionsData)

	// Save expansion state if in "remember" mode
	if config.C.UI.CollectionExpansion == "remember" {
		if err := workspace.SaveExpansionState(collectionsData); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: Failed to save expansion state: %v\n", err)
		}
	}
}

// findParentNode finds the parent node of the given node in the tree
func findParentNode(rootNode *tview.TreeNode, targetNode *tview.TreeNode) *tview.TreeNode {
	var parent *tview.TreeNode

	rootNode.Walk(func(node, currentParent *tview.TreeNode) bool {
		if node == targetNode {
			parent = currentParent
			return false // Stop walking
		}
		return true // Continue walking
	})

	return parent
}
