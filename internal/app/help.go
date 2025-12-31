package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/steipete/gifgrep/internal/model"
	"golang.org/x/term"
)

func helpPrinter(options kong.HelpOptions, ctx *kong.Context) error {
	useColor := helpWantsColor(ctx)
	_, _ = fmt.Fprintln(ctx.Stdout, helpHeader(useColor))
	_, _ = fmt.Fprintln(ctx.Stdout, helpTagline(useColor))
	_, _ = fmt.Fprintln(ctx.Stdout)

	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}

	if options.Summary {
		return nil
	}
	lines := helpExtras(ctx)
	if len(lines) == 0 {
		return nil
	}
	_, _ = fmt.Fprintln(ctx.Stdout)
	for _, line := range lines {
		_, _ = fmt.Fprintln(ctx.Stdout, line)
	}
	return nil
}

func helpHeader(useColor bool) string {
	if !useColor {
		return fmt.Sprintf("%s %s", model.AppName, model.Version)
	}
	return "\x1b[1m\x1b[36m" + model.AppName + "\x1b[0m" + " " + "\x1b[1m" + model.Version + "\x1b[0m"
}

func helpTagline(useColor bool) string {
	if !useColor {
		return model.Tagline
	}
	return "\x1b[90m" + model.Tagline + "\x1b[0m"
}

func helpWantsColor(ctx *kong.Context) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	termEnv := strings.ToLower(strings.TrimSpace(os.Getenv("TERM")))
	if termEnv == "" || termEnv == "dumb" {
		return false
	}

	for i := 0; i < len(ctx.Args); i++ {
		arg := ctx.Args[i]
		if arg == "--no-color" {
			return false
		}
		if strings.HasPrefix(arg, "--color=") {
			val := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(arg, "--color=")))
			if val == "never" {
				return false
			}
			if val == "always" {
				return true
			}
		}
		if arg == "--color" && i+1 < len(ctx.Args) {
			val := strings.ToLower(strings.TrimSpace(ctx.Args[i+1]))
			if val == "never" {
				return false
			}
			if val == "always" {
				return true
			}
			i++
		}
	}

	f, ok := ctx.Stdout.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func helpExtras(ctx *kong.Context) []string {
	selected := ctx.Selected()
	if selected == nil {
		return rootHelpExtras()
	}
	switch selected.Name {
	case "search":
		return searchHelpExtras()
	case "tui":
		return tuiHelpExtras()
	case "still":
		return stillHelpExtras()
	case "sheet":
		return sheetHelpExtras()
	default:
		return rootHelpExtras()
	}
}

func rootHelpExtras() []string {
	return []string{
		"Examples:",
		"  gifgrep cats",
		"  gifgrep search --json cats | jq '.[0].url'",
		"  gifgrep tui cats",
		"  gifgrep still cat.gif --at 1.5s -o still.png",
		"  gifgrep sheet cat.gif --frames 12 --cols 4 -o sheet.png",
		"",
		"Environment:",
		"  TENOR_API_KEY  optional (defaults to Tenor demo key)",
		"  GIPHY_API_KEY  required for --source giphy",
	}
}

func searchHelpExtras() []string {
	return []string{
		"Output:",
		"  Default: <title>\\t<url>",
		"",
		"Examples:",
		"  gifgrep cats | head -n 5",
		"  gifgrep search --json cats | jq '.[] | .url'",
		"  GIPHY_API_KEY=... gifgrep search --source giphy cats",
	}
}

func tuiHelpExtras() []string {
	return []string{
		"Keys:",
		"  /      edit search",
		"  ↑↓     select",
		"  d      download selection",
		"  q      quit",
		"",
		"Examples:",
		"  gifgrep tui cats",
	}
}

func stillHelpExtras() []string {
	return []string{
		"Examples:",
		"  gifgrep still cat.gif --at 0 -o still.png",
		"  gifgrep still https://example.com/cat.gif --at 1.25s -o - > still.png",
	}
}

func sheetHelpExtras() []string {
	return []string{
		"Examples:",
		"  gifgrep sheet cat.gif -o sheet.png",
		"  gifgrep sheet cat.gif --frames 16 --cols 4 --padding 4 -o sheet.png",
	}
}
