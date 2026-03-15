package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/scaffold"
)

// NewCmd creates a new project
type NewCmd struct{}

func (c *NewCmd) Name() string  { return "new" }
func (c *NewCmd) Short() string { return "Create new project from template" }

func (c *NewCmd) Run(ctx *Context, args []string) error {
	// Check for help flag
	if containsHelpFlag(args) {
		help := `Usage: plumego new [options] <project-name>

Create a new Plumego project from a template.

Options:
  --template <name>   Project template (canonical, minimal, api, fullstack, microservice) (default: canonical)
  --module <path>     Go module path (default: project name)
  --dir <path>        Output directory (default: ./<project-name>)
  --force             Overwrite existing directory
  --no-git            Skip git initialization
  --dry-run           Preview without creating files

Templates:
  canonical     Standard Plumego project structure with all features
  minimal       Minimal project structure with basic setup
  api           API-focused project with routes and middleware
  fullstack     Fullstack project with frontend integration
  microservice  Microservice-oriented project structure

Examples:
  plumego new myapp                      # Create project with canonical template
  plumego new myapp --template api       # Create API-focused project
  plumego new myapp --module github.com/myuser/myapp  # Set custom module path
  plumego new myapp --dir ./projects/myapp  # Set custom output directory
  plumego new myapp --dry-run            # Preview without creating files
`
		ctx.Out.Print(help)
		return nil
	}

	var templateFlag *string
	var moduleFlag *string
	var dirFlag *string
	var force *bool
	var noGit *bool
	var dryRun *bool

	positionals, err := ParseFlags("new", args, func(fs *flag.FlagSet) {
		templateFlag = fs.String("template", "canonical", "Project template (canonical, minimal, api, fullstack, microservice)")
		moduleFlag = fs.String("module", "", "Go module path")
		dirFlag = fs.String("dir", "", "Output directory")
		force = fs.Bool("force", false, "Overwrite existing directory")
		noGit = fs.Bool("no-git", false, "Skip git initialization")
		dryRun = fs.Bool("dry-run", false, "Preview without creating")
	})

	if err != nil {
		return NewAppError(1, fmt.Sprintf("invalid flags: %v", err), "")
	}

	if len(positionals) == 0 {
		return NewAppError(1, "project name required", "")
	}

	projectName := positionals[0]
	template := *templateFlag
	module := *moduleFlag
	dir := *dirFlag

	if dir == "" {
		dir = filepath.Join(".", projectName)
	}
	if module == "" {
		module = projectName
	}

	ctx.Out.Verbose(fmt.Sprintf("Creating project: %s", projectName))
	ctx.Out.Verbose(fmt.Sprintf("Template: %s", template))
	ctx.Out.Verbose(fmt.Sprintf("Module: %s", module))
	ctx.Out.Verbose(fmt.Sprintf("Directory: %s", dir))

	if _, err := os.Stat(dir); err == nil && !*force {
		return NewAppError(2, fmt.Sprintf("directory %s already exists (use --force to overwrite)", dir), "")
	}

	validTemplates := []string{"canonical", "minimal", "api", "fullstack", "microservice"}
	valid := false
	for _, t := range validTemplates {
		if template == t {
			valid = true
			break
		}
	}
	if !valid {
		return NewAppError(3, fmt.Sprintf("invalid template: %s (valid: canonical, minimal, api, fullstack, microservice)", template), "")
	}

	if *dryRun {
		files := scaffold.GetTemplateFiles(template)
		return ctx.Out.Print(map[string]any{
			"dry_run":       true,
			"project":       projectName,
			"path":          dir,
			"module":        module,
			"template":      template,
			"files_created": files,
		})
	}

	files, err := scaffold.CreateProject(dir, projectName, module, template, !*noGit)
	if err != nil {
		return NewAppError(1, fmt.Sprintf("failed to create project: %v", err), "")
	}

	return ctx.Out.Success("Project created successfully", map[string]any{
		"project":       projectName,
		"path":          dir,
		"module":        module,
		"template":      template,
		"files_created": files,
		"next_steps": []string{
			fmt.Sprintf("cd %s", projectName),
			"go mod tidy",
			"plumego dev",
		},
	})
}
