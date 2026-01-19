package syntax

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

// SyntaxColors holds the color settings for syntax highlighting
type SyntaxColors struct {
	Keyword  string
	String   string
	Comment  string
	Number   string
	Operator string
	Function string
	Type     string
	Error    string
}

// DefaultSyntaxColors returns the default syntax color settings
func DefaultSyntaxColors() SyntaxColors {
	return SyntaxColors{
		Keyword:  "14", // Bright cyan
		String:   "10", // Bright green
		Comment:  "8",  // Gray
		Number:   "11", // Bright yellow
		Operator: "13", // Bright magenta
		Function: "12", // Bright blue
		Type:     "11", // Bright yellow
		Error:    "9",  // Bright red
	}
}

// ColorSpan represents a colored region of text
type ColorSpan struct {
	Start int    // Start column (rune index)
	End   int    // End column (rune index, exclusive)
	Color string // ANSI color code
}

// Highlighter provides syntax highlighting for source code
type Highlighter struct {
	lexer   chroma.Lexer
	enabled bool
	colors  SyntaxColors
}

// New creates a new Highlighter for the given filename
func New(filename string) *Highlighter {
	h := &Highlighter{
		enabled: true,
		colors:  DefaultSyntaxColors(),
	}
	h.SetFile(filename)
	return h
}

// SetFile updates the lexer based on the filename
func (h *Highlighter) SetFile(filename string) {
	if filename == "" {
		h.lexer = nil
		return
	}
	h.lexer = lexers.Match(filename)
	if h.lexer != nil {
		h.lexer = chroma.Coalesce(h.lexer)
	}
}

// SetEnabled enables or disables syntax highlighting
func (h *Highlighter) SetEnabled(enabled bool) {
	h.enabled = enabled
}

// Enabled returns whether highlighting is enabled
func (h *Highlighter) Enabled() bool {
	return h.enabled
}

// HasLexer returns true if a lexer is available for the current file
func (h *Highlighter) HasLexer() bool {
	return h.lexer != nil
}

// SetColors sets the syntax highlighting colors
func (h *Highlighter) SetColors(colors SyntaxColors) {
	h.colors = colors
}

// GetLineColors returns color spans for a line
// Returns nil if highlighting is disabled or no lexer is available
func (h *Highlighter) GetLineColors(line string) []ColorSpan {
	if !h.enabled || h.lexer == nil {
		return nil
	}

	iterator, err := h.lexer.Tokenise(nil, line)
	if err != nil {
		return nil
	}

	var spans []ColorSpan
	pos := 0
	for _, token := range iterator.Tokens() {
		color := h.tokenColor(token.Type)
		tokenLen := utf8.RuneCountInString(token.Value)
		if color != "" && tokenLen > 0 {
			spans = append(spans, ColorSpan{
				Start: pos,
				End:   pos + tokenLen,
				Color: color,
			})
		}
		pos += tokenLen
	}

	return spans
}

// ColorAt returns the color for a specific column position
// Returns empty string if no color applies
func ColorAt(spans []ColorSpan, col int) string {
	for _, span := range spans {
		if col >= span.Start && col < span.End {
			return span.Color
		}
	}
	return ""
}

// colorToANSI converts a theme color string to an ANSI foreground escape sequence
func colorToANSI(color string) string {
	if strings.HasPrefix(color, "#") {
		r, g, b := parseHexColor(color)
		return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
	}
	n, err := strconv.Atoi(color)
	if err != nil {
		return "\033[37m" // Default to white on error
	}
	if n < 16 {
		if n < 8 {
			return fmt.Sprintf("\033[%dm", 30+n)
		}
		return fmt.Sprintf("\033[%dm", 90+(n-8))
	}
	return fmt.Sprintf("\033[38;5;%dm", n)
}

// parseHexColor parses #RGB or #RRGGBB to r, g, b values
func parseHexColor(hex string) (int, int, int) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		r, _ := strconv.ParseInt(string(hex[0])+string(hex[0]), 16, 32)
		g, _ := strconv.ParseInt(string(hex[1])+string(hex[1]), 16, 32)
		b, _ := strconv.ParseInt(string(hex[2])+string(hex[2]), 16, 32)
		return int(r), int(g), int(b)
	}
	if len(hex) == 6 {
		r, _ := strconv.ParseInt(hex[0:2], 16, 32)
		g, _ := strconv.ParseInt(hex[2:4], 16, 32)
		b, _ := strconv.ParseInt(hex[4:6], 16, 32)
		return int(r), int(g), int(b)
	}
	return 255, 255, 255 // Default to white on error
}

// tokenColor returns the ANSI color code for a token type
func (h *Highlighter) tokenColor(t chroma.TokenType) string {
	switch {
	// Keywords
	case t == chroma.Keyword,
		t == chroma.KeywordConstant,
		t == chroma.KeywordDeclaration,
		t == chroma.KeywordNamespace,
		t == chroma.KeywordPseudo,
		t == chroma.KeywordReserved,
		t == chroma.KeywordType:
		return colorToANSI(h.colors.Keyword)

	// Strings
	case t == chroma.String,
		t == chroma.StringAffix,
		t == chroma.StringBacktick,
		t == chroma.StringChar,
		t == chroma.StringDelimiter,
		t == chroma.StringDoc,
		t == chroma.StringDouble,
		t == chroma.StringEscape,
		t == chroma.StringHeredoc,
		t == chroma.StringInterpol,
		t == chroma.StringOther,
		t == chroma.StringRegex,
		t == chroma.StringSingle,
		t == chroma.StringSymbol:
		return colorToANSI(h.colors.String)

	// Comments
	case t == chroma.Comment,
		t == chroma.CommentHashbang,
		t == chroma.CommentMultiline,
		t == chroma.CommentPreproc,
		t == chroma.CommentPreprocFile,
		t == chroma.CommentSingle,
		t == chroma.CommentSpecial:
		return colorToANSI(h.colors.Comment)

	// Numbers
	case t == chroma.Number,
		t == chroma.NumberBin,
		t == chroma.NumberFloat,
		t == chroma.NumberHex,
		t == chroma.NumberInteger,
		t == chroma.NumberIntegerLong,
		t == chroma.NumberOct:
		return colorToANSI(h.colors.Number)

	// Operators
	case t == chroma.Operator,
		t == chroma.OperatorWord:
		return colorToANSI(h.colors.Operator)

	// Functions
	case t == chroma.NameFunction,
		t == chroma.NameFunctionMagic:
		return colorToANSI(h.colors.Function)

	// Types/Classes
	case t == chroma.NameClass,
		t == chroma.NameBuiltin,
		t == chroma.NameBuiltinPseudo:
		return colorToANSI(h.colors.Type)

	// Constants
	case t == chroma.NameConstant:
		return colorToANSI(h.colors.Number) // Same as numbers

	// Preprocessor
	case t == chroma.GenericHeading,
		t == chroma.GenericSubheading:
		return colorToANSI(h.colors.Type)

	// Errors
	case t == chroma.Error,
		t == chroma.GenericError:
		return colorToANSI(h.colors.Error)

	default:
		return "" // Default terminal color
	}
}
