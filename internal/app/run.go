package app

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func Run(args []string) int {
	opts, query, err := parseArgs(args)
	if err != nil {
		if errors.Is(err, errHelp) || errors.Is(err, errVersion) {
			return 0
		}
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}

	if opts.TUI {
		if err := runTUI(opts, query); err != nil {
			if errors.Is(err, errNotTerminal) {
				_, _ = fmt.Fprintln(os.Stderr, "stdin is not a tty")
			} else {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
			}
			return 1
		}
		return 0
	}

	if strings.TrimSpace(query) == "" {
		printUsage(os.Stderr)
		return 1
	}

	if err := runScript(opts, query); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
