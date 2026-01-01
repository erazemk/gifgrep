package tui

import (
	"time"

	"github.com/steipete/gifgrep/gifdecode"
	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/termcaps"
)

type mode int

const (
	modeBrowse mode = iota
	modeQuery
)

type gifAnimation struct {
	ID     uint32
	RawGIF []byte
	Frames []gifdecode.Frame
	Width  int
	Height int
}

type gifCacheEntry struct {
	RawGIF []byte
	Frames *gifdecode.Frames
	Width  int
	Height int
}

type appState struct {
	query         string
	tagline       string
	headerFlash   string
	headerFlashAt time.Time
	results       []model.Result
	selected      int
	scroll        int
	mode          mode
	status        string
	currentAnim   *gifAnimation
	inline        termcaps.InlineProtocol
	cache         map[string]*gifCacheEntry
	savedPaths    map[string]string
	renderDirty   bool
	lastShowRight bool
	lastRows      int
	lastCols      int
	previewRow    int
	previewCol    int
	lastPreview   struct {
		cols int
		rows int
	}
	itermLast struct {
		row  int
		col  int
		cols int
		rows int
	}
	previewNeedsSend      bool
	previewDirty          bool
	nextImageID           uint32
	activeImageID         uint32
	manualAnim            bool
	manualFrame           int
	manualNext            time.Time
	useSoftwareAnim       bool
	useColor              bool
	opts                  model.Options
	giphyAttributionShown bool
	lastSavedPath         string
}
