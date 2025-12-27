package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spcent/plumego/router"
)

func TestRegisterFromDir(t *testing.T) {
	dir := t.TempDir()

	write := func(path, content string) {
		full := filepath.Join(dir, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", path, err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	write("index.html", "<html>home</html>")
	write("_next/static/app.js", "console.log('hello')")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithCacheControl("public, max-age=31536000")); err != nil {
		t.Fatalf("register: %v", err)
	}

	cases := []struct {
		name         string
		method       string
		path         string
		expectStatus int
		expectBody   string
	}{
		{name: "asset", method: http.MethodGet, path: "/_next/static/app.js", expectStatus: http.StatusOK, expectBody: "console.log('hello')"},
		{name: "home", method: http.MethodGet, path: "/", expectStatus: http.StatusOK, expectBody: "<html>home</html>"},
		{name: "fallback", method: http.MethodGet, path: "/missing/page", expectStatus: http.StatusOK, expectBody: "<html>home</html>"},
		{name: "invalid method", method: http.MethodPost, path: "/", expectStatus: http.StatusMethodNotAllowed},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != tc.expectStatus {
				t.Fatalf("unexpected status: got %d want %d", rec.Code, tc.expectStatus)
			}

			if tc.expectBody != "" && !strings.Contains(rec.Body.String(), tc.expectBody) {
				t.Fatalf("unexpected body: %q", rec.Body.String())
			}

			if tc.name == "asset" {
				head := rec.Header().Get("Cache-Control")
				if head != "public, max-age=31536000" {
					t.Fatalf("cache control not applied: %q", head)
				}
			}
		})
	}
}

func TestRegisterWithPrefix(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("app shell"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir assets: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets/logo.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithPrefix("/app")); err != nil {
		t.Fatalf("register: %v", err)
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/app/assets/logo.png", nil)
	assetRec := httptest.NewRecorder()
	r.ServeHTTP(assetRec, assetReq)
	if assetRec.Code != http.StatusOK {
		t.Fatalf("asset status: %d", assetRec.Code)
	}

	rootReq := httptest.NewRequest(http.MethodGet, "/app", nil)
	rootRec := httptest.NewRecorder()
	r.ServeHTTP(rootRec, rootReq)
	if rootRec.Code != http.StatusOK {
		t.Fatalf("root status: %d", rootRec.Code)
	}
	if !strings.Contains(rootRec.Body.String(), "app shell") {
		t.Fatalf("root body mismatch: %q", rootRec.Body.String())
	}
}

func TestRegisterFromDirMissing(t *testing.T) {
	r := router.NewRouter()
	if err := RegisterFromDir(r, "./does/not/exist"); err == nil {
		t.Fatalf("expected error for missing directory")
	}
}

func TestRegisterEmbedded(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("embedded home"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "bundle.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write bundle: %v", err)
	}

	originalFS, originalRoot := embeddedFS, embeddedRoot
	embeddedFS, embeddedRoot = os.DirFS(dir), "."
	defer func() {
		embeddedFS, embeddedRoot = originalFS, originalRoot
	}()

	r := router.NewRouter()
	if err := RegisterEmbedded(r); err != nil {
		t.Fatalf("register embedded: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/bundle.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "console.log('ok')") {
		t.Fatalf("body mismatch: %q", rec.Body.String())
	}
}

func TestRegisterEmbeddedMissing(t *testing.T) {
	r := router.NewRouter()
	if err := RegisterEmbedded(r); err == nil {
		t.Fatalf("expected error when no embedded assets are present")
	}
}
