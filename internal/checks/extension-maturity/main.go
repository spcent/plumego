package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spcent/plumego/internal/checks/checkutil"
)

type moduleState struct {
	Name   string
	Owner  string
	Status string
	Risk   string
}

type betaCandidate struct {
	Module      string
	EvidenceDoc string
	Blockers    []string
}

func main() {
	report := flag.Bool("report", false, "print deterministic dashboard source data")
	flag.Parse()

	repoRoot, err := os.Getwd()
	if err != nil {
		failf("resolve working directory: %v", err)
	}

	lines, violations, err := maturityReport(repoRoot)
	if err != nil {
		failf("run extension maturity check: %v", err)
	}
	if *report {
		for _, line := range lines {
			fmt.Println(line)
		}
	}
	if len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "extension-maturity check failed:")
	for _, violation := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func maturityViolations(repoRoot string) ([]string, error) {
	_, violations, err := maturityReport(repoRoot)
	return violations, err
}

func maturityReport(repoRoot string) ([]string, []string, error) {
	roots, err := checkutil.ReadRepoExtensionRoots(repoRoot)
	if err != nil {
		return nil, nil, err
	}

	content, err := os.ReadFile(filepath.Join(repoRoot, "docs", "EXTENSION_MATURITY.md"))
	if err != nil {
		return nil, nil, err
	}
	dashboard := string(content)
	candidates, err := readBetaCandidates(filepath.Join(repoRoot, "specs", "extension-beta-evidence.yaml"))
	if err != nil {
		return nil, nil, err
	}

	var paths []string
	for path := range roots {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var report []string
	var violations []string
	for _, path := range paths {
		state, err := readModuleState(filepath.Join(repoRoot, path, "module.yaml"))
		if err != nil {
			return nil, nil, err
		}
		report = append(report, fmt.Sprintf("%s\tstatus=%s\trisk=%s\towner=%s", path, state.Status, state.Risk, state.Owner))
		if state.Name != "" && state.Name != path {
			violations = append(violations, fmt.Sprintf("%s module.yaml name %q does not match declared path", path, state.Name))
		}
		row, ok := findDashboardRow(dashboard, path)
		if !ok {
			violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md missing row for %s", path))
			continue
		}
		if !strings.Contains(row, "| "+state.Status+" |") {
			violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s missing status %q", path, state.Status))
		}
		if !strings.Contains(row, "| "+state.Risk+" |") {
			violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s missing risk %q", path, state.Risk))
		}
		if !strings.Contains(row, "| "+state.Owner+" |") {
			violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s missing owner %q", path, state.Owner))
		}
		if candidate, ok := candidates[path]; ok {
			violations = append(violations, candidateDashboardViolations(row, candidate)...)
		}
	}

	sort.Strings(report)
	sort.Strings(violations)
	return report, violations, nil
}

func readModuleState(path string) (moduleState, error) {
	file, err := os.Open(path)
	if err != nil {
		return moduleState{}, err
	}
	defer file.Close()

	var state moduleState
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "name:") {
			state.Name = yamlScalar(line)
		}
		if strings.HasPrefix(line, "owner:") {
			state.Owner = yamlScalar(line)
		}
		if strings.HasPrefix(line, "status:") {
			state.Status = yamlScalar(line)
		}
		if strings.HasPrefix(line, "risk:") {
			state.Risk = yamlScalar(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return moduleState{}, err
	}
	if state.Status == "" {
		return moduleState{}, fmt.Errorf("%s missing status", path)
	}
	if state.Risk == "" {
		return moduleState{}, fmt.Errorf("%s missing risk", path)
	}
	if state.Owner == "" {
		return moduleState{}, fmt.Errorf("%s missing owner", path)
	}
	return state, nil
}

func readBetaCandidates(path string) (map[string]betaCandidate, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	candidates := map[string]betaCandidate{}
	var current *betaCandidate
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
				candidates[current.Module] = *current
			}
			current = &betaCandidate{Module: yamlScalar(trimmed)}
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
			case "evidence_doc":
				current.EvidenceDoc = trimYAMLValue(value)
			case "blockers":
				current.Blockers = parseInlineList(value)
				if value == "" {
					currentList = key
				}
			}
			continue
		}
		if indent == 6 && currentList == "blockers" && strings.HasPrefix(trimmed, "- ") {
			current.Blockers = append(current.Blockers, trimYAMLValue(strings.TrimPrefix(trimmed, "- ")))
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if current != nil {
		candidates[current.Module] = *current
	}
	return candidates, nil
}

func candidateDashboardViolations(row string, candidate betaCandidate) []string {
	var violations []string
	link := strings.TrimPrefix(filepath.ToSlash(candidate.EvidenceDoc), "docs/")
	if candidate.EvidenceDoc == "" {
		violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s has beta candidate without evidence_doc", candidate.Module))
	} else if !strings.Contains(row, "("+link+")") {
		violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s missing evidence link %s", candidate.Module, link))
	}

	requiredText := map[string]string{
		"release_history_missing": "release history",
		"api_snapshot_missing":    "API snapshot",
		"owner_signoff_missing":   "owner sign-off",
	}
	for _, blocker := range candidate.Blockers {
		text, ok := requiredText[blocker]
		if !ok {
			continue
		}
		if !strings.Contains(row, text) {
			violations = append(violations, fmt.Sprintf("docs/EXTENSION_MATURITY.md row for %s missing blocker text %q", candidate.Module, text))
		}
	}
	return violations
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

func yamlScalar(line string) string {
	_, value, _ := strings.Cut(line, ":")
	return trimYAMLValue(value)
}

func trimYAMLValue(value string) string {
	return strings.Trim(strings.TrimSpace(value), `"'`)
}

func findDashboardRow(dashboard, path string) (string, bool) {
	needle := "| `" + path + "` |"
	for _, line := range strings.Split(dashboard, "\n") {
		if strings.Contains(line, needle) {
			return line, true
		}
	}
	return "", false
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
