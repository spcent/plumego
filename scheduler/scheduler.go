package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/spcent/plumego/log"
)

var (
	// ErrSchedulerClosed indicates scheduler has been stopped.
	ErrSchedulerClosed = errors.New("scheduler: closed")
	// ErrJobExists indicates a job ID is already registered.
	ErrJobExists = errors.New("scheduler: job already exists")
)

// Scheduler coordinates cron, delayed, and retryable jobs.
type Scheduler struct {
	mu     sync.Mutex
	jobs   map[JobID]*job
	queue  scheduleHeap
	wakeCh chan struct{}
	stopCh chan struct{}
	closed bool

	workerCount  int
	workCh       chan *runRequest
	logger       log.StructuredLogger
	panicHandler func(context.Context, JobID, any)
	clock        Clock
	store        Store
	registry     map[string]TaskFunc
	metricsSink  MetricsSink

	wg sync.WaitGroup

	stats schedulerStats
}

// Option configures the scheduler.
type Option func(*Scheduler)

// WithWorkers sets the worker pool size.
func WithWorkers(n int) Option {
	return func(s *Scheduler) {
		if n > 0 {
			s.workerCount = n
		}
	}
}

// WithQueueSize sets the internal work queue size.
func WithQueueSize(n int) Option {
	return func(s *Scheduler) {
		if n > 0 {
			s.workCh = make(chan *runRequest, n)
		}
	}
}

// WithLogger sets the structured logger.
func WithLogger(logger log.StructuredLogger) Option {
	return func(s *Scheduler) {
		s.logger = logger
	}
}

// WithPanicHandler registers a panic handler for task execution.
func WithPanicHandler(handler func(context.Context, JobID, any)) Option {
	return func(s *Scheduler) {
		s.panicHandler = handler
	}
}

// WithMetricsSink registers a metrics sink for scheduler events.
func WithMetricsSink(sink MetricsSink) Option {
	return func(s *Scheduler) {
		s.metricsSink = sink
	}
}

// WithClock injects a custom clock for testing.
func WithClock(clock Clock) Option {
	return func(s *Scheduler) {
		if clock != nil {
			s.clock = clock
		}
	}
}

// WithStore sets a persistence store for delay jobs.
func WithStore(store Store) Option {
	return func(s *Scheduler) {
		s.store = store
	}
}

// New constructs a Scheduler.
func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		jobs:        make(map[JobID]*job),
		queue:       scheduleHeap{},
		wakeCh:      make(chan struct{}, 1),
		stopCh:      make(chan struct{}),
		workerCount: 4,
		workCh:      make(chan *runRequest, 256),
		logger:      log.NewGLogger(),
		clock:       realClock{},
		registry:    make(map[string]TaskFunc),
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// Start launches scheduler goroutines.
func (s *Scheduler) Start() {
	s.loadPersisted()
	s.wg.Add(1)
	go s.runLoop()
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker()
	}
}

// RegisterTask registers a named task for persistence recovery.
func (s *Scheduler) RegisterTask(name string, task TaskFunc) {
	if name == "" || task == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry[name] = task
}

// Stop stops the scheduler and waits for shutdown.
func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.stopCh)
	s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// AddCron registers a cron job.
func (s *Scheduler) AddCron(id JobID, spec string, task TaskFunc, opts ...JobOption) (JobID, error) {
	parsed, err := ParseCronSpec(spec)
	if err != nil {
		return "", err
	}
	return s.addJob(id, jobKindCron, task, parsed, time.Time{}, opts...)
}

// Schedule registers a job to run at a specific time.
func (s *Scheduler) Schedule(id JobID, runAt time.Time, task TaskFunc, opts ...JobOption) (JobID, error) {
	if runAt.IsZero() {
		return "", fmt.Errorf("runAt is required")
	}
	return s.addJob(id, jobKindDelay, task, CronSpec{}, runAt, opts...)
}

// Delay registers a job to run after the specified delay.
func (s *Scheduler) Delay(id JobID, delay time.Duration, task TaskFunc, opts ...JobOption) (JobID, error) {
	return s.Schedule(id, s.clock.Now().Add(delay), task, opts...)
}

