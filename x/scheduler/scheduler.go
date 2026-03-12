package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/spcent/plumego/log"
)

const (
	// DefaultWorkChannelBuffer is the default buffer size for the work channel.
	DefaultWorkChannelBuffer = 256
)

// Error definitions moved to errors.go

// Scheduler coordinates cron, delayed, and retryable jobs.
type Scheduler struct {
	mu      sync.RWMutex
	jobs    map[JobID]*job
	queue   scheduleHeap
	wakeCh  chan struct{}
	stopCh  chan struct{}
	closed  bool
	started bool // guards against multiple Start() calls

	// stopCtx is cancelled when the scheduler is stopped, allowing
	// running tasks to detect shutdown when no per-job timeout is set.
	stopCtx    context.Context
	stopCancel context.CancelFunc

	workerCount  int
	workCh       chan *runRequest
	logger       log.StructuredLogger
	panicHandler func(context.Context, JobID, any)
	clock        Clock
	store        Store
	registry     map[string]TaskFunc
	metricsSink  MetricsSink
	backpressure BackpressureConfig
	rngMu        sync.Mutex
	rng          *rand.Rand

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

// WithRandomSeed sets the scheduler-local random seed used for jitter.
func WithRandomSeed(seed int64) Option {
	return func(s *Scheduler) {
		s.rng = rand.New(rand.NewSource(seed))
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
	ctx, cancel := context.WithCancel(context.Background())
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
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		dependents:       make(map[JobID][]JobID),
		dependencyStatus: make(map[JobID]bool),
		stopCtx:          ctx,
		stopCancel:       cancel,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(s)
		}
	}
	return s
}

// Start launches scheduler goroutines.
// Calling Start more than once on the same Scheduler is a no-op.
func (s *Scheduler) Start() {
	s.mu.Lock()
	if s.started || s.closed {
		s.mu.Unlock()
		return
	}
	s.started = true
	s.mu.Unlock()

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
	s.stopCancel() // signal running tasks via stopCtx
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

// TriggerNow immediately queues a job for execution regardless of its next
// scheduled time. For cron jobs the normal schedule is not affected; the job
// will also run at its next cron-calculated time. For delay jobs the
// behaviour is equivalent to re-running an already-queued job.
//
// Returns ErrJobNotFound if the job does not exist, ErrSchedulerClosed if
// the scheduler has been stopped.
func (s *Scheduler) TriggerNow(id JobID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSchedulerClosed
	}
	j, exists := s.jobs[id]
	if !exists || j.canceled.Load() {
		return ErrJobNotFound
	}
	// Delay jobs are one-shot; reject TriggerNow once they've completed.
	if j.kind == jobKindDelay && j.state == JobStateCompleted {
		return fmt.Errorf("scheduler: delay job %q has already completed", id)
	}
	if j.paused.Load() {
		return fmt.Errorf("scheduler: job %q is paused", id)
	}
	invalidateExisting := j.kind == jobKindDelay
	s.pushScheduleItemLocked(j, s.clock.Now(), invalidateExisting)
	s.wakeScheduler()
	return nil
}

// UpdateCron updates the cron expression of a registered cron job.
// The new schedule takes effect immediately; the next run time is recalculated
// from now. Returns ErrJobNotFound if the job does not exist or is not a cron
// job. Returns an error if the new expression cannot be parsed.
func (s *Scheduler) UpdateCron(id JobID, spec string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return ErrSchedulerClosed
	}
	j, exists := s.jobs[id]
	if !exists || j.canceled.Load() {
		return ErrJobNotFound
	}
	if j.kind != jobKindCron {
		return fmt.Errorf("scheduler: job %q is not a cron job", id)
	}
	parsed, err := ParseCronSpecWithLocation(spec, j.cron.location)
	if err != nil {
		return err
	}
	j.cron = parsed
	j.cronExpr = parsed.Expr()
	next := parsed.Next(s.clock.Now())
	if next.IsZero() {
		return fmt.Errorf("scheduler: new spec %q yields no future run time", spec)
	}
	j.runAt = next
	if !j.paused.Load() {
		s.pushScheduleItemLocked(j, next, true)
		s.wakeScheduler()
	}
	return nil
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

