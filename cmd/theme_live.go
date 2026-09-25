package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
)

// applyThemeLive applies a unified theme to the running UI without restarting
// the application. It updates the global config, refreshes the shared
// ColorManager in place (so every primitive that captured a pointer to it sees
// the new palette), re-colors the existing primitives, and then re-renders the
// content whose text embeds theme colors.
func applyThemeLive(ui *UIOrchestrator, themeName string) error {
	tm := GetThemeManager()
	if err := tm.ApplyTheme(themeName); err != nil {
		return err
	}

	// Keep the existing ColorManager pointer alive instead of replacing it:
	// SetupUI's focus/border closures and several custom primitives captured
	// the original pointer, so copying the new values into it keeps all of
	// them in sync.
	*ui.Colors = *tm.GetColorManager()

	recolorUI(ui)
	refreshThemeSensitiveUI(ui)

	return nil
}

// recolorUI restyles the existing primitives and the collections tree in place.
func recolorUI(ui *UIOrchestrator) {
	c := ui.Colors

	recolorPrimitive(ui.WorkspacePanel, c)
	recolorPrimitive(ui.EnvironmentPanel, c)
	recolorPrimitive(ui.CollectionsTreeView, c)
	recolorPrimitive(ui.MethodURLBar, c)
	recolorPrimitive(ui.RequestDataTabs, c)
	recolorPrimitive(ui.Response, c)
	recolorPrimitive(ui.Footer, c)
	recolorPrimitive(ui.TabPages, c)
	recolorPrimitive(ui.ResponsePages, c)

	// Pages containers do not expose their children, so walk each tab page
	// root explicitly in addition to the top-level containers above.
	for _, p := range []tview.Primitive{
		ui.BodyViewPanel,
		ui.BodyEditPanel,
		ui.BodyContainer,
		ui.AuthTab,
		ui.QueryTab,
		ui.HeadersTab,
		ui.CookiesTab,
		ui.MultipartFieldsTab,
		ui.ResponseTabHeader,
		ui.ResponseInfoBar,
		ui.ResponsePreviewPanel,
		ui.ResponseHeadersPanel,
		ui.ResponseCookiesPanel,
		ui.ResponseTimelinePanel,
		ui.FooterLeft,
		ui.FooterRight,
	} {
		recolorPrimitive(p, c)
	}

	// Dropdowns and inputs are not always reached by the Flex walk (e.g. the
	// URL input lives inside a custom Pages primitive), so style them directly.
	recolorPrimitive(ui.WorkspaceSelector, c)
	recolorPrimitive(ui.EnvDropdown, c)
	recolorPrimitive(ui.MethodDropdown, c)
	recolorPrimitive(ui.ContentTypeDropdown, c)
	recolorPrimitive(ui.URLInput, c)

	// Buttons.
	recolorPrimitive(ui.SendButton, c)
	recolorPrimitive(ui.CurlButton, c)
	recolorPrimitive(ui.EnvConfigButton, c)
	recolorPrimitive(ui.WorkspaceConfigButton, c)
	recolorPrimitive(ui.MultipartAddButton, c)
	recolorPrimitive(ui.MultipartDeleteAllButton, c)
	recolorPrimitive(ui.AddHeaderButton, c)
	recolorPrimitive(ui.DeleteAllHeadersButton, c)
	recolorPrimitive(ui.AddQueryParamButton, c)
	recolorPrimitive(ui.DeleteAllQueryParamsButton, c)
	recolorPrimitive(ui.AddCookieButton, c)
	recolorPrimitive(ui.DeleteAllCookiesButton, c)

	// Hidden-border containers originally use the background as their border
	// color. After the background changes, update them so their borders don't
	// reappear with the previous theme's color.
	setBorderColor(ui.BodyViewPanel, c.Background)
	setBorderColor(ui.BodyEditPanel, c.Background)
	setBorderColor(ui.BodyContainer, c.Background)
	setBorderColor(ui.MultipartFieldsTab, c.Background)
	setBorderColor(ui.QueryTab, c.Background)
	setBorderColor(ui.HeadersTab, c.Background)
	setBorderColor(ui.CookiesTab, c.Background)
	for _, table := range []*tview.Table{ui.ResponseCookiesPanel, ui.ResponseTimelinePanel} {
		setBorderColor(table, c.Background)
	}

	// Real "container" borders.
	for _, p := range []tview.Primitive{
		ui.WorkspacePanel,
		ui.EnvironmentPanel,
		ui.CollectionsTreeView,
		ui.MethodURLBar,
		ui.RequestDataTabs,
		ui.Response,
	} {
		setBorderColor(p, c.Border)
		setTitleColor(p, c.Title)
	}

	// The footer's inner flex carries the visible border; ui.Footer is the
	// outer column that wraps it.
	if ui.Footer != nil && ui.Footer.GetItemCount() > 0 {
		setBorderColor(ui.Footer.GetItem(0), c.Border)
	}

	// Restore the focused panel's highlighted border.
	if ui.MainCycle != nil && ui.NavCurrentContainer < len(ui.MainCycle.panels) {
		ui.SetActiveBorder(ui.MainCycle.panels[ui.NavCurrentContainer])
	}

	recolorCollectionsTree(ui)
}

