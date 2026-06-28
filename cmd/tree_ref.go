package cmd

import (
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// NodeRef identifies a tree node's underlying workspace entity by stable ID.
//
// It is stored in tview.TreeNode references (via SetReference/GetReference) so
// that consumers can re-resolve the live *workspace.Request / *workspace.Collection
// by ID at access time. This replaces storing value copies, which drifted out of
// sync with the workspace data after any edit and were ambiguous when two items
// shared the same name.
//
// Name is a display-only cache used by the name-based path helpers (getNodeName,
// findNodePath, findNodeByPath) for selected-request persistence and search
// navigation. It is NOT a source of truth; data access must always go through
// collectionFromNode / requestFromNode, which resolve by ID.
type NodeRef struct {
	Kind NodeKind
	ID   string
	Name string
}

// NodeKind distinguishes request nodes from collection nodes.
type NodeKind int

const (
	// KindCollection marks a node referencing a workspace.Collection.
	KindCollection NodeKind = iota
	// KindRequest marks a node referencing a workspace.Request.
	KindRequest
)

// collectionFromNode resolves the live *Collection referenced by a tree node,
// or nil if the node is not a collection node or its ID no longer exists in the
// workspace (e.g. it was moved/deleted and the node is stale).
func (ui *UIOrchestrator) collectionFromNode(node *tview.TreeNode) *workspace.Collection {
	if node == nil {
		return nil
	}
	ref, ok := node.GetReference().(NodeRef)
	if !ok || ref.Kind != KindCollection || ref.ID == "" {
		return nil
	}
	return ui.DataManager.GetCollectionByID(ref.ID)
}

// requestFromNode resolves the live *Request referenced by a tree node, or nil
// if the node is not a request node or its ID no longer exists in the workspace.
func (ui *UIOrchestrator) requestFromNode(node *tview.TreeNode) *workspace.Request {
	if node == nil {
		return nil
	}
	ref, ok := node.GetReference().(NodeRef)
	if !ok || ref.Kind != KindRequest || ref.ID == "" {
		return nil
	}
	return ui.DataManager.GetRequestByID(ref.ID)
}

// nodeIsRequest reports whether a tree node references a request.
func nodeIsRequest(node *tview.TreeNode) bool {
	if node == nil {
		return false
	}
	ref, ok := node.GetReference().(NodeRef)
	return ok && ref.Kind == KindRequest && ref.ID != ""
}

// nodeIsCollection reports whether a tree node references a collection.
func nodeIsCollection(node *tview.TreeNode) bool {
	if node == nil {
		return false
	}
	ref, ok := node.GetReference().(NodeRef)
	return ok && ref.Kind == KindCollection && ref.ID != ""
}

// isRootCollectionByID reports whether the collection with the given ID sits
// directly under the workspace root. Root containers are labelled "Collection";
// any container nested inside another is labelled "Folder".
func isRootCollectionByID(ws *workspace.Workspace, id string) bool {
	if ws == nil || id == "" {
		return false
	}
	return workspace.FindParentCollection(&ws.Collections, id) == nil
}

// collectionIconFor returns the (collapsed, expanded) icon pair appropriate for
// a collection node: CollectionIcon at the root level, FolderIcon when nested.
func (ui *UIOrchestrator) collectionIconFor(node *tview.TreeNode) (string, string) {
	if findParentNode(ui.RootNode, node) == ui.RootNode {
		return config.C.UI.CollectionIcon, config.C.UI.CollectionExpandedIcon
	}
	return config.C.UI.FolderIcon, config.C.UI.FolderExpandedIcon
}
