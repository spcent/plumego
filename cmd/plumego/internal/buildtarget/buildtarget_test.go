package buildtarget

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPrefersRootMainPackage(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "main.go"), "package main\nfunc main() {}\n")
	writeFile(t, filepath.Join(tmp, "cmd", "app", "main.go"), "package main\nfunc main() {}\n")

	if got := Default(tmp); got != "." {
		t.Fatalf("Default() = %q, want .", got)
	}
}

func TestDefaultUsesCanonicalCmdAppMainPackage(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "cmd", "app", "app.go"), "package main\nfunc main() {}\n")

	if got := Default(tmp); got != "./cmd/app" {
		t.Fatalf("Default() = %q, want ./cmd/app", got)
	}
	if !HasDefaultEntrypoint(tmp) {
		t.Fatal("expected default entrypoint")
	}
}

func TestHasMainPackageIgnoresCommentsAndStrings(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, "notmain.go"), `package app

const comment = "package main"
`)

	if HasMainPackage(tmp) {
		t.Fatal("comment/string mention should not count as package main")
	}
	if HasDefaultEntrypoint(tmp) {
		t.Fatal("expected no default entrypoint")
	}
}

func writeFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
