package checkutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindDisallowedImportsHonorsBaselineAndSkipsTests(t *testing.T) {
	repo := t.TempDir()

	writeFile(t, filepath.Join(repo, "core", "core.go"), `package core

import (
	"net/http"
	"github.com/spcent/plumego/internal/httpx"
	"github.com/spcent/plumego/x/tenant/resolve"
)

func Use(_ http.ResponseWriter) {
	_, _ = httpx.ClientIP(nil)
	_ = resolve.Options{}
}
`)
	writeFile(t, filepath.Join(repo, "core", "core_test.go"), `package core

import "github.com/spcent/plumego/x/tenant/resolve"
`)

	violations, err := FindDisallowedImports(repo, map[string]struct{}{
		"core/core.go|github.com/spcent/plumego/internal/httpx": {},
	})
	if err != nil {
		t.Fatalf("FindDisallowedImports: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if got := violations[0]; got != "core/core.go|github.com/spcent/plumego/x/tenant/resolve" {
		t.Fatalf("unexpected violation %q", got)
	}
}

func TestFindMissingModuleManifestsHonorsBaseline(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "core", "module.yaml"), validManifest("core", "stable"))
	writeFile(t, filepath.Join(repo, "router", "module.yaml"), validManifest("router", "stable"))
	writeFile(t, filepath.Join(repo, "contract", "module.yaml"), validManifest("contract", "stable"))
	writeFile(t, filepath.Join(repo, "middleware", "module.yaml"), validManifest("middleware", "stable"))
	writeFile(t, filepath.Join(repo, "security", "module.yaml"), validManifest("security", "stable"))
	writeFile(t, filepath.Join(repo, "store", "module.yaml"), validManifest("store", "stable"))
	writeFile(t, filepath.Join(repo, "health", "module.yaml"), validManifest("health", "stable"))
	writeFile(t, filepath.Join(repo, "log", "module.yaml"), validManifest("log", "stable"))
	writeFile(t, filepath.Join(repo, "metrics", "module.yaml"), validManifest("metrics", "stable"))
	writeFile(t, filepath.Join(repo, "x", "ai", "module.yaml"), validManifest("x/ai", "extension"))
	if err := os.MkdirAll(filepath.Join(repo, "x", "ops"), 0o755); err != nil {
		t.Fatalf("mkdir x/ops: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(repo, "x", "scheduler"), 0o755); err != nil {
		t.Fatalf("mkdir x/scheduler: %v", err)
	}

	missing, err := FindMissingModuleManifests(repo, map[string]struct{}{"x/ops": {}})
	if err != nil {
		t.Fatalf("FindMissingModuleManifests: %v", err)
	}
	if len(missing) != 1 || missing[0] != "x/scheduler" {
		t.Fatalf("unexpected missing manifests: %v", missing)
	}
}

func TestValidateModuleManifestsReportsSchemaAndPathViolations(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "core", "module.yaml"), `name: wrong
path: wrong
layer: unstable
status: ga
summary: bad manifest
responsibilities:
  - one
non_goals:
allowed_imports:
  - stdlib
forbidden_imports:
  - x/**
test_commands:
  - go test ./...
review_checklist:
  - stay explicit
agent_hints:
  - keep boundaries
`)

	violations, err := ValidateModuleManifests(repo)
	if err != nil {
		t.Fatalf("ValidateModuleManifests: %v", err)
	}

	joined := strings.Join(violations, "\n")
	for _, want := range []string{
		`required list "non_goals" must not be empty`,
		`invalid layer "unstable"`,
		`path "wrong" does not match directory "core"`,
		`name "wrong" should match module path "core"`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected violation containing %q, got:\n%s", want, joined)
		}
	}
}

func TestValidateModuleManifestsRequiresDeclaredDocPathsToExist(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "core", "module.yaml"), validManifestWithDocPaths("core", "stable", "docs/modules/core/README.md"))
	writeFile(t, filepath.Join(repo, "docs", "modules", "core", "README.md"), "# core\n")

	violations, err := ValidateModuleManifests(repo)
	if err != nil {
		t.Fatalf("ValidateModuleManifests existing docs: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations for existing doc_paths, got %v", violations)
	}

	writeFile(t, filepath.Join(repo, "x", "webhook", "module.yaml"), validManifestWithDocPaths("x/webhook", "extension", "docs/modules/x-webhook/README.md"))

	violations, err = ValidateModuleManifests(repo)
	if err != nil {
		t.Fatalf("ValidateModuleManifests missing docs: %v", err)
	}

	joined := strings.Join(violations, "\n")
	if !strings.Contains(joined, `doc_paths target "docs/modules/x-webhook/README.md" does not exist`) {
		t.Fatalf("expected missing doc_paths violation, got:\n%s", joined)
	}
}

