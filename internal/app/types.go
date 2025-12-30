package app

import "time"

type mode int

const (
	modeBrowse mode = iota
	modeQuery
)

const appName = "gifgrep"

var version = "dev"

type gifResult struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	PreviewURL string   `json:"preview_url"`
	Tags       []string `json:"tags,omitempty"`
	Width      int      `json:"width,omitempty"`
	Height     int      `json:"height,omitempty"`
}

type gifFrame struct {
	PNG     []byte
	DelayMS int
}

type gifFrames struct {
	Frames []gifFrame
	Width  int
	Height int
}

type gifAnimation struct {
	ID     uint32
	Frames []gifFrame
	Width  int
	Height int
}

type cliOptions struct {
	TUI        bool
	JSON       bool
	IgnoreCase bool
	Invert     bool
	Regex      bool
	Number     bool
	Limit      int
	Source     string
	Mood       string
	Color      string
}

type appState struct {
	query       string
	results     []gifResult
	selected    int
	scroll      int
	mode        mode
	status      string
	currentAnim *gifAnimation
	cache       map[string]*gifFrames
	renderDirty bool
	lastRows    int
	lastCols    int
	previewRow  int
	previewCol  int
	lastPreview struct {
		cols int
		rows int
	}
	previewNeedsSend bool
	previewDirty     bool
	nextImageID      uint32
	activeImageID    uint32
	manualAnim       bool
	manualFrame      int
	manualNext       time.Time
	useSoftwareAnim  bool
	opts             cliOptions
}
