package checker

import (
	"os"
	"path/filepath"
	"testing"
)

func writeCheckerFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestCheckStructureAcceptsCanonicalCmdAppEntrypoint(t *testing.T) {
	tmp := t.TempDir()
	writeCheckerFile(t, filepath.Join(tmp, "cmd", "app", "main.go"), "package main\nfunc main() {}\n")
	writeCheckerFile(t, filepath.Join(tmp, "README.md"), "# App\n")
	writeCheckerFile(t, filepath.Join(tmp, ".gitignore"), ".env\n")

	detail := CheckStructure(tmp)
	if detail.Status != "passed" {
		t.Fatalf("status = %q, want passed; issues: %#v", detail.Status, detail.Issues)
	}
	for _, issue := range detail.Issues {
		if issue.Message == "application entrypoint not found" {
			t.Fatalf("canonical cmd/app entrypoint should not warn: %#v", detail.Issues)
		}
	}
}

func TestCheckStructureWarnsWhenEntrypointMissing(t *testing.T) {
	tmp := t.TempDir()
	writeCheckerFile(t, filepath.Join(tmp, "README.md"), "# App\n")
	writeCheckerFile(t, filepath.Join(tmp, ".gitignore"), ".env\n")

	detail := CheckStructure(tmp)
	if detail.Status != "warning" {
		t.Fatalf("status = %q, want warning", detail.Status)
	}
	if len(detail.Issues) != 1 || detail.Issues[0].Message != "application entrypoint not found" {
		t.Fatalf("expected entrypoint issue, got %#v", detail.Issues)
	}
}

func TestCheckSecurityReportsEnvParseError(t *testing.T) {
	tmp := t.TempDir()
	writeCheckerFile(t, filepath.Join(tmp, ".env"), "INVALID\n")

	detail := CheckSecurity(tmp, ".env")
	if detail.Status != "failed" {
		t.Fatalf("status = %q, want failed", detail.Status)
	}
	if len(detail.Issues) != 1 || detail.Issues[0].Severity != "high" {
		t.Fatalf("expected one high-severity parse issue, got %#v", detail.Issues)
	}
}

func TestCheckSecurityUsesSharedRequiredSecrets(t *testing.T) {
	tmp := t.TempDir()
	writeCheckerFile(t, filepath.Join(tmp, ".env"), "WS_SECRET=secret\n")
	t.Setenv("JWT_SECRET", "")

	detail := CheckSecurity(tmp, ".env")
	foundJWT := false
	for _, issue := range detail.Issues {
		if issue.Message == "JWT_SECRET not set in environment or .env" {
			foundJWT = true
		}
	}
	if !foundJWT {
		t.Fatalf("expected JWT_SECRET issue, got %#v", detail.Issues)
	}
}
