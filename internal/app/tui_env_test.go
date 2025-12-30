package app

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"golang.org/x/term"
)

func TestRunTUIWithQuit(t *testing.T) {
	var restored bool
	env := tuiEnv{
		in:  bytes.NewReader([]byte("q")),
		out: io.Discard,
		fd:  1,
		isTerminal: func(int) bool {
			return true
		},
		makeRaw: func(int) (*term.State, error) {
			return &term.State{}, nil
		},
		restore: func(int, *term.State) error {
			restored = true
			return nil
		},
		getSize: func(int) (int, int, error) {
			return 80, 24, nil
		},
		signalCh: make(chan os.Signal),
	}

	if err := runTUIWith(env, cliOptions{Source: "tenor"}, ""); err != nil {
		t.Fatalf("runTUIWith failed: %v", err)
	}
	if !restored {
		t.Fatalf("expected restore to be called")
	}
}

func TestRunTUIWithSearch(t *testing.T) {
	gifData := makeTestGIF()
	withTransport(t, &fakeTransport{gifData: gifData}, func() {
		env := tuiEnv{
			in:  bytes.NewReader([]byte("q")),
			out: io.Discard,
			fd:  1,
			isTerminal: func(int) bool {
				return true
			},
			makeRaw: func(int) (*term.State, error) {
				return &term.State{}, nil
			},
			restore: func(int, *term.State) error {
				return nil
			},
			getSize: func(int) (int, int, error) {
				return 80, 24, nil
			},
			signalCh: make(chan os.Signal),
		}

		if err := runTUIWith(env, cliOptions{Source: "tenor", Limit: 1}, "cats"); err != nil {
			t.Fatalf("runTUIWith search failed: %v", err)
		}
	})
}

func TestRunTUIWithNotTerminal(t *testing.T) {
	env := tuiEnv{
		in:  bytes.NewReader(nil),
		out: io.Discard,
		fd:  1,
		isTerminal: func(int) bool {
			return false
		},
	}

	if err := runTUIWith(env, cliOptions{}, ""); !errors.Is(err, errNotTerminal) {
		t.Fatalf("expected errNotTerminal, got %v", err)
	}
}
