package handler

import (
	"context"
	"sync"
	"time"
)

// OperationInfo tracks an active non-SQL datasource operation.
type OperationInfo struct {
	OperationID string
	Driver      string
	Kind        string
	ConnID      string
	Resource    string
	Summary     string
	StartTime   time.Time
	CancelFunc  context.CancelFunc
}

// OperationRegistry manages active non-SQL operations and cancellation.
type OperationRegistry struct {
	mu         sync.RWMutex
	operations map[string]*OperationInfo
}

// NewOperationRegistry creates an active operation registry.
func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{
		operations: make(map[string]*OperationInfo),
	}
}

// Register adds an operation and returns the operation ID.
func (r *OperationRegistry) Register(info OperationInfo, cancel context.CancelFunc) string {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := info.OperationID
	if id == "" {
		id, _ = generateHistoryID()
	}
	info.OperationID = id
	info.CancelFunc = cancel
	info.StartTime = time.Now().UTC()
	r.operations[id] = &info
	return id
}

// Unregister removes an operation.
func (r *OperationRegistry) Unregister(operationID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.operations, operationID)
}

// Cancel cancels an operation by ID and returns true if it was active.
func (r *OperationRegistry) Cancel(operationID string) bool {
	r.mu.RLock()
	info, ok := r.operations[operationID]
	r.mu.RUnlock()
	if !ok {
		return false
	}
	info.CancelFunc()
	return true
}

// ListActive returns all active operations.
func (r *OperationRegistry) ListActive() []*OperationInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*OperationInfo, 0, len(r.operations))
	for _, info := range r.operations {
		result = append(result, info)
	}
	return result
}

// Count returns the number of active operations.
func (r *OperationRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.operations)
}
