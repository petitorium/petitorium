package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
)

var paramRegex = regexp.MustCompile(`(\w+)="([^"]*)"`)

// textInputTarget represents a detected text input field that can receive a tag insertion.
type textInputTarget struct {
	getText func() string
	setText func(string)
}

// findActiveTextInput detects which text input field currently has focus in the UI.
// It uses hasDescendant to reliably match the focused primitive against known
// input containers (dual-mode inputs, body editor, etc.).
func findActiveTextInput(ui *UIOrchestrator) *textInputTarget {
	focus := ui.App.GetFocus()
	if focus == nil {
		return nil
	}

	// URL input (URLVariableInput embeds Pages → TextView/InputField)
	if ui.URLInput != nil && hasDescendant(ui.URLInput, focus) {
		return &textInputTarget{
			getText: ui.URLInput.GetText,
			setText: ui.URLInput.SetText,
		}
	}

	// Body edit panel
	if ui.BodyEditPanel != nil && ui.BodyEditPanel == focus {
		return &textInputTarget{
			getText: ui.BodyEditPanel.GetText,
			setText: func(s string) { ui.BodyEditPanel.SetText(s, false) },
		}
	}

	// Environment modal JSON editor
	if ui.EnvModalEditor != nil && ui.EnvModalEditor == focus {
		return &textInputTarget{
			getText: ui.EnvModalEditor.GetText,
			setText: func(s string) { ui.EnvModalEditor.SetText(s, false) },
		}
	}

	// Header rows
	for _, row := range currentHeaderRows {
		if row.KeyInput != nil && hasDescendant(row.KeyInput, focus) {
			return &textInputTarget{
				getText: row.KeyInput.GetText,
				setText: row.KeyInput.SetText,
			}
		}
		if row.ValueInput != nil && hasDescendant(row.ValueInput, focus) {
			return &textInputTarget{
				getText: row.ValueInput.GetText,
				setText: row.ValueInput.SetText,
			}
		}
	}

	// Query param rows
	for _, row := range currentQueryRows {
		if row.KeyInput != nil && hasDescendant(row.KeyInput, focus) {
			return &textInputTarget{
				getText: row.KeyInput.GetText,
				setText: row.KeyInput.SetText,
			}
		}
		if row.ValueInput != nil && hasDescendant(row.ValueInput, focus) {
			return &textInputTarget{
				getText: row.ValueInput.GetText,
				setText: row.ValueInput.SetText,
			}
		}
	}

	return nil
}

// parseExistingCommandRunnerTag scans text for the first {{command-runner:run ...}} tag and
// returns its parameters. If no tag is found, all return values are empty.
func parseExistingCommandRunnerTag(text string) (command, outputType, jsonPath string) {
	idx := strings.Index(text, "{{command-runner:run ")
	if idx == -1 {
		return "", "", ""
	}
	end := strings.Index(text[idx:], "}}")
	if end == -1 {
		return "", "", ""
	}

	tag := text[idx : idx+end+2]
	matches := paramRegex.FindAllStringSubmatch(tag, -1)
	params := make(map[string]string)
	for _, m := range matches {
		if len(m) == 3 {
			params[m[1]] = m[2]
		}
	}
	return params["command"], params["type"], params["jsonPath"]
}

