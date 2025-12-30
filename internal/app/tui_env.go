package app

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

type tuiEnv struct {
	in         io.Reader
	out        io.Writer
	fd         int
	isTerminal func(int) bool
	makeRaw    func(int) (*term.State, error)
	restore    func(int, *term.State) error
	getSize    func(int) (int, int, error)
	signalCh   <-chan os.Signal
}

var defaultEnvFn = defaultTUIEnv

func defaultTUIEnv() tuiEnv {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	return tuiEnv{
		in:         os.Stdin,
		out:        os.Stdout,
		fd:         int(os.Stdin.Fd()),
		isTerminal: term.IsTerminal,
		makeRaw:    term.MakeRaw,
		restore:    term.Restore,
		getSize:    term.GetSize,
		signalCh:   sigs,
	}
}
