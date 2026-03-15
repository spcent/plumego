package commands

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/spcent/plumego/cmd/plumego/internal/codegen"
)

// GenerateCmd generates code
type GenerateCmd struct{}

func (c *GenerateCmd) Name() string  { return "generate" }
func (c *GenerateCmd) Short() string { return "Generate middleware, handlers" }

func (c *GenerateCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	outputPath := fs.String("output", "", "Output file path")
	packageName := fs.String("package", "", "Package name")
	methods := fs.String("methods", "GET", "HTTP methods (comma-separated)")
	withTests := fs.Bool("with-tests", false, "Generate test file")
	withValidation := fs.Bool("with-validation", false, "Generate validation")
	force := fs.Bool("force", false, "Overwrite existing files")

	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	if len(positionals) < 2 {
		return ctx.Out.Error("generate type and name required (e.g., plumego generate handler Auth)", 1)
	}

	genType := positionals[0]
	name := positionals[1]

	cwd, err := os.Getwd()
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to get working directory: %v", err), 1)
	}

	ctx.Out.Verbose(fmt.Sprintf("Generating %s: %s", genType, name))

	result, err := codegen.Generate(cwd, codegen.GenerateOptions{
		Type:           genType,
		Name:           name,
		OutputPath:     *outputPath,
		PackageName:    *packageName,
		Methods:        *methods,
		WithTests:      *withTests,
		WithValidation: *withValidation,
		Force:          *force,
	})
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("generation failed: %v", err), 1)
	}

	return ctx.Out.Success(fmt.Sprintf("%s generated successfully", genType), result)
}
