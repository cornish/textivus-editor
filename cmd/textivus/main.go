package main

import (
	"fmt"
	"os"

	"github.com/cornish/textivus-editor/config"
	"github.com/cornish/textivus-editor/editor"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.2.0"

func main() {
	// Parse command line arguments
	args := os.Args[1:]
	var filename string
	asciiMode := false

	// Handle flags
	for _, arg := range args {
		switch arg {
		case "--version", "-v":
			fmt.Printf("textivus %s\n", version)
			os.Exit(0)
		case "--help", "-h":
			printHelp()
			os.Exit(0)
		case "--ascii":
			asciiMode = true
		default:
			if filename == "" && !isFlag(arg) {
				filename = arg
			}
		}
	}

	// Detect terminal capabilities early
	config.InitCapabilities()

	// Migrate config from old location if needed
	config.MigrateConfig()

	// Load configuration
	cfg, configErr := config.Load()

	// Command-line --ascii overrides config
	if asciiMode {
		t := true
		cfg.Editor.AsciiMode = &t
	}

	// Create editor with config
	e := editor.NewWithConfig(cfg)

	// If config had parse errors, show error dialog on startup
	if configErr != nil {
		if loadErr, ok := configErr.(*config.ConfigLoadError); ok {
			e.SetConfigError(loadErr.FilePath, loadErr.Err.Error())
		}
	}

	// Load file if provided
	if filename != "" {
		// Check if file exists
		if _, err := os.Stat(filename); err == nil {
			if err := e.LoadFile(filename); err != nil {
				fmt.Fprintf(os.Stderr, "Error loading file: %v\n", err)
				os.Exit(1)
			}
		} else if os.IsNotExist(err) {
			// New file - just set the filename
			e.SetFilename(filename)
		} else {
			fmt.Fprintf(os.Stderr, "Error accessing file: %v\n", err)
			os.Exit(1)
		}
	}

	// Create and run the Bubbletea program
	p := tea.NewProgram(e, tea.WithAltScreen(), tea.WithMouseAllMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running editor: %v\n", err)
		os.Exit(1)
	}
}

func isFlag(s string) bool {
	return len(s) > 0 && s[0] == '-'
}

func printHelp() {
	fmt.Println("Textivus - A Text Editor for the Rest of Us")
	fmt.Println()
	fmt.Println("Usage: textivus [options] [file]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help     Show this help message")
	fmt.Println("  -v, --version  Show version information")
	fmt.Println("  --ascii        Use ASCII characters for dialogs")
	fmt.Println()
	fmt.Println("Keyboard Shortcuts:")
	fmt.Println("  Ctrl+N         New file")
	fmt.Println("  Ctrl+O         Open file")
	fmt.Println("  Ctrl+W         Close file")
	fmt.Println("  Ctrl+S         Save file")
	fmt.Println("  Ctrl+Q         Quit")
	fmt.Println("  Ctrl+Z         Undo")
	fmt.Println("  Ctrl+Y         Redo")
	fmt.Println("  Ctrl+X         Cut")
	fmt.Println("  Ctrl+C         Copy")
	fmt.Println("  Ctrl+V         Paste")
	fmt.Println("  Ctrl+A         Select all")
	fmt.Println("  Ctrl+F         Find")
	fmt.Println("  Shift+Arrows   Select text")
	fmt.Println("  Ctrl+Arrows    Move by word")
	fmt.Println("  Home/End       Start/end of line")
	fmt.Println("  Ctrl+Home/End  Start/end of file")
	fmt.Println("  F10            Open menu")
	fmt.Println("  F1             Show help")
	fmt.Println("  Alt+F/E/H      Open File/Edit/Help menu")
	fmt.Println()
	fmt.Println("Mouse:")
	fmt.Println("  Click          Position cursor")
	fmt.Println("  Drag           Select text")
	fmt.Println("  Scroll         Scroll viewport")
}
