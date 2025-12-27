package webhookin

import (
	"sync"
	"time"
)

// Deduper is an in-memory TTL-based idempotency gate.
// It is stdlib-only and safe for concurrent use.
type Deduper struct {
	mu    sync.Mutex
	items map[string]time.Time // key -> expiresAt
	ttl   time.Duration

	// best-effort cleanup cadence; no background goroutine required
	lastSweep time.Time
}

func NewDeduper(ttl time.Duration) *Deduper {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	return &Deduper{
		items:     make(map[string]time.Time),
		ttl:       ttl,
		lastSweep: time.Now().UTC(),
	}
}

// SeenBefore marks key as seen and returns true if it was already present and not expired.
func (d *Deduper) SeenBefore(key string) bool {
	now := time.Now().UTC()

	d.mu.Lock()
	defer d.mu.Unlock()

	// occasional sweep to avoid unbounded growth
	if now.Sub(d.lastSweep) > 1*time.Minute {
		for k, exp := range d.items {
			if now.After(exp) {
				delete(d.items, k)
			}
		}
		d.lastSweep = now
	}

	if exp, ok := d.items[key]; ok && now.Before(exp) {
		return true
	}
	d.items[key] = now.Add(d.ttl)
	return false
}
