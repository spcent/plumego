package commands

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/configmgr"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// ConfigCmd manages configuration
type ConfigCmd struct{}

func (c *ConfigCmd) Name() string  { return "config" }
func (c *ConfigCmd) Short() string { return "Configuration management" }

func (c *ConfigCmd) Run(ctx *Context, args []string) error {
	if len(args) == 0 {
		return ctx.Out.Error("config command required (show, validate, init, env)", 1)
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "show":
		return c.runShow(ctx.Out, ctx.EnvFile, subArgs)
	case "validate":
		return c.runValidate(ctx.Out, ctx.EnvFile, subArgs)
	case "init":
		return c.runInit(ctx.Out, subArgs)
	case "env":
		return c.runEnv(ctx.Out, ctx.EnvFile, subArgs)
	default:
		return ctx.Out.Error(fmt.Sprintf("unknown config command: %s", subCmd), 1)
	}
}

func (c *ConfigCmd) runShow(out *output.Formatter, envFile string, args []string) error {
	fs := flag.NewFlagSet("config show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	resolve := fs.Bool("resolve", false, "Resolve environment variables")
	redact := fs.Bool("redact", false, "Redact sensitive values")

	if err := fs.Parse(args); err != nil {
		return out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose("Loading configuration...")
	config, err := configmgr.LoadConfig(cwd, envFile, *resolve)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to load config: %v", err), 1)
	}

	if *redact {
		config = configmgr.RedactSensitive(config)
	}

	return out.Print(config)
}

func (c *ConfigCmd) runValidate(out *output.Formatter, envFile string, args []string) error {
	fs := flag.NewFlagSet("config validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose("Validating configuration...")
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
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose("Generating default configuration files...")
	files, err := configmgr.InitConfig(cwd)
	if err != nil {
		return out.Error(fmt.Sprintf("failed to init config: %v", err), 1)
	}

	return out.Success("Configuration files created", map[string]any{
		"files_created": files,
	})
}

func (c *ConfigCmd) runEnv(out *output.Formatter, envFile string, args []string) error {
	fs := flag.NewFlagSet("config env", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	if err := fs.Parse(args); err != nil {
		return out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	out.Verbose("Loading environment variables...")
	envVars := configmgr.GetEnvVars(cwd, envFile)

	return out.Print(envVars)
}
