package scheduler

import "sync"

// MemoryStore keeps scheduled jobs in memory.
type MemoryStore struct {
	mu   sync.Mutex
	jobs map[JobID]StoredJob
}

// NewMemoryStore constructs an in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{jobs: make(map[JobID]StoredJob)}
}

func (m *MemoryStore) Save(job StoredJob) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.jobs[job.ID] = job
	return nil
}

func (m *MemoryStore) Delete(id JobID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.jobs, id)
	return nil
}

func (m *MemoryStore) List() ([]StoredJob, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]StoredJob, 0, len(m.jobs))
	for _, job := range m.jobs {
		out = append(out, job)
	}
	return out, nil
}
