package scheduler

import (
	"context"
	"time"
)

// TaskFunc is the function signature for scheduled work.
type TaskFunc func(context.Context) error

// JobID identifies a scheduled job.
type JobID string

type jobIDKey struct{}
type jobAttemptKey struct{}
type jobScheduledKey struct{}
type jobLastErrorKey struct{}

// JobIDFromContext returns the job ID stored in the context.
func JobIDFromContext(ctx context.Context) (JobID, bool) {
	if ctx == nil {
		return "", false
	}
	val, ok := ctx.Value(jobIDKey{}).(JobID)
	return val, ok
}

// JobAttemptFromContext returns the attempt number stored in the context.
func JobAttemptFromContext(ctx context.Context) (int, bool) {
	if ctx == nil {
		return 0, false
	}
	val, ok := ctx.Value(jobAttemptKey{}).(int)
	return val, ok
}

// JobScheduledAtFromContext returns the scheduled time stored in the context.
func JobScheduledAtFromContext(ctx context.Context) (time.Time, bool) {
	if ctx == nil {
		return time.Time{}, false
	}
	val, ok := ctx.Value(jobScheduledKey{}).(time.Time)
	return val, ok
}

// JobLastErrorFromContext returns the last error stored in the context.
func JobLastErrorFromContext(ctx context.Context) (error, bool) {
	if ctx == nil {
		return nil, false
	}
	val, ok := ctx.Value(jobLastErrorKey{}).(error)
	return val, ok
}

// OverlapPolicy controls behavior when a job is still running.
type OverlapPolicy int

const (
	// AllowConcurrent allows overlapping runs.
	AllowConcurrent OverlapPolicy = iota
	// SkipIfRunning skips a run if the previous run is still active.
	SkipIfRunning
	// SerialQueue queues the next run to execute after the current run finishes.
	SerialQueue
)

// RetryPolicy controls retry behavior for failed tasks.
type RetryPolicy struct {
	// MaxAttempts includes the initial attempt (e.g. 3 = 1 try + 2 retries).
	MaxAttempts int
	// Backoff returns the delay before the next attempt.
	Backoff func(attempt int) time.Duration
	// MaxBackoff caps the backoff duration (0 = no cap).
	MaxBackoff time.Duration
	// Kind describes the retry strategy (\"fixed\" or \"exponential\") when Backoff is nil.
	Kind string
	// BaseDelay is used by fixed/exponential strategies.
	BaseDelay time.Duration
}

// JobOptions customize a job.
type JobOptions struct {
	Timeout       time.Duration
	OverlapPolicy OverlapPolicy
	RetryPolicy   RetryPolicy
	Paused        bool
	DeadLetter    func(context.Context, JobID, error)
	Group         string
	Tags          []string
	Replace       bool
	TaskName      string
}

// JobOption mutates JobOptions.
type JobOption func(*JobOptions)

// WithTimeout sets a per-run timeout.
func WithTimeout(timeout time.Duration) JobOption {
	return func(opts *JobOptions) {
		opts.Timeout = timeout
	}
}

// WithOverlapPolicy sets the overlap policy.
func WithOverlapPolicy(policy OverlapPolicy) JobOption {
	return func(opts *JobOptions) {
		opts.OverlapPolicy = policy
	}
}

// WithRetryPolicy sets the retry policy.
func WithRetryPolicy(policy RetryPolicy) JobOption {
	return func(opts *JobOptions) {
		opts.RetryPolicy = policy
	}
}

// WithDeadLetter sets a callback invoked when retries are exhausted.
func WithDeadLetter(fn func(context.Context, JobID, error)) JobOption {
	return func(opts *JobOptions) {
		opts.DeadLetter = fn
	}
}

// WithGroup sets a job group name.
func WithGroup(group string) JobOption {
	return func(opts *JobOptions) {
		opts.Group = group
	}
}

// WithTags sets job tags.
func WithTags(tags ...string) JobOption {
	return func(opts *JobOptions) {
		opts.Tags = append([]string(nil), tags...)
	}
}

// ReplaceExisting replaces an existing job with the same ID.
func ReplaceExisting() JobOption {
	return func(opts *JobOptions) {
		opts.Replace = true
	}
}

// WithTaskName associates a job with a registered task name (for persistence).
func WithTaskName(name string) JobOption {
	return func(opts *JobOptions) {
		opts.TaskName = name
	}
}

// RetryFixed returns a fixed-delay retry policy.
func RetryFixed(maxAttempts int, delay time.Duration) RetryPolicy {
	return RetryPolicy{
		MaxAttempts: maxAttempts,
		Kind:        "fixed",
		BaseDelay:   delay,
		Backoff: func(_ int) time.Duration {
			return delay
		},
	}
}

// RetryExponential returns an exponential backoff retry policy.
func RetryExponential(maxAttempts int, base time.Duration, maxBackoff time.Duration) RetryPolicy {
	return RetryPolicy{
		MaxAttempts: maxAttempts,
		MaxBackoff:  maxBackoff,
		Kind:        "exponential",
		BaseDelay:   base,
		Backoff: func(attempt int) time.Duration {
			if attempt <= 1 {
				return base
			}
			backoff := base << (attempt - 1)
			if maxBackoff > 0 && backoff > maxBackoff {
				return maxBackoff
			}
			return backoff
		},
	}
}

func serializeRetry(policy RetryPolicy) RetrySpec {
	return RetrySpec{
		MaxAttempts: policy.MaxAttempts,
		MaxBackoff:  policy.MaxBackoff,
		Kind:        policy.Kind,
		BaseDelay:   policy.BaseDelay,
	}
}

func hydrateRetry(spec RetrySpec) RetryPolicy {
	policy := RetryPolicy{
		MaxAttempts: spec.MaxAttempts,
		MaxBackoff:  spec.MaxBackoff,
		Kind:        spec.Kind,
		BaseDelay:   spec.BaseDelay,
	}
	if policy.Backoff != nil || policy.Kind == "" {
		return policy
	}
	switch policy.Kind {
	case "fixed":
		if policy.BaseDelay <= 0 {
			return policy
		}
		policy.Backoff = func(_ int) time.Duration { return policy.BaseDelay }
	case "exponential":
		if policy.BaseDelay <= 0 {
			return policy
		}
		policy.Backoff = func(attempt int) time.Duration {
			if attempt <= 1 {
				return policy.BaseDelay
			}
			backoff := policy.BaseDelay << (attempt - 1)
			if policy.MaxBackoff > 0 && backoff > policy.MaxBackoff {
				return policy.MaxBackoff
			}
			return backoff
		}
	}
	return policy
}

// Paused marks the job as paused on creation.
func Paused() JobOption {
	return func(opts *JobOptions) {
		opts.Paused = true
	}
}

func defaultJobOptions() JobOptions {
	return JobOptions{
		OverlapPolicy: AllowConcurrent,
	}
}

// JobStatus exposes job runtime state for observability.
type JobStatus struct {
	ID            JobID
	NextRun       time.Time
	LastRun       time.Time
	LastError     error
	Attempt       int
	Paused        bool
	Running       bool
	Kind          string
	OverlapPolicy OverlapPolicy
	Group         string
	Tags          []string
}
