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

// createEnvironmentListPanel creates a list panel for selecting environments using tview.NewList
func createEnvironmentListPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	buttonSelectedColor tcell.Color,
	environments []workspace.Environment,
	onEnvironmentSelected func(*workspace.Environment),
	onCreateNew func(),
	onDelete func(*workspace.Environment),
	onRename func(*workspace.Environment),
) *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(backgroundColor)
	list.SetBorderColor(borderColor)
	list.SetTitleColor(titleColor)
	list.SetMainTextColor(foregroundColor)
	list.SetSelectedBackgroundColor(buttonSelectedColor)
	list.SetSelectedTextColor(foregroundColor)
	list.SetBorder(true).SetTitle(" Environments ")

	// Add "Create New Environment" option at the top
	list.AddItem("➕ Create New Environment", "", 0, onCreateNew)

	// Add all existing environments
	for i, env := range environments {
		env := env // Capture loop variable

		var itemText string
		if env.Name == "Base" {
			itemText = env.Name + " (base environment)"
		} else {
			itemText = env.Name
		}

		list.AddItem(itemText, "", 0, func() {
			onEnvironmentSelected(&environments[i])
		})
	}

	// Add selection change handler to show variables on hover
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index > 0 && index <= len(environments) {
			// Show variables for the selected environment (index 0 is "Create New Environment")
			env := &environments[index-1]
			onEnvironmentSelected(env)
		} else if index == 0 {
			// Clear the JSON editor when "Create New Environment" is selected
			// This is just a button, not an actual environment
			onEnvironmentSelected(nil)
		}
	})

	// Add vim-style navigation (j/k for down/up) and delete (d)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'd':
			currentItem := list.GetCurrentItem()
			if currentItem > 0 && currentItem <= len(environments) {
				env := &environments[currentItem-1]
				if env.Name != "Base" {
					onDelete(env)
				}
			}
			return nil // Consume the event
		case 'n':
			onCreateNew()
			return nil // Consume the event
		case 'r', 'R':
			currentItem := list.GetCurrentItem()
			if currentItem > 0 && currentItem <= len(environments) {
				env := &environments[currentItem-1]
				if env.Name != "Base" {
					onRename(env)
				}
			}
			return nil // Consume the event
		}
		return event
	})

	return list
}

// createWorkspaceListPanel creates a list panel for selecting workspaces using tview.NewList
func createWorkspaceListPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	buttonSelectedColor tcell.Color,
	workspaces []workspace.WorkspaceMetadata,
	currentWorkspace string,
	onWorkspaceSelected func(*workspace.WorkspaceMetadata),
	onCreateNew func(),
	onDelete func(*workspace.WorkspaceMetadata),
	onRename func(*workspace.WorkspaceMetadata),
	onDuplicate func(*workspace.WorkspaceMetadata),
) *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(backgroundColor)
	list.SetBorderColor(borderColor)
	list.SetTitleColor(titleColor)
	list.SetMainTextColor(foregroundColor)
	list.SetSelectedBackgroundColor(buttonSelectedColor)
	list.SetSelectedTextColor(foregroundColor)
	list.SetBorder(true).SetTitle(" Workspaces ")

	// Add "Create New Workspace" option at the top
	list.AddItem("➕ Create New Workspace", "", 0, onCreateNew)

	// Add all existing workspaces
	for i, ws := range workspaces {
		ws := ws // Capture loop variable

		var itemText string
		if ws.Name == currentWorkspace {
			itemText = ws.Name + " (current)"
		} else {
			itemText = ws.Name
		}

		list.AddItem(itemText, "", 0, func() {
			onWorkspaceSelected(&workspaces[i])
		})
	}

	// Add selection change handler to show workspace info on hover
	list.SetChangedFunc(func(index int, mainText string, secondaryText string, shortcut rune) {
		if index > 0 && index <= len(workspaces) {
			// Show info for the selected workspace (index 0 is "Create New Workspace")
			ws := &workspaces[index-1]
			onWorkspaceSelected(ws)
		} else if index == 0 {
			// Clear when "Create New Workspace" is selected
			onWorkspaceSelected(nil)
		}
	})

	// Add vim-style navigation (j/k for down/up) and actions (d delete, n new, r rename, c duplicate)
	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		case 'd':
			currentItem := list.GetCurrentItem()
			if currentItem > 0 && currentItem <= len(workspaces) {
				ws := &workspaces[currentItem-1]
				if ws.Name != currentWorkspace {
					onDelete(ws)
				}
			}
			return nil // Consume the event
		case 'n':
			onCreateNew()
			return nil // Consume the event
		case 'r', 'R':
			currentItem := list.GetCurrentItem()
			if currentItem > 0 && currentItem <= len(workspaces) {
				ws := &workspaces[currentItem-1]
				onRename(ws)
			}
			return nil // Consume the event
		case 'c', 'C':
			currentItem := list.GetCurrentItem()
			if currentItem > 0 && currentItem <= len(workspaces) {
				ws := &workspaces[currentItem-1]
				onDuplicate(ws)
			}
			return nil // Consume the event
		}
		return event
	})

	return list
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
			// Format content with syntax highlighting
			formattedContent := formatBodyContent(content)

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
				(*currentSelectedNode).SetReference(**currentRequest)
				saveCurrentRequest()
			}
		}
		syncBodyContent(currentBodyContent, *bodyEditMode, bodyEditPanel, bodyViewPanel)
	}
}
