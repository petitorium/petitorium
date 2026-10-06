package cmd

import (
	"testing"

	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// TestTreeNodesUseThemeForeground guards against collection-tree node text
// falling back to the terminal's default foreground when unselected: node text
// must use the theme foreground so it matches the rest of the UI.
func TestTreeNodesUseThemeForeground(t *testing.T) {
	saved := config.C
	defer func() { config.C = saved }()

	config.C.Theme.BackgroundColor = "#101010"
	config.C.Theme.ForegroundColor = "#abcdef"
	config.C.Theme.TreeSelectionBackground = "#222222"
	config.C.UI.CollectionExpansion = "expanded"
	config.C.UI.CollectionIcon = "C"
	config.C.UI.CollectionExpandedIcon = "C"
	config.C.UI.FolderIcon = "F"
	config.C.UI.FolderExpandedIcon = "F"
	config.C.UI.SelectedRequestIcon = "R"

	data := &workspace.Workspace{
		Collections: []workspace.Collection{
			{
				ID:   "c1",
				Name: "Col",
				Requests: []workspace.Request{
					{ID: "r1", Name: "Req", Method: "GET"},
				},
			},
		},
	}

	root := tview.NewTreeNode("")
	addWorkspaceToTree(data, root)

	if got := len(root.GetChildren()); got != 1 {
		t.Fatalf("expected 1 collection node, got %d", got)
	}
	collectionNode := root.GetChildren()[0]
	requestNode := collectionNode.GetChildren()[0]

	want := hexToColor(config.C.Theme.ForegroundColor)
	for name, node := range map[string]*tview.TreeNode{
		"collection": collectionNode,
		"request":    requestNode,
	} {
		if fg, _, _ := node.GetTextStyle().Decompose(); fg != want {
			t.Errorf("%s node foreground = %v, want theme foreground %v", name, fg.Hex(), want.Hex())
		}
	}
}
