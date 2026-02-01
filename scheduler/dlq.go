package scheduler

import (
	"sync"
	"time"
)

// DeadLetterEntry represents a job that failed after all retries.
type DeadLetterEntry struct {
	JobID       JobID
	Error       error
	Attempts    int
	FirstFailed time.Time
	LastFailed  time.Time
	TaskName    string
	Group       string
	Tags        []string
}

// DeadLetterQueue manages failed jobs for inspection and requeuing.
type DeadLetterQueue struct {
	mu      sync.RWMutex
	entries map[JobID]*DeadLetterEntry
	maxSize int
}

// NewDeadLetterQueue creates a new dead letter queue.
// maxSize limits the number of entries (0 = unlimited).
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		entries: make(map[JobID]*DeadLetterEntry),
		maxSize: maxSize,
	}
}

// Add adds a failed job to the dead letter queue.
func (d *DeadLetterQueue) Add(entry DeadLetterEntry) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if entry already exists
	if existing, exists := d.entries[entry.JobID]; exists {
		// Update existing entry
		existing.Error = entry.Error
		existing.Attempts = entry.Attempts
		existing.LastFailed = entry.LastFailed
		return
	}

	// Check size limit
	if d.maxSize > 0 && len(d.entries) >= d.maxSize {
		// Remove oldest entry (simple FIFO eviction)
		var oldestID JobID
		var oldestTime time.Time
		for id, e := range d.entries {
			if oldestTime.IsZero() || e.FirstFailed.Before(oldestTime) {
				oldestID = id
				oldestTime = e.FirstFailed
			}
		}
		delete(d.entries, oldestID)
	}

	// Add new entry
	d.entries[entry.JobID] = &entry
}

// Get retrieves a dead letter entry by job ID.
func (d *DeadLetterQueue) Get(jobID JobID) (*DeadLetterEntry, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	entry, ok := d.entries[jobID]
	if !ok {
		return nil, false
	}
	// Return a copy to prevent external modification
	entryCopy := *entry
	return &entryCopy, true
}

// List returns all dead letter entries.
func (d *DeadLetterQueue) List() []DeadLetterEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]DeadLetterEntry, 0, len(d.entries))
	for _, entry := range d.entries {
		result = append(result, *entry)
	}
	return result
}

// Delete removes a dead letter entry.
func (d *DeadLetterQueue) Delete(jobID JobID) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, exists := d.entries[jobID]; exists {
		delete(d.entries, jobID)
		return true
	}
	return false
}

// Clear removes all dead letter entries.
func (d *DeadLetterQueue) Clear() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	count := len(d.entries)
	d.entries = make(map[JobID]*DeadLetterEntry)
	return count
}

// Size returns the current number of entries in the queue.
func (d *DeadLetterQueue) Size() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.entries)
}