// recolorPrimitive restyles one primitive (and, for Flex containers, its
// children recursively) with the given palette.
func recolorPrimitive(p tview.Primitive, c *ColorManager) {
	if primitiveIsNil(p) {
		return
	}

	switch v := p.(type) {
	case *tview.Flex:
		v.SetBackgroundColor(c.Background)
		for i := 0; i < v.GetItemCount(); i++ {
			recolorPrimitive(v.GetItem(i), c)
		}
	case *tview.TextView:
		v.SetBackgroundColor(c.Background)
		v.SetTextColor(c.Foreground)
	case *tview.InputField:
		v.SetBackgroundColor(c.Background)
		v.SetFieldBackgroundColor(c.InputBackground)
		v.SetFieldTextColor(c.Foreground)
		v.SetLabelColor(c.LabelColor)
	case *tview.DropDown:
		v.SetBackgroundColor(c.Background)
		v.SetFieldBackgroundColor(c.InputBackground)
		v.SetFieldTextColor(c.Foreground)
		v.SetLabelColor(c.LabelColor)
	case *tview.TextArea:
		v.SetBackgroundColor(c.Background)
		v.SetTextStyle(tcell.StyleDefault.Background(c.Background).Foreground(c.Foreground))
	case *tview.Table:
		v.SetBackgroundColor(c.Background)
		v.SetSelectedStyle(tcell.StyleDefault.Background(c.Selection).Foreground(c.Foreground))
	case *tview.TreeView:
		v.SetBackgroundColor(c.Background)
	case *tview.Pages:
		v.SetBackgroundColor(c.Background)
	case *CustomButton:
		recolorCustomButton(v, c)
	case *URLVariableInput:
		recolorURLInput(v, c)
	case *HeaderKeyInput:
		recolorHeaderKeyInput(v, c)
	case *HeaderValueInput:
		recolorHeaderValueInput(v, c)
	case *CheckboxPrimitive:
		recolorCheckbox(v, c)
	case *ColorButton:
		v.SetBackgroundColor(c.ButtonBackground)
		v.SetBackgroundColorActivated(c.ButtonSelect)
		v.SetLabelColor(c.Foreground)
		v.SetLabelColorActivated(c.Background)
	case *tview.Box:
		v.SetBackgroundColor(c.Background)
	}
}

// recolorCustomButton applies the standard themed-button palette to a custom
// button while leaving its already-configured selection handlers intact.
func recolorCustomButton(cb *CustomButton, c *ColorManager) {
	cb.SetBackgroundColor(c.ButtonBackground)
	cb.SetBackgroundColorActivated(c.ButtonSelect)
	cb.SetDisabledBackgroundColor(c.Selection)
	cb.SetLabelColor(c.Foreground)
	cb.SetLabelColorActivated(c.Background)
	cb.SetSendingLabelColor(c.BorderFocus)
}

// recolorURLInput re-applies the constructor palette to the dual-mode URL input.
func recolorURLInput(u *URLVariableInput, c *ColorManager) {
	u.colors = c
	u.viewMode.SetBackgroundColor(c.Background)
	u.viewMode.SetTextColor(c.Foreground)
	u.editMode.SetBackgroundColor(c.Background)
	u.editMode.SetFieldBackgroundColor(c.Background)
	u.editMode.SetFieldTextColor(c.Foreground)
	u.Pages.SetBackgroundColor(c.Background)
}

// recolorHeaderKeyInput re-applies the constructor palette to a key input.
func recolorHeaderKeyInput(h *HeaderKeyInput, c *ColorManager) {
	h.colors = c
	h.viewMode.SetBackgroundColor(c.InputBackground)
	h.viewMode.SetTextColor(c.Foreground)
	h.editMode.SetBackgroundColor(c.Background)
	h.editMode.SetFieldBackgroundColor(c.InputBackground)
	h.editMode.SetFieldTextColor(c.Foreground)
	h.Pages.SetBackgroundColor(c.Background)
}

// recolorHeaderValueInput re-applies the constructor palette to a value input.
func recolorHeaderValueInput(h *HeaderValueInput, c *ColorManager) {
	h.colors = c
	h.viewMode.SetBackgroundColor(c.InputBackground)
	h.viewMode.SetTextColor(c.Foreground)
	h.editMode.SetBackgroundColor(c.Background)
	h.editMode.SetFieldBackgroundColor(c.InputBackground)
	h.editMode.SetFieldTextColor(c.Foreground)
	h.Pages.SetBackgroundColor(c.Background)
}

