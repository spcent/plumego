// Package agentmanifest reads and exposes module.yaml manifests.
package agentmanifest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Manifest is the machine-readable module boundary declaration.
type Manifest struct {
	Name    string `yaml:"name"`
	Path    string `yaml:"path"`
	Layer   string `yaml:"layer"`
	Status  string `yaml:"status"`
	Owner   string `yaml:"owner"`
	Risk    string `yaml:"risk"`
	Summary string `yaml:"summary"`

	StrictBoundary string `yaml:"strict_boundary"`

	EntrypointReads []string `yaml:"entrypoint_reads"`

	Responsibilities []string `yaml:"responsibilities"`
	NonGoals         []string `yaml:"non_goals"`

	PublicEntrypoints []string `yaml:"public_entrypoints"`

	AllowedImports   []string `yaml:"allowed_imports"`
	ForbiddenImports []string `yaml:"forbidden_imports"`

	TestCommands []string `yaml:"test_commands"`

	DocPaths []string `yaml:"doc_paths"`

	ChangeRisks     []string `yaml:"change_risks"`
	StopConditions  []string `yaml:"stop_conditions"`
	ReviewChecklist []string `yaml:"review_checklist"`
	AgentHints      []string `yaml:"agent_hints"`
}

// Load reads the module.yaml at <repoRoot>/<modulePath>/module.yaml.
func Load(repoRoot, modulePath string) (*Manifest, error) {
	manifestPath := filepath.Join(repoRoot, modulePath, "module.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no module.yaml found at %s", manifestPath)
		}
		return nil, fmt.Errorf("read module.yaml: %w", err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse module.yaml: %w", err)
	}
	return &m, nil
}

// FindRepoRoot walks up from dir until it finds a directory containing AGENTS.md.
// Returns an error if not found within 8 levels.
func FindRepoRoot(dir string) (string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("resolve dir: %w", err)
	}

	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(abs, "AGENTS.md")); err == nil {
			return abs, nil
		}
		parent := filepath.Dir(abs)
		if parent == abs {
			break
		}
		abs = parent
	}
	return "", fmt.Errorf("could not find repo root (no AGENTS.md within 8 parent directories of %s)", dir)
}

// KnownStableRoots returns the canonical list of stable-root module paths.
func KnownStableRoots() []string {
	return []string{
		"core", "router", "contract", "middleware",
		"security", "store", "health", "log", "metrics",
	}
}

// KnownReferenceApps returns the canonical list of reference application paths.
func KnownReferenceApps() []string {
	return []string{
		"reference/standard-service",
		"reference/with-websocket",
		"reference/with-gateway",
		"reference/with-rest",
		"reference/with-ops",
	}
}

// KnownSpecFiles returns the paths of required spec files relative to repo root.
func KnownSpecFiles() []string {
	return []string{
		"AGENTS.md",
		"specs/task-routing.yaml",
		"specs/dependency-rules.yaml",
		"specs/repo.yaml",
	}
}

// SplitCommand splits a shell command string into name and args.
// It handles simple whitespace separation; it does not interpret shell quoting.
func SplitCommand(cmd string) (string, []string) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}
