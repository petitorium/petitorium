package workspace

import "github.com/google/uuid"

// This file provides identity-based (ID) lookup and mutation helpers for the
// workspace tree. They replace the older value-based matching (by Name /
// Method / URL) which is ambiguous when two items share the same name and
// unsafe after slice shifts caused by append(s[:i], s[i+1:]...) removals.
//
// All finders return pointers that live inside the provided slice's backing
// array, so they remain valid until the caller mutates the tree's structure
// (insert/remove/reorder). After any structural mutation, re-resolve pointers
// by ID rather than caching them across the mutation.

// NewID returns a new unique identifier for a request or collection.
// Always assign an ID at construction time so the item can be resolved by ID
// before the next SaveWorkspace (which runs ensureIDs as a safety net).
func NewID() string {
	return uuid.NewString()
}

// FindRequestByID returns a pointer to the Request with the given ID anywhere
// in the collection tree, or nil if not found.
func FindRequestByID(collections *[]Collection, id string) *Request {
	if collections == nil || id == "" {
		return nil
	}
	for i := range *collections {
		col := &(*collections)[i]
		for j := range col.Requests {
			if col.Requests[j].ID == id {
				return &col.Requests[j]
			}
		}
		if ptr := FindRequestByID(&col.Collections, id); ptr != nil {
			return ptr
		}
	}
	return nil
}

// FindCollectionByID returns a pointer to the Collection with the given ID
// anywhere in the collection tree, or nil if not found.
func FindCollectionByID(collections *[]Collection, id string) *Collection {
	if collections == nil || id == "" {
		return nil
	}
	for i := range *collections {
		col := &(*collections)[i]
		if col.ID == id {
			return col
		}
		if ptr := FindCollectionByID(&col.Collections, id); ptr != nil {
			return ptr
		}
	}
	return nil
}

// FindParentCollectionOfRequest returns the collection that directly contains
// the request with the given ID, or nil if not found.
func FindParentCollectionOfRequest(collections *[]Collection, requestID string) *Collection {
	if collections == nil || requestID == "" {
		return nil
	}
	for i := range *collections {
		col := &(*collections)[i]
		for _, req := range col.Requests {
			if req.ID == requestID {
				return col
			}
		}
		if ptr := FindParentCollectionOfRequest(&col.Collections, requestID); ptr != nil {
			return ptr
		}
	}
	return nil
}

// FindParentCollection returns the collection that directly contains the
// collection with the given childID, or nil if the child is at the root level
// or not found.
func FindParentCollection(collections *[]Collection, childID string) *Collection {
	if collections == nil || childID == "" {
		return nil
	}
	for i := range *collections {
		col := &(*collections)[i]
		for _, sub := range col.Collections {
			if sub.ID == childID {
				return col
			}
		}
		if ptr := FindParentCollection(&col.Collections, childID); ptr != nil {
			return ptr
		}
	}
	return nil
}

// RemoveRequestByID removes the request with the given ID from anywhere in the
// tree. Returns true if the request was found and removed.
func RemoveRequestByID(ws *Workspace, requestID string) bool {
	if ws == nil || requestID == "" {
		return false
	}
	return removeRequestFromCollections(&ws.Collections, requestID)
}

func removeRequestFromCollections(collections *[]Collection, requestID string) bool {
	for i := range *collections {
		col := &(*collections)[i]
		for j := range col.Requests {
			if col.Requests[j].ID == requestID {
				col.Requests = append(col.Requests[:j], col.Requests[j+1:]...)
				return true
			}
		}
		if removeRequestFromCollections(&col.Collections, requestID) {
			return true
		}
	}
	return false
}

// RemoveCollectionByID removes the collection with the given ID from anywhere
// in the tree. Returns true if the collection was found and removed.
func RemoveCollectionByID(ws *Workspace, collectionID string) bool {
	if ws == nil || collectionID == "" {
		return false
	}
	return removeCollectionFromTree(&ws.Collections, collectionID)
}

func removeCollectionFromTree(collections *[]Collection, collectionID string) bool {
	for i := range *collections {
		if (*collections)[i].ID == collectionID {
			*collections = append((*collections)[:i], (*collections)[i+1:]...)
			return true
		}
		if removeCollectionFromTree(&(*collections)[i].Collections, collectionID) {
			return true
		}
	}
	return false
}

// IsDescendantCollection returns true if the collection with childID is a
// descendant of parent (at any depth). Used to prevent moving a collection
// into one of its own descendants (which would create a cycle).
func IsDescendantCollection(parent *Collection, childID string) bool {
	if parent == nil || childID == "" {
		return false
	}
	for i := range parent.Collections {
		sub := &parent.Collections[i]
		if sub.ID == childID {
			return true
		}
		if IsDescendantCollection(sub, childID) {
			return true
		}
	}
	return false
}
