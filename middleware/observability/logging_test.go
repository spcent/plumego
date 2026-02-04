package observability

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	contract "github.com/spcent/plumego/contract"
	log "github.com/spcent/plumego/log"
)

type logEntry struct {
	msg    string
	fields log.Fields
}

type stubLogger struct {
	baseFields log.Fields
	entries    *[]logEntry
	mu         *sync.Mutex
}

func newStubLogger() *stubLogger {
	entries := make([]logEntry, 0)
	return &stubLogger{
		baseFields: log.Fields{},
		entries:    &entries,
		mu:         &sync.Mutex{},
	}
}

func (l *stubLogger) WithFields(fields log.Fields) log.StructuredLogger {
	merged := make(log.Fields, len(l.baseFields)+len(fields))
	for k, v := range l.baseFields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &stubLogger{baseFields: merged, entries: l.entries, mu: l.mu}
}

func (l *stubLogger) Debug(msg string, fields log.Fields) { l.record(msg, fields) }
func (l *stubLogger) Info(msg string, fields log.Fields)  { l.record(msg, fields) }
func (l *stubLogger) Warn(msg string, fields log.Fields)  { l.record(msg, fields) }
func (l *stubLogger) Error(msg string, fields log.Fields) { l.record(msg, fields) }

func (l *stubLogger) DebugCtx(ctx context.Context, msg string, fields log.Fields) {
	l.record(msg, fields)
}
func (l *stubLogger) InfoCtx(ctx context.Context, msg string, fields log.Fields) {
	l.record(msg, fields)
}
func (l *stubLogger) WarnCtx(ctx context.Context, msg string, fields log.Fields) {
	l.record(msg, fields)
}
func (l *stubLogger) ErrorCtx(ctx context.Context, msg string, fields log.Fields) {
	l.record(msg, fields)
}

func (l *stubLogger) record(msg string, fields log.Fields) {
	merged := make(log.Fields, len(l.baseFields)+len(fields))
	for k, v := range l.baseFields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	*l.entries = append(*l.entries, logEntry{msg: msg, fields: merged})
}

type stubMetrics struct {
	mu       sync.Mutex
	observed []RequestMetrics
}

func (m *stubMetrics) Observe(ctx context.Context, metrics RequestMetrics) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.observed = append(m.observed, metrics)
}

type stubSpan struct {
	endedWith RequestMetrics
	ended     bool
}

func (s *stubSpan) End(metrics RequestMetrics) {
	s.endedWith = metrics
	s.ended = true
}

type stubTracer struct {
	started   bool
	received  string
	span      *stubSpan
	startLock sync.Mutex
}

func (t *stubTracer) Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan) {
	t.startLock.Lock()
	defer t.startLock.Unlock()
	t.started = true
	t.received = contract.TraceIDFromContext(ctx)
	t.span = &stubSpan{}
	return ctx, t.span
}

type spanContextSpan struct {
	traceID   string
	spanID    string
	endedWith RequestMetrics
	ended     bool
}

func (s *spanContextSpan) End(metrics RequestMetrics) {
	s.endedWith = metrics
	s.ended = true
}

func (s *spanContextSpan) TraceID() string {
	return s.traceID
}

func (s *spanContextSpan) SpanID() string {
	return s.spanID
}

type spanContextTracer struct {
	span *spanContextSpan
}

func (t *spanContextTracer) Start(ctx context.Context, r *http.Request) (context.Context, TraceSpan) {
	t.span = &spanContextSpan{
		traceID: "trace-ctx",
		spanID:  "span-123",
	}
	return ctx, t.span
}

