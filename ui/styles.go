package ui

import (
	"fmt"
	"strconv"
	"strings"

	"festivus/config"

	"github.com/charmbracelet/lipgloss"
)

// UseTrueColor controls whether hex colors use true color (24-bit) or
// fall back to the nearest 256-color. Set to false for older terminals.
var UseTrueColor = true

// ColorToANSIFg converts a theme color string to an ANSI foreground escape sequence
// Supports: "0"-"255" for indexed colors, "#RGB" or "#RRGGBB" for hex colors
// Hex colors use true color if UseTrueColor is true, otherwise nearest 256-color
func ColorToANSIFg(color string) string {
	if strings.HasPrefix(color, "#") {
		r, g, b := parseHexColor(color)
		if UseTrueColor {
			return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
		}
		// Fall back to nearest 256-color
		return fmt.Sprintf("\033[38;5;%dm", rgbTo256Color(r, g, b))
	}
	n, err := strconv.Atoi(color)
	if err != nil {
		return "\033[37m" // Default to white on error
	}
	if n < 16 {
		// Standard colors: use traditional codes for better compatibility
		if n < 8 {
			return fmt.Sprintf("\033[%dm", 30+n)
		}
		return fmt.Sprintf("\033[%dm", 90+(n-8))
	}
	return fmt.Sprintf("\033[38;5;%dm", n)
}

// ColorToANSIBg converts a theme color string to an ANSI background escape sequence
func ColorToANSIBg(color string) string {
	if strings.HasPrefix(color, "#") {
		r, g, b := parseHexColor(color)
		if UseTrueColor {
			return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
		}
		// Fall back to nearest 256-color
		return fmt.Sprintf("\033[48;5;%dm", rgbTo256Color(r, g, b))
	}
	n, err := strconv.Atoi(color)
	if err != nil {
		return "\033[40m" // Default to black on error
	}
	if n < 16 {
		// Standard colors: use traditional codes for better compatibility
		if n < 8 {
			return fmt.Sprintf("\033[%dm", 40+n)
		}
		return fmt.Sprintf("\033[%dm", 100+(n-8))
	}
	return fmt.Sprintf("\033[48;5;%dm", n)
}

// rgbTo256Color converts RGB values to the nearest 256-color palette index
func rgbTo256Color(r, g, b int) int {
	// Check if it's close to a grayscale color
	if isGrayscale(r, g, b) {
		return rgbToGrayscale(r, g, b)
	}
	// Convert to 6x6x6 color cube (colors 16-231)
	return 16 + 36*rgbTo6(r) + 6*rgbTo6(g) + rgbTo6(b)
}

// rgbTo6 converts an 8-bit color value to a 6-level value (0-5)
// The 6x6x6 cube uses values: 0, 95, 135, 175, 215, 255
func rgbTo6(v int) int {
	if v < 48 {
		return 0
	} else if v < 115 {
		return 1
	} else if v < 155 {
		return 2
	} else if v < 195 {
		return 3
	} else if v < 235 {
		return 4
	}
	return 5
}

// isGrayscale checks if RGB values are close enough to be grayscale
func isGrayscale(r, g, b int) bool {
	// If all components are within 10 of each other, treat as grayscale
	max := r
	min := r
	if g > max {
		max = g
	}
	if b > max {
		max = b
	}
	if g < min {
		min = g
	}
	if b < min {
		min = b
	}
	return (max - min) < 20
}

// rgbToGrayscale converts RGB to nearest grayscale in 232-255 range
func rgbToGrayscale(r, g, b int) int {
	// Average the values
	gray := (r + g + b) / 3
	// Grayscale colors 232-255 represent 24 shades
	// Values: 8, 18, 28, 38, ..., 238 (step of 10)
	if gray < 4 {
		return 16 // Use black from color cube
	}
	if gray > 243 {
		return 231 // Use white from color cube
	}
	// Map to 232-255 range
	return 232 + (gray-8)/10
}

// ColorToANSI returns combined fg+bg ANSI sequence
func ColorToANSI(fg, bg string) string {
	return ColorToANSIBg(bg) + ColorToANSIFg(fg)
}

// parseHexColor parses #RGB or #RRGGBB to r, g, b values
func parseHexColor(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		// #RGB -> #RRGGBB
		r, _ := strconv.ParseInt(string(hex[0])+string(hex[0]), 16, 32)
		g, _ := strconv.ParseInt(string(hex[1])+string(hex[1]), 16, 32)
		b, _ := strconv.ParseInt(string(hex[2])+string(hex[2]), 16, 32)
		return int(r), int(g), int(b)
	}
	if len(hex) == 6 {
		r, _ := strconv.ParseInt(hex[0:2], 16, 32)
		g, _ := strconv.ParseInt(hex[2:4], 16, 32)
		b, _ := strconv.ParseInt(hex[4:6], 16, 32)
		return int(r), int(g), int(b)
	}
	return 255, 255, 255 // Default to white on error
}

