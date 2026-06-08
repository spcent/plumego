package commands

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ServeCmd represents the serve command for static file server
type ServeCmd struct{}

// Name returns the command name
func (c *ServeCmd) Name() string {
	return "serve"
}

// Short returns the command short description
func (c *ServeCmd) Short() string {
	return "Start static file server"
}

type serveOptions struct {
	dir  string
	addr string
}

// Run executes the serve command
func (c *ServeCmd) Run(ctx *Context, args []string) error {
	opts, err := parseServeArgs(args)
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return printCommandHelp(ctx.Out, c)
		}
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	dir, err := filepath.Abs(opts.dir)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return ctx.Out.Error(fmt.Sprintf("directory does not exist: %s", dir), 1)
		}
		return ctx.Out.Error(fmt.Sprintf("failed to inspect directory: %v", err), 1)
	}
	if !info.IsDir() {
		return ctx.Out.Error(fmt.Sprintf("path is not a directory: %s", dir), 1)
	}

	listener, err := net.Listen("tcp", opts.addr)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to listen on %s: %v", opts.addr, err), 1)
	}

	server := &http.Server{
		Handler:           http.FileServer(http.Dir(dir)),
		ReadHeaderTimeout: 5 * time.Second,
	}
	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
			return
		}
		serveErr <- nil
	}()

	addr := listener.Addr().String()
	if err := ctx.Out.Success("Static file server started", map[string]any{
		"addr":      addr,
		"directory": dir,
		"url":       "http://" + addr,
	}); err != nil {
		_ = listener.Close()
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	select {
	case err := <-serveErr:
		if err != nil {
			return ctx.Out.Error(fmt.Sprintf("server error: %v", err), 1)
		}
		return nil
	case <-sigChan:
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return ctx.Out.Error(fmt.Sprintf("server shutdown failed: %v", err), 1)
		}
		if err := <-serveErr; err != nil {
			return ctx.Out.Error(fmt.Sprintf("server error: %v", err), 1)
		}
		return ctx.Out.Success("Static file server stopped", map[string]any{
			"addr":      addr,
			"directory": dir,
		})
	}
}

func parseServeArgs(args []string) (serveOptions, error) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	addrFlag := fs.String("addr", ":8080", "Server address (e.g. :8080 or 0.0.0.0:8080)")
	fs.StringVar(addrFlag, "a", ":8080", "Server address shorthand")
	portFlag := fs.String("port", "", "Listen port (e.g. 3000); shorthand for --addr :PORT")
	fs.StringVar(portFlag, "p", "", "Listen port shorthand")

	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return serveOptions{}, err
	}
	if len(positionals) > 1 {
		return serveOptions{}, fmt.Errorf("serve accepts at most one directory")
	}

	addr := *addrFlag
	if p := *portFlag; p != "" {
		if !strings.HasPrefix(p, ":") {
			p = ":" + p
		}
		addr = p
	}

	dir := "."
	if len(positionals) == 1 {
		dir = positionals[0]
	}
	return serveOptions{dir: dir, addr: addr}, nil
}
