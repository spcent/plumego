package commands

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/checker"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// CheckCmd validates project health
type CheckCmd struct{}

func (c *CheckCmd) Name() string  { return "check" }
func (c *CheckCmd) Short() string { return "Validate project health" }

func (c *CheckCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	configOnly := fs.Bool("config-only", false, "Only check configuration")
	depsOnly := fs.Bool("deps-only", false, "Only check dependencies")
	security := fs.Bool("security", false, "Run security checks")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	ctx.Out.Verbose(fmt.Sprintf("Checking project at: %s", cwd))

	checks := &checker.CheckResult{
		Status: "healthy",
		Checks: make(map[string]checker.CheckDetail),
	}

	if !*depsOnly {
		ctx.Out.Verbose("Running configuration checks...")
		configCheck := checker.CheckConfig(cwd, ctx.EnvFile)
		checks.Checks["config"] = configCheck
		if configCheck.Status == "failed" {
			checks.Status = "unhealthy"
		}
	}

	if !*configOnly {
		ctx.Out.Verbose("Running dependency checks...")
		depsCheck := checker.CheckDependencies(cwd)
		checks.Checks["dependencies"] = depsCheck
		if depsCheck.Status == "failed" {
			checks.Status = "unhealthy"
		} else if depsCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	if *security && !*configOnly && !*depsOnly {
		ctx.Out.Verbose("Running security checks...")
		securityCheck := checker.CheckSecurity(cwd, ctx.EnvFile)
		checks.Checks["security"] = securityCheck
		if securityCheck.Status == "failed" {
			checks.Status = "unhealthy"
		} else if securityCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	if !*configOnly && !*depsOnly {
		ctx.Out.Verbose("Running project structure checks...")
		structureCheck := checker.CheckStructure(cwd)
		checks.Checks["structure"] = structureCheck
		if structureCheck.Status == "warning" && checks.Status == "healthy" {
			checks.Status = "degraded"
		}
	}

	exitCode := 0
	switch checks.Status {
	case "unhealthy":
		exitCode = 1
	case "degraded":
		exitCode = 2
	}

	if exitCode == 0 {
		return ctx.Out.Success("All checks passed", checks)
	}

	if err := ctx.Out.Print(checks); err != nil {
		return err
	}
	return output.Exit(exitCode)
}
