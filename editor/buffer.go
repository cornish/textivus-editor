package editor

import (
	"strings"
	"unicode/utf8"
)

// Buffer implements a gap buffer for efficient text editing.
// The gap is always at the cursor position, making inserts/deletes O(1) amortized.
type Buffer struct {
	data     []byte
	gapStart int // Start of the gap (cursor position in logical text)
	gapEnd   int // End of the gap (exclusive)
}

const initialGapSize = 1024

// NewBuffer creates a new empty buffer.
func NewBuffer() *Buffer {
	data := make([]byte, initialGapSize)
	return &Buffer{
		data:     data,
		gapStart: 0,
		gapEnd:   initialGapSize,
	}
}

// NewBufferFromString creates a buffer initialized with the given text.
func NewBufferFromString(s string) *Buffer {
	b := NewBuffer()
	b.Insert(s)
	return b
}

// Length returns the total number of bytes in the buffer (excluding the gap).
func (b *Buffer) Length() int {
	return len(b.data) - b.gapSize()
}

// gapSize returns the current size of the gap.
func (b *Buffer) gapSize() int {
	return b.gapEnd - b.gapStart
}

// expandGap grows the gap to accommodate at least n more bytes.
func (b *Buffer) expandGap(n int) {
	if b.gapSize() >= n {
		return
	}

	// Calculate new size - at least double the required space
	needed := n - b.gapSize()
	newGapSize := max(initialGapSize, needed*2)

	newData := make([]byte, len(b.data)+newGapSize)

	// Copy data before the gap
	copy(newData[:b.gapStart], b.data[:b.gapStart])

	// Copy data after the gap to the end of new buffer
	newGapEnd := b.gapEnd + newGapSize
	copy(newData[newGapEnd:], b.data[b.gapEnd:])

	b.data = newData
	b.gapEnd = newGapEnd
}

// MoveCursor moves the gap to the specified byte position.
func (b *Buffer) MoveCursor(pos int) {
	if pos < 0 {
		pos = 0
	}
	if pos > b.Length() {
		pos = b.Length()
	}

	if pos == b.gapStart {
		return
	}

	if pos < b.gapStart {
		// Move gap left: shift bytes from [pos, gapStart) to right of gap
		moveCount := b.gapStart - pos
		copy(b.data[b.gapEnd-moveCount:b.gapEnd], b.data[pos:b.gapStart])
		b.gapStart = pos
		b.gapEnd -= moveCount
	} else {
		// Move gap right: shift bytes from [gapEnd, gapEnd+delta) to left of gap
		moveCount := pos - b.gapStart
		copy(b.data[b.gapStart:b.gapStart+moveCount], b.data[b.gapEnd:b.gapEnd+moveCount])
		b.gapStart = pos
		b.gapEnd += moveCount
	}
}

// Insert inserts text at the current cursor position.
func (b *Buffer) Insert(s string) {
	if len(s) == 0 {
		return
	}

	b.expandGap(len(s))
	copy(b.data[b.gapStart:], s)
	b.gapStart += len(s)
}

// InsertRune inserts a single rune at the current cursor position.
func (b *Buffer) InsertRune(r rune) {
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	b.expandGap(n)
	copy(b.data[b.gapStart:], buf[:n])
	b.gapStart += n
}

// DeleteBefore deletes n bytes before the cursor.
func (b *Buffer) DeleteBefore(n int) string {
	if n <= 0 || b.gapStart == 0 {
		return ""
	}
	if n > b.gapStart {
		n = b.gapStart
	}
	deleted := string(b.data[b.gapStart-n : b.gapStart])
	b.gapStart -= n
	return deleted
}

// DeleteAfter deletes n bytes after the cursor.
func (b *Buffer) DeleteAfter(n int) string {
	if n <= 0 || b.gapEnd == len(b.data) {
		return ""
	}
	afterLen := len(b.data) - b.gapEnd
	if n > afterLen {
		n = afterLen
	}
	deleted := string(b.data[b.gapEnd : b.gapEnd+n])
	b.gapEnd += n
	return deleted
}

// DeleteRuneBefore deletes the rune immediately before the cursor.
func (b *Buffer) DeleteRuneBefore() string {
	if b.gapStart == 0 {
		return ""
	}
	// Find the start of the previous rune
	pos := b.gapStart - 1
	for pos > 0 && !utf8.RuneStart(b.data[pos]) {
		pos--
	}
	return b.DeleteBefore(b.gapStart - pos)
}

// DeleteRuneAfter deletes the rune immediately after the cursor.
func (b *Buffer) DeleteRuneAfter() string {
	if b.gapEnd == len(b.data) {
		return ""
	}
	_, size := utf8.DecodeRune(b.data[b.gapEnd:])
	return b.DeleteAfter(size)
}

// String returns the entire buffer contents as a string.
func (b *Buffer) String() string {
	var sb strings.Builder
	sb.Grow(b.Length())
	sb.Write(b.data[:b.gapStart])
	sb.Write(b.data[b.gapEnd:])
	return sb.String()
}

