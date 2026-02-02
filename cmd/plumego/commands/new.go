package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/cmd/plumego/internal/scaffold"
)

// NewCmd creates a new project
type NewCmd struct {
	formatter *output.Formatter
}

func (c *NewCmd) Name() string  { return "new" }
func (c *NewCmd) Short() string { return "Create new project from template" }
func (c *NewCmd) Long() string {
	return `Create a new plumego project from predefined templates.

Templates:
  - minimal:      Minimal HTTP server (default)
  - api:          REST API server with routing
  - fullstack:    API + frontend serving
  - microservice: Microservice with all features

Examples:
  plumego new myapp
  plumego new myapi --template api --module github.com/acme/myapi
  plumego new webapp --template fullstack --dry-run
`
}

func (c *NewCmd) Flags() []Flag {
	return []Flag{
		{Name: "template", Default: "minimal", Usage: "Project template"},
		{Name: "module", Default: "", Usage: "Go module path"},
		{Name: "dir", Default: "", Usage: "Output directory"},
		{Name: "force", Default: false, Usage: "Overwrite existing directory"},
		{Name: "no-git", Default: false, Usage: "Skip git initialization"},
		{Name: "dry-run", Default: false, Usage: "Preview without creating"},
	}
}

func (c *NewCmd) Run(args []string) error {
	c.formatter = output.NewFormatter()
	c.formatter.SetFormat(flagFormat)
	c.formatter.SetQuiet(flagQuiet)
	c.formatter.SetVerbose(flagVerbose)

	if len(args) == 0 {
		return c.formatter.Error("project name required", 1)
	}

	projectName := args[0]
	template := "minimal"
	module := ""
	dir := filepath.Join(".", projectName)
	force := false
	noGit := false
	dryRun := false

	// Parse flags (simplified for demonstration)
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--template":
			if i+1 < len(args) {
				template = args[i+1]
				i++
			}
		case "--module":
			if i+1 < len(args) {
				module = args[i+1]
				i++
			}
		case "--dir":
			if i+1 < len(args) {
				dir = args[i+1]
				i++
			}
		case "--force":
			force = true
		case "--no-git":
			noGit = true
		case "--dry-run":
			dryRun = true
		}
	}

	// Infer module if not provided
	if module == "" {
		module = projectName
	}

	c.formatter.Verbose(fmt.Sprintf("Creating project: %s", projectName))
	c.formatter.Verbose(fmt.Sprintf("Template: %s", template))
	c.formatter.Verbose(fmt.Sprintf("Module: %s", module))
	c.formatter.Verbose(fmt.Sprintf("Directory: %s", dir))

	// Check if directory exists
	if _, err := os.Stat(dir); err == nil && !force {
		return c.formatter.Error(fmt.Sprintf("directory %s already exists (use --force to overwrite)", dir), 2)
	}

	// Validate template
	validTemplates := []string{"minimal", "api", "fullstack", "microservice"}
	valid := false
	for _, t := range validTemplates {
		if template == t {
			valid = true
			break
		}
	}
	if !valid {
		return c.formatter.Error(fmt.Sprintf("invalid template: %s (valid: minimal, api, fullstack, microservice)", template), 3)
	}

	// Dry run - just show what would be created
	if dryRun {
		files := scaffold.GetTemplateFiles(template)
		result := map[string]any{
			"dry_run":       true,
			"project":       projectName,
			"path":          dir,
			"module":        module,
			"template":      template,
			"files_created": files,
		}
		return c.formatter.Print(result)
	}

	// Create project
	files, err := scaffold.CreateProject(dir, projectName, module, template, !noGit)
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to create project: %v", err), 1)
	}

	// Success response
	result := map[string]any{
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
	}

	return c.formatter.Success("Project created successfully", result)
}
