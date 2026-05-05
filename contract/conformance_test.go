package contract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestExternalCodeUsesAPIErrorBuilder(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	var violations []string
	fset := token.NewFileSet()
	err = filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".codex", "node_modules", "vendor":
				return filepath.SkipDir
			}
			if d.Name() == "contract" && filepath.Dir(path) == repoRoot {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		contractNames := contractImportNames(file)
		if len(contractNames) == 0 {
			return nil
		}

		file, err = parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(file, func(node ast.Node) bool {
			lit, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}
			sel, ok := lit.Type.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "APIError" {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, ok := contractNames[ident.Name]; !ok {
				return true
			}
			pos := fset.Position(lit.Pos())
			rel, err := filepath.Rel(repoRoot, pos.Filename)
			if err != nil {
				rel = pos.Filename
			}
			violations = append(violations, filepath.ToSlash(rel)+":"+strconv.Itoa(pos.Line))
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan external APIError literals: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("external code must use contract.NewErrorBuilder instead of contract.APIError literals:\n%s", strings.Join(violations, "\n"))
	}
}

func contractImportNames(file *ast.File) map[string]struct{} {
	names := map[string]struct{}{}
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || path != "github.com/spcent/plumego/contract" {
			continue
		}
		name := "contract"
		if imp.Name != nil {
			if imp.Name.Name == "." || imp.Name.Name == "_" {
				continue
			}
			name = imp.Name.Name
		}
		names[name] = struct{}{}
	}
	return names
}
