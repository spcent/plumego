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

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
