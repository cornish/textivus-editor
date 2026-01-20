package ui

import (
	"strings"
	"unicode/utf8"

	"festivus/syntax"

	"github.com/mattn/go-runewidth"
)

// Viewport handles the scrollable view of the text
type Viewport struct {
	width        int
	height       int
	scrollY      int  // First visible line
	scrollX      int  // First visible column (for horizontal scrolling)
	showLineNum  bool
	wordWrap     bool
	scrollbarWidth int  // Width reserved for scrollbar (0 if disabled)
	styles       Styles
}

// NewViewport creates a new viewport
func NewViewport(styles Styles) *Viewport {
	return &Viewport{
		width:       80,
		height:      24,
		scrollY:     0,
		scrollX:     0,
		showLineNum: false,
		styles:      styles,
	}
}

// SetSize sets the viewport dimensions
func (v *Viewport) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// Width returns the viewport width
func (v *Viewport) Width() int {
	return v.width
}

// Height returns the viewport height
func (v *Viewport) Height() int {
	return v.height
}

// ScrollY returns the current vertical scroll position
func (v *Viewport) ScrollY() int {
	return v.scrollY
}

// SetScrollY sets the vertical scroll position
func (v *Viewport) SetScrollY(y int) {
	if y < 0 {
		y = 0
	}
	v.scrollY = y
}

// ScrollX returns the current horizontal scroll position
func (v *Viewport) ScrollX() int {
	return v.scrollX
}

// SetScrollX sets the horizontal scroll position
func (v *Viewport) SetScrollX(x int) {
	if x < 0 {
		x = 0
	}
	v.scrollX = x
}

// ShowLineNumbers enables or disables line numbers
func (v *Viewport) ShowLineNumbers(show bool) {
	v.showLineNum = show
}

// ShowLineNum returns whether line numbers are enabled
func (v *Viewport) ShowLineNum() bool {
	return v.showLineNum
}

// SetWordWrap enables or disables word wrap
func (v *Viewport) SetWordWrap(wrap bool) {
	v.wordWrap = wrap
}

// WordWrap returns whether word wrap is enabled
func (v *Viewport) WordWrap() bool {
	return v.wordWrap
}

// SetStyles updates the styles for runtime theme changes
func (v *Viewport) SetStyles(styles Styles) {
	v.styles = styles
}

// MoveDownVisual moves the cursor down by one visual line when word wrap is enabled.
// Returns the new line and column position.
func (v *Viewport) MoveDownVisual(lines []string, line, col int) (newLine, newCol int) {
	if !v.wordWrap || line >= len(lines) {
		// No word wrap or at end - just move to next buffer line
		if line < len(lines)-1 {
			return line + 1, col
		}
		return line, col
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		textWidth = 1
	}

	currentLine := lines[line]
	lineRunes := utf8.RuneCountInString(currentLine)

	// Which visual segment is the cursor in?
	segmentIdx := col / textWidth
	segmentCount := (lineRunes + textWidth - 1) / textWidth
	if segmentCount == 0 {
		segmentCount = 1
	}

	// If there's another segment below in the same buffer line, move there
	if segmentIdx < segmentCount-1 {
		// Move to next segment, same relative position
		newCol = (segmentIdx+1)*textWidth + (col % textWidth)
		if newCol > lineRunes {
			newCol = lineRunes
		}
		return line, newCol
	}

	// Otherwise, move to the next buffer line
	if line < len(lines)-1 {
		// Try to maintain column position within the first segment
		colInSegment := col % textWidth
		nextLineRunes := utf8.RuneCountInString(lines[line+1])
		newCol = colInSegment
		if newCol > nextLineRunes {
			newCol = nextLineRunes
		}
		return line + 1, newCol
	}

	// At end of file
	return line, col
}

