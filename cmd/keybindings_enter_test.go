package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/petitorium/petitorium/workspace"
	"github.com/rivo/tview"
)

func TestKeyBindingManagerMatchesEnter(t *testing.T) {
	kbm := NewKeyBindingManager()

	// Check if Enter is bound in tree_view context
	found := false
	for _, b := range kbm.treeViewBindings {
		if b.Key == tcell.KeyEnter && b.Context == "tree_view" {
			found = true
			if b.Description != "Select request in tree" {
				t.Errorf("expected description 'Select request in tree', got '%s'", b.Description)
			}
			break
		}
	}
	if !found {
		t.Error("Enter keybinding not found in tree_view context")
	}

	// Verify the binding matches an Enter event
	event := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	matched := false
	for _, b := range kbm.treeViewBindings {
		if b.Matches(event) {
			matched = true
			break
		}
	}
	if !matched {
		t.Error("KeyBindingManager failed to match Enter key event for tree view")
	}
}

func TestSelectRequestInTree(t *testing.T) {
	// Create a minimal UI orchestrator with a tree view and a request node
	app := tview.NewApplication()
	root := tview.NewTreeNode("root").SetSelectable(false)
	reqNode := tview.NewTreeNode("GET Test").SetSelectable(true)
	reqNode.SetReference(workspace.Request{Name: "Test", Method: "GET", URL: "http://example.com"})
	root.AddChild(reqNode)

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(reqNode)

	var handlerCalled bool
	ui := &UIOrchestrator{
		App:                 app,
		CollectionsTreeView: tree,
		TreeSelectionHandler: func(node *tview.TreeNode) {
			handlerCalled = true
		},
	}

	event := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := selectRequestInTree(ui, event)

	if result != nil {
		t.Error("selectRequestInTree should return nil")
	}
	if !handlerCalled {
		t.Error("TreeSelectionHandler was not called")
	}
}

func TestSelectRequestInTreeNoHandler(t *testing.T) {
	app := tview.NewApplication()
	root := tview.NewTreeNode("root").SetSelectable(false)
	reqNode := tview.NewTreeNode("GET Test").SetSelectable(true)
	reqNode.SetReference(workspace.Request{Name: "Test", Method: "GET", URL: "http://example.com"})
	root.AddChild(reqNode)

	tree := tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(reqNode)

	ui := &UIOrchestrator{
		App:                 app,
		CollectionsTreeView: tree,
		// TreeSelectionHandler is nil
	}

	event := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := selectRequestInTree(ui, event)

	if result != nil {
		t.Error("selectRequestInTree should return nil even when handler is nil")
	}
}
