// Package main implements the docs-sync-contract checker.
//
// It walks all module.yaml files and, for each that declares a
// docs_sync_contract field, verifies that every path listed under
// on_status_change, on_api_change, and on_behavior_change actually exists in
// the repository. A missing file means the contract was written for a doc that
// was later moved, renamed, or never created.
//
// This makes docs_sync_contract machine-readable: agents can query which docs
// must update for a given change type, and CI can verify the referenced paths
// are not stale.
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
		fatalf("resolve working directory: %v", err)
	}

	violations, err := checkDocsSyncContracts(repoRoot)
	if err != nil {
		fatalf("check docs_sync_contract: %v", err)
	}
	if len(violations) == 0 {
		return
	}

	fmt.Fprint(os.Stderr, checkutil.FormatViolations("docs-sync-contract", violations))
	os.Exit(1)
}

// checkDocsSyncContracts walks all module.yaml files in the repository and
// validates that docs_sync_contract path references exist on disk.
func checkDocsSyncContracts(repoRoot string) ([]string, error) {
	var violations []string

	// Walk both stable roots and x/*
	searchDirs := make([]string, 0, 20)
	for _, root := range checkutil.StableRoots {
		searchDirs = append(searchDirs, filepath.Join(repoRoot, root))
	}
	searchDirs = append(searchDirs, filepath.Join(repoRoot, "x"))

	for _, dir := range searchDirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() || d.Name() != "module.yaml" {
				return nil
			}

			viols, err := checkFile(repoRoot, path)
			if err != nil {
				return err
			}
			violations = append(violations, viols...)
			return nil
		})
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
	}

	sort.Strings(violations)
	return violations, nil
}

// checkFile validates the docs_sync_contract entries in a single module.yaml.
func checkFile(repoRoot, manifestPath string) ([]string, error) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	rel, _ := filepath.Rel(repoRoot, manifestPath)
	rel = filepath.ToSlash(rel)

	if !strings.Contains(string(content), "docs_sync_contract") {
		return nil, nil
	}

	// Parse docs_sync_contract paths using a simple line-based scanner.
	// The structure is:
	//   docs_sync_contract:
	//     on_status_change:
	//       - path/to/file
	//     on_api_change:
	//       - path/to/file
	//     on_behavior_change:
	//       - path/to/file
	var violations []string
	inContract := false
	inSection := false

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Detect top-level field (no leading whitespace)
		if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			key := strings.TrimSuffix(strings.TrimSpace(strings.SplitN(line, ":", 2)[0]), "")
			inContract = key == "docs_sync_contract"
			inSection = false
			continue
		}

		if !inContract {
			continue
		}

		// Detect sub-section keys (on_status_change, on_api_change, on_behavior_change)
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") {
			key := strings.TrimSpace(strings.SplitN(line, ":", 2)[0])
			inSection = key == "on_status_change" || key == "on_api_change" || key == "on_behavior_change"
			continue
		}

		if !inSection {
			continue
		}

		// List items under a section
		if strings.HasPrefix(trimmed, "- ") {
			docPath := strings.TrimPrefix(trimmed, "- ")
			docPath = strings.TrimSpace(docPath)
			if docPath == "" {
				continue
			}
			fullPath := filepath.Join(repoRoot, filepath.FromSlash(docPath))
			if _, statErr := os.Stat(fullPath); os.IsNotExist(statErr) {
				violations = append(violations, rel+": docs_sync_contract references missing path "+docPath)
			}
		}
	}

	return violations, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "docs-sync-contract: "+format+"\n", args...)
	os.Exit(1)
}
