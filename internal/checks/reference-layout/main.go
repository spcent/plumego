package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/internal/checks/checkutil"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		failf("resolve working directory: %v", err)
	}

	baseline, err := checkutil.ReadBaseline(filepath.Join(repoRoot, "specs", "check-baseline", "reference-layout-legacy-roots.txt"))
	if err != nil {
		failf("read reference layout baseline: %v", err)
	}

	violations, err := checkutil.FindUnexpectedTopLevelDirs(repoRoot, checkutil.AllowedTopLevelDirs(), baseline)
	if err != nil {
		failf("check top-level layout: %v", err)
	}
	violations = append(violations, requiredPathViolations(repoRoot)...)

	if len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "reference-layout check failed:")
	for _, violation := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func requiredPathViolations(repoRoot string) []string {
	required := []string{
		"docs/CANONICAL_STYLE_GUIDE.md",
		"reference/standard-service",
		"specs/repo.yaml",
		"specs/dependency-rules.yaml",
	}

	var violations []string
	for _, rel := range required {
		if _, err := os.Stat(filepath.Join(repoRoot, rel)); err != nil {
			if os.IsNotExist(err) {
				violations = append(violations, "missing required path "+rel)
				continue
			}
			violations = append(violations, fmt.Sprintf("stat %s: %v", rel, err))
		}
	}
	return violations
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
