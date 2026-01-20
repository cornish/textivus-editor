package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// KeyBinding represents a single action's key bindings
type KeyBinding struct {
	Primary   string `toml:"primary"`
	Alternate string `toml:"alternate,omitempty"`
}

// KeybindingsConfig holds all configurable keybindings
type KeybindingsConfig struct {
	// File operations
	New         KeyBinding `toml:"new"`
	Open        KeyBinding `toml:"open"`
	SaveFile    KeyBinding `toml:"save"`
	SaveAs      KeyBinding `toml:"save_as"`
	Close       KeyBinding `toml:"close"`
	RecentFiles KeyBinding `toml:"recent_files"`
	Quit        KeyBinding `toml:"quit"`

	// Edit operations
	Undo      KeyBinding `toml:"undo"`
	Redo      KeyBinding `toml:"redo"`
	Cut       KeyBinding `toml:"cut"`
	Copy      KeyBinding `toml:"copy"`
	Paste     KeyBinding `toml:"paste"`
	CutLine   KeyBinding `toml:"cut_line"`
	SelectAll KeyBinding `toml:"select_all"`

	// Search operations
	Find     KeyBinding `toml:"find"`
	FindNext KeyBinding `toml:"find_next"`
	Replace  KeyBinding `toml:"replace"`
	GoToLine KeyBinding `toml:"goto_line"`

	// Navigation
	WordLeft  KeyBinding `toml:"word_left"`
	WordRight KeyBinding `toml:"word_right"`
	DocStart  KeyBinding `toml:"doc_start"`
	DocEnd    KeyBinding `toml:"doc_end"`

	// Buffer operations
	NextBuffer KeyBinding `toml:"next_buffer"`
	PrevBuffer KeyBinding `toml:"prev_buffer"`

	// View toggles
	ToggleLineNumbers KeyBinding `toml:"toggle_line_numbers"`

	// Help
	Help KeyBinding `toml:"help"`
}

// DefaultKeybindings returns the default keybinding configuration
func DefaultKeybindings() *KeybindingsConfig {
	return &KeybindingsConfig{
		// File operations
		New:         KeyBinding{Primary: "ctrl+n"},
		Open:        KeyBinding{Primary: "ctrl+o"},
		SaveFile:    KeyBinding{Primary: "ctrl+s"},
		SaveAs:      KeyBinding{Primary: ""},
		Close:       KeyBinding{Primary: "ctrl+w"},
		RecentFiles: KeyBinding{Primary: "ctrl+r"},
		Quit:        KeyBinding{Primary: "ctrl+q"},

		// Edit operations
		Undo:      KeyBinding{Primary: "ctrl+z"},
		Redo:      KeyBinding{Primary: "ctrl+y"},
		Cut:       KeyBinding{Primary: "ctrl+x"},
		Copy:      KeyBinding{Primary: "ctrl+c"},
		Paste:     KeyBinding{Primary: "ctrl+v"},
		CutLine:   KeyBinding{Primary: "ctrl+k"},
		SelectAll: KeyBinding{Primary: "ctrl+a"},

		// Search operations
		Find:     KeyBinding{Primary: "ctrl+f"},
		FindNext: KeyBinding{Primary: "f3"},
		Replace:  KeyBinding{Primary: "ctrl+h"},
		GoToLine: KeyBinding{Primary: "ctrl+g"},

		// Navigation
		WordLeft:  KeyBinding{Primary: "ctrl+left"},
		WordRight: KeyBinding{Primary: "ctrl+right"},
		DocStart:  KeyBinding{Primary: "ctrl+home"},
		DocEnd:    KeyBinding{Primary: "ctrl+end"},

		// Buffer operations
		NextBuffer: KeyBinding{Primary: "alt+>", Alternate: "ctrl+tab"},
		PrevBuffer: KeyBinding{Primary: "alt+<", Alternate: "ctrl+shift+tab"},

		// View toggles
		ToggleLineNumbers: KeyBinding{Primary: "ctrl+l"},

		// Help
		Help: KeyBinding{Primary: "f1"},
	}
}

