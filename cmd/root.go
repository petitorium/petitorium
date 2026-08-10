// Package cmd provides the root command for the Petitorium CLI application.
// It includes the main TUI setup, UI components, and event handling.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
	"github.com/petitorium/petitorium/version"
)

var configFileFlag string
var dataDirFlag string

var rootCmd = &cobra.Command{
	Use:     "petitorium",
	Short:   "A powerful Terminal API Testing Client.",
	Version: version.Version,
	Run:     runTUI,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFileFlag, "config", "c", "", "Path to config file (default: <data-dir>/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&dataDirFlag, "data-dir", "d", "", "Path to data directory (default: OS config dir/petitorium)")

	cobra.OnInitialize(func() {
		// Resolve data directory
		dir, err := config.ResolveAppDir(dataDirFlag)
		if err != nil {
			fmt.Printf("Error resolving data directory: %v\n", err)
			os.Exit(1)
		}
		config.AppDir = dir

		// Auto-create data dir when --data-dir points to a non-existent path
		if dataDirFlag != "" {
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				if err := os.MkdirAll(dir, 0o755); err != nil {
					fmt.Printf("Error creating data directory: %v\n", err)
					os.Exit(1)
				}
				fmt.Printf("Initialized data directory: %s\n", dir)
			}
		}

		// Resolve config file path (defaults to <AppDir>/config.yaml)
		cfgFile, err := config.ResolveConfigFile(configFileFlag)
		if err != nil {
			fmt.Printf("Error resolving config file: %v\n", err)
			os.Exit(1)
		}
		config.ConfigFile = cfgFile

		if err := config.LoadConfig(); err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}
	})
}

func runTUI(cmd *cobra.Command, args []string) {
	// Apply theme colors based on syntax theme
	if config.C.SyntaxTheme != "" {
		tm := GetThemeManager()
		if err := tm.ApplyTheme(config.C.SyntaxTheme); err != nil {
			fmt.Printf("Warning: Failed to apply theme '%s': %v\n", config.C.SyntaxTheme, err)
		}
	}

	// Load data
	workspaceData, dataManager, environmentsData, err := LoadData()
	if err != nil {
		panic(fmt.Sprintf("Failed to load data: %v", err))
	}

	// Setup plugin manager
	pluginDir := filepath.Join(config.AppDir, "plugins", "available")
	os.MkdirAll(pluginDir, 0755)
	pm := plugins.NewPluginManager(&config.C.Plugins, pluginDir)
	defer pm.Close()
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
