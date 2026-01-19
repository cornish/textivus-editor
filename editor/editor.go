package editor

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"festivus/clipboard"
	"festivus/config"
	"festivus/syntax"
	"festivus/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the editor mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeMenu
	ModeFind
	ModeFindReplace
	ModePrompt
	ModeHelp
	ModeAbout
	ModeFileBrowser
	ModeSaveAs
	ModeTheme
)

// FileEntry represents a file or directory in the file browser
type FileEntry struct {
	Name     string
	IsDir    bool
	Size     int64
	Readable bool // For directories: whether we can read/enter it
}

// BoxChars holds characters used for drawing dialog boxes
type BoxChars struct {
	TopLeft     string
	TopRight    string
	BottomLeft  string
	BottomRight string
	Horizontal  string
	Vertical    string
	TeeLeft     string
	TeeRight    string
	Lock        string
	Ellipsis    string
}

// UnicodeBoxChars provides Unicode box drawing characters
var UnicodeBoxChars = BoxChars{
	TopLeft:     "┌",
	TopRight:    "┐",
	BottomLeft:  "└",
	BottomRight: "┘",
	Horizontal:  "─",
	Vertical:    "│",
	TeeLeft:     "├",
	TeeRight:    "┤",
	Lock:        "🔒",
	Ellipsis:    "…",
}

// AsciiBoxChars provides ASCII fallback characters
var AsciiBoxChars = BoxChars{
	TopLeft:     "+",
	TopRight:    "+",
	BottomLeft:  "+",
	BottomRight: "+",
	Horizontal:  "-",
	Vertical:    "|",
	TeeLeft:     "+",
	TeeRight:    "+",
	Lock:        "*",
	Ellipsis:    "...",
}

// PromptAction represents what to do with the prompt result
type PromptAction int

const (
	PromptNone PromptAction = iota
	PromptSaveAs
	PromptOpen
	PromptConfirmNew
	PromptConfirmOpen
	PromptConfirmClose
	PromptConfirmQuit
	PromptConfirmOverwrite
	PromptGoToLine
)

// FestivusQuotes are displayed randomly in the About dialog.
// Feel free to add more Seinfeld Festivus quotes!
var FestivusQuotes = []string{
	"A Festivus for the rest of us!",
	"I got a lot of problems with you people!",
	"I find tinsel distracting.",
	"It's a Festivus miracle!",
	"Serenity now!",
	"The tradition of Festivus begins with the airing of grievances.",
	"Until you pin me, George, Festivus is not over!",
	"No bagel, no bagel, no bagel!",
	"I find your belief system fascinating.",
	"This new holiday is scratching me right where I itch.",
	"Weren't there feats of strength that ended up with you crying?",
	"Instead there's a pole. Requires no decoration.",
	"We don't care and it shows.",
	"You couldn't smooth a silk sheet with a hot babe in it.",
	"Another piece of the puzzle falls into place.",
	"Instead of a tree didn't your father put up an aluminum pole?",
	"Happy Festivus.",
	"What's Festivus?",
	"It's a stupid holiday my father invented. It doesn't exist.",
	"Happy Festivus, Georgie.",
	"Frank invented a holiday? He's so prolific.",
	"As I rained blows upon him, I realized there had to be another way.",
	"But out of that a new holiday was born.",
	"Festivus is back!",
	"I'll get the pole out of the crawl space.",
	"What is that? Is that the pole?",
	"Festivus is your heritage. It's part of who you are.",
	"That's why I hate it.",
	"You're just weak. You're weak.",
	"It's time for the Festivus feats of strength.",
	"I don't really celebrate Christmas. I celebrate Festivus.",
	"I was afraid that I would be persecuted for my beliefs.",
	"It's made from aluminum. Very high strength to weight ratio.",
	"Not the feats of strength!",
	"Oh, please. Somebody stop this.",
	"Stop crying and fight your father.",
	"I give! I give!",
	"This is the best Festivus ever!",
	"When George was growing up his father hated all the commercial religious aspects of Christmas, so he made up his own holiday.",
	"At the Festivus dinner you gather your family around and tell them all the ways they have disappointed you over the past year.",
	"George, you're forgetting how much Festivus has meant to us all.",
}

// Editor is the main Bubbletea model for the text editor
type Editor struct {
	// Core components
	buffer    *Buffer
	cursor    *Cursor
	selection *Selection
	undoStack *UndoStack
	clipboard *clipboard.Clipboard

	// UI components
	menubar   *ui.MenuBar
	statusbar *ui.StatusBar
	viewport  *ui.Viewport
	styles    ui.Styles
	box       BoxChars // Characters for drawing dialog boxes

	// State
	mode     Mode
	filename string
	modified bool
	width    int
	height   int

	// Find mode state
	findQuery  string
	findActive bool

	// Find and Replace mode state
	replaceQuery string
	replaceFocus bool // true = replace field, false = find field

	// Prompt mode state
	promptText      string       // The prompt message
	promptInput     string       // User's input
	promptAction    PromptAction // What to do with the result
	pendingFilename string       // Filename pending confirmation (for overwrite)
	pendingQuit     bool         // Whether to quit after current action

	// Terminal state
	pendingTitle string // Title to set on next render

	// Mouse state
	mouseDown   bool
	mouseStartX int
	mouseStartY int

	// Key throttling
	lastPageKey time.Time

	// About dialog state
	aboutQuote string

	// File browser state (shared with Save As)
	fileBrowserDir      string      // Current directory
	fileBrowserEntries  []FileEntry // Directory contents
	fileBrowserSelected int         // Selected index
	fileBrowserScroll   int         // Scroll offset
	fileBrowserError    string      // Error message to display in dialog

	// Save As state
	saveAsFilename string // Filename input for Save As dialog
	saveAsFocusBrowser bool   // true = focus on browser, false = focus on filename

	// Syntax highlighting
	highlighter *syntax.Highlighter

	// Theme selection state
	themeList  []string // Available themes
	themeIndex int      // Selected theme index

	// Configuration
	config *config.Config
}

// New creates a new editor instance with default config
func New() *Editor {
	return NewWithConfig(config.DefaultConfig())
}

// NewWithConfig creates a new editor instance with the given configuration
func NewWithConfig(cfg *config.Config) *Editor {
	// Create styles from the configured theme
	theme := cfg.Theme.GetResolved()
	styles := ui.NewStyles(theme)
	buf := NewBuffer()

	// Determine ASCII mode: config override or auto-detect
	asciiMode := !detectUTF8Support()
	if cfg != nil && cfg.Editor.AsciiMode != nil {
		asciiMode = *cfg.Editor.AsciiMode
	}

	box := UnicodeBoxChars
	if asciiMode {
		box = AsciiBoxChars
	}

	e := &Editor{
		buffer:      buf,
		cursor:      NewCursor(buf),
		selection:   NewSelection(),
		undoStack:   NewUndoStack(1000),
		clipboard:   clipboard.New(os.Stdout),
		menubar:     ui.NewMenuBar(styles),
		statusbar:   ui.NewStatusBar(styles),
		viewport:    ui.NewViewport(styles),
		highlighter: syntax.New(""), // Initialize with no file
		styles:      styles,
		box:         box,
		mode:        ModeNormal,
		width:       80,
		height:      24,
		config:      cfg,
	}

	// Apply config settings
	if cfg != nil {
		e.viewport.SetWordWrap(cfg.Editor.WordWrap)
		e.viewport.ShowLineNumbers(cfg.Editor.LineNumbers)

		// Update menu checkboxes to reflect config
		if cfg.Editor.WordWrap {
			e.menubar.SetItemLabel(ui.ActionWordWrap, "[x] Word Wrap")
		}
		if cfg.Editor.LineNumbers {
			e.menubar.SetItemLabel(ui.ActionLineNumbers, "[x] Line Numbers")
		}

		// Apply syntax highlighting setting
		if cfg.Editor.SyntaxHighlight {
			e.highlighter.SetEnabled(true)
			e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[x] Syntax Highlight")
		} else {
			e.highlighter.SetEnabled(false)
		}

		// Apply true color setting (default is true)
		if cfg.Editor.TrueColor != nil && !*cfg.Editor.TrueColor {
			ui.UseTrueColor = false
		}

		// Apply theme syntax colors
		e.highlighter.SetColors(syntax.SyntaxColors{
			Keyword:  theme.Syntax.Keyword,
			String:   theme.Syntax.String,
			Comment:  theme.Syntax.Comment,
			Number:   theme.Syntax.Number,
			Operator: theme.Syntax.Operator,
			Function: theme.Syntax.Function,
			Type:     theme.Syntax.Type,
			Error:    theme.UI.ErrorFg,
		})
	}

	return e
}

