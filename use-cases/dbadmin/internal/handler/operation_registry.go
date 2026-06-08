package handler

import (
	"context"
	"time"

	"dbadmin/internal/domain/connection"
)

// OperationKind categorises a non-SQL datasource operation.
type OperationKind string

const (
	OperationKindQuery     OperationKind = "query"
	OperationKindAggregate OperationKind = "aggregate"
	OperationKindExport    OperationKind = "export"
	OperationKindImport    OperationKind = "import"
	OperationKindCommand   OperationKind = "command"
	OperationKindSearch    OperationKind = "search"
	OperationKindDelete    OperationKind = "delete"
)

// OperationInfo tracks an active non-SQL datasource operation.
type OperationInfo struct {
	OperationID string
	Driver      connection.DriverType
	Kind        OperationKind
	ConnID      string
	Resource    string
	Summary     string
	StartTime   time.Time
	CancelFunc  context.CancelFunc
}

// OperationRegistry manages active non-SQL operations and cancellation.
type OperationRegistry struct {
	reg activeRegistry[OperationInfo]
}

// NewOperationRegistry creates an active operation registry.
func NewOperationRegistry() *OperationRegistry {
	return &OperationRegistry{reg: newActiveRegistry[OperationInfo]()}
}

// Register adds an operation and returns the operation ID.
func (r *OperationRegistry) Register(info OperationInfo, cancel context.CancelFunc) string {
	id := info.OperationID
	if id == "" {
		id, _ = generateHistoryID()
	}
	info.OperationID = id
	info.CancelFunc = cancel
	info.StartTime = time.Now().UTC()
	r.reg.put(id, &info)
	return id
}

// Unregister removes an operation.
func (r *OperationRegistry) Unregister(operationID string) {
	r.reg.remove(operationID)
}

// Cancel cancels an operation by ID and returns true if it was active.
func (r *OperationRegistry) Cancel(operationID string) bool {
	info, ok := r.reg.get(operationID)
	if !ok {
		return false
	}
	info.CancelFunc()
	return true
}

// ListActive returns all active operations.
func (r *OperationRegistry) ListActive() []*OperationInfo {
	return r.reg.list()
}

// Count returns the number of active operations.
func (r *OperationRegistry) Count() int {
	return r.reg.count()
}
