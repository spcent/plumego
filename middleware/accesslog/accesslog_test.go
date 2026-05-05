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
	mwtracing "github.com/spcent/plumego/middleware/tracing"
)

type logEntry struct {
	msg    string
	fields log.Fields
}

type stubLogger struct {
	baseFields    log.Fields
	entries       *[]logEntry
	mu            *sync.Mutex
	panicOnRecord bool
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
	return &stubLogger{baseFields: merged, entries: l.entries, mu: l.mu, panicOnRecord: l.panicOnRecord}
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
	if l.panicOnRecord {
		panic("logger panic")
	}
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

func (t *stubTracer) Start(ctx context.Context, r *http.Request) (context.Context, mwtracing.TraceSpan) {
	t.started = true
	t.received = contract.RequestIDFromContext(ctx)
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

func (t *spanContextTracer) Start(ctx context.Context, r *http.Request) (context.Context, mwtracing.TraceSpan) {
	t.span = &spanContextSpan{traceID: "trace-ctx", spanID: "span-123"}
	return ctx, t.span
}

type panicStartTracer struct{}

func (panicStartTracer) Start(ctx context.Context, r *http.Request) (context.Context, mwtracing.TraceSpan) {
	panic("trace start panic")
}

func TestMiddlewareAddsStructuredFields(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}

	mw := Middleware(logger, nil, tracer)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if contract.RequestIDFromContext(r.Context()) == "" {
			t.Fatalf("request id should be present in context")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})

	req := httptest.NewRequest(http.MethodPost, "/structured", nil)
	req.Header.Set(contract.RequestIDHeader, "trace-123")
	rc := contract.RequestContext{RoutePattern: "/structured/:id", RouteName: "structured"}
	req = req.WithContext(contract.WithRequestContext(req.Context(), rc))
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

func TestMiddlewareUsesUpdatedContextForTracer(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}
	mw := Middleware(logger, nil, tracer)

	req := httptest.NewRequest(http.MethodGet, "/context", nil)
	req.Header.Set(contract.RequestIDHeader, "ctx-trace")
	rec := httptest.NewRecorder()
	mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("context")) })).ServeHTTP(rec, req)

	if tracer.received != "ctx-trace" {
		t.Fatalf("tracer should receive request id from context, got %q", tracer.received)
	}
}

func TestMiddlewareCapturesSpanIDFromTracing(t *testing.T) {
	logger := newStubLogger()
	tracer := &spanContextTracer{}
	handler := Middleware(logger, nil, nil)(mwtracing.Middleware(tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

func TestMiddlewareLogsAndEndsTraceOnPanic(t *testing.T) {
	logger := newStubLogger()
	tracer := &stubTracer{}
	handler := Middleware(logger, nil, tracer)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		handler.ServeHTTP(rec, req)
	}()

	if !panicked {
		t.Fatal("expected panic to propagate")
	}
	if len(*logger.entries) != 1 {
		t.Fatalf("expected one log entry, got %d", len(*logger.entries))
	}
	if got := (*logger.entries)[0].fields["status"]; got != http.StatusAccepted {
		t.Fatalf("logged status = %v, want %d", got, http.StatusAccepted)
	}
	if tracer.span == nil || !tracer.span.ended {
		t.Fatalf("expected tracer span to end during panic unwinding")
	}
}

func TestMiddlewarePreservesDownstreamPanicWhenLoggerPanics(t *testing.T) {
	logger := newStubLogger()
	logger.panicOnRecord = true
	handler := Middleware(logger, nil, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("downstream panic")
	}))

	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()
	defer func() {
		if rec := recover(); rec != "downstream panic" {
			t.Fatalf("panic = %v, want downstream panic", rec)
		}
	}()
	handler.ServeHTTP(rec, req)
}

func TestMiddlewareContinuesWhenTracerStartPanics(t *testing.T) {
	logger := newStubLogger()
	handler := Middleware(logger, nil, panicStartTracer{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))

	req := httptest.NewRequest(http.MethodGet, "/trace-start", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	logger.mu.Lock()
	entries := append([]logEntry(nil), (*logger.entries)...)
	logger.mu.Unlock()
	if len(entries) != 1 {
		t.Fatalf("log entries = %d, want 1", len(entries))
	}
	if _, ok := entries[0].fields["span_id"]; ok {
		t.Fatalf("span_id should be omitted after tracer start panic: %+v", entries[0].fields)
	}
}

func TestRedactedLogFieldsMasksSensitiveKeys(t *testing.T) {
	fields := redactedLogFields(map[string]any{
		"method":              http.MethodGet,
		"authorization_token": "secret",
	})

	if fields["method"] != http.MethodGet {
		t.Fatalf("method field = %v, want %s", fields["method"], http.MethodGet)
	}
	if fields["authorization_token"] == "secret" {
		t.Fatal("sensitive access log field was not redacted")
	}
}

func TestMiddlewareRejectsNilLogger(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic when logger is nil")
		}
	}()
	_ = Middleware(nil, nil, nil)
}

func TestMiddlewareEReturnsNilLoggerError(t *testing.T) {
	mw, err := MiddlewareE(nil, nil, nil)
	if !errors.Is(err, ErrNilLogger) {
		t.Fatalf("error = %v, want %v", err, ErrNilLogger)
	}
	if mw != nil {
		t.Fatalf("middleware = %v, want nil", mw)
	}
}

func TestMiddlewareEConstructsMiddleware(t *testing.T) {
	mw, err := MiddlewareE(newStubLogger(), nil, nil)
	if err != nil {
		t.Fatalf("MiddlewareE returned error: %v", err)
	}
	if mw == nil {
		t.Fatalf("expected middleware")
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
func (w *hijackWriter) WriteHeader(_ int)           {}
func (w *hijackWriter) Write(p []byte) (int, error) { return len(p), nil }
func (w *hijackWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijacked = true
	return nil, nil, errors.New("hijack")
}

func TestMiddlewarePreservesHijacker(t *testing.T) {
	logger := newStubLogger()
	mw := Middleware(logger, nil, nil)
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
