package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

type metricsSink struct {
	run   int32
	retry int32
	drop  int32
}

func (s *metricsSink) ObserveRun(jobID JobID, duration time.Duration, err error, queueDelay time.Duration) {
	atomic.AddInt32(&s.run, 1)
}

func (s *metricsSink) ObserveRetry(jobID JobID, attempt int, delay time.Duration) {
	atomic.AddInt32(&s.retry, 1)
}

func (s *metricsSink) ObserveDrop(jobID JobID) {
	atomic.AddInt32(&s.drop, 1)
}

func TestParseCronSpecAndNext(t *testing.T) {
	spec, err := ParseCronSpec("*/15 * * * *")
	if err != nil {
		t.Fatalf("parse cron: %v", err)
	}

	base := time.Date(2025, time.January, 1, 10, 7, 30, 0, time.UTC)
	next := spec.Next(base)
	expected := time.Date(2025, time.January, 1, 10, 15, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestParseCronSpecSeconds(t *testing.T) {
	spec, err := ParseCronSpec("*/10 * * * * *")
	if err != nil {
		t.Fatalf("parse cron: %v", err)
	}
	base := time.Date(2025, time.January, 1, 10, 7, 3, 0, time.UTC)
	next := spec.Next(base)
	expected := time.Date(2025, time.January, 1, 10, 7, 10, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}

	spec, err = ParseCronSpec("* * * * *")
	if err != nil {
		t.Fatalf("parse cron: %v", err)
	}
	base = time.Date(2025, time.January, 1, 10, 7, 59, 0, time.UTC)
	next = spec.Next(base)
	expected = time.Date(2025, time.January, 1, 10, 8, 0, 0, time.UTC)
	if !next.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, next)
	}
}

func TestDelayJobRuns(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	done := make(chan struct{}, 1)
	_, err := s.Delay("delay-1", 25*time.Millisecond, func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected delayed job to run")
	}
}

func TestJobContextValues(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	done := make(chan struct{}, 1)
	_, err := s.Delay("ctx-1", 5*time.Millisecond, func(ctx context.Context) error {
		if id, ok := JobIDFromContext(ctx); !ok || id != "ctx-1" {
			t.Fatalf("expected job id ctx-1")
		}
		if attempt, ok := JobAttemptFromContext(ctx); !ok || attempt < 1 {
			t.Fatalf("expected attempt >= 1")
		}
		done <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected job to run")
	}
}

func TestRetryPolicy(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	var calls int32
	_, err := s.Delay("retry-1", 5*time.Millisecond, func(ctx context.Context) error {
		if atomic.AddInt32(&calls, 1) < 2 {
			return context.Canceled
		}
		return nil
	}, WithRetryPolicy(RetryPolicy{MaxAttempts: 2, Backoff: func(attempt int) time.Duration {
		return 5 * time.Millisecond
	}}))
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", calls)
	}
}

func TestPauseResumeAndStatus(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	var calls int32
	id, err := s.Delay("pause-1", 10*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&calls, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	if !s.Pause(id) {
		t.Fatalf("expected pause to succeed")
	}

	time.Sleep(30 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 0 {
		t.Fatalf("expected no calls while paused")
	}

	if !s.Resume(id) {
		t.Fatalf("expected resume to succeed")
	}

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call after resume, got %d", calls)
	}

	status, ok := s.Status(id)
	if !ok {
		t.Fatalf("expected status")
	}
	if status.LastRun.IsZero() {
		t.Fatalf("expected last run to be set")
	}
}

func TestDeadLetterCallback(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	dead := make(chan JobID, 1)
	_, err := s.Delay("dlq-1", 5*time.Millisecond, func(ctx context.Context) error {
		return context.Canceled
	}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}), WithDeadLetter(func(ctx context.Context, id JobID, err error) {
		dead <- id
	}))
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	select {
	case id := <-dead:
		if id != "dlq-1" {
			t.Fatalf("unexpected job id: %s", id)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected dead letter callback")
	}
}

