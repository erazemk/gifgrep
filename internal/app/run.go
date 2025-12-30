package app

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/steipete/gifgrep/internal/tui"
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
		if opts.GifInput != "" || opts.StillSet || opts.StillsCount > 0 {
			_, _ = fmt.Fprintln(os.Stderr, "--tui cannot be combined with --gif/--still/--stills")
			return 1
		}
		if err := tui.Run(opts, query); err != nil {
			if errors.Is(err, tui.ErrNotTerminal) {
				_, _ = fmt.Fprintln(os.Stderr, "stdin is not a tty")
			} else {
				_, _ = fmt.Fprintln(os.Stderr, err.Error())
			}
			return 1
		}
		return 0
	}

	if opts.GifInput != "" || opts.StillSet || opts.StillsCount > 0 {
		if err := runExtract(opts); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
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
