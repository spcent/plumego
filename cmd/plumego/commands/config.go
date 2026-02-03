package commands

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/configmgr"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// ConfigCmd manages configuration
type ConfigCmd struct {
}

func (c *ConfigCmd) Name() string  { return "config" }
func (c *ConfigCmd) Short() string { return "Configuration management" }
func (c *ConfigCmd) Long() string {
	return `View, validate, and generate configuration files.

Commands:
  show      Display current configuration
  validate  Validate configuration files
  init      Generate default config files
  env       Show environment variables

Examples:
  plumego config show
  plumego config show --resolve --redact
  plumego config validate
  plumego config init
  plumego config env --format json
`
}

func (c *ConfigCmd) Flags() []Flag {
	return []Flag{
		{Name: "resolve", Default: false, Usage: "Resolve environment variables"},
		{Name: "redact", Default: false, Usage: "Redact sensitive values"},
	}
}

func (c *ConfigCmd) Run(ctx *Context, args []string) error {
	out := ctx.Out

	if len(args) == 0 {
		return out.Error("config command required (show, validate, init, env)", 1)
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "show":
		return c.runShow(out, ctx.EnvFile, subArgs)
	case "validate":
		return c.runValidate(out, ctx.EnvFile, subArgs)
	case "init":
		return c.runInit(out, subArgs)
	case "env":
		return c.runEnv(out, ctx.EnvFile, subArgs)
	default:
		return out.Error(fmt.Sprintf("unknown config command: %s", subCmd), 1)
	}
}

func (c *ConfigCmd) runShow(out *output.Formatter, envFile string, args []string) error {
	resolve := false
	redact := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--resolve":
			resolve = true
		case "--redact":
			redact = true
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose("Loading configuration...")
	config, err := configmgr.LoadConfig(cwd, envFile, resolve)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to load config: %v", err), 1)
	}

	if redact {
		config = configmgr.RedactSensitive(config)
	}

	return out.Print(config)
}

func (c *ConfigCmd) runValidate(out *output.Formatter, envFile string, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose(fmt.Sprintf("Validating configuration with args: %v", args))
	result := configmgr.ValidateConfig(cwd, envFile)

	if len(result.Errors) > 0 {
		if err := out.Print(result); err != nil {
			return err
		}
		return output.Exit(1)
	}

	if len(result.Warnings) > 0 {
		if err := out.Print(result); err != nil {
			return err
		}
		return output.Exit(2)
	}

	return out.Success("Configuration is valid", result)
}

func (c *ConfigCmd) runInit(out *output.Formatter, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose(fmt.Sprintf("Generating default configuration files with args: %v", args))
	files, err := configmgr.InitConfig(cwd)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to init config: %v", err), 1)
	}

	result := map[string]any{
		"files_created": files,
	}

	return out.Success("Configuration files created", result)
}

func (c *ConfigCmd) runEnv(out *output.Formatter, envFile string, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose(fmt.Sprintf("Loading environment variables with args: %v", args))
	envVars := configmgr.GetEnvVars(cwd, envFile)

	return out.Print(envVars)
}
