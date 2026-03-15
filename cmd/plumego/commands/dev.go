package commands

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/devserver"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/cmd/plumego/internal/watcher"
)

// DevCmd starts the development server. The newDashboard field allows
// test code to inject a mock without using package-level variables.
type DevCmd struct {
	newDashboard func(devserver.Config) (devserver.DashboardAPI, error)
}

// NewDevCmd constructs a DevCmd wired to the real dashboard constructor.
func NewDevCmd() *DevCmd {
	return &DevCmd{
		newDashboard: func(cfg devserver.Config) (devserver.DashboardAPI, error) {
			return devserver.NewDashboard(cfg)
		},
	}
}

type devOptions struct {
	dir             string
	addr            string
	dashboardAddr   string
	watchPatterns   string
	excludePatterns string
	debounceStr     string
	noReload        bool
	buildCmd        string
	runCmd          string
}

func (c *DevCmd) Name() string  { return "dev" }
func (c *DevCmd) Short() string { return "Start development server with dashboard and hot reload" }

func (c *DevCmd) Run(ctx *Context, args []string) error {
	// Check for help flag
	if containsHelpFlag(args) {
		help := `Usage: plumego dev [options]

Start the Plumego development server with hot reload and dashboard.

Options:
  --dir <path>          Project directory (default: .)
  --addr <address>      Application listen address (default: :8080)
  --dashboard-addr <address>  Dashboard listen address (default: 127.0.0.1:9999)
  --watch <patterns>    Watch patterns (default: **/*.go)
  --exclude <patterns>  Exclude patterns
  --debounce <duration>  Debounce duration (default: 500ms)
  --no-reload           Disable hot reload
  --build-cmd <cmd>     Custom build command
  --run-cmd <cmd>       Custom run command

Examples:
  plumego dev                      # Start dev server in current directory
  plumego dev --addr :3000         # Start dev server on port 3000
  plumego dev --dir ./src          # Start dev server in src directory
  plumego dev --watch "**/*.go,**/*.html"  # Watch Go and HTML files
  plumego dev --no-reload          # Start dev server without hot reload
  plumego dev --build-cmd "go build -o app" --run-cmd "./app"  # Custom build and run commands
`
		ctx.Out.Print(help)
		return nil
	}

	opts, err := parseDevArgs(args)
	if err != nil {
		return NewAppError(1, fmt.Sprintf("invalid flags: %v", err), "")
	}

	return c.runWithContext(context.Background(), ctx.Out, opts)
}

func parseDevArgs(args []string) (devOptions, error) {
	opts := devOptions{}

	_, err := ParseFlags("dev", args, func(fs *flag.FlagSet) {
		fs.StringVar(&opts.dir, "dir", ".", "Project directory")
		fs.StringVar(&opts.addr, "addr", ":8080", "Application listen address")
		fs.StringVar(&opts.dashboardAddr, "dashboard-addr", "127.0.0.1:9999", "Dashboard listen address")
		fs.StringVar(&opts.watchPatterns, "watch", "**/*.go", "Watch patterns")
		fs.StringVar(&opts.excludePatterns, "exclude", "", "Exclude patterns")
		fs.StringVar(&opts.debounceStr, "debounce", "500ms", "Debounce duration")
		fs.BoolVar(&opts.noReload, "no-reload", false, "Disable hot reload")
		fs.StringVar(&opts.buildCmd, "build-cmd", "", "Custom build command")
		fs.StringVar(&opts.runCmd, "run-cmd", "", "Custom run command")
	})

	if err != nil {
		return devOptions{}, err
	}

	return opts, nil
}

func (c *DevCmd) runWithContext(ctx context.Context, out *output.Formatter, opts devOptions) error {
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	debounce, err := time.ParseDuration(opts.debounceStr)
	if err != nil {
		return NewAppError(1, fmt.Sprintf("invalid debounce duration: %v", err), "")
	}

	absDir, err := resolveDir(opts.dir)
	if err != nil {
		return NewAppError(1, err.Error(), "")
	}

	if err := emitDevStart(out, absDir, opts.addr, opts.dashboardAddr); err != nil {
		return err
	}

	uiPath := filepath.Join(getExecutableDir(), "internal", "devserver", "ui")

	cfg := devserver.Config{
		DashboardAddr:     opts.dashboardAddr,
		AppAddr:           opts.addr,
		ProjectDir:        absDir,
		UIPath:            uiPath,
		OutputPassthrough: out.Format() == "text" && !out.IsQuiet(),
	}

	if opts.buildCmd != "" {
		cmd, args, err := parseCommandLine(opts.buildCmd)
		if err != nil {
			return NewAppError(1, fmt.Sprintf("invalid build command: %v", err), "")
		}
		if cmd == "" {
			return NewAppError(1, "build command is empty", "")
		}
		cfg.CustomBuildCmd = cmd
		cfg.CustomBuildArgs = args
	}

	if opts.runCmd != "" {
		cmd, args, err := parseCommandLine(opts.runCmd)
		if err != nil {
			return NewAppError(1, fmt.Sprintf("invalid run command: %v", err), "")
		}
		if cmd == "" {
			return NewAppError(1, "run command is empty", "")
		}
		cfg.CustomRunCmd = cmd
		cfg.CustomRunArgs = args
	}

	dash, err := c.newDashboard(cfg)
	if err != nil {
		return NewAppError(1, fmt.Sprintf("failed to create dashboard: %v", err), "")
	}

	if out.Format() != "text" {
		stopForwarder, err := startDevEventForwarder(runCtx, out, dash.GetPubSub())
		if err != nil {
			return NewAppError(1, fmt.Sprintf("failed to subscribe to dev events: %v", err), "")
		}
		defer stopForwarder()
	}

	if err := dash.Start(runCtx); err != nil {
		return NewAppError(1, fmt.Sprintf("failed to start dashboard: %v", err), "")
	}

	if err := emitDashboardStarted(out, opts.dashboardAddr); err != nil {
		return err
	}

	if err := dash.BuildAndRun(runCtx); err != nil {
		return NewAppError(1, fmt.Sprintf("failed to build and run: %v", err), "")
	}
	if err := emitAppReady(out, opts.addr); err != nil {
		return err
	}

	if opts.noReload {
		if err := emitReloadDisabled(out); err != nil {
			return err
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigChan:
		case <-runCtx.Done():
		}
		signal.Stop(sigChan)
		if err := emitShutdown(out); err != nil {
			return err
		}
		dash.Stop(runCtx)
		return nil
	}

	watches := parsePatterns(opts.watchPatterns)
	excludes := parsePatterns(opts.excludePatterns)

	excludes = append(excludes,
		"**/vendor/**",
		"**/node_modules/**",
		"**/.git/**",
		"**/*_test.go",
		"**/.dev-server",
	)

	w, err := watcher.NewWatcher(absDir, watches, excludes, debounce)
	if err != nil {
		dash.Stop(runCtx)
		return NewAppError(1, fmt.Sprintf("failed to create watcher: %v", err), "")
	}
	defer w.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	if err := emitWatching(out); err != nil {
		return err
	}

	for {
		select {
		case path := <-w.Events():
			if err := emitFileChanged(out, path); err != nil {
				return err
			}

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

			if err := dash.Rebuild(runCtx); err != nil {
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
			dash.Stop(runCtx)
			return nil
		}
	}
}
