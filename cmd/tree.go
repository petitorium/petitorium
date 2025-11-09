package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

// addWorkspaceToTree adds workspace data to the tree root node
func addWorkspaceToTree(data *workspace.Workspace, root *tview.TreeNode) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	selectionBackgroundColor := hexToColor(theme.SelectionBackground)

	// Add collections
	for _, collection := range data.Collections {
		hasChildren := len(collection.Requests) > 0 || len(collection.Collections) > 0

		// Determine expansion state based on configuration
		shouldExpand := false
		shouldShowExpandedIcon := false

		switch config.C.UI.CollectionExpansion {
		case "expanded":
			shouldExpand = hasChildren
			shouldShowExpandedIcon = hasChildren
		case "remember":
			// Use the stored expansion state
			shouldExpand = collection.Expanded && hasChildren
			shouldShowExpandedIcon = collection.Expanded && hasChildren
		default: // "closed" or any other value
			shouldExpand = false
			shouldShowExpandedIcon = false
		}

		nodeText := fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, collection.Name)
		if shouldShowExpandedIcon {
			nodeText = fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, collection.Name)
		}

		collectionNode := tview.NewTreeNode(nodeText).
			SetSelectable(true).
			SetTextStyle(tcell.StyleDefault.Background(backgroundColor)).
			SetSelectedTextStyle(tcell.StyleDefault.Background(selectionBackgroundColor).Foreground(foregroundColor)).
			SetReference(collection)

		// Set expansion state
		if shouldExpand {
			collectionNode.SetExpanded(true)
			// Add children for expanded nodes
			addChildrenToCollectionNode(collectionNode, collection)
		} else {
			// Explicitly set collapsed state for closed collections
			collectionNode.SetExpanded(false)
		}

		root.AddChild(collectionNode)
	}
}

// addChildrenToCollectionNode adds requests and sub-collections to a collection node
func addChildrenToCollectionNode(node *tview.TreeNode, collection workspace.Collection) {
	theme := config.C.Theme
	backgroundColor := hexToColor(theme.BackgroundColor)
	foregroundColor := hexToColor(theme.ForegroundColor)
	selectionBackgroundColor := hexToColor(theme.SelectionBackground)

	// Add requests
	for _, req := range collection.Requests {
		coloredMethod := getColoredMethod(req.Method)
		paddedName := padNameToMinLength(req.Name, 4)
		// Reserve space for selection icon (icon + fixed space)
		iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon) // + 1
		spacePadding := strings.Repeat(" ", iconWidth)
		reqNodeText := fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName)
		requestNode := tview.NewTreeNode(reqNodeText).
			SetSelectable(true).
			SetTextStyle(tcell.StyleDefault.Background(backgroundColor)).
			SetSelectedTextStyle(tcell.StyleDefault.Background(selectionBackgroundColor).Foreground(foregroundColor)).
			SetReference(req)
		node.AddChild(requestNode)
	}

	// Add sub-collections
	for _, subCol := range collection.Collections {
		hasChildren := len(subCol.Requests) > 0 || len(subCol.Collections) > 0

		// Determine expansion state based on configuration
		shouldExpand := false
		shouldShowExpandedIcon := false

		switch config.C.UI.CollectionExpansion {
		case "expanded":
			shouldExpand = hasChildren
			shouldShowExpandedIcon = hasChildren
		case "remember":
			// Use the stored expansion state
			shouldExpand = subCol.Expanded && hasChildren
			shouldShowExpandedIcon = subCol.Expanded && hasChildren
		default: // "closed" or any other value
			shouldExpand = false
			shouldShowExpandedIcon = false
		}

		// Set appropriate icon
		nodeText := fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, subCol.Name)
		if shouldShowExpandedIcon {
			nodeText = fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, subCol.Name)
		}

		subCollectionNode := tview.NewTreeNode(nodeText).
			SetSelectable(true).
			SetTextStyle(tcell.StyleDefault.Background(backgroundColor)).
			SetSelectedTextStyle(tcell.StyleDefault.Background(selectionBackgroundColor).Foreground(foregroundColor)).
			SetReference(subCol)

		// Set expansion state
		if shouldExpand {
			subCollectionNode.SetExpanded(true)
			// Recursively add children for expanded sub-collections
			addChildrenToCollectionNode(subCollectionNode, subCol)
		} else {
			// Explicitly set collapsed state for closed sub-collections
			subCollectionNode.SetExpanded(false)
		}

		node.AddChild(subCollectionNode)
	}
}