// detectUTF8Support checks if the terminal likely supports UTF-8
func detectUTF8Support() bool {
	// Check LANG and LC_ALL environment variables for UTF-8
	for _, env := range []string{"LC_ALL", "LC_CTYPE", "LANG"} {
		val := strings.ToUpper(os.Getenv(env))
		if strings.Contains(val, "UTF-8") || strings.Contains(val, "UTF8") {
			return true
		}
	}
	return false
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Convert to absolute path for consistent directory navigation
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename // Fall back to original if Abs fails
	}

	e.buffer = NewBufferFromString(string(content))
	e.cursor = NewCursor(e.buffer)
	e.selection.Clear()
	e.undoStack.Clear()
	e.filename = absPath
	e.modified = false
	e.updateTitle()
	e.updateMenuState()

	// Update syntax highlighter for new file
	e.highlighter.SetFile(filename)

	return nil
}

// SaveFile saves the buffer to the current filename
// Returns true if save was initiated (might be async if prompting for filename)
func (e *Editor) SaveFile() bool {
	if e.filename == "" {
		// No filename - prompt for one
		e.showPrompt("Save as: ", PromptSaveAs)
		return false
	}

	return e.doSave()
}

// doSave performs the actual file save
func (e *Editor) doSave() bool {
	// Create backup if enabled and file exists
	if e.config != nil && e.config.Editor.BackupOnSave {
		if err := e.createBackup(); err != nil {
			e.statusbar.SetMessage("Backup failed: "+err.Error(), "error")
			return false
		}
	}

	content := e.buffer.String()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		// Clean up Go's error message for user display
		errMsg := err.Error()
		errMsg = strings.TrimPrefix(errMsg, "open ")
		e.statusbar.SetMessage("Save failed: "+errMsg, "error")
		return false
	}

	e.modified = false
	e.statusbar.SetMessage("Saved: "+e.filename, "success")
	e.updateTitle()
	e.updateMenuState()
	return true
}

// createBackup creates a backup copy of the current file (filename~)
func (e *Editor) createBackup() error {
	if e.filename == "" {
		return nil // No file to backup
	}

	// Check if file exists
	if _, err := os.Stat(e.filename); os.IsNotExist(err) {
		return nil // New file, nothing to backup
	}

	// Copy file to backup (filename~)
	backupPath := e.filename + "~"
	src, err := os.ReadFile(e.filename)
	if err != nil {
		return err
	}

	// Preserve original file permissions if possible
	info, err := os.Stat(e.filename)
	mode := os.FileMode(0644)
	if err == nil {
		mode = info.Mode()
	}

	return os.WriteFile(backupPath, src, mode)
}

// doSaveInDialog performs file save, showing errors in the dialog instead of status bar
func (e *Editor) doSaveInDialog() bool {
	// Create backup if enabled and file exists
	if e.config != nil && e.config.Editor.BackupOnSave {
		if err := e.createBackup(); err != nil {
			e.fileBrowserError = "Backup failed: " + err.Error()
			return false
		}
	}

	content := e.buffer.String()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		// Clean up Go's error message for dialog display
		errMsg := err.Error()
		errMsg = strings.TrimPrefix(errMsg, "open ")
		e.fileBrowserError = "Save failed: " + errMsg
		return false
	}

	e.modified = false
	e.fileBrowserError = ""
	e.statusbar.SetMessage("Saved: "+e.filename, "success")
	e.updateMenuState()
	return true
}

// showPrompt displays a prompt for user input
func (e *Editor) showPrompt(text string, action PromptAction) {
	e.promptText = text
	e.promptInput = ""
	e.promptAction = action
	e.mode = ModePrompt
	e.updateViewportSize()
}

// Init implements tea.Model
func (e *Editor) Init() tea.Cmd {
	e.updateTitle()
	e.updateMenuState()
	return tea.Batch(
		tea.EnterAltScreen,
		tea.EnableMouseAllMotion,
	)
}

// Update implements tea.Model
func (e *Editor) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for pending quit (after user confirmed discard)
	if e.pendingQuit {
		return e, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		e.width = msg.Width
		e.height = msg.Height
		e.menubar.SetWidth(msg.Width)
		e.statusbar.SetWidth(msg.Width)
		e.updateViewportSize()
		return e, nil

	case tea.KeyMsg:
		return e.handleKey(msg)

	case tea.MouseMsg:
		// Route mouse to dialog handlers if applicable
		if e.mode == ModeFileBrowser {
			return e.handleFileBrowserMouse(msg)
		}
		if e.mode == ModeSaveAs {
			return e.handleSaveAsMouse(msg)
		}
		if e.mode == ModeTheme {
			return e.handleThemeMouse(msg)
		}
		return e.handleMouse(msg)
	}

	return e, nil
}

// updateViewportSize recalculates the viewport size based on current state
func (e *Editor) updateViewportSize() {
	// Viewport height = total height - menu bar (1) - status bar (1)
	viewportHeight := e.height - 2

	// Subtract find bar if active
	if e.mode == ModeFind {
		viewportHeight--
	}

	// Subtract find/replace bar if active (2 lines)
	if e.mode == ModeFindReplace {
		viewportHeight -= 2
	}

	// Subtract prompt bar if active
	if e.mode == ModePrompt {
		viewportHeight--
	}

	// Note: We no longer subtract dropdown height because it overlays the viewport

	if viewportHeight < 1 {
		viewportHeight = 1
	}

	e.viewport.SetSize(e.width, viewportHeight)
}

