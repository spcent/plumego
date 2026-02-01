package scheduler

import "errors"

// Sentinel errors for the scheduler package.
// Use errors.Is() for comparison.
var (
	// ErrSchedulerClosed indicates the scheduler has been stopped.
	ErrSchedulerClosed = errors.New("scheduler: closed")

	// ErrJobExists indicates a job ID is already registered.
	ErrJobExists = errors.New("scheduler: job already exists")

	// ErrJobNotFound indicates the requested job does not exist.
	ErrJobNotFound = errors.New("scheduler: job not found")

	// ErrTaskNil indicates a nil task function was provided.
	ErrTaskNil = errors.New("scheduler: task cannot be nil")

	// ErrTaskNameEmpty indicates an empty task name was provided.
	ErrTaskNameEmpty = errors.New("scheduler: task name cannot be empty")

	// ErrRunAtRequired indicates a zero-value runAt time was provided.
	ErrRunAtRequired = errors.New("scheduler: runAt is required")

	// ErrSelfDependency indicates a job was configured to depend on itself.
	ErrSelfDependency = errors.New("scheduler: job cannot depend on itself")

	// ErrDependencyNotFound indicates a dependency job does not exist.
	ErrDependencyNotFound = errors.New("scheduler: dependency job not found")

	// ErrInvalidCronExpr indicates an invalid cron expression.
	ErrInvalidCronExpr = errors.New("scheduler: invalid cron expression")

	// ErrInvalidCronField indicates an invalid cron field value.
	ErrInvalidCronField = errors.New("scheduler: invalid cron field")

	// ErrInvalidEveryDuration indicates an invalid @every duration.
	ErrInvalidEveryDuration = errors.New("scheduler: invalid @every duration")
)
