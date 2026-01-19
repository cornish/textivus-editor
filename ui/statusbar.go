package ui

import (
	"fmt"
	"path/filepath"
	"strings"
)

// StatusBar represents the bottom status bar
type StatusBar struct {
	filename    string
	modified    bool
	line        int
	col         int
	totalLines  int
	encoding    string
	wordCount   int
	charCount   int
	message     string // Temporary message to display
	messageType string // "info", "error", "success"
	width       int
	styles      Styles
}

// NewStatusBar creates a new status bar
func NewStatusBar(styles Styles) *StatusBar {
	return &StatusBar{
		filename:   "",
		modified:   false,
		line:       1,
		col:        1,
		totalLines: 1,
		encoding:   "UTF-8",
		styles:     styles,
	}
}

// SetFilename sets the current filename
func (s *StatusBar) SetFilename(filename string) {
	s.filename = filename
}

// SetModified sets whether the buffer has been modified
func (s *StatusBar) SetModified(modified bool) {
	s.modified = modified
}

// SetPosition sets the cursor position (1-indexed for display)
func (s *StatusBar) SetPosition(line, col int) {
	s.line = line + 1 // Convert from 0-indexed to 1-indexed
	s.col = col + 1
}

// SetTotalLines sets the total number of lines
func (s *StatusBar) SetTotalLines(total int) {
	s.totalLines = total
}

// SetEncoding sets the file encoding
func (s *StatusBar) SetEncoding(encoding string) {
	s.encoding = encoding
}

// SetCounts sets the word and character counts
func (s *StatusBar) SetCounts(words, chars int) {
	s.wordCount = words
	s.charCount = chars
}

// SetMessage sets a temporary message to display
func (s *StatusBar) SetMessage(message, msgType string) {
	s.message = message
	s.messageType = msgType
}

// ClearMessage clears the temporary message
func (s *StatusBar) ClearMessage() {
	s.message = ""
	s.messageType = ""
}

// SetWidth sets the width of the status bar
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetStyles updates the styles for runtime theme changes
func (s *StatusBar) SetStyles(styles Styles) {
	s.styles = styles
}

// View renders the status bar
func (s *StatusBar) View() string {
	var sb strings.Builder

	// Get theme colors
	ui := s.styles.Theme.UI
	normalColor := ColorToANSI(ui.StatusFg, ui.StatusBg)
	accentColor := ColorToANSIFg(ui.StatusAccent) + "\033[1m" // Bold
	errorColor := ColorToANSIFg(ui.ErrorFg) + "\033[1m"       // Bold
	resetToNormal := ColorToANSIFg(ui.StatusFg) + "\033[22m"  // Not bold

	// Start with status bar colors
	sb.WriteString(normalColor)

	// Left side: modified indicator + filename
	if s.modified {
		// Accent color for modified indicator
		sb.WriteString(accentColor + "*" + resetToNormal)
	}

	var filename string
	if s.filename == "" {
		filename = "[Untitled]"
	} else {
		filename = filepath.Base(s.filename)
	}
	sb.WriteString(filename)

	// Right side: word count, char count, line:col, encoding
	right := fmt.Sprintf("W:%d C:%d | Ln %d, Col %d | %s", s.wordCount, s.charCount, s.line, s.col, s.encoding)

	// Calculate spacing
	leftLen := len(filename)
	if s.modified {
		leftLen++
	}
	rightLen := len(right)
	centerLen := len(s.message)

	availableSpace := s.width - leftLen - rightLen
	if availableSpace < 0 {
		availableSpace = 0
	}

	// Center message if any
	if s.message != "" && centerLen+4 <= availableSpace {
		leftPad := (availableSpace - centerLen) / 2
		rightPad := availableSpace - centerLen - leftPad
		sb.WriteString(strings.Repeat(" ", leftPad))

		// Render message with appropriate color
		if s.messageType == "error" {
			sb.WriteString(errorColor)
			sb.WriteString(s.message)
			sb.WriteString(resetToNormal)
		} else {
			sb.WriteString(s.message)
		}

		sb.WriteString(strings.Repeat(" ", rightPad))
	} else {
		// No message or not enough space
		sb.WriteString(strings.Repeat(" ", availableSpace))
	}

	sb.WriteString(right)

	// Reset at end
	sb.WriteString("\033[0m")

	return sb.String()
}