func TestLoggingAddsStructuredFields(t *testing.T) {
	logger := newStubLogger()
	metrics := &stubMetrics{}
	tracer := &stubTracer{}

	middleware := Logging(logger, metrics, tracer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contract.TraceIDFromContext(r.Context()) == "" {
			t.Fatalf("trace id should be present in context")
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/structured", nil)
	req.Header.Set("X-Request-ID", "trace-123")
	rc := contract.RequestContext{
		RoutePattern: "/structured/:id",
		RouteName:    "structured",
	}
	req = req.WithContext(context.WithValue(req.Context(), contract.RequestContextKey{}, rc))
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if got := rec.Header().Get("X-Request-ID"); got != "trace-123" {
		t.Fatalf("expected trace header to be propagated, got %q", got)
	}

	if len(*logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(*logger.entries))
	}

	entry := (*logger.entries)[0]
	if entry.msg != "request completed" {
		t.Fatalf("unexpected log message: %s", entry.msg)
	}

	expectedFields := map[string]any{
		"trace_id":    "trace-123",
		"method":      http.MethodPost,
		"path":        "/structured",
		"route":       "/structured/:id",
		"route_name":  "structured",
		"status":      http.StatusCreated,
		"bytes":       len("ok"),
		"user_agent":  "",
		"duration_ms": entry.fields["duration_ms"],
	}

	for key, expected := range expectedFields {
		if entry.fields[key] != expected {
			t.Fatalf("field %s expected %v, got %v", key, expected, entry.fields[key])
		}
	}

	if len(metrics.observed) != 1 {
		t.Fatalf("expected metrics to be recorded")
	}

	if !tracer.started || tracer.span == nil || !tracer.span.ended {
		t.Fatalf("tracer should have started and ended a span")
	}
}

func TestLoggingGeneratesTraceIDWhenMissing(t *testing.T) {
	logger := newStubLogger()
	middleware := Logging(logger, nil, nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contract.TraceIDFromContext(r.Context()) == "" {
			t.Fatalf("generated trace id should be in context")
		}
		w.Write([]byte("generated"))
	})

	req := httptest.NewRequest(http.MethodGet, "/no-trace", nil)
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got == "" {
		t.Fatalf("trace header should be set on response")
	}

	if len(*logger.entries) != 1 {
		t.Fatalf("expected log entry to be recorded")
	}

	if (*logger.entries)[0].fields["trace_id"] == "" {
		t.Fatalf("trace id field should not be empty")
	}
}

func TestLoggingUsesUpdatedContextForTracer(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}

	middleware := Logging(logger, nil, tracer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("context"))
	})

	req := httptest.NewRequest(http.MethodGet, "/context", nil)
	req.Header.Set("X-Request-ID", "ctx-trace")
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	if tracer.received != "ctx-trace" {
		t.Fatalf("tracer should receive trace id from context, got %q", tracer.received)
	}
}

func TestLoggingAddsSpanID(t *testing.T) {
	logger := newStubLogger()
	tracer := &spanContextTracer{}

	middleware := Logging(logger, nil, tracer)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tc := contract.TraceContextFromContext(r.Context())
		if tc == nil || tc.SpanID == "" {
			t.Fatalf("expected span context to be available")
		}
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/span", nil)
	req.Header.Set("X-Request-ID", "trace-ctx")
	rec := httptest.NewRecorder()

	middleware(handler).ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Span-ID"); got != "span-123" {
		t.Fatalf("expected span id header to be set, got %q", got)
	}

	if len(*logger.entries) == 0 {
		t.Fatalf("expected log entry to be recorded")
	}

	if (*logger.entries)[0].fields["span_id"] != "span-123" {
		t.Fatalf("expected span id field to be logged, got %v", (*logger.entries)[0].fields["span_id"])
	}
}

func BenchmarkLogging(b *testing.B) {
	logger := newStubLogger()
	middleware := Logging(logger, nil, nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	req := httptest.NewRequest(http.MethodGet, "/bench", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		middleware(handler).ServeHTTP(rec, req)
	}
}

type hijackWriter struct {
	header   http.Header
	hijacked bool
}

func (w *hijackWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *hijackWriter) WriteHeader(_ int) {}

func (w *hijackWriter) Write(p []byte) (int, error) { return len(p), nil }

func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, errors.New("hijack")
}

func TestLoggingPreservesHijacker(t *testing.T) {
	logger := newStubLogger()
	middleware := Logging(logger, nil, nil)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			t.Fatalf("expected Hijacker support")
		}
		_, _, _ = hj.Hijack()
		w.WriteHeader(http.StatusSwitchingProtocols)
	})

	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := &hijackWriter{}
	middleware(handler).ServeHTTP(rec, req)

	if !rec.hijacked {
		t.Fatalf("expected hijack to be forwarded to underlying writer")
	}
}

type closeNotifyWriter struct {
	header http.Header
	ch     chan bool
}

func (w *closeNotifyWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *closeNotifyWriter) WriteHeader(_ int) {}

func (w *closeNotifyWriter) Write(p []byte) (int, error) { return len(p), nil }

func (w *closeNotifyWriter) CloseNotify() <-chan bool {
	if w.ch == nil {
		w.ch = make(chan bool, 1)
		w.ch <- true
		close(w.ch)
	}
	return w.ch
}
