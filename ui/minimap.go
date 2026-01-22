package ui

import (
	"strings"
)

// MinimapController is an interface for minimap renderers.
// Both the braille-based MinimapRenderer and KittyMinimapRenderer implement this.
type MinimapController interface {
	ColumnRenderer
	SetStyles(styles Styles)
	SetEnabled(enabled bool)
	IsEnabled() bool
	Toggle() bool
	GetMetrics(viewportHeight int, state *RenderState) MinimapMetrics
	RowToVisualLine(row int, metrics MinimapMetrics) int
	ClearImage() string                                                              // Returns escape sequence to clear graphics (Kitty only, empty for braille)
	GetKittySequence(width, height, xOffset, yOffset int, state *RenderState) string // Kitty graphics overlay
}

// MinimapRenderer renders a braille-based minimap of the document.
// Standard width is 8 (1 viewport indicator + 6 braille chars + 1 space).
//
// === MINIMAP SPECIFICATION (TODO: implement) ===
//
// Vertical mapping:
//   - 1 braille dot row = 1 visual line (respects word wrap)
//   - Each braille character = 4 visual lines (braille has 4 dot rows)
//   - Minimap height = ceil(total visual lines / 4)
//   - Minimap may be shorter or taller than viewport - not scaled to fit
//
// Horizontal mapping:
//   - 1 braille dot column = 5 source characters
//   - Each braille character = 10 source characters (2 dot columns × 5 chars)
//   - 6 braille characters = 60 source characters max
//   - Lines longer than 60 chars are truncated (not scaled)
//
// Fill logic:
//   - A dot is ON if there are >= 3 non-whitespace characters in that
//     5-character span (i.e., less than 2 char widths of whitespace)
//
// Viewport indicator:
//   - Option A: Current vertical bar │ on left side
//   - Option B: Reverse video on braille chars within viewport range
//
// Mouse interaction:
//   - Clicking on minimap navigates viewport to that location
type MinimapRenderer struct {
	styles  Styles
	enabled bool
}

// NewMinimapRenderer creates a new minimap renderer.
func NewMinimapRenderer(styles Styles) *MinimapRenderer {
	return &MinimapRenderer{
		styles:  styles,
		enabled: false, // Disabled by default
	}
}

// SetStyles updates the styles for runtime theme changes.
func (r *MinimapRenderer) SetStyles(styles Styles) {
	r.styles = styles
}

// SetEnabled enables or disables the minimap.
func (r *MinimapRenderer) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// IsEnabled returns whether the minimap is enabled.
func (r *MinimapRenderer) IsEnabled() bool {
	return r.enabled
}

// Toggle toggles the minimap on/off.
func (r *MinimapRenderer) Toggle() bool {
	r.enabled = !r.enabled
	return r.enabled
}

