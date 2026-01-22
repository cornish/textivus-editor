# Textivus TODO

## Tier 1: CI, Distribution & Installation

- [x] GitHub Actions CI
  - [x] Build on push/PR
  - [x] Run tests
  - [x] Release workflow (tag-triggered)
- [x] Build targets
  - [x] linux/amd64
  - [x] linux/arm64
  - [x] darwin/amd64
  - [x] darwin/arm64 (Apple Silicon)
- [x] Curl install script (one-liner installation)
- [x] GitHub Releases with pre-built binaries
- [ ] Distribution packages
  - [ ] Homebrew tap (macOS/Linux)
  - [ ] .deb package (Debian/Ubuntu)
  - [ ] .rpm package (Fedora/RHEL)
  - [ ] AUR package (Arch Linux)
- [ ] (maybe) Add `txv` shortcut to distribution packages (Homebrew, .deb, .rpm, AUR)

## Tier 2: MVP Gaps

- [x] Multiple buffers + fast buffer switching
- [x] External file change detection (warn if file modified on disk)
- [x] Test suite
- [x] Tab support (tabs vs spaces, configurable tab width, block indent/dedent)
- [x] Encoding handling (detect and convert non-UTF-8 files)

## Tier 3: v1.1 Features

- [x] Configurable keybindings
- [x] Braille minimap (2x4 pixels per cell for code density)
  - 4 chars wide works best
  - Truncate source lines at ~40 chars before converting
- [x] Kitty graphics minimap (true bitmap for compatible terminals)
- [ ] Expand test suite
  - [ ] Undo/redo operations
  - [ ] Selection logic
  - [ ] Cursor navigation (word movement, line boundaries)
  - [ ] Syntax highlighting (Chroma integration)
- [ ] Highlight search hits in text editor (while search bar open)
- [ ] Highlight search hits in minimap (while search bar open)
- [ ] Highlight search hits in scrollbar (while search bar open)
- [ ] Hotkeys in dialogs (underlined letters for quick access)
- [ ] Rectangular/column selection (block mode)
- [ ] Split views (simple horizontal/vertical)
  - [ ] Horizontal split
  - [ ] Vertical split
  - [ ] Same buffer in multiple views
- [ ] Emoji picker

## Polish

- [ ] Create mini-framework for dialogs (standard components: buttons, file pickers, inputs)
- [ ] Make dialog appearance, function, and conventions consistent throughout
- [ ] Mouse-enable buttons with hotkeys in dialogs
- [ ] Make non-editor cursors consistent throughout
- [ ] Consistent default answers to [y/n] questions
- [ ] Create consistent pattern for dialog vs status bar messages and user input
- [ ] Reconsider right side layout/appearance of status bar
- [ ] Revisit search and replace behavior
- [ ] Address theme colors in minimap and scrollbar
- [ ] Consider displaying encoding compatibility for buffer in Set Encoding dialog
- [ ] Minimap location highlighting barely visible for dense text (e.g., fred.txt with large word-wrapped paragraphs) - review VS Code style
- [ ] Add gutter separation between buffer text and Kitty minimap

## Tier 4: Optional Power-User

- [ ] Vim keybindings (optional modal editing)
- [ ] Fuzzy project navigation (files + content search via ripgrep/fzf-style)
- [ ] Git-aware gutter indicators (modified lines)
- [ ] Macro recording/playback
- [ ] LSP mode (disabled by default, on-demand)
- [ ] Graphical theme editor

---

## Bloat to Avoid

These features are explicitly out of scope to keep Textivus fast and focused:

- Always-on language servers
- Background project indexing
- Built-in plugin marketplace
- Debugging UI / breakpoints
- Embedded terminal multiplexer
- Complex UI layout manager
- Semantic refactors / AST transformations
