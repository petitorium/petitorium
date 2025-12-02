package cmd

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// ColorButton is a custom button implementation that supports background colors
type ColorButton struct {
	*tview.Box
	text                string
	onSelected          func()
	colors              *ColorManager
	isActivated         bool
	backgroundColor     tcell.Color
	activatedColor      tcell.Color
	labelColor          tcell.Color
	labelActivatedColor tcell.Color
	border              bool
	borderColor         tcell.Color
	borderPadding       [4]int
}

// NewColorButton creates a new color button
func NewColorButton(text string, colors *ColorManager) *ColorButton {
	fmt.Printf("DEBUG: NewColorButton called with text='%s', ButtonBackground=%v\n", text, colors.ButtonBackground)
	box := tview.NewBox()
	box.SetBorder(false)
	box.SetBackgroundColor(colors.ButtonBackground)

	cb := &ColorButton{
		Box:                 box,
		text:                text,
		colors:              colors,
		backgroundColor:     colors.ButtonBackground,
		activatedColor:      colors.ButtonSelect,
		labelColor:          colors.Foreground,
		labelActivatedColor: colors.Background,
		isActivated:         false,
	}

	fmt.Printf("DEBUG: ColorButton created with Box bgColor=%v\n", box.GetBackgroundColor())
	return cb
}

// SetSelectedFunc sets the function to call when the button is selected
func (cb *ColorButton) SetSelectedFunc(handler func()) *ColorButton {
	fmt.Printf("DEBUG: SetSelectedFunc called on CustomButton with text='%s'\n", cb.text)
	cb.onSelected = handler
	return cb
}

// SetBackgroundColor sets the normal background color
func (cb *ColorButton) SetBackgroundColor(color tcell.Color) *ColorButton {
	cb.backgroundColor = color
	if !cb.isActivated {
		cb.Box.SetBackgroundColor(color)
	}
	return cb
}

// SetBackgroundColorActivated sets the activated background color
func (cb *ColorButton) SetBackgroundColorActivated(color tcell.Color) *ColorButton {
	cb.activatedColor = color
	return cb
}

// SetLabelColor sets the normal label color
func (cb *ColorButton) SetLabelColor(color tcell.Color) *ColorButton {
	cb.labelColor = color
	return cb
}

// SetLabelColorActivated sets the activated label color
func (cb *ColorButton) SetLabelColorActivated(color tcell.Color) *ColorButton {
	cb.labelActivatedColor = color
	return cb
}

// SetBorder sets whether the button has a border
func (cb *ColorButton) SetBorder(show bool) *ColorButton {
	cb.border = show
	cb.Box.SetBorder(show)
	return cb
}

// SetBorderColor sets the border color
func (cb *ColorButton) SetBorderColor(color tcell.Color) *ColorButton {
	cb.borderColor = color
	cb.Box.SetBorderColor(color)
	return cb
}

// SetBorderPadding sets the border padding
func (cb *ColorButton) SetBorderPadding(top, bottom, left, right int) *ColorButton {
	cb.borderPadding = [4]int{top, bottom, left, right}
	cb.Box.SetBorderPadding(top, bottom, left, right)
	return cb
}

// SetStyle sets the button style (compatibility method)
func (cb *ColorButton) SetStyle(style tcell.Style) *ColorButton {
	// This is a compatibility method - we ignore it since we manage colors directly
	return cb
}

// updateBackground updates the background color based on activation state
func (cb *ColorButton) updateBackground() {
	if cb.isActivated && cb.activatedColor != tcell.ColorDefault {
		cb.Box.SetBackgroundColor(cb.activatedColor)
	} else if cb.backgroundColor != tcell.ColorDefault {
		cb.Box.SetBackgroundColor(cb.backgroundColor)
	}
}

// Draw implements the Primitive interface
func (cb *ColorButton) Draw(screen tcell.Screen) {
	// Debug: Print the background color being used
	boxBg := cb.Box.GetBackgroundColor()
	fmt.Printf("DEBUG: ColorButton Draw - Box bgColor: %v, Custom bgColor: %v, isActivated: %t\n", boxBg, cb.backgroundColor, cb.isActivated)

	cb.Box.Draw(screen)

	x, y, width, height := cb.Box.GetRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Calculate text position (centered)
	textLen := len(cb.text)
	textX := x + (width-textLen)/2
	textY := y + height/2

	// Choose colors based on activation state
	bgColor := cb.Box.GetBackgroundColor()
	fgColor := cb.labelColor
	if cb.isActivated && cb.labelActivatedColor != tcell.ColorDefault {
		fgColor = cb.labelActivatedColor
	}

	// Debug: Print what we're drawing
	fmt.Printf("DEBUG: Drawing text '%s' with bgColor: %v, fgColor: %v\n", cb.text, bgColor, fgColor)

	// Draw each character of the text
	for i, ch := range cb.text {
		if textX+i >= x && textX+i < x+width && textY >= y && textY < y+height {
			screen.SetContent(textX+i, textY, ch, nil, tcell.StyleDefault.Background(bgColor).Foreground(fgColor))
		}
	}
}

// InputHandler implements the Primitive interface
func (cb *ColorButton) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		if event.Key() == tcell.KeyEnter && cb.onSelected != nil {
			cb.isActivated = true
			cb.updateBackground()
			cb.onSelected()
			// Reset activation after a short delay
			go func() {
				// Small delay to show activated state
				// Note: In a real implementation, you'd want to redraw here
				cb.isActivated = false
				cb.updateBackground()
			}()
		}
	}
}

// MouseHandler implements the Primitive interface for mouse events
func (cb *ColorButton) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
			x, y := event.Position()
			bx, by, width, height := cb.Box.GetRect()

			// Check if click is within button bounds
			if x >= bx && x < bx+width && y >= by && y < by+height {
				if cb.onSelected != nil {
					cb.isActivated = true
					cb.updateBackground()
					cb.onSelected()
					// Reset activation after a short delay
					go func() {
						// Small delay to show activated state
						cb.isActivated = false
						cb.updateBackground()
					}()
				}
				return true, nil
			}
		}
		return false, nil
	}
}

// createColorButton creates a button with full background color support
func createColorButton(text string, colors *ColorManager) *ColorButton {
	return NewColorButton(text, colors)
}

// TestCustomButton demonstrates that CustomButton works with background colors
func TestCustomButton(colors *ColorManager) {
	// Create a simple test layout to show the difference
	app := tview.NewApplication()

	// Regular button (no background color)
	regularBtn := tview.NewButton("Regular Button")
	regularBtn.SetLabelColor(colors.Foreground)
	regularBtn.SetBackgroundColorActivated(colors.ButtonSelect)

	// Custom button (with background color)
	customBtn := NewCustomButtonWithColors("Custom Button", colors)
	customBtn.SetSelectedFunc(func() {
		fmt.Println("Custom button clicked!")
	})

	// Layout
	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(tview.NewTextView().SetText("Button Background Test").SetTextAlign(tview.AlignCenter), 1, 0, false).
		AddItem(tview.NewTextView().SetText("Regular (transparent):").SetTextAlign(tview.AlignCenter), 1, 0, false).
		AddItem(regularBtn, 1, 0, false).
		AddItem(tview.NewTextView().SetText("Custom (with bg color):").SetTextAlign(tview.AlignCenter), 1, 0, false).
		AddItem(customBtn, 1, 0, true)

	layout.SetBackgroundColor(colors.Background)

	if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}
