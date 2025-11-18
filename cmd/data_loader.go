package cmd

import (
	"fmt"

	"github.com/hbarral/petitorium/workspace"
)

// LoadData loads all necessary data for the application
func LoadData() (*workspace.Workspace, *DataManager, *[]workspace.Environment, error) {
	workspaceData, err := workspace.LoadWorkspace()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load workspace: %w", err)
	}

	// Create data manager for workspace operations
	dataManager := NewDataManager(workspaceData)

	// Use workspace-specific environments instead of global ones
	environmentsData := workspaceData.Environments

	return workspaceData, dataManager, &environmentsData, nil
}
