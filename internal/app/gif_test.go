package app

import (
	"errors"
	"image"
	"net/http"
	"testing"
)

func TestDecodeGIFFrames(t *testing.T) {
	data := makeTestGIF()
	frames, err := decodeGIFFrames(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames.Frames))
	}
	if frames.Width != 2 || frames.Height != 2 {
		t.Fatalf("unexpected size")
	}
	if frames.Frames[0].DelayMS != 50 || frames.Frames[1].DelayMS != 70 {
		t.Fatalf("unexpected delays: %+v", frames.Frames)
	}

	img := image.NewRGBA(image.Rect(0, 0, 3, 4))
	pngData, err := encodePNG(img)
	if err != nil {
		t.Fatalf("encodePNG failed: %v", err)
	}
	frames, err = decodeGIFFrames(pngData)
	if err != nil {
		t.Fatalf("decodePNG failed: %v", err)
	}
	if len(frames.Frames) != 1 || frames.Width != 3 || frames.Height != 4 {
		t.Fatalf("unexpected png decode result")
	}

	if _, err := decodeGIFFrames([]byte("nope")); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestLoadSelectedImageEdges(t *testing.T) {
	state := &appState{
		results: []gifResult{},
		cache:   map[string]*gifFrames{},
	}
	loadSelectedImage(state)
	if state.currentAnim != nil {
		t.Fatalf("expected nil animation for empty results")
	}

	state.results = []gifResult{{Title: "no preview"}}
	state.selected = 0
	loadSelectedImage(state)
	if state.currentAnim != nil {
		t.Fatalf("expected nil animation for empty preview url")
	}

	state.cache["https://example.test/preview.gif"] = &gifFrames{
		Frames: []gifFrame{{PNG: []byte{1, 2, 3}, DelayMS: 80}},
		Width:  1,
		Height: 1,
	}
	state.results = []gifResult{{Title: "cached", PreviewURL: "https://example.test/preview.gif"}}
	loadSelectedImage(state)
	if state.currentAnim == nil || !state.previewNeedsSend {
		t.Fatalf("expected cached animation")
	}

	badTransport := &fakeTransport{gifData: []byte("not-a-gif")}
	withTransport(t, badTransport, func() {
		state.cache = map[string]*gifFrames{}
		state.results = []gifResult{{Title: "bad", PreviewURL: "https://example.test/preview.gif"}}
		state.selected = 0
		loadSelectedImage(state)
		if state.currentAnim != nil {
			t.Fatalf("expected nil animation on decode error")
		}
	})
}

type errTransport struct{}

func (t *errTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("network")
}

func TestFetchGIFError(t *testing.T) {
	withTransport(t, &errTransport{}, func() {
		if _, err := fetchGIF("https://example.test/preview.gif"); err == nil {
			t.Fatalf("expected fetch error")
		}
	})
}
