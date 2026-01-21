# Keyboard Shortcuts

All shortcuts can be customized via **Options → Keybindings**.

---

## File Operations

| Action | Shortcut |
|--------|----------|
| New file | Ctrl+N |
| Open file | Ctrl+O |
| Recent files | Ctrl+R |
| Save | Ctrl+S |
| Save As | (menu only) |
| Close file | Ctrl+W |
| Quit | Ctrl+Q |

---

## Editing

| Action | Shortcut |
|--------|----------|
| Undo | Ctrl+Z |
| Redo | Ctrl+Y |
| Cut | Ctrl+X |
| Copy | Ctrl+C |
| Paste | Ctrl+V |
| Cut line | Ctrl+K |
| Select all | Ctrl+A |
| Indent | Tab |
| Dedent | Shift+Tab |
| Block indent | Tab (with selection) |
| Block dedent | Shift+Tab (with selection) |

---

## Search

| Action | Shortcut |
|--------|----------|
| Find | Ctrl+F |
| Find next | F3 |
| Find & Replace | Ctrl+H |
| Go to line | Ctrl+G |

---

## Navigation

| Action | Shortcut |
|--------|----------|
| Move by word | Ctrl+Left / Ctrl+Right |
| Start of line | Home |
| End of line | End |
| Start of file | Ctrl+Home |
| End of file | Ctrl+End |
| Page up / down | PgUp / PgDn |

---

## Selection

| Action | Shortcut |
|--------|----------|
| Select character | Shift+Arrow |
| Select word | Ctrl+Shift+Left / Ctrl+Shift+Right |
| Select to line start | Shift+Home |
| Select to line end | Shift+End |
| Select to file start | Ctrl+Shift+Home |
| Select to file end | Ctrl+Shift+End |
| Select all | Ctrl+A |

---

## Buffers

| Action | Shortcut |
|--------|----------|
| Next buffer | Alt+> or Ctrl+Tab |
| Previous buffer | Alt+< or Ctrl+Shift+Tab |
| Buffer 1–9 | Alt+1 through Alt+9 |

---

## View

| Action | Shortcut |
|--------|----------|
| Toggle line numbers | Ctrl+L |

---

## Menus

| Action | Shortcut |
|--------|----------|
| Open File menu | F10 or Alt+F |
| Buffers menu | Alt+B |
| Edit menu | Alt+E |
| Search menu | Alt+S |
| Options menu | Alt+O |
| Help menu | Alt+H |
| Navigate menu | Arrow keys |
| Select item | Enter or underlined letter |
| Close menu | Escape |

---

## Help

| Action | Shortcut |
|--------|----------|
| Show help | F1 |

---

## File Browser

| Action | Key |
|--------|-----|
| Open file / enter directory | Enter |
| Go to parent directory | Backspace |
| Toggle favorite | F |
| Cancel | Escape |

---

## Customizing Keybindings

Keybindings are stored in `~/.config/textivus/keybindings.toml`. Edit via **Options → Keybindings** or manually:

```toml
[save]
primary = "ctrl+s"

[find]
primary = "ctrl+f"
alternate = "f3"
```

Supported modifiers: `ctrl`, `alt`, `shift`
Supported keys: `a`–`z`, `0`–`9`, `f1`–`f12`, `enter`, `tab`, `space`, `home`, `end`, `pgup`, `pgdn`, `left`, `right`, `up`, `down`, `insert`, `delete`, `backspace`, `escape`
