# Festivus TODO

## Tier 0: CI, Distribution & Installation

- [ ] GitHub Actions CI
  - [ ] Build on push/PR
  - [ ] Run tests
  - [ ] Release workflow (tag-triggered)
- [ ] Build targets
  - [ ] linux/amd64
  - [ ] linux/arm64
  - [ ] darwin/amd64
  - [ ] darwin/arm64 (Apple Silicon)
- [ ] Distribution packages
  - [ ] Homebrew tap (macOS/Linux)
  - [ ] .deb package (Debian/Ubuntu)
  - [ ] .rpm package (Fedora/RHEL)
  - [ ] AUR package (Arch Linux)
- [ ] Curl install script (one-liner installation)
- [ ] GitHub Releases with pre-built binaries

## Tier 1: MVP Gaps

- [x] Multiple buffers + fast buffer switching
- [x] External file change detection (warn if file modified on disk)
- [x] Test suite
- [ ] Tab support (tabs vs spaces, configurable tab width)
- [ ] Encoding handling (detect and convert non-UTF-8 files)

## Tier 2: v1.1 Features

- [ ] Rectangular/column selection (block mode)
- [ ] Split views (simple horizontal/vertical)
  - [ ] Horizontal split
  - [ ] Vertical split
  - [ ] Same buffer in multiple views
- [x] Configurable keybindings
- [ ] Braille minimap (2x4 pixels per cell for code density)
  - 4 chars wide works best
  - Truncate source lines at ~40 chars before converting
- [ ] Kitty graphics minimap (true bitmap for compatible terminals)
- [ ] Emoji picker

## Tier 2.5: In Progress

- [ ] Graceful degradation (ASCII fallback for limited terminals)
  - [x] Terminal capability detection (UTF-8, colors, Kitty)
  - [ ] Auto-detect and apply ASCII mode when UTF-8 not supported

## Tier 3: Optional Power-User

- [ ] Fuzzy project navigation (files + content search via ripgrep/fzf-style)
- [ ] Git-aware gutter indicators (modified lines)
- [ ] LSP mode (disabled by default, on-demand)
- [ ] Vim keybindings (optional modal editing)
- [ ] Macro recording/playback
- [ ] Graphical theme editor

---

## Bloat to Avoid

These features are explicitly out of scope to keep Festivus fast and focused:

- Always-on language servers
- Background project indexing
- Built-in plugin marketplace
- Debugging UI / breakpoints
- Embedded terminal multiplexer
- Complex UI layout manager
- Semantic refactors / AST transformations
