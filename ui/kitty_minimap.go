package ui

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/cornish/textivus-editor/syntax"
)

// KittyMinimapRenderer renders a pixel-based minimap using Kitty graphics protocol.
// This provides a VSCode-like minimap with syntax highlighting colors.
//
// Kitty graphics protocol reference:
// https://sw.kovidgoyal.net/kitty/graphics-protocol/
//
// Design (VSCode-style):
//   - Each source character = 1 pixel wide
//   - Each source line = 2 pixels tall (for readability)
//   - Syntax highlighting colors preserved from the highlighter
//   - Viewport indicator = dark gray semi-transparent overlay on visible region
//   - Maximum 120 characters width shown (truncated, not scaled)
type KittyMinimapRenderer struct {
	styles         Styles
	enabled        bool
	useKitty       bool // Whether to use Kitty graphics (vs falling back to braille)
	imageID        uint32
	lineColors     func(line string) []syntax.ColorSpan // Syntax highlighter callback
	lastStartLine  int                                  // First line shown in last render (for click handling)
	lastLinesShown int                                  // Number of lines shown in last render
}

// NewKittyMinimapRenderer creates a new Kitty graphics minimap renderer.
func NewKittyMinimapRenderer(styles Styles, useKitty bool) *KittyMinimapRenderer {
	return &KittyMinimapRenderer{
		styles:   styles,
		enabled:  false,
		useKitty: useKitty,
		imageID:  1001, // Fixed ID for minimap image
	}
}

// SetStyles updates the styles for runtime theme changes.
func (r *KittyMinimapRenderer) SetStyles(styles Styles) {
	r.styles = styles
}

// SetEnabled enables or disables the minimap.
func (r *KittyMinimapRenderer) SetEnabled(enabled bool) {
	r.enabled = enabled
}

// IsEnabled returns whether the minimap is enabled.
func (r *KittyMinimapRenderer) IsEnabled() bool {
	return r.enabled
}

// Toggle toggles the minimap on/off.
func (r *KittyMinimapRenderer) Toggle() bool {
	r.enabled = !r.enabled
	return r.enabled
}

// SetUseKitty enables or disables Kitty graphics mode.
func (r *KittyMinimapRenderer) SetUseKitty(useKitty bool) {
	r.useKitty = useKitty
}

// UseKitty returns whether Kitty graphics mode is active.
func (r *KittyMinimapRenderer) UseKitty() bool {
	return r.useKitty
}

// SetLineColorFunc sets the callback for getting syntax colors for a line.
func (r *KittyMinimapRenderer) SetLineColorFunc(fn func(line string) []syntax.ColorSpan) {
	r.lineColors = fn
}

// Pixel dimensions for the minimap image
const (
	kittyPixelsPerChar  = 1   // 1 pixel per source character
	kittyPixelsPerLine  = 2   // 2 pixels per source line (for visibility)
	kittyMinimapWidth   = 120 // Max source characters shown
	kittyIndicatorAlpha = 80  // Viewport indicator overlay alpha (0-255)
)

// Render implements ColumnRenderer.
// Returns blank spaces for the column area. The actual Kitty graphics
// is rendered separately via GetKittySequence() and appended to View() output.
func (r *KittyMinimapRenderer) Render(width, height int, state *RenderState) []string {
	if !r.enabled || width <= 0 || height <= 0 || state == nil {
		rows := make([]string, height)
		for i := range rows {
			rows[i] = strings.Repeat(" ", width)
		}
		return rows
	}

	if !r.useKitty {
		// Fall back to braille rendering
		return r.renderBrailleFallback(width, height, state)
	}

	// For Kitty graphics, return blank spaces - the actual image
	// is rendered via GetKittySequence() after the main View() output
	rows := make([]string, height)
	for i := range rows {
		rows[i] = strings.Repeat(" ", width)
	}
	return rows
}

