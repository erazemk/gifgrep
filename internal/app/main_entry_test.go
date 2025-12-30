package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"golang.org/x/term"
)

func TestRunArgs(t *testing.T) {
	t.Run("version", func(t *testing.T) {
		if code := Run([]string{"--version"}); code != 0 {
			t.Fatalf("expected exit 0")
		}
	})

	t.Run("help", func(t *testing.T) {
		if code := Run([]string{"--help"}); code != 0 {
			t.Fatalf("expected exit 0")
		}
	})

	t.Run("empty", func(t *testing.T) {
		if code := Run(nil); code != 1 {
			t.Fatalf("expected exit 1")
		}
	})

	t.Run("bad args", func(t *testing.T) {
		if code := Run([]string{"--nope"}); code != 1 {
			t.Fatalf("expected exit 1")
		}
	})

	t.Run("bad source", func(t *testing.T) {
		if code := Run([]string{"--source", "nope", "cats"}); code != 1 {
			t.Fatalf("expected exit 1")
		}
	})

	t.Run("tui", func(t *testing.T) {
		origEnvFn := defaultEnvFn
		defer func() { defaultEnvFn = origEnvFn }()
		defaultEnvFn = func() tuiEnv {
			return tuiEnv{
				in:         bytes.NewReader([]byte("q")),
				out:        io.Discard,
				fd:         1,
				isTerminal: func(int) bool { return true },
				makeRaw:    func(int) (*term.State, error) { return &term.State{}, nil },
				restore:    func(int, *term.State) error { return nil },
				getSize:    func(int) (int, int, error) { return 80, 24, nil },
				signalCh:   make(chan os.Signal),
			}
		}
		if code := Run([]string{"--tui"}); code != 0 {
			t.Fatalf("expected exit 0")
		}
	})
}

func TestDefaultEnv(t *testing.T) {
	env := defaultTUIEnv()
	if env.in == nil || env.out == nil || env.signalCh == nil {
		t.Fatalf("expected env to be initialized")
	}
	if env.isTerminal == nil || env.makeRaw == nil || env.restore == nil || env.getSize == nil {
		t.Fatalf("expected env functions")
	}
}

func TestRunTUIModes(t *testing.T) {
	origEnvFn := defaultEnvFn
	defer func() { defaultEnvFn = origEnvFn }()

	t.Run("not terminal", func(t *testing.T) {
		defaultEnvFn = func() tuiEnv {
			return tuiEnv{
				in:  bytes.NewReader(nil),
				out: io.Discard,
				fd:  1,
				isTerminal: func(int) bool {
					return false
				},
			}
		}
		if code := Run([]string{"--tui"}); code != 1 {
			t.Fatalf("expected exit 1")
		}
	})

	t.Run("ok", func(t *testing.T) {
		defaultEnvFn = func() tuiEnv {
			return tuiEnv{
				in:         bytes.NewReader([]byte("q")),
				out:        io.Discard,
				fd:         1,
				isTerminal: func(int) bool { return true },
				makeRaw: func(int) (*term.State, error) {
					return &term.State{}, nil
				},
				restore:  func(int, *term.State) error { return nil },
				getSize:  func(int) (int, int, error) { return 80, 24, nil },
				signalCh: make(chan os.Signal),
			}
		}
		if err := runTUI(cliOptions{}, ""); err != nil {
			t.Fatalf("expected no error")
		}
	})
}

func TestRunTUIWithFailures(t *testing.T) {
	env := tuiEnv{
		in:         bytes.NewReader(nil),
		out:        io.Discard,
		fd:         1,
		isTerminal: func(int) bool { return true },
		makeRaw: func(int) (*term.State, error) {
			return nil, errors.New("boom")
		},
	}
	if err := runTUIWith(env, cliOptions{}, ""); err == nil {
		t.Fatalf("expected error")
	}
}
