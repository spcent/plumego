package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// RootCmd represents the base command
type RootCmd struct {
	subcommands map[string]Command
	formatter   *output.Formatter
}

// Command interface for all CLI commands
type Command interface {
	Name() string
	Short() string
	Run(ctx *Context, args []string) error
}

// BuildInfo holds build-time version metadata injected from main.
type BuildInfo struct {
	Version   string
	GitCommit string
	BuildDate string
}

// Execute runs the root command
func Execute(info BuildInfo) error {
	root := &RootCmd{
		subcommands: make(map[string]Command),
		formatter:   output.NewFormatter(),
	}

	// Register commands
	root.Register(&NewCmd{})
	root.Register(&GenerateCmd{})
	root.Register(NewDevCmd())
	root.Register(&RoutesCmd{})
	root.Register(&CheckCmd{})
	root.Register(&ConfigCmd{})
	root.Register(&MigrateCmd{})
	root.Register(&TestCmd{})
	root.Register(&BuildCmd{})
	root.Register(&InspectCmd{})
	root.Register(&ServeCmd{})
	root.Register(&VersionCmd{
		Version:   info.Version,
		GitCommit: info.GitCommit,
		BuildDate: info.BuildDate,
	})

	return root.Run(os.Args[1:])
}

// Register adds a command to the root
func (r *RootCmd) Register(cmd Command) {
	r.subcommands[cmd.Name()] = cmd
}

// Run executes the command
func (r *RootCmd) Run(args []string) error {
	// Parse global flags
	global, args, err := r.parseGlobalFlags(args)
	if err != nil {
		return r.formatter.Error(fmt.Sprintf("invalid global flags: %v", err), 1)
	}
	if !output.IsSupportedFormat(global.Format) {
		r.formatter.SetFormat(defaultGlobalFlags().Format)
		return r.formatter.Error(fmt.Sprintf("unsupported output format: %s", global.Format), 1, map[string]any{
			"format":            global.Format,
			"supported_formats": []string{"json", "yaml", "text"},
		})
	}

	// Configure formatter
	r.formatter.SetFormat(global.Format)
	r.formatter.SetQuiet(global.Quiet)
	r.formatter.SetVerbose(global.Verbose)
	r.formatter.SetColor(!global.NoColor)

	if len(args) == 0 {
		return r.showHelp()
	}

	cmdName := args[0]
	if cmdName == "--help" || cmdName == "-h" {
		return r.showHelp()
	}
	if cmdName == "help" {
		if len(args) > 1 {
			return r.showCommandHelp(args[1])
		}
		return r.showHelp()
	}

	cmd, ok := r.subcommands[cmdName]
	if !ok {
		return r.formatter.Error(fmt.Sprintf("unknown command: %s", cmdName), 1, map[string]any{
			"command": cmdName,
			"hint":    "run plumego --help",
		})
	}
	if len(args) > 1 && (args[1] == "--help" || args[1] == "-h") {
		return r.showCommandHelp(cmdName)
	}

	ctx := &Context{
		Out:     r.formatter,
		EnvFile: global.EnvFile,
	}

	return cmd.Run(ctx, args[1:])
}

type globalFlags struct {
	Format  string
	Quiet   bool
	Verbose bool
	NoColor bool
	EnvFile string
}

func defaultGlobalFlags() globalFlags {
	return globalFlags{
		Format:  "json",
		EnvFile: ".env",
	}
}

func (r *RootCmd) parseGlobalFlags(args []string) (globalFlags, []string, error) {
	global := defaultGlobalFlags()

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--help" || arg == "-h" {
			return global, args[i:], nil
		}
		if arg == "--" {
			return global, args[i+1:], nil
		}
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			return global, args[i:], nil
		}

		name, inlineValue, hasInlineValue := strings.Cut(arg, "=")
		switch arg {
		case "--format", "-f":
			if i+1 >= len(args) {
				return global, nil, fmt.Errorf("%s requires a value", arg)
			}
			global.Format = args[i+1]
			i++
		case "--env-file":
			if i+1 >= len(args) {
				return global, nil, fmt.Errorf("%s requires a value", arg)
			}
			global.EnvFile = args[i+1]
			i++
		case "--quiet", "-q":
			global.Quiet = true
		case "--verbose", "-v":
			global.Verbose = true
		case "--no-color":
			global.NoColor = true
		default:
			switch name {
			case "--format", "-f":
				if !hasInlineValue || inlineValue == "" {
					return global, nil, fmt.Errorf("%s requires a value", name)
				}
				global.Format = inlineValue
			case "--env-file":
				if !hasInlineValue || inlineValue == "" {
					return global, nil, fmt.Errorf("%s requires a value", name)
				}
				global.EnvFile = inlineValue
			default:
				return global, nil, fmt.Errorf("unknown global flag: %s", arg)
			}
		}
	}

	return global, nil, nil
}

func (r *RootCmd) showCommandHelp(name string) error {
	cmd, ok := r.subcommands[name]
	if !ok {
		return r.formatter.Error(fmt.Sprintf("unknown command: %s", name), 1, map[string]any{
			"command": name,
			"hint":    "run plumego --help",
		})
	}

	return r.formatter.Print(commandHelp(cmd))
}