// Render implements ColumnRenderer.
// Returns braille representation of the document with viewport indicator.
func (r *MinimapRenderer) Render(width, height int, state *RenderState) []string {
	if !r.enabled || width <= 0 || height <= 0 || state == nil {
		rows := make([]string, height)
		for i := range rows {
			rows[i] = strings.Repeat(" ", width)
		}
		return rows
	}

	// Layout: [indicator][braille chars][space]
	// indicator: 1 char showing if this row is in visible viewport
	// braille: width-2 chars of document content
	// space: 1 char padding on right
	brailleWidth := width - 2
	if brailleWidth < 1 {
		brailleWidth = 1
	}

	// Generate visual lines (respecting word wrap)
	// Each visual line is what actually displays on one screen row
	textWidth := 80 // Approximate text column width for wrapping
	visualLines := r.generateVisualLines(state.Lines, state.WordWrap, textWidth)
	totalVisualLines := len(visualLines)
	if totalVisualLines == 0 {
		totalVisualLines = 1
		visualLines = []string{""}
	}

	// Minimap height = ceil(totalVisualLines / 4)
	// Each braille char represents 4 visual lines
	minimapHeight := (totalVisualLines + 3) / 4

	// Viewport indicator range (in visual lines)
	visibleStart := state.ScrollY
	visibleEnd := state.ScrollY + height

	// Get theme colors
	ui := r.styles.Theme.UI
	indicatorColor := ColorToANSIFg(ui.MinimapIndicator)
	textColor := ColorToANSIFg(ui.MinimapText)
	resetCode := "\033[0m"

	rows := make([]string, height)

	// If minimap is taller than viewport, we need to scroll it
	// Center the minimap view on the current viewport position
	minimapScrollOffset := 0
	if minimapHeight > height {
		// Calculate which minimap row corresponds to viewport center
		viewportCenterLine := state.ScrollY + height/2
		minimapCenterRow := viewportCenterLine / 4
		minimapScrollOffset = minimapCenterRow - height/2
		if minimapScrollOffset < 0 {
			minimapScrollOffset = 0
		}
		if minimapScrollOffset > minimapHeight-height {
			minimapScrollOffset = minimapHeight - height
		}
	}

	for row := 0; row < height; row++ {
		var sb strings.Builder

		minimapRow := row + minimapScrollOffset
		if minimapRow >= minimapHeight {
			// Past end of minimap - empty row
			sb.WriteString(strings.Repeat(" ", width))
			rows[row] = sb.String()
			continue
		}

		// Which visual lines does this minimap row represent?
		// Each minimap row = 4 visual lines (braille has 4 dot rows)
		visualLineStart := minimapRow * 4
		visualLineEnd := visualLineStart + 4
		if visualLineEnd > totalVisualLines {
			visualLineEnd = totalVisualLines
		}

		// Viewport indicator: is any part of this minimap row in the viewport?
		inViewport := visualLineStart < visibleEnd && visualLineEnd > visibleStart
		if inViewport {
			sb.WriteString(indicatorColor)
			sb.WriteString("│")
			sb.WriteString(resetCode)
		} else {
			sb.WriteString(" ")
		}

		// Braille representation: get the 4 visual lines for this row
		var fourLines [4]string
		for i := 0; i < 4; i++ {
			lineIdx := visualLineStart + i
			if lineIdx < totalVisualLines {
				fourLines[i] = visualLines[lineIdx]
			} else {
				fourLines[i] = ""
			}
		}

		sb.WriteString(textColor)
		braille := r.renderBrailleChar(fourLines, brailleWidth)
		sb.WriteString(braille)
		sb.WriteString(resetCode)

		// Right padding
		sb.WriteString(" ")

		rows[row] = sb.String()
	}

	return rows
}

// generateVisualLines converts buffer lines to visual lines respecting word wrap.
func (r *MinimapRenderer) generateVisualLines(lines []string, wordWrap bool, textWidth int) []string {
	if !wordWrap || textWidth <= 0 {
		// No word wrap - visual lines = buffer lines
		return lines
	}

	var visualLines []string
	for _, line := range lines {
		lineRunes := []rune(line)
		if len(lineRunes) == 0 {
			visualLines = append(visualLines, "")
			continue
		}
		// Wrap long lines
		for len(lineRunes) > 0 {
			end := textWidth
			if end > len(lineRunes) {
				end = len(lineRunes)
			}
			visualLines = append(visualLines, string(lineRunes[:end]))
			lineRunes = lineRunes[end:]
		}
	}
	if len(visualLines) == 0 {
		visualLines = []string{""}
	}
	return visualLines
}

// renderBrailleChar renders braille characters for 4 visual lines.
// Each braille char represents 4 rows × 2 columns, where each dot column = 5 source chars.
// A dot is ON if >= 3 non-whitespace chars in that 5-char span.
func (r *MinimapRenderer) renderBrailleChar(fourLines [4]string, brailleWidth int) string {
	var result strings.Builder

	// Each braille char = 10 source chars (2 dot columns × 5 chars each)
	charsPerBraille := 10

	for col := 0; col < brailleWidth; col++ {
		srcColStart := col * charsPerBraille
		srcColMid := srcColStart + 5 // Split point between left and right dot columns

		// Build braille pattern from the 4×2 grid
		// Braille dots are numbered:
		// 1 4
		// 2 5
		// 3 6
		// 7 8
		var pattern rune = 0x2800 // Empty braille

		for rowOffset := 0; rowOffset < 4; rowOffset++ {
			lineRunes := []rune(fourLines[rowOffset])

			// Left dot column (dots 1,2,3,7) - chars [srcColStart, srcColMid)
			if hasEnoughContent(lineRunes, srcColStart, srcColMid, 3) {
				switch rowOffset {
				case 0:
					pattern |= 0x01 // dot 1
				case 1:
					pattern |= 0x02 // dot 2
				case 2:
					pattern |= 0x04 // dot 3
				case 3:
					pattern |= 0x40 // dot 7
				}
			}

			// Right dot column (dots 4,5,6,8) - chars [srcColMid, srcColMid+5)
			if hasEnoughContent(lineRunes, srcColMid, srcColMid+5, 3) {
				switch rowOffset {
				case 0:
					pattern |= 0x08 // dot 4
				case 1:
					pattern |= 0x10 // dot 5
				case 2:
					pattern |= 0x20 // dot 6
				case 3:
					pattern |= 0x80 // dot 8
				}
			}
		}

		result.WriteRune(pattern)
	}

	return result.String()
}