// GetKittySequence returns the Kitty graphics escape sequence to render the minimap.
// This should be appended to the View() output AFTER all normal rendering,
// with cursor positioning to place it at the minimap column location.
// Returns empty string if Kitty graphics is not enabled.
func (r *KittyMinimapRenderer) GetKittySequence(width, height, xOffset, yOffset int, state *RenderState) string {
	if !r.enabled || !r.useKitty || state == nil {
		return ""
	}

	totalLines := len(state.Lines)
	if totalLines == 0 {
		totalLines = 1
	}

	// Image dimensions - fit as many lines as possible in the terminal area
	// Terminal cells are approximately 8x16 pixels (width x height)
	cellPixelHeight := 16
	imgPixelHeight := height * cellPixelHeight
	imgWidth := kittyMinimapWidth

	// Calculate how many source lines can fit in the minimap
	// Each source line = kittyPixelsPerLine pixels
	maxLinesInMinimap := imgPixelHeight / kittyPixelsPerLine

	// Calculate which lines to show (like braille - scroll to keep viewport visible)
	startLine := 0
	if totalLines > maxLinesInMinimap {
		// Center the minimap view on the current viewport
		viewportCenter := state.ScrollY + height/2
		startLine = viewportCenter - maxLinesInMinimap/2
		if startLine < 0 {
			startLine = 0
		}
		if startLine+maxLinesInMinimap > totalLines {
			startLine = totalLines - maxLinesInMinimap
		}
	}
	endLine := startLine + maxLinesInMinimap
	if endLine > totalLines {
		endLine = totalLines
	}

	linesShown := endLine - startLine
	if linesShown < 1 {
		linesShown = 1
	}

	// Always use full image height to avoid Kitty scaling artifacts
	// For short files, the extra space will just be background color
	imgHeight := imgPixelHeight

	// Store startLine for click handling
	r.lastStartLine = startLine
	r.lastLinesShown = linesShown

	// Generate pixel data - pass viewport height for correct highlight
	pixels := r.generatePixelDataWithSyntax(imgWidth, imgHeight, startLine, endLine, height, state)

	// Build the escape sequence with cursor positioning
	var sb strings.Builder

	// Save cursor position
	sb.WriteString("\033[s")

	// Move cursor to minimap position (1-indexed)
	sb.WriteString(fmt.Sprintf("\033[%d;%dH", yOffset+1, xOffset+1))

	// Send Kitty graphics
	sb.WriteString(r.encodeKittyGraphics(pixels, imgWidth, imgHeight, width, height))

	// Restore cursor position
	sb.WriteString("\033[u")

	return sb.String()
}

// renderKittyGraphics generates a VSCode-style minimap using Kitty graphics protocol.
func (r *KittyMinimapRenderer) renderKittyGraphics(width, height int, state *RenderState) []string {
	rows := make([]string, height)

	// Image dimensions:
	// - Width: kittyMinimapWidth pixels (1 pixel per source character, max 120)
	// - Height: total lines * kittyPixelsPerLine (2 pixels per line)
	// But we need to fit within the terminal cell area (width x height cells)

	totalLines := len(state.Lines)
	if totalLines == 0 {
		totalLines = 1
	}

	// Calculate how many lines we can show in the viewport
	// Each terminal row = ~16 pixels typically, each source line = 2 pixels
	// So we can show about 8 source lines per terminal row
	pixelsPerTermRow := 16 // Approximate pixels per terminal row
	maxLinesInView := height * (pixelsPerTermRow / kittyPixelsPerLine)

	// Image pixel dimensions
	imgWidth := kittyMinimapWidth
	imgHeight := totalLines * kittyPixelsPerLine
	if imgHeight > height*pixelsPerTermRow {
		imgHeight = height * pixelsPerTermRow
	}
	if imgHeight < kittyPixelsPerLine {
		imgHeight = kittyPixelsPerLine
	}

	// Calculate which lines to show (center on viewport)
	startLine := 0
	if totalLines > maxLinesInView {
		// Center the view on the current scroll position
		centerLine := state.ScrollY + height/2
		startLine = centerLine - maxLinesInView/2
		if startLine < 0 {
			startLine = 0
		}
		if startLine+maxLinesInView > totalLines {
			startLine = totalLines - maxLinesInView
		}
	}
	endLine := startLine + maxLinesInView
	if endLine > totalLines {
		endLine = totalLines
	}

	// Recalculate image height based on actual lines shown
	linesShown := endLine - startLine
	if linesShown < 1 {
		linesShown = 1
	}
	imgHeight = linesShown * kittyPixelsPerLine

	// Generate the pixel data with syntax colors
	pixels := r.generatePixelDataWithSyntax(imgWidth, imgHeight, startLine, endLine, height, state)

	// Encode as Kitty graphics sequence
	kittySeq := r.encodeKittyGraphics(pixels, imgWidth, imgHeight, width, height)

	// First row contains the Kitty sequence + padding spaces
	rows[0] = kittySeq + strings.Repeat(" ", width)

	// Remaining rows are just spaces (the image overlays them)
	for i := 1; i < height; i++ {
		rows[i] = strings.Repeat(" ", width)
	}

	return rows
}

