package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

func mustNewLocalServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	defer func() {
		if rec := recover(); rec != nil {
			msg := fmt.Sprint(rec)
			if strings.Contains(msg, "failed to listen on a port") || strings.Contains(msg, "operation not permitted") {
				t.Skipf("skipping local listener test in restricted runtime: %s", msg)
			}
			panic(rec)
		}
	}()

	return httptest.NewServer(handler)
}

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

func TestCLI_UnknownCommandExitCode(t *testing.T) {
	_, _, err := runCLI(t, []string{"unknown-command"}, "")
	if err == nil {
		t.Fatalf("expected error for unknown command")
	}

	code, ok := output.ExitCode(err)
	if !ok || code != 1 {
		t.Fatalf("expected exit code 1, got %d (ok=%v)", code, ok)
	}
}

func TestCLI_GlobalFlagsDoNotLeakAcrossRuns(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--quiet", "version"}, "")
	if err != nil {
		t.Fatalf("quiet version command failed: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("expected no output in quiet mode, got: %s", stdout)
	}

	stdout, _, err = runCLI(t, []string{"version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected output after quiet run, got empty output")
	}
}

func TestCLI_MigrateCreateParsesFlagsAfterSubcommand(t *testing.T) {
	tmpDir := t.TempDir()
	migrationDir := filepath.Join(tmpDir, "db", "migrations")

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"migrate", "create", "add_users_table", "--dir", migrationDir,
	}, "")
	if err != nil {
		t.Fatalf("migrate create failed: %v", err)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Directory string `json:"directory"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}

	absDir, err := filepath.Abs(migrationDir)
	if err != nil {
		t.Fatalf("failed to resolve migration dir: %v", err)
	}
	if payload.Data.Directory != absDir {
		t.Fatalf("expected migration directory %q, got %q", absDir, payload.Data.Directory)
	}
}

func TestCLI_InspectParsesFlagsAfterSubcommand(t *testing.T) {
	server := mustNewLocalServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer server.Close()

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"inspect", "health", "--url", server.URL,
	}, "")
	if err != nil {
		t.Fatalf("inspect health failed: %v", err)
	}

	var payload struct {
		Status string         `json:"status"`
		Data   map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}
	if payload.Data["endpoint"] != "/health" {
		t.Fatalf("expected endpoint /health, got %v", payload.Data["endpoint"])
	}
}
