package app

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/steipete/gifgrep/gifdecode"
	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/termcaps"
)

var cursorMoves = regexp.MustCompile(`\x1b\[[0-9]+[ACG]`)

func stripItermCursorMoves(s string) string {
	s = strings.ReplaceAll(s, "\x1b7", "")
	s = strings.ReplaceAll(s, "\x1b8", "")
	s = strings.ReplaceAll(s, "\x1b[K", "")
	s = strings.ReplaceAll(s, "\r", "")
	s = cursorMoves.ReplaceAllString(s, "")
	return s
}

func TestResolveOutputFormatAutoTTY(t *testing.T) {
	prev := isTerminalWriter
	isTerminalWriter = func(_ io.Writer) bool { return true }
	t.Cleanup(func() { isTerminalWriter = prev })

	got := resolveOutputFormat(model.Options{Format: "auto"}, &bytes.Buffer{})
	if got != formatPlain {
		t.Fatalf("expected plain, got %q", got)
	}
}

func TestResolveOutputFormatAutoNonTTY(t *testing.T) {
	prev := isTerminalWriter
	isTerminalWriter = func(_ io.Writer) bool { return false }
	t.Cleanup(func() { isTerminalWriter = prev })

	got := resolveOutputFormat(model.Options{Format: "auto"}, &bytes.Buffer{})
	if got != formatURL {
		t.Fatalf("expected url, got %q", got)
	}
}

func TestRenderPlainNoThumbs(t *testing.T) {
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	renderPlain(out, model.Options{Number: true}, false, termcaps.InlineNone, []model.Result{
		{Title: "A dog", URL: "https://example.test/a.gif"},
	}, 0)
	_ = out.Flush()

	text := buf.String()
	if !strings.Contains(text, "1. A dog") {
		t.Fatalf("missing title: %q", text)
	}
	if !strings.Contains(text, "  https://example.test/a.gif") {
		t.Fatalf("missing url: %q", text)
	}
}

func TestRenderPlainThumbsIncludesBlockGap(t *testing.T) {
	prevFetch := fetchThumb
	prevDecode := decodeThumb
	prevSend := sendThumbKitty
	t.Cleanup(func() {
		fetchThumb = prevFetch
		decodeThumb = prevDecode
		sendThumbKitty = prevSend
	})

	fetchThumb = func(_ string) ([]byte, error) { return []byte("gif"), nil }
	decodeThumb = func(_ []byte) (*gifdecode.Frames, error) {
		return &gifdecode.Frames{Frames: []gifdecode.Frame{{PNG: []byte{1}}}}, nil
	}
	sendThumbKitty = func(out *bufio.Writer, id uint32, _ gifdecode.Frame, _, _ int) {
		_, _ = fmt.Fprintf(out, "<IMG%d>", id)
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	renderPlain(out, model.Options{}, false, termcaps.InlineKitty, []model.Result{
		{Title: "A", URL: "https://example.test/a.gif"},
		{Title: "B", URL: "https://example.test/b.gif"},
	}, 0)
	_ = out.Flush()

	text := buf.String()
	if !strings.Contains(text, "<IMG1>") || !strings.Contains(text, "<IMG2>") {
		t.Fatalf("expected image markers: %q", text)
	}
	if !strings.Contains(text, "\n\n<IMG2>") {
		t.Fatalf("expected blank line between thumb blocks: %q", text)
	}
	if strings.Contains(text, "\n\n\n<IMG2>") {
		t.Fatalf("unexpected extra blank line between thumb blocks: %q", text)
	}
}

func TestRenderPlainThumbsItermUsesRawGIF(t *testing.T) {
	prevFetch := fetchThumb
	prevDecode := decodeThumb
	prevSend := sendThumbIterm
	t.Cleanup(func() {
		fetchThumb = prevFetch
		decodeThumb = prevDecode
		sendThumbIterm = prevSend
	})

	gifData := []byte("GIF89a\x01\x00\x01\x00")
	fetchThumb = func(_ string) ([]byte, error) { return gifData, nil }
	decodeThumb = func(_ []byte) (*gifdecode.Frames, error) {
		t.Fatalf("decodeThumb should not be called for iTerm")
		return nil, nil
	}
	sendThumbIterm = func(out *bufio.Writer, data []byte, _, _ int) {
		if string(data) != string(gifData) {
			t.Fatalf("unexpected inline data")
		}
		_, _ = fmt.Fprint(out, "<ITERM>")
	}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	renderPlain(out, model.Options{}, false, termcaps.InlineIterm, []model.Result{
		{Title: "A", URL: "https://example.test/a.gif"},
	}, 80)
	_ = out.Flush()

	text := buf.String()
	if !strings.Contains(text, "<ITERM>") {
		t.Fatalf("expected iTerm marker: %q", text)
	}
	text = stripItermCursorMoves(text)
	if !strings.HasPrefix(text, "<ITERM>A\n") {
		t.Fatalf("expected title to follow iTerm marker: %q", text)
	}
}

func TestRenderPlainThumbsItermWrapsURLToTerminalWidth(t *testing.T) {
	prevFetch := fetchThumb
	prevSend := sendThumbIterm
	t.Cleanup(func() {
		fetchThumb = prevFetch
		sendThumbIterm = prevSend
	})

	gifData := []byte("GIF89a\x01\x00\x01\x00")
	fetchThumb = func(_ string) ([]byte, error) { return gifData, nil }
	sendThumbIterm = func(_ *bufio.Writer, _ []byte, _, _ int) {}

	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)

	termCols := 40
	url := strings.Repeat("a", 30)
	renderPlain(out, model.Options{}, false, termcaps.InlineIterm, []model.Result{
		{Title: "T", URL: url},
	}, termCols)
	_ = out.Flush()

	text := stripItermCursorMoves(strings.TrimSuffix(buf.String(), "\n"))
	lines := strings.Split(text, "\n")
	if len(lines) != 8 {
		t.Fatalf("expected 8 lines, got %d: %q", len(lines), text)
	}

	// rows=8, cols=16, indent=cols, so text width = 40-16-1 = 23 chars.
	if got := lines[1]; got != strings.Repeat("a", 23) {
		t.Fatalf("unexpected first url line: %q", got)
	}
	if got := lines[2]; got != strings.Repeat("a", 7) {
		t.Fatalf("unexpected second url line: %q", got)
	}
}
