package commands

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/checker"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// CheckCmd validates project health
type CheckCmd struct {
}

func (c *CheckCmd) Name() string  { return "check" }
func (c *CheckCmd) Short() string { return "Validate project health" }
func (c *CheckCmd) Long() string {
	return `Validate project structure, configuration, and dependencies.

Checks:
  - Configuration validation (env vars, config files)
  - Dependency status (go.mod, outdated packages)
  - Security audit (secrets, known vulnerabilities)
  - Project structure (required files, conventions)

Examples:
  plumego check
  plumego check --config-only
  plumego check --security --format json
`
}

func (c *CheckCmd) Flags() []Flag {
	return []Flag{
		{Name: "config-only", Default: false, Usage: "Only check configuration"},
		{Name: "deps-only", Default: false, Usage: "Only check dependencies"},
		{Name: "security", Default: false, Usage: "Run security checks"},
	}
}

func (c *CheckCmd) Run(ctx *Context, args []string) error {
	out := ctx.Out

	// Parse flags
	configOnly := false
	depsOnly := false
	security := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--config-only":
			configOnly = true
		case "--deps-only":
			depsOnly = true
		case "--security":
			security = true
		}
	}

	// Determine current directory
	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose(fmt.Sprintf("Checking project at: %s", cwd))

	// Run checks
	checks := &checker.CheckResult{
		Status: "healthy",
		Checks: make(map[string]checker.CheckDetail),
	}

	// Configuration check
	if !depsOnly {
		out.Verbose("Running configuration checks...")
		configCheck := checker.CheckConfig(cwd, ctx.EnvFile)
		checks.Checks["config"] = configCheck
		if configCheck.Status == "failed" {
			checks.Status = "unhealthy"
		}
	}

	// Dependencies check
	if !configOnly {
		out.Verbose("Running dependency checks...")
		depsCheck := checker.CheckDependencies(cwd)
		checks.Checks["dependencies"] = depsCheck
		if depsCheck.Status == "failed" {
			checks.Status = "unhealthy"
		} else if depsCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	// Security check
	if security && !configOnly && !depsOnly {
		out.Verbose("Running security checks...")
		securityCheck := checker.CheckSecurity(cwd, ctx.EnvFile)
		checks.Checks["security"] = securityCheck
		if securityCheck.Status == "failed" {
			checks.Status = "unhealthy"
		} else if securityCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	// Project structure check
	if !configOnly && !depsOnly {
		out.Verbose("Running project structure checks...")
		structureCheck := checker.CheckStructure(cwd)
		checks.Checks["structure"] = structureCheck
		if structureCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	// Determine exit code
	exitCode := 0
	switch checks.Status {
	case "unhealthy":
		exitCode = 1
	case "degraded":
		exitCode = 2
	}

	if exitCode == 0 {
		return out.Success("All checks passed", checks)
	}

	// Print result and exit with appropriate code
	if err := out.Print(checks); err != nil {
		return err
	}
	return output.Exit(exitCode)
}