// handleKey handles keyboard input
func (e *Editor) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle menu mode
	if e.mode == ModeMenu {
		return e.handleMenuKey(msg)
	}

	// Handle find mode
	if e.mode == ModeFind {
		return e.handleFindKey(msg)
	}

	// Handle find/replace mode
	if e.mode == ModeFindReplace {
		return e.handleFindReplaceKey(msg)
	}

	// Handle prompt mode
	if e.mode == ModePrompt {
		return e.handlePromptKey(msg)
	}

	// Handle help mode - any key dismisses
	if e.mode == ModeHelp {
		e.mode = ModeNormal
		return e, nil
	}

	// Handle about mode - any key dismisses
	if e.mode == ModeAbout {
		e.mode = ModeNormal
		return e, nil
	}

	// Handle theme selection mode
	if e.mode == ModeTheme {
		return e.handleThemeKey(msg)
	}

	// Handle file browser mode
	if e.mode == ModeFileBrowser {
		return e.handleFileBrowserKey(msg)
	}

	// Handle Save As mode
	if e.mode == ModeSaveAs {
		return e.handleSaveAsKey(msg)
	}

	// Clear status message on any key
	e.statusbar.ClearMessage()

	// Get key string for special combinations
	keyStr := msg.String()

	switch msg.Type {
	// Ctrl key combinations
	case tea.KeyCtrlC:
		if e.selection.Active && !e.selection.IsEmpty() {
			e.copy()
		}
		return e, nil

	case tea.KeyCtrlQ:
		return e, e.quitEditor()

	case tea.KeyCtrlS:
		e.SaveFile()
		return e, nil

	case tea.KeyCtrlZ:
		e.undo()
		return e, nil

	case tea.KeyCtrlY:
		e.redo()
		return e, nil

	case tea.KeyCtrlX:
		if e.selection.Active && !e.selection.IsEmpty() {
			e.cut()
		}
		return e, nil

	case tea.KeyCtrlV:
		e.paste()
		return e, nil

	case tea.KeyCtrlA:
		e.selectAll()
		return e, nil

	case tea.KeyCtrlF:
		e.mode = ModeFind
		e.findQuery = ""
		e.findActive = true
		e.updateViewportSize()
		return e, nil

	case tea.KeyCtrlN:
		e.newFile()
		return e, nil

	case tea.KeyCtrlO:
		e.openFile()
		return e, nil

	case tea.KeyCtrlW:
		e.closeFile()
		return e, nil

	case tea.KeyCtrlL:
		e.toggleLineNumbers()
		return e, nil

	case tea.KeyCtrlK:
		e.cutLine()
		return e, nil

	case tea.KeyCtrlG:
		e.showPrompt("Go to line: ", PromptGoToLine)
		return e, nil

	case tea.KeyCtrlH:
		e.showFindReplace()
		return e, nil

	case tea.KeyCtrlHome:
		e.selection.Clear()
		e.cursor.MoveToStart()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyCtrlEnd:
		e.selection.Clear()
		e.cursor.MoveToEnd()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	// Shift+Arrow selection keys
	case tea.KeyShiftLeft:
		e.moveWithSelection(e.cursor.MoveLeft)
		return e, nil

	case tea.KeyShiftRight:
		e.moveWithSelection(e.cursor.MoveRight)
		return e, nil

	case tea.KeyShiftUp:
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveUpVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
				if newLine == e.cursor.Line() && newCol == e.cursor.Col() {
					return false
				}
				e.cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.cursor.MoveUp()
		})
		return e, nil

	case tea.KeyShiftDown:
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveDownVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
				if newLine == e.cursor.Line() && newCol == e.cursor.Col() {
					return false
				}
				e.cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.cursor.MoveDown()
		})
		return e, nil

	case tea.KeyShiftHome:
		e.moveWithSelection(func() bool {
			e.cursor.MoveToLineStart()
			return true
		})
		return e, nil

	case tea.KeyShiftEnd:
		e.moveWithSelection(func() bool {
			e.cursor.MoveToLineEnd()
			return true
		})
		return e, nil

	// Ctrl+Shift combinations
	case tea.KeyCtrlShiftLeft:
		e.moveWithSelection(e.cursor.MoveWordLeft)
		return e, nil

	case tea.KeyCtrlShiftRight:
		e.moveWithSelection(e.cursor.MoveWordRight)
		return e, nil

	case tea.KeyCtrlShiftHome:
		e.moveWithSelection(func() bool {
			e.cursor.MoveToStart()
			return true
		})
		return e, nil

	case tea.KeyCtrlShiftEnd:
		e.moveWithSelection(func() bool {
			e.cursor.MoveToEnd()
			return true
		})
		return e, nil

	// Regular navigation keys
	case tea.KeyEsc:
		e.selection.Clear()
		if e.menubar.IsOpen() {
			e.menubar.Close()
			e.mode = ModeNormal
			e.updateViewportSize()
		}
		return e, nil

	case tea.KeyLeft:
		e.selection.Clear()
		e.cursor.MoveLeft()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyRight:
		e.selection.Clear()
		e.cursor.MoveRight()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyUp:
		e.selection.Clear()
		if e.viewport.WordWrap() {
			newLine, newCol := e.viewport.MoveUpVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
			e.cursor.SetPosition(newLine, newCol)
		} else {
			e.cursor.MoveUp()
		}
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyDown:
		e.selection.Clear()
		if e.viewport.WordWrap() {
			newLine, newCol := e.viewport.MoveDownVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
			e.cursor.SetPosition(newLine, newCol)
		} else {
			e.cursor.MoveDown()
		}
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyHome:
		e.selection.Clear()
		e.cursor.MoveToLineStart()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyEnd:
		e.selection.Clear()
		e.cursor.MoveToLineEnd()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyPgUp:
		// Throttle to prevent key queue buildup
		if time.Since(e.lastPageKey) < 50*time.Millisecond {
			return e, nil
		}
		e.lastPageKey = time.Now()

		// Move cursor up by one page
		pageSize := e.viewport.Height() - 1 // Keep 1 line of context
		for i := 0; i < pageSize; i++ {
			if !e.cursor.MoveUp() {
				break
			}
		}
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyPgDown:
		// Throttle to prevent key queue buildup
		if time.Since(e.lastPageKey) < 50*time.Millisecond {
			return e, nil
		}
		e.lastPageKey = time.Now()

		// Move cursor down by one page
		pageSize := e.viewport.Height() - 1 // Keep 1 line of context
		for i := 0; i < pageSize; i++ {
			if !e.cursor.MoveDown() {
				break
			}
		}
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	// Text editing keys
	case tea.KeyEnter:
		e.insertChar('\n')
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyTab:
		e.insertText("\t")
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyBackspace:
		e.backspace()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyDelete:
		e.delete()
		return e, nil

	case tea.KeySpace:
		e.insertChar(' ')
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	case tea.KeyRunes:
		// Check for Alt+letter combinations first
		if msg.Alt && len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'f', 'F':
				e.mode = ModeMenu
				e.menubar.OpenMenu(0) // File
				e.updateViewportSize()
				return e, nil
			case 'e', 'E':
				e.mode = ModeMenu
				e.menubar.OpenMenu(1) // Edit
				e.updateViewportSize()
				return e, nil
			case 's', 'S':
				e.mode = ModeMenu
				e.menubar.OpenMenu(2) // Search
				e.updateViewportSize()
				return e, nil
			case 'o', 'O':
				e.mode = ModeMenu
				e.menubar.OpenMenu(3) // Options
				e.updateViewportSize()
				return e, nil
			case 'h', 'H':
				e.mode = ModeMenu
				e.menubar.OpenMenu(4) // Help
				e.updateViewportSize()
				return e, nil
			}
		}
		// Regular character input
		for _, r := range msg.Runes {
			e.insertChar(r)
		}
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil
	}

	// Handle keys by string representation (fallback for terminals that report differently)
	switch keyStr {
	// Menu shortcuts
	case "alt+f":
		e.mode = ModeMenu
		e.menubar.OpenMenu(0) // File
		e.updateViewportSize()
		return e, nil
	case "alt+e":
		e.mode = ModeMenu
		e.menubar.OpenMenu(1) // Edit
		e.updateViewportSize()
		return e, nil
	case "alt+s":
		e.mode = ModeMenu
		e.menubar.OpenMenu(2) // Search
		e.updateViewportSize()
		return e, nil
	case "alt+o":
		e.mode = ModeMenu
		e.menubar.OpenMenu(3) // Options
		e.updateViewportSize()
		return e, nil
	case "alt+h":
		e.mode = ModeMenu
		e.menubar.OpenMenu(4) // Help
		e.updateViewportSize()
		return e, nil
	case "f10":
		e.mode = ModeMenu
		e.menubar.OpenMenu(0)
		e.updateViewportSize()
		return e, nil
	case "f1":
		e.showHelp()
		return e, nil
	case "f2":
		e.insertLoremIpsum()
		return e, nil
	case "f3":
		e.findNext()
		return e, nil

	// Fallback ctrl key handling (string-based)
	case "ctrl+s":
		e.SaveFile()
		return e, nil
	case "ctrl+q":
		return e, e.quitEditor()
	case "ctrl+z":
		e.undo()
		return e, nil
	case "ctrl+y":
		e.redo()
		return e, nil
	case "ctrl+x":
		if e.selection.Active && !e.selection.IsEmpty() {
			e.cut()
		}
		return e, nil
	case "ctrl+c":
		if e.selection.Active && !e.selection.IsEmpty() {
			e.copy()
		}
		return e, nil
	case "ctrl+v":
		e.paste()
		return e, nil
	case "ctrl+a":
		e.selectAll()
		return e, nil
	case "ctrl+f":
		e.mode = ModeFind
		e.findQuery = ""
		e.findActive = true
		e.updateViewportSize()
		return e, nil
	case "ctrl+n":
		e.newFile()
		return e, nil
	case "ctrl+o":
		e.openFile()
		return e, nil
	case "ctrl+w":
		e.closeFile()
		return e, nil
	case "ctrl+l":
		e.toggleLineNumbers()
		return e, nil
	case "ctrl+k":
		e.cutLine()
		return e, nil
	case "ctrl+g":
		e.showPrompt("Go to line: ", PromptGoToLine)
		return e, nil
	case "ctrl+h":
		e.showFindReplace()
		return e, nil

	// Word movement
	case "ctrl+left":
		e.selection.Clear()
		e.cursor.MoveWordLeft()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil
	case "ctrl+right":
		e.selection.Clear()
		e.cursor.MoveWordRight()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return e, nil

	// Shift+arrow selection (string-based fallback)
	case "shift+left":
		e.moveWithSelection(e.cursor.MoveLeft)
		return e, nil
	case "shift+right":
		e.moveWithSelection(e.cursor.MoveRight)
		return e, nil
	case "shift+up":
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveUpVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
				if newLine == e.cursor.Line() && newCol == e.cursor.Col() {
					return false
				}
				e.cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.cursor.MoveUp()
		})
		return e, nil
	case "shift+down":
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveDownVisual(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
				if newLine == e.cursor.Line() && newCol == e.cursor.Col() {
					return false
				}
				e.cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.cursor.MoveDown()
		})
		return e, nil
	case "shift+home":
		e.moveWithSelection(func() bool {
			e.cursor.MoveToLineStart()
			return true
		})
		return e, nil
	case "shift+end":
		e.moveWithSelection(func() bool {
			e.cursor.MoveToLineEnd()
			return true
		})
		return e, nil

	// Ctrl+Shift combinations
	case "ctrl+shift+left":
		e.moveWithSelection(e.cursor.MoveWordLeft)
		return e, nil
	case "ctrl+shift+right":
		e.moveWithSelection(e.cursor.MoveWordRight)
		return e, nil
	case "ctrl+shift+home":
		e.moveWithSelection(func() bool {
			e.cursor.MoveToStart()
			return true
		})
		return e, nil
	case "ctrl+shift+end":
		e.moveWithSelection(func() bool {
			e.cursor.MoveToEnd()
			return true
		})
		return e, nil
	}

	return e, nil
}

