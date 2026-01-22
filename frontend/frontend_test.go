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
				t.Fatalf("unexpected body: %q, expected to contain: %q", rec.Body.String(), tc.expectBody)
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
	} else if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "no such file") {
		t.Logf("error message: %v", err)
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
	} else if !strings.Contains(err.Error(), "no embedded frontend assets") {
		t.Logf("error message: %v", err)
	}
}

// Test security and edge cases
func TestSecurityPathTraversal(t *testing.T) {
	dir := t.TempDir()

	// Create index.html
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	// Create a secret file outside the intended directory structure
	secretDir := filepath.Join(dir, "secret")
	if err := os.MkdirAll(secretDir, 0o755); err != nil {
		t.Fatalf("mkdir secret: %v", err)
	}
	if err := os.WriteFile(filepath.Join(secretDir, "secret.txt"), []byte("secret data"), 0o644); err != nil {
		t.Fatalf("write secret: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Try various path traversal attempts
	traversalPaths := []string{
		"/../secret/secret.txt",
		"/../../secret.txt",
		"/../../../etc/passwd",
		"/%2e%2e%2fsecret%2fsecret.txt", // URL encoded
		"/%2e%2e/",                      // URL encoded dot-dot
	}

	for _, path := range traversalPaths {
		t.Run("traversal_"+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			// Should either return 404 or index.html, not the secret file
			if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "secret data") {
				t.Fatalf("path traversal succeeded: %s", path)
			}
		})
	}
}

func TestSpecialCharactersInFilenames(t *testing.T) {
	dir := t.TempDir()

	// Create files with special characters
	files := map[string]string{
		"index.html":            "index",
		"file with spaces.html": "spaces",
		"file-with-dash.html":   "dash",
		"file_with_underscore":  "underscore",
		"file.multiple.dots.js": "dots",
	}

	for filename, content := range files {
		if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", filename, err)
		}
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	for filename, content := range files {
		t.Run(filename, func(t *testing.T) {
			// URL encode the filename for spaces
			path := "/" + filename
			if strings.Contains(filename, " ") {
				path = "/file%20with%20spaces.html"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("failed to serve %s: %d", filename, rec.Code)
			}
			if !strings.Contains(rec.Body.String(), content) {
				t.Fatalf("wrong content for %s: got %q, want %q", filename, rec.Body.String(), content)
			}
		})
	}
}

func TestDirectoryTraversalWithSubdirs(t *testing.T) {
	dir := t.TempDir()

	// Create nested directory structure
	subDir := filepath.Join(dir, "assets", "js")
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(subDir, "app.js"), []byte("app code"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	// Test valid nested path
	req := httptest.NewRequest(http.MethodGet, "/assets/js/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("nested path failed: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "app code") {
		t.Fatalf("nested path content wrong: %q", rec.Body.String())
	}

	// Test directory access (should serve index)
	req = httptest.NewRequest(http.MethodGet, "/assets/js/", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("directory access failed: %d", rec.Code)
	}
}

func TestInvalidPrefixes(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	// Test invalid prefix formats
	invalidPrefixes := []string{
		"/app/../", // Contains path traversal
		"/app/./",  // Contains current directory
	}

	for _, prefix := range invalidPrefixes {
		t.Run("prefix_"+prefix, func(t *testing.T) {
			r := router.NewRouter()
			err := RegisterFromDir(r, dir, WithPrefix(prefix))
			if err == nil {
				t.Fatalf("expected error for prefix %q, got nil", prefix)
			}
		})
	}

	// Test valid prefixes that should work
	validPrefixes := []string{
		"",     // Empty (should default to /)
		"app",  // Missing leading slash (should be fixed)
		"app/", // Missing leading slash with trailing (should be fixed)
	}

	for _, prefix := range validPrefixes {
		t.Run("valid_prefix_"+prefix, func(t *testing.T) {
			r := router.NewRouter()
			err := RegisterFromDir(r, dir, WithPrefix(prefix))
			if err != nil {
				t.Fatalf("prefix %q should work: %v", prefix, err)
			}
		})
	}
}

func TestCacheControlBehavior(t *testing.T) {
	dir := t.TempDir()

	// Create files
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.js"), []byte("asset"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithCacheControl("max-age=3600")); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		expectCache bool
		cacheValue  string
	}{
		{"index", "/", false, ""},
		{"asset", "/asset.js", true, "max-age=3600"},
		{"missing", "/missing.html", false, ""}, // Falls back to index
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			cacheHeader := rec.Header().Get("Cache-Control")
			if tt.expectCache {
				if cacheHeader != tt.cacheValue {
					t.Fatalf("cache header: got %q, want %q", cacheHeader, tt.cacheValue)
				}
			} else {
				if cacheHeader != "" {
					t.Fatalf("unexpected cache header: %q", cacheHeader)
				}
			}
		})
	}
}

func TestIndexCacheControl(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.js"), []byte("asset"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(
		r,
		dir,
		WithCacheControl("public, max-age=3600"),
		WithIndexCacheControl("no-cache"),
	); err != nil {
		t.Fatalf("register: %v", err)
	}

	indexReq := httptest.NewRequest(http.MethodGet, "/", nil)
	indexRec := httptest.NewRecorder()
	r.ServeHTTP(indexRec, indexReq)
	if got := indexRec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Fatalf("index cache header: got %q, want %q", got, "no-cache")
	}

	assetReq := httptest.NewRequest(http.MethodGet, "/asset.js", nil)
	assetRec := httptest.NewRecorder()
	r.ServeHTTP(assetRec, assetReq)
	if got := assetRec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Fatalf("asset cache header: got %q, want %q", got, "public, max-age=3600")
	}
}

