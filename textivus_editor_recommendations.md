# Textivus (Festivus) Editor: Lightweight Feature Recommendations

This document is a **product/design reminder** for building a Linux console (TUI) text editor that feels modern and familiar **without becoming an IDE**.

The goal: **high-leverage editing features** that dramatically improve usability, while keeping scope controlled and implementation maintainable.

---

## Product stance

**Textivus is not an IDE.**

- No always-on language servers
- No background project indexing
- No package manager inside the editor
- No debugger UI
- No embedded terminal multiplexer

**Focus:** fast startup, predictable behavior, minimal dependencies, and high-quality basics.

---

## What "not bloated" means here

- Features are **interactive and user-driven**, not long-running background systems.
- Prefer **small orthogonal capabilities** that compose well.
- Prefer **simple defaults** with optional toggles.
- Avoid features that require large runtimes (node, JVM), heavy plugin ecosystems, or constant indexing.

---

## Tiered roadmap

### Tier 1: Must-have (core editor comfort)

These features provide the highest usability gains with minimal long-term complexity.

#### 1) Multi-level undo/redo
- Reliable, multi-step undo and redo
- Undo across save is ideal (optional: persistent undo file)

**Why:** confidence and speed.

---

#### 2) Incremental search + highlight
- Search as you type
- Highlight all matches
- Next/prev match
- Toggle case / whole-word / regex (optional)

**Bonus:** replace + confirm per match.

---

#### 3) Goto line/column + status line
- `goto line` and `goto line:col`
- Status line: file name, modified flag, cursor line/col, encoding, newline type

**Why:** stack traces and compiler errors.

---

#### 4) Selection model that feels modern
- Character/word/line selection
- Select to end-of-line
- Copy/cut/paste
- Prefer integration with system clipboard where terminal supports it

---

#### 5) Indentation support that "just works"
- Indent/unindent selection
- Auto-indent on newline
- Smart backspace
- Detect tab vs spaces per file (optional but valuable)

---

#### 6) Soft wrap (display wrap) toggle
- Wrap at window width
- Toggleable (many users want to disable it)
- Optional wrap-at-column (80/100)

---

#### 7) Minimal syntax highlighting
- Keywords / strings / comments
- No semantic highlighting requirement
- Limited set of lexers for common languages

**Goal:** readability + error spotting.

---

#### 8) Safe saves + file-change detection
- Unsaved exit prompt
- Atomic save (write temp + rename)
- Detect external modifications and warn/reload

---

#### 9) Multiple buffers (open files) + fast switching
- Several open documents
- Fast buffer switching UX

---

### Tier 2: Strong additions (still low-bloat)

These materially improve usability without pushing into IDE territory.

#### 10) Rectangular (column) selection / block mode
This covers most multiple-cursor use cases with less complexity.

- Column selection mode
- Insert at beginning of selected lines
- Append at end of selected lines

---

#### 11) Split views (simple)
- One vertical split and/or horizontal split
- Limit complexity: no deep window tree required

---

#### 12) Open recent + fuzzy open
- Recent files list
- Fuzzy find / file open dialog

This reduces friction more than almost anything.

---

#### 13) Simple file browser pane (toggleable)
- Lightweight tree/list view
- Navigate, open, create new file

---

#### 14) Configurable keybindings + great defaults
- Discoverable defaults (help screen)
- Optional “vim-ish” mode
- Config file format: TOML/YAML/JSON

---

### Tier 3: Optional power-user extras (keep modular)

These are popular, but can expand scope. Prefer making them optional.

#### 15) Project search via ripgrep
- `rg` integration (external)
- Results list, jump-to-match

---

#### 16) Git-aware polish (still not IDE)
- Modified indicator
- Optional gutter symbols for changed lines (can be external-diff driven)

---

#### 17) LSP mode (only if plugin-like / optional)
- Disabled by default
- On-demand per language/filetype
- No global always-on indexing

---

## Features that tend to create "bloat"

Avoid these in the core or keep them entirely optional.

- Always-on language servers
- Background project indexing
- Built-in plugin marketplace
- Debugging UI / breakpoints / stack panes
- Embedded terminal multiplexer
- Complex UI layout manager
- Semantic refactors / full AST transformations

---

## A practical MVP scope statement

> **Textivus is a fast, friendly console editor.**  
> It focuses on comfort and correctness: undo/redo, search, selection, indentation, syntax highlight, safe saves, and buffer management — **without IDE subsystems**.

---

## Suggested "MVP checklist"

- [ ] Multi-level undo/redo
- [ ] Incremental search + highlight
- [ ] Replace (with confirm)
- [ ] Goto line/column
- [ ] Status line (file / modified / line-col / encoding / newline)
- [ ] Indent/unindent selection
- [ ] Auto-indent + smart backspace
- [ ] Soft wrap toggle
- [ ] Minimal syntax highlighting
- [ ] Multiple buffers + switcher
- [ ] Atomic save + unsaved prompt
- [ ] External file change detection

---

## Suggested "v1.1" checklist

- [ ] Rectangular selection
- [ ] Simple splits
- [ ] Recent files
- [ ] Fuzzy open
- [ ] Simple file browser pane
- [ ] Configurable keys / config file

---

## Notes for Claude Code / implementation guidance

When implementing features, aim for:
- Small pure functions for buffer transforms
- Minimal global state; prefer immutable-ish operations where possible
- Deterministic behavior (avoid hidden background processes)
- Clear instrumentation/logging hooks
- Simple test harness for undo/redo invariants and selection behaviors

---

*End of doc.*
