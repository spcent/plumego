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

	fmt.Fprint(os.Stderr, checkutil.FormatViolations("agent-workflow", violations))
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

	qualityViolations, err := agentQualityControlPlaneViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, qualityViolations...)

	taskQueueViolations, err := taskQueueLifecycleViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, taskQueueViolations...)

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

	taxonomy, err := checkutil.ReadExtensionTaxonomy(repoRoot)
	if err != nil {
		return nil, err
	}
	taxonomyViolations, err := checkutil.FindExtensionTaxonomyCoverageViolations(repoRoot, declaredExtensions, taxonomy)
	if err != nil {
		return nil, err
	}
	violations = append(violations, taxonomyViolations...)

	primerViolations, err := checkutil.FindExtensionPrimerCoverageViolations(repoRoot, taxonomy.CanonicalRoots)
	if err != nil {
		return nil, err
	}
	violations = append(violations, primerViolations...)

	taskRouting, err := checkutil.ReadTaskRouting(repoRoot)
	if err != nil {
		return nil, err
	}
	taskRoutingViolations, err := checkutil.FindTaskRoutingCoverageViolations(repoRoot, taskRouting)
	if err != nil {
		return nil, err
	}
	violations = append(violations, taskRoutingViolations...)

	packageIndex, err := checkutil.ReadPackageIndex(repoRoot)
	if err != nil {
		return nil, err
	}
	packageIndexViolations, err := checkutil.FindPackageIndexCoverageViolations(repoRoot, packageIndex)
	if err != nil {
		return nil, err
	}
	violations = append(violations, packageIndexViolations...)
	violations = append(violations, checkutil.FindExtensionHotspotCoverageViolations(taxonomy, packageIndex)...)

	httpSurfaceViolations, err := checkutil.FindStableHTTPSurfaceViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, httpSurfaceViolations...)

	return violations, nil
}

func taskQueueLifecycleViolations(repoRoot string) ([]string, error) {
	var violations []string

	cardDirs := map[string]string{
		"tasks/cards/active":  "active",
		"tasks/cards/blocked": "blocked",
	}
	for relDir, expectedState := range cardDirs {
		dirViolations, err := taskCardStateViolations(repoRoot, relDir, expectedState)
		if err != nil {
			return nil, err
		}
		violations = append(violations, dirViolations...)
	}

	activeMilestoneViolations, err := activeMilestoneOutcomeViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, activeMilestoneViolations...)

	doneMilestoneViolations, err := doneMilestoneOutcomeViolations(repoRoot)
	if err != nil {
		return nil, err
	}
	violations = append(violations, doneMilestoneViolations...)

	sort.Strings(violations)
	return violations, nil
}

func taskCardStateViolations(repoRoot, relDir, expectedState string) ([]string, error) {
	dir := filepath.Join(repoRoot, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{fmt.Sprintf("%s is missing from the task card lifecycle control plane", relDir)}, nil
		}
		return nil, err
	}

	var violations []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") || name == "README.md" {
			continue
		}
		relPath := filepath.ToSlash(filepath.Join(relDir, name))
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		state, ok := markdownField(string(content), "State")
		if !ok {
			violations = append(violations, fmt.Sprintf("%s is missing State: %s", relPath, expectedState))
			continue
		}
		if state != expectedState {
			violations = append(violations, fmt.Sprintf("%s has State: %s but lives under %s", relPath, state, relDir))
		}
	}
	return violations, nil
}

func activeMilestoneOutcomeViolations(repoRoot string) ([]string, error) {
	const relDir = "tasks/milestones/active"
	dir := filepath.Join(repoRoot, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{relDir + " is missing from the milestone lifecycle control plane"}, nil
		}
		return nil, err
	}

	var violations []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "M-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if hasMarkdownHeading(string(content), "Outcome") {
			relPath := filepath.ToSlash(filepath.Join(relDir, name))
			violations = append(violations, fmt.Sprintf("%s has an Outcome section but still lives under tasks/milestones/active", relPath))
		}
	}
	return violations, nil
}

func doneMilestoneOutcomeViolations(repoRoot string) ([]string, error) {
	const relDir = "tasks/milestones/done"
	dir := filepath.Join(repoRoot, filepath.FromSlash(relDir))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{relDir + " is missing from the milestone lifecycle control plane"}, nil
		}
		return nil, err
	}

	var violations []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasPrefix(name, "M-") || !strings.HasSuffix(name, ".md") {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, err
		}
		if !hasMarkdownHeading(string(content), "Outcome") {
			relPath := filepath.ToSlash(filepath.Join(relDir, name))
			violations = append(violations, fmt.Sprintf("%s is archived but has no Outcome section", relPath))
		}
	}
	return violations, nil
}

func markdownField(content, field string) (string, bool) {
	prefix := field + ":"
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), true
		}
	}
	return "", false
}

func hasMarkdownHeading(content, heading string) bool {
	want := "## " + heading
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func agentQualityControlPlaneViolations(repoRoot string) ([]string, error) {
	const qualityDoc = "docs/AGENT_CODE_QUALITY_RULES.md"
	const qualitySpec = "specs/agent-quality-rules.yaml"

	var violations []string
	for _, relPath := range []string{qualityDoc, qualitySpec} {
		if _, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(relPath))); err != nil {
			if os.IsNotExist(err) {
				violations = append(violations, fmt.Sprintf("%s is missing from the agent quality control plane", relPath))
				continue
			}
			return nil, err
		}
	}

	requiredRefs := map[string][]string{
		"AGENTS.md":                             {qualityDoc, qualitySpec},
		"docs/CODEX_WORKFLOW.md":                {qualityDoc, qualitySpec},
		"docs/README.md":                        {qualityDoc},
		"specs/checks.yaml":                     {qualitySpec},
		"specs/change-recipes/fix-bug.yaml":     {qualityDoc, qualitySpec},
		"specs/change-recipes/review-only.yaml": {qualityDoc},
		"specs/change-recipes/stable-root-boundary-review.yaml": {qualityDoc, qualitySpec},
		"specs/change-recipes/symbol-change.yaml":               {qualityDoc, qualitySpec},
		"specs/repo.yaml": {qualityDoc, qualitySpec},
	}
	for relPath, refs := range requiredRefs {
		content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
		if err != nil {
			return nil, err
		}
		text := string(content)
		for _, ref := range refs {
			if !strings.Contains(text, ref) {
				violations = append(violations, fmt.Sprintf("%s does not reference %s", relPath, ref))
			}
		}
	}

	return violations, nil
}

func failf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
