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

	environmentsData, err := workspace.LoadEnvironments()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load environments: %w", err)
	}

	return workspaceData, dataManager, &environmentsData, nil
}
