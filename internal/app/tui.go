package app

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"time"

	"golang.org/x/term"
)

type inputEvent struct {
	kind keyKind
	ch   rune
}

type keyKind int

const (
	keyRune keyKind = iota
	keyEnter
	keyBackspace
	keyEsc
	keyUp
	keyDown
	keyCtrlC
	keyUnknown
)

var errNotTerminal = errors.New("stdin is not a tty")

func runTUI(opts cliOptions, query string) error {
	env := defaultEnvFn()
	return runTUIWith(env, opts, query)
}

func runTUIWith(env tuiEnv, opts cliOptions, query string) error {
	if env.in == nil {
		env.in = os.Stdin
	}
	if env.out == nil {
		env.out = os.Stdout
	}
	if env.isTerminal == nil {
		env.isTerminal = term.IsTerminal
	}
	if env.makeRaw == nil {
		env.makeRaw = term.MakeRaw
	}
	if env.restore == nil {
		env.restore = term.Restore
	}
	if env.getSize == nil {
		env.getSize = term.GetSize
	}
	if env.fd == 0 {
		env.fd = int(os.Stdin.Fd())
	}
	if !env.isTerminal(env.fd) {
		return errNotTerminal
	}

	oldState, err := env.makeRaw(env.fd)
	if err != nil {
		return err
	}
	if oldState != nil {
		defer func() {
			_ = env.restore(env.fd, oldState)
		}()
	}

	out := bufio.NewWriter(env.out)
	hideCursor(out)
	defer func() {
		showCursor(out)
		clearImages(out)
		_ = out.Flush()
	}()

	sigs := env.signalCh
	if sigs == nil {
		sigs = make(chan os.Signal)
	}

	inputCh := make(chan inputEvent, 16)
	stopCh := make(chan struct{})
	go readInput(env.in, inputCh, stopCh)

	state := &appState{
		mode:            modeQuery,
		status:          "Type a search and press Enter",
		cache:           map[string]*gifFrames{},
		renderDirty:     true,
		nextImageID:     1,
		useSoftwareAnim: useSoftwareAnimation(),
		opts:            opts,
	}
	if cols, rows, err := env.getSize(env.fd); err == nil {
		state.lastRows = rows
		state.lastCols = cols
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	if strings.TrimSpace(query) != "" {
		state.query = query
		state.mode = modeBrowse
		state.status = "Searching..."
		render(state, out, state.lastRows, state.lastCols)
		_ = out.Flush()

		results, err := search(query, opts)
		if err != nil {
			state.status = "Search error: " + err.Error()
		} else {
			results, err = filterResults(results, query, opts)
			if err != nil {
				state.status = "Filter error: " + err.Error()
			} else {
				state.results = results
				state.selected = 0
				state.scroll = 0
				if len(results) == 0 {
					state.status = "No results"
					state.currentAnim = nil
					state.previewDirty = true
				} else {
					state.status = fmt.Sprintf("%d results", len(results))
					loadSelectedImage(state)
				}
			}
		}
		state.renderDirty = true
	}

	for {
		select {
		case <-sigs:
			close(stopCh)
			return nil
		case ev := <-inputCh:
			if handleInput(state, ev, out) {
				close(stopCh)
				return nil
			}
		case <-ticker.C:
		}

		if cols, rows, err := env.getSize(env.fd); err == nil {
			if rows != state.lastRows || cols != state.lastCols {
				state.lastRows = rows
				state.lastCols = cols
				ensureVisible(state)
				state.renderDirty = true
				state.previewDirty = true
			}
		}

		if state.renderDirty {
			render(state, out, state.lastRows, state.lastCols)
			state.renderDirty = false
			_ = out.Flush()
		}

		advanceManualAnimation(state, out)
	}
}

func readInput(r io.Reader, ch chan<- inputEvent, stop <-chan struct{}) {
	reader := bufio.NewReader(r)
	for {
		select {
		case <-stop:
			return
		default:
		}

		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		switch b {
		case 0x03:
			ch <- inputEvent{kind: keyCtrlC}
		case '\r', '\n':
			ch <- inputEvent{kind: keyEnter}
		case 0x7f, 0x08:
			ch <- inputEvent{kind: keyBackspace}
		case 0x1b:
			next, err := reader.ReadByte()
			if err != nil {
				ch <- inputEvent{kind: keyEsc}
				continue
			}
			if next == '[' {
				third, _ := reader.ReadByte()
				switch third {
				case 'A':
					ch <- inputEvent{kind: keyUp}
				case 'B':
					ch <- inputEvent{kind: keyDown}
				default:
					ch <- inputEvent{kind: keyUnknown}
				}
			} else {
				_ = reader.UnreadByte()
				ch <- inputEvent{kind: keyEsc}
			}
		default:
			if b >= 0x20 && b < 0x7f {
				ch <- inputEvent{kind: keyRune, ch: rune(b)}
			}
		}
	}
}

func handleInput(state *appState, ev inputEvent, out *bufio.Writer) bool {
	if ev.kind == keyCtrlC {
		return true
	}
	if ev.kind == keyRune && ev.ch == 'q' {
		return true
	}

	switch state.mode {
	case modeQuery:
		switch ev.kind {
		case keyRune:
			state.query += string(ev.ch)
			state.renderDirty = true
		case keyBackspace:
			if len(state.query) > 0 {
				state.query = state.query[:len(state.query)-1]
				state.renderDirty = true
			}
		case keyEnter:
			if strings.TrimSpace(state.query) == "" {
				state.status = "Empty query"
				state.renderDirty = true
				return false
			}
			state.status = "Searching..."
			render(state, out, state.lastRows, state.lastCols)
			_ = out.Flush()

			results, err := search(state.query, state.opts)
			if err != nil {
				state.status = "Search error: " + err.Error()
			} else {
				results, err = filterResults(results, state.query, state.opts)
				if err != nil {
					state.status = "Filter error: " + err.Error()
				} else {
					state.results = results
					state.selected = 0
					state.scroll = 0
					if len(results) == 0 {
						state.status = "No results"
						state.currentAnim = nil
						state.previewDirty = true
					} else {
						state.status = fmt.Sprintf("%d results", len(results))
						loadSelectedImage(state)
					}
				}
			}
			state.mode = modeBrowse
			state.renderDirty = true
		case keyEsc:
			if len(state.results) > 0 {
				state.mode = modeBrowse
				state.renderDirty = true
			}
		}
	case modeBrowse:
		switch ev.kind {
		case keyRune:
			if ev.ch == '/' {
				state.mode = modeQuery
				state.status = "Type a search and press Enter"
				state.renderDirty = true
				return false
			}
		case keyUp:
			if state.selected > 0 {
				state.selected--
				ensureVisible(state)
				loadSelectedImage(state)
				state.renderDirty = true
			}
		case keyDown:
			if state.selected < len(state.results)-1 {
				state.selected++
				ensureVisible(state)
				loadSelectedImage(state)
				state.renderDirty = true
			}
		case keyEnter:
			state.mode = modeQuery
			state.status = "Type a search and press Enter"
			state.renderDirty = true
		case keyEsc:
			state.mode = modeQuery
			state.renderDirty = true
		}
	}

	return false
}

func ensureVisible(state *appState) {
	listHeight := state.lastRows - 4
	if listHeight < 0 {
		listHeight = 0
	}
	if state.selected < state.scroll {
		state.scroll = state.selected
	}
	if state.selected >= state.scroll+listHeight {
		state.scroll = state.selected - listHeight + 1
	}
	if state.scroll < 0 {
		state.scroll = 0
	}
}

func render(state *appState, out *bufio.Writer, rows, cols int) {
	if rows <= 0 || cols <= 0 {
		return
	}

	showRight := cols >= 70 && rows >= 12
	leftWidth := cols
	if showRight {
		leftWidth = maxInt(24, (cols / 3))
		if leftWidth > cols-2 {
			leftWidth = cols - 2
		}
	}

	if state.currentAnim == nil && state.activeImageID != 0 {
		deleteKittyImage(out, state.activeImageID)
		state.activeImageID = 0
	}

	row := 1
	writeLineAt(out, row, "gifgrep — GIF search (kitty protocol)", leftWidth)
	row++

	modeLabel := "browse"
	if state.mode == modeQuery {
		modeLabel = "query"
	}
	writeLineAt(out, row, fmt.Sprintf("Search [%s]: %s", modeLabel, state.query), leftWidth)
	row++
	writeLineAt(out, row, "Enter=search  / edit  ↑↓ select  q quit", leftWidth)
	row++

	availCols, availRows := availablePreviewSize(rows, cols, leftWidth, showRight)
	previewCols, previewRows := fitPreviewSize(availCols, availRows, state.currentAnim)
	if state.currentAnim == nil {
		previewCols = 0
		previewRows = 0
	}
	listHeight := rows - 4
	if !showRight && previewRows > 0 {
		listHeight = rows - 4 - previewRows - 2
	}
	if listHeight < 0 {
		listHeight = 0
	}

	listStart := row
	for i := 0; i < listHeight; i++ {
		idx := state.scroll + i
		if idx >= 0 && idx < len(state.results) {
			item := state.results[idx]
			label := item.Title
			if label == "" {
				label = item.ID
			}
			prefix := "  "
			if idx == state.selected {
				prefix = "> "
			}
			writeLineAt(out, listStart+i, prefix+label, leftWidth)
		} else {
			writeLineAt(out, listStart+i, "", leftWidth)
		}
	}
	row = listStart + listHeight

	status := state.status
	if status == "" {
		status = fmt.Sprintf("Results: %d", len(state.results))
	}
	writeLineAt(out, row, status, leftWidth)
	row++

	if state.currentAnim != nil {
		if previewCols > 0 && previewRows > 0 {
			if showRight {
				state.previewRow = 4
				state.previewCol = leftWidth + 2
				moveCursor(out, state.previewRow, state.previewCol)
				drawPreview(state, out, previewCols, previewRows, state.previewRow, state.previewCol)
			} else {
				writeLineAt(out, row, "Preview:", leftWidth)
				row++
				state.previewRow = row
				state.previewCol = 1
				for i := 0; i < previewRows; i++ {
					writeLineAt(out, row+i, "", cols)
				}
				moveCursor(out, state.previewRow, state.previewCol)
				drawPreview(state, out, previewCols, previewRows, state.previewRow, state.previewCol)
				row += previewRows
			}
		}
	}

	for row <= rows {
		writeLineAt(out, row, "", cols)
		row++
	}
}

func availablePreviewSize(rows, cols, leftWidth int, showRight bool) (int, int) {
	if rows <= 0 || cols <= 0 {
		return 0, 0
	}
	if showRight {
		availCols := cols - leftWidth - 2
		availRows := rows - 4
		if availCols < 10 || availRows < 6 {
			return 0, 0
		}
		return availCols, availRows
	}
	availCols := cols
	availRows := rows / 3
	if availRows < 6 {
		availRows = 6
	}
	maxRows := rows - 6
	if availRows > maxRows {
		availRows = maxRows
	}
	if availCols < 10 || availRows <= 0 {
		return 0, 0
	}
	return availCols, availRows
}

func fitPreviewSize(availCols, availRows int, anim *gifAnimation) (int, int) {
	if availCols <= 0 || availRows <= 0 {
		return 0, 0
	}
	if anim == nil || anim.Width <= 0 || anim.Height <= 0 {
		return availCols, availRows
	}
	aspect := cellAspectRatio()
	targetCols := availCols
	targetRows := int(math.Round(float64(targetCols) * aspect * float64(anim.Height) / float64(anim.Width)))
	if targetRows > availRows {
		targetRows = availRows
		targetCols = int(math.Round(float64(targetRows) / aspect * float64(anim.Width) / float64(anim.Height)))
	}
	if targetCols < 1 {
		targetCols = 1
	}
	if targetRows < 1 {
		targetRows = 1
	}
	return minInt(targetCols, availCols), minInt(targetRows, availRows)
}

func drawPreview(state *appState, out *bufio.Writer, cols, rows int, row, col int) {
	if state.currentAnim == nil || len(state.currentAnim.Frames) == 0 {
		return
	}
	if state.useSoftwareAnim && len(state.currentAnim.Frames) > 1 {
		drawPreviewSoftware(state, out, cols, rows, row, col)
		return
	}
	if state.previewNeedsSend {
		if state.activeImageID != 0 {
			deleteKittyImage(out, state.activeImageID)
		}
		state.activeImageID = state.currentAnim.ID
		sendKittyAnimation(out, state.currentAnim, cols, rows)
		state.previewNeedsSend = false
		state.previewDirty = false
		state.lastPreview.cols = cols
		state.lastPreview.rows = rows
		return
	}
	if state.previewDirty || state.lastPreview.cols != cols || state.lastPreview.rows != rows {
		placeKittyImage(out, state.activeImageID, cols, rows)
		state.previewDirty = false
		state.lastPreview.cols = cols
		state.lastPreview.rows = rows
	}
}

func writeLineAt(out *bufio.Writer, row int, text string, width int) {
	moveCursor(out, row, 1)
	if width <= 0 {
		_, _ = fmt.Fprint(out, "\x1b[K")
		return
	}
	text = truncateRunes(text, width)
	_, _ = fmt.Fprint(out, text)
	_, _ = fmt.Fprint(out, "\x1b[K")
}

func drawPreviewSoftware(state *appState, out *bufio.Writer, cols, rows int, row, col int) {
	if state.currentAnim == nil || len(state.currentAnim.Frames) == 0 {
		return
	}
	if state.activeImageID != 0 && state.activeImageID != state.currentAnim.ID {
		deleteKittyImage(out, state.activeImageID)
	}
	state.activeImageID = state.currentAnim.ID
	if state.previewNeedsSend {
		state.manualAnim = true
		state.manualFrame = 0
		frame := state.currentAnim.Frames[state.manualFrame]
		saveCursor(out)
		moveCursor(out, row, col)
		sendKittyFrame(out, state.activeImageID, frame, cols, rows)
		restoreCursor(out)
		state.manualNext = time.Now().Add(time.Duration(frame.DelayMS) * time.Millisecond)
		state.previewNeedsSend = false
		state.previewDirty = false
		state.lastPreview.cols = cols
		state.lastPreview.rows = rows
		return
	}
	if state.previewDirty || state.lastPreview.cols != cols || state.lastPreview.rows != rows {
		frame := state.currentAnim.Frames[state.manualFrame]
		saveCursor(out)
		moveCursor(out, row, col)
		sendKittyFrame(out, state.activeImageID, frame, cols, rows)
		restoreCursor(out)
		state.previewDirty = false
		state.lastPreview.cols = cols
		state.lastPreview.rows = rows
	}
}

func advanceManualAnimation(state *appState, out *bufio.Writer) {
	if !state.manualAnim || state.currentAnim == nil {
		return
	}
	if len(state.currentAnim.Frames) <= 1 {
		return
	}
	if state.lastPreview.cols == 0 || state.lastPreview.rows == 0 {
		return
	}
	if state.manualNext.IsZero() || state.previewRow == 0 || state.previewCol == 0 {
		return
	}
	now := time.Now()
	if now.Before(state.manualNext) {
		return
	}
	state.manualFrame = (state.manualFrame + 1) % len(state.currentAnim.Frames)
	frame := state.currentAnim.Frames[state.manualFrame]
	saveCursor(out)
	moveCursor(out, state.previewRow, state.previewCol)
	sendKittyFrame(out, state.activeImageID, frame, state.lastPreview.cols, state.lastPreview.rows)
	restoreCursor(out)
	state.manualNext = now.Add(time.Duration(frame.DelayMS) * time.Millisecond)
	_ = out.Flush()
}

func writeLine(out *bufio.Writer, text string, width int) {
	if width <= 0 {
		_, _ = fmt.Fprint(out, "\r\n")
		return
	}
	text = truncateRunes(text, width)
	_, _ = fmt.Fprint(out, text)
	_, _ = fmt.Fprint(out, "\x1b[K\r\n")
}

func moveCursor(out *bufio.Writer, row, col int) {
	if row < 1 {
		row = 1
	}
	if col < 1 {
		col = 1
	}
	_, _ = fmt.Fprintf(out, "\x1b[%d;%dH", row, col)
}

func saveCursor(out *bufio.Writer) {
	_, _ = fmt.Fprint(out, "\x1b7")
}

func restoreCursor(out *bufio.Writer) {
	_, _ = fmt.Fprint(out, "\x1b8")
}

func hideCursor(out *bufio.Writer) {
	_, _ = fmt.Fprint(out, "\x1b[?25l")
}

func showCursor(out *bufio.Writer) {
	_, _ = fmt.Fprint(out, "\x1b[?25h")
}

func clearImages(out *bufio.Writer) {
	_, _ = fmt.Fprint(out, "\x1b_Ga=d\x1b\\")
}