// ActionName maps action names for display
var ActionNames = map[string]string{
	"new":                 "New File",
	"open":                "Open File",
	"save":                "Save",
	"save_as":             "Save As",
	"close":               "Close",
	"recent_files":        "Recent Files",
	"quit":                "Quit",
	"undo":                "Undo",
	"redo":                "Redo",
	"cut":                 "Cut",
	"copy":                "Copy",
	"paste":               "Paste",
	"cut_line":            "Cut Line",
	"select_all":          "Select All",
	"find":                "Find",
	"find_next":           "Find Next",
	"replace":             "Replace",
	"goto_line":           "Go to Line",
	"word_left":           "Word Left",
	"word_right":          "Word Right",
	"doc_start":           "Document Start",
	"doc_end":             "Document End",
	"next_buffer":         "Next Buffer",
	"prev_buffer":         "Previous Buffer",
	"toggle_line_numbers": "Toggle Line Numbers",
	"help":                "Help",
}

// KeybindingsPath returns the path to the keybindings file
func KeybindingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, configDirName, "keybindings.toml"), nil
}

// LoadKeybindings loads keybindings from disk, returning defaults if not found
func LoadKeybindings() *KeybindingsConfig {
	kb := DefaultKeybindings()

	path, err := KeybindingsPath()
	if err != nil {
		return kb
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return kb
	}

	if _, err := toml.DecodeFile(path, kb); err != nil {
		return kb
	}

	return kb
}

// Save writes keybindings to disk
func (kb *KeybindingsConfig) Save() error {
	path, err := KeybindingsPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("# Textivus keybindings\n")
	f.WriteString("# Format: primary = \"key\", alternate = \"key\" (optional)\n")
	f.WriteString("# Examples: \"ctrl+s\", \"alt+f\", \"f1\", \"ctrl+shift+s\"\n\n")

	encoder := toml.NewEncoder(f)
	return encoder.Encode(kb)
}

// GetBinding returns the KeyBinding for a given action name
func (kb *KeybindingsConfig) GetBinding(action string) KeyBinding {
	switch action {
	case "new":
		return kb.New
	case "open":
		return kb.Open
	case "save":
		return kb.SaveFile
	case "save_as":
		return kb.SaveAs
	case "close":
		return kb.Close
	case "recent_files":
		return kb.RecentFiles
	case "quit":
		return kb.Quit
	case "undo":
		return kb.Undo
	case "redo":
		return kb.Redo
	case "cut":
		return kb.Cut
	case "copy":
		return kb.Copy
	case "paste":
		return kb.Paste
	case "cut_line":
		return kb.CutLine
	case "select_all":
		return kb.SelectAll
	case "find":
		return kb.Find
	case "find_next":
		return kb.FindNext
	case "replace":
		return kb.Replace
	case "goto_line":
		return kb.GoToLine
	case "word_left":
		return kb.WordLeft
	case "word_right":
		return kb.WordRight
	case "doc_start":
		return kb.DocStart
	case "doc_end":
		return kb.DocEnd
	case "next_buffer":
		return kb.NextBuffer
	case "prev_buffer":
		return kb.PrevBuffer
	case "toggle_line_numbers":
		return kb.ToggleLineNumbers
	case "help":
		return kb.Help
	}
	return KeyBinding{}
}

