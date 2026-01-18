package editor

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"festivus/clipboard"
	"festivus/config"
	"festivus/ui"

	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the editor mode
type Mode int

const (
	ModeNormal Mode = iota
	ModeMenu
	ModeFind
	ModePrompt
	ModeAbout
	ModeFileBrowser
	ModeSaveAs
)

// FileEntry represents a file or directory in the file browser
type FileEntry struct {
	Name  string
	IsDir bool
	Size  int64
}

// PromptAction represents what to do with the prompt result
type PromptAction int

const (
	PromptNone PromptAction = iota
	PromptSaveAs
	PromptOpen
	PromptConfirmNew
	PromptConfirmOverwrite
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

	// State
	mode     Mode
	filename string
	modified bool
	width    int
	height   int

	// Find mode state
	findQuery  string
	findActive bool

	// Prompt mode state
	promptText      string       // The prompt message
	promptInput     string       // User's input
	promptAction    PromptAction // What to do with the result
	pendingFilename string       // Filename pending confirmation (for overwrite)

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

	// Save As state
	saveAsFilename string // Filename input for Save As dialog

	// Configuration
	config *config.Config
}

// New creates a new editor instance with default config
func New() *Editor {
	return NewWithConfig(config.DefaultConfig())
}

// NewWithConfig creates a new editor instance with the given configuration
func NewWithConfig(cfg *config.Config) *Editor {
	styles := ui.DefaultStyles()
	buf := NewBuffer()

	e := &Editor{
		buffer:    buf,
		cursor:    NewCursor(buf),
		selection: NewSelection(),
		undoStack: NewUndoStack(1000),
		clipboard: clipboard.New(os.Stdout),
		menubar:   ui.NewMenuBar(styles),
		statusbar: ui.NewStatusBar(styles),
		viewport:  ui.NewViewport(styles),
		styles:    styles,
		mode:      ModeNormal,
		width:     80,
		height:    24,
		config:    cfg,
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
	}

	return e
}

