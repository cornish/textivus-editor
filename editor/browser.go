package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

// handleFileBrowserMouse handles mouse input in file browser mode
func (e *Editor) handleFileBrowserMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Calculate dialog position (must match overlayFileBrowser)
	boxWidth := 52
	visibleHeight := e.fileBrowserVisibleHeight()
	boxHeight := visibleHeight + 6

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
			e.statusbar.SetMessage("Cancelled", "info")
		}
		return e, nil
	}

	// File list starts at line 3 (after title, directory, separator)
	fileListStart := 3
	fileListEnd := fileListStart + visibleHeight

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			// Check if click is in file list area
			if relY >= fileListStart && relY < fileListEnd {
				clickedIdx := e.fileBrowserScroll + (relY - fileListStart)
				if clickedIdx >= 0 && clickedIdx < len(e.fileBrowserEntries) {
					if e.fileBrowserSelected == clickedIdx {
						// Double-click effect: same item clicked again - open it
						if !e.browserEnterDirectory() {
							// Not a directory - open the file
							entry := e.fileBrowserEntries[e.fileBrowserSelected]
							if !entry.IsDir {
								fullPath := filepath.Join(e.fileBrowserDir, entry.Name)
								if err := e.LoadFile(fullPath); err != nil {
									// Show error in dialog, stay open
									e.fileBrowserError = "Open failed: " + err.Error()
								} else {
									e.mode = ModeNormal
									e.fileBrowserError = ""
									e.statusbar.SetMessage("Opened: "+fullPath, "success")
								}
							}
						}
					} else {
						// First click - just select
						e.fileBrowserSelected = clickedIdx
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if relY >= fileListStart && relY < fileListEnd {
			if e.fileBrowserScroll > 0 {
				e.fileBrowserScroll--
				// Keep selection visible
				if e.fileBrowserSelected >= e.fileBrowserScroll+visibleHeight {
					e.fileBrowserSelected = e.fileBrowserScroll + visibleHeight - 1
				}
			}
		}

	case tea.MouseButtonWheelDown:
		if relY >= fileListStart && relY < fileListEnd {
			maxScroll := len(e.fileBrowserEntries) - visibleHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if e.fileBrowserScroll < maxScroll {
				e.fileBrowserScroll++
				// Keep selection visible
				if e.fileBrowserSelected < e.fileBrowserScroll {
					e.fileBrowserSelected = e.fileBrowserScroll
				}
			}
		}
	}

	return e, nil
}

// handleSaveAsMouse handles mouse input in Save As mode
func (e *Editor) handleSaveAsMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Calculate dialog position (must match overlaySaveAs)
	boxWidth := 52
	visibleHeight := e.saveAsVisibleHeight()
	boxHeight := visibleHeight + 7

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
			e.statusbar.SetMessage("Cancelled", "info")
		}
		return e, nil
	}

	// Layout:
	// 0: title border
	// 1: directory line
	// 2: filename input line
	// 3: separator
	// 4 to 4+visibleHeight-1: file list
	// then: separator, status, help, bottom border

	filenameLineY := 2
	fileListStart := 4
	fileListEnd := fileListStart + visibleHeight

	switch msg.Button {
	case tea.MouseButtonLeft:
		if msg.Action == tea.MouseActionPress {
			// Click on filename line - focus filename input
			if relY == filenameLineY {
				e.saveAsFocusBrowser = false
				return e, nil
			}

			// Check if click is in file list area
			if relY >= fileListStart && relY < fileListEnd {
				e.saveAsFocusBrowser = true
				clickedIdx := e.fileBrowserScroll + (relY - fileListStart)
				if clickedIdx >= 0 && clickedIdx < len(e.fileBrowserEntries) {
					if e.fileBrowserSelected == clickedIdx {
						// Double-click effect: same item clicked again
						if !e.browserEnterDirectory() {
							// Not a directory - copy filename to input
							entry := e.fileBrowserEntries[e.fileBrowserSelected]
							if !entry.IsDir {
								e.saveAsFilename = entry.Name
								e.saveAsFocusBrowser = false
								e.fileBrowserError = ""
							}
						}
					} else {
						// First click - just select
						e.fileBrowserSelected = clickedIdx
					}
				}
			}
		}

	case tea.MouseButtonWheelUp:
		if relY >= fileListStart && relY < fileListEnd {
			if e.fileBrowserScroll > 0 {
				e.fileBrowserScroll--
				if e.fileBrowserSelected >= e.fileBrowserScroll+visibleHeight {
					e.fileBrowserSelected = e.fileBrowserScroll + visibleHeight - 1
				}
			}
		}

	case tea.MouseButtonWheelDown:
		if relY >= fileListStart && relY < fileListEnd {
			maxScroll := len(e.fileBrowserEntries) - visibleHeight
			if maxScroll < 0 {
				maxScroll = 0
			}
			if e.fileBrowserScroll < maxScroll {
				e.fileBrowserScroll++
				if e.fileBrowserSelected < e.fileBrowserScroll {
					e.fileBrowserSelected = e.fileBrowserScroll
				}
			}
		}
	}

	return e, nil
}

