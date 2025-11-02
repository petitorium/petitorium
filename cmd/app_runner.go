package cmd

// RunApp starts the TUI application and handles the main event loop
func RunApp(ui *UIOrchestrator) error {
	if err := ui.App.
		SetRoot(ui.Pages, true).
		SetFocus(ui.CollectionsTreeView).
		Run(); err != nil {
		return err
	}
	return nil
}
