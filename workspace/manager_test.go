package workspace

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestGetEffectiveVariablesResolvesReferences(t *testing.T) {
	envs := []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"ipAddress":     "111.222.333.444",
				"dispositivoId": "{{ipAddress}}_abcdefhij",
				"protocol":      "http://",
				"baseURL":       "{{protocol}}api.example.com",
			},
		},
	}

	env := &envs[0]
	effective := env.GetEffectiveVariables(envs)
	ResolveVariableReferences(effective)

	if effective["ipAddress"] != "111.222.333.444" {
		t.Errorf("ipAddress = %q, want %q", effective["ipAddress"], "111.222.333.444")
	}
	if effective["dispositivoId"] != "111.222.333.444_abcdefhij" {
		t.Errorf("dispositivoId = %q, want %q", effective["dispositivoId"], "111.222.333.444_abcdefhij")
	}
	if effective["baseURL"] != "http://api.example.com" {
		t.Errorf("baseURL = %q, want %q", effective["baseURL"], "http://api.example.com")
	}
}

func TestGetEffectiveVariablesMultiLevelChaining(t *testing.T) {
	envs := []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"a": "{{b}}",
				"b": "{{c}}",
				"c": "final",
			},
		},
	}

	env := &envs[0]
	effective := env.GetEffectiveVariables(envs)
	ResolveVariableReferences(effective)

	if effective["a"] != "final" {
		t.Errorf("a = %q, want %q", effective["a"], "final")
	}
	if effective["b"] != "final" {
		t.Errorf("b = %q, want %q", effective["b"], "final")
	}
}

func TestGetEffectiveVariablesCircularReferences(t *testing.T) {
	envs := []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"x": "{{y}}",
				"y": "{{x}}",
			},
		},
	}

	env := &envs[0]
	done := make(chan struct{})
	var effective map[string]string

	go func() {
		effective = env.GetEffectiveVariables(envs)
		ResolveVariableReferences(effective)
		close(done)
	}()

	select {
	case <-done:
		// expected
	case <-time.After(2 * time.Second):
		t.Fatal("GetEffectiveVariables hung on circular references")
	}

	// Circular references resolve to self-referential placeholders.
	if effective["x"] != "{{x}}" {
		t.Errorf("x = %q, want %q", effective["x"], "{{x}}")
	}
	if effective["y"] != "{{x}}" {
		t.Errorf("y = %q, want %q", effective["y"], "{{x}}")
	}
}

func TestGetEffectiveVariablesWithBaseInheritance(t *testing.T) {
	envs := []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"host":     "localhost",
				"endpoint": "{{host}}/api",
			},
		},
		{
			Name: "Dev",
			Base: "Base",
			Variables: map[string]string{
				"host": "dev.example.com",
			},
		},
	}

	dev := &envs[1]
	effective := dev.GetEffectiveVariables(envs)
	ResolveVariableReferences(effective)

	if effective["host"] != "dev.example.com" {
		t.Errorf("host = %q, want %q", effective["host"], "dev.example.com")
	}
	// endpoint references host, which is overridden in Dev
	if effective["endpoint"] != "dev.example.com/api" {
		t.Errorf("endpoint = %q, want %q", effective["endpoint"], "dev.example.com/api")
	}
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
