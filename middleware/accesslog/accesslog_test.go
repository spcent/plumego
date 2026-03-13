package accesslog

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
	"github.com/spcent/plumego/log"
	"github.com/spcent/plumego/metrics"
	mwtracing "github.com/spcent/plumego/middleware/tracing"
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
	return &stubLogger{baseFields: log.Fields{}, entries: &entries, mu: &sync.Mutex{}}
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
func (l *stubLogger) With(key string, value any) log.StructuredLogger {
	return l.WithFields(log.Fields{key: value})
}
func (l *stubLogger) Debug(msg string, fields ...log.Fields) { l.record(msg, first(fields)) }
func (l *stubLogger) Info(msg string, fields ...log.Fields)  { l.record(msg, first(fields)) }
func (l *stubLogger) Warn(msg string, fields ...log.Fields)  { l.record(msg, first(fields)) }
func (l *stubLogger) Error(msg string, fields ...log.Fields) { l.record(msg, first(fields)) }
func (l *stubLogger) DebugCtx(ctx context.Context, msg string, fields ...log.Fields) {
	l.record(msg, first(fields))
}
func (l *stubLogger) InfoCtx(ctx context.Context, msg string, fields ...log.Fields) {
	l.record(msg, first(fields))
}
func (l *stubLogger) WarnCtx(ctx context.Context, msg string, fields ...log.Fields) {
	l.record(msg, first(fields))
}
func (l *stubLogger) ErrorCtx(ctx context.Context, msg string, fields ...log.Fields) {
	l.record(msg, first(fields))
}
func (l *stubLogger) Fatal(msg string, fields ...log.Fields) { l.record(msg, first(fields)) }
func (l *stubLogger) FatalCtx(ctx context.Context, msg string, fields ...log.Fields) {
	l.record(msg, first(fields))
}

func first(fields []log.Fields) log.Fields {
	if len(fields) > 0 {
		return fields[0]
	}
	return nil
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

type stubSpan struct{ ended bool }

func (s *stubSpan) End(status, bytes int, traceID string) { s.ended = true }
func (s *stubSpan) TraceID() string                       { return "" }
func (s *stubSpan) SpanID() string                        { return "" }

type stubTracer struct {
	started  bool
	received string
	span     *stubSpan
}

func (t *stubTracer) Start(ctx context.Context, r *http.Request) (context.Context, metrics.TraceSpan) {
	t.started = true
	t.received = contract.TraceIDFromContext(ctx)
	t.span = &stubSpan{}
	return ctx, t.span
}

type spanContextSpan struct {
	traceID string
	spanID  string
	ended   bool
}

func (s *spanContextSpan) End(status, bytes int, traceID string) { s.ended = true }
func (s *spanContextSpan) TraceID() string                       { return s.traceID }
func (s *spanContextSpan) SpanID() string                        { return s.spanID }

type spanContextTracer struct{ span *spanContextSpan }

func (t *spanContextTracer) Start(ctx context.Context, r *http.Request) (context.Context, metrics.TraceSpan) {
	t.span = &spanContextSpan{traceID: "trace-ctx", spanID: "span-123"}
	return ctx, t.span
}

func TestLoggingAddsStructuredFields(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}

	mw := Logging(logger, nil, tracer)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contract.TraceIDFromContext(r.Context()) == "" {
			t.Fatalf("trace id should be present in context")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/structured", nil)
	req.Header.Set("X-Request-ID", "trace-123")
	rc := contract.RequestContext{RoutePattern: "/structured/:id", RouteName: "structured"}
	req = req.WithContext(context.WithValue(req.Context(), contract.RequestContextKey{}, rc))
	rec := httptest.NewRecorder()
	mw(handler).ServeHTTP(rec, req)

	if len(*logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(*logger.entries))
	}
	entry := (*logger.entries)[0]
	if entry.fields["request_id"] != "trace-123" {
		t.Fatalf("expected request id field")
	}
	if !tracer.started || tracer.span == nil || !tracer.span.ended {
		t.Fatalf("tracer should have started and ended a span")
	}
}

func TestLoggingUsesUpdatedContextForTracer(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}
	mw := Logging(logger, nil, tracer)

	req := httptest.NewRequest(http.MethodGet, "/context", nil)
	req.Header.Set("X-Request-ID", "ctx-trace")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("context")) })).ServeHTTP(rec, req)

	if tracer.received != "ctx-trace" {
		t.Fatalf("tracer should receive trace id from context, got %q", tracer.received)
	}
}

func TestMiddlewareCapturesSpanIDFromTracing(t *testing.T) {
	logger := newStubLogger()
	tracer := &spanContextTracer{}
	handler := Middleware(logger)(mwtracing.Middleware(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodGet, "/span-log", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if len(*logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(*logger.entries))
	}
	if (*logger.entries)[0].fields["span_id"] != "span-123" {
		t.Fatalf("expected span id in access log, got %v", (*logger.entries)[0].fields["span_id"])
	}
}

func TestMiddlewareRejectsNilLogger(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when logger is nil")
		}
	}()
	_ = Middleware(nil)
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
func (w *hijackWriter) WriteHeader(_ int)           {}
func (w *hijackWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, errors.New("hijack")
}

func TestLoggingPreservesHijacker(t *testing.T) {
	logger := newStubLogger()
	mw := Logging(logger, nil, nil)
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
	mw(handler).ServeHTTP(rec, req)

	if !rec.hijacked {
		t.Fatalf("expected hijack to be forwarded to underlying writer")
	}
}
