package devserver

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/spcent/plumego/x/pubsub"
)

func TestAppRunnerStopWaitsThroughSingleOwner(t *testing.T) {
	runner := newHelperAppRunner(t)

	if err := runner.Start(context.Background()); err != nil {
		t.Fatalf("start runner: %v", err)
	}
	if !runner.IsRunning() {
		t.Fatal("runner should be marked running after start")
	}

	if err := runner.Stop(); err != nil {
		t.Fatalf("stop runner: %v", err)
	}
	if runner.IsRunning() {
		t.Fatal("runner should not be marked running after stop")
	}

	if err := runner.Stop(); err != nil {
		t.Fatalf("second stop should be idempotent: %v", err)
	}
}

func TestAppRunnerRejectsStartWhileRunning(t *testing.T) {
	runner := newHelperAppRunner(t)

	if err := runner.Start(context.Background()); err != nil {
		t.Fatalf("start runner: %v", err)
	}
	defer runner.Stop()

	if err := runner.Start(context.Background()); err == nil {
		t.Fatal("second start should fail while runner is already running")
	}
}

func TestAppRunnerStartFailureRestoresStartingState(t *testing.T) {
	runner := NewAppRunner(t.TempDir(), pubsub.New())
	runner.SetOutputPassthrough(false)
	runner.SetCustomCommand(filepath.Join(t.TempDir(), "missing-command"), nil)

	if err := runner.Start(context.Background()); err == nil {
		t.Fatal("expected start failure")
	}
	if runner.IsRunning() {
		t.Fatal("runner should not be marked running after start failure")
	}

	runner.SetCustomCommand(os.Args[0], []string{"-test.run=TestAppRunnerHelperProcess"})
	runner.SetEnv("PLUMEGO_RUNNER_HELPER", "1")
	runner.stopTimeout = 2 * time.Second
	if err := runner.Start(context.Background()); err != nil {
		t.Fatalf("runner should start after failed attempt: %v", err)
	}
	if err := runner.Stop(); err != nil {
		t.Fatalf("stop runner: %v", err)
	}
}

func TestAppRunnerHelperProcess(t *testing.T) {
	if os.Getenv("PLUMEGO_RUNNER_HELPER") != "1" {
		return
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	fmt.Println("runner helper ready")
	<-signals
	os.Exit(0)
}

func newHelperAppRunner(t *testing.T) *AppRunner {
	t.Helper()

	runner := NewAppRunner(t.TempDir(), pubsub.New())
	runner.SetOutputPassthrough(false)
	runner.SetCustomCommand(os.Args[0], []string{"-test.run=TestAppRunnerHelperProcess"})
	runner.SetEnv("PLUMEGO_RUNNER_HELPER", "1")
	runner.stopTimeout = 2 * time.Second
	return runner
}