// generatePixelDataWithSyntax creates RGBA pixel data with actual syntax highlighting colors.
func (r *KittyMinimapRenderer) generatePixelDataWithSyntax(imgWidth, imgHeight, startLine, endLine, viewportHeight int, state *RenderState) []byte {
	// Use RGBA for alpha blending (viewport indicator overlay)
	pixels := make([]byte, imgWidth*imgHeight*4) // RGBA

	// Use a dark background for the minimap (VS Code style)
	// This ensures text of any color is visible
	bgColor := [3]byte{30, 30, 30} // Dark gray background

	// Default text color for unstyled text (light gray)
	defaultTextColor := [3]byte{180, 180, 180}

	// Viewport highlight color (darker gray, semi-transparent overlay)
	viewportHighlight := [3]byte{80, 80, 80}

	// Fill background (fully opaque)
	for i := 0; i < len(pixels); i += 4 {
		pixels[i] = bgColor[0]
		pixels[i+1] = bgColor[1]
		pixels[i+2] = bgColor[2]
		pixels[i+3] = 255 // Fully opaque
	}

	// Viewport range (in source lines) - use actual viewport height
	viewportStart := state.ScrollY
	viewportEnd := state.ScrollY + viewportHeight

	// Render each source line
	for lineIdx := startLine; lineIdx < endLine && lineIdx < len(state.Lines); lineIdx++ {
		line := state.Lines[lineIdx]

		// Get syntax colors for this line
		var colors []syntax.ColorSpan
		if state.LineColors != nil {
			colors = state.LineColors[lineIdx]
		}

		// Calculate pixel row for this line (relative to image)
		relativeLineIdx := lineIdx - startLine
		pyStart := relativeLineIdx * kittyPixelsPerLine
		pyEnd := pyStart + kittyPixelsPerLine

		// Check if this line is in the visible viewport
		inViewport := lineIdx >= viewportStart && lineIdx < viewportEnd

		// Apply viewport highlight FIRST (as background tint)
		if inViewport {
			for py := pyStart; py < pyEnd && py < imgHeight; py++ {
				for px := 0; px < imgWidth; px++ {
					idx := (py*imgWidth + px) * 4
					// Set to viewport highlight color
					pixels[idx] = viewportHighlight[0]
					pixels[idx+1] = viewportHighlight[1]
					pixels[idx+2] = viewportHighlight[2]
				}
			}
		}

		// Render each character on top, accounting for tab width
		tabWidth := state.TabWidth
		if tabWidth <= 0 {
			tabWidth = 4
		}
		runes := []rune(line)
		visualCol := 0 // Track visual column position
		for runeIdx, ru := range runes {
			if visualCol >= imgWidth {
				break // Truncate long lines
			}

			if ru == '\t' {
				// Tab advances to next multiple of tabWidth (or at least tabWidth spaces)
				visualCol += tabWidth - (visualCol % tabWidth)
				continue
			}

			if ru == ' ' {
				visualCol++
				continue
			}

			// Get color for this character (use rune index for syntax lookup)
			charColor := defaultTextColor
			if colors != nil {
				ansiColor := syntax.ColorAt(colors, runeIdx)
				if ansiColor != "" {
					// Parse ANSI color to RGB
					parsed := parseANSIToRGB(ansiColor)
					// Use parsed color if it's not black (black means parsing failed)
					if parsed[0] != 0 || parsed[1] != 0 || parsed[2] != 0 {
						charColor = parsed
					}
				}
			}

			// Draw pixel at visual column position
			if visualCol < imgWidth {
				for py := pyStart; py < pyEnd && py < imgHeight; py++ {
					idx := (py*imgWidth + visualCol) * 4
					pixels[idx] = charColor[0]
					pixels[idx+1] = charColor[1]
					pixels[idx+2] = charColor[2]
					pixels[idx+3] = 255
				}
			}
			visualCol++
		}
	}

	return pixels
}

// blendColor blends two color values with the given alpha.
func blendColor(base, overlay byte, alpha float64) byte {
	result := float64(base)*(1-alpha) + float64(overlay)*alpha
	if result > 255 {
		return 255
	}
	if result < 0 {
		return 0
	}
	return byte(result)
}

