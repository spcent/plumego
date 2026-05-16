package scheduler

import "sync"

// MemoryStore keeps scheduled jobs in memory.
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[JobID]StoredJob
}

// NewMemoryStore constructs an in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{jobs: make(map[JobID]StoredJob)}
}

func (m *MemoryStore) Save(job StoredJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = cloneStoredJob(job)
	return nil
}

func (m *MemoryStore) Delete(id JobID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, id)
	return nil
}

func (m *MemoryStore) List() ([]StoredJob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]StoredJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		out = append(out, cloneStoredJob(job))
	}
	return out, nil
}

func cloneStoredJob(job StoredJob) StoredJob {
	job.Tags = append([]string(nil), job.Tags...)
	job.Dependencies = append([]JobID(nil), job.Dependencies...)
	return job
}
