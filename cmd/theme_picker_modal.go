package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
)

const themePickerPageName = "themePicker"

// ThemePickerModal is a searchable modal for previewing and switching unified
// themes. Moving the selection re-colors the running UI live, Enter persists the
// selection to config, and Esc reverts to the theme that was active on open.
type ThemePickerModal struct {
	*tview.Flex
	ui             *UIOrchestrator
	searchField    *tview.InputField
	table          *tview.Table
	themes         []string
	filtered       []string
	rowThemes      map[int]string
	previousTheme  string
	previewedTheme string
	committed      bool
	updating       bool
	returnFocus    tview.Primitive
}

// NewThemePickerModal creates a theme picker populated with the available
// unified themes. The current theme and focus are captured so previews can be
// reverted and focus restored on close.
func NewThemePickerModal(ui *UIOrchestrator) *ThemePickerModal {
	tm := GetThemeManager()
	previousTheme := tm.GetCurrentTheme()

	m := &ThemePickerModal{
		Flex:           tview.NewFlex().SetDirection(tview.FlexRow),
		ui:             ui,
		table:          tview.NewTable().SetSelectable(true, false),
		themes:         themePickerThemes(),
		rowThemes:      map[int]string{},
		previousTheme:  previousTheme,
		previewedTheme: previousTheme,
		returnFocus:    ui.App.GetFocus(),
	}

	m.SetBackgroundColor(ui.Colors.Background)
	m.table.SetSelectedStyle(tcell.StyleDefault.Background(ui.Colors.Selection).Foreground(ui.Colors.ActiveTab))
	m.table.SetBackgroundColor(ui.Colors.Background)

	m.searchField = createSearchField(" Select Theme: ", "Type to filter themes...", ui.Colors)

	m.searchField.SetChangedFunc(func(text string) {
		m.filterThemes(text)
	})

	m.searchField.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.cancel()
			return nil
		}
		if event.Key() == tcell.KeyDown || event.Key() == tcell.KeyTab {
			if len(m.filtered) > 0 {
				m.focusTable()
			}
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			m.commitCurrentRow()
			return nil
		}
		return event
	})

	m.table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := m.table.GetSelection()
		if event.Key() == tcell.KeyUp && row == m.firstThemeRow() {
			m.ui.App.SetFocus(m.searchField)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			m.ui.App.SetFocus(m.searchField)
			return nil
		}
		if event.Key() == tcell.KeyEnter {
			m.commitCurrentRow()
			return nil
		}
		if event.Rune() == 'j' {
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		}
		if event.Rune() == 'k' {
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		if event.Key() == tcell.KeyEscape || event.Rune() == 'q' {
			m.cancel()
			return nil
		}
		return event
	})

	m.table.SetSelectionChangedFunc(func(row, column int) {
		if m.updating {
			return
		}
		m.previewCurrentRow()
	})

	m.table.SetSelectedFunc(func(row, column int) {
		m.commitCurrentRow()
	})

	m.AddItem(m.searchField, 1, 0, true)
	m.AddItem(m.table, 0, 1, false)

	m.SetBorder(true).SetTitle(" Select Theme ")
	m.SetBorderColor(ui.Colors.BorderFocus)
	m.SetTitleColor(ui.Colors.Title)

	m.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			m.cancel()
			return nil
		}
		if event.Rune() == 'q' && m.ui.App.GetFocus() != m.searchField {
			m.cancel()
			return nil
		}
		return event
	})

	m.filterThemes("")

	return m
}

// themePickerThemes returns the sorted list of unified themes available to the
// picker. Dynamically supported themes are seeded in best-effort so the picker
// includes themes that Chroma can resolve beyond the hand-seeded defaults.
func themePickerThemes() []string {
	tm := GetThemeManager()
	for _, name := range getSupportedUnifiedThemes() {
		// Ignore errors: a theme that Chroma cannot resolve is simply omitted.
		_, _ = tm.GetTheme(name)
	}
	names := tm.GetAvailableThemes()
	sort.Strings(names)
	return names
}

// filterThemes narrows the visible list to themes whose name matches the query
// (case-insensitive) and re-renders the table.
func (m *ThemePickerModal) filterThemes(query string) {
	query = strings.ToLower(query)
	m.filtered = m.filtered[:0]
	for _, name := range m.themes {
		if query == "" || strings.Contains(strings.ToLower(name), query) {
			m.filtered = append(m.filtered, name)
		}
	}
	m.renderThemes("")
}

