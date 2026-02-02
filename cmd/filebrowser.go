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

	fb.setRoot(startPath)

	// Set selected logic
	fb.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Should not happen for valid nodes
		}
		path := reference.(string)

		// Check if it's the ".." entry
		if node.GetText() == ".." {
			fb.setRoot(path) // Re-root to parent
			return
		}

		info, err := os.Stat(path)
		if err != nil {
			return
		}

		if info.IsDir() {
			if node.IsExpanded() {
				node.SetExpanded(false)
			} else {
				node.SetExpanded(true)
				// Load children if not already loaded
				if len(node.GetChildren()) == 0 {
					fb.addNodes(node, path)
				}
			}
		} else {
			if fb.onSelect != nil {
				fb.onSelect(path)
			}
		}
	})

	return fb, nil
}

// setRoot updates the root of the tree to the specified path
func (fb *FileBrowser) setRoot(path string) {
	fb.current = path
	fb.root = tview.NewTreeNode(path).
		SetColor(fb.colors.Title).
		SetSelectable(true).
		SetReference(path)

	fb.tree.SetRoot(fb.root)
	fb.tree.SetCurrentNode(fb.root)

	// Always add ".." if not at root
	parentPath := filepath.Dir(path)
	if parentPath != path {
		parentNode := tview.NewTreeNode("..").
			SetReference(parentPath).
			SetSelectable(true).
			SetColor(fb.colors.Foreground)
		fb.root.AddChild(parentNode)
	}

	fb.addNodes(fb.root, path)
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
			SetSelectable(true)

		if file.IsDir() {
			node.SetColor(fb.colors.Foreground) // Directories
		} else {
			node.SetColor(fb.colors.Foreground) // Files
		}

		// Use tree selection style
		node.SetSelectedTextStyle(tcell.StyleDefault.Background(fb.colors.TreeSelection).Foreground(fb.colors.Foreground))

		target.AddChild(node)
	}
}
