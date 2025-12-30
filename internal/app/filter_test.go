package app

import "testing"

func TestFilterResults(t *testing.T) {
	items := []gifResult{
		{Title: "Happy Cat", Tags: []string{"cute", "cat"}},
		{Title: "Angry Dog", Tags: []string{"dog", "angry"}},
		{Title: "Cat Panic", Tags: []string{"cat", "panic"}},
	}

	out, err := filterResults(items, "cat", cliOptions{Regex: true, IgnoreCase: true})
	if err != nil || len(out) != 2 {
		t.Fatalf("regex filter failed: %v len=%d", err, len(out))
	}

	out, err = filterResults(items, "panic", cliOptions{Mood: "panic", IgnoreCase: true})
	if err != nil || len(out) != 1 {
		t.Fatalf("mood filter failed: %v len=%d", err, len(out))
	}

	out, err = filterResults(items, "panic", cliOptions{Mood: "panic", IgnoreCase: true, Invert: true})
	if err != nil || len(out) != 2 {
		t.Fatalf("invert filter failed: %v len=%d", err, len(out))
	}

	_, err = filterResults(items, "(", cliOptions{Regex: true})
	if err == nil {
		t.Fatalf("expected regex error")
	}
}
