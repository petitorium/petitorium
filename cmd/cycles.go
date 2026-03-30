package cmd

import (
	"github.com/rivo/tview"
)

// findInnermostCycle finds the most specific cycle containing the given primitive
func findInnermostCycle(p tview.Primitive) Cycle {
	if queryParamsCycle != nil && queryParamsCycle.Contains(p) {
		return queryParamsCycle
	}
	if headersCycle != nil && headersCycle.Contains(p) {
		return headersCycle
	}
	if urlBarCycle != nil && urlBarCycle.Contains(p) {
		return urlBarCycle
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

// Global cycle instances
var mainCycle *MainCycle

var urlBarCycle *URLBarCycle

var headersCycle *HeadersCycle

var queryParamsCycle *QueryParamsCycle

var environmentsCycle *EnvironmentsCycle

var workspaceCycle *WorkspaceCycle
