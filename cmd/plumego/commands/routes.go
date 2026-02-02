package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/cmd/plumego/internal/routeanalyzer"
)

type RoutesCmd struct{}

func (c *RoutesCmd) Name() string {
	return "routes"
}

func (c *RoutesCmd) Short() string {
	return "Inspect registered routes"
}

func (c *RoutesCmd) Long() string {
	return `Analyze and display HTTP routes registered in the application.

This command scans the codebase to find route registrations and displays
them in a structured format. Useful for documentation and debugging.

Examples:
  plumego routes
  plumego routes --format json
  plumego routes --method GET
  plumego routes --pattern "/api/*"
  plumego routes --middleware`
}

func (c *RoutesCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: ".", Usage: "Project directory"},
		{Name: "method", Default: "", Usage: "Filter by HTTP method (GET, POST, etc.)"},
		{Name: "pattern", Default: "", Usage: "Filter routes by pattern"},
		{Name: "middleware", Default: "false", Usage: "Show middleware chains"},
		{Name: "group", Default: "", Usage: "Filter by route group"},
		{Name: "sort", Default: "path", Usage: "Sort by: path, method, group"},
	}
}

func (c *RoutesCmd) Run(args []string) error {
	fs := flag.NewFlagSet("routes", flag.ExitOnError)

	dir := fs.String("dir", ".", "Project directory")
	method := fs.String("method", "", "Filter by HTTP method")
	pattern := fs.String("pattern", "", "Filter routes by pattern")
	showMiddleware := fs.Bool("middleware", false, "Show middleware chains")
	group := fs.String("group", "", "Filter by route group")
	sortBy := fs.String("sort", "path", "Sort by: path, method, group")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get absolute directory path
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return output.NewFormatter().Error(fmt.Sprintf("directory not found: %s", absDir), 1)
	}

	// Analyze routes
	routes, err := routeanalyzer.AnalyzeRoutes(absDir, routeanalyzer.AnalyzeOptions{
		Method:         *method,
		Pattern:        *pattern,
		ShowMiddleware: *showMiddleware,
		Group:          *group,
		SortBy:         *sortBy,
	})
	if err != nil {
		return output.NewFormatter().Error(fmt.Sprintf("failed to analyze routes: %v", err), 1)
	}

	result := map[string]any{
		"routes": routes.Routes,
		"total":  routes.Total,
	}

	if *showMiddleware {
		result["middleware_summary"] = routes.MiddlewareSummary
	}

	return output.NewFormatter().Success("Routes analyzed successfully", result)
}
