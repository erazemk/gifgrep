# iTerm2 inline images protocol (gifgrep)

iTerm2 supports inline images via a proprietary **OSC 1337** escape sequence. Unlike Kitty’s protocol, iTerm2 can render images (including animated GIFs) by sending the original file bytes.

## What gets sent

The “classic” form is:

```text
ESC ] 1337 ; File = <key=value;...> : <base64 file bytes> ESC \
```

Keys gifgrep uses:

- `name`: base64-encoded filename (used for UI/Downloads)
- `size`: file size in bytes (progress indicator)
- `inline=1`: render inline instead of downloading to `~/Downloads`
- `width`, `height`: **character cell** size (unitless numbers)
- `preserveAspectRatio`: `1` avoids stretching (TUI); `0` fills the fixed thumb block (CLI `--thumbs`)

## What gifgrep does

- **TUI preview (iTerm2):** sends the preview GIF bytes, sized to the preview cell rectangle (animated GIFs play natively in iTerm2).
- **CLI `--thumbs` (iTerm2):** sends the preview bytes sized to a small fixed cell block (animated when the preview is a GIF).

## Detection

iTerm2 also documents a “feature reporting” protocol for capability detection. gifgrep currently uses environment-based detection (`TERM_PROGRAM=iTerm.app` / `ITERM_SESSION_ID`) with an override via `GIFGREP_INLINE=iterm`.

## Links

- iTerm2 “Inline Images Protocol”: `https://iterm2.com/documentation-images.html`
- `imgcat` utility: `https://iterm2.com/utilities/imgcat`
