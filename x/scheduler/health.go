package scheduler

// HealthSnapshot summarizes scheduler health for probes.
type HealthSnapshot struct {
	Queued        int   // current work-queue depth
	Workers       int   // configured worker pool size
	QueueSize     int   // work-queue capacity
	Dropped       int64 // total dropped executions since start
	RunningJobs   int   // jobs currently executing
	TotalJobs     int   // total registered jobs (including paused/canceled)
	ScheduledJobs int   // active jobs (not paused, not canceled)
}

// Health returns a snapshot of scheduler load.
func (s *Scheduler) Health() HealthSnapshot {
	s.mu.RLock()
	running := 0
	total := len(s.jobs)
	scheduled := 0
	for _, j := range s.jobs {
		if j.running.Load() {
			running++
		}
		if !j.paused.Load() && !j.canceled.Load() {
			scheduled++
		}
	}
	// Read channel metrics inside the same read lock so the snapshot is
	// temporally consistent with the job counts above.
	queued := len(s.workCh)
	queueSize := cap(s.workCh)
	workerCount := s.workerCount
	s.mu.RUnlock()

	stats := s.Stats()
	return HealthSnapshot{
		Queued:        queued,
		Workers:       workerCount,
		QueueSize:     queueSize,
		Dropped:       stats.Dropped,
		RunningJobs:   running,
		TotalJobs:     total,
		ScheduledJobs: scheduled,
	}
}