// showCommandRunnerModal displays a modal for configuring and inserting a command-runner tag.
// It detects the currently focused input field and inserts (or replaces) the constructed tag.
func showCommandRunnerModal(ui *UIOrchestrator) {
	colors := ui.Colors
	app := ui.App
	pages := ui.Pages

	// Detect the active text input before opening the modal
	target := findActiveTextInput(ui)
	if target == nil {
		return
	}

	// Check if the target already contains a command-runner tag
	existingCommand, existingType, existingJSONPath := parseExistingCommandRunnerTag(target.getText())
	isEditing := existingCommand != ""

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Command input
	commandInput := tview.NewInputField().
		SetLabel("Command: ").
		SetFieldWidth(50)
	commandInput.SetFieldBackgroundColor(colors.Border)
	commandInput.SetBackgroundColor(colors.Background)
	commandInput.SetFieldTextColor(colors.Foreground)
	commandInput.SetLabelColor(colors.Foreground)
	if isEditing {
		commandInput.SetText(existingCommand)
	}
	form.AddFormItem(commandInput)

	// Type dropdown
	typeDropdown := tview.NewDropDown().
		SetLabel("Type: ").
		SetOptions([]string{"string", "json"}, nil)
	typeDropdown.SetFieldBackgroundColor(colors.Background)
	typeDropdown.SetBackgroundColor(colors.Background)
	typeDropdown.SetFieldTextColor(colors.Foreground)
	typeDropdown.SetLabelColor(colors.Foreground)
	if isEditing && existingType == "json" {
		typeDropdown.SetCurrentOption(1)
	} else {
		typeDropdown.SetCurrentOption(0)
	}
	form.AddFormItem(typeDropdown)

	// JSONPath input (disabled initially unless editing a json tag)
	jsonPathInput := tview.NewInputField().
		SetLabel("JSONPath: ").
		SetFieldWidth(40)
	jsonPathInput.SetFieldBackgroundColor(colors.Border)
	jsonPathInput.SetBackgroundColor(colors.Background)
	jsonPathInput.SetFieldTextColor(colors.Foreground)
	jsonPathInput.SetLabelColor(colors.Foreground)
	if isEditing && existingType == "json" {
		jsonPathInput.SetDisabled(false)
		jsonPathInput.SetText(existingJSONPath)
	} else {
		jsonPathInput.SetDisabled(true)
	}
	form.AddFormItem(jsonPathInput)

	// Preview output
	previewText := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	previewText.SetBackgroundColor(colors.Background)
	previewText.SetTextColor(colors.Foreground)
	previewText.SetBorder(true).SetTitle(" Preview ")
	previewText.SetTitleColor(colors.Title)
	previewText.SetBorderColor(colors.Border)
	previewText.SetText("(Click Live Preview to run command)")

	// Toggle jsonPath visibility based on type selection
	typeDropdown.SetSelectedFunc(func(text string, index int) {
		if text == "json" {
			jsonPathInput.SetDisabled(false)
		} else {
			jsonPathInput.SetDisabled(true)
			jsonPathInput.SetText("")
		}
	})

	// Build the layout: form on top, preview below
	layout := tview.NewFlex().SetDirection(tview.FlexRow)
	layout.AddItem(form, 0, 1, true)
	layout.AddItem(previewText, 8, 0, false)

	closeModalFunc := func() {
		pages.RemovePage("commandRunnerModal")
		pages.SwitchToPage("main")
	}

	// Helper to run command for live preview
	runPreview := func() {
		cmdText := commandInput.GetText()
		if strings.TrimSpace(cmdText) == "" {
			app.QueueUpdateDraw(func() {
				previewText.SetText("[red]Error: Command is empty[-]")
			})
			return
		}

		outputType, _ := typeDropdown.GetCurrentOption()
		outputTypeStr := "string"
		if outputType == 1 {
			outputTypeStr = "json"
		}
		jsonPath := jsonPathInput.GetText()

		go func() {
			output, err := RunShellCommand(cmdText)
			if err != nil {
				app.QueueUpdateDraw(func() {
					previewText.SetText(fmt.Sprintf("[red]Error: %v[-]", err))
				})
				return
			}

			if outputTypeStr == "json" && jsonPath != "" {
				result := gjson.Get(output, jsonPath)
				output = result.String()
			}

			app.QueueUpdateDraw(func() {
				if output == "" {
					previewText.SetText("(empty output)")
				} else {
					previewText.SetText(output)
				}
			})
		}()
	}

	// Helper to build the tag string
	buildTag := func() string {
		cmdText := commandInput.GetText()
		outputType, _ := typeDropdown.GetCurrentOption()
		outputTypeStr := "string"
		if outputType == 1 {
			outputTypeStr = "json"
		}
		jsonPath := jsonPathInput.GetText()
		return fmt.Sprintf(`{{command-runner:run command="%s" type="%s" jsonPath="%s"}}`,
			cmdText, outputTypeStr, jsonPath)
	}

	// Helper to insert or replace the tag into the target input
	insertTag := func() {
		cmdText := commandInput.GetText()
		if strings.TrimSpace(cmdText) == "" {
			return
		}

		tag := buildTag()
		current := target.getText()

		if isEditing {
			// Replace the first existing tag
			idx := strings.Index(current, "{{command-runner:run ")
			end := strings.Index(current[idx:], "}}")
			if idx != -1 && end != -1 {
				end += idx + 2
				newText := current[:idx] + tag + current[end:]
				target.setText(newText)
			} else {
				target.setText(current + tag)
			}
		} else {
			target.setText(current + tag)
		}

		closeModalFunc()
	}

	insertLabel := "Insert"
	if isEditing {
		insertLabel = "Update"
	}

	// Add buttons
	form.AddButton("Live Preview", runPreview).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))
	form.AddButton(insertLabel, insertTag).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))
	form.AddButton("Cancel", closeModalFunc)

	form.SetCancelFunc(closeModalFunc)

	// Handle Esc key
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeModalFunc()
			return nil
		}
		return event
	})

	layout.SetBorder(true).SetTitle(" Command Runner ")
	layout.SetBorderColor(colors.BorderFocus)
	layout.SetTitleColor(colors.Title)

	modal := createModal(layout, 80, 25, colors.Background)
	pages.AddPage("commandRunnerModal", modal, true, true)
	app.SetFocus(form)
}