// LoadFile loads a file into the editor
func (e *Editor) LoadFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	e.buffer = NewBufferFromString(string(content))
	e.cursor = NewCursor(e.buffer)
	e.selection.Clear()
	e.undoStack.Clear()
	e.filename = filename
	e.modified = false
	e.updateTitle()
	e.updateMenuState()

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
	content := e.buffer.String()
	err := os.WriteFile(e.filename, []byte(content), 0644)
	if err != nil {
		e.statusbar.SetMessage("Error: "+err.Error(), "error")
		return false
	}

	e.modified = false
	e.statusbar.SetMessage("Saved: "+e.filename, "success")
	e.updateTitle()
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

	// Handle prompt mode
	if e.mode == ModePrompt {
		return e.handlePromptKey(msg)
	}

	// Handle about mode - any key dismisses
	if e.mode == ModeAbout {
		e.mode = ModeNormal
		return e, nil
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
		return e, tea.Quit

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
		e.toggleWordWrap()
		return e, nil

	case tea.KeyCtrlL:
		e.toggleLineNumbers()
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
				e.menubar.OpenMenu(0)
				e.updateViewportSize()
				return e, nil
			case 'e', 'E':
				e.mode = ModeMenu
				e.menubar.OpenMenu(1)
				e.updateViewportSize()
				return e, nil
			case 'v', 'V':
				e.mode = ModeMenu
				e.menubar.OpenMenu(2)
				e.updateViewportSize()
				return e, nil
			case 'h', 'H':
				e.mode = ModeMenu
				e.menubar.OpenMenu(3)
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
	case "alt+v":
		e.mode = ModeMenu
		e.menubar.OpenMenu(2) // View
		e.updateViewportSize()
		return e, nil
	case "alt+h":
		e.mode = ModeMenu
		e.menubar.OpenMenu(3) // Help
		e.updateViewportSize()
		return e, nil
	case "f10":
		e.mode = ModeMenu
		e.menubar.OpenMenu(0)
		e.updateViewportSize()
		return e, nil
	case "f1":
		e.showAbout()
		return e, nil
	case "f2":
		e.insertLoremIpsum()
		return e, nil

	// Fallback ctrl key handling (string-based)
	case "ctrl+s":
		e.SaveFile()
		return e, nil
	case "ctrl+q":
		return e, tea.Quit
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
		e.toggleWordWrap()
		return e, nil
	case "ctrl+l":
		e.toggleLineNumbers()
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

// handleFileBrowserKey handles keyboard input in file browser mode
func (e *Editor) handleFileBrowserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Calculate visible height for the file list (box height minus header/footer)
	visibleHeight := e.fileBrowserVisibleHeight()

	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.statusbar.SetMessage("Cancelled", "info")

	case tea.KeyEnter:
		if len(e.fileBrowserEntries) > 0 && e.fileBrowserSelected < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[e.fileBrowserSelected]
			if entry.IsDir {
				// Enter directory
				var newPath string
				if entry.Name == ".." {
					newPath = filepath.Dir(e.fileBrowserDir)
				} else {
					newPath = filepath.Join(e.fileBrowserDir, entry.Name)
				}
				e.loadDirectory(newPath)
			} else {
				// Open file
				fullPath := filepath.Join(e.fileBrowserDir, entry.Name)
				e.mode = ModeNormal
				if err := e.LoadFile(fullPath); err != nil {
					e.statusbar.SetMessage("Error: "+err.Error(), "error")
				} else {
					e.statusbar.SetMessage("Opened: "+fullPath, "success")
				}
			}
		}

	case tea.KeyBackspace:
		// Go to parent directory
		if e.fileBrowserDir != "/" {
			newPath := filepath.Dir(e.fileBrowserDir)
			e.loadDirectory(newPath)
		}

	case tea.KeyUp:
		if e.fileBrowserSelected > 0 {
			e.fileBrowserSelected--
			// Scroll up if needed
			if e.fileBrowserSelected < e.fileBrowserScroll {
				e.fileBrowserScroll = e.fileBrowserSelected
			}
		}

	case tea.KeyDown:
		if e.fileBrowserSelected < len(e.fileBrowserEntries)-1 {
			e.fileBrowserSelected++
			// Scroll down if needed
			if e.fileBrowserSelected >= e.fileBrowserScroll+visibleHeight {
				e.fileBrowserScroll = e.fileBrowserSelected - visibleHeight + 1
			}
		}

	case tea.KeyHome:
		e.fileBrowserSelected = 0
		e.fileBrowserScroll = 0

	case tea.KeyEnd:
		e.fileBrowserSelected = len(e.fileBrowserEntries) - 1
		if e.fileBrowserSelected < 0 {
			e.fileBrowserSelected = 0
		}
		// Adjust scroll to show last item
		if e.fileBrowserSelected >= visibleHeight {
			e.fileBrowserScroll = e.fileBrowserSelected - visibleHeight + 1
		}

	case tea.KeyPgUp:
		e.fileBrowserSelected -= visibleHeight
		if e.fileBrowserSelected < 0 {
			e.fileBrowserSelected = 0
		}
		e.fileBrowserScroll -= visibleHeight
		if e.fileBrowserScroll < 0 {
			e.fileBrowserScroll = 0
		}

	case tea.KeyPgDown:
		e.fileBrowserSelected += visibleHeight
		if e.fileBrowserSelected >= len(e.fileBrowserEntries) {
			e.fileBrowserSelected = len(e.fileBrowserEntries) - 1
		}
		if e.fileBrowserSelected < 0 {
			e.fileBrowserSelected = 0
		}
		e.fileBrowserScroll += visibleHeight
		maxScroll := len(e.fileBrowserEntries) - visibleHeight
		if maxScroll < 0 {
			maxScroll = 0
		}
		if e.fileBrowserScroll > maxScroll {
			e.fileBrowserScroll = maxScroll
		}
	}

	return e, nil
}

// fileBrowserVisibleHeight returns the number of visible file entries in the browser
func (e *Editor) fileBrowserVisibleHeight() int {
	// Box height is based on viewport, minus borders and header/footer
	boxHeight := e.viewport.Height() - 4 // Reserve some margin
	if boxHeight > 20 {
		boxHeight = 20 // Cap at reasonable size
	}
	if boxHeight < 5 {
		boxHeight = 5
	}
	// Subtract header (title + directory + separator) and footer (separator + help)
	return boxHeight - 5
}

