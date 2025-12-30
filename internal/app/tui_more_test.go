package app

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"golang.org/x/term"
)

func TestRunTUIWithDefaultsNotTerminal(t *testing.T) {
	if err := runTUIWith(tuiEnv{}, cliOptions{}, ""); !errors.Is(err, errNotTerminal) {
		t.Fatalf("expected errNotTerminal")
	}
}

func TestRunTUIWithNoRestore(t *testing.T) {
	env := tuiEnv{
		in:         bytes.NewReader([]byte("q")),
		out:        io.Discard,
		fd:         1,
		isTerminal: func(int) bool { return true },
		makeRaw:    func(int) (*term.State, error) { return nil, nil },
		getSize:    func(int) (int, int, error) { return 80, 24, nil },
		signalCh:   make(chan os.Signal),
	}
	if err := runTUIWith(env, cliOptions{}, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTUIWithSearchError(t *testing.T) {
	env := tuiEnv{
		in:         bytes.NewReader([]byte("q")),
		out:        io.Discard,
		fd:         1,
		isTerminal: func(int) bool { return true },
		makeRaw:    func(int) (*term.State, error) { return &term.State{}, nil },
		restore:    func(int, *term.State) error { return nil },
		getSize:    func(int) (int, int, error) { return 80, 24, nil },
		signalCh:   make(chan os.Signal),
	}
	if err := runTUIWith(env, cliOptions{Source: "nope"}, "cats"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunTUIWithSizeError(t *testing.T) {
	env := tuiEnv{
		in:         bytes.NewReader([]byte("q")),
		out:        io.Discard,
		fd:         1,
		isTerminal: func(int) bool { return true },
		makeRaw:    func(int) (*term.State, error) { return &term.State{}, nil },
		restore:    func(int, *term.State) error { return nil },
		getSize:    func(int) (int, int, error) { return 0, 0, errors.New("bad") },
		signalCh:   make(chan os.Signal),
	}
	if err := runTUIWith(env, cliOptions{}, ""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

type emptyTenorTransport struct{}

func (t *emptyTenorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := `{"results":[]}`
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestRunTUIWithEmptyResultsAndSignal(t *testing.T) {
	withTransport(t, &emptyTenorTransport{}, func() {
		sigs := make(chan os.Signal, 1)
		sigs <- os.Interrupt
		env := tuiEnv{
			in:         bytes.NewReader([]byte("q")),
			out:        io.Discard,
			fd:         1,
			isTerminal: func(int) bool { return true },
			makeRaw:    func(int) (*term.State, error) { return &term.State{}, nil },
			restore:    func(int, *term.State) error { return nil },
			getSize:    func(int) (int, int, error) { return 80, 24, nil },
			signalCh:   sigs,
		}
		if err := runTUIWith(env, cliOptions{Source: "tenor"}, "cats"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
