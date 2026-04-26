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

type moduleState struct {
	Name   string
	Status string
	Risk   string
}

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		failf("resolve working directory: %v", err)
	}

	violations, err := maturityViolations(repoRoot)
	if err != nil {
		failf("run extension maturity check: %v", err)
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
	roots, err := checkutil.ReadRepoExtensionRoots(repoRoot)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath.Join(repoRoot, "docs", "EXTENSION_MATURITY.md"))
	if err != nil {
		return nil, err
	}
	dashboard := string(content)

	var paths []string
	for path := range roots {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	var violations []string
	for _, path := range paths {
		state, err := readModuleState(filepath.Join(repoRoot, path, "module.yaml"))
		if err != nil {
			return nil, err
		}
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
	}

	return violations, nil
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
	return state, nil
}

func yamlScalar(line string) string {
	_, value, _ := strings.Cut(line, ":")
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
