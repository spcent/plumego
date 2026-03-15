package commands

import (
	"fmt"
	"os"

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

	// Configure formatter
	r.formatter.SetFormat(global.Format)
	r.formatter.SetQuiet(global.Quiet)
	r.formatter.SetVerbose(global.Verbose)
	r.formatter.SetColor(!global.NoColor)

	if len(args) == 0 {
		return r.showHelp()
	}

	cmdName := args[0]
	if cmdName == "help" || cmdName == "--help" || cmdName == "-h" {
		return r.showHelp()
	}

	cmd, ok := r.subcommands[cmdName]
	if !ok {
		return r.formatter.Error(fmt.Sprintf("unknown command: %s", cmdName), 1, map[string]any{
			"command": cmdName,
			"hint":    "run plumego --help",
		})
	}

	ctx := &Context{
		Out:     r.formatter,
		EnvFile: global.EnvFile,
	}

	// Run the command
	err = cmd.Run(ctx, args[1:])
	if err != nil {
		// Check if it's an AppError
		if appErr, ok := err.(*AppError); ok {
			return r.formatter.Error(appErr.Message, appErr.Code(), map[string]any{
				"detail": appErr.Detail,
			})
		}
		// Otherwise, return the error as is
		return err
	}

	return nil
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
		Format:  "text",
		EnvFile: ".env",
	}
}

func (r *RootCmd) parseGlobalFlags(args []string) (globalFlags, []string, error) {
	global := defaultGlobalFlags()
	remaining := []string{}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 >= len(args) {
				return global, nil, fmt.Errorf("%s requires a value", arg)
			}
			global.Format = args[i+1]
			i++
		case "--quiet", "-q":
			global.Quiet = true
		case "--verbose", "-v":
			global.Verbose = true
		case "--no-color":
			global.NoColor = true
		case "--env-file":
			if i+1 >= len(args) {
				return global, nil, fmt.Errorf("%s requires a value", arg)
			}
			global.EnvFile = args[i+1]
			i++
		default:
			remaining = append(remaining, arg)
		}
	}

	return global, remaining, nil
}

func (r *RootCmd) showHelp() error {
	help := `Plumego CLI - Code Agent Friendly

Usage:
  plumego [global-flags] <command> [command-flags] [args]

Global Flags:
  -f, --format <type>    Output format: json, yaml, text (default: text)
  -q, --quiet            Suppress non-essential output
  -v, --verbose          Detailed logging
      --no-color         Disable color output
      --env-file <path>  Environment file path (default: .env)

Available Commands:
  new         Create new project from template with different boilerplate options
  generate    Generate middleware, handlers, and other components
  dev         Start development server with dashboard and hot reload
  routes      Inspect registered routes and their details
  check       Validate project health and security
  config      Manage project configuration
  migrate     Run database migrations
  test        Run tests with enhanced features
  build       Build application for deployment
  inspect     Inspect running application details
  serve       Start static file server for development
  version     Show version information

Use "plumego <command> --help" for more information about a command.

Examples:
  # Create a new project
  plumego new myapp --template api

  # Generate a handler
  plumego generate handler Auth

  # Start development server
  plumego dev --addr :3000
  
  # Inspect routes
  plumego routes --format json
  
  # Check project health
  plumego check --security
  
  # Serve static files
  plumego serve
  plumego serve ./public --addr :3000

Documentation:
  https://github.com/spcent/plumego/tree/main/docs
`
	r.formatter.Print(help)
	return nil
}
