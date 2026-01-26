package core

import (
	"context"
	"fmt"
	"testing"
)

type stubRunner struct {
	started  bool
	stopped  bool
	startErr error
	stopErr  error
}

func (s *stubRunner) Start(ctx context.Context) error {
	s.started = true
	return s.startErr
}

func (s *stubRunner) Stop(ctx context.Context) error {
	s.stopped = true
	return s.stopErr
}

func TestStartRunners(t *testing.T) {
	t.Run("start all successfully", func(t *testing.T) {
		app := New()
		r1 := &stubRunner{}
		r2 := &stubRunner{}
		r3 := &stubRunner{}
		app.runners = []Runner{r1, r2, r3}

		err := app.startRunners(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !r1.started || !r2.started || !r3.started {
			t.Fatalf("expected all runners to start")
		}
		if len(app.startedRunners) != 3 {
			t.Fatalf("expected 3 started runners, got %d", len(app.startedRunners))
		}
	})

	t.Run("start with nil runners", func(t *testing.T) {
		app := New()
		app.runners = []Runner{nil, &stubRunner{}, nil}

		err := app.startRunners(context.Background())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("start with error stops previous", func(t *testing.T) {
		app := New()
		r1 := &stubRunner{}
		r2 := &stubRunner{startErr: fmt.Errorf("boom")}
		r3 := &stubRunner{}
		app.runners = []Runner{r1, r2, r3}

		err := app.startRunners(context.Background())

		if err == nil {
			t.Fatalf("expected error from runner start")
		}
		if !r1.started {
			t.Fatalf("runner 1 should have started")
		}
		if r3.started {
			t.Fatalf("runner 3 should not have started")
		}
		if !r1.stopped {
			t.Fatalf("runner 1 should be stopped after failure")
		}
	})
}

func TestStopRunners(t *testing.T) {
	t.Run("stop all runners", func(t *testing.T) {
		app := New()
		r1 := &stubRunner{}
		r2 := &stubRunner{}
		r3 := &stubRunner{}
		app.startedRunners = []Runner{r1, r2, r3}

		app.stopRunners(context.Background())

		if !r1.stopped || !r2.stopped || !r3.stopped {
			t.Fatalf("expected all runners to stop")
		}
	})

	t.Run("stop only once", func(t *testing.T) {
		app := New()
		runner := &stubRunner{}
		app.startedRunners = []Runner{runner}

		app.stopRunners(context.Background())
		app.stopRunners(context.Background())

		if !runner.stopped {
			t.Fatalf("runner should be stopped")
		}
	})
}