// saveAsVisibleHeight returns the number of visible directory entries in Save As
func (e *Editor) saveAsVisibleHeight() int {
	boxHeight := e.viewport.Height() - 4
	if boxHeight > 16 {
		boxHeight = 16 // Smaller than file browser
	}
	if boxHeight < 5 {
		boxHeight = 5
	}
	// Subtract header (title + directory + filename input + separator) and footer
	return boxHeight - 6
}

// handleSaveAsKey handles keyboard input in Save As mode
func (e *Editor) handleSaveAsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := e.saveAsVisibleHeight()

	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.statusbar.SetMessage("Cancelled", "info")

	case tea.KeyEnter:
		// Save the file
		if e.saveAsFilename == "" {
			e.statusbar.SetMessage("Please enter a filename", "error")
			return e, nil
		}
		fullPath := filepath.Join(e.fileBrowserDir, e.saveAsFilename)
		// Check if file exists
		if _, err := os.Stat(fullPath); err == nil {
			// File exists - use prompt for confirmation
			e.pendingFilename = fullPath
			e.promptText = "File exists. Overwrite? (y/n): "
			e.promptInput = ""
			e.promptAction = PromptConfirmOverwrite
			e.mode = ModePrompt
			return e, nil
		}
		// Save the file
		e.filename = fullPath
		e.mode = ModeNormal
		if e.doSave() {
			e.updateTitle()
		}

	case tea.KeyBackspace:
		// If filename has text, delete from filename
		// Otherwise, go to parent directory
		if len(e.saveAsFilename) > 0 {
			e.saveAsFilename = e.saveAsFilename[:len(e.saveAsFilename)-1]
		} else if e.fileBrowserDir != "/" {
			newPath := filepath.Dir(e.fileBrowserDir)
			e.loadDirectoryForSaveAs(newPath)
		}

	case tea.KeyTab:
		// Toggle between filename input and directory list
		if e.fileBrowserSelected < 0 {
			e.fileBrowserSelected = 0
		} else {
			e.fileBrowserSelected = -1
		}

	case tea.KeyUp:
		if e.fileBrowserSelected >= 0 {
			if e.fileBrowserSelected > 0 {
				e.fileBrowserSelected--
				if e.fileBrowserSelected < e.fileBrowserScroll {
					e.fileBrowserScroll = e.fileBrowserSelected
				}
			}
		}

	case tea.KeyDown:
		if e.fileBrowserSelected >= 0 {
			if e.fileBrowserSelected < len(e.fileBrowserEntries)-1 {
				e.fileBrowserSelected++
				if e.fileBrowserSelected >= e.fileBrowserScroll+visibleHeight {
					e.fileBrowserScroll = e.fileBrowserSelected - visibleHeight + 1
				}
			}
		} else {
			// Start navigating directories
			e.fileBrowserSelected = 0
		}

	case tea.KeyRight:
		// Enter selected directory
		if e.fileBrowserSelected >= 0 && e.fileBrowserSelected < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[e.fileBrowserSelected]
			if entry.IsDir {
				var newPath string
				if entry.Name == ".." {
					newPath = filepath.Dir(e.fileBrowserDir)
				} else {
					newPath = filepath.Join(e.fileBrowserDir, entry.Name)
				}
				e.loadDirectoryForSaveAs(newPath)
				e.fileBrowserSelected = 0
			}
		}

	case tea.KeyLeft:
		// Go to parent directory
		if e.fileBrowserDir != "/" {
			newPath := filepath.Dir(e.fileBrowserDir)
			e.loadDirectoryForSaveAs(newPath)
			e.fileBrowserSelected = 0
		}

	case tea.KeyRunes:
		// Type into filename
		e.saveAsFilename += string(msg.Runes)
		e.fileBrowserSelected = -1 // Focus back on filename

	case tea.KeySpace:
		e.saveAsFilename += " "
		e.fileBrowserSelected = -1
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
	case ui.ActionSave:
		e.SaveFile()
	case ui.ActionSaveAs:
		e.showSaveAs()
	case ui.ActionRevert:
		e.revertFile()
	case ui.ActionExit:
		return e, tea.Quit
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
	case ui.ActionSelectAll:
		e.selectAll()
	case ui.ActionFind:
		e.mode = ModeFind
		e.findQuery = ""
		e.findActive = true
		e.updateViewportSize()
	case ui.ActionWordWrap:
		e.toggleWordWrap()
	case ui.ActionLineNumbers:
		e.toggleLineNumbers()
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

