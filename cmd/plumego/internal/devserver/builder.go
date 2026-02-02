package devserver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spcent/plumego/pubsub"
)

// Builder manages application builds
type Builder struct {
	dir        string
	outputPath string
	buildCmd   string
	buildArgs  []string

	pubsub *pubsub.InProcPubSub
}

// NewBuilder creates a new builder
func NewBuilder(dir string, ps *pubsub.InProcPubSub) *Builder {
	return &Builder{
		dir:        dir,
		outputPath: filepath.Join(dir, ".dev-server"),
		pubsub:     ps,
	}
}

// SetCustomBuild sets a custom build command
func (b *Builder) SetCustomBuild(cmd string, args []string) {
	b.buildCmd = cmd
	b.buildArgs = args
}

// Build builds the application
func (b *Builder) Build() error {
	start := time.Now()

	// Publish build start event
	b.pubsub.Publish(EventBuildStart, pubsub.Message{
		Topic: EventBuildStart,
		Data: BuildEvent{
			Success: false,
		},
	})

	var cmd *exec.Cmd
	if b.buildCmd != "" {
		// Use custom build command
		cmd = exec.Command(b.buildCmd, b.buildArgs...)
	} else {
		// Default: go build
		cmd = exec.Command("go", "build", "-o", b.outputPath, ".")
	}

	cmd.Dir = b.dir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run build
	err := cmd.Run()
	duration := time.Since(start)

	// Prepare output
	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	if err != nil {
		// Build failed
		errMsg := fmt.Sprintf("Build failed: %v", err)
		if output != "" {
			errMsg = output
		}

		b.pubsub.Publish(EventBuildFail, pubsub.Message{
			Topic: EventBuildFail,
			Data: BuildEvent{
				Success:  false,
				Duration: duration,
				Error:    errMsg,
				Output:   output,
			},
		})

		return fmt.Errorf("build failed: %w\n%s", err, output)
	}

	// Build succeeded
	b.pubsub.Publish(EventBuildSuccess, pubsub.Message{
		Topic: EventBuildSuccess,
		Data: BuildEvent{
			Success:  true,
			Duration: duration,
			Output:   output,
		},
	})

	return nil
}

// Clean removes build artifacts
func (b *Builder) Clean() error {
	if _, err := os.Stat(b.outputPath); err == nil {
		return os.Remove(b.outputPath)
	}
	return nil
}

// OutputPath returns the path to the built binary
func (b *Builder) OutputPath() string {
	return b.outputPath
}

// Verify checks if the build environment is valid
func (b *Builder) Verify() error {
	// Check if go.mod exists
	goModPath := filepath.Join(b.dir, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return fmt.Errorf("go.mod not found in %s", b.dir)
	}

	// Check if main package exists
	hasMain := false
	entries, err := os.ReadDir(b.dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
			// Check if file contains package main
			content, err := os.ReadFile(filepath.Join(b.dir, entry.Name()))
			if err != nil {
				continue
			}
			if strings.Contains(string(content), "package main") {
				hasMain = true
				break
			}
		}
	}

	if !hasMain {
		return fmt.Errorf("no main package found in %s", b.dir)
	}

	return nil
}
