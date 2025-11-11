package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"
)

func getWorkspaceFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "workspace.yaml"), nil
}

func getWorkspaceManagerFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "workspaces.yaml"), nil
}

func getWorkspaceDir(name string) string {
	home, _ := homedir.Dir()
	return filepath.Join(home, ".config", "petitorium", "workspaces", name)
}

func migrateFromOldFormat() error {
	oldPath, err := getWorkspaceFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		// No old file, create default
		return createDefaultWorkspaceManager()
	}

	// Load old workspace
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return err
	}

	var oldWorkspace Workspace
	if err := yaml.Unmarshal(data, &oldWorkspace); err != nil {
		return err
	}

	// Create default workspace with old data
	defaultWorkspace := oldWorkspace
	defaultWorkspace.Name = "Default"
	now := time.Now()
	defaultWorkspace.CreatedAt = now
	defaultWorkspace.UpdatedAt = now

	// Create workspace manager
	manager := &WorkspaceManager{
		CurrentWorkspace: "Default",
		Workspaces:       []Workspace{defaultWorkspace},
	}

	// Save in new format
	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

	// Save workspace data
	if err := SaveWorkspace(&defaultWorkspace); err != nil {
		return err
	}

	// Backup old file
	backupPath := oldPath + ".backup"
	if err := os.Rename(oldPath, backupPath); err != nil {
		return err
	}

	return nil
}

func createDefaultWorkspaceManager() error {
	defaultWorkspace := createDefaultWorkspace()
	manager := &WorkspaceManager{
		CurrentWorkspace: "Default",
		Workspaces:       []Workspace{*defaultWorkspace},
	}

	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

	return SaveWorkspace(defaultWorkspace)
}

func createDefaultWorkspace() *Workspace {
	now := time.Now()
	return &Workspace{
		Name:        "Default",
		Description: "Default workspace created automatically",
		CreatedAt:   now,
		UpdatedAt:   now,
		Collections: []Collection{
			{
				Name: "Example Requests",
				Requests: []Request{
					{
						Name:   "Health Check",
						Method: "GET",
						URL:    "https://httpbin.org/status/200",
					},
					{
						Name:   "Echo",
						Method: "POST",
						URL:    "https://httpbin.org/post",
						Body:   `{"message": "Hello World"}`,
					},
				},
			},
		},
		Environments: []Environment{
			{
				Name: "Base",
				Variables: map[string]string{
					"base_url": "https://api.example.com",
				},
			},
		},
	}
}

// LoadWorkspaceManager loads the workspace manager with all workspaces
func LoadWorkspaceManager() (*WorkspaceManager, error) {
	path, err := getWorkspaceManagerFilePath()
	if err != nil {
		return nil, err
	}

	// Check if new format exists
	if _, err = os.Stat(path); os.IsNotExist(err) {
		// Try to migrate from old format
		if err := migrateFromOldFormat(); err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var manager WorkspaceManager
	if err := yaml.Unmarshal(data, &manager); err != nil {
		return nil, err
	}

	return &manager, nil
}

// LoadWorkspace loads the current workspace
func LoadWorkspace() (*Workspace, error) {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		return nil, err
	}

	if manager.CurrentWorkspace == "" {
		return nil, fmt.Errorf("no current workspace set")
	}

	return LoadWorkspaceByName(manager.CurrentWorkspace)
}

// LoadWorkspaceByName loads a specific workspace by name
func LoadWorkspaceByName(name string) (*Workspace, error) {
	workspaceDir := getWorkspaceDir(name)
	collectionsPath := filepath.Join(workspaceDir, "collections.yaml")

	if _, err := os.Stat(collectionsPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("workspace %s not found", name)
	}

	data, err := os.ReadFile(collectionsPath)
	if err != nil {
		return nil, err
	}

	var collections []Collection
	if err := yaml.Unmarshal(data, &collections); err != nil {
		return nil, err
	}

	// Load environments
	environmentsPath := filepath.Join(workspaceDir, "environments.yaml")
	var environments []Environment
	if envData, err := os.ReadFile(environmentsPath); err == nil {
		if err := yaml.Unmarshal(envData, &environments); err != nil {
			// Not critical, continue with empty environments
		}
	}

	workspace := &Workspace{
		Name:         name,
		Collections:  collections,
		Environments: environments,
	}

	// Load expansion state
	if err := LoadExpansionState(&workspace.Collections); err != nil {
		// Not critical
	}

	return workspace, nil
}

