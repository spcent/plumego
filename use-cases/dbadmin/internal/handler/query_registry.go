package handler

import (
	"context"
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
	reg activeRegistry[QueryInfo]
}

// NewQueryRegistry creates a new query registry.
func NewQueryRegistry() *QueryRegistry {
	return &QueryRegistry{reg: newActiveRegistry[QueryInfo]()}
}

// Register adds a query to the registry.
func (r *QueryRegistry) Register(queryID, connID, database, sql string, cancel context.CancelFunc) {
	r.reg.put(queryID, &QueryInfo{
		QueryID:    queryID,
		ConnID:     connID,
		Database:   database,
		SQL:        sql,
		StartTime:  time.Now(),
		CancelFunc: cancel,
	})
}

// Unregister removes a query from the registry.
func (r *QueryRegistry) Unregister(queryID string) {
	r.reg.remove(queryID)
}

// Cancel cancels a query by ID and returns true if the query was found and cancelled.
func (r *QueryRegistry) Cancel(queryID string) bool {
	info, ok := r.reg.get(queryID)
	if !ok {
		return false
	}
	info.CancelFunc()
	return true
}

// ListActive returns all active queries.
func (r *QueryRegistry) ListActive() []*QueryInfo {
	return r.reg.list()
}

// GetActive returns a specific active query.
func (r *QueryRegistry) GetActive(queryID string) *QueryInfo {
	info, _ := r.reg.get(queryID)
	return info
}

// Count returns the number of active queries.
func (r *QueryRegistry) Count() int {
	return r.reg.count()
}
