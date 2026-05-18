package commands

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"

	"github.com/spcent/plumego/cmd/plumego/internal/codegen"
)

// GenerateCmd generates code
type GenerateCmd struct{}

func (c *GenerateCmd) Name() string  { return "generate" }
func (c *GenerateCmd) Short() string { return "Generate middleware, handlers, models" }

func (c *GenerateCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	outputPath := fs.String("output", "", "Output file path")
	dir := fs.String("dir", ".", "Project directory")
	format := fs.String("format", "json", "Spec output format: json or yaml")
	appPath := fs.String("app", ".", "Application package path to introspect")
	packageName := fs.String("package", "", "Package name")
	methods := fs.String("methods", "GET", "HTTP methods (comma-separated)")
	withTests := fs.Bool("with-tests", false, "Generate test file")
	withValidation := fs.Bool("with-validation", false, "Generate validation")
	force := fs.Bool("force", false, "Overwrite existing files")

	positionals, err := parseInterspersedFlags(fs, args)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	if len(positionals) == 1 && positionals[0] == "spec" {
		absDir, err := resolveDir(*dir)
		if err != nil {
			return ctx.Out.Error(err.Error(), 1)
		}
		return runGenerateSpec(ctx, generateSpecOptions{
			Dir:        absDir,
			OutputPath: *outputPath,
			Format:     *format,
			AppPath:    *appPath,
		})
	}

	if len(positionals) < 2 {
		return ctx.Out.Error("generate type and name required (e.g., plumego generate handler Auth)", 1)
	}
	if positionals[0] == "spec" {
		return ctx.Out.Error(fmt.Sprintf("unexpected arguments for generate spec: %v", positionals[1:]), 1)
	}
	if len(positionals) > 2 {
		return ctx.Out.Error(fmt.Sprintf("unexpected arguments: %v", positionals[2:]), 1)
	}

	genType := positionals[0]
	name := positionals[1]

	absDir, err := resolveDir(*dir)
	if err != nil {
		return ctx.Out.Error(err.Error(), 1)
	}

	resolvedOutputPath := *outputPath
	if resolvedOutputPath != "" && !filepath.IsAbs(resolvedOutputPath) {
		resolvedOutputPath = filepath.Join(absDir, resolvedOutputPath)
	}

	ctx.Out.Verbose(fmt.Sprintf("Generating %s: %s", genType, name))

	result, err := codegen.Generate(absDir, codegen.GenerateOptions{
		Type:           genType,
		Name:           name,
		OutputPath:     resolvedOutputPath,
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
