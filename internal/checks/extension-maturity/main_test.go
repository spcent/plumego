package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ── invalidStatusViolations ───────────────────────────────────────────────────

func TestInvalidStatusViolationsPassWithValidStatuses(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "experimental", Risk: "medium", Owner: "team"},
		"x/bar": {Status: "beta", Risk: "medium", Owner: "team"},
		"x/baz": {Status: "ga", Risk: "low", Owner: "team"},
	}
	if v := invalidStatusViolations(states); len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestInvalidStatusViolationsDetectsInvalidStatus(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "preview", Risk: "medium", Owner: "team"},
	}
	v := invalidStatusViolations(states)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %v", v)
	}
	if !strings.Contains(v[0], "x/foo") || !strings.Contains(v[0], "preview") {
		t.Fatalf("unexpected violation: %q", v[0])
	}
}

func TestInvalidStatusViolationsDetectsMultipleInvalidStatuses(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "stable", Risk: "medium", Owner: "team"},
		"x/bar": {Status: "rc", Risk: "medium", Owner: "team"},
		"x/baz": {Status: "beta", Risk: "medium", Owner: "team"},
	}
	v := invalidStatusViolations(states)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %v", v)
	}
}

// ── betaEvidenceViolations ────────────────────────────────────────────────────

func TestBetaEvidenceViolationsPassWhenAllEvidenced(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
		"x/bar": {Status: "experimental"},
	}
	completed := map[string]bool{"x/foo": true}
	if v := betaEvidenceViolations(states, completed); len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestBetaEvidenceViolationsPassForExperimentalWithoutEvidence(t *testing.T) {
	states := map[string]moduleState{
		"x/bar": {Status: "experimental"},
	}
	completed := map[string]bool{}
	if v := betaEvidenceViolations(states, completed); len(v) != 0 {
		t.Fatalf("expected no violations for experimental module, got %v", v)
	}
}

func TestBetaEvidenceViolationsDetectsMissingEvidence(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
	}
	completed := map[string]bool{}
	v := betaEvidenceViolations(states, completed)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %v", v)
	}
	if !strings.Contains(v[0], "x/foo") {
		t.Fatalf("unexpected violation: %q", v[0])
	}
	if !strings.Contains(v[0], "specs/extension-beta-evidence.yaml") {
		t.Fatalf("violation should reference evidence file: %q", v[0])
	}
}

func TestBetaEvidenceViolationsDetectsMultipleMissing(t *testing.T) {
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
		"x/bar": {Status: "beta"},
		"x/baz": {Status: "experimental"},
	}
	completed := map[string]bool{"x/foo": true}
	v := betaEvidenceViolations(states, completed)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation (only x/bar missing), got %v", v)
	}
	if !strings.Contains(v[0], "x/bar") {
		t.Fatalf("expected violation for x/bar, got %q", v[0])
	}
}

// ── readmeBetaListViolations ──────────────────────────────────────────────────

func TestReadmeBetaListViolationsPass(t *testing.T) {
	readme := "## Current Status\n\n" +
		"Seven `x/*` extension families are **beta** — stable across release refs\n" +
		"and suitable for production adoption: `x/foo`, `x/bar`, and `x/baz`.\n\n" +
		"All remaining `x/*` extensions are **experimental**.\n"
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
		"x/bar": {Status: "beta"},
		"x/baz": {Status: "beta"},
	}
	if v := readmeBetaListViolations(readme, states); len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestReadmeBetaListViolationsDetectsModuleMissingFromReadme(t *testing.T) {
	readme := "Seven `x/*` extension families are **beta** — stable across release refs\n" +
		"and suitable for production adoption: `x/foo`.\n"
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
		"x/bar": {Status: "beta"},
	}
	v := readmeBetaListViolations(readme, states)
	if !containsSubstring(v, "x/bar") {
		t.Fatalf("expected violation mentioning x/bar, got %v", v)
	}
}

func TestReadmeBetaListViolationsDetectsNonBetaListedInReadme(t *testing.T) {
	readme := "Seven `x/*` extension families are **beta** — stable across release refs\n" +
		"and suitable for production adoption: `x/foo` and `x/baz`.\n"
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
		"x/baz": {Status: "experimental"},
	}
	v := readmeBetaListViolations(readme, states)
	if !containsSubstring(v, "x/baz") {
		t.Fatalf("expected violation mentioning x/baz, got %v", v)
	}
}