// Cancel removes a job from the scheduler.
func (s *Scheduler) Cancel(id JobID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[id]
	if !ok {
		return false
	}
	job.canceled = true
	delete(s.jobs, id)
	if s.store != nil {
		_ = s.store.Delete(id)
	}
	return true
}

// Pause marks a job as paused.
func (s *Scheduler) Pause(id JobID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return false
	}
	j.paused = true
	return true
}

// Resume unpauses a job and schedules its next run.
func (s *Scheduler) Resume(id JobID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok || j.canceled {
		return false
	}
	if !j.paused {
		return true
	}
	j.paused = false
	if j.kind == jobKindCron {
		j.runAt = j.cron.Next(time.Now())
		s.queue.PushSchedule(&scheduleItem{runAt: j.runAt, job: j})
	} else if j.kind == jobKindDelay {
		s.queue.PushSchedule(&scheduleItem{runAt: j.runAt, job: j})
	}
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
	return true
}

// Status returns the job status snapshot.
func (s *Scheduler) Status(id JobID) (JobStatus, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok {
		return JobStatus{}, false
	}
	return jobStatusFrom(j), true
}

// List returns status for all jobs.
func (s *Scheduler) List() []JobStatus {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]JobStatus, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, jobStatusFrom(j))
	}
	return out
}

func (s *Scheduler) addJob(id JobID, kind jobKind, task TaskFunc, spec CronSpec, runAt time.Time, opts ...JobOption) (JobID, error) {
	if task == nil {
		return "", fmt.Errorf("task cannot be nil")
	}
	if id == "" {
		id = JobID(fmt.Sprintf("job-%d", time.Now().UnixNano()))
	}
	jobOpts := defaultJobOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&jobOpts)
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return "", ErrSchedulerClosed
	}
	if existing, exists := s.jobs[id]; exists {
		if !jobOpts.Replace {
			return "", ErrJobExists
		}
		existing.canceled = true
		delete(s.jobs, id)
	}

	j := &job{
		id:          id,
		kind:        kind,
		cron:        spec,
		runAt:       runAt,
		fn:          task,
		options:     jobOpts,
		nextAttempt: 1,
	}
	if kind == jobKindCron {
		j.runAt = spec.Next(s.clock.Now())
		j.cronExpr = specString(spec)
	}
	if kind == jobKindDelay && runAt.Before(s.clock.Now()) {
		j.runAt = s.clock.Now()
	}
	j.paused = jobOpts.Paused

	s.jobs[id] = j
	if !j.paused {
		s.pushScheduleLocked(j, j.runAt)
	}
	s.persistJobLocked(j)
	return id, nil
}

func (s *Scheduler) pushScheduleLocked(j *job, runAt time.Time) {
	s.queue.PushSchedule(&scheduleItem{runAt: runAt, job: j})
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Scheduler) runLoop() {
	defer s.wg.Done()
	for {
		item := s.nextDue()
		if item == nil {
			select {
			case <-s.stopCh:
				return
			case <-s.wakeCh:
				continue
			}
		}

		now := s.clock.Now()
		wait := time.Until(item.runAt)
		if wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-s.wakeCh:
				timer.Stop()
				continue
			case <-s.stopCh:
				timer.Stop()
				return
			}
		}

		// Dispatch due items
		for {
			due := s.popDue(now)
			if due == nil {
				break
			}
			s.dispatch(due.job)
		}
	}
}

func (s *Scheduler) nextDue() *scheduleItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	return s.queue.Peek()
}

func (s *Scheduler) popDue(now time.Time) *scheduleItem {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	return s.queue.PopDue(now)
}

func (s *Scheduler) dispatch(j *job) {
	if j == nil || j.canceled || j.paused {
		return
	}

	if j.options.OverlapPolicy == SkipIfRunning && !j.running.CompareAndSwap(false, true) {
		s.scheduleNext(j, time.Now())
		return
	}
	if j.options.OverlapPolicy == SerialQueue {
		if !j.running.CompareAndSwap(false, true) {
			// Use atomic operation to set pending flag to avoid race condition
			j.pending.Store(true)
			return
		}
	}

	req := &runRequest{job: j, enqueuedAt: time.Now()}
	select {
	case s.workCh <- req:
		s.stats.incQueued()
	case <-s.stopCh:
	default:
		s.stats.incDropped()
		if s.metricsSink != nil {
			s.metricsSink.ObserveDrop(j.id)
		}
		if s.logger != nil {
			s.logger.Warn("scheduler queue full", log.Fields{"job_id": j.id})
		}
	}
}

