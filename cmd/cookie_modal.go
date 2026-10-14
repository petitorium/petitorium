package cmd

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showCookiesModal opens a fullscreen modal hosting the cookie jar editor. It
// mirrors the environment/workspace config modals: the cookie list is shown
// behind the cookies panel's config button, and edits are persisted on close.
func showCookiesModal(ui *UIOrchestrator) {
	previousFocus := ui.App.GetFocus()
	EnterModal := ui.EnterModal
	if EnterModal != nil {
		EnterModal()
	}

	colors := ui.Colors
	cookieJar := &ui.WorkspaceData.CookieJar

	cookiesContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	cookiesContainer.SetBorder(true)
	cookiesContainer.SetTitle(" Cookies ")
	cookiesContainer.SetBackgroundColor(colors.Background)
	cookiesContainer.SetBorderColor(colors.BorderFocus)
	cookiesContainer.SetTitleColor(colors.Title)

	cookiesList := tview.NewFlex().SetDirection(tview.FlexRow)
	cookiesList.SetBackgroundColor(colors.Background)

	currentCookieRows = nil
	cookieScrollOffset = 0
	currentCookiesList = cookiesList
	currentCookiesTab = cookiesContainer

	saveCallback := func() {
		syncCookiesFromUI(ui.WorkspaceData)
		SaveCookies(ui.WorkspaceData)
	}

	var refreshCookiesUI func()
	refreshCookiesUI = func() {
		cookiesList.Clear()

		// Each cookie row is followed by a 1-row separator, so a row occupies
		// two rows in the list. Determine how many rows fit in the visible area
		// so the list scrolls instead of overflowing the bottom of the modal.
		_, _, _, listHeight := cookiesList.GetRect()
		if listHeight <= 0 {
			// The list has not been laid out yet (e.g. during initial build).
			// The fullscreen modal is 40 rows tall; after the border, buttons,
			// spacer and header, roughly 33 rows remain for the cookie list.
			listHeight = 33
		}
		visibleRows := listHeight / 2
		if visibleRows < 1 {
			visibleRows = 1
		}

		total := len(currentCookieRows)
		if cookieScrollOffset > total-visibleRows {
			cookieScrollOffset = total - visibleRows
		}
		if cookieScrollOffset < 0 {
			cookieScrollOffset = 0
		}

		start := cookieScrollOffset
		end := start + visibleRows
		if end > total {
			end = total
		}

		for _, cookieRow := range currentCookieRows[start:end] {
			cookiesList.AddItem(cookieRow.Row, rowHeight, 0, false)
			separator := tview.NewBox().SetBackgroundColor(colors.Background)
			separator.SetBorder(false)
			cookiesList.AddItem(separator, 1, 0, false)
		}
		if total == 0 {
			emptyLabel := tview.NewTextView()
			emptyLabel.SetText("No cookies in jar")
			emptyLabel.SetTextColor(colors.Foreground)
			emptyLabel.SetBackgroundColor(colors.Background)
			emptyLabel.SetTextAlign(tview.AlignCenter)
			cookiesList.AddItem(emptyLabel, 1, 0, false)
		}
		cookiesList.AddItem(tview.NewBox().SetBackgroundColor(colors.Background), 0, 1, false)
	}
	refreshCookiesTab = refreshCookiesUI

	buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonRow.SetBackgroundColor(colors.Background)

	addButton := createThemedButton(" Add Cookie ", colors)
	currentAddCookieButton = addButton
	addButton.SetSelectedFunc(func() {
		addCookieRow(cookiesList, colors, refreshCookiesUI, saveCallback, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter, ui.App, ui.Pages)
		// Scroll to the bottom so the newly added row is visible.
		cookieScrollOffset = len(currentCookieRows)
		refreshCookiesUI()
		ui.UpdateFooter()
	})

	clearAllButton := createThemedButton(" Delete All ", colors)
	currentDeleteAllCookiesButton = clearAllButton
	clearAllButton.SetSelectedFunc(func() {
		clearCallback := func() {
			currentCookieRows = nil
			cookieJar.ClearAll()
			refreshCookiesUI()
			saveCallback()
		}
		form := createDeleteAllCookiesConfirm(ui.App, ui.Pages, colors, clearCallback)
		modal := createSizedModal(form, modalSizeForm, tcell.ColorDefault)
		ui.Pages.AddPage("clearAllCookies", modal, true, true)
		ui.App.SetFocus(form)
	})

	buttonRow.AddItem(addButton, 15, 0, false)
	buttonRow.AddItem(tview.NewBox().SetBackgroundColor(colors.Background), 1, 0, false)
	buttonRow.AddItem(clearAllButton, 15, 0, false)
	buttonRow.AddItem(tview.NewBox().SetBackgroundColor(colors.Background), 0, 1, false)

	cookiesContainer.AddItem(buttonRow, 1, 0, false)

	spacer := tview.NewBox().SetBackgroundColor(colors.Background)
	cookiesContainer.AddItem(spacer, 1, 0, false)

	headerRow := createCookiesTableHeader(colors)
	cookiesContainer.AddItem(headerRow, 1, 0, false)
	cookiesContainer.AddItem(cookiesList, 0, 1, false)

	for _, cookie := range cookieJar.Cookies {
		addCookieRowWithData(cookiesList, colors, cookie.Domain, cookie.Name, cookie.Value, cookie.Path, cookie.Secure, cookie.HttpOnly, cookie.Enabled, refreshCookiesUI, saveCallback, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter, ui.App, ui.Pages)
	}

	refreshCookiesUI()

	closeCookiesModal := func() {
		saveCallback()
		ui.Pages.RemovePage("cookies")
		ui.Pages.SwitchToPage("main")
		if ui.RefreshCookiesSummary != nil {
			ui.RefreshCookiesSummary()
		}
		if ui.ExitModal != nil {
			ui.ExitModal()
		}
		if ui.CookiesConfigButton != nil {
			ui.App.SetFocus(ui.CookiesConfigButton)
		} else if previousFocus != nil {
			ui.App.SetFocus(previousFocus)
		}
	}

	// focusCookieElement moves focus to the cookie element described by the
	// current navigation state (NavCurrentCookieRowElement / NavCurrentCookieElement).
	focusCookieElement := func() {
		switch {
		case ui.NavCurrentCookieRowElement == 0:
			ui.App.SetFocus(addButton)
		case ui.NavCurrentCookieRowElement == 1:
			ui.App.SetFocus(clearAllButton)
		case ui.NavCurrentCookieRowElement >= 2 && ui.NavCurrentCookieRowElement-2 < len(currentCookieRows):
			row := currentCookieRows[ui.NavCurrentCookieRowElement-2]
			switch ui.NavCurrentCookieElement {
			case 0:
				ui.App.SetFocus(row.DomainInput)
			case 1:
				ui.App.SetFocus(row.NameInput)
			case 2:
				ui.App.SetFocus(row.ValueInput)
			case 3:
				ui.App.SetFocus(row.PathInput)
			case 4:
				ui.App.SetFocus(row.SecureInput)
			case 5:
				ui.App.SetFocus(row.HttpOnlyInput)
			case 6:
				ui.App.SetFocus(row.Checkbox)
			case 7:
				ui.App.SetFocus(row.DeleteButton)
			}
		default:
			ui.App.SetFocus(addButton)
		}
	}

	// moveCookieFocus advances (delta=+1) or rewinds (delta=-1) the cookie nav
	// state and focuses the resulting element, wrapping within the modal. The
	// Add Cookie and Delete All buttons each count as a single stop; each cookie
	// row counts as eight stops (Domain, Name, Value, Path, Secure, HttpOnly,
	// Checkbox, Delete).
	moveCookieFocus := func(delta int) {
		row := ui.NavCurrentCookieRowElement
		element := ui.NavCurrentCookieElement

		if row == 0 || row == 1 {
			// On a button, move to the next/prev row (buttons are single stops).
			row += delta
			element = 0
		} else {
			// Within a cookie row, move to the next/prev element.
			element += delta
			if element < 0 || element > 7 {
				row += delta
				if delta > 0 {
					element = 0
				} else {
					element = 7
				}
			}
		}

		// Wrap the row index around: 2 buttons + one row per cookie.
		rowCount := 2 + len(currentCookieRows)
		if rowCount == 0 {
			rowCount = 1
		}
		row = ((row % rowCount) + rowCount) % rowCount

		// Clamp the element index for the resulting row.
		if row == 0 || row == 1 {
			element = 0
		} else if element < 0 {
			element = 0
		} else if element > 7 {
			element = 7
		}

		ui.NavCurrentCookieRowElement = row
		ui.NavCurrentCookieElement = element
		focusCookieElement()
		ui.UpdateFooter()
	}

	// isEditingCookieField reports whether the currently focused primitive is a
	// cookie field in edit mode (where typing 'q', 'i' or Esc should reach the
	// field rather than navigate/close the modal).
	isEditingCookieField := func() bool {
		for _, row := range currentCookieRows {
			for _, in := range []interface{ IsEditMode() bool }{
				row.DomainInput, row.NameInput, row.ValueInput, row.PathInput,
				row.SecureInput, row.HttpOnlyInput,
			} {
				if in.IsEditMode() {
					return true
				}
			}
		}
		return false
	}

	cookiesContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if isEditingCookieField() {
			// Let the editing field handle Esc (exit edit mode) and typed text.
			return event
		}
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			closeCookiesModal()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			moveCookieFocus(1)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			moveCookieFocus(-1)
			return nil
		}
		if event.Rune() == 'j' {
			cookieScrollOffset++
			refreshCookiesUI()
			return nil
		}
		if event.Rune() == 'k' {
			if cookieScrollOffset > 0 {
				cookieScrollOffset--
			}
			refreshCookiesUI()
			return nil
		}
		return event
	})

	modal := createSizedModal(cookiesContainer, modalSizeFullscreen, colors.Background)
	ui.Pages.AddPage("cookies", modal, true, true)

	// Reset and initialize the navigation state, then focus the first element
	// (a cookie row's domain field when present, otherwise the add button).
	ui.NavCurrentCookieRowElement = 0
	ui.NavCurrentCookieElement = 0
	if len(currentCookieRows) > 0 && currentCookieRows[0].DomainInput != nil {
		ui.NavCurrentCookieRowElement = 2
	}
	focusCookieElement()
	ui.UpdateFooter()
}
