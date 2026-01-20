package editor

import (
	"github.com/cornish/textivus-editor/ui"
	"strings"

	"github.com/mattn/go-runewidth"
)

// DialogBuilder helps construct consistent dialogs
type DialogBuilder struct {
	box        BoxChars
	width      int      // Total box width including borders
	innerWidth int      // Width inside borders
	lines      []string // Built dialog lines
	themeUI    *themeColors
}

// themeColors holds the resolved theme color escape codes
type themeColors struct {
	dialogStyle      string // Base dialog fg/bg
	selectedStyle    string // Selected item fg/bg
	dialogResetStyle string // Reset to dialog colors after selection
	resetStyle       string // Full reset
}

// NewDialogBuilder creates a new dialog builder
func (e *Editor) NewDialogBuilder(width int) *DialogBuilder {
	themeUI := e.styles.Theme.UI
	return &DialogBuilder{
		box:        e.box,
		width:      width,
		innerWidth: width - 2,
		lines:      make([]string, 0),
		themeUI: &themeColors{
			dialogStyle:      ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg),
			selectedStyle:    ui.ColorToANSI(themeUI.DialogButtonFg, themeUI.DialogButton),
			dialogResetStyle: ui.ColorToANSI(themeUI.DialogFg, themeUI.DialogBg),
			resetStyle:       "\033[0m",
		},
	}
}

// AddTitleBorder adds the top border with an embedded title
func (db *DialogBuilder) AddTitleBorder(title string) {
	titlePadLeft := (db.innerWidth - runewidth.StringWidth(title)) / 2
	titlePadRight := db.innerWidth - runewidth.StringWidth(title) - titlePadLeft
	line := db.box.TopLeft +
		strings.Repeat(db.box.Horizontal, titlePadLeft) +
		title +
		strings.Repeat(db.box.Horizontal, titlePadRight) +
		db.box.TopRight
	db.lines = append(db.lines, line)
}

// AddBottomBorder adds the bottom border
func (db *DialogBuilder) AddBottomBorder() {
	db.lines = append(db.lines, db.box.BottomLeft+strings.Repeat(db.box.Horizontal, db.innerWidth)+db.box.BottomRight)
}

// AddEmptyLine adds an empty line with borders
func (db *DialogBuilder) AddEmptyLine() {
	db.lines = append(db.lines, db.box.Vertical+strings.Repeat(" ", db.innerWidth)+db.box.Vertical)
}

// AddText adds a line of text (left-aligned, padded)
func (db *DialogBuilder) AddText(text string) {
	db.lines = append(db.lines, db.box.Vertical+db.PadText(text)+db.box.Vertical)
}

// AddCenteredText adds a line of centered text
func (db *DialogBuilder) AddCenteredText(text string) {
	db.lines = append(db.lines, db.box.Vertical+db.CenterText(text)+db.box.Vertical)
}

// AddSelectableItem adds an item that can be selected (highlighted when selected)
func (db *DialogBuilder) AddSelectableItem(text string, isSelected bool) {
	var line string
	if isSelected {
		line = db.box.Vertical + db.themeUI.selectedStyle + db.PadText(text) + db.themeUI.dialogResetStyle + db.box.Vertical
	} else {
		line = db.box.Vertical + db.PadText(text) + db.box.Vertical
	}
	db.lines = append(db.lines, line)
}

// AddSeparator adds a horizontal separator line
func (db *DialogBuilder) AddSeparator() {
	db.lines = append(db.lines, db.box.TeeLeft+strings.Repeat(db.box.Horizontal, db.innerWidth)+db.box.TeeRight)
}

// PadText pads text to innerWidth (left-aligned)
func (db *DialogBuilder) PadText(s string) string {
	sw := runewidth.StringWidth(s)
	if sw > db.innerWidth {
		return runewidth.Truncate(s, db.innerWidth, "")
	}
	return s + strings.Repeat(" ", db.innerWidth-sw)
}

// CenterText centers text within innerWidth
func (db *DialogBuilder) CenterText(s string) string {
	sw := runewidth.StringWidth(s)
	if sw >= db.innerWidth {
		return runewidth.Truncate(s, db.innerWidth, "")
	}
	padLeft := (db.innerWidth - sw) / 2
	padRight := db.innerWidth - sw - padLeft
	return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
}

// Height returns the current height of the dialog
func (db *DialogBuilder) Height() int {
	return len(db.lines)
}

// InnerWidth returns the inner width (for external calculations)
func (db *DialogBuilder) InnerWidth() int {
	return db.innerWidth
}

// Lines returns the built dialog lines
func (db *DialogBuilder) Lines() []string {
	return db.lines
}

// Overlay renders the dialog centered on the viewport content
func (db *DialogBuilder) Overlay(viewportContent string, viewportWidth, viewportHeight int) string {
	startX := (viewportWidth - db.width) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (viewportHeight - len(db.lines)) / 2
	if startY < 0 {
		startY = 0
	}

	viewportLines := strings.Split(viewportContent, "\n")

	for i, dialogLine := range db.lines {
		viewportY := startY + i
		if viewportY >= 0 && viewportY < len(viewportLines) {
			// Build the styled dialog line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(db.themeUI.dialogStyle)
			styledLine.WriteString(dialogLine)
			styledLine.WriteString(db.themeUI.resetStyle)

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// DialogPosition calculates the dialog position for mouse handling
type DialogPosition struct {
	StartX    int
	StartY    int
	Width     int
	Height    int
	ListStart int // Y offset where selectable list starts (relative to dialog)
	ListEnd   int // Y offset where selectable list ends
}

// GetPosition returns the dialog's position information for mouse handling
func (db *DialogBuilder) GetPosition(viewportWidth, viewportHeight, listStart, listCount int) DialogPosition {
	startX := (viewportWidth - db.width) / 2
	if startX < 0 {
		startX = 0
	}
	startY := (viewportHeight - len(db.lines)) / 2
	if startY < 0 {
		startY = 0
	}
	return DialogPosition{
		StartX:    startX,
		StartY:    startY,
		Width:     db.width,
		Height:    len(db.lines),
		ListStart: listStart,
		ListEnd:   listStart + listCount,
	}
}

// MouseInDialog checks if mouse coordinates are inside the dialog
// Returns: inside bool, relX int, relY int
func (dp DialogPosition) MouseInDialog(mouseX, mouseY int) (bool, int, int) {
	relX := mouseX - dp.StartX
	relY := mouseY - dp.StartY
	inside := relX >= 0 && relX < dp.Width && relY >= 0 && relY < dp.Height
	return inside, relX, relY
}

// MouseInList checks if mouse Y is in the selectable list area
// Returns the list index or -1 if not in list
func (dp DialogPosition) MouseInList(relY int) int {
	if relY >= dp.ListStart && relY < dp.ListEnd {
		return relY - dp.ListStart
	}
	return -1
}