func SaveWorkspaceManager(manager *WorkspaceManager) error {
	path, err := getWorkspaceManagerFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(manager)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func SaveWorkspace(workspace *Workspace) error {
	workspaceDir := getWorkspaceDir(workspace.Name)
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return err
	}

	// Save collections
	collectionsPath := filepath.Join(workspaceDir, "collections.yaml")
	collectionsData, err := yaml.Marshal(workspace.Collections)
	if err != nil {
		return err
	}
	if err := os.WriteFile(collectionsPath, collectionsData, 0o644); err != nil {
		return err
	}

	// Save environments
	environmentsPath := filepath.Join(workspaceDir, "environments.yaml")
	environmentsData, err := yaml.Marshal(workspace.Environments)
	if err != nil {
		return err
	}
	if err := os.WriteFile(environmentsPath, environmentsData, 0o644); err != nil {
		return err
	}

	// Save expansion state
	if err := SaveExpansionState(&workspace.Collections); err != nil {
		// Not critical
	}

	// Update workspace metadata
	workspace.UpdatedAt = time.Now()

	return nil
}

// Legacy SaveWorkspace for backward compatibility
func SaveWorkspaceLegacy(workspace *Workspace) error {
	path, err := getWorkspaceFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(workspace)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// Helper function to get expansion state file path
func getEnvironmentsFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "environments.yaml"), nil
}

func createDefaultEnvironments() []Environment {
	return []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"base_url": "https://api.example.com",
			},
		},
	}
}

// GetEffectiveVariables returns the effective variables for an environment,
// merging with base environment variables if specified
func (e *Environment) GetEffectiveVariables(environments []Environment) map[string]string {
	effective := make(map[string]string)

	// First, inherit from base environment if specified
	if e.Base != "" {
		for _, env := range environments {
			if env.Name == e.Base {
				// Recursively get base variables (to handle multiple levels of inheritance)
				baseVars := env.GetEffectiveVariables(environments)
				for k, v := range baseVars {
					effective[k] = v
				}
				break
			}
		}
	}

	// Then override with this environment's variables
	for k, v := range e.Variables {
		effective[k] = v
	}

	return effective
}

func LoadEnvironments() ([]Environment, error) {
	path, err := getEnvironmentsFilePath()
	if err != nil {
		return nil, err
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		defaultEnvs := createDefaultEnvironments()
		if err = SaveEnvironments(defaultEnvs); err != nil {
			return nil, err
		}
		return defaultEnvs, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var environments []Environment
	if err := yaml.Unmarshal(data, &environments); err != nil {
		return nil, err
	}

	return environments, nil
}

func SaveEnvironments(environments []Environment) error {
	path, err := getEnvironmentsFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(environments)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func getExpansionStateFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "expansion_state.yaml"), nil
}

