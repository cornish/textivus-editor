package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Editor.WordWrap != false {
		t.Error("DefaultConfig().Editor.WordWrap should be false")
	}
	if cfg.Editor.LineNumbers != false {
		t.Error("DefaultConfig().Editor.LineNumbers should be false")
	}
	if cfg.Editor.SyntaxHighlight != true {
		t.Error("DefaultConfig().Editor.SyntaxHighlight should be true")
	}
	if cfg.Editor.MaxBuffers != 20 {
		t.Errorf("DefaultConfig().Editor.MaxBuffers = %d, want 20", cfg.Editor.MaxBuffers)
	}
	if cfg.Editor.TabWidth != 4 {
		t.Errorf("DefaultConfig().Editor.TabWidth = %d, want 4", cfg.Editor.TabWidth)
	}
	if cfg.Editor.TabsToSpaces != false {
		t.Error("DefaultConfig().Editor.TabsToSpaces should be false")
	}
	if cfg.Theme.Name != "default" {
		t.Errorf("DefaultConfig().Theme.Name = %q, want 'default'", cfg.Theme.Name)
	}
}

func TestAddRecentFile(t *testing.T) {
	cfg := DefaultConfig()

	// Add a file
	cfg.AddRecentFile("/path/to/file1.txt")
	if len(cfg.RecentFiles) != 1 {
		t.Fatalf("RecentFiles length = %d, want 1", len(cfg.RecentFiles))
	}

	// Add another file
	cfg.AddRecentFile("/path/to/file2.txt")
	if len(cfg.RecentFiles) != 2 {
		t.Fatalf("RecentFiles length = %d, want 2", len(cfg.RecentFiles))
	}

	// Most recent should be first
	if !filepath.IsAbs(cfg.RecentFiles[0]) || filepath.Base(cfg.RecentFiles[0]) != "file2.txt" {
		t.Errorf("RecentFiles[0] = %q, want file2.txt to be first", cfg.RecentFiles[0])
	}

	// Re-add file1 - should move to front
	cfg.AddRecentFile("/path/to/file1.txt")
	if len(cfg.RecentFiles) != 2 {
		t.Fatalf("RecentFiles length after re-add = %d, want 2", len(cfg.RecentFiles))
	}
	if filepath.Base(cfg.RecentFiles[0]) != "file1.txt" {
		t.Errorf("RecentFiles[0] after re-add = %q, want file1.txt first", cfg.RecentFiles[0])
	}
}

func TestAddRecentFileMaxLimit(t *testing.T) {
	cfg := DefaultConfig()

	// Add more than MaxRecentFiles
	for i := 0; i < MaxRecentFiles+5; i++ {
		cfg.AddRecentFile("/path/to/file" + string(rune('a'+i)) + ".txt")
	}

	if len(cfg.RecentFiles) != MaxRecentFiles {
		t.Errorf("RecentFiles length = %d, want %d (max)", len(cfg.RecentFiles), MaxRecentFiles)
	}
}

func TestAddRecentDir(t *testing.T) {
	cfg := DefaultConfig()

	cfg.AddRecentDir("/path/to/dir1")
	cfg.AddRecentDir("/path/to/dir2")

	if len(cfg.RecentDirs) != 2 {
		t.Fatalf("RecentDirs length = %d, want 2", len(cfg.RecentDirs))
	}

	// Most recent should be first
	if filepath.Base(cfg.RecentDirs[0]) != "dir2" {
		t.Errorf("RecentDirs[0] = %q, want dir2 first", cfg.RecentDirs[0])
	}
}

func TestAddRecentDirMaxLimit(t *testing.T) {
	cfg := DefaultConfig()

	for i := 0; i < MaxRecentDirs+5; i++ {
		cfg.AddRecentDir("/path/to/dir" + string(rune('a'+i)))
	}

	if len(cfg.RecentDirs) != MaxRecentDirs {
		t.Errorf("RecentDirs length = %d, want %d (max)", len(cfg.RecentDirs), MaxRecentDirs)
	}
}