func TestFallbackDisabled(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithFallback(false)); err != nil {
		t.Fatalf("register: %v", err)
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/missing", nil)
	missingRec := httptest.NewRecorder()
	r.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusNotFound {
		t.Fatalf("missing status: got %d want %d", missingRec.Code, http.StatusNotFound)
	}

	rootReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rootRec := httptest.NewRecorder()
	r.ServeHTTP(rootRec, rootReq)
	if rootRec.Code != http.StatusOK {
		t.Fatalf("root status: got %d want %d", rootRec.Code, http.StatusOK)
	}
}

func TestCustomHeaders(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	headers := map[string]string{
		"X-Frontend": "ok",
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithHeaders(headers)); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if got := rec.Header().Get("X-Frontend"); got != "ok" {
		t.Fatalf("header mismatch: got %q", got)
	}
}

func TestIndexFileValidation(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	r := router.NewRouter()

	// Test invalid index files
	invalidIndexes := []string{
		"../index.html",  // Contains path separator
		"index/../index", // Contains path separator
		"",               // Empty (should default)
	}

	for _, index := range invalidIndexes {
		t.Run("index_"+index, func(t *testing.T) {
			err := RegisterFromDir(r, dir, WithIndex(index))
			if index == "" {
				// Empty should work (defaults to index.html)
				if err != nil {
					t.Fatalf("empty index should work: %v", err)
				}
				return
			}
			// Invalid indexes should fail
			if err == nil {
				t.Fatalf("expected error for index %q, got nil", index)
			}
		})
	}
}

func TestNilFilesystem(t *testing.T) {
	r := router.NewRouter()
	err := RegisterFS(r, nil)
	if err == nil {
		t.Fatal("expected error for nil filesystem")
	}
	if !strings.Contains(err.Error(), "filesystem cannot be nil") {
		t.Fatalf("wrong error message: %v", err)
	}
}

func TestUnreadableDirectory(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "unreadable")
	if err := os.MkdirAll(subDir, 0o000); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.Chmod(subDir, 0o755) // Cleanup

	r := router.NewRouter()
	err := RegisterFromDir(r, subDir)
	if err == nil {
		t.Fatal("expected error for unreadable directory")
	}
	if !strings.Contains(err.Error(), "not readable") {
		t.Fatalf("wrong error message: %v", err)
	}
}

func TestNonDirectoryPath(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notadir")
	if err := os.WriteFile(filePath, []byte("file"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	r := router.NewRouter()
	err := RegisterFromDir(r, filePath)
	if err == nil {
		t.Fatal("expected error for non-directory path")
	}
	if !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("wrong error message: %v", err)
	}
}

func TestEmptyEmbeddedDirectory(t *testing.T) {
	dir := t.TempDir()
	// Create only .keep file
	if err := os.MkdirAll(filepath.Join(dir, "embedded"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "embedded", ".keep"), []byte(""), 0o644); err != nil {
		t.Fatalf("write .keep: %v", err)
	}

	originalFS, originalRoot := embeddedFS, embeddedRoot
	embeddedFS, embeddedRoot = os.DirFS(dir), "embedded"
	defer func() {
		embeddedFS, embeddedRoot = originalFS, originalRoot
	}()

	r := router.NewRouter()
	err := RegisterEmbedded(r)
	if err == nil {
		t.Fatal("expected error for empty embedded directory")
	}
	if !strings.Contains(err.Error(), "no embedded frontend assets") {
		t.Fatalf("wrong error message: %v", err)
	}
}

func TestNestedDirectoriesInEmbedded(t *testing.T) {
	dir := t.TempDir()
	// Create nested structure
	if err := os.MkdirAll(filepath.Join(dir, "embedded", "assets"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "embedded", "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "embedded", "assets", "style.css"), []byte("css"), 0o644); err != nil {
		t.Fatalf("write css: %v", err)
	}

	originalFS, originalRoot := embeddedFS, embeddedRoot
	embeddedFS, embeddedRoot = os.DirFS(dir), "embedded"
	defer func() {
		embeddedFS, embeddedRoot = originalFS, originalRoot
	}()

	r := router.NewRouter()
	if err := RegisterEmbedded(r); err != nil {
		t.Fatalf("register embedded: %v", err)
	}

	// Test nested asset
	req := httptest.NewRequest(http.MethodGet, "/assets/style.css", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("nested asset failed: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "css") {
		t.Fatalf("nested asset content wrong: %q", rec.Body.String())
	}
}

func TestMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("method %s: got status %d, want %d", method, rec.Code, http.StatusMethodNotAllowed)
			}
		})
	}
}

func TestHeadMethod(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.js"), []byte("asset"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []string{"/", "/asset.js"}
	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodHead, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("HEAD %s: got status %d", path, rec.Code)
			}
			// HEAD should not have body
			if rec.Body.Len() > 0 {
				t.Fatalf("HEAD %s: unexpected body length %d", path, rec.Body.Len())
			}
		})
	}
}

func TestPrefixNormalization(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}

	tests := []struct {
		input     string
		expected  string
		shouldErr bool
	}{
		{"/", "/", false},
		{"/app", "/app", false},
		{"/app/", "/app", false},
		{"app", "/app", false},
		{"app/", "/app", false},
		{"/app/../", "", true}, // Path traversal
		{"/app/./", "", true},  // Current directory
		{"", "/", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r := router.NewRouter()
			err := RegisterFromDir(r, dir, WithPrefix(tt.input))

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("expected error for prefix %q, got nil", tt.input)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error for prefix %q: %v", tt.input, err)
				}
			}
		})
	}
}
