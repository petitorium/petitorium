package cmd

import (
	"fmt"
	"sort"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/spf13/cobra"
)

// themesCmd represents the themes command
var themesCmd = &cobra.Command{
	Use:   "themes",
	Short: "List available syntax highlighting themes",
	Long: `List all available Chroma syntax highlighting themes that can be used
in the configuration file's syntaxTheme setting.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available syntax highlighting themes:")
		fmt.Println("=====================================")

		// Get all available themes
		allThemes := styles.Names()
		sort.Strings(allThemes)

		// Get supported unified theming themes
		tm := GetThemeManager()
		supportedThemes := tm.GetAvailableThemes()
		supportedMap := make(map[string]bool)
		for _, theme := range supportedThemes {
			supportedMap[theme] = true
		}

		// Current theme
		currentTheme := GetThemeManager().GetCurrentTheme()

		fmt.Println("\nUnified Theming Supported Themes (★):")
		fmt.Println("=====================================")

		// Print supported themes first
		for i, theme := range supportedThemes {
			marker := "  "
			if theme == currentTheme {
				marker = "* " // Mark current theme
			}

			fmt.Printf("%s★ %-18s", marker, theme)

			// Print 3 themes per line
			if (i+1)%3 == 0 {
				fmt.Println()
			}
		}

		if len(supportedThemes)%3 != 0 {
			fmt.Println()
		}

		fmt.Println("\nAll Available Themes:")
		fmt.Println("====================")

		// Print all themes
		for i, theme := range allThemes {
			marker := "  "
			if theme == currentTheme {
				marker = "* " // Mark current theme
			}
			if supportedMap[theme] {
				marker += "★" // Mark supported themes
			} else {
				marker += " " // Space for alignment
			}

			fmt.Printf("%s%-19s", marker, theme)

			// Print 3 themes per line
			if (i+1)%3 == 0 {
				fmt.Println()
			}
		}

		if len(allThemes)%3 != 0 {
			fmt.Println()
		}

		fmt.Printf("\nCurrent theme: %s\n", currentTheme)
		fmt.Println("\nTo change theme, update 'syntaxTheme' in your config file.")
		fmt.Println("★ = Supported for unified theming (applies theme colors to entire UI)")
		fmt.Println("Popular themes: github-dark, dracula, monokai, solarized-dark, nord, one-dark")
	},
}

func init() {
	rootCmd.AddCommand(themesCmd)
}
