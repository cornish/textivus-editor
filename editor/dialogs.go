package editor

import (
	"fmt"
	"github.com/cornish/textivus-editor/config"
	enc "github.com/cornish/textivus-editor/encoding"
	"github.com/cornish/textivus-editor/ui"
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
// Wide characters that would be split are excluded and replaced with spaces to maintain exact width.
func sliceAnsiString(s string, start, end int) string {
	var result strings.Builder
	visualPos := 0
	outputWidth := 0
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
					result.WriteString(escapeSeq.String())
				}
			}
			continue
		}

		// Regular character - check if it's in our range
		charWidth := runewidth.RuneWidth(r)
		charEnd := visualPos + charWidth

		if end != -1 && visualPos < end && charEnd > end {
			// Character would extend past end boundary - skip it but pad with spaces
			// (This handles wide chars that would be split)
			spacesNeeded := end - visualPos
			result.WriteString(strings.Repeat(" ", spacesNeeded))
			outputWidth += spacesNeeded
			visualPos = charEnd
			continue
		}

		if visualPos >= start && (end == -1 || charEnd <= end) {
			// Character fully within range
			if visualPos > start && outputWidth == 0 {
				// We're starting mid-string, might have skipped a wide char
				// Pad with spaces to maintain alignment
				result.WriteString(strings.Repeat(" ", visualPos-start))
				outputWidth += visualPos - start
			}
			result.WriteRune(r)
			outputWidth += charWidth
		} else if visualPos < start && charEnd > start {
			// Wide character straddles the start boundary - skip it, pad with space
			spacesNeeded := charEnd - start
			if end != -1 && start+spacesNeeded > end {
				spacesNeeded = end - start
			}
			result.WriteString(strings.Repeat(" ", spacesNeeded))
			outputWidth += spacesNeeded
		}

		visualPos = charEnd
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

	// Box dimensions - content is 64 chars, plus 2 for borders = 66
	boxWidth := 66
	innerWidth := boxWidth - 2
	centerText := func(s string) string {
		sLen := runewidth.StringWidth(s) // Use visual width for Unicode
		if sLen >= innerWidth {
			// Truncate by visual width
			return runewidth.Truncate(s, innerWidth, "")
		}
		padLeft := (innerWidth - sLen) / 2
		padRight := innerWidth - sLen - padLeft
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
		// ASCII mode - use asterisk art
		logoLines = []string{
			"     *****  *****  *   *  *****  ***  *   *  *   *   ****      ",
			"       *    *       * *     *     *   *   *  *   *  *          ",
			"       *    ****     *      *     *   *   *  *   *   ***       ",
			"       *    *       * *     *     *    * *   *   *      *      ",
			"       *    *****  *   *    *    ***    *     ***   ****       ",
			"                                                               ",
		}
	} else {
		// Unicode mode - use block art
		logoLines = []string{
			" ████████╗███████╗██╗  ██╗████████╗██╗██╗   ██╗██╗   ██╗███████╗",
			" ╚══██╔══╝██╔════╝╚██╗██╔╝╚══██╔══╝██║██║   ██║██║   ██║██╔════╝",
			"    ██║   █████╗   ╚███╔╝    ██║   ██║██║   ██║██║   ██║███████╗",
			"    ██║   ██╔══╝   ██╔██╗    ██║   ██║╚██╗ ██╔╝██║   ██║╚════██║",
			"    ██║   ███████╗██╔╝ ██╗   ██║   ██║ ╚████╔╝ ╚██████╔╝███████║",
			"    ╚═╝   ╚══════╝╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═══╝   ╚═════╝ ╚══════╝",
		}
	}

	var aboutLines []string

	// Top border with title
	title := " About Textivus "
	titlePadLeft := (innerWidth - len(title)) / 2
	titlePadRight := innerWidth - len(title) - titlePadLeft
	aboutLines = append(aboutLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Empty line
	aboutLines = append(aboutLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Logo lines
	for _, logoLine := range logoLines {
		aboutLines = append(aboutLines, e.box.Vertical+centerText(logoLine)+e.box.Vertical)
	}

	// Content lines
	aboutLines = append(aboutLines,
		e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical,
		e.box.Vertical+centerText("A Text Editor for the Rest of Us")+e.box.Vertical,
		e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical,
		e.box.Vertical+centerText("Version 0.2.0")+e.box.Vertical,
		e.box.Vertical+centerText("github.com/cornish/textivus-editor")+e.box.Vertical,
		e.box.Vertical+centerText("Copyright (c) 2025")+e.box.Vertical,
		e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical,
	)

	// Terminal capabilities
	caps := config.GetCapabilities()
	utf8Status := "No"
	if caps.UTF8Support {
		utf8Status = "Yes"
	}
	kittyStatus := "No"
	if caps.KittyGraphics {
		kittyStatus = "Yes"
	}
	aboutLines = append(aboutLines,
		e.box.Vertical+centerText("─── Terminal ───")+e.box.Vertical,
		e.box.Vertical+centerText(fmt.Sprintf("UTF-8: %s   Colors: %s   Kitty: %s", utf8Status, caps.ColorMode.String(), kittyStatus))+e.box.Vertical,
		e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical,
	)

	// Quote lines
	for _, quoteLine := range quoteLines {
		aboutLines = append(aboutLines, e.box.Vertical+quoteLine+e.box.Vertical)
	}

	// Footer
	aboutLines = append(aboutLines,
		e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical,
		e.box.Vertical+centerText("Press any key or click to close...")+e.box.Vertical,
	)

	// Bottom border
	aboutLines = append(aboutLines, e.box.BottomLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.BottomRight)

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

	// Helper to format a keybinding entry
	fmtKey := func(action, label string) string {
		binding := e.keybindings.GetBinding(action)
		key := config.FormatKeyForDisplay(binding.Primary)
		if key == "" {
			key = "(none)"
		}
		// Pad key to 14 chars, label follows
		keyPadded := padText("  "+key, 15)
		return keyPadded + label
	}

	// Define shortcuts in two columns using current keybindings
	leftCol := []string{
		"  FILE",
		fmtKey("new", "New file"),
		fmtKey("open", "Open file"),
		fmtKey("recent_files", "Recent files"),
		fmtKey("close", "Close file"),
		fmtKey("save", "Save file"),
		fmtKey("quit", "Quit"),
		"",
		"  EDIT",
		fmtKey("undo", "Undo"),
		fmtKey("redo", "Redo"),
		fmtKey("cut", "Cut"),
		fmtKey("copy", "Copy"),
		fmtKey("paste", "Paste"),
		fmtKey("cut_line", "Cut line"),
		fmtKey("select_all", "Select all"),
		"",
		"  SEARCH",
		fmtKey("find", "Find"),
		fmtKey("find_next", "Find next"),
		fmtKey("replace", "Replace"),
	}

	rightCol := []string{
		"  NAVIGATION",
		"  Arrows       Move cursor",
		fmtKey("word_left", "Move word left"),
		fmtKey("word_right", "Move word right"),
		"  Home/End     Start/end of line",
		fmtKey("doc_start", "Start of file"),
		fmtKey("doc_end", "End of file"),
		"  PgUp/PgDn    Page up/down",
		fmtKey("goto_line", "Go to line"),
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
	toggleLnKey := config.FormatKeyForDisplay(e.keybindings.GetBinding("toggle_line_numbers").Primary)
	if toggleLnKey == "" {
		toggleLnKey = "(none)"
	}
	helpLines = append(helpLines, e.box.Vertical+centerText("OPTIONS: "+toggleLnKey+" Line Numbers", innerWidth)+e.box.Vertical)
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

// overlayRecentDirsDialog overlays the recent directories dialog using DialogBuilder
func (e *Editor) overlayRecentDirsDialog(viewportContent string) string {
	if e.config == nil || len(e.config.RecentDirs) == 0 {
		return viewportContent
	}

	// Use DialogBuilder for consistent dialog rendering
	db := e.NewDialogBuilder(60)

	db.AddTitleBorder(" Recent Directories ")
	db.AddEmptyLine()

	// Add recent directories as selectable items
	for i, path := range e.config.RecentDirs {
		// Show truncated path
		display := formatRecentPath(path, db.InnerWidth())
		db.AddSelectableItem(display, i == e.recentDirsIndex)
	}

	db.AddEmptyLine()
	db.AddCenteredText("[Enter] Browse  [Del] Remove  [Esc] Cancel")
	db.AddBottomBorder()

	return db.Overlay(viewportContent, e.width, e.viewport.Height())
}

// overlayConfigErrorDialog overlays the config error dialog
func (e *Editor) overlayConfigErrorDialog(viewportContent string) string {
	boxWidth := 56
	db := e.NewDialogBuilder(boxWidth)

	db.AddTitleBorder(" Config Error ")
	db.AddEmptyLine()

	// Error message - truncate if needed
	errLine := "Error: " + e.configErrorMsg
	if runewidth.StringWidth(errLine) > db.InnerWidth() {
		errLine = runewidth.Truncate(errLine, db.InnerWidth(), "...")
	}
	db.AddText(errLine)

	// File path
	fileLine := "File: " + formatRecentPath(e.configErrorFile, db.InnerWidth()-6)
	db.AddText(fileLine)

	db.AddEmptyLine()

	// Button row with selection highlighting
	buttons := []string{"[ Edit File ]", "[ Use Defaults ]", "[ Quit ]"}
	var buttonRow strings.Builder
	for i, btn := range buttons {
		if i > 0 {
			buttonRow.WriteString("  ")
		}
		if i == e.configErrorChoice {
			buttonRow.WriteString(db.themeUI.selectedStyle)
			buttonRow.WriteString(btn)
			buttonRow.WriteString(db.themeUI.dialogResetStyle)
		} else {
			buttonRow.WriteString(btn)
		}
	}
	db.AddCenteredText(buttonRow.String())

	db.AddBottomBorder()

	return db.Overlay(viewportContent, e.width, e.viewport.Height())
}

// overlaySettingsDialog overlays the settings dialog
func (e *Editor) overlaySettingsDialog(viewportContent string) string {
	boxWidth := 54
	db := e.NewDialogBuilder(boxWidth)

	db.AddTitleBorder(" Settings ")
	db.AddEmptyLine()

	// Settings rows
	const (
		rowWordWrap     = 0
		rowLineNumbers  = 1
		rowSyntax       = 2
		rowScrollbar    = 3
		rowTabsToSpaces = 4
		rowBackupCount  = 5
		rowMaxBuffers   = 6
		rowTabWidth     = 7
		rowSave         = 8
		rowCancel       = 9
	)

	// Helper to format checkbox - pad first, then apply highlighting
	checkbox := func(label string, checked bool, row int) string {
		check := "[ ]"
		if checked {
			check = "[x]"
		}
		line := "  " + check + " " + label
		padded := db.PadText(line)
		if e.settingsIndex == row {
			return db.themeUI.selectedStyle + padded + db.themeUI.dialogResetStyle
		}
		return padded
	}

	// Helper to format number input - pad first, then apply highlighting
	numberInput := func(label string, value int, row int) string {
		valStr := fmt.Sprintf("%2d", value)
		line := "  " + label + ": [" + valStr + "] [-][+]"
		padded := db.PadText(line)
		if e.settingsIndex == row {
			return db.themeUI.selectedStyle + padded + db.themeUI.dialogResetStyle
		}
		return padded
	}

	// Checkboxes
	db.lines = append(db.lines, db.box.Vertical+checkbox("Word Wrap", e.settingsWordWrap, rowWordWrap)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+checkbox("Line Numbers", e.settingsLineNumbers, rowLineNumbers)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+checkbox("Syntax Highlighting", e.settingsSyntax, rowSyntax)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+checkbox("Scrollbar", e.settingsScrollbar, rowScrollbar)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+checkbox("Tabs to Spaces", e.settingsTabsToSpaces, rowTabsToSpaces)+db.box.Vertical)

	db.AddEmptyLine()

	// Number inputs
	db.lines = append(db.lines, db.box.Vertical+numberInput("Backup Count", e.settingsBackupCount, rowBackupCount)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+db.PadText("    0=disabled, 1=file~, N=rotating")+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+numberInput("Max Buffers", e.settingsMaxBuffers, rowMaxBuffers)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+db.PadText("    0=unlimited")+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+numberInput("Tab Width", e.settingsTabWidth, rowTabWidth)+db.box.Vertical)
	db.lines = append(db.lines, db.box.Vertical+db.PadText("    1-16 columns")+db.box.Vertical)

	db.AddEmptyLine()

	// Buttons - center them properly
	saveBtnText := "[ Save ]"
	cancelBtnText := "[ Cancel ]"
	buttonContent := saveBtnText + "    " + cancelBtnText // 8 + 4 + 10 = 22 chars
	paddedLine := db.CenterText(buttonContent)
	// Now apply highlighting by replacing the button text
	if e.settingsIndex == rowSave {
		paddedLine = strings.Replace(paddedLine, saveBtnText, db.themeUI.selectedStyle+saveBtnText+db.themeUI.dialogResetStyle, 1)
	}
	if e.settingsIndex == rowCancel {
		paddedLine = strings.Replace(paddedLine, cancelBtnText, db.themeUI.selectedStyle+cancelBtnText+db.themeUI.dialogResetStyle, 1)
	}
	db.lines = append(db.lines, db.box.Vertical+paddedLine+db.box.Vertical)

	db.AddBottomBorder()

	return db.Overlay(viewportContent, e.width, e.viewport.Height())
}

// overlayEncodingDialog overlays the encoding selection dialog
func (e *Editor) overlayEncodingDialog(viewportContent string) string {
	boxWidth := 50
	db := e.NewDialogBuilder(boxWidth)

	db.AddTitleBorder(" Save As Encoding ")
	db.AddEmptyLine()

	// Get list of supported encodings
	encodings := enc.GetSupportedEncodings()

	// Current encoding for marking
	currentEncoding := "utf-8"
	if e.activeDoc().encoding != nil {
		currentEncoding = e.activeDoc().encoding.ID
	}

	// Add encodings as selectable items
	for i, encoding := range encodings {
		prefix := "   "
		if encoding.ID == currentEncoding {
			prefix = " * "
		}
		display := prefix + encoding.Name
		if encoding.Description != "" {
			// Truncate description if needed
			maxDescLen := db.InnerWidth() - len(display) - 3
			if maxDescLen > 10 {
				desc := encoding.Description
				if len(desc) > maxDescLen {
					desc = desc[:maxDescLen-3] + "..."
				}
				display += " - " + desc
			}
		}
		db.AddSelectableItem(display, i == e.encodingIndex)
	}

	db.AddEmptyLine()
	db.AddCenteredText("Changes encoding used when saving")
	db.AddCenteredText("[Enter] Select  [Esc] Cancel")
	db.AddBottomBorder()

	return db.Overlay(viewportContent, e.width, e.viewport.Height())
}

// overlayKeybindingsDialog overlays the keybindings configuration dialog
func (e *Editor) overlayKeybindingsDialog(viewportContent string) string {
	boxWidth := 64
	innerWidth := boxWidth - 2 // 62

	actions := config.AllActions()
	actionCount := len(actions)

	// Calculate visible items based on viewport height
	visibleItems := e.viewport.Height() - 8
	if visibleItems > actionCount {
		visibleItems = actionCount
	}
	if visibleItems < 5 {
		visibleItems = 5
	}

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

	// Get theme colors
	themeUI := e.styles.Theme.UI
	dialogStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	selectedStyle := ui.ColorToANSI(themeUI.DialogButtonFg, themeUI.DialogButton)
	dialogResetStyle := ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg)
	resetStyle := "\033[0m"

	// Column widths: space(1) + Action(21) + |(1) + Primary(19) + |(1) + Alternate(19) = 62 = innerWidth
	actionWidth := 21
	keyWidth := 19

	var dialogLines []string

	// Top border with title
	title := " Keybindings "
	titlePadLeft := (innerWidth - len(title)) / 2
	titlePadRight := innerWidth - len(title) - titlePadLeft
	dialogLines = append(dialogLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Header row
	header := " " + padText("Action", actionWidth) + e.box.Vertical + padText(" Primary", keyWidth) + e.box.Vertical + padText(" Alternate", keyWidth)
	dialogLines = append(dialogLines, e.box.Vertical+header+e.box.Vertical)

	// Action list with scrolling
	for i := 0; i < visibleItems; i++ {
		idx := e.kbDialogScroll + i
		if idx >= actionCount {
			// Empty row if we run out of items
			dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)
			continue
		}

		action := actions[idx]
		actionName := config.ActionNames[action]
		binding := e.keybindings.GetBinding(action)

		primaryStr := config.FormatKeyForDisplay(binding.Primary)
		if primaryStr == "" {
			primaryStr = "(none)"
		}
		alternateStr := config.FormatKeyForDisplay(binding.Alternate)
		if alternateStr == "" {
			alternateStr = "(none)"
		}

		isSelected := idx == e.kbDialogIndex

		var line string
		if isSelected {
			// Build line with highlighting for selected row
			actionPart := " " + padText(actionName, actionWidth)

			// Format key fields - show brackets around active field, cursor when editing
			var primaryDisplay, alternateDisplay string
			if e.kbDialogEditField == 0 {
				if e.kbDialogEditing {
					primaryDisplay = "[" + padText("_", keyWidth-2) + "]"
				} else {
					primaryDisplay = "[" + padText(primaryStr, keyWidth-2) + "]"
				}
				alternateDisplay = " " + padText(alternateStr, keyWidth-1)
			} else {
				primaryDisplay = " " + padText(primaryStr, keyWidth-1)
				if e.kbDialogEditing {
					alternateDisplay = "[" + padText("_", keyWidth-2) + "]"
				} else {
					alternateDisplay = "[" + padText(alternateStr, keyWidth-2) + "]"
				}
			}

			line = e.box.Vertical + selectedStyle + actionPart +
				e.box.Vertical + primaryDisplay +
				e.box.Vertical + alternateDisplay + dialogResetStyle + e.box.Vertical
		} else {
			// Normal unselected row
			line = e.box.Vertical + " " + padText(actionName, actionWidth) +
				e.box.Vertical + " " + padText(primaryStr, keyWidth-1) +
				e.box.Vertical + " " + padText(alternateStr, keyWidth-1) + e.box.Vertical
		}
		dialogLines = append(dialogLines, line)
	}

	// Empty line
	dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)

	// Message line (errors, info)
	if e.kbDialogMessage != "" {
		msgText := centerText(e.kbDialogMessage, innerWidth)
		var msgLine string
		if e.kbDialogMsgError {
			// Show errors in red
			errorStyle := ui.ColorToANSIFg(themeUI.ErrorFg) + "\033[1m" // Bold red
			msgLine = e.box.Vertical + errorStyle + msgText + dialogResetStyle + e.box.Vertical
		} else {
			msgLine = e.box.Vertical + msgText + e.box.Vertical
		}
		dialogLines = append(dialogLines, msgLine)
	}

	// Footer with instructions
	footer := "[Enter] Edit  [<][>] Field  [R]eset  [Esc] Close"
	if e.kbDialogEditing {
		footer = "Press key to bind  [Esc] Cancel  [Del] Clear"
	} else if e.kbDialogConfirm {
		footer = "" // No footer during confirmation - message has the prompt
	}
	dialogLines = append(dialogLines, e.box.Vertical+centerText(footer, innerWidth)+e.box.Vertical)

	// Scroll indicator if needed
	scrollInfo := ""
	if actionCount > visibleItems {
		scrollInfo = centerText("["+strings.Repeat("^", min(1, e.kbDialogScroll))+strings.Repeat("v", min(1, actionCount-e.kbDialogScroll-visibleItems))+"]", innerWidth)
		if e.kbDialogScroll > 0 || e.kbDialogScroll+visibleItems < actionCount {
			dialogLines = append(dialogLines, e.box.Vertical+scrollInfo+e.box.Vertical)
		}
	}

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
			var styledLine strings.Builder
			styledLine.WriteString(dialogStyle)
			styledLine.WriteString(dialogLine)
			styledLine.WriteString(resetStyle)

			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}
