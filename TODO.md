# Festivus TODO

## Tier 1: MVP Gaps

- [x] Multiple buffers + fast buffer switching
- [x] External file change detection (warn if file modified on disk)

## Tier 2: v1.1 Features

- [ ] Rectangular/column selection (block mode)
- [ ] Split views (simple horizontal/vertical)
- [ ] Configurable keybindings
- [ ] Braille minimap (2x4 pixels per cell for code density)
  - 4 chars wide works best
  - Truncate source lines at ~40 chars before converting
- [ ] Graceful degradation (ASCII fallback for limited terminals)

## Tier 3: Optional Power-User

- [ ] Fuzzy file open
- [ ] Project search via ripgrep integration
- [ ] Git-aware gutter indicators (modified lines)
- [ ] LSP mode (disabled by default, on-demand)
- [ ] Vim keybindings (optional modal editing)
- [ ] Kitty graphics minimap (true bitmap for compatible terminals)

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
