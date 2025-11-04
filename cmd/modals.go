package cmd

import (
	"encoding/json"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// showEnvironmentModal displays the environment variables modal
func showEnvironmentModal(
	ui *UIOrchestrator,
) {
	// Get current selected environment
	currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
	var env *workspace.Environment
	var modalTitle string

	if currentEnvIndex == 0 {
		// Base Environment selected - find and edit the "Base" environment
		for i := range *ui.EnvironmentsData {
			if (*ui.EnvironmentsData)[i].Name == "Base" {
				env = &(*ui.EnvironmentsData)[i]
				modalTitle = " Environment: Base "
				break
			}
		}
		// If Base environment doesn't exist, create new environment
		if env == nil {
			env = nil
			modalTitle = " Create New Environment "
		}
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
		// Other environment is selected - edit existing environment
		env = &(*ui.EnvironmentsData)[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0
		modalTitle = " Environment: " + env.Name + " "
	} else {
		// Fallback - create new environment
		env = nil
		modalTitle = " Create New Environment "
	}

	ui.Header.SetTitle(modalTitle)

	// Create JSON editor for environment variables
	var jsonBytes []byte
	if env != nil {
		// Convert existing environment variables to JSON
		effectiveVars := env.GetEffectiveVariables(*ui.EnvironmentsData)
		jsonBytes, _ = json.MarshalIndent(effectiveVars, "", "  ")
	} else {
		// Start with empty JSON for new environment
		jsonBytes = []byte("{}")
	}

	// Create JSON editor
	jsonEditor := createTextArea(" Environment Variables (JSON) ", ui.Colors.Background, ui.Colors.Border, ui.Colors.Title, ui.Colors.Foreground)
	jsonEditor.SetText(string(jsonBytes), false)

	// Create left panel (environment list)
	var selectedEnvironment *workspace.Environment
	var leftPanel *tview.List

	leftPanel = createEnvironmentListPanel(
		ui.Colors.Background,
		ui.Colors.Border,
		ui.Colors.BorderFocus,
		ui.Colors.Title,
		ui.Colors.Foreground,
		ui.Colors.ButtonSelect,
		*ui.EnvironmentsData,
		func(env *workspace.Environment) {
			selectedEnvironment = env
			// Update the JSON editor with the selected environment's variables
			effectiveVars := env.GetEffectiveVariables(*ui.EnvironmentsData)
			jsonBytes, _ := json.MarshalIndent(effectiveVars, "", "  ")
			jsonEditor.SetText(string(jsonBytes), false)
		},
		func() {
			// Handle create new environment - directly create "New Environment"
			newEnv := workspace.Environment{
				Name:      "New Environment",
				Base:      "Base",
				Variables: make(map[string]string),
			}

			// Add to environments
			*ui.EnvironmentsData = append(*ui.EnvironmentsData, newEnv)

			// Save environments
			if err := workspace.SaveEnvironments(*ui.EnvironmentsData); err != nil {
				// Handle error
				return
			}

			// Refresh the environment dropdown
			updateEnvironmentDropdown(ui.EnvDropdown, *ui.EnvironmentsData)

			// Refresh the environment list in place
			newLeftPanel := createEnvironmentListPanel(
				ui.Colors.Background,
				ui.Colors.Border,
				ui.Colors.BorderFocus,
				ui.Colors.Title,
				ui.Colors.Foreground,
				ui.Colors.ButtonSelect,
				*ui.EnvironmentsData,
				func(env *workspace.Environment) {
					selectedEnvironment = env
					// Update the JSON editor with the selected environment's variables
					effectiveVars := env.GetEffectiveVariables(*ui.EnvironmentsData)
					jsonBytes, _ := json.MarshalIndent(effectiveVars, "", "  ")
					jsonEditor.SetText(string(jsonBytes), false)
				},
				func() {
					// This will be replaced by the outer function
				},
			)

			// Replace the left panel with the updated one
			content := tview.NewFlex().
				AddItem(newLeftPanel, 0, 4, false). // 40% for left panel
				AddItem(jsonEditor, 0, 6, false)    // 60% for JSON editor

			modal := createModal(content, 120, 40, ui.Colors.Background)
			ui.Pages.RemovePage("envVariables")
			ui.Pages.AddPage("envVariables", modal, true, true)
			ui.UpdateFooter()
			ui.App.SetFocus(newLeftPanel)
		},
	)

	// Create split layout: left 40%, right 60%
	content := tview.NewFlex().
		AddItem(leftPanel, 0, 4, false). // 40% for left panel
		AddItem(jsonEditor, 0, 6, false) // 60% for JSON editor

	modal := createModal(content, 120, 40, ui.Colors.Background)
	ui.Pages.AddPage("envVariables", modal, true, true)
	ui.UpdateFooter()
	ui.App.SetFocus(leftPanel)

	// Add keybinding to close modal with Escape, q, or Q using the keybinding manager
	ui.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// First try modal keybindings
		if result := ui.KeyManager.HandleKeyEvent(ui, event, "modal"); result != event {
			// Save the JSON back to environment variables if editing existing environment
			if selectedEnvironment != nil {
				jsonText := jsonEditor.GetText()
				var newVars map[string]string
				if err := json.Unmarshal([]byte(jsonText), &newVars); err != nil {
					// If JSON is invalid, keep the original variables
					// Could show an error message here
				} else {
					selectedEnvironment.Variables = newVars
					if saveErr := workspace.SaveEnvironments(*ui.EnvironmentsData); saveErr != nil {
						// Handle save error
					}
				}
			}

			// Close modal and return focus to config button
			ui.Pages.RemovePage("envVariables")
			ui.UpdateFooter()
			ui.App.SetFocus(ui.EnvConfigButton)
			return nil
		}
		return event
	})
}
