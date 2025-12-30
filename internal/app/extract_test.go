package app

import (
	"bytes"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/steipete/gifgrep/internal/model"
	"github.com/steipete/gifgrep/internal/testutil"
)

func TestReadInputFileAndURL(t *testing.T) {
	data := testutil.MakeTestGIF()
	tmp := filepath.Join(t.TempDir(), "sample.gif")
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	got, err := readInput(tmp)
	if err != nil {
		t.Fatalf("readInput file: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("file data mismatch")
	}

	fileURL := "file://" + tmp
	got, err = readInput(fileURL)
	if err != nil {
		t.Fatalf("readInput file url: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Fatalf("file url data mismatch")
	}
}

func TestReadInputHTTP(t *testing.T) {
	data := testutil.MakeTestGIF()
	testutil.WithTransport(t, &testutil.FakeTransport{GIFData: data}, func() {
		got, err := readInput("https://example.test/preview.gif")
		if err != nil {
			t.Fatalf("readInput http: %v", err)
		}
		if !bytes.Equal(got, data) {
			t.Fatalf("http data mismatch")
		}
	})
}

func TestRunExtractContactSheet(t *testing.T) {
	data := testutil.MakeTestGIF()
	inPath := filepath.Join(t.TempDir(), "in.gif")
	if err := os.WriteFile(inPath, data, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "out.png")

	opts := model.Options{
		GifInput:      inPath,
		StillsCount:   2,
		StillsCols:    2,
		StillsPadding: 1,
		OutPath:       outPath,
	}
	if err := runExtract(opts); err != nil {
		t.Fatalf("runExtract failed: %v", err)
	}
	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatalf("png decode failed: %v", err)
	}
	if img.Bounds().Dx() != 5 || img.Bounds().Dy() != 2 {
		t.Fatalf("unexpected sheet size %dx%d", img.Bounds().Dx(), img.Bounds().Dy())
	}
}

func TestRunExtractStill(t *testing.T) {
	data := testutil.MakeTestGIF()
	inPath := filepath.Join(t.TempDir(), "in.gif")
	if err := os.WriteFile(inPath, data, 0o644); err != nil {
		t.Fatalf("write input: %v", err)
	}
	outPath := filepath.Join(t.TempDir(), "still.png")

	opts := model.Options{
		GifInput: inPath,
		StillSet: true,
		StillAt:  60 * time.Millisecond,
		OutPath:  outPath,
	}
	if err := runExtract(opts); err != nil {
		t.Fatalf("runExtract failed: %v", err)
	}
	out, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if !bytes.HasPrefix(out, []byte{0x89, 'P', 'N', 'G'}) {
		t.Fatalf("expected png output")
	}
}