// MoveUpVisual moves the cursor up by one visual line when word wrap is enabled.
// Returns the new line and column position.
func (v *Viewport) MoveUpVisual(lines []string, line, col int) (newLine, newCol int) {
	if !v.wordWrap {
		// No word wrap - just move to previous buffer line
		if line > 0 {
			return line - 1, col
		}
		return line, col
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		textWidth = 1
	}

	// Which visual segment is the cursor in?
	segmentIdx := col / textWidth

	// If we're not in the first segment, move to the previous segment
	if segmentIdx > 0 {
		// Move to previous segment, same relative position
		newCol = (segmentIdx-1)*textWidth + (col % textWidth)
		return line, newCol
	}

	// Otherwise, move to the last segment of the previous buffer line
	if line > 0 {
		prevLine := lines[line-1]
		prevLineRunes := utf8.RuneCountInString(prevLine)
		prevSegmentCount := (prevLineRunes + textWidth - 1) / textWidth
		if prevSegmentCount == 0 {
			prevSegmentCount = 1
		}

		// Move to last segment of previous line, try to maintain column position
		colInSegment := col % textWidth
		lastSegmentStart := (prevSegmentCount - 1) * textWidth
		newCol = lastSegmentStart + colInSegment
		if newCol > prevLineRunes {
			newCol = prevLineRunes
		}
		return line - 1, newCol
	}

	// At start of file
	return line, col
}

// EnsureCursorVisible scrolls the viewport to ensure the cursor is visible
func (v *Viewport) EnsureCursorVisible(cursorLine, cursorCol int) {
	// Vertical scrolling - word wrap uses visual lines
	if cursorLine < v.scrollY {
		v.scrollY = cursorLine
	}
	if cursorLine >= v.scrollY+v.height {
		v.scrollY = cursorLine - v.height + 1
	}

	// Horizontal scrolling (only when word wrap is off)
	if !v.wordWrap {
		textWidth := v.width
		if v.showLineNum {
			textWidth -= 5 // Line number width
		}

		if cursorCol < v.scrollX {
			v.scrollX = cursorCol
		}
		if cursorCol >= v.scrollX+textWidth {
			v.scrollX = cursorCol - textWidth + 1
		}
	} else {
		v.scrollX = 0 // No horizontal scroll with word wrap
	}
}

// EnsureCursorVisibleWrapped scrolls the viewport to ensure cursor is visible (word-wrap aware)
// lines parameter is needed to calculate visual line positions
func (v *Viewport) EnsureCursorVisibleWrapped(lines []string, cursorLine, cursorCol int) {
	if !v.wordWrap {
		v.EnsureCursorVisible(cursorLine, cursorCol)
		return
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		textWidth = 1
	}

	// Calculate visual line position of cursor
	visualLine := 0
	for i := 0; i < cursorLine && i < len(lines); i++ {
		visualLine += v.countWrappedLines(lines[i], textWidth)
	}

	// Add offset within the current line based on cursor column
	if cursorLine < len(lines) {
		lineLen := utf8.RuneCountInString(lines[cursorLine])
		if lineLen > 0 && cursorCol > 0 {
			visualLine += cursorCol / textWidth
		}
	}

	// Scroll to show cursor
	if visualLine < v.scrollY {
		v.scrollY = visualLine
	}
	if visualLine >= v.scrollY+v.height {
		v.scrollY = visualLine - v.height + 1
	}

	v.scrollX = 0 // No horizontal scroll with word wrap
}

// LineNumberWidth returns the width of the line number column
func (v *Viewport) LineNumberWidth() int {
	if v.showLineNum {
		return 5
	}
	return 0
}

// SetScrollbarWidth sets the width reserved for the scrollbar
func (v *Viewport) SetScrollbarWidth(width int) {
	if width < 0 {
		width = 0
	}
	v.scrollbarWidth = width
}

// ScrollbarWidth returns the width reserved for the scrollbar
func (v *Viewport) ScrollbarWidth() int {
	return v.scrollbarWidth
}

// TextWidth returns the width available for text (viewport width minus line numbers and scrollbar)
func (v *Viewport) TextWidth() int {
	return v.width - v.LineNumberWidth() - v.scrollbarWidth
}

