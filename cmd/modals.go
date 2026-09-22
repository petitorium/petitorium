package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/workspace"
)

// modalSize defines a standard modal dimension. New modals must use one of
// the predefined sizes below via createSizedModal. Fall back to createModal
// with explicit dimensions only when the height must adapt to the content
// (e.g. a list sized to its number of entries).
type modalSize struct {
	Width  int
	Height int
}

// Standard modal sizes.
var (
	// modalSizeConfirm: yes/no confirmations, deletions, short notices, progress.
	modalSizeConfirm = modalSize{Width: 50, Height: 8}
	// modalSizeForm: single-field forms (rename, duplicate, create item, row editors).
	modalSizeForm = modalSize{Width: 50, Height: 10}
	// modalSizeEditor: multi-field forms (new request, duplicate request, move item).
	modalSizeEditor = modalSize{Width: 60, Height: 15}
	// modalSizeSearch: search overlays with a results list.
	modalSizeSearch = modalSize{Width: 100, Height: 20}
	// modalSizeLarge: complex editors (tag editor, command runner, file pickers).
	modalSizeLarge = modalSize{Width: 80, Height: 25}
	// modalSizeFullscreen: split-panel modals (environments, workspaces, marketplace).
	modalSizeFullscreen = modalSize{Width: 120, Height: 40}
)

