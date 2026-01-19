package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Theme holds complete color theme settings
// This is the format for theme TOML files in ~/.config/festivus/themes/
type Theme struct {
	// Metadata
	Name        string `toml:"name"`
	Description string `toml:"description"`
	Author      string `toml:"author"`

	// UI Colors
	UI UIColors `toml:"ui"`

	// Syntax highlighting colors
	Syntax SyntaxColors `toml:"syntax"`
}

// UIColors holds UI color settings
type UIColors struct {
	MenuBg           string `toml:"menu_bg"`
	MenuFg           string `toml:"menu_fg"`
	MenuHighlightBg  string `toml:"menu_highlight_bg"`
	MenuHighlightFg  string `toml:"menu_highlight_fg"`
	StatusBg         string `toml:"status_bg"`
	StatusFg         string `toml:"status_fg"`
	StatusAccent     string `toml:"status_accent"`
	SelectionBg      string `toml:"selection_bg"`
	SelectionFg      string `toml:"selection_fg"`
	LineNumber       string `toml:"line_number"`
	LineNumberActive string `toml:"line_number_active"`
	ErrorFg          string `toml:"error_fg"`
	DisabledFg       string `toml:"disabled_fg"`
	// Dialog colors
	DialogBg       string `toml:"dialog_bg"`
	DialogFg       string `toml:"dialog_fg"`
	DialogBorder   string `toml:"dialog_border"`
	DialogTitle    string `toml:"dialog_title"`
	DialogButton   string `toml:"dialog_button"`
	DialogButtonFg string `toml:"dialog_button_fg"`
}

// SyntaxColors holds syntax highlighting color settings
type SyntaxColors struct {
	Keyword  string `toml:"keyword"`
	String   string `toml:"string"`
	Comment  string `toml:"comment"`
	Number   string `toml:"number"`
	Operator string `toml:"operator"`
	Function string `toml:"function"`
	Type     string `toml:"type"`
}

