package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"

	"github.com/hbarral/petitorium/config"
	"github.com/hbarral/petitorium/workspace"
)

var rootCmd = &cobra.Command{
	Use:   "petitorium",
	Short: "A powerful TUI for API interaction and testing.",
	Run:   runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := config.LoadConfig(); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	})
}

func runTUI(cmd *cobra.Command, args []string) {
	backgroundColor, foregroundColor, borderColor, borderFocusColor, titleColor, selectionBackgroundColor := setupTheme()

	collectionsData, err := workspace.LoadCollections()
	if err != nil {
		panic(fmt.Sprintf("Failed to load collections: %v", err))
	}

	app := tview.NewApplication().
		EnableMouse(true)

	header := createPanel(" Petitorium ", backgroundColor, borderColor, titleColor, foregroundColor)
	// Change the root node's text to an empty string. It's cleaner.
	rootNode := tview.NewTreeNode("").SetSelectable(false)
	request := createPanel(" Request ", backgroundColor, borderColor, titleColor, foregroundColor)
	response := createPanel(" Response ", backgroundColor, borderColor, titleColor, foregroundColor)
	footer := createPanel("", backgroundColor, borderColor, titleColor, foregroundColor)
	footer.SetText(" [Tab] Cycle Focus | [Enter] Select / Toggle | [q] Quit | [n] New Collection | [r] New Request | [R] Rename | [m] Move Item | [d] Delete | [h] Collapse | [l] Expand")

	rightSide := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(request, 0, 1, false).
		AddItem(response, 0, 1, false)

	grid := tview.NewGrid().
		SetRows(3, 0, 3).
		SetColumns(30, 0).
		SetBorders(false)

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

	var lastSelectedRequestNode *tview.TreeNode

	collectionsTreeView.SetSelectedFunc(func(node *tview.TreeNode) {
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(selectionBackgroundColor).Foreground(foregroundColor))

		// Remove icon from previously selected request only when another request is selected
		if lastSelectedRequestNode != nil && lastSelectedRequestNode != node {
			if reference := lastSelectedRequestNode.GetReference(); reference != nil {
				if req, ok := reference.(workspace.Request); ok {
					// Only remove icon if the new selection is also a request
					if newReference := node.GetReference(); newReference != nil {
						if _, isRequest := newReference.(workspace.Request); isRequest {
							coloredMethod := getColoredMethod(req.Method)
							paddedName := padNameToMinLength(req.Name, 4)
							// Restore to reserved space (icon width + fixed space)
							iconWidth := getIconDisplayWidth(config.C.UI.SelectedRequestIcon) // + 1
							spacePadding := strings.Repeat(" ", iconWidth)
							lastSelectedRequestNode.SetText(fmt.Sprintf("%s%s%s", spacePadding, coloredMethod, paddedName))
						}
					}
				}
			}
		}

		// Add icon to newly selected node if it's a request
		reference := node.GetReference()
		if req, ok := reference.(workspace.Request); ok {
			coloredMethod := getColoredMethod(req.Method)
			paddedName := padNameToMinLength(req.Name, 4)
			// Create colored icon with configured color, always add space after icon
			iconColor := config.C.UI.SelectedRequestIconColor
			coloredIcon := fmt.Sprintf("[%s]%s[-:-:-]", iconColor, config.C.UI.SelectedRequestIcon)
			node.SetText(fmt.Sprintf("%s%s%s", coloredIcon, coloredMethod, paddedName))

			var sb strings.Builder
			sb.WriteString(fmt.Sprintf("Method: %s\n", req.Method))
			sb.WriteString(fmt.Sprintf("URL: %s", req.URL))
			if req.Body != "" {
				sb.WriteString(fmt.Sprintf("\n\n-- Body --\n%s", req.Body))
			}
			request.SetText(sb.String()).SetTextAlign(tview.AlignLeft)
		} else if col, ok := reference.(workspace.Collection); ok {
			// Handle collection expansion
			expanded := !node.IsExpanded()
			node.SetExpanded(expanded)

			if expanded {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
				if len(node.GetChildren()) == 0 {
					addChildrenToCollectionNode(node, col)
				}
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&collectionsData, col.Name, true)
				}
			} else {
				node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
				node.ClearChildren()
				if config.C.UI.CollectionExpansion == "remember" {
					updateCollectionExpansionState(&collectionsData, col.Name, false)
				}
			}
		}

		// Track the last selected request node
		if reference := node.GetReference(); reference != nil {
			if _, ok := reference.(workspace.Request); ok {
				lastSelectedRequestNode = node
			}
		}
	})

	// For "closed" mode, we need to ensure all collections start collapsed
	// For "remember" mode, the expansion state is handled within addCollectionsToTree
	if config.C.UI.CollectionExpansion == "closed" {
		// Temporarily set config to "closed" for tree building, then restore
		originalExpansion := config.C.UI.CollectionExpansion
		config.C.UI.CollectionExpansion = "closed"
		addCollectionsToTree(collectionsData, rootNode)
		config.C.UI.CollectionExpansion = originalExpansion
	} else {
		addCollectionsToTree(collectionsData, rootNode)
	}

	grid.AddItem(header, 0, 0, 1, 2, 0, 0, false)
	grid.AddItem(footer, 2, 0, 1, 2, 0, 0, false)
	grid.AddItem(collectionsTreeView, 1, 0, 1, 1, 0, 0, true)
	grid.AddItem(rightSide, 1, 1, 1, 1, 0, 0, false)

	currentFocus := 0
	panels := []tview.Primitive{collectionsTreeView, request, response}

	setFocusStyle(panels[currentFocus], true, borderColor, borderFocusColor)

	// Create pages for modals
	pages := tview.NewPages()

	// Add main page
	pages.AddPage("main", grid, true, true)

	// Helper function to center modals
	createModal := func(p tview.Primitive, width, height int) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, height, 1, true).
				AddItem(nil, 0, 1, false), width, 1, false).
			AddItem(nil, 0, 1, false)
	}

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Check if we're on a modal page (not main)
		currentPage, _ := pages.GetFrontPage()
		if currentPage != "main" {
			// Let the form handle its own input
			return event
		}

		if event.Rune() == 'q' || event.Rune() == 'Q' {
			// Save expansion state before quitting if in "remember" mode
			if config.C.UI.CollectionExpansion == "remember" {
				if err := workspace.SaveExpansionState(collectionsData); err != nil {
					fmt.Printf("Warning: Failed to save expansion state: %v\n", err)
				}
			}
			app.Stop()
			return nil
		}
		if event.Key() == tcell.KeyTab {
			setFocusStyle(panels[currentFocus], false, borderColor, borderFocusColor)
			currentFocus = (currentFocus + 1) % len(panels)
			app.SetFocus(panels[currentFocus])
			setFocusStyle(panels[currentFocus], true, borderColor, borderFocusColor)
			return nil
		}
		if event.Key() == tcell.KeyBacktab {
			setFocusStyle(panels[currentFocus], false, borderColor, borderFocusColor)
			currentFocus = (currentFocus - 1 + len(panels)) % len(panels)
			app.SetFocus(panels[currentFocus])
			setFocusStyle(panels[currentFocus], true, borderColor, borderFocusColor)
			return nil
		}

		if event.Rune() == 'n' {
			// New collection - check if we're adding to root or nested
			node := collectionsTreeView.GetCurrentNode()
			var parentCollection *workspace.Collection = nil
			var parentNode *tview.TreeNode = rootNode

			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Find the actual collection in the data to get a proper pointer
					parentCollection = findCollectionByName(&collectionsData, col.Name)
					parentNode = node
				}
			}

			form := createCollectionForm(app, pages, &collectionsData, rootNode, parentNode, collectionsTreeView, parentCollection)
			modal := createModal(form, 40, 10)
			pages.AddPage("newCollection", modal, true, true)
			app.SetFocus(form)
			return nil
		}

		if event.Rune() == 'r' {
			// New request - check if a collection or request is selected
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				var selectedCollection *workspace.Collection

				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Collection is selected
					selectedCollection = &col
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Request is selected - find its parent collection
					selectedCollection = findParentCollectionOfRequest(&collectionsData, req.Name, req.Method, req.URL)
				}

				if selectedCollection != nil {
					form := createRequestForm(app, pages, selectedCollection, &collectionsData, rootNode, collectionsTreeView)
					modal := createModal(form, 60, 14)
					pages.AddPage("newRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Rename functionality (Shift+R)
		if event.Rune() == 'R' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Rename collection
					form := createRenameCollectionForm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 10)
					pages.AddPage("renameCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Rename request - need to find parent collection
					form := createRenameRequestForm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 10)
					pages.AddPage("renameRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Move collection/request functionality (M)
		if event.Rune() == 'm' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Move collection
					form := createMoveCollectionForm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 40, 12)
					pages.AddPage("moveCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := node.GetReference().(workspace.Request); ok {
					// Move request
					form := createMoveRequestForm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView)
					modal := createModal(form, 40, 10)
					pages.AddPage("moveRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Delete functionality (d)
		if event.Rune() == 'd' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()

				if col, ok := reference.(workspace.Collection); ok {
					// Delete collection with confirmation
					form := createDeleteCollectionConfirm(app, pages, &col, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 50, 8)
					pages.AddPage("deleteCollection", modal, true, true)
					app.SetFocus(form)
					return nil
				} else if req, ok := reference.(workspace.Request); ok {
					// Delete request with confirmation
					form := createDeleteRequestConfirm(app, pages, &req, &collectionsData, rootNode, collectionsTreeView, node)
					modal := createModal(form, 50, 8)
					pages.AddPage("deleteRequest", modal, true, true)
					app.SetFocus(form)
					return nil
				}
			}
		}

		// Vim-style collection navigation: 'h' to collapse, 'l' to expand
		if event.Rune() == 'h' || event.Rune() == 'l' {
			node := collectionsTreeView.GetCurrentNode()
			if node != nil {
				if col, ok := node.GetReference().(workspace.Collection); ok {
					// Check if this is a nested collection (has a parent that is also a collection)
					parentNode := findParentNode(rootNode, node)
					isNestedCollection := parentNode != nil && parentNode != rootNode

					if event.Rune() == 'h' {
						if isNestedCollection && parentNode.IsExpanded() {
							// For nested collections, collapse the parent collection
							if parentCol, ok := parentNode.GetReference().(workspace.Collection); ok {
								parentNode.SetExpanded(false)
								parentNode.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, parentCol.Name))
								parentNode.ClearChildren()
								if config.C.UI.CollectionExpansion == "remember" {
									updateCollectionExpansionState(&collectionsData, parentCol.Name, false)
								}
								// Move selection to the parent collection
								collectionsTreeView.SetCurrentNode(parentNode)
								return nil
							}
						} else if node.IsExpanded() {
							// For root collections, collapse the current collection
							node.SetExpanded(false)
							node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
							node.ClearChildren()
							if config.C.UI.CollectionExpansion == "remember" {
								updateCollectionExpansionState(&collectionsData, col.Name, false)
							}
							return nil
						}
					} else if event.Rune() == 'l' && !node.IsExpanded() {
						// Expand collection (works for both root and nested collections)
						node.SetExpanded(true)
						node.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionExpandedIcon, col.Name))
						if len(node.GetChildren()) == 0 {
							addChildrenToCollectionNode(node, col)
						}
						if config.C.UI.CollectionExpansion == "remember" {
							updateCollectionExpansionState(&collectionsData, col.Name, true)
						}
						return nil
					}
				} else if _, ok := node.GetReference().(workspace.Request); ok {
					// Handle request navigation - 'h' collapses parent collection
					if event.Rune() == 'h' {
						parentNode := findParentNode(rootNode, node)
						if parentNode != nil {
							if col, ok := parentNode.GetReference().(workspace.Collection); ok && parentNode.IsExpanded() {
								// Collapse parent collection
								parentNode.SetExpanded(false)
								parentNode.SetText(fmt.Sprintf("%s %s", config.C.UI.CollectionIcon, col.Name))
								parentNode.ClearChildren()
								if config.C.UI.CollectionExpansion == "remember" {
									updateCollectionExpansionState(&collectionsData, col.Name, false)
								}
								// Move selection to the parent collection
								collectionsTreeView.SetCurrentNode(parentNode)
								return nil
							}
						}
					}
					// 'l' on a request does nothing (requests can't be expanded)
				}
			}
		}

		return event
	})

	if err := app.SetRoot(pages, true).SetFocus(collectionsTreeView).Run(); err != nil {
		panic(err)
	}
}
