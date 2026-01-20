package config

import (
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// configDirName is the name of the config directory
const configDirName = "textivus"

// oldConfigDirName is the legacy config directory name (for migration)
const oldConfigDirName = "festivus"

// configMigrated tracks whether migration has been attempted this session
var configMigrated bool

// MigrateConfig checks for old festivus config and migrates to textivus
// This should be called once at startup before any config operations
func MigrateConfig() error {
	if configMigrated {
		return nil
	}
	configMigrated = true

	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil // Can't determine config dir, skip migration
		}
		configDir = filepath.Join(home, ".config")
	}

	oldDir := filepath.Join(configDir, oldConfigDirName)
	newDir := filepath.Join(configDir, configDirName)

	// Check if old dir exists and new dir doesn't
	oldInfo, oldErr := os.Stat(oldDir)
	_, newErr := os.Stat(newDir)

	if oldErr == nil && oldInfo.IsDir() && os.IsNotExist(newErr) {
		// Old exists, new doesn't - migrate by copying
		if err := copyDir(oldDir, newDir); err != nil {
			return err
		}
	}

	return nil
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// Config holds the editor configuration
type Config struct {
	Editor        EditorConfig `toml:"editor"`
	Theme         ThemeConfig  `toml:"theme"`
	RecentFiles   []string     `toml:"recent_files,omitempty"`   // Recently opened files (max 10)
	RecentDirs    []string     `toml:"recent_dirs,omitempty"`    // Recently visited directories (max 10)
	FavoriteFiles []string     `toml:"favorite_files,omitempty"` // User-favorited files (max 50)
	FavoriteDirs  []string     `toml:"favorite_dirs,omitempty"`  // User-favorited directories (max 50)
}

// MaxRecentFiles is the maximum number of recent files to track
const MaxRecentFiles = 10

// MaxRecentDirs is the maximum number of recent directories to track
const MaxRecentDirs = 10

// MaxFavorites is the maximum number of favorite files or directories
const MaxFavorites = 50

// AddRecentFile adds a file to the recent files list
func (c *Config) AddRecentFile(path string) {
	// Make path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Remove if already in list (will re-add at top)
	newList := make([]string, 0, MaxRecentFiles)
	for _, f := range c.RecentFiles {
		if f != absPath {
			newList = append(newList, f)
		}
	}

	// Add to front
	c.RecentFiles = append([]string{absPath}, newList...)

	// Trim to max
	if len(c.RecentFiles) > MaxRecentFiles {
		c.RecentFiles = c.RecentFiles[:MaxRecentFiles]
	}
}

// AddRecentDir adds a directory to the recent directories list
func (c *Config) AddRecentDir(path string) {
	// Make path absolute
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Remove if already in list (will re-add at top)
	newList := make([]string, 0, MaxRecentDirs)
	for _, d := range c.RecentDirs {
		if d != absPath {
			newList = append(newList, d)
		}
	}

	// Add to front
	c.RecentDirs = append([]string{absPath}, newList...)

	// Trim to max
	if len(c.RecentDirs) > MaxRecentDirs {
		c.RecentDirs = c.RecentDirs[:MaxRecentDirs]
	}
}

// AddFavoriteFile adds a file to favorites (if not already present)
func (c *Config) AddFavoriteFile(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Check if already favorited
	for _, f := range c.FavoriteFiles {
		if f == absPath {
			return false // Already exists
		}
	}

	// Check max limit
	if len(c.FavoriteFiles) >= MaxFavorites {
		return false // At limit
	}

	c.FavoriteFiles = append(c.FavoriteFiles, absPath)
	return true
}

// RemoveFavoriteFile removes a file from favorites
func (c *Config) RemoveFavoriteFile(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for i, f := range c.FavoriteFiles {
		if f == absPath {
			c.FavoriteFiles = append(c.FavoriteFiles[:i], c.FavoriteFiles[i+1:]...)
			return true
		}
	}
	return false
}

