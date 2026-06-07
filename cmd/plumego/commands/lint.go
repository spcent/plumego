package commands

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LintCmd checks user projects for plumego anti-patterns.
type LintCmd struct{}

func (c *LintCmd) Name() string { return "lint" }
func (c *LintCmd) Short() string {
	return "Check project for plumego anti-patterns and boundary violations"
}

type lintViolation struct {
	File    string `json:"file" yaml:"file"`
	Line    int    `json:"line" yaml:"line"`
	Check   string `json:"check" yaml:"check"`
	Level   string `json:"level" yaml:"level"` // error, warning
	Message string `json:"message" yaml:"message"`
}

type lintResult struct {
	Status     string          `json:"status" yaml:"status"` // clean, violations
	Dir        string          `json:"dir" yaml:"dir"`
	Violations []lintViolation `json:"violations" yaml:"violations"`
	Stats      map[string]int  `json:"stats" yaml:"stats"`
}

func (c *LintCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("lint", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", ".", "Project directory")
	checks := fs.String("check", "all", "Checks to run: boundaries,globals,init-effects,handlers,all")
	strict := fs.Bool("strict", false, "Treat warnings as errors")

	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}
	if len(positionals) > 0 {
		return ctx.Out.Error(fmt.Sprintf("unexpected arguments: %v", positionals), 1)
	}

	absDir, err := resolveDir(*dir)
	if err != nil {
		return ctx.Out.Error(err.Error(), 1)
	}

	enabled := parseCheckList(*checks)

	var violations []lintViolation

	if enabled["globals"] || enabled["all"] {
		v, err := checkGlobals(absDir)
		if err != nil {
			return ctx.Out.Error(fmt.Sprintf("globals check: %v", err), 1)
		}
		violations = append(violations, v...)
	}

	if enabled["init-effects"] || enabled["all"] {
		v, err := checkInitSideEffects(absDir)
		if err != nil {
			return ctx.Out.Error(fmt.Sprintf("init-effects check: %v", err), 1)
		}
		violations = append(violations, v...)
	}

	if enabled["handlers"] || enabled["all"] {
		v, err := checkHandlerSignatures(absDir)
		if err != nil {
			return ctx.Out.Error(fmt.Sprintf("handlers check: %v", err), 1)
		}
		violations = append(violations, v...)
	}

	if enabled["boundaries"] || enabled["all"] {
		v, err := checkBoundaries(absDir)
		if err != nil {
			return ctx.Out.Error(fmt.Sprintf("boundaries check: %v", err), 1)
		}
		violations = append(violations, v...)
	}

	stats := map[string]int{"error": 0, "warning": 0}
	for _, v := range violations {
		stats[v.Level]++
	}

	status := "clean"
	if len(violations) > 0 {
		status = "violations"
	}

	result := lintResult{
		Status:     status,
		Dir:        absDir,
		Violations: violations,
		Stats:      stats,
	}
	if result.Violations == nil {
		result.Violations = []lintViolation{}
	}

	hasErrors := stats["error"] > 0
	hasWarnings := stats["warning"] > 0
	failOnWarning := *strict && hasWarnings

	switch {
	case status == "clean":
		return ctx.Out.Success("No lint violations found", result)
	case hasErrors || failOnWarning:
		return ctx.Out.Error(fmt.Sprintf("%d error(s), %d warning(s)", stats["error"], stats["warning"]), 1, result)
	default:
		return ctx.Out.Warning(fmt.Sprintf("%d warning(s) (use --strict to fail on warnings)", stats["warning"]), 1, result)
	}
}

func parseCheckList(s string) map[string]bool {
	m := make(map[string]bool)
	for _, part := range strings.Split(s, ",") {
		m[strings.TrimSpace(part)] = true
	}
	return m
}

// checkGlobals detects mutable package-level variables in handler and middleware packages.
// plumego rule: no hidden globals (AGENTS.md §2).
func checkGlobals(root string) ([]lintViolation, error) {
	var violations []lintViolation

	err := walkGoFiles(root, func(path string, fset *token.FileSet, f *ast.File) {
		// Only flag packages that look like handlers or middleware.
		if !isHandlerOrMiddlewarePkg(filepath.Dir(path)) {
			return
		}
		for _, decl := range f.Decls {
			gd, ok := decl.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, spec := range gd.Specs {
				vs, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for _, name := range vs.Names {
					if name.Name == "_" {
						continue
					}
					pos := fset.Position(name.Pos())
					violations = append(violations, lintViolation{
						File:    relPath(root, pos.Filename),
						Line:    pos.Line,
						Check:   "globals",
						Level:   "error",
						Message: fmt.Sprintf("package-level variable %q in handler/middleware package (plumego: no hidden globals)", name.Name),
					})
				}
			}
		}
	})
	return violations, err
}

