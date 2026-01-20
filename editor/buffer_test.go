package editor

import (
	"strings"
	"testing"
)

func TestNewBuffer(t *testing.T) {
	b := NewBuffer()
	if b.Length() != 0 {
		t.Errorf("NewBuffer().Length() = %d, want 0", b.Length())
	}
	if b.String() != "" {
		t.Errorf("NewBuffer().String() = %q, want empty string", b.String())
	}
}

func TestNewBufferFromString(t *testing.T) {
	tests := []struct {
		input string
	}{
		{""},
		{"hello"},
		{"hello\nworld"},
		{"line1\nline2\nline3"},
		{"unicode: 日本語 émojis 🎉"},
	}

	for _, tt := range tests {
		b := NewBufferFromString(tt.input)
		if got := b.String(); got != tt.input {
			t.Errorf("NewBufferFromString(%q).String() = %q, want %q", tt.input, got, tt.input)
		}
		if got := b.Length(); got != len(tt.input) {
			t.Errorf("NewBufferFromString(%q).Length() = %d, want %d", tt.input, got, len(tt.input))
		}
	}
}

func TestBufferInsert(t *testing.T) {
	b := NewBuffer()
	b.Insert("hello")
	if got := b.String(); got != "hello" {
		t.Errorf("after Insert('hello'), String() = %q, want 'hello'", got)
	}

	b.Insert(" world")
	if got := b.String(); got != "hello world" {
		t.Errorf("after Insert(' world'), String() = %q, want 'hello world'", got)
	}
}

func TestBufferInsertRune(t *testing.T) {
	b := NewBuffer()
	b.InsertRune('H')
	b.InsertRune('i')
	b.InsertRune('!')
	if got := b.String(); got != "Hi!" {
		t.Errorf("after InsertRune sequence, String() = %q, want 'Hi!'", got)
	}

	// Test unicode rune
	b.InsertRune('🎉')
	if got := b.String(); got != "Hi!🎉" {
		t.Errorf("after InsertRune('🎉'), String() = %q, want 'Hi!🎉'", got)
	}
}

func TestBufferMoveCursor(t *testing.T) {
	b := NewBufferFromString("hello world")

	// Move to middle
	b.MoveCursor(5)
	if pos := b.CursorPosition(); pos != 5 {
		t.Errorf("CursorPosition() = %d, want 5", pos)
	}

	// Insert at cursor
	b.Insert(",")
	if got := b.String(); got != "hello, world" {
		t.Errorf("after Insert at pos 5, String() = %q, want 'hello, world'", got)
	}

	// Move to start
	b.MoveCursor(0)
	b.Insert("Say: ")
	if got := b.String(); got != "Say: hello, world" {
		t.Errorf("after Insert at pos 0, String() = %q, want 'Say: hello, world'", got)
	}

	// Boundary tests
	b.MoveCursor(-10) // Should clamp to 0
	if pos := b.CursorPosition(); pos != 0 {
		t.Errorf("CursorPosition() after -10 = %d, want 0", pos)
	}

	b.MoveCursor(1000) // Should clamp to length
	if pos := b.CursorPosition(); pos != b.Length() {
		t.Errorf("CursorPosition() after 1000 = %d, want %d", pos, b.Length())
	}
}

func TestBufferDeleteBefore(t *testing.T) {
	b := NewBufferFromString("hello world")
	b.MoveCursor(5)

	deleted := b.DeleteBefore(2)
	if deleted != "lo" {
		t.Errorf("DeleteBefore(2) returned %q, want 'lo'", deleted)
	}
	if got := b.String(); got != "hel world" {
		t.Errorf("after DeleteBefore(2), String() = %q, want 'hel world'", got)
	}

	// Delete more than available
	b.MoveCursor(2)
	deleted = b.DeleteBefore(10)
	if deleted != "he" {
		t.Errorf("DeleteBefore(10) at pos 2 returned %q, want 'he'", deleted)
	}
	if got := b.String(); got != "l world" {
		t.Errorf("after DeleteBefore(10), String() = %q, want 'l world'", got)
	}

	// Delete at position 0
	b.MoveCursor(0)
	deleted = b.DeleteBefore(1)
	if deleted != "" {
		t.Errorf("DeleteBefore(1) at pos 0 returned %q, want empty", deleted)
	}
}