// CountVisualLines returns the total number of visual lines when word wrap is enabled
func (v *Viewport) CountVisualLines(lines []string) int {
	if !v.wordWrap {
		return len(lines)
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		return len(lines)
	}

	total := 0
	for _, line := range lines {
		lineLen := len([]rune(line))
		if lineLen == 0 {
			total++
		} else {
			total += (lineLen + textWidth - 1) / textWidth
		}
	}
	return total
}

// VisualLineToBufferLine converts a visual line index to a buffer line index
// Returns the buffer line and the wrap offset within that line
func (v *Viewport) VisualLineToBufferLine(lines []string, visualLine int) (bufferLine int, wrapOffset int) {
	if !v.wordWrap || visualLine < 0 {
		if visualLine >= len(lines) {
			return len(lines) - 1, 0
		}
		if visualLine < 0 {
			return 0, 0
		}
		return visualLine, 0
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		return visualLine, 0
	}

	currentVisual := 0
	for i, line := range lines {
		lineLen := len([]rune(line))
		var linesForThis int
		if lineLen == 0 {
			linesForThis = 1
		} else {
			linesForThis = (lineLen + textWidth - 1) / textWidth
		}

		if currentVisual+linesForThis > visualLine {
			// The visual line is within this buffer line
			return i, visualLine - currentVisual
		}
		currentVisual += linesForThis
	}

	// Past the end - return last line
	if len(lines) > 0 {
		return len(lines) - 1, 0
	}
	return 0, 0
}

// RenderLine renders a single line with optional selection highlighting
type SelectionRange struct {
	Start int // Start column (inclusive)
	End   int // End column (exclusive), -1 for end of line
}

// Render renders the visible portion of the text
// lineColors is an optional map of line index to color spans for syntax highlighting
func (v *Viewport) Render(lines []string, cursorLine, cursorCol int, selection map[int]SelectionRange, lineColors map[int][]syntax.ColorSpan) string {
	if v.wordWrap {
		return v.renderWrapped(lines, cursorLine, cursorCol, selection, lineColors)
	}
	return v.renderNoWrap(lines, cursorLine, cursorCol, selection, lineColors)
}

// renderNoWrap renders without word wrap (original behavior)
func (v *Viewport) renderNoWrap(lines []string, cursorLine, cursorCol int, selection map[int]SelectionRange, lineColors map[int][]syntax.ColorSpan) string {
	var sb strings.Builder

	endLine := v.scrollY + v.height
	if endLine > len(lines) {
		endLine = len(lines)
	}

	for lineIdx := v.scrollY; lineIdx < endLine; lineIdx++ {
		if lineIdx > v.scrollY {
			sb.WriteString("\n")
		}

		// Line number
		if v.showLineNum {
			lineNumStyle := v.styles.LineNumber
			if lineIdx == cursorLine {
				lineNumStyle = v.styles.LineNumberActive
			}
			sb.WriteString(lineNumStyle.Render(padLeft(itoa(lineIdx+1), 4)))
		}

		// Line content
		line := ""
		if lineIdx < len(lines) {
			line = lines[lineIdx]
		}

		// Get syntax colors for this line
		var colors []syntax.ColorSpan
		if lineColors != nil {
			colors = lineColors[lineIdx]
		}

		// Apply horizontal scroll
		displayLine := v.renderLineContent(line, lineIdx, cursorLine, cursorCol, selection, colors)
		sb.WriteString(displayLine)
	}

	// Fill remaining lines if buffer is shorter than viewport
	textWidth := v.TextWidth()
	for lineIdx := endLine; lineIdx < v.scrollY+v.height; lineIdx++ {
		if lineIdx > v.scrollY || endLine > v.scrollY {
			sb.WriteString("\n")
		}
		if v.showLineNum {
			sb.WriteString(v.styles.LineNumber.Render("    "))
		}
		// Render ~ and pad to full width
		sb.WriteString(v.styles.Subtle.Render("~"))
		if textWidth > 1 {
			sb.WriteString(strings.Repeat(" ", textWidth-1))
		}
	}

	return sb.String()
}