// SaveExpansionState saves the expansion state of collections
func SaveExpansionState(collections *[]Collection) error {
	path, err := getExpansionStateFilePath()
	if err != nil {
		return err
	}

	if err = os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	// Create a simple map for expansion state
	expansionState := make(map[string]bool)

	var collectExpansionState func(collections *[]Collection, prefix string)
	collectExpansionState = func(collections *[]Collection, prefix string) {
		for _, col := range *collections {
			fullName := prefix + col.Name
			expansionState[fullName] = col.Expanded
			if len(col.Collections) > 0 {
				collectExpansionState(&col.Collections, fullName+"/")
			}
		}
	}

	collectExpansionState(collections, "")

	data, err := yaml.Marshal(expansionState)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

// LoadExpansionState loads the expansion state and applies it to collections
func LoadExpansionState(collections *[]Collection) error {
	path, err := getExpansionStateFilePath()
	if err != nil {
		return err
	}

	if _, err = os.Stat(path); os.IsNotExist(err) {
		// No expansion state file exists yet, that's fine
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var expansionState map[string]bool
	if err := yaml.Unmarshal(data, &expansionState); err != nil {
		return err
	}

	var applyExpansionState func(collections *[]Collection, prefix string)
	applyExpansionState = func(collections *[]Collection, prefix string) {
		for i := range *collections {
			fullName := prefix + (*collections)[i].Name
			if expanded, exists := expansionState[fullName]; exists {
				(*collections)[i].Expanded = expanded
			}
			if len((*collections)[i].Collections) > 0 {
				applyExpansionState(&(*collections)[i].Collections, fullName+"/")
			}
		}
	}

	applyExpansionState(collections, "")
	return nil
}

// DeleteWorkspace deletes a workspace (with safety checks)
func DeleteWorkspace(name string) error {
	if name == "" {
		return fmt.Errorf("workspace name cannot be empty")
	}

	manager, err := LoadWorkspaceManager()
	if err != nil {
		return err
	}

	// Don't allow deleting the last workspace
	if len(manager.Workspaces) <= 1 {
		return fmt.Errorf("cannot delete the last workspace")
	}

	// Find the workspace
	workspaceIndex := -1
	for i, ws := range manager.Workspaces {
		if ws.Name == name {
			workspaceIndex = i
			break
		}
	}

	if workspaceIndex == -1 {
		return fmt.Errorf("workspace %s not found", name)
	}

	// Don't allow deleting the current workspace
	if manager.CurrentWorkspace == name {
		return fmt.Errorf("cannot delete the current workspace, switch to another workspace first")
	}

	// Remove from manager
	manager.Workspaces = append(manager.Workspaces[:workspaceIndex], manager.Workspaces[workspaceIndex+1:]...)

	// Save manager
	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

	// Remove the workspace directory
	workspaceDir := getWorkspaceDir(name)
	if err := os.RemoveAll(workspaceDir); err != nil {
		return fmt.Errorf("failed to remove workspace directory: %w", err)
	}

	return nil
}

// DuplicateWorkspace creates a copy of an existing workspace
func DuplicateWorkspace(sourceName, targetName string) (*Workspace, error) {
	if sourceName == "" || targetName == "" {
		return nil, fmt.Errorf("workspace names cannot be empty")
	}

	if sourceName == targetName {
		return nil, fmt.Errorf("target name must be different from source name")
	}

	manager, err := LoadWorkspaceManager()
	if err != nil {
		return nil, err
	}

	// Check if target name already exists
	for _, ws := range manager.Workspaces {
		if ws.Name == targetName {
			return nil, fmt.Errorf("workspace %s already exists", targetName)
		}
	}

	// Find the source workspace
	var sourceWorkspace *Workspace
	for _, ws := range manager.Workspaces {
		if ws.Name == sourceName {
			sourceWorkspace = &ws
			break
		}
	}

	if sourceWorkspace == nil {
		return nil, fmt.Errorf("source workspace %s not found", sourceName)
	}

	// Create duplicate workspace
	now := time.Now()
	duplicateWorkspace := &Workspace{
		Name:         targetName,
		Description:  fmt.Sprintf("Copy of %s", sourceWorkspace.Description),
		Collections:  make([]Collection, len(sourceWorkspace.Collections)),
		Environments: make([]Environment, len(sourceWorkspace.Environments)),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Deep copy collections
	for i, col := range sourceWorkspace.Collections {
		duplicateWorkspace.Collections[i] = Collection{
			Name:     col.Name,
			Requests: make([]Request, len(col.Requests)),
			Expanded: col.Expanded,
		}
		copy(duplicateWorkspace.Collections[i].Requests, col.Requests)
		// Note: Nested collections would need recursive copying, but keeping simple for now
	}

	// Deep copy environments
	for i, env := range sourceWorkspace.Environments {
		duplicateWorkspace.Environments[i] = Environment{
			Name:      env.Name,
			Base:      env.Base,
			Variables: make(map[string]string),
		}
		for k, v := range env.Variables {
			duplicateWorkspace.Environments[i].Variables[k] = v
		}
	}

	// Add to manager
	manager.Workspaces = append(manager.Workspaces, *duplicateWorkspace)

	// Save manager
	if err := SaveWorkspaceManager(manager); err != nil {
		return nil, err
	}

	// Save workspace data
	if err := SaveWorkspace(duplicateWorkspace); err != nil {
		return nil, err
	}

	return duplicateWorkspace, nil
}

func CreateWorkspace(name string) (*Workspace, error) {
	if name == "" {
		return nil, fmt.Errorf("workspace name cannot be empty")
	}

	manager, err := LoadWorkspaceManager()
	if err != nil {
		return nil, err
	}

	// Check if workspace already exists
	for _, ws := range manager.Workspaces {
		if ws.Name == name {
			return nil, fmt.Errorf("workspace %s already exists", name)
		}
	}

	// Create new workspace
	now := time.Now()
	newWorkspace := &Workspace{
		Name:        name,
		Description: "",
		CreatedAt:   now,
		UpdatedAt:   now,
		Collections: []Collection{
			{
				Name: "Example Requests",
				Requests: []Request{
					{
						Name:   "Health Check",
						Method: "GET",
						URL:    "https://httpbin.org/status/200",
					},
				},
			},
		},
		Environments: []Environment{
			{
				Name: "Base",
				Variables: map[string]string{
					"base_url": "https://api.example.com",
				},
			},
		},
	}

	// Add to manager
	manager.Workspaces = append(manager.Workspaces, *newWorkspace)

	// Save manager
	if err := SaveWorkspaceManager(manager); err != nil {
		return nil, err
	}

	// Save workspace data
	if err := SaveWorkspace(newWorkspace); err != nil {
		return nil, err
	}

	return newWorkspace, nil
}

// SwitchWorkspace switches to the specified workspace
func SwitchWorkspace(name string) error {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		fmt.Printf("LoadWorkspaceManager error: %v\n", err)
		return err
	}

	// Check if workspace exists
	found := false
	for _, ws := range manager.Workspaces {
		if ws.Name == name {
			found = true
			break
		}
	}

	if !found {
		fmt.Printf("Workspace '%s' not found in available workspaces\n", name)
		return fmt.Errorf("workspace %s not found", name)
	}

	fmt.Printf("Setting current workspace to: '%s'\n", name)
	manager.CurrentWorkspace = name

	if err := SaveWorkspaceManager(manager); err != nil {
		fmt.Printf("SaveWorkspaceManager error: %v\n", err)
		return err
	}

	fmt.Printf("SwitchWorkspace completed successfully\n")
	return nil
}

// ListWorkspaces returns a list of all workspace names
func ListWorkspaces() ([]string, error) {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(manager.Workspaces))
	for i, ws := range manager.Workspaces {
		names[i] = ws.Name
	}

	return names, nil
}

// RenameWorkspace renames a workspace
func RenameWorkspace(oldName, newName string) error {
	if oldName == "" || newName == "" {
		return fmt.Errorf("workspace names cannot be empty")
	}

	if oldName == newName {
		return fmt.Errorf("new name must be different from current name")
	}

	manager, err := LoadWorkspaceManager()
	if err != nil {
		return err
	}

	// Check if new name already exists
	for _, ws := range manager.Workspaces {
		if ws.Name == newName {
			return fmt.Errorf("workspace %s already exists", newName)
		}
	}

	// Find and update the workspace
	found := false
	for i, ws := range manager.Workspaces {
		if ws.Name == oldName {
			manager.Workspaces[i].Name = newName
			manager.Workspaces[i].UpdatedAt = time.Now()
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("workspace %s not found", oldName)
	}

	// Update current workspace if it was renamed
	if manager.CurrentWorkspace == oldName {
		manager.CurrentWorkspace = newName
	}

	// Save manager
	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

	// Rename the workspace directory
	oldDir := getWorkspaceDir(oldName)
	newDir := getWorkspaceDir(newName)

	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename workspace directory: %w", err)
	}

	return nil
}
