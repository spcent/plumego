package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spcent/plumego/contract"
)

type stubSpan struct {
	ended     bool
	status    int
	bytes     int
	requestID string
	traceID   string
	spanID    string
}

func (s *stubSpan) End(status, bytes int, requestID string) {
	s.ended = true
	s.status = status
	s.bytes = bytes
	s.requestID = requestID
}

func (s *stubSpan) TraceID() string { return s.traceID }
func (s *stubSpan) SpanID() string  { return s.spanID }

type stubTracer struct {
	span *stubSpan
}

func (t *stubTracer) Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan) {
	t.span = &stubSpan{traceID: "trace-1", spanID: "span-1"}
	ctx = contract.WithTraceContext(ctx, contract.TraceContext{
		TraceID: contract.TraceID(t.span.traceID),
		SpanID:  contract.SpanID(t.span.spanID),
	})
	return ctx, t.span
}

func TestPrepareRequestAddsFallbackRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
	rec := httptest.NewRecorder()

	prepared := PrepareRequest(rec, req)

	if prepared.RequestID == "" {
		t.Fatal("expected fallback request id to be generated")
	}
	if got := rec.Header().Get(contract.RequestIDHeader); got == "" {
		t.Fatalf("expected %s header to be set", contract.RequestIDHeader)
	}
	if got := contract.RequestIDFromContext(prepared.Request.Context()); got != prepared.RequestID {
		t.Fatalf("expected request context id %q, got %q", prepared.RequestID, got)
	}
}

func TestBeginTraceAttachesSpanHeaderAndContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/trace", nil)
	rec := httptest.NewRecorder()
	prepared := PrepareRequest(rec, req)
	tracer := &stubTracer{}

	r, span, spanID := BeginTrace(rec, prepared, tracer.Start)

	if span == nil {
		t.Fatal("expected span to be returned")
	}
	if spanID != "span-1" {
		t.Fatalf("expected span id %q, got %q", "span-1", spanID)
	}
	if got := rec.Header().Get(SpanIDHeader); got != "span-1" {
		t.Fatalf("expected span header %q, got %q", "span-1", got)
	}
	tc := contract.TraceContextFromContext(r.Context())
	if tc == nil || string(tc.SpanID) != "span-1" {
		t.Fatalf("expected trace context with span id, got %+v", tc)
	}
}

func TestPreparedRequestCompleteAndEndTrace(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		RoutePattern: "/users/:id",
	}))
	rec := httptest.NewRecorder()
	prepared := PrepareRequest(rec, req)
	prepared.Recorder.WriteHeader(http.StatusCreated)
	_, _ = prepared.Recorder.Write([]byte("ok"))

	metrics := prepared.Complete(prepared.Request)
	if metrics.ObservedPath() != "/users/:id" {
		t.Fatalf("expected observed path %q, got %q", "/users/:id", metrics.ObservedPath())
	}

	span := &stubSpan{}
	EndTrace(span, metrics)
	if !span.ended {
		t.Fatal("expected span to be ended")
	}
	if span.status != http.StatusCreated || span.bytes != len("ok") {
		t.Fatalf("unexpected span end payload: status=%d bytes=%d", span.status, span.bytes)
	}
	if span.requestID != metrics.RequestID {
		t.Fatalf("expected request id %q, got %q", metrics.RequestID, span.requestID)
	}
}