// parseANSIToRGB attempts to parse an ANSI color escape sequence to RGB.
func parseANSIToRGB(ansi string) [3]byte {
	// Common ANSI 256-color format: \033[38;5;XXXm or \033[38;2;R;G;Bm
	// Try to extract RGB values

	// Check for true color (24-bit): \033[38;2;R;G;Bm
	var r, g, b int
	if n, _ := fmt.Sscanf(ansi, "\033[38;2;%d;%d;%dm", &r, &g, &b); n == 3 {
		return [3]byte{byte(r), byte(g), byte(b)}
	}

	// Check for 256-color: \033[38;5;XXXm
	var colorIdx int
	if n, _ := fmt.Sscanf(ansi, "\033[38;5;%dm", &colorIdx); n == 1 {
		return ansi256ToRGB(colorIdx)
	}

	// Basic ANSI colors (30-37, 90-97)
	var code int
	if n, _ := fmt.Sscanf(ansi, "\033[%dm", &code); n == 1 {
		return ansiBasicToRGB(code)
	}

	return [3]byte{200, 200, 200} // Default gray
}

// ansi256ToRGB converts a 256-color palette index to RGB.
func ansi256ToRGB(idx int) [3]byte {
	if idx < 16 {
		// Standard colors
		return ansiBasicToRGB(idx + 30)
	}
	if idx < 232 {
		// 6x6x6 color cube
		idx -= 16
		r := (idx / 36) * 51
		g := ((idx / 6) % 6) * 51
		b := (idx % 6) * 51
		return [3]byte{byte(r), byte(g), byte(b)}
	}
	// Grayscale
	gray := (idx-232)*10 + 8
	return [3]byte{byte(gray), byte(gray), byte(gray)}
}

// ansiBasicToRGB converts basic ANSI color codes to RGB.
func ansiBasicToRGB(code int) [3]byte {
	// Map basic ANSI colors to approximate RGB
	colors := map[int][3]byte{
		30: {0, 0, 0},       // Black
		31: {205, 49, 49},   // Red
		32: {13, 188, 121},  // Green
		33: {229, 229, 16},  // Yellow
		34: {36, 114, 200},  // Blue
		35: {188, 63, 188},  // Magenta
		36: {17, 168, 205},  // Cyan
		37: {229, 229, 229}, // White
		90: {102, 102, 102}, // Bright Black
		91: {241, 76, 76},   // Bright Red
		92: {35, 209, 139},  // Bright Green
		93: {245, 245, 67},  // Bright Yellow
		94: {59, 142, 234},  // Bright Blue
		95: {214, 112, 214}, // Bright Magenta
		96: {41, 184, 219},  // Bright Cyan
		97: {255, 255, 255}, // Bright White
	}
	if rgb, ok := colors[code]; ok {
		return rgb
	}
	return [3]byte{200, 200, 200}
}

// encodeKittyGraphics creates the Kitty graphics protocol escape sequence.
// Format: \033_G<control>;base64data\033\\
func (r *KittyMinimapRenderer) encodeKittyGraphics(pixels []byte, imgWidth, imgHeight, cellCols, cellRows int) string {
	// Base64 encode the pixel data
	b64Data := base64.StdEncoding.EncodeToString(pixels)

	// Build the control string
	// a=T: transmit and display
	// f=32: RGBA format (4 bytes per pixel)
	// s=width, v=height: pixel dimensions
	// c=cols, r=rows: cell dimensions to occupy
	// i=id: image ID for updates
	// q=2: suppress response
	control := fmt.Sprintf("a=T,f=32,s=%d,v=%d,c=%d,r=%d,i=%d,q=2",
		imgWidth, imgHeight, cellCols, cellRows, r.imageID)

	// Kitty allows chunked transmission for large images
	// For simplicity, send in chunks of 4096 bytes
	const chunkSize = 4096
	var sb strings.Builder

	if len(b64Data) <= chunkSize {
		// Single chunk - use m=0 (no more data)
		sb.WriteString(fmt.Sprintf("\033_G%s;%s\033\\", control, b64Data))
	} else {
		// Multiple chunks
		for i := 0; i < len(b64Data); i += chunkSize {
			end := i + chunkSize
			if end > len(b64Data) {
				end = len(b64Data)
			}
			chunk := b64Data[i:end]

			if i == 0 {
				// First chunk - include control data, m=1 (more data follows)
				sb.WriteString(fmt.Sprintf("\033_G%s,m=1;%s\033\\", control, chunk))
			} else if end >= len(b64Data) {
				// Last chunk - m=0 (no more data)
				sb.WriteString(fmt.Sprintf("\033_Gm=0;%s\033\\", chunk))
			} else {
				// Middle chunk - m=1 (more data follows)
				sb.WriteString(fmt.Sprintf("\033_Gm=1;%s\033\\", chunk))
			}
		}
	}

	return sb.String()
}

