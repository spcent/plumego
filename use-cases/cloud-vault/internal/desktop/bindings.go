package desktop

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/spcent/plumego/log"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsBindings exposes desktop functions to the frontend via Wails bindings.
type WailsBindings struct {
	config  *Config
	server  *LocalServer
	service *Service
	tray    *TrayManager
	logger  log.StructuredLogger
	ctx     context.Context
}

// NewWailsBindings creates a new Wails bindings instance.
func NewWailsBindings(config *Config, server *LocalServer, service *Service, tray *TrayManager, logger log.StructuredLogger) *WailsBindings {
	return &WailsBindings{
		config:  config,
		server:  server,
		service: service,
		tray:    tray,
		logger:  logger,
	}
}

// SetContext sets the Wails context (called by Wails runtime).
func (w *WailsBindings) SetContext(ctx context.Context) {
	w.ctx = ctx
}

// GetRuntimeInfo returns desktop runtime information.
func (w *WailsBindings) GetRuntimeInfo() (*RuntimeInfo, error) {
	return w.service.GetRuntimeInfo(w.ctx)
}

// SelectDirectory opens a native directory selection dialog.
func (w *WailsBindings) SelectDirectory(title, defaultPath string) (string, error) {
	if title == "" {
		title = "Select Directory"
	}
	if defaultPath == "" {
		defaultPath = w.config.DataDir
	}

	absDefault, err := filepath.Abs(defaultPath)
	if err != nil {
		absDefault = w.config.DataDir
	}

	selected, err := runtime.OpenDirectoryDialog(w.ctx, runtime.OpenDialogOptions{
		Title:                title,
		DefaultDirectory:     absDefault,
		CanCreateDirectories: true,
		ShowHiddenFiles:      false,
	})

	if err != nil {
		w.logger.Error("directory selection failed", map[string]any{"error": err.Error()})
		return "", err
	}

	if selected == "" {
		w.logger.Info("directory selection cancelled")
		return "", nil
	}

	w.logger.Info("directory selected", map[string]any{"path": selected})
	return selected, nil
}

// ScanPreview performs a preview scan of a directory.
func (w *WailsBindings) ScanPreview(req *ScanPreviewRequest) (*ScanPreviewResponse, error) {
	w.logger.Info("scan preview requested", map[string]any{
		"sourceDir": req.SourceDir,
		"recursive": req.Recursive,
	})

	resp, err := w.service.ScanPreview(w.ctx, req)
	if err != nil {
		w.logger.Error("scan preview failed", map[string]any{"error": err.Error()})
		return nil, err
	}

	w.logger.Info("scan preview completed", map[string]any{
		"totalFiles":    resp.TotalFiles,
		"markdownFiles": resp.MarkdownFiles,
		"totalSize":     resp.TotalSize,
	})

	return resp, nil
}

// OpenDataDirectory opens the data directory in the file manager.
func (w *WailsBindings) OpenDataDirectory() error {
	return w.service.OpenDataDir(w.ctx)
}

// GetDataDirInfo returns information about the data directory.
func (w *WailsBindings) GetDataDirInfo() (*DataDirInfo, error) {
	return w.service.GetDataDirInfo(w.ctx)
}

// ShowNotification displays a desktop notification.
func (w *WailsBindings) ShowNotification(title, message string) error {
	if !w.config.NativeNotifications {
		w.logger.Info("notification skipped (disabled)", map[string]any{
			"title":   title,
			"message": message,
		})
		return nil
	}

	// Try Wails native notification first
	err := runtime.SendNotification(w.ctx, runtime.NotificationOptions{
		Title: title,
		Body:  message,
	})

	if err != nil {
		// Fall back to systray notification
		w.logger.Warn("wails notification failed, using systray", map[string]any{"error": err.Error()})
		return w.tray.ShowNotification(title, message)
	}

	w.logger.Info("notification shown", map[string]any{
		"title":   title,
		"message": message,
	})
	return nil
}

// ImportProgress tracks import job progress.
type ImportProgress struct {
	JobID          string `json:"job_id"`
	Status         string `json:"status"` // "pending", "running", "completed", "failed"
	TotalFiles     int    `json:"total_files"`
	ProcessedFiles int    `json:"processed_files"`
	SuccessFiles   int    `json:"success_files"`
	SkippedFiles   int    `json:"skipped_files"`
	FailedFiles    int    `json:"failed_files"`
	CurrentFile    string `json:"current_file"`
	StartTime      string `json:"start_time"`
	ElapsedSeconds int    `json:"elapsed_seconds"`
}

// StartImport initiates an import job with progress tracking.
func (w *WailsBindings) StartImport(jobID string, req *ScanPreviewRequest) error {
	w.logger.Info("import started", map[string]any{
		"jobID":     jobID,
		"sourceDir": req.SourceDir,
	})

	// The actual import logic will be handled by the importer service
	// This binding just triggers it and sets up progress tracking

	// Emit initial progress event
	w.emitProgress(jobID, &ImportProgress{
		JobID:      jobID,
		Status:     "running",
		StartTime:  time.Now().Format(time.RFC3339),
	})

	return nil
}

// emitProgress sends a progress event to the frontend.
func (w *WailsBindings) emitProgress(jobID string, progress *ImportProgress) {
	if w.ctx == nil {
		return
	}

	runtime.EventsEmit(w.ctx, fmt.Sprintf("import:progress:%s", jobID), progress)
}

// GetConfig returns the current desktop configuration.
func (w *WailsBindings) GetConfig() (*Config, error) {
	return w.config, nil
}

// UpdateConfig updates the desktop configuration.
func (w *WailsBindings) UpdateConfig(newConfig *Config) error {
	// Only allow updating certain fields
	w.config.CloseToTray = newConfig.CloseToTray
	w.config.NativeNotifications = newConfig.NativeNotifications
	w.config.Import = newConfig.Import

	w.logger.Info("config updated", map[string]any{
		"closeToTray":         w.config.CloseToTray,
		"nativeNotifications": w.config.NativeNotifications,
	})

	return nil
}

// Quit initiates application shutdown.
func (w *WailsBindings) Quit() {
	w.logger.Info("quit requested")
	w.tray.Quit()
}

// MinimizeToTray hides the window and minimizes to tray.
func (w *WailsBindings) MinimizeToTray() {
	w.logger.Info("minimize to tray requested")
	w.tray.HideWindow()
	if w.ctx != nil {
		runtime.WindowHide(w.ctx)
	}
}

// RestoreFromTray shows the window from tray.
func (w *WailsBindings) RestoreFromTray() {
	w.logger.Info("restore from tray requested")
	w.tray.ShowWindow()
	if w.ctx != nil {
		runtime.WindowUnminimise(w.ctx)
		runtime.WindowShow(w.ctx)
	}
}

// IsWindowVisible returns whether the window is currently visible.
func (w *WailsBindings) IsWindowVisible() bool {
	return w.tray.IsVisible()
}
