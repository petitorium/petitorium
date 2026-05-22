package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CheckboxPrimitive struct {
	*tview.TextView
	enabled     bool
	onChar      string
	offChar     string
	changedFunc func(bool)
	colors      *ColorManager
}

func AppCheckbox(onChar, offChar string, enabled bool, colors *ColorManager) *CheckboxPrimitive {
	cb := &CheckboxPrimitive{
		TextView: tview.NewTextView(),
		enabled:  enabled,
		onChar:   onChar,
		offChar:  offChar,
		colors:   colors,
	}
	cb.SetBorder(false)
	cb.SetBackgroundColor(colors.Background)
	cb.SetText(onChar)
	cb.SetTextAlign(tview.AlignCenter)
	if enabled {
		cb.SetTextColor(colors.Foreground)
	} else {
		cb.SetTextColor(tcell.ColorGray)
	}
	cb.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Rune() == ' ' {
			cb.toggle()
			return nil
		}
		return event
	})
	cb.SetFocusFunc(func() {
		cb.SetBackgroundColor(colors.SelectedBackground)
	})
	cb.SetBlurFunc(func() {
		cb.SetBackgroundColor(colors.Background)
	})
	return cb
}

func (cb *CheckboxPrimitive) toggle() {
	cb.enabled = !cb.enabled
	if cb.enabled {
		cb.SetText(cb.onChar)
		cb.SetTextColor(cb.colors.Foreground)
		cb.SetBackgroundColor(cb.colors.Background)
	} else {
		cb.SetText(cb.offChar)
		cb.SetTextColor(tcell.ColorGray)
		cb.SetBackgroundColor(cb.colors.Background)
	}
	if cb.changedFunc != nil {
		cb.changedFunc(cb.enabled)
	}
}

func (cb *CheckboxPrimitive) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
			x, y := event.Position()
			bx, by, width, height := cb.Box.GetRect()

			// Check if click is within button bounds
			if x >= bx && x < bx+width && y >= by && y < by+height {
				cb.toggle()
				return true, cb
			}
		}

		return false, nil
	}
}

func (cb *CheckboxPrimitive) IsEnabled() bool {
	return cb.enabled
}

func (cb *CheckboxPrimitive) SetEnabled(enabled bool) *CheckboxPrimitive {
	cb.enabled = enabled
	if enabled {
		cb.SetText(cb.onChar)
		cb.SetTextColor(cb.colors.Foreground)
	} else {
		cb.SetText(cb.offChar)
		cb.SetTextColor(tcell.ColorGray)
	}
	if cb.changedFunc != nil {
		cb.changedFunc(enabled)
	}
	return cb
}

func (cb *CheckboxPrimitive) SetChangedFunc(handler func(bool)) *CheckboxPrimitive {
	cb.changedFunc = handler
	return cb
}