// generateVisualLines converts buffer lines to visual lines respecting word wrap.
func (r *KittyMinimapRenderer) generateVisualLines(lines []string, wordWrap bool, textWidth int) []string {
	if !wordWrap || textWidth <= 0 {
		return lines
	}

	var visualLines []string
	for _, line := range lines {
		lineRunes := []rune(line)
		if len(lineRunes) == 0 {
			visualLines = append(visualLines, "")
			continue
		}
		for len(lineRunes) > 0 {
			end := textWidth
			if end > len(lineRunes) {
				end = len(lineRunes)
			}
			visualLines = append(visualLines, string(lineRunes[:end]))
			lineRunes = lineRunes[end:]
		}
	}
	if len(visualLines) == 0 {
		visualLines = []string{""}
	}
	return visualLines
}

// renderBrailleFallback renders using braille characters when Kitty is not available.
// This is a simplified version - delegates to the core braille logic.
func (r *KittyMinimapRenderer) renderBrailleFallback(width, height int, state *RenderState) []string {
	rows := make([]string, height)
	visualLineCount := 0

	// Layout: [indicator][braille chars][space]
	brailleWidth := width - 2
	if brailleWidth < 1 {
		brailleWidth = 1
	}

	// Generate visual lines
	textWidth := 80
	visualLines := r.generateVisualLines(state.Lines, state.WordWrap, textWidth)
	totalVisualLines := len(visualLines)
	if totalVisualLines == 0 {
		totalVisualLines = 1
		visualLines = []string{""}
	}

	minimapHeight := (totalVisualLines + 3) / 4

	// Viewport indicator range
	visibleStart := state.ScrollY
	visibleEnd := state.ScrollY + height

	// Get theme colors
	ui := r.styles.Theme.UI
	indicatorColor := ColorToANSIFg(ui.MinimapIndicator)
	textColor := ColorToANSIFg(ui.MinimapText)
	resetCode := "\033[0m"

	// Scroll offset if minimap is taller than viewport
	minimapScrollOffset := 0
	if minimapHeight > height {
		viewportCenterLine := state.ScrollY + height/2
		minimapCenterRow := viewportCenterLine / 4
		minimapScrollOffset = minimapCenterRow - height/2
		if minimapScrollOffset < 0 {
			minimapScrollOffset = 0
		}
		if minimapScrollOffset > minimapHeight-height {
			minimapScrollOffset = minimapHeight - height
		}
	}

	for row := 0; row < height; row++ {
		var sb strings.Builder
		minimapRow := row + minimapScrollOffset

		if minimapRow >= minimapHeight {
			sb.WriteString(strings.Repeat(" ", width))
			rows[row] = sb.String()
			continue
		}

		visualLineStart := minimapRow * 4
		visualLineEnd := visualLineStart + 4
		if visualLineEnd > totalVisualLines {
			visualLineEnd = totalVisualLines
		}

		inViewport := visualLineStart < visibleEnd && visualLineEnd > visibleStart
		if inViewport {
			sb.WriteString(indicatorColor)
			sb.WriteString("│")
			sb.WriteString(resetCode)
		} else {
			sb.WriteString(" ")
		}

		var fourLines [4]string
		for i := 0; i < 4; i++ {
			lineIdx := visualLineStart + i
			if lineIdx < totalVisualLines {
				fourLines[i] = visualLines[lineIdx]
			}
		}

		sb.WriteString(textColor)
		braille := renderBrailleChars(fourLines, brailleWidth)
		sb.WriteString(braille)
		sb.WriteString(resetCode)

		sb.WriteString(" ")
		rows[row] = sb.String()
		visualLineCount++
	}

	return rows
}

// hexToRGB converts a hex color string to RGB bytes.
func hexToRGB(hex string) [3]byte {
	var r, g, b byte
	if len(hex) == 7 && hex[0] == '#' {
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	} else if len(hex) == 4 && hex[0] == '#' {
		// Short form #RGB
		fmt.Sscanf(hex, "#%1x%1x%1x", &r, &g, &b)
		r, g, b = r*17, g*17, b*17
	}
	return [3]byte{r, g, b}
}

