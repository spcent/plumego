package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spcent/plumego/router"
)

// writeTestFile creates a file at dir/relPath with the given content,
// creating intermediate directories as needed.
func writeTestFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func TestRegisterFromDir(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>home</html>")
	writeTestFile(t, dir, "_next/static/app.js", "console.log('hello')")

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
	writeTestFile(t, dir, "index.html", "app shell")
	writeTestFile(t, dir, "assets/logo.png", "png")

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

func TestNewMountFSRegister(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "mount index")
	writeTestFile(t, dir, "assets/app.js", "mount asset")

	mount, err := NewMountFS(http.Dir(dir), WithPrefix("/app"))
	if err != nil {
		t.Fatalf("new mount: %v", err)
	}
	if got := mount.Prefix(); got != "/app" {
		t.Fatalf("prefix: got %q want %q", got, "/app")
	}
	if mount.Handler() == nil {
		t.Fatal("expected mount handler")
	}

	r := router.NewRouter()
	if err := mount.Register(r); err != nil {
		t.Fatalf("register mount: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app/assets/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("asset status: got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "mount asset") {
		t.Fatalf("unexpected asset body: %q", rec.Body.String())
	}
}

func TestNewHandlerFS(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "handler index")

	h, err := NewHandlerFS(http.Dir(dir))
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "handler index") {
		t.Fatalf("unexpected body: %q", rec.Body.String())
	}
}

