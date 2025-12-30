package stills

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"testing"
	"time"

	"github.com/steipete/gifgrep/gifdecode"
	"github.com/steipete/gifgrep/internal/testutil"
)

func TestFrameAtPNG(t *testing.T) {
	data := testutil.MakeTestGIF()
	decoded, err := gifdecode.Decode(data, gifdecode.DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	pngData, idx, err := FrameAtPNG(decoded, 0)
	if err != nil {
		t.Fatalf("FrameAtPNG failed: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected idx 0, got %d", idx)
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if !isWhite(img.At(0, 0)) {
		t.Fatalf("expected white pixel at (0,0)")
	}

	pngData, idx, err = FrameAtPNG(decoded, 60*time.Millisecond)
	if err != nil {
		t.Fatalf("FrameAtPNG failed: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected idx 1, got %d", idx)
	}
	img, err = png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if !isWhite(img.At(1, 1)) {
		t.Fatalf("expected white pixel at (1,1)")
	}
}

func TestContactSheetDimensions(t *testing.T) {
	data := testutil.MakeTestGIF()
	decoded, err := gifdecode.Decode(data, gifdecode.DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	pngData, err := ContactSheet(decoded, SheetOptions{Count: 2, Columns: 2, Padding: 1, Background: color.Transparent})
	if err != nil {
		t.Fatalf("ContactSheet failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 5 {
		t.Fatalf("expected width 5, got %d", img.Bounds().Dx())
	}
	if img.Bounds().Dy() != 2 {
		t.Fatalf("expected height 2, got %d", img.Bounds().Dy())
	}
}

func TestFrameIndexAtBounds(t *testing.T) {
	frames := []gifdecode.Frame{{Delay: 10 * time.Millisecond}, {Delay: 20 * time.Millisecond}}
	idx, err := FrameIndexAt(frames, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("FrameIndexAt failed: %v", err)
	}
	if idx != 1 {
		t.Fatalf("expected last frame, got %d", idx)
	}
}

func TestFrameIndexAtErrors(t *testing.T) {
	if _, err := FrameIndexAt(nil, 0); err == nil {
		t.Fatalf("expected error on empty frames")
	}
	frames := []gifdecode.Frame{{Delay: 10 * time.Millisecond}}
	idx, err := FrameIndexAt(frames, -time.Second)
	if err != nil {
		t.Fatalf("FrameIndexAt failed: %v", err)
	}
	if idx != 0 {
		t.Fatalf("expected first frame, got %d", idx)
	}
}

func TestContactSheetDefaults(t *testing.T) {
	data := testutil.MakeTestGIF()
	decoded, err := gifdecode.Decode(data, gifdecode.DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	pngData, err := ContactSheet(decoded, SheetOptions{Count: 2})
	if err != nil {
		t.Fatalf("ContactSheet failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		t.Fatalf("expected non-zero sheet size")
	}
}

func TestContactSheetErrors(t *testing.T) {
	data := testutil.MakeTestGIF()
	decoded, err := gifdecode.Decode(data, gifdecode.DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if _, err := ContactSheet(decoded, SheetOptions{Count: 0}); err == nil {
		t.Fatalf("expected invalid count error")
	}
	if _, err := ContactSheet(nil, SheetOptions{Count: 1}); err == nil {
		t.Fatalf("expected no frames error")
	}
}

func TestFrameAtPNGNil(t *testing.T) {
	if _, _, err := FrameAtPNG(nil, 0); err == nil {
		t.Fatalf("expected error for nil frames")
	}
	empty := &gifdecode.Frames{}
	if _, _, err := FrameAtPNG(empty, 0); err == nil {
		t.Fatalf("expected error for empty frames")
	}
}

func TestContactSheetZeroDelay(t *testing.T) {
	pngData := makeSolidPNG(color.White, 1, 1)
	decoded := &gifdecode.Frames{
		Frames: []gifdecode.Frame{
			{PNG: pngData, Delay: 0},
			{PNG: pngData, Delay: 0},
		},
		Width:  1,
		Height: 1,
	}
	out, err := ContactSheet(decoded, SheetOptions{Count: 2, Columns: 2, Padding: -1})
	if err != nil {
		t.Fatalf("ContactSheet failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 1 {
		t.Fatalf("unexpected sheet size %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestContactSheetBadPNG(t *testing.T) {
	decoded := &gifdecode.Frames{
		Frames: []gifdecode.Frame{{PNG: []byte("not-png"), Delay: 10 * time.Millisecond}},
		Width:  0,
		Height: 0,
	}
	if _, err := ContactSheet(decoded, SheetOptions{Count: 1}); err == nil {
		t.Fatalf("expected error for invalid png")
	}
}

func TestContactSheetClampAndFallbackSize(t *testing.T) {
	pngData := makeSolidPNG(color.Black, 1, 1)
	decoded := &gifdecode.Frames{
		Frames: []gifdecode.Frame{
			{PNG: pngData, Delay: 10 * time.Millisecond},
			{PNG: pngData, Delay: 10 * time.Millisecond},
		},
		Width:  0,
		Height: 0,
	}
	out, err := ContactSheet(decoded, SheetOptions{Count: 5, Columns: 2, Padding: 1})
	if err != nil {
		t.Fatalf("ContactSheet failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 3 || img.Bounds().Dy() != 1 {
		t.Fatalf("unexpected sheet size %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestContactSheetSingleFrame(t *testing.T) {
	data := testutil.MakeTestGIF()
	decoded, err := gifdecode.Decode(data, gifdecode.DefaultOptions())
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	out, err := ContactSheet(decoded, SheetOptions{Count: 1, Columns: 1, Padding: 0})
	if err != nil {
		t.Fatalf("ContactSheet failed: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 2 || img.Bounds().Dy() != 2 {
		t.Fatalf("unexpected sheet size %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func makeSolidPNG(c color.Color, w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: c}, image.Point{}, draw.Src)
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

func isWhite(c color.Color) bool {
	r, g, b, a := c.RGBA()
	return a > 0 && r == 0xffff && g == 0xffff && b == 0xffff
}