func TestFindUnexpectedTopLevelDirsHonorsBaseline(t *testing.T) {
	repo := t.TempDir()
	for _, dir := range []string{"core", "docs", "x", "legacy"} {
		if err := os.MkdirAll(filepath.Join(repo, dir), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	unexpected, err := FindUnexpectedTopLevelDirs(repo, AllowedTopLevelDirs(), map[string]struct{}{
		"legacy": {},
	})
	if err != nil {
		t.Fatalf("FindUnexpectedTopLevelDirs: %v", err)
	}
	if len(unexpected) != 0 {
		t.Fatalf("expected no unexpected dirs, got %v", unexpected)
	}

	if err := os.MkdirAll(filepath.Join(repo, "rogue"), 0o755); err != nil {
		t.Fatalf("mkdir rogue: %v", err)
	}
	unexpected, err = FindUnexpectedTopLevelDirs(repo, AllowedTopLevelDirs(), map[string]struct{}{
		"legacy": {},
	})
	if err != nil {
		t.Fatalf("FindUnexpectedTopLevelDirs second pass: %v", err)
	}
	if len(unexpected) != 1 || unexpected[0] != "rogue" {
		t.Fatalf("unexpected dirs: %v", unexpected)
	}
}

func TestReadRepoExtensionRootsParsesDeclaredExtensionPaths(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "specs", "repo.yaml"), `layers:
  stable:
    paths:
      - core
  extension:
    paths:
      - x/ai
      - x/fileapi
      - x/websocket
`)

	roots, err := ReadRepoExtensionRoots(repo)
	if err != nil {
		t.Fatalf("ReadRepoExtensionRoots: %v", err)
	}

	for _, want := range []string{"x/ai", "x/fileapi", "x/websocket"} {
		if _, ok := roots[want]; !ok {
			t.Fatalf("expected %s in parsed extension roots, got %v", want, roots)
		}
	}
	if len(roots) != 3 {
		t.Fatalf("expected exactly 3 roots, got %d: %v", len(roots), roots)
	}
}

func TestFindOrphanedExtensionRootsReportsUndeclaredXDirs(t *testing.T) {
	repo := t.TempDir()
	for _, dir := range []string{"x/ai", "x/fileapi", "x/websocket"} {
		if err := os.MkdirAll(filepath.Join(repo, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	orphans, err := FindOrphanedExtensionRoots(repo, map[string]struct{}{
		"x/ai":        {},
		"x/websocket": {},
	})
	if err != nil {
		t.Fatalf("FindOrphanedExtensionRoots: %v", err)
	}
	if len(orphans) != 1 || orphans[0] != "x/fileapi" {
		t.Fatalf("unexpected orphaned extension roots: %v", orphans)
	}
}

func TestFindEmptyMisleadingDirsFlagsEmptyPackageLikeDirs(t *testing.T) {
	repo := t.TempDir()
	for _, dir := range []string{
		"contract/protocol",
		"x/fileapi",
		"x/data/file/migrations",
		"x/data/file/testdata",
	} {
		if err := os.MkdirAll(filepath.Join(repo, filepath.FromSlash(dir)), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	writeFile(t, filepath.Join(repo, "x", "fileapi", "handler.go"), "package fileapi\n")

	empty, err := FindEmptyMisleadingDirs(repo)
	if err != nil {
		t.Fatalf("FindEmptyMisleadingDirs: %v", err)
	}

	if len(empty) != 1 || empty[0] != "contract/protocol" {
		t.Fatalf("unexpected empty misleading dirs: %v", empty)
	}
}

func TestReadCanonicalExtensionEntrypointsParsesCanonicalRoots(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "specs", "extension-entrypoints.yaml"), `families:
  gateway:
    canonical_entrypoint: x/gateway
  resource_api:
    canonical_entrypoint: x/rest
`)

	roots, err := ReadCanonicalExtensionEntrypoints(repo)
	if err != nil {
		t.Fatalf("ReadCanonicalExtensionEntrypoints: %v", err)
	}
	if len(roots) != 2 || roots[0] != "x/gateway" || roots[1] != "x/rest" {
		t.Fatalf("unexpected canonical entrypoints: %v", roots)
	}
}

func TestFindExtensionPrimerCoverageViolationsRequiresDocPaths(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "x", "gateway", "module.yaml"), validManifestWithDocPaths("x/gateway", "extension", "docs/modules/x-gateway/README.md"))
	writeFile(t, filepath.Join(repo, "docs", "modules", "x-gateway", "README.md"), "# x/gateway\n")
	writeFile(t, filepath.Join(repo, "x", "rest", "module.yaml"), validManifest("x/rest", "extension"))

	violations, err := FindExtensionPrimerCoverageViolations(repo, []string{"x/gateway", "x/rest", "x/fileapi"})
	if err != nil {
		t.Fatalf("FindExtensionPrimerCoverageViolations: %v", err)
	}

	joined := strings.Join(violations, "\n")
	if !strings.Contains(joined, `x/rest is a canonical extension entrypoint`) {
		t.Fatalf("expected missing doc_paths violation for x/rest, got:\n%s", joined)
	}
	if !strings.Contains(joined, `x/fileapi is a canonical extension entrypoint but has no module.yaml`) {
		t.Fatalf("expected missing module.yaml violation for x/fileapi, got:\n%s", joined)
	}
	if strings.Contains(joined, "x/gateway") && !strings.Contains(joined, "x/fileapi") && !strings.Contains(joined, "x/rest") {
		t.Fatalf("unexpected violation set: %s", joined)
	}
}

func TestReadPackageIndexParsesPackagesAndStartPaths(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "specs", "package-index.yaml"), `packages:
  x/fileapi:
    start_with:
      - x/fileapi/module.yaml
      - x/fileapi/handler.go
  contract:
    start_with:
      - contract/module.yaml
`)

	entries, err := ReadPackageIndex(repo)
	if err != nil {
		t.Fatalf("ReadPackageIndex: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 package index entries, got %d: %v", len(entries), entries)
	}
	if got := entries["x/fileapi"].StartWith; len(got) != 2 {
		t.Fatalf("expected 2 start_with paths for x/fileapi, got %v", got)
	}
}

