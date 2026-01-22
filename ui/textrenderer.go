package ui

import (
	"strings"
	"unicode/utf8"

	"github.com/cornish/textivus-editor/syntax"
	"github.com/mattn/go-runewidth"
)

// TextRenderer renders the main text content column.
// This is the flexible column that displays document content with
// syntax highlighting, cursor, and selection.
type TextRenderer struct {
	styles Styles
}

// NewTextRenderer creates a new text renderer.
func NewTextRenderer(styles Styles) *TextRenderer {
	return &TextRenderer{styles: styles}
}

// SetStyles updates the styles for runtime theme changes.
func (r *TextRenderer) SetStyles(styles Styles) {
	r.styles = styles
}

// Render implements ColumnRenderer.
// Renders document text with syntax highlighting, cursor, and selection.
func (r *TextRenderer) Render(width, height int, state *RenderState) []string {
	if width <= 0 || height <= 0 || state == nil {
		rows := make([]string, height)
		for i := range rows {
			rows[i] = strings.Repeat(" ", width)
		}
		return rows
	}

	if state.WordWrap {
		return r.renderWrapped(width, height, state)
	}
	return r.renderNoWrap(width, height, state)
}

// renderNoWrap renders without word wrap.
func (r *TextRenderer) renderNoWrap(width, height int, state *RenderState) []string {
	rows := make([]string, height)

	endLine := state.ScrollY + height
	if endLine > len(state.Lines) {
		endLine = len(state.Lines)
	}

	for row := 0; row < height; row++ {
		lineIdx := state.ScrollY + row

		if lineIdx < len(state.Lines) {
			line := state.Lines[lineIdx]

			// Get syntax colors for this line
			var colors []syntax.ColorSpan
			if state.LineColors != nil {
				colors = state.LineColors[lineIdx]
			}

			// Render line content with selection and cursor
			rows[row] = r.renderLineContent(line, lineIdx, width, state, colors)
		} else {
			// Past end of file - render empty line marker
			rows[row] = r.renderEmptyLine(width)
		}
	}

	return rows
}

// renderWrapped renders with word wrap enabled.
func (r *TextRenderer) renderWrapped(width, height int, state *RenderState) []string {
	rows := make([]string, height)
	visualLineCount := 0

	tabWidth := state.TabWidth
	if tabWidth <= 0 {
		tabWidth = 4
	}

	// Skip lines until we reach scrollY visual lines
	logicalLine := 0
	visualLinesSkipped := 0
	startOffset := 0

	if state.ScrollY > 0 {
		for logicalLine < len(state.Lines) && visualLinesSkipped < state.ScrollY {
			line := state.Lines[logicalLine]
			wrappedCount := countWrappedLinesLocal(line, width, tabWidth)
			if visualLinesSkipped+wrappedCount > state.ScrollY {
				break
			}
			visualLinesSkipped += wrappedCount
			logicalLine++
		}
		startOffset = state.ScrollY - visualLinesSkipped
	}

	// Render visible lines
	for visualLineCount < height && logicalLine < len(state.Lines) {
		line := state.Lines[logicalLine]
		sel := state.Selection[logicalLine]
		wrappedLines := wrapLineLocal(line, width, tabWidth)

		var colors []syntax.ColorSpan
		if state.LineColors != nil {
			colors = state.LineColors[logicalLine]
		}

		// Track starting column for each wrapped segment
		segmentStartCol := 0
		for wrapIdx := 0; wrapIdx < len(wrappedLines) && visualLineCount < height; wrapIdx++ {
			if logicalLine == 0 || visualLinesSkipped < state.ScrollY {
				if wrapIdx < startOffset {
					segmentStartCol += utf8.RuneCountInString(wrappedLines[wrapIdx])
					continue
				}
			}

			rows[visualLineCount] = r.renderWrappedSegment(
				wrappedLines[wrapIdx], logicalLine, segmentStartCol,
				state.CursorLine, state.CursorCol, sel, width, tabWidth, colors,
			)
			visualLineCount++
			segmentStartCol += utf8.RuneCountInString(wrappedLines[wrapIdx])
		}

		logicalLine++
		startOffset = 0
	}

	// Fill remaining lines with empty markers
	for visualLineCount < height {
		rows[visualLineCount] = r.renderEmptyLine(width)
		visualLineCount++
	}

	return rows
}

