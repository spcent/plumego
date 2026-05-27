package appmiddleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	plumelog "github.com/spcent/plumego/log"
)

func discardLogger() plumelog.StructuredLogger {
	return plumelog.NewLogger(plumelog.LoggerConfig{Format: plumelog.LoggerFormatDiscard})
}

func TestRequireWriteKey(t *testing.T) {
	okHandler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("empty key is a no-op — all requests pass", func(t *testing.T) {
		mw := RequireWriteKey("", discardLogger())(okHandler)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("correct key passes through to handler", func(t *testing.T) {
		mw := RequireWriteKey("secret", discardLogger())(okHandler)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(WriteKeyHeader, "secret")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("missing header returns 401", func(t *testing.T) {
		mw := RequireWriteKey("secret", discardLogger())(okHandler)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})

	t.Run("wrong key returns 401", func(t *testing.T) {
		mw := RequireWriteKey("secret", discardLogger())(okHandler)
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(WriteKeyHeader, "wrong")
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
		}
	})
}
