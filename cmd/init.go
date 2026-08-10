package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium/config"
)

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Create a default configuration file",
	Long: `Creates a default configuration file.
You can customize your theme and settings by editing this file.

Default location: <data-dir>/config.yaml
Use --config to specify a different file path.
Use --data-dir to specify a different data directory.

A positional [path] argument can be used to create the config in a specific
location. If the path is an existing directory, config.yaml is created inside
it. If the path does not exist, it is treated as a file path.

Examples:
  petitorium init                    # <data-dir>/config.yaml
  petitorium init .                  # ./config.yaml
  petitorium init ./pet.yaml         # ./pet.yaml
  petitorium init ~/my-configs/      # ~/my-configs/config.yaml`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configPath := config.ConfigFile

		if len(args) > 0 {
			resolved, err := resolveInitPath(args[0])
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			configPath = resolved
		}

		path, err := config.InitConfig(configPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("If the file already exists, you can edit it at: %s\n", path)
			os.Exit(1)
		}
		fmt.Printf("✅ Configuration file created at: %s\n", path)
	},
}

// resolveInitPath resolves a positional [path] argument for the init command.
// If the path is an existing directory, "config.yaml" is appended to it.
// Otherwise the path is treated as a file path.
func resolveInitPath(p string) (string, error) {
	expanded, err := homedir.Expand(p)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", p, err)
	}
	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", err
	}
	if info, err := os.Stat(abs); err == nil && info.IsDir() {
		return filepath.Join(abs, "config.yaml"), nil
	}
	return abs, nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
