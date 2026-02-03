package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/devserver"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/cmd/plumego/internal/watcher"
)

type DevCmd struct{}

func (c *DevCmd) Name() string {
	return "dev"
}

func (c *DevCmd) Short() string {
	return "Start development server with dashboard and hot reload"
}

func (c *DevCmd) Long() string {
	return `Start a development server with web dashboard and automatic hot reload.

This command runs your application with a powerful web-based dashboard that provides:
- Real-time log streaming with filtering
- Auto-discovered route browser
- Performance metrics and health monitoring
- Manual build triggers and app control
- Hot reload on file changes (< 5 seconds)

The dashboard is built with plumego itself (dogfooding), demonstrating the
framework's capabilities while providing an enhanced development experience.

Examples:
  plumego dev                                    # Dashboard at :9999, app at :8080
  plumego dev --addr :3000                       # Custom app port
  plumego dev --dashboard-addr :8888             # Custom dashboard port
  plumego dev --watch "**/*.go,**/*.yaml"        # Custom watch patterns
  plumego dev --debounce 1s                      # Slower rebuild trigger`
}

func (c *DevCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: ".", Usage: "Project directory"},
		{Name: "addr", Default: ":8080", Usage: "Application listen address (sets APP_ADDR)"},
		{Name: "dashboard-addr", Default: ":9999", Usage: "Dashboard listen address"},
		{Name: "watch", Default: "**/*.go", Usage: "Watch patterns (comma-separated)"},
		{Name: "exclude", Default: "", Usage: "Exclude patterns (comma-separated)"},
		{Name: "debounce", Default: "500ms", Usage: "Debounce duration for file changes"},
	}
}

func (c *DevCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("dev", flag.ExitOnError)

	dir := fs.String("dir", ".", "Project directory")
	addr := fs.String("addr", ":8080", "Application listen address")
	dashboardAddr := fs.String("dashboard-addr", ":9999", "Dashboard listen address")
	watchPatterns := fs.String("watch", "**/*.go", "Watch patterns")
	excludePatterns := fs.String("exclude", "", "Exclude patterns")
	debounceStr := fs.String("debounce", "500ms", "Debounce duration")

	if err := fs.Parse(args); err != nil {
		return err
	}

	return c.run(ctx.Out, *dir, *addr, *dashboardAddr, *watchPatterns, *excludePatterns, *debounceStr)
}

func (c *DevCmd) run(out *output.Formatter, dir, addr, dashboardAddr, watchPatterns, excludePatterns, debounceStr string) error {
	// Parse debounce duration
	debounce, err := time.ParseDuration(debounceStr)
	if err != nil {
		return out.Error(fmt.Sprintf("invalid debounce duration: %v", err), 1)
	}

	// Get absolute directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return out.Error(fmt.Sprintf("directory not found: %s", absDir), 1)
	}

	if err := emitDevStart(out, absDir, addr, dashboardAddr); err != nil {
		return err
	}

	// Get UI path (embedded or disk)
	uiPath := filepath.Join(getExecutableDir(), "internal", "devserver", "ui")

	// Create dashboard
	dash, err := devserver.NewDashboard(devserver.Config{
		DashboardAddr: dashboardAddr,
		AppAddr:       addr,
		ProjectDir:    absDir,
		UIPath:        uiPath,
	})
	if err != nil {
		return out.Error(fmt.Sprintf("failed to create dashboard: %v", err), 1)
	}

	ctx := context.Background()

	// Start dashboard server
	if err := dash.Start(ctx); err != nil {
		return out.Error(fmt.Sprintf("failed to start dashboard: %v", err), 1)
	}

	if err := emitDashboardStarted(out, dashboardAddr); err != nil {
		return err
	}

	// Build and run the application
	if err := dash.BuildAndRun(ctx); err != nil {
		return out.Error(fmt.Sprintf("failed to build and run: %v", err), 1)
	}
	if err := emitAppReady(out, addr); err != nil {
		return err
	}

	// Parse watch patterns
	watches := parsePatterns(watchPatterns)
	excludes := parsePatterns(excludePatterns)

	// Add sensible default excludes
	excludes = append(excludes,
		"**/vendor/**",
		"**/node_modules/**",
		"**/.git/**",
		"**/*_test.go",
		"**/.dev-server",
	)

	// Start file watcher
	w, err := watcher.NewWatcher(absDir, watches, excludes, debounce)
	if err != nil {
		dash.Stop(ctx)
		return out.Error(fmt.Sprintf("failed to create watcher: %v", err), 1)
	}
	defer w.Close()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if err := emitWatching(out); err != nil {
		return err
	}

	// Main event loop
	for {
		select {
		case path := <-w.Events():
			if err := emitFileChanged(out, path); err != nil {
				return err
			}

			// Publish file change event to dashboard
			dash.PublishEvent(devserver.EventFileChange, devserver.FileChangeEvent{
				Path:   path,
				Action: "modify",
			})

			if out.Format() != "text" {
				if err := out.Event(output.Event{
					Event:   "reloading",
					Message: "Reloading application",
					Data: map[string]any{
						"reason": "file_changed",
						"path":   path,
					},
				}); err != nil {
					return err
				}
			}

			// Trigger rebuild and restart
			if err := dash.Rebuild(ctx); err != nil {
				if err := emitReloadFailed(out, err); err != nil {
					return err
				}
			} else {
				if err := emitReloadComplete(out); err != nil {
					return err
				}
			}

		case err := <-w.Errors():
			if err := emitWatcherError(out, err); err != nil {
				return err
			}

		case <-sigChan:
			if err := emitShutdown(out); err != nil {
				return err
			}
			dash.Stop(ctx)
			return nil
		}
	}
}

