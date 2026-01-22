package editor

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cornish/textivus-editor/clipboard"
	"github.com/cornish/textivus-editor/config"
	enc "github.com/cornish/textivus-editor/encoding"
	"github.com/cornish/textivus-editor/syntax"
	"github.com/cornish/textivus-editor/ui"

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
	ModeRecentFiles
	ModeRecentDirs
	ModeKeybindings
	ModeConfigError
	ModeSettings
	ModeEncoding
)

// FileEntry represents a file or directory in the file browser
type FileEntry struct {
	Name       string
	IsDir      bool
	Size       int64
	Readable   bool   // For directories: whether we can read/enter it
	IsFavorite bool   // Whether this item is favorited
	FullPath   string // Full path (used in favorites view)
	IsSpecial  bool   // True for special entries like "★ Favorites" or ".."
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
	PromptThemeCopyName
	PromptFileChanged // File changed on disk - reload?
)

// fileCheckMsg is sent periodically to check for external file changes
type fileCheckMsg struct{}

// fileCheckInterval is how often to check for external file changes
const fileCheckInterval = 30 * time.Second

// fileCheckCmd returns a command that sends a fileCheckMsg after the interval
func fileCheckCmd() tea.Cmd {
	return tea.Tick(fileCheckInterval, func(t time.Time) tea.Msg {
		return fileCheckMsg{}
	})
}

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

// Document holds the state for a single open file/buffer
type Document struct {
	buffer      *Buffer
	cursor      *Cursor
	selection   *Selection
	undoStack   *UndoStack
	filename    string
	modified    bool
	scrollY     int // viewport scroll position for this document
	highlighter *syntax.Highlighter
	modTime     time.Time     // file modification time when loaded/saved
	encoding    *enc.Encoding // detected file encoding
}

// Editor is the main Bubbletea model for the text editor
type Editor struct {
	// Documents (multiple buffer support)
	documents []*Document
	activeIdx int

	// Shared components
	clipboard *clipboard.Clipboard

	// UI components
	menubar   *ui.MenuBar
	statusbar *ui.StatusBar
	viewport  *ui.Viewport
	scrollbar *ui.Scrollbar
	styles    ui.Styles
	box       BoxChars // Characters for drawing dialog boxes

	// Column-based rendering
	compositor       *ui.Compositor
	lineNumRenderer  *ui.LineNumberRenderer
	textRenderer     *ui.TextRenderer
	minimapRenderer  ui.MinimapController
	scrollbarAdapter *ui.ScrollbarColumnAdapter

	// State
	mode   Mode
	width  int
	height int

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
	pendingTitle   string // Title to set on next render
	pendingEscapes string // Escape sequences to output on next render (e.g., clear Kitty graphics)

	// Mouse state
	mouseDown   bool
	mouseStartX int
	mouseStartY int

	// Key throttling
	lastPageKey time.Time

	// About dialog state
	aboutQuote string

	// File browser state (shared with Save As)
	fileBrowserDir       string      // Current directory
	fileBrowserEntries   []FileEntry // Directory contents
	fileBrowserSelected  int         // Selected index
	fileBrowserScroll    int         // Scroll offset
	fileBrowserError     string      // Error message to display in dialog
	fileBrowserFavorites bool        // true = showing favorites virtual directory

	// Save As state
	saveAsFilename     string // Filename input for Save As dialog
	saveAsFocusBrowser bool   // true = focus on browser, false = focus on filename

	// Theme selection state
	themeList       []string // Available themes
	themeIndex      int      // Selected theme index
	themeExportName string   // Theme name being exported/copied

	// Recent files dialog state
	recentFilesIndex int // Selected index in recent files dialog

	// Recent directories dialog state
	recentDirsIndex int // Selected index in recent dirs dialog

	// Configuration
	config      *config.Config
	keybindings *config.KeybindingsConfig

	// Keybindings dialog state
	kbDialogIndex     int    // Selected action index
	kbDialogScroll    int    // Scroll offset
	kbDialogEditing   bool   // true = waiting for key input
	kbDialogEditField int    // 0 = primary, 1 = alternate
	kbDialogMessage   string // Message to show in dialog (errors, etc.)
	kbDialogMsgError  bool   // true = message is an error (show in red)
	kbDialogConfirm   bool   // true = waiting for y/n confirmation

	// Config error dialog state
	configErrorFile   string // Path to malformed config file
	configErrorMsg    string // Error message from parser
	configErrorChoice int    // 0=Edit, 1=Defaults, 2=Quit

	// Settings dialog state
	settingsIndex        int  // Selected row in settings dialog
	settingsWordWrap     bool // Temporary value while editing
	settingsLineNumbers  bool
	settingsSyntax       bool
	settingsScrollbar    bool
	settingsBackupCount  int
	settingsMaxBuffers   int
	settingsTabWidth     int
	settingsTabsToSpaces bool

	// Encoding dialog state
	encodingIndex int // Selected encoding index
}

// activeDoc returns the currently active document
func (e *Editor) activeDoc() *Document {
	if len(e.documents) == 0 {
		return nil
	}
	return e.documents[e.activeIdx]
}

// switchToBuffer switches to the document at the given index
func (e *Editor) switchToBuffer(idx int) {
	if idx < 0 || idx >= len(e.documents) || idx == e.activeIdx {
		return
	}
	// Save current scroll position
	e.activeDoc().scrollY = e.viewport.ScrollY()

	// Switch
	e.activeIdx = idx

	// Restore new doc's scroll position
	e.viewport.SetScrollY(e.activeDoc().scrollY)

	// Update title, menu, and status
	e.updateTitle()
	e.updateMenuState()

	// Check if file changed on disk
	if e.fileChangedOnDisk() {
		e.statusbar.SetMessage("Warning: File changed on disk!", "error")
	} else {
		e.statusbar.SetMessage("", "")
	}
}

// nextBuffer switches to the next buffer (wraps around)
func (e *Editor) nextBuffer() {
	if len(e.documents) <= 1 {
		return
	}
	nextIdx := (e.activeIdx + 1) % len(e.documents)
	e.switchToBuffer(nextIdx)
}

// prevBuffer switches to the previous buffer (wraps around)
func (e *Editor) prevBuffer() {
	if len(e.documents) <= 1 {
		return
	}
	prevIdx := (e.activeIdx - 1 + len(e.documents)) % len(e.documents)
	e.switchToBuffer(prevIdx)
}

// bufferCount returns the number of open buffers
func (e *Editor) bufferCount() int {
	return len(e.documents)
}

// findBufferByFilename returns the index of a buffer with the given filename, or -1 if not found
func (e *Editor) findBufferByFilename(filename string) int {
	for i, doc := range e.documents {
		if doc.filename == filename {
			return i
		}
	}
	return -1
}

// matchesBinding checks if a key string matches a configured action
func (e *Editor) matchesBinding(keyStr string, action string) bool {
	return e.keybindings.GetBinding(action).Matches(keyStr)
}

// handleConfigurableBinding checks if the key matches any configurable binding and executes the action
// Returns (true, cmd) if handled, (false, nil) otherwise
func (e *Editor) handleConfigurableBinding(keyStr string, msg tea.KeyMsg) (bool, tea.Cmd) {
	// File operations
	if e.matchesBinding(keyStr, "new") {
		e.newFile()
		return true, nil
	}
	if e.matchesBinding(keyStr, "open") {
		e.openFile()
		return true, nil
	}
	if e.matchesBinding(keyStr, "save") {
		e.SaveFile()
		return true, nil
	}
	if e.matchesBinding(keyStr, "close") {
		e.closeFile()
		return true, nil
	}
	if e.matchesBinding(keyStr, "recent_files") {
		e.showRecentFiles()
		return true, nil
	}
	if e.matchesBinding(keyStr, "quit") {
		return true, e.quitEditor()
	}

	// Edit operations
	if e.matchesBinding(keyStr, "undo") {
		e.undo()
		return true, nil
	}
	if e.matchesBinding(keyStr, "redo") {
		e.redo()
		return true, nil
	}
	if e.matchesBinding(keyStr, "cut") {
		if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
			e.cut()
		}
		return true, nil
	}
	if e.matchesBinding(keyStr, "copy") {
		if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
			e.copy()
		}
		return true, nil
	}
	if e.matchesBinding(keyStr, "paste") {
		e.paste()
		return true, nil
	}
	if e.matchesBinding(keyStr, "cut_line") {
		e.cutLine()
		return true, nil
	}
	if e.matchesBinding(keyStr, "select_all") {
		e.selectAll()
		return true, nil
	}

	// Search operations
	if e.matchesBinding(keyStr, "find") {
		e.mode = ModeFind
		e.findQuery = ""
		e.findActive = true
		e.updateViewportSize()
		return true, nil
	}
	if e.matchesBinding(keyStr, "find_next") {
		e.findNext()
		return true, nil
	}
	if e.matchesBinding(keyStr, "replace") {
		e.showFindReplace()
		return true, nil
	}
	if e.matchesBinding(keyStr, "goto_line") {
		e.showPrompt("Go to line: ", PromptGoToLine)
		return true, nil
	}

	// Navigation
	if e.matchesBinding(keyStr, "word_left") {
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveWordLeft()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return true, nil
	}
	if e.matchesBinding(keyStr, "word_right") {
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveWordRight()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return true, nil
	}
	if e.matchesBinding(keyStr, "doc_start") {
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveToStart()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return true, nil
	}
	if e.matchesBinding(keyStr, "doc_end") {
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveToEnd()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return true, nil
	}

	// Buffer operations
	if e.matchesBinding(keyStr, "next_buffer") {
		if e.bufferCount() > 1 {
			e.nextBuffer()
		}
		return true, nil
	}
	if e.matchesBinding(keyStr, "prev_buffer") {
		if e.bufferCount() > 1 {
			e.prevBuffer()
		}
		return true, nil
	}

	// View toggles
	if e.matchesBinding(keyStr, "toggle_line_numbers") {
		e.toggleLineNumbers()
		return true, nil
	}

	// Help
	if e.matchesBinding(keyStr, "help") {
		e.showHelp()
		return true, nil
	}

	return false, nil
}

// fileChangedOnDisk checks if the file has been modified externally since last load/save
func (e *Editor) fileChangedOnDisk() bool {
	doc := e.activeDoc()
	if doc.filename == "" || doc.modTime.IsZero() {
		return false
	}
	fileInfo, err := os.Stat(doc.filename)
	if err != nil {
		return false // File doesn't exist or can't be read
	}
	return fileInfo.ModTime().After(doc.modTime)
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

	// Determine ASCII mode: config override or auto-detect from capabilities
	caps := config.GetCapabilities()
	asciiMode := caps.ShouldUseASCII(cfg.Editor.AsciiMode)

	box := UnicodeBoxChars
	if asciiMode {
		box = AsciiBoxChars
	}

	// Create the initial document
	buf := NewBuffer()
	doc := &Document{
		buffer:      buf,
		cursor:      NewCursor(buf),
		selection:   NewSelection(),
		undoStack:   NewUndoStack(1000),
		highlighter: syntax.New(""), // Initialize with no file
		filename:    "",
		modified:    false,
		scrollY:     0,
		encoding:    enc.GetEncodingByID("utf-8"), // Default to UTF-8
	}

	scrollbar := ui.NewScrollbar(styles)

	// Create minimap renderer - use Kitty graphics when available
	var minimapRenderer ui.MinimapController
	if caps.KittyGraphics {
		minimapRenderer = ui.NewKittyMinimapRenderer(styles, true)
	} else {
		minimapRenderer = ui.NewMinimapRenderer(styles)
	}

	e := &Editor{
		documents:   []*Document{doc},
		activeIdx:   0,
		clipboard:   clipboard.New(os.Stdout),
		menubar:     ui.NewMenuBar(styles),
		statusbar:   ui.NewStatusBar(styles),
		viewport:    ui.NewViewport(styles),
		scrollbar:   scrollbar,
		styles:      styles,
		box:         box,
		mode:        ModeNormal,
		width:       80,
		height:      24,
		config:      cfg,
		keybindings: config.LoadKeybindings(),
		// Initialize column renderers
		lineNumRenderer:  ui.NewLineNumberRenderer(styles),
		textRenderer:     ui.NewTextRenderer(styles),
		minimapRenderer:  minimapRenderer,
		scrollbarAdapter: ui.NewScrollbarColumnAdapter(scrollbar),
	}

	// Initialize compositor with default dimensions
	e.compositor = ui.NewCompositor(80, 22) // Will be resized on first render

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
			e.activeDoc().highlighter.SetEnabled(true)
			e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[x] Syntax Highlight")
		} else {
			e.activeDoc().highlighter.SetEnabled(false)
		}

		// Apply true color setting (default is true)
		if cfg.Editor.TrueColor != nil && !*cfg.Editor.TrueColor {
			ui.UseTrueColor = false
		}

		// Apply scrollbar setting
		if cfg.Editor.Scrollbar {
			e.scrollbar.SetEnabled(true)
			e.menubar.SetItemLabel(ui.ActionScrollbar, "[x] Scrollbar")
		}
		// Update viewport to account for scrollbar width
		e.viewport.SetScrollbarWidth(e.scrollbar.Width())

		// Apply minimap setting
		if cfg.Editor.Minimap {
			e.minimapRenderer.SetEnabled(true)
			e.menubar.SetItemLabel(ui.ActionMinimap, "[x] Minimap")
		}

		// Apply theme syntax colors
		e.activeDoc().highlighter.SetColors(syntax.SyntaxColors{
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

	// Setup compositor columns AFTER config is applied
	e.setupCompositorColumns()

	return e
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	// Convert to absolute path for consistent comparison
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename // Fall back to original if Abs fails
	}

	// Check if file is already open - switch to it
	if idx := e.findBufferByFilename(absPath); idx >= 0 {
		e.switchToBuffer(idx)
		e.statusbar.SetMessage("Switched to existing buffer", "info")
		return nil
	}

	// Read file content and get mod time
	rawContent, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	fileInfo, err := os.Stat(absPath)
	var modTime time.Time
	if err == nil {
		modTime = fileInfo.ModTime()
	}

	// Detect encoding
	detection := enc.Detect(rawContent)
	detectedEnc := detection.Encoding

	// Convert to UTF-8 if needed
	content, err := enc.DecodeToUTF8(rawContent, detectedEnc)
	if err != nil {
		// Fall back to raw content if decoding fails
		content = rawContent
		detectedEnc = enc.GetEncodingByID("utf-8")
	}

	// Decide whether to reuse current buffer or create new one
	// Only reuse the initial empty buffer (when there's just 1 document)
	// If user has created additional buffers, respect them
	currentDoc := e.activeDoc()
	reuseCurrentBuffer := len(e.documents) == 1 &&
		currentDoc.filename == "" &&
		!currentDoc.modified &&
		currentDoc.buffer.LineCount() == 1 &&
		len(currentDoc.buffer.Lines()[0]) == 0

	if reuseCurrentBuffer {
		// Reuse current buffer
		currentDoc.buffer = NewBufferFromString(string(content))
		currentDoc.cursor = NewCursor(currentDoc.buffer)
		currentDoc.selection.Clear()
		currentDoc.undoStack.Clear()
		currentDoc.scrollY = 0
		currentDoc.filename = absPath
		currentDoc.modified = false
		currentDoc.modTime = modTime
		currentDoc.highlighter.SetFile(filename)
		currentDoc.encoding = detectedEnc
	} else {
		// Check buffer limit before creating new document
		maxBuffers := 20 // default
		if e.config != nil && e.config.Editor.MaxBuffers > 0 {
			maxBuffers = e.config.Editor.MaxBuffers
		}
		if maxBuffers > 0 && len(e.documents) >= maxBuffers {
			return fmt.Errorf("buffer limit reached (%d)", maxBuffers)
		}

		// Create new document
		buf := NewBufferFromString(string(content))
		doc := &Document{
			buffer:      buf,
			cursor:      NewCursor(buf),
			selection:   NewSelection(),
			undoStack:   NewUndoStack(1000),
			highlighter: syntax.New(filename),
			filename:    absPath,
			modified:    false,
			scrollY:     0,
			modTime:     modTime,
			encoding:    detectedEnc,
		}
		e.documents = append(e.documents, doc)
		e.activeIdx = len(e.documents) - 1
	}

	// Warn if encoding is unsupported
	if detectedEnc != nil && !detectedEnc.Supported {
		e.statusbar.SetMessage("Warning: Unsupported encoding "+detectedEnc.Name, "error")
	}

	e.viewport.SetScrollY(0)
	e.updateTitle()
	e.updateMenuState()

	// Track in recent files and directories
	if e.config != nil {
		e.config.AddRecentFile(absPath)
		e.config.AddRecentDir(filepath.Dir(absPath))
		go e.config.Save()
	}

	return nil
}

