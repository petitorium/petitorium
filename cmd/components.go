package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// createInlineEditInput creates an input field for inline editing of environment names
func createInlineEditInput(
	backgroundColor,
	borderColor,
	titleColor,
	foregroundColor tcell.Color,
	currentName string,
	onSave func(string),
	onCancel func(),
) *tview.InputField {
	editInput := createInputField("", backgroundColor, borderColor, titleColor, foregroundColor)
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

// createEnvironmentListPanel creates a list panel for selecting and managing environments
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
	onRenameEnvironment func(*workspace.Environment),
	onRemoveEnvironment func(*workspace.Environment),
) *tview.Flex {
	list := tview.NewFlex().SetDirection(tview.FlexRow)
	list.SetBackgroundColor(backgroundColor)
	list.SetBorderColor(borderColor)
	list.SetTitleColor(titleColor)

	// Add "Create New Environment" option at the top
	createNewButton := createButton("➕ Create New Environment", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
	createNewButton.SetSelectedFunc(onCreateNew)
	list.AddItem(createNewButton, 1, 0, false)

	// Add separator
	separator := tview.NewBox().SetBackgroundColor(backgroundColor)
	list.AddItem(separator, 1, 0, false)

	// Add all existing environments
	for i, env := range environments {
		env := env // Capture loop variable

		// Create a container for each environment with buttons
		envContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

		// Create the main environment button
		var buttonText string
		if env.Name == "Base" {
			buttonText = env.Name + " (base environment)"
		} else {
			buttonText = env.Name
		}

		envButton := createButton(buttonText, backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
		envButton.SetSelectedFunc(func() {
			onEnvironmentSelected(&environments[i])
		})
		envContainer.AddItem(envButton, 0, 1, false)

		// Add rename and remove buttons for non-base environments
		if env.Name != "Base" {
			renameButton := createButton("✏", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
			renameButton.SetSelectedFunc(func() {
				// Start inline editing - replace button with input field
				envContainer.Clear()

				editInput := createInlineEditInput(
					backgroundColor,
					borderColor,
					titleColor,
					foregroundColor,
					env.Name,
					func(newName string) {
						// Update the environment name
						environments[i].Name = newName
						onRenameEnvironment(&environments[i])
						// Revert to button view will be handled by onCancel
					},
					func() {
						// Revert to button view
						envContainer.Clear()

						// Recreate the environment button with current name
						var buttonText string
						if environments[i].Name == "Base" {
							buttonText = environments[i].Name + " (base environment)"
						} else {
							buttonText = environments[i].Name
						}

						envButton := createButton(buttonText, backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
						envButton.SetSelectedFunc(func() {
							onEnvironmentSelected(&environments[i])
						})
						envContainer.AddItem(envButton, 0, 1, false)

						// Re-add rename and remove buttons
						renameBtn := createButton("✏", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
						renameBtn.SetSelectedFunc(func() {
							// Recursively start inline editing again
							envContainer.Clear()

							editInput2 := createInlineEditInput(
								backgroundColor,
								borderColor,
								titleColor,
								foregroundColor,
								environments[i].Name,
								func(newName string) {
									environments[i].Name = newName
									onRenameEnvironment(&environments[i])
								},
								func() {
									// This will be called to revert
									envContainer.Clear()

									var btnText string
									if environments[i].Name == "Base" {
										btnText = environments[i].Name + " (base environment)"
									} else {
										btnText = environments[i].Name
									}

									finalEnvButton := createButton(btnText, backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
									finalEnvButton.SetSelectedFunc(func() {
										onEnvironmentSelected(&environments[i])
									})
									envContainer.AddItem(finalEnvButton, 0, 1, false)

									finalRenameButton := createButton("✏", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
									finalRenameButton.SetSelectedFunc(func() {
										onRenameEnvironment(&environments[i])
									})
									envContainer.AddItem(finalRenameButton, 3, 0, false)

									finalRemoveButton := createButton("✕", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
									finalRemoveButton.SetSelectedFunc(func() {
										onRemoveEnvironment(&environments[i])
									})
									envContainer.AddItem(finalRemoveButton, 3, 0, false)
								},
							)
							envContainer.AddItem(editInput2, 0, 1, false)
						})
						envContainer.AddItem(renameBtn, 3, 0, false)

						removeBtn := createButton("✕", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
						removeBtn.SetSelectedFunc(func() {
							onRemoveEnvironment(&environments[i])
						})
						envContainer.AddItem(removeBtn, 3, 0, false)
					},
				)
				envContainer.AddItem(editInput, 0, 1, false)
			})
			envContainer.AddItem(renameButton, 3, 0, false)

			removeButton := createButton("✕", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
			removeButton.SetSelectedFunc(func() {
				onRemoveEnvironment(&environments[i])
			})
			envContainer.AddItem(removeButton, 3, 0, false)
		}

		list.AddItem(envContainer, 1, 0, false)
	}

	list.SetBorder(true).SetTitle(" Environments ")
	return list
}

// createEnvironmentPanel creates the environment panel with dropdown and config button
func createEnvironmentPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	selectionBackgroundColor,
	activeTabColor,
	buttonSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) (*tview.Flex, *tview.DropDown, *tview.Button, *CustomButton) {
	// Create environment dropdown
	envDropdown := createDropDown(
		"",
		[]string{"Base Environment"},
		backgroundColor,
		backgroundColor,
		backgroundColor,
		foregroundColor,
		activeTabColor,
		backgroundColor,
		selectionBackgroundColor,
		dropdownFocusedBackgroundColor,
	)
	envDropdown.SetBorder(false)
	envDropdown.SetCurrentOption(0)

	// Create config button
	configButton := createButton("⚙", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)

	separator := tview.NewBox().
		SetBackgroundColor(backgroundColor)

	indicator := createCustomButton("▼", backgroundColor, backgroundColor, borderFocusColor, borderFocusColor)

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
	container.SetBackgroundColor(backgroundColor)
	container.SetBorderColor(borderColor)
	container.SetTitleColor(titleColor)

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