func TestMountRegisterNilRouter(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")

	mount, err := NewMountFS(http.Dir(dir))
	if err != nil {
		t.Fatalf("new mount: %v", err)
	}
	if err := mount.Register(nil); err == nil {
		t.Fatal("expected error for nil router")
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

// TestRegisterFS_HTTPFileSystem verifies that RegisterFS works with http.Dir.
func TestRegisterFS_HTTPFileSystem(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "embedded home")
	writeTestFile(t, dir, "bundle.js", "console.log('ok')")

	r := router.NewRouter()
	if err := RegisterFS(r, http.Dir(dir)); err != nil {
		t.Fatalf("register: %v", err)
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

// TestRegisterEmbeddedMissing verifies that RegisterEmbedded returns an error
// when the embedded/ directory contains no real assets (only .keep).
func TestRegisterEmbeddedMissing(t *testing.T) {
	r := router.NewRouter()
	if err := RegisterEmbedded(r); err == nil {
		t.Fatalf("expected error when no embedded assets are present")
	} else if !strings.Contains(err.Error(), "no embedded frontend assets") {
		t.Logf("error message: %v", err)
	}
}

// TestRegisterFS_NestedDirectories verifies that files in nested subdirectories
// are served correctly when using RegisterFS.
func TestRegisterFS_NestedDirectories(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")
	writeTestFile(t, dir, "assets/style.css", "css")

	r := router.NewRouter()
	if err := RegisterFS(r, http.Dir(dir)); err != nil {
		t.Fatalf("register: %v", err)
	}

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

// Test security and edge cases
func TestSecurityPathTraversal(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")
	writeTestFile(t, dir, "secret/secret.txt", "secret data")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	traversalPaths := []string{
		"/../secret/secret.txt",
		"/../../secret.txt",
		"/../../../etc/passwd",
		"/%2e%2e%2fsecret%2fsecret.txt",
		"/%2e%2e/",
	}

	for _, path := range traversalPaths {
		t.Run("traversal_"+path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK && strings.Contains(rec.Body.String(), "secret data") {
				t.Fatalf("path traversal succeeded: %s", path)
			}
		})
	}
}

func TestSpecialCharactersInFilenames(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"index.html":            "index",
		"file with spaces.html": "spaces",
		"file-with-dash.html":   "dash",
		"file_with_underscore":  "underscore",
		"file.multiple.dots.js": "dots",
	}

	for filename, content := range files {
		writeTestFile(t, dir, filename, content)
	}

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	for filename, content := range files {
		t.Run(filename, func(t *testing.T) {
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
	writeTestFile(t, dir, "assets/js/app.js", "app code")
	writeTestFile(t, dir, "index.html", "index")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/assets/js/app.js", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("nested path failed: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "app code") {
		t.Fatalf("nested path content wrong: %q", rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/assets/js/", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("directory access failed: %d", rec.Code)
	}
}

func TestInvalidPrefixes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")

	invalidPrefixes := []string{
		"/app/../",
		"/app/./",
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

	validPrefixes := []string{
		"",
		"app",
		"app/",
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
	writeTestFile(t, dir, "index.html", "index")
	writeTestFile(t, dir, "asset.js", "asset")

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
		{"missing", "/missing.html", false, ""},
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
	writeTestFile(t, dir, "index.html", "index")
	writeTestFile(t, dir, "asset.js", "asset")

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
	writeTestFile(t, dir, "index.html", "index")

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
	writeTestFile(t, dir, "index.html", "index")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithHeaders(map[string]string{"X-Frontend": "ok"})); err != nil {
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
	writeTestFile(t, dir, "index.html", "index")

	r := router.NewRouter()

	invalidIndexes := []string{
		"../index.html",
		"index/../index",
		"",
	}

	for _, index := range invalidIndexes {
		t.Run("index_"+index, func(t *testing.T) {
			err := RegisterFromDir(r, dir, WithIndex(index))
			if index == "" {
				if err != nil {
					t.Fatalf("empty index should work: %v", err)
				}
				return
			}
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
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on Windows due to permission model differences")
	}
	if os.Getuid() == 0 {
		t.Skip("skipping test when running as root")
	}

	dir := t.TempDir()
	subDir := filepath.Join(dir, "unreadable")
	if err := os.MkdirAll(subDir, 0o000); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	defer os.Chmod(subDir, 0o755)

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

func TestMethodNotAllowed(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")

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
	writeTestFile(t, dir, "index.html", "index")
	writeTestFile(t, dir, "asset.js", "asset")

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
			if rec.Body.Len() > 0 {
				t.Fatalf("HEAD %s: unexpected body length %d", path, rec.Body.Len())
			}
		})
	}
}

func TestPrefixNormalization(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "index")

	tests := []struct {
		input     string
		shouldErr bool
	}{
		{"/", false},
		{"/app", false},
		{"/app/", false},
		{"app", false},
		{"app/", false},
		{"/app/../", true},
		{"/app/./", true},
		{"", false},
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

func TestPrecompressedFiles(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>index</html>")
	writeTestFile(t, dir, "app.js", "console.log('original')")
	writeTestFile(t, dir, "app.js.gz", "gzipped content")
	writeTestFile(t, dir, "app.js.br", "brotli content")
	writeTestFile(t, dir, "style.css", "body { color: red; }")
	writeTestFile(t, dir, "style.css.gz", "gzipped css")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithPrecompressed(true)); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		name           string
		path           string
		acceptEncoding string
		expectBody     string
		expectEncoding string
		expectVary     bool
	}{
		{
			name:           "brotli preferred",
			path:           "/app.js",
			acceptEncoding: "br, gzip, deflate",
			expectBody:     "brotli content",
			expectEncoding: "br",
			expectVary:     true,
		},
		{
			name:           "gzip when br not accepted",
			path:           "/app.js",
			acceptEncoding: "gzip, deflate",
			expectBody:     "gzipped content",
			expectEncoding: "gzip",
			expectVary:     true,
		},
		{
			name:           "original when no encoding accepted",
			path:           "/app.js",
			acceptEncoding: "",
			expectBody:     "console.log('original')",
			expectEncoding: "",
			expectVary:     false,
		},
		{
			name:           "only gzip available",
			path:           "/style.css",
			acceptEncoding: "br, gzip",
			expectBody:     "gzipped css",
			expectEncoding: "gzip",
			expectVary:     true,
		},
		{
			name:           "no precompressed version",
			path:           "/index.html",
			acceptEncoding: "gzip, br",
			expectBody:     "<html>index</html>",
			expectEncoding: "",
			expectVary:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tc.acceptEncoding)
			}
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status: got %d want %d", rec.Code, http.StatusOK)
			}

			if !strings.Contains(rec.Body.String(), tc.expectBody) {
				t.Fatalf("body: got %q, expected to contain %q", rec.Body.String(), tc.expectBody)
			}

			gotEncoding := rec.Header().Get("Content-Encoding")
			if gotEncoding != tc.expectEncoding {
				t.Fatalf("Content-Encoding: got %q want %q", gotEncoding, tc.expectEncoding)
			}

			hasVary := rec.Header().Get("Vary") != ""
			if hasVary != tc.expectVary {
				t.Fatalf("Vary header: got %v want %v", hasVary, tc.expectVary)
			}
		})
	}
}

func TestPrecompressedDisabled(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>index</html>")
	writeTestFile(t, dir, "app.js", "original")
	writeTestFile(t, dir, "app.js.gz", "gzipped")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithPrecompressed(false)); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	if rec.Body.String() != "original" {
		t.Fatalf("should serve original file when precompressed disabled, got: %q", rec.Body.String())
	}
	if rec.Header().Get("Content-Encoding") != "" {
		t.Fatalf("should not set Content-Encoding when precompressed disabled")
	}
}

func TestCustomNotFoundPage(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>home</html>")
	writeTestFile(t, dir, "404.html", "<html>Custom 404 Page</html>")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir,
		WithNotFoundPage("404.html"),
		WithFallback(false),
	); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if !strings.Contains(rec.Body.String(), "Custom 404 Page") {
		t.Fatalf("should serve custom 404 page, got: %q", rec.Body.String())
	}
}

