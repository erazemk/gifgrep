package app

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestHelpers(t *testing.T) {
	if truncateRunes("héllö", 3) != "hél" {
		t.Fatalf("truncateRunes failed")
	}
	if truncateRunes("hello", 0) != "" {
		t.Fatalf("truncateRunes width 0 failed")
	}
	if runeLen("héllö") != 5 {
		t.Fatalf("runeLen failed")
	}
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	moveCursor(out, 2, 3)
	saveCursor(out)
	restoreCursor(out)
	hideCursor(out)
	showCursor(out)
	clearImages(out)
	_ = out.Flush()
	s := buf.String()
	if !strings.Contains(s, "\x1b[2;3H") || !strings.Contains(s, "\x1b[?25l") || !strings.Contains(s, "\x1b[?25h") {
		t.Fatalf("cursor sequences missing")
	}
	if !strings.Contains(s, "\x1b_Ga=d\x1b\\") {
		t.Fatalf("clear images missing")
	}
	buf.Reset()
	moveCursor(out, 0, 0)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "\x1b[1;1H") {
		t.Fatalf("expected clamped cursor move")
	}
	if clampDelay(5) != 20 || clampDelay(2000) != 1000 || clampDelay(300) != 300 {
		t.Fatalf("clampDelay failed")
	}
	if maxInt(3, 2) != 3 || maxInt(1, 4) != 4 {
		t.Fatalf("max failed")
	}
	if minInt(3, 2) != 2 || minInt(1, 4) != 1 {
		t.Fatalf("min failed")
	}

	t.Setenv("GIFGREP_CELL_ASPECT", "0.7")
	if cellAspectRatio() != 0.7 {
		t.Fatalf("cellAspectRatio env override failed")
	}

	t.Setenv("GIFGREP_SOFTWARE_ANIM", "true")
	if !useSoftwareAnimation() {
		t.Fatalf("expected software animation")
	}

	t.Setenv("GIFGREP_SOFTWARE_ANIM", "")
	t.Setenv("TERM_PROGRAM", "xterm")
	t.Setenv("TERM", "xterm-256color")
	if useSoftwareAnimation() {
		t.Fatalf("expected software animation to be false")
	}
}
