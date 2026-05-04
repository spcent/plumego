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
	root.Register(&ServeCmd{})
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

func writeTinyCanonicalProject(t *testing.T, dir string) {
	t.Helper()

	appDir := filepath.Join(dir, "cmd", "app")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		t.Fatalf("mkdir app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tiny\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	mainGo := []byte(`package main

import "fmt"

func main() {
	fmt.Println("ok")
}
`)
	if err := os.WriteFile(filepath.Join(appDir, "main.go"), mainGo, 0644); err != nil {
		t.Fatalf("write cmd/app/main.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Tiny\n"), 0644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(".env\nbin/\n"), 0644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "env.example"), []byte("APP_ADDR=:8080\n"), 0644); err != nil {
		t.Fatalf("write env.example: %v", err)
	}
}

func writeTinyTestProject(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/testproject\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc.go"), []byte(`package testproject

func Add(a, b int) int {
	return a + b
}
`), 0644); err != nil {
		t.Fatalf("write calc.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "calc_test.go"), []byte(`package testproject

import "testing"

func TestAdd(t *testing.T) {
	if Add(2, 3) != 5 {
		t.Fatal("bad add")
	}
}
`), 0644); err != nil {
		t.Fatalf("write calc_test.go: %v", err)
	}
}

type cliJSONEnvelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
	Error   json.RawMessage `json:"error,omitempty"`
}

type inspectHealthData struct {
	Endpoint string `json:"endpoint"`
}

func TestCLI_VersionJSONOutput(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "json", "version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}

	if payload.Status != "success" {
		t.Fatalf("expected status success, got %v", payload.Status)
	}
	if len(payload.Data) == 0 {
		t.Fatalf("expected data in output, got nil")
	}
}

func TestCLI_DefaultFormatIsJSON(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("default output should be json: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}
}

func TestCLI_GlobalInlineFormatFlag(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format=json", "version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("inline format output should be json: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}
}

func TestCLI_GlobalFlagsStopAtCommandToken(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"version", "--format", "text"}, "")
	if err == nil {
		t.Fatalf("expected command-local --format to fail instead of being consumed globally")
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("expected json error envelope: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
}

func TestCLI_JSONEnvelopeIsCommandOutput(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "json", "version"}, "")
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}

	if payload.Status != "success" {
		t.Fatalf("expected CLI status field, got %v", payload.Status)
	}
	if payload.Message == "" {
		t.Fatalf("expected CLI message field, got: %#v", payload)
	}
	if len(payload.Error) != 0 {
		t.Fatalf("CLI success output should not mimic HTTP error envelope: %#v", payload)
	}
}

func TestCLI_InvalidFormatFailsClosed(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "bogus", "version"}, "")
	if err == nil {
		t.Fatalf("expected invalid format error")
	}

	code, ok := output.ExitCode(err)
	if !ok || code != 1 {
		t.Fatalf("expected exit code 1, got %d (ok=%v)", code, ok)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("invalid format error should be json: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "unsupported output format") {
		t.Fatalf("unexpected invalid format response: %#v", payload)
	}
}

func TestCLI_CommandHelpReturnsUsage(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"new", "--help"}, "")
	if err != nil {
		t.Fatalf("command help failed: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "plumego [global-flags] new") {
		t.Fatalf("expected command usage, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Create new project from template") {
		t.Fatalf("expected command summary, got: %s", stdout)
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

func TestCLI_ConfigShowRedactsSecretsByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WS_SECRET", "")

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("WS_SECRET=super-secret-value\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	stdout, _, err := runCLI(t, []string{"--format", "json", "config", "show", "--resolve"}, tmpDir)
	if err != nil {
		t.Fatalf("config show failed: %v", err)
	}
	if strings.Contains(stdout, "super-secret-value") {
		t.Fatalf("config show leaked secret: %s", stdout)
	}

	var payload struct {
		Config struct {
			Security map[string]any `json:"security"`
		} `json:"config"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Config.Security["ws_secret"] != "***REDACTED***" {
		t.Fatalf("expected redacted ws_secret, got %#v", payload.Config.Security["ws_secret"])
	}
}

func TestCLI_ConfigShowSecretsRequiresExplicitFlag(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("WS_SECRET", "")

	if err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte("WS_SECRET=super-secret-value\n"), 0644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	stdout, _, err := runCLI(t, []string{"--format", "json", "config", "show", "--resolve", "--show-secrets"}, tmpDir)
	if err != nil {
		t.Fatalf("config show --show-secrets failed: %v", err)
	}

	var payload struct {
		Config struct {
			Security map[string]any `json:"security"`
		} `json:"config"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Config.Security["ws_secret"] != "super-secret-value" {
		t.Fatalf("expected raw ws_secret, got %#v", payload.Config.Security["ws_secret"])
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
		Status   string `json:"status"`
		ExitCode int    `json:"exit_code"`
		Data     struct {
			Valid  bool `json:"valid"`
			Errors []struct {
				Field string `json:"field"`
			} `json:"errors"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" || payload.ExitCode != 1 {
		t.Fatalf("unexpected validation envelope: %#v", payload)
	}
	if payload.Data.Valid {
		t.Fatalf("expected valid=false, got true")
	}
	if len(payload.Data.Errors) == 0 {
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

func TestCLI_NewAcceptsScenarioTemplatesDryRun(t *testing.T) {
	templates := []string{
		"rest-api",
		"tenant-api",
		"gateway",
		"realtime",
		"ai-service",
		"ops-service",
	}

	for _, template := range templates {
		t.Run(template, func(t *testing.T) {
			stdout, _, err := runCLI(t, []string{
				"--format", "json",
				"new", "--template", template, "--dry-run", "trust-check",
			}, t.TempDir())
			if err != nil {
				t.Fatalf("new dry-run failed: %v\noutput: %s", err, stdout)
			}

			var payload struct {
				DryRun       bool     `json:"dry_run"`
				Template     string   `json:"template"`
				FilesCreated []string `json:"files_created"`
			}
			if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
				t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
			}
			if !payload.DryRun {
				t.Fatalf("expected dry_run=true, got %#v", payload)
			}
			if payload.Template != template {
				t.Fatalf("expected template %q, got %q", template, payload.Template)
			}
			if len(payload.FilesCreated) == 0 {
				t.Fatalf("expected file preview for %q, got none", template)
			}
		})
	}
}

func TestCLI_NewRejectsInvalidTemplate(t *testing.T) {
	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"new", "--template", "invalid-template-name", "--dry-run", "trust-check",
	}, t.TempDir())
	if err == nil {
		t.Fatalf("expected invalid template error")
	}

	code, ok := output.ExitCode(err)
	if !ok || code != 3 {
		t.Fatalf("expected exit code 3, got %d (ok=%v)", code, ok)
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
	if !strings.Contains(payload.Message, "invalid-template-name") {
		t.Fatalf("expected invalid template in message, got %q", payload.Message)
	}
	if !strings.Contains(payload.Message, "rest-api") {
		t.Fatalf("expected supported scenario templates in message, got %q", payload.Message)
	}
}

func TestCLI_NewWritesLocalPlumegoReplaceWhenRunFromCheckout(t *testing.T) {
	root := findPlumegoModuleRoot(".")
	if root == "" {
		t.Skip("not running from a plumego checkout")
	}

	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "demo")
	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"new", "--dir", projectDir, "--no-git", "demo",
	}, "")
	if err != nil {
		t.Fatalf("new failed: %v\noutput: %s", err, stdout)
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "go.mod"))
	if err != nil {
		t.Fatalf("read generated go.mod: %v", err)
	}
	want := "replace github.com/spcent/plumego => " + filepath.ToSlash(root)
	if !strings.Contains(string(data), want) {
		t.Fatalf("generated go.mod missing local replace %q:\n%s", want, string(data))
	}
}

