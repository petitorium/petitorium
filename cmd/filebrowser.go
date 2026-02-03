package cmd

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// FileBrowser manages the file selection interface
type FileBrowser struct {
	tree     *tview.TreeView
	root     *tview.TreeNode
	current  string // Current path
	colors   *ColorManager
	onSelect func(string)
}

// createFileBrowser creates a new file browser component
func createFileBrowser(startPath string, colors *ColorManager, onSelect func(string)) (*FileBrowser, error) {
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}

	// Ensure startPath is absolute
	startPath, err := filepath.Abs(startPath)
	if err != nil {
		return nil, err
	}

	// If startPath is a file, use its parent directory
	info, err := os.Stat(startPath)
	if err == nil && !info.IsDir() {
		startPath = filepath.Dir(startPath)
	}

	fb := &FileBrowser{
		colors:   colors,
		onSelect: onSelect,
		current:  startPath,
	}

	fb.tree = tview.NewTreeView()
	fb.tree.SetBackgroundColor(colors.Background)
	fb.tree.SetBorder(true)
	fb.tree.SetBorderColor(colors.BorderFocus)
	fb.tree.SetTitle(" File Browser ")
	fb.tree.SetTitleColor(colors.Title)
	fb.tree.SetGraphics(false)

	fb.setRoot(startPath)

	// Handle selection logic
	fb.tree.SetSelectedFunc(fb.handleSelect)

	// Set up keyboard navigation
	fb.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEnter || event.Rune() == 'l' {
			node := fb.tree.GetCurrentNode()
			if node != nil {
				reference := node.GetReference()
				if reference != nil {
					path := reference.(string)
					// If it's the ".." entry, handle rerooting
					if node.GetText() == ".." {
						fb.setRoot(path)
						return nil
					}

					info, err := os.Stat(path)
					if err == nil && info.IsDir() {
						if node.IsExpanded() {
							node.SetExpanded(false)
							node.ClearChildren()
							dummy := tview.NewTreeNode("").
								SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background))
							node.AddChild(dummy) // Restore dummy
							// Update icon to closed folder
							text := node.GetText()
							node.SetText(strings.Replace(text, "📂", "📁", 1))
						} else {
							node.SetExpanded(true)
							node.ClearChildren() // Remove dummy
							fb.addNodes(node, path)
							// Update icon to open folder
							text := node.GetText()
							node.SetText(strings.Replace(text, "📁", "📂", 1))
						}
						return nil
					}
				}
				// For files, trigger selection
				if event.Key() == tcell.KeyEnter {
					// Use handleSelect for files
					fb.handleSelect(node)
					return nil
				}
				if event.Rune() == 'l' {
					fb.handleSelect(node)
					return nil
				}
			}
			return nil
		}

		switch event.Rune() {
		case 'h':
			node := fb.tree.GetCurrentNode()
			if node != nil {
				if node.IsExpanded() {
					node.SetExpanded(false)
					node.ClearChildren()
					dummy := tview.NewTreeNode("").
						SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background))
					node.AddChild(dummy)
					text := node.GetText()
					node.SetText(strings.Replace(text, "📂", "📁", 1))
				} else {
					parent := fb.findParentNode(fb.tree.GetRoot(), node)
					if parent != nil {
						fb.tree.SetCurrentNode(parent)
					}
				}
			}
			return nil
		case 'g':
			nodes := fb.getVisibleNodes()
			if len(nodes) > 0 {
				fb.tree.SetCurrentNode(nodes[0])
			}
			return nil
		case 'G':
			nodes := fb.getVisibleNodes()
			if len(nodes) > 0 {
				fb.tree.SetCurrentNode(nodes[len(nodes)-1])
			}
			return nil
		case 'j':
			return tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
		case 'k':
			return tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone)
		}
		return event
	})

	return fb, nil
}

// handleSelect handles the selection of a node
func (fb *FileBrowser) handleSelect(node *tview.TreeNode) {
	reference := node.GetReference()
	if reference == nil {
		return // Should not happen for valid nodes
	}
	path := reference.(string)

	// Check if it's the ".." entry
	if node.GetText() == ".." {
		fb.setRoot(path)
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		return
	}

	if info.IsDir() {
		// Directory expansion is handled in InputCapture
		return
	}

	if fb.onSelect != nil {
		fb.onSelect(path)
	}
}

// findParentNode finds the parent of a node in the tree
func (fb *FileBrowser) findParentNode(root, target *tview.TreeNode) *tview.TreeNode {
	if root == nil || target == nil || root == target {
		return nil
	}
	for _, child := range root.GetChildren() {
		if child == target {
			return root
		}
		if parent := fb.findParentNode(child, target); parent != nil {
			return parent
		}
	}
	return nil
}

// getVisibleNodes returns a list of all visible nodes in the tree
func (fb *FileBrowser) getVisibleNodes() []*tview.TreeNode {
	var nodes []*tview.TreeNode
	var collect func(*tview.TreeNode)
	collect = func(node *tview.TreeNode) {
		if node == nil {
			return
		}
		nodes = append(nodes, node)
		if node.IsExpanded() {
			for _, child := range node.GetChildren() {
				collect(child)
			}
		}
	}
	collect(fb.tree.GetRoot())
	return nodes
}

// setRoot updates the root of the tree to the specified path
func (fb *FileBrowser) setRoot(path string) {
	fb.current = path
	fb.root = tview.NewTreeNode(path).
		SetColor(fb.colors.Title).
		SetSelectable(true).
		SetExpanded(true).
		SetReference(path).
		SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background))

	fb.tree.SetRoot(fb.root)
	fb.tree.SetCurrentNode(fb.root)

	fb.addNodes(fb.root, path)

	// Always add ".." if not at root
	parentPath := filepath.Dir(path)
	if parentPath != path {
		parentNode := tview.NewTreeNode("..").
			SetReference(parentPath).
			SetSelectable(true).
			SetColor(fb.colors.Foreground).
			SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background)).
			SetSelectedTextStyle(tcell.StyleDefault.Background(fb.colors.TreeSelection).Foreground(fb.colors.Foreground))

		// We want ".." to be at the top. Since addNodes just finished,
		// we can prepend it by getting children, clearing, and re-adding.
		children := fb.root.GetChildren()
		fb.root.ClearChildren()
		fb.root.AddChild(parentNode)
		for _, child := range children {
			fb.root.AddChild(child)
		}
	}
}

// addNodes adds child nodes to the given parent node based on the directory path
func (fb *FileBrowser) addNodes(target *tview.TreeNode, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}

	// Sort files: directories first, then files
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir() && !files[j].IsDir() {
			return true
		}
		if !files[i].IsDir() && files[j].IsDir() {
			return false
		}
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		// Skip hidden files
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		displayName := file.Name()
		if file.IsDir() {
			displayName = "📁 " + displayName
		} else {
			displayName = "📄 " + displayName
		}

		node := tview.NewTreeNode(displayName).
			SetReference(filepath.Join(path, file.Name())).
			SetSelectable(true).
			SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background))

		if file.IsDir() {
			node.SetColor(fb.colors.Foreground) // Directories
			dummy := tview.NewTreeNode("").
				SetTextStyle(tcell.StyleDefault.Background(fb.colors.Background))
			node.AddChild(dummy) // Add dummy child to make it expandable
			node.SetExpanded(false)
		} else {
			node.SetColor(fb.colors.Foreground) // Files
		}

		// Use tree selection style
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(fb.colors.TreeSelection).Foreground(fb.colors.Foreground))

		target.AddChild(node)
	}
}
