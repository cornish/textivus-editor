package config

import (
	"os"
	"strings"
)

// ColorMode represents the terminal color capability
type ColorMode int

const (
	Color16      ColorMode = iota // Basic 16 colors
	Color256                      // 256 color palette
	ColorTrueColor                // 24-bit true color
)

// TermCapabilities holds detected terminal capabilities
type TermCapabilities struct {
	UTF8Support   bool      // Terminal supports UTF-8
	ColorMode     ColorMode // Color capability level
	KittyGraphics bool      // Kitty graphics protocol support
}

// String returns a human-readable description of the color mode
func (c ColorMode) String() string {
	switch c {
	case Color16:
		return "16 colors"
	case Color256:
		return "256 colors"
	case ColorTrueColor:
		return "TrueColor (24-bit)"
	default:
		return "unknown"
	}
}

// DetectCapabilities detects terminal capabilities from environment variables
func DetectCapabilities() *TermCapabilities {
	caps := &TermCapabilities{
		UTF8Support:   detectUTF8Support(),
		ColorMode:     detectColorMode(),
		KittyGraphics: detectKittyGraphics(),
	}
	return caps
}

// detectUTF8Support checks if the terminal supports UTF-8
func detectUTF8Support() bool {
	// Check LC_ALL, LC_CTYPE, and LANG for UTF-8 indicators
	for _, envVar := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		val := strings.ToUpper(os.Getenv(envVar))
		if val != "" {
			if strings.Contains(val, "UTF-8") || strings.Contains(val, "UTF8") {
				return true
			}
		}
	}
	return false
}

// detectColorMode detects the terminal's color capability
func detectColorMode() ColorMode {
	// Check COLORTERM for truecolor support
	colorterm := strings.ToLower(os.Getenv("COLORTERM"))
	if colorterm == "truecolor" || colorterm == "24bit" {
		return ColorTrueColor
	}

	// Check TERM for 256-color support
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "256color") || strings.Contains(term, "256-color") {
		return Color256
	}

	// Some terminals set truecolor capability via TERM
	if strings.Contains(term, "truecolor") || strings.Contains(term, "24bit") {
		return ColorTrueColor
	}

	// Known terminals that support truecolor
	trueColorTerms := []string{"xterm-direct", "iterm2", "vte"}
	for _, t := range trueColorTerms {
		if strings.Contains(term, t) {
			return ColorTrueColor
		}
	}

	// Default to 16 colors for safety
	return Color16
}

// detectKittyGraphics checks if running in Kitty terminal with graphics support
func detectKittyGraphics() bool {
	// KITTY_WINDOW_ID is set when running inside Kitty
	return os.Getenv("KITTY_WINDOW_ID") != ""
}

// ShouldUseASCII returns true if ASCII mode should be used based on capabilities
// Takes into account both auto-detection and user override
func (c *TermCapabilities) ShouldUseASCII(override *bool) bool {
	if override != nil {
		return *override
	}
	return !c.UTF8Support
}

// ShouldUseTrueColor returns true if TrueColor should be used based on capabilities
// Takes into account both auto-detection and user override
func (c *TermCapabilities) ShouldUseTrueColor(override *bool) bool {
	if override != nil {
		return *override
	}
	return c.ColorMode == ColorTrueColor
}

// GlobalCapabilities holds the detected capabilities (set at startup)
var GlobalCapabilities *TermCapabilities

// InitCapabilities detects and stores terminal capabilities
// Should be called once at startup
func InitCapabilities() {
	GlobalCapabilities = DetectCapabilities()
}

// GetCapabilities returns the global capabilities, detecting if needed
func GetCapabilities() *TermCapabilities {
	if GlobalCapabilities == nil {
		InitCapabilities()
	}
	return GlobalCapabilities
}
