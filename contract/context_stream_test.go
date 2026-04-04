package contract

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestStreamJSONWithChannel(t *testing.T) {
	ctx := NewCtx(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil), nil)
	ch := make(chan any, 2)
	ch <- map[string]any{"id": 1}
	ch <- map[string]any{"id": 2}
	close(ch)

	if err := ctx.Stream(StreamConfig{Format: StreamFormatJSON, Source: (<-chan any)(ch)}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := ctx.W.(*httptest.ResponseRecorder).Body.String()
	lines := strings.Split(strings.TrimSpace(body), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	var v map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &v); err != nil {
		t.Fatalf("invalid json line: %v", err)
	}
	if int(v["id"].(float64)) != 1 {
		t.Fatalf("expected id=1, got %v", v["id"])
	}
}

func TestStreamTextWithGenerator(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	items := []string{"a", "b", "c"}
	i := 0
	gen := func() (string, error) {
		if i >= len(items) {
			return "", io.EOF
		}
		v := items[i]
		i++
		return v, nil
	}

	if err := ctx.Stream(StreamConfig{Format: StreamFormatText, Source: gen}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got, want := rec.Body.String(), "a\nb\nc\n"; got != want {
		t.Fatalf("unexpected body:\nwant: %q\ngot:  %q", want, got)
	}
}

func TestStreamSSEWithGenerator(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	items := []SSEEvent{
		{ID: "1", Event: "tick", Data: "hello"},
		{ID: "2", Event: "tick", Data: "world"},
	}
	i := 0
	gen := func() (SSEEvent, error) {
		if i >= len(items) {
			return SSEEvent{}, io.EOF
		}
		v := items[i]
		i++
		return v, nil
	}

	if err := ctx.Stream(StreamConfig{Format: StreamFormatSSE, Source: gen}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := rec.Body.String()
	if !strings.Contains(out, "id: 1\n") || !strings.Contains(out, "data: hello\n") {
		t.Fatalf("unexpected sse output: %q", out)
	}
	if rec.Header().Get("Content-Type") != "text/event-stream" {
		t.Fatalf("expected text/event-stream, got %q", rec.Header().Get("Content-Type"))
	}
}

func TestStreamJSONWithRetry(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	calls := 0
	gen := func() (any, error) {
		calls++
		switch calls {
		case 1:
			return nil, errors.New("transient")
		case 2:
			return map[string]string{"ok": "1"}, nil
		default:
			return nil, io.EOF
		}
	}

	if err := ctx.Stream(StreamConfig{Format: StreamFormatJSON, Source: gen, MaxRetry: 1, RetryDelay: time.Millisecond}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
	if !strings.Contains(rec.Body.String(), `{"ok":"1"}`) {
		t.Fatalf("missing expected payload: %q", rec.Body.String())
	}
}

func TestStreamTextWithRetryExhausted(t *testing.T) {
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, httptest.NewRequest(http.MethodGet, "/", nil), nil)

	failErr := errors.New("still failing")
	gen := func() (string, error) {
		return "", failErr
	}

	err := ctx.Stream(StreamConfig{Format: StreamFormatText, Source: gen, MaxRetry: 1, RetryDelay: time.Millisecond})
	if !errors.Is(err, failErr) {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestStreamSSEWithChannelCanceled(t *testing.T) {
	baseReq := httptest.NewRequest(http.MethodGet, "/", nil)
	cancelCtx, cancel := context.WithCancel(baseReq.Context())
	cancel()
	req := baseReq.WithContext(cancelCtx)
	rec := httptest.NewRecorder()
	ctx := NewCtx(rec, req, nil)

	ch := make(chan SSEEvent)
	defer close(ch)

	err := ctx.Stream(StreamConfig{Format: StreamFormatSSE, Source: (<-chan SSEEvent)(ch)})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}