func TestBufferDeleteAfter(t *testing.T) {
	b := NewBufferFromString("hello world")
	b.MoveCursor(5)

	deleted := b.DeleteAfter(1)
	if deleted != " " {
		t.Errorf("DeleteAfter(1) returned %q, want ' '", deleted)
	}
	if got := b.String(); got != "helloworld" {
		t.Errorf("after DeleteAfter(1), String() = %q, want 'helloworld'", got)
	}

	// Delete more than available
	b.MoveCursor(5)
	deleted = b.DeleteAfter(100)
	if deleted != "world" {
		t.Errorf("DeleteAfter(100) returned %q, want 'world'", deleted)
	}
	if got := b.String(); got != "hello" {
		t.Errorf("after DeleteAfter(100), String() = %q, want 'hello'", got)
	}
}

func TestBufferDeleteRuneBefore(t *testing.T) {
	b := NewBufferFromString("hi🎉!")
	b.MoveCursor(b.Length()) // End of string

	// Delete '!'
	deleted := b.DeleteRuneBefore()
	if deleted != "!" {
		t.Errorf("DeleteRuneBefore() returned %q, want '!'", deleted)
	}

	// Delete emoji (multi-byte)
	deleted = b.DeleteRuneBefore()
	if deleted != "🎉" {
		t.Errorf("DeleteRuneBefore() returned %q, want '🎉'", deleted)
	}

	if got := b.String(); got != "hi" {
		t.Errorf("after deleting emoji, String() = %q, want 'hi'", got)
	}
}

func TestBufferDeleteRuneAfter(t *testing.T) {
	b := NewBufferFromString("🎉hi")
	b.MoveCursor(0)

	// Delete emoji at start
	deleted := b.DeleteRuneAfter()
	if deleted != "🎉" {
		t.Errorf("DeleteRuneAfter() returned %q, want '🎉'", deleted)
	}

	if got := b.String(); got != "hi" {
		t.Errorf("after DeleteRuneAfter(), String() = %q, want 'hi'", got)
	}
}

func TestBufferSubstring(t *testing.T) {
	b := NewBufferFromString("hello world")

	tests := []struct {
		start, end int
		want       string
	}{
		{0, 5, "hello"},
		{6, 11, "world"},
		{0, 11, "hello world"},
		{3, 8, "lo wo"},
		{-5, 5, "hello"},  // Clamped start
		{6, 100, "world"}, // Clamped end
		{8, 3, ""},        // Invalid range
	}

	for _, tt := range tests {
		got := b.Substring(tt.start, tt.end)
		if got != tt.want {
			t.Errorf("Substring(%d, %d) = %q, want %q", tt.start, tt.end, got, tt.want)
		}
	}
}

func TestBufferByteAt(t *testing.T) {
	b := NewBufferFromString("hello")

	tests := []struct {
		pos  int
		want byte
	}{
		{0, 'h'},
		{1, 'e'},
		{4, 'o'},
		{-1, 0},  // Out of bounds
		{100, 0}, // Out of bounds
	}

	for _, tt := range tests {
		got := b.ByteAt(tt.pos)
		if got != tt.want {
			t.Errorf("ByteAt(%d) = %c, want %c", tt.pos, got, tt.want)
		}
	}
}

func TestBufferLineCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 1},
		{"hello", 1},
		{"hello\n", 2},
		{"hello\nworld", 2},
		{"line1\nline2\nline3", 3},
		{"\n\n\n", 4},
	}

	for _, tt := range tests {
		b := NewBufferFromString(tt.input)
		if got := b.LineCount(); got != tt.want {
			t.Errorf("LineCount() for %q = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBufferLines(t *testing.T) {
	b := NewBufferFromString("line1\nline2\nline3")
	lines := b.Lines()

	want := []string{"line1", "line2", "line3"}
	if len(lines) != len(want) {
		t.Fatalf("Lines() returned %d lines, want %d", len(lines), len(want))
	}

	for i, line := range lines {
		if line != want[i] {
			t.Errorf("Lines()[%d] = %q, want %q", i, line, want[i])
		}
	}
}

func TestBufferPositionToLineCol(t *testing.T) {
	b := NewBufferFromString("hello\nworld\ntest")

	tests := []struct {
		pos      int
		wantLine int
		wantCol  int
	}{
		{0, 0, 0},   // Start of first line
		{5, 0, 5},   // End of first line (before newline)
		{6, 1, 0},   // Start of second line
		{11, 1, 5},  // End of second line
		{12, 2, 0},  // Start of third line
		{16, 2, 4},  // End of buffer
		{-1, 0, 0},  // Negative (clamped)
		{100, 2, 4}, // Beyond end (clamped)
	}

	for _, tt := range tests {
		line, col := b.PositionToLineCol(tt.pos)
		if line != tt.wantLine || col != tt.wantCol {
			t.Errorf("PositionToLineCol(%d) = (%d, %d), want (%d, %d)",
				tt.pos, line, col, tt.wantLine, tt.wantCol)
		}
	}
}

func TestBufferLineColToPosition(t *testing.T) {
	b := NewBufferFromString("hello\nworld\ntest")

	tests := []struct {
		line, col int
		want      int
	}{
		{0, 0, 0},
		{0, 5, 5},
		{1, 0, 6},
		{1, 5, 11},
		{2, 0, 12},
		{2, 4, 16},
		{0, 100, 5},  // Column beyond line end
		{-1, 0, 0},   // Negative line
		{100, 0, 16}, // Line beyond end
	}

	for _, tt := range tests {
		got := b.LineColToPosition(tt.line, tt.col)
		if got != tt.want {
			t.Errorf("LineColToPosition(%d, %d) = %d, want %d",
				tt.line, tt.col, got, tt.want)
		}
	}
}

func TestBufferLineStartEndOffset(t *testing.T) {
	b := NewBufferFromString("hello\nworld\ntest")

	tests := []struct {
		line      int
		wantStart int
		wantEnd   int
	}{
		{0, 0, 5},
		{1, 6, 11},
		{2, 12, 16},
	}

	for _, tt := range tests {
		start := b.LineStartOffset(tt.line)
		end := b.LineEndOffset(tt.line)
		if start != tt.wantStart {
			t.Errorf("LineStartOffset(%d) = %d, want %d", tt.line, start, tt.wantStart)
		}
		if end != tt.wantEnd {
			t.Errorf("LineEndOffset(%d) = %d, want %d", tt.line, end, tt.wantEnd)
		}
	}
}

func TestBufferReplace(t *testing.T) {
	b := NewBufferFromString("hello world")

	b.Replace(6, 11, "there")
	if got := b.String(); got != "hello there" {
		t.Errorf("after Replace(6, 11, 'there'), String() = %q, want 'hello there'", got)
	}

	b.Replace(0, 5, "hi")
	if got := b.String(); got != "hi there" {
		t.Errorf("after Replace(0, 5, 'hi'), String() = %q, want 'hi there'", got)
	}

	// Replace with swapped indices
	b.Replace(3, 0, "OH ")
	if got := b.String(); got != "OH there" {
		t.Errorf("after Replace(3, 0, 'OH '), String() = %q, want 'OH there'", got)
	}
}

func TestBufferRuneCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 5},
		{"hello🎉", 6}, // 5 ASCII + 1 emoji
		{"日本語", 3},
		{"🎉🎊🎁", 3},
	}

	for _, tt := range tests {
		b := NewBufferFromString(tt.input)
		if got := b.RuneCount(); got != tt.want {
			t.Errorf("RuneCount() for %q = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBufferWordCount(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{"  hello   world  ", 2},
		{"one\ntwo\nthree", 3},
		{"word1\t\tword2", 2},
		{"日本語 テスト", 2}, // Japanese words
	}

	for _, tt := range tests {
		b := NewBufferFromString(tt.input)
		if got := b.WordCount(); got != tt.want {
			t.Errorf("WordCount() for %q = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestBufferGapExpansion(t *testing.T) {
	b := NewBuffer()

	// Insert a large amount of text to trigger gap expansion
	largeText := strings.Repeat("x", 10000)
	b.Insert(largeText)

	if got := b.Length(); got != 10000 {
		t.Errorf("Length() after large insert = %d, want 10000", got)
	}
	if got := b.String(); got != largeText {
		t.Error("String() after large insert doesn't match")
	}
}

func TestBufferCursorMovementWithGap(t *testing.T) {
	// Test that gap movement preserves content
	b := NewBufferFromString("abcdefghij")

	// Move cursor around and verify content stays intact
	positions := []int{5, 0, 10, 3, 7, 1, 9}
	for _, pos := range positions {
		b.MoveCursor(pos)
		if got := b.String(); got != "abcdefghij" {
			t.Errorf("String() after MoveCursor(%d) = %q, want 'abcdefghij'", pos, got)
		}
	}
}
