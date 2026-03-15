package commands

import (
	"flag"
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
	// Check for help flag first
	if containsHelpFlag(args) {
		// Show help
		help := `Usage: plumego serve [options] [directory]

Start a static file server for serving files from a directory.

Options:
  -a, --addr <address>  Server address in the format host:port (default: :8080)

Arguments:
  [directory]  Directory to serve files from (default: current directory)

Examples:
  plumego serve                    # Serve current directory on :8080
  plumego serve ./public           # Serve public directory on :8080
  plumego serve ./dist --addr :3000  # Serve dist directory on :3000
  plumego serve --addr 127.0.0.1:8080  # Serve current directory on specific IP

Features:
  - Serves all files in the specified directory and its subdirectories
  - Automatically generates directory listings for directories without an index.html file
  - Supports common file types with appropriate MIME types
  - Handles range requests for large files

Notes:
  - This server is intended for development use only, not for production deployment
  - For production, consider using a dedicated web server like Nginx or Caddy
`
		ctx.Out.Print(help)
		return nil
	}

	// Parse arguments
	var addrFlag *string

	positionals, err := ParseFlags("serve", args, func(fs *flag.FlagSet) {
		addrFlag = fs.String("addr", ":8080", "Server address")
	})

	if err != nil {
		return NewAppError(1, fmt.Sprintf("invalid flags: %v", err), "")
	}

	// Get directory from positional arguments
	dir := "."
	if len(positionals) > 0 {
		dir = positionals[0]
	}

	addr := *addrFlag

	// Make directory absolute
	dir, err = filepath.Abs(dir)
	if err != nil {
		return NewAppError(1, fmt.Sprintf("invalid directory: %v", err), "")
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return NewAppError(1, fmt.Sprintf("directory does not exist: %s", dir), "")
	}

	// Create file server
	fs := http.FileServer(http.Dir(dir))

	// Start server
	ctx.Out.Print(fmt.Sprintf("Starting static file server at http://localhost%s\n", addr))
	ctx.Out.Print(fmt.Sprintf("Serving directory: %s\n", dir))

	// Start HTTP server
	if err := http.ListenAndServe(addr, fs); err != nil {
		return NewAppError(1, fmt.Sprintf("server error: %v", err), "")
	}

	return nil
}