// PruneTerminalJobs removes terminal jobs from in-memory registry to bound
// memory growth in long-running processes.
//
// A job is considered terminal when its state is completed, failed, or canceled.
// Running jobs are never pruned. Jobs that still have live dependents are kept.
//
// limit controls how many jobs to remove:
//   - limit <= 0: prune all eligible jobs
//   - limit > 0: prune at most limit jobs
//
// Returns the number of pruned jobs.
func (s *Scheduler) PruneTerminalJobs(limit int) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	pruned := 0
	for id, j := range s.jobs {
		if j == nil || j.running.Load() || !isTerminalState(j.state) {
			continue
		}
		if s.hasLiveDependentsLocked(id) {
			continue
		}
		// Mark canceled defensively so stale queued pointers (if any) cannot run.
		j.canceled.Store(true)
		delete(s.jobs, id)
		s.cleanupDependencyTrackingLocked(id, j)
		pruned++
		if limit > 0 && pruned >= limit {
			break
		}
	}
	return pruned
}

// PauseByGroup pauses all jobs in the specified group.
// Returns the number of jobs paused.
func (s *Scheduler) PauseByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool { return j.options.Group == group && !j.paused.Load() },
		func(_ JobID, j *job) { j.paused.Store(true) },
	)
}

// PauseByTags pauses all jobs that have ALL of the specified tags.
// Returns the number of jobs paused.
func (s *Scheduler) PauseByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool { return hasAllTags(j.options.Tags, tags) && !j.paused.Load() },
		func(_ JobID, j *job) { j.paused.Store(true) },
	)
}

// ResumeByGroup resumes all paused jobs in the specified group.
// Returns the number of jobs resumed.
func (s *Scheduler) ResumeByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool { return j.options.Group == group && j.paused.Load() && !j.canceled.Load() },
		func(_ JobID, j *job) { s.resumeJobLocked(j) },
	)
}

// ResumeByTags resumes all paused jobs that have ALL of the specified tags.
// Returns the number of jobs resumed.
func (s *Scheduler) ResumeByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool {
			return hasAllTags(j.options.Tags, tags) && j.paused.Load() && !j.canceled.Load()
		},
		func(_ JobID, j *job) { s.resumeJobLocked(j) },
	)
}

// CancelByGroup cancels all jobs in the specified group.
// Returns the number of jobs canceled.
func (s *Scheduler) CancelByGroup(group string) int {
	if group == "" {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool { return j.options.Group == group },
		func(id JobID, j *job) { s.cancelJobLocked(id, j) },
	)
}

