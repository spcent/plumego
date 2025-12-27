package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
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
	r.Static("/static", tmpDir)
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
