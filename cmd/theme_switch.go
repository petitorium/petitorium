package cmd

import (
	"fmt"
	"os"

	"github.com/petitorium/petitorium/config"
	"github.com/spf13/cobra"
)

// themeSwitchCmd represents the theme switch command
var themeSwitchCmd = &cobra.Command{
	Use:   "switch [theme-name]",
	Short: "Switch to a different unified theme",
	Long: `Switch to a different unified theme that applies syntax highlighting colors to the entire UI.
This will update your configuration file with the new theme settings.

Use 'petitorium themes' to see available themes.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		themeName := args[0]

		// Get theme manager
		tm := GetThemeManager()

		// Check if theme exists
		if _, err := tm.GetTheme(themeName); err != nil {
			fmt.Printf("Error: Theme '%s' not found.\n", themeName)
			fmt.Println("Use 'petitorium themes' to see available themes.")
			os.Exit(1)
		}

		// Apply the theme
		if err := tm.ApplyTheme(themeName); err != nil {
			fmt.Printf("Error applying theme: %v\n", err)
			os.Exit(1)
		}

		// Save the updated config
		if err := config.SaveConfig(&config.C); err != nil {
			fmt.Printf("Error saving configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Theme switched to: %s\n", themeName)
		fmt.Println("The new theme will be applied when you restart Petitorium.")
	},
}

func init() {
	themesCmd.AddCommand(themeSwitchCmd)
}
