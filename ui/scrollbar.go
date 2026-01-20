package ui

import (
	"strings"
)

// Scrollbar represents a vertical scrollbar displayed on the right side of the editor
type Scrollbar struct {
	height  int
	enabled bool
	styles  Styles
}

// NewScrollbar creates a new scrollbar instance
func NewScrollbar(styles Styles) *Scrollbar {
	return &Scrollbar{
		height:  24,
		enabled: false,
		styles:  styles,
	}
}

// Width returns the scrollbar width (1 character, or 0 if disabled)
func (s *Scrollbar) Width() int {
	if !s.enabled {
		return 0
	}
	return 1
}

// SetHeight sets the scrollbar height
func (s *Scrollbar) SetHeight(height int) {
	if height > 0 {
		s.height = height
	}
}

// Height returns the scrollbar height
func (s *Scrollbar) Height() int {
	return s.height
}

// SetEnabled enables or disables the scrollbar
func (s *Scrollbar) SetEnabled(enabled bool) {
	s.enabled = enabled
}

// IsEnabled returns whether the scrollbar is enabled
func (s *Scrollbar) IsEnabled() bool {
	return s.enabled
}

// Toggle toggles the scrollbar on/off
func (s *Scrollbar) Toggle() bool {
	s.enabled = !s.enabled
	return s.enabled
}

// SetStyles updates the styles for runtime theme changes
func (s *Scrollbar) SetStyles(styles Styles) {
	s.styles = styles
}

// Render renders the scrollbar as a slice of strings, one per viewport row
// viewportStart is the first visible line, viewportHeight is the number of visible lines,
// totalLines is the total number of lines in the document
func (s *Scrollbar) Render(viewportStart, viewportHeight, totalLines int) []string {
	if !s.enabled || s.height <= 0 {
		return nil
	}

	result := make([]string, s.height)
	ui := s.styles.Theme.UI

	// Get ANSI colors
	trackColor := ColorToANSIFg(ui.ScrollbarTrack)
	thumbColor := ColorToANSIFg(ui.ScrollbarThumb)

	// Handle edge cases
	if totalLines <= 0 {
		totalLines = 1
	}
	if viewportHeight <= 0 {
		viewportHeight = 1
	}
	if viewportStart < 0 {
		viewportStart = 0
	}

	// Calculate thumb size (proportional to visible content)
	// Use int64 to avoid overflow with large files
	thumbSize := int((int64(viewportHeight) * int64(s.height)) / int64(totalLines))
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > s.height {
		thumbSize = s.height
	}

	// Calculate thumb position
	var thumbStart int
	if totalLines <= viewportHeight {
		// Everything fits - thumb fills track
		thumbStart = 0
		thumbSize = s.height
	} else {
		// Calculate position based on scroll progress (0.0 to 1.0)
		// Maximum scroll position is totalLines - viewportHeight
		maxScroll := totalLines - viewportHeight
		if maxScroll <= 0 {
			maxScroll = 1
		}

		// Clamp viewportStart to valid range
		if viewportStart > maxScroll {
			viewportStart = maxScroll
		}

		// Calculate thumb position: map scroll position to thumb range
		// thumbStart ranges from 0 to (height - thumbSize)
		thumbRange := s.height - thumbSize
		if thumbRange <= 0 {
			thumbStart = 0
		} else {
			thumbStart = int((int64(viewportStart) * int64(thumbRange)) / int64(maxScroll))
		}

		// Final bounds check
		if thumbStart < 0 {
			thumbStart = 0
		}
		if thumbStart > s.height-thumbSize {
			thumbStart = s.height - thumbSize
		}
	}

	thumbEnd := thumbStart + thumbSize
	if thumbEnd > s.height {
		thumbEnd = s.height
	}

	// Render each row
	for row := 0; row < s.height; row++ {
		var sb strings.Builder

		if row >= thumbStart && row < thumbEnd {
			sb.WriteString(thumbColor)
			sb.WriteString("┃")
		} else {
			sb.WriteString(trackColor)
			sb.WriteString("│")
		}

		sb.WriteString("\033[0m")
		result[row] = sb.String()
	}

	return result
}

// RowToLine converts a scrollbar row to the corresponding visual line index
// This is consistent with the thumb position calculation in Render
func (s *Scrollbar) RowToLine(row int, totalLines, viewportHeight int) int {
	if totalLines <= 0 || s.height <= 0 {
		return 0
	}
	if row < 0 {
		row = 0
	}
	if row >= s.height {
		row = s.height - 1
	}

	// If everything fits, always return 0
	if totalLines <= viewportHeight {
		return 0
	}

	// Calculate thumb size (same as in Render)
	thumbSize := int((int64(viewportHeight) * int64(s.height)) / int64(totalLines))
	if thumbSize < 1 {
		thumbSize = 1
	}
	if thumbSize > s.height {
		thumbSize = s.height
	}

	// Thumb range and max scroll (same as in Render)
	thumbRange := s.height - thumbSize
	maxScroll := totalLines - viewportHeight

	if thumbRange <= 0 || maxScroll <= 0 {
		return 0
	}

	// Inverse of: thumbStart = (viewportStart * thumbRange) / maxScroll
	// So: viewportStart = (thumbStart * maxScroll) / thumbRange
	// We use 'row' as the clicked position, centering on it
	scrollPos := int((int64(row) * int64(maxScroll)) / int64(thumbRange))

	// Return the line at the center of where the viewport would be
	line := scrollPos + viewportHeight/2

	if line < 0 {
		return 0
	}
	if line >= totalLines {
		return totalLines - 1
	}
	return line
}
