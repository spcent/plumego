package bodylimit

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spcent/plumego/log"
)

type panicLogger struct{}

func (panicLogger) WithFields(log.Fields) log.StructuredLogger      { return panicLogger{} }
func (panicLogger) With(string, any) log.StructuredLogger           { return panicLogger{} }
func (panicLogger) Debug(string, ...log.Fields)                     {}
func (panicLogger) Info(string, ...log.Fields)                      {}
func (panicLogger) Warn(string, ...log.Fields)                      { panic("logger panic") }
func (panicLogger) Error(string, ...log.Fields)                     {}
func (panicLogger) DebugCtx(context.Context, string, ...log.Fields) {}
func (panicLogger) InfoCtx(context.Context, string, ...log.Fields)  {}
func (panicLogger) WarnCtx(context.Context, string, ...log.Fields)  { panic("logger panic") }
func (panicLogger) ErrorCtx(context.Context, string, ...log.Fields) {}
func (panicLogger) Fatal(string, ...log.Fields)                     {}
func (panicLogger) FatalCtx(context.Context, string, ...log.Fields) {}

func TestBodyLimit(t *testing.T) {
	mw := BodyLimit(10, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodPost, "/", strings.NewReader("this body is definitely too long"))
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
}

func TestMiddlewareWithConfigMatchesBodyLimit(t *testing.T) {
	for _, tt := range []struct {
		name string
		mw   func() func(http.Handler) http.Handler
	}{
		{name: "positional", mw: func() func(http.Handler) http.Handler {
			return BodyLimit(5, nil)
		}},
		{name: "config", mw: func() func(http.Handler) http.Handler {
			return MiddlewareWithConfig(Config{MaxBytes: 5})
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			handler := tt.mw()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := io.ReadAll(r.Body)
				if err != nil {
					return
				}
				t.Fatal("expected body read to fail")
			}))

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusRequestEntityTooLarge {
				t.Fatalf("expected 413, got %d", rec.Code)
			}
		})
	}
}

func TestDefaultConfigReturnsConfiguredBodyLimit(t *testing.T) {
	cfg := DefaultConfig(12, nil)
	if cfg.MaxBytes != 12 {
		t.Fatalf("MaxBytes = %d, want 12", cfg.MaxBytes)
	}
	if cfg.Logger != nil {
		t.Fatalf("Logger = %v, want nil", cfg.Logger)
	}
}

func TestBodyLimitAllowsExactLimit(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("hello"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "hello" {
		t.Fatalf("expected exact body to pass through, got %q", got)
	}
}

func TestBodyLimitDetectsSingleReadOverrun(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 16)
		n, err := r.Body.Read(buf)
		if n != 5 {
			t.Fatalf("first read bytes = %d, want 5", n)
		}
		if err != errRequestTooLarge {
			t.Fatalf("first read error = %v, want %v", err, errRequestTooLarge)
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "REQUEST_BODY_TOO_LARGE") {
		t.Fatalf("expected structured body limit error, got %q", rec.Body.String())
	}
}

func TestBodyLimitSuppressesDownstreamWritesAfterLimitError(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "handler saw "+err.Error(), http.StatusBadRequest)
			return
		}
		t.Fatal("expected body read to fail")
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "REQUEST_BODY_TOO_LARGE") {
		t.Fatalf("expected structured body limit error, got %q", body)
	}
	if strings.Contains(body, "handler saw") || strings.Contains(body, errRequestTooLarge.Error()) {
		t.Fatalf("downstream error write polluted body limit response: %q", body)
	}
}

func TestBodyLimitLoggerPanicDoesNotBlockLimitError(t *testing.T) {
	mw := BodyLimit(5, panicLogger{})
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
	if !strings.Contains(rec.Body.String(), "REQUEST_BODY_TOO_LARGE") {
		t.Fatalf("expected structured body limit error, got %q", rec.Body.String())
	}
}

func TestBodyLimitDownstreamWriteAfterOverrunReportsConsumed(t *testing.T) {
	mw := BodyLimit(5, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err == nil {
			t.Fatal("expected body read to fail")
		}
		n, writeErr := w.Write([]byte("late handler body"))
		if writeErr != nil {
			t.Fatalf("post-overrun write error = %v, want nil", writeErr)
		}
		if n != len("late handler body") {
			t.Fatalf("post-overrun write bytes = %d, want %d", n, len("late handler body"))
		}
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("toolong"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "REQUEST_BODY_TOO_LARGE") {
		t.Fatalf("expected structured body limit error, got %q", body)
	}
	if strings.Contains(body, "late handler body") {
		t.Fatalf("post-overrun write polluted body limit response: %q", body)
	}
}

func TestBodyLimitDisabledPassesThrough(t *testing.T) {
	mw := BodyLimit(0, nil)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		_, _ = w.Write(body)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("unlimited"))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Body.String(); got != "unlimited" {
		t.Fatalf("expected disabled limit to pass body through, got %q", got)
	}
}

func TestBodyLimitResponseWriterUnwrap(t *testing.T) {
	underlying := httptest.NewRecorder()
	w := &bodyLimitResponseWriter{ResponseWriter: underlying}

	if got := w.Unwrap(); got != underlying {
		t.Fatalf("Unwrap() = %v, want underlying writer", got)
	}
}