func TestCLI_GenerateHandlerUsesCanonicalDefaultPath(t *testing.T) {
	tmpDir := t.TempDir()
	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"generate", "handler", "User", "--methods", "GET",
	}, tmpDir)
	if err != nil {
		t.Fatalf("generate handler failed: %v\noutput: %s", err, stdout)
	}

	generatedPath := filepath.Join(tmpDir, "internal", "handler", "user.go")
	if _, err := os.Stat(generatedPath); err != nil {
		t.Fatalf("expected generated handler at %s: %v", generatedPath, err)
	}
}

func TestCLI_GenerateRejectsUnsupportedMethod(t *testing.T) {
	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"generate", "handler", "User", "--methods", "TRACE",
	}, t.TempDir())
	if err == nil {
		t.Fatalf("expected unsupported method to fail")
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "unsupported HTTP method") {
		t.Fatalf("unexpected generation error: %#v", payload)
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
		Status string            `json:"status"`
		Data   inspectHealthData `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}
	if payload.Data.Endpoint != "/health" {
		t.Fatalf("expected endpoint /health, got %v", payload.Data.Endpoint)
	}
}

func TestCLI_InspectUsesCanonicalDebugEndpoints(t *testing.T) {
	tests := []struct {
		name        string
		subcommand  string
		wantPath    string
		wantMessage string
	}{
		{name: "routes", subcommand: "routes", wantPath: "/_debug/routes.json", wantMessage: "Routes retrieved"},
		{name: "config", subcommand: "config", wantPath: "/_debug/config", wantMessage: "Configuration retrieved"},
		{name: "info", subcommand: "info", wantPath: "/_debug/info", wantMessage: "Application info retrieved"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mustNewLocalServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != tt.wantPath {
					t.Fatalf("unexpected path %q", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
			}))
			defer server.Close()

			stdout, _, err := runCLI(t, []string{
				"--format", "json",
				"inspect", tt.subcommand, "--url", server.URL,
			}, "")
			if err != nil {
				t.Fatalf("inspect %s failed: %v", tt.subcommand, err)
			}

			var payload struct {
				Status  string `json:"status"`
				Message string `json:"message"`
				Data    struct {
					OK bool `json:"ok"`
				} `json:"data"`
			}
			if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
				t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
			}
			if payload.Status != "success" {
				t.Fatalf("expected success status, got %q", payload.Status)
			}
			if payload.Message != tt.wantMessage {
				t.Fatalf("expected message %q, got %q", tt.wantMessage, payload.Message)
			}
		})
	}
}

func TestCLI_BuildDefaultsToCanonicalCmdApp(t *testing.T) {
	tmpDir := t.TempDir()
	writeTinyCanonicalProject(t, tmpDir)
	outputPath := filepath.Join(tmpDir, "bin", "tiny")

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"build", "--dir", tmpDir, "--output", outputPath,
	}, "")
	if err != nil {
		t.Fatalf("build failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Target string `json:"target"`
			Binary string `json:"binary"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success status, got %q", payload.Status)
	}
	if payload.Data.Target != "./cmd/app" {
		t.Fatalf("expected build target ./cmd/app, got %q", payload.Data.Target)
	}
	if payload.Data.Binary != outputPath {
		t.Fatalf("expected binary %q, got %q", outputPath, payload.Data.Binary)
	}
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("expected build output at %s: %v", outputPath, err)
	}
}

