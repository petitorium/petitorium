package cmd

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// themesCmd represents the themes command
var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available Petitorium themes",
	Long: `List all available themes that can be used
in the configuration file's syntaxTheme setting.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available Petitorium themes:")
		fmt.Println("=====================================")

		// Get supported themes
		tm := GetThemeManager()
		supportedThemes := tm.GetAvailableThemes()
		sort.Strings(supportedThemes)

		// Current theme
		currentTheme := tm.GetCurrentTheme()

		// Print supported themes
		for i, theme := range supportedThemes {
			marker := "  "
			if theme == currentTheme {
				marker = "* " // Mark current theme
			}

			fmt.Printf("%s%-18s", marker, theme)

			// Print 3 themes per line
			if (i+1)%3 == 0 {
				fmt.Println()
			}
		}

		if len(supportedThemes)%3 != 0 {
			fmt.Println()
		}

		fmt.Printf("\nCurrent theme: %s\n", currentTheme)
		fmt.Println("\nTo change theme, update 'syntaxTheme' in your config file, or use 'petitorium switch <theme>'.")
	},
}

func init() {
	rootCmd.AddCommand(themesCmd)
}
