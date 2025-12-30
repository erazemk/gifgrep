package app

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/search"
)

var (
	errHelp    = errors.New("help")
	errVersion = errors.New("version")
)

func parseArgs(args []string) (model.Options, string, error) {
	var opts model.Options
	var showHelp bool
	var showVersion bool
	var stillRaw string
	fs := flag.NewFlagSet(model.AppName, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.BoolVar(&showHelp, "help", false, "help")
	fs.BoolVar(&showHelp, "h", false, "help")
	fs.BoolVar(&showVersion, "version", false, "version")
	fs.BoolVar(&opts.TUI, "tui", false, "interactive mode")
	fs.BoolVar(&opts.JSON, "json", false, "json output")
	fs.BoolVar(&opts.IgnoreCase, "i", false, "ignore case")
	fs.BoolVar(&opts.Invert, "v", false, "invert vibe")
	fs.BoolVar(&opts.Regex, "E", false, "regex search")
	fs.BoolVar(&opts.Number, "n", false, "number results")
	fs.IntVar(&opts.Limit, "m", 20, "max results")
	fs.StringVar(&opts.Source, "source", "auto", "source: auto|tenor|giphy")
	fs.StringVar(&opts.Mood, "mood", "", "mood filter")
	fs.StringVar(&opts.Color, "color", "auto", "color: auto|always|never")
	fs.StringVar(&opts.GifInput, "gif", "", "gif input path or URL")
	fs.StringVar(&stillRaw, "still", "", "extract still at time (e.g. 1.5s)")
	fs.IntVar(&opts.StillsCount, "stills", 0, "contact sheet frame count")
	fs.IntVar(&opts.StillsCols, "stills-cols", 0, "contact sheet columns")
	fs.IntVar(&opts.StillsPadding, "stills-padding", 2, "contact sheet padding (px)")
	fs.StringVar(&opts.OutPath, "out", "", "output path or '-' for stdout")

	if err := fs.Parse(args); err != nil {
		return opts, "", errors.New("bad args")
	}

	if showHelp {
		printUsage(os.Stdout)
		return opts, "", errHelp
	}
	if showVersion {
		_, _ = fmt.Fprintf(os.Stdout, "%s %s\n", model.AppName, model.Version)
		return opts, "", errVersion
	}

	if stillRaw != "" {
		parsed, err := parseDurationValue(stillRaw)
		if err != nil {
			return opts, "", errors.New("bad args")
		}
		opts.StillSet = true
		opts.StillAt = parsed
	}

	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	return opts, query, nil
}

func printUsage(w io.Writer) {
	_, _ = fmt.Fprintf(w, "%s %s\n\n", model.AppName, model.Version)
	_, _ = fmt.Fprintln(w, "Usage:")
	_, _ = fmt.Fprintln(w, "  gifgrep [flags] <query>")
	_, _ = fmt.Fprintln(w, "  gifgrep --tui [flags] <query>")
	_, _ = fmt.Fprintln(w, "  gifgrep --gif <path|url> --still <time> [--out <file>]")
	_, _ = fmt.Fprintln(w, "  gifgrep --gif <path|url> --stills <N> [--stills-cols <N>] [--out <file>]")
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Flags:")
	_, _ = fmt.Fprintln(w, "  -i            ignore case")
	_, _ = fmt.Fprintln(w, "  -v            invert vibe (exclude mood)")
	_, _ = fmt.Fprintln(w, "  -E            regex filter over title+tags")
	_, _ = fmt.Fprintln(w, "  -n            number results")
	_, _ = fmt.Fprintln(w, "  -m <N>        max results")
	_, _ = fmt.Fprintln(w, "  --mood <s>    mood filter")
	_, _ = fmt.Fprintln(w, "  --json        json output")
	_, _ = fmt.Fprintln(w, "  --tui         interactive mode")
	_, _ = fmt.Fprintln(w, "  --source <s>  source (auto, tenor, giphy)")
	_, _ = fmt.Fprintln(w, "  --gif <s>     gif input path or URL")
	_, _ = fmt.Fprintln(w, "  --still <s>   extract still at time (e.g. 1.5s)")
	_, _ = fmt.Fprintln(w, "  --stills <N>  contact sheet frame count")
	_, _ = fmt.Fprintln(w, "  --stills-cols <N>    contact sheet columns")
	_, _ = fmt.Fprintln(w, "  --stills-padding <N> contact sheet padding (px)")
	_, _ = fmt.Fprintln(w, "  --out <s>     output path or '-' for stdout")
	_, _ = fmt.Fprintln(w, "  --version     show version")
	_, _ = fmt.Fprintln(w, "  -h, --help    show help")
}

func parseDurationValue(raw string) (time.Duration, error) {
	if raw == "" {
		return 0, errors.New("empty duration")
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d, nil
	}
	if secs, err := strconv.ParseFloat(raw, 64); err == nil {
		if secs < 0 {
			return 0, errors.New("negative duration")
		}
		return time.Duration(secs * float64(time.Second)), nil
	}
	return 0, errors.New("invalid duration")
}

func runScript(opts model.Options, query string) error {
	results, err := search.Search(query, opts)
	if err != nil {
		return err
	}
	results, err = search.FilterResults(results, query, opts)
	if err != nil {
		return err
	}

	if opts.JSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for i, res := range results {
		prefix := ""
		if opts.Number {
			prefix = fmt.Sprintf("%d\t", i+1)
		}
		_, _ = fmt.Fprintf(os.Stdout, "%s%s\n", prefix, res.URL)
	}
	return nil
}