// SaveFile saves the buffer to the current filename
// Returns true if save was initiated (might be async if prompting for filename)
func (e *Editor) SaveFile() bool {
	if e.activeDoc().filename == "" {
		// No filename - prompt for one
		e.showPrompt("Save as: ", PromptSaveAs)
		return false
	}

	// Check for external changes
	if e.fileChangedOnDisk() {
		e.showPrompt("File changed on disk. Overwrite? (y/N): ", PromptFileChanged)
		return false
	}

	return e.doSave()
}

// doSave performs the actual file save
func (e *Editor) doSave() bool {
	// Create backup if enabled and file exists
	if e.config != nil && e.config.Editor.BackupCount > 0 {
		if err := e.createBackup(); err != nil {
			e.statusbar.SetMessage("Backup failed: "+err.Error(), "error")
			return false
		}
	}

	content := e.activeDoc().buffer.String()
	var outputData []byte
	var encErr error
	docEnc := e.activeDoc().encoding

	// Encode to original encoding if supported, otherwise save as UTF-8
	if docEnc != nil && docEnc.Supported {
		outputData, encErr = enc.EncodeFromUTF8([]byte(content), docEnc)
		if encErr != nil {
			e.statusbar.SetMessage("Encoding failed, saving as UTF-8", "warning")
			outputData = []byte(content)
			e.activeDoc().encoding = enc.GetEncodingByID("utf-8")
		}
	} else {
		// Unsupported encoding - convert to UTF-8
		outputData = []byte(content)
		if docEnc != nil && !docEnc.Supported {
			e.activeDoc().encoding = enc.GetEncodingByID("utf-8")
			e.statusbar.SetMessage("Converted from "+docEnc.Name+" to UTF-8", "info")
		}
	}

	err := os.WriteFile(e.activeDoc().filename, outputData, 0644)
	if err != nil {
		// Clean up Go's error message for user display
		errMsg := err.Error()
		errMsg = strings.TrimPrefix(errMsg, "open ")
		e.statusbar.SetMessage("Save failed: "+errMsg, "error")
		return false
	}

	// Update stored mod time after successful save
	if fileInfo, err := os.Stat(e.activeDoc().filename); err == nil {
		e.activeDoc().modTime = fileInfo.ModTime()
	}

	e.activeDoc().modified = false
	if encErr == nil && (docEnc == nil || docEnc.Supported) {
		e.statusbar.SetMessage("Saved: "+e.activeDoc().filename, "success")
	}
	e.updateTitle()
	e.updateMenuState()

	// Track directory in recent dirs
	if e.config != nil {
		e.config.AddRecentDir(filepath.Dir(e.activeDoc().filename))
		go e.config.Save()
	}

	return true
}

// createBackup creates a backup copy of the current file
// With backup_count=1: creates filename~
// With backup_count>1: creates filename~1~ (newest) through filename~N~ (oldest)
func (e *Editor) createBackup() error {
	if e.activeDoc().filename == "" {
		return nil // No file to backup
	}

	// Check if file exists
	if _, err := os.Stat(e.activeDoc().filename); os.IsNotExist(err) {
		return nil // New file, nothing to backup
	}

	// Read current file content
	src, err := os.ReadFile(e.activeDoc().filename)
	if err != nil {
		return err
	}

	// Preserve original file permissions if possible
	info, err := os.Stat(e.activeDoc().filename)
	mode := os.FileMode(0644)
	if err == nil {
		mode = info.Mode()
	}

	backupCount := 1
	if e.config != nil {
		backupCount = e.config.Editor.BackupCount
	}

	if backupCount == 1 {
		// Simple backup: filename~
		return os.WriteFile(e.activeDoc().filename+"~", src, mode)
	}

	// Numbered backups: rotate existing backups
	// Delete oldest backup if it exists
	oldestBackup := fmt.Sprintf("%s~%d~", e.activeDoc().filename, backupCount)
	os.Remove(oldestBackup) // Ignore error if doesn't exist

	// Rotate backups: ~2~ becomes ~3~, ~1~ becomes ~2~, etc.
	for i := backupCount - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s~%d~", e.activeDoc().filename, i)
		newPath := fmt.Sprintf("%s~%d~", e.activeDoc().filename, i+1)
		if _, err := os.Stat(oldPath); err == nil {
			os.Rename(oldPath, newPath)
		}
	}

	// Write new backup as ~1~ (newest)
	return os.WriteFile(fmt.Sprintf("%s~1~", e.activeDoc().filename), src, mode)
}

// doSaveInDialog performs file save, showing errors in the dialog instead of status bar
func (e *Editor) doSaveInDialog() bool {
	// Create backup if enabled and file exists
	if e.config != nil && e.config.Editor.BackupCount > 0 {
		if err := e.createBackup(); err != nil {
			e.fileBrowserError = "Backup failed: " + err.Error()
			return false
		}
	}

	content := e.activeDoc().buffer.String()
	var outputData []byte
	docEnc := e.activeDoc().encoding

	// Encode to original encoding if supported, otherwise save as UTF-8
	if docEnc != nil && docEnc.Supported {
		var encErr error
		outputData, encErr = enc.EncodeFromUTF8([]byte(content), docEnc)
		if encErr != nil {
			outputData = []byte(content)
			e.activeDoc().encoding = enc.GetEncodingByID("utf-8")
		}
	} else {
		outputData = []byte(content)
		if docEnc != nil && !docEnc.Supported {
			e.activeDoc().encoding = enc.GetEncodingByID("utf-8")
		}
	}

	err := os.WriteFile(e.activeDoc().filename, outputData, 0644)
	if err != nil {
		// Clean up Go's error message for dialog display
		errMsg := err.Error()
		errMsg = strings.TrimPrefix(errMsg, "open ")
		e.fileBrowserError = "Save failed: " + errMsg
		return false
	}

	e.activeDoc().modified = false
	e.fileBrowserError = ""
	e.statusbar.SetMessage("Saved: "+e.activeDoc().filename, "success")
	e.updateMenuState()

	// Track directory in recent dirs
	if e.config != nil {
		e.config.AddRecentDir(filepath.Dir(e.activeDoc().filename))
		go e.config.Save()
	}

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
		fileCheckCmd(), // Start periodic file change detection
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

	case fileCheckMsg:
		// Periodic check for external file changes
		if e.fileChangedOnDisk() && e.mode == ModeNormal {
			e.statusbar.SetMessage("File changed on disk!", "error")
		}
		return e, fileCheckCmd() // Schedule next check

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
		if e.mode == ModeRecentFiles {
			return e.handleRecentFilesMouse(msg)
		}
		if e.mode == ModeRecentDirs {
			return e.handleRecentDirsMouse(msg)
		}
		if e.mode == ModeKeybindings {
			return e.handleKeybindingsMouse(msg)
		}
		if e.mode == ModeConfigError {
			return e.handleConfigErrorMouse(msg)
		}
		if e.mode == ModeSettings {
			return e.handleSettingsMouse(msg)
		}
		if e.mode == ModeEncoding {
			return e.handleEncodingMouse(msg)
		}
		if e.mode == ModeHelp {
			return e.handleHelpMouse(msg)
		}
		if e.mode == ModeAbout {
			return e.handleAboutMouse(msg)
		}
		return e.handleMouse(msg)
	}

	return e, nil
}

// setupCompositorColumns configures the compositor columns based on current settings.
func (e *Editor) setupCompositorColumns() {
	columns := []ui.Column{
		// Line numbers (fixed width 5)
		{
			Width:    5,
			Flexible: false,
			Enabled:  e.viewport.ShowLineNum(),
			Renderer: e.lineNumRenderer,
		},
		// Text content (flexible)
		{
			Width:    0,
			Flexible: true,
			Enabled:  true,
			Renderer: e.textRenderer,
		},
		// Minimap (fixed width 8)
		{
			Width:    ui.MinimapWidth(),
			Flexible: false,
			Enabled:  e.minimapRenderer.IsEnabled(),
			Renderer: e.minimapRenderer,
		},
		// Scrollbar (fixed width 1)
		{
			Width:    1,
			Flexible: false,
			Enabled:  e.scrollbar.IsEnabled(),
			Renderer: e.scrollbarAdapter,
		},
	}
	e.compositor.SetColumns(columns)
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
	e.scrollbar.SetHeight(viewportHeight)
	e.compositor.SetSize(e.width, viewportHeight)
}

