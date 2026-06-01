package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium-plugin-sdk/types"
)

// showTagEditorModal displays a plugin-driven form for editing (or inserting) a
// template variable tag.  When dt is nil the modal is opened in "insert" mode
// for the given plugin (defaulting to "command-runner").
func showTagEditorModal(ui *UIOrchestrator, target *textInputTarget, dt *DetectedTag) {
	colors := ui.Colors
	app := ui.App
	pages := ui.Pages

	pluginName := "command-runner"
	var rawTag string
	if dt != nil {
		pluginName = dt.Plugin
		rawTag = dt.Raw
	}

	// Look up the plugin.
	var schema *types.TagEditorSchema
	var displayLabel string
	if ui.PluginManager != nil {
		if plg, ok := ui.PluginManager.GetPlugin(pluginName); ok {
			if capable, ok := plg.(types.TagEditorCapable); ok {
				ctx := detectFieldContext(ui, target)
				res, err := capable.GetTagDetails(rawTag, ctx)
				if err == nil && res != nil {
					displayLabel = res.DisplayLabel
					schema = res.Schema
				}
			}
		}
	}

	// Fallback: if no schema was returned, fall back to the hardcoded
	// command-runner modal so the user is never stuck.
	if schema == nil && pluginName == "command-runner" {
		if dt != nil {
			showCommandRunnerModalForTag(ui, target, *dt)
		} else {
			showCommandRunnerModal(ui)
		}
		return
	}

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Keep references to every form item so we can read values and handle
	// dependencies.
	formItems := make(map[string]tview.FormItem)
	fieldMap := make(map[string]types.TagField)

	// Track disabled states to avoid redundant SetDisabled calls, which in
	// tview trigger a finished callback and can steal focus.
	disabledStates := make(map[string]bool)

	// Helper to evaluate dependencies for all fields.
	refreshDependencies := func() {}

	for _, f := range schema.Fields {
		fieldMap[f.Key] = f
		label := f.Label
		if f.Required {
			label += " *"
		}

		switch f.FieldType {
		case "text":
			inp := tview.NewInputField().SetLabel(label).SetFieldWidth(50)
			inp.SetFieldBackgroundColor(colors.Border)
			inp.SetBackgroundColor(colors.Background)
			inp.SetFieldTextColor(colors.Foreground)
			inp.SetLabelColor(colors.Foreground)
			inp.SetText(f.DefaultValue)
			formItems[f.Key] = inp
			form.AddFormItem(inp)

		case "textarea":
			// tview.Form doesn't have a native textarea, so we use a
			// multi-line InputField by widening it.
			inp := tview.NewInputField().SetLabel(label).SetFieldWidth(60)
			inp.SetFieldBackgroundColor(colors.Border)
			inp.SetBackgroundColor(colors.Background)
			inp.SetFieldTextColor(colors.Foreground)
			inp.SetLabelColor(colors.Foreground)
			inp.SetText(f.DefaultValue)
			formItems[f.Key] = inp
			form.AddFormItem(inp)

		case "dropdown":
			dd := tview.NewDropDown().SetLabel(label)
			dd.SetOptions(f.Options, nil)
			dd.SetFieldBackgroundColor(colors.Background)
			dd.SetBackgroundColor(colors.Background)
			dd.SetFieldTextColor(colors.Foreground)
			dd.SetLabelColor(colors.Foreground)
			// Set current option from default value.
			for i, opt := range f.Options {
				if opt == f.DefaultValue {
					dd.SetCurrentOption(i)
					break
				}
			}
			formItems[f.Key] = dd
			form.AddFormItem(dd)

		case "checkbox":
			chk := tview.NewCheckbox().SetLabel(label)
			chk.SetBackgroundColor(colors.Background)
			chk.SetFieldBackgroundColor(colors.Background)
			chk.SetLabelColor(colors.Foreground)
			chk.SetChecked(f.DefaultValue == "true")
			formItems[f.Key] = chk
			form.AddFormItem(chk)
		}
	}

	// refreshDependencies evaluates each field's depends_on/depends_value and
	// enables or disables the widget accordingly.
	refreshDependencies = func() {
		for key, f := range fieldMap {
			item := formItems[key]
			if item == nil {
				continue
			}
			if f.DependsOn == "" {
				continue
			}
			depItem := formItems[f.DependsOn]
			if depItem == nil {
				continue
			}
			var currentValue string
			switch w := depItem.(type) {
			case *tview.DropDown:
				_, opt := w.GetCurrentOption()
				currentValue = opt
			case *tview.InputField:
				currentValue = w.GetText()
			case *tview.Checkbox:
				if w.IsChecked() {
					currentValue = "true"
				} else {
					currentValue = "false"
				}
			}

			enabled := currentValue == f.DependsValue
			shouldDisable := !enabled
			if disabledStates[key] == shouldDisable {
				continue
			}
			disabledStates[key] = shouldDisable

			switch w := item.(type) {
			case *tview.InputField:
				w.SetDisabled(shouldDisable)
			case *tview.DropDown:
				w.SetDisabled(shouldDisable)
			case *tview.Checkbox:
				w.SetDisabled(shouldDisable)
			}
		}
	}

	// Wire dependency re-evaluation into every interactive widget.
	for key, item := range formItems {
		k := key
		switch w := item.(type) {
		case *tview.DropDown:
			w.SetSelectedFunc(func(text string, index int) {
				_ = text
				_ = index
				refreshDependencies()
			})
		case *tview.Checkbox:
			w.SetChangedFunc(func(checked bool) {
				_ = checked
				refreshDependencies()
			})
		case *tview.InputField:
			w.SetChangedFunc(func(text string) {
				_ = text
				refreshDependencies()
			})
		}
		_ = k
	}

	// Run once to set initial disabled states.
	refreshDependencies()

	closeModalFunc := func() {
		pages.RemovePage("tagEditorModal")
		pages.SwitchToPage("main")
	}

	// Collect current form values into a map.
	collectValues := func() map[string]string {
		values := make(map[string]string)
		for key, item := range formItems {
			switch w := item.(type) {
			case *tview.InputField:
				values[key] = w.GetText()
			case *tview.DropDown:
				_, opt := w.GetCurrentOption()
				values[key] = opt
			case *tview.Checkbox:
				if w.IsChecked() {
					values[key] = "true"
				} else {
					values[key] = "false"
				}
			}
		}
		return values
	}

	// Save handler: call plugin.UpdateTag and replace the exact tag.
	saveTag := func() {
		values := collectValues()

		var newTag string
		if ui.PluginManager != nil {
			if plg, ok := ui.PluginManager.GetPlugin(pluginName); ok {
				if capable, ok := plg.(types.TagEditorCapable); ok {
					res, err := capable.UpdateTag(rawTag, values)
					if err == nil && res != nil {
						newTag = res.NewRawTag
					}
				}
			}
		}

		if newTag == "" {
			// If the plugin didn't return a tag, reconstruct a best-effort one.
			newTag = reconstructTag(pluginName, "run", values)
		}

		current := target.getText()
		if dt != nil {
			// Replace exact tag at precise offsets.
			newText := current[:dt.Start] + newTag + current[dt.End:]
			target.setText(newText)
		} else {
			target.setText(current + newTag)
		}
		closeModalFunc()
	}

	// Delete handler: remove the tag from the text.
	deleteTag := func() {
		if dt == nil {
			closeModalFunc()
			return
		}
		current := target.getText()
		newText := current[:dt.Start] + current[dt.End:]
		target.setText(newText)
		closeModalFunc()
	}

	insertLabel := "Insert"
	if dt != nil {
		insertLabel = "Update"
	}

	form.AddButton(insertLabel, saveTag).
		SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))
	if dt != nil {
		form.AddButton("Delete", deleteTag).
			SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Error).Foreground(colors.Foreground))
	}
	form.AddButton("Cancel", closeModalFunc)

	form.SetCancelFunc(closeModalFunc)
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeModalFunc()
			return nil
		}
		return event
	})

	title := " Tag Editor "
	if displayLabel != "" {
		title = fmt.Sprintf(" %s ", displayLabel)
	}
	form.SetBorder(true).SetTitle(title)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)

	modal := createModal(form, 80, 20, colors.Background)
	pages.AddPage("tagEditorModal", modal, true, true)
	app.SetFocus(form)
}

// detectFieldContext tries to determine which part of the request the focused
// input belongs to.  It returns a string like "url", "header", "query", "body".
func detectFieldContext(ui *UIOrchestrator, target *textInputTarget) string {
	if target == nil {
		return "body"
	}
	// Compare against known input components by checking the getter results.
	if ui.URLInput != nil && target.getText() == ui.URLInput.GetText() {
		return "url"
	}
	if ui.BodyEditPanel != nil && target.getText() == ui.BodyEditPanel.GetText() {
		return "body"
	}
	return "body"
}

// reconstructTag builds a raw tag string from plugin name, action, and values.
// This is a best-effort fallback when the plugin does not implement UpdateTag.
func reconstructTag(plugin, action string, values map[string]string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "{{%s:%s", plugin, action)
	for k, v := range values {
		fmt.Fprintf(&b, " %s=\"%s\"", k, v)
	}
	b.WriteString("}}")
	return b.String()
}