// saveConfig saves the current settings to the config file
func (e *Editor) saveConfig() {
	if e.config == nil {
		e.config = config.DefaultConfig()
	}
	e.config.Editor.WordWrap = e.viewport.WordWrap()
	e.config.Editor.LineNumbers = e.viewport.ShowLineNum()
	// Save in background - don't block the UI
	go e.config.Save()
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
	e.updateTitle()
	e.updateMenuState()
	e.statusbar.SetMessage("New file", "info")
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
		// TODO: Could add a two-step prompt here
		e.statusbar.SetMessage("Save changes first (Ctrl+S)", "error")
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

// showFileBrowser initializes and displays the file browser dialog
func (e *Editor) showFileBrowser() {
	// Start in current working directory, or home directory as fallback
	startDir, err := os.Getwd()
	if err != nil {
		startDir, err = os.UserHomeDir()
		if err != nil {
			startDir = "/"
		}
	}

	e.fileBrowserDir = startDir
	e.fileBrowserSelected = 0
	e.fileBrowserScroll = 0
	e.loadDirectory(startDir)
	e.mode = ModeFileBrowser
}

// showSaveAs initializes and displays the Save As dialog
func (e *Editor) showSaveAs() {
	// Start in current file's directory, or current working directory
	startDir := ""
	if e.filename != "" {
		startDir = filepath.Dir(e.filename)
		e.saveAsFilename = filepath.Base(e.filename)
	} else {
		var err error
		startDir, err = os.Getwd()
		if err != nil {
			startDir, err = os.UserHomeDir()
			if err != nil {
				startDir = "/"
			}
		}
		e.saveAsFilename = ""
	}

	e.fileBrowserDir = startDir
	e.fileBrowserSelected = -1 // No selection initially (focus on filename input)
	e.fileBrowserScroll = 0
	e.loadDirectoryForSaveAs(startDir)
	e.mode = ModeSaveAs
}

// loadDirectoryForSaveAs reads the contents of a directory for the Save As dialog
// Only shows directories (not files) since user types the filename
func (e *Editor) loadDirectoryForSaveAs(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		e.statusbar.SetMessage("Error reading directory: "+err.Error(), "error")
		return
	}

	e.fileBrowserEntries = make([]FileEntry, 0)

	// Add parent directory entry if not at root
	if path != "/" {
		e.fileBrowserEntries = append(e.fileBrowserEntries, FileEntry{
			Name:  "..",
			IsDir: true,
			Size:  0,
		})
	}

	// Only include directories
	var dirs []FileEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, FileEntry{
				Name:  entry.Name(),
				IsDir: true,
				Size:  0,
			})
		}
	}

	// Sort directories alphabetically (case-insensitive)
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})

	e.fileBrowserEntries = append(e.fileBrowserEntries, dirs...)
	e.fileBrowserDir = path
	e.fileBrowserScroll = 0
}

// loadDirectory reads the contents of a directory and populates the file browser
func (e *Editor) loadDirectory(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		e.statusbar.SetMessage("Error reading directory: "+err.Error(), "error")
		return
	}

	e.fileBrowserEntries = make([]FileEntry, 0, len(entries)+1)

	// Add parent directory entry if not at root
	if path != "/" {
		e.fileBrowserEntries = append(e.fileBrowserEntries, FileEntry{
			Name:  "..",
			IsDir: true,
			Size:  0,
		})
	}

	// Convert to FileEntry and separate dirs from files
	var dirs, files []FileEntry
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		fe := FileEntry{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
		}
		if entry.IsDir() {
			dirs = append(dirs, fe)
		} else {
			files = append(files, fe)
		}
	}

	// Sort directories and files alphabetically (case-insensitive)
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].Name) < strings.ToLower(dirs[j].Name)
	})
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].Name) < strings.ToLower(files[j].Name)
	})

	// Add directories first, then files
	e.fileBrowserEntries = append(e.fileBrowserEntries, dirs...)
	e.fileBrowserEntries = append(e.fileBrowserEntries, files...)

	e.fileBrowserDir = path
	e.fileBrowserSelected = 0
	e.fileBrowserScroll = 0
}