// IsFavoriteFile checks if a file is in favorites
func (c *Config) IsFavoriteFile(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for _, f := range c.FavoriteFiles {
		if f == absPath {
			return true
		}
	}
	return false
}

// AddFavoriteDir adds a directory to favorites (if not already present)
func (c *Config) AddFavoriteDir(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	// Check if already favorited
	for _, d := range c.FavoriteDirs {
		if d == absPath {
			return false // Already exists
		}
	}

	// Check max limit
	if len(c.FavoriteDirs) >= MaxFavorites {
		return false // At limit
	}

	c.FavoriteDirs = append(c.FavoriteDirs, absPath)
	return true
}

// RemoveFavoriteDir removes a directory from favorites
func (c *Config) RemoveFavoriteDir(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for i, d := range c.FavoriteDirs {
		if d == absPath {
			c.FavoriteDirs = append(c.FavoriteDirs[:i], c.FavoriteDirs[i+1:]...)
			return true
		}
	}
	return false
}

// IsFavoriteDir checks if a directory is in favorites
func (c *Config) IsFavoriteDir(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	for _, d := range c.FavoriteDirs {
		if d == absPath {
			return true
		}
	}
	return false
}

// ToggleFavorite toggles favorite status for a file or directory
// Returns (isNowFavorite, wasChanged)
func (c *Config) ToggleFavorite(path string, isDir bool) (bool, bool) {
	if isDir {
		if c.IsFavoriteDir(path) {
			return false, c.RemoveFavoriteDir(path)
		}
		return true, c.AddFavoriteDir(path)
	}
	if c.IsFavoriteFile(path) {
		return false, c.RemoveFavoriteFile(path)
	}
	return true, c.AddFavoriteFile(path)
}

// EditorConfig holds editor-specific settings
type EditorConfig struct {
	WordWrap        bool  `toml:"word_wrap"`
	LineNumbers     bool  `toml:"line_numbers"`
	SyntaxHighlight bool  `toml:"syntax_highlight"`
	TrueColor       *bool `toml:"true_color"`     // nil = auto (true), false = force 256-color
	AsciiMode       *bool `toml:"ascii_mode"`     // nil = auto-detect, true/false = override
	BackupCount     int   `toml:"backup_count"`   // 0=disabled, 1=filename~, >1=filename~1~ through filename~N~
	Scrollbar       bool  `toml:"scrollbar"`      // Show scrollbar
	MaxBuffers      int   `toml:"max_buffers"`    // Maximum open buffers (0=unlimited, default 20)
	TabWidth        int   `toml:"tab_width"`      // Display width of tabs (default 4)
	TabsToSpaces    bool  `toml:"tabs_to_spaces"` // Insert spaces instead of tab characters
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
			SyntaxHighlight: true,  // Enabled by default
			MaxBuffers:      20,    // Default max open buffers
			TabWidth:        4,     // Default tab width
			TabsToSpaces:    false, // Use real tabs by default
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
	return filepath.Join(configDir, configDirName, "config.toml"), nil
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
	return filepath.Join(configDir, configDirName, "themes"), nil
}

// ConfigLoadError holds details about a config loading error
type ConfigLoadError struct {
	FilePath string
	Err      error
}

func (e *ConfigLoadError) Error() string {
	return e.Err.Error()
}

// Load reads the configuration from disk
// Returns default config if file doesn't exist
// Returns ConfigLoadError if file exists but has parse errors
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
		return cfg, &ConfigLoadError{FilePath: path, Err: err}
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
	f.WriteString("# Textivus configuration\n\n")

	// Encode config as TOML
	encoder := toml.NewEncoder(f)
	return encoder.Encode(c)
}

// GetResolved loads and returns the complete theme
func (t *ThemeConfig) GetResolved() Theme {
	return LoadTheme(t.Name)
}
