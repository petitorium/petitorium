package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

// createInlineEditInput creates an input field for inline editing of environment names
func createInlineEditInput(
	colors *ColorManager,
	currentName string,
	onSave func(string),
	onCancel func(),
) *tview.InputField {
	editInput := createInputField("", colors)
	editInput.SetText(currentName)
	editInput.SetBorder(false)
	editInput.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			newName := strings.TrimSpace(editInput.GetText())
			if newName != "" && newName != currentName {
				onSave(newName)
			} else {
				onCancel()
			}
		} else {
			onCancel()
		}
	})
	return editInput
}

// createRenameEnvironmentPanel creates a simple panel for renaming environments
func createRenameEnvironmentPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor tcell.Color,
	currentName string,
	onRename func(string),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(backgroundColor)
	form.SetBorderColor(borderColor)
	form.SetTitleColor(titleColor)
	form.SetFieldBackgroundColor(backgroundColor)
	form.SetFieldTextColor(foregroundColor)
	form.SetLabelColor(foregroundColor)
	form.SetButtonBackgroundColor(backgroundColor)
	form.SetButtonTextColor(foregroundColor)

	form.AddInputField("New Name", currentName, 20, nil, nil)

	form.AddButton("Rename", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) != "" && newName != currentName {
			onRename(newName)
		}
	})

	form.AddButton("Cancel", func() {
		// Just close the form - handled by parent
	})

	form.SetBorder(true).SetTitle(" Rename Environment ")
	return form
}

// createEnvironmentPanel creates the environment panel with dropdown and config button
func createEnvironmentPanel(
	colors *ColorManager,
) (*tview.Flex, *tview.DropDown, *tview.Button, *CustomButton) {
	// Create environment dropdown
	envDropdown := createDropDownWithOpenOnFocus(
		"",
		[]string{"Base Environment"},
		colors,
		false, // Don't open on focus
	)
	envDropdown.SetBorder(false)
	envDropdown.SetCurrentOption(0)

	// Prevent opening on focus
	envDropdown.SetFocusFunc(func() {
		// Close the dropdown if it's open
		// In tview, we can't directly close, but perhaps we can set it to not open
		// Actually, since SetOpenOnFocus doesn't exist, we can try to override
		// For now, let's leave it and see
	})

	// Create config button with themed background color
	var configButton *CustomButton = createThemedButton(config.C.UI.ConfigButtonIcon, colors)

	separator := tview.NewBox().
		SetBackgroundColor(colors.Background)

	indicator := createCustomButton(config.C.UI.DropdownIndicator, colors.Background, colors.Background, colors.BorderFocus, colors.BorderFocus)

	// Create container
	container := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(separator, 1, 0, false).
			AddItem(envDropdown, 14, 0, false).
			AddItem(indicator, 2, 0, true), 17, 0, false).
		AddItem(separator, 0, 1, false).
		AddItem(configButton, 3, 0, true)

	container.SetBorder(true)
	container.SetTitle(" Environment ")
	container.SetBackgroundColor(colors.Background)
	container.SetBorderColor(colors.Border)
	container.SetTitleColor(colors.Title)

	// Explicit return with correct types
	var nilButton *tview.Button = nil
	return container, envDropdown, nilButton, configButton
}

func createWorkspacePanel(
	colors *ColorManager,
) (*tview.Flex, *tview.DropDown, *tview.Button, *CustomButton) {
	// Create workspace dropdown
	workspaceNames := []string{"Default"}
	// Try to load actual workspace names
	if workspaces, err := workspace.ListWorkspaces(); err == nil {
		workspaceNames = workspaces
	}

	workspaceDropdown := createDropDownWithOpenOnFocus(
		"",
		workspaceNames,
		colors,
		false, // Don't open on focus
	)
	workspaceDropdown.SetBorder(false)

	// Set current option based on current workspace
	currentWorkspace := "Default"
	if manager, err := workspace.LoadWorkspaceManager(); err == nil && manager.CurrentWorkspace != "" {
		currentWorkspace = manager.CurrentWorkspace
	}

	// Find the index of current workspace
	currentIndex := 0
	for i, name := range workspaceNames {
		if name == currentWorkspace {
			currentIndex = i
			break
		}
	}
	workspaceDropdown.SetCurrentOption(currentIndex)

	// Create config button with themed background color
	configButton := createThemedButton(config.C.UI.ConfigButtonIcon, colors)

	separator := tview.NewBox().
		SetBackgroundColor(colors.Background)

	// Create container
	container := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(separator, 1, 0, false).
			AddItem(workspaceDropdown, 14, 0, false), 15, 0, false).
		AddItem(separator, 0, 1, false).
		AddItem(configButton, 3, 0, true)

	container.SetBorder(true)
	container.SetTitle(" Workspace ")
	container.SetBackgroundColor(colors.Background)
	container.SetBorderColor(colors.Border)
	container.SetTitleColor(colors.Title)

	return container, workspaceDropdown, nil, configButton // configButton is *CustomButton, nil is for unused *tview.Button
}

// syncBodyContent syncs body content between view and edit panels
func syncBodyContent(content string, bodyEditMode bool, bodyEditPanel *tview.TextArea, bodyViewPanel *tview.TextView) {
	if bodyEditMode {
		bodyEditPanel.SetText(content, false)
	} else {
		bodyViewPanel.Clear()
		if content != "" {
			// Format content with syntax highlighting and variable highlighting
			formattedContent := FormatBodyContentWithVariables(content)

			// Set content with proper handling
			bodyViewPanel.SetText(formattedContent)
			bodyViewPanel.SetTextAlign(tview.AlignLeft)

			// Force proper rendering and scrolling
			go func() {
				// Small delay to ensure content is set before scrolling
				bodyViewPanel.ScrollToEnd()
				bodyViewPanel.ScrollToBeginning()
			}()
		} else {
			bodyViewPanel.SetText("")
			bodyViewPanel.SetTextAlign(tview.AlignLeft)
		}
	}
}

// switchBodyMode switches between view and edit modes for body content
func switchBodyMode(
	bodyEditMode *bool,
	bodyContainer *tview.Flex,
	bodyViewPanel *tview.TextView,
	bodyEditPanel *tview.TextArea,
	currentBodyContent string,
	currentRequest **workspace.Request,
	currentSelectedNode **tview.TreeNode,
	saveCurrentRequest func(),
) {
	*bodyEditMode = !*bodyEditMode
	bodyContainer.Clear()

	if *bodyEditMode {
		// Switch to edit mode
		bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
		bodyEditPanel.SetText(currentBodyContent, false)
	} else {
		// Switch to view mode
		bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
		// Update body content from edit panel if we were editing
		if *currentRequest != nil {
			currentBodyContent = bodyEditPanel.GetText()
			(*currentRequest).Body = currentBodyContent
			if *currentSelectedNode != nil {
				saveCurrentRequest()
			}
		}
		syncBodyContent(currentBodyContent, *bodyEditMode, bodyEditPanel, bodyViewPanel)
	}
}
