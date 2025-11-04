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
	collectionsData *[]workspace.Collection,
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
	addCollectionsToOptions(*collectionsData, "")

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
			*collectionsData = append(*collectionsData, newCollection)
		} else {
			// Find target collection and add to it
			targetName := strings.Split(location, " → ")[0]
			var addToCollection func(collections *[]workspace.Collection) bool
			addToCollection = func(collections *[]workspace.Collection) bool {
				for i := range *collections {
					if (*collections)[i].Name == targetName {
						(*collections)[i].Collections = append((*collections)[i].Collections, newCollection)
						return true
					}
					if len((*collections)[i].Collections) > 0 {
						if addToCollection(&(*collections)[i].Collections) {
							return true
						}
					}
				}
				return false
			}
			addToCollection(collectionsData)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
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
	collectionsData *[]workspace.Collection,
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
			*collectionsData = append(*collectionsData, newCollection)
		}

		// Rebuild the entire tree to reflect changes
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
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
	collectionsData *[]workspace.Collection,
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

		// Find and update the actual collection in collectionsData
		actualCollection := findCollectionByName(collectionsData, selectedCollection.Name)

		if actualCollection != nil {
			// Add request to the actual collection in collectionsData
			actualCollection.Requests = append(actualCollection.Requests, newRequest)

			// Rebuild the entire tree to reflect changes
			rootNode.ClearChildren()
			addCollectionsToTree(*collectionsData, rootNode)

			// Save workspace
			if err := workspace.SaveCollections(*collectionsData); err != nil {
				// Handle error
			}
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
	collectionsData *[]workspace.Collection,
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

		// Find and update the actual collection in collectionsData
		for i := range *collectionsData {
			if (*collectionsData)[i].Name == selectedCollection.Name {
				(*collectionsData)[i].Name = newName
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
		if err := workspace.SaveCollections(*collectionsData); err != nil {
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
	collectionsData *[]workspace.Collection,
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

		// Find and update the request in collectionsData
		for i := range *collectionsData {
			for j := range (*collectionsData)[i].Requests {
				if (*collectionsData)[i].Requests[j].Name == selectedRequest.Name &&
					(*collectionsData)[i].Requests[j].Method == selectedRequest.Method &&
					(*collectionsData)[i].Requests[j].URL == selectedRequest.URL {
					(*collectionsData)[i].Requests[j].Name = newName

					// Update tree node
					coloredMethod := getColoredMethod(selectedRequest.Method)
					paddedName := padNameToMinLength(newName, 4)
					node.SetText(fmt.Sprintf("%s %s", coloredMethod, paddedName))

					// Update node reference
					updatedRequest := *selectedRequest
					updatedRequest.Name = newName
					node.SetReference(updatedRequest)

					// Save workspace
					if err := workspace.SaveCollections(*collectionsData); err != nil {
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

func createRenameEnvironmentForm(app *tview.Application,
	pages *tview.Pages,
	selectedEnvironment *workspace.Environment,
	environmentsData *[]workspace.Environment,
	envDropdown *tview.DropDown,
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

	form.AddInputField("Environment Name", selectedEnvironment.Name, 30, nil, nil)

	form.AddButton("Save", func() {
		newName := form.GetFormItem(0).(*tview.InputField).GetText()
		if strings.TrimSpace(newName) == "" {
			return
		}

		// Find and update the environment in environmentsData
		for i := range *environmentsData {
			if (*environmentsData)[i].Name == selectedEnvironment.Name {
				(*environmentsData)[i].Name = newName
				break
			}
		}

		// Save environments
		if err := workspace.SaveEnvironments(*environmentsData); err != nil {
			// Handle error
		}

		// Update environment dropdown
		updateEnvironmentDropdown(envDropdown, *environmentsData)

		pages.RemovePage("renameEnvironment")
		pages.SwitchToPage("main")
		app.SetFocus(envDropdown)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("renameEnvironment")
		pages.SwitchToPage("main")
		app.SetFocus(envDropdown)
	})

	form.SetBorder(true).SetTitle(" Rename Environment ")
	return form
}

func createMoveCollectionForm(app *tview.Application,
	pages *tview.Pages,
	selectedCollection *workspace.Collection,
	collectionsData *[]workspace.Collection,
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

	// Get all available collections for movement targets
	var targetOptions []string
	targetOptions = append(targetOptions, "(Root Level)")

	var addCollectionsToOptions func(collections []workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections []workspace.Collection, prefix string) {
		for _, col := range collections {
			if col.Name != selectedCollection.Name { // Don't allow moving into itself
				targetOptions = append(targetOptions, prefix+col.Name)
				if len(col.Collections) > 0 {
					addCollectionsToOptions(col.Collections, prefix+col.Name+" → ")
				}
			}
		}
	}
	addCollectionsToOptions(*collectionsData, "")

	form.AddDropDown("Move to", targetOptions, 0, nil)

	form.AddButton("Move", func() {
		_, target := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()

		// Remove collection from current location
		var removeFromCollection func(collections *[]workspace.Collection) bool
		removeFromCollection = func(collections *[]workspace.Collection) bool {
			for i := range *collections {
				if (*collections)[i].Name == selectedCollection.Name {
					// Remove from current location
					*collections = append((*collections)[:i], (*collections)[i+1:]...)
					return true
				}
				if len((*collections)[i].Collections) > 0 {
					if removeFromCollection(&(*collections)[i].Collections) {
						return true
					}
				}
			}
			return false
		}

		// Remove from current location
		removeFromCollection(collectionsData)

		// Add to target location
		if target == "(Root Level)" {
			// Add to root
			*collectionsData = append(*collectionsData, *selectedCollection)
		} else {
			// Find target collection and add to it
			targetName := strings.Split(target, " → ")[0]
			var addToCollection func(collections *[]workspace.Collection) bool
			addToCollection = func(collections *[]workspace.Collection) bool {
				for i := range *collections {
					if (*collections)[i].Name == targetName {
						(*collections)[i].Collections = append((*collections)[i].Collections, *selectedCollection)
						return true
					}
					if len((*collections)[i].Collections) > 0 {
						if addToCollection(&(*collections)[i].Collections) {
							return true
						}
					}
				}
				return false
			}
			addToCollection(collectionsData)
		}

		// Rebuild the tree
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
			// Handle error
		}

		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("moveCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Move Collection ")
	return form
}

func createMoveRequestForm(app *tview.Application,
	pages *tview.Pages,
	selectedRequest *workspace.Request,
	collectionsData *[]workspace.Collection,
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

	// Get all available collections for movement targets
	var targetOptions []string

	var addCollectionsToOptions func(collections []workspace.Collection, prefix string)
	addCollectionsToOptions = func(collections []workspace.Collection, prefix string) {
		for _, col := range collections {
			targetOptions = append(targetOptions, prefix+col.Name)
			if len(col.Collections) > 0 {
				addCollectionsToOptions(col.Collections, prefix+col.Name+" → ")
			}
		}
	}
	addCollectionsToOptions(*collectionsData, "")

	form.AddDropDown("Move to", targetOptions, 0, nil)

	cancelFunc := func() {
		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Move", func() {
		_, target := form.GetFormItem(0).(*tview.DropDown).GetCurrentOption()

		// Remove request from current location
		var removeFromCollection func(collections *[]workspace.Collection) bool
		removeFromCollection = func(collections *[]workspace.Collection) bool {
			for i := range *collections {
				// Check if request is in this collection
				for j := len((*collections)[i].Requests) - 1; j >= 0; j-- {
					req := (*collections)[i].Requests[j]
					if req.Name == selectedRequest.Name && req.Method == selectedRequest.Method && req.URL == selectedRequest.URL {
						// Remove from current location
						(*collections)[i].Requests = append((*collections)[i].Requests[:j], (*collections)[i].Requests[j+1:]...)
						return true
					}
				}
				// Check nested collections
				if len((*collections)[i].Collections) > 0 {
					if removeFromCollection(&(*collections)[i].Collections) {
						return true
					}
				}
			}
			return false
		}

		// Remove from current location
		removeFromCollection(collectionsData)

		// Add to target location
		targetName := strings.Split(target, " → ")[0]
		var addToCollection func(collections *[]workspace.Collection) bool
		addToCollection = func(collections *[]workspace.Collection) bool {
			for i := range *collections {
				if (*collections)[i].Name == targetName {
					(*collections)[i].Requests = append((*collections)[i].Requests, *selectedRequest)
					return true
				}
				if len((*collections)[i].Collections) > 0 {
					if addToCollection(&(*collections)[i].Collections) {
						return true
					}
				}
			}
			return false
		}
		addToCollection(collectionsData)

		// Rebuild the tree
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
			// Handle error
		}

		cancelFunc()
	})

	form.AddButton("Cancel", func() {
		pages.RemovePage("moveRequest")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Move Request ")
	return form
}

func createDeleteCollectionConfirm(app *tview.Application,
	pages *tview.Pages,
	selectedCollection *workspace.Collection,
	collectionsData *[]workspace.Collection,
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

	form.AddTextView("", fmt.Sprintf("Are you sure you want to delete the collection '%s'?\nThis will also delete all nested collections and requests.", selectedCollection.Name), 0, 2, false, false)

	form.AddButton("Delete", func() {
		// Remove collection from data
		deleteCollectionFromData(collectionsData, selectedCollection.Name)

		// Rebuild tree from updated data
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
			// Handle error
		}

		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	})

	cancelFunc := func() {
		pages.RemovePage("deleteCollection")
		pages.SwitchToPage("main")
		app.SetFocus(collectionsTreeView)
	}

	form.AddButton("Cancel", cancelFunc)

	form.SetCancelFunc(cancelFunc)

	form.SetBorder(true).SetTitle(" Delete Collection ")
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

func createDeleteRequestConfirm(app *tview.Application,
	pages *tview.Pages,
	selectedRequest *workspace.Request,
	collectionsData *[]workspace.Collection,
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
		deleteRequestFromData(collectionsData, selectedRequest.Name)

		// Rebuild tree from updated data
		rootNode.ClearChildren()
		addCollectionsToTree(*collectionsData, rootNode)

		// Save workspace
		if err := workspace.SaveCollections(*collectionsData); err != nil {
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