// renderThemes rebuilds the table rows. When selectName is non-empty and present
// in the filtered list, that row is re-selected; otherwise the current theme is
// selected when present, falling back to the first row.
func (m *ThemePickerModal) renderThemes(selectName string) {
	m.table.Clear()
	m.rowThemes = map[int]string{}

	if len(m.filtered) == 0 {
		m.table.SetCell(0, 0, tview.NewTableCell(" No matching themes ").
			SetTextColor(m.ui.Colors.Placeholder).
			SetSelectable(false).
			SetExpansion(1).
			SetAlign(tview.AlignLeft))
		return
	}

	current := GetThemeManager().GetCurrentTheme()

	for i, name := range m.filtered {
		row := i
		m.rowThemes[row] = name

		marker := ""
		if name == current {
			marker = "*"
		}

		m.table.SetCell(row, 0, tview.NewTableCell(" "+name+" ").
			SetTextColor(m.ui.Colors.Foreground).
			SetExpansion(3).
			SetAlign(tview.AlignLeft))
		m.table.SetCell(row, 1, tview.NewTableCell(" "+marker+" ").
			SetTextColor(m.ui.Colors.Placeholder).
			SetExpansion(1).
			SetAlign(tview.AlignCenter))
	}

	selected := 0
	for row, name := range m.rowThemes {
		if selectName != "" && name == selectName {
			selected = row
			break
		}
		if name == m.previousTheme {
			selected = row
		}
	}

	m.updating = true
	m.table.Select(selected, 0)
	m.updating = false
}

// firstThemeRow returns the row index of the first selectable theme row. The
// picker table has no header row, so the first theme is always row zero.
func (m *ThemePickerModal) firstThemeRow() int {
	return 0
}

// focusTable moves focus to the results table if there is at least one theme.
func (m *ThemePickerModal) focusTable() {
	if len(m.filtered) > 0 {
		m.ui.App.SetFocus(m.table)
	}
}

// previewCurrentRow applies the currently highlighted theme live.
func (m *ThemePickerModal) previewCurrentRow() {
	row, _ := m.table.GetSelection()
	name := m.rowThemes[row]
	if name == "" || name == m.previewedTheme {
		return
	}

	if err := applyThemeLive(m.ui, name); err != nil {
		return
	}

	m.previewedTheme = name
	m.renderThemes(name)
	m.restyleModal()
}

// commitCurrentRow applies and persists the currently selected theme.
func (m *ThemePickerModal) commitCurrentRow() {
	row, _ := m.table.GetSelection()
	name := m.rowThemes[row]
	if name == "" {
		return
	}

	if err := applyThemeLive(m.ui, name); err != nil {
		return
	}
	m.previewedTheme = name
	m.committed = true

	if err := config.SaveConfig(&config.C); err != nil {
		m.close()
		showErrorModalWithFocus(m.ui.App, m.ui.Pages, fmt.Sprintf("Failed to save theme: %v", err), m.returnFocus, m.ui.Colors)
		return
	}

	m.close()
}

// cancel reverts any live preview and closes the picker without saving.
func (m *ThemePickerModal) cancel() {
	if !m.committed && m.previewedTheme != m.previousTheme {
		_ = applyThemeLive(m.ui, m.previousTheme)
	}
	m.close()
}

// restyleModal re-applies the current palette to the picker chrome and search
// field after a live preview changes ui.Colors.
func (m *ThemePickerModal) restyleModal() {
	c := m.ui.Colors
	m.SetBackgroundColor(c.Background)
	m.SetBorderColor(c.BorderFocus)
	m.SetTitleColor(c.Title)
	m.table.SetSelectedStyle(tcell.StyleDefault.Background(c.Selection).Foreground(c.ActiveTab))
	m.table.SetBackgroundColor(c.Background)
	m.searchField.SetLabelColor(c.LabelColor)
	m.searchField.SetPlaceholderStyle(tcell.StyleDefault.Background(c.InputBackground).Foreground(c.Placeholder))
	m.searchField.SetFieldBackgroundColor(c.InputBackground)
	m.searchField.SetFieldTextColor(c.Foreground)
	m.searchField.SetBackgroundColor(c.Background)
}

// close removes the picker and restores the focus captured on open.
func (m *ThemePickerModal) close() {
	m.ui.Pages.RemovePage(themePickerPageName)
	m.ui.Pages.SwitchToPage("main")
	if m.returnFocus != nil {
		m.ui.App.SetFocus(m.returnFocus)
	}
	if m.ui.ExitModal != nil {
		m.ui.ExitModal()
	}
	m.ui.UpdateFooter()
}

// showThemePickerModal opens the theme picker.
func showThemePickerModal(ui *UIOrchestrator) {
	ui.EnterModal()
	m := NewThemePickerModal(ui)
	ui.Pages.AddPage(themePickerPageName, createSizedModal(m, modalSizeSearch, ui.Colors.Background), true, true)
	ui.App.SetFocus(m.searchField)
	ui.UpdateFooter()
}
