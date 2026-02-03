package commands

import (
	"bytes"
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spcent/plumego/cmd/plumego/internal/devserver"
	"github.com/spcent/plumego/cmd/plumego/internal/output"
	"github.com/spcent/plumego/pubsub"
)

type fakeBuilder struct {
	cmd  string
	args []string
}

func (b *fakeBuilder) SetCustomBuild(cmd string, args []string) {
	b.cmd = cmd
	b.args = args
}

type fakeRunner struct {
	cmd         string
	args        []string
	passthrough bool
}

func (r *fakeRunner) SetCustomCommand(cmd string, args []string) {
	r.cmd = cmd
	r.args = args
}

func (r *fakeRunner) SetOutputPassthrough(enabled bool) {
	r.passthrough = enabled
}

type fakeDashboard struct {
	builder devserver.BuilderAPI
	runner  devserver.RunnerAPI
	pubsub  *pubsub.InProcPubSub

	started bool
	stopped bool
	built   bool
}

func (d *fakeDashboard) Start(context.Context) error {
	d.started = true
	return nil
}

func (d *fakeDashboard) Stop(context.Context) error {
	d.stopped = true
	return nil
}

func (d *fakeDashboard) BuildAndRun(context.Context) error {
	d.built = true
	return nil
}

func (d *fakeDashboard) Rebuild(context.Context) error { return nil }

func (d *fakeDashboard) PublishEvent(string, interface{}) {}

func (d *fakeDashboard) GetPubSub() *pubsub.InProcPubSub { return d.pubsub }

func (d *fakeDashboard) GetBuilder() devserver.BuilderAPI { return d.builder }

func (d *fakeDashboard) GetRunner() devserver.RunnerAPI { return d.runner }

func TestParseDevArgs(t *testing.T) {
	opts, err := parseDevArgs([]string{
		"--no-reload",
		"--build-cmd", "go build -o .dev-server ./cmd/api",
		"--run-cmd", "./.dev-server",
	})
	if err != nil {
		t.Fatalf("parseDevArgs failed: %v", err)
	}

	if !opts.noReload {
		t.Fatalf("expected noReload true")
	}
	if opts.buildCmd == "" || opts.runCmd == "" {
		t.Fatalf("expected build and run commands to be set")
	}
}

func TestDevRunNoReload(t *testing.T) {
	tmpDir := t.TempDir()

	builder := &fakeBuilder{}
	runner := &fakeRunner{}
	dash := &fakeDashboard{
		builder: builder,
		runner:  runner,
		pubsub:  pubsub.New(),
	}

	origFactory := newDevDashboard
	newDevDashboard = func(cfg devserver.Config) (devserver.DashboardAPI, error) {
		return dash, nil
	}
	t.Cleanup(func() { newDevDashboard = origFactory })

	out := output.NewFormatter()
	out.SetFormat("text")
	var buf bytes.Buffer
	out.SetWriters(&buf, &buf)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	cmd := &DevCmd{}
	err := cmd.runWithContext(ctx, out, devOptions{
		dir:           tmpDir,
		addr:          ":8080",
		dashboardAddr: ":9999",
		debounceStr:   "500ms",
		noReload:      true,
	})
	if err != nil {
		t.Fatalf("runWithContext failed: %v", err)
	}

	if !dash.started || !dash.built {
		t.Fatalf("expected dashboard to start and build")
	}
	if !strings.Contains(buf.String(), "Auto reload disabled") {
		t.Fatalf("expected no-reload message, got: %s", buf.String())
	}
}

func TestDevRunBuildCmd(t *testing.T) {
	tmpDir := t.TempDir()

	builder := &fakeBuilder{}
	runner := &fakeRunner{}
	dash := &fakeDashboard{
		builder: builder,
		runner:  runner,
		pubsub:  pubsub.New(),
	}

	origFactory := newDevDashboard
	newDevDashboard = func(cfg devserver.Config) (devserver.DashboardAPI, error) {
		if filepath.Clean(cfg.ProjectDir) != filepath.Clean(tmpDir) {
			t.Fatalf("expected project dir %s, got %s", tmpDir, cfg.ProjectDir)
		}
		return dash, nil
	}
	t.Cleanup(func() { newDevDashboard = origFactory })

	out := output.NewFormatter()
	out.SetFormat("text")

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	cmd := &DevCmd{}
	err := cmd.runWithContext(ctx, out, devOptions{
		dir:           tmpDir,
		addr:          ":8080",
		dashboardAddr: ":9999",
		debounceStr:   "500ms",
		noReload:      true,
		buildCmd:      "go build -o .dev-server ./cmd/api",
	})
	if err != nil {
		t.Fatalf("runWithContext failed: %v", err)
	}

	if builder.cmd != "go" {
		t.Fatalf("expected build cmd 'go', got %q", builder.cmd)
	}
	if len(builder.args) < 3 || builder.args[0] != "build" {
		t.Fatalf("unexpected build args: %#v", builder.args)
	}
}
