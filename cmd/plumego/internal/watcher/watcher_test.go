package watcher

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShouldWatchIncludeAndExcludePatterns(t *testing.T) {
	w := &Watcher{
		include: []string{"**/*.go"},
		exclude: []string{"**/vendor/**", "**/*_test.go"},
	}

	tests := []struct {
		path string
		want bool
	}{
		{path: "main.go", want: true},
		{path: "internal/app/routes.go", want: true},
		{path: "internal/app/routes_test.go", want: false},
		{path: "vendor/example/main.go", want: false},
		{path: "README.md", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := w.shouldWatch(tt.path); got != tt.want {
				t.Fatalf("shouldWatch(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestWatcherEmitsDebouncedModifyEvent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "main.go")
	if err := os.WriteFile(path, []byte("package main\n"), 0644); err != nil {
		t.Fatalf("write initial file: %v", err)
	}

	w, err := NewWatcher(tmp, []string{"**/*.go"}, nil, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("new watcher: %v", err)
	}
	defer w.Close()

	time.Sleep(50 * time.Millisecond)
	if err := os.WriteFile(path, []byte("package main\n\nfunc main() {}\n"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	select {
	case got := <-w.Events():
		if got != "main.go" {
			t.Fatalf("event = %q, want main.go", got)
		}
	case err := <-w.Errors():
		t.Fatalf("watcher error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for watcher event")
	}
}