// buildRenderState creates a RenderState for the compositor from current editor state.
func (e *Editor) buildRenderState() *ui.RenderState {
	lines := e.activeDoc().buffer.Lines()

	// Build selection map
	selectionMap := make(map[int]ui.SelectionRange)
	if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
		start, end := e.activeDoc().selection.Normalize()
		startLine, startCol := e.activeDoc().buffer.PositionToLineCol(start)
		endLine, endCol := e.activeDoc().buffer.PositionToLineCol(end)

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

	// Generate syntax highlighting colors
	// When minimap is enabled, generate for all lines; otherwise just visible lines
	var lineColors map[int][]syntax.ColorSpan
	if e.activeDoc().highlighter.Enabled() && e.activeDoc().highlighter.HasLexer() {
		lineColors = make(map[int][]syntax.ColorSpan)
		startLine := 0
		endLine := len(lines)
		// If minimap is disabled, only generate for visible lines (performance)
		if !e.minimapRenderer.IsEnabled() {
			startLine = e.viewport.ScrollY()
			endLine = startLine + e.viewport.Height()
			if endLine > len(lines) {
				endLine = len(lines)
			}
		}
		for i := startLine; i < endLine; i++ {
			colors := e.activeDoc().highlighter.GetLineColors(lines[i])
			if len(colors) > 0 {
				lineColors[i] = colors
			}
		}
	}

	// Calculate total visual lines
	totalVisualLines := len(lines)
	if e.viewport.WordWrap() {
		totalVisualLines = e.viewport.CountVisualLines(lines)
	}

	return &ui.RenderState{
		Lines:            lines,
		CursorLine:       e.activeDoc().cursor.Line(),
		CursorCol:        e.activeDoc().cursor.Col(),
		ScrollY:          e.viewport.ScrollY(),
		ScrollX:          e.viewport.ScrollX(),
		Selection:        selectionMap,
		LineColors:       lineColors,
		WordWrap:         e.viewport.WordWrap(),
		TabWidth:         e.config.Editor.TabWidth,
		TotalLines:       len(lines),
		TotalVisualLines: totalVisualLines,
		Styles:           e.styles,
	}
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

	// Handle config error mode
	if e.mode == ModeConfigError {
		return e.handleConfigErrorKey(msg)
	}

	// Handle settings mode
	if e.mode == ModeSettings {
		return e.handleSettingsKey(msg)
	}

	// Handle encoding selection mode
	if e.mode == ModeEncoding {
		return e.handleEncodingKey(msg)
	}

	// Handle theme selection mode
	if e.mode == ModeTheme {
		return e.handleThemeKey(msg)
	}

	// Handle recent files mode
	if e.mode == ModeRecentFiles {
		return e.handleRecentFilesKey(msg)
	}

	// Handle recent directories mode
	if e.mode == ModeRecentDirs {
		return e.handleRecentDirsKey(msg)
	}

	// Handle keybindings mode
	if e.mode == ModeKeybindings {
		return e.handleKeybindingsKey(msg)
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

	// Get key string for matching against configurable bindings
	keyStr := msg.String()

	// Check configurable keybindings first
	if handled, cmd := e.handleConfigurableBinding(keyStr, msg); handled {
		return e, cmd
	}

	switch msg.Type {

	// Shift+Arrow selection keys
	case tea.KeyShiftLeft:
		e.moveWithSelection(e.activeDoc().cursor.MoveLeft)
		return e, nil

	case tea.KeyShiftRight:
		e.moveWithSelection(e.activeDoc().cursor.MoveRight)
		return e, nil

	case tea.KeyShiftUp:
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveUpVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
				if newLine == e.activeDoc().cursor.Line() && newCol == e.activeDoc().cursor.Col() {
					return false
				}
				e.activeDoc().cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.activeDoc().cursor.MoveUp()
		})
		return e, nil

	case tea.KeyShiftDown:
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveDownVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
				if newLine == e.activeDoc().cursor.Line() && newCol == e.activeDoc().cursor.Col() {
					return false
				}
				e.activeDoc().cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.activeDoc().cursor.MoveDown()
		})
		return e, nil

	case tea.KeyShiftHome:
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToLineStart()
			return true
		})
		return e, nil

	case tea.KeyShiftEnd:
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToLineEnd()
			return true
		})
		return e, nil

	// Ctrl+Shift combinations
	case tea.KeyCtrlShiftLeft:
		e.moveWithSelection(e.activeDoc().cursor.MoveWordLeft)
		return e, nil

	case tea.KeyCtrlShiftRight:
		e.moveWithSelection(e.activeDoc().cursor.MoveWordRight)
		return e, nil

	case tea.KeyCtrlShiftHome:
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToStart()
			return true
		})
		return e, nil

	case tea.KeyCtrlShiftEnd:
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToEnd()
			return true
		})
		return e, nil

	// Regular navigation keys
	case tea.KeyEsc:
		e.activeDoc().selection.Clear()
		if e.menubar.IsOpen() {
			e.menubar.Close()
			e.mode = ModeNormal
			e.updateViewportSize()
		}
		return e, nil

	case tea.KeyLeft:
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveLeft()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyRight:
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveRight()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyUp:
		e.activeDoc().selection.Clear()
		if e.viewport.WordWrap() {
			newLine, newCol := e.viewport.MoveUpVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
			e.activeDoc().cursor.SetPosition(newLine, newCol)
		} else {
			e.activeDoc().cursor.MoveUp()
		}
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyDown:
		e.activeDoc().selection.Clear()
		if e.viewport.WordWrap() {
			newLine, newCol := e.viewport.MoveDownVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
			e.activeDoc().cursor.SetPosition(newLine, newCol)
		} else {
			e.activeDoc().cursor.MoveDown()
		}
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyHome:
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveToLineStart()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyEnd:
		e.activeDoc().selection.Clear()
		e.activeDoc().cursor.MoveToLineEnd()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
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
			if !e.activeDoc().cursor.MoveUp() {
				break
			}
		}
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
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
			if !e.activeDoc().cursor.MoveDown() {
				break
			}
		}
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	// Text editing keys
	case tea.KeyEnter:
		e.insertChar('\n')
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyTab:
		// If there's a selection, indent all selected lines
		if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
			e.indentLines()
		} else {
			// No selection - insert tab/spaces based on config
			e.insertText(e.getIndentString())
		}
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyShiftTab:
		// Dedent current line or all selected lines
		e.dedentLines()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyBackspace:
		e.backspace()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return e, nil

	case tea.KeyDelete:
		e.delete()
		return e, nil

	case tea.KeySpace:
		e.insertChar(' ')
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
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
			case 'b', 'B':
				e.mode = ModeMenu
				e.menubar.OpenMenu(1) // Buffers
				e.updateViewportSize()
				return e, nil
			case 'e', 'E':
				e.mode = ModeMenu
				e.menubar.OpenMenu(2) // Edit
				e.updateViewportSize()
				return e, nil
			case 's', 'S':
				e.mode = ModeMenu
				e.menubar.OpenMenu(3) // Search
				e.updateViewportSize()
				return e, nil
			case 'o', 'O':
				e.mode = ModeMenu
				e.menubar.OpenMenu(4) // Options
				e.updateViewportSize()
				return e, nil
			case 'h', 'H':
				e.mode = ModeMenu
				e.menubar.OpenMenu(5) // Help
				e.updateViewportSize()
				return e, nil
			case '<': // Alt+< (same as nano)
				if e.bufferCount() > 1 {
					e.prevBuffer()
				}
				return e, nil
			case '>': // Alt+> (same as nano)
				if e.bufferCount() > 1 {
					e.nextBuffer()
				}
				return e, nil
			}
		}
		// Regular character input - skip control characters (ASCII 0-31 except tab)
		for _, r := range msg.Runes {
			if r >= 32 || r == '\t' {
				e.insertChar(r)
			}
		}
		if len(msg.Runes) > 0 {
			e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		}
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
	case "alt+b":
		e.mode = ModeMenu
		e.menubar.OpenMenu(1) // Buffers
		e.updateViewportSize()
		return e, nil
	case "alt+e":
		e.mode = ModeMenu
		e.menubar.OpenMenu(2) // Edit
		e.updateViewportSize()
		return e, nil
	case "alt+s":
		e.mode = ModeMenu
		e.menubar.OpenMenu(3) // Search
		e.updateViewportSize()
		return e, nil
	case "alt+o":
		e.mode = ModeMenu
		e.menubar.OpenMenu(4) // Options
		e.updateViewportSize()
		return e, nil
	case "alt+h":
		e.mode = ModeMenu
		e.menubar.OpenMenu(5) // Help
		e.updateViewportSize()
		return e, nil
	case "f10":
		e.mode = ModeMenu
		e.menubar.OpenMenu(0)
		e.updateViewportSize()
		return e, nil
	case "f2":
		e.insertLoremIpsum()
		return e, nil

	// Shift+arrow selection (string-based fallback)
	case "shift+left":
		e.moveWithSelection(e.activeDoc().cursor.MoveLeft)
		return e, nil
	case "shift+right":
		e.moveWithSelection(e.activeDoc().cursor.MoveRight)
		return e, nil
	case "shift+up":
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveUpVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
				if newLine == e.activeDoc().cursor.Line() && newCol == e.activeDoc().cursor.Col() {
					return false
				}
				e.activeDoc().cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.activeDoc().cursor.MoveUp()
		})
		return e, nil
	case "shift+down":
		e.moveWithSelection(func() bool {
			if e.viewport.WordWrap() {
				newLine, newCol := e.viewport.MoveDownVisual(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
				if newLine == e.activeDoc().cursor.Line() && newCol == e.activeDoc().cursor.Col() {
					return false
				}
				e.activeDoc().cursor.SetPosition(newLine, newCol)
				return true
			}
			return e.activeDoc().cursor.MoveDown()
		})
		return e, nil
	case "shift+home":
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToLineStart()
			return true
		})
		return e, nil
	case "shift+end":
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToLineEnd()
			return true
		})
		return e, nil

	// Ctrl+Shift combinations
	case "ctrl+shift+left":
		e.moveWithSelection(e.activeDoc().cursor.MoveWordLeft)
		return e, nil
	case "ctrl+shift+right":
		e.moveWithSelection(e.activeDoc().cursor.MoveWordRight)
		return e, nil
	case "ctrl+shift+home":
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToStart()
			return true
		})
		return e, nil
	case "ctrl+shift+end":
		e.moveWithSelection(func() bool {
			e.activeDoc().cursor.MoveToEnd()
			return true
		})
		return e, nil
	}

	return e, nil
}

// moveWithSelection moves the cursor while extending the selection
func (e *Editor) moveWithSelection(move func() bool) {
	if !e.activeDoc().selection.Active {
		e.activeDoc().selection.Start(e.activeDoc().cursor.ByteOffset())
	}
	move()
	e.activeDoc().selection.Update(e.activeDoc().cursor.ByteOffset())
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
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
				e.promptText = "File exists. Overwrite? (y/N): "
				e.promptInput = ""
				e.promptAction = PromptConfirmOverwrite
				e.mode = ModePrompt // Stay in prompt mode
				return
			}
			e.activeDoc().filename = input
			e.doSave()
		} else {
			e.statusbar.SetMessage("Save cancelled - no filename", "info")
		}

	case PromptConfirmOverwrite:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.activeDoc().filename = e.pendingFilename
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
			e.activeDoc().modified = false // Discard changes
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

	case PromptFileChanged:
		if strings.ToLower(input) == "y" || strings.ToLower(input) == "yes" {
			e.doSave() // Overwrite the external changes
		} else {
			e.statusbar.SetMessage("Save cancelled", "info")
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
		totalLines := e.activeDoc().buffer.LineCount()
		if lineNum < 1 {
			e.statusbar.SetMessage("Line number must be at least 1", "error")
			return
		}
		if lineNum > totalLines {
			e.statusbar.SetMessage(fmt.Sprintf("Line %d exceeds file length (%d lines)", lineNum, totalLines), "error")
			return
		}
		// Convert to 0-indexed
		e.activeDoc().cursor.SetPosition(lineNum-1, 0)
		e.activeDoc().selection.Clear()
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		e.statusbar.SetMessage(fmt.Sprintf("Jumped to line %d", lineNum), "info")

	case PromptThemeCopyName:
		if input == "" {
			e.statusbar.SetMessage("Cancelled - no name provided", "info")
			return
		}
		// Get the source theme and export with new name
		theme := config.GetTheme(e.themeExportName)
		theme.Name = input // Update theme name
		path, err := config.ExportTheme(theme, input)
		if err != nil {
			e.statusbar.SetMessage("Error copying theme: "+err.Error(), "error")
		} else {
			// Open the theme file in a buffer for editing
			if err := e.LoadFile(path); err != nil {
				e.statusbar.SetMessage("Theme saved but couldn't open: "+err.Error(), "error")
			} else {
				e.statusbar.SetMessage("Theme saved and opened for editing: "+path, "success")
			}
		}
		e.themeExportName = ""
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

			// Check if click is on minimap
			if e.minimapRenderer.IsEnabled() && y >= 0 && y < e.viewport.Height() {
				// Calculate minimap position (before scrollbar)
				scrollbarWidth := 0
				if e.scrollbar.IsEnabled() {
					scrollbarWidth = e.scrollbar.Width()
				}
				minimapStartX := e.width - scrollbarWidth - ui.MinimapWidth()
				minimapEndX := e.width - scrollbarWidth

				if msg.X >= minimapStartX && msg.X < minimapEndX {
					lines := e.activeDoc().buffer.Lines()

					// Get minimap metrics and convert click to visual line
					renderState := e.buildRenderState()
					metrics := e.minimapRenderer.GetMetrics(e.viewport.Height(), renderState)
					visualLine := e.minimapRenderer.RowToVisualLine(y, metrics)

					// Convert visual line to buffer line
					var targetLine int
					if e.viewport.WordWrap() {
						targetLine, _ = e.viewport.VisualLineToBufferLine(lines, visualLine)
					} else {
						targetLine = visualLine
					}

					e.activeDoc().cursor.SetPosition(targetLine, 0)
					e.activeDoc().selection.Clear()
					e.viewport.EnsureCursorVisibleWrapped(lines, e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
					return e, nil
				}
			}

			// Check if click is on scrollbar
			if e.scrollbar.IsEnabled() && y >= 0 && y < e.viewport.Height() {
				scrollbarStartX := e.width - e.scrollbar.Width()
				if msg.X >= scrollbarStartX {
					lines := e.activeDoc().buffer.Lines()

					// Calculate total lines - use visual lines if word wrap is enabled
					totalLines := len(lines)
					if e.viewport.WordWrap() {
						totalLines = e.viewport.CountVisualLines(lines)
					}

					// Convert scrollbar row to visual line
					visualLine := e.scrollbar.RowToLine(y, totalLines, e.viewport.Height())

					// Convert visual line to buffer line
					var targetLine int
					if e.viewport.WordWrap() {
						targetLine, _ = e.viewport.VisualLineToBufferLine(lines, visualLine)
					} else {
						targetLine = visualLine
					}

					e.activeDoc().cursor.SetPosition(targetLine, 0)
					e.activeDoc().selection.Clear()
					e.viewport.EnsureCursorVisibleWrapped(lines, e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
					return e, nil
				}
			}

			// Handle click in editor area
			if y >= 0 && y < e.viewport.Height() {
				line, col := e.viewport.PositionFromClickWrapped(e.activeDoc().buffer.Lines(), msg.X, y)
				e.activeDoc().cursor.SetPosition(line, col)
				e.activeDoc().selection.Clear()
				e.mouseDown = true
				e.mouseStartX = msg.X
				e.mouseStartY = y
			}
		} else if msg.Action == tea.MouseActionRelease {
			e.mouseDown = false
		} else if msg.Action == tea.MouseActionMotion && e.mouseDown {
			// Drag selection
			if y >= 0 && y < e.viewport.Height() {
				if !e.activeDoc().selection.Active {
					startLine, startCol := e.viewport.PositionFromClickWrapped(e.activeDoc().buffer.Lines(), e.mouseStartX, e.mouseStartY)
					startPos := e.activeDoc().buffer.LineColToPosition(startLine, startCol)
					e.activeDoc().selection.Start(startPos)
				}
				line, col := e.viewport.PositionFromClickWrapped(e.activeDoc().buffer.Lines(), msg.X, y)
				e.activeDoc().cursor.SetPosition(line, col)
				e.activeDoc().selection.Update(e.activeDoc().cursor.ByteOffset())
			}
		}

	case tea.MouseButtonWheelUp:
		e.viewport.ScrollUp()

	case tea.MouseButtonWheelDown:
		e.viewport.ScrollDownWrapped(e.activeDoc().buffer.Lines())
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
	case ui.ActionRecentFiles:
		e.showRecentFiles()
	case ui.ActionRecentDirs:
		e.showRecentDirs()
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
	case ui.ActionScrollbar:
		e.toggleScrollbar()
	case ui.ActionMinimap:
		e.toggleMinimap()
	case ui.ActionTheme:
		e.showThemeDialog()
	case ui.ActionKeybindings:
		e.showKeybindingsDialog()
	case ui.ActionSettings:
		e.showSettingsDialog()
	case ui.ActionBuffer1:
		e.switchToBuffer(0)
	case ui.ActionBuffer2:
		e.switchToBuffer(1)
	case ui.ActionBuffer3:
		e.switchToBuffer(2)
	case ui.ActionBuffer4:
		e.switchToBuffer(3)
	case ui.ActionBuffer5:
		e.switchToBuffer(4)
	case ui.ActionBuffer6:
		e.switchToBuffer(5)
	case ui.ActionBuffer7:
		e.switchToBuffer(6)
	case ui.ActionBuffer8:
		e.switchToBuffer(7)
	case ui.ActionBuffer9:
		e.switchToBuffer(8)
	case ui.ActionBuffer10:
		e.switchToBuffer(9)
	case ui.ActionBuffer11:
		e.switchToBuffer(10)
	case ui.ActionBuffer12:
		e.switchToBuffer(11)
	case ui.ActionBuffer13:
		e.switchToBuffer(12)
	case ui.ActionBuffer14:
		e.switchToBuffer(13)
	case ui.ActionBuffer15:
		e.switchToBuffer(14)
	case ui.ActionBuffer16:
		e.switchToBuffer(15)
	case ui.ActionBuffer17:
		e.switchToBuffer(16)
	case ui.ActionBuffer18:
		e.switchToBuffer(17)
	case ui.ActionBuffer19:
		e.switchToBuffer(18)
	case ui.ActionBuffer20:
		e.switchToBuffer(19)
	case ui.ActionHelp:
		e.showHelp()
	case ui.ActionAbout:
		e.showAbout()
	case ui.ActionSetEncoding:
		e.showEncodingDialog()
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
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())

	// Save to config
	e.saveConfig()
}

