package workspace

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mitchellh/go-homedir"
	"gopkg.in/yaml.v3"

	"github.com/petitorium/petitorium/config"
)

func getWorkspaceFilePath() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(home, ".config", "petitorium")
	return filepath.Join(configDir, "workspace.yaml"), nil
}

// writeFileAtomic writes data to path atomically: it writes to a temp file in
// the same directory, fsyncs it, then renames it over the target. This prevents
// a torn/partial file (and thus a corrupt workspace) if the process is
// interrupted mid-write. The rename is atomic on POSIX filesystems.
func writeFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	success := false
	defer func() {
		if !success {
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, perm); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	success = true
	return nil
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

	manager := &WorkspaceManager{
		CurrentWorkspace: "Default",
		Workspaces: []WorkspaceMetadata{
			{
				Name:        defaultWorkspace.Name,
				Description: defaultWorkspace.Description,
				CreatedAt:   defaultWorkspace.CreatedAt,
				UpdatedAt:   defaultWorkspace.UpdatedAt,
			},
		},
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
		Workspaces: []WorkspaceMetadata{
			{
				Name:        defaultWorkspace.Name,
				Description: defaultWorkspace.Description,
				CreatedAt:   defaultWorkspace.CreatedAt,
				UpdatedAt:   defaultWorkspace.UpdatedAt,
			},
		},
	}

	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

	return SaveWorkspace(defaultWorkspace)
}