// moveWithSelection moves the cursor while extending the selection
func (e *Editor) moveWithSelection(move func() bool) {
	if !e.selection.Active {
		e.selection.Start(e.cursor.ByteOffset())
	}
	move()
	e.selection.Update(e.cursor.ByteOffset())
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
}

// handleMenuKey handles keyboard input in menu mode
func (e *Editor) handleMenuKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		e.menubar.Close()
		e.mode = ModeNormal
		e.updateViewportSize()

	case tea.KeyEnter:
		action := e.menubar.Select()
		e.mode = ModeNormal
		e.updateViewportSize()
		return e.executeAction(action)

	case tea.KeyUp:
		e.menubar.PrevItem()

	case tea.KeyDown:
		e.menubar.NextItem()

	case tea.KeyLeft:
		e.menubar.PrevMenu()
		e.updateViewportSize() // Dropdown height may change

	case tea.KeyRight:
		e.menubar.NextMenu()
		e.updateViewportSize() // Dropdown height may change

	case tea.KeyRunes:
		// Handle hotkey letter press
		if len(msg.Runes) == 1 {
			action := e.menubar.SelectByHotKey(msg.Runes[0])
			if action != ui.ActionNone {
				e.mode = ModeNormal
				e.updateViewportSize()
				return e.executeAction(action)
			}
		}
	}

	return e, nil
}

// handlePromptKey handles keyboard input in prompt mode
func (e *Editor) handlePromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.updateViewportSize()
		e.statusbar.SetMessage("Cancelled", "info")

	case tea.KeyEnter:
		e.executePrompt()
		// If quit was confirmed, exit immediately
		if e.pendingQuit {
			return e, tea.Quit
		}
		e.mode = ModeNormal
		e.updateViewportSize()

	case tea.KeyBackspace:
		if len(e.promptInput) > 0 {
			e.promptInput = e.promptInput[:len(e.promptInput)-1]
		}

	case tea.KeyRunes:
		e.promptInput += string(msg.Runes)

	case tea.KeySpace:
		e.promptInput += " "
	}

	return e, nil
}

// executePrompt handles the prompt result based on the action
func (e *Editor) executePrompt() {
	input := strings.TrimSpace(e.promptInput)

	switch e.promptAction {
	case PromptSaveAs:
		if input != "" {
			// Check if file already exists
			if _, err := os.Stat(input); err == nil {
				// File exists - ask for confirmation
				e.pendingFilename = input
				e.promptText = "File exists. Overwrite? (y/n): "
				e.promptInput = ""
				e.promptAction = PromptConfirmOverwrite
				e.mode = ModePrompt // Stay in prompt mode
				return
			}
			e.filename = input
			e.doSave()
		} else {
			e.statusbar.SetMessage("Save cancelled - no filename", "info")
		}

	case PromptConfirmOverwrite:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.filename = e.pendingFilename
			e.doSave()
		} else {
			e.statusbar.SetMessage("Save cancelled", "info")
		}
		e.pendingFilename = ""

	case PromptOpen:
		if input != "" {
			if err := e.LoadFile(input); err != nil {
				e.statusbar.SetMessage("Error: "+err.Error(), "error")
			} else {
				e.statusbar.SetMessage("Opened: "+input, "success")
				e.updateTitle()
			}
		}

	case PromptConfirmNew:
		// input should be "y" or "n"
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.doNewFile()
		} else {
			e.statusbar.SetMessage("Cancelled", "info")
		}

	case PromptConfirmOpen:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.modified = false // Discard changes
			e.showFileBrowser()
		} else {
			e.statusbar.SetMessage("Cancelled", "info")
		}

	case PromptConfirmClose:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.doCloseFile()
		} else {
			e.statusbar.SetMessage("Cancelled", "info")
		}

	case PromptConfirmQuit:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.pendingQuit = true
		} else {
			e.statusbar.SetMessage("Cancelled", "info")
		}

	case PromptGoToLine:
		if input == "" {
			e.statusbar.SetMessage("Cancelled", "info")
			return
		}
		lineNum, err := strconv.Atoi(input)
		if err != nil {
			e.statusbar.SetMessage("Invalid line number", "error")
			return
		}
		totalLines := e.buffer.LineCount()
		if lineNum < 1 {
			e.statusbar.SetMessage("Line number must be at least 1", "error")
			return
		}
		if lineNum > totalLines {
			e.statusbar.SetMessage(fmt.Sprintf("Line %d exceeds file length (%d lines)", lineNum, totalLines), "error")
			return
		}
		// Convert to 0-indexed
		e.cursor.SetPosition(lineNum-1, 0)
		e.selection.Clear()
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		e.statusbar.SetMessage(fmt.Sprintf("Jumped to line %d", lineNum), "info")
	}
}

// handleFindKey handles keyboard input in find mode
func (e *Editor) handleFindKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.findActive = false
		e.updateViewportSize()

	case tea.KeyEnter:
		e.findNext()

	case tea.KeyBackspace:
		if len(e.findQuery) > 0 {
			e.findQuery = e.findQuery[:len(e.findQuery)-1]
		}

	case tea.KeyRunes:
		e.findQuery += string(msg.Runes)

	case tea.KeySpace:
		e.findQuery += " "
	}

	return e, nil
}