// toggleLineNumbers toggles line numbers on/off
func (e *Editor) toggleLineNumbers() {
	show := !e.viewport.ShowLineNum()
	e.viewport.ShowLineNumbers(show)

	// Update compositor columns
	e.setupCompositorColumns()

	// Update menu checkbox
	if show {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[x] Line Numbers")
		e.statusbar.SetMessage("Line numbers enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[ ] Line Numbers")
		e.statusbar.SetMessage("Line numbers disabled", "info")
	}

	// Ensure cursor stays visible after toggle (text width changes)
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())

	// Save to config
	e.saveConfig()
}

// toggleSyntaxHighlight toggles syntax highlighting on/off
func (e *Editor) toggleSyntaxHighlight() {
	enabled := !e.activeDoc().highlighter.Enabled()
	e.activeDoc().highlighter.SetEnabled(enabled)

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

// toggleScrollbar toggles the code scrollbar on/off
func (e *Editor) toggleScrollbar() {
	enabled := e.scrollbar.Toggle()

	// Update viewport to account for scrollbar width change
	e.viewport.SetScrollbarWidth(e.scrollbar.Width())

	// Update compositor columns
	e.setupCompositorColumns()

	// Update menu checkbox
	if enabled {
		e.menubar.SetItemLabel(ui.ActionScrollbar, "[x] Scrollbar")
		e.statusbar.SetMessage("Scrollbar enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionScrollbar, "[ ] Scrollbar")
		e.statusbar.SetMessage("Scrollbar disabled", "info")
	}

	// Ensure cursor stays visible after toggle (text width changes)
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())

	// Save to config
	e.saveConfig()
}

// toggleMinimap toggles the minimap on/off
func (e *Editor) toggleMinimap() {
	enabled := e.minimapRenderer.Toggle()

	// Update compositor columns
	e.setupCompositorColumns()

	// Update menu checkbox
	if enabled {
		e.menubar.SetItemLabel(ui.ActionMinimap, "[x] Minimap")
		e.statusbar.SetMessage("Minimap enabled", "info")
	} else {
		e.menubar.SetItemLabel(ui.ActionMinimap, "[ ] Minimap")
		e.statusbar.SetMessage("Minimap disabled", "info")
		// Clear Kitty graphics image if applicable
		e.pendingEscapes += e.minimapRenderer.ClearImage()
	}

	// Ensure cursor stays visible after toggle (text width changes)
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())

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
	e.config.Editor.SyntaxHighlight = e.activeDoc().highlighter.Enabled()
	e.config.Editor.Scrollbar = e.scrollbar.IsEnabled()
	e.config.Editor.Minimap = e.minimapRenderer.IsEnabled()
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
	e.scrollbar.SetStyles(styles)
	e.lineNumRenderer.SetStyles(styles)
	e.textRenderer.SetStyles(styles)
	e.minimapRenderer.SetStyles(styles)
	e.styles = styles

	// Update syntax highlighter colors
	e.activeDoc().highlighter.SetColors(syntax.SyntaxColors{
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
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "e", "E":
			// Edit: export theme to file and open in buffer
			if e.themeIndex >= 0 && e.themeIndex < len(e.themeList) {
				themeName := e.themeList[e.themeIndex]
				theme := config.GetTheme(themeName)
				path, err := config.ExportTheme(theme, themeName)
				if err != nil {
					e.statusbar.SetMessage("Error exporting theme: "+err.Error(), "error")
				} else {
					// Open the theme file in a buffer for editing
					if err := e.LoadFile(path); err != nil {
						e.statusbar.SetMessage("Theme saved but couldn't open: "+err.Error(), "error")
					} else {
						e.statusbar.SetMessage("Theme exported and opened for editing", "success")
					}
				}
				e.mode = ModeNormal
			}
		case "c", "C":
			// Copy: prompt for new name
			if e.themeIndex >= 0 && e.themeIndex < len(e.themeList) {
				e.themeExportName = e.themeList[e.themeIndex]
				e.promptText = "New theme name: "
				e.promptInput = e.themeExportName + "_copy"
				e.promptAction = PromptThemeCopyName
				e.mode = ModePrompt
			}
		}
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

// showRecentFiles opens the recent files dialog
func (e *Editor) showRecentFiles() {
	if e.config == nil || len(e.config.RecentFiles) == 0 {
		e.statusbar.SetMessage("No recent files", "info")
		return
	}
	e.recentFilesIndex = 0
	e.mode = ModeRecentFiles
}

// handleRecentFilesKey handles key events in the recent files dialog
func (e *Editor) handleRecentFilesKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	recentCount := 0
	if e.config != nil {
		recentCount = len(e.config.RecentFiles)
	}

	switch msg.Type {
	case tea.KeyUp:
		if e.recentFilesIndex > 0 {
			e.recentFilesIndex--
		}
	case tea.KeyDown:
		if e.recentFilesIndex < recentCount-1 {
			e.recentFilesIndex++
		}
	case tea.KeyEnter:
		// Open selected file
		if e.recentFilesIndex >= 0 && e.recentFilesIndex < recentCount {
			path := e.config.RecentFiles[e.recentFilesIndex]
			if err := e.LoadFile(path); err != nil {
				e.statusbar.SetMessage("Open failed: "+err.Error(), "error")
			} else {
				e.statusbar.SetMessage("Opened: "+path, "success")
			}
		}
		e.mode = ModeNormal
	case tea.KeyEsc:
		e.mode = ModeNormal
	case tea.KeyDelete, tea.KeyBackspace:
		// Remove selected file from recent list
		if e.recentFilesIndex >= 0 && e.recentFilesIndex < recentCount {
			e.config.RecentFiles = append(
				e.config.RecentFiles[:e.recentFilesIndex],
				e.config.RecentFiles[e.recentFilesIndex+1:]...,
			)
			go e.config.Save()
			// Adjust index if needed
			if e.recentFilesIndex >= len(e.config.RecentFiles) {
				e.recentFilesIndex = len(e.config.RecentFiles) - 1
			}
			// Close dialog if list is now empty
			if len(e.config.RecentFiles) == 0 {
				e.mode = ModeNormal
				e.statusbar.SetMessage("Recent files cleared", "info")
			}
		}
	}
	return e, nil
}

// handleRecentFilesMouse handles mouse input in the recent files dialog
func (e *Editor) handleRecentFilesMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	recentCount := 0
	if e.config != nil {
		recentCount = len(e.config.RecentFiles)
	}
	if recentCount == 0 {
		return e, nil
	}

	// Calculate dialog position (must match overlayRecentFilesDialog)
	boxWidth := 60
	boxHeight := recentCount + 5 // title, empty, items..., empty, footer, bottom

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

	// File list starts at line 2 (after title border and empty line)
	listStart := 2
	listEnd := listStart + recentCount

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			if relY >= listStart && relY < listEnd {
				clickedIdx := relY - listStart
				if clickedIdx >= 0 && clickedIdx < recentCount {
					if e.recentFilesIndex == clickedIdx {
						// Double-click effect: open file
						path := e.config.RecentFiles[e.recentFilesIndex]
						if err := e.LoadFile(path); err != nil {
							e.statusbar.SetMessage("Open failed: "+err.Error(), "error")
						} else {
							e.statusbar.SetMessage("Opened: "+path, "success")
						}
						e.mode = ModeNormal
					} else {
						e.recentFilesIndex = clickedIdx
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if e.recentFilesIndex > 0 {
			e.recentFilesIndex--
		}

	case tea.MouseButtonWheelDown:
		if e.recentFilesIndex < recentCount-1 {
			e.recentFilesIndex++
		}
	}

	return e, nil
}

// showRecentDirs opens the recent directories dialog
func (e *Editor) showRecentDirs() {
	if e.config == nil || len(e.config.RecentDirs) == 0 {
		e.statusbar.SetMessage("No recent directories", "info")
		return
	}
	e.recentDirsIndex = 0
	e.mode = ModeRecentDirs
}

// handleRecentDirsKey handles key events in the recent directories dialog
func (e *Editor) handleRecentDirsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	recentCount := 0
	if e.config != nil {
		recentCount = len(e.config.RecentDirs)
	}

	switch msg.Type {
	case tea.KeyUp:
		if e.recentDirsIndex > 0 {
			e.recentDirsIndex--
		}
	case tea.KeyDown:
		if e.recentDirsIndex < recentCount-1 {
			e.recentDirsIndex++
		}
	case tea.KeyEnter:
		// Navigate to selected directory in file browser
		if e.recentDirsIndex >= 0 && e.recentDirsIndex < recentCount {
			path := e.config.RecentDirs[e.recentDirsIndex]
			e.fileBrowserDir = path
			e.fileBrowserSelected = 0
			e.fileBrowserScroll = 0
			e.fileBrowserError = ""
			e.loadDirectory(path)
			e.mode = ModeFileBrowser
			e.statusbar.SetMessage("Browsing: "+path, "info")
		}
	case tea.KeyEsc:
		e.mode = ModeNormal
	case tea.KeyDelete, tea.KeyBackspace:
		// Remove selected directory from recent list
		if e.recentDirsIndex >= 0 && e.recentDirsIndex < recentCount {
			e.config.RecentDirs = append(
				e.config.RecentDirs[:e.recentDirsIndex],
				e.config.RecentDirs[e.recentDirsIndex+1:]...,
			)
			go e.config.Save()
			// Adjust index if needed
			if e.recentDirsIndex >= len(e.config.RecentDirs) {
				e.recentDirsIndex = len(e.config.RecentDirs) - 1
			}
			// Close dialog if list is now empty
			if len(e.config.RecentDirs) == 0 {
				e.mode = ModeNormal
				e.statusbar.SetMessage("Recent directories cleared", "info")
			}
		}
	}
	return e, nil
}

// handleRecentDirsMouse handles mouse input in the recent directories dialog
func (e *Editor) handleRecentDirsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	recentCount := 0
	if e.config != nil {
		recentCount = len(e.config.RecentDirs)
	}
	if recentCount == 0 {
		return e, nil
	}

	// Calculate dialog position (must match overlayRecentDirsDialog)
	boxWidth := 60
	boxHeight := recentCount + 5 // title, empty, items..., empty, footer, bottom

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

	// Directory list starts at line 2 (after title border and empty line)
	listStart := 2
	listEnd := listStart + recentCount

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			if relY >= listStart && relY < listEnd {
				clickedIdx := relY - listStart
				if clickedIdx >= 0 && clickedIdx < recentCount {
					if e.recentDirsIndex == clickedIdx {
						// Double-click effect: open directory in browser
						path := e.config.RecentDirs[e.recentDirsIndex]
						e.fileBrowserDir = path
						e.fileBrowserSelected = 0
						e.fileBrowserScroll = 0
						e.fileBrowserError = ""
						e.loadDirectory(path)
						e.mode = ModeFileBrowser
						e.statusbar.SetMessage("Browsing: "+path, "info")
					} else {
						e.recentDirsIndex = clickedIdx
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if e.recentDirsIndex > 0 {
			e.recentDirsIndex--
		}

	case tea.MouseButtonWheelDown:
		if e.recentDirsIndex < recentCount-1 {
			e.recentDirsIndex++
		}
	}

	return e, nil
}

// handleConfigErrorKey handles key events in the config error dialog
func (e *Editor) handleConfigErrorKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyLeft:
		if e.configErrorChoice > 0 {
			e.configErrorChoice--
		}
	case tea.KeyRight:
		if e.configErrorChoice < 2 {
			e.configErrorChoice++
		}
	case tea.KeyEnter:
		return e.executeConfigErrorChoice()
	case tea.KeyEsc:
		// Escape = Use Defaults
		e.configErrorChoice = 1
		return e.executeConfigErrorChoice()
	case tea.KeyRunes:
		// Hotkeys: E=Edit, D=Defaults, Q=Quit
		switch string(msg.Runes) {
		case "e", "E":
			e.configErrorChoice = 0
			return e.executeConfigErrorChoice()
		case "d", "D":
			e.configErrorChoice = 1
			return e.executeConfigErrorChoice()
		case "q", "Q":
			e.configErrorChoice = 2
			return e.executeConfigErrorChoice()
		}
	}
	return e, nil
}

