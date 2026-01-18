package editor

import (
	"unicode"
	"unicode/utf8"
)

// Position represents a position in the text as line and column (both 0-indexed).
type Position struct {
	Line int
	Col  int
}

// Cursor manages the cursor position within a buffer.
type Cursor struct {
	buf *Buffer
	pos int // Byte offset in the buffer
}

// NewCursor creates a new cursor for the given buffer.
func NewCursor(buf *Buffer) *Cursor {
	return &Cursor{
		buf: buf,
		pos: 0,
	}
}

// Position returns the current cursor position as line and column.
func (c *Cursor) Position() Position {
	line, col := c.buf.PositionToLineCol(c.pos)
	return Position{Line: line, Col: col}
}

// ByteOffset returns the current cursor position as a byte offset.
func (c *Cursor) ByteOffset() int {
	return c.pos
}

// SetByteOffset sets the cursor to a specific byte offset.
func (c *Cursor) SetByteOffset(offset int) {
	if offset < 0 {
		offset = 0
	}
	if offset > c.buf.Length() {
		offset = c.buf.Length()
	}
	c.pos = offset
	c.buf.MoveCursor(offset)
}

// SetPosition sets the cursor to a specific line and column.
func (c *Cursor) SetPosition(line, col int) {
	c.pos = c.buf.LineColToPosition(line, col)
	c.buf.MoveCursor(c.pos)
}

// MoveLeft moves the cursor left by one character (rune).
func (c *Cursor) MoveLeft() bool {
	if c.pos == 0 {
		return false
	}
	// Move back through UTF-8 bytes to find the previous rune start
	c.pos--
	for c.pos > 0 && !utf8.RuneStart(c.buf.ByteAt(c.pos)) {
		c.pos--
	}
	c.buf.MoveCursor(c.pos)
	return true
}

// MoveRight moves the cursor right by one character (rune).
func (c *Cursor) MoveRight() bool {
	if c.pos >= c.buf.Length() {
		return false
	}
	_, size := c.buf.RuneAt(c.pos)
	if size == 0 {
		return false
	}
	c.pos += size
	c.buf.MoveCursor(c.pos)
	return true
}

// MoveUp moves the cursor up one line, trying to maintain the column.
func (c *Cursor) MoveUp() bool {
	line, col := c.buf.PositionToLineCol(c.pos)
	if line == 0 {
		return false
	}
	c.pos = c.buf.LineColToPosition(line-1, col)
	c.buf.MoveCursor(c.pos)
	return true
}

// MoveDown moves the cursor down one line, trying to maintain the column.
func (c *Cursor) MoveDown() bool {
	line, col := c.buf.PositionToLineCol(c.pos)
	if line >= c.buf.LineCount()-1 {
		return false
	}
	c.pos = c.buf.LineColToPosition(line+1, col)
	c.buf.MoveCursor(c.pos)
	return true
}

// MoveToLineStart moves the cursor to the start of the current line.
func (c *Cursor) MoveToLineStart() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	c.pos = c.buf.LineStartOffset(line)
	c.buf.MoveCursor(c.pos)
}

// MoveToLineEnd moves the cursor to the end of the current line.
func (c *Cursor) MoveToLineEnd() {
	line, _ := c.buf.PositionToLineCol(c.pos)
	c.pos = c.buf.LineEndOffset(line)
	c.buf.MoveCursor(c.pos)
}

// MoveToStart moves the cursor to the start of the buffer.
func (c *Cursor) MoveToStart() {
	c.pos = 0
	c.buf.MoveCursor(c.pos)
}

// MoveToEnd moves the cursor to the end of the buffer.
func (c *Cursor) MoveToEnd() {
	c.pos = c.buf.Length()
	c.buf.MoveCursor(c.pos)
}

// MoveWordLeft moves the cursor to the start of the previous word.
func (c *Cursor) MoveWordLeft() bool {
	if c.pos == 0 {
		return false
	}

	// Skip any whitespace/punctuation before the word
	for c.pos > 0 {
		c.pos--
		for c.pos > 0 && !utf8.RuneStart(c.buf.ByteAt(c.pos)) {
			c.pos--
		}
		r, _ := c.buf.RuneAt(c.pos)
		if isWordChar(r) {
			break
		}
	}

	// Move to the start of the word
	for c.pos > 0 {
		prevPos := c.pos - 1
		for prevPos > 0 && !utf8.RuneStart(c.buf.ByteAt(prevPos)) {
			prevPos--
		}
		r, _ := c.buf.RuneAt(prevPos)
		if !isWordChar(r) {
			break
		}
		c.pos = prevPos
	}

	c.buf.MoveCursor(c.pos)
	return true
}

// MoveWordRight moves the cursor to the start of the next word.
func (c *Cursor) MoveWordRight() bool {
	if c.pos >= c.buf.Length() {
		return false
	}

	// Skip current word
	for c.pos < c.buf.Length() {
		r, size := c.buf.RuneAt(c.pos)
		if !isWordChar(r) {
			break
		}
		c.pos += size
	}

	// Skip whitespace/punctuation to next word
	for c.pos < c.buf.Length() {
		r, size := c.buf.RuneAt(c.pos)
		if isWordChar(r) {
			break
		}
		c.pos += size
	}

	c.buf.MoveCursor(c.pos)
	return true
}

// isWordChar returns true if the rune is part of a word.
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// Line returns the current line number (0-indexed).
func (c *Cursor) Line() int {
	line, _ := c.buf.PositionToLineCol(c.pos)
	return line
}

// Col returns the current column number (0-indexed).
func (c *Cursor) Col() int {
	_, col := c.buf.PositionToLineCol(c.pos)
	return col
}

// Sync ensures the buffer's gap is at the cursor position.
func (c *Cursor) Sync() {
	c.buf.MoveCursor(c.pos)
}
