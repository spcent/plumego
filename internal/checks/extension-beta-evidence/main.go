package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spcent/plumego/internal/checks/checkutil"
)

const requiredReleaseCount = 2

var knownBlockers = map[string]struct{}{
	"release_history_missing": {},
	"api_snapshot_missing":    {},
	"owner_signoff_missing":   {},
}

type candidate struct {
	Module        string
	Owner         string
	CurrentStatus string
	EvidenceDoc   string
	ReleaseRefs   []string
	APISnapshots  []string
	OwnerSignoff  string
	Blockers      []string
}

type moduleManifest struct {
	Name   string
	Owner  string
	Status string
}

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		failf("resolve working directory: %v", err)
	}

	report, violations, err := run(repoRoot)
	if err != nil {
		failf("run extension beta evidence check: %v", err)
	}
	for _, line := range report {
		fmt.Println(line)
	}
	if len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "extension-beta-evidence check failed:")
	for _, violation := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func run(repoRoot string) ([]string, []string, error) {
	roots, err := checkutil.ReadRepoExtensionRoots(repoRoot)
	if err != nil {
		return nil, nil, err
	}

	candidates, err := readCandidates(filepath.Join(repoRoot, "specs", "extension-beta-evidence.yaml"))
	if err != nil {
		return nil, nil, err
	}
	if len(candidates) == 0 {
		return nil, []string{"specs/extension-beta-evidence.yaml has no candidates"}, nil
	}

	var report []string
	var violations []string
	for _, cand := range candidates {
		candViolations, err := validateCandidate(repoRoot, roots, cand)
		if err != nil {
			return nil, nil, err
		}
		violations = append(violations, candViolations...)
		report = append(report, candidateReport(cand))
	}
	sort.Strings(report)
	sort.Strings(violations)
	return report, violations, nil
}

