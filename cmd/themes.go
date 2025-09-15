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

		// Current theme
		currentTheme := getSyntaxTheme()

		// Print themes
		for i, theme := range allThemes {
			marker := "  "
			if theme == currentTheme {
				marker = "* " // Mark current theme
			}

			fmt.Printf("%s%-20s", marker, theme)

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
		fmt.Println("Popular themes: github-dark, dracula, monokai, solarized-dark, nord, one-dark")
	},
}

func init() {
	rootCmd.AddCommand(themesCmd)
}