// formatFileSize formats a file size in human-readable format
func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case size >= GB:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%d B", size)
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
	viewportContent := e.viewport.Render(lines, e.cursor.Line(), e.cursor.Col(), selectionMap)

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

	sb.WriteString(viewportContent)
	sb.WriteString("\n")

	// Find bar if active
	if e.mode == ModeFind {
		findContent := "Find: " + e.findQuery + "_"
		padding := e.width - len(findContent)
		if padding < 0 {
			padding = 0
		}
		sb.WriteString("\033[44;97m") // Dark blue bg, white text
		sb.WriteString(findContent)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Prompt bar if active
	if e.mode == ModePrompt {
		promptContent := e.promptText + e.promptInput + "_"
		padding := e.width - len(promptContent)
		if padding < 0 {
			padding = 0
		}
		sb.WriteString("\033[44;97m") // Dark blue bg, white text
		sb.WriteString(promptContent)
		sb.WriteString(strings.Repeat(" ", padding))
		sb.WriteString("\033[0m\n")
	}

	// Status bar
	e.statusbar.SetPosition(e.cursor.Line(), e.cursor.Col())
	e.statusbar.SetFilename(e.filename)
	e.statusbar.SetModified(e.modified)
	e.statusbar.SetTotalLines(e.buffer.LineCount())
	sb.WriteString(e.statusbar.View())

	return sb.String()
}

// SetFilename sets the filename for the editor
func (e *Editor) SetFilename(filename string) {
	e.filename = filename
}

