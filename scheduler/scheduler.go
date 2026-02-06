package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	log "github.com/spcent/plumego/log"
)

const (
	// DefaultWorkChannelBuffer is the default buffer size for the work channel.
	DefaultWorkChannelBuffer = 256
)

// Error definitions moved to errors.go

// Scheduler coordinates cron, delayed, and retryable jobs.
type Scheduler struct {
	mu     sync.RWMutex
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
	backpressure BackpressureConfig

	// dependents tracks which jobs depend on each job (reverse mapping)
	// Key: dependency JobID, Value: slice of dependent JobIDs
	dependents map[JobID][]JobID

	// dependencyStatus tracks completion status of dependencies
	// Key: JobID, Value: true if completed successfully, false if failed
	dependencyStatus map[JobID]bool

	// dlq manages failed jobs for inspection and requeuing
	dlq *DeadLetterQueue

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

// WithBackpressure configures the backpressure behavior when the work queue is full.
func WithBackpressure(config BackpressureConfig) Option {
	return func(s *Scheduler) {
		s.backpressure = config
	}
}

// WithDeadLetterQueue enables dead letter queue for failed jobs.
// maxSize limits the number of entries (0 = unlimited).
func WithDeadLetterQueue(maxSize int) Option {
	return func(s *Scheduler) {
		s.dlq = NewDeadLetterQueue(maxSize)
	}
}

// New constructs a Scheduler.
func New(opts ...Option) *Scheduler {
	s := &Scheduler{
		jobs:             make(map[JobID]*job),
		queue:            scheduleHeap{},
		wakeCh:           make(chan struct{}, 1),
		stopCh:           make(chan struct{}),
		workerCount:      4,
		workCh:           make(chan *runRequest, DefaultWorkChannelBuffer),
		logger:           log.NewGLogger(),
		clock:            realClock{},
		registry:         make(map[string]TaskFunc),
		backpressure:     DefaultBackpressureConfig(),
		dependents:       make(map[JobID][]JobID),
		dependencyStatus: make(map[JobID]bool),
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
// Returns an error if the name is empty or the task is nil.
func (s *Scheduler) RegisterTask(name string, task TaskFunc) error {
	if name == "" {
		return ErrTaskNameEmpty
	}
	if task == nil {
		return ErrTaskNil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.registry[name] = task
	return nil
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
// Supports standard cron syntax, @descriptors (@hourly, @daily, etc.), and @every duration.
// Examples:
//   - "0 * * * *"  - every hour
//   - "@daily"     - every day at midnight
//   - "@every 5m"  - every 5 minutes
func (s *Scheduler) AddCron(id JobID, spec string, task TaskFunc, opts ...JobOption) (JobID, error) {
	parsed, err := ParseCronSpec(spec)
	if err != nil {
		return "", err
	}
	return s.addJob(id, jobKindCron, task, parsed, time.Time{}, opts...)
}

// AddCronWithLocation registers a cron job with a specific timezone.
// This allows scheduling jobs in timezones other than UTC.
// Example: AddCronWithLocation("daily-report", "0 9 * * *", task, time.FixedZone("EST", -5*3600))
func (s *Scheduler) AddCronWithLocation(id JobID, spec string, task TaskFunc, location *time.Location, opts ...JobOption) (JobID, error) {
	if location == nil {
		location = time.UTC
	}
	parsed, err := ParseCronSpecWithLocation(spec, location)
	if err != nil {
		return "", err
	}
	return s.addJob(id, jobKindCron, task, parsed, time.Time{}, opts...)
}

// Schedule registers a job to run at a specific time.
func (s *Scheduler) Schedule(id JobID, runAt time.Time, task TaskFunc, opts ...JobOption) (JobID, error) {
	if runAt.IsZero() {
		return "", ErrRunAtRequired
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
	j, ok := s.jobs[id]
	if !ok {
		return false
	}
	s.cancelJobLocked(id, j)
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
	j.paused.Store(true)
	return true
}

// Resume unpauses a job and schedules its next run.
func (s *Scheduler) Resume(id JobID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	j, ok := s.jobs[id]
	if !ok || j.canceled.Load() {
		return false
	}
	if !j.paused.Load() {
		return true
	}
	s.resumeJobLocked(j)
	return true
}

// Status returns the job status snapshot.
func (s *Scheduler) Status(id JobID) (JobStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[id]
	if !ok {
		return JobStatus{}, false
	}
	return jobStatusFrom(j), true
}

// List returns status for all jobs.
func (s *Scheduler) List() []JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]JobStatus, 0, len(s.jobs))
	for _, j := range s.jobs {
		out = append(out, jobStatusFrom(j))
	}
	return out
}

// PauseByGroup pauses all jobs in the specified group.
// Returns the number of jobs paused.
func (s *Scheduler) PauseByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, j := range s.jobs {
		if j.options.Group == group && !j.paused.Load() {
			j.paused.Store(true)
			count++
		}
	}
	return count
}

// PauseByTags pauses all jobs that have ALL of the specified tags.
// Returns the number of jobs paused.
func (s *Scheduler) PauseByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, j := range s.jobs {
		if hasAllTags(j.options.Tags, tags) && !j.paused.Load() {
			j.paused.Store(true)
			count++
		}
	}
	return count
}

// ResumeByGroup resumes all paused jobs in the specified group.
// Returns the number of jobs resumed.
func (s *Scheduler) ResumeByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, j := range s.jobs {
		if j.options.Group == group && j.paused.Load() && !j.canceled.Load() {
			s.resumeJobLocked(j)
			count++
		}
	}
	return count
}