// renderWrapped renders with word wrap enabled
func (v *Viewport) renderWrapped(lines []string, cursorLine, cursorCol int, selection map[int]SelectionRange, lineColors map[int][]syntax.ColorSpan) string {
	var sb strings.Builder
	textWidth := v.TextWidth()
	visualLineCount := 0

	// Skip lines until we reach scrollY visual lines
	logicalLine := 0
	visualLinesSkipped := 0
	startOffset := 0

	// Only do the counting loop if we're scrolled down
	if v.scrollY > 0 {
		// First, count visual lines to find where to start
		for logicalLine < len(lines) && visualLinesSkipped < v.scrollY {
			line := lines[logicalLine]
			wrappedCount := v.countWrappedLines(line, textWidth)
			if visualLinesSkipped+wrappedCount > v.scrollY {
				// Start partway through this logical line
				break
			}
			visualLinesSkipped += wrappedCount
			logicalLine++
		}

		// Calculate offset within the current logical line
		startOffset = v.scrollY - visualLinesSkipped
	}

	// Render visible lines
	for visualLineCount < v.height && logicalLine < len(lines) {
		line := lines[logicalLine]
		sel := selection[logicalLine]
		wrappedLines := v.wrapLine(line, textWidth)

		// Get syntax colors for this line
		var colors []syntax.ColorSpan
		if lineColors != nil {
			colors = lineColors[logicalLine]
		}

		for wrapIdx := 0; wrapIdx < len(wrappedLines) && visualLineCount < v.height; wrapIdx++ {
			// Skip lines before our start offset
			if logicalLine == 0 || visualLinesSkipped < v.scrollY {
				if wrapIdx < startOffset {
					continue
				}
			}

			if visualLineCount > 0 {
				sb.WriteString("\n")
			}

			// Line number (only on first wrapped segment)
			if v.showLineNum {
				if wrapIdx == 0 {
					lineNumStyle := v.styles.LineNumber
					if logicalLine == cursorLine {
						lineNumStyle = v.styles.LineNumberActive
					}
					sb.WriteString(lineNumStyle.Render(padLeft(itoa(logicalLine+1), 4)))
				} else {
					sb.WriteString(v.styles.LineNumber.Render("    "))
				}
			}

			// Calculate the column range for this wrapped segment
			segmentStartCol := wrapIdx * textWidth

			// Render the wrapped segment
			content := v.renderWrappedSegment(wrappedLines[wrapIdx], logicalLine, segmentStartCol,
				cursorLine, cursorCol, sel, textWidth, colors)
			sb.WriteString(content)

			visualLineCount++
		}

		logicalLine++
		startOffset = 0 // Reset for subsequent lines
	}

	// Fill remaining lines
	for visualLineCount < v.height {
		if visualLineCount > 0 {
			sb.WriteString("\n")
		}
		if v.showLineNum {
			sb.WriteString(v.styles.LineNumber.Render("    "))
		}
		// Render ~ and pad to full width
		sb.WriteString(v.styles.Subtle.Render("~"))
		if textWidth > 1 {
			sb.WriteString(strings.Repeat(" ", textWidth-1))
		}
		visualLineCount++
	}

	return sb.String()
}

// countWrappedLines returns how many visual lines a logical line takes
func (v *Viewport) countWrappedLines(line string, textWidth int) int {
	if textWidth <= 0 {
		return 1
	}
	lineLen := utf8.RuneCountInString(line)
	if lineLen == 0 {
		return 1
	}
	return (lineLen + textWidth - 1) / textWidth
}

// wrapLine splits a line into wrapped segments
func (v *Viewport) wrapLine(line string, textWidth int) []string {
	if textWidth <= 0 {
		return []string{line}
	}
	runes := []rune(line)
	if len(runes) == 0 {
		return []string{""}
	}

	var segments []string
	for i := 0; i < len(runes); i += textWidth {
		end := i + textWidth
		if end > len(runes) {
			end = len(runes)
		}
		segments = append(segments, string(runes[i:end]))
	}
	return segments
}