// handleFileBrowserKey handles keyboard input in file browser mode
func (e *Editor) handleFileBrowserKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := e.fileBrowserVisibleHeight()

	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.statusbar.SetMessage("Cancelled", "info")

	case tea.KeyEnter:
		if !e.browserEnterDirectory() {
			// Not a directory - open the file
			if e.fileBrowserSelected >= 0 && e.fileBrowserSelected < len(e.fileBrowserEntries) {
				entry := e.fileBrowserEntries[e.fileBrowserSelected]
				if !entry.IsDir {
					fullPath := filepath.Join(e.fileBrowserDir, entry.Name)
					if err := e.LoadFile(fullPath); err != nil {
						// Show error in dialog, stay open
						e.fileBrowserError = "Open failed: " + err.Error()
					} else {
						e.mode = ModeNormal
						e.fileBrowserError = ""
						e.statusbar.SetMessage("Opened: "+fullPath, "success")
					}
				}
			}
		}

	case tea.KeyBackspace:
		e.browserGoToParent()

	case tea.KeyUp:
		e.browserNavigateUp()

	case tea.KeyDown:
		e.browserNavigateDown(visibleHeight)

	case tea.KeyHome:
		e.browserNavigateHome()

	case tea.KeyEnd:
		e.browserNavigateEnd(visibleHeight)

	case tea.KeyPgUp:
		e.browserNavigatePgUp(visibleHeight)

	case tea.KeyPgDown:
		e.browserNavigatePgDown(visibleHeight)
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
	// Subtract header (title + directory + separator) and footer (separator + status + help)
	return boxHeight - 6
}

// Shared file browser navigation functions

// browserNavigateUp moves selection up one item
func (e *Editor) browserNavigateUp() {
	if e.fileBrowserSelected > 0 {
		e.fileBrowserSelected--
		if e.fileBrowserSelected < e.fileBrowserScroll {
			e.fileBrowserScroll = e.fileBrowserSelected
		}
	}
}

// browserNavigateDown moves selection down one item
func (e *Editor) browserNavigateDown(visibleHeight int) {
	if e.fileBrowserSelected < len(e.fileBrowserEntries)-1 {
		e.fileBrowserSelected++
		if e.fileBrowserSelected >= e.fileBrowserScroll+visibleHeight {
			e.fileBrowserScroll = e.fileBrowserSelected - visibleHeight + 1
		}
	}
}

// browserNavigateHome moves selection to first item
func (e *Editor) browserNavigateHome() {
	e.fileBrowserSelected = 0
	e.fileBrowserScroll = 0
}

// browserNavigateEnd moves selection to last item
func (e *Editor) browserNavigateEnd(visibleHeight int) {
	e.fileBrowserSelected = len(e.fileBrowserEntries) - 1
	if e.fileBrowserSelected < 0 {
		e.fileBrowserSelected = 0
	}
	if e.fileBrowserSelected >= visibleHeight {
		e.fileBrowserScroll = e.fileBrowserSelected - visibleHeight + 1
	}
}

