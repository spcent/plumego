package conformance_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestStaticChecksForbiddenPatterns(t *testing.T) {
	repoRoot := repositoryRoot(t)
	middlewareRoot := filepath.Join(repoRoot, "middleware")

	fset := token.NewFileSet()
	err := filepath.WalkDir(middlewareRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileNode, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}

		for _, decl := range fileNode.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			constructor := functionReturnsMiddleware(fn)

			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				if isSelectorCall(call, "context", "WithValue") && bindsBusinessDTO(call) {
					t.Fatalf("forbidden DTO binding into request context in %s (%s)", path, fn.Name.Name)
				}

				if isSelectorCall(call, "http", "Error") {
					t.Fatalf("forbidden direct http.Error in middleware JSON path: %s (%s)", path, fn.Name.Name)
				}

				if constructor && (isSelectorCall(call, "os", "Getenv") || isSelectorCall(call, "os", "LookupEnv")) {
					t.Fatalf("forbidden env read in middleware constructor: %s (%s)", path, fn.Name.Name)
				}
				return true
			})
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk middleware directory: %v", err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(currentFile), "..", ".."))
}

func functionReturnsMiddleware(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	for _, result := range fn.Type.Results.List {
		if exprContainsMiddlewareType(result.Type) {
			return true
		}
	}
	return false
}

func exprContainsMiddlewareType(expr ast.Expr) bool {
	switch x := expr.(type) {
	case *ast.Ident:
		return x.Name == "Middleware"
	case *ast.SelectorExpr:
		if x.Sel != nil && x.Sel.Name == "Middleware" {
			return true
		}
		return exprContainsMiddlewareType(x.X)
	case *ast.StarExpr:
		return exprContainsMiddlewareType(x.X)
	case *ast.ArrayType:
		return exprContainsMiddlewareType(x.Elt)
	default:
		return false
	}
}

func isSelectorCall(call *ast.CallExpr, pkg, fn string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel == nil || sel.Sel.Name != fn {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	return ok && ident.Name == pkg
}

func bindsBusinessDTO(call *ast.CallExpr) bool {
	if len(call.Args) < 3 {
		return false
	}
	return isBusinessPayloadExpression(call.Args[2])
}

func isBusinessPayloadExpression(expr ast.Expr) bool {
	switch v := expr.(type) {
	case *ast.CompositeLit:
		return true
	case *ast.UnaryExpr:
		if v.Op == token.AND {
			return isBusinessPayloadExpression(v.X)
		}
	case *ast.Ident:
		name := strings.ToLower(v.Name)
		for _, part := range []string{"dto", "payload", "request", "body", "command"} {
			if strings.Contains(name, part) {
				return true
			}
		}
	case *ast.CallExpr:
		if sel, ok := v.Fun.(*ast.SelectorExpr); ok && sel.Sel != nil {
			name := strings.ToLower(sel.Sel.Name)
			return strings.Contains(name, "decode") || strings.Contains(name, "bind")
		}
	}
	return false
}
