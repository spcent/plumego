package store

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/spcent/plumego/net/mq"
)

type MemStore struct {
	mu      sync.RWMutex
	tasks   map[string]*memRecord
	dedupe  map[string]string
	nowFunc func() time.Time
}

type memRecord struct {
	task        mq.Task
	status      string
	lastError   string
	lastErrorAt time.Time
}

func NewMemory(cfg MemConfig) *MemStore {
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &MemStore{
		tasks:   make(map[string]*memRecord),
		dedupe:  make(map[string]string),
		nowFunc: cfg.Now,
	}
}

var _ mq.TaskStore = (*MemStore)(nil)

func (m *MemStore) Insert(_ context.Context, task mq.Task) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tasks[task.ID]; exists {
		return mq.ErrDuplicateTask
	}
	if task.DedupeKey != "" {
		key := dedupeKey(task.TenantID, task.DedupeKey)
		if _, exists := m.dedupe[key]; exists {
			return mq.ErrDuplicateTask
		}
		m.dedupe[key] = task.ID
	}

	m.tasks[task.ID] = &memRecord{task: task, status: mq.TaskStatusQueued}
	return nil
}

func (m *MemStore) Reserve(_ context.Context, opts mq.ReserveOptions) ([]mq.Task, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if opts.Limit <= 0 {
		return nil, nil
	}
	if opts.Now.IsZero() {
		opts.Now = m.nowFunc()
	}
	if opts.Lease <= 0 {
		opts.Lease = mq.DefaultLeaseDuration
	}

	candidates := make([]*memRecord, 0)
	for _, rec := range m.tasks {
		if rec.status != mq.TaskStatusQueued {
			continue
		}
		if !rec.task.ExpiresAt.IsZero() && !rec.task.ExpiresAt.After(opts.Now) {
			rec.status = mq.TaskStatusExpired
			continue
		}
		if rec.task.AvailableAt.After(opts.Now) {
			continue
		}
		if len(opts.Topics) > 0 && !topicAllowed(rec.task.Topic, opts.Topics) {
			continue
		}
		candidates = append(candidates, rec)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].task.Priority == candidates[j].task.Priority {
			return candidates[i].task.CreatedAt.Before(candidates[j].task.CreatedAt)
		}
		return candidates[i].task.Priority > candidates[j].task.Priority
	})

	limit := opts.Limit
	if limit > len(candidates) {
		limit = len(candidates)
	}

	selected := make([]mq.Task, 0, limit)
	for i := 0; i < limit; i++ {
		rec := candidates[i]
		rec.status = mq.TaskStatusLeased
		rec.task.LeaseOwner = opts.ConsumerID
		rec.task.LeaseUntil = opts.Now.Add(opts.Lease)
		rec.task.Attempts++
		rec.task.UpdatedAt = opts.Now
		selected = append(selected, rec.task)
	}

	return selected, nil
}

func (m *MemStore) Ack(_ context.Context, taskID, consumerID string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.tasks[taskID]
	if !ok {
		return mq.ErrTaskNotFound
	}
	if rec.status != mq.TaskStatusLeased || rec.task.LeaseOwner != consumerID || (!rec.task.LeaseUntil.IsZero() && rec.task.LeaseUntil.Before(now)) {
		return mq.ErrLeaseLost
	}

	rec.status = mq.TaskStatusDone
	rec.task.LeaseOwner = ""
	rec.task.LeaseUntil = time.Time{}
	rec.task.UpdatedAt = now
	return nil
}

func (m *MemStore) Release(_ context.Context, taskID, consumerID string, opts mq.ReleaseOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.tasks[taskID]
	if !ok {
		return mq.ErrTaskNotFound
	}
	if rec.status != mq.TaskStatusLeased || rec.task.LeaseOwner != consumerID || (!rec.task.LeaseUntil.IsZero() && rec.task.LeaseUntil.Before(opts.Now)) {
		return mq.ErrLeaseLost
	}

	rec.status = mq.TaskStatusQueued
	rec.task.LeaseOwner = ""
	rec.task.LeaseUntil = time.Time{}
	rec.task.AvailableAt = opts.RetryAt
	rec.task.UpdatedAt = opts.Now
	rec.lastError = opts.Reason
	rec.lastErrorAt = opts.Now
	return nil
}

func (m *MemStore) MoveToDLQ(_ context.Context, taskID, consumerID, reason string, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.tasks[taskID]
	if !ok {
		return mq.ErrTaskNotFound
	}
	if rec.status != mq.TaskStatusLeased || rec.task.LeaseOwner != consumerID || (!rec.task.LeaseUntil.IsZero() && rec.task.LeaseUntil.Before(now)) {
		return mq.ErrLeaseLost
	}

	rec.status = mq.TaskStatusDead
	rec.task.LeaseOwner = ""
	rec.task.LeaseUntil = time.Time{}
	rec.task.UpdatedAt = now
	rec.lastError = reason
	rec.lastErrorAt = now
	return nil
}

func (m *MemStore) ExtendLease(_ context.Context, taskID, consumerID string, lease time.Duration, now time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.tasks[taskID]
	if !ok {
		return mq.ErrTaskNotFound
	}
	if rec.status != mq.TaskStatusLeased || rec.task.LeaseOwner != consumerID || (!rec.task.LeaseUntil.IsZero() && rec.task.LeaseUntil.Before(now)) {
		return mq.ErrLeaseLost
	}

	rec.task.LeaseUntil = now.Add(lease)
	rec.task.UpdatedAt = now
	return nil
}

func (m *MemStore) Stats(_ context.Context) (mq.Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stats mq.Stats
	for _, rec := range m.tasks {
		switch rec.status {
		case mq.TaskStatusQueued:
			stats.Queued++
		case mq.TaskStatusLeased:
			stats.Leased++
		case mq.TaskStatusDead:
			stats.Dead++
		case mq.TaskStatusExpired:
			stats.Expired++
		}
	}
	return stats, nil
}

func dedupeKey(tenantID, dedupe string) string {
	return tenantID + ":" + dedupe
}

func topicAllowed(topic string, topics []string) bool {
	for _, t := range topics {
		if t == topic {
			return true
		}
	}
	return false
}
