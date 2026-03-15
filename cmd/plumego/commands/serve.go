package commands

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

// Run executes the serve command
func (c *ServeCmd) Run(ctx *Context, args []string) error {
	// Parse arguments
	var dir string
	var addr string
	var err error

	// Default values
	dir = "."
	addr = ":8080"

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--help", "-h":
			// Show help
			help := `Usage: plumego serve [options] [directory]

Start static file server

Options:
  -a, --addr <address>  Server address (default: :8080)

Examples:
  plumego serve
  plumego serve ./public
  plumego serve ./public --addr :3000
`
			ctx.Out.Print(help)
			return nil
		case "--addr", "-a":
			if i+1 >= len(args) {
				return ctx.Out.Error("--addr requires a value", 1)
			}
			addr = args[i+1]
			i++
		default:
			// Treat as directory
			dir = arg
		}
	}

	// Make directory absolute
	dir, err = filepath.Abs(dir)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return ctx.Out.Error(fmt.Sprintf("directory does not exist: %s", dir), 1)
	}

	// Create file server
	fs := http.FileServer(http.Dir(dir))

	// Start server
	ctx.Out.Print(fmt.Sprintf("Starting static file server at http://localhost%s\n", addr))
	ctx.Out.Print(fmt.Sprintf("Serving directory: %s\n", dir))

	// Start HTTP server
	if err := http.ListenAndServe(addr, fs); err != nil {
		return ctx.Out.Error(fmt.Sprintf("server error: %v", err), 1)
	}

	return nil
}
