package commands

import (
	"bytes"
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

func (c *BuildCmd) Name() string  { return "build" }
func (c *BuildCmd) Short() string { return "Build application with optimizations" }

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

	absDir, err := resolveDir(*dir)
	if err != nil {
		return ctx.Out.Error(err.Error(), 1)
	}

	absOutput, err := filepath.Abs(*outputPath)
	if err != nil {
		return ctx.Out.Error(fmt.Sprintf("invalid output path: %v", err), 1)
	}

	outputDir := filepath.Dir(absOutput)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ctx.Out.Error(fmt.Sprintf("failed to create output directory: %v", err), 1)
	}

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

	goVersion, err := getGoVersion()
	if err != nil {
		goVersion = "unknown"
	}

	gitCommit := getGitCommit(absDir)
	startTime := time.Now()

	cmd := exec.Command("go", buildArgs...)
	cmd.Dir = absDir
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	if ctx.Out.IsVerbose() {
		ctx.Out.Verbose(fmt.Sprintf("Building: go %s", strings.Join(buildArgs, " ")))
	}

	buildCommand := fmt.Sprintf("go %s", strings.Join(buildArgs, " "))
	if err := cmd.Run(); err != nil {
		buildOutput := strings.TrimSpace(stdoutBuf.String() + "\n" + stderrBuf.String())
		if ctx.Out.Format() == "text" && !ctx.Out.IsQuiet() && buildOutput != "" {
			_ = ctx.Out.Textf("%s\n", buildOutput)
		}

		errData := map[string]any{
			"command": buildCommand,
		}
		if buildOutput != "" {
			errData["output"] = buildOutput
		}
		return ctx.Out.Error(fmt.Sprintf("build failed: %v", err), 1, errData)
	}

	buildOutput := strings.TrimSpace(stdoutBuf.String() + "\n" + stderrBuf.String())
	if ctx.Out.Format() == "text" && !ctx.Out.IsQuiet() && buildOutput != "" {
		_ = ctx.Out.Textf("%s\n", buildOutput)
	}

	buildTime := time.Since(startTime)

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
	if buildOutput != "" {
		result["build_output"] = buildOutput
	}

	return ctx.Out.Success("Build completed successfully", result)
}

func getGoVersion() (string, error) {
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

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
