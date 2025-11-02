package cmd

import (
	"encoding/json"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// showEnvironmentModal displays the environment variables modal
func showEnvironmentModal(
	app *tview.Application,
	pages *tview.Pages,
	environmentsData []workspace.Environment,
	envDropdown *tview.DropDown,
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	buttonSelectedColor tcell.Color,
	header *tview.TextView,
) {
	// Get current selected environment
	currentEnvIndex, _ := envDropdown.GetCurrentOption()
	var env *workspace.Environment
	var modalTitle string

	if currentEnvIndex == 0 {
		// Base Environment selected - find and edit the "Base" environment
		for i := range environmentsData {
			if environmentsData[i].Name == "Base" {
				env = &environmentsData[i]
				modalTitle = " Environment: Base "
				break
			}
		}
		// If Base environment doesn't exist, create new environment
		if env == nil {
			env = nil
			modalTitle = " Create New Environment "
		}
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(environmentsData) {
		// Other environment is selected - edit existing environment
		env = &environmentsData[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0
		modalTitle = " Environment: " + env.Name + " "
	} else {
		// Fallback - create new environment
		env = nil
		modalTitle = " Create New Environment "
	}

	header.SetTitle(modalTitle)

	// Create JSON editor for environment variables
	var jsonBytes []byte
	if env != nil {
		// Convert existing environment variables to JSON
		effectiveVars := env.GetEffectiveVariables(environmentsData)
		jsonBytes, _ = json.MarshalIndent(effectiveVars, "", "  ")
	} else {
		// Start with empty JSON for new environment
		jsonBytes = []byte("{}")
	}

	// Create JSON editor
	jsonEditor := createTextArea(" Environment Variables (JSON) ", backgroundColor, borderColor, titleColor, foregroundColor)
	jsonEditor.SetText(string(jsonBytes), false)

	// Create left panel (environment list)
	var selectedEnvironment *workspace.Environment
	var leftPanel *tview.Flex

	leftPanel = createEnvironmentListPanel(
		backgroundColor,
		borderColor,
		borderFocusColor,
		titleColor,
		foregroundColor,
		buttonSelectedColor,
		environmentsData,
		func(env *workspace.Environment) {
			selectedEnvironment = env
			// Update the JSON editor with the selected environment's variables
			effectiveVars := env.GetEffectiveVariables(environmentsData)
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
			environmentsData = append(environmentsData, newEnv)

			// Save environments
			if err := workspace.SaveEnvironments(environmentsData); err != nil {
				// Handle error
				return
			}

			// Refresh the environment dropdown
			updateEnvironmentDropdown(envDropdown, environmentsData)

			// Refresh the environment list in place
			newLeftPanel := createEnvironmentListPanel(
				backgroundColor,
				borderColor,
				borderFocusColor,
				titleColor,
				foregroundColor,
				buttonSelectedColor,
				environmentsData,
				func(env *workspace.Environment) {
					selectedEnvironment = env
					// Update the JSON editor with the selected environment's variables
					effectiveVars := env.GetEffectiveVariables(environmentsData)
					jsonBytes, _ := json.MarshalIndent(effectiveVars, "", "  ")
					jsonEditor.SetText(string(jsonBytes), false)
				},
				func() {
					// This will be replaced by the outer function
				},
				func(env *workspace.Environment) {
					// Handle rename environment - show rename form
					renameForm := createRenameEnvironmentPanel(
						backgroundColor,
						borderColor,
						borderFocusColor,
						titleColor,
						foregroundColor,
						env.Name,
						func(newName string) {
							// Update the environment name
							env.Name = newName

							// Save environments
							if err := workspace.SaveEnvironments(environmentsData); err != nil {
								// Handle error
								return
							}

							// Refresh the environment dropdown
							updateEnvironmentDropdown(envDropdown, environmentsData)

							// Refresh the environment list in place
						},
					)

					// Replace the left panel content with the rename form
					leftPanel.Clear()
					leftPanel.AddItem(renameForm, 0, 1, false)
					leftPanel.SetTitle(" Rename Environment ")

					// Clear the right panel
					jsonEditor.SetText("{}", false)
					selectedEnvironment = nil
				},
				func(env *workspace.Environment) {
					// Handle remove environment
					// Remove the environment from the slice
					for i, e := range environmentsData {
						if e.Name == env.Name {
							environmentsData = append(environmentsData[:i], environmentsData[i+1:]...)
							break
						}
					}

					// Save environments
					if err := workspace.SaveEnvironments(environmentsData); err != nil {
						// Handle error
						return
					}

					// Refresh the environment dropdown
					updateEnvironmentDropdown(envDropdown, environmentsData)

					// Close and reopen modal to refresh the list
					// Keep modal open - environment list will be refreshed when modal is reopened
				},
			)

			// Replace the left panel with the updated one
			content := tview.NewFlex().
				AddItem(newLeftPanel, 0, 4, false). // 40% for left panel
				AddItem(jsonEditor, 0, 6, false)    // 60% for JSON editor

			modal := createModal(content, 120, 40, backgroundColor)
			pages.RemovePage("envVariables")
			pages.AddPage("envVariables", modal, true, true)
			app.SetFocus(newLeftPanel)
		},
		func(env *workspace.Environment) {
			// Handle rename environment - show rename form
			renameForm := createRenameEnvironmentPanel(
				backgroundColor,
				borderColor,
				borderFocusColor,
				titleColor,
				foregroundColor,
				env.Name,
				func(newName string) {
					// Update the environment name
					env.Name = newName

					// Save environments
					if err := workspace.SaveEnvironments(environmentsData); err != nil {
						// Handle error
						return
					}

					// Refresh the environment dropdown
					updateEnvironmentDropdown(envDropdown, environmentsData)

					// Refresh the environment list in place
					newLeftPanel := createEnvironmentListPanel(
						backgroundColor,
						borderColor,
						borderFocusColor,
						titleColor,
						foregroundColor,
						buttonSelectedColor,
						environmentsData,
						func(env *workspace.Environment) {
							selectedEnvironment = env
							// Update the JSON editor with the selected environment's variables
							effectiveVars := env.GetEffectiveVariables(environmentsData)
							jsonBytes, _ := json.MarshalIndent(effectiveVars, "", "  ")
							jsonEditor.SetText(string(jsonBytes), false)
						},
						func() {
							// This will be replaced by the outer function
						},
						func(env *workspace.Environment) {
							// This will be replaced by the outer function
						},
						func(env *workspace.Environment) {
							// This will be replaced by the outer function
						},
					)

					// Replace the left panel with the updated one
					content := tview.NewFlex().
						AddItem(newLeftPanel, 0, 4, false). // 40% for left panel
						AddItem(jsonEditor, 0, 6, false)    // 60% for JSON editor

					modal := createModal(content, 120, 40, backgroundColor)
					pages.RemovePage("envVariables")
					pages.AddPage("envVariables", modal, true, true)
					app.SetFocus(newLeftPanel)
				},
			)

			// Replace the left panel content with the rename form
			leftPanel.Clear()
			leftPanel.AddItem(renameForm, 0, 1, false)
			leftPanel.SetTitle(" Rename Environment ")

			// Clear the right panel
			jsonEditor.SetText("{}", false)
			selectedEnvironment = nil
		},
		func(env *workspace.Environment) {
			// Handle remove environment
			// Remove the environment from the slice
			for i, e := range environmentsData {
				if e.Name == env.Name {
					environmentsData = append(environmentsData[:i], environmentsData[i+1:]...)
					break
				}
			}

			// Save environments
			if err := workspace.SaveEnvironments(environmentsData); err != nil {
				// Handle error
				return
			}

			// Refresh the environment dropdown
			updateEnvironmentDropdown(envDropdown, environmentsData)

			// Close and reopen modal to refresh the list
			// Keep modal open - environment list will be refreshed when modal is reopened
		},
	)

	// Create split layout: left 40%, right 60%
	content := tview.NewFlex().
		AddItem(leftPanel, 0, 4, false). // 40% for left panel
		AddItem(jsonEditor, 0, 6, false) // 60% for JSON editor

	modal := createModal(content, 120, 40, backgroundColor)
	pages.AddPage("envVariables", modal, true, true)
	app.SetFocus(jsonEditor)

	// Add keybinding to close modal with Escape, q, or Q
	pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Key() == 'q' || event.Key() == 'Q' {
			// Save the JSON back to environment variables if editing existing environment
			if selectedEnvironment != nil {
				jsonText := jsonEditor.GetText()
				var newVars map[string]string
				if err := json.Unmarshal([]byte(jsonText), &newVars); err != nil {
					// If JSON is invalid, keep the original variables
					// Could show an error message here
				} else {
					selectedEnvironment.Variables = newVars
					if saveErr := workspace.SaveEnvironments(environmentsData); saveErr != nil {
						// Handle save error
					}
				}
			}

			// Keep modal open - environment list will be refreshed when modal is reopened
			app.SetFocus(envDropdown)
			return nil
		}
		return event
	})
}