// recolorCheckbox re-applies the constructor palette to a checkbox primitive.
func recolorCheckbox(cb *CheckboxPrimitive, c *ColorManager) {
	cb.colors = c
	cb.SetBackgroundColor(c.Background)
	if cb.IsEnabled() {
		cb.SetTextColor(c.Foreground)
	} else {
		cb.SetTextColor(tcell.ColorGray)
	}
}

// refreshThemeSensitiveUI re-renders content that embeds theme colors, such as
// syntax-highlighted bodies, HTTP method coloring, response tables and the
// tab/footer indicators.
func refreshThemeSensitiveUI(ui *UIOrchestrator) {
	ui.UpdateFooter()
	updateTabHeader(RequestTabDisplayNames, ui.TabHeader, ui.CurrentTabIndex, ui.Colors)
	updateResponseTabHeader(ui.ResponseTabHeader, ui.CurrentResponseTabIndex, ui.Colors)

	if ui.CurrentRequestID != "" {
		ui.refreshCurrentRequest()
	}
	req := ui.CurrentRequest

	var body string
	if req != nil {
		body = req.Body
	} else {
		body = ui.CurrentBodyContent
	}
	syncBodyContent(body, ui.BodyEditMode, ui.BodyEditPanel, ui.BodyViewPanel)

	if req != nil {
		saveCallback := func() {
			saveCurrentRequest(ui.CurrentRequest, ui.WorkspaceData)
		}
		focusSetter := func(p tview.Primitive) {
			ui.App.SetFocus(p)
		}

		setHeadersInUI(ui.Colors, req.Headers, saveCallback, focusSetter, ui.UpdateFooter)
		setQueryParamsInUI(ui.Colors, req.QueryParams, saveCallback, focusSetter, ui.UpdateFooter)

		if req.ContentType == "Multipart" {
			updateMultipartFieldsFromBody(req.Body, ui)
		}
	}

	RefreshCookiesTab(ui.WorkspaceData.CookieJar.Cookies, ui.Colors, ui.App, ui.Pages)

	if ui.LastResponse != nil {
		updateResponseTabs(
			ui.LastResponse,
			ui.LastResponseTime,
			ui.Response,
			ui.ResponseTabHeader,
			&ui.ResponseInfoBar,
			&ui.ResponseTimeText,
			&ui.LastResponseTime,
			ui.ResponsePreviewPanel,
			ui.ResponseHeadersPanel,
			ui.ResponseCookiesPanel,
			ui.ResponseTimelinePanel,
			ui.Colors,
			ui.CopyResponse,
			ui.SaveResponse,
		)
	}
}

// recolorCollectionsTree updates tree node text and selection styles in place so
// the current selection and expansion state survive a live theme change.
func recolorCollectionsTree(ui *UIOrchestrator) {
	if ui.RootNode == nil {
		return
	}

	c := ui.Colors
	bg := c.Background
	selBg := c.TreeSelection
	fg := c.Foreground

	selectedNode := ui.CollectionsTreeView.GetCurrentNode()

	var walk func(*tview.TreeNode)
	walk = func(node *tview.TreeNode) {
		node.SetTextStyle(tcell.StyleDefault.Background(bg))
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(selBg).Foreground(fg))

		if req := ui.requestFromNode(node); req != nil {
			coloredMethod := getColoredMethod(req.Method)
			paddedName := padNameToMinLength(req.Name, 4)

			showIcon := nodeIsRequest(node) && (node == selectedNode || node == ui.LastSelectedRequestNode)
			if showIcon {
				iconColor := config.C.UI.SelectedRequestIconColor
				coloredIcon := fmt.Sprintf("[%s]%s[-:-:-]", iconColor, config.C.UI.SelectedRequestIcon)
				node.SetText(fmt.Sprintf("%s%s%s", coloredIcon, coloredMethod, paddedName))
			} else {
				iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon)
				spacePadding := strings.Repeat(" ", iconWidth)
				node.SetText(fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName))
			}
		}

		for _, child := range node.GetChildren() {
			walk(child)
		}
	}

	walk(ui.RootNode)
}

// primitiveIsNil reports whether p is nil, including the case where a non-nil
// interface wraps a nil typed pointer.
func primitiveIsNil(p tview.Primitive) bool {
	if p == nil {
		return true
	}
	rv := reflect.ValueOf(p)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}

// setBorderColor is a small helper over the Box border-color setter so callers
// can pass any primitive that embeds *tview.Box.
func setBorderColor(p tview.Primitive, color tcell.Color) {
	if primitiveIsNil(p) {
		return
	}
	if b, ok := p.(interface{ SetBorderColor(tcell.Color) *tview.Box }); ok {
		b.SetBorderColor(color)
	}
}

// setTitleColor is a small helper over the Box title-color setter.
func setTitleColor(p tview.Primitive, color tcell.Color) {
	if primitiveIsNil(p) {
		return
	}
	if b, ok := p.(interface{ SetTitleColor(tcell.Color) *tview.Box }); ok {
		b.SetTitleColor(color)
	}
}
