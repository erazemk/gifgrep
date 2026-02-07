package clipboard

import (
	"errors"
	"testing"
)

func TestCopyFileEmptyPath(t *testing.T) {
	if err := CopyFile(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestCopyCommandDarwin(t *testing.T) {
	cmd, args, err := copyCommand("darwin", "/tmp/a.gif")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cmd != "osascript" {
		t.Fatalf("expected osascript, got %q", cmd)
	}
	if len(args) != 2 || args[0] != "-e" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestCopyCommandLinuxXclip(t *testing.T) {
	prev := lookPath
	lookPath = func(name string) (string, error) {
		if name == "xclip" {
			return "/usr/bin/xclip", nil
		}
		return "", errors.New("not found")
	}
	t.Cleanup(func() { lookPath = prev })

	cmd, args, err := copyCommand("linux", "/tmp/a.gif")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cmd != "xclip" {
		t.Fatalf("expected xclip, got %q", cmd)
	}
	if len(args) != 6 || args[4] != "-i" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestCopyCommandLinuxWlCopy(t *testing.T) {
	prev := lookPath
	lookPath = func(name string) (string, error) {
		if name == "wl-copy" {
			return "/usr/bin/wl-copy", nil
		}
		return "", errors.New("not found")
	}
	t.Cleanup(func() { lookPath = prev })

	cmd, args, err := copyCommand("linux", "/tmp/a.gif")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if cmd != "sh" {
		t.Fatalf("expected sh, got %q", cmd)
	}
	if len(args) != 2 || args[0] != "-c" {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestCopyCommandLinuxNoTool(t *testing.T) {
	prev := lookPath
	lookPath = func(string) (string, error) { return "", errors.New("not found") }
	t.Cleanup(func() { lookPath = prev })

	_, _, err := copyCommand("linux", "/tmp/a.gif")
	if err == nil {
		t.Fatal("expected error when no tool available")
	}
}