// handleMouse handles mouse input
func (e *Editor) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Adjust for menu bar offset
	y := msg.Y - 1
	if e.menubar.IsOpen() {
		y -= e.menubar.DropdownHeight()
	}

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			// Check if click is on menu bar
			if msg.Y == 0 {
				handled, action := e.menubar.HandleClick(msg.X, 0)
				if handled {
					// Always update viewport size after menu bar interaction
					e.updateViewportSize()
					if action != ui.ActionNone {
						return e.executeAction(action)
					}
					if e.menubar.IsOpen() {
						e.mode = ModeMenu
					} else {
						e.mode = ModeNormal
					}
					return e, nil
				}
			}

			// Check if click is on menu dropdown
			if e.menubar.IsOpen() && msg.Y > 0 && msg.Y <= e.menubar.Height() {
				handled, action := e.menubar.HandleClick(msg.X, msg.Y)
				if handled {
					e.mode = ModeNormal
					e.updateViewportSize()
					if action != ui.ActionNone {
						return e.executeAction(action)
					}
					return e, nil
				}
			}

			// Close menu if clicking elsewhere
			if e.menubar.IsOpen() {
				e.menubar.Close()
				e.mode = ModeNormal
				e.updateViewportSize()
			}

			// Handle click in editor area
			if y >= 0 && y < e.viewport.Height() {
				line, col := e.viewport.PositionFromClickWrapped(e.buffer.Lines(), msg.X, y)
				e.cursor.SetPosition(line, col)
				e.selection.Clear()
				e.mouseDown = true
				e.mouseStartX = msg.X
				e.mouseStartY = y
			}
		} else if msg.Action == tea.MouseActionRelease {
			e.mouseDown = false
		} else if msg.Action == tea.MouseActionMotion && e.mouseDown {
			// Drag selection
			if y >= 0 && y < e.viewport.Height() {
				if !e.selection.Active {
					startLine, startCol := e.viewport.PositionFromClickWrapped(e.buffer.Lines(), e.mouseStartX, e.mouseStartY)
					startPos := e.buffer.LineColToPosition(startLine, startCol)
					e.selection.Start(startPos)
				}
				line, col := e.viewport.PositionFromClickWrapped(e.buffer.Lines(), msg.X, y)
				e.cursor.SetPosition(line, col)
				e.selection.Update(e.cursor.ByteOffset())
			}
		}

	case tea.MouseButtonWheelUp:
		e.viewport.ScrollUp()

	case tea.MouseButtonWheelDown:
		e.viewport.ScrollDownWrapped(e.buffer.Lines())
	}

	return e, nil
}

// executeAction executes a menu action
func (e *Editor) executeAction(action ui.MenuAction) (tea.Model, tea.Cmd) {
	switch action {
	case ui.ActionNew:
		e.newFile()
	case ui.ActionOpen:
		e.openFile()
	case ui.ActionClose:
		e.closeFile()
	case ui.ActionSave:
		e.SaveFile()
	case ui.ActionSaveAs:
		e.showSaveAs()
	case ui.ActionRevert:
		e.revertFile()
	case ui.ActionExit:
		return e, e.quitEditor()
	case ui.ActionUndo:
		e.undo()
	case ui.ActionRedo:
		e.redo()
	case ui.ActionCut:
		e.cut()
	case ui.ActionCopy:
		e.copy()
	case ui.ActionPaste:
		e.paste()
	case ui.ActionCutLine:
		e.cutLine()
	case ui.ActionSelectAll:
		e.selectAll()
	case ui.ActionFind:
		e.mode = ModeFind
		e.findQuery = ""
		e.findActive = true
		e.updateViewportSize()
	case ui.ActionFindNext:
		e.findNext()
	case ui.ActionReplace:
		e.showFindReplace()
	case ui.ActionGoToLine:
		e.showPrompt("Go to line: ", PromptGoToLine)
	case ui.ActionWordWrap:
		e.toggleWordWrap()
	case ui.ActionLineNumbers:
		e.toggleLineNumbers()
	case ui.ActionSyntaxHighlight:
		e.toggleSyntaxHighlight()
	case ui.ActionTheme:
		e.showThemeDialog()
	case ui.ActionHelp:
		e.showHelp()
	case ui.ActionAbout:
		e.showAbout()
	}
	return e, nil
}

// toggleWordWrap toggles word wrap on/off
func (e *Editor) toggleWordWrap() {
	wrap := !e.viewport.WordWrap()
	e.viewport.SetWordWrap(wrap)

	// Update menu checkbox
	if wrap {
		e.menubar.SetItemLabel(ui.ActionWordWrap, "[x] Word Wrap")
		e.statusbar.SetMessage("Word wrap enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionWordWrap, "[ ] Word Wrap")
		e.statusbar.SetMessage("Word wrap disabled", "info")
	}

	// Ensure cursor stays visible after toggle
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())

	// Save to config
	e.saveConfig()
}

// toggleLineNumbers toggles line numbers on/off
func (e *Editor) toggleLineNumbers() {
	show := !e.viewport.ShowLineNum()
	e.viewport.ShowLineNumbers(show)

	// Update menu checkbox
	if show {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[x] Line Numbers")
		e.statusbar.SetMessage("Line numbers enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[ ] Line Numbers")
		e.statusbar.SetMessage("Line numbers disabled", "info")
	}

	// Ensure cursor stays visible after toggle (text width changes)
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())

	// Save to config
	e.saveConfig()
}

// toggleSyntaxHighlight toggles syntax highlighting on/off
func (e *Editor) toggleSyntaxHighlight() {
	enabled := !e.highlighter.Enabled()
	e.highlighter.SetEnabled(enabled)

	// Update menu checkbox
	if enabled {
		e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[x] Syntax Highlight")
		e.statusbar.SetMessage("Syntax highlighting enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[ ] Syntax Highlight")
		e.statusbar.SetMessage("Syntax highlighting disabled", "info")
	}

	// Save to config
	e.saveConfig()
}

// saveConfig saves the current settings to the config file
func (e *Editor) saveConfig() {
	if e.config == nil {
		e.config = config.DefaultConfig()
	}
	e.config.Editor.WordWrap = e.viewport.WordWrap()
	e.config.Editor.LineNumbers = e.viewport.ShowLineNum()
	e.config.Editor.SyntaxHighlight = e.highlighter.Enabled()
	// Save in background - don't block the UI
	go e.config.Save()
}

// applyTheme changes the current theme and updates all UI components
func (e *Editor) applyTheme(themeName string) {
	// Load the theme
	theme := config.LoadTheme(themeName)

	// Create new styles from the theme
	styles := ui.NewStyles(theme)

	// Update all UI components
	e.menubar.SetStyles(styles)
	e.statusbar.SetStyles(styles)
	e.viewport.SetStyles(styles)
	e.styles = styles

	// Update syntax highlighter colors
	e.highlighter.SetColors(syntax.SyntaxColors{
		Keyword:  theme.Syntax.Keyword,
		String:   theme.Syntax.String,
		Comment:  theme.Syntax.Comment,
		Number:   theme.Syntax.Number,
		Operator: theme.Syntax.Operator,
		Function: theme.Syntax.Function,
		Type:     theme.Syntax.Type,
	})

	// Update config and save
	if e.config == nil {
		e.config = config.DefaultConfig()
	}
	e.config.Theme.Name = themeName
	go e.config.Save()

	e.statusbar.SetMessage("Theme: "+themeName, "info")
}