// errOnPathFS wraps http.FileSystem and returns os.ErrPermission for a specific
// path, simulating a server-side IO error (not a "file not found" error).
type errOnPathFS struct {
	base    http.FileSystem
	errPath string
}

func (f *errOnPathFS) Open(name string) (http.File, error) {
	if name == f.errPath {
		return nil, os.ErrPermission
	}
	return f.base.Open(name)
}

// TestCustomErrorPage verifies that when a server IO error occurs and an error
// page is configured, the error page is served with the correct 500 status code.
func TestCustomErrorPage(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>home</html>")
	writeTestFile(t, dir, "500.html", "<html>Server Error</html>")

	fs := &errOnPathFS{base: http.Dir(dir), errPath: "bad-file"}

	r := router.NewRouter()
	if err := RegisterFS(r, fs, WithErrorPage("500.html"), WithFallback(false)); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/bad-file", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 status, got: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Server Error") {
		t.Fatalf("expected error page content, got: %q", rec.Body.String())
	}
}

// TestCustomErrorPageFallback verifies that when the custom error page itself
// is missing, serveError falls back to a JSON error response.
func TestCustomErrorPageFallback(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>home</html>")
	// 500.html intentionally not created.

	fs := &errOnPathFS{base: http.Dir(dir), errPath: "bad-file"}

	r := router.NewRouter()
	if err := RegisterFS(r, fs, WithErrorPage("500.html"), WithFallback(false)); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/bad-file", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 status, got: %d", rec.Code)
	}
	// Fallback is JSON; it must not contain HTML from the (missing) error page.
	body := rec.Body.String()
	if strings.Contains(body, "<html>") {
		t.Fatalf("should not serve missing error page HTML, got: %q", body)
	}
	if !strings.Contains(body, `"error"`) {
		t.Fatalf("expected JSON error response, got: %q", body)
	}
}

// TestCustomNotFoundPageStatus verifies that a custom 404 page is served with
// the correct 404 status code rather than 200 OK.
func TestCustomNotFoundPageStatus(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>home</html>")
	writeTestFile(t, dir, "404.html", "<html>Custom 404 Page</html>")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir,
		WithNotFoundPage("404.html"),
		WithFallback(false),
	); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 status, got: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Custom 404 Page") {
		t.Fatalf("expected custom 404 page content, got: %q", rec.Body.String())
	}
}

func TestCustomMIMETypes(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>index</html>")
	writeTestFile(t, dir, "app.wasm", "wasm binary content")
	writeTestFile(t, dir, "manifest.json", `{"name":"app"}`)

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir, WithMIMETypes(map[string]string{
		".wasm": "application/wasm",
		".json": "application/json; charset=utf-8",
	})); err != nil {
		t.Fatalf("register: %v", err)
	}

	tests := []struct {
		path       string
		expectType string
	}{
		{"/app.wasm", "application/wasm"},
		{"/manifest.json", "application/json; charset=utf-8"},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status: %d", rec.Code)
			}

			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, tc.expectType) {
				t.Fatalf("Content-Type: got %q, expected to contain %q", contentType, tc.expectType)
			}
		})
	}
}

func TestMIMETypesWithPrecompressed(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, dir, "index.html", "<html>index</html>")
	writeTestFile(t, dir, "app.wasm", "wasm original")
	writeTestFile(t, dir, "app.wasm.br", "wasm compressed")

	r := router.NewRouter()
	if err := RegisterFromDir(r, dir,
		WithPrecompressed(true),
		WithMIMETypes(map[string]string{
			".wasm": "application/wasm",
		}),
	); err != nil {
		t.Fatalf("register: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app.wasm", nil)
	req.Header.Set("Accept-Encoding", "br")
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "wasm compressed") {
		t.Fatalf("should serve compressed version")
	}
	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/wasm") {
		t.Fatalf("Content-Type should be set for original file type, got: %q", contentType)
	}
	if rec.Header().Get("Content-Encoding") != "br" {
		t.Fatalf("should set Content-Encoding")
	}
}

// TestAcceptsToken verifies the token-level Accept-Encoding parser.
func TestAcceptsToken(t *testing.T) {
	tests := []struct {
		header string
		token  string
		want   bool
	}{
		{"br, gzip, deflate", "br", true},
		{"gzip, deflate", "br", false},
		{"gzip, deflate", "gzip", true},
		{"GZIP", "gzip", true},               // case-insensitive
		{"gzip;q=0.9, br;q=1.0", "br", true}, // quality factors ignored
		{"", "gzip", false},                  // empty header
		{"brotli", "br", false},              // no substring matching
	}

	for _, tt := range tests {
		t.Run(tt.header+"/"+tt.token, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.header != "" {
				req.Header.Set("Accept-Encoding", tt.header)
			}
			got := acceptsToken(req, tt.token)
			if got != tt.want {
				t.Fatalf("acceptsToken(%q, %q) = %v, want %v", tt.header, tt.token, got, tt.want)
			}
		})
	}
}
