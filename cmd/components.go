package cmd

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/workspace"
)

// UIComponents holds all UI components for the application
type UIComponents struct {
	Header                *tview.TextView
	RootNode              *tview.TreeNode
	MethodURLBar          *tview.Flex
	MethodDropdown        *tview.DropDown
	URLInput              *tview.InputField
	SendButton            *tview.Button
	BodyViewPanel         *tview.TextView
	BodyEditPanel         *tview.TextArea
	Response              *tview.Flex
	Footer                *tview.TextView
	CollectionsTreeView   *tview.TreeView
	ResponsePages         *tview.Pages
	ResponseTabHeader     *tview.Flex
	ResponseInfoBar       *tview.Flex
	ResponsePreviewPanel  *tview.TextView
	ResponseHeadersPanel  *tview.TextView
	ResponseCookiesPanel  *tview.TextView
	ResponseTimelinePanel *tview.TextView
	EnvironmentPanel      *tview.Flex
	EnvDropdown           *tview.DropDown
	EnvConfigButton       *tview.Button
	EnvIndicatorButton    *CustomButton
}

// createRenameEnvironmentPanel creates a simple panel for renaming environments
func createRenameEnvironmentPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor tcell.Color,
	currentName string,
	onRename func(string),
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

	form.AddInputField("New Name", currentName, 20, nil, nil)

	form.AddButton("Rename", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) != "" && newName != currentName {
			onRename(newName)
		}
	})

	form.AddButton("Cancel", func() {
		// Just close the form - handled by parent
	})

	form.SetBorder(true).SetTitle(" Rename Environment ")
	return form
}

// createEnvironmentListPanel creates a list panel for selecting and managing environments
func createEnvironmentListPanel(
	backgroundColor,
	borderColor,
	borderFocusColor,
	titleColor,
	foregroundColor,
	buttonSelectedColor tcell.Color,
	environments []workspace.Environment,
	onEnvironmentSelected func(*workspace.Environment),
	onCreateNew func(),
	onRenameEnvironment func(*workspace.Environment),
	onRemoveEnvironment func(*workspace.Environment),
) *tview.Flex {
	list := tview.NewFlex().SetDirection(tview.FlexRow)
	list.SetBackgroundColor(backgroundColor)
	list.SetBorderColor(borderColor)
	list.SetTitleColor(titleColor)

	// Add "Create New Environment" option at the top
	createNewButton := createButton("➕ Create New Environment", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
	createNewButton.SetSelectedFunc(onCreateNew)
	list.AddItem(createNewButton, 1, 0, false)

	// Add separator
	separator := tview.NewBox().SetBackgroundColor(backgroundColor)
	list.AddItem(separator, 1, 0, false)

	// Add all existing environments
	for i, env := range environments {
		env := env // Capture loop variable

		// Create a container for each environment with buttons
		envContainer := tview.NewFlex().SetDirection(tview.FlexColumn)

		// Create the main environment button
		var buttonText string
		if env.Name == "Base" {
			buttonText = env.Name + " (base environment)"
		} else {
			buttonText = env.Name
		}

		envButton := createButton(buttonText, backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
		envButton.SetSelectedFunc(func() {
			onEnvironmentSelected(&environments[i])
		})
		envContainer.AddItem(envButton, 0, 1, false)

		// Add rename and remove buttons for non-base environments
		if env.Name != "Base" {
			renameButton := createButton("✏", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
			renameButton.SetSelectedFunc(func() {
				onRenameEnvironment(&environments[i])
			})
			envContainer.AddItem(renameButton, 3, 0, false)

			removeButton := createButton("✕", backgroundColor, borderColor, titleColor, foregroundColor, buttonSelectedColor)
			removeButton.SetSelectedFunc(func() {
				onRemoveEnvironment(&environments[i])
			})
			envContainer.AddItem(removeButton, 3, 0, false)
		}

		list.AddItem(envContainer, 1, 0, false)
	}

	list.SetBorder(true).SetTitle(" Environments ")
	return list
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
		[]string{"Base Environment"},
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
) *UIComponents {
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

	return &UIComponents{
		Header:                header,
		RootNode:              rootNode,
		MethodURLBar:          methodURLBar,
		MethodDropdown:        methodDropdown,
		URLInput:              urlInput,
		SendButton:            sendButton,
		BodyViewPanel:         bodyViewPanel,
		BodyEditPanel:         bodyEditPanel,
		Response:              response,
		Footer:                footer,
		CollectionsTreeView:   collectionsTreeView,
		ResponsePages:         responsePages,
		ResponseTabHeader:     responseTabHeader,
		ResponseInfoBar:       responseInfoBar,
		ResponsePreviewPanel:  responsePreviewPanel,
		ResponseHeadersPanel:  responseHeadersPanel,
		ResponseCookiesPanel:  responseCookiesPanel,
		ResponseTimelinePanel: responseTimelinePanel,
		EnvironmentPanel:      environmentPanel,
		EnvDropdown:           envDropdown,
		EnvConfigButton:       envConfigButton,
		EnvIndicatorButton:    envIndicatorButton,
	}
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
