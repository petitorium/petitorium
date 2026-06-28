package cmd

import (
	"encoding/json"
	"fmt"
	"time"

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

	// Determine initial JSON content (own variables, not effective/merged)
	var jsonBytes []byte
	if env != nil {
		jsonBytes, _ = json.MarshalIndent(env.Variables, "", "  ")
	} else {
		jsonBytes = []byte("{}")
	}
	initialContent := string(jsonBytes)

	// Create the view panel (highlighted, read-only) and the edit panel (raw).
	envViewPanel := createPanel(" Environment Variables (VIEW) ", ui.Colors, &PanelOptions{HasBorder: &[]bool{true}[0]})
	envEditPanel := createTextArea(" Environment Variables (EDIT) ", ui.Colors.Background, ui.Colors.Border, ui.Colors.Title, ui.Colors.Foreground)
	envEditPanel.SetWordWrap(false)

	// Highlight the border of the focused env panel (view or edit), matching the
	// environment list panel's focus/blur behaviour.
	envViewPanel.SetFocusFunc(func() { envViewPanel.SetBorderColor(ui.Colors.BorderFocus) })
	envViewPanel.SetBlurFunc(func() { envViewPanel.SetBorderColor(ui.Colors.Border) })
	envEditPanel.SetFocusFunc(func() { envEditPanel.SetBorderColor(ui.Colors.BorderFocus) })
	envEditPanel.SetBlurFunc(func() { envEditPanel.SetBorderColor(ui.Colors.Border) })

	// Swap container between view and edit. View mode is the default.
	envContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	envContainer.SetBackgroundColor(ui.Colors.Background)
	envContainer.AddItem(envViewPanel, 0, 1, false)

	// Modal state
	ui.EnvModalCurrentContent = initialContent
	ui.EnvModalEditMode = false

	// SyncEnvModalContent renders content into the currently active panel.
	ui.SyncEnvModalContent = func(content string) {
		ui.EnvModalCurrentContent = content
		if ui.EnvModalEditMode {
			envEditPanel.SetText(content, false)
		} else {
			envViewPanel.Clear()
			if content != "" {
				envViewPanel.SetText(FormatBodyContentWithVariables(content))
				envViewPanel.SetTextAlign(tview.AlignLeft)
			} else {
				envViewPanel.SetText("")
				envViewPanel.SetTextAlign(tview.AlignLeft)
			}
		}
	}

	// SwitchEnvModalMode toggles between view (highlighted) and edit (raw).
	ui.SwitchEnvModalMode = func() {
		ui.EnvModalEditMode = !ui.EnvModalEditMode
		envContainer.Clear()
		if ui.EnvModalEditMode {
			envContainer.AddItem(envEditPanel, 0, 1, false)
			envEditPanel.SetText(ui.EnvModalCurrentContent, false)
			ui.App.SetFocus(envEditPanel)
		} else {
			ui.EnvModalCurrentContent = envEditPanel.GetText()
			envContainer.AddItem(envViewPanel, 0, 1, false)
			ui.SyncEnvModalContent(ui.EnvModalCurrentContent)
			ui.App.SetFocus(envViewPanel)
		}
	}

	// Initialize both panels and render the view.
	envEditPanel.SetText(initialContent, false)
	ui.SyncEnvModalContent(initialContent)

	// Create error display
	errorText := tview.NewTextView()
	errorText.SetTextColor(ui.Colors.BorderFocus) // Red color for errors
	errorText.SetBackgroundColor(ui.Colors.Background)
	errorText.SetDynamicColors(true)
	errorText.SetText("")

	// Create status bar with error display
	statusBar := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(errorText, 1, 0, false)

	// Create left panel (environment list)
	var selectedEnvironment *workspace.Environment
	var leftPanel *tview.List

	// Define callbacks
	onEnvironmentSelected := func(env *workspace.Environment) {
		selectedEnvironment = env
		var content string
		if env != nil {
			b, _ := json.MarshalIndent(env.Variables, "", "  ")
			content = string(b)
		} else {
			content = "{}"
		}
		envEditPanel.SetText(content, false)
		ui.SyncEnvModalContent(content)
	}

	// Function to save environment variables
	saveEnvironmentVariables := func() error {
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

		if envToSave == nil {
			return fmt.Errorf("no environment selected to save")
		}

		jsonText := sanitizeCommandRunnerTagsInJSON(envEditPanel.GetText())
		var newVars map[string]string
		if err := json.Unmarshal([]byte(jsonText), &newVars); err != nil {
			return fmt.Errorf("invalid JSON: %v", err)
		}

		envToSave.Variables = newVars
		ui.WorkspaceData.Environments = *ui.EnvironmentsData
		if saveErr := workspace.SaveWorkspace(ui.WorkspaceData); saveErr != nil {
			return fmt.Errorf("failed to save workspace: %v", saveErr)
		}

		return nil
	}

	// closeEnvModal tears down the modal and restores focus to the config button.
	closeEnvModal := func() {
		ui.EnvModalEditor = nil
		ui.EnvModalViewPanel = nil
		ui.EnvModalContainer = nil
		ui.EnvModalEditMode = false
		ui.SyncEnvModalContent = nil
		ui.SwitchEnvModalMode = nil
		ui.Pages.RemovePage("envVariables")
		ui.App.SetFocus(ui.EnvConfigButton)
	}

	// openExternalEnvEditor launches $EDITOR on the current content and
	// returns to view mode to show the highlighted result.
	openExternalEnvEditor := func() {
		currentContent := envEditPanel.GetText()
		ui.App.Suspend(func() {
			modifiedContent, err := openInExternalEditor(currentContent, "json")
			if err != nil {
				errorText.SetText(fmt.Sprintf("Error opening external editor: %v", err))
				return
			}
			errorText.SetText("")
			ui.EnvModalCurrentContent = modifiedContent
			envEditPanel.SetText(modifiedContent, false)
			if ui.EnvModalEditMode {
				ui.SwitchEnvModalMode() // edit -> view (syncs highlighted view)
			} else {
				ui.SyncEnvModalContent(modifiedContent)
			}
		})
	}

	// activeEnvPanel returns the currently visible env panel.
	activeEnvPanel := func() tview.Primitive {
		if ui.EnvModalEditMode {
			return envEditPanel
		}
		return envViewPanel
	}

	// makeContentCapture wires Tab/Esc/q/i/F4 handling for the modal content.
	makeContentCapture := func(lp *tview.List) func(event *tcell.EventKey) *tcell.EventKey {
		return func(event *tcell.EventKey) *tcell.EventKey {
			// F4: external editor (only when an env panel is focused)
			if event.Key() == tcell.KeyF4 {
				f := ui.App.GetFocus()
				if f == envViewPanel || f == envEditPanel {
					openExternalEnvEditor()
					return nil
				}
				return event
			}
			// Tab: toggle focus between the list and the active env panel
			if event.Key() == tcell.KeyTab {
				if ui.App.GetFocus() == lp {
					ui.App.SetFocus(activeEnvPanel())
				} else {
					ui.App.SetFocus(lp)
				}
				return nil
			}
			// 'i': enter edit mode (only from the view panel)
			if event.Rune() == 'i' && ui.App.GetFocus() == envViewPanel {
				ui.SwitchEnvModalMode()
				return nil
			}
			// Esc: exit edit -> view when editing; otherwise save + close
			if event.Key() == tcell.KeyEsc {
				if ui.App.GetFocus() == envEditPanel {
					ui.SwitchEnvModalMode()
					return nil
				}
				if err := saveEnvironmentVariables(); err != nil {
					errorText.SetText(fmt.Sprintf("Error saving: [red]%v", err))
					return nil
				}
				closeEnvModal()
				return nil
			}
			// 'q': save + close (not while editing, where 'q' types)
			if event.Rune() == 'q' && ui.App.GetFocus() != envEditPanel {
				if err := saveEnvironmentVariables(); err != nil {
					errorText.SetText(fmt.Sprintf("Error saving: [red]%v", err))
					return nil
				}
				closeEnvModal()
				return nil
			}
			return event
		}
	}

	// buildModalContent assembles the modal content (list | env container + status bar).
	buildModalContent := func(lp *tview.List) *tview.Flex {
		c := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewFlex().
				AddItem(lp, 0, 4, false).
				AddItem(envContainer, 0, 6, false),
				0, 1, false).
			AddItem(statusBar, 1, 0, false)
		c.SetInputCapture(makeContentCapture(lp))
		return c
	}

	onEnvironmentChosen := func(name string) {
		// 1. Save changes if any
		if err := saveEnvironmentVariables(); err != nil {
			errorText.SetText(fmt.Sprintf("Error saving: %v", err))
			return
		}

		// 2. Switch environment
		if name == "Base" {
			ui.EnvDropdown.SetCurrentOption(0)
		} else {
			for i, env := range *ui.EnvironmentsData {
				if env.Name == name {
					ui.EnvDropdown.SetCurrentOption(i + 1)
					break
				}
			}
		}

		// 3. Close modal
		closeEnvModal()
	}

	var onCreateNew func()
	var onDelete func(*workspace.Environment)
	var onRename func(*workspace.Environment)
	var onClone func(*workspace.Environment)

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
		modal := createModal(form, 24, 10, ui.Colors.Background)
		ui.Pages.AddPage("renameEnvironment", modal, true, true)
		ui.App.SetFocus(form)
	}

	onClone = func(env *workspace.Environment) {
		currentFocus := ui.App.GetFocus()
		form := createCloneEnvironmentForm(ui.App, ui.Pages, env, ui.EnvironmentsData, ui.WorkspaceData, ui.EnvDropdown, ui.EnvConfigButton, ui.Colors, currentFocus)
		form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyEsc {
				ui.Pages.RemovePage("cloneEnvironment")
				ui.App.SetFocus(currentFocus)
				return nil
			}
			return event
		})
		modal := createModal(form, 24, 10, ui.Colors.Background)
		ui.Pages.AddPage("cloneEnvironment", modal, true, true)
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
			onEnvironmentChosen,
			onCreateNew,
			onDelete,
			onRename,
			onClone,
		)

		// Select the newly created environment (index = number of environments, since 0 is "Create New")
		newLeftPanel.SetCurrentItem(len(*ui.EnvironmentsData))

		// Rebuild the modal content with the refreshed list.
		content := buildModalContent(newLeftPanel)

		modal := createModal(content, 120, 40, ui.Colors.Background)
		ui.EnvModalEditor = nil
		ui.Pages.RemovePage("envVariables")
		ui.Pages.AddPage("envVariables", modal, true, true)
		ui.EnvModalEditor = envEditPanel
		ui.EnvModalViewPanel = envViewPanel
		ui.EnvModalContainer = envContainer
		ui.EnvModalEditMode = false
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
		onEnvironmentChosen,
		onCreateNew,
		onDelete,
		onRename,
		onClone,
	)

	// Build the modal content (list | env container + status bar) with key handling.
	content := buildModalContent(leftPanel)

	// Find the index of the selected environment in the list (add 1 because index 0 is "Create New Environment")
	if env != nil {
		for i, listEnv := range *ui.EnvironmentsData {
			if listEnv.Name == env.Name {
				leftPanel.SetCurrentItem(i + 1) // +1 because index 0 is "Create New Environment"
				break
			}
		}
	}

	modal := createModal(content, 120, 40, ui.Colors.Background)
	ui.Pages.AddPage("envVariables", modal, true, true)
	ui.EnvModalEditor = envEditPanel
	ui.EnvModalViewPanel = envViewPanel
	ui.EnvModalContainer = envContainer
	ui.EnvModalEditMode = false
	ui.UpdateFooter()
	ui.App.SetFocus(leftPanel)
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
	workspaceInfo := createPanel(" Workspace Information ", ui.Colors, nil)
	workspaceInfo.SetText("Select a workspace to view its information")

	// Create left panel (workspace list)
	var leftPanel *tview.List

	// Define callbacks
	onWorkspaceSelected := func(ws *workspace.WorkspaceMetadata) {
		if ws != nil {
			// Load full workspace data to show information
			fullWorkspace, err := workspace.LoadWorkspaceByName(ws.Name)
			if err != nil {
				workspaceInfo.SetText(fmt.Sprintf("Error loading workspace: %v", err))
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
			workspaceInfo.SetText(info)
		} else {
			workspaceInfo.SetText("Select a workspace to view its information")
		}
	}

	var onCreateNew func()
	var onDelete func(*workspace.WorkspaceMetadata)
	var onRename func(*workspace.WorkspaceMetadata)
	var onDuplicate func(*workspace.WorkspaceMetadata)

	onDelete = func(ws *workspace.WorkspaceMetadata) {
		currentFocus := ui.App.GetFocus()
		form := createDeleteWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceConfigButton, ui.Colors)
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
		form := createRenameWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceConfigButton, ui.Colors)
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
		form := createDuplicateWorkspaceForm(ui.App, ui.Pages, ws.Name, ui.WorkspaceConfigButton, ui.Colors)
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
		form := createNewWorkspaceForm(ui.App, ui.Pages, ui, ui.RootNode, ui.CollectionsTreeView, ui.Colors)
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
		func(name string) {
			ui.SwitchWorkspace(name)
			ui.Pages.RemovePage("workspaceModal")
			ui.App.SetFocus(ui.WorkspaceConfigButton)
		},
		onCreateNew,
		onDelete,
		onRename,
		onDuplicate,
	)

	// Create split layout: left 40%, right 60%
	content := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewFlex().
			AddItem(leftPanel, 0, 4, false).     // 40% for left panel
			AddItem(workspaceInfo, 0, 6, false), // 60% for workspace info
			0, 1, false)

	content.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
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

