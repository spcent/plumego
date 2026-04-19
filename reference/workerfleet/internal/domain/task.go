package domain

import "time"

type TaskPhase string

const (
	TaskPhaseUnknown    TaskPhase = "unknown"
	TaskPhaseQueued     TaskPhase = "queued"
	TaskPhasePreparing  TaskPhase = "preparing"
	TaskPhaseRunning    TaskPhase = "running"
	TaskPhaseFinalizing TaskPhase = "finalizing"
	TaskPhaseSucceeded  TaskPhase = "succeeded"
	TaskPhaseFailed     TaskPhase = "failed"
	TaskPhaseCanceled   TaskPhase = "canceled"
)

type ActiveTask struct {
	TaskID    TaskID
	TaskType  string
	Phase     TaskPhase
	PhaseName string
	StartedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
}

type TaskReport struct {
	TaskID    TaskID
	TaskType  string
	Phase     TaskPhase
	PhaseName string
	StartedAt time.Time
	UpdatedAt time.Time
	Metadata  map[string]string
}
