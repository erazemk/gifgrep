package app

import (
	"encoding"
	"errors"
	"strconv"
	"strings"
	"time"
)

type DurationValue time.Duration

var _ encoding.TextUnmarshaler = (*DurationValue)(nil)

func (d *DurationValue) UnmarshalText(text []byte) error {
	raw := strings.TrimSpace(string(text))
	if raw == "" {
		return errors.New("empty duration")
	}
	if parsed, err := time.ParseDuration(raw); err == nil {
		*d = DurationValue(parsed)
		return nil
	}
	if secs, err := strconv.ParseFloat(raw, 64); err == nil {
		if secs < 0 {
			return errors.New("negative duration")
		}
		*d = DurationValue(time.Duration(secs * float64(time.Second)))
		return nil
	}
	return errors.New("invalid duration")
}
