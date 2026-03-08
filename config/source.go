package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// Source represents a configuration source that can load and watch for changes.
type Source interface {
	// Load loads configuration data from the source.
	Load(ctx context.Context) (map[string]any, error)
	// Watch returns a channel that sends configuration updates or errors.
	// The channel is closed when the context is cancelled or watching stops.
	Watch(ctx context.Context) <-chan WatchResult
	// Name returns a human-readable name for the source.
	Name() string
}

// WatchResult represents a configuration update or error from watching a source.
type WatchResult struct {
	// Data contains the updated configuration, nil if there was an error.
	Data map[string]any
	// Err contains any error that occurred during watching.
	Err error
}

// File format constants.
const (
	FormatJSON = "json"
	FormatTOML = "toml"
	FormatEnv  = "env"
)

// EnvSource implements Source for environment variables.
type EnvSource struct {
	prefix string
}

// NewEnvSource creates a new environment variable source.
// If prefix is non-empty, only environment variables with that prefix are loaded,
// and the prefix is stripped from the key names.
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{prefix: prefix}
}

// Load loads configuration from environment variables.
func (e *EnvSource) Load(ctx context.Context) (map[string]any, error) {
	data := make(map[string]any)

	for _, env := range os.Environ() {
		key, value, ok := parseEnvLine(env)
		if !ok {
			continue
		}

		if e.prefix != "" {
			trimmed, found := strings.CutPrefix(key, e.prefix)
			if !found {
				continue
			}
			key = trimmed
		}

		data[toSnakeCase(key)] = value
	}

	return data, nil
}

// Watch watches for environment variable changes.
// Environment variables don't typically change during runtime,
// so the returned channel is immediately closed.
func (e *EnvSource) Watch(ctx context.Context) <-chan WatchResult {
	results := make(chan WatchResult)
	close(results)
	return results
}

// Name returns the source name.
func (e *EnvSource) Name() string {
	return "env"
}

// FileSource implements Source for configuration files.
type FileSource struct {
	path          string
	format        string
	watch         bool
	watchInterval time.Duration // polling interval; defaults to time.Second
	mu            sync.Mutex    // protects lastData and lastMod
	lastData      map[string]any
	lastMod       time.Time
}

// NewFileSource creates a new file configuration source.
// Supported formats: FormatJSON, FormatEnv.
func NewFileSource(path string, format string, watch bool) *FileSource {
	return &FileSource{
		path:          path,
		format:        strings.ToLower(format), // normalize once at construction
		watch:         watch,
		watchInterval: time.Second,
	}
}

// WithWatchInterval sets the file-polling interval for hot-reload.
// Panics if d <= 0. Returns the receiver for chaining.
func (f *FileSource) WithWatchInterval(d time.Duration) *FileSource {
	if d <= 0 {
		panic("config: FileSource watch interval must be positive")
	}
	f.watchInterval = d
	return f
}

// Load loads configuration from the file.
func (f *FileSource) Load(ctx context.Context) (map[string]any, error) {
	data, err := f.loadFile()
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	f.lastData = data
	if info, err := os.Stat(f.path); err == nil {
		f.lastMod = info.ModTime()
	}
	f.mu.Unlock()
	return data, nil
}

// Watch watches for file changes.
// Returns a channel that sends WatchResult containing updates or errors.
func (f *FileSource) Watch(ctx context.Context) <-chan WatchResult {
	results := make(chan WatchResult, 1)

	if !f.watch {
		close(results)
		return results
	}

	go func() {
		defer close(results)

		ticker := time.NewTicker(f.watchInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				info, statErr := os.Stat(f.path)
				if statErr != nil {
					continue
				}

				f.mu.Lock()
				modChanged := info.ModTime().After(f.lastMod)
				f.mu.Unlock()

				if !modChanged {
					continue
				}

				data, err := f.loadFile()
				if err != nil {
					select {
					case results <- WatchResult{Err: fmt.Errorf("failed to load updated file: %w", err)}:
					case <-ctx.Done():
						return
					}
					continue
				}

				f.mu.Lock()
				changed := !mapsEqual(f.lastData, data)
				if changed {
					f.lastData = data
				}
				f.lastMod = info.ModTime()
				f.mu.Unlock()

				if changed {
					select {
					case results <- WatchResult{Data: data}:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return results
}

// Name returns the source name.
func (f *FileSource) Name() string {
	return fmt.Sprintf("file:%s", f.path)
}

// loadFile loads configuration from file based on format.
func (f *FileSource) loadFile() (map[string]any, error) {
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", f.path, err)
	}

	switch f.format {
	case FormatJSON:
		return f.loadJSON(content)
	case FormatEnv:
		return f.loadEnvFile(content)
	case FormatTOML:
		return nil, fmt.Errorf("TOML format not yet implemented: %s", f.format)
	default:
		return nil, fmt.Errorf("unsupported format: %s", f.format)
	}
}

// loadJSON loads JSON configuration.
func (f *FileSource) loadJSON(content []byte) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return data, nil
}

// loadEnvFile loads environment file format.
func (f *FileSource) loadEnvFile(content []byte) (map[string]any, error) {
	data := make(map[string]any)
	for _, line := range strings.Split(string(content), "\n") {
		key, value, ok := parseEnvLine(line)
		if !ok {
			continue
		}
		data[toSnakeCase(key)] = value
	}
	return data, nil
}

// mapsEqual compares two maps for equality.
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists || !valuesEqual(valueA, valueB) {
			return false
		}
	}

	return true
}
