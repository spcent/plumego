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

func TestValidateStableBoundaryDeclarationsReportsWhenMissing(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "core", "module.yaml"), validManifestWithStrictBoundary("core", "stable", "kernel"))
	writeFile(t, filepath.Join(repo, "router", "module.yaml"), validManifest("router", "stable"))
	// other stable roots absent → skipped (FindMissingModuleManifests handles them)

	violations, err := ValidateStableBoundaryDeclarations(repo)
	if err != nil {
		t.Fatalf("ValidateStableBoundaryDeclarations: %v", err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d: %v", len(violations), violations)
	}
	if !strings.Contains(violations[0], "router/module.yaml") || !strings.Contains(violations[0], "strict_boundary") {
		t.Fatalf("unexpected violation: %q", violations[0])
	}
}

func TestValidateStableBoundaryDeclarationsPassesWhenPresent(t *testing.T) {
	repo := t.TempDir()
	writeFile(t, filepath.Join(repo, "core", "module.yaml"), validManifestWithStrictBoundary("core", "stable", "kernel"))
	writeFile(t, filepath.Join(repo, "router", "module.yaml"), validManifestWithStrictBoundary("router", "stable", "route_structure"))

	violations, err := ValidateStableBoundaryDeclarations(repo)
	if err != nil {
		t.Fatalf("ValidateStableBoundaryDeclarations: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got: %v", violations)
	}
}

func TestValidateXFamilyTaxonomyDetectsViolations(t *testing.T) {
	repo := t.TempDir()

	// x/messaging lacks subordinate_families → violation
	writeFile(t, filepath.Join(repo, "x", "messaging", "module.yaml"), validManifest("x/messaging", "extension"))

	// x/mq declares an unknown parent_family → violation
	writeFile(t, filepath.Join(repo, "x", "mq", "module.yaml"), validManifestWithParentFamily("x/mq", "extension", "x/nonexistent"))

	// x/gateway has subordinate_families → no violation
	writeFile(t, filepath.Join(repo, "x", "gateway", "module.yaml"), validManifestWithSubordinateFamilies("x/gateway", "extension"))

	// x/pubsub declares a valid parent_family → no violation
	writeFile(t, filepath.Join(repo, "x", "pubsub", "module.yaml"), validManifestWithParentFamily("x/pubsub", "extension", "x/messaging"))

	violations, err := ValidateXFamilyTaxonomy(repo)
	if err != nil {
		t.Fatalf("ValidateXFamilyTaxonomy: %v", err)
	}
	// expect exactly 2: x/messaging missing subordinate_families, x/mq unknown parent
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(violations), violations)
	}
	joined := strings.Join(violations, "\n")
	if !strings.Contains(joined, "x/messaging/module.yaml") || !strings.Contains(joined, "subordinate_families") {
		t.Fatalf("expected subordinate_families violation for x/messaging, got:\n%s", joined)
	}
	if !strings.Contains(joined, "x/mq/module.yaml") || !strings.Contains(joined, "x/nonexistent") {
		t.Fatalf("expected parent_family violation for x/mq, got:\n%s", joined)
	}
}

func TestValidateXFamilyTaxonomyPassesForValidSetup(t *testing.T) {
	repo := t.TempDir()

	writeFile(t, filepath.Join(repo, "x", "messaging", "module.yaml"), validManifestWithSubordinateFamilies("x/messaging", "extension"))
	writeFile(t, filepath.Join(repo, "x", "mq", "module.yaml"), validManifestWithParentFamily("x/mq", "extension", "x/messaging"))
	writeFile(t, filepath.Join(repo, "x", "rest", "module.yaml"), validManifest("x/rest", "extension"))

	violations, err := ValidateXFamilyTaxonomy(repo)
	if err != nil {
		t.Fatalf("ValidateXFamilyTaxonomy: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got: %v", violations)
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

func validManifestWithStrictBoundary(path, layer, boundary string) string {
	return validManifest(path, layer) + "strict_boundary: " + boundary + "\n"
}

func validManifestWithParentFamily(path, layer, parent string) string {
	return validManifest(path, layer) + "parent_family: " + parent + "\n"
}

func validManifestWithSubordinateFamilies(path, layer string) string {
	return validManifest(path, layer) + "subordinate_families:\n  - package: x/sub\n    role: subordinate\n"
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