func createDefaultWorkspace() *Workspace {
	now := time.Now()
	return &Workspace{
		Name:                "Default",
		Description:         "Default workspace created automatically",
		CreatedAt:           now,
		UpdatedAt:           now,
		SelectedEnvironment: "",
		Collections: []Collection{
			{
				Name: "Example Requests",
				Requests: []Request{
					{
						Name:        "Health Check",
						Method:      http.MethodGet,
						URL:         "https://httpbin.org/status/200",
						ContentType: "JSON",
					},
					{
						Name:        "Echo",
						Method:      http.MethodPost,
						URL:         "https://httpbin.org/post",
						Body:        `{"message": "Hello World"}`,
						ContentType: "JSON",
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

// migrateWorkspaceContentTypes sets default content types for requests that don't have them
// and migrates legacy tag formats.  It returns true if any data was modified.
func migrateWorkspaceContentTypes(workspace *Workspace) bool {
	migrateCollectionsContentTypes(workspace.Collections)
	trimResponseHistory(workspace.Collections)
	return migrateOldCommandRunnerTags(workspace.Collections)
}

// ensureIDs assigns a stable UUID to every Request and Collection that lacks
// one. It returns true if any ID was assigned so callers can persist the
// change. This migrates workspaces created before stable IDs existed; without
// unique IDs, move/delete/find operations can't reliably distinguish items
// that share a name.
func ensureIDs(ws *Workspace) bool {
	if ws == nil {
		return false
	}
	return ensureCollectionIDs(&ws.Collections)
}

func ensureCollectionIDs(collections *[]Collection) bool {
	modified := false
	for i := range *collections {
		col := &(*collections)[i]
		if col.ID == "" {
			col.ID = uuid.NewString()
			modified = true
		}
		for j := range col.Requests {
			if col.Requests[j].ID == "" {
				col.Requests[j].ID = uuid.NewString()
				modified = true
			}
		}
		if ensureCollectionIDs(&col.Collections) {
			modified = true
		}
	}
	return modified
}

// migrateCollectionsContentTypes recursively migrates content types for all collections
func migrateCollectionsContentTypes(collections []Collection) {
	for i := range collections {
		// Migrate requests in this collection
		for j := range collections[i].Requests {
			if collections[i].Requests[j].ContentType == "" {
				// Default to JSON for backward compatibility
				collections[i].Requests[j].ContentType = "JSON"
			}
		}
		// Recursively migrate nested collections
		migrateCollectionsContentTypes(collections[i].Collections)
	}
}

// migrateOldCommandRunnerTags converts legacy {{command-runner ...}} tags to the
// new namespace format {{command-runner:run ...}} across all requests.
func migrateOldCommandRunnerTags(collections []Collection) bool {
	modified := false
	oldPrefix := "{{command-runner "
	newPrefix := "{{command-runner:run "

	for i := range collections {
		for j := range collections[i].Requests {
			req := &collections[i].Requests[j]
			if strings.Contains(req.URL, oldPrefix) {
				req.URL = strings.ReplaceAll(req.URL, oldPrefix, newPrefix)
				modified = true
			}
			if strings.Contains(req.Body, oldPrefix) {
				req.Body = strings.ReplaceAll(req.Body, oldPrefix, newPrefix)
				modified = true
			}
			for k, e := range req.Headers {
				if strings.Contains(e.Value, oldPrefix) {
					req.Headers[k] = Entry{
						Value:   strings.ReplaceAll(e.Value, oldPrefix, newPrefix),
						Enabled: e.Enabled,
					}
					modified = true
				}
			}
			for k, e := range req.QueryParams {
				if strings.Contains(e.Value, oldPrefix) {
					req.QueryParams[k] = Entry{
						Value:   strings.ReplaceAll(e.Value, oldPrefix, newPrefix),
						Enabled: e.Enabled,
					}
					modified = true
				}
			}
		}
		if migrateOldCommandRunnerTags(collections[i].Collections) {
			modified = true
		}
	}
	return modified
}

const maxResponseBodySize = 1024 * 1024

func isBinaryResponse(body string, headers map[string][]string) bool {
	if headers == nil {
		return false
	}
	ct, ok := headers["Content-Type"]
	if !ok || len(ct) == 0 {
		return false
	}
	contentType := strings.ToLower(strings.TrimSpace(ct[0]))
	binaryPrefixes := []string{
		"image/",
		"audio/",
		"video/",
		"application/pdf",
		"application/zip",
		"application/gzip",
		"application/x-tar",
		"application/x-rar-compressed",
		"application/octet-stream",
		"application/msword",
		"application/vnd.ms-excel",
		"application/vnd.ms-powerpoint",
		"application/vnd.openxmlformats-officedocument",
	}
	for _, prefix := range binaryPrefixes {
		if strings.HasPrefix(contentType, prefix) {
			return true
		}
	}
	return false
}

func shouldSkipResponseBody(body string, headers map[string][]string) bool {
	return len(body) > maxResponseBodySize || isBinaryResponse(body, headers)
}

// trimResponseHistory trims response history to the configured limit
func trimResponseHistory(collections []Collection) {
	maxHistory := config.C.MaxResponseHistory

	var trimCollection func(cols []Collection)
	trimCollection = func(cols []Collection) {
		for i := range cols {
			for j := range cols[i].Requests {
				req := &cols[i].Requests[j]
				if maxHistory > 0 && len(req.ResponseHistory) > maxHistory {
					req.ResponseHistory = req.ResponseHistory[len(req.ResponseHistory)-maxHistory:]
				}
				for k := range req.ResponseHistory {
					resp := &req.ResponseHistory[k]
					if shouldSkipResponseBody(resp.Body, resp.Headers) {
						resp.Body = fmt.Sprintf("[Response body skipped - %d bytes]", len(resp.Body))
					}
				}
			}
			trimCollection(cols[i].Collections)
		}
	}

	trimCollection(collections)
}

// LoadWorkspaceByName loads a specific workspace by name
func LoadWorkspaceByName(name string) (*Workspace, error) {
	workspaceDir := getWorkspaceDir(name)
	workspacePath := filepath.Join(workspaceDir, "workspace.yaml")

	// Try to load workspace metadata first
	var workspace *Workspace
	if workspaceData, err := os.ReadFile(workspacePath); err == nil {
		if err := yaml.Unmarshal(workspaceData, &workspace); err == nil {
			if workspace.Name != name {
				workspace.Name = name
				if updatedData, err := yaml.Marshal(workspace); err == nil {
					os.WriteFile(workspacePath, updatedData, 0o644)
				}
			}
			// Successfully loaded workspace metadata
			// Load expansion state
			if err := LoadExpansionState(&workspace.Collections); err != nil {
				// Not critical
			}
			// Migrate: assign stable IDs to any items that lack them, default content
			// types, and legacy tag formats.
			modified := ensureIDs(workspace)
			if migrateWorkspaceContentTypes(workspace) {
				modified = true
			}
			if modified {
				if updatedData, err := yaml.Marshal(workspace); err == nil {
					os.WriteFile(workspacePath, updatedData, 0o644)
				}
				collectionsPath := filepath.Join(workspaceDir, "collections.yaml")
				if collectionsData, err := yaml.Marshal(workspace.Collections); err == nil {
					os.WriteFile(collectionsPath, collectionsData, 0o644)
				}
			}
			return workspace, nil
		}
	}

	// Fallback to old format: load collections and environments separately
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

	// Load environments - use default environments since workspace.yaml should contain them
	defaultEnvs := []Environment{
		{
			Name: "Base",
			Variables: map[string]string{
				"base_url": "https://api.example.com",
			},
		},
	}

	workspace = &Workspace{
		Name:                name,
		Collections:         collections,
		Environments:        defaultEnvs,
		SelectedEnvironment: "",
	}

	// Load expansion state
	if err := LoadExpansionState(&workspace.Collections); err != nil {
		// Not critical
	}

	// Migrate existing requests to have stable IDs, default content types and tag formats
	modified := ensureIDs(workspace)
	if migrateWorkspaceContentTypes(workspace) {
		modified = true
	}
	if modified {
		if updatedData, err := yaml.Marshal(workspace); err == nil {
			os.WriteFile(workspacePath, updatedData, 0o644)
		}
		if collectionsData, err := yaml.Marshal(workspace.Collections); err == nil {
			os.WriteFile(collectionsPath, collectionsData, 0o644)
		}
	}

	return workspace, nil
}

func SaveWorkspaceManager(manager *WorkspaceManager) error {
	path, err := getWorkspaceManagerFilePath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(manager)
	if err != nil {
		return err
	}

	return writeFileAtomic(path, data, 0o644)
}

func SaveWorkspace(workspace *Workspace) error {
	// Safety net: guarantee every item has a stable ID before persisting, so the
	// on-disk format is always consistent even if a constructor forgot to set one.
	ensureIDs(workspace)

	workspaceDir := getWorkspaceDir(workspace.Name)
	if err := os.MkdirAll(workspaceDir, 0o755); err != nil {
		return err
	}

	// Save workspace metadata
	workspacePath := filepath.Join(workspaceDir, "workspace.yaml")
	workspace.UpdatedAt = time.Now()
	workspaceData, err := yaml.Marshal(workspace)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(workspacePath, workspaceData, 0o644); err != nil {
		return err
	}

	// Save collections
	collectionsPath := filepath.Join(workspaceDir, "collections.yaml")
	collectionsData, err := yaml.Marshal(workspace.Collections)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(collectionsPath, collectionsData, 0o644); err != nil {
		return err
	}

	// Save expansion state
	if err := SaveExpansionState(&workspace.Collections); err != nil {
		// Not critical
	}

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
// merging with base environment variables if specified.
// {{key}} cross-references among the merged values are NOT resolved here;
// callers should run ResolveVariableReferences afterwards when needed.
func (e *Environment) GetEffectiveVariables(environments []Environment) map[string]string {
	return e.getRawVariables(environments)
}

// getRawVariables merges this environment's variables with its base
// environments without resolving {{key}} placeholders.
func (e *Environment) getRawVariables(environments []Environment) map[string]string {
	raw := make(map[string]string)

	if e.Base != "" {
		for _, env := range environments {
			if env.Name == e.Base {
				baseVars := env.getRawVariables(environments)
				for k, v := range baseVars {
					raw[k] = v
				}
				break
			}
		}
	}

	for k, v := range e.Variables {
		raw[k] = v
	}

	return raw
}

// ResolveVariableReferences performs iterative in-place substitution of
// {{key}} placeholders using other values in the same map.
// It runs at most len(vars) passes, stopping early when no changes occur.
func ResolveVariableReferences(vars map[string]string) {
	n := len(vars)
	if n == 0 {
		return
	}

	keys := make([]string, 0, n)
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i := 0; i < n; i++ {
		changed := false
		next := make(map[string]string, n)
		for _, key := range keys {
			resolved := vars[key]
			for _, otherKey := range keys {
				placeholder := "{{" + otherKey + "}}"
				resolved = strings.ReplaceAll(resolved, placeholder, vars[otherKey])
			}
			next[key] = resolved
			if resolved != vars[key] {
				changed = true
			}
		}
		for _, key := range keys {
			vars[key] = next[key]
		}
		if !changed {
			break
		}
	}
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

	data, err := yaml.Marshal(environments)
	if err != nil {
		return err
	}

	return writeFileAtomic(path, data, 0o644)
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

	return writeFileAtomic(path, data, 0o644)
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

	var sourceMetadata *WorkspaceMetadata
	for _, ws := range manager.Workspaces {
		if ws.Name == sourceName {
			sourceMetadata = &ws
			break
		}
	}

	if sourceMetadata == nil {
		return nil, fmt.Errorf("source workspace %s not found", sourceName)
	}

	sourceWorkspace, err := LoadWorkspaceByName(sourceName)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	duplicateWorkspace := &Workspace{
		Name:                targetName,
		Description:         fmt.Sprintf("Copy of %s", sourceMetadata.Description),
		Collections:         make([]Collection, len(sourceWorkspace.Collections)),
		Environments:        make([]Environment, len(sourceWorkspace.Environments)),
		SelectedEnvironment: sourceWorkspace.SelectedEnvironment,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	for i, col := range sourceWorkspace.Collections {
		duplicateWorkspace.Collections[i] = Collection{
			Name:     col.Name,
			Requests: make([]Request, len(col.Requests)),
			Expanded: col.Expanded,
		}
		copy(duplicateWorkspace.Collections[i].Requests, col.Requests)
	}

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

	manager.Workspaces = append(manager.Workspaces, WorkspaceMetadata{
		Name:        duplicateWorkspace.Name,
		Description: duplicateWorkspace.Description,
		CreatedAt:   duplicateWorkspace.CreatedAt,
		UpdatedAt:   duplicateWorkspace.UpdatedAt,
	})

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

	now := time.Now()
	newWorkspace := &Workspace{
		Name:                name,
		Description:         "",
		CookieJar:           CookieJar{Cookies: []Cookie{}},
		CreatedAt:           now,
		UpdatedAt:           now,
		SelectedEnvironment: "",
		Collections: []Collection{
			{
				Name: "Example Requests",
				Requests: []Request{
					{
						Name:        "Health Check",
						Method:      http.MethodGet,
						URL:         "https://httpbin.org/status/200",
						ContentType: "JSON",
					},
					{
						Name:        "Echo",
						Method:      http.MethodPost,
						URL:         "https://httpbin.org/post",
						Body:        `{"message": "Hello World"}`,
						ContentType: "JSON",
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

	manager.Workspaces = append(manager.Workspaces, WorkspaceMetadata{
		Name:        newWorkspace.Name,
		Description: newWorkspace.Description,
		CreatedAt:   newWorkspace.CreatedAt,
		UpdatedAt:   newWorkspace.UpdatedAt,
	})

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

func SwitchWorkspace(name string) error {
	manager, err := LoadWorkspaceManager()
	if err != nil {
		return err
	}

	found := false
	for _, ws := range manager.Workspaces {
		if ws.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("workspace %s not found", name)
	}

	manager.CurrentWorkspace = name

	if err := SaveWorkspaceManager(manager); err != nil {
		return err
	}

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

	// Load the workspace and update its internal Name to match the new name
	workspacePath := filepath.Join(getWorkspaceDir(oldName), "workspace.yaml")
	var ws Workspace
	if data, err := os.ReadFile(workspacePath); err == nil {
		if err := yaml.Unmarshal(data, &ws); err == nil {
			ws.Name = newName
			ws.UpdatedAt = time.Now()
			updatedData, err := yaml.Marshal(&ws)
			if err == nil {
				os.WriteFile(workspacePath, updatedData, 0o644)
			}
		}
	}

	// Rename the workspace directory
	oldDir := getWorkspaceDir(oldName)
	newDir := getWorkspaceDir(newName)

	if err := os.Rename(oldDir, newDir); err != nil {
		return fmt.Errorf("failed to rename workspace directory: %w", err)
	}

	return nil
}