// checkInitSideEffects detects init() functions that register globals or start goroutines.
// plumego rule: no init() registration (AGENTS.md §2).
func checkInitSideEffects(root string) ([]lintViolation, error) {
	var violations []lintViolation

	err := walkGoFiles(root, func(path string, fset *token.FileSet, f *ast.File) {
		// Skip main packages — they may legitimately use init().
		if f.Name.Name == "main" {
			return
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "init" || fn.Recv != nil {
				continue
			}
			if fn.Body == nil || len(fn.Body.List) == 0 {
				continue
			}
			pos := fset.Position(fn.Pos())
			// Inspect body for suspicious patterns: go statements, assignments to package vars.
			hasSideEffect := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch n.(type) {
				case *ast.GoStmt:
					hasSideEffect = true
				case *ast.AssignStmt:
					hasSideEffect = true
				case *ast.SendStmt:
					hasSideEffect = true
				}
				return !hasSideEffect
			})
			if hasSideEffect {
				violations = append(violations, lintViolation{
					File:    relPath(root, pos.Filename),
					Line:    pos.Line,
					Check:   "init-effects",
					Level:   "error",
					Message: "init() with side effects (goroutine launch, assignment, or channel send) — plumego prohibits init() registration",
				})
			}
		}
	})
	return violations, err
}

// checkHandlerSignatures warns about handler functions not using net/http's canonical shape.
// plumego canonical: func(http.ResponseWriter, *http.Request).
func checkHandlerSignatures(root string) ([]lintViolation, error) {
	var violations []lintViolation

	err := walkGoFiles(root, func(path string, fset *token.FileSet, f *ast.File) {
		if !isHandlerOrMiddlewarePkg(filepath.Dir(path)) {
			return
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil {
				continue
			}
			name := fn.Name.Name
			if !strings.HasSuffix(name, "Handler") && !strings.HasSuffix(name, "Handle") {
				continue
			}
			if fn.Type.Params == nil {
				continue
			}
			params := fn.Type.Params.List
			if len(params) != 2 {
				pos := fset.Position(fn.Pos())
				violations = append(violations, lintViolation{
					File:    relPath(root, pos.Filename),
					Line:    pos.Line,
					Check:   "handlers",
					Level:   "warning",
					Message: fmt.Sprintf("function %q looks like a handler but does not have (http.ResponseWriter, *http.Request) signature", name),
				})
			}
		}
	})
	return violations, err
}

// checkBoundaries detects cross-layer import violations in the project.
// Flags middleware packages that import from service or domain packages.
func checkBoundaries(root string) ([]lintViolation, error) {
	var violations []lintViolation

	err := walkGoFiles(root, func(path string, fset *token.FileSet, f *ast.File) {
		dir := filepath.Dir(path)
		pkgClass := classifyPackage(dir)
		if pkgClass == "" {
			return
		}
		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			violation := boundaryViolation(pkgClass, importPath)
			if violation == "" {
				continue
			}
			pos := fset.Position(imp.Pos())
			violations = append(violations, lintViolation{
				File:    relPath(root, pos.Filename),
				Line:    pos.Line,
				Check:   "boundaries",
				Level:   "error",
				Message: fmt.Sprintf("%s package imports %q — %s", pkgClass, importPath, violation),
			})
		}
	})
	return violations, err
}

// boundaryViolation returns a description of the boundary rule being violated, or "".
func boundaryViolation(pkgClass, importPath string) string {
	parts := strings.Split(importPath, "/")
	lastPart := ""
	if len(parts) > 0 {
		lastPart = parts[len(parts)-1]
	}

	switch pkgClass {
	case "middleware":
		// Middleware must be transport-only: no service or domain imports.
		if lastPart == "service" || lastPart == "services" {
			return "middleware must be transport-only; move business logic to a domain service"
		}
		if lastPart == "domain" {
			return "middleware must be transport-only; domain types belong in the service layer"
		}
		if lastPart == "dto" || lastPart == "dtos" {
			return "middleware must not assemble DTOs; delegate to handlers"
		}
	case "handler":
		// Handlers should not import infrastructure packages directly.
		if lastPart == "store" || lastPart == "stores" || lastPart == "repository" || lastPart == "repo" {
			return "handler should depend on a service interface, not infrastructure directly"
		}
	}
	return ""
}

// classifyPackage returns "middleware" or "handler" based on directory name, or "" if unclassified.
func classifyPackage(dir string) string {
	base := filepath.Base(dir)
	switch {
	case base == "middleware" || strings.HasPrefix(base, "middleware"):
		return "middleware"
	case base == "handler" || base == "handlers":
		return "handler"
	}
	return ""
}

// isHandlerOrMiddlewarePkg reports whether dir is likely a handler or middleware package.
func isHandlerOrMiddlewarePkg(dir string) bool {
	return classifyPackage(dir) != ""
}

// walkGoFiles walks all non-test .go files under root and calls fn for each parsed file.
func walkGoFiles(root string, fn func(path string, fset *token.FileSet, f *ast.File)) error {
	fset := token.NewFileSet()
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", "vendor", "node_modules", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			// Skip unparseable files; they may have build-tag exclusions.
			return nil
		}
		fn(path, fset, f)
		return nil
	})
}

// relPath returns path relative to root, or path itself if rel fails.
func relPath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}