// showThemeDialog opens the theme selection dialog
func (e *Editor) showThemeDialog() {
	// Build list of available themes (built-in + user themes)
	e.themeList = config.ThemeNames()
	userThemes := config.ListUserThemes()
	for _, ut := range userThemes {
		// Only add if not already in the list
		found := false
		for _, t := range e.themeList {
			if t == ut {
				found = true
				break
			}
		}
		if !found {
			e.themeList = append(e.themeList, ut)
		}
	}

	// Find current theme index
	currentTheme := "default"
	if e.config != nil && e.config.Theme.Name != "" {
		currentTheme = e.config.Theme.Name
	}
	e.themeIndex = 0
	for i, name := range e.themeList {
		if name == currentTheme {
			e.themeIndex = i
			break
		}
	}

	e.mode = ModeTheme
}

// handleThemeKey handles key events in the theme selection dialog
func (e *Editor) handleThemeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if e.themeIndex > 0 {
			e.themeIndex--
		}
	case tea.KeyDown:
		if e.themeIndex < len(e.themeList)-1 {
			e.themeIndex++
		}
	case tea.KeyEnter:
		// Apply selected theme and close dialog
		if e.themeIndex >= 0 && e.themeIndex < len(e.themeList) {
			e.applyTheme(e.themeList[e.themeIndex])
		}
		e.mode = ModeNormal
	case tea.KeyEsc:
		// Cancel - just close dialog
		e.mode = ModeNormal
	}
	return e, nil
}

// handleThemeMouse handles mouse input in the theme selection dialog
func (e *Editor) handleThemeMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Calculate dialog position (must match overlayThemeDialog)
	boxWidth := 40
	themeCount := len(e.themeList)
	// Dialog structure: title, empty, themes..., empty, footer, bottom border
	boxHeight := themeCount + 5

	startX := (e.width - boxWidth) / 2
	startY := (e.viewport.Height() - boxHeight) / 2

	// Adjust mouse Y for menu bar
	mouseY := msg.Y - 1

	// Calculate relative position within dialog
	relX := msg.X - startX
	relY := mouseY - startY

	// Check if click is outside dialog - close it
	if relX < 0 || relX >= boxWidth || relY < 0 || relY >= boxHeight {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.mode = ModeNormal
		}
		return e, nil
	}

	// Theme list starts at line 2 (after title border and empty line)
	themeListStart := 2
	themeListEnd := themeListStart + themeCount

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			// Check if click is in theme list area
			if relY >= themeListStart && relY < themeListEnd {
				clickedIdx := relY - themeListStart
				if clickedIdx >= 0 && clickedIdx < themeCount {
					if e.themeIndex == clickedIdx {
						// Double-click effect: same item clicked again - apply it
						e.applyTheme(e.themeList[e.themeIndex])
						e.mode = ModeNormal
					} else {
						// First click - just select
						e.themeIndex = clickedIdx
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if e.themeIndex > 0 {
			e.themeIndex--
		}

	case tea.MouseButtonWheelDown:
		if e.themeIndex < themeCount-1 {
			e.themeIndex++
		}
	}

	return e, nil
}

// showHelp opens the Help dialog with keyboard shortcuts
func (e *Editor) showHelp() {
	e.mode = ModeHelp
}

// showAbout opens the About dialog with a random quote
func (e *Editor) showAbout() {
	e.mode = ModeAbout
	e.aboutQuote = FestivusQuotes[rand.Intn(len(FestivusQuotes))]
}

// Text manipulation methods

func (e *Editor) insertChar(r rune) {
	// Delete selection first if any
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
	}

	// Record for undo
	entry := &UndoEntry{
		Position:     e.cursor.ByteOffset(),
		Inserted:     string(r),
		CursorBefore: e.cursor.ByteOffset(),
	}

	e.cursor.Sync()
	e.buffer.InsertRune(r)
	e.cursor.MoveRight()

	entry.CursorAfter = e.cursor.ByteOffset()
	e.undoStack.Push(entry)
	e.modified = true
}

func (e *Editor) insertText(s string) {
	if s == "" {
		return
	}

	// Delete selection first if any
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
	}

	entry := &UndoEntry{
		Position:     e.cursor.ByteOffset(),
		Inserted:     s,
		CursorBefore: e.cursor.ByteOffset(),
	}

	e.cursor.Sync()
	e.buffer.Insert(s)
	e.cursor.SetByteOffset(e.cursor.ByteOffset() + len(s))

	entry.CursorAfter = e.cursor.ByteOffset()
	e.undoStack.Push(entry)
	e.modified = true
}

func (e *Editor) backspace() {
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	if e.cursor.ByteOffset() == 0 {
		return
	}

	// Sync cursor position to buffer gap
	e.cursor.Sync()

	// Get info about what we're about to delete (the rune before cursor)
	pos := e.cursor.ByteOffset()
	deleted := e.buffer.DeleteRuneBefore()
	if deleted == "" {
		return
	}

	// Update cursor position to match new gap position
	newPos := e.buffer.CursorPosition()
	e.cursor.SetByteOffset(newPos)

	entry := &UndoEntry{
		Position:     newPos,
		Deleted:      deleted,
		CursorBefore: pos,
		CursorAfter:  newPos,
	}

	e.undoStack.Push(entry)
	e.modified = true
}

func (e *Editor) delete() {
	if e.selection.Active && !e.selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	if e.cursor.ByteOffset() >= e.buffer.Length() {
		return
	}

	_, size := e.buffer.RuneAt(e.cursor.ByteOffset())
	if size == 0 {
		return
	}

	entry := &UndoEntry{
		Position:     e.cursor.ByteOffset(),
		Deleted:      e.buffer.Substring(e.cursor.ByteOffset(), e.cursor.ByteOffset()+size),
		CursorBefore: e.cursor.ByteOffset(),
		CursorAfter:  e.cursor.ByteOffset(),
	}

	e.cursor.Sync()
	e.buffer.DeleteAfter(size)
	e.undoStack.Push(entry)
	e.modified = true
}

func (e *Editor) deleteSelection() {
	if !e.selection.Active || e.selection.IsEmpty() {
		return
	}

	start, end := e.selection.Normalize()
	text := e.buffer.Substring(start, end)

	entry := &UndoEntry{
		Position:     start,
		Deleted:      text,
		CursorBefore: e.cursor.ByteOffset(),
		CursorAfter:  start,
	}

	e.buffer.Replace(start, end, "")
	e.cursor.SetByteOffset(start)
	e.selection.Clear()
	e.undoStack.Push(entry)
	e.modified = true
}

func (e *Editor) undo() {
	entry := e.undoStack.Undo()
	if entry == nil {
		return
	}

	// Reverse the operation
	if entry.Inserted != "" {
		// Was an insertion - delete it
		e.buffer.Replace(entry.Position, entry.Position+len(entry.Inserted), "")
	}
	if entry.Deleted != "" {
		// Was a deletion - insert it back
		e.buffer.MoveCursor(entry.Position)
		e.buffer.Insert(entry.Deleted)
	}

	e.cursor.SetByteOffset(entry.CursorBefore)
	e.selection.Clear()
	e.modified = true
}

func (e *Editor) redo() {
	entry := e.undoStack.Redo()
	if entry == nil {
		return
	}

	// Replay the operation
	if entry.Deleted != "" {
		// Was a deletion - delete it again
		e.buffer.Replace(entry.Position, entry.Position+len(entry.Deleted), "")
	}
	if entry.Inserted != "" {
		// Was an insertion - insert it again
		e.buffer.MoveCursor(entry.Position)
		e.buffer.Insert(entry.Inserted)
	}

	e.cursor.SetByteOffset(entry.CursorAfter)
	e.selection.Clear()
	e.modified = true
}

func (e *Editor) cut() {
	if !e.selection.Active || e.selection.IsEmpty() {
		return
	}

	text := e.selection.GetText(e.buffer)
	e.clipboard.Copy(text)
	e.deleteSelection()
}

