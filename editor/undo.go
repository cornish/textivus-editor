package editor

import "time"

// UndoEntry represents a single change that can be undone/redone.
type UndoEntry struct {
	// Position where the change occurred (byte offset)
	Position int
	// Text that was deleted (empty for pure insertions)
	Deleted string
	// Text that was inserted (empty for pure deletions)
	Inserted string
	// Cursor position before the change
	CursorBefore int
	// Cursor position after the change
	CursorAfter int
	// Timestamp of the change (for grouping)
	Timestamp time.Time
}

// UndoStack manages undo and redo operations.
type UndoStack struct {
	undoStack []*UndoEntry
	redoStack []*UndoEntry
	maxSize   int
	// Grouping: changes within this duration are grouped together
	groupingInterval time.Duration
	lastChange       time.Time
}

// NewUndoStack creates a new undo stack with the given maximum size.
func NewUndoStack(maxSize int) *UndoStack {
	return &UndoStack{
		undoStack:        make([]*UndoEntry, 0, maxSize),
		redoStack:        make([]*UndoEntry, 0, maxSize),
		maxSize:          maxSize,
		groupingInterval: 500 * time.Millisecond,
	}
}

// Push adds a new entry to the undo stack.
// This clears the redo stack since we're making a new change.
func (u *UndoStack) Push(entry *UndoEntry) {
	entry.Timestamp = time.Now()

	// Try to merge with the last entry if it's recent and compatible
	if u.shouldMerge(entry) {
		last := u.undoStack[len(u.undoStack)-1]
		u.mergeEntries(last, entry)
	} else {
		u.undoStack = append(u.undoStack, entry)

		// Trim if over max size
		if len(u.undoStack) > u.maxSize {
			u.undoStack = u.undoStack[1:]
		}
	}

	// Clear redo stack on new change
	u.redoStack = u.redoStack[:0]
	u.lastChange = entry.Timestamp
}

// shouldMerge returns true if the new entry should be merged with the last one.
func (u *UndoStack) shouldMerge(entry *UndoEntry) bool {
	if len(u.undoStack) == 0 {
		return false
	}

	last := u.undoStack[len(u.undoStack)-1]

	// Check if within grouping interval
	if time.Since(last.Timestamp) > u.groupingInterval {
		return false
	}

	// Merge consecutive character insertions at adjacent positions
	if last.Deleted == "" && entry.Deleted == "" {
		// Both are pure insertions
		expectedPos := last.Position + len(last.Inserted)
		if entry.Position == expectedPos && len(entry.Inserted) == 1 {
			// Check if we should break on word boundaries
			r := rune(entry.Inserted[0])
			if r == ' ' || r == '\n' || r == '\t' {
				return false
			}
			return true
		}
	}

	// Merge consecutive character deletions at the same or adjacent positions
	if last.Inserted == "" && entry.Inserted == "" {
		// Backspace: deleting character before cursor
		if entry.Position == last.Position-len(entry.Deleted) {
			return true
		}
		// Delete: deleting character at cursor
		if entry.Position == last.Position {
			return true
		}
	}

	return false
}

// mergeEntries merges the new entry into the existing last entry.
func (u *UndoStack) mergeEntries(last, entry *UndoEntry) {
	if last.Deleted == "" && entry.Deleted == "" {
		// Merge insertions
		last.Inserted += entry.Inserted
		last.CursorAfter = entry.CursorAfter
	} else if last.Inserted == "" && entry.Inserted == "" {
		// Merge deletions
		if entry.Position < last.Position {
			// Backspace: prepend deleted text
			last.Deleted = entry.Deleted + last.Deleted
			last.Position = entry.Position
		} else {
			// Delete key: append deleted text
			last.Deleted += entry.Deleted
		}
		last.CursorAfter = entry.CursorAfter
	}
	last.Timestamp = entry.Timestamp
}

// Undo pops and returns the last entry from the undo stack, or nil if empty.
// The entry is moved to the redo stack.
func (u *UndoStack) Undo() *UndoEntry {
	if len(u.undoStack) == 0 {
		return nil
	}

	entry := u.undoStack[len(u.undoStack)-1]
	u.undoStack = u.undoStack[:len(u.undoStack)-1]
	u.redoStack = append(u.redoStack, entry)

	return entry
}

// Redo pops and returns the last entry from the redo stack, or nil if empty.
// The entry is moved back to the undo stack.
func (u *UndoStack) Redo() *UndoEntry {
	if len(u.redoStack) == 0 {
		return nil
	}

	entry := u.redoStack[len(u.redoStack)-1]
	u.redoStack = u.redoStack[:len(u.redoStack)-1]
	u.undoStack = append(u.undoStack, entry)

	return entry
}

// CanUndo returns true if there are entries to undo.
func (u *UndoStack) CanUndo() bool {
	return len(u.undoStack) > 0
}

// CanRedo returns true if there are entries to redo.
func (u *UndoStack) CanRedo() bool {
	return len(u.redoStack) > 0
}

// Clear clears both the undo and redo stacks.
func (u *UndoStack) Clear() {
	u.undoStack = u.undoStack[:0]
	u.redoStack = u.redoStack[:0]
}

// BreakMerge forces the next change to not merge with previous ones.
func (u *UndoStack) BreakMerge() {
	u.lastChange = time.Time{}
}

// SetGroupingInterval sets the interval for grouping changes.
func (u *UndoStack) SetGroupingInterval(d time.Duration) {
	u.groupingInterval = d
}
