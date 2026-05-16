package scheduler

import "time"

// MetricsSink receives scheduler runtime events.
type MetricsSink interface {
	ObserveRun(jobID JobID, duration time.Duration, err error, queueDelay time.Duration)
	ObserveRetry(jobID JobID, attempt int, delay time.Duration)
	ObserveDrop(jobID JobID)
}