// executeConfigErrorChoice executes the selected config error action
func (e *Editor) executeConfigErrorChoice() (tea.Model, tea.Cmd) {
	switch e.configErrorChoice {
	case 0: // Edit File
		e.mode = ModeNormal
		if err := e.LoadFile(e.configErrorFile); err != nil {
			e.statusbar.SetMessage("Could not open config: "+err.Error(), "error")
		} else {
			e.statusbar.SetMessage("Edit config file, then restart Textivus", "info")
		}
	case 1: // Use Defaults
		e.mode = ModeNormal
		e.statusbar.SetMessage("Using default settings", "info")
	case 2: // Quit
		return e, tea.Quit
	}
	return e, nil
}

// handleConfigErrorMouse handles mouse input in the config error dialog
func (e *Editor) handleConfigErrorMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Dialog dimensions (must match overlayConfigErrorDialog)
	boxWidth := 56
	boxHeight := 9

	startX := (e.width - boxWidth) / 2
	startY := (e.viewport.Height() - boxHeight) / 2

	// Adjust mouse Y for menu bar
	mouseY := msg.Y - 1

	// Calculate relative position within dialog
	relX := msg.X - startX
	relY := mouseY - startY

	// Check if click is outside dialog - treat as "Use Defaults"
	if relX < 0 || relX >= boxWidth || relY < 0 || relY >= boxHeight {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.configErrorChoice = 1
			return e.executeConfigErrorChoice()
		}
		return e, nil
	}

	// Button row is at line 7 (0-indexed)
	buttonRowY := 7
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		if relY == buttonRowY {
			// Buttons: [ Edit File ]  [ Use Defaults ]  [ Quit ]
			// Approximate positions within inner width (54)
			innerX := relX - 1 // Account for border
			if innerX >= 2 && innerX < 15 {
				e.configErrorChoice = 0
				return e.executeConfigErrorChoice()
			} else if innerX >= 17 && innerX < 34 {
				e.configErrorChoice = 1
				return e.executeConfigErrorChoice()
			} else if innerX >= 36 && innerX < 46 {
				e.configErrorChoice = 2
				return e.executeConfigErrorChoice()
			}
		}
	}

	return e, nil
}

// showSettingsDialog opens the settings dialog
func (e *Editor) showSettingsDialog() {
	// Load current values into dialog state
	if e.config != nil {
		e.settingsWordWrap = e.config.Editor.WordWrap
		e.settingsLineNumbers = e.config.Editor.LineNumbers
		e.settingsSyntax = e.config.Editor.SyntaxHighlight
		e.settingsScrollbar = e.config.Editor.Scrollbar
		e.settingsBackupCount = e.config.Editor.BackupCount
		e.settingsMaxBuffers = e.config.Editor.MaxBuffers
		e.settingsTabWidth = e.config.Editor.TabWidth
		if e.settingsTabWidth <= 0 {
			e.settingsTabWidth = 4
		}
		e.settingsTabsToSpaces = e.config.Editor.TabsToSpaces
	}
	e.settingsIndex = 0
	e.mode = ModeSettings
}

// handleSettingsKey handles key events in the settings dialog
func (e *Editor) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Settings rows: 0-4 = checkboxes, 5-7 = numbers, 8 = Save, 9 = Cancel
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
		maxRow          = 9
	)

	switch msg.Type {
	case tea.KeyUp:
		if e.settingsIndex > 0 {
			e.settingsIndex--
		}
	case tea.KeyDown:
		if e.settingsIndex < maxRow {
			e.settingsIndex++
		}
	case tea.KeyLeft:
		// Decrease number inputs or navigate to Save button
		switch e.settingsIndex {
		case rowBackupCount:
			if e.settingsBackupCount > 0 {
				e.settingsBackupCount--
			}
		case rowMaxBuffers:
			if e.settingsMaxBuffers > 0 {
				e.settingsMaxBuffers--
			}
		case rowTabWidth:
			if e.settingsTabWidth > 1 {
				e.settingsTabWidth--
			}
		case rowCancel:
			e.settingsIndex = rowSave
		}
	case tea.KeyRight:
		// Increase number inputs or navigate to Cancel button
		switch e.settingsIndex {
		case rowBackupCount:
			if e.settingsBackupCount < 99 {
				e.settingsBackupCount++
			}
		case rowMaxBuffers:
			if e.settingsMaxBuffers < 99 {
				e.settingsMaxBuffers++
			}
		case rowTabWidth:
			if e.settingsTabWidth < 16 {
				e.settingsTabWidth++
			}
		case rowSave:
			e.settingsIndex = rowCancel
		}
	case tea.KeyEnter, tea.KeySpace:
		switch e.settingsIndex {
		case rowWordWrap:
			e.settingsWordWrap = !e.settingsWordWrap
		case rowLineNumbers:
			e.settingsLineNumbers = !e.settingsLineNumbers
		case rowSyntax:
			e.settingsSyntax = !e.settingsSyntax
		case rowScrollbar:
			e.settingsScrollbar = !e.settingsScrollbar
		case rowTabsToSpaces:
			e.settingsTabsToSpaces = !e.settingsTabsToSpaces
		case rowSave:
			e.saveSettings()
			e.mode = ModeNormal
			e.statusbar.SetMessage("Settings saved", "success")
		case rowCancel:
			e.mode = ModeNormal
		}
	case tea.KeyEsc:
		e.mode = ModeNormal
	}
	return e, nil
}

// saveSettings applies and saves the settings to config
func (e *Editor) saveSettings() {
	if e.config == nil {
		return
	}

	// Apply to config
	e.config.Editor.WordWrap = e.settingsWordWrap
	e.config.Editor.LineNumbers = e.settingsLineNumbers
	e.config.Editor.SyntaxHighlight = e.settingsSyntax
	e.config.Editor.Scrollbar = e.settingsScrollbar
	e.config.Editor.BackupCount = e.settingsBackupCount
	e.config.Editor.MaxBuffers = e.settingsMaxBuffers
	e.config.Editor.TabWidth = e.settingsTabWidth
	e.config.Editor.TabsToSpaces = e.settingsTabsToSpaces

	// Apply to current editor state
	e.viewport.SetWordWrap(e.settingsWordWrap)
	e.viewport.ShowLineNumbers(e.settingsLineNumbers)
	e.activeDoc().highlighter.SetEnabled(e.settingsSyntax)
	e.scrollbar.SetEnabled(e.settingsScrollbar)
	e.viewport.SetScrollbarWidth(e.scrollbar.Width())

	// Update compositor columns to reflect changes
	e.setupCompositorColumns()

	// Update menu checkboxes to reflect new state
	if e.settingsWordWrap {
		e.menubar.SetItemLabel(ui.ActionWordWrap, "[x] Word Wrap")
	} else {
		e.menubar.SetItemLabel(ui.ActionWordWrap, "[ ] Word Wrap")
	}
	if e.settingsLineNumbers {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[x] Line Numbers")
	} else {
		e.menubar.SetItemLabel(ui.ActionLineNumbers, "[ ] Line Numbers")
	}
	if e.settingsSyntax {
		e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[x] Syntax Highlight")
	} else {
		e.menubar.SetItemLabel(ui.ActionSyntaxHighlight, "[ ] Syntax Highlight")
	}
	if e.settingsScrollbar {
		e.menubar.SetItemLabel(ui.ActionScrollbar, "[x] Scrollbar")
	} else {
		e.menubar.SetItemLabel(ui.ActionScrollbar, "[ ] Scrollbar")
	}

	// Save to disk
	go e.config.Save()
}

// handleSettingsMouse handles mouse input in the settings dialog
func (e *Editor) handleSettingsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Dialog dimensions (must match overlaySettingsDialog)
	// title + empty + 5 checkboxes + empty + 3 numbers with help + empty + buttons + bottom
	boxWidth := 54
	boxHeight := 18

	startX := (e.width - boxWidth) / 2
	startY := (e.viewport.Height() - boxHeight) / 2

	mouseY := msg.Y - 1 // Adjust for menu bar
	relX := msg.X - startX
	relY := mouseY - startY

	// Click outside = cancel
	if relX < 0 || relX >= boxWidth || relY < 0 || relY >= boxHeight {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.mode = ModeNormal
		}
		return e, nil
	}

	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		// Row mapping (0-indexed from content start at line 2):
		// 0-4: checkboxes (rows 0-4)
		// 5: empty line
		// 6: Backup Count (row 5)
		// 7: help text
		// 8: Max Buffers (row 6)
		// 9: help text
		// 10: Tab Width (row 7)
		// 11: help text
		// 12: empty line
		// 13: buttons (rows 8, 9)

		contentRow := relY - 2
		if contentRow >= 0 && contentRow <= 4 {
			// Checkbox rows
			e.settingsIndex = contentRow
			switch contentRow {
			case 0:
				e.settingsWordWrap = !e.settingsWordWrap
			case 1:
				e.settingsLineNumbers = !e.settingsLineNumbers
			case 2:
				e.settingsSyntax = !e.settingsSyntax
			case 3:
				e.settingsScrollbar = !e.settingsScrollbar
			case 4:
				e.settingsTabsToSpaces = !e.settingsTabsToSpaces
			}
		} else if contentRow == 6 {
			e.settingsIndex = 5 // Backup Count
		} else if contentRow == 8 {
			e.settingsIndex = 6 // Max Buffers
		} else if contentRow == 10 {
			e.settingsIndex = 7 // Tab Width
		} else if contentRow == 14 {
			// Button row
			innerX := relX - 1
			if innerX >= 12 && innerX < 22 {
				e.saveSettings()
				e.mode = ModeNormal
				e.statusbar.SetMessage("Settings saved", "success")
			} else if innerX >= 28 && innerX < 40 {
				e.mode = ModeNormal
			}
		}
	}

	return e, nil
}

// showEncodingDialog opens the encoding selection dialog
func (e *Editor) showEncodingDialog() {
	// Find the current encoding index
	encodings := enc.GetSupportedEncodings()
	currentID := "utf-8"
	if e.activeDoc().encoding != nil {
		currentID = e.activeDoc().encoding.ID
	}

	e.encodingIndex = 0
	for i, encoding := range encodings {
		if encoding.ID == currentID {
			e.encodingIndex = i
			break
		}
	}

	e.mode = ModeEncoding
}

// handleEncodingKey handles key events in the encoding selection dialog
func (e *Editor) handleEncodingKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	encodings := enc.GetSupportedEncodings()
	count := len(encodings)

	switch msg.Type {
	case tea.KeyUp:
		if e.encodingIndex > 0 {
			e.encodingIndex--
		}
	case tea.KeyDown:
		if e.encodingIndex < count-1 {
			e.encodingIndex++
		}
	case tea.KeyHome:
		e.encodingIndex = 0
	case tea.KeyEnd:
		e.encodingIndex = count - 1
	case tea.KeyEsc:
		e.mode = ModeNormal
	case tea.KeyEnter:
		// Apply the selected encoding
		selectedEnc := encodings[e.encodingIndex]
		e.applyEncoding(selectedEnc)
		e.mode = ModeNormal
	}

	return e, nil
}

