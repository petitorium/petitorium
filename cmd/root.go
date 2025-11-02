// Package cmd provides the root command for the Petitorium CLI application.
// It includes the main TUI setup, UI components, and event handling.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hbarral/petitorium/config"
)

var rootCmd = &cobra.Command{
	Use:   "petitorium",
	Short: "A powerful TUI for API interaction and testing.",
	Run:   runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(func() {
		if err := config.LoadConfig(); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	})
}

func runTUI(cmd *cobra.Command, args []string) {
	// Load data
	collectionsData, dataManager, environmentsData, err := LoadData()
	if err != nil {
		panic(fmt.Sprintf("Failed to load data: %v", err))
	}

	// Setup UI
	ui, err := SetupUI(collectionsData, dataManager, environmentsData)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup UI: %v", err))
	}

	// Setup event handlers
	SetupEventHandlers(ui)

	// Run app
	if err := RunApp(ui); err != nil {
		panic(err)
	}
}
