package commands

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintCmd_CleanProject(t *testing.T) {
	dir := t.TempDir()
	// A clean project: handler with canonical signature, no globals, no init effects.
	mustWriteFile(t, filepath.Join(dir, "internal", "handler", "user.go"), `
package handler

import "net/http"

func UserHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
`)
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir}, "")
	if err != nil {
		t.Fatalf("lint on clean project failed: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, `"status": "clean"`) {
		t.Fatalf("expected clean status, got: %s", stdout)
	}
}

func TestLintCmd_DetectsGlobalInHandlerPkg(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "handler", "api.go"), `
package handler

import "sync"

var cache sync.Map // global — should be flagged

func Handle() {}
`)
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir, "--check", "globals"}, "")
	if err == nil {
		t.Fatalf("expected lint to fail on global variable")
	}
	if !strings.Contains(stdout, "globals") {
		t.Fatalf("expected globals violation, got: %s", stdout)
	}
}

func TestLintCmd_DetectsInitSideEffect(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "service", "loader.go"), `
package service

var registry []string

func init() {
	registry = append(registry, "item") // side effect
}
`)
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir, "--check", "init-effects"}, "")
	if err == nil {
		t.Fatalf("expected lint to fail on init side effect")
	}
	if !strings.Contains(stdout, "init-effects") {
		t.Fatalf("expected init-effects violation, got: %s", stdout)
	}
}

func TestLintCmd_DetectsBoundaryViolation(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "middleware", "auth.go"), `
package middleware

import (
	"net/http"
	"example.com/app/service"
)

func Auth(next http.Handler) http.Handler {
	_ = service.New()
	return next
}
`)
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir, "--check", "boundaries"}, "")
	if err == nil {
		t.Fatalf("expected lint to fail on boundary violation")
	}
	if !strings.Contains(stdout, "boundaries") {
		t.Fatalf("expected boundaries violation, got: %s", stdout)
	}
}

func TestLintCmd_StrictModeFailsOnWarnings(t *testing.T) {
	dir := t.TempDir()
	// A handler function with a non-standard signature generates a warning.
	mustWriteFile(t, filepath.Join(dir, "handler", "api.go"), `
package handler

func ItemHandler(id string) string { // wrong signature for a handler
	return id
}
`)
	// Without --strict, should exit 0 or with warning code.
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir, "--check", "handlers", "--strict"}, "")
	if err == nil {
		t.Fatalf("expected strict lint to fail on handler warning")
	}
	if !strings.Contains(stdout, `"status": "error"`) {
		t.Fatalf("expected error status in strict mode, got: %s", stdout)
	}
}

func TestLintCmd_SelectiveChecks(t *testing.T) {
	dir := t.TempDir()
	mustWriteFile(t, filepath.Join(dir, "handler", "api.go"), `
package handler

var db interface{} // global — would be flagged by --check globals

func Handle() {}
`)
	// --check boundaries should pass; the global is not a boundary issue.
	stdout, _, err := runCLI(t, []string{"--format", "json", "lint", "--dir", dir, "--check", "boundaries"}, "")
	if err != nil {
		t.Fatalf("expected boundaries check to pass when only global exists: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, `"status": "clean"`) {
		t.Fatalf("expected clean status for boundaries-only check, got: %s", stdout)
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
