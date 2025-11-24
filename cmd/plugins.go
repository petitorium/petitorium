// Package cmd provides plugin management commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium/config"
	"github.com/petitorium/petitorium/plugins"
)

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Manage plugins",
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available plugins",
	Run: func(cmd *cobra.Command, args []string) {
		home, _ := homedir.Dir()
		pluginDir := filepath.Join(home, ".config", "petitorium", "plugins", "available")

		entries, err := os.ReadDir(pluginDir)
		if err != nil {
			fmt.Printf("Error reading plugin directory: %v\n", err)
			return
		}

		fmt.Println("Available plugins:")
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".so") {
				name := strings.TrimSuffix(entry.Name(), ".so")
				status := "disabled"
				for _, enabled := range config.C.Plugins.Enabled {
					if enabled == name {
						status = "enabled"
						break
					}
				}
				fmt.Printf("  - %s (%s)\n", name, status)
			}
		}
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		home, _ := homedir.Dir()
		pluginDir := filepath.Join(home, ".config", "petitorium", "plugins", "available")
		pm := plugins.NewPluginManager(&config.C.Plugins, pluginDir)

		if err := pm.EnablePlugin(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := config.SaveConfig(&config.C); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Plugin %s enabled\n", name)
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <name>",
	Short: "Disable a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		home, _ := homedir.Dir()
		pluginDir := filepath.Join(home, ".config", "petitorium", "plugins", "available")
		pm := plugins.NewPluginManager(&config.C.Plugins, pluginDir)

		if err := pm.DisablePlugin(name); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if err := config.SaveConfig(&config.C); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Plugin %s disabled\n", name)
	},
}

var configCmd = &cobra.Command{
	Use:   "config <name>",
	Short: "Show config for a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		if cfg, exists := config.C.Plugins.Config[name]; exists {
			fmt.Printf("Config for %s: %v\n", name, cfg)
		} else {
			fmt.Printf("No config found for %s\n", name)
		}
	},
}

func init() {
	pluginsCmd.AddCommand(listCmd)
	pluginsCmd.AddCommand(enableCmd)
	pluginsCmd.AddCommand(disableCmd)
	pluginsCmd.AddCommand(configCmd)
	rootCmd.AddCommand(pluginsCmd)
}
