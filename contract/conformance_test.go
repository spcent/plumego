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
	err = walkExternalContractGoFiles(repoRoot, fset, func(path string, file *ast.File, contractNames map[string]struct{}) error {
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

func walkExternalContractGoFiles(repoRoot string, fset *token.FileSet, fn func(path string, file *ast.File, contractNames map[string]struct{}) error) error {
	return filepath.WalkDir(repoRoot, func(path string, d os.DirEntry, err error) error {
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
			if filepath.Dir(path) == repoRoot {
				switch d.Name() {
				case "cmd", "core", "health", "internal", "log", "metrics", "middleware", "reference", "router", "security", "store", "x":
				default:
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(src), `"github.com/spcent/plumego/contract"`) {
			return nil
		}

		file, err := parser.ParseFile(fset, path, src, 0)
		if err != nil {
			return err
		}
		contractNames := contractImportNames(file)
		if len(contractNames) == 0 {
			return nil
		}

		return fn(path, file, contractNames)
	})
}

func TestExternalTypedErrorsUseCanonicalContractCodes(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	allowedCodesForType := map[string]map[string]struct{}{
		"TypeValidation": {
			"CodeValidationError": {},
			"CodeBadRequest":      {},
			"CodeInvalidRequest":  {},
			"CodeInvalidJSON":     {},
			"CodeInvalidQuery":    {},
		},
		"TypeRequired":         {"CodeRequired": {}},
		"TypeInvalidFormat":    {"CodeInvalidFormat": {}},
		"TypeOutOfRange":       {"CodeOutOfRange": {}},
		"TypeDuplicate":        {"CodeDuplicate": {}},
		"TypeUnauthorized":     {"CodeUnauthorized": {}},
		"TypeForbidden":        {"CodeForbidden": {}},
		"TypeInvalidToken":     {"CodeInvalidToken": {}},
		"TypeExpiredToken":     {"CodeExpiredToken": {}},
		"TypeNotFound":         {"CodeResourceNotFound": {}},
		"TypeConflict":         {"CodeConflict": {}},
		"TypeAlreadyExists":    {"CodeAlreadyExists": {}},
		"TypeGone":             {"CodeGone": {}},
		"TypeInternal":         {"CodeInternalError": {}},
		"TypeUnavailable":      {"CodeUnavailable": {}},
		"TypeTimeout":          {"CodeTimeout": {}},
		"TypeRateLimited":      {"CodeRateLimited": {}},
		"TypeMaintenance":      {"CodeMaintenance": {}},
		"TypeMethodNotAllowed": {"CodeMethodNotAllowed": {}},
		"TypeNotImplemented":   {"CodeNotImplemented": {}},
		"TypeBadGateway":       {"CodeBadGateway": {}},
		"TypeGatewayTimeout":   {"CodeGatewayTimeout": {}},
	}

	var violations []string
	fset := token.NewFileSet()
	err = walkExternalContractGoFiles(repoRoot, fset, func(path string, file *ast.File, contractNames map[string]struct{}) error {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "Build" {
				return true
			}
			chain, ok := contractErrorBuilderChain(call, contractNames)
			if !ok {
				return true
			}

			typeName := ""
			typeIndex := -1
			codeName := ""
			codeIndex := -1
			for i, step := range chain {
				switch step.name {
				case "Type":
					if len(step.args) != 1 {
						continue
					}
					if name, ok := contractSelector(step.args[0], contractNames, "Type"); ok {
						typeName = name
						typeIndex = i
					}
				case "Code":
					if len(step.args) != 1 {
						continue
					}
					if name, ok := contractSelector(step.args[0], contractNames, "Code"); ok {
						codeName = name
						codeIndex = i
					}
				}
			}
			if typeName == "" || codeName == "" || codeIndex < typeIndex {
				return true
			}

			allowedCodes, ok := allowedCodesForType[typeName]
			if !ok {
				return true
			}
			if _, ok := allowedCodes[codeName]; ok {
				return true
			}
			pos := fset.Position(call.Pos())
			rel, err := filepath.Rel(repoRoot, pos.Filename)
			if err != nil {
				rel = pos.Filename
			}
			violations = append(violations, filepath.ToSlash(rel)+":"+strconv.Itoa(pos.Line)+
				" uses contract."+typeName+" with contract."+codeName+"; use a type-compatible contract code or an extension-owned code")
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan external typed error code overrides: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("typed contract errors must not override contract-owned codes across type families:\n%s", strings.Join(violations, "\n"))
	}
}

type errorBuilderStep struct {
	name string
	args []ast.Expr
}

func contractErrorBuilderChain(call *ast.CallExpr, contractNames map[string]struct{}) ([]errorBuilderStep, bool) {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil, false
	}
	if sel.Sel.Name == "NewErrorBuilder" {
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return nil, false
		}
		_, ok = contractNames[ident.Name]
		return nil, ok
	}

	prevCall, ok := sel.X.(*ast.CallExpr)
	if !ok {
		return nil, false
	}
	chain, ok := contractErrorBuilderChain(prevCall, contractNames)
	if !ok {
		return nil, false
	}
	return append(chain, errorBuilderStep{name: sel.Sel.Name, args: call.Args}), true
}

func contractSelector(expr ast.Expr, contractNames map[string]struct{}, prefix string) (string, bool) {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || !strings.HasPrefix(sel.Sel.Name, prefix) {
		return "", false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	if _, ok := contractNames[ident.Name]; !ok {
		return "", false
	}
	return sel.Sel.Name, true
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
