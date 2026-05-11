package cmd

import (
	"bytes"
	"fmt"
	"net/http"
	netURL "net/url"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/gdamore/tcell/v2"

	"github.com/petitorium/petitorium/config"
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
	case http.MethodGet:
		colorHex = config.C.MethodColors.GET
	case http.MethodPost:
		colorHex = config.C.MethodColors.POST
	case http.MethodPut:
		colorHex = config.C.MethodColors.PUT
	case http.MethodPatch:
		colorHex = config.C.MethodColors.PATCH
	case http.MethodDelete:
		colorHex = config.C.MethodColors.DELETE
	case http.MethodOptions:
		colorHex = config.C.MethodColors.OPTIONS
	case http.MethodHead:
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
	case http.MethodGet:
		return "GET  "
	case http.MethodPost:
		return "POST "
	case http.MethodPut:
		return "PUT  "
	case http.MethodDelete:
		return "DEL  "
	case http.MethodPatch:
		return "PAT  "
	case http.MethodHead:
		return "HEAD "
	case http.MethodOptions:
		return "OPT  "
	case http.MethodTrace:
		return "TRC  "
	case http.MethodConnect:
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
	case http.MethodGet:
		return "GET "
	case http.MethodPost:
		return http.MethodPost
	case http.MethodPut:
		return "PUT "
	case http.MethodDelete:
		return "DEL "
	case http.MethodPatch:
		return "PAT "
	case http.MethodHead:
		return http.MethodHead
	case http.MethodOptions:
		return "OPT "
	case http.MethodTrace:
		return "TRC "
	case http.MethodConnect:
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

// FormatBodyContentWithVariables formats body content with environment variable highlighting
func FormatBodyContentWithVariables(content string) string {
	if content == "" {
		return ""
	}

	// First, highlight environment variables
	content = highlightEnvironmentVariables(content)

	// Then apply syntax highlighting
	return formatBodyContent(content)
}

// highlightEnvironmentVariables highlights {{variable}} patterns with special background colors
func highlightEnvironmentVariables(content string) string {
	variableRegex := regexp.MustCompile(`\{\{[^}]+\}\}`)

	// Find all variable positions
	matches := variableRegex.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return content
	}

	// Build result with proper spacing
	var result strings.Builder
	lastEnd := 0

	for i, match := range matches {
		start, end := match[0], match[1]

		// Add text before this variable
		result.WriteString(content[lastEnd:start])

		// Extract variable name (remove {{ and }})
		varName := content[start+2 : end-2]

		// Render variable with background color (same as URL component)
		result.WriteString(fmt.Sprintf("[%s:%s:-]%s[-:-:-]",
			config.C.Theme.DropdownFocusedBackground,
			config.C.Theme.BorderFocusColor,
			varName))

		// Add space only if next character is another variable (no text between)
		if i < len(matches)-1 && end == matches[i+1][0] {
			result.WriteString(" ")
		}

		lastEnd = end
	}

	// Add remaining text after last variable
	result.WriteString(content[lastEnd:])

	return result.String()
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

		// Only escape brackets in text that could be misinterpreted as tview formatting
		// JSON structural brackets and content don't need escaping
		// text = strings.ReplaceAll(text, "[", "[[")
		// text = strings.ReplaceAll(text, "]", "]]")

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

	// JSON brackets don't need escaping for tview as they're not valid color commands
	// result = strings.ReplaceAll(result, "[", "[[")
	// result = strings.ReplaceAll(result, "]", "]]")

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

// getSupportedUnifiedThemes returns a list of themes that work well for unified theming
func getSupportedUnifiedThemes() []string {
	return []string{
		"tokyonight-night",
		"github-dark",
		"dracula",
		"monokai",
		"solarized-dark",
		"nord",
		"one-dark",
		"vim",
		"xcode-dark",
		"gruvbox",
		"catppuccin-mocha",
		"doom-one",
		"evergarden",
		"modus-vivendi",
	}
}

// tokenTypeToChroma converts a string token type name to chroma.TokenType
func tokenTypeToChroma(tokenName string) chroma.TokenType {
	switch strings.ToLower(tokenName) {
	case "background":
		return chroma.Background
	case "text", "foreground":
		return chroma.Text
	case "comment":
		return chroma.Comment
	case "keyword":
		return chroma.Keyword
	case "string":
		return chroma.String
	case "number":
		return chroma.Number
	case "name":
		return chroma.Name
	case "operator":
		return chroma.Operator
	case "punctuation":
		return chroma.Punctuation
	case "literal":
		return chroma.Literal
	case "error":
		return chroma.Error
	default:
		return chroma.Text // Default fallback
	}
}

// extractThemeColors extracts UI colors from a Chroma syntax highlighting theme
func extractThemeColors(themeName string) (background, foreground, border, borderFocus, title, selectionBackground, activeTab, buttonSelected, dropdownFocused string) {
	// Use the ThemeManager to get unified theme colors
	tm := GetThemeManager()
	theme, err := tm.GetTheme(themeName)
	if err != nil {
		// Fallback to default colors if theme not found
		return "#102529", "#e4e4e4", "#95CEDA", "#FF9F77", "#EBEBEB", "#1B4248", "#FF9F77", "#FFD700", "#636DA6"
	}

	return theme.UIColors.Background,
		theme.UIColors.Foreground,
		theme.UIColors.Border,
		theme.UIColors.BorderFocus,
		theme.UIColors.Title,
		theme.UIColors.Selection,
		theme.UIColors.ActiveTab,
		theme.UIColors.ButtonSelected,
		theme.UIColors.DropdownFocused
}

// applyFallback applies a fallback color if the input color is invalid
func applyFallback(color, fallback string) string {
	if color == "#ffffff" || color == "" {
		return fallback
	}
	return color
}

// isLightColor determines if a hex color is light (for theme detection)
func isLightColor(hexColor string) bool {
	if len(hexColor) < 7 {
		return false
	}

	// Convert hex to RGB
	r, g, b := hexToRGB(hexColor)
	if r == -1 || g == -1 || b == -1 {
		return false
	}

	// Calculate luminance (perceived brightness)
	luminance := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 255.0

	// Consider light if luminance > 0.5
	return luminance > 0.5
}

// hexToRGB converts hex color to RGB values
func hexToRGB(hex string) (int, int, int) {
	if len(hex) < 7 || hex[0] != '#' {
		return -1, -1, -1
	}

	r, err1 := strconv.ParseInt(hex[1:3], 16, 32)
	g, err2 := strconv.ParseInt(hex[3:5], 16, 32)
	b, err3 := strconv.ParseInt(hex[5:7], 16, 32)

	if err1 != nil || err2 != nil || err3 != nil {
		return -1, -1, -1
	}

	return int(r), int(g), int(b)
}

// openInExternalEditor opens content in an external editor and returns the modified content
// ext is the file extension (e.g., "json", "txt") used for the temporary file
func openInExternalEditor(content string, ext string) (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
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

	tmpDir := os.TempDir()
	pattern := fmt.Sprintf("petitorium-*.%s", ext)
	tmpFile, err := os.CreateTemp(tmpDir, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", fmt.Errorf("failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("editor exited with error: %v", err)
	}

	modifiedContent, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", fmt.Errorf("failed to read modified content: %v", err)
	}

	return string(modifiedContent), nil
}

