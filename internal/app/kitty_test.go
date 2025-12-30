package app

import (
	"bufio"
	"bytes"
	"strings"
	"testing"
)

func TestKittySequences(t *testing.T) {
	var buf bytes.Buffer
	out := bufio.NewWriter(&buf)
	sendKittyData(out, kittyData{
		Action:      "T",
		ID:          7,
		Data:        []byte{1, 2, 3},
		Cols:        2,
		Rows:        3,
		PlacementID: 1,
		NoCursor:    true,
	})
	sendKittyAnimDelay(out, 7, 80)
	sendKittyAnimStart(out, 7)
	placeKittyImage(out, 7, 2, 3)
	deleteKittyImage(out, 7)
	_ = out.Flush()

	s := buf.String()
	if !strings.Contains(s, "a=T") || !strings.Contains(s, "i=7") {
		t.Fatalf("missing kitty header")
	}
	if !strings.Contains(s, "a=a") || !strings.Contains(s, "a=p") || !strings.Contains(s, "a=d") {
		t.Fatalf("missing kitty controls")
	}

	buf.Reset()
	sendKittyAnimation(out, &gifAnimation{
		ID: 2,
		Frames: []gifFrame{
			{PNG: []byte{1, 2, 3}, DelayMS: 80},
			{PNG: []byte{4, 5, 6}, DelayMS: 90},
		},
	}, 5, 4)
	_ = out.Flush()
	if !strings.Contains(buf.String(), "a=f") {
		t.Fatalf("expected frame data")
	}

	buf.Reset()
	sendKittyAnimDelay(out, 7, 0)
	placeKittyImage(out, 0, 2, 3)
	deleteKittyImage(out, 0)
	_ = out.Flush()
	if buf.Len() != 0 {
		t.Fatalf("expected no output for no-op calls")
	}

	buf.Reset()
	large := make([]byte, 5000)
	for i := range large {
		large[i] = byte(i % 251)
	}
	sendKittyData(out, kittyData{Action: "f", ID: 9, Data: large})
	_ = out.Flush()
	if !strings.Contains(buf.String(), "a=f") {
		t.Fatalf("expected chunked frame data")
	}
}
