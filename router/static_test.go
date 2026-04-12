package router

import (
	"io"
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

	r := NewRouter()
	if err := r.Static("/static", tmpDir); err != nil {
		t.Fatalf("Static returned error: %v", err)
	}
	r.Freeze()

	// test existing file
	req := httptest.NewRequest("GET", "/static/js/app.js", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if string(body) != "console.log('hello');" {
		t.Errorf("unexpected body: %s", body)
	}

	// test non-existing file
	req2 := httptest.NewRequest("GET", "/static/img/logo.png", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Result().StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 for missing file, got %d", w2.Result().StatusCode)
	}
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

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if string(body) != "console.log('embedded');" {
		t.Fatalf("unexpected body: %s", body)
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
