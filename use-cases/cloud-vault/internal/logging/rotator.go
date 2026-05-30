package logging

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RotatingFileWriter writes log data to files with automatic rotation.
// When a log file reaches MaxSizeBytes, it's renamed with a timestamp suffix
// and a new file is created. Old files beyond MaxBackups are deleted.
type RotatingFileWriter struct {
	mu          sync.Mutex
	dir         string
	filename    string
	maxSize     int64
	maxBackups  int
	compress    bool
	currentFile *os.File
	currentSize int64
}

// RotatingWriterConfig configures the rotating file writer.
type RotatingWriterConfig struct {
	// Dir is the directory where log files are stored.
	Dir string
	// Filename is the base name of the log file (e.g., "app.log").
	Filename string
	// MaxSizeBytes is the maximum size in bytes before rotation. Default: 100MB.
	MaxSizeBytes int64
	// MaxBackups is the maximum number of old log files to retain. Default: 5.
	MaxBackups int
	// Compress determines whether rotated files are gzipped. Default: false.
	Compress bool
}

// NewRotatingFileWriter creates a new RotatingFileWriter with the given configuration.
func NewRotatingFileWriter(cfg RotatingWriterConfig) (*RotatingFileWriter, error) {
	if cfg.Dir == "" {
		cfg.Dir = "."
	}
	if cfg.Filename == "" {
		cfg.Filename = "app.log"
	}
	if cfg.MaxSizeBytes <= 0 {
		cfg.MaxSizeBytes = 100 * 1024 * 1024 // 100MB
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 5
	}

	// Ensure directory exists
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	rw := &RotatingFileWriter{
		dir:        cfg.Dir,
		filename:   cfg.Filename,
		maxSize:    cfg.MaxSizeBytes,
		maxBackups: cfg.MaxBackups,
		compress:   cfg.Compress,
	}

	if err := rw.openCurrentFile(); err != nil {
		return nil, err
	}

	return rw, nil
}

// Write implements io.Writer. It writes data to the current log file,
// rotating if necessary.
func (rw *RotatingFileWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Check if rotation is needed
	if rw.currentSize+int64(len(p)) > rw.maxSize {
		if err := rw.rotate(); err != nil {
			return 0, fmt.Errorf("rotate log file: %w", err)
		}
	}

	n, err = rw.currentFile.Write(p)
	rw.currentSize += int64(n)
	return n, err
}

// Close closes the current log file.
func (rw *RotatingFileWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	if rw.currentFile != nil {
		return rw.currentFile.Close()
	}
	return nil
}

// openCurrentFile opens or creates the current log file.
func (rw *RotatingFileWriter) openCurrentFile() error {
	path := filepath.Join(rw.dir, rw.filename)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Get current file size
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return fmt.Errorf("stat log file: %w", err)
	}

	rw.currentFile = f
	rw.currentSize = info.Size()
	return nil
}

// rotate closes the current file, renames it with a timestamp,
// opens a new file, and cleans up old backups.
func (rw *RotatingFileWriter) rotate() error {
	// Close current file
	if err := rw.currentFile.Close(); err != nil {
		return fmt.Errorf("close current file: %w", err)
	}

	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102-150405")
	oldPath := filepath.Join(rw.dir, rw.filename)
	newPath := filepath.Join(rw.dir, fmt.Sprintf("%s.%s", rw.filename, timestamp))

	if err := os.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("rename log file: %w", err)
	}

	// Compress if enabled
	if rw.compress {
		go rw.compressFile(newPath)
	}

	// Open new file
	if err := rw.openCurrentFile(); err != nil {
		return err
	}

	// Clean up old backups
	go rw.cleanupOldBackups()

	return nil
}

// compressFile gzips a log file asynchronously.
func (rw *RotatingFileWriter) compressFile(path string) {
	src, err := os.Open(path)
	if err != nil {
		return
	}
	defer src.Close()

	dst, err := os.Create(path + ".gz")
	if err != nil {
		return
	}
	defer dst.Close()

	gz := gzip.NewWriter(dst)
	defer gz.Close()

	if _, err := io.Copy(gz, src); err != nil {
		return
	}

	// Close gzip to flush
	if err := gz.Close(); err != nil {
		return
	}
	if err := dst.Close(); err != nil {
		return
	}
	if err := src.Close(); err != nil {
		return
	}

	// Remove uncompressed file
	os.Remove(path)
}

// cleanupOldBackups removes old log files beyond MaxBackups.
func (rw *RotatingFileWriter) cleanupOldBackups() {
	pattern := filepath.Join(rw.dir, rw.filename+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	if len(matches) <= rw.maxBackups {
		return
	}

	// Sort by modification time (oldest first)
	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo
	for _, match := range matches {
		info, err := os.Stat(match)
		if err != nil {
			continue
		}
		files = append(files, fileInfo{path: match, modTime: info.ModTime()})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Remove oldest files
	toDelete := len(files) - rw.maxBackups
	for i := 0; i < toDelete; i++ {
		os.Remove(files[i].path)
	}
}

// CurrentFilePath returns the path to the current log file.
func (rw *RotatingFileWriter) CurrentFilePath() string {
	return filepath.Join(rw.dir, rw.filename)
}

// ListBackups returns a list of backup log file paths.
func (rw *RotatingFileWriter) ListBackups() ([]string, error) {
	pattern := filepath.Join(rw.dir, rw.filename+".*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Filter out .gz files if they're duplicates
	var result []string
	seen := make(map[string]bool)
	for _, match := range matches {
		base := strings.TrimSuffix(match, ".gz")
		if !seen[base] {
			seen[base] = true
			result = append(result, match)
		}
	}

	return result, nil
}
