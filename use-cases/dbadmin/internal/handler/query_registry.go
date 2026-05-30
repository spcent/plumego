package handler

import (
	"context"
	"sync"
	"time"
)

// QueryInfo tracks an active query execution.
type QueryInfo struct {
	QueryID    string
	ConnID     string
	Database   string
	SQL        string
	StartTime  time.Time
	CancelFunc context.CancelFunc
}

// QueryRegistry manages active queries and provides cancellation capability.
type QueryRegistry struct {
	mu      sync.RWMutex
	queries map[string]*QueryInfo
}

// NewQueryRegistry creates a new query registry.
func NewQueryRegistry() *QueryRegistry {
	return &QueryRegistry{
		queries: make(map[string]*QueryInfo),
	}
}

// Register adds a query to the registry and returns a context that can be cancelled.
func (r *QueryRegistry) Register(queryID, connID, database, sql string, cancel context.CancelFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.queries[queryID] = &QueryInfo{
		QueryID:    queryID,
		ConnID:     connID,
		Database:   database,
		SQL:        sql,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	}
}

// Unregister removes a query from the registry.
func (r *QueryRegistry) Unregister(queryID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.queries, queryID)
}

// Cancel cancels a query by ID and returns true if the query was found and cancelled.
func (r *QueryRegistry) Cancel(queryID string) bool {
	r.mu.RLock()
	info, exists := r.queries[queryID]
	r.mu.RUnlock()

	if !exists {
		return false
	}

	info.CancelFunc()
	return true
}

// ListActive returns all active queries.
func (r *QueryRegistry) ListActive() []*QueryInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*QueryInfo, 0, len(r.queries))
	for _, info := range r.queries {
		result = append(result, info)
	}
	return result
}

// GetActive returns a specific active query.
func (r *QueryRegistry) GetActive(queryID string) *QueryInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.queries[queryID]
}

// Count returns the number of active queries.
func (r *QueryRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.queries)
}