func (s *Scheduler) scheduleNext(j *job, base time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || j.canceled {
		return
	}
	var next time.Time
	switch j.kind {
	case jobKindCron:
		next = j.cron.Next(base)
	case jobKindDelay:
		// delay jobs run only once unless retried
		return
	}
	if next.IsZero() {
		return
	}
	j.runAt = next
	s.queue.PushSchedule(&scheduleItem{runAt: next, job: j})
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Scheduler) worker() {
	defer s.wg.Done()
	for {
		select {
		case <-s.stopCh:
			return
		case req := <-s.workCh:
			if req == nil || req.job == nil {
				continue
			}
			s.execute(req.job, req.enqueuedAt)
		}
	}
}

func (s *Scheduler) execute(j *job, enqueuedAt time.Time) {
	ctx := context.Background()
	cancel := func() {}
	if j.options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, j.options.Timeout)
	}
	queueDelay := time.Since(enqueuedAt)
	s.stats.addQueueDelay(queueDelay)
	start := time.Now()
	s.mu.Lock()
	attempt := j.nextAttempt
	scheduledAt := j.runAt
	lastErr := j.lastError
	s.mu.Unlock()
	ctx = context.WithValue(ctx, jobIDKey{}, j.id)
	ctx = context.WithValue(ctx, jobAttemptKey{}, attempt)
	ctx = context.WithValue(ctx, jobScheduledKey{}, scheduledAt)
	if lastErr != nil {
		ctx = context.WithValue(ctx, jobLastErrorKey{}, lastErr)
	}
	err := func() (err error) {
		defer func() {
			if recovered := recover(); recovered != nil {
				if s.panicHandler != nil {
					s.panicHandler(ctx, j.id, recovered)
				}
				err = fmt.Errorf("panic: %v", recovered)
			}
		}()
		return j.fn(ctx)
	}()
	cancel()

	if j.options.OverlapPolicy == SkipIfRunning || j.options.OverlapPolicy == SerialQueue {
		j.running.Store(false)
	}

	s.mu.Lock()
	j.lastRun = start
	j.lastError = err
	s.mu.Unlock()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			s.stats.incTimeout()
		}
		s.stats.incFailure()
		if s.metricsSink != nil {
			s.metricsSink.ObserveRun(j.id, time.Since(start), err, queueDelay)
		}
		s.handleFailure(j, err)
		return
	}

	s.stats.incSuccess()
	if s.metricsSink != nil {
		s.metricsSink.ObserveRun(j.id, time.Since(start), nil, queueDelay)
	}
	s.mu.Lock()
	j.nextAttempt = 1
	// Use atomic operation to read and clear pending flag
	pending := j.options.OverlapPolicy == SerialQueue && j.pending.Swap(false)
	if pending {
		s.queue.PushSchedule(&scheduleItem{runAt: time.Now(), job: j})
		select {
		case s.wakeCh <- struct{}{}:
		default:
		}
	}
	s.mu.Unlock()

	if j.kind == jobKindCron && !pending {
		s.scheduleNext(j, start)
	}
	if j.kind == jobKindDelay && s.store != nil {
		_ = s.store.Delete(j.id)
	}
}

