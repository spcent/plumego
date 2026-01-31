package scheduler

import (
	"sync/atomic"
	"time"
)

type schedulerStats struct {
	queued            atomic.Int64
	dropped           atomic.Int64
	success           atomic.Int64
	failure           atomic.Int64
	retry             atomic.Int64
	timeout           atomic.Int64
	queueDelayTotalNs atomic.Int64
}

func (s *schedulerStats) incQueued() {
	s.queued.Add(1)
}

func (s *schedulerStats) incDropped() {
	s.dropped.Add(1)
}

func (s *schedulerStats) incSuccess() {
	s.success.Add(1)
}

func (s *schedulerStats) incFailure() {
	s.failure.Add(1)
}

func (s *schedulerStats) incRetry() {
	s.retry.Add(1)
}

func (s *schedulerStats) incTimeout() {
	s.timeout.Add(1)
}

func (s *schedulerStats) addQueueDelay(d time.Duration) {
	s.queueDelayTotalNs.Add(d.Nanoseconds())
}

// StatsSnapshot captures scheduler runtime stats.
type StatsSnapshot struct {
	Queued        int64
	Dropped       int64
	Success       int64
	Failure       int64
	Retry         int64
	Timeout       int64
	QueueDelayAvg time.Duration
}

// Stats returns a snapshot of scheduler stats.
func (s *Scheduler) Stats() StatsSnapshot {
	queued := s.stats.queued.Load()
	totalDelay := s.stats.queueDelayTotalNs.Load()
	avg := time.Duration(0)
	if queued > 0 {
		avg = time.Duration(totalDelay / queued)
	}
	return StatsSnapshot{
		Queued:        queued,
		Dropped:       s.stats.dropped.Load(),
		Success:       s.stats.success.Load(),
		Failure:       s.stats.failure.Load(),
		Retry:         s.stats.retry.Load(),
		Timeout:       s.stats.timeout.Load(),
		QueueDelayAvg: avg,
	}
}