// browserNavigatePgUp moves selection up by a page
func (e *Editor) browserNavigatePgUp(visibleHeight int) {
	e.fileBrowserSelected -= visibleHeight
	if e.fileBrowserSelected < 0 {
		e.fileBrowserSelected = 0
	}
	e.fileBrowserScroll -= visibleHeight
	if e.fileBrowserScroll < 0 {
		e.fileBrowserScroll = 0
	}
}

// browserNavigatePgDown moves selection down by a page
func (e *Editor) browserNavigatePgDown(visibleHeight int) {
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

// browserGoToParent navigates to parent directory
func (e *Editor) browserGoToParent() {
	if e.fileBrowserDir != "/" {
		newPath := filepath.Clean(filepath.Dir(e.fileBrowserDir))
		e.loadDirectory(newPath)
	}
}

// browserEnterDirectory enters the selected directory if it's a readable directory
// Returns true if navigation occurred, false otherwise
func (e *Editor) browserEnterDirectory() bool {
	if len(e.fileBrowserEntries) == 0 || e.fileBrowserSelected < 0 || e.fileBrowserSelected >= len(e.fileBrowserEntries) {
		return false
	}
	entry := e.fileBrowserEntries[e.fileBrowserSelected]
	if !entry.IsDir {
		return false
	}
	if !entry.Readable {
		e.fileBrowserError = "Permission denied: " + entry.Name
		return true // Handled, but with error
	}
	e.fileBrowserError = "" // Clear error
	var newPath string
	if entry.Name == ".." {
		newPath = filepath.Dir(e.fileBrowserDir)
	} else {
		newPath = filepath.Join(e.fileBrowserDir, entry.Name)
	}
	e.loadDirectory(filepath.Clean(newPath))
	return true
}

// saveAsVisibleHeight returns the number of visible file entries in Save As
func (e *Editor) saveAsVisibleHeight() int {
	boxHeight := e.viewport.Height() - 4
	if boxHeight > 18 {
		boxHeight = 18
	}
	if boxHeight < 5 {
		boxHeight = 5
	}
	// Subtract header (title + directory + filename + separator) and footer (separator + status + help)
	return boxHeight - 7
}

// handleSaveAsKey handles keyboard input in Save As mode
func (e *Editor) handleSaveAsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	visibleHeight := e.saveAsVisibleHeight()

	switch msg.Type {
	case tea.KeyEsc:
		e.mode = ModeNormal
		e.statusbar.SetMessage("Cancelled", "info")

	case tea.KeyTab:
		// Toggle focus between filename field and browser
		e.saveAsFocusBrowser = !e.saveAsFocusBrowser

	case tea.KeyEnter:
		if e.saveAsFocusBrowser {
			// Focus on browser - handle directory navigation or file selection
			if !e.browserEnterDirectory() {
				// Not a directory - copy filename to input
				if e.fileBrowserSelected >= 0 && e.fileBrowserSelected < len(e.fileBrowserEntries) {
					entry := e.fileBrowserEntries[e.fileBrowserSelected]
					if !entry.IsDir {
						e.saveAsFilename = entry.Name
						e.saveAsFocusBrowser = false
						e.fileBrowserError = ""
					}
				}
			}
		} else {
			// Focus on filename - save the file
			if e.saveAsFilename == "" {
				e.fileBrowserError = "Enter a filename"
				return e, nil
			}
			fullPath := filepath.Join(e.fileBrowserDir, e.saveAsFilename)
			// Check if file exists
			if _, err := os.Stat(fullPath); err == nil {
				// File exists - prompt for confirmation
				e.pendingFilename = fullPath
				e.promptText = "Overwrite? (y/n): "
				e.promptInput = ""
				e.promptAction = PromptConfirmOverwrite
				e.mode = ModePrompt
				return e, nil
			}
			// Save the file - try first, only close dialog on success
			oldFilename := e.filename
			e.filename = fullPath
			if e.doSaveInDialog() {
				e.mode = ModeNormal
				e.updateTitle()
			} else {
				// Save failed - restore filename and keep dialog open
				e.filename = oldFilename
			}
		}

	case tea.KeyBackspace:
		if e.saveAsFocusBrowser {
			e.browserGoToParent()
		} else {
			// Delete from filename
			if len(e.saveAsFilename) > 0 {
				e.saveAsFilename = e.saveAsFilename[:len(e.saveAsFilename)-1]
			}
		}

	case tea.KeyUp:
		if e.saveAsFocusBrowser {
			e.browserNavigateUp()
		} else {
			// Switch focus to browser
			e.saveAsFocusBrowser = true
		}

	case tea.KeyDown:
		if e.saveAsFocusBrowser {
			e.browserNavigateDown(visibleHeight)
		} else {
			// Switch focus to browser
			e.saveAsFocusBrowser = true
		}

	case tea.KeyHome:
		if e.saveAsFocusBrowser {
			e.browserNavigateHome()
		}

	case tea.KeyEnd:
		if e.saveAsFocusBrowser {
			e.browserNavigateEnd(visibleHeight)
		}

	case tea.KeyPgUp:
		if e.saveAsFocusBrowser {
			e.browserNavigatePgUp(visibleHeight)
		}

	case tea.KeyPgDown:
		if e.saveAsFocusBrowser {
			e.browserNavigatePgDown(visibleHeight)
		}

	case tea.KeyRunes:
		// Always type into filename field, switch focus there
		e.saveAsFilename += string(msg.Runes)
		e.saveAsFocusBrowser = false

	case tea.KeySpace:
		e.saveAsFilename += " "
		e.saveAsFocusBrowser = false
	}

	return e, nil
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
	e.fileBrowserError = "" // Clear any previous error
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
	e.fileBrowserSelected = 0
	e.fileBrowserScroll = 0
	e.fileBrowserError = "" // Clear any previous error
	e.saveAsFocusBrowser = false // Start with focus on filename field
	e.loadDirectory(startDir)
	e.mode = ModeSaveAs
}