// renderLineContent renders a single line's content with selection and cursor (no wrap).
func (r *TextRenderer) renderLineContent(line string, lineIdx, width int, state *RenderState, colors []syntax.ColorSpan) string {
	runes := []rune(line)
	var sb strings.Builder

	// Get ANSI codes for cursor and selection
	ui := r.styles.Theme.UI
	cursorCode := "\033[7m" // Reverse video for cursor
	selectionBg := ColorToANSIBg(ui.SelectionBg)
	selectionFg := ColorToANSIFg(ui.SelectionFg)
	resetCode := "\033[0m"

	// Apply horizontal scroll
	visibleStart := state.ScrollX
	visualCol := 0
	runeIdx := 0

	// Skip to scroll position
	tabWidth := state.TabWidth
	if tabWidth <= 0 {
		tabWidth = 4
	}
	for runeIdx < len(runes) && visualCol < visibleStart {
		ru := runes[runeIdx]
		if ru == '\t' {
			visualCol += tabWidth
		} else {
			visualCol += runewidth.RuneWidth(ru)
		}
		runeIdx++
	}

	// Get selection range for this line
	sel, hasSelection := state.Selection[lineIdx]

	// Render visible portion
	outputCol := 0
	for runeIdx < len(runes) && outputCol < width {
		ru := runes[runeIdx]
		rw := runewidth.RuneWidth(ru)

		char := string(ru)
		if ru == '\t' {
			char = strings.Repeat(" ", tabWidth) // Render tab as spaces
			rw = tabWidth
		}

		if outputCol+rw > width {
			break
		}

		isCursor := lineIdx == state.CursorLine && runeIdx == state.CursorCol
		isSelected := hasSelection && runeIdx >= sel.Start && (sel.End == -1 || runeIdx < sel.End)

		if isCursor {
			sb.WriteString(cursorCode)
			sb.WriteString(char)
			sb.WriteString(resetCode)
		} else if isSelected {
			sb.WriteString(selectionBg)
			sb.WriteString(selectionFg)
			sb.WriteString(char)
			sb.WriteString(resetCode)
		} else {
			syntaxColor := syntax.ColorAt(colors, runeIdx)
			if syntaxColor != "" {
				sb.WriteString(syntaxColor)
				sb.WriteString(char)
				sb.WriteString(resetCode)
			} else {
				sb.WriteString(char)
			}
		}

		visualCol += rw
		outputCol += rw
		runeIdx++
	}

	// Render cursor at end of line if needed
	if lineIdx == state.CursorLine && runeIdx == state.CursorCol {
		sb.WriteString(cursorCode)
		sb.WriteString(" ")
		sb.WriteString(resetCode)
		outputCol++
	} else if hasSelection && runeIdx >= sel.Start && (sel.End == -1 || runeIdx < sel.End) {
		sb.WriteString(selectionBg)
		sb.WriteString(selectionFg)
		sb.WriteString(" ")
		sb.WriteString(resetCode)
		outputCol++
	}

	// Pad to full width
	if outputCol < width {
		padding := width - outputCol
		sb.WriteString(strings.Repeat(" ", padding))
	}

	return sb.String()
}

