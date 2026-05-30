package diagnostics

import (
	"archive/zip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"cloud-vault/internal/config"
	"cloud-vault/internal/database"
	"cloud-vault/internal/storage"
	"cloud-vault/internal/version"
)

// Service handles diagnostic bundle generation.
type Service struct {
	cfg       config.Config
	db        *database.DB
	store     storage.ObjectStorage
	outputDir string
}

// NewService creates a new diagnostics service.
func NewService(cfg config.Config, db *database.DB, store storage.ObjectStorage, outputDir string) *Service {
	return &Service{
		cfg:       cfg,
		db:        db,
		store:     store,
		outputDir: outputDir,
	}
}

// GenerateBundle creates a diagnostic bundle zip file containing:
// - manifest.json: bundle metadata and file list
// - config.json: redacted configuration
// - system_info.json: runtime and system information
// - build_info.json: version and build metadata
// - logs.txt: recent log entries (if available)
//
// The bundle excludes: database files, storage objects, Markdown content,
// API keys, passwords, tokens, and other secrets.
func (s *Service) GenerateBundle(ctx context.Context) (*Bundle, error) {
	// Generate unique filename
	now := time.Now()
	filename := fmt.Sprintf("diagnostics_%s.zip", now.Format("20060102_150405"))
	outputPath := filepath.Join(s.outputDir, filename)

	// Ensure output directory exists
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("create output directory: %w", err)
	}

	// Create zip file
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("create zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Collect manifest entries
	manifest := struct {
		GeneratedAt string   `json:"generated_at"`
		Files       []string `json:"files"`
		Excludes    []string `json:"excludes"`
	}{
		GeneratedAt: now.UTC().Format(time.RFC3339),
		Files:       []string{},
		Excludes: []string{
			"database files (*.db, *.sqlite)",
			"storage objects (Markdown files, attachments)",
			"API keys and secrets",
			"passwords and tokens",
			"session data",
		},
	}

	// 1. Add redacted configuration
	if err := addJSONFile(zipWriter, "config.json", RedactConfig(s.cfg)); err != nil {
		return nil, fmt.Errorf("add config.json: %w", err)
	}
	manifest.Files = append(manifest.Files, "config.json")

	// 2. Add system information
	sysInfo := collectSystemInfo()
	if err := addJSONFile(zipWriter, "system_info.json", sysInfo); err != nil {
		return nil, fmt.Errorf("add system_info.json: %w", err)
	}
	manifest.Files = append(manifest.Files, "system_info.json")

	// 3. Add build information
	buildInfo := BuildInfo{
		Version:   version.Version,
		Commit:    version.Commit,
		BuildTime: version.BuildTime,
		Channel:   version.Channel,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
	if err := addJSONFile(zipWriter, "build_info.json", buildInfo); err != nil {
		return nil, fmt.Errorf("add build_info.json: %w", err)
	}
	manifest.Files = append(manifest.Files, "build_info.json")

	// 4. Add recent logs (if log directory exists)
	logDir := filepath.Join(filepath.Dir(s.cfg.DB.Path), "logs")
	if _, err := os.Stat(logDir); err == nil {
		logs, err := collectRecentLogs(logDir, 1000)
		if err == nil && len(logs) > 0 {
			if err := addTextFile(zipWriter, "logs.txt", logs); err != nil {
				return nil, fmt.Errorf("add logs.txt: %w", err)
			}
			manifest.Files = append(manifest.Files, "logs.txt")
		}
	}

	// 5. Add manifest last
	if err := addJSONFile(zipWriter, "manifest.json", manifest); err != nil {
		return nil, fmt.Errorf("add manifest.json: %w", err)
	}

	// Close zip to flush
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("close zip writer: %w", err)
	}
	if err := zipFile.Close(); err != nil {
		return nil, fmt.Errorf("close zip file: %w", err)
	}

	// Get file size
	stat, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("stat output file: %w", err)
	}

	return &Bundle{
		ID:           filename,
		Filename:     filename,
		CreatedAt:    now,
		Size:         stat.Size(),
		DownloadPath: outputPath,
	}, nil
}

// addJSONFile adds a JSON-encoded file to the zip archive.
func addJSONFile(w *zip.Writer, name string, data interface{}) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// addTextFile adds a plain text file to the zip archive.
func addTextFile(w *zip.Writer, name string, content string) error {
	f, err := w.Create(name)
	if err != nil {
		return err
	}
	_, err = f.Write([]byte(content))
	return err
}

// collectSystemInfo gathers runtime system information.
func collectSystemInfo() SystemInfo {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemInfo{
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		NumGoroutine: runtime.NumGoroutine(),
		MemAllocMB:   m.Alloc / 1024 / 1024,
		MemSysMB:     m.Sys / 1024 / 1024,
		Uptime:       "unknown", // Would need app start time to calculate
	}
}

// collectRecentLogs reads the most recent N lines from log files.
// Logs are redacted to remove sensitive information.
func collectRecentLogs(logDir string, maxLines int) (string, error) {
	// Look for common log file patterns
	patterns := []string{"*.log", "app.log", "server.log"}

	var allLogs []string
	for _, pattern := range patterns {
		matches, err := filepath.Glob(filepath.Join(logDir, pattern))
		if err != nil {
			continue
		}

		for _, logFile := range matches {
			content, err := os.ReadFile(logFile)
			if err != nil {
				continue
			}

			// Split into lines and redact
			lines := splitLines(string(content))
			for _, line := range lines {
				redacted := RedactLogLine(line)
				allLogs = append(allLogs, redacted)
			}
		}
	}

	// Keep only the most recent lines
	if len(allLogs) > maxLines {
		allLogs = allLogs[len(allLogs)-maxLines:]
	}

	// Join back into single string
	result := ""
	for _, line := range allLogs {
		result += line + "\n"
	}

	return result, nil
}

// splitLines splits text into lines, handling both \n and \r\n.
func splitLines(text string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			line := text[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	// Add last line if no trailing newline
	if start < len(text) {
		lines = append(lines, text[start:])
	}
	return lines
}

// ListBundles returns all available diagnostic bundles.
func (s *Service) ListBundles() ([]Bundle, error) {
	entries, err := os.ReadDir(s.outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []Bundle{}, nil
		}
		return nil, fmt.Errorf("read output directory: %w", err)
	}

	var bundles []Bundle
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "diagnostics_") || !strings.HasSuffix(name, ".zip") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		bundles = append(bundles, Bundle{
			ID:           name,
			Filename:     name,
			CreatedAt:    info.ModTime(),
			Size:         info.Size(),
			DownloadPath: filepath.Join(s.outputDir, name),
		})
	}

	return bundles, nil
}

// GetBundle retrieves a specific diagnostic bundle by filename.
func (s *Service) GetBundle(filename string) (*Bundle, error) {
	// Sanitize filename to prevent path traversal
	cleanName := filepath.Base(filename)
	if cleanName != filename {
		return nil, fmt.Errorf("invalid filename")
	}

	path := filepath.Join(s.outputDir, cleanName)
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("bundle not found: %w", err)
	}

	return &Bundle{
		ID:           cleanName,
		Filename:     cleanName,
		CreatedAt:    info.ModTime(),
		Size:         info.Size(),
		DownloadPath: path,
	}, nil
}
