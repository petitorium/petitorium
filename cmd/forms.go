package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

func createCollectionFormWithLocation(ui *UIOrchestrator) *tview.Form {
	app := ui.App
	pages := ui.Pages
	workspaceData := ui.WorkspaceData
	rootNode := ui.RootNode
	collectionsTreeView := ui.CollectionsTreeView
	colors := ui.Colors

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Build the location dropdown with the full nested path as display text and
	// capture the target collection IDs in parallel. Resolution on Save is by
	// ID (not by splitting the display path), so a collection placed at
	// "A → B → C" lands inside C at the correct depth.
	var locationOptions []string
	var targetCollectionIDs []string
	locationOptions = append(locationOptions, "(Root Level)")

	var addCollectionsToOptions func(collections *[]workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections *[]workspace.Collection, prefix string) {
		for i := range *collections {
			col := &(*collections)[i]
			locationOptions = append(locationOptions, prefix+col.Name)
			targetCollectionIDs = append(targetCollectionIDs, col.ID)
			if len(col.Collections) > 0 {
				addCollectionsToOptions(&col.Collections, prefix+col.Name+" → ")
			}
		}
	}
	addCollectionsToOptions(&workspaceData.Collections, "")

	// Pre-fill the Location dropdown to the currently-selected collection (or
	// the parent of a selected request), falling back to Root Level.
	defaultLocationIdx := 0
	if node := collectionsTreeView.GetCurrentNode(); node != nil {
		var selectedColID string
		if col := ui.collectionFromNode(node); col != nil {
			selectedColID = col.ID
		} else if req := ui.requestFromNode(node); req != nil {
			if parent := ui.DataManager.FindParentCollectionOfRequest(req.ID); parent != nil {
				selectedColID = parent.ID
			}
		}
		for i, id := range targetCollectionIDs {
			if id == selectedColID {
				defaultLocationIdx = i + 1 // +1 because dropdown index 0 is Root
				break
			}
		}
	}

	form.AddInputField("Name: ", "", 0, nil, nil).SetFieldBackgroundColor(colors.Border)
	// The title follows the selected Location: Root creates a "Collection",
	// any nested target creates a "Folder".
	form.AddDropDown("Location:", locationOptions, defaultLocationIdx, func(_ string, optionIndex int) {
		if optionIndex == 0 {
			form.SetTitle(" New Collection ")
		} else {
			form.SetTitle(" New Folder ")
		}
	})

	cancelFunc := func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		selectedIdx, _ := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		if strings.TrimSpace(name) == "" {
			return
		}

		newCollection := workspace.Collection{ID: workspace.NewID(), Name: name}

		if selectedIdx == 0 || selectedIdx-1 >= len(targetCollectionIDs) {
			// Add to root level
			workspaceData.Collections = append(workspaceData.Collections, newCollection)
		} else {
			targetID := targetCollectionIDs[selectedIdx-1]
			target := workspace.FindCollectionByID(&workspaceData.Collections, targetID)
			if target != nil {
				target.Collections = append(target.Collections, newCollection)
			} else {
				debugLog("new collection: target id %q not found, appending to root", targetID)
				workspaceData.Collections = append(workspaceData.Collections, newCollection)
			}
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("new collection: failed to save workspace: %v", err)
		}

		cancelFunc()
	}).SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))

	form.AddButton("Cancel", func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true)
	if defaultLocationIdx == 0 {
		form.SetTitle(" New Collection ")
	} else {
		form.SetTitle(" New Folder ")
	}
	return form
}

