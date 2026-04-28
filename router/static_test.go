package router

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/spcent/plumego/contract"
)

// helper: create a temporary file with given content
func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	fullPath := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return fullPath
}

func TestStatic(t *testing.T) {
	tmpDir := t.TempDir()
	// create sample static files
	createTempFile(t, tmpDir, "js/app.js", "console.log('hello');")
	createTempFile(t, tmpDir, "css/style.css", "body{color:red;}")
	createTempFile(t, tmpDir, "safe..name.txt", "safe")

	r := NewRouter()
	if err := r.Static("/static", tmpDir); err != nil {
		t.Fatalf("Static returned error: %v", err)
	}
	r.Freeze()

	// test existing file
	w := serveRouter(r, http.MethodGet, "/static/js/app.js")
	assertResponseStatus(t, w, http.StatusOK)
	assertResponseBody(t, w, "console.log('hello');")

	w = serveRouter(r, http.MethodGet, "/static/safe..name.txt")
	assertResponseStatus(t, w, http.StatusOK)
	assertResponseBody(t, w, "safe")

	// test non-existing file
	w = serveRouter(r, http.MethodGet, "/static/img/logo.png")
	assertResponseStatus(t, w, http.StatusNotFound)
}

func TestStaticNormalizesTrailingSlashPrefix(t *testing.T) {
	tmpDir := t.TempDir()
	createTempFile(t, tmpDir, "js/app.js", "console.log('hello');")

	r := NewRouter()
	if err := r.Static("/static/", tmpDir); err != nil {
		t.Fatalf("Static returned error: %v", err)
	}
	r.Freeze()

	w := serveRouter(r, http.MethodGet, "/static/js/app.js")
	assertResponseStatus(t, w, http.StatusOK)
	assertResponseBody(t, w, "console.log('hello');")

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Path != "/static/*filepath" {
		t.Fatalf("expected normalized static route path %q, got %q", "/static/*filepath", routes[0].Path)
	}
}

func TestStaticRootPrefixRegistersSingleSlashWildcard(t *testing.T) {
	tmpDir := t.TempDir()
	createTempFile(t, tmpDir, "app.js", "console.log('root');")

	r := NewRouter()
	if err := r.Static("/", tmpDir); err != nil {
		t.Fatalf("Static returned error: %v", err)
	}

	routes := r.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	if routes[0].Path != "/*filepath" {
		t.Fatalf("root static route path = %q, want %q", routes[0].Path, "/*filepath")
	}

	w := serveRouter(r, http.MethodGet, "/app.js")
	assertResponseStatus(t, w, http.StatusOK)
	assertResponseBody(t, w, "console.log('root');")
}

func TestStaticRejectsSymlinkEscape(t *testing.T) {
	tmpDir := t.TempDir()
	outsideDir := t.TempDir()
	outsideFile := createTempFile(t, outsideDir, "secret.txt", "secret")

	if err := os.Symlink(outsideFile, filepath.Join(tmpDir, "link.txt")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}

	r := NewRouter()
	if err := r.Static("/static", tmpDir); err != nil {
		t.Fatalf("Static returned error: %v", err)
	}
	r.Freeze()

	w := serveRouter(r, http.MethodGet, "/static/link.txt")
	assertResponseStatus(t, w, http.StatusNotFound)
}

func TestStaticFS(t *testing.T) {
	r := NewRouter()
	if err := r.StaticFS("/assets", http.FS(fstest.MapFS{
		"app.js": &fstest.MapFile{
			Data: []byte("console.log('embedded');"),
			Mode: fs.ModePerm,
		},
	})); err != nil {
		t.Fatalf("StaticFS returned error: %v", err)
	}
	r.Freeze()

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assertResponseStatus(t, w, http.StatusOK)
	assertResponseBody(t, w, "console.log('embedded');")
}

func TestStaticFSRejectsNilFileSystem(t *testing.T) {
	r := NewRouter()

	err := r.StaticFS("/assets", nil)
	if err == nil {
		t.Fatalf("expected nil filesystem registration to fail")
	}
}

func TestGetFilePathFromRequestRejectsUnsafePaths(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantOK  bool
		wantVal string
	}{
		{name: "valid nested path", path: "css/site.css", wantOK: true, wantVal: "css/site.css"},
		{name: "empty path", path: "", wantOK: false},
		{name: "null byte", path: "css/site.css\x00", wantOK: false},
		{name: "traversal", path: "../secret.txt", wantOK: false},
		{name: "absolute path", path: "/etc/passwd", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			ctx := contract.WithRequestContext(req.Context(), contract.RequestContext{
				Params: map[string]string{"filepath": tt.path},
			})
			req = req.WithContext(ctx)

			got, ok := getFilePathFromRequest(req)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantVal {
				t.Fatalf("path = %q, want %q", got, tt.wantVal)
			}
		})
	}
}
