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

// View renders the status bar
func (s *StatusBar) View() string {
	var sb strings.Builder

	// Start with dark blue background, white text (matching menu bar style)
	sb.WriteString("\033[44;97m")

	// Left side: modified indicator + filename
	if s.modified {
		// Bright cyan for modified indicator
		sb.WriteString("\033[96;1m*\033[97;22m")
	}

	var filename string
	if s.filename == "" {
		filename = "[Untitled]"
	} else {
		filename = filepath.Base(s.filename)
	}
	sb.WriteString(filename)

	// Right side: line:col, encoding
	right := fmt.Sprintf("Ln %d, Col %d | %s", s.line, s.col, s.encoding)

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
			sb.WriteString("\033[91;1m") // Bright red, bold
			sb.WriteString(s.message)
			sb.WriteString("\033[97;22m") // Back to white, not bold
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