// ResumeByTags resumes all paused jobs that have ALL of the specified tags.
// Returns the number of jobs resumed.
func (s *Scheduler) ResumeByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for _, j := range s.jobs {
		if hasAllTags(j.options.Tags, tags) && j.paused.Load() && !j.canceled.Load() {
			s.resumeJobLocked(j)
			count++
		}
	}
	return count
}

// CancelByGroup cancels all jobs in the specified group.
// Returns the number of jobs canceled.
func (s *Scheduler) CancelByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for id, j := range s.jobs {
		if j.options.Group == group {
			s.cancelJobLocked(id, j)
			count++
		}
	}
	return count
}

// CancelByTags cancels all jobs that have ALL of the specified tags.
// Returns the number of jobs canceled.
func (s *Scheduler) CancelByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	count := 0
	for id, j := range s.jobs {
		if hasAllTags(j.options.Tags, tags) {
			s.cancelJobLocked(id, j)
			count++
		}
	}
	return count
}

// hasAllTags checks if jobTags contains all of the required tags.
func hasAllTags(jobTags, required []string) bool {
	if len(required) == 0 {
		return false
	}
	tagSet := make(map[string]bool, len(jobTags))
	for _, tag := range jobTags {
		tagSet[tag] = true
	}
	for _, req := range required {
		if !tagSet[req] {
			return false
		}
	}
	return true
}

