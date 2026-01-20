// Package encoding provides character encoding detection and conversion.
package encoding

import (
	"bytes"
	"io"
	"strings"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Encoding represents a character encoding with metadata
type Encoding struct {
	Name        string            // Display name
	ID          string            // Internal identifier
	Encoder     encoding.Encoding // x/text encoder (nil for UTF-8)
	Aliases     []string          // Alternative names from chardet
	Supported   bool              // Whether we support encoding/decoding
	Description string            // Brief description
}

// DetectionResult holds the result of encoding detection
type DetectionResult struct {
	Encoding   *Encoding
	Confidence int  // 0-100
	HasBOM     bool // Whether a BOM was detected
}

// SupportedEncodings is the list of encodings we fully support
var SupportedEncodings = []*Encoding{
	{
		Name:        "UTF-8",
		ID:          "utf-8",
		Encoder:     nil, // Native, no conversion needed
		Aliases:     []string{"UTF-8", "utf8"},
		Supported:   true,
		Description: "Unicode (default)",
	},
	{
		Name:        "UTF-8 BOM",
		ID:          "utf-8-bom",
		Encoder:     nil,
		Aliases:     []string{},
		Supported:   true,
		Description: "Unicode with byte order mark",
	},
	{
		Name:        "UTF-16 LE",
		ID:          "utf-16-le",
		Encoder:     unicode.UTF16(unicode.LittleEndian, unicode.UseBOM),
		Aliases:     []string{"UTF-16LE"},
		Supported:   true,
		Description: "Unicode 16-bit (Little Endian)",
	},
	{
		Name:        "UTF-16 BE",
		ID:          "utf-16-be",
		Encoder:     unicode.UTF16(unicode.BigEndian, unicode.UseBOM),
		Aliases:     []string{"UTF-16BE"},
		Supported:   true,
		Description: "Unicode 16-bit (Big Endian)",
	},
	{
		Name:        "ISO-8859-1",
		ID:          "iso-8859-1",
		Encoder:     charmap.ISO8859_1,
		Aliases:     []string{"ISO-8859-1", "latin1", "Latin-1"},
		Supported:   true,
		Description: "Western European (Latin-1)",
	},
	{
		Name:        "Windows-1252",
		ID:          "windows-1252",
		Encoder:     charmap.Windows1252,
		Aliases:     []string{"windows-1252", "CP1252"},
		Supported:   true,
		Description: "Western European (Windows)",
	},
	{
		Name:        "ISO-8859-15",
		ID:          "iso-8859-15",
		Encoder:     charmap.ISO8859_15,
		Aliases:     []string{"ISO-8859-15", "latin9", "Latin-9"},
		Supported:   true,
		Description: "Western European (Latin-9, with €)",
	},
	{
		Name:        "Shift-JIS",
		ID:          "shift-jis",
		Encoder:     japanese.ShiftJIS,
		Aliases:     []string{"Shift_JIS", "SJIS", "MS_Kanji"},
		Supported:   true,
		Description: "Japanese",
	},
	{
		Name:        "EUC-JP",
		ID:          "euc-jp",
		Encoder:     japanese.EUCJP,
		Aliases:     []string{"EUC-JP"},
		Supported:   true,
		Description: "Japanese (Unix)",
	},
	{
		Name:        "GBK",
		ID:          "gbk",
		Encoder:     simplifiedchinese.GBK,
		Aliases:     []string{"GBK", "GB2312", "GB-2312"},
		Supported:   true,
		Description: "Simplified Chinese",
	},
	{
		Name:        "GB18030",
		ID:          "gb18030",
		Encoder:     simplifiedchinese.GB18030,
		Aliases:     []string{"GB18030"},
		Supported:   true,
		Description: "Simplified Chinese (extended)",
	},
	{
		Name:        "EUC-KR",
		ID:          "euc-kr",
		Encoder:     korean.EUCKR,
		Aliases:     []string{"EUC-KR"},
		Supported:   true,
		Description: "Korean",
	},
}

// UTF8 BOM bytes
var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// UTF-16 BOMs
var utf16LEBOM = []byte{0xFF, 0xFE}
var utf16BEBOM = []byte{0xFE, 0xFF}

// GetEncodingByID returns an encoding by its ID
func GetEncodingByID(id string) *Encoding {
	id = strings.ToLower(id)
	for _, enc := range SupportedEncodings {
		if strings.ToLower(enc.ID) == id {
			return enc
		}
	}
	return nil
}

// GetEncodingByName returns an encoding by name or alias
func GetEncodingByName(name string) *Encoding {
	name = strings.ToLower(name)
	for _, enc := range SupportedEncodings {
		if strings.ToLower(enc.Name) == name || strings.ToLower(enc.ID) == name {
			return enc
		}
		for _, alias := range enc.Aliases {
			if strings.ToLower(alias) == name {
				return enc
			}
		}
	}
	return nil
}

// Detect attempts to detect the encoding of the given data
func Detect(data []byte) *DetectionResult {
	result := &DetectionResult{
		Encoding:   GetEncodingByID("utf-8"),
		Confidence: 100,
		HasBOM:     false,
	}

	// Check for BOMs first (most reliable)
	if bytes.HasPrefix(data, utf8BOM) {
		result.Encoding = GetEncodingByID("utf-8-bom")
		result.HasBOM = true
		return result
	}
	if bytes.HasPrefix(data, utf16BEBOM) {
		result.Encoding = GetEncodingByID("utf-16-be")
		result.HasBOM = true
		return result
	}
	if bytes.HasPrefix(data, utf16LEBOM) {
		result.Encoding = GetEncodingByID("utf-16-le")
		result.HasBOM = true
		return result
	}

	// Check if valid UTF-8
	if isValidUTF8(data) {
		result.Encoding = GetEncodingByID("utf-8")
		result.Confidence = 100
		return result
	}

	// Use chardet for detection
	detector := chardet.NewTextDetector()
	detected, err := detector.DetectBest(data)
	if err != nil || detected == nil {
		// Fall back to Latin-1 (always valid)
		result.Encoding = GetEncodingByID("iso-8859-1")
		result.Confidence = 50
		return result
	}

	// Map chardet result to our encoding
	enc := GetEncodingByName(detected.Charset)
	if enc != nil {
		result.Encoding = enc
		result.Confidence = detected.Confidence
	} else {
		// Unsupported encoding detected
		result.Encoding = &Encoding{
			Name:      detected.Charset,
			ID:        strings.ToLower(detected.Charset),
			Supported: false,
		}
		result.Confidence = detected.Confidence
	}

	return result
}

// isValidUTF8 checks if data is valid UTF-8
func isValidUTF8(data []byte) bool {
	// Check for invalid UTF-8 sequences
	for i := 0; i < len(data); {
		if data[i] < 0x80 {
			// ASCII
			i++
			continue
		}

		// Multi-byte sequence
		var size int
		switch {
		case data[i]&0xE0 == 0xC0:
			size = 2
		case data[i]&0xF0 == 0xE0:
			size = 3
		case data[i]&0xF8 == 0xF0:
			size = 4
		default:
			return false
		}

		if i+size > len(data) {
			return false
		}

		// Check continuation bytes
		for j := 1; j < size; j++ {
			if data[i+j]&0xC0 != 0x80 {
				return false
			}
		}

		i += size
	}
	return true
}

// DecodeToUTF8 decodes data from the given encoding to UTF-8
func DecodeToUTF8(data []byte, enc *Encoding) ([]byte, error) {
	if enc == nil || enc.Encoder == nil {
		// Already UTF-8 or UTF-8 BOM
		if enc != nil && enc.ID == "utf-8-bom" && bytes.HasPrefix(data, utf8BOM) {
			return data[3:], nil // Strip BOM
		}
		return data, nil
	}

	// Handle UTF-16 BOMs
	if enc.ID == "utf-16-le" && bytes.HasPrefix(data, utf16LEBOM) {
		data = data[2:]
	} else if enc.ID == "utf-16-be" && bytes.HasPrefix(data, utf16BEBOM) {
		data = data[2:]
	}

	reader := transform.NewReader(bytes.NewReader(data), enc.Encoder.NewDecoder())
	return io.ReadAll(reader)
}

// EncodeFromUTF8 encodes UTF-8 data to the given encoding
func EncodeFromUTF8(data []byte, enc *Encoding) ([]byte, error) {
	if enc == nil || enc.Encoder == nil {
		// UTF-8 or UTF-8 BOM
		if enc != nil && enc.ID == "utf-8-bom" {
			return append(utf8BOM, data...), nil
		}
		return data, nil
	}

	var buf bytes.Buffer

	// Add BOM for UTF-16
	if enc.ID == "utf-16-le" {
		buf.Write(utf16LEBOM)
	} else if enc.ID == "utf-16-be" {
		buf.Write(utf16BEBOM)
	}

	writer := transform.NewWriter(&buf, enc.Encoder.NewEncoder())
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetSupportedEncodings returns all supported encodings
func GetSupportedEncodings() []*Encoding {
	return SupportedEncodings
}