// Built-in themes
var builtinThemes = map[string]Theme{
	"default": {
		Name:        "default",
		Description: "Classic DOS EDIT style - blue with cyan highlights",
		Author:      "Festivus",
		UI: UIColors{
			MenuBg:           "4",  // Dark blue
			MenuFg:           "15", // Bright white
			MenuHighlightBg:  "6",  // Cyan
			MenuHighlightFg:  "16", // True black
			StatusBg:         "4",  // Dark blue
			StatusFg:         "15", // Bright white
			StatusAccent:     "14", // Bright cyan
			SelectionBg:      "6",  // Cyan
			SelectionFg:      "0",  // Black
			LineNumber:       "8",  // Gray
			LineNumberActive: "3",  // Yellow
			ErrorFg:          "9",  // Bright red
			DisabledFg:       "8",  // Gray
			DialogBg:         "7",  // Light gray
			DialogFg:         "0",  // Black
			DialogBorder:     "0",  // Black
			DialogTitle:      "4",  // Blue
			DialogButton:     "2",  // Green
			DialogButtonFg:   "15", // White
		},
		Syntax: SyntaxColors{
			Keyword:  "14", // Bright cyan
			String:   "10", // Bright green
			Comment:  "8",  // Gray
			Number:   "11", // Bright yellow
			Operator: "13", // Bright magenta
			Function: "12", // Bright blue
			Type:     "11", // Bright yellow
		},
	},
	"dark": {
		Name:        "dark",
		Description: "Modern dark theme with muted colors",
		Author:      "Festivus",
		UI: UIColors{
			MenuBg:           "236", // Dark gray
			MenuFg:           "252", // Light gray
			MenuHighlightBg:  "24",  // Dark cyan
			MenuHighlightFg:  "15",  // Bright white
			StatusBg:         "236", // Dark gray
			StatusFg:         "252", // Light gray
			StatusAccent:     "43",  // Teal
			SelectionBg:      "24",  // Dark cyan
			SelectionFg:      "15",  // Bright white
			LineNumber:       "240", // Medium gray
			LineNumberActive: "250", // Lighter gray
			ErrorFg:          "203", // Soft red
			DisabledFg:       "240", // Medium gray
			DialogBg:         "238", // Darker gray
			DialogFg:         "252", // Light gray
			DialogBorder:     "245", // Medium gray
			DialogTitle:      "43",  // Teal
			DialogButton:     "24",  // Dark cyan
			DialogButtonFg:   "15",  // White
		},
		Syntax: SyntaxColors{
			Keyword:  "176", // Purple
			String:   "114", // Green
			Comment:  "245", // Gray
			Number:   "215", // Orange
			Operator: "80",  // Cyan
			Function: "75",  // Light blue
			Type:     "222", // Yellow
		},
	},
	"light": {
		Name:        "light",
		Description: "Light theme for bright environments",
		Author:      "Festivus",
		UI: UIColors{
			MenuBg:           "254", // Light gray
			MenuFg:           "235", // Dark gray
			MenuHighlightBg:  "32",  // Blue
			MenuHighlightFg:  "15",  // White
			StatusBg:         "254", // Light gray
			StatusFg:         "235", // Dark gray
			StatusAccent:     "26",  // Blue
			SelectionBg:      "153", // Light blue
			SelectionFg:      "0",   // Black
			LineNumber:       "249", // Medium gray
			LineNumberActive: "235", // Dark gray
			ErrorFg:          "160", // Red
			DisabledFg:       "249", // Medium gray
			DialogBg:         "255", // White
			DialogFg:         "235", // Dark gray
			DialogBorder:     "240", // Gray
			DialogTitle:      "26",  // Blue
			DialogButton:     "32",  // Blue
			DialogButtonFg:   "15",  // White
		},
		Syntax: SyntaxColors{
			Keyword:  "26",  // Blue
			String:   "28",  // Green
			Comment:  "245", // Gray
			Number:   "166", // Orange
			Operator: "90",  // Magenta
			Function: "26",  // Blue
			Type:     "30",  // Teal
		},
	},
	"monokai": {
		Name:        "monokai",
		Description: "Monokai-inspired dark theme",
		Author:      "Festivus",
		UI: UIColors{
			MenuBg:           "235", // Dark background
			MenuFg:           "231", // White
			MenuHighlightBg:  "208", // Orange
			MenuHighlightFg:  "16",  // Black
			StatusBg:         "235", // Dark background
			StatusFg:         "231", // White
			StatusAccent:     "208", // Orange
			SelectionBg:      "59",  // Gray
			SelectionFg:      "231", // White
			LineNumber:       "59",  // Gray
			LineNumberActive: "231", // White
			ErrorFg:          "197", // Pink-red
			DisabledFg:       "59",  // Gray
			DialogBg:         "237", // Slightly lighter bg
			DialogFg:         "231", // White
			DialogBorder:     "208", // Orange
			DialogTitle:      "208", // Orange
			DialogButton:     "64",  // Olive green
			DialogButtonFg:   "231", // White
		},
		Syntax: SyntaxColors{
			Keyword:  "197", // Pink-red
			String:   "186", // Yellow
			Comment:  "59",  // Gray
			Number:   "141", // Purple
			Operator: "197", // Pink-red
			Function: "81",  // Light blue
			Type:     "81",  // Light blue
		},
	},
}

// DefaultTheme returns the default DOS EDIT theme
func DefaultTheme() Theme {
	return builtinThemes["default"]
}

// LoadTheme loads a theme by name
// Checks user themes directory first, then falls back to built-in themes
func LoadTheme(name string) Theme {
	if name == "" {
		return DefaultTheme()
	}

	// Try loading from user themes directory
	theme, err := loadUserTheme(name)
	if err == nil {
		return theme
	}

	// Fall back to built-in theme
	if builtin, ok := builtinThemes[name]; ok {
		return builtin
	}

	// Default if not found
	return DefaultTheme()
}

// loadUserTheme attempts to load a theme from the user's themes directory
func loadUserTheme(name string) (Theme, error) {
	themesDir, err := ThemesDir()
	if err != nil {
		return Theme{}, err
	}

	themePath := filepath.Join(themesDir, name+".toml")
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return Theme{}, err
	}

	var theme Theme
	if _, err := toml.DecodeFile(themePath, &theme); err != nil {
		return Theme{}, err
	}

	// Merge with default theme to fill in any missing values
	return mergeWithDefault(theme), nil
}