// handleEncodingMouse handles mouse input in the encoding selection dialog
func (e *Editor) handleEncodingMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	encodings := enc.GetSupportedEncodings()
	encodingCount := len(encodings)

	// Dialog dimensions (must match overlayEncodingDialog)
	boxWidth := 50
	// title + empty + encodings + empty + help + footer + bottom border
	boxHeight := encodingCount + 6

	startX := (e.width - boxWidth) / 2
	startY := (e.viewport.Height() - boxHeight) / 2

	mouseY := msg.Y - 1 // Adjust for menu bar
	relX := msg.X - startX
	relY := mouseY - startY

	// Click outside = cancel
	if relX < 0 || relX >= boxWidth || relY < 0 || relY >= boxHeight {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.mode = ModeNormal
		}
		return e, nil
	}

	// Encoding items start at row 2 (after title and empty line)
	itemRow := relY - 2
	if itemRow >= 0 && itemRow < encodingCount {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.encodingIndex = itemRow
		}
		// Double-click to select
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionRelease {
			if e.encodingIndex == itemRow {
				selectedEnc := encodings[e.encodingIndex]
				e.applyEncoding(selectedEnc)
				e.mode = ModeNormal
			}
		}
	}

	return e, nil
}

// applyEncoding changes the encoding the document will be saved as
// This does NOT reload the file - it just changes the save encoding
func (e *Editor) applyEncoding(newEnc *enc.Encoding) {
	doc := e.activeDoc()
	if doc == nil {
		return
	}

	oldEnc := doc.encoding
	if oldEnc != nil && oldEnc.ID == newEnc.ID {
		// Same encoding, nothing to do
		return
	}

	// Just change the encoding - content stays the same
	doc.encoding = newEnc
	e.statusbar.SetMessage("Will save as "+newEnc.Name, "info")
}

// showKeybindingsDialog opens the keybindings configuration dialog
func (e *Editor) showKeybindingsDialog() {
	e.kbDialogIndex = 0
	e.kbDialogScroll = 0
	e.kbDialogEditing = false
	e.kbDialogEditField = 0
	e.kbDialogMessage = ""
	e.kbDialogMsgError = false
	e.kbDialogConfirm = false
	e.mode = ModeKeybindings
}

// handleKeybindingsKey handles key events in the keybindings dialog
func (e *Editor) handleKeybindingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	actions := config.AllActions()
	actionCount := len(actions)

	// If we're in confirmation mode, handle y/n
	if e.kbDialogConfirm {
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'y', 'Y':
				// Confirmed - reset to defaults
				e.keybindings = config.DefaultKeybindings()
				go e.keybindings.Save()
				e.kbDialogMessage = "Reset to defaults"
				e.kbDialogMsgError = false
			case 'n', 'N':
				e.kbDialogMessage = ""
			}
		} else if msg.Type == tea.KeyEsc {
			e.kbDialogMessage = ""
		}
		e.kbDialogConfirm = false
		return e, nil
	}

	// If we're in editing mode (waiting for a key), capture the key
	if e.kbDialogEditing {
		// Escape cancels editing mode
		if msg.Type == tea.KeyEsc {
			e.kbDialogEditing = false
			return e, nil
		}

		// Delete/Backspace clears the binding
		if msg.Type == tea.KeyDelete || msg.Type == tea.KeyBackspace {
			action := actions[e.kbDialogIndex]
			binding := e.keybindings.GetBinding(action)
			if e.kbDialogEditField == 0 {
				binding.Primary = ""
			} else {
				binding.Alternate = ""
			}
			e.keybindings.SetBinding(action, binding)
			e.kbDialogEditing = false
			go e.keybindings.Save()
			return e, nil
		}

		// Convert key press to a key string
		keyStr := e.keyMsgToString(msg)
		if keyStr != "" {
			action := actions[e.kbDialogIndex]
			binding := e.keybindings.GetBinding(action)

			// Check for conflicts
			conflicts := e.checkKeyConflict(keyStr, action)
			if len(conflicts) > 0 {
				conflictNames := make([]string, len(conflicts))
				for i, c := range conflicts {
					conflictNames[i] = config.ActionNames[c]
				}
				e.kbDialogMessage = "Conflict: " + strings.Join(conflictNames, ", ")
				e.kbDialogMsgError = true
				e.kbDialogEditing = false
				return e, nil
			}

			if e.kbDialogEditField == 0 {
				binding.Primary = keyStr
			} else {
				binding.Alternate = keyStr
			}
			e.keybindings.SetBinding(action, binding)
			e.kbDialogEditing = false
			e.kbDialogMessage = ""
			e.kbDialogMsgError = false
			go e.keybindings.Save()
		}
		return e, nil
	}

	// Normal navigation mode
	switch msg.Type {
	case tea.KeyUp:
		if e.kbDialogIndex > 0 {
			e.kbDialogIndex--
			e.ensureKbDialogVisible()
		}
	case tea.KeyDown:
		if e.kbDialogIndex < actionCount-1 {
			e.kbDialogIndex++
			e.ensureKbDialogVisible()
		}
	case tea.KeyLeft:
		// Switch to primary binding field
		e.kbDialogEditField = 0
	case tea.KeyRight:
		// Switch to alternate binding field
		e.kbDialogEditField = 1
	case tea.KeyEnter:
		// Start editing the selected binding field
		e.kbDialogEditing = true
		e.kbDialogMessage = ""
		e.kbDialogMsgError = false
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.kbDialogMessage = ""
	case tea.KeyRunes:
		switch string(msg.Runes) {
		case "r", "R":
			// Ask for confirmation before resetting
			e.kbDialogMessage = "Reset all keybindings to defaults? (y/N)"
			e.kbDialogMsgError = false
			e.kbDialogConfirm = true
		}
	}
	return e, nil
}

// ensureKbDialogVisible adjusts scroll to keep selected item visible
func (e *Editor) ensureKbDialogVisible() {
	visibleItems := e.viewport.Height() - 8 // Account for dialog chrome
	if visibleItems < 5 {
		visibleItems = 5
	}

	if e.kbDialogIndex < e.kbDialogScroll {
		e.kbDialogScroll = e.kbDialogIndex
	} else if e.kbDialogIndex >= e.kbDialogScroll+visibleItems {
		e.kbDialogScroll = e.kbDialogIndex - visibleItems + 1
	}
}

// handleKeybindingsMouse handles mouse input in the keybindings dialog
func (e *Editor) handleKeybindingsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// If editing, any click cancels
	if e.kbDialogEditing {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			e.kbDialogEditing = false
		}
		return e, nil
	}

	actions := config.AllActions()
	actionCount := len(actions)

	// Calculate dialog dimensions (must match overlayKeybindingsDialog)
	boxWidth := 64
	visibleItems := e.viewport.Height() - 8
	if visibleItems > actionCount {
		visibleItems = actionCount
	}
	if visibleItems < 5 {
		visibleItems = 5
	}
	boxHeight := visibleItems + 6 // title, header, items, empty, footer, bottom

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

	// Item list starts at line 2 (after title border and header)
	listStart := 2
	listEnd := listStart + visibleItems

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			if relY >= listStart && relY < listEnd {
				clickedIdx := e.kbDialogScroll + relY - listStart
				if clickedIdx >= 0 && clickedIdx < actionCount {
					// Determine which field was clicked based on X position
					// Layout: "│ Action Name (20)  │ Primary (18) │ Alternate (18) │"
					primaryStart := 23
					alternateStart := 43

					if e.kbDialogIndex == clickedIdx {
						// Same item clicked - check field and start editing
						if relX >= alternateStart {
							e.kbDialogEditField = 1
						} else if relX >= primaryStart {
							e.kbDialogEditField = 0
						}
						e.kbDialogEditing = true
						e.statusbar.SetMessage("Press key for binding (Esc=cancel, Del=clear)", "info")
					} else {
						// Different item - just select it
						e.kbDialogIndex = clickedIdx
						if relX >= alternateStart {
							e.kbDialogEditField = 1
						} else if relX >= primaryStart {
							e.kbDialogEditField = 0
						}
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if e.kbDialogIndex > 0 {
			e.kbDialogIndex--
			e.ensureKbDialogVisible()
		}

	case tea.MouseButtonWheelDown:
		if e.kbDialogIndex < actionCount-1 {
			e.kbDialogIndex++
			e.ensureKbDialogVisible()
		}
	}

	return e, nil
}

// handleHelpMouse handles mouse input in help mode
func (e *Editor) handleHelpMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Help dialog dimensions (must match overlayHelpDialog)
	boxWidth := 72
	// Height: top border + empty + 21 rows + empty + 2 options + empty + footer + bottom
	boxHeight := 29

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

	// Any click inside the dialog also closes it
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		e.mode = ModeNormal
	}

	return e, nil
}

// handleAboutMouse handles mouse input in about mode
func (e *Editor) handleAboutMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// About dialog dimensions (must match overlayAboutDialog)
	boxWidth := 66
	// Height: top border + empty + 6 logo + 7 content + 2 quote + 2 footer + bottom
	boxHeight := 20

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

	// Any click inside the dialog also closes it
	if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
		e.mode = ModeNormal
	}

	return e, nil
}

// keyMsgToString converts a tea.KeyMsg to a keybinding string like "ctrl+s"
func (e *Editor) keyMsgToString(msg tea.KeyMsg) string {
	var parts []string

	// Add modifiers
	if msg.Alt {
		parts = append(parts, "alt")
	}

	// Check for ctrl from the key type
	switch msg.Type {
	case tea.KeyCtrlA:
		parts = append(parts, "ctrl")
		parts = append(parts, "a")
	case tea.KeyCtrlB:
		parts = append(parts, "ctrl")
		parts = append(parts, "b")
	case tea.KeyCtrlC:
		parts = append(parts, "ctrl")
		parts = append(parts, "c")
	case tea.KeyCtrlD:
		parts = append(parts, "ctrl")
		parts = append(parts, "d")
	case tea.KeyCtrlE:
		parts = append(parts, "ctrl")
		parts = append(parts, "e")
	case tea.KeyCtrlF:
		parts = append(parts, "ctrl")
		parts = append(parts, "f")
	case tea.KeyCtrlG:
		parts = append(parts, "ctrl")
		parts = append(parts, "g")
	case tea.KeyCtrlH:
		parts = append(parts, "ctrl")
		parts = append(parts, "h")
	// Note: tea.KeyCtrlI is same as tea.KeyTab (handled below)
	case tea.KeyCtrlJ:
		parts = append(parts, "ctrl")
		parts = append(parts, "j")
	case tea.KeyCtrlK:
		parts = append(parts, "ctrl")
		parts = append(parts, "k")
	case tea.KeyCtrlL:
		parts = append(parts, "ctrl")
		parts = append(parts, "l")
	case tea.KeyCtrlM:
		parts = append(parts, "ctrl")
		parts = append(parts, "m")
	case tea.KeyCtrlN:
		parts = append(parts, "ctrl")
		parts = append(parts, "n")
	case tea.KeyCtrlO:
		parts = append(parts, "ctrl")
		parts = append(parts, "o")
	case tea.KeyCtrlP:
		parts = append(parts, "ctrl")
		parts = append(parts, "p")
	case tea.KeyCtrlQ:
		parts = append(parts, "ctrl")
		parts = append(parts, "q")
	case tea.KeyCtrlR:
		parts = append(parts, "ctrl")
		parts = append(parts, "r")
	case tea.KeyCtrlS:
		parts = append(parts, "ctrl")
		parts = append(parts, "s")
	case tea.KeyCtrlT:
		parts = append(parts, "ctrl")
		parts = append(parts, "t")
	case tea.KeyCtrlU:
		parts = append(parts, "ctrl")
		parts = append(parts, "u")
	case tea.KeyCtrlV:
		parts = append(parts, "ctrl")
		parts = append(parts, "v")
	case tea.KeyCtrlW:
		parts = append(parts, "ctrl")
		parts = append(parts, "w")
	case tea.KeyCtrlX:
		parts = append(parts, "ctrl")
		parts = append(parts, "x")
	case tea.KeyCtrlY:
		parts = append(parts, "ctrl")
		parts = append(parts, "y")
	case tea.KeyCtrlZ:
		parts = append(parts, "ctrl")
		parts = append(parts, "z")
	case tea.KeyF1:
		parts = append(parts, "f1")
	case tea.KeyF2:
		parts = append(parts, "f2")
	case tea.KeyF3:
		parts = append(parts, "f3")
	case tea.KeyF4:
		parts = append(parts, "f4")
	case tea.KeyF5:
		parts = append(parts, "f5")
	case tea.KeyF6:
		parts = append(parts, "f6")
	case tea.KeyF7:
		parts = append(parts, "f7")
	case tea.KeyF8:
		parts = append(parts, "f8")
	case tea.KeyF9:
		parts = append(parts, "f9")
	case tea.KeyF10:
		parts = append(parts, "f10")
	case tea.KeyF11:
		parts = append(parts, "f11")
	case tea.KeyF12:
		parts = append(parts, "f12")
	case tea.KeyHome:
		parts = append(parts, "home")
	case tea.KeyEnd:
		parts = append(parts, "end")
	case tea.KeyLeft:
		parts = append(parts, "left")
	case tea.KeyRight:
		parts = append(parts, "right")
	case tea.KeyUp:
		parts = append(parts, "up")
	case tea.KeyDown:
		parts = append(parts, "down")
	case tea.KeyTab:
		parts = append(parts, "tab")
	case tea.KeyShiftTab:
		parts = append(parts, "shift")
		parts = append(parts, "tab")
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			r := msg.Runes[0]
			parts = append(parts, strings.ToLower(string(r)))
		}
	default:
		return "" // Unsupported key
	}

	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "+")
}

