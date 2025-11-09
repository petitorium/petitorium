package cmd

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
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

	form.AddInputField("Collection Name", "", 21, nil, nil)
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

	form.AddInputField("Request Name", "", 43, nil, nil)
	form.AddDropDown("Method", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}, 0, nil)
	form.AddInputField("URL", "", 43, nil, nil)
	form.AddInputField("Body", "", 43, nil, nil)

	form.AddButton("Save", func() {
		name := form.GetFormItem(0).(*tview.InputField).GetText()
		_, method := form.GetFormItem(1).(*tview.DropDown).GetCurrentOption()
		url := form.GetFormItem(2).(*tview.InputField).GetText()
		body := form.GetFormItem(3).(*tview.InputField).GetText()

		if strings.TrimSpace(name) == "" || strings.TrimSpace(url) == "" {
			return
		}

		newRequest := workspace.Request{
			Name:   name,
			Method: method,
			URL:    url,
			Body:   body,
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
	envDropdown *tview.DropDown,
	envConfigButton *tview.Button,
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

		// Save environments
		if err := workspace.SaveEnvironments(*environmentsData); err != nil {
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
	envDropdown *tview.DropDown,
	envConfigButton *tview.Button,
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

		// Save environments
		if err := workspace.SaveEnvironments(*environmentsData); err != nil {
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
