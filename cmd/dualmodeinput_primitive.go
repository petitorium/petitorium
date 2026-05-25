package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
)

type HeaderKeyInput struct {
	*tview.Pages
	viewMode      *tview.TextView
	editMode      *tview.InputField
	currentMode   string
	rawText       string
	onChanged     func(string)
	onModeChange  func()
	colors        *ColorManager
	variableRegex *regexp.Regexp
	lastWidth     int
}

func AppInputDualMode(colors *ColorManager, primitiveComponent ...tview.Primitive) *HeaderKeyInput {
	variableRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)

	var vm *tview.TextView
	var em *tview.InputField

	if len(primitiveComponent) > 0 && primitiveComponent[0] != nil {
		vm = primitiveComponent[0].(*tview.TextView)
	} else {
		vm = tview.NewTextView().
			SetDynamicColors(true).
			SetWordWrap(false).
			SetScrollable(false)

		vm.SetBackgroundColor(colors.InputBackground)
		vm.SetTextColor(colors.Foreground)
		vm.SetBorderPadding(0, 0, 0, 0)
		vm.SetTextAlign(tview.AlignCenter)

		vm.SetFocusFunc(func() {
			vm.SetBackgroundColor(colors.SelectedBackground)
			vm.SetTextColor(colors.SelectedForeground)
		})

		vm.SetBlurFunc(func() {
			vm.SetBackgroundColor(colors.InputBackground)
			vm.SetTextColor(colors.Foreground)
		})
	}

	if len(primitiveComponent) > 1 && primitiveComponent[1] != nil {
		em = primitiveComponent[1].(*tview.InputField)
	} else {
		em = tview.NewInputField()
		em.SetBackgroundColor(colors.Background)
		em.SetFieldBackgroundColor(colors.InputBackground)
		em.SetFieldTextColor(colors.Foreground)
		em.SetBorder(false)

		em.SetFocusFunc(func() {
			em.SetFieldBackgroundColor(colors.Error)
		})

		em.SetBlurFunc(func() {
			em.SetFieldBackgroundColor(colors.InputBackground)
		})
	}

	pages := tview.NewPages()
	pages.SetBackgroundColor(colors.Background)
	pages.AddPage("view", vm, true, true)
	pages.AddPage("edit", em, true, false)

	input := &HeaderKeyInput{
		Pages:         pages,
		viewMode:      vm,
		editMode:      em,
		currentMode:   "view",
		rawText:       "",
		colors:        colors,
		variableRegex: variableRegex,
	}

	em.SetChangedFunc(func(text string) {
		input.rawText = text
		if input.onChanged != nil {
			input.onChanged(text)
		}
	})

	em.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEsc {
			input.switchToViewMode()
		}
	})

	vm.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'i' {
			input.switchToEditMode()
			return nil
		}
		return event
	})

	return input
}

func (h *HeaderKeyInput) switchToViewMode() {
	h.currentMode = "view"
	h.viewMode.SetBackgroundColor(h.colors.InputBackground)
	h.Pages.SwitchToPage("view")
	h.updateViewMode()
	if h.onModeChange != nil {
		h.onModeChange()
	}
}

func (h *HeaderKeyInput) switchToEditMode() {
	h.currentMode = "edit"
	h.editMode.SetBackgroundColor(h.colors.Background)
	h.editMode.SetFieldBackgroundColor(h.colors.InputBackground)
	h.editMode.SetText(h.rawText)
	h.Pages.SwitchToPage("edit")
	if h.onModeChange != nil {
		h.onModeChange()
	}
}

func (h *HeaderKeyInput) Focus(delegate func(p tview.Primitive)) {
	h.Pages.Focus(delegate)
}

func (h *HeaderKeyInput) Draw(screen tcell.Screen) {
	if h.currentMode == "view" {
		_, _, width, _ := h.viewMode.GetInnerRect()
		if width > 0 && width != h.lastWidth {
			h.lastWidth = width
			h.updateViewModeWithWidth(width)
		}
	}
	h.Pages.Draw(screen)
}

func (h *HeaderKeyInput) HasFocus() bool {
	if h.currentMode == "edit" {
		return h.editMode.HasFocus()
	}
	return h.viewMode.HasFocus()
}

func (h *HeaderKeyInput) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event.Rune() == 'i' && h.currentMode == "view" {
			h.switchToEditMode()
			setFocus(h.editMode)
			return
		}

		if event.Key() == tcell.KeyTab || event.Key() == tcell.KeyBacktab {
			return
		}
		if h.Pages.InputHandler() != nil {
			h.Pages.InputHandler()(event, setFocus)
		}
	}
}

func (h *HeaderKeyInput) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return h.Pages.MouseHandler()
}

func (h *HeaderKeyInput) updateViewMode() {
	h.lastWidth = 0
	h.updateViewModeWithWidth(0)
}

func (h *HeaderKeyInput) updateViewModeWithWidth(width int) {
	if h.rawText == "" {
		h.viewMode.SetText("")
		return
	}

	matches := h.variableRegex.FindAllStringIndex(h.rawText, -1)
	if len(matches) == 0 {
		text := h.rawText
		if width > 0 {
			text = TruncateTaggedString(text, width)
		}
		h.viewMode.SetText(text)
		return
	}

	var result strings.Builder
	lastEnd := 0

	for i, match := range matches {
		start, end := match[0], match[1]

		result.WriteString(h.rawText[lastEnd:start])

		varName := h.rawText[start+2 : end-2]

		result.WriteString(fmt.Sprintf("[%s:%s:-]%s[-:-:-]",
			config.C.Theme.DropdownFocusedBackground,
			config.C.Theme.BorderFocusColor,
			varName))

		if i < len(matches)-1 && end == matches[i+1][0] {
			result.WriteString(" ")
		}

		lastEnd = end
	}

	result.WriteString(h.rawText[lastEnd:])

	renderedText := result.String()
	if width > 0 {
		renderedText = TruncateTaggedString(renderedText, width)
	}

	h.viewMode.SetText(renderedText)
}

func (h *HeaderKeyInput) SetText(text string) {
	h.rawText = text
	if h.currentMode == "view" {
		h.updateViewMode()
	} else {
		h.editMode.SetText(text)
	}
	if h.onChanged != nil {
		h.onChanged(text)
	}
}

func (h *HeaderKeyInput) GetText() string {
	return h.rawText
}

func (h *HeaderKeyInput) SetChangedFunc(callback func(string)) {
	h.onChanged = callback
}

func (h *HeaderKeyInput) SetInputCapture(capture func(*tcell.EventKey) *tcell.EventKey) {
	h.viewMode.SetInputCapture(capture)
	h.editMode.SetInputCapture(capture)
}

func (h *HeaderKeyInput) IsEditMode() bool {
	return h.currentMode == "edit"
}
