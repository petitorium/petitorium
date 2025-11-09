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

	// For now, return the environments from the workspace
	// TODO: Update this when workspace switching is implemented
	environmentsData := workspaceData.Environments

	return workspaceData, dataManager, &environmentsData, nil
}
