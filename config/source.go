package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// EnvSource implements Source for environment variables
type EnvSource struct {
	prefix string
}

// NewEnvSource creates a new environment variable source
func NewEnvSource(prefix string) *EnvSource {
	return &EnvSource{
		prefix: prefix,
	}
}

// Load loads configuration from environment variables
func (e *EnvSource) Load(ctx context.Context) (map[string]any, error) {
	data := make(map[string]any)

	for _, env := range os.Environ() {
		key, value := e.parseEnvVar(env)

		if e.prefix != "" && !strings.HasPrefix(key, e.prefix) {
			continue
		}

		// Remove prefix for internal key
		if e.prefix != "" {
			key = strings.TrimPrefix(key, e.prefix)
		}

		// Always convert to snake_case for consistency
		key = toSnakeCase(key)

		data[key] = value
	}

	return data, nil
}

// Watch watches for environment variable changes
func (e *EnvSource) Watch(ctx context.Context) (<-chan map[string]any, <-chan error) {
	updates := make(chan map[string]any, 1)
	errs := make(chan error, 1)

	// Environment variables don't typically change during runtime
	// Return closed channels to indicate no updates
	close(updates)
	close(errs)

	return updates, errs
}

// Name returns the source name
func (e *EnvSource) Name() string {
	return "env"
}

// parseEnvVar parses an environment variable string
func (e *EnvSource) parseEnvVar(env string) (key, value string) {
	parts := strings.SplitN(env, "=", 2)
	if len(parts) != 2 {
		return "", ""
	}

	key = strings.TrimSpace(parts[0])
	value = strings.TrimSpace(parts[1])

	// Handle quoted values with proper escaping
	if len(value) >= 2 {
		quote := value[0]
		if (quote == '"' || quote == '\'') && value[len(value)-1] == quote {
			value = value[1 : len(value)-1]
			// Handle escaped characters
			value = strings.ReplaceAll(value, fmt.Sprintf("\\%c", quote), string(quote))
		}
	}

	return key, value
}

// FileSource implements Source for configuration files
type FileSource struct {
	path     string
	format   string
	watch    bool
	lastData map[string]any
	lastMod  time.Time
}

// Format constants
const (
	FormatJSON = "json"
	FormatTOML = "toml"
	FormatEnv  = "env"
)

// NewFileSource creates a new file configuration source
func NewFileSource(path string, format string, watch bool) *FileSource {
	return &FileSource{
		path:   path,
		format: format,
		watch:  watch,
	}
}

// Load loads configuration from file
func (f *FileSource) Load(ctx context.Context) (map[string]any, error) {
	data, err := f.loadFile()
	if err != nil {
		return nil, err
	}

	f.lastData = data
	if info, err := os.Stat(f.path); err == nil {
		f.lastMod = info.ModTime()
	}
	return data, nil
}

// Watch watches for file changes
func (f *FileSource) Watch(ctx context.Context) (<-chan map[string]any, <-chan error) {
	updates := make(chan map[string]any, 1)
	errs := make(chan error, 1)

	if !f.watch {
		// Return empty channels if watching is disabled
		close(updates)
		close(errs)
		return updates, errs
	}

	go func() {
		defer close(updates)
		defer close(errs)

		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if info, err := os.Stat(f.path); err == nil {
					if info.ModTime().After(f.lastMod) {
						data, err := f.loadFile()
						if err != nil {
							errs <- fmt.Errorf("failed to load updated file: %w", err)
							continue
						}

						// Only send update if data actually changed
						if !mapsEqual(f.lastData, data) {
							f.lastData = data
							f.lastMod = info.ModTime()
							select {
							case updates <- data:
							case <-ctx.Done():
								return
							}
						} else {
							f.lastMod = info.ModTime()
						}
					}
				}
			}
		}
	}()

	return updates, errs
}

// Name returns the source name
func (f *FileSource) Name() string {
	return fmt.Sprintf("file:%s", f.path)
}

// loadFile loads configuration from file based on format
func (f *FileSource) loadFile() (map[string]any, error) {
	content, err := os.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", f.path, err)
	}

	switch strings.ToLower(f.format) {
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

// loadJSON loads JSON configuration
func (f *FileSource) loadJSON(content []byte) (map[string]any, error) {
	var data map[string]any
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data, nil
}

// loadEnvFile loads environment file format
func (f *FileSource) loadEnvFile(content []byte) (map[string]any, error) {
	data := make(map[string]any)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		key = toSnakeCase(key)
		data[key] = value
	}

	return data, nil
}

// mapsEqual compares two maps for equality
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		// Type-aware comparison
		if !valuesEqual(valueA, valueB) {
			return false
		}
	}

	return true
}

// toSnakeCase function is now defined in config.go