// Styles contains all the styles used in the editor
type Styles struct {
	// The theme these styles were generated from
	Theme config.Theme

	// Menu bar styles
	MenuBar            lipgloss.Style
	MenuItem           lipgloss.Style
	MenuItemActive     lipgloss.Style
	MenuDropdown       lipgloss.Style
	MenuOption         lipgloss.Style
	MenuOptionActive   lipgloss.Style
	MenuOptionDisabled lipgloss.Style

	// Status bar styles
	StatusBar      lipgloss.Style
	StatusModified lipgloss.Style

	// Editor styles
	Editor           lipgloss.Style
	LineNumber       lipgloss.Style
	LineNumberActive lipgloss.Style
	Selection        lipgloss.Style
	Cursor           lipgloss.Style

	// Dialog styles
	DialogBox         lipgloss.Style
	DialogTitle       lipgloss.Style
	DialogText        lipgloss.Style
	DialogButton      lipgloss.Style
	DialogButtonFocus lipgloss.Style
	DialogInput       lipgloss.Style
	DialogList        lipgloss.Style
	DialogListItem    lipgloss.Style
	DialogListActive  lipgloss.Style

	// General styles
	Subtle lipgloss.Style
	Error  lipgloss.Style
}

// NewStyles creates a Styles configuration from a theme
func NewStyles(theme config.Theme) Styles {
	ui := theme.UI

	return Styles{
		Theme: theme,

		// Menu bar
		MenuBar: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuBg)).
			Foreground(lipgloss.Color(ui.MenuFg)).
			Bold(true),

		MenuItem: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuBg)).
			Foreground(lipgloss.Color(ui.MenuFg)).
			Padding(0, 2),

		MenuItemActive: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuHighlightBg)).
			Foreground(lipgloss.Color(ui.MenuHighlightFg)).
			Bold(true).
			Padding(0, 2),

		// Dropdown menu
		MenuDropdown: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuBg)).
			Foreground(lipgloss.Color(ui.MenuFg)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.MenuFg)),

		MenuOption: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuBg)).
			Foreground(lipgloss.Color(ui.MenuFg)).
			Padding(0, 1),

		MenuOptionActive: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuHighlightBg)).
			Foreground(lipgloss.Color(ui.MenuHighlightFg)).
			Padding(0, 1),

		MenuOptionDisabled: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.MenuBg)).
			Foreground(lipgloss.Color(ui.DisabledFg)).
			Padding(0, 1),

		// Status bar
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.StatusBg)).
			Foreground(lipgloss.Color(ui.StatusFg)),

		StatusModified: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.StatusAccent)).
			Background(lipgloss.Color(ui.StatusBg)).
			Bold(true),

		// Editor - use terminal defaults
		Editor: lipgloss.NewStyle(),

		LineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.LineNumber)).
			PaddingRight(1),

		LineNumberActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.LineNumberActive)).
			Bold(true).
			PaddingRight(1),

		// Selection
		Selection: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.SelectionBg)).
			Foreground(lipgloss.Color(ui.SelectionFg)),

		Cursor: lipgloss.NewStyle().
			Reverse(true),

		// Dialog styles
		DialogBox: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogBg)).
			Foreground(lipgloss.Color(ui.DialogFg)).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color(ui.DialogBorder)).
			Padding(1, 2),

		DialogTitle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.DialogTitle)).
			Bold(true),

		DialogText: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.DialogFg)),

		DialogButton: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogBg)).
			Foreground(lipgloss.Color(ui.DialogFg)).
			Padding(0, 2),

		DialogButtonFocus: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogButton)).
			Foreground(lipgloss.Color(ui.DialogButtonFg)).
			Bold(true).
			Padding(0, 2),

		DialogInput: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogFg)).
			Foreground(lipgloss.Color(ui.DialogBg)),

		DialogList: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogBg)).
			Foreground(lipgloss.Color(ui.DialogFg)),

		DialogListItem: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogBg)).
			Foreground(lipgloss.Color(ui.DialogFg)).
			Padding(0, 1),

		DialogListActive: lipgloss.NewStyle().
			Background(lipgloss.Color(ui.DialogButton)).
			Foreground(lipgloss.Color(ui.DialogButtonFg)).
			Padding(0, 1),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.DisabledFg)),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color(ui.ErrorFg)).
			Bold(true),
	}
}

// DefaultStyles returns the default style configuration (DOS EDIT theme)
func DefaultStyles() Styles {
	return NewStyles(config.DefaultTheme())
}
