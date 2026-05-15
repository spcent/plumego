package router

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

func TestRouterPrintIncludesWildcardLabel(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, MethodAny, "/any/*path", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	var buf bytes.Buffer
	r.Print(&buf)

	assertOutputContains(t, buf.String(), "ANY    /any/*path   [wildcard]")
}

type routerCallbackWriter struct {
	buf bytes.Buffer
	fn  func()
}

func (w *routerCallbackWriter) Write(p []byte) (int, error) {
	if w.fn != nil {
		fn := w.fn
		w.fn = nil
		fn()
	}
	return w.buf.Write(p)
}

func TestRouterPrintWritesAfterSnapshot(t *testing.T) {
	r := NewRouter()
	mustAddRoute(r, http.MethodGet, "/users", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))

	writer := &routerCallbackWriter{fn: r.Freeze}
	done := make(chan struct{})
	go func() {
		defer close(done)
		r.Print(writer)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Print blocked while writer called back into router")
	}

	assertOutputContains(t, writer.buf.String(), "GET    /users")
}