func (s *Scheduler) handleFailure(j *job, err error) {
	s.mu.Lock()
	policy := j.options.RetryPolicy
	dlq := j.options.DeadLetter
	attempt := j.nextAttempt
	if policy.MaxAttempts > 1 && attempt < policy.MaxAttempts {
		j.nextAttempt++
	}
	s.mu.Unlock()

	if policy.MaxAttempts <= 1 {
		if dlq != nil {
			dlq(context.Background(), j.id, err)
		}
		if j.kind == jobKindCron {
			s.scheduleNext(j, s.clock.Now())
		}
		if j.kind == jobKindDelay && s.store != nil {
			_ = s.store.Delete(j.id)
		}
		return
	}

	if attempt >= policy.MaxAttempts {
		if dlq != nil {
			dlq(context.Background(), j.id, err)
		}
		if j.kind == jobKindCron {
			s.scheduleNext(j, s.clock.Now())
		}
		if j.kind == jobKindDelay && s.store != nil {
			_ = s.store.Delete(j.id)
		}
		return
	}

	backoff := time.Duration(0)
	if policy.Backoff != nil {
		backoff = policy.Backoff(attempt + 1)
		if policy.MaxBackoff > 0 && backoff > policy.MaxBackoff {
			backoff = policy.MaxBackoff
		}
	} else if policy.Kind != "" {
		switch policy.Kind {
		case "fixed":
			backoff = policy.BaseDelay
		case "exponential":
			if policy.BaseDelay > 0 {
				backoff = policy.BaseDelay << attempt
			}
		}
		if policy.MaxBackoff > 0 && backoff > policy.MaxBackoff {
			backoff = policy.MaxBackoff
		}
	}
	s.mu.Lock()
	s.queue.PushSchedule(&scheduleItem{runAt: s.clock.Now().Add(backoff), job: j})
	s.mu.Unlock()
	if s.metricsSink != nil {
		s.metricsSink.ObserveRetry(j.id, attempt+1, backoff)
	}

	select {
	case s.wakeCh <- struct{}{}:
	default:
	}

	if s.logger != nil {
		s.logger.Warn("scheduler job failed", log.Fields{
			"job_id":  j.id,
			"error":   err,
			"attempt": attempt,
		})
	}
	s.stats.incRetry()
}

func (s *Scheduler) logError(msg string, err error) {
	if s.logger == nil || err == nil {
		return
	}
	s.logger.Error(msg, log.Fields{"error": err})
}

func specString(spec CronSpec) string {
	_ = spec
	return "cron"
}

func jobStatusFrom(j *job) JobStatus {
	kind := "delay"
	if j.kind == jobKindCron {
		kind = "cron"
	}
	return JobStatus{
		ID:            j.id,
		NextRun:       j.runAt,
		LastRun:       j.lastRun,
		LastError:     j.lastError,
		Attempt:       j.nextAttempt,
		Paused:        j.paused,
		Running:       j.running.Load(),
		Kind:          kind,
		OverlapPolicy: j.options.OverlapPolicy,
		Group:         j.options.Group,
		Tags:          append([]string(nil), j.options.Tags...),
	}
}

func (s *Scheduler) persistJobLocked(j *job) {
	if s.store == nil || j == nil {
		return
	}
	if j.kind != jobKindDelay {
		return
	}
	if j.options.TaskName == "" {
		return
	}
	job := StoredJob{
		ID:        j.id,
		Kind:      "delay",
		RunAt:     j.runAt,
		TaskName:  j.options.TaskName,
		Group:     j.options.Group,
		Tags:      append([]string(nil), j.options.Tags...),
		Timeout:   j.options.Timeout,
		Overlap:   j.options.OverlapPolicy,
		Retry:     serializeRetry(j.options.RetryPolicy),
		CreatedAt: s.clock.Now(),
	}
	_ = s.store.Save(job)
}

func (s *Scheduler) loadPersisted() {
	if s.store == nil {
		return
	}
	stored, err := s.store.List()
	if err != nil {
		s.logError("scheduler load persisted jobs failed", err)
		return
	}
	for _, item := range stored {
		if item.TaskName == "" {
			continue
		}
		task := s.lookupTask(item.TaskName)
		if task == nil {
			continue
		}
		retry := hydrateRetry(item.Retry)
		opts := []JobOption{
			WithTimeout(item.Timeout),
			WithOverlapPolicy(item.Overlap),
			WithRetryPolicy(retry),
			WithGroup(item.Group),
			WithTags(item.Tags...),
			WithTaskName(item.TaskName),
			ReplaceExisting(),
		}
		_, _ = s.Schedule(item.ID, item.RunAt, task, opts...)
	}
}

func (s *Scheduler) lookupTask(name string) TaskFunc {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.registry[name]
}
