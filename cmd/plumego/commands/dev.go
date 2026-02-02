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

func (c *DevCmd) Run(args []string) error {
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

	return c.run(*dir, *addr, *dashboardAddr, *watchPatterns, *excludePatterns, *debounceStr)
}

func (c *DevCmd) run(dir, addr, dashboardAddr, watchPatterns, excludePatterns, debounceStr string) error {
	// Parse debounce duration
	debounce, err := time.ParseDuration(debounceStr)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("invalid debounce duration: %v", err), 1)
	}

	// Get absolute directory
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return output.NewFormatter().Error(fmt.Sprintf("directory not found: %s", absDir), 1)
	}

	fmt.Printf("🚀 Starting Plumego Dev Server\n")
	fmt.Printf("   Project: %s\n", absDir)
	fmt.Printf("   App URL: http://localhost%s\n", addr)
	fmt.Printf("   Dashboard URL: http://localhost%s\n", dashboardAddr)
	fmt.Println()

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
		return output.NewFormatter().Error(fmt.Sprintf("failed to create dashboard: %v", err), 1)
	}

	ctx := context.Background()

	// Start dashboard server
	if err := dash.Start(ctx); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to start dashboard: %v", err), 1)
	}

	fmt.Printf("✓ Dashboard started at http://localhost%s\n\n", dashboardAddr)

	// Build and run the application
	if err := dash.BuildAndRun(ctx); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to build and run: %v", err), 1)
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
		return output.NewFormatter().Error(fmt.Sprintf("failed to create watcher: %v", err), 1)
	}
	defer w.Close()

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("👀 Watching for changes...")
	fmt.Println("   Press Ctrl+C to stop")

	// Main event loop
	for {
		select {
		case path := <-w.Events():
			fmt.Printf("\n📝 File changed: %s\n", path)

			// Publish file change event to dashboard
			dash.PublishEvent(devserver.EventFileChange, devserver.FileChangeEvent{
				Path:   path,
				Action: "modify",
			})

			// Trigger rebuild and restart
			if err := dash.Rebuild(ctx); err != nil {
				fmt.Printf("❌ Reload failed: %v\n", err)
			} else {
				fmt.Println("✓ Reload complete\n")
			}

		case err := <-w.Errors():
			fmt.Printf("⚠️  Watcher error: %v\n", err)

		case <-sigChan:
			fmt.Println("\n\n🛑 Shutting down...")
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
