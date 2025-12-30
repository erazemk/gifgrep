package gifdecode

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"testing"
	"time"
)

func TestDecodeGIFFrames(t *testing.T) {
	data := makeTestGIF(2)
	frames, err := Decode(data, DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames.Frames))
	}
	if frames.Width != 2 || frames.Height != 2 {
		t.Fatalf("unexpected size")
	}
	if frames.Frames[0].Delay != 50*time.Millisecond || frames.Frames[1].Delay != 70*time.Millisecond {
		t.Fatalf("unexpected delays: %+v", frames.Frames)
	}
}

func TestDecodeFallbackAndStrict(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 3, 4))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}
	frames, err := Decode(buf.Bytes(), DefaultOptions())
	if err != nil {
		t.Fatalf("decode png failed: %v", err)
	}
	if len(frames.Frames) != 1 || frames.Width != 3 || frames.Height != 4 {
		t.Fatalf("unexpected png decode result")
	}

	_, err = Decode(buf.Bytes(), Options{StrictGIF: true})
	if err == nil {
		t.Fatalf("expected strict gif error")
	}
}

func TestDecodeLimits(t *testing.T) {
	data := makeTestGIF(3)

	opts := DefaultOptions()
	opts.MaxFrames = 2
	frames, err := Decode(data, opts)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(frames.Frames))
	}

	opts = DefaultOptions()
	opts.MaxPixels = 1
	_, err = Decode(data, opts)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}

	opts = DefaultOptions()
	opts.MaxBytes = 1
	_, err = Decode(data, opts)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestBackgroundDisposalUsesColor(t *testing.T) {
	data := makeBackgroundGIF()
	frames, err := Decode(data, DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) < 2 {
		t.Fatalf("expected 2 frames")
	}
	img, err := png.Decode(bytes.NewReader(frames.Frames[1].PNG))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	r, g, b, _ := img.At(0, 0).RGBA()
	if r != 0 || g != 0 || b != 0 {
		t.Fatalf("expected background black, got %d %d %d", r, g, b)
	}
	r, g, b, _ = img.At(1, 1).RGBA()
	if r == 0 && g == 0 && b == 0 {
		t.Fatalf("expected second frame to draw at offset")
	}
}

func TestDecodeFrameOffsetComposite(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	frame1 := image.NewPaletted(image.Rect(0, 0, 4, 4), pal)
	frame2 := image.NewPaletted(image.Rect(1, 1, 3, 3), pal)
	frame2.SetColorIndex(1, 1, 1)
	frame2.SetColorIndex(2, 2, 1)

	anim := &gif.GIF{
		Image:    []*image.Paletted{frame1, frame2},
		Delay:    []int{5, 5},
		Disposal: []byte{gif.DisposalNone, gif.DisposalNone},
		Config: image.Config{
			Width:      4,
			Height:     4,
			ColorModel: pal,
		},
		BackgroundIndex: 0,
	}

	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, anim)

	decoded, err := Decode(buf.Bytes(), DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(decoded.Frames) != 2 {
		t.Fatalf("expected 2 frames, got %d", len(decoded.Frames))
	}

	img, err := png.Decode(bytes.NewReader(decoded.Frames[1].PNG))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}

	rr, gg, bb, _ := img.At(0, 0).RGBA()
	if rr != 0 || gg != 0 || bb != 0 {
		t.Fatalf("expected background to stay black at (0,0), got %d %d %d", rr, gg, bb)
	}

	rr, gg, bb, _ = img.At(1, 1).RGBA()
	if rr != 0xffff || gg != 0xffff || bb != 0xffff {
		t.Fatalf("expected offset pixel to be white at (1,1), got %d %d %d", rr, gg, bb)
	}

	rr, gg, bb, _ = img.At(3, 3).RGBA()
	if rr != 0 || gg != 0 || bb != 0 {
		t.Fatalf("expected background to stay black at (3,3), got %d %d %d", rr, gg, bb)
	}
}

