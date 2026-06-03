//go:build !wails

package desktop

import (
	"context"
	"embed"
	"fmt"
	"os"
	"sync"

	"github.com/getlantern/systray"
	"github.com/spcent/plumego/log"
)

//go:embed resources
var resources embed.FS

// TrayManager manages the system tray icon and menu.
type TrayManager struct {
	config     *Config
	server     *LocalServer
	service    *Service
	logger     log.StructuredLogger
	cancelFunc context.CancelFunc
	mu         sync.Mutex
	isVisible  bool
}

// NewTrayManager creates a new tray manager.
func NewTrayManager(config *Config, server *LocalServer, service *Service, logger log.StructuredLogger) *TrayManager {
	return &TrayManager{
		config:    config,
		server:    server,
		service:   service,
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

// Start initializes the system tray.
func (t *TrayManager) Start() {
	systray.Run(t.onReady, t.onExit)
}

// onReady is called when the systray is ready.
func (t *TrayManager) onReady() {
	t.logger.Info("system tray ready")

	// Set icon
	icon, err := resources.ReadFile("resources/icon.ico")
	if err != nil {
		t.logger.Error("failed to load icon", map[string]any{"error": err.Error()})
		// Use fallback
		icon = []byte{}
	}
	systray.SetIcon(icon)
	systray.SetTitle(t.config.AppName)
	systray.SetTooltip(fmt.Sprintf("%s - Running", t.config.AppName))

	// Build menu
	t.buildMenu()
}

// buildMenu creates the system tray menu.
func (t *TrayManager) buildMenu() {
	// Show/Hide Window
	mShow := systray.AddMenuItem("Show Window", "Show the main window")
	go func() {
		for range mShow.ClickedCh {
			t.logger.Info("show window clicked")
			// Window management will be handled by Wails
		}
	}()

	// Open Data Directory
	mOpenDir := systray.AddMenuItem("Open Data Directory", "Open the data directory in file manager")
	go func() {
		for range mOpenDir.ClickedCh {
			t.logger.Info("open data directory clicked")
			if err := t.service.OpenDataDir(context.Background()); err != nil {
				t.logger.Error("failed to open data directory", map[string]any{"error": err.Error()})
			}
		}
	}()

	// Scan Preview
	mScan := systray.AddMenuItem("Scan Import Directory", "Preview files in import directory")
	go func() {
		for range mScan.ClickedCh {
			t.logger.Info("scan preview clicked")
			// Will be implemented via Wails dialog
		}
	}()

	systray.AddSeparator()

	// Runtime Info
	mInfo := systray.AddMenuItem("Runtime Info", "Show runtime information")
	go func() {
		for range mInfo.ClickedCh {
			t.logger.Info("runtime info clicked")
			info, err := t.service.GetRuntimeInfo(context.Background())
			if err != nil {
				t.logger.Error("failed to get runtime info", map[string]any{"error": err.Error()})
				return
			}
			t.logger.Info("runtime info", map[string]any{
				"mode":    info.Mode,
				"version": info.Version,
				"port":    info.Port,
				"dataDir": info.DataDir,
			})
		}
	}()

	systray.AddSeparator()

	// Close to Tray toggle
	var mCloseToTray *systray.MenuItem
	if t.config.CloseToTray {
		mCloseToTray = systray.AddMenuItemCheckbox("Close to Tray", "Minimize to tray instead of closing", true)
	} else {
		mCloseToTray = systray.AddMenuItemCheckbox("Close to Tray", "Minimize to tray instead of closing", false)
	}
	go func() {
		for range mCloseToTray.ClickedCh {
			t.mu.Lock()
			t.config.CloseToTray = !t.config.CloseToTray
			checked := t.config.CloseToTray
			t.mu.Unlock()

			if checked {
				mCloseToTray.Check()
			} else {
				mCloseToTray.Uncheck()
			}
			t.logger.Info("close to tray toggled", map[string]any{"enabled": checked})
		}
	}()

	systray.AddSeparator()

	// Quit
	mQuit := systray.AddMenuItem("Quit", "Quit the application")
	go func() {
		<-mQuit.ClickedCh
		t.logger.Info("quit clicked")
		t.Quit()
	}()
}

// onExit is called when the systray is exiting.
func (t *TrayManager) onExit() {
	t.logger.Info("system tray exiting")
}

// HideWindow hides the main window (close to tray).
func (t *TrayManager) HideWindow() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.isVisible = false
	t.logger.Info("window hidden")
}

// ShowWindow shows the main window.
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
	systray.Quit()
}

// ShowNotification displays a desktop notification.
// systray does not have a native notification API; the primary path uses Wails'
// runtime.SendNotification. This fallback logs the notification so that it is
// visible in the application log (and can be wired to a platform notifier later).
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

// ensureIcon ensures the icon resource exists.
func ensureIcon() error {
	// Create resources directory if it doesn't exist
	if err := os.MkdirAll("internal/desktop/resources", 0755); err != nil {
		return err
	}

	// Check if icon exists
	if _, err := os.Stat("internal/desktop/resources/icon.ico"); os.IsNotExist(err) {
		// Create a minimal placeholder icon
		// In production, this should be a real icon file
		return os.WriteFile("internal/desktop/resources/icon.ico", []byte{}, 0644)
	}
	return nil
}
