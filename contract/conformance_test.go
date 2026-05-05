package contract

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
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
	for _, root := range conformanceScanRoots(repoRoot) {
		rootPath := filepath.Join(repoRoot, root)
		info, err := os.Stat(rootPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
		if !info.IsDir() {
			return nil
		}

		err = filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				switch d.Name() {
				case ".git", ".codex", "node_modules", "vendor":
					return filepath.SkipDir
				}
				if filepath.Base(path) == "contract" && filepath.Dir(path) == repoRoot {
					return filepath.SkipDir
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
		if err != nil {
			return err
		}
	}
	return nil
}

func conformanceScanRoots(repoRoot string) []string {
	spec, err := os.ReadFile(filepath.Join(repoRoot, "specs", "repo.yaml"))
	if err != nil {
		return []string{"cmd", "core", "health", "internal", "log", "metrics", "middleware", "reference", "router", "security", "store", "x"}
	}

	roots := map[string]struct{}{}
	inPathList := false
	for _, line := range strings.Split(string(spec), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasSuffix(trimmed, "paths:") {
			inPathList = true
			continue
		}
		if inPathList && !strings.HasPrefix(trimmed, "- ") && trimmed != "" {
			inPathList = false
		}
		if !inPathList || !strings.HasPrefix(trimmed, "- ") {
			continue
		}
		path := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		path = strings.Trim(path, `"'`)
		if path == "" {
			continue
		}
		root := strings.Split(path, "/")[0]
		if root != "" && root != "contract" {
			roots[root] = struct{}{}
		}
	}
	for _, root := range []string{"cmd", "internal", "reference", "x"} {
		roots[root] = struct{}{}
	}

	out := make([]string, 0, len(roots))
	for root := range roots {
		out = append(out, root)
	}
	sort.Strings(out)
	return out
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

func TestExternalValidateStructUsageIsAllowlisted(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	allowed := map[string]int{
		"reference/workerfleet/internal/handler/worker_heartbeat.go#Handler.HeartbeatWorker": 2,
		"reference/workerfleet/internal/handler/worker_register.go#Handler.RegisterWorker":   1,
		"x/messaging/api.go#Service.HandleBatchSend":                                         1,
		"x/messaging/api.go#Service.HandleSend":                                              1,
		"x/ops/ops.go#Handler.handleQueueReplay":                                             1,
	}
	actual := map[string]int{}

	fset := token.NewFileSet()
	err = walkExternalContractGoFiles(repoRoot, fset, func(path string, file *ast.File, contractNames map[string]struct{}) error {
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			rel = path
		}
		rel = filepath.ToSlash(rel)

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			key := rel + "#" + funcDeclName(fn)
			ast.Inspect(fn.Body, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "ValidateStruct" {
					return true
				}
				ident, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				if _, ok := contractNames[ident.Name]; !ok {
					return true
				}
				actual[key]++
				return true
			})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan external ValidateStruct usage: %v", err)
	}

	var violations []string
	for path, count := range actual {
		if allowed[path] != count {
			violations = append(violations, path+" uses contract.ValidateStruct "+strconv.Itoa(count)+" time(s); expected "+strconv.Itoa(allowed[path]))
		}
	}
	for path, count := range allowed {
		if actual[path] != count {
			violations = append(violations, path+" uses contract.ValidateStruct "+strconv.Itoa(actual[path])+" time(s); expected "+strconv.Itoa(count))
		}
	}
	if len(violations) > 0 {
		t.Fatalf("external contract.ValidateStruct usage must stay on the stable compatibility allowlist:\n%s", strings.Join(violations, "\n"))
	}
}

func TestExternalWriteResponseUsesSuccessStatuses(t *testing.T) {
	repoRoot, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}

	var violations []string
	fset := token.NewFileSet()
	err = walkExternalContractGoFiles(repoRoot, fset, func(path string, file *ast.File, contractNames map[string]struct{}) error {
		httpNames := packageImportNames(file, "net/http")
		if len(httpNames) == 0 {
			return nil
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) < 3 {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok || sel.Sel.Name != "WriteResponse" {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, ok := contractNames[ident.Name]; !ok {
				return true
			}
			if statusIsKnownNonSuccess(call.Args[2], httpNames) {
				pos := fset.Position(call.Args[2].Pos())
				rel, err := filepath.Rel(repoRoot, pos.Filename)
				if err != nil {
					rel = pos.Filename
				}
				violations = append(violations, filepath.ToSlash(rel)+":"+strconv.Itoa(pos.Line))
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatalf("scan external WriteResponse statuses: %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("external contract.WriteResponse calls must use known 2xx status literals/selectors; use WriteError for errors:\n%s", strings.Join(violations, "\n"))
	}
}

func statusIsKnownNonSuccess(expr ast.Expr, httpNames map[string]struct{}) bool {
	if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.INT {
		status, err := strconv.Atoi(lit.Value)
		return err == nil && (status < 200 || status > 299)
	}
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	if _, ok := httpNames[ident.Name]; !ok {
		return false
	}
	return !strings.HasPrefix(sel.Sel.Name, "StatusOK") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusCreated") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusAccepted") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusNonAuthoritativeInfo") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusNoContent") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusResetContent") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusPartialContent") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusMultiStatus") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusAlreadyReported") &&
		!strings.HasPrefix(sel.Sel.Name, "StatusIMUsed")
}

func funcDeclName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return fn.Name.Name
	}
	return typeExprName(fn.Recv.List[0].Type) + "." + fn.Name.Name
}

func typeExprName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeExprName(t.X)
	case *ast.IndexExpr:
		return typeExprName(t.X)
	case *ast.IndexListExpr:
		return typeExprName(t.X)
	default:
		return "unknown"
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
	return packageImportNames(file, "github.com/spcent/plumego/contract")
}

func packageImportNames(file *ast.File, importPath string) map[string]struct{} {
	names := map[string]struct{}{}
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil || path != importPath {
			continue
		}
		name := filepath.Base(importPath)
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
