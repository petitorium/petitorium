package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"

	"github.com/hbarral/petitorium/config"
)

// hexToColor converts a hex color string to tcell.Color
func hexToColor(hexStr string) tcell.Color {
	hexStr = strings.TrimPrefix(hexStr, "#")
	if colorValue, err := strconv.ParseInt(hexStr, 16, 32); err == nil {
		return tcell.NewHexColor(int32(colorValue))
	}

	return tcell.ColorWhite
}

// strToRune converts the first character of a string to a rune
func strToRune(s string) rune {
	for _, r := range s {
		return r
	}
	return ' '
}

// getColoredMethod returns the method name with appropriate color formatting
func getColoredMethod(method string) string {
	var colorHex string

	switch method {
	case "GET":
		colorHex = config.C.MethodColors.GET
	case "POST":
		colorHex = config.C.MethodColors.POST
	case "PUT":
		colorHex = config.C.MethodColors.PUT
	case "PATCH":
		colorHex = config.C.MethodColors.PATCH
	case "DELETE":
		colorHex = config.C.MethodColors.DELETE
	case "OPTIONS":
		colorHex = config.C.MethodColors.OPTIONS
	case "HEAD":
		colorHex = config.C.MethodColors.HEAD
	default:
		colorHex = config.C.MethodColors.Default
	}

	// Convert hex color to tview color format [::b]...[-:-:-]
	return fmt.Sprintf("[%s]%s[-:-:-]", colorHex, getPaddedMethodName(method))
}

// getPaddedMethodName returns a padded method name for consistent alignment
func getPaddedMethodName(method string) string {
	switch method {
	case "GET":
		return "GET  "
	case "POST":
		return "POST "
	case "PUT":
		return "PUT  "
	case "DELETE":
		return "DEL  "
	case "PATCH":
		return "PAT  "
	case "HEAD":
		return "HEAD "
	case "OPTIONS":
		return "OPT  "
	case "TRACE":
		return "TRC  "
	case "CONNECT":
		return "CON  "
	default:
		// For unknown methods, take first 3 characters and pad to 5
		if len(method) >= 3 {
			return method[:3] + "  "
		}
		return method + strings.Repeat(" ", 5-len(method))
	}
}

// getShortMethodName returns a shortened version of the HTTP method
func getShortMethodName(method string) string {
	switch method {
	case "GET":
		return "GET "
	case "POST":
		return "POST"
	case "PUT":
		return "PUT "
	case "DELETE":
		return "DEL "
	case "PATCH":
		return "PAT "
	case "HEAD":
		return "HEAD"
	case "OPTIONS":
		return "OPT "
	case "TRACE":
		return "TRC "
	case "CONNECT":
		return "CON "
	default:
		// For unknown methods, take first 3 characters or pad to 4
		if len(method) >= 3 {
			return method[:3] + " "
		}
		return method + strings.Repeat(" ", 4-len(method))
	}
}

// padNameToMinLength pads a name with spaces to ensure it has at least minLength characters
func padNameToMinLength(name string, minLength int) string {
	if len(name) >= minLength {
		return name
	}
	return name + strings.Repeat(" ", minLength-len(name))
}

// getIconDisplayWidth calculates the display width of an icon string
// This handles common cases where icons might include spaces or combining characters
func getIconDisplayWidth(icon string) int {
	if icon == "" {
		return 0
	}

	// Count all runes including spaces - spaces are part of the display
	return len([]rune(icon))
}

// formatBodyContent formats body content with syntax highlighting using Chroma
func formatBodyContent(content string) string {
	if content == "" {
		return ""
	}

	content = strings.TrimSpace(content)

	// For JSON, use Chroma themes with tview colors
	if strings.HasPrefix(content, "{") || strings.HasPrefix(content, "[") {
		// DO NOT use json.MarshalIndent as it reorders keys alphabetically
		// Keep original JSON structure and just apply syntax highlighting
		formatted := formatWithChromaTheme(content, "json", getSyntaxTheme())
		return formatted
	}

	// For XML, use Chroma themes
	if strings.HasPrefix(content, "<?xml") || (strings.HasPrefix(content, "<") && !strings.Contains(content, "<html")) {
		formatted := formatWithChromaTheme(content, "xml", getSyntaxTheme())
		return formatted
	}

	// For other content types, detect and use appropriate Chroma lexer
	if strings.Contains(content, "package ") || strings.Contains(content, "import ") || strings.Contains(content, "func ") {
		// Go code
		formatted := formatWithChromaTheme(content, "go", getSyntaxTheme())
		return formatted
	} else if strings.Contains(content, "<html") || strings.Contains(content, "<!DOCTYPE") {
		// HTML
		formatted := formatWithChromaTheme(content, "html", getSyntaxTheme())
		return formatted
	} else if strings.Contains(content, "def ") || strings.Contains(content, "class ") {
		// Python
		formatted := formatWithChromaTheme(content, "python", getSyntaxTheme())
		return formatted
	}

	// Default: return as plain text
	return content
}

// isAlphaNumeric checks if a rune is alphanumeric
func isAlphaNumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// isDigit checks if a rune is a digit
func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

// chromaToTviewFormatter converts Chroma output to tview hex colors
type chromaToTviewFormatter struct {
	style *chroma.Style
}