func createRequestForm(app *tview.Application,
	pages *tview.Pages,
	selectedCollection *workspace.Collection,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	colors *ColorManager,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddInputField("Name: ", "", 0, nil, nil).SetFieldBackgroundColor(colors.Border)
	form.AddDropDown("Method: ", workspace.HTTPMethods, 0, nil)
	form.AddInputField("URL: ", "", 0, nil, nil)

	bodyInput := tview.NewInputField().
		SetLabel("Body: ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(colors.Border).
		SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.LabelColor)).
		SetPlaceholder("Enter JSON data or edit later...")

	// Multipart fields management interface
	multipartContainer := tview.NewFlex().SetDirection(tview.FlexRow)
	multipartContainer.SetBorder(true).SetTitle("Multipart Fields")
	multipartFieldsList := tview.NewFlex().SetDirection(tview.FlexRow)
	multipartContainer.AddItem(multipartFieldsList, 0, 1, true)

	// Add field button
	addFieldButton := tview.NewButton("Add Field").SetSelectedFunc(func() {
		addMultipartField(multipartFieldsList, colors)
	})
	multipartContainer.AddItem(addFieldButton, 1, 0, false)

	// Initially hide multipart container (will be handled by form management)

	contentTypeDropdown := tview.NewDropDown().
		SetLabel("Content Type: ").
		SetOptions([]string{"JSON", "Multipart", "No Body"}, nil).
		SetCurrentOption(0)
	contentTypeDropdown.SetSelectedFunc(func(text string, index int) {
		switch text {
		case "JSON":
			bodyInput.SetPlaceholder("Enter JSON data or edit later.")
		case "Multipart":
			bodyInput.SetPlaceholder("You can edit later.")
		case "No Body":
			bodyInput.SetPlaceholder("(No body for this request)")
		}
	})

	form.AddFormItem(contentTypeDropdown)
	form.AddFormItem(bodyInput)

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		_, method := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		url := form.GetFormItem(2).(*tview.InputField).GetText()
		_, contentType := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
		body := form.GetFormItem(4).(*tview.InputField).GetText()

		if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
			return
		}

		// For multipart, the body field contains the formatted field data

		newRequest := workspace.Request{
			ID:          workspace.NewID(),
			Name:        name,
			Method:      method,
			URL:         url,
			ContentType: contentType,
			Body:        body,
		}

		if selectedCollection == nil {
			// Add to first collection if no collection selected
			if len(workspaceData.Collections) > 0 {
				workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, newRequest)
			} else {
				// Create a default collection
				defaultCollection := workspace.Collection{
					ID:       workspace.NewID(),
					Name:     "Requests",
					Requests: []workspace.Request{newRequest},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else {
			// selectedCollection is already a live pointer into workspaceData.
			selectedCollection.Requests = append(selectedCollection.Requests, newRequest)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("new request: failed to save workspace: %v", err)
		}

		pages.RemovePage("newRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("newRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("newRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.SetCancelFunc(cancelFunc)

	// Add F4 support for external editor on Body field
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF4 {
			// Check if we're on the Body field (index 3)
			formItemIndex, _ := form.GetFocusedItemIndex()
			if formItemIndex == 3 { // Body field is at index 3
				bodyField := form.GetFormItem(3).(*tview.InputField)
				currentBody := bodyField.GetText()

				// Suspend the app to open external editor
				app.Suspend(func() {
					modifiedContent, err := openInExternalEditor(currentBody, "json")
					if err != nil {
						// Could show error but for now just continue
						return
					}

					// Update the body field with the edited content
					bodyField.SetText(modifiedContent)
				})

				return nil
			}
		}
		return event
	})

	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))
	form.SetBorder(true).SetTitle(" New Request ")

	return form
}

func createRenameCollectionForm(app *tview.Application,
	pages *tview.Pages,
	selectedCollection *workspace.Collection,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	node *tview.TreeNode,
	colors *ColorManager,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddInputField("Name: ", selectedCollection.Name, 0, nil, nil)
	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		// Update the collection name in-place via the live pointer (works for
		// nested collections too, unlike the previous name-matching loop).
		selectedCollection.Name = newName

		// Update tree node (keep Collection icon at root, Folder icon when nested)
		closedIcon, expandedIcon := config.C.UI.CollectionIcon, config.C.UI.CollectionExpandedIcon
		if findParentNode(rootNode, node) != rootNode {
			closedIcon, expandedIcon = config.C.UI.FolderIcon, config.C.UI.FolderExpandedIcon
		}
		expanded := node.IsExpanded()
		if expanded {
			node.SetText(fmt.Sprintf("%s %s", expandedIcon, newName))
		} else {
			node.SetText(fmt.Sprintf("%s %s", closedIcon, newName))
		}

		// Refresh the node's cached display name (the ID is unchanged).
		node.SetReference(NodeRef{Kind: KindCollection, ID: selectedCollection.ID, Name: newName})

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("rename collection: failed to save workspace: %v", err)
		}

		pages.RemovePage("renameCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})
	cancelFunc := func() {
		pages.RemovePage("renameCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true)
	if isRootCollectionByID(workspaceData, selectedCollection.ID) {
		form.SetTitle(" Rename Collection ")
	} else {
		form.SetTitle(" Rename Folder ")
	}
	return form
}

func createRenameRequestForm(app *tview.Application,
	pages *tview.Pages,
	selectedRequest *workspace.Request,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	node *tview.TreeNode,
	colors *ColorManager,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddInputField("Name: ", selectedRequest.Name, 0, nil, nil).SetFieldBackgroundColor(colors.Border)

	cancelFunc := func() {
		pages.RemovePage("renameRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		// Update the request name in-place via the live pointer.
		selectedRequest.Name = newName

		// Update tree node
		coloredMethod := getColoredMethod(selectedRequest.Method)
		paddedName := padNameToMinLength(newName, 4)
		node.SetText(fmt.Sprintf("%s %s", coloredMethod, paddedName))

		// Refresh the node's cached display name (the ID is unchanged).
		node.SetReference(NodeRef{Kind: KindRequest, ID: selectedRequest.ID, Name: newName})

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("rename request: failed to save workspace: %v", err)
		}

		cancelFunc()
	}).SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))

	form.AddButton("Cancel", func() {
		cancelFunc()
	})

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Rename Request ")
	return form
}