// cutLine cuts the entire current line (like nano's Ctrl+K)
func (e *Editor) cutLine() {
	line := e.cursor.Line()
	lineStart := e.buffer.LineStartOffset(line)
	lineEnd := e.buffer.LineEndOffset(line)

	// Include the newline character if this isn't the last line
	if lineEnd < e.buffer.Length() {
		lineEnd++ // Include the \n
	}

	// If the line is empty and it's the only line, nothing to cut
	if lineStart == lineEnd {
		e.statusbar.SetMessage("Nothing to cut", "info")
		return
	}

	// Get the line content
	text := e.buffer.Substring(lineStart, lineEnd)

	// Copy to clipboard
	e.clipboard.Copy(text)

	// Record for undo
	entry := &UndoEntry{
		Position:     lineStart,
		Deleted:      text,
		CursorBefore: e.cursor.ByteOffset(),
		CursorAfter:  lineStart,
	}

	// Delete the line
	e.buffer.Replace(lineStart, lineEnd, "")
	e.cursor.SetByteOffset(lineStart)
	e.selection.Clear()
	e.undoStack.Push(entry)
	e.modified = true

	e.statusbar.SetMessage("Line cut", "info")
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
}

func (e *Editor) copy() {
	if !e.selection.Active || e.selection.IsEmpty() {
		return
	}

	text := e.selection.GetText(e.buffer)
	e.clipboard.Copy(text)
	e.statusbar.SetMessage("Copied", "info")
}

func (e *Editor) paste() {
	text, err := e.clipboard.Paste()
	if err != nil || text == "" {
		return
	}

	e.insertText(text)
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
}

func (e *Editor) selectAll() {
	e.selection.SelectAll(e.buffer)
	e.cursor.MoveToEnd()
}

func (e *Editor) newFile() {
	// Check for unsaved changes
	if e.modified {
		e.showPrompt("Unsaved changes. Discard? (y/n): ", PromptConfirmNew)
		return
	}
	e.doNewFile()
}

func (e *Editor) doNewFile() {
	e.buffer = NewBuffer()
	e.cursor = NewCursor(e.buffer)
	e.selection.Clear()
	e.undoStack.Clear()
	e.filename = ""
	e.modified = false
	e.highlighter.SetFile("") // Clear syntax highlighter
	e.updateTitle()
	e.updateMenuState()
	e.statusbar.SetMessage("New file", "info")
}

// closeFile closes the current file (same as new, but different messaging)
func (e *Editor) closeFile() {
	if e.modified {
		e.showPrompt("Unsaved changes. Discard? (y/n): ", PromptConfirmClose)
		return
	}
	e.doCloseFile()
}

func (e *Editor) doCloseFile() {
	e.buffer = NewBuffer()
	e.cursor = NewCursor(e.buffer)
	e.selection.Clear()
	e.undoStack.Clear()
	e.filename = ""
	e.modified = false
	e.highlighter.SetFile("")
	e.updateTitle()
	e.updateMenuState()
	e.statusbar.SetMessage("File closed", "info")
}

// quitEditor exits the editor, checking for unsaved changes
func (e *Editor) quitEditor() tea.Cmd {
	if e.modified {
		e.showPrompt("Unsaved changes. Quit anyway? (y/n): ", PromptConfirmQuit)
		return nil
	}
	return tea.Quit
}

// updateTitle sets the terminal title
func (e *Editor) updateTitle() {
	e.pendingTitle = "festivus"
	if e.filename != "" {
		e.pendingTitle += " - " + e.filename
	} else {
		e.pendingTitle += " - [Untitled]"
	}
}

// getTitle returns the current title for the terminal
func (e *Editor) getTitle() string {
	return e.pendingTitle
}

// updateMenuState updates disabled states for menu items
func (e *Editor) updateMenuState() {
	// Revert is disabled if there's no file to revert to
	e.menubar.SetItemDisabled(ui.ActionRevert, e.filename == "")
}

// openFile prompts for a filename to open
func (e *Editor) openFile() {
	if e.modified {
		e.showPrompt("Unsaved changes. Discard? (y/n): ", PromptConfirmOpen)
		return
	}
	e.showFileBrowser()
}

// revertFile reloads the file from disk
func (e *Editor) revertFile() {
	if e.filename == "" {
		e.statusbar.SetMessage("No file to revert", "error")
		return
	}
	if err := e.LoadFile(e.filename); err != nil {
		e.statusbar.SetMessage("Error: "+err.Error(), "error")
	} else {
		e.statusbar.SetMessage("Reverted to saved version", "success")
	}
}