// showErrorModal displays an error message modal
func showErrorModal(pages *tview.Pages, message string, colors *ColorManager) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
		}).
		SetBackgroundColor(colors.Background).
		SetTextColor(colors.Foreground).
		SetButtonBackgroundColor(colors.ButtonBackground).
		SetButtonTextColor(colors.Foreground)
	pages.AddPage("error", modal, true, true)
}

// showSuccessModal displays a success message modal
func showSuccessModal(pages *tview.Pages, message string, colors *ColorManager) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("success")
		}).
		SetBackgroundColor(colors.Background).
		SetTextColor(colors.Foreground).
		SetButtonBackgroundColor(colors.ButtonBackground).
		SetButtonTextColor(colors.Foreground)
	pages.AddPage("success", modal, true, true)
}

// showSuccessModalWithFocus displays a success message modal and returns focus to a specific primitive
func showSuccessModalWithFocus(app *tview.Application, pages *tview.Pages, message string, returnFocus tview.Primitive, colors *ColorManager) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("success")
			if returnFocus != nil {
				app.SetFocus(returnFocus)
			}
		}).
		SetBackgroundColor(colors.Background).
		SetTextColor(colors.Foreground).
		SetButtonBackgroundColor(colors.ButtonBackground).
		SetButtonTextColor(colors.Foreground)
	pages.AddPage("success", modal, true, true)
}

