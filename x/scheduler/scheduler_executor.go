package scheduler

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/spcent/plumego/log"
)

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
	ctx = WithJobID(ctx, j.id)
	ctx = WithJobAttempt(ctx, attempt)
	ctx = WithJobScheduledAt(ctx, scheduledAt)
	if lastErr != nil {
		ctx = WithJobLastError(ctx, lastErr)
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