func TestCLI_TestCoverUsesTempProfileByDefault(t *testing.T) {
	tmpDir := t.TempDir()
	writeTinyTestProject(t, tmpDir)

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"test", "--dir", tmpDir, "--cover",
	}, "")
	if err != nil {
		t.Fatalf("test --cover failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			CoveragePercent float64 `json:"coverage_percent"`
			CoverageProfile string  `json:"coverage_profile"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" || payload.Data.CoveragePercent <= 0 {
		t.Fatalf("unexpected coverage payload: %#v", payload)
	}
	if payload.Data.CoverageProfile != "" {
		t.Fatalf("default temp coverage profile should not be exposed, got %q", payload.Data.CoverageProfile)
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "coverage.out")); !os.IsNotExist(err) {
		t.Fatalf("default coverage should not write coverage.out in project root")
	}
}

func TestCLI_TestFailureIncludesStructuredFailures(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/failtest\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "fail_test.go"), []byte(`package failtest

import "testing"

func TestFailure(t *testing.T) {
	t.Fatal("boom")
}
`), 0644); err != nil {
		t.Fatalf("write fail_test.go: %v", err)
	}

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"test", "--dir", tmpDir,
	}, "")
	if err == nil {
		t.Fatalf("expected failing tests")
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Status   string `json:"status"`
			Failures []struct {
				Package string `json:"package"`
				Test    string `json:"test"`
			} `json:"failures"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" || payload.Data.Status != "failed" || len(payload.Data.Failures) == 0 {
		t.Fatalf("unexpected failure payload: %#v", payload)
	}
	if payload.Data.Failures[0].Test != "TestFailure" {
		t.Fatalf("expected TestFailure in structured failures, got %#v", payload.Data.Failures)
	}
}

