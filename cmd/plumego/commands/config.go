package commands

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/configmgr"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// ConfigCmd manages configuration
type ConfigCmd struct {
	formatter *output.Formatter
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

func (c *ConfigCmd) Run(args []string) error {
	c.formatter = output.NewFormatter()
	c.formatter.SetFormat(flagFormat)
	c.formatter.SetQuiet(flagQuiet)
	c.formatter.SetVerbose(flagVerbose)

	if len(args) == 0 {
		return c.formatter.Error("config command required (show, validate, init, env)", 1)
	}

	subCmd := args[0]
	subArgs := args[1:]

	switch subCmd {
	case "show":
		return c.runShow(subArgs)
	case "validate":
		return c.runValidate(subArgs)
	case "init":
		return c.runInit(subArgs)
	case "env":
		return c.runEnv(subArgs)
	default:
		return c.formatter.Error(fmt.Sprintf("unknown config command: %s", subCmd), 1)
	}
}

func (c *ConfigCmd) runShow(args []string) error {
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
		return c.formatter.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	c.formatter.Verbose("Loading configuration...")
	config, err := configmgr.LoadConfig(cwd, flagEnvFile, resolve)
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to load config: %v", err), 1)
	}

	if redact {
		config = configmgr.RedactSensitive(config)
	}

	return c.formatter.Print(config)
}

func (c *ConfigCmd) runValidate(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	c.formatter.Verbose("Validating configuration...")
	result := configmgr.ValidateConfig(cwd, flagEnvFile)

	if len(result.Errors) > 0 {
		if err := c.formatter.Print(result); err != nil {
			return err
		}
		os.Exit(1)
		return nil
	}

	if len(result.Warnings) > 0 {
		if err := c.formatter.Print(result); err != nil {
			return err
		}
		os.Exit(2)
		return nil
	}

	return c.formatter.Success("Configuration is valid", result)
}

func (c *ConfigCmd) runInit(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	c.formatter.Verbose("Generating default configuration files...")
	files, err := configmgr.InitConfig(cwd)
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to init config: %v", err), 1)
	}

	result := map[string]interface{}{
		"files_created": files,
	}

	return c.formatter.Success("Configuration files created", result)
}

func (c *ConfigCmd) runEnv(args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	c.formatter.Verbose("Loading environment variables...")
	envVars := configmgr.GetEnvVars(cwd, flagEnvFile)

	return c.formatter.Print(envVars)
}
