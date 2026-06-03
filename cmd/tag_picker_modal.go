package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/tidwall/gjson"
)

// showTagPickerModal displays a small list modal when multiple tags are detected
// in the focused input.  The user selects a tag and the onSelect callback is
// invoked with the chosen DetectedTag.
func showTagPickerModal(ui *UIOrchestrator, tags []DetectedTag, onSelect func(tag DetectedTag)) {
	colors := ui.Colors
	app := ui.App
	pages := ui.Pages

	list := tview.NewList()
	list.SetBackgroundColor(colors.Background)
	list.SetMainTextColor(colors.Foreground)
	list.SetSecondaryTextColor(colors.Placeholder)
	list.SetSelectedBackgroundColor(colors.Selection)
	list.SetSelectedTextColor(colors.Foreground)
	list.SetHighlightFullLine(true)

	for i, tag := range tags {
		// Build a short preview from the raw tag inner params
		preview := tag.Inner
		if len(preview) > 50 {
			preview = preview[:47] + "..."
		}
		label := fmt.Sprintf("[%d] %s:%s", i+1, tag.Plugin, tag.Action)
		list.AddItem(label, preview, rune('0'+i+1), nil)
	}

	previousFocus := app.GetFocus()
	closeModalFunc := func() {
		pages.RemovePage("tagPickerModal")
		pages.SwitchToPage("main")
		if previousFocus != nil {
			app.SetFocus(previousFocus)
		}
	}

	list.SetSelectedFunc(func(idx int, mainText string, secondaryText string, shortcut rune) {
		if idx >= 0 && idx < len(tags) {
			closeModalFunc()
			onSelect(tags[idx])
		}
	})

	list.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			closeModalFunc()
			return nil
		}
		// Number keys 1-9 for quick selection
		r := event.Rune()
		if r >= '1' && r <= '9' {
			idx := int(r - '1')
			if idx < len(tags) {
				closeModalFunc()
				onSelect(tags[idx])
				return nil
			}
		}
		return event
	})

	list.SetBorder(true).SetTitle(" Select Tag ")
	list.SetBorderColor(colors.BorderFocus)
	list.SetTitleColor(colors.Title)

	height := len(tags) + 4
	if height > 20 {
		height = 20
	}
	modal := createModal(list, 70, height, colors.Background)
	pages.AddPage("tagPickerModal", modal, true, true)
	app.SetFocus(list)
}

// openTagEditorForField is the entry point for the tag editor keybinding.
// It scans the focused field and either:
//   - opens the editor directly (0 or 1 tag)
//   - shows the picker first (2+ tags)
func openTagEditorForField(ui *UIOrchestrator) {
	target := findActiveTextInput(ui)
	if target == nil {
		return
	}

	text := target.getText()
	tags := scanTags(text)

	switch len(tags) {
	case 0:
		// No existing tag — open editor for a new tag insertion.
		showTagEditorModal(ui, target, nil)
	case 1:
		// Exactly one tag — edit it directly.
		showTagEditorModal(ui, target, &tags[0])
	default:
		// Multiple tags — show picker first.
		showTagPickerModal(ui, tags, func(tag DetectedTag) {
			showTagEditorModal(ui, target, &tag)
		})
	}
}

// showCommandRunnerModalForTag opens the command-runner modal pre-filled with
// the specific tag identified by dt.  On save it replaces that exact tag in the
// parent text using the precise Start/End offsets.
//
// TODO: This is a temporary bridge until Phase 5 replaces it with the dynamic
// plugin-driven form renderer.
func showCommandRunnerModalForTag(ui *UIOrchestrator, target *textInputTarget, dt DetectedTag) {
	colors := ui.Colors
	app := ui.App
	pages := ui.Pages
	forEnvJSON := ui.EnvModalEditor != nil

	// Pre-fill from the detected tag's inner params.
	matches := paramRegex.FindAllStringSubmatch(dt.Inner, -1)
	params := make(map[string]string)
	for _, m := range matches {
		if len(m) == 3 {
			params[m[1]] = m[2]
		}
	}
	existingCommand := params["command"]
	existingType := params["type"]
	existingJSONPath := params["jsonPath"]
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
		tag := fmt.Sprintf(`{{command-runner:run command="%s" type="%s" jsonPath="%s"}}`,
			cmdText, outputTypeStr, jsonPath)
		if forEnvJSON {
			tag = strings.ReplaceAll(tag, `"`, `\"`)
		}
		return tag
	}

	// Helper to insert or replace the exact tag in the target input
	insertTag := func() {
		cmdText := commandInput.GetText()
		if strings.TrimSpace(cmdText) == "" {
			return
		}

		tag := buildTag()
		current := target.getText()

		if forEnvJSON {
			if isEditing {
				ui.EnvModalEditor.Replace(dt.Start, dt.End, tag)
			} else {
				row, col, _, _ := ui.EnvModalEditor.GetCursor()
				offset := cursorByteOffset(current, row, col)
				ui.EnvModalEditor.Replace(offset, offset, tag)
			}
			closeModalFunc()
			return
		}

		if isEditing {
			// Replace the exact tag using precise byte offsets.
			newText := current[:dt.Start] + tag + current[dt.End:]
			target.setText(newText)
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
