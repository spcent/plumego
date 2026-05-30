package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeMinimalRepo creates a minimal repo structure with AGENTS.md, a spec
// file, and one stable-root module.yaml so doctor and agents can find the root.
func writeMinimalRepo(t *testing.T, dir string) {
	t.Helper()

	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Agents\n"), 0644); err != nil {
		t.Fatalf("write AGENTS.md: %v", err)
	}

	specsDir := filepath.Join(dir, "specs")
	if err := os.MkdirAll(specsDir, 0755); err != nil {
		t.Fatalf("mkdir specs: %v", err)
	}
	for _, name := range []string{"task-routing.yaml", "dependency-rules.yaml", "repo.yaml"} {
		if err := os.WriteFile(filepath.Join(specsDir, name), []byte("version: 1\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

// writeModuleYAML writes a minimal module.yaml under dir/<module>/module.yaml.
// testCmds overrides the default test_commands list; pass nil to use the default.
func writeModuleYAML(t *testing.T, repoDir, module string, testCmds []string) {
	t.Helper()

	modDir := filepath.Join(repoDir, module)
	if err := os.MkdirAll(modDir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", module, err)
	}

	cmds := testCmds
	if cmds == nil {
		cmds = []string{"go version"}
	}

	var cmdLines string
	for _, c := range cmds {
		cmdLines += "  - " + c + "\n"
	}

	content := "name: " + module + "\npath: " + module + "\nlayer: stable\nstatus: ga\n" +
		"summary: test module\nresponsibilities:\n  - testing\n" +
		"allowed_imports:\n  - stdlib\nforbidden_imports:\n  - x/**\n" +
		"test_commands:\n" + cmdLines +
		"agent_hints:\n  - keep it simple\n"
	if err := os.WriteFile(filepath.Join(modDir, "module.yaml"), []byte(content), 0644); err != nil {
		t.Fatalf("write module.yaml: %v", err)
	}
}

// ─── agents verify ────────────────────────────────────────────────────────────

func TestCLI_AgentsVerifyRequiresChangedFlag(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{"--format", "json", "agents", "verify", "--dir", tmp}, "")
	if err == nil {
		t.Fatalf("expected error when --changed is missing")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "--changed") {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestCLI_AgentsVerifyMissingManifestReturnsError(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "verify", "--changed", "nonexistent", "--dir", tmp,
	}, "")
	if err == nil {
		t.Fatalf("expected error for missing module.yaml")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
}

func TestCLI_AgentsVerifyRunsTestCommands(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)
	writeModuleYAML(t, tmp, "log", nil)

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "verify", "--changed", "log", "--dir", tmp,
	}, "")
	if err != nil {
		t.Fatalf("agents verify failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Data    struct {
			Module   string `json:"module"`
			Commands []struct {
				Command string `json:"command"`
				Status  string `json:"status"`
			} `json:"commands"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success, got %q\noutput: %s", payload.Status, stdout)
	}
	if payload.Data.Module != "log" {
		t.Fatalf("expected module=log, got %q", payload.Data.Module)
	}
	if len(payload.Data.Commands) == 0 {
		t.Fatal("expected at least one command result")
	}
	for _, cmd := range payload.Data.Commands {
		if cmd.Status != "passed" {
			t.Fatalf("command %q failed", cmd.Command)
		}
	}
}

func TestCLI_AgentsVerifyReportsFailedGate(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)
	// Use a command that will always fail
	writeModuleYAML(t, tmp, "badmod", []string{"/nonexistent/cmd/zzz"})

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "verify", "--changed", "badmod", "--dir", tmp,
	}, "")
	if err == nil {
		t.Fatalf("expected verify to fail for a failing gate command")
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Commands []struct {
				Status string `json:"status"`
			} `json:"commands"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
	if len(payload.Data.Commands) == 0 {
		t.Fatal("expected command results in error payload")
	}
	if payload.Data.Commands[0].Status != "failed" {
		t.Fatalf("expected first command to be failed, got %q", payload.Data.Commands[0].Status)
	}
}

func TestCLI_AgentsVerifyRejectsUnexpectedArguments(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "verify", "extra", "--changed", "log", "--dir", tmp,
	}, "")
	if err == nil {
		t.Fatalf("expected unexpected arguments error")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "unexpected arguments") {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

// ─── agents explain ───────────────────────────────────────────────────────────

func TestCLI_AgentsExplainRequiresModuleFlag(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{"--format", "json", "agents", "explain", "--dir", tmp}, "")
	if err == nil {
		t.Fatalf("expected error when --module is missing")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "--module") {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestCLI_AgentsExplainReturnsManifestFields(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)
	writeModuleYAML(t, tmp, "core", nil)

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "explain", "--module", "core", "--dir", tmp,
	}, "")
	if err != nil {
		t.Fatalf("agents explain failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Name             string   `json:"name"`
			Layer            string   `json:"layer"`
			Summary          string   `json:"summary"`
			Responsibilities []string `json:"responsibilities"`
			TestCommands     []string `json:"test_commands"`
			AgentHints       []string `json:"agent_hints"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success, got %q", payload.Status)
	}
	if payload.Data.Name != "core" {
		t.Fatalf("expected name=core, got %q", payload.Data.Name)
	}
	if payload.Data.Layer != "stable" {
		t.Fatalf("expected layer=stable, got %q", payload.Data.Layer)
	}
	if len(payload.Data.Responsibilities) == 0 {
		t.Fatal("expected non-empty responsibilities")
	}
	if len(payload.Data.TestCommands) == 0 {
		t.Fatal("expected non-empty test_commands")
	}
	if len(payload.Data.AgentHints) == 0 {
		t.Fatal("expected non-empty agent_hints")
	}
}

func TestCLI_AgentsExplainMissingModuleReturnsError(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{
		"--format", "json", "agents", "explain", "--module", "doesnotexist", "--dir", tmp,
	}, "")
	if err == nil {
		t.Fatalf("expected error for missing module")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
}

func TestCLI_AgentsUnknownSubcommandReturnsError(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	stdout, _, err := runCLI(t, []string{"--format", "json", "agents", "unknown"}, "")
	if err == nil {
		t.Fatalf("expected error for unknown agents subcommand")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "unknown agents subcommand") {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

// ─── doctor ───────────────────────────────────────────────────────────────────

func TestCLI_DoctorPassesOnFullRepo(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	// All known stable roots with module.yaml
	for _, root := range []string{"core", "router", "contract", "middleware", "security", "store", "health", "log", "metrics"} {
		writeModuleYAML(t, tmp, root, nil)
	}

	// All known reference apps with AGENT_TASKS.md
	for _, ref := range []string{
		"reference/standard-service",
		"reference/with-websocket",
		"reference/with-gateway",
		"reference/with-rest",
		"reference/with-ops",
	} {
		refDir := filepath.Join(tmp, ref)
		if err := os.MkdirAll(refDir, 0755); err != nil {
			t.Fatalf("mkdir %s: %v", ref, err)
		}
		if err := os.WriteFile(filepath.Join(refDir, "AGENT_TASKS.md"), []byte("# Agent Tasks\n"), 0644); err != nil {
			t.Fatalf("write AGENT_TASKS.md: %v", err)
		}
	}

	// Agent-first docs
	docsDir := filepath.Join(tmp, "docs", "concepts")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	docsFiles := []struct{ dir, name string }{
		{docsDir, "agent-first.md"},
		{docsDir, "core-boundary.md"},
		{docsDir, "extension-boundary.md"},
	}
	for _, f := range docsFiles {
		if err := os.WriteFile(filepath.Join(f.dir, f.name), []byte("# Doc\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}

	stdout, _, err := runCLI(t, []string{"--format", "json", "doctor", "--dir", tmp}, "")
	if err != nil {
		t.Fatalf("doctor failed: %v\noutput: %s", err, stdout)
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
			Passed int    `json:"passed"`
			Warned int    `json:"warned"`
			Failed int    `json:"failed"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "success" {
		t.Fatalf("expected success, got %q\noutput: %s", payload.Status, stdout)
	}
	if payload.Data.Status != "ready" {
		t.Fatalf("expected status=ready, got %q", payload.Data.Status)
	}
	if payload.Data.Failed != 0 {
		t.Fatalf("expected 0 failed checks, got %d", payload.Data.Failed)
	}
}

func TestCLI_DoctorFailsWhenManifestsMissing(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)
	// No module.yaml files, so doctor should fail

	stdout, _, err := runCLI(t, []string{"--format", "json", "doctor", "--dir", tmp}, "")
	if err == nil {
		t.Fatalf("expected doctor to fail when module.yaml files are missing")
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
			Failed int    `json:"failed"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" {
		t.Fatalf("expected error status, got %q", payload.Status)
	}
	if payload.Data.Failed == 0 {
		t.Fatal("expected failed > 0 when module.yaml files are missing")
	}
}

func TestCLI_DoctorWarnsWhenAgentTasksMissing(t *testing.T) {
	tmp := t.TempDir()
	writeMinimalRepo(t, tmp)

	// All stable roots with module.yaml (so no failures)
	for _, root := range []string{"core", "router", "contract", "middleware", "security", "store", "health", "log", "metrics"} {
		writeModuleYAML(t, tmp, root, nil)
	}
	// Also write the agent-first docs
	docsDir := filepath.Join(tmp, "docs", "concepts")
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		t.Fatalf("mkdir docs: %v", err)
	}
	for _, f := range []struct{ dir, name string }{
		{docsDir, "agent-first.md"},
		{docsDir, "core-boundary.md"},
		{docsDir, "extension-boundary.md"},
	} {
		if err := os.WriteFile(filepath.Join(f.dir, f.name), []byte("# Doc\n"), 0644); err != nil {
			t.Fatalf("write %s: %v", f.name, err)
		}
	}
	// No AGENT_TASKS.md for reference apps → should produce warnings

	stdout, _, err := runCLI(t, []string{"--format", "json", "doctor", "--dir", tmp}, "")
	// doctor with warnings returns exit code 2
	if err == nil {
		t.Fatalf("expected doctor to exit non-zero when AGENT_TASKS.md files are missing")
	}

	var payload struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
			Warned int    `json:"warned"`
			Failed int    `json:"failed"`
		} `json:"data"`
	}
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "warning" {
		t.Fatalf("expected warning status, got %q\noutput: %s", payload.Status, stdout)
	}
	if payload.Data.Warned == 0 {
		t.Fatal("expected warned > 0 when AGENT_TASKS.md files are missing")
	}
	if payload.Data.Failed != 0 {
		t.Fatalf("expected 0 failed, got %d", payload.Data.Failed)
	}
}

func TestCLI_DoctorRejectsUnexpectedArguments(t *testing.T) {
	stdout, _, err := runCLI(t, []string{"--format", "json", "doctor", "extra"}, "")
	if err == nil {
		t.Fatalf("expected unexpected arguments error")
	}

	var payload cliJSONEnvelope
	if e := json.Unmarshal([]byte(stdout), &payload); e != nil {
		t.Fatalf("parse output: %v\n%s", e, stdout)
	}
	if payload.Status != "error" || !strings.Contains(payload.Message, "unexpected arguments") {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