func parsePatterns(s string) []string {
	if s == "" {
		return nil
	}

	patterns := strings.Split(s, ",")
	result := make([]string, 0, len(patterns))

	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}

	return result
}

func emitDevStart(out *output.Formatter, absDir, addr, dashboardAddr string) error {
	if out.Format() == "text" {
		if err := out.Textln("🚀 Starting Plumego Dev Server"); err != nil {
			return err
		}
		if err := out.Textln(fmt.Sprintf("   Project: %s", absDir)); err != nil {
			return err
		}
		if err := out.Textln(fmt.Sprintf("   App URL: http://localhost%s", addr)); err != nil {
			return err
		}
		if err := out.Textln(fmt.Sprintf("   Dashboard URL: http://localhost%s", dashboardAddr)); err != nil {
			return err
		}
		return out.Textln("")
	}

	return out.Event(output.Event{
		Event:   "starting",
		Message: "Plumego Dev Server starting",
		Data: map[string]any{
			"project":        absDir,
			"app_addr":       addr,
			"dashboard_addr": dashboardAddr,
		},
	})
}

func emitDashboardStarted(out *output.Formatter, dashboardAddr string) error {
	if out.Format() == "text" {
		if err := out.Textln(fmt.Sprintf("✓ Dashboard started at http://localhost%s", dashboardAddr)); err != nil {
			return err
		}
		return out.Textln("")
	}

	return out.Event(output.Event{
		Event:   "dashboard_started",
		Message: "Dashboard started",
		Data: map[string]any{
			"url": fmt.Sprintf("http://localhost%s", dashboardAddr),
		},
	})
}

func emitAppReady(out *output.Formatter, addr string) error {
	if out.Format() == "text" {
		return nil
	}

	return out.Event(output.Event{
		Event:   "ready",
		Message: "Application ready",
		Data: map[string]any{
			"url": fmt.Sprintf("http://localhost%s", addr),
		},
	})
}

func emitWatching(out *output.Formatter) error {
	if out.Format() == "text" {
		if err := out.Textln("👀 Watching for changes..."); err != nil {
			return err
		}
		return out.Textln("   Press Ctrl+C to stop")
	}

	return out.Event(output.Event{
		Event:   "watching",
		Message: "Watching for changes",
	})
}

func emitFileChanged(out *output.Formatter, path string) error {
	if out.Format() == "text" {
		if err := out.Textln(""); err != nil {
			return err
		}
		return out.Textln(fmt.Sprintf("📝 File changed: %s", path))
	}

	return out.Event(output.Event{
		Event:   "file_changed",
		Message: "File changed",
		Data: map[string]any{
			"path": path,
		},
	})
}

func emitReloadFailed(out *output.Formatter, reloadErr error) error {
	if out.Format() == "text" {
		return out.Textln(fmt.Sprintf("❌ Reload failed: %v", reloadErr))
	}

	return out.Event(output.Event{
		Event:   "reload_failed",
		Level:   "error",
		Message: "Reload failed",
		Data: map[string]any{
			"error": reloadErr.Error(),
		},
	})
}

func emitReloadComplete(out *output.Formatter) error {
	if out.Format() == "text" {
		if err := out.Textln("✓ Reload complete"); err != nil {
			return err
		}
		return out.Textln("")
	}

	return out.Event(output.Event{
		Event:   "reload_complete",
		Message: "Reload complete",
	})
}

func emitWatcherError(out *output.Formatter, watchErr error) error {
	if out.Format() == "text" {
		return out.Textln(fmt.Sprintf("⚠️  Watcher error: %v", watchErr))
	}

	return out.Event(output.Event{
		Event:   "watcher_error",
		Level:   "warn",
		Message: "Watcher error",
		Data: map[string]any{
			"error": watchErr.Error(),
		},
	})
}

func emitShutdown(out *output.Formatter) error {
	if out.Format() == "text" {
		if err := out.Textln(""); err != nil {
			return err
		}
		if err := out.Textln(""); err != nil {
			return err
		}
		return out.Textln("🛑 Shutting down...")
	}

	return out.Event(output.Event{
		Event:   "stopped",
		Message: "Shutting down",
		Data: map[string]any{
			"code": 0,
		},
	})
}

// getExecutableDir returns the directory containing the plumego executable
func getExecutableDir() string {
	ex, err := os.Executable()
	if err != nil {
		// Fallback to working directory
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(ex)
}
