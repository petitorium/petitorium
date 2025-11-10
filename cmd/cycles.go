package cmd

import (
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
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

// RequestCycle handles cycling within request panel elements
type RequestCycle struct {
	elements []tview.Primitive
	current  int
	parent   Cycle
	// children     []tview.Primitive
	// currentChild int
}

func (c *RequestCycle) Next() tview.Primitive {
	if c.current < len(c.elements)-1 {
		c.current++
		return c.elements[c.current]
	} else {
		c.current = 0
		return nil // wrap, let parent handle
	}
}

func (c *RequestCycle) Prev() tview.Primitive {
	if c.current > 0 {
		c.current--
		return c.elements[c.current]
	} else {
		c.current = len(c.elements) - 1
		return nil
	}
}

func (c *RequestCycle) GetCurrent() tview.Primitive {
	return c.elements[c.current]
}

func (c *RequestCycle) Contains(p tview.Primitive) bool {
	for _, elem := range c.elements {
		if elem == p {
			return true
		}
	}
	return false
}

func (c *RequestCycle) GetParent() Cycle {
	return c.parent
}

// func (c *RequestCycle) GetChildren() []tview.Primitive {
// 	return c.children
// }

// func (c *RequestCycle) GetCurrentChildIndex() int {
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

// findInnermostCycle finds the most specific cycle containing the given primitive
func findInnermostCycle(p tview.Primitive) Cycle {
	if headersCycle != nil && headersCycle.Contains(p) {
		return headersCycle
	}
	if requestCycle != nil && requestCycle.Contains(p) {
		return requestCycle
	}
	if environmentsCycle != nil && environmentsCycle.Contains(p) {
		return environmentsCycle
	}
	if workspaceCycle != nil && workspaceCycle.Contains(p) {
		return workspaceCycle
	}
	if mainCycle != nil && mainCycle.Contains(p) {
		return mainCycle
	}
	return nil
}

// findRequestPtr finds the request pointer in workspace collections
func findRequestPtr(data *workspace.Workspace, req workspace.Request) *workspace.Request {
	// Search in collections
	for i := range data.Collections {
		for j := range data.Collections[i].Requests {
			if data.Collections[i].Requests[j].Name == req.Name &&
				data.Collections[i].Requests[j].Method == req.Method &&
				data.Collections[i].Requests[j].URL == req.URL &&
				data.Collections[i].Requests[j].Body == req.Body {
				return &data.Collections[i].Requests[j]
			}
		}
		if ptr := findRequestPtrInCollections(data.Collections[i].Collections, req); ptr != nil {
			return ptr
		}
	}
	return nil
}

// Helper function to find in nested collections
func findRequestPtrInCollections(data []workspace.Collection, req workspace.Request) *workspace.Request {
	for i := range data {
		for j := range data[i].Requests {
			if data[i].Requests[j].Name == req.Name &&
				data[i].Requests[j].Method == req.Method &&
				data[i].Requests[j].URL == req.URL &&
				data[i].Requests[j].Body == req.Body {
				return &data[i].Requests[j]
			}
		}
		if ptr := findRequestPtrInCollections(data[i].Collections, req); ptr != nil {
			return ptr
		}
	}
	return nil
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

// Global cycle instances
var mainCycle *MainCycle

var requestCycle *RequestCycle

var headersCycle *HeadersCycle

var environmentsCycle *EnvironmentsCycle

var workspaceCycle *WorkspaceCycle