// loadDirectory reads the contents of a directory and populates the file browser
func (e *Editor) loadDirectory(path string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		e.fileBrowserError = "Cannot open: " + err.Error()
		return
	}

	// Clear any previous error on success
	e.fileBrowserError = ""

	e.fileBrowserEntries = make([]FileEntry, 0, len(entries)+1)

	// Add parent directory entry if not at root
	if path != "/" {
		e.fileBrowserEntries = append(e.fileBrowserEntries, FileEntry{
			Name:     "..",
			IsDir:    true,
			Size:     0,
			Readable: true, // Parent is always readable (we came from there)
		})
	}

	// Convert to FileEntry and separate dirs from files
	// Note: We avoid calling entry.Info() for directories as it can hang
	// on stale network mounts. Only call Info() for files (to get size).
	var dirs, files []FileEntry
	for _, entry := range entries {
		if entry.IsDir() {
			// For directories, don't call Info() - it can hang on stale mounts
			// We'll check readability when they try to enter
			dirs = append(dirs, FileEntry{
				Name:     entry.Name(),
				IsDir:    true,
				Size:     0,
				Readable: true, // Assume readable, show error on enter
			})
		} else {
			// For files, get info for size (less likely to hang)
			info, err := entry.Info()
			if err != nil {
				continue
			}
			files = append(files, FileEntry{
				Name:     entry.Name(),
				IsDir:    false,
				Size:     info.Size(),
				Readable: true,
			})
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

	e.fileBrowserDir = filepath.Clean(path)
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

// overlayFileBrowser overlays the file browser dialog centered on the viewport
func (e *Editor) overlayFileBrowser(viewportContent string) string {
	// Box dimensions
	boxWidth := 52
	visibleHeight := e.fileBrowserVisibleHeight()
	boxHeight := visibleHeight + 6 // +6 for header (3), status (1), and footer (2)

	// Get theme colors for internal styling
	themeUI := e.styles.Theme.UI
	selectedStyle := "\033[" + colorToSGR(themeUI.DialogButtonFg, themeUI.DialogButton) + "m"
	dialogResetStyle := "\033[" + colorToSGR(themeUI.DialogFg, themeUI.DialogBg) + "m"
	errorStyle := "\033[" + colorToSGRSingle(themeUI.ErrorFg, true) + "m"

	// Helper to pad/truncate text to exact display width (Unicode-aware)
	innerWidth := boxWidth - 2 // Account for left and right borders
	padText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw >= width {
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
	// Truncate string to fit display width, with prefix
	truncateWithPrefix := func(s string, width int, prefix string) string {
		sw := runewidth.StringWidth(s)
		if sw <= width {
			return s
		}
		// Need to truncate from the start, showing "..." prefix
		prefixWidth := runewidth.StringWidth(prefix)
		targetWidth := width - prefixWidth
		// Find how many runes from the end fit
		runes := []rune(s)
		endWidth := 0
		startIdx := len(runes)
		for i := len(runes) - 1; i >= 0 && endWidth < targetWidth; i-- {
			rw := runewidth.RuneWidth(runes[i])
			if endWidth+rw <= targetWidth {
				endWidth += rw
				startIdx = i
			} else {
				break
			}
		}
		return prefix + string(runes[startIdx:])
	}

	// Truncate directory path if too long
	dirDisplay := e.fileBrowserDir
	maxDirLen := innerWidth - 12 // "Directory: " prefix
	dirDisplay = truncateWithPrefix(dirDisplay, maxDirLen, e.box.Ellipsis)

	// Build the dialog lines
	var dialogLines []string

	// Top border with title
	title := " Open File "
	titlePadLeft := (boxWidth - 2 - len(title)) / 2
	titlePadRight := boxWidth - 2 - len(title) - titlePadLeft
	dialogLines = append(dialogLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Directory line
	dialogLines = append(dialogLines, e.box.Vertical+padText(" Directory: "+dirDisplay, innerWidth)+e.box.Vertical)

	// Separator
	dialogLines = append(dialogLines, e.box.TeeLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.TeeRight)

	// File list
	lockWidth := runewidth.StringWidth(e.box.Lock)
	for i := 0; i < visibleHeight; i++ {
		idx := e.fileBrowserScroll + i
		if idx < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[idx]
			// Truncate filename if needed (leave room for size column)
			nameWidth := 36
			name := entry.Name
			if runewidth.StringWidth(name) > nameWidth {
				name = runewidth.Truncate(name, nameWidth-1, e.box.Ellipsis)
			}
			// Pad name to fixed width
			namePadded := name + strings.Repeat(" ", nameWidth-runewidth.StringWidth(name))

			var line string
			if entry.IsDir {
				if !entry.Readable {
					// Unreadable directory - show lock before name
					line = " " + e.box.Lock + " " + namePadded + fmt.Sprintf("%6s ", "<DIR>")
				} else {
					line = strings.Repeat(" ", lockWidth+2) + namePadded + fmt.Sprintf("%6s ", "<DIR>")
				}
			} else {
				line = strings.Repeat(" ", lockWidth+2) + namePadded + fmt.Sprintf("%6s ", formatFileSize(entry.Size))
			}
			// Pad line to inner width using Unicode-aware width
			line = padText(line, innerWidth)
			// Style the line
			if idx == e.fileBrowserSelected {
				// Selected: use theme button colors
				line = selectedStyle + line + dialogResetStyle
			} else if entry.IsDir && !entry.Readable {
				// Unreadable: dim/gray
				line = "\033[2m" + line + "\033[22m"
			}
			dialogLines = append(dialogLines, e.box.Vertical+line+e.box.Vertical)
		} else {
			// Empty line
			dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)
		}
	}

	// Separator
	dialogLines = append(dialogLines, e.box.TeeLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.TeeRight)

	// Status/error line
	statusLine := ""
	if e.fileBrowserError != "" {
		statusLine = e.fileBrowserError
		if len(statusLine) > innerWidth {
			statusLine = statusLine[:innerWidth]
		}
		statusLine = errorStyle + padText(statusLine, innerWidth) + dialogResetStyle
	} else {
		statusLine = padText("", innerWidth)
	}
	dialogLines = append(dialogLines, e.box.Vertical+statusLine+e.box.Vertical)

	// Help line
	helpText := "Click/Enter: Open  Esc: Cancel  Bksp: Parent"
	dialogLines = append(dialogLines, e.box.Vertical+centerText(helpText, innerWidth)+e.box.Vertical)

	// Bottom border
	dialogLines = append(dialogLines, e.box.BottomLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.BottomRight)

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
			// Build the styled line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(dialogResetStyle)
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
	boxHeight := visibleHeight + 7 // +7 for header (4 with filename), status (1), and footer (2)

	// Get theme colors for internal styling
	themeUI := e.styles.Theme.UI
	selectedStyle := "\033[" + colorToSGR(themeUI.DialogButtonFg, themeUI.DialogButton) + "m"
	dialogResetStyle := "\033[" + colorToSGR(themeUI.DialogFg, themeUI.DialogBg) + "m"
	errorStyle := "\033[" + colorToSGRSingle(themeUI.ErrorFg, true) + "m"

	// Helper to pad/truncate text to exact display width (Unicode-aware)
	innerWidth := boxWidth - 2 // Account for left and right borders
	padText := func(s string, width int) string {
		sw := runewidth.StringWidth(s)
		if sw >= width {
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
	// Truncate string to fit display width, with prefix
	truncateWithPrefix := func(s string, width int, prefix string) string {
		sw := runewidth.StringWidth(s)
		if sw <= width {
			return s
		}
		// Need to truncate from the start, showing "..." prefix
		prefixWidth := runewidth.StringWidth(prefix)
		targetWidth := width - prefixWidth
		// Find how many runes from the end fit
		runes := []rune(s)
		endWidth := 0
		startIdx := len(runes)
		for i := len(runes) - 1; i >= 0 && endWidth < targetWidth; i-- {
			rw := runewidth.RuneWidth(runes[i])
			if endWidth+rw <= targetWidth {
				endWidth += rw
				startIdx = i
			} else {
				break
			}
		}
		return prefix + string(runes[startIdx:])
	}

	// Truncate directory path if too long
	dirDisplay := e.fileBrowserDir
	maxDirLen := innerWidth - 12 // "Directory: " prefix
	dirDisplay = truncateWithPrefix(dirDisplay, maxDirLen, e.box.Ellipsis)

	// Build the dialog lines
	var dialogLines []string

	// Top border with title
	title := " Save As "
	titlePadLeft := (boxWidth - 2 - len(title)) / 2
	titlePadRight := boxWidth - 2 - len(title) - titlePadLeft
	dialogLines = append(dialogLines, e.box.TopLeft+strings.Repeat(e.box.Horizontal, titlePadLeft)+title+strings.Repeat(e.box.Horizontal, titlePadRight)+e.box.TopRight)

	// Directory line
	dialogLines = append(dialogLines, e.box.Vertical+padText(" Directory: "+dirDisplay, innerWidth)+e.box.Vertical)

	// Filename input line - show block cursor when focused
	filenameDisplay := e.saveAsFilename
	editAreaWidth := innerWidth - 11 // " Filename: " prefix is 11 chars
	fnWidth := runewidth.StringWidth(filenameDisplay)
	if fnWidth > editAreaWidth-1 { // -1 for cursor
		// Truncate from start to show end of filename
		filenameDisplay = runewidth.TruncateLeft(filenameDisplay, fnWidth-(editAreaWidth-1), "")
		fnWidth = runewidth.StringWidth(filenameDisplay)
	}
	var filenameLine string
	if !e.saveAsFocusBrowser {
		// Focused - show filename with block cursor at end
		cursor := "\033[7m \033[27m" // Reverse video space (block cursor)
		padding := editAreaWidth - fnWidth - 1
		if padding < 0 {
			padding = 0
		}
		filenameLine = " Filename: " + filenameDisplay + cursor + strings.Repeat(" ", padding)
	} else {
		// Not focused - just show filename
		filenameLine = padText(" Filename: "+filenameDisplay, innerWidth)
	}
	dialogLines = append(dialogLines, e.box.Vertical+filenameLine+e.box.Vertical)

	// Separator
	dialogLines = append(dialogLines, e.box.TeeLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.TeeRight)

	// File list
	lockWidth := runewidth.StringWidth(e.box.Lock)
	for i := 0; i < visibleHeight; i++ {
		idx := e.fileBrowserScroll + i
		if idx < len(e.fileBrowserEntries) {
			entry := e.fileBrowserEntries[idx]
			// Truncate filename if needed (leave room for size column)
			nameWidth := 36
			name := entry.Name
			if runewidth.StringWidth(name) > nameWidth {
				name = runewidth.Truncate(name, nameWidth-1, e.box.Ellipsis)
			}
			// Pad name to fixed width
			namePadded := name + strings.Repeat(" ", nameWidth-runewidth.StringWidth(name))

			var line string
			if entry.IsDir {
				if !entry.Readable {
					// Unreadable directory - show lock before name
					line = " " + e.box.Lock + " " + namePadded + fmt.Sprintf("%6s ", "<DIR>")
				} else {
					line = strings.Repeat(" ", lockWidth+2) + namePadded + fmt.Sprintf("%6s ", "<DIR>")
				}
			} else {
				line = strings.Repeat(" ", lockWidth+2) + namePadded + fmt.Sprintf("%6s ", formatFileSize(entry.Size))
			}
			// Pad line to inner width using Unicode-aware width
			line = padText(line, innerWidth)
			// Style the line
			if idx == e.fileBrowserSelected && e.saveAsFocusBrowser {
				// Selected: use theme button colors
				line = selectedStyle + line + dialogResetStyle
			} else if entry.IsDir && !entry.Readable {
				// Unreadable: dim/gray
				line = "\033[2m" + line + "\033[22m"
			}
			dialogLines = append(dialogLines, e.box.Vertical+line+e.box.Vertical)
		} else {
			// Empty line
			dialogLines = append(dialogLines, e.box.Vertical+strings.Repeat(" ", innerWidth)+e.box.Vertical)
		}
	}

	// Separator
	dialogLines = append(dialogLines, e.box.TeeLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.TeeRight)

	// Status/error line
	statusLine := ""
	if e.fileBrowserError != "" {
		statusLine = e.fileBrowserError
		if runewidth.StringWidth(statusLine) > innerWidth {
			statusLine = runewidth.Truncate(statusLine, innerWidth, "")
		}
		statusLine = errorStyle + padText(statusLine, innerWidth) + dialogResetStyle
	} else {
		statusLine = padText("", innerWidth)
	}
	dialogLines = append(dialogLines, e.box.Vertical+statusLine+e.box.Vertical)

	// Help line - changes based on focus
	var helpText string
	if e.saveAsFocusBrowser {
		helpText = "Click/Enter: Select  Tab: Switch  Esc: Cancel"
	} else {
		helpText = "Enter: Save  Tab: Browse  Esc: Cancel"
	}
	dialogLines = append(dialogLines, e.box.Vertical+centerText(helpText, innerWidth)+e.box.Vertical)

	// Bottom border
	dialogLines = append(dialogLines, e.box.BottomLeft+strings.Repeat(e.box.Horizontal, innerWidth)+e.box.BottomRight)

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
			// Build the styled line with theme colors
			var styledLine strings.Builder
			styledLine.WriteString(dialogResetStyle)
			styledLine.WriteString(dialogLine)
			styledLine.WriteString("\033[0m")

			// Overlay on viewport line
			viewportLines[viewportY] = overlayLineAt(styledLine.String(), viewportLines[viewportY], startX)
		}
	}

	return strings.Join(viewportLines, "\n")
}