func TestBackgroundIndexOutOfRangeUsesTransparent(t *testing.T) {
	data := makeOutOfRangeBackgroundGIF()
	frames, err := Decode(data, DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(frames.Frames) < 2 {
		t.Fatalf("expected 2 frames")
	}
	img, err := png.Decode(bytes.NewReader(frames.Frames[1].PNG))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	r, g, b, a := img.At(0, 0).RGBA()
	_ = r
	_ = g
	_ = b
	if a != 0 {
		t.Fatalf("expected transparent background, got alpha %d", a)
	}
	r, g, b, a = img.At(1, 1).RGBA()
	_ = r
	_ = g
	_ = b
	if a == 0 {
		t.Fatalf("expected second frame to draw at offset")
	}
}

func TestDecodeDelayClamp(t *testing.T) {
	data := makeDelayGIF([]int{0, 1, 300})
	opts := DefaultOptions()
	opts.DefaultDelay = 100 * time.Millisecond
	opts.MinDelay = 50 * time.Millisecond
	opts.MaxDelay = 200 * time.Millisecond
	frames, err := Decode(data, opts)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if frames.Frames[0].Delay != 100*time.Millisecond {
		t.Fatalf("expected default delay, got %v", frames.Frames[0].Delay)
	}
	if frames.Frames[1].Delay != 50*time.Millisecond {
		t.Fatalf("expected min delay, got %v", frames.Frames[1].Delay)
	}
	if frames.Frames[2].Delay != 200*time.Millisecond {
		t.Fatalf("expected max delay, got %v", frames.Frames[2].Delay)
	}
}

func TestDecodeInvalidData(t *testing.T) {
	if _, err := Decode([]byte("nope"), DefaultOptions()); err == nil {
		t.Fatalf("expected decode error")
	}
}

func TestDecodeUnlimitedBytesAndPixels(t *testing.T) {
	data := makeTestGIF(1)
	opts := DefaultOptions()
	opts.MaxBytes = -1
	opts.MaxPixels = -1
	if _, err := Decode(data, opts); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
}

func TestDecodeGIFErrors(t *testing.T) {
	opts := DefaultOptions()
	if _, err := decodeGIF(&gif.GIF{}, opts); !errors.Is(err, ErrNoFrames) {
		t.Fatalf("expected ErrNoFrames, got %v", err)
	}

	pal := color.Palette{color.Black}
	zero := image.NewPaletted(image.Rect(0, 0, 0, 0), pal)
	g := &gif.GIF{Image: []*image.Paletted{zero}}
	if _, err := decodeGIF(g, opts); !errors.Is(err, ErrInvalidSize) {
		t.Fatalf("expected ErrInvalidSize, got %v", err)
	}
}

func TestDecodeGIFDelayMissingEntries(t *testing.T) {
	pal := color.Palette{color.Black, color.White}
	frame1 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	frame2 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	g := &gif.GIF{
		Image: []*image.Paletted{frame1, frame2},
		Delay: []int{5},
		Config: image.Config{
			Width:      2,
			Height:     2,
			ColorModel: pal,
		},
	}
	opts := DefaultOptions()
	opts.DefaultDelay = 90 * time.Millisecond
	frames, err := decodeGIF(g, opts)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if frames.Frames[1].Delay != 90*time.Millisecond {
		t.Fatalf("expected default delay, got %v", frames.Frames[1].Delay)
	}
}

func TestSingleFrameTooLarge(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode failed: %v", err)
	}
	opts := DefaultOptions()
	opts.MaxPixels = 1
	if _, err := Decode(buf.Bytes(), opts); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestBackgroundColorNoPalette(t *testing.T) {
	g := &gif.GIF{Config: image.Config{ColorModel: color.GrayModel}}
	if bg := backgroundColor(g); bg != color.Transparent {
		t.Fatalf("expected transparent background")
	}
}

func TestReadAllLimitError(t *testing.T) {
	_, err := readAllLimit(errReader{}, 10)
	if err == nil {
		t.Fatalf("expected read error")
	}
}

func TestWithDefaultsMaxDelayClamp(t *testing.T) {
	opts := Options{MinDelay: 200 * time.Millisecond, MaxDelay: 100 * time.Millisecond}
	opts = opts.withDefaults()
	if opts.MaxDelay != opts.MinDelay {
		t.Fatalf("expected MaxDelay to be clamped to MinDelay")
	}
}

func makeTestGIF(count int) []byte {
	pal := color.Palette{color.Black, color.White}
	frames := make([]*image.Paletted, 0, count)
	delays := make([]int, 0, count)
	disposal := make([]byte, 0, count)
	for i := 0; i < count; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
		frame.SetColorIndex(i%2, i%2, 1)
		frames = append(frames, frame)
		delays = append(delays, 5+i*2)
		disposal = append(disposal, gif.DisposalNone)
	}
	g := &gif.GIF{
		Image:    frames,
		Delay:    delays,
		Disposal: disposal,
		Config: image.Config{
			Width:      2,
			Height:     2,
			ColorModel: pal,
		},
		BackgroundIndex: 0,
	}
	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

func makeBackgroundGIF() []byte {
	pal := color.Palette{color.Black, color.White}
	frame1 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	frame1.SetColorIndex(0, 0, 1)
	frame2 := image.NewPaletted(image.Rect(1, 1, 2, 2), pal)
	frame2.SetColorIndex(1, 1, 1)

	g := &gif.GIF{
		Image:    []*image.Paletted{frame1, frame2},
		Delay:    []int{5, 5},
		Disposal: []byte{gif.DisposalBackground, gif.DisposalNone},
		Config: image.Config{
			Width:      2,
			Height:     2,
			ColorModel: pal,
		},
		BackgroundIndex: 0,
	}
	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

func makeOutOfRangeBackgroundGIF() []byte {
	pal := color.Palette{color.Black, color.White}
	frame1 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	frame1.SetColorIndex(0, 0, 1)
	frame2 := image.NewPaletted(image.Rect(1, 1, 2, 2), pal)
	frame2.SetColorIndex(1, 1, 1)

	g := &gif.GIF{
		Image:    []*image.Paletted{frame1, frame2},
		Delay:    []int{5, 5},
		Disposal: []byte{gif.DisposalBackground, gif.DisposalNone},
		Config: image.Config{
			Width:      2,
			Height:     2,
			ColorModel: pal,
		},
		BackgroundIndex: 9,
	}
	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

func makeDelayGIF(delays []int) []byte {
	pal := color.Palette{color.Black, color.White}
	frames := make([]*image.Paletted, 0, len(delays))
	for i := range delays {
		frame := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
		frame.SetColorIndex(i%2, i%2, 1)
		frames = append(frames, frame)
	}
	g := &gif.GIF{
		Image: frames,
		Delay: delays,
		Config: image.Config{
			Width:      2,
			Height:     2,
			ColorModel: pal,
		},
		BackgroundIndex: 0,
	}
	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, g)
	return buf.Bytes()
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) {
	return 0, errors.New("boom")
}

func makeMediumGIF() []byte {
	pal := color.Palette{color.Black, color.White}
	frames := make([]*image.Paletted, 0, 8)
	delays := make([]int, 0, 8)
	for i := 0; i < 8; i++ {
		frame := image.NewPaletted(image.Rect(0, 0, 80, 60), pal)
		frame.SetColorIndex(i%80, i%60, 1)
		frames = append(frames, frame)
		delays = append(delays, 4+i)
	}
	g := &gif.GIF{
		Image: frames,
		Delay: delays,
		Config: image.Config{
			Width:      80,
			Height:     60,
			ColorModel: pal,
		},
		BackgroundIndex: 0,
	}
	var buf bytes.Buffer
	_ = gif.EncodeAll(&buf, g)
	return buf.Bytes()
}
