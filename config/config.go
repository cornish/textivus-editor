package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the editor configuration
type Config struct {
	Editor EditorConfig `toml:"editor"`
	Theme  ThemeConfig  `toml:"theme"`
}

// EditorConfig holds editor-specific settings
type EditorConfig struct {
	WordWrap        bool  `toml:"word_wrap"`
	LineNumbers     bool  `toml:"line_numbers"`
	SyntaxHighlight bool  `toml:"syntax_highlight"`
	TrueColor       *bool `toml:"true_color"`    // nil = auto (true), false = force 256-color
	AsciiMode       *bool `toml:"ascii_mode"`    // nil = auto-detect, true/false = override
	BackupOnSave    bool  `toml:"backup_on_save"` // Create filename~ backup before saving
}

// ThemeConfig holds the theme reference in the main config
// Just references a theme by name - the actual colors come from theme files
type ThemeConfig struct {
	Name string `toml:"name"` // Theme name (built-in or from themes/ directory)
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Editor: EditorConfig{
			WordWrap:        false,
			LineNumbers:     false,
			SyntaxHighlight: true, // Enabled by default
		},
		Theme: ThemeConfig{
			Name: "default",
		},
	}
}

// ConfigPath returns the path to the config file
func ConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "festivus", "config.toml"), nil
}

// ThemesDir returns the path to the user themes directory
func ThemesDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "festivus", "themes"), nil
}

// Load reads the configuration from disk
// Returns default config if file doesn't exist
func Load() (*Config, error) {
	cfg := DefaultConfig()

	path, err := ConfigPath()
	if err != nil {
		return cfg, nil // Return defaults on error
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil // Return defaults if no config file
	}

	// Parse the config file
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return cfg, err // Return defaults but also the error
	}

	return cfg, nil
}

// Save writes the configuration to disk
func (c *Config) Save() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Create/overwrite the file
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write header comment
	f.WriteString("# Festivus configuration\n\n")

	// Encode config as TOML
	encoder := toml.NewEncoder(f)
	return encoder.Encode(c)
}

// GetResolved loads and returns the complete theme
func (t *ThemeConfig) GetResolved() Theme {
	return LoadTheme(t.Name)
}
