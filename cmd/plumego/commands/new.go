package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/scaffold"
)

// NewCmd creates a new project
type NewCmd struct{}

func (c *NewCmd) Name() string  { return "new" }
func (c *NewCmd) Short() string { return "Create new project from template" }

func (c *NewCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	templateFlag := fs.String("template", "canonical", "Project template (canonical, minimal, api, fullstack, microservice)")
	moduleFlag := fs.String("module", "", "Go module path")
	dirFlag := fs.String("dir", "", "Output directory")
	force := fs.Bool("force", false, "Overwrite existing directory")
	noGit := fs.Bool("no-git", false, "Skip git initialization")
	dryRun := fs.Bool("dry-run", false, "Preview without creating")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	positionals := fs.Args()
	if len(positionals) == 0 {
		return ctx.Out.Error("project name required", 1)
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
		return ctx.Out.Error(fmt.Sprintf("directory %s already exists (use --force to overwrite)", dir), 2)
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
		return ctx.Out.Error(fmt.Sprintf("invalid template: %s (valid: canonical, minimal, api, fullstack, microservice)", template), 3)
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
		return ctx.Out.Error(fmt.Sprintf("failed to create project: %v", err), 1)
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
