package workspace

import "testing"

// buildTestTree constructs a workspace with the following structure:
//
//	-root
//	  colA (id "a")
//	    reqA1 (id "a1")
//	    reqA2 (id "a2")
//	  colB (id "b")
//	    colB1 (id "b1")
//	      reqB1a (id "b1a")
func buildTestTree() *Workspace {
	return &Workspace{
		Name: "test",
		Collections: []Collection{
			{
				ID:   "a",
				Name: "ColA",
				Requests: []Request{
					{ID: "a1", Name: "ReqA1", Method: "GET", URL: "http://a1"},
					{ID: "a2", Name: "ReqA2", Method: "POST", URL: "http://a2"},
				},
			},
			{
				ID:   "b",
				Name: "ColB",
				Collections: []Collection{
					{
						ID:   "b1",
						Name: "ColB1",
						Requests: []Request{
							{ID: "b1a", Name: "ReqB1a", Method: "GET", URL: "http://b1a"},
						},
					},
				},
			},
		},
	}
}

func TestFindRequestByID(t *testing.T) {
	ws := buildTestTree()

	cases := []struct {
		id      string
		wantNil bool
		wantURL string
	}{
		{"a1", false, "http://a1"},
		{"b1a", false, "http://b1a"},
		{"a2", false, "http://a2"},
		{"missing", true, ""},
		{"", true, ""},
	}
	for _, c := range cases {
		got := FindRequestByID(&ws.Collections, c.id)
		if c.wantNil {
			if got != nil {
				t.Errorf("FindRequestByID(%q) = %v, want nil", c.id, got)
			}
			continue
		}
		if got == nil {
			t.Errorf("FindRequestByID(%q) = nil, want non-nil", c.id)
			continue
		}
		if got.URL != c.wantURL {
			t.Errorf("FindRequestByID(%q).URL = %q, want %q", c.id, got.URL, c.wantURL)
		}
	}
}

func TestFindCollectionByID(t *testing.T) {
	ws := buildTestTree()

	if col := FindCollectionByID(&ws.Collections, "b1"); col == nil || col.Name != "ColB1" {
		t.Errorf("FindCollectionByID(b1) = %+v, want ColB1", col)
	}
	if col := FindCollectionByID(&ws.Collections, "missing"); col != nil {
		t.Errorf("FindCollectionByID(missing) = %+v, want nil", col)
	}
}

func TestFindParentCollectionOfRequest(t *testing.T) {
	ws := buildTestTree()

	if parent := FindParentCollectionOfRequest(&ws.Collections, "b1a"); parent == nil || parent.ID != "b1" {
		t.Errorf("parent of b1a = %+v, want b1", parent)
	}
	if parent := FindParentCollectionOfRequest(&ws.Collections, "a1"); parent == nil || parent.ID != "a" {
		t.Errorf("parent of a1 = %+v, want a", parent)
	}
	if parent := FindParentCollectionOfRequest(&ws.Collections, "missing"); parent != nil {
		t.Errorf("parent of missing = %+v, want nil", parent)
	}
}

func TestIsDescendantCollection(t *testing.T) {
	ws := buildTestTree()
	colB := FindCollectionByID(&ws.Collections, "b")

	if !IsDescendantCollection(colB, "b1") {
		t.Error("b1 should be a descendant of b")
	}
	if IsDescendantCollection(colB, "a") {
		t.Error("a should not be a descendant of b")
	}
}

// TestMoveRequestByIDNoLossOrDuplication simulates the core of the move
// operation: snapshot the request, remove it from its source by ID, then append
// the snapshot to a target collection. This is the exact sequence that used to
// corrupt data when *selectedRequest was dereferenced after the removal.
func TestMoveRequestByIDNoLossOrDuplication(t *testing.T) {
	ws := buildTestTree()

	// Move reqA1 from ColA into nested ColB1.
	src := FindRequestByID(&ws.Collections, "a1")
	if src == nil {
		t.Fatal("source request not found")
	}
	// Snapshot BEFORE removal (the fix).
	savedReq := *src

	if !RemoveRequestByID(ws, "a1") {
		t.Fatal("RemoveRequestByID returned false")
	}

	target := FindCollectionByID(&ws.Collections, "b1")
	if target == nil {
		t.Fatal("target collection not found")
	}
	target.Requests = append(target.Requests, savedReq)

	// The source collection should now have exactly one request (a2).
	colA := FindCollectionByID(&ws.Collections, "a")
	if len(colA.Requests) != 1 {
		t.Errorf("ColA has %d requests after move, want 1", len(colA.Requests))
	}
	if colA.Requests[0].ID != "a2" {
		t.Errorf("ColA[0].ID = %q, want a2 (neighbour must not be duplicated into source)", colA.Requests[0].ID)
	}

	// The target should now contain the moved request exactly once.
	count := 0
	for _, r := range target.Requests {
		if r.ID == "a1" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("moved request a1 appears %d times in target, want 1", count)
	}

	// Total request count across the workspace must be unchanged (no loss).
	total := countRequests(ws.Collections)
	if total != 3 {
		t.Errorf("total requests = %d, want 3 (no loss/no duplication)", total)
	}
}