func TestReadmeBetaListViolationsReportsMissingParagraph(t *testing.T) {
	readme := "# Plumego\n\nSome text without the beta families paragraph.\n"
	states := map[string]moduleState{
		"x/foo": {Status: "beta"},
	}
	v := readmeBetaListViolations(readme, states)
	if len(v) == 0 {
		t.Fatal("expected violation for missing beta paragraph, got none")
	}
}

// ── parseReadmeBetaModules ────────────────────────────────────────────────────

func TestParseReadmeBetaModulesExtractsModules(t *testing.T) {
	content := "Some text.\n\n" +
		"Seven `x/*` extension families are **beta** — stable across release refs\n" +
		"and suitable for production: `x/frontend`, `x/gateway`, and `x/rest`.\n\n" +
		"All remaining are **experimental**.\n"
	got := parseReadmeBetaModules(content)
	want := []string{"x/frontend", "x/gateway", "x/rest"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Fatalf("index %d: got %q, want %q", i, got[i], w)
		}
	}
}

func TestParseReadmeBetaModulesSkipsWildcard(t *testing.T) {
	content := "Seven `x/*` extension families are **beta** — test\n" +
		"caveats: `x/foo`.\n"
	got := parseReadmeBetaModules(content)
	for _, m := range got {
		if strings.Contains(m, "*") {
			t.Fatalf("result should not include wildcard entry, got %v", got)
		}
	}
}

func TestParseReadmeBetaModulesStopsAtBlankLine(t *testing.T) {
	content := "Seven `x/*` extension families are **beta** — test\n" +
		"caveats: `x/foo`.\n\n" +
		"This paragraph should not be scanned: `x/bar`.\n"
	got := parseReadmeBetaModules(content)
	for _, m := range got {
		if m == "x/bar" {
			t.Fatalf("parser should stop at blank line, got %v", got)
		}
	}
}

// ── surfaceCandidateStatusViolations ─────────────────────────────────────────

