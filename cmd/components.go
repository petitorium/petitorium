package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// createEnvironmentCreationPanel creates the left panel for environment creation
func createEnvironmentCreationPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor tcell.Color,
	environments []workspace.Environment,
	onEnvironmentCreated func(),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(backgroundColor)
	form.SetBorderColor(borderColor)
	form.SetTitleColor(titleColor)
	form.SetFieldBackgroundColor(backgroundColor)
	form.SetFieldTextColor(foregroundColor)
	form.SetLabelColor(foregroundColor)
	form.SetButtonBackgroundColor(backgroundColor)
	form.SetButtonTextColor(foregroundColor)

	// Get base environment options
	baseOptions := []string{"None"}
	for _, env := range environments {
		baseOptions = append(baseOptions, env.Name)
	}

	form.AddInputField("Name", "", 20, nil, nil)
	form.AddDropDown("Base", baseOptions, 0, nil)

	form.AddButton("Create", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		baseIndex, _ := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()

		if strings.TrimSpace(name) == "" {
			return
		}

		// Check if environment name already exists
		for _, env := range environments {
			if env.Name == name {
				return // Name already exists
			}
		}

		// Create new environment
		newEnv := workspace.Environment{
			Name:      name,
			Variables: make(map[string]string),
		}

		// Set base if selected
		if baseIndex > 0 {
			newEnv.Base = baseOptions[baseIndex]
		}

		// Add to environments
		environments = append(environments, newEnv)

		// Save environments
		if err := workspace.SaveEnvironments(environments); err != nil {
			// Handle error
			return
		}

		// Clear form
		form.GetFormItem(0).(*tview.InputField).SetText("")
		form.GetFormItem(1).(*tview.DropDown).SetCurrentOption(0)

		// Notify parent
		if onEnvironmentCreated != nil {
			onEnvironmentCreated()
		}
	})

	form.SetBorder(true).SetTitle(" Create Environment ")
	return form
}

// createEnvironmentPanel creates the environment panel with dropdown and config button
func createEnvironmentPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	selectionBackgroundColor,
	activeTabColor,
	buttonSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) (*tview.Flex, *tview.DropDown, *tview.Button, *CustomButton) {
	// Create environment dropdown
	envDropdown := createDropDown(
		"",
		[]string{"No Environment", "Development", "Staging", "Production"},
		backgroundColor,
		backgroundColor,
		backgroundColor,
		foregroundColor,
		activeTabColor,
		backgroundColor,
		selectionBackgroundColor,
		dropdownFocusedBackgroundColor,
	)
	envDropdown.SetBorder(false)
	envDropdown.SetCurrentOption(0)

	// Create config button
	configButton := createButton("⚙", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)

	separator := tview.NewBox().
		SetBackgroundColor(backgroundColor)

	indicator := createCustomButton("▼", backgroundColor, backgroundColor, borderFocusColor, borderFocusColor)

	// Create container
	container := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(separator, 1, 0, false).
			AddItem(envDropdown, 14, 0, false).
			AddItem(indicator, 2, 0, true), 17, 0, false).
		AddItem(separator, 0, 1, false).
		AddItem(configButton, 3, 0, true)

	container.SetBorder(true)
	container.SetTitle(" Environment ")
	container.SetBackgroundColor(backgroundColor)
	container.SetBorderColor(borderColor)
	container.SetTitleColor(titleColor)

	return container, envDropdown, configButton, indicator
}