// SetBinding sets the KeyBinding for a given action name
func (kb *KeybindingsConfig) SetBinding(action string, binding KeyBinding) {
	switch action {
	case "new":
		kb.New = binding
	case "open":
		kb.Open = binding
	case "save":
		kb.SaveFile = binding
	case "save_as":
		kb.SaveAs = binding
	case "close":
		kb.Close = binding
	case "recent_files":
		kb.RecentFiles = binding
	case "quit":
		kb.Quit = binding
	case "undo":
		kb.Undo = binding
	case "redo":
		kb.Redo = binding
	case "cut":
		kb.Cut = binding
	case "copy":
		kb.Copy = binding
	case "paste":
		kb.Paste = binding
	case "cut_line":
		kb.CutLine = binding
	case "select_all":
		kb.SelectAll = binding
	case "find":
		kb.Find = binding
	case "find_next":
		kb.FindNext = binding
	case "replace":
		kb.Replace = binding
	case "goto_line":
		kb.GoToLine = binding
	case "word_left":
		kb.WordLeft = binding
	case "word_right":
		kb.WordRight = binding
	case "doc_start":
		kb.DocStart = binding
	case "doc_end":
		kb.DocEnd = binding
	case "next_buffer":
		kb.NextBuffer = binding
	case "prev_buffer":
		kb.PrevBuffer = binding
	case "toggle_line_numbers":
		kb.ToggleLineNumbers = binding
	case "help":
		kb.Help = binding
	}
}

// AllActions returns a list of all action names in display order
func AllActions() []string {
	return []string{
		"new", "open", "save", "save_as", "close", "recent_files", "quit",
		"undo", "redo", "cut", "copy", "paste", "cut_line", "select_all",
		"find", "find_next", "replace", "goto_line",
		"word_left", "word_right", "doc_start", "doc_end",
		"next_buffer", "prev_buffer",
		"toggle_line_numbers",
		"help",
	}
}

// Matches checks if a key string matches this binding (primary or alternate)
func (b KeyBinding) Matches(key string) bool {
	key = strings.ToLower(key)
	return (b.Primary != "" && strings.ToLower(b.Primary) == key) ||
		(b.Alternate != "" && strings.ToLower(b.Alternate) == key)
}

// DisplayString returns a human-readable string for the binding
func (b KeyBinding) DisplayString() string {
	if b.Primary == "" && b.Alternate == "" {
		return "(none)"
	}
	if b.Alternate == "" {
		return FormatKeyForDisplay(b.Primary)
	}
	return FormatKeyForDisplay(b.Primary) + " / " + FormatKeyForDisplay(b.Alternate)
}

// FormatKeyForDisplay converts a key string to a more readable format
func FormatKeyForDisplay(key string) string {
	if key == "" {
		return ""
	}
	// Capitalize modifiers
	key = strings.ReplaceAll(key, "ctrl+", "Ctrl+")
	key = strings.ReplaceAll(key, "alt+", "Alt+")
	key = strings.ReplaceAll(key, "shift+", "Shift+")
	// Capitalize special keys
	key = strings.ReplaceAll(key, "home", "Home")
	key = strings.ReplaceAll(key, "end", "End")
	key = strings.ReplaceAll(key, "left", "Left")
	key = strings.ReplaceAll(key, "right", "Right")
	key = strings.ReplaceAll(key, "tab", "Tab")
	// F-keys
	for i := 1; i <= 12; i++ {
		old := strings.ToLower(string(rune('f')) + string(rune('0'+i)))
		if i >= 10 {
			old = "f" + string(rune('0'+i/10)) + string(rune('0'+i%10))
		} else {
			old = "f" + string(rune('0'+i))
		}
		new := "F" + old[1:]
		key = strings.ReplaceAll(key, old, new)
	}
	return key
}

// FindConflicts checks for key conflicts and returns a map of conflicting actions
func (kb *KeybindingsConfig) FindConflicts() map[string][]string {
	conflicts := make(map[string][]string)
	keyToActions := make(map[string][]string)

	for _, action := range AllActions() {
		binding := kb.GetBinding(action)
		if binding.Primary != "" {
			key := strings.ToLower(binding.Primary)
			keyToActions[key] = append(keyToActions[key], action)
		}
		if binding.Alternate != "" {
			key := strings.ToLower(binding.Alternate)
			keyToActions[key] = append(keyToActions[key], action)
		}
	}

	for key, actions := range keyToActions {
		if len(actions) > 1 {
			conflicts[key] = actions
		}
	}

	return conflicts
}