// CancelByTags cancels all jobs that have ALL of the specified tags.
// Returns the number of jobs canceled.
func (s *Scheduler) CancelByTags(tags ...string) int {
	if len(tags) == 0 {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.applyBatchLocked(
		func(_ JobID, j *job) bool { return hasAllTags(j.options.Tags, tags) },
		func(id JobID, j *job) { s.cancelJobLocked(id, j) },
	)
}

// applyBatchLocked iterates all jobs and applies mutator to matched jobs.
// s.mu must be held by caller.
func (s *Scheduler) applyBatchLocked(match func(JobID, *job) bool, mutate func(JobID, *job)) int {
	if match == nil || mutate == nil {
		return 0
	}
	count := 0
	for id, j := range s.jobs {
		if !match(id, j) {
			continue
		}
		mutate(id, j)
		count++
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

	matcher := newJobQueryMatcher(query)

	// Collect and filter jobs
	var filtered []JobStatus
	for _, j := range s.jobs {
		status := jobStatusFrom(j)
		if !matcher.match(status) {
			continue
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
	less := buildJobStatusLess(orderBy)
	if less == nil {
		return
	}
	sort.SliceStable(jobs, func(i, j int) bool {
		if ascending {
			return less(jobs[i], jobs[j])
		}
		return less(jobs[j], jobs[i])
	})
}

type jobQueryMatcher struct {
	group   string
	tags    []string
	kinds   map[string]struct{}
	running *bool
	paused  *bool
	states  map[JobState]struct{}
}

func newJobQueryMatcher(query JobQuery) jobQueryMatcher {
	m := jobQueryMatcher{
		group:   query.Group,
		tags:    query.Tags,
		running: query.Running,
		paused:  query.Paused,
	}
	if len(query.Kinds) > 0 {
		m.kinds = make(map[string]struct{}, len(query.Kinds))
		for _, kind := range query.Kinds {
			m.kinds[kind] = struct{}{}
		}
	}
	if len(query.States) > 0 {
		m.states = make(map[JobState]struct{}, len(query.States))
		for _, state := range query.States {
			m.states[state] = struct{}{}
		}
	}
	return m
}

func (m jobQueryMatcher) match(status JobStatus) bool {
	if m.group != "" && status.Group != m.group {
		return false
	}
	if len(m.tags) > 0 && !hasAllTags(status.Tags, m.tags) {
		return false
	}
	if len(m.kinds) > 0 {
		if _, ok := m.kinds[status.Kind]; !ok {
			return false
		}
	}
	if m.running != nil && status.Running != *m.running {
		return false
	}
	if m.paused != nil && status.Paused != *m.paused {
		return false
	}
	if len(m.states) > 0 {
		if _, ok := m.states[status.State]; !ok {
			return false
		}
	}
	return true
}

func buildJobStatusLess(orderBy string) func(a, b JobStatus) bool {
	switch orderBy {
	case "id":
		return func(a, b JobStatus) bool {
			return a.ID < b.ID
		}
	case "next_run":
		return func(a, b JobStatus) bool {
			if a.NextRun.Equal(b.NextRun) {
				return a.ID < b.ID
			}
			return a.NextRun.Before(b.NextRun)
		}
	case "last_run":
		return func(a, b JobStatus) bool {
			if a.LastRun.Equal(b.LastRun) {
				return a.ID < b.ID
			}
			return a.LastRun.Before(b.LastRun)
		}
	case "group":
		return func(a, b JobStatus) bool {
			if a.Group == b.Group {
				return a.ID < b.ID
			}
			return a.Group < b.Group
		}
	default:
		return nil
	}
}

func (s *Scheduler) addJob(id JobID, kind jobKind, task TaskFunc, spec CronSpec, runAt time.Time, opts ...JobOption) (JobID, error) {
	if task == nil {
		return "", ErrTaskNil
	}
	if id == "" {
		id = JobID(fmt.Sprintf("job-%d", s.clock.Now().UnixNano()))
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
	var existing *job
	if current, exists := s.jobs[id]; exists {
		existing = current
		if !jobOpts.Replace && !existing.canceled.Load() {
			return "", ErrJobExists
		}
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
		j.cronExpr = spec.Expr()
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
	if s.wouldCreateDependencyCycleLocked(id, jobOpts.Dependencies) {
		return "", fmt.Errorf("%w: %s", ErrDependencyCycle, id)
	}
	if existing != nil {
		// Use cancelJobLocked to ensure full cleanup: dependency maps,
		// dependents reverse-index, and the persistence store are all
		// cleared before the replacement job is registered.
		s.cancelJobLocked(id, existing)
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
	s.pushScheduleItemLocked(j, runAt, true)
	s.wakeScheduler()
}

func (s *Scheduler) pushScheduleItemLocked(j *job, runAt time.Time, invalidateExisting bool) {
	if j == nil {
		return
	}
	version := uint64(0)
	if invalidateExisting {
		j.scheduleVersion++
		version = j.scheduleVersion
	}
	s.queue.PushSchedule(&scheduleItem{runAt: runAt, job: j, version: version})
}

func (s *Scheduler) wouldCreateDependencyCycleLocked(newID JobID, dependencies []JobID) bool {
	if len(dependencies) == 0 {
		return false
	}
	for _, depID := range dependencies {
		if depID == newID {
			return true
		}
		visited := map[JobID]bool{}
		if s.depPathExistsLocked(depID, newID, visited) {
			return true
		}
	}
	return false
}

func (s *Scheduler) depPathExistsLocked(current, target JobID, visited map[JobID]bool) bool {
	if current == target {
		return true
	}
	if visited[current] {
		return false
	}
	visited[current] = true
	j, ok := s.jobs[current]
	if !ok {
		return false
	}
	for _, depID := range j.options.Dependencies {
		if s.depPathExistsLocked(depID, target, visited) {
			return true
		}
	}
	return false
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

		wait := item.runAt.Sub(s.clock.Now())
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

		// Sample clock after the wait so popDue sees an accurate "now".
		now := s.clock.Now()

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
	for {
		item := s.queue.PopDue(now)
		if item == nil {
			return nil
		}
		if item.job == nil {
			continue
		}
		if item.version > 0 && item.job.scheduleVersion != item.version {
			continue
		}
		return item
	}
}

func (s *Scheduler) dispatch(j *job) {
	if j == nil || j.canceled.Load() || j.paused.Load() {
		return
	}

	if j.options.OverlapPolicy == SkipIfRunning && !j.running.CompareAndSwap(false, true) {
		s.scheduleNext(j, s.clock.Now())
		return
	}
	if j.options.OverlapPolicy == SerialQueue {
		if !j.running.CompareAndSwap(false, true) {
			// Use atomic operation to set pending flag to avoid race condition
			j.pending.Store(true)
			return
		}
	}

	req := &runRequest{job: j, enqueuedAt: s.clock.Now()}

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
		if s.backpressure.Timeout <= 0 {
			// No timeout configured — fall back to non-blocking drop.
			select {
			case s.workCh <- req:
				s.stats.incQueued()
			case <-s.stopCh:
			default:
				s.handleBackpressure(j)
			}
		} else {
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
	// dispatch() set running=true via CAS before calling us, but the job never
	// actually executed. Clear the flag so the job can be dispatched again.
	if j.options.OverlapPolicy == SkipIfRunning || j.options.OverlapPolicy == SerialQueue {
		j.running.Store(false)
	}
	// Reschedule cron jobs for their next occurrence so they are not silently
	// lost when the worker pool is saturated.
	if j.kind == jobKindCron {
		s.scheduleNext(j, s.clock.Now())
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
	// Apply jitter: randomise execution start within [0, Jitter).
	if j.options.Jitter > 0 {
		next = next.Add(s.nextJitter(j.options.Jitter))
	}
	j.runAt = next
	s.setJobStateIfIdleLocked(j, JobStateScheduled)
	s.pushScheduleItemLocked(j, next, true)
	s.wakeScheduler()
}

func (s *Scheduler) nextJitter(max time.Duration) time.Duration {
	if max <= 0 {
		return 0
	}
	s.rngMu.Lock()
	defer s.rngMu.Unlock()
	if s.rng == nil {
		s.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return time.Duration(s.rng.Int63n(int64(max)))
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
	// Derive execution context from the scheduler's stop context so that
	// running tasks are notified when the scheduler shuts down.
	ctx := s.stopCtx
	cancel := func() {}
	if j.options.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, j.options.Timeout)
	}
	queueDelay := s.clock.Now().Sub(enqueuedAt)
	s.stats.addQueueDelay(queueDelay)
	start := s.clock.Now()
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

	elapsed := s.clock.Now().Sub(start)

	if err != nil {
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			// Per-job timeout fired.
			s.stats.incTimeout()
		case errors.Is(err, context.Canceled):
			// Scheduler shutting down; do not count as a task timeout.
			s.stats.incCanceled()
		}
		s.stats.incFailure()
		if s.metricsSink != nil {
			s.metricsSink.ObserveRun(j.id, elapsed, err, queueDelay)
		}
		s.handleFailure(j, err)
		return
	}

	s.stats.incSuccess()
	if s.metricsSink != nil {
		s.metricsSink.ObserveRun(j.id, elapsed, nil, queueDelay)
	}
	s.mu.Lock()
	j.nextAttempt = 1
	j.runCount++
	maxRunsReached := j.options.MaxRuns > 0 && j.runCount >= int64(j.options.MaxRuns)
	if j.kind == jobKindDelay || maxRunsReached {
		s.setJobStateLocked(j, JobStateCompleted)
	}
	// Mark dependency as succeeded so dependent-job checks see the correct status.
	s.dependencyStatus[j.id] = true
	// Use atomic operation to read and clear pending flag.
	// Guard with !j.running.Load(): if dispatch() stole the running slot between
	// j.running.Store(false) (line above) and here, the pending execution is
	// already handled by that dispatch — don't re-queue again.
	pending := j.options.OverlapPolicy == SerialQueue && j.pending.Swap(false)
	if pending && !maxRunsReached && !j.running.Load() {
		s.setJobStateIfIdleLocked(j, JobStateScheduled)
		s.pushScheduleItemLocked(j, s.clock.Now(), true)
		s.wakeScheduler()
	}
	s.mu.Unlock()

	if j.kind == jobKindCron && !pending && !maxRunsReached {
		s.scheduleNext(j, start)
	}
	if j.kind == jobKindDelay && s.store != nil {
		_ = s.store.Delete(j.id)
	}

	// Trigger dependent jobs on success. Must happen BEFORE retireJobLocked so
	// that the dependents map entry for this job is still intact.
	s.triggerDependentJobs(j.id)

	// Retire the job after dependents have been triggered. retireJobLocked only
	// stops future scheduling; it preserves the Completed state and does not
	// re-notify dependents (they were already handled above as a success).
	if maxRunsReached {
		s.mu.Lock()
		s.retireJobLocked(j.id, j)
		s.mu.Unlock()
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
	s.pushScheduleItemLocked(j, s.clock.Now().Add(backoff), true)
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
	// Only notify dependents of failure after all retry attempts are exhausted.
	// Moving this here (rather than on every failure) prevents dependents from
	// being marked failed when the job succeeds on a subsequent retry attempt.
	s.mu.Lock()
	s.dependencyStatus[j.id] = false
	s.mu.Unlock()
	s.handleDependentJobsOnFailure(j)
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
		CronExpr:      j.cronExpr,
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
		ID:               j.id,
		Kind:             "delay",
		RunAt:            j.runAt,
		TaskName:         j.options.TaskName,
		Group:            j.options.Group,
		Tags:             append([]string(nil), j.options.Tags...),
		Timeout:          j.options.Timeout,
		Overlap:          j.options.OverlapPolicy,
		Retry:            serializeRetry(j.options.RetryPolicy),
		CreatedAt:        s.clock.Now(),
		Dependencies:     append([]JobID(nil), j.options.Dependencies...),
		DependencyPolicy: j.options.DependencyPolicy,
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

	// Two-pass loading: first restore jobs without dependencies so that
	// dependency targets exist when the second pass runs.
	var withDeps []StoredJob
	for _, item := range stored {
		if item.TaskName == "" {
			continue
		}
		if s.lookupTask(item.TaskName) == nil {
			continue
		}
		if len(item.Dependencies) > 0 {
			withDeps = append(withDeps, item)
			continue
		}
		if _, err := s.Schedule(item.ID, item.RunAt, s.lookupTask(item.TaskName), s.buildLoadOpts(item)...); err != nil {
			if s.logger != nil {
				s.logger.Warn("scheduler: failed to restore job", log.Fields{
					"job_id": item.ID,
					"error":  err.Error(),
				})
			}
		}
	}

	// Second pass: jobs with dependencies (their targets are now registered).
	for _, item := range withDeps {
		task := s.lookupTask(item.TaskName)
		if task == nil {
			continue
		}
		opts := append(s.buildLoadOpts(item), WithDependsOn(item.DependencyPolicy, item.Dependencies...))
		if _, err := s.Schedule(item.ID, item.RunAt, task, opts...); err != nil {
			if s.logger != nil {
				s.logger.Warn("scheduler: failed to restore job with dependencies", log.Fields{
					"job_id":       item.ID,
					"dependencies": item.Dependencies,
					"error":        err.Error(),
				})
			}
		}
	}
}

// buildLoadOpts constructs the JobOption slice for a persisted job (without dependencies).
func (s *Scheduler) buildLoadOpts(item StoredJob) []JobOption {
	return []JobOption{
		WithTimeout(item.Timeout),
		WithOverlapPolicy(item.Overlap),
		WithRetryPolicy(hydrateRetry(item.Retry)),
		WithGroup(item.Group),
		WithTags(item.Tags...),
		WithTaskName(item.TaskName),
		ReplaceExisting(),
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
			s.pushScheduleItemLocked(depJob, depJob.runAt, true)
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
		s.applyDependentPolicyLocked(depJobID, depJob, failedJob.id, dependencyEventFailure)
	}
}

// addToDeadLetterQueue adds a failed job to the dead letter queue.
func (s *Scheduler) addToDeadLetterQueue(j *job, err error, attempts int) {
	now := s.clock.Now()
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}
	entry := DeadLetterEntry{
		JobID:        j.id,
		Error:        err,
		ErrorMessage: errMsg,
		Attempts:     attempts,
		FirstFailed:  now,
		LastFailed:   now,
		TaskName:     j.options.TaskName,
		Group:        j.options.Group,
		Tags:         append([]string(nil), j.options.Tags...), // defensive copy
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
	// Notify dependent jobs that were waiting for this job before cleaning
	// up the dependents map entry, so we still have the list of waiters.
	if deps := s.dependents[id]; len(deps) > 0 {
		s.notifyDependentsCanceledLocked(id, deps)
	}
	// Clean up dependency tracking to prevent memory leaks.
	s.cleanupDependencyTrackingLocked(id, j)
}

// retireJobLocked is called when a job has exhausted its MaxRuns quota after a
// SUCCESSFUL execution. Unlike cancelJobLocked it:
//   - Preserves the JobStateCompleted state (does not overwrite with Canceled).
//   - Does not notify dependents via DependencyFailurePolicy; the caller has
//     already set dependencyStatus[id]=true and triggerDependentJobs will handle
//     them as a normal successful completion.
//   - Still cleans up the schedule queue slot, store entry, and dependency maps
//     so the job does not run again.
//
// Must be called with s.mu held.
func (s *Scheduler) retireJobLocked(id JobID, j *job) {
	j.canceled.Store(true)
	// State is already JobStateCompleted — do not override.
	if s.store != nil {
		_ = s.store.Delete(id)
	}
	// Clean up dependency tracking (same as cancelJobLocked, minus dependent notification).
	s.cleanupDependencyTrackingLocked(id, j)
}

// notifyDependentsCanceledLocked applies DependencyFailurePolicy to each job
// that was waiting for canceledID to complete. Called from cancelJobLocked;
// s.mu must be held.
func (s *Scheduler) notifyDependentsCanceledLocked(canceledID JobID, dependentIDs []JobID) {
	for _, depJobID := range dependentIDs {
		depJob, exists := s.jobs[depJobID]
		if !exists || depJob.canceled.Load() {
			continue
		}
		s.applyDependentPolicyLocked(depJobID, depJob, canceledID, dependencyEventCancellation)
	}
}

type dependencyEvent struct {
	noun  string
	field string
}

var (
	dependencyEventFailure      = dependencyEvent{noun: "failure", field: "failed_dependency"}
	dependencyEventCancellation = dependencyEvent{noun: "cancellation", field: "canceled_dependency"}
)

func (s *Scheduler) applyDependentPolicyLocked(depJobID JobID, depJob *job, triggerID JobID, event dependencyEvent) {
	switch depJob.options.DependencyPolicy {
	case DependencyFailureSkip:
		s.logDependentPolicy("skipping dependent job due to dependency "+event.noun, depJobID, triggerID, event.field, true)
		s.skipDependentJobLocked(depJobID, depJob)
	case DependencyFailureCancel:
		s.logDependentPolicy("canceling dependent job due to dependency "+event.noun, depJobID, triggerID, event.field, true)
		s.cancelJobLocked(depJobID, depJob)
	case DependencyFailureContinue:
		s.logDependentPolicy("continuing dependent job despite dependency "+event.noun, depJobID, triggerID, event.field, false)
		// DependencyFailureContinue ignores the triggering dependency outcome for
		// this scheduling decision only.
		if s.allDependenciesMetExceptLocked(depJob, triggerID) {
			s.setJobStateIfIdleLocked(depJob, JobStateScheduled)
			s.pushScheduleItemLocked(depJob, depJob.runAt, true)
			s.wakeScheduler()
		}
	}
}

func (s *Scheduler) skipDependentJobLocked(depJobID JobID, depJob *job) {
	switch depJob.kind {
	case jobKindCron:
		// Reschedule cron job at its next natural trigger time.
		next := depJob.cron.Next(s.clock.Now())
		depJob.runAt = next
		s.setJobStateIfIdleLocked(depJob, JobStateScheduled)
		s.pushScheduleItemLocked(depJob, next, true)
		s.wakeScheduler()
	case jobKindDelay:
		// A delay job runs once; if its dependency is skipped/canceled with Skip
		// policy, mark it failed and remove persisted entry.
		s.setJobStateLocked(depJob, JobStateFailed)
		if s.store != nil {
			_ = s.store.Delete(depJobID)
		}
	}
}

func (s *Scheduler) logDependentPolicy(msg string, depJobID JobID, triggerID JobID, triggerField string, warn bool) {
	if s.logger == nil {
		return
	}
	fields := log.Fields{"job_id": depJobID}
	fields[triggerField] = triggerID
	if warn {
		s.logger.Warn(msg, fields)
		return
	}
	s.logger.Info(msg, fields)
}

func (s *Scheduler) cleanupDependencyTrackingLocked(id JobID, j *job) {
	s.cleanupDependencyLinksLocked(id, j)
	delete(s.dependencyStatus, id)
}

func (s *Scheduler) cleanupDependencyLinksLocked(id JobID, j *job) {
	delete(s.dependents, id)
	if j == nil {
		return
	}
	for _, depID := range j.options.Dependencies {
		deps := s.dependents[depID]
		filtered := deps[:0]
		for _, d := range deps {
			if d != id {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			delete(s.dependents, depID)
		} else {
			s.dependents[depID] = filtered
		}
	}
}

func (s *Scheduler) hasLiveDependentsLocked(id JobID) bool {
	deps := s.dependents[id]
	if len(deps) == 0 {
		return false
	}
	live := deps[:0]
	for _, depID := range deps {
		depJob, exists := s.jobs[depID]
		if !exists || depJob == nil {
			continue
		}
		if depJob.canceled.Load() || isTerminalState(depJob.state) {
			continue
		}
		live = append(live, depID)
	}
	if len(live) == 0 {
		delete(s.dependents, id)
		return false
	}
	s.dependents[id] = live
	return true
}

func isTerminalState(state JobState) bool {
	switch state {
	case JobStateCompleted, JobStateFailed, JobStateCanceled:
		return true
	default:
		return false
	}
}

// resumeJobLocked resumes a paused job and reschedules it. Must be called with s.mu held.
func (s *Scheduler) resumeJobLocked(j *job) {
	j.paused.Store(false)

	switch j.kind {
	case jobKindCron:
		j.runAt = j.cron.Next(s.clock.Now())
		s.pushScheduleLocked(j, j.runAt)
	case jobKindDelay:
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

// allDependenciesMetExceptLocked is like allDependenciesMetLocked but skips
// the dependency identified by exceptID. Used for DependencyFailureContinue:
// the failed dependency is ignored so the dependent job can still proceed if
// all other dependencies are satisfied.
// Must be called with s.mu held.
func (s *Scheduler) allDependenciesMetExceptLocked(j *job, exceptID JobID) bool {
	for _, requiredDepID := range j.options.Dependencies {
		if requiredDepID == exceptID {
			continue // intentionally ignore this dependency's outcome
		}
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

	// Re-add the job as a delay task with immediate execution.
	// Copy opts to avoid mutating the caller's slice.
	allOpts := make([]JobOption, len(opts)+1)
	copy(allOpts, opts)
	allOpts[len(opts)] = ReplaceExisting()
	id, err := s.Delay(entry.JobID, 0, task, allOpts...)
	if err != nil {
		return "", err
	}
	// Remove from DLQ only after the job has been successfully re-registered.
	s.dlq.Delete(jobID)
	return id, nil
}

// DeadLetterQueueSize returns the current size of the dead letter queue.
// Returns 0 if DLQ is not enabled.
func (s *Scheduler) DeadLetterQueueSize() int {
	if s.dlq == nil {
		return 0
	}
	return s.dlq.Size()
}
