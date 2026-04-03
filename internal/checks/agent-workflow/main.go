package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spcent/plumego/internal/checks/checkutil"
)

func main() {
	repoRoot, err := os.Getwd()
	if err != nil {
		failf("resolve working directory: %v", err)
	}

	violations, err := workflowViolations(repoRoot)
	if err != nil {
		failf("run agent workflow check: %v", err)
	}
	if len(violations) == 0 {
		return
	}

	fmt.Fprintln(os.Stderr, "agent-workflow check failed:")
	for _, violation := range violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func workflowViolations(repoRoot string) ([]string, error) {
	recipesDir := filepath.Join(repoRoot, "specs", "change-recipes")
	entries, err := os.ReadDir(recipesDir)
	if err != nil {
		return nil, err
	}

	var recipePaths []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		recipePaths = append(recipePaths, filepath.ToSlash(filepath.Join("specs", "change-recipes", entry.Name())))
	}
	sort.Strings(recipePaths)

	var violations []string
	if len(recipePaths) == 0 {
		violations = append(violations, "specs/change-recipes has no recipe files")
		return violations, nil
	}

	repoSpecPath := filepath.Join(repoRoot, "specs", "repo.yaml")
	content, err := os.ReadFile(repoSpecPath)
	if err != nil {
		return nil, err
	}
	repoSpec := string(content)

	if !strings.Contains(repoSpec, "specs/change-recipes") {
		violations = append(violations, "specs/repo.yaml does not declare specs/change-recipes as a machine-readable workflow source")
	}

	for _, recipePath := range recipePaths {
		if !strings.Contains(repoSpec, recipePath) {
			violations = append(violations, fmt.Sprintf("specs/repo.yaml does not reference workflow recipe %s", recipePath))
		}
	}

	declaredExtensions, err := checkutil.ReadRepoExtensionRoots(repoRoot)
	if err != nil {
		return nil, err
	}
	orphans, err := checkutil.FindOrphanedExtensionRoots(repoRoot, declaredExtensions)
	if err != nil {
		return nil, err
	}
	for _, orphan := range orphans {
		violations = append(violations, fmt.Sprintf("extension root %s exists in x/ but is not declared in specs/repo.yaml layers.extension.paths", orphan))
	}

	emptyDirs, err := checkutil.FindEmptyMisleadingDirs(repoRoot)
	if err != nil {
		return nil, err
	}
	for _, dir := range emptyDirs {
		violations = append(violations, fmt.Sprintf("empty directory %s is misleading in a package tree; remove it or add real contents", dir))
	}

	canonicalEntrypoints, err := checkutil.ReadCanonicalExtensionEntrypoints(repoRoot)
	if err != nil {
		return nil, err
	}
	primerViolations, err := checkutil.FindExtensionPrimerCoverageViolations(repoRoot, canonicalEntrypoints)
	if err != nil {
		return nil, err
	}
	violations = append(violations, primerViolations...)

	packageIndex, err := checkutil.ReadPackageIndex(repoRoot)
	if err != nil {
		return nil, err
	}
	packageIndexViolations, err := checkutil.FindPackageIndexCoverageViolations(repoRoot, packageIndex)
	if err != nil {
		return nil, err
	}
	violations = append(violations, packageIndexViolations...)

	httpSurfaceViolations, err := checkutil.FindStableHTTPSurfaceViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, httpSurfaceViolations...)

	return violations, nil
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
