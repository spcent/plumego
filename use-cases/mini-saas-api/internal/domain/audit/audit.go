// Package audit owns the append-only, tenant-scoped audit log.
//
// Entries record who did what to which resource. Detail is a short free-text
// summary; credentials, tokens, and request payloads must never be stored.
package audit

import (
	"context"
	"sync"
	"time"

	"mini-saas-api/internal/domain/ident"
)

// Entry is one immutable audit record.
type Entry struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	ActorID      string    `json:"actor_id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id"`
	Detail       string    `json:"detail,omitempty"`
	At           time.Time `json:"at"`
}

// Recorder is an in-memory, per-tenant ring of audit entries.
// When a tenant exceeds maxPerTenant entries the oldest are dropped.
type Recorder struct {
	mu           sync.RWMutex
	byTenant     map[string][]Entry
	maxPerTenant int
}

// NewRecorder returns a Recorder retaining up to maxPerTenant entries per
// tenant (minimum 1).
func NewRecorder(maxPerTenant int) *Recorder {
	if maxPerTenant < 1 {
		maxPerTenant = 1
	}
	return &Recorder{
		byTenant:     make(map[string][]Entry),
		maxPerTenant: maxPerTenant,
	}
}

// Record appends an entry, filling ID and At.
func (r *Recorder) Record(_ context.Context, e Entry) error {
	id, err := ident.New()
	if err != nil {
		return err
	}
	e.ID = id
	e.At = time.Now().UTC()

	r.mu.Lock()
	defer r.mu.Unlock()
	entries := append(r.byTenant[e.TenantID], e)
	if len(entries) > r.maxPerTenant {
		entries = entries[len(entries)-r.maxPerTenant:]
	}
	r.byTenant[e.TenantID] = entries
	return nil
}

// List returns up to limit entries for the tenant, newest first.
// limit <= 0 returns all retained entries.
func (r *Recorder) List(_ context.Context, tenantID string, limit int) ([]Entry, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	entries := r.byTenant[tenantID]
	n := len(entries)
	if limit <= 0 || limit > n {
		limit = n
	}
	out := make([]Entry, 0, limit)
	for i := n - 1; i >= n-limit; i-- {
		out = append(out, entries[i])
	}
	return out, nil
}
