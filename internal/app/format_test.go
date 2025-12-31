package app

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/steipete/gifgrep/internal/model"
)

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

	renderPlain(out, model.Options{Number: true}, false, false, []model.Result{
		{Title: "A dog", URL: "https://example.test/a.gif"},
	})
	_ = out.Flush()

	text := buf.String()
	if !strings.Contains(text, "1. A dog") {
		t.Fatalf("missing title: %q", text)
	}
	if !strings.Contains(text, "  https://example.test/a.gif") {
		t.Fatalf("missing url: %q", text)
	}
}