// renderWrappedSegment renders a single wrapped segment of a line.
func (r *TextRenderer) renderWrappedSegment(segment string, lineIdx, segmentStartCol, cursorLine, cursorCol int, sel SelectionRange, width, tabWidth int, colors []syntax.ColorSpan) string {
	var sb strings.Builder
	runes := []rune(segment)

	// Get ANSI codes for cursor and selection
	ui := r.styles.Theme.UI
	cursorCode := "\033[7m" // Reverse video for cursor
	selectionBg := ColorToANSIBg(ui.SelectionBg)
	selectionFg := ColorToANSIFg(ui.SelectionFg)
	resetCode := "\033[0m"

	if tabWidth <= 0 {
		tabWidth = 4
	}

	outputCol := 0
	for i, ru := range runes {
		col := segmentStartCol + i
		isCursor := lineIdx == cursorLine && col == cursorCol
		isSelected := sel.Start <= col && (sel.End == -1 || col < sel.End)

		char := string(ru)
		charWidth := runewidth.RuneWidth(ru)
		if ru == '\t' {
			char = strings.Repeat(" ", tabWidth)
			charWidth = tabWidth
		}

		if isCursor {
			sb.WriteString(cursorCode)
			sb.WriteString(char)
			sb.WriteString(resetCode)
		} else if isSelected {
			sb.WriteString(selectionBg)
			sb.WriteString(selectionFg)
			sb.WriteString(char)
			sb.WriteString(resetCode)
		} else {
			syntaxColor := syntax.ColorAt(colors, col)
			if syntaxColor != "" {
				sb.WriteString(syntaxColor)
				sb.WriteString(char)
				sb.WriteString(resetCode)
			} else {
				sb.WriteString(char)
			}
		}
		outputCol += charWidth
	}

	// Cursor at end of segment
	segmentEndCol := segmentStartCol + len(runes)
	if lineIdx == cursorLine && cursorCol == segmentEndCol && segmentEndCol%width == 0 && len(runes) == width {
		// Cursor is at wrap point, don't show here
	} else if lineIdx == cursorLine && cursorCol >= segmentStartCol && cursorCol <= segmentEndCol && outputCol < width {
		if cursorCol == segmentEndCol {
			sb.WriteString(cursorCode)
			sb.WriteString(" ")
			sb.WriteString(resetCode)
			outputCol++
		}
	}

	// Pad to full width
	if outputCol < width {
		sb.WriteString(strings.Repeat(" ", width-outputCol))
	}

	return sb.String()
}

// renderEmptyLine renders an empty line marker (~).
func (r *TextRenderer) renderEmptyLine(width int) string {
	var sb strings.Builder
	// Use dim gray for empty line marker
	sb.WriteString("\033[90m") // Dim gray
	sb.WriteString("~")
	sb.WriteString("\033[0m")
	if width > 1 {
		sb.WriteString(strings.Repeat(" ", width-1))
	}
	return sb.String()
}

// Helper functions (local copies to avoid dependency issues)

// countWrappedLinesLocal counts how many visual lines a buffer line takes.
// Accounts for tabs and wide characters.
func countWrappedLinesLocal(line string, width, tabWidth int) int {
	if width <= 0 {
		return 1
	}
	visualWidth := calculateVisualWidth(line, tabWidth)
	if visualWidth == 0 {
		return 1
	}
	return (visualWidth + width - 1) / width
}

// wrapLineLocal splits a line into segments that fit within width visual columns.
// Accounts for tabs and wide characters.
func wrapLineLocal(line string, width, tabWidth int) []string {
	if width <= 0 {
		return []string{line}
	}
	if tabWidth <= 0 {
		tabWidth = 4
	}
	runes := []rune(line)
	if len(runes) == 0 {
		return []string{""}
	}

	var segments []string
	var currentSegment strings.Builder
	currentWidth := 0

	for _, r := range runes {
		charWidth := runewidth.RuneWidth(r)
		if r == '\t' {
			charWidth = tabWidth
		}

		if currentWidth+charWidth > width {
			// Start a new segment
			segments = append(segments, currentSegment.String())
			currentSegment.Reset()
			currentWidth = 0
		}

		currentSegment.WriteRune(r)
		currentWidth += charWidth
	}

	// Don't forget the last segment
	if currentSegment.Len() > 0 {
		segments = append(segments, currentSegment.String())
	}

	if len(segments) == 0 {
		return []string{""}
	}
	return segments
}

// calculateVisualWidth returns the visual width of a string,
// accounting for tabs and wide characters.
func calculateVisualWidth(s string, tabWidth int) int {
	if tabWidth <= 0 {
		tabWidth = 4
	}
	width := 0
	for _, r := range s {
		if r == '\t' {
			width += tabWidth
		} else {
			width += runewidth.RuneWidth(r)
		}
	}
	return width
}
