package tui

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/steipete/gifgrep/gifdecode"
	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/termcaps"
)

func TestRenderDeletesOldImage(t *testing.T) {
	state := &appState{
		mode:          modeBrowse,
		results:       []model.Result{{Title: "A"}},
		activeImageID: 7,
		inline:        termcaps.InlineKitty,
		currentAnim:   nil,
		opts:          model.Options{Source: "tenor"},
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	render(state, out, 10, 60)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "a=d") {
		t.Fatalf("expected delete kitty image")
	}
}

func TestRenderDoesNotClearItermPreviewEveryRender(t *testing.T) {
	prev := clearPreviewAreaFn
	t.Cleanup(func() { clearPreviewAreaFn = prev })

	var clears int
	clearPreviewAreaFn = func(_ *bufio.Writer, _ layout) { clears++ }

	state := &appState{
		mode:   modeBrowse,
		inline: termcaps.InlineIterm,
		results: []model.Result{
			{Title: "A"},
		},
		currentAnim: &gifAnimation{
			ID:     1,
			RawGIF: []byte("GIF89a\x01\x00\x01\x00"),
			Width:  1,
			Height: 1,
		},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	render(state, out, 20, 100)
	render(state, out, 20, 100)

	if clears != 1 {
		t.Fatalf("expected 1 clear for iterm split preview, got %d", clears)
	}
}

func TestRenderKeepsClearingKittyPreview(t *testing.T) {
	prev := clearPreviewAreaFn
	t.Cleanup(func() { clearPreviewAreaFn = prev })

	var clears int
	clearPreviewAreaFn = func(_ *bufio.Writer, _ layout) { clears++ }

	state := &appState{
		mode:   modeBrowse,
		inline: termcaps.InlineKitty,
		results: []model.Result{
			{Title: "A"},
		},
		currentAnim: &gifAnimation{
			ID:     1,
			Frames: []gifdecode.Frame{{PNG: []byte{1, 2, 3}, Delay: 80 * time.Millisecond}},
			Width:  200,
			Height: 100,
		},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	render(state, out, 20, 100)
	render(state, out, 20, 100)

	if clears != 2 {
		t.Fatalf("expected 2 clears for kitty split preview, got %d", clears)
	}
}

func TestDrawPreviewItermClearsOldRectOnResend(t *testing.T) {
	prev := clearItermRectFn
	t.Cleanup(func() { clearItermRectFn = prev })

	var clears int
	clearItermRectFn = func(_ *bufio.Writer, _, _, _, _ int) { clears++ }

	state := &appState{
		inline: termcaps.InlineIterm,
		currentAnim: &gifAnimation{
			ID:     1,
			RawGIF: []byte("GIF89a\x01\x00\x01\x00"),
			Width:  1,
			Height: 1,
		},
		previewNeedsSend: true,
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	drawPreview(state, out, 20, 8, 2, 2) // first send (no clear)
	if clears != 1 {
		t.Fatalf("expected 1 clear on first send, got %d", clears)
	}

	state.previewDirty = true
	drawPreview(state, out, 20, 8, 2, 2) // resend -> should clear previous rect
	if clears != 2 {
		t.Fatalf("expected 2 clears after resend, got %d", clears)
	}
}

func TestDrawPreviewItermClearsNewRectOnExpand(t *testing.T) {
	prev := clearItermRectFn
	t.Cleanup(func() { clearItermRectFn = prev })

	var clears int
	clearItermRectFn = func(_ *bufio.Writer, _, _, _, _ int) { clears++ }

	state := &appState{
		inline: termcaps.InlineIterm,
		currentAnim: &gifAnimation{
			ID:     1,
			RawGIF: []byte("GIF89a\x01\x00\x01\x00"),
			Width:  1,
			Height: 1,
		},
		previewNeedsSend: true,
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	drawPreview(state, out, 10, 4, 2, 2) // first send (no clear)
	state.previewDirty = true
	drawPreview(state, out, 20, 8, 2, 2) // expand -> should clear old + new rect
	if clears != 3 {
		t.Fatalf("expected 3 clears on expand resend, got %d", clears)
	}
}

func TestRenderItermErasesContentOnSend(t *testing.T) {
	prev := eraseItermContentAreaFn
	t.Cleanup(func() { eraseItermContentAreaFn = prev })

	var clears int
	eraseItermContentAreaFn = func(_ *bufio.Writer, _ layout) { clears++ }

	state := &appState{
		mode:   modeBrowse,
		inline: termcaps.InlineIterm,
		results: []model.Result{
			{Title: "A"},
		},
		currentAnim:      &gifAnimation{ID: 1, RawGIF: []byte("GIF89a\x01\x00\x01\x00"), Width: 400, Height: 100},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	render(state, out, 20, 120)
	if clears != 1 {
		t.Fatalf("expected 1 erase on send, got %d", clears)
	}

	render(state, out, 20, 120)
	if clears != 1 {
		t.Fatalf("expected no extra erase without send, got %d", clears)
	}
}

func TestRenderItermHardClearsOnShrink(t *testing.T) {
	prev := clearItermScreenFn
	t.Cleanup(func() { clearItermScreenFn = prev })

	var clears int
	clearItermScreenFn = func(_ *bufio.Writer) { clears++ }

	state := &appState{
		mode:   modeBrowse,
		inline: termcaps.InlineIterm,
		results: []model.Result{
			{Title: "A"},
		},
		currentAnim:      &gifAnimation{ID: 1, RawGIF: []byte("GIF89a\x01\x00\x01\x00"), Width: 800, Height: 100},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	render(state, out, 20, 120) // first send sets itermLast
	if clears != 0 {
		t.Fatalf("expected no hard clear on first send, got %d", clears)
	}

	// Switch to tall/narrow GIF => preview shrinks, should hard clear to avoid leftover image cells.
	state.currentAnim = &gifAnimation{ID: 2, RawGIF: []byte("GIF89a\x01\x00\x01\x00"), Width: 100, Height: 800}
	state.previewNeedsSend = true
	render(state, out, 20, 120)
	if clears != 1 {
		t.Fatalf("expected 1 hard clear on shrink, got %d", clears)
	}
}

func TestRenderWithPreviewRight(t *testing.T) {
	state := &appState{
		mode:    modeBrowse,
		results: []model.Result{{Title: "A"}},
		inline:  termcaps.InlineKitty,
		currentAnim: &gifAnimation{
			ID:     1,
			Frames: []gifdecode.Frame{{PNG: []byte{1, 2, 3}, Delay: 80 * time.Millisecond}},
			Width:  200,
			Height: 100,
		},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	render(state, out, 20, 90)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "a=T") {
		t.Fatalf("expected kitty image data")
	}
}

func TestRenderWithPreviewBottom(t *testing.T) {
	state := &appState{
		mode:    modeBrowse,
		results: []model.Result{{Title: "A"}},
		inline:  termcaps.InlineKitty,
		currentAnim: &gifAnimation{
			ID:     2,
			Frames: []gifdecode.Frame{{PNG: []byte{1, 2, 3}, Delay: 80 * time.Millisecond}},
			Width:  200,
			Height: 100,
		},
		previewNeedsSend: true,
		opts:             model.Options{Source: "tenor"},
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	render(state, out, 24, 60)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "Preview") {
		t.Fatalf("expected preview label")
	}
	if !strings.Contains(buf.String(), "a=T") {
		t.Fatalf("expected kitty image data")
	}
}

func TestDrawPreviewPlacement(t *testing.T) {
	state := &appState{
		inline: termcaps.InlineKitty,
		currentAnim: &gifAnimation{
			ID:     3,
			Frames: []gifdecode.Frame{{PNG: []byte{1, 2, 3}, Delay: 80 * time.Millisecond}},
		},
		previewNeedsSend: false,
		previewDirty:     false,
		activeImageID:    3,
		opts:             model.Options{Source: "tenor"},
		lastPreview: struct {
			cols int
			rows int
		}{cols: 1, rows: 1},
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	drawPreview(state, out, 10, 6, 2, 2)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "a=p") {
		t.Fatalf("expected kitty placement")
	}
}

func TestWriteLineAtClears(t *testing.T) {
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	writeLineAt(out, 1, 1, "hi", 0)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "\x1b[K") {
		t.Fatalf("expected clear line")
	}
}

func TestDrawHintsDoesNotColorWords(t *testing.T) {
	state := &appState{
		useColor: true,
	}
	layout := layout{rows: 10, cols: 120, hintsRow: 10, listCol: 1, listWidth: 120}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	drawHints(out, state, layout)
	_ = out.Flush()

	text := buf.String()
	if strings.Contains(text, "\x1b[1m\x1b[36mD") {
		t.Fatalf("unexpected coloring inside words")
	}
	if strings.Contains(text, "D\x1b[0mownload") || strings.Contains(text, "E\x1b[0mdit") {
		t.Fatalf("unexpected ANSI reset inside words")
	}
	if !strings.Contains(text, "Download") || !strings.Contains(text, "Edit") {
		t.Fatalf("expected hint labels")
	}
}
