package desktop

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spcent/plumego/log"
)

// RuntimeInfo contains desktop runtime information.
type RuntimeInfo struct {
	Mode         string `json:"mode"`
	Version      string `json:"version"`
	BaseURL      string `json:"base_url"`
	Port         int    `json:"port"`
	DataDir      string `json:"data_dir"`
	StoragePath  string `json:"storage_path"`
	Platform     string `json:"platform"`
	PID          int    `json:"pid"`
	StartTime    string `json:"start_time"`
	TokenEnabled bool   `json:"token_enabled"`
}

// Service provides desktop-specific functionality.
type Service struct {
	config    *Config
	server    *LocalServer
	startTime time.Time
	logger    log.StructuredLogger
}

// NewService creates a new desktop service.
func NewService(config *Config, server *LocalServer, logger log.StructuredLogger) *Service {
	return &Service{
		config:    config,
		server:    server,
		startTime: time.Now(),
		logger:    logger,
	}
}

// GetRuntimeInfo returns current runtime information.
func (s *Service) GetRuntimeInfo(ctx context.Context) (*RuntimeInfo, error) {
	info := &RuntimeInfo{
		Mode:        "desktop",
		Version:     "0.9.0",
		DataDir:     s.config.DataDir,
		StoragePath: filepath.Join(s.config.DataDir, "objects"),
		Platform:    getPlatform(),
		PID:         os.Getpid(),
		StartTime:   s.startTime.Format(time.RFC3339),
	}

	if s.server != nil {
		info.Port = s.server.Port()
		info.BaseURL = s.server.BaseURL()
		info.TokenEnabled = true
	}

	return info, nil
}

// ScanPreviewRequest contains parameters for scan preview.
type ScanPreviewRequest struct {
	SourceDir     string `json:"source_dir"`
	Recursive     bool   `json:"recursive"`
	MaxFileSizeMB int    `json:"max_file_size_mb"`
	IgnoreHidden  bool   `json:"ignore_hidden"`
	IgnoreEmpty   bool   `json:"ignore_empty"`
}

// ScanPreviewResponse contains scan preview results.
type ScanPreviewResponse struct {
	TotalFiles    int64    `json:"total_files"`
	TotalSize     int64    `json:"total_size"`
	MarkdownFiles int64    `json:"markdown_files"`
	EmptyFiles    int64    `json:"empty_files"`
	LargeFiles    int64    `json:"large_files"`
	SkippedFiles  int64    `json:"skipped_files"`
	SampleFiles   []string `json:"sample_files"`
	Errors        []string `json:"errors"`
}

// ScanPreview performs a preview scan of a directory.
func (s *Service) ScanPreview(ctx context.Context, req *ScanPreviewRequest) (*ScanPreviewResponse, error) {
	if req.SourceDir == "" {
		return nil, fmt.Errorf("source_dir is required")
	}

	absPath, err := filepath.Abs(req.SourceDir)
	if err != nil {
		return nil, fmt.Errorf("get absolute path: %w", err)
	}

	safeRoot, err := filepath.Abs(s.config.DataDir)
	if err != nil {
		return nil, fmt.Errorf("get safe root path: %w", err)
	}
	safeRootWithSep := safeRoot
	if !strings.HasSuffix(safeRootWithSep, string(os.PathSeparator)) {
		safeRootWithSep += string(os.PathSeparator)
	}
	if absPath != safeRoot && !strings.HasPrefix(absPath, safeRootWithSep) {
		return nil, fmt.Errorf("source_dir is outside allowed directory")
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("stat directory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	resp := &ScanPreviewResponse{
		SampleFiles: make([]string, 0, 10),
		Errors:      make([]string, 0),
	}

	maxSize := int64(req.MaxFileSizeMB) * 1024 * 1024
	if maxSize == 0 {
		maxSize = 50 * 1024 * 1024 // default 50MB
	}

	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("error accessing %s: %v", path, err))
			return nil
		}

		// Skip directories
		if info.IsDir() {
			if !req.Recursive && path != absPath {
				return filepath.SkipDir
			}
			return nil
		}

		resp.TotalFiles++
		resp.TotalSize += info.Size()

		// Skip hidden files if requested
		if req.IgnoreHidden && isHidden(info.Name()) {
			resp.SkippedFiles++
			return nil
		}

		// Skip empty files if requested
		if req.IgnoreEmpty && info.Size() == 0 {
			resp.EmptyFiles++
			resp.SkippedFiles++
			return nil
		}

		// Check file size
		if maxSize > 0 && info.Size() > maxSize {
			resp.LargeFiles++
			resp.SkippedFiles++
			return nil
		}

		// Check if markdown file
		ext := filepath.Ext(info.Name())
		if ext == ".md" || ext == ".markdown" {
			resp.MarkdownFiles++
			if len(resp.SampleFiles) < 10 {
				relPath, _ := filepath.Rel(absPath, path)
				resp.SampleFiles = append(resp.SampleFiles, relPath)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	return resp, nil
}

// SelectDirectoryRequest contains parameters for directory selection.
type SelectDirectoryRequest struct {
	Title   string `json:"title"`
	Default string `json:"default"`
}

// SelectDirectoryResponse contains the selected directory path.
type SelectDirectoryResponse struct {
	Selected bool   `json:"selected"`
	Path     string `json:"path"`
}

// ShowNotificationRequest contains parameters for showing a notification.
type ShowNotificationRequest struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

// OpenDataDir opens the data directory in the file manager.
func (s *Service) OpenDataDir(ctx context.Context) error {
	return OpenDataDirectory(s.config.DataDir)
}

// GetDataDirInfo returns information about the data directory.
func (s *Service) GetDataDirInfo(ctx context.Context) (*DataDirInfo, error) {
	return GetDataDirInfo(s.config.DataDir)
}

// Handler returns an HTTP handler for desktop API endpoints.
func (s *Service) Handler() http.Handler {
	mux := http.NewServeMux()

	// Runtime info endpoint
	mux.HandleFunc("/api/v1/system/runtime", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info, err := s.GetRuntimeInfo(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": info,
		})
	})

	// Scan preview endpoint
	mux.HandleFunc("/api/v1/import-jobs/scan-preview", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ScanPreviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		resp, err := s.ScanPreview(r.Context(), &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": resp,
		})
	})

	// Data directory info endpoint
	mux.HandleFunc("/api/v1/desktop/data-dir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		info, err := s.GetDataDirInfo(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"data": info,
		})
	})

	// Open data directory endpoint
	mux.HandleFunc("/api/v1/desktop/open-data-dir", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := s.OpenDataDir(r.Context()); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"success": true,
		})
	})

	return mux
}

// getPlatform returns the current platform identifier.
func getPlatform() string {
	return runtime.GOOS
}

// isHidden checks if a file is hidden (starts with .).
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}
