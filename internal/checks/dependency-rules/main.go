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

	baseline, err := checkutil.ReadBaseline(filepath.Join(repoRoot, "specs", "check-baseline", "dependency-rules.txt"))
	if err != nil {
		failf("read dependency baseline: %v", err)
	}

	violations, err := checkutil.FindDisallowedImports(repoRoot, baseline)
	if err != nil {
		failf("run dependency rules check: %v", err)
	}
	controlPlaneViolations, err := checkutil.ValidateManifestDependencyRuleConsistency(repoRoot)
	if err != nil {
		failf("validate dependency rule control plane: %v", err)
	}
	violations = append(violations, controlPlaneViolations...)
	if len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "dependency-rules check failed:")
	for _, violation := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
