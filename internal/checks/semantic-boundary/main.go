// Package main implements the semantic-boundary checker.
//
// It detects two anti-patterns in extension package source (x/* production
// .go files, excluding _test.go):
//
//  1. init() side-effect registration — extension packages must not use init()
//     because it creates hidden, order-dependent global state.
//
//  2. Panic-only constructors — exported New* functions that can only signal
//     failure via panic instead of returning an error.
//
// Both baselines live under specs/check-baseline/:
//
//	semantic-boundary-init.txt
//	semantic-boundary-panic-ctors.txt
//
// Each baseline file contains one "file:symbol" entry per line (blank lines
// and lines starting with # are ignored).
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spcent/plumego/internal/checks/checkutil"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		fatalf("resolve working directory: %v", err)
	}

	initBaseline, err := checkutil.ReadBaseline(
		filepath.Join(repoRoot, "specs", "check-baseline", "semantic-boundary-init.txt"),
	)
	if err != nil {
		fatalf("read init baseline: %v", err)
	}
	panicBaseline, err := checkutil.ReadBaseline(
		filepath.Join(repoRoot, "specs", "check-baseline", "semantic-boundary-panic-ctors.txt"),
	)
	if err != nil {
		fatalf("read panic-ctors baseline: %v", err)
	}

	initViolations, panicViolations, err := scanExtensionPackages(repoRoot, initBaseline, panicBaseline)
	if err != nil {
		fatalf("scan extension packages: %v", err)
	}

	var all []string
	all = append(all, initViolations...)
	all = append(all, panicViolations...)
	if len(all) == 0 {
		return
	}

	fmt.Fprint(os.Stderr, checkutil.FormatViolations("semantic-boundary", all))
	os.Exit(1)
}

// scanExtensionPackages walks x/* and checks each non-test .go file for
// init() functions and panic-only New* constructors.
func scanExtensionPackages(repoRoot string, initBaseline, panicBaseline map[string]struct{}) (initViols, panicViols []string, err error) {
	xDir := filepath.Join(repoRoot, "x")
	fset := token.NewFileSet()

	err = filepath.WalkDir(xDir, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}

		f, parseErr := parser.ParseFile(fset, path, nil, 0)
		if parseErr != nil {
			return nil // skip unparseable files
		}

		rel, _ := filepath.Rel(repoRoot, path)
		rel = filepath.ToSlash(rel)

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			name := fn.Name.Name

			// Check 1: init() side-effect registration.
			if name == "init" && fn.Recv == nil {
				key := rel + ":init"
				if _, allowed := initBaseline[key]; !allowed {
					initViols = append(initViols, rel+": init() function — extension packages must not use init() for side-effect registration (FIX: move registration to an explicit constructor or bootstrap call)")
				}
			}

			// Check 2: panic-only New* constructors.
			if !strings.HasPrefix(name, "New") || fn.Recv != nil {
				continue
			}
			if !isExported(name) {
				continue
			}
			if hasErrorReturn(fn) {
				continue
			}
			if bodyContainsPanic(fn) {
				key := rel + ":" + name
				if _, allowed := panicBaseline[key]; !allowed {
					panicViols = append(panicViols, rel+": "+name+"() panics without returning an error — use New"+strings.TrimPrefix(name, "New")+"(...) (T, error) instead (FIX: add error return and replace panic with return nil, err)")
				}
			}
		}
		return nil
	})
	if err != nil && os.IsNotExist(err) {
		return nil, nil, nil
	}

	sort.Strings(initViols)
	sort.Strings(panicViols)
	return initViols, panicViols, err
}

// hasErrorReturn reports whether fn's result list contains an error type.
func hasErrorReturn(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, field := range fn.Type.Results.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "error" {
			return true
		}
	}
	return false
}

// bodyContainsPanic reports whether the function body contains a panic() call.
func bodyContainsPanic(fn *ast.FuncDecl) bool {
	if fn.Body == nil {
		return false
	}
	found := false
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if found {
			return false
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := call.Fun.(*ast.Ident)
		if ok && ident.Name == "panic" {
			found = true
			return false
		}
		return true
	})
	return found
}

func isExported(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= 'A' && name[0] <= 'Z'
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "semantic-boundary: "+format+"\n", args...)
	os.Exit(1)
}