func createDeleteEnvironmentConfirm(
	app *tview.Application,
	pages *tview.Pages,
	selectedEnvironment *workspace.Environment,
	environmentsData *[]workspace.Environment,
	workspaceData *workspace.Workspace,
	envDropdown *tview.DropDown,
	envConfigButton *CustomButton,
	colors *ColorManager,
	currentFocus tview.Primitive,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete the environment '%s'?", selectedEnvironment.Name), 0, 1, false, false)

	form.AddButton("Delete", func() {
		oldName := selectedEnvironment.Name
		// Find and remove the environment
		for i, e := range *environmentsData {
			if e.Name == oldName {
				*environmentsData = append((*environmentsData)[:i], (*environmentsData)[i+1:]...)
				break
			}
		}

		// Update selected environment if it was the one deleted
		if workspaceData.SelectedEnvironment == oldName {
			workspaceData.SelectedEnvironment = ""
		}

		// Save workspace data (which includes environments)
		workspaceData.Environments = *environmentsData
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		// Refresh the environment dropdown
		updateEnvironmentDropdown(envDropdown, *environmentsData)

		pages.RemovePage("deleteEnvironment")
		// Refresh the envVariables modal if it's open
		if pages.HasPage("envVariables") {
			// Since the modal is still open, we need to refresh it
			// But for simplicity, just close it or refresh
			// Actually, since we removed the env, the modal might need to be updated
			// But to keep it simple, perhaps remove the page and the user can reopen
			pages.RemovePage("envVariables")
			app.SetFocus(envConfigButton)
		}
	})

	cancelFunc := func() {
		pages.RemovePage("deleteEnvironment")
		app.SetFocus(currentFocus)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Delete Environment ")
	return form
}

func createRenameEnvironmentForm(
	app *tview.Application,
	pages *tview.Pages,
	selectedEnvironment *workspace.Environment,
	environmentsData *[]workspace.Environment,
	workspaceData *workspace.Workspace,
	envDropdown *tview.DropDown,
	envConfigButton *CustomButton,
	colors *ColorManager,
	currentFocus tview.Primitive,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddInputField("Name", selectedEnvironment.Name, 15, nil, nil)
	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		oldName := selectedEnvironment.Name
		// Find and update the actual environment in environmentsData
		for i := range *environmentsData {
			if (*environmentsData)[i].Name == oldName {
				(*environmentsData)[i].Name = newName
				break
			}
		}

		// Update selected environment if it was the one renamed
		if workspaceData.SelectedEnvironment == oldName {
			workspaceData.SelectedEnvironment = newName
		}

		// Update dropdown
		updateEnvironmentDropdown(envDropdown, *environmentsData)

		// Save workspace data (which includes environments)
		workspaceData.Environments = *environmentsData
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		pages.RemovePage("renameEnvironment")
		// Since the modal is still open, refresh it
		if pages.HasPage("envVariables") {
			pages.RemovePage("envVariables")
			app.SetFocus(envConfigButton)
		}
	})
	cancelFunc := func() {
		pages.RemovePage("renameEnvironment")
		app.SetFocus(currentFocus)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Rename Environment ")
	return form
}

func createCloneEnvironmentForm(
	app *tview.Application,
	pages *tview.Pages,
	selectedEnvironment *workspace.Environment,
	environmentsData *[]workspace.Environment,
	workspaceData *workspace.Workspace,
	envDropdown *tview.DropDown,
	envConfigButton *CustomButton,
	colors *ColorManager,
	currentFocus tview.Primitive,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddInputField("Name", selectedEnvironment.Name+" Copy", 15, nil, nil)
	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		// Check if name already exists
		for _, env := range *environmentsData {
			if env.Name == newName {
				return // Name already exists
			}
		}

		// Create new environment by cloning the selected one
		newEnv := workspace.Environment{
			Name:      newName,
			Base:      selectedEnvironment.Base,
			Variables: make(map[string]string),
		}
		// Copy variables
		for k, v := range selectedEnvironment.Variables {
			newEnv.Variables[k] = v
		}

		// Add to environments
		*environmentsData = append(*environmentsData, newEnv)

		// Update dropdown
		updateEnvironmentDropdown(envDropdown, *environmentsData)

		// Save workspace data (which includes environments)
		workspaceData.Environments = *environmentsData
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		pages.RemovePage("cloneEnvironment")
		// Since the modal is still open, refresh it
		if pages.HasPage("envVariables") {
			pages.RemovePage("envVariables")
			app.SetFocus(envConfigButton)
		}
	})
	form.AddButton("Cancel", func() {
		pages.RemovePage("cloneEnvironment")
		app.SetFocus(currentFocus)
	})

	form.SetBorder(true).SetTitle(" Clone Environment ")
	return form
}

func createDeleteCollectionConfirm(app *tview.Application, pages *tview.Pages, selectedCollection *workspace.Collection, workspaceData *workspace.Workspace, rootNode *tview.TreeNode, collectionsTreeView *tview.TreeView, node *tview.TreeNode, colors *ColorManager, dataManager *DataManager) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	isRoot := isRootCollectionByID(workspaceData, selectedCollection.ID)
	kind := "collection"
	if !isRoot {
		kind = "folder"
	}
	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete the %s '%s'?\nThis will also delete all nested folders and requests.", kind, selectedCollection.Name), 0, 2, false, false)

	form.AddButton("Delete", func() {
		workspace.RemoveCollectionByID(workspaceData, selectedCollection.ID)

		workspaceData.SelectedRequest = nil

		dataManager.UpdateWorkspaceData(workspaceData)

		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		collectionsTreeView.SetRoot(rootNode)

		if len(rootNode.GetChildren()) > 0 {
			collectionsTreeView.SetCurrentNode(rootNode.GetChildren()[0])
		}

		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("delete collection: failed to save workspace: %v", err)
		}

		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}).SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))

	form.AddButton("Cancel", func() {
		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true)
	if isRoot {
		form.SetTitle(" Delete Collection ")
	} else {
		form.SetTitle(" Delete Folder ")
	}
	return form
}