// renderWrappedSegment renders a single wrapped segment of a line
func (v *Viewport) renderWrappedSegment(segment string, lineIdx, segmentStartCol, cursorLine, cursorCol int, sel SelectionRange, textWidth int, colors []syntax.ColorSpan) string {
	var sb strings.Builder
	runes := []rune(segment)

	for i, r := range runes {
		col := segmentStartCol + i
		isCursor := lineIdx == cursorLine && col == cursorCol
		isSelected := sel.Start <= col && (sel.End == -1 || col < sel.End)

		char := string(r)
		if r == '\t' {
			char = "    "
		}

		if isCursor {
			sb.WriteString(v.styles.Cursor.Render(char))
		} else if isSelected {
			sb.WriteString(v.styles.Selection.Render(char))
		} else {
			// Apply syntax color if available
			syntaxColor := syntax.ColorAt(colors, col)
			if syntaxColor != "" {
				sb.WriteString(syntaxColor)
				sb.WriteString(char)
				sb.WriteString("\033[0m") // Reset
			} else {
				sb.WriteString(char)
			}
		}
	}

	// Cursor at end of segment
	segmentEndCol := segmentStartCol + len(runes)
	renderedCursorAtEnd := false
	if lineIdx == cursorLine && cursorCol == segmentEndCol && segmentEndCol%textWidth == 0 && len(runes) == textWidth {
		// Cursor is at wrap point, don't show here
	} else if lineIdx == cursorLine && cursorCol >= segmentStartCol && cursorCol <= segmentEndCol && len(runes) < textWidth {
		if cursorCol == segmentEndCol {
			sb.WriteString(v.styles.Cursor.Render(" "))
			renderedCursorAtEnd = true
		}
	}

	// Pad to full width (account for cursor space if rendered)
	contentLen := len(runes)
	if renderedCursorAtEnd {
		contentLen++
	}
	if contentLen < textWidth {
		sb.WriteString(strings.Repeat(" ", textWidth-contentLen))
	}

	return v.styles.Editor.Render(sb.String())
}

// renderLineContent renders a single line's content with selection and cursor
func (v *Viewport) renderLineContent(line string, lineIdx, cursorLine, cursorCol int, selection map[int]SelectionRange, colors []syntax.ColorSpan) string {
	textWidth := v.TextWidth()

	// Convert line to runes for proper unicode handling
	runes := []rune(line)

	var sb strings.Builder

	// Apply horizontal scroll - find the actual column position
	visibleStart := v.scrollX

	visualCol := 0
	runeIdx := 0

	// Skip to scroll position (using visual columns for scrolling)
	for runeIdx < len(runes) && visualCol < visibleStart {
		r := runes[runeIdx]
		if r == '\t' {
			visualCol += 4
		} else {
			visualCol += runewidth.RuneWidth(r)
		}
		runeIdx++
	}

	// Get selection range for this line
	sel, hasSelection := selection[lineIdx]

	// Render visible portion
	outputCol := 0
	for runeIdx < len(runes) && outputCol < textWidth {
		r := runes[runeIdx]
		rw := runewidth.RuneWidth(r)

		char := string(r)
		if r == '\t' {
			char = "    " // Render tab as 4 spaces
			rw = 4
		}

		if outputCol+rw > textWidth {
			break
		}

		// Check if this position should be rendered with cursor or selection
		// Use runeIdx for comparison since cursorCol is the character index
		isCursor := lineIdx == cursorLine && runeIdx == cursorCol
		isSelected := hasSelection && runeIdx >= sel.Start && (sel.End == -1 || runeIdx < sel.End)

		if isCursor {
			sb.WriteString(v.styles.Cursor.Render(char))
		} else if isSelected {
			sb.WriteString(v.styles.Selection.Render(char))
		} else {
			// Apply syntax color if available
			syntaxColor := syntax.ColorAt(colors, runeIdx)
			if syntaxColor != "" {
				sb.WriteString(syntaxColor)
				sb.WriteString(char)
				sb.WriteString("\033[0m") // Reset
			} else {
				sb.WriteString(char)
			}
		}

		visualCol += rw
		outputCol += rw
		runeIdx++
	}

	// Render cursor at end of line if needed
	if lineIdx == cursorLine && runeIdx == cursorCol {
		sb.WriteString(v.styles.Cursor.Render(" "))
		outputCol++
	} else if hasSelection && runeIdx >= sel.Start && (sel.End == -1 || runeIdx < sel.End) {
		// Selection extends past end of line
		sb.WriteString(v.styles.Selection.Render(" "))
		outputCol++
	}

	// Pad to full width
	if outputCol < textWidth {
		padding := textWidth - outputCol
		sb.WriteString(strings.Repeat(" ", padding))
	}

	return v.styles.Editor.Render(sb.String())
}

