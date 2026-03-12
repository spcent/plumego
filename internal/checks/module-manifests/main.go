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

	baseline, err := checkutil.ReadBaseline(filepath.Join(repoRoot, "specs", "check-baseline", "missing-module-manifests.txt"))
	if err != nil {
		failf("read module manifest baseline: %v", err)
	}

	missing, err := checkutil.FindMissingModuleManifests(repoRoot, baseline)
	if err != nil {
		failf("find missing module manifests: %v", err)
	}
	violations, err := checkutil.ValidateModuleManifests(repoRoot)
	if err != nil {
		failf("validate module manifests: %v", err)
	}

	if len(missing) == 0 && len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "module-manifests check failed:")
	for _, item := range missing {
		fmt.Fprintf(os.Stderr, "- missing module.yaml for %s\n", item)
	}
	for _, item := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", item)
	}
	os.Exit(1)
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