func TestReplaceExistingDelay(t *testing.T) {
	s := New(WithWorkers(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	done := make(chan string, 2)
	_, err := s.Delay("replace-1", 20*time.Millisecond, func(ctx context.Context) error {
		done <- "old"
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	_, err = s.Delay("replace-1", 10*time.Millisecond, func(ctx context.Context) error {
		done <- "new"
		return nil
	}, ReplaceExisting())
	if err != nil {
		t.Fatalf("delay replace: %v", err)
	}

	select {
	case val := <-done:
		if val != "new" {
			t.Fatalf("expected replacement to run, got %s", val)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected replacement to run")
	}
}

func TestQueueBackpressureDrops(t *testing.T) {
	block := make(chan struct{})
	s := New(WithWorkers(1), WithQueueSize(1))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()
	defer close(block)

	_, err := s.Delay("blocker", 5*time.Millisecond, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	_, err = s.Delay("queued", 5*time.Millisecond, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	_, err = s.Delay("dropper", 5*time.Millisecond, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	time.Sleep(30 * time.Millisecond)
	stats := s.Stats()
	if stats.Dropped == 0 {
		t.Fatalf("expected dropped tasks due to backpressure")
	}
}

func TestPanicHandler(t *testing.T) {
	called := make(chan JobID, 1)
	s := New(WithWorkers(1), WithPanicHandler(func(ctx context.Context, id JobID, recovered any) {
		called <- id
	}))
	s.Start()
	defer func() { _ = s.Stop(context.Background()) }()

	_, err := s.Delay("panic-1", 5*time.Millisecond, func(ctx context.Context) error {
		panic("boom")
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	select {
	case id := <-called:
		if id != "panic-1" {
			t.Fatalf("unexpected job id: %s", id)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected panic handler to be called")
	}
}

func TestPersistenceReload(t *testing.T) {
	store := NewMemoryStore()
	done := make(chan struct{}, 1)

	s := New(WithWorkers(1), WithStore(store))
	if err := s.RegisterTask("persisted", func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("register task: %v", err)
	}
	s.Start()
	_, err := s.Delay("persist-1", 10*time.Millisecond, func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	}, WithTaskName("persisted"))
	if err != nil {
		t.Fatalf("delay: %v", err)
	}
	_ = s.Stop(context.Background())

	s2 := New(WithWorkers(1), WithStore(store))
	if err := s2.RegisterTask("persisted", func(ctx context.Context) error {
		done <- struct{}{}
		return nil
	}); err != nil {
		t.Fatalf("register task: %v", err)
	}
	s2.Start()
	defer func() { _ = s2.Stop(context.Background()) }()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected persisted job to run")
	}
}

func TestMetricsSink(t *testing.T) {
	var sink metricsSink
	block := make(chan struct{})

	sched := New(WithWorkers(1), WithQueueSize(1), WithMetricsSink(&sink))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	_, err := sched.Delay("sink-2", 5*time.Millisecond, func(ctx context.Context) error {
		return context.Canceled
	}, WithRetryPolicy(RetryPolicy{MaxAttempts: 2, Backoff: func(_ int) time.Duration { return 5 * time.Millisecond }}))
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	_, err = sched.Delay("sink-1", 5*time.Millisecond, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	_, err = sched.Delay("sink-3", 5*time.Millisecond, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	time.AfterFunc(20*time.Millisecond, func() { close(block) })
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&sink.run) > 0 && atomic.LoadInt32(&sink.retry) > 0 && atomic.LoadInt32(&sink.drop) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	if atomic.LoadInt32(&sink.run) == 0 {
		t.Fatalf("expected run metrics")
	}
	if atomic.LoadInt32(&sink.retry) == 0 {
		t.Fatalf("expected retry metrics")
	}
	if atomic.LoadInt32(&sink.drop) == 0 {
		t.Fatalf("expected drop metrics")
	}
}
