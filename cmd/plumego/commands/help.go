package commands

import (
	"fmt"
	"strings"
)

var stableCommandOrder = []string{
	"new",
	"generate",
	"dev",
	"routes",
	"check",
	"config",
	"migrate",
	"test",
	"build",
	"inspect",
	"serve",
	"agents",
	"doctor",
	"version",
}

type commandHelpProvider interface {
	Help() CommandHelp
}

type CommandHelp struct {
	Args        string
	Subcommands []HelpItem
	Flags       []HelpItem
	Examples    []string
}

type HelpItem struct {
	Name        string
	Description string
}

func commandHelp(cmd Command) string {
	spec := CommandHelp{Args: "[command-flags] [args]"}
	if provider, ok := cmd.(commandHelpProvider); ok {
		spec = provider.Help()
	}
	args := strings.TrimSpace(spec.Args)
	if args == "" {
		args = "[command-flags] [args]"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Usage:\n  plumego [global-flags] %s %s\n\n", cmd.Name(), args)
	fmt.Fprintf(&b, "Summary:\n  %s\n", cmd.Short())
	writeHelpItems(&b, "Subcommands", spec.Subcommands)
	writeHelpItems(&b, "Command Flags", spec.Flags)
	writeExamples(&b, spec.Examples)
	b.WriteString(globalHelpFooter())
	return b.String()
}

func writeHelpItems(b *strings.Builder, title string, items []HelpItem) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "\n%s:\n", title)
	width := 0
	for _, item := range items {
		if len(item.Name) > width {
			width = len(item.Name)
		}
	}
	for _, item := range items {
		fmt.Fprintf(b, "  %-*s  %s\n", width, item.Name, item.Description)
	}
}

func writeExamples(b *strings.Builder, examples []string) {
	if len(examples) == 0 {
		return
	}
	b.WriteString("\nExamples:\n")
	for _, example := range examples {
		fmt.Fprintf(b, "  %s\n", example)
	}
}

func (c *NewCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags] <project-name>",
		Flags: []HelpItem{
			{"--template <name>", "Project template"},
			{"--module <path>", "Go module path"},
			{"--dir <path>", "Output directory"},
			{"--force", "Overwrite existing directory"},
			{"--no-git", "Skip git initialization"},
			{"--dry-run", "Preview without creating files"},
		},
		Examples: []string{
			"plumego new myapp --template canonical",
			"plumego new myapi --template api --module github.com/acme/myapi",
		},
	}
}

func (c *GenerateCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags] <type> <name>",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--output <path>", "Output file path"},
			{"--package <name>", "Package name"},
			{"--methods <list>", "Handler HTTP methods: GET,POST,PUT,DELETE"},
			{"--with-tests", "Generate tests where supported"},
			{"--with-validation", "Generate validation where supported"},
			{"--force", "Overwrite existing files"},
		},
		Examples: []string{
			"plumego generate middleware RateLimit",
			"plumego generate handler User --methods GET,POST,PUT,DELETE --with-tests",
			"plumego generate model Invoice --with-validation",
		},
	}
}

func (c *DevCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags]",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--addr <addr>", "Application listen address"},
			{"--dashboard-addr <addr>", "Dashboard listen address"},
			{"--dashboard-token <token>", "Token required for dashboard action APIs"},
			{"--watch <patterns>", "Comma-separated watch patterns"},
			{"--exclude <patterns>", "Comma-separated exclude patterns"},
			{"--debounce <duration>", "Reload debounce duration"},
			{"--no-reload", "Disable hot reload"},
			{"--build-cmd <command>", "Custom build command"},
			{"--run-cmd <command>", "Custom run command"},
		},
	}
}

func (c *RoutesCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags]",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--method <method>", "Filter by HTTP method"},
			{"--pattern <text>", "Filter routes by path substring"},
			{"--middleware", "Unsupported: middleware extraction is not available"},
			{"--group <name>", "Unsupported: route group filtering is not available"},
			{"--sort <field>", "Sort by path or method"},
		},
	}
}

func (c *CheckCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags]",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--config-only", "Only check configuration"},
			{"--deps-only", "Only check dependencies"},
			{"--security", "Run security checks"},
			{"--updates", "Check for available dependency updates"},
		},
	}
}

func (c *ConfigCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "<show|validate|init|env> [command-flags]",
		Subcommands: []HelpItem{
			{"show [--dir <path>] [--resolve] [--show-secrets]", "Show configuration"},
			{"validate [--dir <path>]", "Validate configuration"},
			{"init [--dir <path>]", "Create default configuration files"},
			{"env [--dir <path>]", "Show environment variables"},
		},
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
		},
	}
}

func (c *MigrateCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "<create|status|up|down> [command-flags] [args]",
		Subcommands: []HelpItem{
			{"create <name>", "Create offline up/down migration files"},
			{"status", "Show migration status using a registered database driver"},
			{"up", "Apply pending migrations"},
			{"down", "Roll back applied migrations"},
		},
		Flags: []HelpItem{
			{"--dir <path>", "Migrations directory"},
			{"--driver <name>", "Registered database/sql driver name"},
			{"--db-url <url>", "Database connection string"},
			{"--steps <n>", "Number of migrations to apply or roll back"},
		},
	}
}

func (c *TestCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags] [packages...]",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--race", "Enable race detector"},
			{"--cover", "Enable coverage"},
			{"--coverprofile <path>", "Coverage profile path"},
			{"--bench", "Run benchmarks"},
			{"--timeout <duration>", "Test timeout"},
			{"--tags <list>", "Build tags"},
			{"--run <pattern>", "Run pattern"},
			{"--short", "Run short tests"},
		},
	}
}

func (c *BuildCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags] [target]",
		Flags: []HelpItem{
			{"--dir <path>", "Project directory"},
			{"--output <path>", "Output binary path"},
			{"--ldflags <flags>", "Go linker flags"},
			{"--tags <list>", "Build tags"},
			{"--race", "Enable race detector"},
			{"--trimpath", "Remove filesystem paths"},
		},
	}
}

func (c *InspectCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[health|metrics|routes|config|info] [command-flags]",
		Subcommands: []HelpItem{
			{"health", "Probe health endpoints"},
			{"metrics", "Fetch metrics endpoint"},
			{"routes", "Fetch debug route data"},
			{"config", "Fetch debug config data"},
			{"info", "Fetch debug app info"},
		},
		Flags: []HelpItem{
			{"--url <url>", "Application URL"},
			{"--auth <value>", "Authorization header value"},
			{"--timeout <duration>", "Request timeout"},
		},
	}
}

func (c *ServeCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags] [directory]",
		Flags: []HelpItem{
			{"-a, --addr <addr>", "Server address"},
		},
		Examples: []string{
			"plumego serve",
			"plumego serve ./public",
			"plumego serve ./public --addr :3000",
		},
	}
}

func (c *VersionCmd) Help() CommandHelp {
	return CommandHelp{
		Args: "[command-flags]",
		Flags: []HelpItem{
			{"None.", ""},
		},
	}
}
