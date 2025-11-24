// Package cmd provides the root command for the Petitorium CLI application.
// It includes the main TUI setup, UI components, and event handling.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
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
	workspaceData, dataManager, environmentsData, err := LoadData()
	if err != nil {
		panic(fmt.Sprintf("Failed to load data: %v", err))
	}

	// Setup plugin manager
	home, _ := homedir.Dir()
	pluginDir := filepath.Join(home, ".config", "petitorium", "plugins", "available")
	os.MkdirAll(pluginDir, 0755)
	pm := plugins.NewPluginManager(&config.C.Plugins, pluginDir)
	if err := pm.LoadPlugins(); err != nil {
		fmt.Printf("Warning: Failed to load plugins: %v\n", err)
		fmt.Printf("Plugin directory: %s\n", pluginDir)
		fmt.Printf("Make sure plugins are built and copied to the plugin directory.\n")
	}

	// Setup UI
	ui, err := SetupUI(workspaceData, dataManager, environmentsData)
	if err != nil {
		panic(fmt.Sprintf("Failed to setup UI: %v", err))
	}
	ui.PluginManager = pm

	// Setup event handlers
	SetupEventHandlers(ui)

	// Run app
	if err := RunApp(ui); err != nil {
		panic(err)
	}
}