// createSizedModal creates a centered modal dialog with a standard size.
func createSizedModal(p tview.Primitive, size modalSize, backgroundColor tcell.Color) tview.Primitive {
	return createModal(p, size.Width, size.Height, backgroundColor)
}

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
	} else if currentEnvIndex > 0 && currentEnvIndex <= len(*ui.EnvironmentsData) {
		// Other environment is selected - edit existing environment
		env = &(*ui.EnvironmentsData)[currentEnvIndex-1] // -1 because dropdown has "Base Environment" at index 0
	}

	currentEnvName := ""
	if env != nil {
		currentEnvName = env.Name
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

	// Create the marketplace-style environment table and search field.
	searchField := createSearchField(" Search Environments: ", "", ui.Colors)
	table := tview.NewTable().SetSelectable(true, false).SetFixed(1, 0)
	table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	table.SetBackgroundColor(ui.Colors.Background)

	var selectedEnvironment *workspace.Environment
	var filtered []workspace.Environment

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

	// filterEnvironments rebuilds the environment table for the given query.
	filterEnvironments := func(query string) {
		table.Clear()
		filtered = nil
		query = strings.ToLower(strings.TrimSpace(query))

		headers := []string{"Name", "Base", "Vars", "Current"}
		for i, h := range headers {
			table.SetCell(0, i, tview.NewTableCell(" "+h+" ").
				SetTextColor(ui.Colors.Title).
				SetSelectable(false).
				SetExpansion(1).
				SetAlign(tview.AlignCenter))
		}
		table.GetCell(0, 0).SetAlign(tview.AlignLeft)

		for _, e := range *ui.EnvironmentsData {
			if query != "" && !strings.Contains(strings.ToLower(e.Name), query) {
				continue
			}

			filtered = append(filtered, e)
			row := table.GetRowCount()

			base := e.Base
			if base == "" {
				base = "-"
			}
			currentText := ""
			if e.Name == currentEnvName {
				currentText = "*"
			}

			table.SetCell(row, 0, tview.NewTableCell(" "+e.Name+" ").
				SetExpansion(3).
				SetTextColor(ui.Colors.Foreground).
				SetAlign(tview.AlignLeft))
			table.SetCell(row, 1, tview.NewTableCell(" "+base+" ").
				SetExpansion(1).
				SetTextColor(ui.Colors.Foreground).
				SetAlign(tview.AlignCenter))
			table.SetCell(row, 2, tview.NewTableCell(" "+strconv.Itoa(len(e.Variables))+" ").
				SetExpansion(1).
				SetTextColor(ui.Colors.Foreground).
				SetAlign(tview.AlignCenter))
			table.SetCell(row, 3, tview.NewTableCell(" "+currentText+" ").
				SetExpansion(1).
				SetTextColor(ui.Colors.Foreground).
				SetAlign(tview.AlignCenter))
		}
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
	makeContentCapture := func() func(event *tcell.EventKey) *tcell.EventKey {
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
			// Tab: toggle focus between the environment table and the active env panel
			if event.Key() == tcell.KeyTab {
				if ui.App.GetFocus() == table {
					ui.App.SetFocus(activeEnvPanel())
				} else {
					ui.App.SetFocus(table)
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

	// buildModalContent assembles the modal content (search | table + editor + status bar).
	buildModalContent := func() *tview.Flex {
		tableAndEditor := tview.NewFlex().
			AddItem(table, 0, 2, true).
			AddItem(envContainer, 0, 3, false)

		c := tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(searchField, 1, 0, true).
			AddItem(tableAndEditor, 0, 1, false).
			AddItem(statusBar, 1, 0, false)
		c.SetBackgroundColor(ui.Colors.Background)
		c.SetBorder(true).SetTitle(" Environment Configuration ")
		c.SetBorderColor(ui.Colors.BorderFocus)
		c.SetTitleColor(ui.Colors.Title)
		c.SetInputCapture(makeContentCapture())
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
		modal := createSizedModal(form, modalSizeConfirm, ui.Colors.Background)
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
		modal := createSizedModal(form, modalSizeForm, ui.Colors.Background)
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
		modal := createSizedModal(form, modalSizeForm, ui.Colors.Background)
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
			return
		}

		// Refresh the environment dropdown
		updateEnvironmentDropdown(ui.EnvDropdown, *ui.EnvironmentsData)

		// Refresh the environment table in place and select the new environment.
		filterEnvironments(searchField.GetText())
		for i, e := range filtered {
			if e.Name == "New Environment" {
				table.Select(i+1, 0)
				break
			}
		}
		ui.App.SetFocus(table)
	}

	table.SetSelectionChangedFunc(func(row, column int) {
		if row > 0 && row-1 < len(filtered) {
			onEnvironmentSelected(&filtered[row-1])
		}
	})

	table.SetSelectedFunc(func(row, column int) {
		if row > 0 && row-1 < len(filtered) {
			onEnvironmentChosen(filtered[row-1].Name)
		}
	})

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		if event.Key() == tcell.KeyUp && row == 1 {
			ui.App.SetFocus(searchField)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			ui.App.SetFocus(searchField)
			return nil
		}
		if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		switch event.Rune() {
		case 'N':
			onCreateNew()
			return nil
		case 'd':
			if row > 0 && row-1 < len(filtered) {
				env := &filtered[row-1]
				if env.Name != "Base" {
					onDelete(env)
				}
			}
			return nil
		case 'r':
			if row > 0 && row-1 < len(filtered) {
				env := &filtered[row-1]
				if env.Name != "Base" {
					onRename(env)
				}
			}
			return nil
		case 'c', 'C':
			if row > 0 && row-1 < len(filtered) {
				env := &filtered[row-1]
				if env.Name != "Base" {
					onClone(env)
				}
			}
			return nil
		}
		return event
	})

	searchField.SetChangedFunc(func(text string) {
		filterEnvironments(text)
	})

	searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyDown {
			if table.GetRowCount() > 1 {
				ui.App.SetFocus(table)
			}
			return nil
		}
		return event
	})

	// Build the modal content (search | table + editor + status bar) with key handling.
	content := buildModalContent()

	// Populate the table and select the active environment.
	filterEnvironments("")
	if env != nil {
		for i, e := range filtered {
			if e.Name == env.Name {
				table.Select(i+1, 0)
				break
			}
		}
	} else if len(filtered) > 0 {
		table.Select(1, 0)
	}

	modal := createSizedModal(content, modalSizeFullscreen, ui.Colors.Background)
	ui.Pages.AddPage("envVariables", modal, true, true)
	ui.EnvModalEditor = envEditPanel
	ui.EnvModalViewPanel = envViewPanel
	ui.EnvModalContainer = envContainer
	ui.EnvModalEditMode = false
	ui.UpdateFooter()
	ui.App.SetFocus(searchField)
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

	modal := createSizedModal(textView, modalSizeConfirm, colors.Background)
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

	modal := createSizedModal(flex, modalSizeConfirm, colors.Background)
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
