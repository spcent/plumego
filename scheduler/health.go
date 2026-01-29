package scheduler

// HealthSnapshot summarizes scheduler health for probes.
type HealthSnapshot struct {
	Queued      int
	Workers     int
	QueueSize   int
	Dropped     int64
	RunningJobs int
}

// Health returns a snapshot of scheduler load.
func (s *Scheduler) Health() HealthSnapshot {
	s.mu.Lock()
	running := 0
	for _, j := range s.jobs {
		if j.running.Load() {
			running++
		}
	}
	s.mu.Unlock()

	stats := s.Stats()
	return HealthSnapshot{
		Queued:      len(s.workCh),
		Workers:     s.workerCount,
		QueueSize:   cap(s.workCh),
		Dropped:     stats.Dropped,
		RunningJobs: running,
	}
}