// Substring returns a portion of the buffer from byte positions start to end.
func (b *Buffer) Substring(start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > b.Length() {
		end = b.Length()
	}
	if start >= end {
		return ""
	}

	var sb strings.Builder
	sb.Grow(end - start)

	// Handle the portion before the gap
	if start < b.gapStart {
		beforeEnd := min(end, b.gapStart)
		sb.Write(b.data[start:beforeEnd])
	}

	// Handle the portion after the gap
	if end > b.gapStart {
		afterStart := max(start, b.gapStart)
		// Convert logical position to physical position (after gap)
		physicalStart := afterStart - b.gapStart + b.gapEnd
		physicalEnd := end - b.gapStart + b.gapEnd
		sb.Write(b.data[physicalStart:physicalEnd])
	}

	return sb.String()
}

// ByteAt returns the byte at the given logical position.
func (b *Buffer) ByteAt(pos int) byte {
	if pos < 0 || pos >= b.Length() {
		return 0
	}
	if pos < b.gapStart {
		return b.data[pos]
	}
	return b.data[pos-b.gapStart+b.gapEnd]
}

// RuneAt returns the rune at the given byte position.
func (b *Buffer) RuneAt(pos int) (rune, int) {
	if pos < 0 || pos >= b.Length() {
		return 0, 0
	}
	if pos < b.gapStart {
		return utf8.DecodeRune(b.data[pos:b.gapStart])
	}
	physPos := pos - b.gapStart + b.gapEnd
	return utf8.DecodeRune(b.data[physPos:])
}

// CursorPosition returns the current cursor position (byte offset).
func (b *Buffer) CursorPosition() int {
	return b.gapStart
}

// Lines returns all lines in the buffer.
func (b *Buffer) Lines() []string {
	return strings.Split(b.String(), "\n")
}

// LineCount returns the number of lines in the buffer.
func (b *Buffer) LineCount() int {
	count := 1
	for i := 0; i < b.gapStart; i++ {
		if b.data[i] == '\n' {
			count++
		}
	}
	for i := b.gapEnd; i < len(b.data); i++ {
		if b.data[i] == '\n' {
			count++
		}
	}
	return count
}

// LineStartOffset returns the byte offset of the start of the given line (0-indexed).
func (b *Buffer) LineStartOffset(line int) int {
	if line <= 0 {
		return 0
	}

	currentLine := 0
	for i := 0; i < b.gapStart; i++ {
		if b.data[i] == '\n' {
			currentLine++
			if currentLine == line {
				return i + 1
			}
		}
	}
	for i := b.gapEnd; i < len(b.data); i++ {
		if b.data[i] == '\n' {
			currentLine++
			if currentLine == line {
				return i - b.gapEnd + b.gapStart + 1
			}
		}
	}
	return b.Length()
}

// LineEndOffset returns the byte offset of the end of the given line (0-indexed).
// This is the position just before the newline, or the end of the buffer.
func (b *Buffer) LineEndOffset(line int) int {
	currentLine := 0
	for i := 0; i < b.gapStart; i++ {
		if b.data[i] == '\n' {
			if currentLine == line {
				return i
			}
			currentLine++
		}
	}
	for i := b.gapEnd; i < len(b.data); i++ {
		if b.data[i] == '\n' {
			if currentLine == line {
				return i - b.gapEnd + b.gapStart
			}
			currentLine++
		}
	}
	return b.Length()
}

// PositionToLineCol converts a byte offset to line and column (both 0-indexed).
func (b *Buffer) PositionToLineCol(pos int) (line, col int) {
	if pos < 0 {
		return 0, 0
	}
	if pos > b.Length() {
		pos = b.Length()
	}

	line = 0
	lineStart := 0

	// Check bytes before the gap
	limit := min(pos, b.gapStart)
	for i := 0; i < limit; i++ {
		if b.data[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}

	// If position is before or at the gap, we're done
	if pos <= b.gapStart {
		col = pos - lineStart
		return
	}

	// Check bytes after the gap
	physPos := pos - b.gapStart + b.gapEnd
	for i := b.gapEnd; i < physPos; i++ {
		if b.data[i] == '\n' {
			line++
			lineStart = i - b.gapEnd + b.gapStart + 1
		}
	}

	col = pos - lineStart
	return
}

// LineColToPosition converts line and column (both 0-indexed) to byte offset.
func (b *Buffer) LineColToPosition(line, col int) int {
	if line < 0 {
		line = 0
	}
	if col < 0 {
		col = 0
	}

	start := b.LineStartOffset(line)
	end := b.LineEndOffset(line)

	pos := start + col
	if pos > end {
		pos = end
	}
	return pos
}

// Replace replaces text between start and end positions with new text.
func (b *Buffer) Replace(start, end int, text string) {
	if start > end {
		start, end = end, start
	}
	if start < 0 {
		start = 0
	}
	if end > b.Length() {
		end = b.Length()
	}

	// Move cursor to end and delete backward to start
	b.MoveCursor(end)
	b.DeleteBefore(end - start)
	b.Insert(text)
}
