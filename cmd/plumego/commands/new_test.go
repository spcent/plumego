package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

func TestNewCmdParseFlags(t *testing.T) {
	cmd := &NewCmd{}

	// Test case 1: No arguments
	args := []string{}
	ctx := &Context{
		Out: output.NewFormatter(),
	}
	err := cmd.Run(ctx, args)
	if err == nil {
		t.Fatalf("expected error for no project name, got nil")
	}

	// Test case 2: Invalid template
	args = []string{"myapp", "--template", "invalid"}
	err = cmd.Run(ctx, args)
	if err == nil {
		t.Fatalf("expected error for invalid template, got nil")
	}

	// Test case 3: Dry run
	args = []string{"myapp", "--dry-run"}
	var buf bytes.Buffer
	ctx.Out.SetWriters(&buf, &buf)
	err = cmd.Run(ctx, args)
	if err != nil {
		t.Fatalf("expected no error for dry run, got: %v", err)
	}
}

func TestNewCmdCreateProject(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "testapp")

	cmd := &NewCmd{}
	ctx := &Context{
		Out: output.NewFormatter(),
	}

	// Test case: Create project with default template
	args := []string{"testapp", "--dir", projectDir, "--no-git"}
	err := cmd.Run(ctx, args)
	if err != nil {
		t.Fatalf("expected no error for project creation, got: %v", err)
	}

	// Verify project was created
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		t.Fatalf("expected project directory to be created, got error: %v", err)
	}

	// Verify go.mod file exists
	goModFile := filepath.Join(projectDir, "go.mod")
	if _, err := os.Stat(goModFile); os.IsNotExist(err) {
		t.Fatalf("expected go.mod file to be created, got error: %v", err)
	}
}
