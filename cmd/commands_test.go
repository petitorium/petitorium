package cmd

import "testing"

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
