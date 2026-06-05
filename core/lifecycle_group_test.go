package core

import (
	"context"
	"errors"
	"testing"
)

// recorder is a Lifecycle that records the order of Start/Stop calls into a
// shared slice and can be configured to fail any phase.
type recorder struct {
	name     string
	log      *[]string
	startErr error
	stopErr  error
	readyErr error
}

func (r *recorder) Start(context.Context) error {
	*r.log = append(*r.log, "start:"+r.name)
	return r.startErr
}

func (r *recorder) Stop(context.Context) error {
	*r.log = append(*r.log, "stop:"+r.name)
	return r.stopErr
}

func (r *recorder) Ready(context.Context) error { return r.readyErr }

func TestLifecycleGroupStartsInOrderStopsInReverse(t *testing.T) {
	var log []string
	g := NewLifecycleGroup()
	g.Add("a", &recorder{name: "a", log: &log}).
		Add("b", &recorder{name: "b", log: &log}).
		Add("c", &recorder{name: "c", log: &log})

	if err := g.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := g.Stop(context.Background()); err != nil {
		t.Fatalf("Stop: %v", err)
	}

	want := []string{"start:a", "start:b", "start:c", "stop:c", "stop:b", "stop:a"}
	if len(log) != len(want) {
		t.Fatalf("call log = %v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("call log = %v, want %v", log, want)
		}
	}
}

func TestLifecycleGroupRollsBackOnStartFailure(t *testing.T) {
	var log []string
	boom := errors.New("boom")
	g := NewLifecycleGroup()
	g.Add("a", &recorder{name: "a", log: &log}).
		Add("b", &recorder{name: "b", log: &log, startErr: boom}).
		Add("c", &recorder{name: "c", log: &log})

	err := g.Start(context.Background())
	if err == nil || !errors.Is(err, boom) {
		t.Fatalf("Start error = %v, want wrapping %v", err, boom)
	}
	// a started then b failed: a must be rolled back, c never started.
	want := []string{"start:a", "start:b", "stop:a"}
	if len(log) != len(want) {
		t.Fatalf("call log = %v, want %v", log, want)
	}
	for i := range want {
		if log[i] != want[i] {
			t.Fatalf("call log = %v, want %v", log, want)
		}
	}
	// A subsequent Stop must not re-stop anything (nothing is running).
	log = nil
	if err := g.Stop(context.Background()); err != nil {
		t.Fatalf("Stop after rollback: %v", err)
	}
	if len(log) != 0 {
		t.Fatalf("Stop after rollback ran %v, want no-op", log)
	}
}

func TestLifecycleGroupStopAggregatesErrors(t *testing.T) {
	var log []string
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	g := NewLifecycleGroup()
	g.Add("a", &recorder{name: "a", log: &log, stopErr: e1}).
		Add("b", &recorder{name: "b", log: &log, stopErr: e2})

	if err := g.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	err := g.Stop(context.Background())
	if err == nil || !errors.Is(err, e1) || !errors.Is(err, e2) {
		t.Fatalf("Stop error = %v, want both e1 and e2", err)
	}
}

func TestLifecycleGroupReadyAggregates(t *testing.T) {
	notReady := errors.New("warming up")
	g := NewLifecycleGroup()
	var log []string
	g.Add("a", &recorder{name: "a", log: &log}).
		Add("b", &recorder{name: "b", log: &log, readyErr: notReady})

	if err := g.Ready(context.Background()); err == nil || !errors.Is(err, notReady) {
		t.Fatalf("Ready error = %v, want wrapping %v", err, notReady)
	}

	// All ready -> nil.
	g2 := NewLifecycleGroup()
	g2.Add("a", &recorder{name: "a", log: &log})
	if err := g2.Ready(context.Background()); err != nil {
		t.Fatalf("Ready with all-ready = %v, want nil", err)
	}
}

func TestLifecycleGroupSecondStartFails(t *testing.T) {
	var log []string
	g := NewLifecycleGroup()
	g.Add("a", &recorder{name: "a", log: &log})
	if err := g.Start(context.Background()); err != nil {
		t.Fatalf("first Start: %v", err)
	}
	if err := g.Start(context.Background()); err == nil {
		t.Fatal("second Start should fail")
	}
}

func TestComponentAdapterOptionalPhases(t *testing.T) {
	var started, stopped bool
	c := Component{
		Name:    "worker",
		OnStart: func(context.Context) error { started = true; return nil },
		OnStop:  func(context.Context) error { stopped = true; return nil },
		// OnReady nil -> always ready.
	}
	ctx := context.Background()
	if err := c.Start(ctx); err != nil || !started {
		t.Fatalf("Start: err=%v started=%v", err, started)
	}
	if err := c.Ready(ctx); err != nil {
		t.Fatalf("Ready with nil OnReady should be nil, got %v", err)
	}
	if err := c.Stop(ctx); err != nil || !stopped {
		t.Fatalf("Stop: err=%v stopped=%v", err, stopped)
	}
}

func TestLifecycleGroupAddNilIsNoop(t *testing.T) {
	g := NewLifecycleGroup()
	g.Add("nil", nil)
	if err := g.Start(context.Background()); err != nil {
		t.Fatalf("Start with only nil component: %v", err)
	}
}

// Compile-time assertions: the group and the adapter satisfy Lifecycle, and a
// group composes into another group.
var (
	_ Lifecycle    = (*LifecycleGroup)(nil)
	_ Lifecycle    = Component{}
	_ Starter      = (*LifecycleGroup)(nil)
	_ Stopper      = (*LifecycleGroup)(nil)
	_ ReadyChecker = (*LifecycleGroup)(nil)
)
