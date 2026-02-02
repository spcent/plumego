package routeanalyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Route represents a single HTTP route
type Route struct {
	Method     string   `json:"method" yaml:"method"`
	Path       string   `json:"path" yaml:"path"`
	Handler    string   `json:"handler,omitempty" yaml:"handler,omitempty"`
	Group      string   `json:"group,omitempty" yaml:"group,omitempty"`
	Middleware []string `json:"middleware,omitempty" yaml:"middleware,omitempty"`
	File       string   `json:"file,omitempty" yaml:"file,omitempty"`
	Line       int      `json:"line,omitempty" yaml:"line,omitempty"`
}

// AnalyzeResult contains the analysis result
type AnalyzeResult struct {
	Routes            []Route           `json:"routes" yaml:"routes"`
	Total             int               `json:"total" yaml:"total"`
	MiddlewareSummary map[string]int    `json:"middleware_summary,omitempty" yaml:"middleware_summary,omitempty"`
}

// AnalyzeOptions contains options for route analysis
type AnalyzeOptions struct {
	Method         string
	Pattern        string
	ShowMiddleware bool
	Group          string
	SortBy         string
}

// AnalyzeRoutes analyzes routes in the given directory
func AnalyzeRoutes(dir string, opts AnalyzeOptions) (*AnalyzeResult, error) {
	routes := []Route{}
	middlewareCount := make(map[string]int)

	// Walk through Go files
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, node_modules, .git
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == "node_modules" || name == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process .go files
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Parse the file
		fileRoutes, err := parseFileForRoutes(path, dir)
		if err != nil {
			// Silently skip files that can't be parsed
			return nil
		}

		routes = append(routes, fileRoutes...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Filter routes
	filtered := []Route{}
	for _, route := range routes {
		// Filter by method
		if opts.Method != "" && route.Method != opts.Method {
			continue
		}

		// Filter by pattern
		if opts.Pattern != "" && !strings.Contains(route.Path, opts.Pattern) {
			continue
		}

		// Filter by group
		if opts.Group != "" && route.Group != opts.Group {
			continue
		}

		filtered = append(filtered, route)

		// Count middleware
		if opts.ShowMiddleware {
			for _, mw := range route.Middleware {
				middlewareCount[mw]++
			}
		}
	}

	// Sort routes
	sortRoutes(filtered, opts.SortBy)

	result := &AnalyzeResult{
		Routes: filtered,
		Total:  len(filtered),
	}

	if opts.ShowMiddleware && len(middlewareCount) > 0 {
		result.MiddlewareSummary = middlewareCount
	}

	return result, nil
}

func parseFileForRoutes(path, baseDir string) ([]Route, error) {
	routes := []Route{}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	relPath, _ := filepath.Rel(baseDir, path)

	// Walk the AST to find route registrations
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for method calls like r.GET(), app.Post(), etc.
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				method := sel.Sel.Name

				// Check if it's an HTTP method
				httpMethod := strings.ToUpper(method)
				if isHTTPMethod(httpMethod) {
					route := extractRoute(call, httpMethod, fset, relPath)
					if route != nil {
						routes = append(routes, *route)
					}
				}
			}
		}
		return true
	})

	return routes, nil
}

func extractRoute(call *ast.CallExpr, method string, fset *token.FileSet, file string) *Route {
	if len(call.Args) < 2 {
		return nil
	}

	// First argument should be the path
	var path string
	if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
		path = strings.Trim(lit.Value, `"`)
	} else {
		return nil
	}

	// Second argument is the handler
	var handler string
	switch h := call.Args[1].(type) {
	case *ast.Ident:
		handler = h.Name
	case *ast.SelectorExpr:
		if x, ok := h.X.(*ast.Ident); ok {
			handler = x.Name + "." + h.Sel.Name
		} else {
			handler = h.Sel.Name
		}
	case *ast.FuncLit:
		handler = "anonymous"
	default:
		handler = "unknown"
	}

	pos := fset.Position(call.Pos())

	return &Route{
		Method:  method,
		Path:    path,
		Handler: handler,
		File:    file,
		Line:    pos.Line,
	}
}

func isHTTPMethod(method string) bool {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS", "CONNECT", "TRACE"}
	for _, m := range methods {
		if method == m {
			return true
		}
	}
	return false
}

func sortRoutes(routes []Route, sortBy string) {
	switch sortBy {
	case "method":
		sort.Slice(routes, func(i, j int) bool {
			if routes[i].Method == routes[j].Method {
				return routes[i].Path < routes[j].Path
			}
			return routes[i].Method < routes[j].Method
		})
	case "group":
		sort.Slice(routes, func(i, j int) bool {
			if routes[i].Group == routes[j].Group {
				return routes[i].Path < routes[j].Path
			}
			return routes[i].Group < routes[j].Group
		})
	default: // path
		sort.Slice(routes, func(i, j int) bool {
			return routes[i].Path < routes[j].Path
		})
	}
}
