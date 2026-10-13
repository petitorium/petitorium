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
	currentCookiesList = cookiesList
	currentCookiesTab = cookiesContainer

	saveCallback := func() {
		syncCookiesFromUI(ui.WorkspaceData)
		SaveCookies(ui.WorkspaceData)
	}

	var refreshCookiesUI func()
	refreshCookiesUI = func() {
		cookiesList.Clear()
		for _, cookieRow := range currentCookieRows {
			cookiesList.AddItem(cookieRow.Row, rowHeight, 0, false)
			separator := tview.NewBox().SetBackgroundColor(colors.Background)
			separator.SetBorder(false)
			cookiesList.AddItem(separator, 1, 0, false)
		}
		if len(currentCookieRows) == 0 {
			emptyLabel := tview.NewTextView()
			emptyLabel.SetText("No cookies in jar")
			emptyLabel.SetTextColor(colors.Foreground)
			emptyLabel.SetBackgroundColor(colors.Background)
			emptyLabel.SetTextAlign(tview.AlignCenter)
			cookiesList.AddItem(emptyLabel, 1, 0, false)
		}
	}
	refreshCookiesTab = refreshCookiesUI

	buttonRow := tview.NewFlex().SetDirection(tview.FlexColumn)
	buttonRow.SetBackgroundColor(colors.Background)

	addButton := createThemedButton(" Add Cookie ", colors)
	currentAddCookieButton = addButton
	addButton.SetSelectedFunc(func() {
		addCookieRow(cookiesList, colors, refreshCookiesUI, saveCallback, func(p tview.Primitive) { ui.App.SetFocus(p) }, ui.UpdateFooter, ui.App, ui.Pages)
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
	buttonRow.AddItem(nil, 1, 0, false)
	buttonRow.AddItem(clearAllButton, 15, 0, false)
	buttonRow.AddItem(nil, 0, 1, false)

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
		if ui.ExitModal != nil {
			ui.ExitModal()
		}
		if ui.CookiesConfigButton != nil {
			ui.App.SetFocus(ui.CookiesConfigButton)
		} else if previousFocus != nil {
			ui.App.SetFocus(previousFocus)
		}
	}

	cookiesContainer.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			closeCookiesModal()
			return nil
		}
		return event
	})

	modal := createSizedModal(cookiesContainer, modalSizeFullscreen, colors.Background)
	ui.Pages.AddPage("cookies", modal, true, true)

	// Focus the first cookie input if present, otherwise the add button.
	if len(currentCookieRows) > 0 && currentCookieRows[0].DomainInput != nil {
		ui.App.SetFocus(currentCookieRows[0].DomainInput)
	} else {
		ui.App.SetFocus(addButton)
	}
}
