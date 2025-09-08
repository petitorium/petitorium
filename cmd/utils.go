package cmd

import (
	"fmt"
	"strconv"
	"strings"

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
