package editor

import (
	"festivus/ui"
	"strings"

	"github.com/mattn/go-runewidth"
)

// overlayLineAt overlays the dropdown line on top of the viewport line at the given offset,
// preserving viewport content on both sides of the dropdown (including ANSI color codes)
func overlayLineAt(dropLine, viewportLine string, offset int) string {
	// Calculate the visual width of the dropdown line (strip ANSI codes)
	dropWidth := visualWidth(dropLine)

	// Extract prefix and suffix from viewport line, preserving ANSI codes
	prefix := sliceAnsiString(viewportLine, 0, offset)
	suffix := sliceAnsiString(viewportLine, offset+dropWidth, -1)

	// Build the result: prefix + dropdown + suffix
	var result strings.Builder

	// Prefix: viewport content before the dropdown (or spaces if line is short)
	prefixWidth := visualWidth(prefix)
	result.WriteString(prefix)
	if prefixWidth < offset {
		// Viewport line is shorter than offset - add padding
		result.WriteString(strings.Repeat(" ", offset-prefixWidth))
	}

	// The dropdown itself
	result.WriteString(dropLine)

	// Suffix: viewport content after the dropdown (with ANSI codes preserved)
	if suffix != "" {
		result.WriteString(suffix)
	}

	return result.String()
}

// sliceAnsiString extracts a substring from an ANSI-coded string based on visual positions.
// start and end are visual column positions (0-indexed). Use end=-1 for "to the end".
// ANSI escape codes are preserved and passed through correctly.
func sliceAnsiString(s string, start, end int) string {
	var result strings.Builder
	visualPos := 0
	inEscape := false
	var escapeSeq strings.Builder

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			escapeSeq.Reset()
			escapeSeq.WriteRune(r)
			continue
		}

		if inEscape {
			escapeSeq.WriteRune(r)
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
				// Include escape sequences that appear within our range
				// or at the boundary (to preserve color state)
				if end == -1 || visualPos < end {
					if visualPos >= start || visualPos == 0 {
						result.WriteString(escapeSeq.String())
					} else if visualPos < start {
						// Escape sequence before our range - still include it
						// to maintain proper color state
						result.WriteString(escapeSeq.String())
					}
				}
			}
			continue
		}

		// Regular character - check if it's in our range
		charWidth := runewidth.RuneWidth(r)
		if visualPos >= start && (end == -1 || visualPos < end) {
			result.WriteRune(r)
		}
		visualPos += charWidth
	}

	return result.String()
}

// stripAnsi removes ANSI escape sequences from a string
func stripAnsi(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}

// visualWidth calculates the visible width of a string (ignoring ANSI codes)
func visualWidth(s string) int {
	return runewidth.StringWidth(stripAnsi(s))
}

