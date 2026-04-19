// fix-test-ctx replaces context.Background() with t.Context() (or b.Context())
// inside test and benchmark functions in all *_test.go files.
//
// Rules:
//   - Replace only inside functions whose parameter list includes *testing.T,
//     *testing.B, or testing.TB. The variable name (always "t" or "b") is
//     preserved across nested scopes introduced by t.Run closures.
//   - Closures that declare their own *testing.T parameter get their own
//     replacement pass; the outer pass stops at that boundary so the inner "t"
//     is used, not the outer one.
//   - Closures without a testing parameter are treated as part of the enclosing
//     scope (goroutines, helpers, etc.), so they use the nearest testing var in
//     scope — which means their context also respects the test lifetime.
//   - Functions with only *testing.M (TestMain) are never modified.
//   - After replacement, the "context" import is removed if it is no longer
//     referenced anywhere in the file.
package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}

	var files, total int
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			switch filepath.Base(path) {
			case "vendor", ".git", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		n, ferr := processFile(path)
		if ferr != nil {
			fmt.Fprintf(os.Stderr, "error: %s: %v\n", path, ferr)
			return nil
		}
		if n > 0 {
			files++
			total += n
			fmt.Printf("  %s: %d\n", path, n)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "walk: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\n%d replacements in %d files\n", total, files)
}

func processFile(path string) (int, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	n := replaceInFile(f)
	if n == 0 {
		return 0, nil
	}

	if !usesContextPackage(f) {
		removeContextImport(f)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return 0, fmt.Errorf("format: %w", err)
	}
	return n, os.WriteFile(path, buf.Bytes(), 0644)
}

// replaceInFile drives the replacement for one parsed file.
// It walks top-level declarations and function literals, calling replaceInScope
// for each function that has a qualifying testing parameter.
func replaceInFile(f *ast.File) int {
	total := 0
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Body != nil {
				if v := testingVarName(fn.Type); v != "" {
					total += replaceInScope(fn.Body, v)
				}
			}
			return true // descend to find nested FuncLit
		case *ast.FuncLit:
			if v := testingVarName(fn.Type); v != "" {
				total += replaceInScope(fn.Body, v)
			}
			return true
		}
		return true
	})
	return total
}

// replaceInScope replaces context.Background() → varName.Context() inside root,
// stopping at nested FuncLit nodes that declare their own testing parameter
// (those are handled by a separate replaceInFile pass).
func replaceInScope(root ast.Node, varName string) int {
	count := 0
	ast.Inspect(root, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if n == root {
			return true
		}
		// Stop at nested closures that own their testing var.
		if fl, ok := n.(*ast.FuncLit); ok {
			if testingVarName(fl.Type) != "" {
				return false
			}
			return true // no own testing var → belongs to this scope
		}
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		if isContextBackground(call) {
			call.Fun = &ast.SelectorExpr{
				X:   ast.NewIdent(varName),
				Sel: ast.NewIdent("Context"),
			}
			count++
			return false
		}
		return true
	})
	return count
}

// testingVarName returns the first parameter name whose type is *testing.T,
// *testing.B, or testing.TB. Returns "" if none found.
func testingVarName(ft *ast.FuncType) string {
	if ft == nil || ft.Params == nil {
		return ""
	}
	for _, field := range ft.Params.List {
		switch baseTestingType(field.Type) {
		case "T", "B", "TB":
			if len(field.Names) > 0 {
				return field.Names[0].Name
			}
		}
	}
	return ""
}

// baseTestingType resolves *testing.X or testing.X to "X", else "".
func baseTestingType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return baseTestingType(t.X)
	case *ast.SelectorExpr:
		if id, ok := t.X.(*ast.Ident); ok && id.Name == "testing" {
			return t.Sel.Name
		}
	}
	return ""
}

// isContextBackground reports whether call is context.Background().
func isContextBackground(call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	id, ok := sel.X.(*ast.Ident)
	return ok && id.Name == "context" && sel.Sel.Name == "Background"
}

// usesContextPackage reports whether any context.X selector remains in the file.
func usesContextPackage(f *ast.File) bool {
	found := false
	ast.Inspect(f, func(n ast.Node) bool {
		if found {
			return false
		}
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		id, ok := sel.X.(*ast.Ident)
		if ok && id.Name == "context" {
			found = true
		}
		return !found
	})
	return found
}

// removeContextImport removes the "context" import specifier from the file.
func removeContextImport(f *ast.File) {
	for i, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.IMPORT {
			continue
		}
		for j, spec := range gen.Specs {
			imp, ok := spec.(*ast.ImportSpec)
			if !ok {
				continue
			}
			if strings.Trim(imp.Path.Value, `"`) == "context" {
				gen.Specs = append(gen.Specs[:j], gen.Specs[j+1:]...)
				if len(gen.Specs) == 0 {
					f.Decls = append(f.Decls[:i], f.Decls[i+1:]...)
				}
				return
			}
		}
	}
}
