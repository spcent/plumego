package http

import (
	"errors"
	"net/http"
	"testing"
)

func TestLoggingMiddlewareReportsSuccess(t *testing.T) {
	var got RequestLogEntry
	logged := false

	mw := Logging(func(entry RequestLogEntry) {
		logged = true
		got = entry
	})

	next := mw(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			Status:     "204 No Content",
			StatusCode: http.StatusNoContent,
			Body:       http.NoBody,
		}, nil
	})

	req, _ := http.NewRequest(http.MethodGet, "https://example.com/path", nil)
	resp, err := next(req)
	if err != nil {
		t.Fatalf("next returned error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status code = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !logged {
		t.Fatal("expected logging callback to be invoked")
	}
	if got.Method != http.MethodGet {
		t.Fatalf("method = %q, want %q", got.Method, http.MethodGet)
	}
	if got.Host != "example.com" {
		t.Fatalf("host = %q, want %q", got.Host, "example.com")
	}
	if got.Status != "204 No Content" {
		t.Fatalf("status = %q, want %q", got.Status, "204 No Content")
	}
	if got.Err != nil {
		t.Fatalf("err = %v, want nil", got.Err)
	}
	if got.Duration < 0 {
		t.Fatalf("duration = %s, want non-negative", got.Duration)
	}
}

func TestLoggingMiddlewareReportsError(t *testing.T) {
	wantErr := errors.New("boom")
	var got RequestLogEntry

	mw := Logging(func(entry RequestLogEntry) {
		got = entry
	})

	next := mw(func(req *http.Request) (*http.Response, error) {
		return nil, wantErr
	})

	req, _ := http.NewRequest(http.MethodPost, "https://example.com/path", nil)
	resp, err := next(req)
	if !errors.Is(err, wantErr) {
		t.Fatalf("err = %v, want %v", err, wantErr)
	}
	if resp != nil {
		t.Fatalf("resp = %#v, want nil", resp)
	}
	if !errors.Is(got.Err, wantErr) {
		t.Fatalf("logged err = %v, want %v", got.Err, wantErr)
	}
	if got.Status != "" {
		t.Fatalf("status = %q, want empty", got.Status)
	}
}
