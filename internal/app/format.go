package app

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/steipete/gifgrep/gifdecode"
	"github.com/steipete/gifgrep/internal/iterm"
	"github.com/steipete/gifgrep/internal/kitty"
	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/termcaps"
	"golang.org/x/term"
)

type outputFormat string

const (
	formatAuto    outputFormat = "auto"
	formatPlain   outputFormat = "plain"
	formatTSV     outputFormat = "tsv"
	formatMD      outputFormat = "md"
	formatURL     outputFormat = "url"
	formatComment outputFormat = "comment"
	formatJSON    outputFormat = "json"
)

type thumbsMode string

const (
	thumbsAuto   thumbsMode = "auto"
	thumbsAlways thumbsMode = "always"
	thumbsNever  thumbsMode = "never"
)

var isTerminalWriter = func(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

var (
	fetchThumb  = fetchURL
	decodeThumb = func(data []byte) (*gifdecode.Frames, error) {
		decodeOpts := gifdecode.DefaultOptions()
		decodeOpts.MaxFrames = 1
		return gifdecode.Decode(data, decodeOpts)
	}
	sendThumbKitty = func(out *bufio.Writer, id uint32, frame gifdecode.Frame, cols, rows int) {
		kitty.SendFrame(out, id, frame, cols, rows)
	}
	sendThumbIterm = func(out *bufio.Writer, data []byte, cols, rows int) {
		iterm.SendInlineFile(out, iterm.File{
			Name:        thumbInlineName(data),
			Data:        data,
			WidthCells:  cols,
			HeightCells: rows,
			Stretch:     true,
		})
	}
)

func resolveOutputFormat(opts model.Options, stdout io.Writer) outputFormat {
	if opts.JSON {
		return formatJSON
	}
	f := outputFormat(strings.ToLower(strings.TrimSpace(opts.Format)))
	if f == "" || f == formatAuto {
		if isTerminalWriter(stdout) {
			return formatPlain
		}
		return formatURL
	}
	return f
}

func resolveThumbsMode(opts model.Options) thumbsMode {
	m := thumbsMode(strings.ToLower(strings.TrimSpace(opts.Thumbs)))
	if m == "" {
		return thumbsAuto
	}
	return m
}

func thumbsProtocol(opts model.Options, stdout io.Writer, format outputFormat) termcaps.InlineProtocol {
	if format != formatPlain {
		return termcaps.InlineNone
	}
	if !isTerminalWriter(stdout) {
		return termcaps.InlineNone
	}
	switch resolveThumbsMode(opts) {
	case thumbsNever:
		return termcaps.InlineNone
	case thumbsAlways:
		return termcaps.DetectInlineRobust(os.Getenv)
	case thumbsAuto:
		return termcaps.DetectInlineRobust(os.Getenv)
	}
	return termcaps.DetectInlineRobust(os.Getenv)
}

func renderPlain(
	out *bufio.Writer,
	opts model.Options,
	useColor bool,
	thumbs termcaps.InlineProtocol,
	results []model.Result,
	termCols int,
) {
	nextID := uint32(1)
	withThumbs := thumbs != termcaps.InlineNone
	for i, res := range results {
		title := normalizeTitle(res)
		url := res.URL

		nPrefix := ""
		if opts.Number {
			nPrefix = fmt.Sprintf("%d. ", i+1)
		}

		if withThumbs && renderThumbBlock(out, thumbs, nextID, res, nPrefix, title, url, useColor, termCols) == nil {
			nextID++
			if i < len(results)-1 {
				if thumbs == termcaps.InlineIterm {
					_, _ = fmt.Fprint(out, "\r\x1b[K\n")
				} else {
					_, _ = fmt.Fprintln(out)
				}
			}
			continue
		}

		if useColor {
			title = "\x1b[1m" + nPrefix + title + "\x1b[0m"
			url = "\x1b[36m" + url + "\x1b[0m"
		} else {
			title = nPrefix + title
		}
		_, _ = fmt.Fprintln(out, title)
		_, _ = fmt.Fprintln(out, "  "+url)
		_, _ = fmt.Fprintln(out)
	}
}

func renderThumbBlock(out *bufio.Writer, thumbs termcaps.InlineProtocol, id uint32, res model.Result, nPrefix, title, url string, useColor bool, termCols int) error {
	src := res.PreviewURL
	if src == "" {
		src = res.URL
	}
	data, err := fetchThumb(src)
	if err != nil {
		return err
	}

	cols := 16
	rows := 8
	if w, h := thumbDims(data, res); w > 0 && h > 0 {
		if thumbs != termcaps.InlineIterm {
			rows = clampInt(3, 10, int(float64(cols)*0.5*float64(h)/float64(w)))
		}
	}

	switch thumbs {
	case termcaps.InlineNone:
		return fmt.Errorf("inline thumbnails not supported")
	case termcaps.InlineIterm:
		if !isSupportedItermImage(data) && src != res.URL && res.URL != "" {
			if fallback, err := fetchThumb(res.URL); err == nil {
				data = fallback
			}
		}
		if !isSupportedItermImage(data) {
			return fmt.Errorf("unsupported image")
		}
		if len(data) == 0 {
			return fmt.Errorf("empty image")
		}
		_, _ = fmt.Fprint(out, "\r")
		sendThumbIterm(out, data, cols, rows)
		if rows > 1 {
			_, _ = fmt.Fprintf(out, "\x1b[%dA", rows-1)
		}
	case termcaps.InlineKitty:
		decoded, err := decodeThumb(data)
		if err != nil {
			return err
		}
		if decoded == nil || len(decoded.Frames) == 0 {
			return fmt.Errorf("no frames")
		}
		sendThumbKitty(out, id, decoded.Frames[0], cols, rows)
	}

	indentCols := cols + 2
	if thumbs == termcaps.InlineIterm {
		indentCols = cols
	}
	textWidth := termCols - indentCols - 1
	if termCols <= 0 || textWidth <= 0 {
		textWidth = 0
	}

	titleLine := nPrefix + title
	titleLine = truncateText(titleLine, textWidth)

	urlLines := []string{url}
	if textWidth > 0 {
		urlLines = wrapText(url, textWidth)
	}
	if len(urlLines) > rows-1 {
		urlLines = urlLines[:rows-1]
		urlLines[len(urlLines)-1] = truncateText(urlLines[len(urlLines)-1]+"…", textWidth)
	}

	for r := 0; r < rows; r++ {
		line := ""
		switch r {
		case 0:
			line = titleLine
			if useColor {
				line = "\x1b[1m" + line + "\x1b[0m"
			}
		default:
			if i := r - 1; i >= 0 && i < len(urlLines) {
				line = urlLines[i]
				if useColor {
					line = "\x1b[36m" + line + "\x1b[0m"
				}
			}
		}
		if thumbs == termcaps.InlineIterm {
			col := indentCols + 1
			if col < 1 {
				col = 1
			}
			_, _ = fmt.Fprintf(out, "\x1b[%dG", col)
			_, _ = fmt.Fprint(out, line)
			_, _ = fmt.Fprint(out, "\x1b[K\n")
			continue
		}

		_, _ = fmt.Fprint(out, strings.Repeat(" ", indentCols))
		_, _ = fmt.Fprintln(out, line)
	}
	return nil
}

func normalizeTitle(res model.Result) string {
	label := strings.Join(strings.Fields(res.Title), " ")
	if label == "" {
		label = strings.Join(strings.Fields(res.ID), " ")
	}
	if label == "" {
		label = "untitled"
	}
	return label
}

func clampInt(minVal, maxVal, v int) int {
	if v < minVal {
		return minVal
	}
	if v > maxVal {
		return maxVal
	}
	return v
}

func thumbInlineName(data []byte) string {
	if isGIFData(data) {
		return "thumb.gif"
	}
	if isPNGData(data) {
		return "thumb.png"
	}
	if isJPEGData(data) {
		return "thumb.jpg"
	}
	return "thumb.bin"
}

func thumbDims(data []byte, res model.Result) (int, int) {
	if isGIFData(data) {
		if len(data) < 10 {
			return 0, 0
		}
		w := int(binary.LittleEndian.Uint16(data[6:8]))
		h := int(binary.LittleEndian.Uint16(data[8:10]))
		return w, h
	}
	if isPNGData(data) {
		if len(data) < 24 {
			return 0, 0
		}
		w := int(binary.BigEndian.Uint32(data[16:20]))
		h := int(binary.BigEndian.Uint32(data[20:24]))
		return w, h
	}
	if res.Width > 0 && res.Height > 0 {
		return res.Width, res.Height
	}
	return 0, 0
}

func isGIFData(data []byte) bool {
	if len(data) < 6 {
		return false
	}
	hdr := string(data[:6])
	return hdr == "GIF87a" || hdr == "GIF89a"
}

func isPNGData(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4e && data[3] == 0x47 &&
		data[4] == 0x0d && data[5] == 0x0a && data[6] == 0x1a && data[7] == 0x0a
}

func isJPEGData(data []byte) bool {
	return len(data) >= 3 && data[0] == 0xff && data[1] == 0xd8 && data[2] == 0xff
}

func isSupportedItermImage(data []byte) bool {
	return isGIFData(data) || isPNGData(data) || isJPEGData(data)
}

func truncateText(s string, width int) string {
	if width <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= width {
		return s
	}
	if width <= 1 {
		return "…"
	}
	return string(r[:width-1]) + "…"
}

func wrapText(s string, width int) []string {
	if width <= 0 || s == "" {
		return []string{s}
	}
	var out []string
	r := []rune(s)
	for len(r) > 0 {
		n := width
		if n > len(r) {
			n = len(r)
		}
		out = append(out, string(r[:n]))
		r = r[n:]
	}
	return out
}
