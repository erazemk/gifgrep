package app

import (
	"os"
	"strconv"
	"strings"
)

func truncateRunes(s string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= width {
		return s
	}
	return string(runes[:width])
}

func runeLen(s string) int {
	return len([]rune(s))
}

func clampDelay(delay int) int {
	if delay < 20 {
		return 20
	}
	if delay > 1000 {
		return 1000
	}
	return delay
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func cellAspectRatio() float64 {
	if raw := strings.TrimSpace(os.Getenv("GIFGREP_CELL_ASPECT")); raw != "" {
		if v, err := strconv.ParseFloat(raw, 64); err == nil && v > 0.1 && v < 2 {
			return v
		}
	}
	return 0.5
}

func useSoftwareAnimation() bool {
	if raw := strings.TrimSpace(os.Getenv("GIFGREP_SOFTWARE_ANIM")); raw != "" {
		raw = strings.ToLower(raw)
		return raw == "1" || raw == "true" || raw == "yes"
	}
	termProgram := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(termProgram, "ghostty") || strings.Contains(term, "ghostty") {
		return true
	}
	return false
}
