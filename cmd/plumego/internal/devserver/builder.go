package devserver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/buildtarget"
	"github.com/spcent/plumego/x/pubsub"
)

// Builder manages application builds
type Builder struct {
	dir        string
	outputPath string
	buildCmd   string
	buildArgs  []string

	pubsub *pubsub.InProcBroker
}

// NewBuilder creates a new builder
func NewBuilder(dir string, ps *pubsub.InProcBroker) *Builder {
	return &Builder{
		dir:        dir,
		outputPath: filepath.Join(dir, ".dev-server"),
		pubsub:     ps,
	}
}

// HasCustomBuild reports whether a custom build command is configured.
func (b *Builder) HasCustomBuild() bool {
	return b.buildCmd != ""
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
		cmd = exec.Command("go", "build", "-o", b.outputPath, buildtarget.Default(b.dir))
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

	if b.buildCmd != "" {
		return nil
	}

	if !buildtarget.HasDefaultEntrypoint(b.dir) {
		return fmt.Errorf("no main package found in %s or %s", b.dir, filepath.Join(b.dir, "cmd", "app"))
	}

	return nil
}