func TestFavoriteFiles(t *testing.T) {
	cfg := DefaultConfig()

	// Add favorite
	added := cfg.AddFavoriteFile("/path/to/fav.txt")
	if !added {
		t.Error("AddFavoriteFile should return true on first add")
	}
	if len(cfg.FavoriteFiles) != 1 {
		t.Fatalf("FavoriteFiles length = %d, want 1", len(cfg.FavoriteFiles))
	}

	// Check if favorited
	if !cfg.IsFavoriteFile("/path/to/fav.txt") {
		t.Error("IsFavoriteFile should return true for added file")
	}

	// Try to add duplicate
	added = cfg.AddFavoriteFile("/path/to/fav.txt")
	if added {
		t.Error("AddFavoriteFile should return false for duplicate")
	}
	if len(cfg.FavoriteFiles) != 1 {
		t.Errorf("FavoriteFiles length after duplicate = %d, want 1", len(cfg.FavoriteFiles))
	}

	// Remove favorite
	removed := cfg.RemoveFavoriteFile("/path/to/fav.txt")
	if !removed {
		t.Error("RemoveFavoriteFile should return true")
	}
	if len(cfg.FavoriteFiles) != 0 {
		t.Errorf("FavoriteFiles length after remove = %d, want 0", len(cfg.FavoriteFiles))
	}

	// Check not favorited
	if cfg.IsFavoriteFile("/path/to/fav.txt") {
		t.Error("IsFavoriteFile should return false after removal")
	}

	// Remove non-existent
	removed = cfg.RemoveFavoriteFile("/path/to/nonexistent.txt")
	if removed {
		t.Error("RemoveFavoriteFile should return false for non-existent file")
	}
}

func TestFavoriteDirs(t *testing.T) {
	cfg := DefaultConfig()

	// Add favorite dir
	added := cfg.AddFavoriteDir("/path/to/favdir")
	if !added {
		t.Error("AddFavoriteDir should return true on first add")
	}

	// Check if favorited
	if !cfg.IsFavoriteDir("/path/to/favdir") {
		t.Error("IsFavoriteDir should return true for added dir")
	}

	// Remove favorite dir
	removed := cfg.RemoveFavoriteDir("/path/to/favdir")
	if !removed {
		t.Error("RemoveFavoriteDir should return true")
	}

	if cfg.IsFavoriteDir("/path/to/favdir") {
		t.Error("IsFavoriteDir should return false after removal")
	}
}

func TestFavoriteMaxLimit(t *testing.T) {
	cfg := DefaultConfig()

	// Add max favorites
	for i := 0; i < MaxFavorites; i++ {
		cfg.AddFavoriteFile("/path/to/file" + string(rune(i)) + ".txt")
	}

	if len(cfg.FavoriteFiles) != MaxFavorites {
		t.Fatalf("FavoriteFiles length = %d, want %d", len(cfg.FavoriteFiles), MaxFavorites)
	}

	// Try to add one more
	added := cfg.AddFavoriteFile("/path/to/onemore.txt")
	if added {
		t.Error("AddFavoriteFile should return false when at max limit")
	}
	if len(cfg.FavoriteFiles) != MaxFavorites {
		t.Errorf("FavoriteFiles length after limit = %d, want %d", len(cfg.FavoriteFiles), MaxFavorites)
	}
}

func TestToggleFavorite(t *testing.T) {
	cfg := DefaultConfig()

	// Toggle on (file)
	isNow, changed := cfg.ToggleFavorite("/path/to/file.txt", false)
	if !isNow || !changed {
		t.Error("ToggleFavorite should add file and return (true, true)")
	}

	// Toggle off (file)
	isNow, changed = cfg.ToggleFavorite("/path/to/file.txt", false)
	if isNow || !changed {
		t.Error("ToggleFavorite should remove file and return (false, true)")
	}

	// Toggle on (dir)
	isNow, changed = cfg.ToggleFavorite("/path/to/dir", true)
	if !isNow || !changed {
		t.Error("ToggleFavorite should add dir and return (true, true)")
	}

	// Toggle off (dir)
	isNow, changed = cfg.ToggleFavorite("/path/to/dir", true)
	if isNow || !changed {
		t.Error("ToggleFavorite should remove dir and return (false, true)")
	}
}

func TestConfigLoadError(t *testing.T) {
	err := &ConfigLoadError{
		FilePath: "/path/to/config.toml",
		Err:      os.ErrNotExist,
	}

	if err.Error() != os.ErrNotExist.Error() {
		t.Errorf("ConfigLoadError.Error() = %q, want %q", err.Error(), os.ErrNotExist.Error())
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}

	if !filepath.IsAbs(path) {
		t.Errorf("ConfigPath() = %q, want absolute path", path)
	}

	if filepath.Base(path) != "config.toml" {
		t.Errorf("ConfigPath() base = %q, want 'config.toml'", filepath.Base(path))
	}

	if !contains(path, "festivus") {
		t.Errorf("ConfigPath() = %q, should contain 'festivus'", path)
	}
}

func TestThemesDir(t *testing.T) {
	dir, err := ThemesDir()
	if err != nil {
		t.Fatalf("ThemesDir() error: %v", err)
	}

	if !filepath.IsAbs(dir) {
		t.Errorf("ThemesDir() = %q, want absolute path", dir)
	}

	if filepath.Base(dir) != "themes" {
		t.Errorf("ThemesDir() base = %q, want 'themes'", filepath.Base(dir))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
