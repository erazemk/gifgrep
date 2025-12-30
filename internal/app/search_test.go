package app

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestFetchTenorAndGIF(t *testing.T) {
	gifData := makeTestGIF()
	withTransport(t, &fakeTransport{gifData: gifData}, func() {
		if _, err := search("cats", cliOptions{Source: "nope"}); err == nil {
			t.Fatalf("expected unknown source error")
		}
		out, err := fetchTenorV1("cats", cliOptions{Limit: 1})
		if err != nil {
			t.Fatalf("fetchTenorV1 failed: %v", err)
		}
		if len(out) != 1 {
			t.Fatalf("expected 1 result")
		}
		if out[0].PreviewURL == "" || out[0].URL == "" {
			t.Fatalf("missing URLs")
		}
		data, err := fetchGIF(out[0].PreviewURL)
		if err != nil {
			t.Fatalf("fetchGIF failed: %v", err)
		}
		if len(data) == 0 {
			t.Fatalf("expected gif data")
		}
		if _, err := fetchGIF("https://example.test/missing.gif"); err == nil {
			t.Fatalf("expected fetchGIF error")
		}
	})
}

type badTenorTransport struct{}

func (t *badTenorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := "not-json"
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

type statusTenorTransport struct{}

func (t *statusTenorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 500,
		Body:       io.NopCloser(strings.NewReader("oops")),
	}, nil
}

func TestFetchTenorErrors(t *testing.T) {
	withTransport(t, &badTenorTransport{}, func() {
		if _, err := fetchTenorV1("cats", cliOptions{Limit: 1}); err == nil {
			t.Fatalf("expected json error")
		}
	})
	withTransport(t, &statusTenorTransport{}, func() {
		if _, err := fetchTenorV1("cats", cliOptions{Limit: 1}); err == nil {
			t.Fatalf("expected status error")
		}
	})
}

type noMediaTransport struct{}

func (t *noMediaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	body := `{"results":[{"id":"1","title":"No Media","media":[]},{"id":"2","title":"Gif Only","media":[{"gif":{"url":"https://example.test/full.gif","dims":[10,5]}}]}]}`
	return &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func TestFetchTenorMediaFallbacks(t *testing.T) {
	withTransport(t, &noMediaTransport{}, func() {
		results, err := fetchTenorV1("cats", cliOptions{Limit: 2})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected one result, got %d", len(results))
		}
		if results[0].PreviewURL == "" {
			t.Fatalf("expected preview fallback")
		}
	})
}
