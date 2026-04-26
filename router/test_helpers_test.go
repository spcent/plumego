package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mustAddRoute(r *Router, method, path string, handler http.Handler, opts ...RouteOption) {
	if err := r.AddRoute(method, path, handler, opts...); err != nil {
		panic(err)
	}
}

func serveRouter(r *Router, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func assertResponseStatus(t testing.TB, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d", want, rec.Code)
	}
}

func assertResponseBody(t testing.TB, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := rec.Body.String(); got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func assertTrimmedResponseBody(t testing.TB, rec *httptest.ResponseRecorder, want string) {
	t.Helper()
	if got := strings.TrimSpace(rec.Body.String()); got != want {
		t.Fatalf("expected body %q, got %q", want, got)
	}
}

func assertResponseHeader(t testing.TB, rec *httptest.ResponseRecorder, key, want string) {
	t.Helper()
	if got := rec.Header().Get(key); got != want {
		t.Fatalf("expected header %s %q, got %q", key, want, got)
	}
}
