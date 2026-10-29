package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// clipScreen wraps a tcell.Screen and discards any content written outside a
// given rectangle. It provides CSS `overflow: hidden` semantics for tview
// primitives whose children would otherwise paint past their bounds (for
// example a Flex list with more rows than fit in the panel).
type clipScreen struct {
	tcell.Screen
	x, y, width, height int
}

func (s *clipScreen) inClip(x, y int) bool {
	return x >= s.x && x < s.x+s.width && y >= s.y && y < s.y+s.height
}

// SetContent only forwards cells that are inside the clip rectangle.
func (s *clipScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	if !s.inClip(x, y) {
		return
	}
	s.Screen.SetContent(x, y, primary, combining, style)
}

// GetContent returns a blank cell outside the clip rectangle.
func (s *clipScreen) GetContent(x, y int) (rune, []rune, tcell.Style, int) {
	if !s.inClip(x, y) {
		return ' ', nil, tcell.StyleDefault, 1
	}
	return s.Screen.GetContent(x, y)
}

// Fill only fills the clip rectangle.
func (s *clipScreen) Fill(primary rune, style tcell.Style) {
	for y := s.y; y < s.y+s.height; y++ {
		for x := s.x; x < s.x+s.width; x++ {
			s.Screen.SetContent(x, y, primary, nil, style)
		}
	}
}

// Clear only clears the clip rectangle.
func (s *clipScreen) Clear() {
	s.Fill(' ', tcell.StyleDefault)
}

// ShowCursor hides the cursor when it would be placed outside the clip area.
func (s *clipScreen) ShowCursor(x, y int) {
	if !s.inClip(x, y) {
		s.Screen.HideCursor()
		return
	}
	s.Screen.ShowCursor(x, y)
}

// clipBox is a tview primitive that clips its child to its own bounds. It is a
// drop-in container for a list Flex so that overflowing rows are hidden instead
// of being drawn over neighbouring panels.
type clipBox struct {
	*tview.Box
	child     tview.Primitive
	rowHeight int // height in rows of a single list row (always 1)
	rowStride int // list rows occupied by one data row (2 for headers/query, 1 for multipart)
	scrollRow int // number of list rows scrolled off the top
}

// newClipBox wraps child in a clipping container with the given background
// colour.
func newClipBox(child tview.Primitive, backgroundColor tcell.Color) *clipBox {
	c := &clipBox{
		Box:       tview.NewBox(),
		child:     child,
		rowHeight: 1,
		rowStride: 1,
	}
	c.SetBackgroundColor(backgroundColor)
	return c
}

// SetRowStride sets how many list rows each data row occupies. This is used to
// translate between data-row indices and scroll offsets.
func (c *clipBox) SetRowStride(stride int) {
	if stride < 1 {
		stride = 1
	}
	c.rowStride = stride
}

// SetScrollRow sets the scroll offset in list rows, clamped to non-negative.
func (c *clipBox) SetScrollRow(row int) {
	if row < 0 {
		row = 0
	}
	c.scrollRow = row
}

// ScrollDataRowsBy adjusts the scroll offset by delta data rows.
func (c *clipBox) ScrollDataRowsBy(delta int) {
	c.scrollRow += delta * c.rowStride
	if c.scrollRow < 0 {
		c.scrollRow = 0
	}
}

// VisibleDataRows returns how many data rows fit in the visible area.
func (c *clipBox) VisibleDataRows() int {
	_, _, _, height := c.GetInnerRect()
	if height <= 0 {
		return 0
	}
	visible := height / c.rowStride
	if visible < 1 {
		visible = 1
	}
	return visible
}

// scrollDataRowIntoView ensures the given 0-indexed data row is fully visible,
// adjusting the scroll offset if needed.
func (c *clipBox) scrollDataRowIntoView(row int) {
	visible := c.VisibleDataRows()
	if visible <= 0 {
		return
	}
	top := c.scrollRow / c.rowStride
	if row < top {
		c.scrollRow = row * c.rowStride
	} else if row >= top+visible {
		c.scrollRow = (row - visible + 1) * c.rowStride
	}
	if c.scrollRow < 0 {
		c.scrollRow = 0
	}
}

// ClampScroll clamps the scroll offset so that at most the given number of data
// rows can be scrolled past the top of the view.
func (c *clipBox) ClampScroll(totalDataRows int) {
	visible := c.VisibleDataRows()
	if visible < 1 {
		visible = 1
	}
	max := totalDataRows - visible
	if max < 0 {
		max = 0
	}
	maxScroll := max * c.rowStride
	if c.scrollRow > maxScroll {
		c.scrollRow = maxScroll
	}
	if c.scrollRow < 0 {
		c.scrollRow = 0
	}
}

// Draw clips the child's rendering to this primitive's inner rectangle and
// applies the current scroll offset.
func (c *clipBox) Draw(screen tcell.Screen) {
	c.Box.Draw(screen)

	x, y, width, height := c.GetInnerRect()
	if width <= 0 || height <= 0 {
		return
	}

	// Shift the child up by the scroll offset. Rows above the clip area are
	// discarded by clipScreen; rows below it are discarded as well, giving the
	// same effect as CSS overflow: hidden plus a vertical scroll.
	offset := c.scrollRow * c.rowHeight
	c.child.SetRect(x, y-offset, width, height+offset)
	c.child.Draw(&clipScreen{
		Screen: screen,
		x:      x,
		y:      y,
		width:  width,
		height: height,
	})
}

// Focus delegates to the child so focus reaches the rows inside the list.
func (c *clipBox) Focus(delegate func(p tview.Primitive)) {
	c.child.Focus(delegate)
}

// HasFocus reports whether the child (or one of its descendants) has focus.
func (c *clipBox) HasFocus() bool {
	return c.child.HasFocus()
}

// Blur forwards blur to the child.
func (c *clipBox) Blur() {
	c.child.Blur()
}

// InputHandler forwards key events to the child.
func (c *clipBox) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return c.child.InputHandler()
}

// MouseHandler forwards mouse events to the child.
func (c *clipBox) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (bool, tview.Primitive) {
	return c.child.MouseHandler()
}

// PasteHandler forwards paste events to the child.
func (c *clipBox) PasteHandler() func(text string, setFocus func(p tview.Primitive)) {
	return c.child.PasteHandler()
}