func TestFindPackageIndexCoverageViolationsRequiresExistingPackageAndStartFiles(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "x", "fileapi", "module.yaml"), "name: x/fileapi\n")
	writeFile(t, filepath.Join(repo, "contract", "module.yaml"), "name: contract\n")

	violations, err := FindPackageIndexCoverageViolations(repo, map[string]packageIndexEntry{
		"x/fileapi": {
			Path:      "x/fileapi",
			StartWith: []string{"x/fileapi/module.yaml", "x/fileapi/handler.go"},
		},
		"contract": {
			Path:      "contract",
			StartWith: nil,
		},
		"x/missing": {
			Path:      "x/missing",
			StartWith: []string{"x/missing/module.yaml"},
		},
	})
	if err != nil {
		t.Fatalf("FindPackageIndexCoverageViolations: %v", err)
	}

	joined := strings.Join(violations, "\n")
	for _, want := range []string{
		`specs/package-index.yaml package x/fileapi references missing start_with path x/fileapi/handler.go`,
		`specs/package-index.yaml package contract must declare at least one start_with path`,
		`specs/package-index.yaml package x/missing does not exist in the repository`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected violation containing %q, got:\n%s", want, joined)
		}
	}
}

func TestFindStableHTTPSurfaceViolationsFlagsAppFacingHandlersOutsideCoreAndRouter(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "store", "blob", "handler.go"), `package blob

import "net/http"

func Upload(w http.ResponseWriter, r *http.Request) {}
`)
	writeFile(t, filepath.Join(repo, "health", "admin", "routes.go"), `package admin

import "github.com/spcent/plumego/router"

func RegisterRoutes(r *router.Router) {}
`)
	writeFile(t, filepath.Join(repo, "core", "routing.go"), `package core

import "net/http"

type App struct{}

func (a *App) HandleFunc(pattern string, handler http.HandlerFunc) {}
`)
	writeFile(t, filepath.Join(repo, "metrics", "exporter.go"), `package metrics

import "net/http"

type Exporter struct{}

func (e *Exporter) Handler() http.Handler { return nil }
`)

	violations, err := FindStableHTTPSurfaceViolations(repo)
	if err != nil {
		t.Fatalf("FindStableHTTPSurfaceViolations: %v", err)
	}

	joined := strings.Join(violations, "\n")
	for _, want := range []string{
		`stable package health/admin/routes.go exposes route registration helper RegisterRoutes`,
		`stable package store/blob/handler.go exposes app-facing HTTP handler surface Upload`,
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("expected violation containing %q, got:\n%s", want, joined)
		}
	}
	for _, unwanted := range []string{
		"core/routing.go",
		"metrics/exporter.go",
	} {
		if strings.Contains(joined, unwanted) {
			t.Fatalf("did not expect %q in violations:\n%s", unwanted, joined)
		}
	}
}

func validManifest(path, layer string) string {
	return "name: " + path + "\n" +
		"path: " + path + "\n" +
		"layer: " + layer + "\n" +
		"status: ga\n" +
		"summary: example\n" +
		"responsibilities:\n  - keep scope tight\n" +
		"non_goals:\n  - do not sprawl\n" +
		"allowed_imports:\n  - stdlib\n" +
		"forbidden_imports:\n  - x/**\n" +
		"test_commands:\n  - go test ./...\n" +
		"review_checklist:\n  - stay explicit\n" +
		"agent_hints:\n  - keep modules small\n"
}

func validManifestWithDocPaths(path, layer string, docPaths ...string) string {
	manifest := validManifest(path, layer)
	if len(docPaths) == 0 {
		return manifest
	}
	manifest += "doc_paths:\n"
	for _, docPath := range docPaths {
		manifest += "  - " + docPath + "\n"
	}
	return manifest
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
