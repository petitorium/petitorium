package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/workspace"
)

// showEnvironmentModal displays the environment variables modal
func showEnvironmentModal(
	ui *UIOrchestrator,
) {
	// Get current selected environment
	currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
	var env *workspace.Environment

	if currentEnvIndex == 0 {
		// Base Environment selected - find and edit the "Base" environment
		for i := range *ui.EnvironmentsData {
			if (*ui.EnvironmentsData)[i].Name == "Base" {
				env = &(*ui.EnvironmentsData)[i]
				break
			}
		}
		// If Base environment doesn't exist, create new environment
		if env == nil {
			env = nil
		}
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
		// Other environment is selected - edit existing environment
		env = &(*ui.EnvironmentsData)[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0
	} else {
		// Fallback - create new environment
		env = nil
	}

	// Create JSON editor for environment variables
	var jsonBytes []byte
	if env != nil {
		// Convert existing environment variables to JSON (own variables, not effective)
		jsonBytes, _ = json.MarshalIndent(env.Variables, "", "  ")
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

	// Define callbacks
	onEnvironmentSelected := func(env *workspace.Environment) {
		selectedEnvironment = env
		if env != nil {
			// Update the JSON editor with the selected environment's own variables (not effective/merged)
			jsonBytes, _ := json.MarshalIndent(env.Variables, "", "  ")
			jsonEditor.SetText(string(jsonBytes), false)
		} else {
			// Clear the JSON editor when no environment is selected
			jsonEditor.SetText("{}", false)
		}
	}

	var onCreateNew func()
	var onDelete func(*workspace.Environment)
	var onRename func(*workspace.Environment)

	onDelete = func(env *workspace.Environment) {
		currentFocus := ui.App.GetFocus()
		form := createDeleteEnvironmentConfirm(ui.App, ui.Pages, env, ui.EnvironmentsData, ui.WorkspaceData, ui.EnvDropdown, ui.EnvConfigButton, ui.Colors, currentFocus)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("deleteEnvironment")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 50, 8, ui.Colors.Background)
		ui.Pages.AddPage("deleteEnvironment", modal, true, true)
		ui.App.SetFocus(form)
	}

	onRename = func(env *workspace.Environment) {
		currentFocus := ui.App.GetFocus()
		form := createRenameEnvironmentForm(ui.App, ui.Pages, env, ui.EnvironmentsData, ui.WorkspaceData, ui.EnvDropdown, ui.EnvConfigButton, ui.Colors, currentFocus)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("renameEnvironment")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 25, 10, ui.Colors.Background)
		ui.Pages.AddPage("renameEnvironment", modal, true, true)
		ui.App.SetFocus(form)
	}

	onCreateNew = func() {
		// Handle create new environment - directly create "New Environment"
		newEnv := workspace.Environment{
			Name:      "New Environment",
			Base:      "Base",
			Variables: make(map[string]string),
		}

		// Add to environments
		*ui.EnvironmentsData = append(*ui.EnvironmentsData, newEnv)

		// Save workspace data (which includes environments)
		ui.WorkspaceData.Environments = *ui.EnvironmentsData
		if err := workspace.SaveWorkspace(ui.WorkspaceData); err != nil {
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
			onEnvironmentSelected,
			onCreateNew,
			onDelete,
			onRename,
		)

		// Select the newly created environment (index = number of environments, since 0 is "Create New")
		newLeftPanel.SetCurrentItem(len(*ui.EnvironmentsData))

		// Replace the left panel with the updated one
		content := tview.NewFlex().
			AddItem(newLeftPanel, 0, 4, false). // 40% for left panel
			AddItem(jsonEditor, 0, 6, false)    // 60% for JSON editor

		content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyTab {
				if ui.App.GetFocus() == newLeftPanel {
					ui.App.SetFocus(jsonEditor)
				} else {
					ui.App.SetFocus(newLeftPanel)
				}
				return nil
			}
			return event
		})

		modal := createModal(content, 120, 40, ui.Colors.Background)
		ui.Pages.RemovePage("envVariables")
		ui.Pages.AddPage("envVariables", modal, true, true)
		ui.UpdateFooter()
		ui.App.SetFocus(newLeftPanel)
	}

	leftPanel = createEnvironmentListPanel(
		ui.Colors.Background,
		ui.Colors.Border,
		ui.Colors.BorderFocus,
		ui.Colors.Title,
		ui.Colors.Foreground,
		ui.Colors.ButtonSelect,
		*ui.EnvironmentsData,
		onEnvironmentSelected,
		onCreateNew,
		onDelete,
		onRename,
	)

	// Function to save environment variables
	saveEnvironmentVariables := func() {
		// Determine which environment to save
		var envToSave *workspace.Environment
		if selectedEnvironment != nil {
			// Use the environment selected in the list
			envToSave = selectedEnvironment
		} else {
			// Fallback: use the environment selected in the dropdown
			currentEnvIndex, _ := ui.EnvDropdown.GetCurrentOption()
			if currentEnvIndex == 0 {
				// Base environment
				for i := range *ui.EnvironmentsData {
					if (*ui.EnvironmentsData)[i].Name == "Base" {
						envToSave = &(*ui.EnvironmentsData)[i]
						break
					}
				}
			} else if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
				// Other environment
				envToSave = &(*ui.EnvironmentsData)[currentEnvIndex-1]
			}
		}

		if envToSave != nil {
			jsonText := jsonEditor.GetText()
			var newVars map[string]string
			if err := json.Unmarshal([]byte(jsonText), &newVars); err != nil {
				// If JSON is invalid, keep the original variables
			} else {
				envToSave.Variables = newVars
				ui.WorkspaceData.Environments = *ui.EnvironmentsData
				if saveErr := workspace.SaveWorkspace(ui.WorkspaceData); saveErr != nil {
					// Handle save error - could show a message but for now ignore
				}
			}
		}
	}

	// Create save button
	saveButton := createButton("Save", ui.Colors)
	saveButton.SetSelectedFunc(func() {
		// Save environment variables
		saveEnvironmentVariables()
		// Keep modal open so user can continue editing
	})

	// Create button container
	buttonContainer := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().SetBackgroundColor(ui.Colors.Background), 0, 1, false).
		AddItem(saveButton, 10, 0, true).
		AddItem(tview.NewBox().SetBackgroundColor(ui.Colors.Background), 0, 1, false)

	// Create split layout: left 40%, right 60%, with button bar at bottom
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(leftPanel, 0, 4, false).  // 40% for left panel
			AddItem(jsonEditor, 0, 6, false), // 60% for JSON editor
							0, 9, false).
		AddItem(buttonContainer, 1, 0, false) // Button bar at bottom

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if ui.App.GetFocus() == leftPanel {
				ui.App.SetFocus(jsonEditor)
			} else {
				ui.App.SetFocus(leftPanel)
			}
			return nil
		}
		return event
	})

	modal := createModal(content, 120, 40, ui.Colors.Background)
	ui.Pages.AddPage("envVariables", modal, true, true)
	ui.UpdateFooter()

	// Set initial selection on the environment list to match the currently selected environment
	if env != nil {
		// Find the index of the selected environment in the list (add 1 because index 0 is "Create New Environment")
		for i, listEnv := range *ui.EnvironmentsData {
			if listEnv.Name == env.Name {
				leftPanel.SetCurrentItem(i + 1) // +1 because index 0 is "Create New Environment"
				break
			}
		}
	}

	ui.App.SetFocus(leftPanel)

	// Add keybinding to close modal with Escape, q, or Q
	ui.Pages.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		currentPage, _ := ui.Pages.GetFrontPage()
		if currentPage != "envVariables" {
			return event
		}
		// Handle modal closing with Escape
		if event.Key() == tcell.KeyEscape {
			// Save environment variables before closing
			saveEnvironmentVariables()

			// Close modal and return focus to config button
			ui.Pages.RemovePage("envVariables")
			ui.UpdateFooter()
			ui.App.SetFocus(ui.EnvConfigButton)
			return nil
		}
		return event
	})
}

