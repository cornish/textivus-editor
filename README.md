# Festivus

**A Text Editor for the Rest of Us**

Festivus is a terminal-based text editor inspired by the classic DOS EDIT, built with Go and the [Bubbletea](https://github.com/charmbracelet/bubbletea) TUI framework.

![Festivus Screenshot](screenshot.png)

## Features

- **Instant startup** - No bloat, just editing
- **Classic DOS EDIT styling** - Dark blue menu bar and status bar with cyan highlights
- **Modern keyboard shortcuts** - Ctrl+S, Ctrl+C, Ctrl+V, Ctrl+Z, etc.
- **Configurable keybindings** - Customize shortcuts via Options menu
- **Multiple buffers** - Edit multiple files with fast switching (Alt+< / Alt+>)
- **Recent files** - Quick access to recently opened files (Ctrl+R)
- **Mouse support** - Click to position cursor, drag to select, scroll wheel
- **Shift+Arrow selection** - Select text the modern way
- **Word wrap** - Toggle via Options menu
- **Line numbers** - Toggle via Options menu or Ctrl+L
- **Syntax highlighting** - Auto-detected by file extension
- **Find & Replace** - Ctrl+F to find, Ctrl+H to find and replace
- **Go to Line** - Ctrl+G to jump to a specific line
- **Cut Line** - Ctrl+K to cut the entire current line (like nano)
- **Word & Character counts** - Displayed in the status bar
- **Clipboard support** - Native X11/Wayland support, OSC52 for SSH
- **Undo/Redo** - Ctrl+Z / Ctrl+Y with full history

## Installation

### From Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/cornish/festivus.git
cd festivus
go build
./festivus [filename]
```

### Clipboard Support (Linux)

For clipboard integration with other applications, install one of:

```bash
# X11
sudo apt install xclip
# or
sudo apt install xsel

# Wayland
sudo apt install wl-clipboard
```

Without these tools, copy/paste will only work within Festivus.

## Keyboard Shortcuts

### File Operations
| Action | Shortcut |
|--------|----------|
| New | Ctrl+N |
| Open | Ctrl+O |
| Recent Files | Ctrl+R |
| Save | Ctrl+S |
| Close | Ctrl+W |
| Quit | Ctrl+Q |

### Buffers
| Action | Shortcut |
|--------|----------|
| Previous Buffer | Alt+< |
| Next Buffer | Alt+> |
| Buffer 1-9 | Alt+1 through Alt+9 |

### Editing
| Action | Shortcut |
|--------|----------|
| Undo | Ctrl+Z |
| Redo | Ctrl+Y |
| Cut | Ctrl+X |
| Copy | Ctrl+C |
| Paste | Ctrl+V |
| Cut Line | Ctrl+K |
| Select All | Ctrl+A |

### Search
| Action | Shortcut |
|--------|----------|
| Find | Ctrl+F |
| Find Next | F3 |
| Replace | Ctrl+H |
| Go to Line | Ctrl+G |

### Navigation
| Action | Shortcut |
|--------|----------|
| Start of file | Ctrl+Home |
| End of file | Ctrl+End |
| Start of line | Home |
| End of line | End |
| Word left/right | Ctrl+Left/Right |
| Page up/down | PgUp/PgDn |

### Selection
| Action | Shortcut |
|--------|----------|
| Select with cursor | Shift+Arrow |
| Select word | Ctrl+Shift+Left/Right |
| Select to line start/end | Shift+Home/End |
| Select to file start/end | Ctrl+Shift+Home/End |

### Options
| Action | Shortcut |
|--------|----------|
| Toggle Line Numbers | Ctrl+L |

## Menu Navigation

- **F10** or click to open File menu
- **Alt+F** File, **Alt+B** Buffers, **Alt+E** Edit, **Alt+S** Search, **Alt+O** Options, **Alt+H** Help
- Arrow keys to navigate within menus
- Press underlined letter to select item
- Enter to select, Escape to close

All keyboard shortcuts can be customized via **Options → Keybindings**.

## Status Bar

The status bar shows:
- Filename (with * if modified)
- Word count (W:xxx)
- Character count (C:xxx)
- Current line and column
- File encoding (UTF-8)

## Configuration

Festivus stores its configuration in `~/.config/festivus/config.toml`:

```toml
[editor]
word_wrap = false
line_numbers = false
syntax_highlight = true
true_color = true    # Set to false for older terminals
backup_count = 0     # 0=disabled, 1=filename~, 2+=numbered (filename~1~ newest)
max_buffers = 20     # Maximum open buffers (0=unlimited)

[theme]
name = "default"  # or "dark", "light", "monokai", "nord", "dracula", "gruvbox", "solarized", "catppuccin"
```

### Keybindings

Keybindings are stored in `~/.config/festivus/keybindings.toml` and can be edited via **Options → Keybindings**. Each action supports a primary and alternate binding:

```toml
[save]
primary = "ctrl+s"

[find]
primary = "ctrl+f"
alternate = "f3"
```

## Themes

Festivus supports color themes with 9 built-in options:
- **default** - Classic DOS EDIT style (blue/cyan)
- **dark** - Modern dark theme
- **light** - Light theme for bright environments
- **monokai** - Monokai-inspired dark theme
- **nord** - Arctic, north-bluish palette
- **dracula** - Dark theme with vibrant colors
- **gruvbox** - Retro groove color scheme
- **solarized** - Precision colors (dark variant)
- **catppuccin** - Soothing pastel theme (Mocha)

Switch themes at runtime via the **Options** menu, or set the default in your config file.

In the theme dialog, press **E** to edit a theme or **C** to copy it with a new name - the theme file will open directly in Festivus for editing.

### Custom Themes

Create custom themes in `~/.config/festivus/themes/`:

```toml
# ~/.config/festivus/themes/mytheme.toml
name = "mytheme"
description = "My custom theme"
author = "Your Name"

[ui]
menu_bg = "#3B4252"
menu_fg = "#ECEFF4"
menu_highlight_bg = "#5E81AC"
menu_highlight_fg = "#ECEFF4"
status_bg = "#3B4252"
status_fg = "#ECEFF4"
status_accent = "#88C0D0"
selection_bg = "#4C566A"
selection_fg = "#ECEFF4"
line_number = "#4C566A"
line_number_active = "#D8DEE9"
error_fg = "#BF616A"
disabled_fg = "#4C566A"
dialog_bg = "#3B4252"
dialog_fg = "#ECEFF4"
dialog_border = "#4C566A"
dialog_title = "#88C0D0"
dialog_button = "#5E81AC"
dialog_button_fg = "#ECEFF4"

[syntax]
keyword = "#81A1C1"
string = "#A3BE8C"
comment = "#616E88"
number = "#B48EAD"
operator = "#81A1C1"
function = "#88C0D0"
type = "#8FBCBB"
```

Then reference it in your config: `name = "mytheme"`

### Color Formats

Colors can be specified as:
- **ANSI 16 colors**: `"0"` - `"15"`
- **256-color palette**: `"16"` - `"255"`
- **Hex colors**: `"#RGB"` or `"#RRGGBB"`

### True Color Support

Hex colors use 24-bit "true color" by default, which requires a modern terminal:
- iTerm2, Alacritty, Kitty, Windows Terminal
- GNOME Terminal, Konsole, VS Code terminal

For older terminals, set `true_color = false` in your config to automatically convert hex colors to the nearest 256-color match.

## Why "Festivus"?

> "A Festivus for the rest of us!"

Named after the holiday from Seinfeld, because every text editor tries to be Vim or Emacs. Festivus is for the rest of us who just want to edit text.

## Built With

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [go-runewidth](https://github.com/mattn/go-runewidth) - Unicode width calculation
- [Chroma](https://github.com/alecthomas/chroma) - Syntax highlighting

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

Contributions welcome! Feel free to submit issues and pull requests.

---

*"I got a lot of problems with you people!"* - Frank Costanza
