# Textivus
[![Release](https://img.shields.io/github/v/release/cornish/textivus-editor?display_name=tag&sort=semver)](https://github.com/cornish/textivus-editor/releases)
[![CI](https://github.com/cornish/textivus-editor/actions/workflows/ci.yml/badge.svg)](https://github.com/cornish/textivus-editor/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**Textivus** is a fast, friendly **terminal (TUI) text editor for Linux** (and macOS) inspired by the simplicity of **nano/micro** and the familiarity of **DOS EDIT** — with modern comforts like multi-file buffers, incremental find/replace, and syntax highlighting.

> **A text editor for the rest of us!**

**Status:** Early but usable. Expect breaking changes, but minor ones, before v1.0 (especially to defaults and configuration). See the changelog for upgrade notes.

![Textivus terminal text editor screenshot](docs/screenshot.png)

---

## Quick start

### Install (Linux/macOS)
```sh
curl -fsSL https://raw.githubusercontent.com/cornish/textivus-editor/main/install.sh | sh
```

Or download binaries from **GitHub Releases**.

### Run
```sh
textivus README.md
```

---

## Why Textivus?

Textivus is for people who want a **comfortable, familiar editor in the terminal** without turning their console into an IDE.

- **Instant startup** (no indexing, no plugins, no waiting)
- **Modern shortcuts** by default (Ctrl+S / Ctrl+C / Ctrl+V / Ctrl+Z)
- **Classic look** (DOS EDIT-inspired menu/status bars)
- **Great for SSH** (includes OSC52 clipboard support)
- **Sane features**: just enough power, without bloat

---

## Features

- **Instant startup** — no bloat, just editing
- **Customizable theming** — built-in DOS EDIT, light, dark and other themes; fully customizable
- **Modern keyboard shortcuts** — Ctrl+S, Ctrl+C, Ctrl+V, Ctrl+Z, etc.
- **Configurable keybindings** — customize shortcuts via Options menu
- **Multiple encodings supported** — UTF-8/UTF-16, Western European, and CJK encodings (Shift-JIS, EUC-JP, GBK/GB18030, EUC-KR)
- **Multiple buffers** — edit multiple files with fast switching (Alt+< / Alt+>)
- **Recent files & directories** — quick access from menus
- **Favorites** — star frequently-used files/directories
- **Mouse support** — mouse supported, but optional; click to move cursor, drag to select, scroll wheel
- **Shift+Arrow selection** — select text the modern way
- **Word wrap** — toggle via Options menu
- **Line numbers** — toggle via Options menu or Ctrl+L
- **Syntax highlighting** — auto-detected by file extension
- **Find & Replace** — Ctrl+F to find, Ctrl+H to find and replace
- **Go to Line** — Ctrl+G to jump to a specific line
- **Cut Line** — Ctrl+K cuts the entire current line (like nano)
- **Word & character counts** — displayed in the status bar
- **Clipboard support**
  - System clipboard integration:
    - X11: `xclip` / `xsel` *(install required)*
    - Wayland: `wl-clipboard` (`wl-copy`, `wl-paste`) *(install required)*
    - macOS: `pbcopy` / `pbpaste` *(built-in)*
  - **OSC52 clipboard** support for remote SSH sessions
- **Undo/Redo** — Ctrl+Z / Ctrl+Y with full history

---

## Supported encodings

Textivus can open and save files in the following encodings:

| Encoding | ID | Notes / Aliases |
|---|---|---|
| UTF-8 | `utf-8` | Default |
| UTF-8 BOM | `utf-8-bom` | UTF-8 with byte order mark |
| UTF-16 LE | `utf-16-le` | `UTF-16LE` |
| UTF-16 BE | `utf-16-be` | `UTF-16BE` |
| ISO-8859-1 (Latin-1) | `iso-8859-1` | `latin1` |
| Windows-1252 | `windows-1252` | `CP1252` |
| ISO-8859-15 (Latin-9) | `iso-8859-15` | Includes `€` |
| Shift-JIS | `shift-jis` | `SJIS`, `MS_Kanji` |
| EUC-JP | `euc-jp` |  |
| GBK | `gbk` | `GB2312` |
| GB18030 | `gb18030` |  |
| EUC-KR | `euc-kr` |  |

---

## Non-goals

Textivus is **not an IDE**.

- No project indexing
- No always-on language servers
- No plugin marketplace
- No telemetry

---

## Keyboard shortcuts (essentials)

- **Save:** Ctrl+S  
- **Open:** Ctrl+O  
- **Quit:** Ctrl+Q  
- **Find:** Ctrl+F  
- **Replace:** Ctrl+H  
- **Go to line:** Ctrl+G  
- **Undo / Redo:** Ctrl+Z / Ctrl+Y  
- **Cut line:** Ctrl+K  
- **Switch buffers:** Alt+< / Alt+>

Full shortcuts list: **[docs/shortcuts.md](docs/shortcuts.md)**

---

## Clipboard support (Linux)

For clipboard integration with other applications, install one of:

```sh
# X11
sudo apt install xclip
# or
sudo apt install xsel

# Wayland
sudo apt install wl-clipboard
```

Without these tools, copy/paste will still work **inside Textivus**, but won’t integrate with other apps.

---

## Build from source

Requires **Go 1.21+**.

```sh
git clone https://github.com/cornish/textivus-editor.git
cd textivus-editor
go build
./textivus [filename]
```

---

## License

MIT