func commandHelp(cmd Command) string {
	header := fmt.Sprintf(`Usage:
  plumego [global-flags] %s [command-flags] [args]

Summary:
  %s
`, cmd.Name(), cmd.Short())

	var body string
	switch cmd.Name() {
	case "new":
		body = `
Command Flags:
      --template <name>  Project template
      --module <path>    Go module path
      --dir <path>       Output directory
      --force            Overwrite existing directory
      --no-git           Skip git initialization
      --dry-run          Preview without creating files

Examples:
  plumego new myapp --template canonical
  plumego new myapi --template api --module github.com/acme/myapi
`
	case "generate":
		body = `
Usage:
  plumego [global-flags] generate <middleware|handler|model> <name> [command-flags]

Command Flags:
      --dir <path>       Project directory
      --output <path>    Output file path
      --package <name>   Package name
      --methods <list>   Handler HTTP methods: GET,POST,PUT,DELETE
      --with-tests       Generate tests where supported
      --with-validation  Generate validation where supported
      --force            Overwrite existing files
`
	case "dev":
		body = `
Command Flags:
      --dir <path>                 Project directory
      --addr <addr>                Application listen address
      --dashboard-addr <addr>      Dashboard listen address
      --dashboard-token <token>    Token required for dashboard action APIs
      --watch <patterns>           Comma-separated watch patterns
      --exclude <patterns>         Comma-separated exclude patterns
      --debounce <duration>        Reload debounce duration
      --no-reload                  Disable hot reload
      --build-cmd <command>        Custom build command
      --run-cmd <command>          Custom run command
`
	case "routes":
		body = `
Command Flags:
      --dir <path>       Project directory
      --method <method>  Filter by HTTP method
      --pattern <text>   Filter routes by path substring
      --middleware       Include middleware summary
      --sort <field>     Sort by path, method, or group
`
	case "check":
		body = `
Command Flags:
      --config-only  Only check configuration
      --deps-only    Only check dependencies
      --security     Run security checks
`
	case "config":
		body = `
Subcommands:
  show      Show resolved configuration
  validate  Validate configuration
  init      Create default configuration files
  env       Show environment variables

Command Flags:
      --resolve       Resolve env-file values in config show
      --show-secrets  Show raw sensitive values in config show
`
	case "migrate":
		body = `
Subcommands:
  create <name>  Create offline up/down migration files
  status         Show migration status using a registered database driver
  up             Apply pending migrations
  down           Roll back applied migrations

Command Flags:
      --dir <path>    Migrations directory
      --driver <name> Registered database/sql driver name
      --db-url <url>  Database connection string
      --steps <n>     Number of migrations to apply or roll back
`
	case "test":
		body = `
Command Flags:
      --dir <path>           Project directory
      --race                 Enable race detector
      --cover                Enable coverage
      --coverprofile <path>  Coverage profile path
      --bench                Run benchmarks
      --timeout <duration>   Test timeout
      --tags <list>          Build tags
      --run <pattern>        Run pattern
      --short                Run short tests
`
	case "build":
		body = `
Command Flags:
      --dir <path>       Project directory
      --output <path>    Output binary path
      --ldflags <flags>  Go linker flags
      --tags <list>      Build tags
      --race             Enable race detector
      --trimpath         Remove filesystem paths
`
	case "inspect":
		body = `
Subcommands:
  health   Probe health endpoints
  metrics  Fetch metrics endpoint
  routes   Fetch debug route data
  config   Fetch debug config data
  info     Fetch debug app info

Command Flags:
      --url <url>           Application URL
      --auth <token>        Authorization header value
      --timeout <duration>  Request timeout
`
	case "serve":
		body = `
Usage:
  plumego [global-flags] serve [command-flags] [directory]

Command Flags:
  -a, --addr <addr>  Server address
`
	case "version":
		body = `
Command Flags:
  None.
`
	default:
		body = ""
	}

	return header + body + globalHelpFooter()
}

func globalHelpFooter() string {
	return `
Global Flags:
  -f, --format <type>    Output format: json, yaml, text (default: json)
  -q, --quiet            Suppress non-essential output
  -v, --verbose          Detailed logging
      --no-color         Disable color output
      --env-file <path>  Environment file path (default: .env)
`
}

func (r *RootCmd) showHelp() error {
	help := `Plumego CLI - Code Agent Friendly

Usage:
  plumego [global-flags] <command> [command-flags] [args]

Global Flags:
  -f, --format <type>    Output format: json, yaml, text (default: json)
  -q, --quiet            Suppress non-essential output
  -v, --verbose          Detailed logging
      --no-color         Disable color output
      --env-file <path>  Environment file path (default: .env)

Available Commands:
  new         Create new project from template
  generate    Generate middleware, handlers
  dev         Start development server with hot reload
  routes      Inspect registered routes
  check       Validate project health
  config      Configuration management
  migrate     Database migrations
  test        Enhanced test running
  build       Build application
  inspect     Inspect running application
  serve       Start static file server
  version     Show version information

Use "plumego <command> --help" for more information about a command.

Examples:
  plumego new myapp --template api
  plumego generate handler Auth
  plumego dev --addr :3000
  plumego --format json routes
  plumego check --security
  plumego serve
  plumego serve ./public --addr :3000

Documentation:
  https://github.com/spcent/plumego/tree/main/docs
`
	return r.formatter.Print(help)
}
