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