// mergeWithDefault fills in any missing theme values with defaults
func mergeWithDefault(theme Theme) Theme {
	def := DefaultTheme()

	if theme.Name == "" {
		theme.Name = def.Name
	}

	// UI colors
	if theme.UI.MenuBg == "" {
		theme.UI.MenuBg = def.UI.MenuBg
	}
	if theme.UI.MenuFg == "" {
		theme.UI.MenuFg = def.UI.MenuFg
	}
	if theme.UI.MenuHighlightBg == "" {
		theme.UI.MenuHighlightBg = def.UI.MenuHighlightBg
	}
	if theme.UI.MenuHighlightFg == "" {
		theme.UI.MenuHighlightFg = def.UI.MenuHighlightFg
	}
	if theme.UI.StatusBg == "" {
		theme.UI.StatusBg = def.UI.StatusBg
	}
	if theme.UI.StatusFg == "" {
		theme.UI.StatusFg = def.UI.StatusFg
	}
	if theme.UI.StatusAccent == "" {
		theme.UI.StatusAccent = def.UI.StatusAccent
	}
	if theme.UI.SelectionBg == "" {
		theme.UI.SelectionBg = def.UI.SelectionBg
	}
	if theme.UI.SelectionFg == "" {
		theme.UI.SelectionFg = def.UI.SelectionFg
	}
	if theme.UI.LineNumber == "" {
		theme.UI.LineNumber = def.UI.LineNumber
	}
	if theme.UI.LineNumberActive == "" {
		theme.UI.LineNumberActive = def.UI.LineNumberActive
	}
	if theme.UI.ErrorFg == "" {
		theme.UI.ErrorFg = def.UI.ErrorFg
	}
	if theme.UI.DisabledFg == "" {
		theme.UI.DisabledFg = def.UI.DisabledFg
	}
	if theme.UI.DialogBg == "" {
		theme.UI.DialogBg = def.UI.DialogBg
	}
	if theme.UI.DialogFg == "" {
		theme.UI.DialogFg = def.UI.DialogFg
	}
	if theme.UI.DialogBorder == "" {
		theme.UI.DialogBorder = def.UI.DialogBorder
	}
	if theme.UI.DialogTitle == "" {
		theme.UI.DialogTitle = def.UI.DialogTitle
	}
	if theme.UI.DialogButton == "" {
		theme.UI.DialogButton = def.UI.DialogButton
	}
	if theme.UI.DialogButtonFg == "" {
		theme.UI.DialogButtonFg = def.UI.DialogButtonFg
	}

	// Syntax colors
	if theme.Syntax.Keyword == "" {
		theme.Syntax.Keyword = def.Syntax.Keyword
	}
	if theme.Syntax.String == "" {
		theme.Syntax.String = def.Syntax.String
	}
	if theme.Syntax.Comment == "" {
		theme.Syntax.Comment = def.Syntax.Comment
	}
	if theme.Syntax.Number == "" {
		theme.Syntax.Number = def.Syntax.Number
	}
	if theme.Syntax.Operator == "" {
		theme.Syntax.Operator = def.Syntax.Operator
	}
	if theme.Syntax.Function == "" {
		theme.Syntax.Function = def.Syntax.Function
	}
	if theme.Syntax.Type == "" {
		theme.Syntax.Type = def.Syntax.Type
	}

	return theme
}

// ThemeNames returns the list of built-in theme names
func ThemeNames() []string {
	return []string{"default", "dark", "light", "monokai"}
}

// ListUserThemes returns a list of user-defined theme names
func ListUserThemes() []string {
	themesDir, err := ThemesDir()
	if err != nil {
		return nil
	}

	entries, err := os.ReadDir(themesDir)
	if err != nil {
		return nil
	}

	var themes []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) == ".toml" {
			themes = append(themes, name[:len(name)-5]) // Remove .toml extension
		}
	}
	return themes
}
