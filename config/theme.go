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
	// Scrollbar colors
	ScrollbarTrack string `toml:"scrollbar_track"` // Scrollbar track color
	ScrollbarThumb string `toml:"scrollbar_thumb"` // Scrollbar thumb color
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
			ScrollbarTrack:   "8",  // Gray
			ScrollbarThumb:   "6",  // Cyan
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
			ScrollbarTrack:   "240", // Medium gray
			ScrollbarThumb:   "43",  // Teal
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
			ScrollbarTrack:   "249", // Medium gray
			ScrollbarThumb:   "32",  // Blue
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
			ScrollbarTrack:   "59",  // Gray
			ScrollbarThumb:   "208", // Orange
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
	"nord": {
		Name:        "nord",
		Description: "Arctic, north-bluish color palette",
		Author:      "Arctic Ice Studio",
		UI: UIColors{
			MenuBg:           "#3B4252", // nord1
			MenuFg:           "#ECEFF4", // nord6
			MenuHighlightBg:  "#5E81AC", // nord10
			MenuHighlightFg:  "#ECEFF4", // nord6
			StatusBg:         "#3B4252", // nord1
			StatusFg:         "#ECEFF4", // nord6
			StatusAccent:     "#88C0D0", // nord8
			SelectionBg:      "#4C566A", // nord3
			SelectionFg:      "#ECEFF4", // nord6
			LineNumber:       "#4C566A", // nord3
			LineNumberActive: "#D8DEE9", // nord4
			ErrorFg:          "#BF616A", // nord11
			DisabledFg:       "#4C566A", // nord3
			DialogBg:         "#3B4252", // nord1
			DialogFg:         "#ECEFF4", // nord6
			DialogBorder:     "#4C566A", // nord3
			DialogTitle:      "#88C0D0", // nord8
			DialogButton:     "#5E81AC", // nord10
			DialogButtonFg:   "#ECEFF4", // nord6
			ScrollbarTrack:   "#4C566A", // nord3
			ScrollbarThumb:   "#5E81AC", // nord10
		},
		Syntax: SyntaxColors{
			Keyword:  "#81A1C1", // nord9
			String:   "#A3BE8C", // nord14
			Comment:  "#616E88", // nord3 brightened
			Number:   "#B48EAD", // nord15
			Operator: "#81A1C1", // nord9
			Function: "#88C0D0", // nord8
			Type:     "#8FBCBB", // nord7
		},
	},
	"dracula": {
		Name:        "dracula",
		Description: "Dark theme with vibrant colors",
		Author:      "Zeno Rocha",
		UI: UIColors{
			MenuBg:           "#282A36", // background
			MenuFg:           "#F8F8F2", // foreground
			MenuHighlightBg:  "#BD93F9", // purple
			MenuHighlightFg:  "#282A36", // background
			StatusBg:         "#282A36", // background
			StatusFg:         "#F8F8F2", // foreground
			StatusAccent:     "#FF79C6", // pink
			SelectionBg:      "#44475A", // selection
			SelectionFg:      "#F8F8F2", // foreground
			LineNumber:       "#6272A4", // comment
			LineNumberActive: "#F8F8F2", // foreground
			ErrorFg:          "#FF5555", // red
			DisabledFg:       "#6272A4", // comment
			DialogBg:         "#282A36", // background
			DialogFg:         "#F8F8F2", // foreground
			DialogBorder:     "#6272A4", // comment
			DialogTitle:      "#BD93F9", // purple
			DialogButton:     "#50FA7B", // green
			DialogButtonFg:   "#282A36", // background
			ScrollbarTrack:   "#6272A4", // comment
			ScrollbarThumb:   "#BD93F9", // purple
		},
		Syntax: SyntaxColors{
			Keyword:  "#FF79C6", // pink
			String:   "#F1FA8C", // yellow
			Comment:  "#6272A4", // comment
			Number:   "#BD93F9", // purple
			Operator: "#FF79C6", // pink
			Function: "#50FA7B", // green
			Type:     "#8BE9FD", // cyan
		},
	},
	"gruvbox": {
		Name:        "gruvbox",
		Description: "Retro groove color scheme",
		Author:      "morhetz",
		UI: UIColors{
			MenuBg:           "#282828", // bg0
			MenuFg:           "#EBDBB2", // fg1
			MenuHighlightBg:  "#D79921", // yellow
			MenuHighlightFg:  "#282828", // bg0
			StatusBg:         "#282828", // bg0
			StatusFg:         "#EBDBB2", // fg1
			StatusAccent:     "#D79921", // yellow
			SelectionBg:      "#504945", // bg2
			SelectionFg:      "#EBDBB2", // fg1
			LineNumber:       "#665C54", // bg3
			LineNumberActive: "#EBDBB2", // fg1
			ErrorFg:          "#FB4934", // bright red
			DisabledFg:       "#665C54", // bg3
			DialogBg:         "#3C3836", // bg1
			DialogFg:         "#EBDBB2", // fg1
			DialogBorder:     "#665C54", // bg3
			DialogTitle:      "#FABD2F", // bright yellow
			DialogButton:     "#98971A", // green
			DialogButtonFg:   "#EBDBB2", // fg1
			ScrollbarTrack:   "#665C54", // bg3
			ScrollbarThumb:   "#D79921", // yellow
		},
		Syntax: SyntaxColors{
			Keyword:  "#FB4934", // bright red
			String:   "#B8BB26", // bright green
			Comment:  "#928374", // gray
			Number:   "#D3869B", // bright purple
			Operator: "#FE8019", // bright orange
			Function: "#FABD2F", // bright yellow
			Type:     "#83A598", // bright blue
		},
	},
	"solarized": {
		Name:        "solarized",
		Description: "Precision colors for machines and people",
		Author:      "Ethan Schoonover",
		UI: UIColors{
			MenuBg:           "#002B36", // base03
			MenuFg:           "#839496", // base0
			MenuHighlightBg:  "#268BD2", // blue
			MenuHighlightFg:  "#FDF6E3", // base3
			StatusBg:         "#002B36", // base03
			StatusFg:         "#839496", // base0
			StatusAccent:     "#2AA198", // cyan
			SelectionBg:      "#073642", // base02
			SelectionFg:      "#93A1A1", // base1
			LineNumber:       "#586E75", // base01
			LineNumberActive: "#93A1A1", // base1
			ErrorFg:          "#DC322F", // red
			DisabledFg:       "#586E75", // base01
			DialogBg:         "#073642", // base02
			DialogFg:         "#839496", // base0
			DialogBorder:     "#586E75", // base01
			DialogTitle:      "#268BD2", // blue
			DialogButton:     "#2AA198", // cyan
			DialogButtonFg:   "#FDF6E3", // base3
			ScrollbarTrack:   "#586E75", // base01
			ScrollbarThumb:   "#268BD2", // blue
		},
		Syntax: SyntaxColors{
			Keyword:  "#859900", // green
			String:   "#2AA198", // cyan
			Comment:  "#586E75", // base01
			Number:   "#D33682", // magenta
			Operator: "#859900", // green
			Function: "#268BD2", // blue
			Type:     "#B58900", // yellow
		},
	},
	"catppuccin": {
		Name:        "catppuccin",
		Description: "Soothing pastel theme (Mocha)",
		Author:      "Catppuccin",
		UI: UIColors{
			MenuBg:           "#1E1E2E", // base
			MenuFg:           "#CDD6F4", // text
			MenuHighlightBg:  "#CBA6F7", // mauve
			MenuHighlightFg:  "#1E1E2E", // base
			StatusBg:         "#1E1E2E", // base
			StatusFg:         "#CDD6F4", // text
			StatusAccent:     "#F5C2E7", // pink
			SelectionBg:      "#45475A", // surface1
			SelectionFg:      "#CDD6F4", // text
			LineNumber:       "#6C7086", // overlay0
			LineNumberActive: "#CDD6F4", // text
			ErrorFg:          "#F38BA8", // red
			DisabledFg:       "#6C7086", // overlay0
			DialogBg:         "#313244", // surface0
			DialogFg:         "#CDD6F4", // text
			DialogBorder:     "#585B70", // surface2
			DialogTitle:      "#CBA6F7", // mauve
			DialogButton:     "#89B4FA", // blue
			DialogButtonFg:   "#1E1E2E", // base
			ScrollbarTrack:   "#6C7086", // overlay0
			ScrollbarThumb:   "#CBA6F7", // mauve
		},
		Syntax: SyntaxColors{
			Keyword:  "#CBA6F7", // mauve
			String:   "#A6E3A1", // green
			Comment:  "#6C7086", // overlay0
			Number:   "#FAB387", // peach
			Operator: "#89DCEB", // sky
			Function: "#89B4FA", // blue
			Type:     "#F9E2AF", // yellow
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
// If a built-in theme with the same name exists, use it for defaults
func mergeWithDefault(theme Theme) Theme {
	// Try to find a built-in theme with the same name first
	def, exists := builtinThemes[theme.Name]
	if !exists {
		def = DefaultTheme()
	}

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
	if theme.UI.ScrollbarTrack == "" {
		theme.UI.ScrollbarTrack = def.UI.ScrollbarTrack
	}
	if theme.UI.ScrollbarThumb == "" {
		theme.UI.ScrollbarThumb = def.UI.ScrollbarThumb
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
	return []string{"default", "dark", "light", "monokai", "nord", "dracula", "gruvbox", "solarized", "catppuccin"}
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

// ExportTheme saves a theme to the user's themes directory as a TOML file
// Returns the full path to the saved file
func ExportTheme(theme Theme, filename string) (string, error) {
	themesDir, err := ThemesDir()
	if err != nil {
		return "", err
	}

	// Create themes directory if it doesn't exist
	if err := os.MkdirAll(themesDir, 0755); err != nil {
		return "", err
	}

	// Ensure filename has .toml extension
	if filepath.Ext(filename) != ".toml" {
		filename = filename + ".toml"
	}

	themePath := filepath.Join(themesDir, filename)

	// Create/overwrite the file
	f, err := os.Create(themePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// Encode theme as TOML
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(theme); err != nil {
		return "", err
	}

	return themePath, nil
}

// GetTheme returns a theme by name (built-in or user), without applying it
func GetTheme(name string) Theme {
	// Try user themes first
	theme, err := loadUserTheme(name)
	if err == nil {
		return theme
	}

	// Fall back to built-in
	if builtin, ok := builtinThemes[name]; ok {
		return builtin
	}

	return DefaultTheme()
}

// ThemeFilePath returns the path to a user theme file, or empty if it doesn't exist
func ThemeFilePath(name string) string {
	themesDir, err := ThemesDir()
	if err != nil {
		return ""
	}

	themePath := filepath.Join(themesDir, name+".toml")
	if _, err := os.Stat(themePath); os.IsNotExist(err) {
		return ""
	}
	return themePath
}