// reservedKeys are keys that cannot be remapped (menu shortcuts, etc.)
var reservedKeys = map[string]string{
	"alt+f":  "File menu",
	"alt+b":  "Buffers menu",
	"alt+e":  "Edit menu",
	"alt+s":  "Search menu",
	"alt+o":  "Options menu",
	"alt+h":  "Help menu",
	"alt+<":  "Previous buffer",
	"alt+>":  "Next buffer",
	"f10":    "Open menu",
	"escape": "Cancel/Close",
	"esc":    "Cancel/Close",
}

// checkKeyConflict checks if a key is already bound to another action or reserved
func (e *Editor) checkKeyConflict(keyStr, currentAction string) []string {
	var conflicts []string
	keyLower := strings.ToLower(keyStr)

	// Check reserved keys first
	if desc, reserved := reservedKeys[keyLower]; reserved {
		conflicts = append(conflicts, desc)
		return conflicts
	}

	for _, action := range config.AllActions() {
		if action == currentAction {
			continue
		}
		binding := e.keybindings.GetBinding(action)
		if strings.ToLower(binding.Primary) == keyLower ||
			strings.ToLower(binding.Alternate) == keyLower {
			conflicts = append(conflicts, action)
		}
	}
	return conflicts
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
	if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
		e.deleteSelection()
	}

	// Record for undo
	entry := &UndoEntry{
		Position:     e.activeDoc().cursor.ByteOffset(),
		Inserted:     string(r),
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
	}

	e.activeDoc().cursor.Sync()
	e.activeDoc().buffer.InsertRune(r)
	e.activeDoc().cursor.MoveRight()

	entry.CursorAfter = e.activeDoc().cursor.ByteOffset()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true
}

func (e *Editor) insertText(s string) {
	if s == "" {
		return
	}

	// Delete selection first if any
	if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
		e.deleteSelection()
	}

	entry := &UndoEntry{
		Position:     e.activeDoc().cursor.ByteOffset(),
		Inserted:     s,
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
	}

	e.activeDoc().cursor.Sync()
	e.activeDoc().buffer.Insert(s)
	e.activeDoc().cursor.SetByteOffset(e.activeDoc().cursor.ByteOffset() + len(s))

	entry.CursorAfter = e.activeDoc().cursor.ByteOffset()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true
}

// getIndentString returns the string to use for one level of indentation
func (e *Editor) getIndentString() string {
	if e.config.Editor.TabsToSpaces {
		width := e.config.Editor.TabWidth
		if width <= 0 {
			width = 4
		}
		return strings.Repeat(" ", width)
	}
	return "\t"
}

// indentLines indents all lines in the current selection
func (e *Editor) indentLines() {
	doc := e.activeDoc()
	sel := doc.selection

	// If no selection, just insert tab at cursor
	if !sel.Active || sel.IsEmpty() {
		e.insertText(e.getIndentString())
		return
	}

	// Get the line range of the selection
	startPos, endPos := sel.Normalize()
	startLine, _ := doc.buffer.PositionToLineCol(startPos)
	endLine, endCol := doc.buffer.PositionToLineCol(endPos)

	// If selection ends at column 0, don't include that line
	if endCol == 0 && endLine > startLine {
		endLine--
	}

	indent := e.getIndentString()

	// Build the undo entry for all changes
	var deleted strings.Builder
	var inserted strings.Builder

	// Calculate the range we're modifying
	rangeStart := doc.buffer.LineStartOffset(startLine)
	rangeEnd := doc.buffer.LineEndOffset(endLine)
	if rangeEnd < doc.buffer.Length() {
		// Include the newline if present
		if r, size := doc.buffer.RuneAt(rangeEnd); r == '\n' {
			rangeEnd += size
		}
	}

	// Get original text
	originalText := doc.buffer.Substring(rangeStart, rangeEnd)
	deleted.WriteString(originalText)

	// Build new text with indentation
	lines := strings.Split(originalText, "\n")
	for i, line := range lines {
		if i < len(lines)-1 || line != "" { // Don't indent empty trailing line
			inserted.WriteString(indent)
		}
		inserted.WriteString(line)
		if i < len(lines)-1 {
			inserted.WriteString("\n")
		}
	}

	// Create undo entry
	entry := &UndoEntry{
		Position:     rangeStart,
		Deleted:      deleted.String(),
		Inserted:     inserted.String(),
		CursorBefore: doc.cursor.ByteOffset(),
	}

	// Replace the text
	doc.cursor.Sync()
	doc.buffer.Replace(rangeStart, rangeEnd, inserted.String())

	// Update selection to cover the indented lines
	newEndPos := rangeStart + len(inserted.String())
	if newEndPos > 0 && inserted.String()[len(inserted.String())-1] == '\n' {
		newEndPos-- // Don't include trailing newline in selection
	}
	sel.Anchor = rangeStart
	sel.Cursor = newEndPos

	// Position cursor at end of selection
	doc.cursor.SetByteOffset(newEndPos)

	entry.CursorAfter = doc.cursor.ByteOffset()
	doc.undoStack.Push(entry)
	doc.modified = true
}

// dedentLines removes one level of indentation from all lines in the selection
func (e *Editor) dedentLines() {
	doc := e.activeDoc()
	sel := doc.selection

	// If no selection, dedent current line
	if !sel.Active || sel.IsEmpty() {
		line, _ := doc.buffer.PositionToLineCol(doc.cursor.ByteOffset())
		sel.Start(doc.buffer.LineStartOffset(line))
		sel.Update(doc.buffer.LineEndOffset(line))
	}

	// Get the line range of the selection
	startPos, endPos := sel.Normalize()
	startLine, _ := doc.buffer.PositionToLineCol(startPos)
	endLine, endCol := doc.buffer.PositionToLineCol(endPos)

	// If selection ends at column 0, don't include that line
	if endCol == 0 && endLine > startLine {
		endLine--
	}

	tabWidth := e.config.Editor.TabWidth
	if tabWidth <= 0 {
		tabWidth = 4
	}

	// Calculate the range we're modifying
	rangeStart := doc.buffer.LineStartOffset(startLine)
	rangeEnd := doc.buffer.LineEndOffset(endLine)
	if rangeEnd < doc.buffer.Length() {
		if r, size := doc.buffer.RuneAt(rangeEnd); r == '\n' {
			rangeEnd += size
		}
	}

	// Get original text
	originalText := doc.buffer.Substring(rangeStart, rangeEnd)

	// Build new text with dedentation
	var inserted strings.Builder
	lines := strings.Split(originalText, "\n")
	changed := false

	for i, line := range lines {
		newLine := line
		if len(line) > 0 {
			if line[0] == '\t' {
				// Remove one tab
				newLine = line[1:]
				changed = true
			} else if line[0] == ' ' {
				// Remove up to tabWidth spaces
				spacesToRemove := 0
				for j := 0; j < len(line) && j < tabWidth && line[j] == ' '; j++ {
					spacesToRemove++
				}
				if spacesToRemove > 0 {
					newLine = line[spacesToRemove:]
					changed = true
				}
			}
		}
		inserted.WriteString(newLine)
		if i < len(lines)-1 {
			inserted.WriteString("\n")
		}
	}

	if !changed {
		// Nothing to dedent
		sel.Clear()
		return
	}

	// Create undo entry
	entry := &UndoEntry{
		Position:     rangeStart,
		Deleted:      originalText,
		Inserted:     inserted.String(),
		CursorBefore: doc.cursor.ByteOffset(),
	}

	// Replace the text
	doc.cursor.Sync()
	doc.buffer.Replace(rangeStart, rangeEnd, inserted.String())

	// Update selection to cover the dedented lines
	newEndPos := rangeStart + len(inserted.String())
	if newEndPos > 0 && len(inserted.String()) > 0 && inserted.String()[len(inserted.String())-1] == '\n' {
		newEndPos--
	}
	sel.Anchor = rangeStart
	sel.Cursor = newEndPos

	// Position cursor at end of selection
	doc.cursor.SetByteOffset(newEndPos)

	entry.CursorAfter = doc.cursor.ByteOffset()
	doc.undoStack.Push(entry)
	doc.modified = true
}

func (e *Editor) backspace() {
	if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	if e.activeDoc().cursor.ByteOffset() == 0 {
		return
	}

	// Sync cursor position to buffer gap
	e.activeDoc().cursor.Sync()

	// Get info about what we're about to delete (the rune before cursor)
	pos := e.activeDoc().cursor.ByteOffset()
	deleted := e.activeDoc().buffer.DeleteRuneBefore()
	if deleted == "" {
		return
	}

	// Update cursor position to match new gap position
	newPos := e.activeDoc().buffer.CursorPosition()
	e.activeDoc().cursor.SetByteOffset(newPos)

	entry := &UndoEntry{
		Position:     newPos,
		Deleted:      deleted,
		CursorBefore: pos,
		CursorAfter:  newPos,
	}

	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true
}

func (e *Editor) delete() {
	if e.activeDoc().selection.Active && !e.activeDoc().selection.IsEmpty() {
		e.deleteSelection()
		return
	}

	if e.activeDoc().cursor.ByteOffset() >= e.activeDoc().buffer.Length() {
		return
	}

	_, size := e.activeDoc().buffer.RuneAt(e.activeDoc().cursor.ByteOffset())
	if size == 0 {
		return
	}

	entry := &UndoEntry{
		Position:     e.activeDoc().cursor.ByteOffset(),
		Deleted:      e.activeDoc().buffer.Substring(e.activeDoc().cursor.ByteOffset(), e.activeDoc().cursor.ByteOffset()+size),
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
		CursorAfter:  e.activeDoc().cursor.ByteOffset(),
	}

	e.activeDoc().cursor.Sync()
	e.activeDoc().buffer.DeleteAfter(size)
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true
}

func (e *Editor) deleteSelection() {
	if !e.activeDoc().selection.Active || e.activeDoc().selection.IsEmpty() {
		return
	}

	start, end := e.activeDoc().selection.Normalize()
	text := e.activeDoc().buffer.Substring(start, end)

	entry := &UndoEntry{
		Position:     start,
		Deleted:      text,
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
		CursorAfter:  start,
	}

	e.activeDoc().buffer.Replace(start, end, "")
	e.activeDoc().cursor.SetByteOffset(start)
	e.activeDoc().selection.Clear()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true
}

func (e *Editor) undo() {
	entry := e.activeDoc().undoStack.Undo()
	if entry == nil {
		return
	}

	// Reverse the operation
	if entry.Inserted != "" {
		// Was an insertion - delete it
		e.activeDoc().buffer.Replace(entry.Position, entry.Position+len(entry.Inserted), "")
	}
	if entry.Deleted != "" {
		// Was a deletion - insert it back
		e.activeDoc().buffer.MoveCursor(entry.Position)
		e.activeDoc().buffer.Insert(entry.Deleted)
	}

	e.activeDoc().cursor.SetByteOffset(entry.CursorBefore)
	e.activeDoc().selection.Clear()
	e.activeDoc().modified = true
}

func (e *Editor) redo() {
	entry := e.activeDoc().undoStack.Redo()
	if entry == nil {
		return
	}

	// Replay the operation
	if entry.Deleted != "" {
		// Was a deletion - delete it again
		e.activeDoc().buffer.Replace(entry.Position, entry.Position+len(entry.Deleted), "")
	}
	if entry.Inserted != "" {
		// Was an insertion - insert it again
		e.activeDoc().buffer.MoveCursor(entry.Position)
		e.activeDoc().buffer.Insert(entry.Inserted)
	}

	e.activeDoc().cursor.SetByteOffset(entry.CursorAfter)
	e.activeDoc().selection.Clear()
	e.activeDoc().modified = true
}

func (e *Editor) cut() {
	if !e.activeDoc().selection.Active || e.activeDoc().selection.IsEmpty() {
		return
	}

	text := e.activeDoc().selection.GetText(e.activeDoc().buffer)
	e.clipboard.Copy(text)
	e.deleteSelection()
}

// cutLine cuts the entire current line (like nano's Ctrl+K)
func (e *Editor) cutLine() {
	line := e.activeDoc().cursor.Line()
	lineStart := e.activeDoc().buffer.LineStartOffset(line)
	lineEnd := e.activeDoc().buffer.LineEndOffset(line)

	// Include the newline character if this isn't the last line
	if lineEnd < e.activeDoc().buffer.Length() {
		lineEnd++ // Include the \n
	}

	// If the line is empty and it's the only line, nothing to cut
	if lineStart == lineEnd {
		e.statusbar.SetMessage("Nothing to cut", "info")
		return
	}

	// Get the line content
	text := e.activeDoc().buffer.Substring(lineStart, lineEnd)

	// Copy to clipboard
	e.clipboard.Copy(text)

	// Record for undo
	entry := &UndoEntry{
		Position:     lineStart,
		Deleted:      text,
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
		CursorAfter:  lineStart,
	}

	// Delete the line
	e.activeDoc().buffer.Replace(lineStart, lineEnd, "")
	e.activeDoc().cursor.SetByteOffset(lineStart)
	e.activeDoc().selection.Clear()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true

	e.statusbar.SetMessage("Line cut", "info")
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
}

func (e *Editor) copy() {
	if !e.activeDoc().selection.Active || e.activeDoc().selection.IsEmpty() {
		return
	}

	text := e.activeDoc().selection.GetText(e.activeDoc().buffer)
	e.clipboard.Copy(text)
	e.statusbar.SetMessage("Copied", "info")
}