// renderBrailleChars renders braille characters for 4 visual lines.
func renderBrailleChars(fourLines [4]string, brailleWidth int) string {
	var result strings.Builder
	charsPerBraille := 10

	for col := 0; col < brailleWidth; col++ {
		srcColStart := col * charsPerBraille
		srcColMid := srcColStart + 5

		var pattern rune = 0x2800

		for rowOffset := 0; rowOffset < 4; rowOffset++ {
			lineRunes := []rune(fourLines[rowOffset])

			if hasEnoughContentLocal(lineRunes, srcColStart, srcColMid, 3) {
				switch rowOffset {
				case 0:
					pattern |= 0x01
				case 1:
					pattern |= 0x02
				case 2:
					pattern |= 0x04
				case 3:
					pattern |= 0x40
				}
			}

			if hasEnoughContentLocal(lineRunes, srcColMid, srcColMid+5, 3) {
				switch rowOffset {
				case 0:
					pattern |= 0x08
				case 1:
					pattern |= 0x10
				case 2:
					pattern |= 0x20
				case 3:
					pattern |= 0x80
				}
			}
		}

		result.WriteRune(pattern)
	}

	return result.String()
}

// hasEnoughContentLocal checks if a line has enough non-whitespace characters.
func hasEnoughContentLocal(lineRunes []rune, start, end, threshold int) bool {
	if start < 0 {
		start = 0
	}
	if end > len(lineRunes) {
		end = len(lineRunes)
	}
	count := 0
	for i := start; i < end; i++ {
		if i < len(lineRunes) {
			r := lineRunes[i]
			if r != ' ' && r != '\t' {
				count++
				if count >= threshold {
					return true
				}
			}
		}
	}
	return false
}

// GetMetrics calculates minimap metrics for mouse interaction.
func (r *KittyMinimapRenderer) GetMetrics(viewportHeight int, state *RenderState) MinimapMetrics {
	if r.useKitty {
		// For Kitty graphics, use the stored line info from last render
		totalLines := len(state.Lines)
		if totalLines == 0 {
			totalLines = 1
		}
		return MinimapMetrics{
			TotalVisualLines:    totalLines,
			MinimapHeight:       r.lastLinesShown,
			MinimapScrollOffset: r.lastStartLine,
			ViewportHeight:      viewportHeight,
		}
	}

	// Braille fallback metrics
	textWidth := 80
	visualLines := r.generateVisualLines(state.Lines, state.WordWrap, textWidth)
	totalVisualLines := len(visualLines)
	if totalVisualLines == 0 {
		totalVisualLines = 1
	}

	minimapHeight := (totalVisualLines + 3) / 4

	minimapScrollOffset := 0
	if minimapHeight > viewportHeight {
		viewportCenterLine := state.ScrollY + viewportHeight/2
		minimapCenterRow := viewportCenterLine / 4
		minimapScrollOffset = minimapCenterRow - viewportHeight/2
		if minimapScrollOffset < 0 {
			minimapScrollOffset = 0
		}
		if minimapScrollOffset > minimapHeight-viewportHeight {
			minimapScrollOffset = minimapHeight - viewportHeight
		}
	}

	return MinimapMetrics{
		TotalVisualLines:    totalVisualLines,
		MinimapHeight:       minimapHeight,
		MinimapScrollOffset: minimapScrollOffset,
		ViewportHeight:      viewportHeight,
	}
}

// RowToVisualLine converts a minimap row click to a visual line index.
func (r *KittyMinimapRenderer) RowToVisualLine(row int, metrics MinimapMetrics) int {
	if r.useKitty {
		// For Kitty graphics, each terminal row shows multiple source lines
		// Terminal cells are ~16 pixels tall, each source line is 2 pixels
		linesPerTermRow := 16 / kittyPixelsPerLine // = 8 lines per terminal row

		// Calculate which source line was clicked
		sourceLine := r.lastStartLine + (row * linesPerTermRow)
		if sourceLine < 0 {
			return 0
		}
		if sourceLine >= metrics.TotalVisualLines {
			return metrics.TotalVisualLines - 1
		}
		return sourceLine
	}

	// Braille fallback
	minimapRow := row + metrics.MinimapScrollOffset
	visualLine := minimapRow * 4
	if visualLine < 0 {
		return 0
	}
	if visualLine >= metrics.TotalVisualLines {
		return metrics.TotalVisualLines - 1
	}
	return visualLine
}

// ClearImage sends a Kitty graphics command to delete the minimap image.
// This should be called when disabling the minimap or exiting.
func (r *KittyMinimapRenderer) ClearImage() string {
	if !r.useKitty {
		return ""
	}
	// Delete image by ID
	return fmt.Sprintf("\033_Ga=d,d=i,i=%d\033\\", r.imageID)
}
