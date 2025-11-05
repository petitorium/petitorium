package cmd

import (
	"fmt"

	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

// updateCollectionExpansionState updates the expansion state of a collection in the collections data
func updateCollectionExpansionState(collectionsData *[]workspace.Collection, collectionName string, expanded bool) {
	var updateExpansion func(collections *[]workspace.Collection) bool
	updateExpansion = func(collections *[]workspace.Collection) bool {
		for i := range *collections {
			if (*collections)[i].Name == collectionName {
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

func deleteCollectionFromData(collectionsData *[]workspace.Collection, collectionName string) {
	for i := len(*collectionsData) - 1; i >= 0; i-- {
		if (*collectionsData)[i].Name == collectionName {
			*collectionsData = append((*collectionsData)[:i], (*collectionsData)[i+1:]...)
			return
		}
		// Check nested collections recursively
		deleteCollectionFromNestedData(&(*collectionsData)[i].Collections, collectionName)
	}
}

func deleteRequestFromData(workspaceData *workspace.Workspace, requestName string) {
	// Delete from root requests
	for i := len(workspaceData.Requests) - 1; i >= 0; i-- {
		if workspaceData.Requests[i].Name == requestName {
			workspaceData.Requests = append(workspaceData.Requests[:i], workspaceData.Requests[i+1:]...)
		}
	}
	// Delete from collections
	for i := range workspaceData.Collections {
		deleteRequestFromCollection(&workspaceData.Collections[i], requestName)
	}
}

func deleteCollectionFromNestedData(collections *[]workspace.Collection, collectionName string) {
	for i := len(*collections) - 1; i >= 0; i-- {
		if (*collections)[i].Name == collectionName {
			*collections = append((*collections)[:i], (*collections)[i+1:]...)
			return
		}
		// Check further nested collections
		deleteCollectionFromNestedData(&(*collections)[i].Collections, collectionName)
	}
}

func deleteRequestFromCollection(collection *workspace.Collection, requestName string) {
	// Check requests in this collection
	for i := len(collection.Requests) - 1; i >= 0; i-- {
		if collection.Requests[i].Name == requestName {
			collection.Requests = append(collection.Requests[:i], collection.Requests[i+1:]...)
			return
		}
	}
	// Check nested collections
	for i := range collection.Collections {
		deleteRequestFromCollection(&collection.Collections[i], requestName)
	}
}

// findCollectionByName recursively finds a collection by name in the collections data
func findCollectionByName(collections *[]workspace.Collection, name string) *workspace.Collection {
	for i := range *collections {
		if (*collections)[i].Name == name {
			return &(*collections)[i]
		}
		// Search in nested collections
		if found := findCollectionByName(&(*collections)[i].Collections, name); found != nil {
			return found
		}
	}
	return nil
}

// findParentCollectionOfRequest finds the parent collection that contains a specific request
func findParentCollectionOfRequest(collections *[]workspace.Collection, requestName string, requestMethod string, requestURL string) *workspace.Collection {
	for i := range *collections {
		// Check if the request is in this collection
		for _, req := range (*collections)[i].Requests {
			if req.Name == requestName && req.Method == requestMethod && req.URL == requestURL {
				return &(*collections)[i]
			}
		}
		// Search in nested collections
		if found := findParentCollectionOfRequest(&(*collections)[i].Collections, requestName, requestMethod, requestURL); found != nil {
			return found
		}
	}
	return nil
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
