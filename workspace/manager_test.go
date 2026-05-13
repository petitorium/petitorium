package workspace

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestWorkspaceNameMatchesDirectory(t *testing.T) {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		t.Skip("No workspace manager available")
	}

	for _, ws := range manager.Workspaces {
		workspaceDir := getWorkspaceDir(ws.Name)
		workspacePath := filepath.Join(workspaceDir, "workspace.yaml")

		if _, err := os.Stat(workspacePath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(workspacePath)
		if err != nil {
			continue
		}

		var workspace Workspace
		if err := yaml.Unmarshal(data, &workspace); err != nil {
			continue
		}

		if workspace.Name != ws.Name {
			t.Errorf("Workspace name mismatch: directory name=%s, workspace.Name=%s", ws.Name, workspace.Name)
		}
	}
}

func TestCreateWorkspaceSetsCorrectName(t *testing.T) {
	testName := "test-workspace-name-check"

	// Clean up if exists
	dir := getWorkspaceDir(testName)
	os.RemoveAll(dir)
	removeFromManager(testName)

	ws, err := CreateWorkspace(testName)
	if err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	if ws.Name != testName {
		t.Errorf("Created workspace has name=%s, expected %s", ws.Name, testName)
	}

	// Verify the name in the file matches
	workspacePath := filepath.Join(dir, "workspace.yaml")
	data, err := os.ReadFile(workspacePath)
	if err != nil {
		t.Fatalf("Failed to read workspace file: %v", err)
	}

	var loaded Workspace
	if err := yaml.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal workspace: %v", err)
	}

	if loaded.Name != testName {
		t.Errorf("Workspace file has name=%s, expected %s", loaded.Name, testName)
	}

	// Clean up
	os.RemoveAll(dir)
	removeFromManager(testName)
}

func removeFromManager(name string) {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		return
	}
	for i, ws := range manager.Workspaces {
		if ws.Name == name {
			manager.Workspaces = append(manager.Workspaces[:i], manager.Workspaces[i+1:]...)
			break
		}
	}
	SaveWorkspaceManager(manager)
}