// overlayLineAt overlays the dropdown line on top of the viewport line at the given offset,
// preserving viewport content on both sides of the dropdown
func overlayLineAt(dropLine, viewportLine string, offset int) string {
	// Calculate the visual width of the dropdown line (strip ANSI codes)
	dropWidth := visualWidth(dropLine)

	// Get the viewport content as runes (stripped of ANSI for positioning)
	vpRunes := []rune(stripAnsi(viewportLine))

	// Build the result: prefix + dropdown + suffix
	var result strings.Builder

	// Prefix: viewport content before the dropdown (or spaces if line is short)
	if offset > 0 {
		if len(vpRunes) >= offset {
			// Use viewport content as prefix
			result.WriteString(string(vpRunes[:offset]))
		} else {
			// Viewport line is shorter than offset - use what we have plus padding
			result.WriteString(string(vpRunes))
			result.WriteString(strings.Repeat(" ", offset-len(vpRunes)))
		}
	}

	// The dropdown itself
	result.WriteString(dropLine)

	// Suffix: viewport content after the dropdown
	suffixStart := offset + dropWidth
	if suffixStart < len(vpRunes) {
		result.WriteString(string(vpRunes[suffixStart:]))
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
	return len([]rune(stripAnsi(s)))
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

	aboutLines := []string{
		strings.Repeat(" ", boxWidth),
		" ███████╗███████╗███████╗████████╗██╗██╗   ██╗██╗   ██╗███████╗ ",
		" ██╔════╝██╔════╝██╔════╝╚══██╔══╝██║██║   ██║██║   ██║██╔════╝ ",
		" █████╗  █████╗  ███████╗   ██║   ██║██║   ██║██║   ██║███████╗ ",
		" ██╔══╝  ██╔══╝  ╚════██║   ██║   ██║╚██╗ ██╔╝██║   ██║╚════██║ ",
		" ██║     ███████╗███████║   ██║   ██║ ╚████╔╝ ╚██████╔╝███████║ ",
		" ╚═╝     ╚══════╝╚══════╝   ╚═╝   ╚═╝  ╚═══╝   ╚═════╝ ╚══════╝ ",
		strings.Repeat(" ", boxWidth),
		centerText("A Text Editor for the Rest of Us"),
		strings.Repeat(" ", boxWidth),
		centerText("Version 0.1.0"),
		centerText("github.com/cornish/festivus"),
		centerText("Copyright (c) 2025"),
		strings.Repeat(" ", boxWidth),
	}
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

	for i, aboutLine := range aboutLines {
		viewportY := startY + i
		if viewportY >= 0 && viewportY < len(viewportLines) {
			// Build the styled about line with cyan background
			var styledLine strings.Builder
			styledLine.WriteString("\033[46;30m") // Cyan bg, black text
			styledLine.WriteString(aboutLine)
			styledLine.WriteString("\033[0m")

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// overlayFileBrowser overlays the file browser dialog centered on the viewport
func (e *Editor) overlayFileBrowser(viewportContent string) string {
	// Box dimensions
	boxWidth := 52
	visibleHeight := e.fileBrowserVisibleHeight()
	boxHeight := visibleHeight + 5 // +5 for header (3) and footer (2)

	// Helper to pad/truncate text to box width (minus borders)
	innerWidth := boxWidth - 2 // Account for left and right borders
	padText := func(s string, width int) string {
		if len(s) > width {
			return s[:width]
		}
		return s + strings.Repeat(" ", width-len(s))
	}
	centerText := func(s string, width int) string {
		if len(s) >= width {
			return s[:width]
		}
		padLeft := (width - len(s)) / 2
		padRight := width - len(s) - padLeft
		return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
	}

	// Truncate directory path if too long
	dirDisplay := e.fileBrowserDir
	maxDirLen := innerWidth - 12 // "Directory: " prefix
	if len(dirDisplay) > maxDirLen {
		dirDisplay = "..." + dirDisplay[len(dirDisplay)-maxDirLen+3:]
	}

	// Build the dialog lines
	var dialogLines []string

	// Top border with title
	title := " Open File "
	titlePadLeft := (boxWidth - 2 - len(title)) / 2
	titlePadRight := boxWidth - 2 - len(title) - titlePadLeft
	dialogLines = append(dialogLines, "┌"+strings.Repeat("─", titlePadLeft)+title+strings.Repeat("─", titlePadRight)+"┐")

	// Directory line
	dialogLines = append(dialogLines, "│"+padText(" Directory: "+dirDisplay, innerWidth)+"│")

	// Separator
	dialogLines = append(dialogLines, "├"+strings.Repeat("─", innerWidth)+"┤")

	// File list
	for i := 0; i < visibleHeight; i++ {
		idx := e.fileBrowserScroll + i
		if idx < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[idx]
			var line string
			if entry.IsDir {
				line = fmt.Sprintf(" %-36s %8s ", entry.Name, "<DIR>")
			} else {
				line = fmt.Sprintf(" %-36s %8s ", entry.Name, formatFileSize(entry.Size))
			}
			// Truncate if needed
			if len(line) > innerWidth {
				line = line[:innerWidth]
			} else {
				line = padText(line, innerWidth)
			}
			// Highlight selected item
			if idx == e.fileBrowserSelected {
				// Selected: white on blue
				line = "\033[44;97m" + line + "\033[46;30m"
			}
			dialogLines = append(dialogLines, "│"+line+"│")
		} else {
			// Empty line
			dialogLines = append(dialogLines, "│"+strings.Repeat(" ", innerWidth)+"│")
		}
	}

	// Separator
	dialogLines = append(dialogLines, "├"+strings.Repeat("─", innerWidth)+"┤")

	// Help line
	helpText := "[Enter] Open  [Esc] Cancel  [Bksp] Parent"
	dialogLines = append(dialogLines, "│"+centerText(helpText, innerWidth)+"│")

	// Bottom border
	dialogLines = append(dialogLines, "└"+strings.Repeat("─", innerWidth)+"┘")

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
			// Build the styled line with cyan background
			var styledLine strings.Builder
			styledLine.WriteString("\033[46;30m") // Cyan bg, black text
			styledLine.WriteString(dialogLine)
			styledLine.WriteString("\033[0m")

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}

// overlaySaveAs overlays the Save As dialog centered on the viewport
func (e *Editor) overlaySaveAs(viewportContent string) string {
	// Box dimensions
	boxWidth := 52
	visibleHeight := e.saveAsVisibleHeight()
	boxHeight := visibleHeight + 7 // +7 for header (4 with filename) and footer (3)

	// Helper to pad/truncate text to box width (minus borders)
	innerWidth := boxWidth - 2 // Account for left and right borders
	padText := func(s string, width int) string {
		if len(s) > width {
			return s[:width]
		}
		return s + strings.Repeat(" ", width-len(s))
	}
	centerText := func(s string, width int) string {
		if len(s) >= width {
			return s[:width]
		}
		padLeft := (width - len(s)) / 2
		padRight := width - len(s) - padLeft
		return strings.Repeat(" ", padLeft) + s + strings.Repeat(" ", padRight)
	}

	// Truncate directory path if too long
	dirDisplay := e.fileBrowserDir
	maxDirLen := innerWidth - 12 // "Directory: " prefix
	if len(dirDisplay) > maxDirLen {
		dirDisplay = "..." + dirDisplay[len(dirDisplay)-maxDirLen+3:]
	}

	// Build the dialog lines
	var dialogLines []string

	// Top border with title
	title := " Save As "
	titlePadLeft := (boxWidth - 2 - len(title)) / 2
	titlePadRight := boxWidth - 2 - len(title) - titlePadLeft
	dialogLines = append(dialogLines, "┌"+strings.Repeat("─", titlePadLeft)+title+strings.Repeat("─", titlePadRight)+"┐")

	// Directory line
	dialogLines = append(dialogLines, "│"+padText(" Directory: "+dirDisplay, innerWidth)+"│")

	// Filename input line (highlighted when focused)
	filenameDisplay := e.saveAsFilename
	if len(filenameDisplay) > innerWidth-12 {
		filenameDisplay = filenameDisplay[len(filenameDisplay)-innerWidth+12:]
	}
	filenameLabel := " Filename: " + filenameDisplay
	if e.fileBrowserSelected < 0 {
		// Focused - show cursor
		filenameLabel += "_"
	}
	filenameLine := padText(filenameLabel, innerWidth)
	if e.fileBrowserSelected < 0 {
		// Highlight filename input when focused
		filenameLine = "\033[44;97m" + filenameLine + "\033[46;30m"
	}
	dialogLines = append(dialogLines, "│"+filenameLine+"│")

	// Separator
	dialogLines = append(dialogLines, "├"+strings.Repeat("─", innerWidth)+"┤")

	// Directory list (only directories, for navigation)
	for i := 0; i < visibleHeight; i++ {
		idx := e.fileBrowserScroll + i
		if idx < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[idx]
			line := fmt.Sprintf(" %-46s ", entry.Name+"/")
			if len(line) > innerWidth {
				line = line[:innerWidth]
			} else {
				line = padText(line, innerWidth)
			}
			// Highlight selected item
			if idx == e.fileBrowserSelected {
				// Selected: white on blue
				line = "\033[44;97m" + line + "\033[46;30m"
			}
			dialogLines = append(dialogLines, "│"+line+"│")
		} else {
			// Empty line
			dialogLines = append(dialogLines, "│"+strings.Repeat(" ", innerWidth)+"│")
		}
	}

	// Separator
	dialogLines = append(dialogLines, "├"+strings.Repeat("─", innerWidth)+"┤")

	// Help lines
	helpText1 := "[Enter] Save  [Esc] Cancel  [Tab] Switch"
	helpText2 := "[←/→] Navigate dirs  [↑/↓] Select dir"
	dialogLines = append(dialogLines, "│"+centerText(helpText1, innerWidth)+"│")
	dialogLines = append(dialogLines, "│"+centerText(helpText2, innerWidth)+"│")

	// Bottom border
	dialogLines = append(dialogLines, "└"+strings.Repeat("─", innerWidth)+"┘")

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
			// Build the styled line with cyan background
			var styledLine strings.Builder
			styledLine.WriteString("\033[46;30m") // Cyan bg, black text
			styledLine.WriteString(dialogLine)
			styledLine.WriteString("\033[0m")

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}