func TestSurfaceCandidateStatusViolationsPassWhenMatching(t *testing.T) {
	root := t.TempDir()
	writeTestModuleYAML(t, root, "x/data/file", "beta")

	cands := []surfaceCandidateStatus{
		{Module: "x/data", Surface: "file", Package: "x/data/file", CurrentStatus: "beta"},
	}
	v, err := surfaceCandidateStatusViolations(root, cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestSurfaceCandidateStatusViolationsDetectsMismatch(t *testing.T) {
	root := t.TempDir()
	writeTestModuleYAML(t, root, "x/data/file", "beta")

	cands := []surfaceCandidateStatus{
		{Module: "x/data", Surface: "file", Package: "x/data/file", CurrentStatus: "experimental"},
	}
	v, err := surfaceCandidateStatusViolations(root, cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %v", v)
	}
	if !strings.Contains(v[0], "x/data:file") {
		t.Fatalf("violation should mention surface label, got %q", v[0])
	}
	if !strings.Contains(v[0], "experimental") || !strings.Contains(v[0], "beta") {
		t.Fatalf("violation should mention both statuses, got %q", v[0])
	}
}

func TestSurfaceCandidateStatusViolationsSkipsWhenNoPackageManifest(t *testing.T) {
	root := t.TempDir()
	// No module.yaml written for x/data/file — should be silently skipped.
	cands := []surfaceCandidateStatus{
		{Module: "x/data", Surface: "file", Package: "x/data/file", CurrentStatus: "experimental"},
	}
	v, err := surfaceCandidateStatusViolations(root, cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("expected no violations when package has no module.yaml, got %v", v)
	}
}

func TestSurfaceCandidateStatusViolationsSkipsWhenPackageEmpty(t *testing.T) {
	root := t.TempDir()
	cands := []surfaceCandidateStatus{
		{Module: "x/data", Surface: "file", Package: "", CurrentStatus: "experimental"},
	}
	v, err := surfaceCandidateStatusViolations(root, cands)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("expected no violations when package is empty, got %v", v)
	}
}

// ── readCompletedBetaModules ──────────────────────────────────────────────────

func TestReadCompletedBetaModulesRecognizesBetaCandidate(t *testing.T) {
	content := "version: 1\ncandidates:\n  - module: x/foo\n    current_status: beta\n    evidence_doc: docs/evidence/extension/x-foo.md\n    blockers: []\n"
	path := writeTemp(t, content)

	got, err := readCompletedBetaModules(path)
	if err != nil {
		t.Fatal(err)
	}
	if !got["x/foo"] {
		t.Fatalf("expected x/foo to be completed, got %v", got)
	}
}

func TestReadCompletedBetaModulesExcludesWhenHasBlockers(t *testing.T) {
	content := "version: 1\ncandidates:\n  - module: x/foo\n    current_status: beta\n    evidence_doc: docs/evidence/extension/x-foo.md\n    blockers:\n      - api_snapshot_missing\n"
	path := writeTemp(t, content)

	got, err := readCompletedBetaModules(path)
	if err != nil {
		t.Fatal(err)
	}
	if got["x/foo"] {
		t.Fatalf("x/foo should not be completed when it has blockers")
	}
}

func TestReadCompletedBetaModulesExcludesExperimental(t *testing.T) {
	content := "version: 1\ncandidates:\n  - module: x/foo\n    current_status: experimental\n    evidence_doc: docs/evidence/extension/x-foo.md\n    blockers: []\n"
	path := writeTemp(t, content)

	got, err := readCompletedBetaModules(path)
	if err != nil {
		t.Fatal(err)
	}
	if got["x/foo"] {
		t.Fatalf("x/foo with experimental status should not be completed")
	}
}

func TestReadCompletedBetaModulesReadsSurfaceCandidates(t *testing.T) {
	content := "version: 1\n" +
		"candidates:\n  - module: x/bar\n    current_status: experimental\n    evidence_doc: docs/evidence/extension/x-bar.md\n    blockers: []\n" +
		"surface_candidates:\n  - module: x/baz\n    surface: svc\n    package: x/baz\n    current_status: beta\n    evidence_doc: docs/evidence/extension/x-baz.md\n    blockers: []\n"
	path := writeTemp(t, content)

	got, err := readCompletedBetaModules(path)
	if err != nil {
		t.Fatal(err)
	}
	if got["x/bar"] {
		t.Fatalf("x/bar is experimental, should not be completed")
	}
	if !got["x/baz"] {
		t.Fatalf("x/baz surface_candidate with beta status should be completed")
	}
}

// ── readSurfaceCandidateStatuses ──────────────────────────────────────────────

func TestReadSurfaceCandidateStatusesParsesFields(t *testing.T) {
	content := "version: 1\nsurface_candidates:\n  - module: x/data\n    surface: file\n    package: x/data/file\n    current_status: experimental\n    evidence_doc: docs/evidence/extension/x-data.md\n    blockers: []\n"
	path := writeTemp(t, content)

	got, err := readSurfaceCandidateStatuses(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d: %v", len(got), got)
	}
	c := got[0]
	if c.Module != "x/data" || c.Surface != "file" || c.Package != "x/data/file" || c.CurrentStatus != "experimental" {
		t.Fatalf("unexpected candidate: %+v", c)
	}
	if c.Label() != "x/data:file" {
		t.Fatalf("unexpected label: %q", c.Label())
	}
}

func TestReadSurfaceCandidateStatusesSkipsOtherSections(t *testing.T) {
	content := "version: 1\n" +
		"candidates:\n  - module: x/foo\n    current_status: beta\n    blockers: []\n" +
		"surface_candidates:\n  - module: x/bar\n    surface: svc\n    package: x/bar\n    current_status: beta\n    blockers: []\n" +
		"subpackage_candidates:\n  - module: x/baz\n    subpackage: sub\n    current_tier: stable\n    blockers: []\n"
	path := writeTemp(t, content)

	got, err := readSurfaceCandidateStatuses(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Module != "x/bar" {
		t.Fatalf("expected only x/bar surface candidate, got %v", got)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeTestModuleYAML(t *testing.T, repoRoot, relPath, status string) {
	t.Helper()
	fullPath := filepath.Join(repoRoot, filepath.FromSlash(relPath), "module.yaml")
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	content := "name: " + relPath + "\npath: " + relPath + "\nlayer: extension\nstatus: " + status + "\nowner: test\nrisk: medium\n" +
		"summary: test\nresponsibilities:\n  - test\nnon_goals:\n  - test\n" +
		"allowed_imports:\n  - stdlib\nforbidden_imports:\n  - x/other\n" +
		"test_commands:\n  - go test ./...\nreview_checklist:\n  - check\nagent_hints:\n  - hint\n"
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
}

func writeTemp(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.yaml")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}
	return f.Name()
}

func containsSubstring(violations []string, sub string) bool {
	for _, v := range violations {
		if strings.Contains(v, sub) {
			return true
		}
	}
	return false
}
