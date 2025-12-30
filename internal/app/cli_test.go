package app

import (
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestParseArgs(t *testing.T) {
	_, _, err := parseArgs([]string{"--help"})
	if !errors.Is(err, errHelp) {
		t.Fatalf("expected errHelp, got %v", err)
	}

	_, _, err = parseArgs([]string{"--version"})
	if !errors.Is(err, errVersion) {
		t.Fatalf("expected errVersion, got %v", err)
	}

	opts, query, err := parseArgs([]string{"--tui", "-i", "-v", "-E", "-n", "-m", "5", "--source", "tenor", "--mood", "angry", "--color", "always", "cats"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.TUI || !opts.IgnoreCase || !opts.Invert || !opts.Regex || !opts.Number {
		t.Fatalf("flags not parsed")
	}
	if opts.Limit != 5 || opts.Source != "tenor" || opts.Mood != "angry" || opts.Color != "always" {
		t.Fatalf("options not parsed")
	}
	if query != "cats" {
		t.Fatalf("unexpected query: %q", query)
	}

	_, _, err = parseArgs([]string{"--nope"})
	if err == nil {
		t.Fatalf("expected error for bad args")
	}
}

func TestRunScriptOutput(t *testing.T) {
	gifData := makeTestGIF()
	withTransport(t, &fakeTransport{gifData: gifData}, func() {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		t.Cleanup(func() {
			os.Stdout = oldStdout
		})

		err := runScript(cliOptions{Number: true, Limit: 1, Source: "tenor"}, "cats")
		_ = w.Close()
		if err != nil {
			t.Fatalf("runScript failed: %v", err)
		}
		out, _ := io.ReadAll(r)
		if !strings.Contains(string(out), "1\t") {
			t.Fatalf("expected numbered output")
		}

		r2, w2, _ := os.Pipe()
		os.Stdout = w2
		err = runScript(cliOptions{JSON: true, Limit: 1, Source: "tenor"}, "cats")
		_ = w2.Close()
		if err != nil {
			t.Fatalf("runScript json failed: %v", err)
		}
		out2, _ := io.ReadAll(r2)
		if !strings.Contains(string(out2), "\"preview_url\"") {
			t.Fatalf("expected json output")
		}
	})
}

func TestRunScriptError(t *testing.T) {
	if err := runScript(cliOptions{Source: "nope"}, "cats"); err == nil {
		t.Fatalf("expected error")
	}
}
