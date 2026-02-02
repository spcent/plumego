package watcher

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Watcher watches for file changes
type Watcher struct {
	dir      string
	include  []string
	exclude  []string
	debounce time.Duration

	events chan string
	errors chan error
	done   chan bool

	lastChange time.Time
	pending    string
}

// NewWatcher creates a new file watcher
func NewWatcher(dir string, include, exclude []string, debounce time.Duration) (*Watcher, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	w := &Watcher{
		dir:      absDir,
		include:  include,
		exclude:  exclude,
		debounce: debounce,
		events:   make(chan string, 10),
		errors:   make(chan error, 10),
		done:     make(chan bool),
	}

	go w.watch()

	return w, nil
}

// Events returns the events channel
func (w *Watcher) Events() <-chan string {
	return w.events
}

// Errors returns the errors channel
func (w *Watcher) Errors() <-chan error {
	return w.errors
}

// Close stops the watcher
func (w *Watcher) Close() error {
	close(w.done)
	return nil
}

func (w *Watcher) watch() {
	// Keep track of file modification times
	fileModTimes := make(map[string]time.Time)

	// Initial scan
	w.scanFiles(fileModTimes)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.scanFiles(fileModTimes)

			// Check if we should emit a pending event
			if w.pending != "" && time.Since(w.lastChange) > w.debounce {
				w.events <- w.pending
				w.pending = ""
			}

		case <-w.done:
			return
		}
	}
}

func (w *Watcher) scanFiles(modTimes map[string]time.Time) {
	filepath.Walk(w.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories
		if info.IsDir() {
			// Check if we should skip this directory
			relPath, _ := filepath.Rel(w.dir, path)
			if w.shouldExclude(relPath) {
				return filepath.SkipDir
			}
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(w.dir, path)
		if err != nil {
			return nil
		}

		// Check if file matches patterns
		if !w.shouldWatch(relPath) {
			return nil
		}

		// Check modification time
		modTime := info.ModTime()
		lastModTime, exists := modTimes[path]

		if !exists {
			// New file
			modTimes[path] = modTime
		} else if modTime.After(lastModTime) {
			// File modified
			modTimes[path] = modTime
			w.lastChange = time.Now()
			w.pending = relPath
		}

		return nil
	})
}

func (w *Watcher) shouldWatch(path string) bool {
	// Check includes
	if len(w.include) > 0 {
		matched := false
		for _, pattern := range w.include {
			if matchPattern(pattern, path) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	// Check excludes
	if w.shouldExclude(path) {
		return false
	}

	return true
}

func (w *Watcher) shouldExclude(path string) bool {
	for _, pattern := range w.exclude {
		if matchPattern(pattern, path) {
			return true
		}
	}
	return false
}

func matchPattern(pattern, path string) bool {
	// Simple pattern matching
	// Supports:
	//   - **/*.go (recursive)
	//   - *.go (single level)
	//   - dir/file.go (exact match)

	// Handle ** (recursive match)
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix == "" {
				return true
			}

			// Match suffix
			if strings.HasPrefix(suffix, "*.") {
				ext := strings.TrimPrefix(suffix, "*")
				return strings.HasSuffix(path, ext)
			}

			return strings.HasSuffix(path, suffix)
		}
	}

	// Handle * (single level)
	if strings.Contains(pattern, "*") {
		if strings.HasPrefix(pattern, "*.") {
			ext := strings.TrimPrefix(pattern, "*")
			return strings.HasSuffix(path, ext)
		}

		// Simple glob matching
		return simpleGlob(pattern, path)
	}

	// Exact match
	return pattern == path
}

func simpleGlob(pattern, str string) bool {
	// Very simple glob implementation
	if pattern == "*" {
		return true
	}

	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == str
	}

	// Check prefix
	if parts[0] != "" && !strings.HasPrefix(str, parts[0]) {
		return false
	}

	// Check suffix
	if parts[len(parts)-1] != "" && !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}

	// Check middle parts
	currentPos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		idx := strings.Index(str[currentPos:], part)
		if idx == -1 {
			return false
		}
		currentPos += idx + len(part)
	}

	return true
}
