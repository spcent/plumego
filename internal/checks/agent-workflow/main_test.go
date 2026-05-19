package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestAgentQualityControlPlaneViolationsPass(t *testing.T) {
	repoRoot := t.TempDir()
	writeAgentQualityControlPlane(t, repoRoot, nil)

	violations, err := agentQualityControlPlaneViolations(repoRoot)
	if err != nil {
		t.Fatalf("agentQualityControlPlaneViolations returned error: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestAgentQualityControlPlaneViolationsReportsMissingFile(t *testing.T) {
	repoRoot := t.TempDir()
	writeAgentQualityControlPlane(t, repoRoot, []string{"specs/agent-quality-rules.yaml"})

	violations, err := agentQualityControlPlaneViolations(repoRoot)
	if err != nil {
		t.Fatalf("agentQualityControlPlaneViolations returned error: %v", err)
	}
	if !slices.Contains(violations, "specs/agent-quality-rules.yaml is missing from the agent quality control plane") {
		t.Fatalf("expected missing quality spec violation, got %v", violations)
	}
}

func TestAgentQualityControlPlaneViolationsReportsMissingReference(t *testing.T) {
	repoRoot := t.TempDir()
	writeAgentQualityControlPlane(t, repoRoot, nil)
	writeFile(t, repoRoot, "AGENTS.md", "docs/AGENT_CODE_QUALITY_RULES.md\n")

	violations, err := agentQualityControlPlaneViolations(repoRoot)
	if err != nil {
		t.Fatalf("agentQualityControlPlaneViolations returned error: %v", err)
	}
	if !slices.Contains(violations, "AGENTS.md does not reference specs/agent-quality-rules.yaml") {
		t.Fatalf("expected missing AGENTS quality spec reference, got %v", violations)
	}
}

func TestTaskQueueLifecycleViolationsPass(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, repoRoot, "tasks/cards/active/README.md", "# Active\n")
	writeFile(t, repoRoot, "tasks/cards/active/0001-ready.md", "# Card 0001\nState: active\n")
	writeFile(t, repoRoot, "tasks/cards/blocked/0002-waiting.md", "# Card 0002\nState: blocked\n")
	writeFile(t, repoRoot, "tasks/cards/done/0003-finished.md", "# Card 0003\nState: done\n")
	writeFile(t, repoRoot, "tasks/milestones/active/.gitkeep", "")
	writeFile(t, repoRoot, "tasks/milestones/active/M-002-next/M-002.md", "# M-002\n")
	writeFile(t, repoRoot, "tasks/milestones/done/M-001-done/M-001.md", "# M-001\n\n## Outcome\n\nDone.\n")

	violations, err := taskQueueLifecycleViolations(repoRoot)
	if err != nil {
		t.Fatalf("taskQueueLifecycleViolations returned error: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("expected no violations, got %v", violations)
	}
}

func TestTaskQueueLifecycleViolationsReportDirectoryDrift(t *testing.T) {
	repoRoot := t.TempDir()
	writeFile(t, repoRoot, "tasks/cards/active/0001-waiting.md", "# Card 0001\nState: blocked\n")
	writeFile(t, repoRoot, "tasks/cards/blocked/0002-ready.md", "# Card 0002\nState: active\n")
	writeFile(t, repoRoot, "tasks/cards/done/0003-stale.md", "# Card 0003\nState: active\n")
	writeFile(t, repoRoot, "tasks/milestones/active/M-001-next/M-001.md", "# M-001\n\n## Outcome\n\nDone.\n")
	writeFile(t, repoRoot, "tasks/milestones/done/M-002-done/M-002.md", "# M-002\n")

	violations, err := taskQueueLifecycleViolations(repoRoot)
	if err != nil {
		t.Fatalf("taskQueueLifecycleViolations returned error: %v", err)
	}
	expected := []string{
		"tasks/cards/active/0001-waiting.md has State: blocked but lives under tasks/cards/active",
		"tasks/cards/blocked/0002-ready.md has State: active but lives under tasks/cards/blocked",
		"tasks/cards/done/0003-stale.md has State: active but lives under tasks/cards/done",
		"tasks/milestones/active/M-001-next/M-001.md has an Outcome section but still lives under tasks/milestones/active",
		"tasks/milestones/done/M-002-done/M-002.md is archived but has no Outcome section",
	}
	for _, want := range expected {
		if !slices.Contains(violations, want) {
			t.Fatalf("expected violation %q, got %v", want, violations)
		}
	}
}

func writeAgentQualityControlPlane(t *testing.T, repoRoot string, omit []string) {
	t.Helper()

	files := map[string]string{
		"docs/AGENT_CODE_QUALITY_RULES.md":                      "quality rules\n",
		"specs/agent-quality-rules.yaml":                        "version: 1\n",
		"AGENTS.md":                                             "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
		"docs/CODEX_WORKFLOW.md":                                "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
		"docs/README.md":                                        "docs/AGENT_CODE_QUALITY_RULES.md\n",
		"specs/checks.yaml":                                     "specs/agent-quality-rules.yaml\n",
		"specs/change-recipes/fix-bug.yaml":                     "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
		"specs/change-recipes/review-only.yaml":                 "docs/AGENT_CODE_QUALITY_RULES.md\n",
		"specs/change-recipes/stable-root-boundary-review.yaml": "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
		"specs/change-recipes/symbol-change.yaml":               "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
		"specs/repo.yaml":                                       "docs/AGENT_CODE_QUALITY_RULES.md\nspecs/agent-quality-rules.yaml\n",
	}
	for path, content := range files {
		if slices.Contains(omit, path) {
			continue
		}
		writeFile(t, repoRoot, path, content)
	}
}

func writeFile(t *testing.T, repoRoot, relPath, content string) {
	t.Helper()

	path := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("create parent dir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(path, []byte(strings.TrimLeft(content, "\n")), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}
