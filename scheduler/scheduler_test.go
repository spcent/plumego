package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
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

func TestBackpressureDrop(t *testing.T) {
	// Test default drop behavior when queue is full
	var dropped atomic.Int32
	config := BackpressureConfig{
		Policy: BackpressureDrop,
		OnBackpressure: func(jobID JobID) {
			dropped.Add(1)
		},
	}

	sched := New(
		WithWorkers(1),
		WithQueueSize(1),
		WithBackpressure(config),
	)
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Block the worker
	block := make(chan struct{})
	_, err := sched.Delay("block", 5*time.Millisecond, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	// Wait for job to be picked up by worker
	time.Sleep(20 * time.Millisecond)

	// Schedule jobs to fill the queue and trigger drops
	for i := 0; i < 10; i++ {
		_, _ = sched.Delay(JobID("job-"+string(rune(i))), 5*time.Millisecond, func(ctx context.Context) error {
			return nil
		})
	}

	close(block)
	time.Sleep(50 * time.Millisecond)

	// Verify drops occurred
	stats := sched.Stats()
	if stats.Dropped == 0 {
		t.Fatalf("expected dropped jobs, got 0")
	}
	if dropped.Load() == 0 {
		t.Fatalf("expected backpressure callback to be invoked")
	}
}

func TestBackpressureBlock(t *testing.T) {
	// Test blocking behavior - all jobs should be queued eventually
	config := BackpressureConfig{
		Policy: BackpressureBlock,
	}

	sched := New(
		WithWorkers(1),
		WithQueueSize(2),
		WithBackpressure(config),
	)
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var executed atomic.Int32
	block := make(chan struct{})

	// Block the worker temporarily
	_, err := sched.Delay("block", 5*time.Millisecond, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	// Wait for job to be picked up
	time.Sleep(20 * time.Millisecond)

	// Schedule multiple jobs - they should block but not drop
	done := make(chan struct{})
	go func() {
		for i := 0; i < 5; i++ {
			_, _ = sched.Delay(JobID("job-"+string(rune(i))), 5*time.Millisecond, func(ctx context.Context) error {
				executed.Add(1)
				return nil
			})
		}
		close(done)
	}()

	// Unblock after a delay
	time.Sleep(50 * time.Millisecond)
	close(block)

	// Wait for all scheduling to complete
	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatalf("scheduling timed out")
	}

	// Wait for execution
	time.Sleep(200 * time.Millisecond)

	stats := sched.Stats()
	if stats.Dropped > 0 {
		t.Fatalf("expected no drops with blocking policy, got %d", stats.Dropped)
	}
	if executed.Load() == 0 {
		t.Fatalf("expected jobs to execute")
	}
}

func TestBackpressureBlockTimeout(t *testing.T) {
	// Test timeout blocking behavior
	var dropped atomic.Int32
	config := BackpressureConfig{
		Policy:  BackpressureBlockTimeout,
		Timeout: 30 * time.Millisecond,
		OnBackpressure: func(jobID JobID) {
			dropped.Add(1)
		},
	}

	sched := New(
		WithWorkers(1),
		WithQueueSize(1),
		WithBackpressure(config),
	)
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Block the worker
	block := make(chan struct{})
	_, err := sched.Delay("block", 5*time.Millisecond, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("delay: %v", err)
	}

	// Wait for job to be picked up
	time.Sleep(20 * time.Millisecond)

	// Schedule jobs - some should timeout
	for i := 0; i < 5; i++ {
		_, _ = sched.Delay(JobID("job-"+string(rune(i))), 5*time.Millisecond, func(ctx context.Context) error {
			return nil
		})
	}

	// Don't unblock immediately to force timeouts
	time.Sleep(100 * time.Millisecond)
	close(block)
	time.Sleep(100 * time.Millisecond)

	stats := sched.Stats()
	if dropped.Load() == 0 {
		t.Fatalf("expected backpressure callback for timeouts")
	}
	if stats.Dropped == 0 {
		t.Fatalf("expected some jobs to be dropped due to timeout")
	}
}

func TestBatchOperationsByGroup(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Create jobs in different groups
	_, _ = sched.AddCron("group-a-1", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-a"))

	_, _ = sched.AddCron("group-a-2", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-a"))

	_, _ = sched.AddCron("group-b-1", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-b"))

	time.Sleep(20 * time.Millisecond)

	// Test PauseByGroup
	paused := sched.PauseByGroup("group-a")
	if paused != 2 {
		t.Fatalf("expected 2 jobs paused, got %d", paused)
	}

	// Verify jobs are paused
	status, ok := sched.Status("group-a-1")
	if !ok || !status.Paused {
		t.Fatalf("expected group-a-1 to be paused")
	}

	status, ok = sched.Status("group-b-1")
	if !ok || status.Paused {
		t.Fatalf("expected group-b-1 to not be paused")
	}

	// Test ResumeByGroup
	resumed := sched.ResumeByGroup("group-a")
	if resumed != 2 {
		t.Fatalf("expected 2 jobs resumed, got %d", resumed)
	}

	status, ok = sched.Status("group-a-1")
	if !ok || status.Paused {
		t.Fatalf("expected group-a-1 to be resumed")
	}

	// Test CancelByGroup
	canceled := sched.CancelByGroup("group-a")
	if canceled != 2 {
		t.Fatalf("expected 2 jobs canceled, got %d", canceled)
	}

	status, ok = sched.Status("group-a-1")
	if !ok || status.State != JobStateCanceled {
		t.Fatalf("expected group-a-1 to be canceled")
	}

	_, ok = sched.Status("group-b-1")
	if !ok {
		t.Fatalf("expected group-b-1 to still exist")
	}
}

func TestBatchOperationsByTags(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Create jobs with different tags
	_, _ = sched.AddCron("tagged-1", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithTags("urgent", "cleanup"))

	_, _ = sched.AddCron("tagged-2", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithTags("urgent", "cleanup"))

	_, _ = sched.AddCron("tagged-3", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithTags("urgent"))

	time.Sleep(20 * time.Millisecond)

	// Test PauseByTags - only jobs with ALL tags
	paused := sched.PauseByTags("urgent", "cleanup")
	if paused != 2 {
		t.Fatalf("expected 2 jobs paused, got %d", paused)
	}

	status, ok := sched.Status("tagged-1")
	if !ok || !status.Paused {
		t.Fatalf("expected tagged-1 to be paused")
	}

	status, ok = sched.Status("tagged-3")
	if !ok || status.Paused {
		t.Fatalf("expected tagged-3 to not be paused (missing 'cleanup' tag)")
	}

	// Test ResumeByTags
	resumed := sched.ResumeByTags("urgent", "cleanup")
	if resumed != 2 {
		t.Fatalf("expected 2 jobs resumed, got %d", resumed)
	}

	// Test CancelByTags
	canceled := sched.CancelByTags("urgent", "cleanup")
	if canceled != 2 {
		t.Fatalf("expected 2 jobs canceled, got %d", canceled)
	}

	status, ok = sched.Status("tagged-1")
	if !ok || status.State != JobStateCanceled {
		t.Fatalf("expected tagged-1 to be canceled")
	}

	_, ok = sched.Status("tagged-3")
	if !ok {
		t.Fatalf("expected tagged-3 to still exist")
	}
}

func TestQueryJobs(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Create diverse jobs for testing
	_, _ = sched.AddCron("cron-a-1", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-a"), WithTags("urgent"))

	_, _ = sched.AddCron("cron-a-2", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-a"), WithTags("urgent", "cleanup"))

	_, _ = sched.Delay("delay-b-1", 5*time.Second, func(ctx context.Context) error {
		return nil
	}, WithGroup("group-b"), WithTags("batch"))

	_, _ = sched.AddCron("cron-b-2", "* * * * *", func(ctx context.Context) error {
		return nil
	}, WithGroup("group-b"))

	time.Sleep(20 * time.Millisecond)

	// Pause one job for testing
	sched.Pause("cron-a-2")

	// Test 1: Filter by group
	result := sched.QueryJobs(JobQuery{Group: "group-a"})
	if result.Total != 2 {
		t.Fatalf("expected 2 jobs in group-a, got %d", result.Total)
	}

	// Test 2: Filter by tags
	result = sched.QueryJobs(JobQuery{Tags: []string{"urgent", "cleanup"}})
	if result.Total != 1 {
		t.Fatalf("expected 1 job with both tags, got %d", result.Total)
	}

	// Test 3: Filter by kind
	result = sched.QueryJobs(JobQuery{Kinds: []string{"cron"}})
	if result.Total != 3 {
		t.Fatalf("expected 3 cron jobs, got %d", result.Total)
	}

	result = sched.QueryJobs(JobQuery{Kinds: []string{"delay"}})
	if result.Total != 1 {
		t.Fatalf("expected 1 delay job, got %d", result.Total)
	}

	// Test 4: Filter by paused status
	paused := true
	result = sched.QueryJobs(JobQuery{Paused: &paused})
	if result.Total != 1 {
		t.Fatalf("expected 1 paused job, got %d", result.Total)
	}

	notPaused := false
	result = sched.QueryJobs(JobQuery{Paused: &notPaused})
	if result.Total != 3 {
		t.Fatalf("expected 3 not paused jobs, got %d", result.Total)
	}

	// Test 5: Sorting by ID
	result = sched.QueryJobs(JobQuery{OrderBy: "id", Ascending: true})
	if len(result.Jobs) < 2 {
		t.Fatalf("expected at least 2 jobs")
	}
	if result.Jobs[0].ID > result.Jobs[1].ID {
		t.Fatalf("expected jobs sorted by ID ascending")
	}

	// Test 6: Pagination
	result = sched.QueryJobs(JobQuery{Limit: 2, OrderBy: "id", Ascending: true})
	if len(result.Jobs) != 2 {
		t.Fatalf("expected 2 jobs with limit, got %d", len(result.Jobs))
	}
	if result.Total != 4 {
		t.Fatalf("expected total count of 4, got %d", result.Total)
	}

	result = sched.QueryJobs(JobQuery{Offset: 2, OrderBy: "id", Ascending: true})
	if len(result.Jobs) != 2 {
		t.Fatalf("expected 2 jobs with offset, got %d", len(result.Jobs))
	}

	// Test 7: Combined filters
	result = sched.QueryJobs(JobQuery{
		Group:  "group-a",
		Paused: &notPaused,
	})
	if result.Total != 1 {
		t.Fatalf("expected 1 job (group-a, not paused), got %d", result.Total)
	}
}

func TestJobDependenciesSuccess(t *testing.T) {
	sched := New(WithWorkers(2))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var execOrder []string
	var mu sync.Mutex

	// Job A - no dependencies
	_, err := sched.Delay("job-a", 10*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		execOrder = append(execOrder, "job-a")
		mu.Unlock()
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to add job-a: %v", err)
	}

	// Job B - depends on A
	_, err = sched.Delay("job-b", 10*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		execOrder = append(execOrder, "job-b")
		mu.Unlock()
		return nil
	}, WithDependsOn(DependencyFailureSkip, "job-a"))
	if err != nil {
		t.Fatalf("failed to add job-b: %v", err)
	}

	// Job C - depends on A and B
	_, err = sched.Delay("job-c", 10*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		execOrder = append(execOrder, "job-c")
		mu.Unlock()
		return nil
	}, WithDependsOn(DependencyFailureSkip, "job-a", "job-b"))
	if err != nil {
		t.Fatalf("failed to add job-c: %v", err)
	}

	// Wait for all jobs to complete
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(execOrder) != 3 {
		t.Fatalf("expected 3 jobs executed, got %d: %v", len(execOrder), execOrder)
	}

	// Verify execution order: A must come before B and C
	aIndex := -1
	bIndex := -1
	cIndex := -1
	for i, job := range execOrder {
		switch job {
		case "job-a":
			aIndex = i
		case "job-b":
			bIndex = i
		case "job-c":
			cIndex = i
		}
	}

	if aIndex == -1 || bIndex == -1 || cIndex == -1 {
		t.Fatalf("not all jobs executed: %v", execOrder)
	}

	if aIndex > bIndex {
		t.Fatalf("job-a must execute before job-b")
	}
	if bIndex > cIndex {
		t.Fatalf("job-b must execute before job-c")
	}
}

func TestJobDependencyFailureSkip(t *testing.T) {
	sched := New(WithWorkers(2))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var executed atomic.Int32

	// Job A - will fail
	_, err := sched.Delay("job-a", 10*time.Millisecond, func(ctx context.Context) error {
		return fmt.Errorf("intentional failure")
	})
	if err != nil {
		t.Fatalf("failed to add job-a: %v", err)
	}

	// Job B - depends on A with Skip policy
	_, err = sched.Delay("job-b", 10*time.Millisecond, func(ctx context.Context) error {
		executed.Add(1)
		return nil
	}, WithDependsOn(DependencyFailureSkip, "job-a"))
	if err != nil {
		t.Fatalf("failed to add job-b: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 0 {
		t.Fatalf("expected job-b to be skipped, but it executed")
	}

	// Verify job-b still exists (not canceled)
	_, ok := sched.Status("job-b")
	if !ok {
		t.Fatalf("expected job-b to still exist after skip")
	}
}

func TestJobDependencyFailureCancel(t *testing.T) {
	sched := New(WithWorkers(2))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var executed atomic.Int32

	// Job A - will fail
	_, err := sched.Delay("job-a", 10*time.Millisecond, func(ctx context.Context) error {
		return fmt.Errorf("intentional failure")
	})
	if err != nil {
		t.Fatalf("failed to add job-a: %v", err)
	}

	// Job B - depends on A with Cancel policy
	_, err = sched.Delay("job-b", 10*time.Millisecond, func(ctx context.Context) error {
		executed.Add(1)
		return nil
	}, WithDependsOn(DependencyFailureCancel, "job-a"))
	if err != nil {
		t.Fatalf("failed to add job-b: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 0 {
		t.Fatalf("expected job-b to not execute")
	}

	// Verify job-b is canceled
	status, ok := sched.Status("job-b")
	if !ok || status.State != JobStateCanceled {
		t.Fatalf("expected job-b to be canceled")
	}
}

func TestJobDependencyFailureContinue(t *testing.T) {
	sched := New(WithWorkers(2))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var executed atomic.Int32

	// Job A - will fail
	_, err := sched.Delay("job-a", 10*time.Millisecond, func(ctx context.Context) error {
		return fmt.Errorf("intentional failure")
	})
	if err != nil {
		t.Fatalf("failed to add job-a: %v", err)
	}

	// Job B - depends on A with Continue policy
	_, err = sched.Delay("job-b", 10*time.Millisecond, func(ctx context.Context) error {
		executed.Add(1)
		return nil
	}, WithDependsOn(DependencyFailureContinue, "job-a"))
	if err != nil {
		t.Fatalf("failed to add job-b: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if executed.Load() != 1 {
		t.Fatalf("expected job-b to execute despite dependency failure")
	}
}

func TestJobDependencyValidation(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	// Test self-dependency
	_, err := sched.Delay("job-a", 10*time.Millisecond, func(ctx context.Context) error {
		return nil
	}, WithDependsOn(DependencyFailureSkip, "job-a"))
	if !errors.Is(err, ErrSelfDependency) {
		t.Fatalf("expected ErrSelfDependency, got: %v", err)
	}

	// Test non-existent dependency
	_, err = sched.Delay("job-b", 10*time.Millisecond, func(ctx context.Context) error {
		return nil
	}, WithDependsOn(DependencyFailureSkip, "non-existent"))
	if !errors.Is(err, ErrDependencyNotFound) {
		t.Fatalf("expected ErrDependencyNotFound, got: %v", err)
	}
}

func TestJobStateTransitions(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	block := make(chan struct{})
	_, err := sched.Delay("state-running", 0, func(ctx context.Context) error {
		<-block
		return nil
	})
	if err != nil {
		t.Fatalf("failed to add job: %v", err)
	}

	if !waitForState(t, sched, "state-running", JobStateRunning, 500*time.Millisecond) {
		t.Fatalf("expected state-running to reach running state")
	}

	close(block)

	if !waitForState(t, sched, "state-running", JobStateCompleted, 500*time.Millisecond) {
		t.Fatalf("expected state-running to reach completed state")
	}

	var attempts atomic.Int32
	_, err = sched.Delay("state-retry", 0, func(ctx context.Context) error {
		if attempts.Add(1) == 1 {
			return fmt.Errorf("fail once")
		}
		return nil
	}, WithRetryPolicy(RetryPolicy{
		MaxAttempts: 2,
		Backoff: func(attempt int) time.Duration {
			return 50 * time.Millisecond
		},
	}))
	if err != nil {
		t.Fatalf("failed to add retry job: %v", err)
	}

	if !waitForState(t, sched, "state-retry", JobStateRetrying, 200*time.Millisecond) {
		t.Fatalf("expected state-retry to enter retrying state")
	}

	_, err = sched.Delay("state-failed", 0, func(ctx context.Context) error {
		return fmt.Errorf("always fail")
	}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
	if err != nil {
		t.Fatalf("failed to add failed job: %v", err)
	}

	if !waitForState(t, sched, "state-failed", JobStateFailed, 500*time.Millisecond) {
		t.Fatalf("expected state-failed to reach failed state")
	}
}

func waitForState(t *testing.T, sched *Scheduler, id JobID, state JobState, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		status, ok := sched.Status(id)
		if ok && status.State == state {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}

func TestCronDescriptors(t *testing.T) {
	tests := []struct {
		name       string
		descriptor string
		wantErr    bool
	}{
		{"hourly", "@hourly", false},
		{"daily", "@daily", false},
		{"midnight", "@midnight", false},
		{"weekly", "@weekly", false},
		{"monthly", "@monthly", false},
		{"yearly", "@yearly", false},
		{"annually", "@annually", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseCronSpec(tt.descriptor)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCronSpec(%q) error = %v, wantErr %v", tt.descriptor, err, tt.wantErr)
			}
			if err == nil {
				// Verify it produces valid next times
				now := time.Now()
				next := spec.Next(now)
				if next.IsZero() || !next.After(now) {
					t.Fatalf("invalid next time for %q", tt.descriptor)
				}
			}
		})
	}
}

func TestCronEveryDuration(t *testing.T) {
	tests := []struct {
		name     string
		expr     string
		wantErr  bool
		interval time.Duration
	}{
		{"5 minutes", "@every 5m", false, 5 * time.Minute},
		{"1 hour", "@every 1h", false, time.Hour},
		{"30 seconds", "@every 30s", false, 30 * time.Second},
		{"complex", "@every 1h30m", false, 90 * time.Minute},
		{"invalid", "@every invalid", true, 0},
		{"zero", "@every 0s", true, 0},
		{"negative", "@every -5m", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := ParseCronSpec(tt.expr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseCronSpec(%q) error = %v, wantErr %v", tt.expr, err, tt.wantErr)
			}
			if err == nil {
				now := time.Now()
				next := spec.Next(now)
				expectedNext := now.Add(tt.interval)
				diff := next.Sub(expectedNext)
				if diff < 0 {
					diff = -diff
				}
				if diff > time.Second {
					t.Fatalf("@every %v: expected next ~%v, got %v", tt.interval, expectedNext, next)
				}
			}
		})
	}
}

func TestCronWithTimezone(t *testing.T) {
	// Create a timezone 5 hours ahead of UTC
	ahead := time.FixedZone("TEST+5", 5*3600)

	// Parse cron spec for 9 AM in the ahead timezone
	spec, err := ParseCronSpecWithLocation("0 9 * * *", ahead)
	if err != nil {
		t.Fatalf("ParseCronSpecWithLocation failed: %v", err)
	}

	// Start at 8 AM in the ahead timezone
	now := time.Date(2025, 1, 1, 8, 0, 0, 0, ahead)
	next := spec.Next(now)

	// Should be 9 AM in the ahead timezone
	if next.Hour() != 9 || next.Minute() != 0 {
		t.Fatalf("expected 9:00, got %02d:%02d", next.Hour(), next.Minute())
	}

	// Verify it's in the correct timezone
	if next.Location().String() != ahead.String() {
		t.Fatalf("expected timezone %v, got %v", ahead, next.Location())
	}
}

func TestAddCronWithDescriptors(t *testing.T) {
	sched := New(WithWorkers(1))
	sched.Start()
	defer func() { _ = sched.Stop(context.Background()) }()

	var executed atomic.Int32

	// Test @every descriptor
	_, err := sched.AddCron("every-test", "@every 50ms", func(ctx context.Context) error {
		executed.Add(1)
		return nil
	})
	if err != nil {
		t.Fatalf("AddCron with @every failed: %v", err)
	}

	// Test @hourly descriptor
	_, err = sched.AddCron("hourly-test", "@hourly", func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("AddCron with @hourly failed: %v", err)
	}

	// Wait for @every job to run
	time.Sleep(150 * time.Millisecond)

	if executed.Load() < 2 {
		t.Fatalf("expected at least 2 executions, got %d", executed.Load())
	}
}

func TestErrorHandlingStandardization(t *testing.T) {
	t.Run("ErrTaskNameEmpty", func(t *testing.T) {
		sched := New()
		err := sched.RegisterTask("", func(ctx context.Context) error { return nil })
		if !errors.Is(err, ErrTaskNameEmpty) {
			t.Fatalf("expected ErrTaskNameEmpty, got %v", err)
		}
	})

	t.Run("ErrTaskNil", func(t *testing.T) {
		sched := New()
		err := sched.RegisterTask("test", nil)
		if !errors.Is(err, ErrTaskNil) {
			t.Fatalf("expected ErrTaskNil, got %v", err)
		}
	})

	t.Run("ErrTaskNilInSchedule", func(t *testing.T) {
		sched := New()
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		_, err := sched.Schedule("test", time.Now().Add(time.Hour), nil)
		if !errors.Is(err, ErrTaskNil) {
			t.Fatalf("expected ErrTaskNil, got %v", err)
		}
	})

	t.Run("ErrRunAtRequired", func(t *testing.T) {
		sched := New()
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		task := func(ctx context.Context) error { return nil }
		_, err := sched.Schedule("test", time.Time{}, task)
		if !errors.Is(err, ErrRunAtRequired) {
			t.Fatalf("expected ErrRunAtRequired, got %v", err)
		}
	})

	t.Run("ErrJobExists", func(t *testing.T) {
		sched := New()
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		task := func(ctx context.Context) error { return nil }
		_, err := sched.AddCron("test", "* * * * *", task)
		if err != nil {
			t.Fatalf("first AddCron failed: %v", err)
		}

		_, err = sched.AddCron("test", "* * * * *", task)
		if !errors.Is(err, ErrJobExists) {
			t.Fatalf("expected ErrJobExists, got %v", err)
		}
	})

	t.Run("ErrSchedulerClosed", func(t *testing.T) {
		sched := New()
		sched.Start()
		_ = sched.Stop(context.Background())

		task := func(ctx context.Context) error { return nil }
		_, err := sched.AddCron("test", "* * * * *", task)
		if !errors.Is(err, ErrSchedulerClosed) {
			t.Fatalf("expected ErrSchedulerClosed, got %v", err)
		}
	})

	t.Run("ErrSelfDependency", func(t *testing.T) {
		sched := New()
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		task := func(ctx context.Context) error { return nil }
		_, err := sched.Delay("job1", time.Second, task,
			WithDependsOn(DependencyFailureSkip, "job1"),
		)
		if !errors.Is(err, ErrSelfDependency) {
			t.Fatalf("expected ErrSelfDependency, got %v", err)
		}
	})

	t.Run("ErrDependencyNotFound", func(t *testing.T) {
		sched := New()
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		task := func(ctx context.Context) error { return nil }
		_, err := sched.Delay("job1", time.Second, task,
			WithDependsOn(DependencyFailureSkip, "nonexistent"),
		)
		if !errors.Is(err, ErrDependencyNotFound) {
			t.Fatalf("expected ErrDependencyNotFound, got %v", err)
		}
	})

	t.Run("ErrInvalidCronExpr", func(t *testing.T) {
		tests := []struct {
			name string
			expr string
		}{
			{"too few fields", "* * *"},
			{"too many fields", "* * * * * * *"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseCronSpec(tt.expr)
				if !errors.Is(err, ErrInvalidCronExpr) {
					t.Fatalf("expected ErrInvalidCronExpr for %q, got %v", tt.expr, err)
				}
			})
		}
	})

	t.Run("ErrInvalidCronField", func(t *testing.T) {
		tests := []struct {
			name string
			expr string
		}{
			{"invalid minute", "99 * * * *"},
			{"invalid hour", "* 25 * * *"},
			{"invalid day", "* * 32 * *"},
			{"invalid month", "* * * 13 *"},
			{"invalid dow", "* * * * 8"},
			{"invalid range", "5-2 * * * *"},
			{"invalid step", "*/0 * * * *"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseCronSpec(tt.expr)
				if !errors.Is(err, ErrInvalidCronField) {
					t.Fatalf("expected ErrInvalidCronField for %q, got %v", tt.expr, err)
				}
			})
		}
	})

	t.Run("ErrInvalidEveryDuration", func(t *testing.T) {
		tests := []struct {
			name string
			expr string
		}{
			{"invalid format", "@every invalid"},
			{"zero duration", "@every 0s"},
			{"negative duration", "@every -5m"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseCronSpec(tt.expr)
				if !errors.Is(err, ErrInvalidEveryDuration) {
					t.Fatalf("expected ErrInvalidEveryDuration for %q, got %v", tt.expr, err)
				}
			})
		}
	})
}

func TestDeadLetterQueue(t *testing.T) {
	t.Run("DLQ stores failed jobs", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// Add a job that will fail
		_, err := sched.Delay("fail-job", 10*time.Millisecond, func(ctx context.Context) error {
			return fmt.Errorf("intentional failure")
		}, WithRetryPolicy(RetryPolicy{MaxAttempts: 2}))
		if err != nil {
			t.Fatalf("Delay failed: %v", err)
		}

		// Wait for job to fail and be added to DLQ
		time.Sleep(100 * time.Millisecond)

		// Check DLQ size
		size := sched.DeadLetterQueueSize()
		if size != 1 {
			t.Fatalf("expected DLQ size 1, got %d", size)
		}

		// Get the dead letter entry
		entry, exists := sched.GetDeadLetter("fail-job")
		if !exists {
			t.Fatal("expected dead letter entry to exist")
		}
		if entry.JobID != "fail-job" {
			t.Fatalf("expected job ID 'fail-job', got %s", entry.JobID)
		}
		if entry.Error == nil || entry.Error.Error() != "intentional failure" {
			t.Fatalf("expected error 'intentional failure', got %v", entry.Error)
		}
		if entry.Attempts != 2 {
			t.Fatalf("expected 2 attempts, got %d", entry.Attempts)
		}
	})

	t.Run("List all DLQ entries", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// Add multiple failing jobs
		for i := 0; i < 3; i++ {
			jobID := fmt.Sprintf("fail-job-%d", i)
			_, err := sched.Delay(JobID(jobID), 10*time.Millisecond, func(ctx context.Context) error {
				return fmt.Errorf("failure %d", i)
			}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
			if err != nil {
				t.Fatalf("Delay failed: %v", err)
			}
		}

		// Wait for all jobs to fail
		time.Sleep(150 * time.Millisecond)

		// List all entries
		entries := sched.ListDeadLetters()
		if len(entries) != 3 {
			t.Fatalf("expected 3 DLQ entries, got %d", len(entries))
		}
	})

	t.Run("Delete DLQ entry", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// Add a failing job
		_, err := sched.Delay("fail-job", 10*time.Millisecond, func(ctx context.Context) error {
			return fmt.Errorf("intentional failure")
		}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
		if err != nil {
			t.Fatalf("Delay failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Verify entry exists
		if sched.DeadLetterQueueSize() != 1 {
			t.Fatal("expected DLQ size 1")
		}

		// Delete the entry
		deleted := sched.DeleteDeadLetter("fail-job")
		if !deleted {
			t.Fatal("expected delete to return true")
		}

		// Verify entry is gone
		if sched.DeadLetterQueueSize() != 0 {
			t.Fatal("expected DLQ size 0 after delete")
		}
	})

	t.Run("Clear all DLQ entries", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// Add multiple failing jobs
		for i := 0; i < 5; i++ {
			jobID := fmt.Sprintf("fail-job-%d", i)
			_, err := sched.Delay(JobID(jobID), 10*time.Millisecond, func(ctx context.Context) error {
				return fmt.Errorf("failure")
			}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
			if err != nil {
				t.Fatalf("Delay failed: %v", err)
			}
		}

		time.Sleep(150 * time.Millisecond)

		// Verify entries exist
		if sched.DeadLetterQueueSize() != 5 {
			t.Fatalf("expected DLQ size 5, got %d", sched.DeadLetterQueueSize())
		}

		// Clear all entries
		cleared := sched.ClearDeadLetters()
		if cleared != 5 {
			t.Fatalf("expected to clear 5 entries, got %d", cleared)
		}

		// Verify all entries are gone
		if sched.DeadLetterQueueSize() != 0 {
			t.Fatal("expected DLQ size 0 after clear")
		}
	})

	t.Run("Requeue from DLQ", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		var executed atomic.Int32

		// Add a failing job
		_, err := sched.Delay("fail-job", 10*time.Millisecond, func(ctx context.Context) error {
			return fmt.Errorf("intentional failure")
		}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
		if err != nil {
			t.Fatalf("Delay failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Verify entry in DLQ
		if sched.DeadLetterQueueSize() != 1 {
			t.Fatal("expected DLQ size 1")
		}

		// Requeue with a working task
		_, err = sched.RequeueDeadLetter("fail-job", func(ctx context.Context) error {
			executed.Add(1)
			return nil
		})
		if err != nil {
			t.Fatalf("RequeueDeadLetter failed: %v", err)
		}

		// Wait for requeued job to execute
		time.Sleep(100 * time.Millisecond)

		// Verify job executed
		if executed.Load() != 1 {
			t.Fatalf("expected 1 execution, got %d", executed.Load())
		}

		// Verify entry removed from DLQ
		if sched.DeadLetterQueueSize() != 0 {
			t.Fatalf("expected DLQ size 0 after requeue, got %d", sched.DeadLetterQueueSize())
		}
	})

	t.Run("DLQ max size eviction", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(3), // Max 3 entries
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// Add 5 failing jobs (should keep only 3)
		for i := 0; i < 5; i++ {
			jobID := fmt.Sprintf("fail-job-%d", i)
			_, err := sched.Delay(JobID(jobID), time.Duration(i*10)*time.Millisecond, func(ctx context.Context) error {
				return fmt.Errorf("failure")
			}, WithRetryPolicy(RetryPolicy{MaxAttempts: 1}))
			if err != nil {
				t.Fatalf("Delay failed: %v", err)
			}
		}

		time.Sleep(200 * time.Millisecond)

		// Verify max size is enforced
		size := sched.DeadLetterQueueSize()
		if size != 3 {
			t.Fatalf("expected DLQ size 3 (max), got %d", size)
		}
	})

	t.Run("DLQ disabled returns nil/false", func(t *testing.T) {
		sched := New(WithWorkers(1))
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		// All DLQ operations should return nil/false/0
		if sched.DeadLetterQueueSize() != 0 {
			t.Fatal("expected size 0 when DLQ disabled")
		}
		if sched.ListDeadLetters() != nil {
			t.Fatal("expected nil when DLQ disabled")
		}
		if _, ok := sched.GetDeadLetter("test"); ok {
			t.Fatal("expected false when DLQ disabled")
		}
		if sched.DeleteDeadLetter("test") {
			t.Fatal("expected false when DLQ disabled")
		}
		if sched.ClearDeadLetters() != 0 {
			t.Fatal("expected 0 when DLQ disabled")
		}
	})

	t.Run("Legacy callback still works with DLQ", func(t *testing.T) {
		sched := New(
			WithWorkers(1),
			WithDeadLetterQueue(10),
		)
		sched.Start()
		defer func() { _ = sched.Stop(context.Background()) }()

		var callbackCalled atomic.Bool

		// Add a failing job with legacy callback
		_, err := sched.Delay("fail-job", 10*time.Millisecond, func(ctx context.Context) error {
			return fmt.Errorf("intentional failure")
		},
			WithRetryPolicy(RetryPolicy{MaxAttempts: 1}),
			WithDeadLetter(func(ctx context.Context, id JobID, err error) {
				callbackCalled.Store(true)
			}),
		)
		if err != nil {
			t.Fatalf("Delay failed: %v", err)
		}

		time.Sleep(100 * time.Millisecond)

		// Verify both DLQ and callback were triggered
		if sched.DeadLetterQueueSize() != 1 {
			t.Fatal("expected DLQ size 1")
		}
		if !callbackCalled.Load() {
			t.Fatal("expected legacy callback to be called")
		}
	})
}
