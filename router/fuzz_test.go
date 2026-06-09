package router

import (
	"net/http"
	"strings"
	"testing"
)

func FuzzRoutePathNormalizationDoesNotPanic(f *testing.F) {
	for _, seed := range []string{
		"",
		"/",
		"users/:id",
		"///api/users/:id",
		"/files/*path",
		"/files//readme",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, raw string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("route path helpers panicked for %q: %v", raw, recovered)
			}
		}()

		path := canonicalRoutePath(raw)
		segments := compilePathSegments(path)
		_ = validateRouteSegments(path, segments)
		_ = fastNormalizePath(raw)
	})
}

func FuzzReverseRoutingURLDoesNotPanic(f *testing.F) {
	for _, seed := range []struct {
		id   string
		rest string
	}{
		{id: "42", rest: "docs/readme.md"},
		{id: "a b", rest: "nested/file name.txt"},
		{id: "slash", rest: "a/b/c"},
		{id: "symbols", rest: "q?x=1#frag"},
	} {
		f.Add(seed.id, seed.rest)
	}

	r := NewRouter()
	if err := r.AddRoute(http.MethodGet, "/files/:id/*path", http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), WithRouteName("files.show")); err != nil {
		f.Fatalf("add route failed: %v", err)
	}

	f.Fuzz(func(t *testing.T, id, rest string) {
		defer func() {
			if recovered := recover(); recovered != nil {
				t.Fatalf("URL panicked for id=%q rest=%q: %v", id, rest, recovered)
			}
		}()

		got, err := r.URL("files.show", "id", id, "path", rest)
		if err != nil {
			// Empty params (and other invalid inputs) are expected errors, not panics.
			return
		}
		if got == "" {
			t.Fatalf("URL returned empty for non-empty params id=%q rest=%q", id, rest)
		}
		if !strings.HasPrefix(got, "/files/") {
			t.Fatalf("URL = %q, want /files/ prefix", got)
		}
	})
}