func readCandidates(path string) ([]candidate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var candidates []candidate
	var current *candidate
	var currentList string
	inCandidates := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.TrimRight(scanner.Text(), " \t")
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		indent := len(raw) - len(strings.TrimLeft(raw, " "))

		if indent == 0 && trimmed == "candidates:" {
			inCandidates = true
			continue
		}
		if !inCandidates {
			continue
		}
		if indent == 0 {
			break
		}

		if indent == 2 && strings.HasPrefix(trimmed, "- module:") {
			if current != nil {
				candidates = append(candidates, *current)
			}
			current = &candidate{Module: yamlScalar(trimmed)}
			currentList = ""
			continue
		}
		if current == nil {
			continue
		}

		if indent == 4 {
			key, value, ok := strings.Cut(trimmed, ":")
			if !ok {
				continue
			}
			key = strings.TrimSpace(key)
			value = strings.TrimSpace(value)
			currentList = ""
			switch key {
			case "owner":
				current.Owner = trimYAMLValue(value)
			case "current_status":
				current.CurrentStatus = trimYAMLValue(value)
			case "evidence_doc":
				current.EvidenceDoc = trimYAMLValue(value)
			case "owner_signoff":
				current.OwnerSignoff = trimYAMLValue(value)
			case "release_refs":
				current.ReleaseRefs = parseInlineList(value)
				if value == "" {
					currentList = key
				}
			case "api_snapshots":
				current.APISnapshots = parseInlineList(value)
				if value == "" {
					currentList = key
				}
			case "blockers":
				current.Blockers = parseInlineList(value)
				if value == "" {
					currentList = key
				}
			}
			continue
		}

		if indent == 6 && currentList != "" && strings.HasPrefix(trimmed, "- ") {
			value := trimYAMLValue(strings.TrimSpace(strings.TrimPrefix(trimmed, "- ")))
			switch currentList {
			case "release_refs":
				current.ReleaseRefs = append(current.ReleaseRefs, value)
			case "api_snapshots":
				current.APISnapshots = append(current.APISnapshots, value)
			case "blockers":
				current.Blockers = append(current.Blockers, value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil {
		candidates = append(candidates, *current)
	}
	return candidates, nil
}

func validateCandidate(repoRoot string, roots map[string]struct{}, cand candidate) ([]string, error) {
	var violations []string
	if cand.Module == "" {
		return []string{"candidate missing module"}, nil
	}
	if _, ok := roots[cand.Module]; !ok {
		violations = append(violations, fmt.Sprintf("%s is not declared in specs/repo.yaml extension paths", cand.Module))
	}

	manifest, err := readModuleManifest(filepath.Join(repoRoot, cand.Module, "module.yaml"))
	if err != nil {
		return nil, err
	}
	if manifest.Name != cand.Module {
		violations = append(violations, fmt.Sprintf("%s module.yaml name %q does not match evidence module", cand.Module, manifest.Name))
	}
	if cand.Owner == "" {
		violations = append(violations, fmt.Sprintf("%s missing owner", cand.Module))
	} else if manifest.Owner != cand.Owner {
		violations = append(violations, fmt.Sprintf("%s owner mismatch: evidence %q, module.yaml %q", cand.Module, cand.Owner, manifest.Owner))
	}
	if cand.CurrentStatus == "" {
		violations = append(violations, fmt.Sprintf("%s missing current_status", cand.Module))
	} else if manifest.Status != cand.CurrentStatus {
		violations = append(violations, fmt.Sprintf("%s status mismatch: evidence %q, module.yaml %q", cand.Module, cand.CurrentStatus, manifest.Status))
	}
	if cand.EvidenceDoc == "" {
		violations = append(violations, fmt.Sprintf("%s missing evidence_doc", cand.Module))
	} else if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(cand.EvidenceDoc))); err != nil {
		if os.IsNotExist(err) {
			violations = append(violations, fmt.Sprintf("%s evidence_doc does not exist: %s", cand.Module, cand.EvidenceDoc))
		} else {
			return nil, err
		}
	}

	violations = append(violations, blockerViolations(cand)...)
	return violations, nil
}

func blockerViolations(cand candidate) []string {
	expected := expectedBlockers(cand)
	actual := stringSet(cand.Blockers)

	var violations []string
	for blocker := range actual {
		if _, ok := knownBlockers[blocker]; !ok {
			violations = append(violations, fmt.Sprintf("%s has unknown blocker %q", cand.Module, blocker))
		}
	}
	for blocker := range expected {
		if _, ok := actual[blocker]; !ok {
			violations = append(violations, fmt.Sprintf("%s missing blocker %q", cand.Module, blocker))
		}
	}
	for blocker := range actual {
		if _, ok := expected[blocker]; !ok {
			violations = append(violations, fmt.Sprintf("%s has stale blocker %q", cand.Module, blocker))
		}
	}
	return violations
}

func expectedBlockers(cand candidate) map[string]struct{} {
	expected := map[string]struct{}{}
	if len(cand.ReleaseRefs) < requiredReleaseCount {
		expected["release_history_missing"] = struct{}{}
	}
	if len(cand.APISnapshots) < requiredReleaseCount {
		expected["api_snapshot_missing"] = struct{}{}
	}
	if !isOwnerSignoffComplete(cand.OwnerSignoff) {
		expected["owner_signoff_missing"] = struct{}{}
	}
	return expected
}

func isOwnerSignoffComplete(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "signed", "complete", "approved":
		return true
	default:
		return false
	}
}

func candidateReport(cand candidate) string {
	blockers := append([]string(nil), cand.Blockers...)
	sort.Strings(blockers)
	if len(blockers) == 0 {
		blockers = []string{"none"}
	}
	return fmt.Sprintf("%s\tstatus=%s\towner=%s\trelease_refs=%d\tapi_snapshots=%d\tblockers=%s",
		cand.Module,
		cand.CurrentStatus,
		cand.Owner,
		len(cand.ReleaseRefs),
		len(cand.APISnapshots),
		strings.Join(blockers, ","),
	)
}

func readModuleManifest(path string) (moduleManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return moduleManifest{}, err
	}
	defer file.Close()

	var manifest moduleManifest
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		switch {
		case strings.HasPrefix(line, "name:"):
			manifest.Name = yamlScalar(line)
		case strings.HasPrefix(line, "owner:"):
			manifest.Owner = yamlScalar(line)
		case strings.HasPrefix(line, "status:"):
			manifest.Status = yamlScalar(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return moduleManifest{}, err
	}
	return manifest, nil
}

func parseInlineList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return nil
	}
	value = strings.TrimPrefix(value, "[")
	value = strings.TrimSuffix(value, "]")

	var out []string
	for _, part := range strings.Split(value, ",") {
		if item := trimYAMLValue(part); item != "" {
			out = append(out, item)
		}
	}
	return out
}

func stringSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		out[value] = struct{}{}
	}
	return out
}

func yamlScalar(line string) string {
	_, value, _ := strings.Cut(line, ":")
	return trimYAMLValue(value)
}

func trimYAMLValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
