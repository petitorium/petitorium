package cmd

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func TestFxKeybinding(t *testing.T) {
	kbm := NewKeyBindingManager()

	// Check if 'f' is bound in response_view context
	found := false
	for _, b := range kbm.responseViewBindings {
		if b.Rune == 'f' && b.Context == "response_view" {
			found = true
			if b.Description != "Open response in fx" {
				t.Errorf("expected description 'Open response in fx', got '%s'", b.Description)
			}
			break
		}
	}

	if !found {
		t.Error("fx keybinding ('f') not found in response_view context")
	}
}

func TestOpenResponseInFxAction(t *testing.T) {
	// Mock openInFxFunc
	originalFxFunc := openInFxFunc
	var capturedContent string
	openInFxFunc = func(content string) error {
		capturedContent = content
		return nil
	}
	defer func() { openInFxFunc = originalFxFunc }()

	// Mock UIOrchestrator
	ui := &UIOrchestrator{
		App: tview.NewApplication(),
		LastResponse: &HTTPResponse{
			Body: `{"test": "data"}`,
		},
		Suspend: func(f func()) bool {
			f()
			return true
		},
	}

	// Trigger action
	openResponseInFx(ui, nil)

	// Note: since ui.App.Suspend runs the function asynchronously or in a way that
	// might require the app to be running, this might not capture immediately
	// if Suspend logic is complex.
	// However, tview.Application.Suspend calls the function immediately if not running.

	if capturedContent != `{"test": "data"}` {
		t.Errorf("Expected captured content '{\"test\": \"data\"}', got '%s'", capturedContent)
	}
}

func TestHandleFxKey(t *testing.T) {
	// This test verifies that the key handler identifies the 'f' key correctly
	kbm := NewKeyBindingManager()

	event := tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone)

	// Check if 'f' matches any binding in response_view context
	found := false
	for _, b := range kbm.responseViewBindings {
		if b.Matches(event) {
			found = true
			if b.Description != "Open response in fx" {
				t.Errorf("expected description 'Open response in fx', got '%s'", b.Description)
			}
			break
		}
	}

	if !found {
		t.Error("KeyBindingManager failed to match 'f' key event for response view")
	}
}

func TestCommandPaletteKeybinding(t *testing.T) {
	kbm := NewKeyBindingManager()

	event := tcell.NewEventKey(tcell.KeyCtrlP, 0, tcell.ModCtrl)

	found := false
	for _, b := range kbm.globalBindings {
		if b.Matches(event) {
			found = true
			if b.Description != "Open command palette" {
				t.Errorf("expected description 'Open command palette', got '%s'", b.Description)
			}
			break
		}
	}

	if !found {
		t.Error("Ctrl+P not bound in global context for the command palette")
	}
}

func TestShowCommandRunnerModalActionPassesThroughOnButton(t *testing.T) {
	app := tview.NewApplication()
	btn := createThemedButton(" Send ", &ColorManager{})
	app.SetFocus(btn)

	ui := &UIOrchestrator{
		App:        app,
		SendButton: btn,
		Pages:      tview.NewPages(),
	}

	event := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	result := showCommandRunnerModalAction(ui, event)

	if result != event {
		t.Error("showCommandRunnerModalAction should return the event when focus is on a button")
	}
}
