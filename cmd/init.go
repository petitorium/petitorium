package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/petitorium/petitorium/config"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default configuration file",
	Long: `Creates a default configuration file at ~/.config/petitorium/config.yaml.
You can customize your theme and settings by editing this file.`,
	Run: func(cmd *cobra.Command, args []string) {
		path, err := config.InitConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("If the file already exists, you can edit it at: %s\n", path)
			os.Exit(1)
		}
		fmt.Printf("✅ Configuration file created at: %s\n", path)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
