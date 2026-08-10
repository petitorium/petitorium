// Package cmd provides plugin management commands.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium-plugin-sdk/types"
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
		pluginDir := filepath.Join(config.AppDir, "plugins", "available")

		entries, err := os.ReadDir(pluginDir)
		if err != nil {
			fmt.Printf("Error reading plugin directory: %v\n", err)
			return
		}

		fmt.Println("Available plugins:")
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if strings.HasSuffix(name, ".sha256") {
				continue
			}
			status := "disabled"
			for _, enabled := range config.C.Plugins.Enabled {
				if enabled == name {
					status = "enabled"
					break
				}
			}
			fmt.Printf("  - %s (%s)\n", name, status)
		}
	},
}

var enableCmd = &cobra.Command{
	Use:   "enable <name>",
	Short: "Enable a plugin",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		pluginDir := filepath.Join(config.AppDir, "plugins", "available")
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
		pluginDir := filepath.Join(config.AppDir, "plugins", "available")
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

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for plugins in the registry",
	Run: func(cmd *cobra.Command, args []string) {
		query := ""
		if len(args) > 0 {
			query = strings.ToLower(args[0])
		}

		client := plugins.NewRegistryClient(config.C.Plugins.RegistryURL)
		available, err := client.ListPlugins()
		if err != nil {
			fmt.Printf("Error fetching plugins: %v\n", err)
			return
		}

		fmt.Println("Search results:")
		for _, p := range available {
			if query == "" || strings.Contains(strings.ToLower(p.Name), query) || strings.Contains(strings.ToLower(p.Description), query) {
				fmt.Printf("  - %s (%s): %s\n", p.Name, plugins.LatestVersion(p), p.Description)
			}
		}
	},
}

var installVersionFlag string

var installCmd = &cobra.Command{
	Use:   "install <name>",
	Short: "Install a plugin from the registry",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := plugins.NewRegistryClient(config.C.Plugins.RegistryURL)
		available, err := client.ListPlugins()
		if err != nil {
			fmt.Printf("Error fetching plugins: %v\n", err)
			return
		}

		var target *types.RegistryPlugin
		for _, p := range available {
			if p.Name == name {
				target = &p
				break
			}
		}

		if target == nil {
			fmt.Printf("Plugin %s not found in registry\n", name)
			return
		}

		version := installVersionFlag
		if version == "" {
			version = plugins.LatestVersion(*target)
		} else {
			found := false
			for _, v := range plugins.AvailableVersions(*target) {
				if v == version {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Version %s not available for plugin %s\n", version, name)
				return
			}
		}

		pluginDir := filepath.Join(config.AppDir, "plugins", "available")
		pm := plugins.NewPluginManager(&config.C.Plugins, pluginDir)

		fmt.Printf("Installing %s (%s)...\n", target.Name, version)
		if err := pm.InstallPluginVersion(*target, version); err != nil {
			fmt.Printf("Error installing plugin: %v\n", err)
			return
		}

		if err := config.SaveConfig(&config.C); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Printf("Plugin %s installed successfully\n", name)
	},
}

func init() {
	pluginsCmd.AddCommand(listCmd)
	pluginsCmd.AddCommand(enableCmd)
	pluginsCmd.AddCommand(disableCmd)
	pluginsCmd.AddCommand(configCmd)
	pluginsCmd.AddCommand(searchCmd)
	installCmd.Flags().StringVarP(&installVersionFlag, "version", "v", "", "specific version to install (default: latest)")
	pluginsCmd.AddCommand(installCmd)
	rootCmd.AddCommand(pluginsCmd)
}
