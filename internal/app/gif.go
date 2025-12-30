package app

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"
)

func fetchGIF(gifURL string) ([]byte, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", gifURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gifgrep")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return raw, nil
}

func decodeGIFFrames(data []byte) (*gifFrames, error) {
	g, err := gif.DecodeAll(bytes.NewReader(data))
	if err != nil {
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		pngData, err := encodePNG(img)
		if err != nil {
			return nil, err
		}
		return &gifFrames{
			Frames: []gifFrame{{PNG: pngData, DelayMS: 0}},
			Width:  img.Bounds().Dx(),
			Height: img.Bounds().Dy(),
		}, nil
	}

	bounds := image.Rect(0, 0, g.Config.Width, g.Config.Height)
	canvas := image.NewRGBA(bounds)
	prev := image.NewRGBA(bounds)
	frames := make([]gifFrame, 0, len(g.Image))

	maxFrames := 60
	for i, frame := range g.Image {
		if i >= maxFrames {
			break
		}
		disposal := gif.DisposalNone
		if i < len(g.Disposal) {
			disposal = int(g.Disposal[i])
		}
		if disposal == gif.DisposalPrevious {
			copy(prev.Pix, canvas.Pix)
		}

		draw.Draw(canvas, frame.Bounds(), frame, image.Point{}, draw.Over)
		pngData, err := encodePNG(canvas)
		if err != nil {
			return nil, err
		}
		delay := 0
		if i < len(g.Delay) {
			delay = g.Delay[i] * 10
		}
		if delay <= 0 {
			delay = 80
		}
		frames = append(frames, gifFrame{PNG: pngData, DelayMS: delay})

		switch disposal {
		case gif.DisposalBackground:
			draw.Draw(canvas, frame.Bounds(), &image.Uniform{C: color.Transparent}, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			copy(canvas.Pix, prev.Pix)
		}
	}

	return &gifFrames{
		Frames: frames,
		Width:  g.Config.Width,
		Height: g.Config.Height,
	}, nil
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