// TestMoveRequestByIDTwice verifies the "compounding after the first move"
// scenario: a second move of a different request must also be lossless.
func TestMoveRequestByIDTwice(t *testing.T) {
	ws := buildTestTree()

	// First move: a1 -> ColB1.
	moveRequest(t, ws, "a1", "b1")
	// Second move: a2 -> ColB1.
	moveRequest(t, ws, "a2", "b1")

	colA := FindCollectionByID(&ws.Collections, "a")
	if len(colA.Requests) != 0 {
		t.Errorf("ColA has %d requests after two moves, want 0", len(colA.Requests))
	}
	target := FindCollectionByID(&ws.Collections, "b1")
	if len(target.Requests) != 3 {
		t.Errorf("ColB1 has %d requests after two moves, want 3", len(target.Requests))
	}
	if total := countRequests(ws.Collections); total != 3 {
		t.Errorf("total requests = %d, want 3", total)
	}
}

func moveRequest(t *testing.T, ws *Workspace, requestID, targetID string) {
	t.Helper()
	src := FindRequestByID(&ws.Collections, requestID)
	if src == nil {
		t.Fatalf("source %q not found", requestID)
	}
	savedReq := *src
	if !RemoveRequestByID(ws, requestID) {
		t.Fatalf("RemoveRequestByID(%q) false", requestID)
	}
	target := FindCollectionByID(&ws.Collections, targetID)
	if target == nil {
		t.Fatalf("target %q not found", targetID)
	}
	target.Requests = append(target.Requests, savedReq)
}

func countRequests(collections []Collection) int {
	n := 0
	for i := range collections {
		n += len(collections[i].Requests)
		n += countRequests(collections[i].Collections)
	}
	return n
}

// TestMoveCollectionByID verifies that moving a collection into a nested parent
// re-resolves the target by ID after the source removal shifts the backing array.
func TestMoveCollectionByID(t *testing.T) {
	ws := buildTestTree()

	// Move ColA into ColB (which sits at a higher index, so removing ColA shifts
	// ColB left by one slot — exactly the case where a stale pointer would land
	// on the wrong collection).
	src := FindCollectionByID(&ws.Collections, "a")
	if src == nil {
		t.Fatal("source collection not found")
	}
	savedCol := *src
	RemoveCollectionByID(ws, savedCol.ID)

	newParent := FindCollectionByID(&ws.Collections, "b")
	if newParent == nil {
		t.Fatal("target parent not found after removal")
	}
	newParent.Collections = append(newParent.Collections, savedCol)

	// ColA should no longer be at the root level.
	for _, c := range ws.Collections {
		if c.ID == "a" {
			t.Fatal("ColA should have been removed from root level")
		}
	}
	// ColA should now be a child of ColB.
	colB := FindCollectionByID(&ws.Collections, "b")
	found := false
	for _, c := range colB.Collections {
		if c.ID == "a" {
			found = true
		}
	}
	if !found {
		t.Error("ColA was not appended to ColB after move")
	}
}

// TestEnsureIDsAssignsAndPreserves verifies that ensureIDs fills in missing IDs
// without disturbing existing ones.
func TestEnsureIDsAssignsAndPreserves(t *testing.T) {
	ws := &Workspace{
		Name: "test",
		Collections: []Collection{
			{ID: "keep-me", Name: "Col", Requests: []Request{
				{Name: "Req"}, // no ID
			}},
		},
	}

	if !ensureIDs(ws) {
		t.Error("ensureIDs reported no changes, but a request lacked an ID")
	}
	if ws.Collections[0].ID != "keep-me" {
		t.Errorf("existing collection ID changed: got %q, want keep-me", ws.Collections[0].ID)
	}
	if ws.Collections[0].Requests[0].ID == "" {
		t.Error("ensureIDs did not assign an ID to the request")
	}
	// Running again must be a no-op.
	if ensureIDs(ws) {
		t.Error("ensureIDs reported changes on a fully-IDed workspace")
	}
}
