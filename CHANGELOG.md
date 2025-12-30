# Changelog

## 0.1.0 - 2025-12-30
### Added
- Scriptable CLI that outputs GIF URLs, with optional JSON metadata and numbered results.
- Tenor search backend with configurable `TENOR_API_KEY` (defaults to the public demo key).
- Grep-flavored filters: ignore case, regex over title+tags, mood filter, invert mood matches, and max results.
- Interactive TUI mode with query/browse states, arrow-key navigation, and status line.
- Inline preview using the Kitty graphics protocol (no alt-screen), with automatic cleanup on exit.
- Animated GIF decoding with frame compositing/disposal handling, delay clamping, and a frame cap for performance.
- Software animation fallback for terminals that do not play kitty animations (auto-detect Ghostty or `GIFGREP_SOFTWARE_ANIM=1`).
- Aspect-ratio-aware preview sizing with configurable cell aspect (`GIFGREP_CELL_ASPECT`).
- In-memory preview cache keyed by URL to keep browsing fast.
- Responsive layout that splits list/preview on wide terminals or stacks preview under the list on narrow ones.

### Developer Experience
- Makefile + justfile targets for fmt/lint/test/check/build/cover/snap.
- gofumpt formatting and golangci-lint ruleset.
- GitHub Actions CI for format check, lint, and tests.
- `scripts/ghostty-web-snap.mjs` to generate Ghostty-web screenshots for visual checks.
