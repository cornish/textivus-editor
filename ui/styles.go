package ui

import "github.com/charmbracelet/lipgloss"

// Styles contains all the styles used in the editor
type Styles struct {
	// Menu bar styles
	MenuBar          lipgloss.Style
	MenuItem         lipgloss.Style
	MenuItemActive   lipgloss.Style
	MenuDropdown     lipgloss.Style
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

	// General styles
	Subtle lipgloss.Style
	Error  lipgloss.Style
}

// DefaultStyles returns the default style configuration
func DefaultStyles() Styles {
	return Styles{
		// Menu bar - dark blue background (classic DOS EDIT style)
		MenuBar: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).  // Dark blue
			Foreground(lipgloss.Color("15")). // White text
			Bold(true),

		MenuItem: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).  // Dark blue
			Foreground(lipgloss.Color("15")). // White text
			Padding(0, 2),

		MenuItemActive: lipgloss.NewStyle().
			Background(lipgloss.Color("6")).  // Cyan background
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true).
			Padding(0, 2),

		// Dropdown menu - dark blue background (classic DOS style)
		MenuDropdown: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).   // Dark blue
			Foreground(lipgloss.Color("15")).  // White
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("15")),

		MenuOption: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).  // Dark blue
			Foreground(lipgloss.Color("15")). // White
			Padding(0, 1),

		MenuOptionActive: lipgloss.NewStyle().
			Background(lipgloss.Color("6")).  // Cyan
			Foreground(lipgloss.Color("0")).  // Black text
			Bold(true).
			Padding(0, 1),

		MenuOptionDisabled: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).  // Dark blue
			Foreground(lipgloss.Color("8")).  // Gray
			Padding(0, 1),

		// Status bar - dark blue background (classic DOS style)
		StatusBar: lipgloss.NewStyle().
			Background(lipgloss.Color("4")).  // Dark blue
			Foreground(lipgloss.Color("15")), // White text

		StatusModified: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")). // Bright cyan
			Background(lipgloss.Color("4")).  // Keep dark blue background
			Bold(true),

		// Editor - use terminal defaults
		Editor: lipgloss.NewStyle(),

		LineNumber: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			PaddingRight(1),

		LineNumberActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true).
			PaddingRight(1),

		// Selection - cyan background (classic DOS style)
		Selection: lipgloss.NewStyle().
			Background(lipgloss.Color("6")).  // Cyan
			Foreground(lipgloss.Color("0")),  // Black text

		Cursor: lipgloss.NewStyle().
			Reverse(true),

		Subtle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),

		Error: lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true),
	}
}
