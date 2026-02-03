package commands

import (
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type BuildCmd struct{}

func (c *BuildCmd) Name() string {
	return "build"
}

func (c *BuildCmd) Short() string {
	return "Build application with optimizations"
}

func (c *BuildCmd) Long() string {
	return `Build the Go application with plumego-specific optimizations.

This command wraps 'go build' with useful defaults and options for
building production-ready binaries.

Examples:
  plumego build
  plumego build --output ./bin/app
  plumego build --ldflags "-X main.version=1.0.0"
  plumego build --tags prod
  plumego build --format json`
}

func (c *BuildCmd) Flags() []Flag {
	return []Flag{
		{Name: "dir", Default: ".", Usage: "Project directory"},
		{Name: "output", Default: "./bin/app", Usage: "Output binary path"},
		{Name: "ldflags", Default: "", Usage: "Go linker flags"},
		{Name: "tags", Default: "", Usage: "Build tags (comma-separated)"},
		{Name: "race", Default: "false", Usage: "Enable race detector"},
		{Name: "trimpath", Default: "true", Usage: "Remove file system paths from binary"},
	}
}

func (c *BuildCmd) Run(ctx *Context, args []string) error {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	dir := fs.String("dir", ".", "Project directory")
	outputPath := fs.String("output", "./bin/app", "Output binary path")
	ldflags := fs.String("ldflags", "", "Go linker flags")
	tags := fs.String("tags", "", "Build tags")
	race := fs.Bool("race", false, "Enable race detector")
	trimpath := fs.Bool("trimpath", true, "Remove file system paths")

	if err := fs.Parse(args); err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid flags: %v", err), 1)
	}

	// Get absolute directory
	absDir, err := filepath.Abs(*dir)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid directory: %v", err), 1)
	}

	// Check if directory exists
	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		return ctx.Out.Error(fmt.Sprintf("directory not found: %s", absDir), 1)
	}

	// Get output path
	absOutput, err := filepath.Abs(*outputPath)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid output path: %v", err), 1)
	}

	// Create output directory
	outputDir := filepath.Dir(absOutput)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to create output directory: %v", err), 1)
	}

	// Build command
	buildArgs := []string{"build"}

	if *race {
		buildArgs = append(buildArgs, "-race")
	}

	if *trimpath {
		buildArgs = append(buildArgs, "-trimpath")
	}

	if *ldflags != "" {
		buildArgs = append(buildArgs, "-ldflags", *ldflags)
	}

	if *tags != "" {
		buildArgs = append(buildArgs, "-tags", *tags)
	}

	buildArgs = append(buildArgs, "-o", absOutput)

	// Get Go version
	goVersion, err := getGoVersion()
	if err != nil {
		goVersion = "unknown"
	}

	// Get git commit
	gitCommit := getGitCommit(absDir)

	startTime := time.Now()

	// Run build
	cmd := exec.Command("go", buildArgs...)
	cmd.Dir = absDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if flagVerbose {
		ctx.Out.Verbose(fmt.Sprintf("Building: go %s", strings.Join(buildArgs, " ")))
	}

	if err := cmd.Run(); err != nil {
		return ctx.Out.Error(fmt.Sprintf("build failed: %v", err), 1)
	}

	buildTime := time.Since(startTime)

	// Get binary size
	fileInfo, err := os.Stat(absOutput)
	var sizeBytes int64
	if err == nil {
		sizeBytes = fileInfo.Size()
	}

	result := map[string]any{
		"binary":        absOutput,
		"size_bytes":    sizeBytes,
		"size_mb":       float64(sizeBytes) / (1024 * 1024),
		"build_time_ms": buildTime.Milliseconds(),
		"go_version":    goVersion,
		"git_commit":    gitCommit,
		"race_detector": *race,
		"trimpath":      *trimpath,
	}

	if *tags != "" {
		result["build_tags"] = strings.Split(*tags, ",")
	}

	return ctx.Out.Success("Build completed successfully", result)
}

func getGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse "go version go1.24.0 linux/amd64"
	parts := strings.Fields(string(output))
	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go"), nil
	}

	return string(output), nil
}

func getGitCommit(dir string) string {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}