// QueryJobs filters and sorts jobs based on the provided query criteria.
// Returns a JobQueryResult with matched jobs and total count.
func (s *Scheduler) QueryJobs(query JobQuery) JobQueryResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Collect and filter jobs
	var filtered []JobStatus
	for _, j := range s.jobs {
		status := jobStatusFrom(j)

		// Apply filters
		if query.Group != "" && status.Group != query.Group {
			continue
		}
		if len(query.Tags) > 0 && !hasAllTags(status.Tags, query.Tags) {
			continue
		}
		if len(query.Kinds) > 0 {
			matched := false
			for _, kind := range query.Kinds {
				if status.Kind == kind {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if query.Running != nil && status.Running != *query.Running {
			continue
		}
		if query.Paused != nil && status.Paused != *query.Paused {
			continue
		}
		if len(query.States) > 0 {
			matched := false
			for _, state := range query.States {
				if status.State == state {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		filtered = append(filtered, status)
	}

	total := len(filtered)

	// Sort results
	if query.OrderBy != "" {
		sortJobStatuses(filtered, query.OrderBy, query.Ascending)
	}

	// Apply pagination (clamp negative values to 0)
	offset := query.Offset
	if offset < 0 {
		offset = 0
	}
	if offset > 0 {
		if offset >= len(filtered) {
			filtered = nil
		} else {
			filtered = filtered[offset:]
		}
	}
	limit := query.Limit
	if limit < 0 {
		limit = 0
	}
	if limit > 0 && len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return JobQueryResult{
		Jobs:  filtered,
		Total: total,
	}
}

// sortJobStatuses sorts job statuses based on the specified field and direction.
func sortJobStatuses(jobs []JobStatus, orderBy string, ascending bool) {
	less := func(i, j int) bool {
		switch orderBy {
		case "id":
			if ascending {
				return jobs[i].ID < jobs[j].ID
			}
			return jobs[i].ID > jobs[j].ID
		case "next_run":
			if ascending {
				return jobs[i].NextRun.Before(jobs[j].NextRun)
			}
			return jobs[i].NextRun.After(jobs[j].NextRun)
		case "last_run":
			if ascending {
				return jobs[i].LastRun.Before(jobs[j].LastRun)
			}
			return jobs[i].LastRun.After(jobs[j].LastRun)
		case "group":
			if ascending {
				return jobs[i].Group < jobs[j].Group
			}
			return jobs[i].Group > jobs[j].Group
		default:
			return false
		}
	}

	// Simple bubble sort (good enough for scheduler use case)
	n := len(jobs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if less(j+1, j) {
				jobs[j], jobs[j+1] = jobs[j+1], jobs[j]
			}
		}
	}
}

func (s *Scheduler) addJob(id JobID, kind jobKind, task TaskFunc, spec CronSpec, runAt time.Time, opts ...JobOption) (JobID, error) {
	if task == nil {
		return "", ErrTaskNil
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
		if !jobOpts.Replace && !existing.canceled.Load() {
			return "", ErrJobExists
		}
		existing.canceled.Store(true)
		s.setJobStateLocked(existing, JobStateCanceled)
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
	j.paused.Store(jobOpts.Paused)
	s.setJobStateLocked(j, JobStateQueued)

	// Validate dependencies
	for _, depID := range jobOpts.Dependencies {
		if depID == id {
			return "", ErrSelfDependency
		}
		if _, exists := s.jobs[depID]; !exists {
			return "", fmt.Errorf("%w: %s", ErrDependencyNotFound, depID)
		}
	}

	s.jobs[id] = j

	// Build dependents map (reverse mapping)
	for _, depID := range jobOpts.Dependencies {
		s.dependents[depID] = append(s.dependents[depID], id)
	}

	if !j.paused.Load() && len(jobOpts.Dependencies) == 0 {
		// Only schedule if no dependencies or dependencies already met
		s.pushScheduleLocked(j, j.runAt)
	} else if len(jobOpts.Dependencies) > 0 {
		// Job has dependencies, will be triggered when dependencies complete
		if s.logger != nil {
			s.logger.Info("job has dependencies, waiting for completion", log.Fields{
				"job_id":       id,
				"dependencies": jobOpts.Dependencies,
			})
		}
	}

	s.persistJobLocked(j)
	return id, nil
}

func (s *Scheduler) pushScheduleLocked(j *job, runAt time.Time) {
	if j != nil && !j.canceled.Load() {
		s.setJobStateIfIdleLocked(j, JobStateScheduled)
	}
	s.queue.PushSchedule(&scheduleItem{runAt: runAt, job: j})
	s.wakeScheduler()
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
	s.mu.RLock()
	defer s.mu.RUnlock()
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
	if j == nil || j.canceled.Load() || j.paused.Load() {
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

	// Handle backpressure based on configured policy
	switch s.backpressure.Policy {
	case BackpressureDrop:
		// Default behavior: try to send, drop if full
		select {
		case s.workCh <- req:
			s.stats.incQueued()
		case <-s.stopCh:
		default:
			s.handleBackpressure(j)
		}

	case BackpressureBlock:
		// Block indefinitely until space is available
		select {
		case s.workCh <- req:
			s.stats.incQueued()
		case <-s.stopCh:
		}

	case BackpressureBlockTimeout:
		// Block with timeout
		if s.backpressure.Timeout <= 0 {
			// Fallback to drop if timeout is not configured
			select {
			case s.workCh <- req:
				s.stats.incQueued()
			case <-s.stopCh:
			default:
				s.handleBackpressure(j)
			}
			return
		}

		timer := time.NewTimer(s.backpressure.Timeout)
		defer timer.Stop()

		select {
		case s.workCh <- req:
			s.stats.incQueued()
		case <-timer.C:
			s.handleBackpressure(j)
		case <-s.stopCh:
		}
	}
}

func (s *Scheduler) handleBackpressure(j *job) {
	s.stats.incDropped()
	if s.metricsSink != nil {
		s.metricsSink.ObserveDrop(j.id)
	}
	if s.logger != nil {
		s.logger.Warn("scheduler queue full, job dropped", log.Fields{"job_id": j.id})
	}
	if s.backpressure.OnBackpressure != nil {
		s.backpressure.OnBackpressure(j.id)
	}
}

func (s *Scheduler) scheduleNext(j *job, base time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || j.canceled.Load() {
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
	s.setJobStateIfIdleLocked(j, JobStateScheduled)
	s.queue.PushSchedule(&scheduleItem{runAt: next, job: j})
	s.wakeScheduler()
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
	s.setJobStateLocked(j, JobStateRunning)
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
		// Mark dependency as failed
		s.mu.Lock()
		s.dependencyStatus[j.id] = false
		s.mu.Unlock()
		// Handle dependent jobs on failure
		s.handleDependentJobsOnFailure(j)
		s.handleFailure(j, err)
		return
	}

	s.stats.incSuccess()
	if s.metricsSink != nil {
		s.metricsSink.ObserveRun(j.id, time.Since(start), nil, queueDelay)
	}
	s.mu.Lock()
	j.nextAttempt = 1
	if j.kind == jobKindDelay {
		s.setJobStateLocked(j, JobStateCompleted)
	}
	// Mark dependency as succeeded
	s.dependencyStatus[j.id] = true
	// Use atomic operation to read and clear pending flag
	pending := j.options.OverlapPolicy == SerialQueue && j.pending.Swap(false)
	if pending {
		s.setJobStateIfIdleLocked(j, JobStateScheduled)
		s.queue.PushSchedule(&scheduleItem{runAt: time.Now(), job: j})
		s.wakeScheduler()
	}
	s.mu.Unlock()

	if j.kind == jobKindCron && !pending {
		s.scheduleNext(j, start)
	}
	if j.kind == jobKindDelay && s.store != nil {
		_ = s.store.Delete(j.id)
	}

	// Trigger dependent jobs on success
	s.triggerDependentJobs(j.id)
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

	retriesExhausted := policy.MaxAttempts <= 1 || attempt >= policy.MaxAttempts
	if retriesExhausted {
		finalAttempt := attempt
		if policy.MaxAttempts <= 1 {
			finalAttempt = 1
		}
		s.handleRetriesExhausted(j, err, finalAttempt, dlq)
		return
	}

	s.mu.Lock()
	s.setJobStateLocked(j, JobStateRetrying)
	s.mu.Unlock()

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
				backoff = safeExponentialBackoff(policy.BaseDelay, policy.MaxBackoff, attempt)
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

	s.wakeScheduler()

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

// handleRetriesExhausted handles the case when a job has no more retry attempts left.
func (s *Scheduler) handleRetriesExhausted(j *job, err error, attempts int, deadLetterCb func(context.Context, JobID, error)) {
	if j.kind == jobKindDelay {
		s.mu.Lock()
		s.setJobStateLocked(j, JobStateFailed)
		s.mu.Unlock()
	}
	if s.dlq != nil {
		s.addToDeadLetterQueue(j, err, attempts)
	}
	if deadLetterCb != nil {
		deadLetterCb(context.Background(), j.id, err)
	}
	if j.kind == jobKindCron {
		s.scheduleNext(j, s.clock.Now())
	}
	if j.kind == jobKindDelay && s.store != nil {
		_ = s.store.Delete(j.id)
	}
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
		Paused:        j.paused.Load(),
		Running:       j.running.Load(),
		Kind:          kind,
		OverlapPolicy: j.options.OverlapPolicy,
		Group:         j.options.Group,
		Tags:          append([]string(nil), j.options.Tags...),
		State:         j.state,
		StateUpdated:  j.stateAt,
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
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.registry[name]
}

// triggerDependentJobs checks and triggers jobs that depend on the completed job.
func (s *Scheduler) triggerDependentJobs(completedJobID JobID) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dependentIDs := s.dependents[completedJobID]
	if len(dependentIDs) == 0 {
		return
	}

	for _, depJobID := range dependentIDs {
		depJob, exists := s.jobs[depJobID]
		if !exists || depJob.canceled.Load() || depJob.paused.Load() {
			continue
		}

		if s.allDependenciesMetLocked(depJob) {
			s.setJobStateIfIdleLocked(depJob, JobStateScheduled)
			s.queue.PushSchedule(&scheduleItem{runAt: depJob.runAt, job: depJob})
			s.wakeScheduler()
			if s.logger != nil {
				s.logger.Info("triggering dependent job", log.Fields{
					"job_id":     depJobID,
					"dependency": completedJobID,
				})
			}
		}
	}
}

// handleDependentJobsOnFailure handles dependent jobs when a dependency fails.
func (s *Scheduler) handleDependentJobsOnFailure(failedJob *job) {
	s.mu.Lock()
	defer s.mu.Unlock()

	dependentIDs := s.dependents[failedJob.id]
	if len(dependentIDs) == 0 {
		return
	}

	for _, depJobID := range dependentIDs {
		depJob, exists := s.jobs[depJobID]
		if !exists || depJob.canceled.Load() {
			continue
		}

		switch depJob.options.DependencyPolicy {
		case DependencyFailureSkip:
			if s.logger != nil {
				s.logger.Warn("skipping dependent job due to dependency failure", log.Fields{
					"job_id":            depJobID,
					"failed_dependency": failedJob.id,
				})
			}
			if depJob.kind == jobKindCron {
				next := depJob.cron.Next(s.clock.Now())
				depJob.runAt = next
				s.setJobStateIfIdleLocked(depJob, JobStateScheduled)
				s.queue.PushSchedule(&scheduleItem{runAt: next, job: depJob})
				s.wakeScheduler()
			}

		case DependencyFailureCancel:
			if s.logger != nil {
				s.logger.Warn("canceling dependent job due to dependency failure", log.Fields{
					"job_id":            depJobID,
					"failed_dependency": failedJob.id,
				})
			}
			s.cancelJobLocked(depJobID, depJob)

		case DependencyFailureContinue:
			if s.logger != nil {
				s.logger.Info("continuing dependent job despite dependency failure", log.Fields{
					"job_id":            depJobID,
					"failed_dependency": failedJob.id,
				})
			}
			s.dependencyStatus[failedJob.id] = true
			if s.allDependenciesMetLocked(depJob) {
				s.setJobStateIfIdleLocked(depJob, JobStateScheduled)
				s.queue.PushSchedule(&scheduleItem{runAt: depJob.runAt, job: depJob})
				s.wakeScheduler()
			}
		}
	}
}

// addToDeadLetterQueue adds a failed job to the dead letter queue.
func (s *Scheduler) addToDeadLetterQueue(j *job, err error, attempts int) {
	now := s.clock.Now()
	entry := DeadLetterEntry{
		JobID:       j.id,
		Error:       err,
		Attempts:    attempts,
		FirstFailed: now,
		LastFailed:  now,
		TaskName:    j.options.TaskName,
		Group:       j.options.Group,
		Tags:        j.options.Tags,
	}
	s.dlq.Add(entry)
}

func (s *Scheduler) setJobStateLocked(j *job, state JobState) {
	if j == nil {
		return
	}
	if j.state == state {
		return
	}
	j.state = state
	j.stateAt = s.clock.Now()
}

func (s *Scheduler) setJobStateIfIdleLocked(j *job, state JobState) {
	if j == nil {
		return
	}
	if j.running.Load() {
		return
	}
	if j.state == JobStateRetrying {
		return
	}
	s.setJobStateLocked(j, state)
}

// wakeScheduler sends a non-blocking signal to the scheduler run loop.
func (s *Scheduler) wakeScheduler() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

// cancelJobLocked cancels a single job. Must be called with s.mu held.
func (s *Scheduler) cancelJobLocked(id JobID, j *job) {
	j.canceled.Store(true)
	s.setJobStateLocked(j, JobStateCanceled)
	if s.store != nil {
		_ = s.store.Delete(id)
	}
}

// resumeJobLocked resumes a paused job and reschedules it. Must be called with s.mu held.
func (s *Scheduler) resumeJobLocked(j *job) {
	j.paused.Store(false)
	if j.kind == jobKindCron {
		j.runAt = j.cron.Next(s.clock.Now())
		s.pushScheduleLocked(j, j.runAt)
	} else if j.kind == jobKindDelay {
		s.pushScheduleLocked(j, j.runAt)
	}
}

// allDependenciesMetLocked checks if all dependencies for a job have completed successfully.
// Must be called with s.mu held.
func (s *Scheduler) allDependenciesMetLocked(j *job) bool {
	for _, requiredDepID := range j.options.Dependencies {
		if status, ok := s.dependencyStatus[requiredDepID]; !ok || !status {
			return false
		}
	}
	return true
}

// ListDeadLetters returns all dead letter entries.
// Returns nil if DLQ is not enabled.
func (s *Scheduler) ListDeadLetters() []DeadLetterEntry {
	if s.dlq == nil {
		return nil
	}
	return s.dlq.List()
}

// GetDeadLetter retrieves a specific dead letter entry.
// Returns nil if DLQ is not enabled or entry not found.
func (s *Scheduler) GetDeadLetter(jobID JobID) (*DeadLetterEntry, bool) {
	if s.dlq == nil {
		return nil, false
	}
	return s.dlq.Get(jobID)
}

// DeleteDeadLetter removes a dead letter entry.
// Returns false if DLQ is not enabled or entry not found.
func (s *Scheduler) DeleteDeadLetter(jobID JobID) bool {
	if s.dlq == nil {
		return false
	}
	return s.dlq.Delete(jobID)
}

// ClearDeadLetters removes all dead letter entries.
// Returns the number of entries cleared.
func (s *Scheduler) ClearDeadLetters() int {
	if s.dlq == nil {
		return 0
	}
	return s.dlq.Clear()
}

// RequeueDeadLetter retries a job from the dead letter queue.
// The job must be re-registered with the same ID using AddCron or Delay.
// Returns ErrJobNotFound if the entry doesn't exist in DLQ.
func (s *Scheduler) RequeueDeadLetter(jobID JobID, task TaskFunc, opts ...JobOption) (JobID, error) {
	if s.dlq == nil {
		return "", ErrJobNotFound
	}

	entry, exists := s.dlq.Get(jobID)
	if !exists {
		return "", ErrJobNotFound
	}

	// Remove from DLQ
	s.dlq.Delete(jobID)

	// Re-add the job as a delay task with immediate execution.
	// Copy opts to avoid mutating the caller's slice.
	allOpts := make([]JobOption, len(opts)+1)
	copy(allOpts, opts)
	allOpts[len(opts)] = ReplaceExisting()
	return s.Delay(entry.JobID, 0, task, allOpts...)
}

// DeadLetterQueueSize returns the current size of the dead letter queue.
// Returns 0 if DLQ is not enabled.
func (s *Scheduler) DeadLetterQueueSize() int {
	if s.dlq == nil {
		return 0
	}
	return s.dlq.Size()
}
