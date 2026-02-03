package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

func runCLI(t *testing.T, args []string, cwd string) (string, string, error) {
	t.Helper()

	root := &RootCmd{
		subcommands: make(map[string]Command),
		formatter:   output.NewFormatter(),
	}

	root.Register(&NewCmd{})
	root.Register(&GenerateCmd{})
	root.Register(&DevCmd{})
	root.Register(&RoutesCmd{})
	root.Register(&CheckCmd{})
	root.Register(&ConfigCmd{})
	root.Register(&MigrateCmd{})
	root.Register(&TestCmd{})
	root.Register(&BuildCmd{})
	root.Register(&InspectCmd{})
	root.Register(&VersionCmd{})

	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	root.formatter.SetWriters(&outBuf, &errBuf)

	if cwd != "" {
		prev, err := os.Getwd()
		if err != nil {
			t.Fatalf("getwd failed: %v", err)
		}
		if err := os.Chdir(cwd); err != nil {
			t.Fatalf("chdir failed: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Chdir(prev)
		})
	}

	err := root.Run(args)
	return outBuf.String(), errBuf.String(), err
}

func TestCLI_VersionJSONOutput(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "json", "version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}

	if payload["status"] != "success" {
		t.Fatalf("expected status success, got %v", payload["status"])
	}
	if payload["data"] == nil {
		t.Fatalf("expected data in output, got nil")
	}
}

func TestCLI_ConfigShowJSONOutput(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, err := runCLI(t, []string{"--format", "json", "config", "show"}, tmpDir)
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}

	var payload struct {
		Config map[string]any    `json:"config"`
		Source map[string]string `json:"source"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}

	if _, ok := payload.Config["app"]; !ok {
		t.Fatalf("expected app config, got: %#v", payload.Config)
	}
	if _, ok := payload.Config["security"]; !ok {
		t.Fatalf("expected security config, got: %#v", payload.Config)
	}
}

func TestCLI_ConfigValidateExitCode(t *testing.T) {
	tmpDir := t.TempDir()

	stdout, _, err := runCLI(t, []string{"--format", "json", "config", "validate"}, tmpDir)
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}

	code, ok := output.ExitCode(err)
	if !ok || code != 1 {
		t.Fatalf("expected exit code 1, got %v (ok=%v)", code, ok)
	}

	var payload struct {
		Valid  bool `json:"valid"`
		Errors []struct {
			Field string `json:"field"`
		} `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}
	if payload.Valid {
		t.Fatalf("expected valid=false, got true")
	}
	if len(payload.Errors) == 0 {
		t.Fatalf("expected at least one error, got none")
	}
}