func (e *Editor) paste() {
	text, err := e.clipboard.Paste()
	if err != nil || text == "" {
		return
	}

	e.insertText(text)
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
}

func (e *Editor) selectAll() {
	e.activeDoc().selection.SelectAll(e.activeDoc().buffer)
	e.activeDoc().cursor.MoveToEnd()
}

func (e *Editor) newFile() {
	// Creates a new buffer - doesn't affect the current buffer
	e.doNewFile()
}

func (e *Editor) doNewFile() {
	// Check buffer limit
	maxBuffers := 20 // default
	if e.config != nil && e.config.Editor.MaxBuffers > 0 {
		maxBuffers = e.config.Editor.MaxBuffers
	}
	if maxBuffers > 0 && len(e.documents) >= maxBuffers {
		e.statusbar.SetMessage(fmt.Sprintf("Buffer limit reached (%d)", maxBuffers), "error")
		return
	}

	// Create a new document
	buf := NewBuffer()
	doc := &Document{
		buffer:      buf,
		cursor:      NewCursor(buf),
		selection:   NewSelection(),
		undoStack:   NewUndoStack(100),
		filename:    "",
		modified:    false,
		scrollY:     0,
		highlighter: syntax.New(""),
		encoding:    enc.GetEncodingByID("utf-8"), // Default to UTF-8
	}
	e.documents = append(e.documents, doc)
	e.activeIdx = len(e.documents) - 1

	e.updateTitle()
	e.updateMenuState()
	e.statusbar.SetMessage("New file", "info")
}

// closeFile closes the current file (same as new, but different messaging)
func (e *Editor) closeFile() {
	if e.activeDoc().modified {
		e.showPrompt("Unsaved changes. Discard? (y/N): ", PromptConfirmClose)
		return
	}
	e.doCloseFile()
}

func (e *Editor) doCloseFile() {
	if len(e.documents) > 1 {
		// Multiple buffers - remove current and switch to another
		e.documents = append(e.documents[:e.activeIdx], e.documents[e.activeIdx+1:]...)
		if e.activeIdx >= len(e.documents) {
			e.activeIdx = len(e.documents) - 1
		}
		// Restore scroll position of new active buffer
		e.viewport.SetScrollY(e.activeDoc().scrollY)
		e.statusbar.SetMessage("Buffer closed", "info")
	} else {
		// Single buffer - reset to empty
		e.activeDoc().buffer = NewBuffer()
		e.activeDoc().cursor = NewCursor(e.activeDoc().buffer)
		e.activeDoc().selection.Clear()
		e.activeDoc().undoStack.Clear()
		e.activeDoc().filename = ""
		e.activeDoc().modified = false
		e.activeDoc().scrollY = 0
		e.activeDoc().highlighter.SetFile("")
		e.activeDoc().encoding = enc.GetEncodingByID("utf-8")
		e.viewport.SetScrollY(0)
		e.statusbar.SetMessage("File closed", "info")
	}
	e.updateTitle()
	e.updateMenuState()
}

// quitEditor exits the editor, checking for unsaved changes in ALL buffers
func (e *Editor) quitEditor() tea.Cmd {
	// Check all buffers for unsaved changes
	unsavedCount := 0
	for _, doc := range e.documents {
		if doc.modified {
			unsavedCount++
		}
	}
	if unsavedCount > 0 {
		msg := "Unsaved changes. Quit anyway? (y/N): "
		if unsavedCount > 1 {
			msg = fmt.Sprintf("%d unsaved buffers. Quit anyway? (y/N): ", unsavedCount)
		}
		e.showPrompt(msg, PromptConfirmQuit)
		return nil
	}
	return tea.Quit
}

// updateTitle sets the terminal title
func (e *Editor) updateTitle() {
	e.pendingTitle = "textivus"
	if e.activeDoc().filename != "" {
		e.pendingTitle += " - " + e.activeDoc().filename
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
	e.menubar.SetItemDisabled(ui.ActionRevert, e.activeDoc().filename == "")

	// Update buffers menu
	var names []string
	for _, doc := range e.documents {
		name := "[Untitled]"
		if doc.filename != "" {
			name = filepath.Base(doc.filename)
		}
		if doc.modified {
			name = "*" + name
		}
		names = append(names, name)
	}
	e.menubar.SetBuffers(names, e.activeIdx)
}

// openFile prompts for a filename to open
func (e *Editor) openFile() {
	if e.activeDoc().modified {
		e.showPrompt("Unsaved changes. Discard? (y/N): ", PromptConfirmOpen)
		return
	}
	e.showFileBrowser()
}

// revertFile reloads the file from disk
func (e *Editor) revertFile() {
	if e.activeDoc().filename == "" {
		e.statusbar.SetMessage("No file to revert", "error")
		return
	}
	if err := e.LoadFile(e.activeDoc().filename); err != nil {
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
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
	e.statusbar.SetMessage("Inserted lorem ipsum", "info")
}

func (e *Editor) findNext() {
	if e.findQuery == "" {
		return
	}

	content := e.activeDoc().buffer.String()
	startPos := e.activeDoc().cursor.ByteOffset() + 1
	if startPos >= len(content) {
		startPos = 0
	}

	// Search from cursor position
	idx := strings.Index(content[startPos:], e.findQuery)
	if idx >= 0 {
		pos := startPos + idx
		e.activeDoc().cursor.SetByteOffset(pos)
		e.activeDoc().selection.Active = true
		e.activeDoc().selection.Anchor = pos
		e.activeDoc().selection.Cursor = pos + len(e.findQuery)
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
		return
	}

	// Wrap around
	idx = strings.Index(content[:startPos], e.findQuery)
	if idx >= 0 {
		e.activeDoc().cursor.SetByteOffset(idx)
		e.activeDoc().selection.Active = true
		e.activeDoc().selection.Anchor = idx
		e.activeDoc().selection.Cursor = idx + len(e.findQuery)
		e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
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

	content := e.activeDoc().buffer.String()
	startPos := e.activeDoc().cursor.ByteOffset()

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
		CursorBefore: e.activeDoc().cursor.ByteOffset(),
		CursorAfter:  idx + len(e.replaceQuery),
	}

	// Perform the replacement
	e.activeDoc().buffer.Replace(idx, idx+len(e.findQuery), e.replaceQuery)
	e.activeDoc().cursor.SetByteOffset(idx + len(e.replaceQuery))
	e.activeDoc().selection.Clear()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true

	e.statusbar.SetMessage("Replaced", "info")
	e.viewport.EnsureCursorVisibleWrapped(e.activeDoc().buffer.Lines(), e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
}

// replaceAll replaces all occurrences with a single undo entry
func (e *Editor) replaceAll() {
	if e.findQuery == "" {
		e.statusbar.SetMessage("No search term", "error")
		return
	}

	content := e.activeDoc().buffer.String()
	count := strings.Count(content, e.findQuery)
	if count == 0 {
		e.statusbar.SetMessage("Not found", "error")
		return
	}

	// Store original content for undo
	originalContent := content
	cursorBefore := e.activeDoc().cursor.ByteOffset()

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
	e.activeDoc().buffer = NewBufferFromString(newContent)
	e.activeDoc().cursor = NewCursor(e.activeDoc().buffer)
	e.activeDoc().selection.Clear()
	e.activeDoc().undoStack.Push(entry)
	e.activeDoc().modified = true

	e.statusbar.SetMessage(fmt.Sprintf("Replaced %d occurrences", count), "info")
}

// View implements tea.Model
func (e *Editor) View() string {
	var sb strings.Builder

	// Set terminal title using OSC escape sequence
	if e.pendingTitle != "" {
		sb.WriteString(fmt.Sprintf("\033]0;%s\007", e.pendingTitle))
	}

	// Output any pending escape sequences (e.g., Kitty graphics cleanup)
	if e.pendingEscapes != "" {
		sb.WriteString(e.pendingEscapes)
		e.pendingEscapes = ""
	}

	// Menu bar
	sb.WriteString(e.menubar.View())
	sb.WriteString("\n")

	// Render editor content using compositor
	renderState := e.buildRenderState()
	viewportContent := e.compositor.Render(renderState)

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

	// If recent files dialog is open, overlay it centered on the viewport
	if e.mode == ModeRecentFiles {
		viewportContent = e.overlayRecentFilesDialog(viewportContent)
	}

	// If recent directories dialog is open, overlay it centered on the viewport
	if e.mode == ModeRecentDirs {
		viewportContent = e.overlayRecentDirsDialog(viewportContent)
	}

	// If keybindings dialog is open, overlay it centered on the viewport
	if e.mode == ModeKeybindings {
		viewportContent = e.overlayKeybindingsDialog(viewportContent)
	}

	// If config error dialog is open, overlay it centered on the viewport
	if e.mode == ModeConfigError {
		viewportContent = e.overlayConfigErrorDialog(viewportContent)
	}

	// If settings dialog is open, overlay it centered on the viewport
	if e.mode == ModeSettings {
		viewportContent = e.overlaySettingsDialog(viewportContent)
	}

	// If encoding dialog is open, overlay it centered on the viewport
	if e.mode == ModeEncoding {
		viewportContent = e.overlayEncodingDialog(viewportContent)
	}

	sb.WriteString(viewportContent)
	sb.WriteString("\n")

	// Get theme colors for input bars
	barColor := ui.ColorToANSI(e.styles.Theme.UI.MenuFg, e.styles.Theme.UI.MenuBg)

	// Find bar if active
	if e.mode == ModeFind {
		findContent := "Find: " + e.findQuery
		cursor := "▂" // Lower quarter block cursor
		padding := e.width - len(findContent) - 1
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(findContent)
		sb.WriteString(cursor)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Find/Replace bar if active (two lines)
	if e.mode == ModeFindReplace {
		cursor := "▂" // Lower quarter block cursor

		// Line 1: Find field
		findLine := "Find: " + e.findQuery
		findCursorStr := ""
		if !e.replaceFocus {
			findCursorStr = cursor
		}
		findPadding := e.width - len(findLine) - 1
		if findPadding < 0 {
			findPadding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(findLine)
		sb.WriteString(findCursorStr)
		if e.replaceFocus {
			sb.WriteString(" ") // Space where cursor would be
		}
		sb.WriteString(strings.Repeat(" ", findPadding))
		sb.WriteString("\033[0m\n")

		// Line 2: Replace field with hints
		replaceLine := "Replace: " + e.replaceQuery
		replaceCursorStr := ""
		if e.replaceFocus {
			replaceCursorStr = cursor
		}
		hints := " [Tab] Switch [Enter] Replace [Ctrl+A] All"
		availSpace := e.width - len(replaceLine) - 1 - len(hints)
		if availSpace < 0 {
			availSpace = 0
			hints = ""
		}
		sb.WriteString(barColor)
		sb.WriteString(replaceLine)
		sb.WriteString(replaceCursorStr)
		if !e.replaceFocus {
			sb.WriteString(" ") // Space where cursor would be
		}
		sb.WriteString(strings.Repeat(" ", availSpace))
		sb.WriteString(hints)
		sb.WriteString("\033[0m\n")
	}

	// Prompt bar if active
	if e.mode == ModePrompt {
		promptContent := e.promptText + e.promptInput
		cursor := "▂" // Lower quarter block cursor
		padding := e.width - len(promptContent) - 1
		if padding < 0 {
			padding = 0
		}
		sb.WriteString(barColor)
		sb.WriteString(promptContent)
		sb.WriteString(cursor)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Status bar
	e.statusbar.SetPosition(e.activeDoc().cursor.Line(), e.activeDoc().cursor.Col())
	e.statusbar.SetFilename(e.activeDoc().filename)
	e.statusbar.SetModified(e.activeDoc().modified)
	e.statusbar.SetTotalLines(e.activeDoc().buffer.LineCount())
	e.statusbar.SetCounts(e.activeDoc().buffer.WordCount(), e.activeDoc().buffer.RuneCount())
	e.statusbar.SetBufferInfo(e.activeIdx, len(e.documents))
	// Set encoding display
	docEnc := e.activeDoc().encoding
	if docEnc != nil {
		e.statusbar.SetEncoding(docEnc.Name, docEnc.Supported)
	} else {
		e.statusbar.SetEncoding("UTF-8", true)
	}
	sb.WriteString(e.statusbar.View())

	// Append Kitty graphics minimap if enabled (rendered as overlay with cursor positioning)
	if e.minimapRenderer.IsEnabled() {
		// Calculate minimap position
		// X offset: width - scrollbar (if enabled) - minimap width
		xOffset := e.width - ui.MinimapWidth()
		if e.scrollbar.IsEnabled() {
			xOffset -= e.scrollbar.Width()
		}
		// Y offset: 1 for menu bar (viewport starts at row 2, which is index 1)
		yOffset := 1
		kittySeq := e.minimapRenderer.GetKittySequence(ui.MinimapWidth(), e.viewport.Height(), xOffset, yOffset, renderState)
		sb.WriteString(kittySeq)
	}

	return sb.String()
}

// SetFilename sets the filename for the editor
func (e *Editor) SetFilename(filename string) {
	// Convert to absolute path for consistent directory navigation
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename // Fall back to original if Abs fails
	}
	e.activeDoc().filename = absPath
	e.activeDoc().highlighter.SetFile(absPath) // Update syntax highlighter
}

// SetConfigError sets the config error state and shows the error dialog
func (e *Editor) SetConfigError(filePath, errMsg string) {
	e.configErrorFile = filePath
	e.configErrorMsg = errMsg
	e.configErrorChoice = 1 // Default to "Use Defaults"
	e.mode = ModeConfigError
}
