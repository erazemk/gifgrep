# gifgrep

CLI GIF search with two brains:
- Scriptable mode (default): URLs/JSON for pipes
- TUI mode (`--tui`): arrow-key browse + kitty preview

## Requirements
- Terminal with Kitty graphics (Kitty, Ghostty) for inline preview
- Go 1.21+

## Usage
Scriptable:
```bash
export TENOR_API_KEY=your_key # optional; falls back to LIVDSRZULELA

gifgrep cats

gifgrep cats --json | jq '.[] | .url'
```

TUI:
```bash
gifgrep --tui cats
```

## Flags (subset)
- `-m N` max results
- `-n` number results
- `-i` ignore case (filters)
- `-E` regex filter (title+tags)
- `-v` invert vibe (exclude `--mood` matches)
- `--mood angry`
- `--source tenor`

## pnpm scripts
```bash
pnpm start
pnpm tenor
pnpm test
pnpm check
pnpm build
```
