package commands

import (
	"fmt"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/codegen"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
)

// GenerateCmd generates code
type GenerateCmd struct {
	formatter *output.Formatter
}

func (c *GenerateCmd) Name() string  { return "generate" }
func (c *GenerateCmd) Short() string { return "Generate components, middleware, handlers" }
func (c *GenerateCmd) Long() string {
	return `Generate boilerplate code for plumego components.

Types:
  component   Component with full lifecycle
  middleware  HTTP middleware
  handler     HTTP handler
  model       Data model with validation

Examples:
  plumego generate component Auth
  plumego generate middleware RateLimit
  plumego generate handler User --methods GET,POST,PUT,DELETE
  plumego generate model User --with-validation
`
}

func (c *GenerateCmd) Flags() []Flag {
	return []Flag{
		{Name: "output", Default: "", Usage: "Output file path"},
		{Name: "package", Default: "", Usage: "Package name"},
		{Name: "methods", Default: "GET", Usage: "HTTP methods (comma-separated)"},
		{Name: "with-tests", Default: false, Usage: "Generate test file"},
		{Name: "with-validation", Default: false, Usage: "Generate validation"},
		{Name: "force", Default: false, Usage: "Overwrite existing files"},
	}
}

func (c *GenerateCmd) Run(args []string) error {
	c.formatter = output.NewFormatter()
	c.formatter.SetFormat(flagFormat)
	c.formatter.SetQuiet(flagQuiet)
	c.formatter.SetVerbose(flagVerbose)

	if len(args) < 2 {
		return c.formatter.Error("generate type and name required (e.g., plumego generate component Auth)", 1)
	}

	genType := args[0]
	name := args[1]
	restArgs := args[2:]

	// Parse flags
	outputPath := ""
	packageName := ""
	methods := "GET"
	withTests := false
	withValidation := false
	force := false

	for i := 0; i < len(restArgs); i++ {
		switch restArgs[i] {
		case "--output":
			if i+1 < len(restArgs) {
				outputPath = restArgs[i+1]
				i++
			}
		case "--package":
			if i+1 < len(restArgs) {
				packageName = restArgs[i+1]
				i++
			}
		case "--methods":
			if i+1 < len(restArgs) {
				methods = restArgs[i+1]
				i++
			}
		case "--with-tests":
			withTests = true
		case "--with-validation":
			withValidation = true
		case "--force":
			force = true
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	c.formatter.Verbose(fmt.Sprintf("Generating %s: %s", genType, name))

	opts := codegen.GenerateOptions{
		Type:           genType,
		Name:           name,
		OutputPath:     outputPath,
		PackageName:    packageName,
		Methods:        methods,
		WithTests:      withTests,
		WithValidation: withValidation,
		Force:          force,
	}

	result, err := codegen.Generate(cwd, opts)
	if err != nil {
		return c.formatter.Error(fmt.Sprintf("generation failed: %v", err), 1)
	}

	return c.formatter.Success(fmt.Sprintf("%s generated successfully", genType), result)
}