// openInFxFunc is used for testing to mock fx calls
var openInFxFunc = openInFx

// openInFx opens JSON content in fx for interactive viewing
func openInFx(content string) error {
	// Check if fx is available
	if _, err := exec.LookPath("fx"); err != nil {
		return fmt.Errorf("fx not found. Please install fx: https://github.com/antonmedv/fx")
	}

	// Create a temporary file
	tmpDir := os.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "petitorium-response-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write content to temporary file
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temporary file: %v", err)
	}
	tmpFile.Close()

	// Open in fx
	cmd := exec.Command("fx", tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// generateCurlCommand generates a curl command string from HTTP request components
func generateCurlCommand(method, urlStr string, headers map[string]string, body, contentType string, queryParams map[string]string) string {
	var cmd strings.Builder
	cmd.WriteString("curl")

	// Add method if not GET
	if method != http.MethodGet {
		cmd.WriteString(" -X ")
		cmd.WriteString(method)
	}

	// Add headers
	for key, value := range headers {
		// Skip Content-Type for multipart requests as -F sets it automatically
		if contentType == "Multipart" && (key == "Content-Type" || key == "content-type") {
			continue
		}
		cmd.WriteString(" -H '")
		cmd.WriteString(key)
		cmd.WriteString(": ")
		cmd.WriteString(value)
		cmd.WriteString("'")
	}

	// Add body if present
	if body != "" {
		if contentType == "Multipart" {
			// Parse multipart fields and add as -F options
			fields := strings.Split(body, "&")
			for _, field := range fields {
				field = strings.TrimSpace(field)
				if field == "" {
					continue
				}
				parts := strings.SplitN(field, "=", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					if strings.HasPrefix(value, "file:") {
						filePath := strings.TrimPrefix(value, "file:")
						cmd.WriteString(" -F '")
						cmd.WriteString(name)
						cmd.WriteString("=@")
						cmd.WriteString(filePath)
						cmd.WriteString("'")
					} else {
						cmd.WriteString(" -F '")
						cmd.WriteString(name)
						cmd.WriteString("=")
						cmd.WriteString(value)
						cmd.WriteString("'")
					}
				}
			}
		} else {
			// Escape single quotes in body by replacing ' with '\''
			escapedBody := strings.ReplaceAll(body, "'", "'\\''")
			cmd.WriteString(" -d '")
			cmd.WriteString(escapedBody)
			cmd.WriteString("'")
		}
	}

	// Build final URL with query params
	finalURL := urlStr
	if len(queryParams) > 0 {
		u, err := netURL.Parse(urlStr)
		if err == nil {
			q := u.Query()
			for key, value := range queryParams {
				q.Set(key, value)
			}
			u.RawQuery = q.Encode()
			finalURL = u.String()
		}
	}

	// Add URL (must be last)
	cmd.WriteString(" '")
	cmd.WriteString(finalURL)
	cmd.WriteString("'")

	return cmd.String()
}

// substituteVariables replaces {{variable}} placeholders with values from the environment variables map
func substituteVariables(text string, variables map[string]string) string {
	if variables == nil {
		return text
	}

	result := text
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// substituteVariablesInHeaders replaces {{variable}} placeholders in header values
func substituteVariablesInHeaders(headers map[string]string, variables map[string]string) map[string]string {
	if variables == nil {
		return headers
	}

	result := make(map[string]string)
	for key, value := range headers {
		result[key] = substituteVariables(value, variables)
	}
	return result
}

// StripTags removes tview color tags from a string
func StripTags(text string) string {
	var result strings.Builder
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '[' {
			if i+1 < len(runes) && runes[i+1] == '[' {
				result.WriteRune('[')
				i++
				continue
			}
			// Skip until ]
			for i < len(runes) && runes[i] != ']' {
				i++
			}
			continue
		}
		result.WriteRune(runes[i])
	}
	return result.String()
}

// TruncateTaggedString truncates a string with color tags while preserving tags and accounting for their zero-width
func TruncateTaggedString(text string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}

	stripped := StripTags(text)
	if len(stripped) <= maxWidth {
		return text
	}

	// We need space for ellipsis
	if maxWidth <= 3 {
		if len(stripped) > maxWidth {
			return stripped[:maxWidth]
		}
		return stripped
	}

	var result strings.Builder
	visibleCount := 0
	inTag := false
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '[' {
			// Check for escaped [[
			if i+1 < len(runes) && runes[i+1] == '[' {
				if visibleCount < maxWidth-3 {
					result.WriteRune('[')
					result.WriteRune('[')
					visibleCount++
				}
				i++
				continue
			}
			inTag = true
			result.WriteRune(r)
			continue
		}

		if inTag {
			result.WriteRune(r)
			if r == ']' {
				inTag = false
			}
			continue
		}

		// Visible character
		if visibleCount < maxWidth-3 {
			result.WriteRune(r)
			visibleCount++
		} else {
			result.WriteString("...")
			// Close any potential color tags
			result.WriteString("[-:-:-]")
			break
		}
	}

	return result.String()
}