// hasEnoughContent checks if a line has at least `threshold` non-whitespace characters
// in the given column range [start, end).
func hasEnoughContent(lineRunes []rune, start, end, threshold int) bool {
	if start < 0 {
		start = 0
	}
	if end > len(lineRunes) {
		end = len(lineRunes)
	}
	count := 0
	for i := start; i < end; i++ {
		if i < len(lineRunes) {
			r := lineRunes[i]
			if r != ' ' && r != '\t' {
				count++
				if count >= threshold {
					return true
				}
			}
		}
	}
	return false
}

// MinimapWidth returns the standard width for the minimap column.
func MinimapWidth() int {
	return 8 // 1 indicator + 6 braille + 1 space
}

// MinimapMetrics holds metrics for mouse interaction with minimap.
type MinimapMetrics struct {
	TotalVisualLines    int // Total visual lines in document
	MinimapHeight       int // Height of minimap in rows (ceil(visual lines / 4))
	MinimapScrollOffset int // Current scroll offset of minimap view
	ViewportHeight      int // Height of viewport
}

// GetMetrics calculates minimap metrics for a given state.
func (r *MinimapRenderer) GetMetrics(viewportHeight int, state *RenderState) MinimapMetrics {
	// Generate visual lines to get accurate count
	textWidth := 80
	visualLines := r.generateVisualLines(state.Lines, state.WordWrap, textWidth)
	totalVisualLines := len(visualLines)
	if totalVisualLines == 0 {
		totalVisualLines = 1
	}

	minimapHeight := (totalVisualLines + 3) / 4

	// Calculate scroll offset (same logic as in Render)
	minimapScrollOffset := 0
	if minimapHeight > viewportHeight {
		viewportCenterLine := state.ScrollY + viewportHeight/2
		minimapCenterRow := viewportCenterLine / 4
		minimapScrollOffset = minimapCenterRow - viewportHeight/2
		if minimapScrollOffset < 0 {
			minimapScrollOffset = 0
		}
		if minimapScrollOffset > minimapHeight-viewportHeight {
			minimapScrollOffset = minimapHeight - viewportHeight
		}
	}

	return MinimapMetrics{
		TotalVisualLines:    totalVisualLines,
		MinimapHeight:       minimapHeight,
		MinimapScrollOffset: minimapScrollOffset,
		ViewportHeight:      viewportHeight,
	}
}

// RowToVisualLine converts a minimap row click to a visual line index.
// The row is relative to the viewport (0 = top of visible minimap area).
func (r *MinimapRenderer) RowToVisualLine(row int, metrics MinimapMetrics) int {
	// Account for minimap scroll offset
	minimapRow := row + metrics.MinimapScrollOffset
	// Each minimap row = 4 visual lines
	visualLine := minimapRow * 4
	if visualLine < 0 {
		return 0
	}
	if visualLine >= metrics.TotalVisualLines {
		return metrics.TotalVisualLines - 1
	}
	return visualLine
}

// ClearImage returns an empty string for braille renderer (no graphics to clear).
func (r *MinimapRenderer) ClearImage() string {
	return ""
}

// GetKittySequence returns empty for braille renderer (no Kitty graphics).
func (r *MinimapRenderer) GetKittySequence(width, height, xOffset, yOffset int, state *RenderState) string {
	return ""
}