// setupUIComponents creates and configures all UI components
func setupUIComponents(
	backgroundColor,
	foregroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	selectionBackgroundColor,
	activeTabColor,
	buttonSelectedColor,
	dropdownFocusedBackgroundColor tcell.Color,
) (
	*tview.TextView, *tview.TreeNode, *tview.Flex, *tview.DropDown, *tview.InputField, *tview.Button, *tview.TextView, *tview.TextArea, *tview.Flex, *tview.TextView, *tview.TreeView, *tview.Pages, *tview.Flex, *tview.Flex, *tview.TextView, *tview.TextView, *tview.TextView, *tview.TextView, *tview.Flex, *tview.DropDown, *tview.Button, *CustomButton,
) {
	// Create header panel
	header := createPanel(" Petitorium ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create root node for tree
	rootNode := tview.NewTreeNode("").SetSelectable(false)

	// Create unified method+URL+Send bar
	methodURLBar, methodDropdown, urlInput, sendButton := createMethodURLBar(
		"",
		backgroundColor,
		borderColor,
		titleColor,
		foregroundColor,
		selectionBackgroundColor,
		activeTabColor,
		buttonSelectedColor,
		dropdownFocusedBackgroundColor,
	)

	// Create both view and edit panels for body
	bodyViewPanel := createPanel(" Body (VIEW) ", backgroundColor, backgroundColor, titleColor, foregroundColor)
	bodyEditPanel := createTextArea(" Body (EDIT) ", backgroundColor, borderColor, titleColor, foregroundColor)

	// Create response tabs panel
	response, responsePages, responseTabHeader, _, responseInfoBar, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel := createResponseTabs(
		backgroundColor, borderColor, borderFocusColor, titleColor, foregroundColor, activeTabColor, selectionBackgroundColor,
		nil, nil, // No initial response or last request time
	)

	// Create footer panel
	footer := createPanel("", backgroundColor, borderColor, titleColor, foregroundColor)
	footer.SetText(" (Tab) Cycle Focus | Body Tab: (i) Insert (Esc) Normal (hjkl) Nav | (F4) External Editor | (q) Quit | Collections: (n) New Collection | (r) New Request | (R) Rename | (m) Move Item | (d) Delete")

	// Create environment panel with dropdown and config button
	environmentPanel, envDropdown, envConfigButton, envIndicatorButton := createEnvironmentPanel(
		backgroundColor,
		borderColor,
		borderFocusColor,
		titleColor,
		foregroundColor,
		selectionBackgroundColor,
		activeTabColor,
		buttonSelectedColor,
		dropdownFocusedBackgroundColor,
	)

	// Create collections tree view
	collectionsTreeView := tview.NewTreeView().
		SetRoot(rootNode).
		SetCurrentNode(rootNode)

	collectionsTreeView.
		SetGraphics(false).
		SetTopLevel(0)

	collectionsTreeView.
		SetBorder(true).
		SetTitle(" Collections ").
		SetBackgroundColor(backgroundColor).
		SetBorderColor(borderColor).
		SetTitleColor(titleColor).
		SetBorderPadding(0, 0, 0, 0)

	return header, rootNode, methodURLBar, methodDropdown, urlInput, sendButton, bodyViewPanel, bodyEditPanel, response, footer, collectionsTreeView, responsePages, responseTabHeader, responseInfoBar, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel, environmentPanel, envDropdown, envConfigButton, envIndicatorButton
}

// setupRequestPanel creates the unified request panel
func setupRequestPanel(methodURLBar, requestDataTabs *tview.Flex, backgroundColor tcell.Color) *tview.Flex {
	requestPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(methodURLBar, 3, 0, false).
		AddItem(requestDataTabs, 0, 1, false)
	requestPanel.SetBorder(false)
	requestPanel.SetBackgroundColor(backgroundColor)

	return requestPanel
}

// setupRightSide creates the right side layout
func setupRightSide(requestPanel *tview.Flex, response *tview.Flex) *tview.Flex {
	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(requestPanel, 0, 1, false).
		AddItem(response, 0, 1, false)

	return rightSide
}

// setupLayout creates the main grid layout
func setupLayout(header, footer *tview.TextView, collectionsTreeView *tview.TreeView, environmentPanel *tview.Flex, rightSide *tview.Flex) *tview.Grid {
	leftSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(environmentPanel, 3, 0, false).
		AddItem(collectionsTreeView, 0, 1, false)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(leftSide, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	return grid
}

// setupPanels creates the panels slice for cycling
func setupPanels(environmentPanel *tview.Flex, collectionsTreeView *tview.TreeView, methodURLBar *tview.Flex, requestDataTabs *tview.Flex, response *tview.TextView) []tview.Primitive {
	return []tview.Primitive{environmentPanel, collectionsTreeView, methodURLBar, requestDataTabs, response}
}

// setupCycles initializes the navigation cycles
func setupCycles(panels []tview.Primitive) {
	mainCycle = &MainCycle{
		panels:  panels,
		current: 1, // start with collections
	}

	headersCycle = &HeadersCycle{
		inputs:  []tview.Primitive{},
		current: 0,
		parent:  mainCycle,
	}
}

// syncBodyContent syncs body content between view and edit panels
func syncBodyContent(content string, bodyEditMode bool, bodyEditPanel *tview.TextArea, bodyViewPanel *tview.TextView) {
	if bodyEditMode {
		bodyEditPanel.SetText(content, false)
	} else {
		bodyViewPanel.Clear()
		if content != "" {
			// Format content with syntax highlighting
			formattedContent := formatBodyContent(content)

			// Set content with proper handling
			bodyViewPanel.SetText(formattedContent)
			bodyViewPanel.SetTextAlign(tview.AlignLeft)

			// Force proper rendering and scrolling
			go func() {
				// Small delay to ensure content is set before scrolling
				bodyViewPanel.ScrollToEnd()
				bodyViewPanel.ScrollToBeginning()
			}()
		} else {
			bodyViewPanel.SetText("")
			bodyViewPanel.SetTextAlign(tview.AlignLeft)
		}
	}
}

// switchBodyMode switches between view and edit modes for body content
func switchBodyMode(
	bodyEditMode *bool,
	bodyContainer *tview.Flex,
	bodyViewPanel *tview.TextView,
	bodyEditPanel *tview.TextArea,
	currentBodyContent string,
	currentRequest **workspace.Request,
	currentSelectedNode **tview.TreeNode,
	saveCurrentRequest func(),
) {
	*bodyEditMode = !*bodyEditMode
	bodyContainer.Clear()

	if *bodyEditMode {
		// Switch to edit mode
		bodyContainer.AddItem(bodyEditPanel, 0, 1, false)
		bodyEditPanel.SetText(currentBodyContent, false)
	} else {
		// Switch to view mode
		bodyContainer.AddItem(bodyViewPanel, 0, 1, false)
		// Update body content from edit panel if we were editing
		if *currentRequest != nil {
			currentBodyContent = bodyEditPanel.GetText()
			(*currentRequest).Body = currentBodyContent
			if *currentSelectedNode != nil {
				(*currentSelectedNode).SetReference(**currentRequest)
				saveCurrentRequest()
			}
		}
		syncBodyContent(currentBodyContent, *bodyEditMode, bodyEditPanel, bodyViewPanel)
	}
}