func (e *Editor) insertLoremIpsum() {
	lorem := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris.

Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident.

Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.

Nemo enim ipsam voluptatem quia voluptas sit aspernatur aut odit aut fugit, sed quia consequuntur magni dolores eos qui ratione voluptatem sequi nesciunt.

Neque porro quisquam est, qui dolorem ipsum quia dolor sit amet, consectetur, adipisci velit, sed quia non numquam eius modi tempora incidunt ut labore et dolore magnam aliquam quaerat voluptatem.
`
	e.insertText(lorem)
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
	e.statusbar.SetMessage("Inserted lorem ipsum", "info")
}

func (e *Editor) findNext() {
	if e.findQuery == "" {
		return
	}

	content := e.buffer.String()
	startPos := e.cursor.ByteOffset() + 1
	if startPos >= len(content) {
		startPos = 0
	}

	// Search from cursor position
	idx := strings.Index(content[startPos:], e.findQuery)
	if idx >= 0 {
		pos := startPos + idx
		e.cursor.SetByteOffset(pos)
		e.selection.Active = true
		e.selection.Anchor = pos
		e.selection.Cursor = pos + len(e.findQuery)
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return
	}

	// Wrap around
	idx = strings.Index(content[:startPos], e.findQuery)
	if idx >= 0 {
		e.cursor.SetByteOffset(idx)
		e.selection.Active = true
		e.selection.Anchor = idx
		e.selection.Cursor = idx + len(e.findQuery)
		e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
		return
	}

	e.statusbar.SetMessage("Not found", "error")
}

// showFindReplace opens the find and replace bar
func (e *Editor) showFindReplace() {
	e.mode = ModeFindReplace
	e.replaceFocus = false // Start with focus on find field
	e.updateViewportSize()
}

// handleFindReplaceKey handles keyboard input in find/replace mode
func (e *Editor) handleFindReplaceKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.updateViewportSize()
		return e, nil

	case tea.KeyTab:
		// Switch between find and replace fields
		e.replaceFocus = !e.replaceFocus
		return e, nil

	case tea.KeyEnter:
		// Replace next occurrence
		e.replaceNext()
		return e, nil

	case tea.KeyCtrlA:
		// Replace all
		e.replaceAll()
		return e, nil

	case tea.KeyBackspace:
		if e.replaceFocus {
			if len(e.replaceQuery) > 0 {
				e.replaceQuery = e.replaceQuery[:len(e.replaceQuery)-1]
			}
		} else {
			if len(e.findQuery) > 0 {
				e.findQuery = e.findQuery[:len(e.findQuery)-1]
			}
		}
		return e, nil

	case tea.KeyRunes:
		if e.replaceFocus {
			e.replaceQuery += string(msg.Runes)
		} else {
			e.findQuery += string(msg.Runes)
		}
		return e, nil

	case tea.KeySpace:
		if e.replaceFocus {
			e.replaceQuery += " "
		} else {
			e.findQuery += " "
		}
		return e, nil
	}

	// Handle string-based keys
	switch msg.String() {
	case "ctrl+a":
		e.replaceAll()
		return e, nil
	}

	return e, nil
}

// replaceNext finds the next occurrence and replaces it
func (e *Editor) replaceNext() {
	if e.findQuery == "" {
		e.statusbar.SetMessage("No search term", "error")
		return
	}

	content := e.buffer.String()
	startPos := e.cursor.ByteOffset()

	// Search from cursor position
	idx := strings.Index(content[startPos:], e.findQuery)
	if idx < 0 {
		// Wrap around
		idx = strings.Index(content[:startPos], e.findQuery)
		if idx < 0 {
			e.statusbar.SetMessage("Not found", "error")
			return
		}
	} else {
		idx = startPos + idx
	}

	// Create undo entry for the replacement
	entry := &UndoEntry{
		Position:     idx,
		Deleted:      e.findQuery,
		Inserted:     e.replaceQuery,
		CursorBefore: e.cursor.ByteOffset(),
		CursorAfter:  idx + len(e.replaceQuery),
	}

	// Perform the replacement
	e.buffer.Replace(idx, idx+len(e.findQuery), e.replaceQuery)
	e.cursor.SetByteOffset(idx + len(e.replaceQuery))
	e.selection.Clear()
	e.undoStack.Push(entry)
	e.modified = true

	e.statusbar.SetMessage("Replaced", "info")
	e.viewport.EnsureCursorVisibleWrapped(e.buffer.Lines(), e.cursor.Line(), e.cursor.Col())
}

// replaceAll replaces all occurrences with a single undo entry
func (e *Editor) replaceAll() {
	if e.findQuery == "" {
		e.statusbar.SetMessage("No search term", "error")
		return
	}

	content := e.buffer.String()
	count := strings.Count(content, e.findQuery)
	if count == 0 {
		e.statusbar.SetMessage("Not found", "error")
		return
	}

	// Store original content for undo
	originalContent := content
	cursorBefore := e.cursor.ByteOffset()

	// Replace all occurrences
	newContent := strings.ReplaceAll(content, e.findQuery, e.replaceQuery)

	// Create a single undo entry for the entire operation
	entry := &UndoEntry{
		Position:     0,
		Deleted:      originalContent,
		Inserted:     newContent,
		CursorBefore: cursorBefore,
		CursorAfter:  0,
	}

	// Replace the entire buffer
	e.buffer = NewBufferFromString(newContent)
	e.cursor = NewCursor(e.buffer)
	e.selection.Clear()
	e.undoStack.Push(entry)
	e.modified = true

	e.statusbar.SetMessage(fmt.Sprintf("Replaced %d occurrences", count), "info")
}

// View implements tea.Model
func (e *Editor) View() string {
	var sb strings.Builder

	// Set terminal title using OSC escape sequence
	if e.pendingTitle != "" {
		sb.WriteString(fmt.Sprintf("\033]0;%s\007", e.pendingTitle))
	}

	// Menu bar
	sb.WriteString(e.menubar.View())
	sb.WriteString("\n")

	// Build selection map for viewport
	selectionMap := make(map[int]ui.SelectionRange)
	if e.selection.Active && !e.selection.IsEmpty() {
		start, end := e.selection.Normalize()
		startLine, startCol := e.buffer.PositionToLineCol(start)
		endLine, endCol := e.buffer.PositionToLineCol(end)

		for line := startLine; line <= endLine; line++ {
			sr := ui.SelectionRange{Start: 0, End: -1}
			if line == startLine {
				sr.Start = startCol
			}
			if line == endLine {
				sr.End = endCol
			}
			selectionMap[line] = sr
		}
	}

	// Editor viewport
	lines := e.buffer.Lines()

	// Generate syntax highlighting colors for visible lines
	var lineColors map[int][]syntax.ColorSpan
	if e.highlighter.Enabled() && e.highlighter.HasLexer() {
		lineColors = make(map[int][]syntax.ColorSpan)
		// Calculate visible line range
		startLine := e.viewport.ScrollY()
		endLine := startLine + e.viewport.Height()
		if endLine > len(lines) {
			endLine = len(lines)
		}
		for i := startLine; i < endLine; i++ {
			colors := e.highlighter.GetLineColors(lines[i])
			if len(colors) > 0 {
				lineColors[i] = colors
			}
		}
	}

	viewportContent := e.viewport.Render(lines, e.cursor.Line(), e.cursor.Col(), selectionMap, lineColors)

	// If menu dropdown is open, overlay it on top of the viewport
	if e.menubar.IsOpen() {
		dropdownLines, offset := e.menubar.RenderDropdown()
		if len(dropdownLines) > 0 {
			viewportLines := strings.Split(viewportContent, "\n")
			// Overlay dropdown on viewport, preserving text on both sides
			for i, dropLine := range dropdownLines {
				if i < len(viewportLines) {
					viewportLines[i] = overlayLineAt(dropLine, viewportLines[i], offset)
				}
			}
			viewportContent = strings.Join(viewportLines, "\n")
		}
	}

	// If help dialog is open, overlay it centered on the viewport
	if e.mode == ModeHelp {
		viewportContent = e.overlayHelpDialog(viewportContent)
	}

	// If about dialog is open, overlay it centered on the viewport
	if e.mode == ModeAbout {
		viewportContent = e.overlayAboutDialog(viewportContent)
	}

	// If file browser is open, overlay it centered on the viewport
	if e.mode == ModeFileBrowser {
		viewportContent = e.overlayFileBrowser(viewportContent)
	}

	// If Save As dialog is open, overlay it centered on the viewport
	if e.mode == ModeSaveAs {
		viewportContent = e.overlaySaveAs(viewportContent)
	}

	// If theme selection dialog is open, overlay it centered on the viewport
	if e.mode == ModeTheme {
		viewportContent = e.overlayThemeDialog(viewportContent)
	}

	sb.WriteString(viewportContent)
	sb.WriteString("\n")

	// Get theme colors for input bars
	barColor := ui.ColorToANSI(e.styles.Theme.UI.MenuFg, e.styles.Theme.UI.MenuBg)

	// Find bar if active
	if e.mode == ModeFind {
		findContent := "Find: " + e.findQuery + "_"
		padding := e.width - len(findContent)
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(findContent)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Find/Replace bar if active (two lines)
	if e.mode == ModeFindReplace {
		// Line 1: Find field
		findCursor := ""
		replaceCursor := ""
		if !e.replaceFocus {
			findCursor = "_"
		} else {
			replaceCursor = "_"
		}
		findLine := "Find: " + e.findQuery + findCursor
		findPadding := e.width - len(findLine) + len(findCursor) // adjust for cursor char
		if findPadding < 0 {
			findPadding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(findLine)
		sb.WriteString(strings.Repeat(" ", findPadding))
		sb.WriteString("\033[0m\n")

		// Line 2: Replace field with hints
		replaceLine := "Replace: " + e.replaceQuery + replaceCursor
		hints := " [Tab] Switch [Enter] Replace [Ctrl+A] All"
		availSpace := e.width - len(replaceLine) + len(replaceCursor) - len(hints)
		if availSpace < 0 {
			availSpace = 0
			hints = ""
		}
		sb.WriteString(barColor)
		sb.WriteString(replaceLine)
		sb.WriteString(strings.Repeat(" ", availSpace))
		sb.WriteString(hints)
		sb.WriteString("\033[0m\n")
	}

	// Prompt bar if active
	if e.mode == ModePrompt {
		promptContent := e.promptText + e.promptInput + "_"
		padding := e.width - len(promptContent)
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(promptContent)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Status bar
	e.statusbar.SetPosition(e.cursor.Line(), e.cursor.Col())
	e.statusbar.SetFilename(e.filename)
	e.statusbar.SetModified(e.modified)
	e.statusbar.SetTotalLines(e.buffer.LineCount())
	e.statusbar.SetCounts(e.buffer.WordCount(), e.buffer.RuneCount())
	sb.WriteString(e.statusbar.View())

	return sb.String()
}

// SetFilename sets the filename for the editor
func (e *Editor) SetFilename(filename string) {
	// Convert to absolute path for consistent directory navigation
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename // Fall back to original if Abs fails
	}
	e.filename = absPath
	e.highlighter.SetFile(absPath) // Update syntax highlighter
}

