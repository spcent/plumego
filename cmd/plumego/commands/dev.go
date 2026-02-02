package commands

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
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
	return "Start development server with hot reload"
}

func (c *DevCmd) Long() string {
	return `Start a development server with automatic hot reload.

This command runs your application and watches for file changes,
automatically rebuilding and restarting when Go files are modified.

New: Use --dashboard to enable the web-based development dashboard!

Examples:
  plumego dev
  plumego dev --addr :3000
  plumego dev --dashboard :9999
  plumego dev --watch "**/*.go,**/*.yaml"
  plumego dev --no-reload`
}

func (c *DevCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: ".", Usage: "Project directory"},
		{Name: "addr", Default: ":8080", Usage: "Listen address (sets APP_ADDR)"},
		{Name: "dashboard", Default: "", Usage: "Enable dashboard on specified address (e.g., :9999)"},
		{Name: "watch", Default: "**/*.go", Usage: "Watch patterns (comma-separated)"},
		{Name: "exclude", Default: "", Usage: "Exclude patterns (comma-separated)"},
		{Name: "no-reload", Default: "false", Usage: "Disable hot reload"},
		{Name: "build-cmd", Default: "", Usage: "Custom build command"},
		{Name: "run-cmd", Default: "", Usage: "Custom run command"},
		{Name: "debounce", Default: "500ms", Usage: "Debounce duration for file changes"},
	}
}

func (c *DevCmd) Run(args []string) error {
	fs := flag.NewFlagSet("dev", flag.ExitOnError)

	dir := fs.String("dir", ".", "Project directory")
	addr := fs.String("addr", ":8080", "Listen address")
	dashboardAddr := fs.String("dashboard", "", "Dashboard address")
	watchPatterns := fs.String("watch", "**/*.go", "Watch patterns")
	excludePatterns := fs.String("exclude", "", "Exclude patterns")
	noReload := fs.Bool("no-reload", false, "Disable hot reload")
	buildCmd := fs.String("build-cmd", "", "Custom build command")
	runCmd := fs.String("run-cmd", "", "Custom run command")
	debounceStr := fs.String("debounce", "500ms", "Debounce duration")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// If dashboard is enabled, use new dashboard mode
	if *dashboardAddr != "" {
		return c.runWithDashboard(*dir, *addr, *dashboardAddr, *watchPatterns, *excludePatterns, *debounceStr)
	}

	// Otherwise, use legacy mode (backward compatible)
	return c.runLegacyMode(*dir, *addr, *watchPatterns, *excludePatterns, *noReload, *buildCmd, *runCmd, *debounceStr)
}

