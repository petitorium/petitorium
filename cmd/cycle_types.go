package cmd

import (
	"github.com/rivo/tview"
)

// Cycle interface for Tab navigation
type Cycle interface {
	Next() tview.Primitive
	Prev() tview.Primitive
	GetCurrent() tview.Primitive
	Contains(p tview.Primitive) bool
	GetParent() Cycle
}

// MainCycle handles cycling between main panels
type MainCycle struct {
	panels   []tview.Primitive
	current  int
	children []tview.Primitive
}

func (c *MainCycle) Next() tview.Primitive {
	c.current = (c.current + 1) % len(c.panels)
	return c.panels[c.current]
}

func (c *MainCycle) Prev() tview.Primitive {
	c.current = (c.current - 1 + len(c.panels)) % len(c.panels)
	return c.panels[c.current]
}

func (c *MainCycle) GetCurrent() tview.Primitive {
	return c.panels[c.current]
}

func (c *MainCycle) Contains(p tview.Primitive) bool {
	for _, panel := range c.panels {
		if panel == p {
			return true
		}
	}
	return false
}

func (c *MainCycle) GetParent() Cycle {
	return nil
}

// URLBarCycle handles cycling within URL bar elements
type URLBarCycle struct {
	elements []tview.Primitive
	current  int
	parent   Cycle
	// children     []tview.Primitive
	// currentChild int
}

func (c *URLBarCycle) Next() tview.Primitive {
	if c.current < len(c.elements)-1 {
		c.current++
		return c.elements[c.current]
	} else {
		c.current = 0
		return nil // wrap, let parent handle
	}
}

func (c *URLBarCycle) Prev() tview.Primitive {
	if c.current > 0 {
		c.current--
		return c.elements[c.current]
	} else {
		c.current = len(c.elements) - 1
		return nil
	}
}

func (c *URLBarCycle) GetCurrent() tview.Primitive {
	return c.elements[c.current]
}

func (c *URLBarCycle) Contains(p tview.Primitive) bool {
	for _, elem := range c.elements {
		if elem == p {
			return true
		}
	}
	return false
}

func (c *URLBarCycle) GetParent() Cycle {
	return c.parent
}

// func (c *URLBarCycle) GetChildren() []tview.Primitive {
// 	return c.children
// }

// func (c *URLBarCycle) GetCurrentChildIndex() int {
// 	return c.currentChild
// }

// HeadersCycle handles cycling through header inputs
type HeadersCycle struct {
	inputs   []tview.Primitive
	current  int
	parent   Cycle
	children []tview.Primitive
}

func (c *HeadersCycle) Next() tview.Primitive {
	if c.current < len(c.inputs)-1 {
		c.current++
		return c.inputs[c.current]
	} else {
		c.current = 0
		return nil
	}
}

func (c *HeadersCycle) Prev() tview.Primitive {
	if c.current > 0 {
		c.current--
		return c.inputs[c.current]
	} else {
		c.current = len(c.inputs) - 1
		return nil
	}
}

func (c *HeadersCycle) GetCurrent() tview.Primitive {
	return c.inputs[c.current]
}

func (c *HeadersCycle) Contains(p tview.Primitive) bool {
	for _, input := range c.inputs {
		if input == p {
			return true
		}
	}
	return false
}

func (c *HeadersCycle) GetParent() Cycle {
	return c.parent
}

func (c *HeadersCycle) UpdateInputs() {
	c.inputs = []tview.Primitive{}
	for _, row := range currentHeaderRows {
		c.inputs = append(c.inputs, row.KeyInput, row.ValueInput)
	}
}

// EnvironmentsCycle handles cycling through header inputs
type EnvironmentsCycle struct {
	inputs   []tview.Primitive
	current  int
	parent   Cycle
	children []tview.Primitive
}

func (c *EnvironmentsCycle) Next() tview.Primitive {
	if c.current < len(c.inputs)-1 {
		c.current++
		return c.inputs[c.current]
	} else {
		c.current = 0
		return nil
	}
}

func (c *EnvironmentsCycle) Prev() tview.Primitive {
	if c.current > 0 {
		c.current--
		return c.inputs[c.current]
	} else {
		c.current = len(c.inputs) - 1
		return nil
	}
}

func (c *EnvironmentsCycle) GetCurrent() tview.Primitive {
	return c.inputs[c.current]
}

func (c *EnvironmentsCycle) Contains(p tview.Primitive) bool {
	for _, input := range c.inputs {
		if input == p {
			return true
		}
	}
	return false
}

func (c *EnvironmentsCycle) GetParent() Cycle {
	return c.parent
}

func (c *EnvironmentsCycle) UpdateInputs() {
	c.inputs = []tview.Primitive{}
	for _, row := range currentHeaderRows {
		c.inputs = append(c.inputs, row.KeyInput, row.ValueInput)
	}
}

// WorkspaceCycle handles cycling through workspace panel inputs
type WorkspaceCycle struct {
	inputs   []tview.Primitive
	current  int
	parent   Cycle
	children []tview.Primitive
}

func (c *WorkspaceCycle) Next() tview.Primitive {
	if c.current < len(c.inputs)-1 {
		c.current++
		return c.inputs[c.current]
	} else {
		c.current = 0
		return nil
	}
}

func (c *WorkspaceCycle) Prev() tview.Primitive {
	if c.current > 0 {
		c.current--
		return c.inputs[c.current]
	} else {
		c.current = len(c.inputs) - 1
		return nil
	}
}

func (c *WorkspaceCycle) GetCurrent() tview.Primitive {
	return c.inputs[c.current]
}

func (c *WorkspaceCycle) Contains(p tview.Primitive) bool {
	for _, input := range c.inputs {
		if input == p {
			return true
		}
	}
	return false
}

func (c *WorkspaceCycle) GetParent() Cycle {
	return c.parent
}
