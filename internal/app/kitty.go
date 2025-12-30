package app

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"strings"
)

type kittyData struct {
	Action      string
	ID          uint32
	Data        []byte
	Cols        int
	Rows        int
	PlacementID int
	DelayMS     int
	NoCursor    bool
}

func sendKittyAnimation(out *bufio.Writer, anim *gifAnimation, cols, rows int) {
	if anim == nil || len(anim.Frames) == 0 {
		return
	}
	base := anim.Frames[0]
	sendKittyData(out, kittyData{
		Action:      "T",
		ID:          anim.ID,
		Data:        base.PNG,
		Cols:        cols,
		Rows:        rows,
		PlacementID: 1,
		NoCursor:    true,
	})
	for i := 1; i < len(anim.Frames); i++ {
		frame := anim.Frames[i]
		sendKittyData(out, kittyData{
			Action:  "f",
			ID:      anim.ID,
			Data:    frame.PNG,
			DelayMS: frame.DelayMS,
		})
	}
	sendKittyAnimDelay(out, anim.ID, clampDelay(base.DelayMS))
	sendKittyAnimStart(out, anim.ID)
}

func sendKittyFrame(out *bufio.Writer, id uint32, frame gifFrame, cols, rows int) {
	sendKittyData(out, kittyData{
		Action:      "T",
		ID:          id,
		Data:        frame.PNG,
		Cols:        cols,
		Rows:        rows,
		PlacementID: 1,
		NoCursor:    true,
	})
}

func sendKittyData(out *bufio.Writer, data kittyData) {
	encoded := base64.StdEncoding.EncodeToString(data.Data)
	const chunkSize = 4096
	first := true
	for len(encoded) > 0 {
		chunk := encoded
		if len(chunk) > chunkSize {
			chunk = chunk[:chunkSize]
		}
		encoded = encoded[len(chunk):]
		more := 0
		if len(encoded) > 0 {
			more = 1
		}
		if first {
			params := []string{
				fmt.Sprintf("a=%s", data.Action),
				"f=100",
				fmt.Sprintf("i=%d", data.ID),
				fmt.Sprintf("m=%d", more),
				"q=2",
			}
			if data.Cols > 0 {
				params = append(params, fmt.Sprintf("c=%d", data.Cols))
			}
			if data.Rows > 0 {
				params = append(params, fmt.Sprintf("r=%d", data.Rows))
			}
			if data.PlacementID > 0 {
				params = append(params, fmt.Sprintf("p=%d", data.PlacementID))
			}
			if data.NoCursor {
				params = append(params, "C=1")
			}
			if data.Action == "f" && data.DelayMS > 0 {
				params = append(params, fmt.Sprintf("z=%d", data.DelayMS))
			}
			_, _ = fmt.Fprintf(out, "\x1b_G%s;", strings.Join(params, ","))
			first = false
		} else {
			if data.Action == "f" {
				_, _ = fmt.Fprintf(out, "\x1b_Ga=f,m=%d;", more)
			} else {
				_, _ = fmt.Fprintf(out, "\x1b_Gm=%d;", more)
			}
		}
		_, _ = fmt.Fprint(out, chunk)
		_, _ = fmt.Fprint(out, "\x1b\\")
	}
}

func sendKittyAnimDelay(out *bufio.Writer, id uint32, delayMS int) {
	if delayMS <= 0 {
		return
	}
	_, _ = fmt.Fprintf(out, "\x1b_Ga=a,i=%d,r=1,z=%d,q=2\x1b\\", id, delayMS)
}

func sendKittyAnimStart(out *bufio.Writer, id uint32) {
	_, _ = fmt.Fprintf(out, "\x1b_Ga=a,i=%d,s=3,v=1,q=2\x1b\\", id)
}

func placeKittyImage(out *bufio.Writer, id uint32, cols, rows int) {
	if id == 0 {
		return
	}
	_, _ = fmt.Fprintf(out, "\x1b_Ga=p,i=%d,p=1,c=%d,r=%d,C=1,q=2\x1b\\", id, cols, rows)
}

func deleteKittyImage(out *bufio.Writer, id uint32) {
	if id == 0 {
		return
	}
	_, _ = fmt.Fprintf(out, "\x1b_Ga=d,d=I,i=%d,q=2\x1b\\", id)
}
