package app

import "time"

func loadSelectedImage(state *appState) {
	if state.selected < 0 || state.selected >= len(state.results) {
		state.currentAnim = nil
		state.previewDirty = true
		return
	}
	item := state.results[state.selected]
	if item.PreviewURL == "" {
		state.currentAnim = nil
		state.previewDirty = true
		return
	}
	frames, ok := state.cache[item.PreviewURL]
	if !ok {
		data, err := fetchGIF(item.PreviewURL)
		if err != nil {
			state.status = "Image error: " + err.Error()
			state.currentAnim = nil
			return
		}
		decoded, err := decodeGIFFrames(data)
		if err != nil {
			state.status = "Image error: " + err.Error()
			state.currentAnim = nil
			return
		}
		state.cache[item.PreviewURL] = decoded
		frames = decoded
	}
	state.currentAnim = &gifAnimation{
		ID:     state.nextImageID,
		Frames: frames.Frames,
		Width:  frames.Width,
		Height: frames.Height,
	}
	state.nextImageID++
	state.manualAnim = false
	state.manualFrame = 0
	state.manualNext = time.Time{}
	state.previewNeedsSend = true
	state.previewDirty = true
}