// runWithDashboard runs dev server with web dashboard
func (c *DevCmd) runWithDashboard(dir, addr, dashboardAddr, watchPatterns, excludePatterns, debounceStr string) error {
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

	fmt.Printf("🚀 Starting Plumego Dev Server with Dashboard\n")
	fmt.Printf("   Project: %s\n", absDir)
	fmt.Printf("   App URL: http://localhost%s\n", addr)
	fmt.Printf("   Dashboard URL: http://localhost%s\n", dashboardAddr)
	fmt.Println()

	// Get UI path
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

	// Start dashboard
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
	excludes = append(excludes, "**/vendor/**", "**/node_modules/**", "**/.git/**", "**/*_test.go")

	// Watch for file changes
	w, err := watcher.NewWatcher(absDir, watches, excludes, debounce)
	if err != nil {
		dash.Stop(ctx)
		return output.NewFormatter().Error(fmt.Sprintf("failed to create watcher: %v", err), 1)
	}
	defer w.Close()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Println("👀 Watching for changes...")
	fmt.Println("   Press Ctrl+C to stop")

	// Event loop
	for {
		select {
		case path := <-w.Events():
			fmt.Printf("\n📝 File changed: %s\n", path)

			// Publish file change event
			dash.PublishEvent(devserver.EventFileChange, devserver.FileChangeEvent{
				Path:   path,
				Action: "modify",
			})

			// Rebuild and restart
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

// runLegacyMode runs dev server in legacy mode (backward compatible)
func (c *DevCmd) runLegacyMode(dir, addr, watchPatterns, excludePatterns string, noReload bool, buildCmd, runCmd, debounceStr string) error {
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

	// Set environment variable for address
	os.Setenv("APP_ADDR", addr)
	os.Setenv("APP_DEBUG", "true")

	// Parse watch patterns
	watches := parsePatterns(watchPatterns)
	excludes := parsePatterns(excludePatterns)

	// Add default excludes
	excludes = append(excludes, "**/vendor/**", "**/node_modules/**", "**/.git/**", "**/*_test.go")

	if flagVerbose {
		fmt.Printf("Starting development server\n")
		fmt.Printf("  Directory: %s\n", absDir)
		fmt.Printf("  Address: %s\n", addr)
		fmt.Printf("  Watch patterns: %v\n", watches)
		fmt.Printf("  Exclude patterns: %v\n", excludes)
	}

	// Create development server
	devServer := &DevServer{
		Dir:      absDir,
		Addr:     addr,
		BuildCmd: buildCmd,
		RunCmd:   runCmd,
		NoReload: noReload,
		Watch:    watches,
		Exclude:  excludes,
		Debounce: debounce,
	}

	return devServer.Run()
}

type DevServer struct {
	Dir      string
	Addr     string
	BuildCmd string
	RunCmd   string
	NoReload bool
	Watch    []string
	Exclude  []string
	Debounce time.Duration

	process *os.Process
}

func (d *DevServer) Run() error {
	// Build and start initially
	if err := d.build(); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("initial build failed: %v", err), 1)
	}

	if err := d.start(); err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to start: %v", err), 1)
	}

	if d.NoReload {
		// Just wait for signals
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		d.stop()
		return nil
	}

	// Watch for file changes
	w, err := watcher.NewWatcher(d.Dir, d.Watch, d.Exclude, d.Debounce)
	if err != nil {
		d.stop()
		return output.NewFormatter().Error(fmt.Sprintf("failed to create watcher: %v", err), 1)
	}
	defer w.Close()

	// Handle signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	fmt.Printf("Development server running at http://localhost%s\n", d.Addr)
	fmt.Println("Watching for changes...")

	for {
		select {
		case event := <-w.Events():
			fmt.Printf("File changed: %s\n", event)
			fmt.Println("Reloading...")

			d.stop()

			if err := d.build(); err != nil {
				fmt.Printf("Build failed: %v\n", err)
				continue
			}

			if err := d.start(); err != nil {
				fmt.Printf("Failed to start: %v\n", err)
				continue
			}

			fmt.Println("Reload complete")

		case err := <-w.Errors():
			fmt.Printf("Watcher error: %v\n", err)

		case <-sigChan:
			fmt.Println("\nShutting down...")
			d.stop()
			return nil
		}
	}
}

func (d *DevServer) build() error {
	if d.BuildCmd != "" {
		// Use custom build command
		parts := strings.Fields(d.BuildCmd)
		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = d.Dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	// Default: go build
	cmd := exec.Command("go", "build", "-o", "./.dev-server", ".")
	cmd.Dir = d.Dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (d *DevServer) start() error {
	var cmd *exec.Cmd

	if d.RunCmd != "" {
		// Use custom run command
		parts := strings.Fields(d.RunCmd)
		cmd = exec.Command(parts[0], parts[1:]...)
	} else {
		// Default: run the built binary
		binaryPath := filepath.Join(d.Dir, ".dev-server")
		cmd = exec.Command(binaryPath)
	}

	cmd.Dir = d.Dir
	cmd.Env = os.Environ()

	// Capture stdout/stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		return err
	}

	d.process = cmd.Process

	// Stream output
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintln(os.Stderr, scanner.Text())
		}
	}()

	// Wait for process in background
	go func() {
		cmd.Wait()
	}()

	return nil
}

func (d *DevServer) stop() {
	if d.process != nil {
		// Try graceful shutdown first
		d.process.Signal(syscall.SIGTERM)

		// Wait a bit
		done := make(chan bool)
		go func() {
			d.process.Wait()
			done <- true
		}()

		select {
		case <-done:
			// Process exited
		case <-time.After(3 * time.Second):
			// Force kill
			d.process.Kill()
			d.process.Wait()
		}

		d.process = nil
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
