package editor

// Selection represents a text selection in the buffer.
// The selection spans from Anchor to Cursor, where Anchor is where the selection
// started and Cursor is the current position (and can be before or after Anchor).
type Selection struct {
	Active bool // Whether there is an active selection
	Anchor int  // Byte offset where selection started
	Cursor int  // Byte offset where selection ends (current cursor position)
}

// NewSelection creates a new inactive selection.
func NewSelection() *Selection {
	return &Selection{
		Active: false,
		Anchor: 0,
		Cursor: 0,
	}
}

// Start begins a new selection at the given position.
func (s *Selection) Start(pos int) {
	s.Active = true
	s.Anchor = pos
	s.Cursor = pos
}

// Update updates the cursor end of the selection.
func (s *Selection) Update(pos int) {
	if s.Active {
		s.Cursor = pos
	}
}

// Clear clears the selection.
func (s *Selection) Clear() {
	s.Active = false
	s.Anchor = 0
	s.Cursor = 0
}

// StartPos returns the start position (lower of Anchor and Cursor).
func (s *Selection) StartPos() int {
	if s.Anchor < s.Cursor {
		return s.Anchor
	}
	return s.Cursor
}

// EndPos returns the end position (higher of Anchor and Cursor).
func (s *Selection) EndPos() int {
	if s.Anchor > s.Cursor {
		return s.Anchor
	}
	return s.Cursor
}

// Length returns the length of the selection in bytes.
func (s *Selection) Length() int {
	if !s.Active {
		return 0
	}
	return s.EndPos() - s.StartPos()
}

// Contains returns true if the given position is within the selection.
func (s *Selection) Contains(pos int) bool {
	if !s.Active {
		return false
	}
	return pos >= s.StartPos() && pos < s.EndPos()
}

// IsEmpty returns true if the selection is empty or inactive.
func (s *Selection) IsEmpty() bool {
	return !s.Active || s.Anchor == s.Cursor
}

// GetText returns the selected text from the given buffer.
func (s *Selection) GetText(buf *Buffer) string {
	if !s.Active || s.IsEmpty() {
		return ""
	}
	return buf.Substring(s.StartPos(), s.EndPos())
}

// SelectAll selects all text in the buffer.
func (s *Selection) SelectAll(buf *Buffer) {
	s.Active = true
	s.Anchor = 0
	s.Cursor = buf.Length()
}

// SelectWord selects the word at the given position in the buffer.
func (s *Selection) SelectWord(buf *Buffer, pos int) {
	if buf.Length() == 0 {
		return
	}

	// Clamp position to valid range
	if pos < 0 {
		pos = 0
	}
	if pos >= buf.Length() {
		pos = buf.Length() - 1
	}

	// Find word boundaries
	r, _ := buf.RuneAt(pos)

	// Determine what kind of "word" we're in
	var inWord func(rune) bool
	if isWordChar(r) {
		inWord = isWordChar
	} else if r == ' ' || r == '\t' {
		// Select contiguous whitespace
		inWord = func(r rune) bool { return r == ' ' || r == '\t' }
	} else {
		// Select contiguous punctuation/symbols
		inWord = func(r rune) bool { return !isWordChar(r) && r != ' ' && r != '\t' && r != '\n' }
	}

	// Find start of word
	start := pos
	for start > 0 {
		prevR, _ := buf.RuneAt(start - 1)
		if prevR == 0 {
			break
		}
		// Scan back for rune start
		prevPos := start - 1
		for prevPos > 0 {
			b := buf.ByteAt(prevPos)
			if b&0xC0 != 0x80 { // Not a continuation byte
				break
			}
			prevPos--
		}
		r, _ := buf.RuneAt(prevPos)
		if !inWord(r) {
			break
		}
		start = prevPos
	}

	// Find end of word
	end := pos
	for end < buf.Length() {
		r, size := buf.RuneAt(end)
		if r == 0 || !inWord(r) {
			break
		}
		end += size
	}

	s.Active = true
	s.Anchor = start
	s.Cursor = end
}

// SelectLine selects the entire line at the given position in the buffer.
func (s *Selection) SelectLine(buf *Buffer, pos int) {
	line, _ := buf.PositionToLineCol(pos)
	start := buf.LineStartOffset(line)
	end := buf.LineEndOffset(line)

	// Include the newline if there is one
	if end < buf.Length() {
		r, size := buf.RuneAt(end)
		if r == '\n' {
			end += size
		}
	}

	s.Active = true
	s.Anchor = start
	s.Cursor = end
}

// Normalize returns the selection with start <= end.
func (s *Selection) Normalize() (start, end int) {
	start, end = s.StartPos(), s.EndPos()
	return
}