func createMoveCollectionForm(ui *UIOrchestrator, selectedCollection *workspace.Collection) *tview.Form {
	app := ui.App
	pages := ui.Pages
	workspaceData := ui.WorkspaceData
	rootNode := ui.RootNode
	collectionsTreeView := ui.CollectionsTreeView
	colors := ui.Colors
	dataManager := ui.DataManager

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Collect possible parent collections, including nested ones (excluding self
	// and the selected collection's descendants, which would create a cycle). We
	// capture IDs (for re-resolution after removal) and path-prefixed names (for
	// the dropdown display). Pointers are NOT captured because RemoveCollectionByID
	// shifts the Collections backing array (append(s[:i], s[i+1:]...)),
	// invalidating any pre-captured pointer.
	var possibleParentIDs []string
	var possibleParentNames []string
	var addPossibleParents func(collections *[]workspace.Collection, prefix string)
	addPossibleParents = func(collections *[]workspace.Collection, prefix string) {
		for i := range *collections {
			col := &(*collections)[i]
			// A collection is a valid parent unless it IS the selected collection
			// or a descendant of it (moving into a descendant creates a cycle).
			if col.ID != selectedCollection.ID && !workspace.IsDescendantCollection(selectedCollection, col.ID) {
				possibleParentIDs = append(possibleParentIDs, col.ID)
				possibleParentNames = append(possibleParentNames, prefix+col.Name)
			}
			// Recurse to find deeper valid parents, but skip the selected
			// collection's own subtree (its descendants are all excluded above).
			if col.ID != selectedCollection.ID && len(col.Collections) > 0 {
				addPossibleParents(&col.Collections, prefix+col.Name+" → ")
			}
		}
	}
	addPossibleParents(&workspaceData.Collections, "")

	// Create dropdown for selecting new parent
	parentOptions := make([]string, len(possibleParentNames)+1)
	parentOptions[0] = "Root"
	copy(parentOptions[1:], possibleParentNames)

	parentDropdown := tview.NewDropDown().
		SetLabel("Move to: ").
		SetOptions(parentOptions, nil).
		SetCurrentOption(0)
	parentDropdown.SetBackgroundColor(colors.Background)
	parentDropdown.SetFieldBackgroundColor(colors.Background)
	parentDropdown.SetFieldTextColor(colors.Foreground)
	parentDropdown.SetLabelColor(colors.Foreground)

	form.AddFormItem(parentDropdown)

	form.AddButton("Move", func() {
		selectedIndex, _ := parentDropdown.GetCurrentOption()

		// Snapshot the collection value BEFORE removal. selectedCollection is a
		// live pointer into the Collections backing array, which
		// RemoveCollectionByID shifts via append(s[:i], s[i+1:]...); reading
		// *selectedCollection after removal would return shifted memory.
		savedCol := *selectedCollection

		// Remove first (shifts the Collections backing array).
		workspace.RemoveCollectionByID(workspaceData, savedCol.ID)

		// Re-resolve the target parent by ID AFTER removal. A pointer captured
		// before removal would now address shifted memory.
		if selectedIndex > 0 && selectedIndex-1 < len(possibleParentIDs) {
			targetID := possibleParentIDs[selectedIndex-1]
			newParent := workspace.FindCollectionByID(&workspaceData.Collections, targetID)
			if newParent == nil {
				debugLog("move collection: target parent id %q not found after removal, appending to root", targetID)
				workspaceData.Collections = append(workspaceData.Collections, savedCol)
			} else {
				newParent.Collections = append(newParent.Collections, savedCol)
			}
		} else {
			workspaceData.Collections = append(workspaceData.Collections, savedCol)
		}

		dataManager.UpdateWorkspaceData(workspaceData)

		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("move collection: failed to save workspace: %v", err)
		}

		// Clear cached selection state: the tree is about to be rebuilt, so the
		// cached *Request / *tview.TreeNode would otherwise dangle.
		ui.clearSelectionState()

		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		collectionsTreeView.SetRoot(rootNode)

		if len(rootNode.GetChildren()) > 0 {
			collectionsTreeView.SetCurrentNode(rootNode.GetChildren()[0])
		}

		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	// Handle Esc key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true)
	if isRootCollectionByID(workspaceData, selectedCollection.ID) {
		form.SetTitle(" Move Collection ")
	} else {
		form.SetTitle(" Move Folder ")
	}
	return form
}

func createMoveRequestForm(ui *UIOrchestrator, selectedRequest *workspace.Request) *tview.Form {
	app := ui.App
	pages := ui.Pages
	workspaceData := ui.WorkspaceData
	rootNode := ui.RootNode
	collectionsTreeView := ui.CollectionsTreeView
	colors := ui.Colors
	dataManager := ui.DataManager

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Get all collections as possible targets, including nested. We capture IDs
	// (not pointers) so the target can be re-resolved by ID after the source
	// request is removed, avoiding any stale-pointer risk across mutation.
	var collectionOptions []string
	var targetCollectionIDs []string
	collectionOptions = append(collectionOptions, "Root")

	var addCollectionsToOptions func(collections *[]workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections *[]workspace.Collection, prefix string) {
		for i := range *collections {
			col := &(*collections)[i]
			collectionOptions = append(collectionOptions, prefix+col.Name)
			targetCollectionIDs = append(targetCollectionIDs, col.ID)
			if len(col.Collections) > 0 {
				addCollectionsToOptions(&col.Collections, prefix+col.Name+" → ")
			}
		}
	}
	addCollectionsToOptions(&workspaceData.Collections, "")

	collectionDropdown := tview.NewDropDown().
		SetLabel("Move to: ").
		SetOptions(collectionOptions, nil).
		SetCurrentOption(0)
	collectionDropdown.SetBackgroundColor(colors.Background)
	collectionDropdown.SetFieldBackgroundColor(colors.Background)
	collectionDropdown.SetFieldTextColor(colors.Foreground)
	collectionDropdown.SetLabelColor(colors.Foreground)

	form.AddFormItem(collectionDropdown)

	form.AddButton("Move", func() {
		selectedIndex, _ := collectionDropdown.GetCurrentOption()

		// selectedRequest is a live pointer (resolved by ID in moveItem).
		// Snapshot the value BEFORE removal: RemoveRequestByID shifts the
		// Requests backing array via append(s[:j], s[j+1:]...), so reading
		// *selectedRequest afterwards would return shifted/neighbouring memory,
		// which was the direct cause of lost and duplicated requests.
		savedReq := *selectedRequest

		if selectedIndex == 0 {
			workspace.RemoveRequestByID(workspaceData, savedReq.ID)
			if len(workspaceData.Collections) > 0 {
				workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, savedReq)
			} else {
				defaultCollection := workspace.Collection{
					ID:       workspace.NewID(),
					Name:     "Requests",
					Requests: []workspace.Request{savedReq},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else if selectedIndex > 0 && selectedIndex <= len(targetCollectionIDs) {
			workspace.RemoveRequestByID(workspaceData, savedReq.ID)

			// Re-resolve the target collection by ID after removal.
			targetCollection := workspace.FindCollectionByID(&workspaceData.Collections, targetCollectionIDs[selectedIndex-1])
			if targetCollection != nil {
				targetCollection.Requests = append(targetCollection.Requests, savedReq)
			} else {
				debugLog("move request: target collection id %q not found, appending to first collection", targetCollectionIDs[selectedIndex-1])
				if len(workspaceData.Collections) > 0 {
					workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, savedReq)
				}
			}
		}

		workspaceData.SelectedRequest = nil

		dataManager.UpdateWorkspaceData(workspaceData)

		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("move request: failed to save workspace: %v", err)
		}

		// Clear cached selection state: the tree is about to be rebuilt, so the
		// cached *Request / *tview.TreeNode would otherwise dangle.
		ui.clearSelectionState()

		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		collectionsTreeView.SetRoot(rootNode)

		if len(rootNode.GetChildren()) > 0 {
			collectionsTreeView.SetCurrentNode(rootNode.GetChildren()[0])
		}

		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}).SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))

	cancelFunc := func() {
		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	// Handle Esc key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Move Request ")
	return form
}

func createDeleteRequestConfirm(app *tview.Application,
	pages *tview.Pages,
	selectedRequest *workspace.Request,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	node *tview.TreeNode,
	colors *ColorManager,
	dataManager *DataManager,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete\nthe request '%s'?", selectedRequest.Name), 0, 2, false, false)

	form.AddButton("Delete", func() {
		// selectedRequest is a live pointer resolved by ID; remove by ID.
		workspace.RemoveRequestByID(workspaceData, selectedRequest.ID)

		workspaceData.SelectedRequest = nil

		dataManager.UpdateWorkspaceData(workspaceData)

		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		collectionsTreeView.SetRoot(rootNode)

		if len(rootNode.GetChildren()) > 0 {
			collectionsTreeView.SetCurrentNode(rootNode.GetChildren()[0])
		}

		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("delete request: failed to save workspace: %v", err)
		}

		pages.RemovePage("deleteRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("deleteRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))
	form.SetBorder(true).SetTitle(" Delete Request ")

	return form
}

func createWorkspaceManagementForm(
	app *tview.Application,
	pages *tview.Pages,
	workspaceData *workspace.Workspace,
	ui *UIOrchestrator,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	colors *ColorManager,
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Get current workspace info
	manager, err := workspace.LoadWorkspaceManager()
	currentWorkspace := "Default"
	if err == nil && manager.CurrentWorkspace != "" {
		currentWorkspace = manager.CurrentWorkspace
	}

	// Display current workspace
	form.AddTextView("Current Workspace", fmt.Sprintf(" %s ", currentWorkspace), 30, 1, true, false)

	// Workspace actions
	form.AddButton("Create New Workspace", func() {
		pages.RemovePage("workspaceMenu")
		createForm := createNewWorkspaceForm(app, pages, ui, rootNode, collectionsTreeView, colors)
		modal := createSizedModal(createForm, modalSizeForm, tcell.ColorDefault)
		pages.AddPage("createWorkspace", modal, true, true)
		app.SetFocus(createForm)
	})

	form.AddButton("Switch Workspace", func() {
		pages.RemovePage("workspaceMenu")
		showWorkspaceModal(&UIOrchestrator{App: app, Pages: pages, WorkspaceConfigButton: ui.WorkspaceConfigButton, Colors: colors, WorkspaceData: workspaceData, RootNode: rootNode, CollectionsTreeView: collectionsTreeView})
	})

	form.AddButton("Rename Current Workspace", func() {
		pages.RemovePage("workspaceMenu")
		renameForm := createRenameWorkspaceForm(app, pages, currentWorkspace, ui.WorkspaceConfigButton, colors)
		modal := createSizedModal(renameForm, modalSizeForm, tcell.ColorDefault)
		pages.AddPage("renameWorkspace", modal, true, true)
		app.SetFocus(renameForm)
	})

	form.AddButton("Duplicate Workspace", func() {
		pages.RemovePage("workspaceMenu")
		duplicateForm := createDuplicateWorkspaceForm(app, pages, currentWorkspace, ui.WorkspaceConfigButton, colors)
		modal := createSizedModal(duplicateForm, modalSizeForm, tcell.ColorDefault)
		pages.AddPage("duplicateWorkspace", modal, true, true)
		app.SetFocus(duplicateForm)
	})

	form.AddButton("Delete Workspace", func() {
		pages.RemovePage("workspaceMenu")
		deleteForm := createDeleteWorkspaceForm(app, pages, currentWorkspace, ui.WorkspaceConfigButton, colors)
		modal := createSizedModal(deleteForm, modalSizeConfirm, tcell.ColorDefault)
		pages.AddPage("deleteWorkspace", modal, true, true)
		app.SetFocus(deleteForm)
	})

	form.AddButton("Close", func() {
		pages.RemovePage("workspaceMenu")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetBorder(true).SetTitle(" Workspace Management ")
	return form
}

func createNewWorkspaceForm(
	app *tview.Application,
	pages *tview.Pages,
	ui *UIOrchestrator,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	colors *ColorManager,
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	nameInput := tview.NewInputField().
		SetLabel("Workspace Name: ").
		SetFieldWidth(30)

	form.AddFormItem(nameInput)

	form.AddButton("Create", func() {
		name := strings.TrimSpace(nameInput.GetText())
		if name == "" {
			return
		}

		_, err := workspace.CreateWorkspace(name)
		if err != nil {
			return
		}

		ui.WorkspaceNames, _ = workspace.ListWorkspaces()
		ui.WorkspaceSelector.SetOptions(ui.WorkspaceNames, nil)
		ui.WorkspaceSelector.SetSelectedFunc(func(text string, index int) {
			if text != "" {
				ui.SwitchWorkspace(text)
			}
		})

		pages.RemovePage("createWorkspace")
		pages.SwitchToPage("main")

		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("createWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetBorder(true).SetTitle(" Create New Workspace ")

	// Add input capture to handle 'q' to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			pages.RemovePage("createWorkspace")
			pages.SwitchToPage("main")
			app.SetFocus(collectionsTreeView)
			return nil
		}
		return event
	})

	return form
}

func createRenameWorkspaceForm(
	app *tview.Application,
	pages *tview.Pages,
	currentName string,
	workspaceSelector *CustomButton,
	colors *ColorManager,
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	nameInput := tview.NewInputField().
		SetLabel("New Name: ").
		SetText(currentName)

	form.AddFormItem(nameInput)

	form.AddButton("Rename", func() {
		newName := strings.TrimSpace(nameInput.GetText())
		if newName == "" || newName == currentName {
			return
		}

		err := workspace.RenameWorkspace(currentName, newName)
		if err != nil {
			// Show error - for now just ignore
			return
		}

		pages.RemovePage("renameWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	cancelFunc := func() {
		pages.RemovePage("renameWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Rename Workspace ")

	// Add input capture to handle 'q' to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' || event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	return form
}

func createDuplicateWorkspaceForm(
	app *tview.Application,
	pages *tview.Pages,
	sourceName string,
	workspaceSelector *CustomButton,
	colors *ColorManager,
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	nameInput := tview.NewInputField().
		SetLabel("NeX Workspace Name: ").
		SetFieldWidth(25)

	form.AddFormItem(nameInput)

	form.AddButton("Duplicate", func() {
		targetName := strings.TrimSpace(nameInput.GetText())
		if targetName == "" {
			return
		}

		_, err := workspace.DuplicateWorkspace(sourceName, targetName)
		if err != nil {
			// Show error - for now just ignore
			return
		}

		pages.RemovePage("duplicateWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("duplicateWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.SetBorder(true).SetTitle(" Duplicate Workspace ")
	return form
}

func createDeleteWorkspaceForm(
	app *tview.Application,
	pages *tview.Pages,
	workspaceName string,
	workspaceSelector *CustomButton,
	colors *ColorManager,
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	message := fmt.Sprintf("Are you sure you want to delete '%s'?", workspaceName) + "\n" + "This action cannot be undone."
	form.AddTextView("", message, 40, 3, true, false)

	form.AddButton("Delete", func() {
		err := workspace.DeleteWorkspace(workspaceName)
		if err != nil {
			// Show error - for now just ignore
			return
		}

		pages.RemovePage("deleteWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("deleteWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.SetBorder(true).SetTitle(" Delete Workspace ")
	return form
}

// createDuplicateRequestForm creates a form for duplicating an existing request
func createDuplicateRequestForm(
	app *tview.Application,
	pages *tview.Pages,
	originalRequest *workspace.Request,
	selectedCollection *workspace.Collection,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	colors *ColorManager,
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetFieldBackgroundColor(colors.Background)
	form.SetFieldTextColor(colors.Foreground)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Pre-populate form with original request data, adding "(Copy)" to name
	duplicatedName := originalRequest.Name + " (Copy)"

	// Find method index for dropdown
	methodIndex := 0
	for i, method := range workspace.HTTPMethods {
		if method == originalRequest.Method {
			methodIndex = i
			break
		}
	}

	form.AddInputField("Name: ", duplicatedName, 41, nil, nil)
	form.AddDropDown("Method: ", workspace.HTTPMethods, methodIndex, nil)
	form.AddInputField("URL: ", originalRequest.URL, 41, nil, nil)

	bodyInput := tview.NewInputField().
		SetLabel("Body: ").
		SetFieldWidth(41).
		SetText(originalRequest.Body).
		SetPlaceholderStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.LabelColor)).
		SetPlaceholder("Enter JSON data...")

	contentTypeDropdown := tview.NewDropDown().
		SetLabel("Content Type: ").
		SetOptions([]string{"JSON", "Multipart", "No Body"}, nil).
		SetCurrentOption(0)
	contentTypeDropdown.SetSelectedFunc(func(text string, index int) {
		switch text {
		case "JSON":
			bodyInput.SetPlaceholder("Enter JSON data...")
		case "Multipart":
			bodyInput.SetPlaceholder("Multipart UI coming soon. For now: name1=value1&name2=file:/path/to/file")
		case "No Body":
			bodyInput.SetPlaceholder("(No body for this request)")
		}
	})

	form.AddFormItem(contentTypeDropdown)
	form.AddFormItem(bodyInput)

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		_, method := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		url := form.GetFormItem(2).(*tview.InputField).GetText()
		_, contentType := form.GetFormItem(3).(*tview.DropDown).GetCurrentOption()
		body := form.GetFormItem(4).(*tview.InputField).GetText()

		if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
			return
		}

		// Create new request with duplicated data (excluding response history)
		newRequest := workspace.Request{
			ID:          workspace.NewID(),
			Name:        name,
			Method:      method,
			URL:         url,
			ContentType: contentType,
			Body:        body,
			Headers:     make(map[string]workspace.Entry),
		}

		for key, value := range originalRequest.Headers {
			newRequest.Headers[key] = workspace.Entry{Value: value.Value, Enabled: value.Enabled}
		}

		if selectedCollection == nil {
			// Add to first collection if no collection selected
			if len(workspaceData.Collections) > 0 {
				workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, newRequest)
			} else {
				// Create a default collection
				defaultCollection := workspace.Collection{
					ID:       workspace.NewID(),
					Name:     "Requests",
					Requests: []workspace.Request{newRequest},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else {
			// Find and update the actual collection in workspaceData
			// selectedCollection is already a live pointer into workspaceData.
			selectedCollection.Requests = append(selectedCollection.Requests, newRequest)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			debugLog("duplicate request: failed to save workspace: %v", err)
		}

		pages.RemovePage("duplicateRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})
	form.AddButton("Cancel", func() {
		pages.RemovePage("duplicateRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("duplicateRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.SetCancelFunc(cancelFunc)

	// Add F4 support for external editor on Body field
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF4 {
			// Check if we're on the Body field (index 3)
			formItemIndex, _ := form.GetFocusedItemIndex()
			if formItemIndex == 3 { // Body field is at index 3
				bodyField := form.GetFormItem(3).(*tview.InputField)
				currentBody := bodyField.GetText()

				// Suspend the app to open external editor
				app.Suspend(func() {
					modifiedContent, err := openInExternalEditor(currentBody, "json")
					if err != nil {
						// Could show error but for now just continue
						return
					}

					// Update the body field with the edited content
					bodyField.SetText(modifiedContent)
				})

				return nil
			}
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Duplicate Request ")
	form.SetFieldBackgroundColor(colors.Border)
	form.SetButtonActivatedStyle(tcell.StyleDefault.Background(colors.Border).Foreground(colors.Foreground))

	return form
}

func createDeleteAllHeadersConfirm(
	app *tview.Application,
	pages *tview.Pages,
	colors *ColorManager,
	deleteCallback func(),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", "Are you sure you want to delete all headers?", 0, 1, false, false)

	form.AddButton("Delete", func() {
		deleteCallback()
		pages.RemovePage("deleteAllHeaders")
	})

	cancelFunc := func() {
		pages.RemovePage("deleteAllHeaders")
	}

	form.AddButton("Cancel", cancelFunc)

	// Handle Esc key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Delete All Headers ")
	return form
}

func createDeleteCookieConfirm(
	app *tview.Application,
	pages *tview.Pages,
	colors *ColorManager,
	cookieName string,
	deleteCallback func(),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", "Delete cookie '"+cookieName+"'?", 0, 1, false, false)

	form.AddButton("Delete", func() {
		deleteCallback()
		pages.RemovePage("deleteCookie")
	})

	cancelFunc := func() {
		pages.RemovePage("deleteCookie")
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Delete Cookie ")
	return form
}

func createDeleteAllCookiesConfirm(
	app *tview.Application,
	pages *tview.Pages,
	colors *ColorManager,
	deleteCallback func(),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", "Are you sure you want to clear all cookies?", 0, 1, false, false)

	form.AddButton("Delete", func() {
		deleteCallback()
		pages.RemovePage("clearAllCookies")
	})

	cancelFunc := func() {
		pages.RemovePage("clearAllCookies")
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Clear All Cookies ")
	return form
}

func createDeleteAllQueryParamsConfirm(
	app *tview.Application,
	pages *tview.Pages,
	colors *ColorManager,
	deleteCallback func(),
) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", "Are you sure you want to delete all query parameters?", 0, 1, false, false)

	form.AddButton("Delete", func() {
		deleteCallback()
		pages.RemovePage("deleteAllQueryParams")
	})

	cancelFunc := func() {
		pages.RemovePage("deleteAllQueryParams")
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Delete All Query Parameters ")
	return form
}

// collectMultipartFields collects all multipart fields from the UI and formats them as a string
func collectMultipartFields(fieldsList *tview.Flex) string {
	var fields []string
	for i := 0; i < fieldsList.GetItemCount(); i++ {
		fieldContainer := fieldsList.GetItem(i).(*tview.Flex)
		if fieldContainer.GetItemCount() >= 4 {
			// Extract field values from the controls
			fieldControls := fieldContainer.GetItem(0).(*tview.Flex)
			nameInput := fieldControls.GetItem(0).(*tview.InputField)
			typeDropdown := fieldControls.GetItem(1).(*tview.DropDown)
			valueInput := fieldControls.GetItem(2).(*tview.InputField)

			name := nameInput.GetText()
			_, fieldType := typeDropdown.GetCurrentOption()
			value := valueInput.GetText()

			if name != "" && value != "" {
				if fieldType == "file" {
					fields = append(fields, name+"=file:"+value)
				} else {
					fields = append(fields, name+"="+value)
				}
			}
		}
	}
	return strings.Join(fields, "&")
}

// openMultipartFieldsModal opens a modal for managing multipart fields
func openMultipartFieldsModal(app *tview.Application, pages *tview.Pages, bodyInput *tview.InputField) {
	modal := tview.NewModal().
		SetText("Multipart Fields Management\n\nUse the body field with format:\nname1=value1&name2=file:/path/to/file\n\nClick OK to continue editing the body field.").
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			pages.RemovePage("multipartModal")
			app.SetFocus(bodyInput)
		})

	pages.AddPage("multipartModal", modal, true, true)
}

// createDeleteMultipartFieldConfirm creates a confirmation dialog for deleting a single
// multipart field. On confirm or cancel the modal page is removed and the previously-focused
// primitive (captured before opening the modal) is restored per the AGENTS.md focus rule.
func createDeleteMultipartFieldConfirm(app *tview.Application, pages *tview.Pages, colors *ColorManager, deleteCallback func(), previousFocus tview.Primitive) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", "Are you sure you want to delete this multipart field?", 0, 3, false, false)

	closeModalFunc := func() {
		pages.RemovePage("deleteMultipartField")
		if previousFocus != nil {
			app.SetFocus(previousFocus)
		}
	}

	form.AddButton("Delete", func() {
		deleteCallback()
		pages.RemovePage("deleteMultipartField")
		if previousFocus != nil {
			app.SetFocus(previousFocus)
		}
	})

	form.AddButton("Cancel", closeModalFunc)
	form.SetCancelFunc(closeModalFunc)

	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			closeModalFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Delete Multipart Field ")
	return form
}

// addMultipartField adds a new multipart field to the fields list
func addMultipartField(fieldsList *tview.Flex, colors *ColorManager) {
	fieldContainer := tview.NewFlex().SetDirection(tview.FlexRow)

	// Field controls in a horizontal layout
	fieldControls := tview.NewFlex().SetDirection(tview.FlexColumn)

	// Field name input
	nameInput := tview.NewInputField().
		SetLabel("Name: ").
		SetFieldWidth(8)
	fieldControls.AddItem(nameInput, 5, 0, false)

	// Field type dropdown
	typeDropdown := tview.NewDropDown().
		SetLabel("Type: ").
		SetOptions([]string{"text", "text_multiline", "file"}, nil).
		SetCurrentOption(0)
	fieldControls.AddItem(typeDropdown, 12, 0, false)

	// Value input (changes based on type)
	valueInput := tview.NewInputField().
		SetLabel("Value: ").
		SetFieldWidth(8)
	fieldControls.AddItem(valueInput, 5, 0, false)

	// Remove button
	// removeButton := tview.NewButton("Remove").SetSelectedFunc(func() {
	// 	// Find and remove this field container from the parent
	// 	for i := 0; i < fieldsList.GetItemCount(); i++ {
	// 		if fieldsList.GetItem(i) == fieldContainer {
	// 			fieldsList.RemoveItem(fieldsList.GetItem(i))
	// 			break
	// 		}
	// 	}
	// })

	// fieldControls.AddItem(removeButton, 10, 0, false)

	fieldContainer.AddItem(fieldControls, 1, 0, false)
	fieldsList.AddItem(fieldContainer, 1, 0, false)

	// Update value input based on type selection
	typeDropdown.SetSelectedFunc(func(text string, index int) {
		switch text {
		case "text":
			valueInput.SetLabel("Value: ")
			valueInput.SetPlaceholder("Enter text value")
		case "text_multiline":
			valueInput.SetLabel("Value: ")
			valueInput.SetPlaceholder("Enter multiline text")
		case "file":
			valueInput.SetLabel("File: ")
			valueInput.SetPlaceholder("Enter absolute file path")
		}
	})
}