func TestParseServeArgsInterspersedFlags(t *testing.T) {
	opts, err := parseServeArgs([]string{"./public", "--addr", "127.0.0.1:0"})
	if err != nil {
		t.Fatalf("parseServeArgs failed: %v", err)
	}
	if opts.dir != "./public" || opts.addr != "127.0.0.1:0" {
		t.Fatalf("unexpected serve options: %#v", opts)
	}
}

func TestCLI_ServeInvalidDirectoryUsesErrorEnvelope(t *testing.T) {
	tmpDir := t.TempDir()
	missingDir := filepath.Join(tmpDir, "missing")

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"serve", "--addr", "127.0.0.1:0", missingDir,
	}, "")
	if err == nil {
		t.Fatalf("expected invalid directory error")
	}

	var payload cliJSONEnvelope
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "directory does not exist") {
		t.Fatalf("unexpected serve error payload: %#v", payload)
	}
}

func TestCLI_ServeHelpReturnsUsage(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"serve", "--help"}, "")
	if err != nil {
		t.Fatalf("serve help failed: %v\noutput: %s", err, stdout)
	}
	if !strings.Contains(stdout, "plumego [global-flags] serve") {
		t.Fatalf("expected serve usage, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Start static file server") {
		t.Fatalf("expected serve summary, got: %s", stdout)
	}
}

func TestCLI_CheckAcceptsCanonicalCmdAppEntrypoint(t *testing.T) {
	tmpDir := t.TempDir()
	writeTinyCanonicalProject(t, tmpDir)

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"check",
	}, tmpDir)
	if err != nil {
		t.Fatalf("check failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
			Checks struct {
				Structure struct {
					Status string `json:"status"`
					Issues []struct {
						Message string `json:"message"`
					} `json:"issues"`
				} `json:"structure"`
			} `json:"checks"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "success" || payload.Data.Status != "healthy" {
		t.Fatalf("expected healthy success, got %#v", payload)
	}
	if payload.Data.Checks.Structure.Status != "passed" {
		t.Fatalf("expected structure passed, got %#v", payload.Data.Checks.Structure)
	}
	for _, issue := range payload.Data.Checks.Structure.Issues {
		if strings.Contains(issue.Message, "entrypoint") || strings.Contains(issue.Message, "main.go") {
			t.Fatalf("canonical cmd/app entrypoint should not warn, got issue %q", issue.Message)
		}
	}
}

func TestCLI_CheckDegradedUsesWarningEnvelope(t *testing.T) {
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/degraded\n\ngo 1.24\n"), 0644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	stdout, _, err := runCLI(t, []string{
		"--format", "json",
		"check",
	}, tmpDir)
	if err == nil {
		t.Fatalf("expected degraded check to return exit error")
	}
	code, ok := output.ExitCode(err)
	if !ok || code != 2 {
		t.Fatalf("expected exit code 2, got %d (ok=%v)", code, ok)
	}

	var payload struct {
		Status   string `json:"status"`
		ExitCode int    `json:"exit_code"`
		Data     struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse output: %v\noutput: %s", err, stdout)
	}
	if payload.Status != "warning" || payload.ExitCode != 2 || payload.Data.Status != "degraded" {
		t.Fatalf("unexpected degraded envelope: %#v", payload)
	}
}
