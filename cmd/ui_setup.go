package cmd

import (
	"github.com/rivo/tview"
)

// UIComponents holds all UI components for the application
type UIComponents struct {
	WorkspacePanel          *tview.Flex
	WorkspaceSelector       *tview.DropDown
	WorkspaceConfigButton   *CustomButton
	RootNode                *tview.TreeNode
	MethodURLBar            *tview.Flex
	MethodDropdown          *tview.DropDown
	URLInput                *URLVariableInput
	SendButton              *CustomButton
	CurlButton              *CustomButton
	BodyViewPanel           *tview.TextView
	BodyEditPanel           *tview.TextArea
	Response                *tview.Flex
	Footer                  *tview.Flex
	FooterLeft              *tview.TextView
	FooterRight             *tview.TextView
	CollectionsTreeView     *tview.TreeView
	CollectionSearchPanel   *tview.Flex
	CollectionSearchInput   *tview.InputField
	CollectionSearchResults *tview.List
	ResponsePages           *tview.Pages
	ResponseTabHeader       *tview.Flex
	ResponseInfoBar         *tview.Flex
	ResponseTimeText        *tview.TextView
	ResponsePreviewPanel    *tview.TextView
	ResponseHeadersPanel    tview.Primitive
	ResponseCookiesPanel    *tview.TextView
	ResponseTimelinePanel   *tview.TextView
	EnvironmentPanel        *tview.Flex
	EnvDropdown             *tview.DropDown
	EnvConfigButton         *CustomButton
	EnvIndicatorButton      *CustomButton
}

// setupUIComponents creates and configures all UI components
func setupUIComponents(colors *ColorManager, app *tview.Application) *UIComponents {
	// Create workspace panel with dropdown and config button
	workspacePanel, workspaceSelector, _, workspaceConfigButton := createWorkspacePanel(colors)

	// Create root node for tree
	rootNode := tview.NewTreeNode("").SetSelectable(false)

	// Create unified method+URL+Send bar
	methodURLBar, methodDropdown, urlInput, sendButton, curlButton := createMethodURLBar(
		"",
		colors,
		app,
	)

	// Create both view and edit panels for body
	bodyViewPanel := createPanel(" Body (VIEW) ", colors, &PanelOptions{HasBorder: &[]bool{true}[0], BorderColor: &colors.Background})
	bodyEditPanel := createTextArea(" Body (EDIT) ", colors.Background, colors.Border, colors.Title, colors.Foreground)

	// Create response tabs panel
	response, responsePages, responseTabHeader, _, responseInfoBar, responsePreviewPanel, responseHeadersPanel, responseCookiesPanel, responseTimelinePanel, responseTimeText := createResponseTabs(
		colors,
		nil, nil, // No initial response or last request time
		nil, // tabCallback will be set later
		nil, // copyCallback will be set later
	)

	activeBoder := false
	// Create footer panel
	footerLeft := createPanel("", colors, nil)
	footerLeft.SetBorder(activeBoder)

	footerRight := createPanel("", colors, nil)
	footerRight.SetBorder(activeBoder).
		SetTitleColor(colors.Title)

	footerFlex := tview.NewFlex().
		AddItem(footerLeft, 0, 1, false).
		AddItem(footerRight, 40, 0, false)

	footerFlex.SetBorder(true).
		SetBorderColor(colors.Border).
		SetBackgroundColor(colors.Background)

	footer := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(footerFlex, 0, 1, false)

	// Create environment panel with dropdown and config button
	environmentPanel, envDropdown, _, configButtonRaw := createEnvironmentPanel(
		colors,
	)
	// Direct assignment since we know the type
	envConfigButton := configButtonRaw
	envIndicatorButton := configButtonRaw // Using same button for both

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
		SetBackgroundColor(colors.Background).
		SetBorderColor(colors.Border).
		SetTitleColor(colors.Title).
		SetBorderPadding(0, 0, 0, 0)

	collectionSearchPanel, collectionSearchInput, collectionSearchResults := createCollectionSearchPanel(colors)

	return &UIComponents{
		WorkspacePanel:          workspacePanel,
		WorkspaceSelector:       workspaceSelector,
		WorkspaceConfigButton:   workspaceConfigButton,
		RootNode:                rootNode,
		MethodURLBar:            methodURLBar,
		MethodDropdown:          methodDropdown,
		URLInput:                urlInput,
		SendButton:              sendButton,
		CurlButton:              curlButton,
		BodyViewPanel:           bodyViewPanel,
		BodyEditPanel:           bodyEditPanel,
		Response:                response,
		Footer:                  footer,
		FooterLeft:              footerLeft,
		FooterRight:             footerRight,
		CollectionsTreeView:     collectionsTreeView,
		CollectionSearchPanel:   collectionSearchPanel,
		CollectionSearchInput:   collectionSearchInput,
		CollectionSearchResults: collectionSearchResults,
		ResponsePages:           responsePages,
		ResponseTabHeader:       responseTabHeader,
		ResponseInfoBar:         responseInfoBar,
		ResponseTimeText:        responseTimeText,
		ResponsePreviewPanel:    responsePreviewPanel,
		ResponseHeadersPanel:    responseHeadersPanel,
		ResponseCookiesPanel:    responseCookiesPanel,
		ResponseTimelinePanel:   responseTimelinePanel,
		EnvironmentPanel:        environmentPanel,
		EnvDropdown:             envDropdown,
		EnvConfigButton:         envConfigButton,
		EnvIndicatorButton:      envIndicatorButton,
	}
}

// setupRequestPanel creates the unified request panel
func setupRequestPanel(methodURLBar, requestDataTabs *tview.Flex, colors *ColorManager) *tview.Flex {
	requestPanel := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(methodURLBar, 3, 0, false).
		AddItem(requestDataTabs, 0, 1, false)
	requestPanel.SetBorder(false)
	requestPanel.SetBackgroundColor(colors.Background)

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
