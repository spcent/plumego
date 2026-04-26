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

func TestStaticChecksConfigDefaults(t *testing.T) {
	repoRoot := repositoryRoot(t)
	middlewareRoot := filepath.Join(repoRoot, "middleware")

	type configType struct {
		key  string
		path string
		name string
	}

	allowedInlineDefaults := map[string]string{
		"compression.GzipConfig": "stable package-specific Gzip constructor applies defaults inline",
		"cors.CORSOptions":       "stable CORSOptions constructor applies defaults through withDefaults",
		"timeout.TimeoutConfig":  "stable package-specific Timeout constructor applies defaults inline",
	}
	configTypes := []configType{}
	defaultMethods := map[string]bool{}
	defaultFuncs := map[string]map[string]bool{}

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
		pkg := fileNode.Name.Name

		for _, decl := range fileNode.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok || !typeSpec.Name.IsExported() || !isConfigTypeName(typeSpec.Name.Name) {
						continue
					}
					if _, ok := typeSpec.Type.(*ast.StructType); !ok {
						continue
					}
					configTypes = append(configTypes, configType{
						key:  pkg + "." + typeSpec.Name.Name,
						path: path,
						name: typeSpec.Name.Name,
					})
				}
			case *ast.FuncDecl:
				if d.Recv == nil {
					if defaultFuncs[pkg] == nil {
						defaultFuncs[pkg] = map[string]bool{}
					}
					defaultFuncs[pkg][d.Name.Name] = true
					continue
				}
				receiver := receiverTypeName(d.Recv)
				if receiver == "" {
					continue
				}
				if d.Name.Name == "WithDefaults" || d.Name.Name == "Default" {
					defaultMethods[pkg+"."+receiver] = true
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk middleware directory: %v", err)
	}

	for _, cfg := range configTypes {
		pkg := strings.Split(cfg.key, ".")[0]
		if defaultMethods[cfg.key] || defaultFuncs[pkg]["Default"+cfg.name] {
			continue
		}
		if reason, ok := allowedInlineDefaults[cfg.key]; ok {
			t.Logf("allowing inline defaults for %s: %s", cfg.key, reason)
			continue
		}
		t.Fatalf("exported middleware config %s in %s lacks Default%s, Default, or WithDefaults", cfg.key, cfg.path, cfg.name)
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

func isConfigTypeName(name string) bool {
	return strings.HasSuffix(name, "Config") || strings.HasSuffix(name, "Options")
}

func receiverTypeName(recv *ast.FieldList) string {
	if recv == nil || len(recv.List) == 0 {
		return ""
	}
	expr := recv.List[0].Type
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	ident, ok := expr.(*ast.Ident)
	if !ok {
		return ""
	}
	return ident.Name
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
