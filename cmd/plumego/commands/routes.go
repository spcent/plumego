package commands

import (
	"flag"
	"fmt"
	"io"

	"github.com/spcent/plumego/cmd/plumego/internal/routeanalyzer"
)

type RoutesCmd struct{}

func (c *RoutesCmd) Name() string  { return "routes" }
func (c *RoutesCmd) Short() string { return "Inspect registered routes" }

func (c *RoutesCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("routes", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", ".", "Project directory")
	method := fs.String("method", "", "Filter by HTTP method")
	pattern := fs.String("pattern", "", "Filter routes by pattern")
	showMiddleware := fs.Bool("middleware", false, "Show middleware chains")
	group := fs.String("group", "", "Filter by route group")
	sortBy := fs.String("sort", "path", "Sort by: path, method, group")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	absDir, err := resolveDir(*dir)
	if err != nil {
		return ctx.Out.Error(err.Error(), 1)
	}

	routes, err := routeanalyzer.AnalyzeRoutes(absDir, routeanalyzer.AnalyzeOptions{
		Method:         *method,
		Pattern:        *pattern,
		ShowMiddleware: *showMiddleware,
		Group:          *group,
		SortBy:         *sortBy,
	})
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to analyze routes: %v", err), 1)
	}

	result := map[string]any{
		"routes": routes.Routes,
		"total":  routes.Total,
	}

	if *showMiddleware {
		result["middleware_summary"] = routes.MiddlewareSummary
	}

	return ctx.Out.Success("Routes analyzed successfully", result)
}
