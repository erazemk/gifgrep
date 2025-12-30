package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/steipete/gifgrep/internal/model"
)

func TestSanitizeFilename(t *testing.T) {
	name := sanitizeFilename("Hello / weird:name?.gif")
	if strings.ContainsAny(name, "/:\\") {
		t.Fatalf("unexpected separators: %q", name)
	}
	if name == "" {
		t.Fatalf("expected name")
	}
}

func TestUniqueFilePath(t *testing.T) {
	dir := t.TempDir()
	first, err := uniqueFilePath(dir, "sample.gif")
	if err != nil {
		t.Fatalf("uniqueFilePath: %v", err)
	}
	if err := os.WriteFile(first, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	second, err := uniqueFilePath(dir, "sample.gif")
	if err != nil {
		t.Fatalf("uniqueFilePath: %v", err)
	}
	if first == second {
		t.Fatalf("expected unique path")
	}
	if filepath.Dir(second) != dir {
		t.Fatalf("expected same dir")
	}
}

func TestFilenameForResult(t *testing.T) {
	item := model.Result{Title: "My Cool GIF"}
	name := filenameForResult(item)
	if !strings.HasSuffix(strings.ToLower(name), ".gif") {
		t.Fatalf("expected gif extension")
	}
}