// overlayAboutDialog overlays the about dialog centered on the viewport
func (e *Editor) overlayAboutDialog(viewportContent string) string {
	// Use the stored quote (selected when dialog opened)
	quote := e.aboutQuote
	if quote == "" {
		quote = "A Festivus for the rest of us!"
	}

	// ASCII art from festivus.txt - art is 62 chars, box is 64 for padding
	boxWidth := 64
	centerText := func(s string) string {
		sLen := len(s)
		if sLen >= boxWidth {
			// Truncate if too long
			return s[:boxWidth]
		}
		padLeft := (boxWidth - sLen) / 2
		padRight := boxWidth - sLen - padLeft
		return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
	}

	// Split quote into lines if too long (max 60 chars per line)
	maxLineWidth := 60
	var quoteLines []string
	quotedText := "\"" + quote + "\""
	if len(quotedText) <= maxLineWidth {
		// Fits on one line
		quoteLines = []string{centerText(quotedText)}
	} else {
		// Split at word boundary
		words := strings.Fields(quote)
		line1 := "\""
		line2 := ""
		for _, word := range words {
			testLine := line1
			if len(line1) > 1 {
				testLine += " "
			}
			testLine += word
			if line2 == "" && len(testLine) <= maxLineWidth {
				line1 = testLine
			} else {
				if line2 != "" {
					line2 += " "
				}
				line2 += word
			}
		}
		line2 += "\""
		quoteLines = []string{centerText(line1), centerText(line2)}
	}

	// Choose logo based on ASCII mode
	var logoLines []string
	if e.box.Lock == "*" {
		// ASCII mode - use asterisk art (64 chars wide to match boxWidth)
		logoLines = []string{
			"      *****  *****   ****  *****  ***  *   *  *   *   ****      ",
			"      *      *      *        *     *   *   *  *   *  *          ",
			"      ****   ****    ***     *     *   *   *  *   *   ***       ",
			"      *      *          *    *     *    * *   *   *      *      ",
			"      *      *****  ****     *    ***    *     ***   ****       ",
			"                                                                ",
		}
	} else {
		// Unicode mode - use block art
		logoLines = []string{
			" ███████╗███████╗███████╗████████╗██╗██╗   ██╗██╗   ██╗███████╗ ",
			" ██╔════╝██╔════╝██╔════╝╚══██╔══╝██║██║   ██║██║   ██║██╔════╝ ",
			" █████╗  █████╗  ███████╗   ██║   ██║██║   ██║██║   ██║███████╗ ",
			" ██╔══╝  ██╔══╝  ╚════██║   ██║   ██║╚██╗ ██╔╝██║   ██║╚════██║ ",
			" ██║     ███████╗███████║   ██║   ██║ ╚████╔╝ ╚██████╔╝███████║ ",
			" ╚═╝     ╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═══╝   ╚═════╝ ╚══════╝ ",
		}
	}

	aboutLines := []string{strings.Repeat(" ", boxWidth)}
	aboutLines = append(aboutLines, logoLines...)
	aboutLines = append(aboutLines,
		strings.Repeat(" ", boxWidth),
		centerText("A Text Editor for the Rest of Us"),
		strings.Repeat(" ", boxWidth),
		centerText("Version 0.1.0"),
		centerText("github.com/cornish/festivus"),
		centerText("Copyright (c) 2025"),
		strings.Repeat(" ", boxWidth),
	)
	aboutLines = append(aboutLines, quoteLines...)
	aboutLines = append(aboutLines,
		strings.Repeat(" ", boxWidth),
		centerText("Press any key to continue..."),
		strings.Repeat(" ", boxWidth),
	)
	boxHeight := len(aboutLines)

	// Calculate centering
	startX := (e.width - boxWidth) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (e.viewport.Height() - boxHeight) / 2
	if startY < 0 {
		startY = 0
	}

	viewportLines := strings.Split(viewportContent, "\n")

	// Get theme dialog colors
	themeUI := e.styles.Theme.UI
	dialogStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	resetStyle := "\033[0m"

	for i, aboutLine := range aboutLines {
		viewportY := startY + i
		if viewportY >= 0 && viewportY < len(viewportLines) {
			// Build the styled about line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(dialogStyle)
			styledLine.WriteString(aboutLine)
			styledLine.WriteString(resetStyle)

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// overlayHelpDialog overlays the help dialog centered on the viewport
func (e *Editor) overlayHelpDialog(viewportContent string) string {
	// Two-column layout for keyboard shortcuts
	boxWidth := 72
	innerWidth := boxWidth - 2 // 70
	colWidth := 33             // Each column width
	// Layout: colWidth (33) + separator "  │ " (4) + colWidth (33) = 70

	padText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw > width {
			return runewidth.Truncate(s, width, "")
		}
		return s + strings.Repeat(" ", width-sw)
	}

	centerText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw >= width {
			return runewidth.Truncate(s, width, "")
		}
		padLeft := (width - sw) / 2
		padRight := width - sw - padLeft
		return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
	}

	// Define shortcuts in two columns
	leftCol := []string{
		"  FILE",
		"  Ctrl+N       New file",
		"  Ctrl+O       Open file",
		"  Ctrl+R       Recent files",
		"  Ctrl+W       Close file",
		"  Ctrl+S       Save file",
		"  Ctrl+Q       Quit",
		"",
		"  EDIT",
		"  Ctrl+Z       Undo",
		"  Ctrl+Y       Redo",
		"  Ctrl+X       Cut",
		"  Ctrl+C       Copy",
		"  Ctrl+V       Paste",
		"  Ctrl+K       Cut line",
		"  Ctrl+A       Select all",
		"",
		"  SEARCH",
		"  Ctrl+F       Find",
		"  F3           Find next",
		"  Ctrl+H       Replace",
	}

	rightCol := []string{
		"  NAVIGATION",
		"  Arrows       Move cursor",
		"  Ctrl+Left/Right  Move by word",
		"  Home/End     Start/end of line",
		"  Ctrl+Home    Start of file",
		"  Ctrl+End     End of file",
		"  PgUp/PgDn    Page up/down",
		"  Ctrl+G       Go to line",
		"",
		"  SELECTION",
		"  Shift+Arrows    Select text",
		"  Ctrl+Shift+L/R  Select word",
		"  Shift+Home/End  Select to line",
		"  MOUSE: Click, Drag, Scroll",
	}

	// Build help lines
	var helpLines []string

	// Top border with title
	title := " Keyboard Shortcuts "
	titlePadLeft := (innerWidth - len(title)) / 2
	titlePadRight := innerWidth - len(title) - titlePadLeft
	helpLines = append(helpLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Empty line
	helpLines = append(helpLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Build two-column content
	maxRows := len(leftCol)
	if len(rightCol) > maxRows {
		maxRows = len(rightCol)
	}

	colSep := "  " + e.box.Vertical + " "
	for i := 0; i < maxRows; i++ {
		left := ""
		right := ""
		if i < len(leftCol) {
			left = leftCol[i]
		}
		if i < len(rightCol) {
			right = rightCol[i]
		}
		line := padText(left, colWidth) + colSep + padText(right, colWidth)
		helpLines = append(helpLines, e.box.Vertical+line+e.box.Vertical)
	}

	// Empty line
	helpLines = append(helpLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Options section
	helpLines = append(helpLines, e.box.Vertical+centerText("OPTIONS: Ctrl+L Line Numbers", innerWidth)+e.box.Vertical)
	helpLines = append(helpLines, e.box.Vertical+centerText("MENUS: F10 or Alt+F/E/O/H", innerWidth)+e.box.Vertical)

	// Empty line
	helpLines = append(helpLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Footer
	helpLines = append(helpLines, e.box.Vertical+centerText("Press any key to continue...", innerWidth)+e.box.Vertical)

	// Bottom border
	helpLines = append(helpLines, e.box.BottomLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.BottomRight)

	boxHeight := len(helpLines)

	// Calculate centering
	startX := (e.width - boxWidth) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (e.viewport.Height() - boxHeight) / 2
	if startY < 0 {
		startY = 0
	}

	viewportLines := strings.Split(viewportContent, "\n")

	// Get theme dialog colors
	themeUI := e.styles.Theme.UI
	dialogStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	resetStyle := "\033[0m"

	for i, helpLine := range helpLines {
		viewportY := startY + i
		if viewportY >= 0 && viewportY < len(viewportLines) {
			// Build the styled help line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(dialogStyle)
			styledLine.WriteString(helpLine)
			styledLine.WriteString(resetStyle)

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// overlayThemeDialog overlays the theme selection dialog centered on the viewport
func (e *Editor) overlayThemeDialog(viewportContent string) string {
	boxWidth := 40
	innerWidth := boxWidth - 2

	padText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw > width {
			return runewidth.Truncate(s, width, "")
		}
		return s + strings.Repeat(" ", width-sw)
	}

	centerText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw >= width {
			return runewidth.Truncate(s, width, "")
		}
		padLeft := (width - sw) / 2
		padRight := width - sw - padLeft
		return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
	}

	// Get the theme colors for the dialog
	themeUI := e.styles.Theme.UI
	dialogStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	selectedStyle := ui.ColorToANSI(themeUI.DialogButtonFg, themeUI.DialogButton)
	dialogResetStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	resetStyle := "\033[0m"

	// Current theme name for marking
	currentTheme := "default"
	if e.config != nil && e.config.Theme.Name != "" {
		currentTheme = e.config.Theme.Name
	}

	// Build dialog lines (plain text, color applied in overlay loop)
	var dialogLines []string

	// Top border with title
	title := " Select Theme "
	titlePadLeft := (innerWidth - len(title)) / 2
	titlePadRight := innerWidth - len(title) - titlePadLeft
	dialogLines = append(dialogLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Empty line
	dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Theme list - need internal styling for selected item
	for i, name := range e.themeList {
		// Mark current theme with asterisk, selected with highlight
		prefix := "   "
		if name == currentTheme {
			prefix = " * "
		}
		displayName := prefix + name

		var line string
		if i == e.themeIndex {
			// Selected item - highlighted (internal styling needed)
			line = e.box.Vertical + selectedStyle + padText(displayName, innerWidth) + dialogResetStyle + e.box.Vertical
		} else {
			// Normal item
			line = e.box.Vertical + padText(displayName, innerWidth) + e.box.Vertical
		}
		dialogLines = append(dialogLines, line)
	}

	// Empty line
	dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Footer
	footerText := centerText("[Enter] Select [E]dit [C]opy [Esc]", innerWidth)
	dialogLines = append(dialogLines, e.box.Vertical+footerText+e.box.Vertical)

	// Bottom border
	dialogLines = append(dialogLines, e.box.BottomLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.BottomRight)

	boxHeight := len(dialogLines)

	// Calculate centering
	startX := (e.width - boxWidth) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (e.viewport.Height() - boxHeight) / 2
	if startY < 0 {
		startY = 0
	}

	viewportLines := strings.Split(viewportContent, "\n")

	for i, dialogLine := range dialogLines {
		viewportY := startY + i
		if viewportY >= 0 && viewportY < len(viewportLines) {
			// Build the styled dialog line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(dialogStyle)
			styledLine.WriteString(dialogLine)
			styledLine.WriteString(resetStyle)

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// overlayRecentFilesDialog overlays the recent files dialog using DialogBuilder
func (e *Editor) overlayRecentFilesDialog(viewportContent string) string {
	if e.config == nil || len(e.config.RecentFiles) == 0 {
		return viewportContent
	}

	// Use DialogBuilder for consistent dialog rendering
	db := e.NewDialogBuilder(60)

	db.AddTitleBorder(" Recent Files ")
	db.AddEmptyLine()

	// Add recent files as selectable items
	for i, path := range e.config.RecentFiles {
		// Show just filename with truncated path
		display := formatRecentPath(path, db.InnerWidth())
		db.AddSelectableItem(display, i == e.recentFilesIndex)
	}

	db.AddEmptyLine()
	db.AddCenteredText("[Enter] Open  [Del] Remove  [Esc] Cancel")
	db.AddBottomBorder()

	return db.Overlay(viewportContent, e.width, e.viewport.Height())
}

// formatRecentPath formats a path to fit within the given width
func formatRecentPath(path string, maxWidth int) string {
	// Try to show as much of the path as possible
	if runewidth.StringWidth(path) <= maxWidth {
		return path
	}

	// Truncate from the left, showing the end of the path
	runes := []rune(path)
	for len(runes) > 0 && runewidth.StringWidth("..."+string(runes)) > maxWidth {
		runes = runes[1:]
	}
	return "..." + string(runes)
}
