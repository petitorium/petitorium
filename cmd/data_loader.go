package cmd

import (
	"fmt"

	"github.com/hbarral/petitorium/workspace"
)

// LoadData loads all necessary data for the application
func LoadData() (*[]workspace.Collection, *DataManager, *[]workspace.Environment, error) {
	collectionsData, err := workspace.LoadCollections()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load collections: %w", err)
	}

	// Create data manager for collections operations
	dataManager := NewDataManager(collectionsData)

	environmentsData, err := workspace.LoadEnvironments()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to load environments: %w", err)
	}

	return &collectionsData, dataManager, &environmentsData, nil
}