// ScrollUp scrolls the viewport up by one line
func (v *Viewport) ScrollUp() {
	if v.scrollY > 0 {
		v.scrollY--
	}
}

// ScrollDown scrolls the viewport down by one line
func (v *Viewport) ScrollDown(totalLines int) {
	if v.scrollY < totalLines-v.height {
		v.scrollY++
	}
}

// ScrollDownWrapped scrolls the viewport down (word-wrap aware)
func (v *Viewport) ScrollDownWrapped(lines []string) {
	maxScroll := v.totalVisualLines(lines) - v.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollY < maxScroll {
		v.scrollY++
	}
}

// PageUp scrolls up by one page
func (v *Viewport) PageUp() {
	v.scrollY -= v.height
	if v.scrollY < 0 {
		v.scrollY = 0
	}
}

// PageDown scrolls down by one page
func (v *Viewport) PageDown(totalLines int) {
	v.scrollY += v.height
	maxScroll := totalLines - v.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollY > maxScroll {
		v.scrollY = maxScroll
	}
}

// PageDownWrapped scrolls down by one page (word-wrap aware)
func (v *Viewport) PageDownWrapped(lines []string) {
	v.scrollY += v.height
	maxScroll := v.totalVisualLines(lines) - v.height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if v.scrollY > maxScroll {
		v.scrollY = maxScroll
	}
}

// totalVisualLines calculates the total number of visual lines with word wrap
func (v *Viewport) totalVisualLines(lines []string) int {
	if !v.wordWrap {
		return len(lines)
	}
	textWidth := v.TextWidth()
	if textWidth <= 0 {
		return len(lines)
	}
	total := 0
	for _, line := range lines {
		total += v.countWrappedLines(line, textWidth)
	}
	return total
}

// Helper functions

func padLeft(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	negative := n < 0
	if negative {
		n = -n
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// PositionFromClick converts a click position to buffer line and column
func (v *Viewport) PositionFromClick(x, y int) (line, col int) {
	line = v.scrollY + y
	col = v.scrollX + x - v.LineNumberWidth()
	if col < 0 {
		col = 0
	}
	return
}

// PositionFromClickWrapped converts a click position to buffer line and column (word-wrap aware)
func (v *Viewport) PositionFromClickWrapped(lines []string, x, y int) (line, col int) {
	if !v.wordWrap {
		return v.PositionFromClick(x, y)
	}

	textWidth := v.TextWidth()
	if textWidth <= 0 {
		textWidth = 1
	}

	// Calculate which visual line was clicked
	targetVisualLine := v.scrollY + y

	// Find the logical line and offset within it
	visualLine := 0
	for logicalLine := 0; logicalLine < len(lines); logicalLine++ {
		lineWrappedCount := v.countWrappedLines(lines[logicalLine], textWidth)

		if visualLine+lineWrappedCount > targetVisualLine {
			// Click is within this logical line
			line = logicalLine
			// Calculate which wrapped segment and column
			segmentIndex := targetVisualLine - visualLine
			col = segmentIndex*textWidth + (x - v.LineNumberWidth())
			if col < 0 {
				col = 0
			}
			// Clamp to line length
			lineLen := utf8.RuneCountInString(lines[logicalLine])
			if col > lineLen {
				col = lineLen
			}
			return
		}
		visualLine += lineWrappedCount
	}

	// Click is past end of file
	if len(lines) > 0 {
		line = len(lines) - 1
		col = utf8.RuneCountInString(lines[line])
	}
	return
}