// showWorkspaceModal displays a modal for workspace configuration with a split-panel layout
func showWorkspaceModal(
	ui *UIOrchestrator,
) {
	// Load workspace manager
	manager, err := workspace.LoadWorkspaceManager()
	if err != nil {
		// Handle error - could show a message but for now just return
		return
	}

	// Create workspace info display (right panel)
	workspaceInfo := createTextArea(" Workspace Information ", ui.Colors.Background, ui.Colors.Border, ui.Colors.Title, ui.Colors.Foreground)
	workspaceInfo.SetText("Select a workspace to view its information", false)
	workspaceInfo.SetBorder(true)

	// Create left panel (workspace list)
	var selectedWorkspace *workspace.WorkspaceMetadata
	var leftPanel *tview.List

	// Define callbacks
	onWorkspaceSelected := func(ws *workspace.WorkspaceMetadata) {
		selectedWorkspace = ws
		if ws != nil {
			// Load full workspace data to show information
			fullWorkspace, err := workspace.LoadWorkspaceByName(ws.Name)
			if err != nil {
				workspaceInfo.SetText(fmt.Sprintf("Error loading workspace: %v", err), false)
				return
			}

			// Display workspace information
			info := fmt.Sprintf("Name: %s\nDescription: %s\nCollections: %d\nEnvironments: %d\nCreated: %s\nUpdated: %s",
				fullWorkspace.Name,
				fullWorkspace.Description,
				len(fullWorkspace.Collections),
				len(fullWorkspace.Environments),
				fullWorkspace.CreatedAt.Format("2006-01-02 15:04:05"),
				fullWorkspace.UpdatedAt.Format("2006-01-02 15:04:05"))
			workspaceInfo.SetText(info, false)
		} else {
			workspaceInfo.SetText("Select a workspace to view its information", false)
		}
	}

	var onCreateNew func()
	var onDelete func(*workspace.WorkspaceMetadata)
	var onRename func(*workspace.WorkspaceMetadata)
	var onDuplicate func(*workspace.WorkspaceMetadata)

	onDelete = func(ws *workspace.WorkspaceMetadata) {
		currentFocus := ui.App.GetFocus()
		form := createDeleteWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceSelector, ui.Colors)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("deleteWorkspace")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 50, 8, ui.Colors.Background)
		ui.Pages.AddPage("deleteWorkspace", modal, true, true)
		ui.App.SetFocus(form)
	}

	onRename = func(ws *workspace.WorkspaceMetadata) {
		currentFocus := ui.App.GetFocus()
		form := createRenameWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceSelector, ui.Colors)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("renameWorkspace")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 50, 8, ui.Colors.Background)
		ui.Pages.AddPage("renameWorkspace", modal, true, true)
		ui.App.SetFocus(form)
	}

	onDuplicate = func(ws *workspace.WorkspaceMetadata) {
		currentFocus := ui.App.GetFocus()
		form := createDuplicateWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceSelector, ui.Colors)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("duplicateWorkspace")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 50, 10, ui.Colors.Background)
		ui.Pages.AddPage("duplicateWorkspace", modal, true, true)
		ui.App.SetFocus(form)
	}

	onCreateNew = func() {
		currentFocus := ui.App.GetFocus()
		form := createNewWorkspaceForm(ui.App, ui.Pages, ui.WorkspaceSelector, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("createWorkspace")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 50, 8, ui.Colors.Background)
		ui.Pages.AddPage("createWorkspace", modal, true, true)
		ui.App.SetFocus(form)
	}

	leftPanel = createWorkspaceListPanel(
		ui.Colors.Background,
		ui.Colors.Border,
		ui.Colors.BorderFocus,
		ui.Colors.Title,
		ui.Colors.Foreground,
		ui.Colors.ButtonSelect,
		manager.Workspaces,
		manager.CurrentWorkspace,
		onWorkspaceSelected,
		onCreateNew,
		onDelete,
		onRename,
		onDuplicate,
	)

	// Create switch workspace button
	switchButton := createButton("Switch to Selected", ui.Colors)
	switchButton.SetSelectedFunc(func() {
		if selectedWorkspace != nil && selectedWorkspace.Name != manager.CurrentWorkspace {
			// Switch to the selected workspace
			if err := workspace.SwitchWorkspace(selectedWorkspace.Name); err != nil {
				// Handle error - could show message
				return
			}

			// Update UI
			ui.WorkspaceSelector.SetCurrentOption(findWorkspaceIndex(manager.Workspaces, selectedWorkspace.Name))

			// Reload workspace data
			if newWorkspace, err := workspace.LoadWorkspace(); err == nil {
				ui.WorkspaceData = newWorkspace
				ui.EnvironmentsData = &newWorkspace.Environments
				updateEnvironmentDropdown(ui.EnvDropdown, *ui.EnvironmentsData)
				refreshCollectionsTree(ui)
			}

			// Close modal
			ui.Pages.RemovePage("workspaceModal")
			ui.App.SetFocus(ui.WorkspaceConfigButton)
			ui.UpdateFooter()
		}
	})

	// Create close button
	closeButton := createButton("Close", ui.Colors)
	closeButton.SetSelectedFunc(func() {
		ui.Pages.RemovePage("workspaceModal")
		ui.App.SetFocus(ui.WorkspaceConfigButton)
	})

	// Create button container
	buttonContainer := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewBox().SetBackgroundColor(ui.Colors.Background), 0, 1, false).
		AddItem(switchButton, 18, 0, true).
		AddItem(tview.NewBox().SetBackgroundColor(ui.Colors.Background), 0, 1, false).
		AddItem(closeButton, 8, 0, true).
		AddItem(tview.NewBox().SetBackgroundColor(ui.Colors.Background), 0, 1, false)

	// Create split layout: left 40%, right 60%, with button bar at bottom
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(leftPanel, 0, 4, false).     // 40% for left panel
			AddItem(workspaceInfo, 0, 6, false), // 60% for workspace info
							0, 9, false).
		AddItem(buttonContainer, 1, 0, false) // Button bar at bottom

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyTab {
			if ui.App.GetFocus() == leftPanel {
				ui.App.SetFocus(workspaceInfo)
			} else {
				ui.App.SetFocus(leftPanel)
			}
			return nil
		}
		if event.Key() == tcell.KeyEsc {
			ui.Pages.RemovePage("workspaceModal")
			ui.App.SetFocus(ui.WorkspaceConfigButton)
			return nil
		}
		return event
	})

	modal := createModal(content, 120, 40, ui.Colors.Background)
	ui.Pages.RemovePage("workspaceModal")
	ui.Pages.AddPage("workspaceModal", modal, true, true)
	ui.UpdateFooter()
	ui.App.SetFocus(leftPanel)
}

// Helper function to find workspace index in dropdown
func findWorkspaceIndex(workspaces []workspace.WorkspaceMetadata, name string) int {
	for i, ws := range workspaces {
		if ws.Name == name {
			return i + 1 // +1 because dropdown has "Create New Workspace" at index 0
		}
	}
	return 0
}
