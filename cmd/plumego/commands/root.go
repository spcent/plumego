package commands

import (
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// Global flags
var (
	flagFormat   string
	flagQuiet    bool
	flagVerbose  bool
	flagNoColor  bool
	flagConfig   string
	flagEnvFile  string
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
	Long() string
	Flags() []Flag
	Run(args []string) error
}

// Flag represents a command flag
type Flag struct {
	Name      string
	Short     string
	Default   interface{}
	Usage     string
	Required  bool
}

// Execute runs the root command
func Execute() error {
	root := &RootCmd{
		subcommands: make(map[string]Command),
		formatter:   output.NewFormatter(),
	}

	// Register commands
	root.Register(&NewCmd{})
	root.Register(&GenerateCmd{})
	root.Register(&DevCmd{})
	root.Register(&RoutesCmd{})
	root.Register(&CheckCmd{})
	root.Register(&ConfigCmd{})
	root.Register(&TestCmd{})
	root.Register(&BuildCmd{})
	root.Register(&InspectCmd{})
	root.Register(&VersionCmd{})

	return root.Run(os.Args[1:])
}

// Register adds a command to the root
func (r *RootCmd) Register(cmd Command) {
	r.subcommands[cmd.Name()] = cmd
}

// Run executes the command
func (r *RootCmd) Run(args []string) error {
	// Parse global flags
	args = r.parseGlobalFlags(args)

	// Configure formatter
	r.formatter.SetFormat(flagFormat)
	r.formatter.SetQuiet(flagQuiet)
	r.formatter.SetVerbose(flagVerbose)
	r.formatter.SetColor(!flagNoColor)

	if len(args) == 0 {
		return r.showHelp()
	}

	cmdName := args[0]
	cmd, ok := r.subcommands[cmdName]
	if !ok {
		return r.showHelp()
	}

	return cmd.Run(args[1:])
}

func (r *RootCmd) parseGlobalFlags(args []string) []string {
	// Simple flag parsing for demonstration
	// In production, use pflag or similar
	remaining := []string{}
	flagFormat = "json"
	flagConfig = ".plumego.yaml"
	flagEnvFile = ".env"

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				flagFormat = args[i+1]
				i++
			}
		case "--quiet", "-q":
			flagQuiet = true
		case "--verbose", "-v":
			flagVerbose = true
		case "--no-color":
			flagNoColor = true
		case "--config", "-c":
			if i+1 < len(args) {
				flagConfig = args[i+1]
				i++
			}
		case "--env-file":
			if i+1 < len(args) {
				flagEnvFile = args[i+1]
				i++
			}
		default:
			remaining = append(remaining, arg)
		}
	}

	return remaining
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
  -c, --config <path>    Config file path (default: .plumego.yaml)
      --env-file <path>  Environment file path (default: .env)

Available Commands:
  new         Create new project from template
  generate    Generate components, middleware, handlers
  dev         Start development server with hot reload
  routes      Inspect registered routes
  check       Validate project health
  config      Configuration management
  test        Enhanced test running
  build       Build application
  inspect     Inspect running application
  version     Show version information

Use "plumego <command> --help" for more information about a command.

Examples:
  plumego new myapp --template api
  plumego generate component Auth
  plumego dev --addr :3000
  plumego routes --format json
  plumego check --security

Documentation:
  https://github.com/spcent/plumego/tree/main/docs
`
	r.formatter.Print(help)
	return nil
}