// showErrorModalWithFocus displays an error message modal and returns focus to a specific primitive
func showErrorModalWithFocus(app *tview.Application, pages *tview.Pages, message string, returnFocus tview.Primitive, colors *ColorManager) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("error")
			if returnFocus != nil {
				app.SetFocus(returnFocus)
			}
		}).
		SetBackgroundColor(colors.Background).
		SetTextColor(colors.Foreground).
		SetButtonBackgroundColor(colors.ButtonBackground).
		SetButtonTextColor(colors.Foreground)
	pages.AddPage("error", modal, true, true)
}

// showProgressModal displays a progress modal with an animated progress bar.
// Returns the text view and a stop function that should be called when the operation completes.
func showProgressModal(app *tview.Application, pages *tview.Pages, title string, message string, colors *ColorManager) (*tview.TextView, func()) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	textView.SetBackgroundColor(colors.Background)
	textView.SetTextColor(colors.Foreground)
	textView.SetBorder(true)
	textView.SetTitle(title)
	textView.SetTitleColor(colors.Title)
	textView.SetBorderColor(colors.BorderFocus)

	barFrames := []string{
		"[#bb9af8]      ",
		"[#bb9af8]█     ",
		"[#bb9af8]██    ",
		"[#bb9af8]███   ",
		"[#bb9af8]████  ",
		"[#bb9af8]█████ ",
		"[#bb9af8]██████",
		"[#bb9af8]█████ ",
		"[#bb9af8]████  ",
		"[#bb9af8]███   ",
		"[#bb9af8]██    ",
		"[#bb9af8]█     ",
	}
	spinner := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	stopBar := make(chan struct{})
	go func() {
		i := 0
		for {
			select {
			case <-stopBar:
				return
			default:
				app.QueueUpdateDraw(func() {
					textView.SetText(fmt.Sprintf("\n  %s %s\n\n  %s",
						spinner[i%len(spinner)], message, barFrames[i%len(barFrames)]))
				})
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	modal := createModal(textView, 40, 7, colors.Background)
	pages.AddPage("progress", modal, true, true)

	stop := func() {
		close(stopBar)
	}
	return textView, stop
}

// showConfirmModal displays a confirmation modal with buttons
func showConfirmModal(
	pages *tview.Pages,
	title string,
	message string,
	buttons []string,
	colors *ColorManager,
	onButton func(buttonIndex int),
) *tview.Form {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(message)
	textView.SetBackgroundColor(colors.Background)
	textView.SetTextColor(colors.Foreground)

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	for i, btn := range buttons {
		btn := btn
		i := i
		form.AddButton(btn, func() {
			pages.RemovePage("confirm")
			onButton(i)
		})
	}

	form.SetCancelFunc(func() {
		pages.RemovePage("confirm")
		onButton(0)
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(colors.Background)
	flex.AddItem(textView, 0, 1, false)
	flex.AddItem(form, 0, 1, false)

	flex.SetBorder(true).SetTitle(title)
	flex.SetBorderColor(colors.BorderFocus)
	flex.SetTitleColor(colors.Title)

	modal := createModal(flex, 50, 10, colors.Background)
	pages.AddPage("confirm", modal, true, true)

	return form
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
