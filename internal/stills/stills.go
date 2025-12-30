package stills

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"time"

	"github.com/steipete/gifgrep/gifdecode"
)

var (
	ErrNoFrames     = errors.New("no frames")
	ErrInvalidCount = errors.New("invalid count")
	ErrInvalidSheet = errors.New("invalid sheet size")
)

type SheetOptions struct {
	Count      int
	Columns    int
	Padding    int
	Background color.Color
}

func FrameIndexAt(frames []gifdecode.Frame, at time.Duration) (int, error) {
	if len(frames) == 0 {
		return -1, ErrNoFrames
	}
	if at < 0 {
		at = 0
	}
	total := totalDuration(frames)
	if total <= 0 {
		return 0, nil
	}
	elapsed := time.Duration(0)
	for i, frame := range frames {
		elapsed += frame.Delay
		if at < elapsed {
			return i, nil
		}
	}
	return len(frames) - 1, nil
}

func FrameAtPNG(decoded *gifdecode.Frames, at time.Duration) ([]byte, int, error) {
	if decoded == nil {
		return nil, -1, ErrNoFrames
	}
	idx, err := FrameIndexAt(decoded.Frames, at)
	if err != nil {
		return nil, -1, err
	}
	return decoded.Frames[idx].PNG, idx, nil
}

func ContactSheet(decoded *gifdecode.Frames, opts SheetOptions) ([]byte, error) {
	if decoded == nil || len(decoded.Frames) == 0 {
		return nil, ErrNoFrames
	}
	if opts.Count <= 0 {
		return nil, ErrInvalidCount
	}

	if opts.Count > len(decoded.Frames) {
		opts.Count = len(decoded.Frames)
	}
	if opts.Columns <= 0 {
		opts.Columns = int(math.Ceil(math.Sqrt(float64(opts.Count))))
	}
	if opts.Padding < 0 {
		opts.Padding = 0
	}
	if opts.Background == nil {
		opts.Background = color.Transparent
	}

	frameWidth := decoded.Width
	frameHeight := decoded.Height
	if frameWidth <= 0 || frameHeight <= 0 {
		img, err := decodePNG(decoded.Frames[0].PNG)
		if err != nil {
			return nil, err
		}
		b := img.Bounds()
		frameWidth, frameHeight = b.Dx(), b.Dy()
	}
	if frameWidth <= 0 || frameHeight <= 0 {
		return nil, ErrInvalidSheet
	}

	rows := int(math.Ceil(float64(opts.Count) / float64(opts.Columns)))

	sheetWidth := frameWidth*opts.Columns + opts.Padding*(opts.Columns-1)
	sheetHeight := frameHeight*rows + opts.Padding*(rows-1)

	sheet := image.NewRGBA(image.Rect(0, 0, sheetWidth, sheetHeight))
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{C: opts.Background}, image.Point{}, draw.Src)

	indices := sampleIndices(decoded.Frames, opts.Count)
	for i, idx := range indices {
		if idx < 0 || idx >= len(decoded.Frames) {
			continue
		}
		img, err := decodePNG(decoded.Frames[idx].PNG)
		if err != nil {
			return nil, err
		}
		col := i % opts.Columns
		row := i / opts.Columns
		x := col * (frameWidth + opts.Padding)
		y := row * (frameHeight + opts.Padding)
		rect := image.Rect(x, y, x+frameWidth, y+frameHeight)
		draw.Draw(sheet, rect, img, img.Bounds().Min, draw.Over)
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, sheet); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func totalDuration(frames []gifdecode.Frame) time.Duration {
	var total time.Duration
	for _, frame := range frames {
		total += frame.Delay
	}
	return total
}

func sampleIndices(frames []gifdecode.Frame, count int) []int {
	if count <= 0 {
		return nil
	}
	if len(frames) == 0 {
		return []int{}
	}
	if count == 1 {
		return []int{0}
	}
	total := totalDuration(frames)
	if total <= 0 {
		indices := make([]int, 0, count)
		max := len(frames) - 1
		for i := 0; i < count; i++ {
			idx := int(math.Round(float64(max) * float64(i) / float64(count-1)))
			indices = append(indices, idx)
		}
		return indices
	}
	indices := make([]int, 0, count)
	for i := 0; i < count; i++ {
		at := time.Duration(float64(total) * float64(i) / float64(count-1))
		idx, _ := FrameIndexAt(frames, at)
		indices = append(indices, idx)
	}
	return indices
}

func decodePNG(data []byte) (image.Image, error) {
	return png.Decode(bytes.NewReader(data))
}
