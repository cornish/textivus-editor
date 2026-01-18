package clipboard

import (
	"io"
	"os"

	"github.com/atotto/clipboard"
	"github.com/aymanbagabas/go-osc52/v2"
)

// Clipboard provides unified clipboard access with OSC52 support for SSH.
type Clipboard struct {
	// Internal clipboard for when no system clipboard is available
	internal string
	// Whether we're likely in an SSH session
	isSSH bool
	// Output writer for OSC52 sequences (typically os.Stdout)
	output io.Writer
}

// New creates a new Clipboard instance.
func New(output io.Writer) *Clipboard {
	if output == nil {
		output = os.Stdout
	}
	return &Clipboard{
		isSSH:  isSSHSession(),
		output: output,
	}
}

// isSSHSession detects if we're running in an SSH session.
func isSSHSession() bool {
	// Check common SSH environment variables
	if os.Getenv("SSH_TTY") != "" {
		return true
	}
	if os.Getenv("SSH_CLIENT") != "" {
		return true
	}
	if os.Getenv("SSH_CONNECTION") != "" {
		return true
	}
	return false
}

// Copy copies the given text to the clipboard.
// In SSH sessions, it uses OSC52 escape sequences.
// Locally, it tries the system clipboard first, then falls back to OSC52.
func (c *Clipboard) Copy(text string) error {
	// Always store internally as a last resort
	c.internal = text

	if c.isSSH {
		// In SSH, always use OSC52
		return c.copyOSC52(text)
	}

	// Try system clipboard first
	err := clipboard.WriteAll(text)
	if err != nil {
		// Fall back to OSC52
		return c.copyOSC52(text)
	}

	return nil
}

// copyOSC52 copies text using OSC52 escape sequence.
func (c *Clipboard) copyOSC52(text string) error {
	seq := osc52.New(text)
	_, err := io.WriteString(c.output, seq.String())
	return err
}

// Paste returns text from the clipboard.
// Note: OSC52 paste (OSC52 query) is not widely supported.
// We rely on the system clipboard or the internal buffer.
func (c *Clipboard) Paste() (string, error) {
	// Try system clipboard first
	text, err := clipboard.ReadAll()
	if err == nil && text != "" {
		return text, nil
	}

	// Fall back to internal clipboard
	return c.internal, nil
}

// HasContent returns true if there's content available to paste.
func (c *Clipboard) HasContent() bool {
	// Check system clipboard
	text, err := clipboard.ReadAll()
	if err == nil && text != "" {
		return true
	}

	// Check internal clipboard
	return c.internal != ""
}

// Clear clears the internal clipboard.
func (c *Clipboard) Clear() {
	c.internal = ""
}

// IsSSH returns true if we're in an SSH session.
func (c *Clipboard) IsSSH() bool {
	return c.isSSH
}
