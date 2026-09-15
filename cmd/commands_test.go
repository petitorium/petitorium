package cmd

import (
	"reflect"
	"testing"
)

func TestGetCommandsRegistry(t *testing.T) {
	commands := getCommands()

	if len(commands) == 0 {
		t.Fatal("command registry is empty")
	}

	seenIDs := make(map[string]bool)
	seenCategories := make(map[string]bool)
	lastCategory := ""

	for _, cmd := range commands {
		if cmd.ID == "" {
			t.Errorf("command %q has empty ID", cmd.Label)
		}
		if seenIDs[cmd.ID] {
			t.Errorf("duplicate command ID %q", cmd.ID)
		}
		seenIDs[cmd.ID] = true

		if cmd.Label == "" {
			t.Errorf("command %q has empty Label", cmd.ID)
		}
		if cmd.Category == "" {
			t.Errorf("command %q has empty Category", cmd.ID)
		}
		if cmd.Description == "" {
			t.Errorf("command %q has empty Description", cmd.ID)
		}
		if cmd.Handler == nil {
			t.Errorf("command %q has nil Handler", cmd.ID)
		}

		// Categories must be contiguous so the palette can group them in order.
		if cmd.Category != lastCategory {
			if seenCategories[cmd.Category] {
				t.Errorf("category %q appears in multiple non-adjacent blocks", cmd.Category)
			}
			seenCategories[cmd.Category] = true
			lastCategory = cmd.Category
		}
	}
}

func TestQuitCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var quit *Command
	for i := range commands {
		if commands[i].ID == "app.quit" {
			quit = &commands[i]
			break
		}
	}
	if quit == nil {
		t.Fatal("app.quit command not found in registry")
	}

	if reflect.ValueOf(quit.Handler).Pointer() != reflect.ValueOf(quitApplication).Pointer() {
		t.Error("app.quit should use the real quitApplication handler, not a dummy")
	}
}

func TestOpenMarketplaceCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var open *Command
	for i := range commands {
		if commands[i].ID == "plugins.openMarketplace" {
			open = &commands[i]
			break
		}
	}
	if open == nil {
		t.Fatal("plugins.openMarketplace command not found in registry")
	}

	if reflect.ValueOf(open.Handler).Pointer() != reflect.ValueOf(openMarketplace).Pointer() {
		t.Error("plugins.openMarketplace should use the real openMarketplace handler, not a dummy")
	}
}

func TestNewCollectionCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var newCol *Command
	for i := range commands {
		if commands[i].ID == "file.newCollection" {
			newCol = &commands[i]
			break
		}
	}
	if newCol == nil {
		t.Fatal("file.newCollection command not found in registry")
	}

	if reflect.ValueOf(newCol.Handler).Pointer() != reflect.ValueOf(openNewCollectionForm).Pointer() {
		t.Error("file.newCollection should use the real openNewCollectionForm handler, not a dummy")
	}
}

func TestNewRequestCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var newReq *Command
	for i := range commands {
		if commands[i].ID == "file.newRequest" {
			newReq = &commands[i]
			break
		}
	}
	if newReq == nil {
		t.Fatal("file.newRequest command not found in registry")
	}

	if reflect.ValueOf(newReq.Handler).Pointer() != reflect.ValueOf(newRequestCommand).Pointer() {
		t.Error("file.newRequest should use the real newRequestCommand handler, not a dummy")
	}
}

func TestDuplicateRequestCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var dup *Command
	for i := range commands {
		if commands[i].ID == "file.duplicateRequest" {
			dup = &commands[i]
			break
		}
	}
	if dup == nil {
		t.Fatal("file.duplicateRequest command not found in registry")
	}

	if reflect.ValueOf(dup.Handler).Pointer() != reflect.ValueOf(duplicateRequestCommand).Pointer() {
		t.Error("file.duplicateRequest should use the real duplicateRequestCommand handler, not a dummy")
	}
}

func TestRenameItemCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var rename *Command
	for i := range commands {
		if commands[i].ID == "edit.renameItem" {
			rename = &commands[i]
			break
		}
	}
	if rename == nil {
		t.Fatal("edit.renameItem command not found in registry")
	}

	if reflect.ValueOf(rename.Handler).Pointer() != reflect.ValueOf(renameItemCommand).Pointer() {
		t.Error("edit.renameItem should use the real renameItemCommand handler, not a dummy")
	}
}

func TestMoveItemCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var move *Command
	for i := range commands {
		if commands[i].ID == "edit.moveItem" {
			move = &commands[i]
			break
		}
	}
	if move == nil {
		t.Fatal("edit.moveItem command not found in registry")
	}

	if reflect.ValueOf(move.Handler).Pointer() != reflect.ValueOf(moveItemCommand).Pointer() {
		t.Error("edit.moveItem should use the real moveItemCommand handler, not a dummy")
	}
}

func TestDeleteItemCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var del *Command
	for i := range commands {
		if commands[i].ID == "edit.deleteItem" {
			del = &commands[i]
			break
		}
	}
	if del == nil {
		t.Fatal("edit.deleteItem command not found in registry")
	}

	if reflect.ValueOf(del.Handler).Pointer() != reflect.ValueOf(deleteItemCommand).Pointer() {
		t.Error("edit.deleteItem should use the real deleteItemCommand handler, not a dummy")
	}
}

func TestJumpToWorkspaceCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToWorkspace" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToWorkspace command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToWorkspaceCommand).Pointer() {
		t.Error("view.jumpToWorkspace should use the real jumpToWorkspaceCommand handler, not a dummy")
	}
}

func TestJumpToEnvironmentCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToEnvironment" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToEnvironment command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToEnvironmentCommand).Pointer() {
		t.Error("view.jumpToEnvironment should use the real jumpToEnvironmentCommand handler, not a dummy")
	}
}

func TestJumpToCollectionsCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToCollections" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToCollections command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToCollectionsCommand).Pointer() {
		t.Error("view.jumpToCollections should use the real jumpToCollectionsCommand handler, not a dummy")
	}
}

func TestJumpToURLBarCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToURLBar" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToURLBar command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToURLBarCommand).Pointer() {
		t.Error("view.jumpToURLBar should use the real jumpToURLBarCommand handler, not a dummy")
	}
}

func TestJumpToRequestCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToRequest" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToRequest command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToRequestCommand).Pointer() {
		t.Error("view.jumpToRequest should use the real jumpToRequestCommand handler, not a dummy")
	}
}

func TestJumpToResponseCommandUsesRealHandler(t *testing.T) {
	commands := getCommands()

	var jump *Command
	for i := range commands {
		if commands[i].ID == "view.jumpToResponse" {
			jump = &commands[i]
			break
		}
	}
	if jump == nil {
		t.Fatal("view.jumpToResponse command not found in registry")
	}

	if reflect.ValueOf(jump.Handler).Pointer() != reflect.ValueOf(jumpToResponseCommand).Pointer() {
		t.Error("view.jumpToResponse should use the real jumpToResponseCommand handler, not a dummy")
	}
}
