package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/workspace"
)

func createCollectionFormWithLocation(
	app *tview.Application,
	pages *tview.Pages,
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

	// Get all available collections for location targets
	var locationOptions []string
	locationOptions = append(locationOptions, "(Root Level)")

	var addCollectionsToOptions func(collections []workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections []workspace.Collection, prefix string) {
		for _, col := range collections {
			locationOptions = append(locationOptions, prefix+col.Name)
			if len(col.Collections) > 0 {
				addCollectionsToOptions(col.Collections, prefix+col.Name+" → ")
			}
		}
	}
	addCollectionsToOptions(workspaceData.Collections, "")

	form.AddInputField("Collection Name", "", 30, nil, nil)
	form.AddDropDown("Location", locationOptions, 0, nil)

	cancelFunc := func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		_, location := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		if strings.TrimSpace(name) == "" {
			return
		}

		newCollection := workspace.Collection{Name: name}

		if location == "(Root Level)" {
			// Add to root level
			workspaceData.Collections = append(workspaceData.Collections, newCollection)
		} else {
			// Find target collection and add to it
			targetName := strings.Split(location, " → ")[0]
			var addToCollection func(collections []workspace.Collection) bool
			addToCollection = func(collections []workspace.Collection) bool {
				for i := range collections {
					if collections[i].Name == targetName {
						collections[i].Collections = append(collections[i].Collections, newCollection)
						return true
					}
					if len(collections[i].Collections) > 0 {
						if addToCollection(collections[i].Collections) {
							return true
						}
					}
				}
				return false
			}
			addToCollection(workspaceData.Collections)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		cancelFunc()
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" New Collection ")
	return form
}

