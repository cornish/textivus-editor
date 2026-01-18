# Festivus

**A Text Editor for the Rest of Us**

Festivus is a terminal-based text editor inspired by the classic DOS EDIT, built with Go and the [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI framework.

![Festivus Screenshot](screenshot.png)

## Features

- **Instant startup** - No bloat, just editing
- **Classic DOS EDIT styling** - Dark blue menu bar and status bar with cyan highlights
- **Modern keyboard shortcuts** - Ctrl+S, Ctrl+C, Ctrl+V, Ctrl+Z, etc.
- **Mouse support** - Click to position cursor, drag to select
- **Shift+Arrow selection** - Select text the modern way
- **Word wrap** - Toggle via View menu
- **Line numbers** - Toggle via View menu
- **Find** - Ctrl+F to search
- **Clipboard support** - Works over SSH via OSC52
- **Undo/Redo** - Ctrl+Z / Ctrl+Y

## Installation

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/cornish/festivus.git
cd festivus
go build
./festivus [filename]
```

## Keyboard Shortcuts

| Action | Shortcut |
|--------|----------|
| Save | Ctrl+S |
| Open | Ctrl+O |
| New | Ctrl+N |
| Quit | Ctrl+Q |
| Undo | Ctrl+Z |
| Redo | Ctrl+Y |
| Cut | Ctrl+X |
| Copy | Ctrl+C |
| Paste | Ctrl+V |
| Select All | Ctrl+A |
| Find | Ctrl+F |
| Start of file | Ctrl+Home |
| End of file | Ctrl+End |
| Select with cursor | Shift+Arrow |

## Menu Navigation

- Click menu items with mouse
- Alt+F, Alt+E, Alt+V, Alt+H to open menus (when implemented)
- Arrow keys to navigate within menus
- Enter to select, Escape to close

## Why "Festivus"?

> "A Festivus for the rest of us!"

Named after the holiday from Seinfeld, because every text editor tries to be Vim or Emacs. Festivus is for the rest of us who just want to edit text.

## Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [go-runewidth](https://github.com/mattn/go-runewidth) - Unicode width calculation

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Feel free to submit issues and pull requests.

---

*"I got a lot of problems with you people!"* - Frank Costanza
