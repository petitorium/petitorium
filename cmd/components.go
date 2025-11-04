package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
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
) *tview.List {
	list := tview.NewList()
	list.SetBackgroundColor(backgroundColor)
	list.SetBorderColor(borderColor)
	list.SetTitleColor(titleColor)
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
	envDropdown := createDropDown(
		"",
		[]string{"Base Environment"},
		colors,
	)
	envDropdown.SetBorder(false)
	envDropdown.SetCurrentOption(0)

	// Create config button
	configButton := createButton("⚙", colors)

	separator := tview.NewBox().
		SetBackgroundColor(colors.Background)

	indicator := createCustomButton("▼", colors.Background, colors.Background, colors.BorderFocus, colors.BorderFocus)

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

	return container, envDropdown, configButton, indicator
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