func createCollectionForm(app *tview.Application,
	pages *tview.Pages,
	workspaceData *workspace.Workspace,
	rootNode *tview.TreeNode,
	parentNode *tview.TreeNode,
	collectionsTreeView *tview.TreeView,
	parentCollection *workspace.Collection,
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

	form.AddInputField("Collection Name", "", 21, nil, nil)

	cancelFunc := func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(name) == "" {
			return
		}

		newCollection := workspace.Collection{Name: name}

		if parentCollection != nil {
			// Add to nested collection
			parentCollection.Collections = append(parentCollection.Collections, newCollection)
		} else {
			// Add to root level
			workspaceData.Collections = append(workspaceData.Collections, newCollection)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		cancelFunc()
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("newCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" New Collection ")
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

	form.AddInputField("Request Name", "", 41, nil, nil)
	form.AddDropDown("Method", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}, 0, nil)
	form.AddInputField("URL", "", 41, nil, nil)
	bodyInput := tview.NewInputField().
		SetLabel("Body: ").
		SetFieldWidth(41).
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

		// For multipart, the body field contains the formatted field data

		newRequest := workspace.Request{
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
					Name:     "Requests",
					Requests: []workspace.Request{newRequest},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else {
			// Find and update the actual collection in workspaceData
			actualCollection := findCollectionByName(&workspaceData.Collections, selectedCollection.Name)

			if actualCollection != nil {
				// Add request to the actual collection in workspaceData
				actualCollection.Requests = append(actualCollection.Requests, newRequest)
			}
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
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
					modifiedContent, err := openInExternalEditor(currentBody)
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

	form.AddInputField("Collection Name", selectedCollection.Name, 30, nil, nil)
	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		// Find and update the actual collection in workspaceData
		for i := range workspaceData.Collections {
			if workspaceData.Collections[i].Name == selectedCollection.Name {
				workspaceData.Collections[i].Name = newName
				break
			}
		}

		// Update tree node
		expanded := node.IsExpanded()
		if expanded {
			node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, newName))
		} else {
			node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, newName))
		}

		// Update node reference
		updatedCollection := *selectedCollection
		updatedCollection.Name = newName
		node.SetReference(updatedCollection)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		pages.RemovePage("renameCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})
	form.AddButton("Cancel", func() {
		pages.RemovePage("renameCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetBorder(true).SetTitle("Rename Collection")
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

	form.AddInputField("Request Name", selectedRequest.Name, 30, nil, nil)

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

		// Find and update the request in workspaceData
		// Check collections
		for i := range workspaceData.Collections {
			for j := range workspaceData.Collections[i].Requests {
				if workspaceData.Collections[i].Requests[j].Name == selectedRequest.Name &&
					workspaceData.Collections[i].Requests[j].Method == selectedRequest.Method &&
					workspaceData.Collections[i].Requests[j].URL == selectedRequest.URL {
					workspaceData.Collections[i].Requests[j].Name = newName

					// Update tree node
					coloredMethod := getColoredMethod(selectedRequest.Method)
					paddedName := padNameToMinLength(newName, 4)
					node.SetText(fmt.Sprintf("%s %s", coloredMethod, paddedName))

					// Update node reference
					updatedRequest := *selectedRequest
					updatedRequest.Name = newName
					node.SetReference(updatedRequest)

					// Save workspace
					if err := workspace.SaveWorkspace(workspaceData); err != nil {
						// Handle error
					}

					cancelFunc()
					return
				}
			}
		}
	})

	form.AddButton("Cancel", func() {
		cancelFunc()
	})

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
		// Find and remove the environment
		for i, e := range *environmentsData {
			if e.Name == selectedEnvironment.Name {
				*environmentsData = append((*environmentsData)[:i], (*environmentsData)[i+1:]...)
				break
			}
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

		// Find and update the actual environment in environmentsData
		for i := range *environmentsData {
			if (*environmentsData)[i].Name == selectedEnvironment.Name {
				(*environmentsData)[i].Name = newName
				break
			}
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
	form.AddButton("Cancel", func() {
		pages.RemovePage("renameEnvironment")
		app.SetFocus(currentFocus)
	})

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

func createDeleteCollectionConfirm(app *tview.Application, pages *tview.Pages, selectedCollection *workspace.Collection, workspaceData *workspace.Workspace, rootNode *tview.TreeNode, collectionsTreeView *tview.TreeView, node *tview.TreeNode, colors *ColorManager) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete the collection '%s'?\nThis will also delete all nested collections and requests.", selectedCollection.Name), 0, 2, false, false)

	form.AddButton("Delete", func() {
		// Remove collection from data
		deleteCollectionFromData(&workspaceData.Collections, selectedCollection.Name)

		// Rebuild tree from updated data
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetBorder(true).SetTitle(" Delete Collection ")
	return form
}

// Helper function to check if a collection is a descendant of another
func isDescendant(parent, child *workspace.Collection) bool {
	for _, col := range parent.Collections {
		if col.Name == child.Name || isDescendant(&col, child) {
			return true
		}
	}
	return false
}

// Helper function to remove a collection from its parent
func removeCollectionFromParent(workspaceData *workspace.Workspace, name string) {
	for i := range workspaceData.Collections {
		col := &workspaceData.Collections[i]
		if col.Name == name {
			workspaceData.Collections = append(workspaceData.Collections[:i], workspaceData.Collections[i+1:]...)
			return
		}
		removeCollectionFromParentNested(&col.Collections, name)
	}
}

// Helper function to remove a collection from nested collections
func removeCollectionFromParentNested(collections *[]workspace.Collection, name string) {
	for i := range *collections {
		if (*collections)[i].Name == name {
			*collections = append((*collections)[:i], (*collections)[i+1:]...)
			return
		}
		removeCollectionFromParentNested(&(*collections)[i].Collections, name)
	}
}

func createMoveCollectionForm(app *tview.Application, pages *tview.Pages, selectedCollection *workspace.Collection, workspaceData *workspace.Workspace, rootNode *tview.TreeNode, collectionsTreeView *tview.TreeView, node *tview.TreeNode, colors *ColorManager) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Get all possible parent collections (excluding self and descendants)
	var possibleParents []*workspace.Collection
	for i := range workspaceData.Collections {
		col := &workspaceData.Collections[i]
		if col.Name != selectedCollection.Name && !isDescendant(col, selectedCollection) {
			possibleParents = append(possibleParents, col)
		}
	}

	// Create dropdown for selecting new parent
	parentOptions := make([]string, len(possibleParents)+1)
	parentOptions[0] = "Root"
	for i, col := range possibleParents {
		parentOptions[i+1] = col.Name
	}

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
		var newParent *workspace.Collection
		if selectedIndex > 0 {
			newParent = possibleParents[selectedIndex-1]
		}

		// Remove from current parent
		removeCollectionFromParent(workspaceData, selectedCollection.Name)

		// Add to new parent
		if newParent != nil {
			newParent.Collections = append(newParent.Collections, *selectedCollection)
		} else {
			// Add to root
			workspaceData.Collections = append(workspaceData.Collections, *selectedCollection)
		}

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		// Rebuild tree
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetBorder(true).SetTitle(" Move Collection ")
	return form
}

// Helper function to remove a request from collections
func removeRequestFromCollections(workspaceData *workspace.Workspace, name, method, url string) {
	// Remove from collections
	for i := range workspaceData.Collections {
		col := &workspaceData.Collections[i]
		for j, req := range col.Requests {
			if req.Name == name && req.Method == method && req.URL == url {
				col.Requests = append(col.Requests[:j], col.Requests[j+1:]...)
				return
			}
		}
		// Recursive for nested
		if len(col.Collections) > 0 {
			if removeRequestFromNestedCollections(&col.Collections, name, method, url) {
				return
			}
		}
	}
}

func removeRequestFromNestedCollections(collections *[]workspace.Collection, name, method, url string) bool {
	for i := range *collections {
		col := &(*collections)[i]
		for j, req := range col.Requests {
			if req.Name == name && req.Method == method && req.URL == url {
				col.Requests = append(col.Requests[:j], col.Requests[j+1:]...)
				return true
			}
		}
		if len(col.Collections) > 0 {
			if removeRequestFromNestedCollections(&col.Collections, name, method, url) {
				return true
			}
		}
	}
	return false
}

func createMoveRequestForm(app *tview.Application, pages *tview.Pages, selectedRequest *workspace.Request, workspaceData *workspace.Workspace, rootNode *tview.TreeNode, collectionsTreeView *tview.TreeView, colors *ColorManager) *tview.Form {
	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	// Get all collections as possible targets, including nested
	var collectionOptions []string
	var targetCollections []*workspace.Collection
	collectionOptions = append(collectionOptions, "Root")

	var addCollectionsToOptions func(collections *[]workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections *[]workspace.Collection, prefix string) {
		for i := range *collections {
			col := &(*collections)[i]
			collectionOptions = append(collectionOptions, prefix+col.Name)
			targetCollections = append(targetCollections, col)
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

		if selectedIndex == 0 {
			// Move to first collection - first remove from current location, then add to first collection
			removeRequestFromCollections(workspaceData, selectedRequest.Name, selectedRequest.Method, selectedRequest.URL)
			if len(workspaceData.Collections) > 0 {
				workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, *selectedRequest)
			} else {
				// Create default collection
				defaultCollection := workspace.Collection{
					Name:     "Requests",
					Requests: []workspace.Request{*selectedRequest},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else if selectedIndex > 0 && selectedIndex <= len(targetCollections) {
			targetCollection := targetCollections[selectedIndex-1]

			// Remove from current collection
			removeRequestFromCollections(workspaceData, selectedRequest.Name, selectedRequest.Method, selectedRequest.URL)

			// Add to target collection
			targetCollection.Requests = append(targetCollection.Requests, *selectedRequest)
		}

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
		}

		// Rebuild tree
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Refresh the tree view
		collectionsTreeView.SetRoot(rootNode)

		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
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
) *tview.Form {

	form := tview.NewForm()
	form.SetBackgroundColor(colors.Background)
	form.SetBorderColor(colors.BorderFocus)
	form.SetTitleColor(colors.Title)
	form.SetLabelColor(colors.Foreground)
	form.SetButtonBackgroundColor(colors.Background)
	form.SetButtonTextColor(colors.Foreground)

	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete the request '%s'?", selectedRequest.Name), 0, 1, false, false)

	form.AddButton("Delete", func() {
		// Remove request from data
		deleteRequestFromData(workspaceData, selectedRequest.Name)

		// Rebuild tree from updated data
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
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

	// Handle Esc key to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			cancelFunc()
			return nil
		}
		return event
	})

	form.SetBorder(true).SetTitle(" Delete Request ")
	return form
}

func createWorkspaceManagementForm(
	app *tview.Application,
	pages *tview.Pages,
	workspaceData *workspace.Workspace,
	workspaceSelector *tview.DropDown,
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
		createForm := createNewWorkspaceForm(app, pages, workspaceSelector, rootNode, collectionsTreeView, colors)
		modal := createModal(createForm, 50, 8, tcell.ColorDefault)
		pages.AddPage("createWorkspace", modal, true, true)
		app.SetFocus(createForm)
	})

	form.AddButton("Switch Workspace", func() {
		pages.RemovePage("workspaceMenu")
		// Focus on workspace selector
		app.SetFocus(workspaceSelector)
	})

	form.AddButton("Rename Current Workspace", func() {
		pages.RemovePage("workspaceMenu")
		renameForm := createRenameWorkspaceForm(app, pages, currentWorkspace, workspaceSelector, colors)
		modal := createModal(renameForm, 50, 8, tcell.ColorDefault)
		pages.AddPage("renameWorkspace", modal, true, true)
		app.SetFocus(renameForm)
	})

	form.AddButton("Duplicate Workspace", func() {
		pages.RemovePage("workspaceMenu")
		duplicateForm := createDuplicateWorkspaceForm(app, pages, currentWorkspace, workspaceSelector, colors)
		modal := createModal(duplicateForm, 50, 10, tcell.ColorDefault)
		pages.AddPage("duplicateWorkspace", modal, true, true)
		app.SetFocus(duplicateForm)
	})

	form.AddButton("Delete Workspace", func() {
		pages.RemovePage("workspaceMenu")
		deleteForm := createDeleteWorkspaceForm(app, pages, currentWorkspace, workspaceSelector, colors)
		modal := createModal(deleteForm, 50, 8, tcell.ColorDefault)
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
	workspaceSelector *tview.DropDown,
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
			// Show error - for now just ignore
			return
		}

		// Update workspace selector
		workspaceNames, _ := workspace.ListWorkspaces()
		workspaceSelector.SetOptions(workspaceNames, nil)

		// Switch to new workspace
		err = workspace.SwitchWorkspace(name)
		if err == nil {
			workspaceSelector.SetCurrentOption(len(workspaceNames) - 1)
		}

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
	workspaceSelector *tview.DropDown,
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
		SetText(currentName).
		SetFieldWidth(30)

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

		// Update workspace selector
		workspaceNames, _ := workspace.ListWorkspaces()
		workspaceSelector.SetOptions(workspaceNames, nil)

		// Update current selection
		for i, name := range workspaceNames {
			if name == newName {
				workspaceSelector.SetCurrentOption(i)
				break
			}
		}

		pages.RemovePage("renameWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("renameWorkspace")
		pages.SwitchToPage("main")
		app.SetFocus(workspaceSelector)
	})

	form.SetBorder(true).SetTitle(" Rename Workspace ")

	// Add input capture to handle 'q' to cancel
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' || event.Rune() == 'Q' {
			pages.RemovePage("renameWorkspace")
			pages.SwitchToPage("main")
			app.SetFocus(workspaceSelector)
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
	workspaceSelector *tview.DropDown,
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
		SetLabel("New Workspace Name: ").
		SetFieldWidth(30)

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

		// Update workspace selector
		workspaceNames, _ := workspace.ListWorkspaces()
		workspaceSelector.SetOptions(workspaceNames, nil)

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
	workspaceSelector *tview.DropDown,
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

	form.AddTextView("Delete Workspace", fmt.Sprintf(" Are you sure you want to delete '%s'? ", workspaceName), 40, 2, true, false)
	form.AddTextView("", " This action cannot be undone. ", 40, 1, true, false)

	form.AddButton("Delete", func() {
		err := workspace.DeleteWorkspace(workspaceName)
		if err != nil {
			// Show error - for now just ignore
			return
		}

		// Update workspace selector
		workspaceNames, _ := workspace.ListWorkspaces()
		workspaceSelector.SetOptions(workspaceNames, nil)

		// Switch to first available workspace
		if len(workspaceNames) > 0 {
			err = workspace.SwitchWorkspace(workspaceNames[0])
			if err == nil {
				workspaceSelector.SetCurrentOption(0)
			}
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
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	methodIndex := 0
	for i, method := range methods {
		if method == originalRequest.Method {
			methodIndex = i
			break
		}
	}

	form.AddInputField("Request Name", duplicatedName, 41, nil, nil)
	form.AddDropDown("Method", methods, methodIndex, nil)
	form.AddInputField("URL", originalRequest.URL, 41, nil, nil)

	bodyInput := tview.NewInputField().
		SetLabel("Body: ").
		SetFieldWidth(41).
		SetText(originalRequest.Body).
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
			Name:        name,
			Method:      method,
			URL:         url,
			ContentType: contentType,
			Body:        body,
			Headers:     make(map[string]string), // Copy headers from original
		}

		// Copy headers from original request
		for key, value := range originalRequest.Headers {
			newRequest.Headers[key] = value
		}

		if selectedCollection == nil {
			// Add to first collection if no collection selected
			if len(workspaceData.Collections) > 0 {
				workspaceData.Collections[0].Requests = append(workspaceData.Collections[0].Requests, newRequest)
			} else {
				// Create a default collection
				defaultCollection := workspace.Collection{
					Name:     "Requests",
					Requests: []workspace.Request{newRequest},
				}
				workspaceData.Collections = append(workspaceData.Collections, defaultCollection)
			}
		} else {
			// Find and update the actual collection in workspaceData
			actualCollection := findCollectionByName(&workspaceData.Collections, selectedCollection.Name)

			if actualCollection != nil {
				// Add request to the actual collection in workspaceData
				actualCollection.Requests = append(actualCollection.Requests, newRequest)
			}
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addWorkspaceToTree(workspaceData, rootNode)

		// Save workspace
		if err := workspace.SaveWorkspace(workspaceData); err != nil {
			// Handle error
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
					modifiedContent, err := openInExternalEditor(currentBody)
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