// Format implements the chroma.Formatter interface
func (f *chromaToTviewFormatter) Format(w *bytes.Buffer, style *chroma.Style, iterator chroma.Iterator) error {
	if style == nil {
		style = f.style
	}

	for token := iterator(); token != chroma.EOF; token = iterator() {
		entry := style.Get(token.Type)

		// Get the text value
		text := token.String()

		// More robust escaping for tview
		text = strings.ReplaceAll(text, "[", "[[")
		text = strings.ReplaceAll(text, "]", "]]")

		// Apply color if available
		if !entry.IsZero() {
			if entry.Colour.IsSet() {
				hexColor := chromaColorToHex(entry.Colour)
				// Use more explicit color formatting
				w.WriteString(fmt.Sprintf("[%s:-:-]%s[-:-:-]", hexColor, text))
			} else {
				w.WriteString(text)
			}
		} else {
			w.WriteString(text)
		}
	}

	return nil
}

// chromaColorToHex converts a Chroma color to hex string for tview
func chromaColorToHex(color chroma.Colour) string {
	if !color.IsSet() {
		return "#ffffff"
	}

	r, g, b := color.Red(), color.Green(), color.Blue()
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

// formatWithChromaTheme formats content using Chroma themes but outputting tview colors
func formatWithChromaTheme(content string, lexerName string, themeName string) string {
	// Get lexer
	lexer := lexers.Get(lexerName)
	if lexer == nil {
		return formatJSONSimple(content) // fallback to simple formatting
	}

	// Get style
	style := styles.Get(themeName)
	if style == nil {
		return formatJSONSimple(content) // fallback to simple formatting
	}

	// Create our custom formatter
	formatter := &chromaToTviewFormatter{style: style}

	// Tokenize
	iterator, err := lexer.Tokenise(nil, content)
	if err != nil {
		return formatJSONSimple(content) // fallback to simple formatting
	}

	// Format
	var buf bytes.Buffer
	err = formatter.Format(&buf, style, iterator)
	if err != nil {
		return formatJSONSimple(content) // fallback to simple formatting
	}

	return buf.String()
}

// formatJSONSimple provides basic JSON highlighting without complex formatting
func formatJSONSimple(content string) string {
	// Simple JSON highlighting that preserves original order
	result := content

	// Escape any existing brackets first
	result = strings.ReplaceAll(result, "[", "[[")
	result = strings.ReplaceAll(result, "]", "]]")

	// Add basic JSON highlighting with minimal color tags
	result = strings.ReplaceAll(result, `"`, `[green]"[-]`)
	result = strings.ReplaceAll(result, `: true`, `: [yellow]true[-]`)
	result = strings.ReplaceAll(result, `: false`, `: [yellow]false[-]`)
	result = strings.ReplaceAll(result, `: null`, `: [yellow]null[-]`)

	return result
}

// getSyntaxTheme returns the current syntax highlighting theme
func getSyntaxTheme() string {
	// Get theme from config or use default
	if config.C.SyntaxTheme != "" {
		// Validate that the theme exists
		if styles.Get(config.C.SyntaxTheme) != nil {
			return config.C.SyntaxTheme
		}
	}

	// Default themes to try in order (popular themes with good colors)
	themes := []string{"github-dark", "dracula", "monokai", "solarized-dark", "nord", "one-dark", "vim", "github"}

	for _, theme := range themes {
		if styles.Get(theme) != nil {
			return theme
		}
	}

	// Fallback
	return "github"
}

// getAvailableThemes returns a list of all available Chroma themes
func getAvailableThemes() []string {
	var availableThemes []string

	// Popular themes that work well with syntax highlighting
	popularThemes := []string{
		"github", "github-dark", "dracula", "monokai", "solarized-dark", "solarized-light",
		"nord", "one-dark", "vim", "xcode", "xcode-dark", "native", "fruity", "emacs",
		"friendly", "colorful", "autumn", "murphy", "manni", "perldoc", "pastie",
		"borland", "trac", "default", "igor", "lovelace", "paraiso-dark", "paraiso-light",
		"rrt", "tango", "bw", "api", "material", "zenburn", "rainbow_dash", "algol",
		"algol_nu", "arduino", "average", "base16-snazzy", "coffee", "gruvbox",
		"gruvbox-light", "hrdark", "hr_high_contrast", "monokai-light", "monokailight",
		"murphy", "native", "onedark", "pygments", "rainbowdash", "solarized-dark256",
		"solarized-light", "swapoff", "trac", "vs", "witchhazel",
	}

	// Check which themes are actually available
	for _, theme := range popularThemes {
		if styles.Get(theme) != nil {
			availableThemes = append(availableThemes, theme)
		}
	}

	return availableThemes
}

// openInExternalEditor opens content in an external editor and returns the modified content
func openInExternalEditor(content string) (string, error) {
	// Get editor from environment variables, fallback to sensible defaults
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		// Try common editors in order of preference
		editors := []string{"nvim", "vim", "nano", "code", "subl", "gedit"}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				editor = e
				break
			}
		}
	}
	if editor == "" {
		return "", fmt.Errorf("no editor found. Please set EDITOR or VISUAL environment variable")
	}

	// Create a temporary file
	tmpDir := os.TempDir()
	tmpFile, err := ioutil.TempFile(tmpDir, "petitorium-body-*.json")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write content to temporary file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %v", err)
	}

	// Read back the content
	modifiedContent, err := ioutil.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read modified content: %v", err)
	}

	return string(modifiedContent), nil
}
