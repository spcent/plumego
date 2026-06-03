//go:build wails

package desktop

import (
	"context"
	"sync"

	"github.com/spcent/plumego/log"
)

// TrayManager is a Wails-compatible tray facade.
//
// Wails already links macOS application delegate code. The systray package also
// links an AppDelegate implementation, so Wails builds use this facade to avoid
// duplicate native symbols.
type TrayManager struct {
	config     *Config
	logger     log.StructuredLogger
	cancelFunc context.CancelFunc
	mu         sync.Mutex
	isVisible  bool
}

// NewTrayManager creates a tray facade for Wails builds.
func NewTrayManager(config *Config, _ *LocalServer, _ *Service, logger log.StructuredLogger) *TrayManager {
	return &TrayManager{
		config:    config,
		logger:    logger,
		isVisible: true,
	}
}

// SetCancelFunc sets the cancel function for graceful shutdown.
func (t *TrayManager) SetCancelFunc(cancel context.CancelFunc) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.cancelFunc = cancel
}

// Start is intentionally a no-op in Wails builds.
func (t *TrayManager) Start() {
	t.logger.Info("system tray skipped in wails build")
}

// HideWindow records window visibility state.
func (t *TrayManager) HideWindow() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isVisible = false
	t.logger.Info("window hidden")
}

// ShowWindow records window visibility state.
func (t *TrayManager) ShowWindow() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isVisible = true
	t.logger.Info("window shown")
}

// IsVisible returns whether the window is visible.
func (t *TrayManager) IsVisible() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isVisible
}

// ShouldCloseToTray returns whether to close to tray.
func (t *TrayManager) ShouldCloseToTray() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.config.CloseToTray
}

// Quit performs a full application shutdown.
func (t *TrayManager) Quit() {
	t.mu.Lock()
	cancel := t.cancelFunc
	t.mu.Unlock()

	if cancel != nil {
		t.logger.Info("initiating shutdown")
		cancel()
	}
}

// ShowNotification logs the notification fallback for Wails builds.
func (t *TrayManager) ShowNotification(title, message string) error {
	if !t.config.NativeNotifications {
		t.logger.Info("notification (disabled)", map[string]any{
			"title":   title,
			"message": message,
		})
		return nil
	}

	t.logger.Info("desktop notification", map[string]any{
		"title":   title,
		"message": message,
	})
	return nil
}
