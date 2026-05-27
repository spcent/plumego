package domain

import "time"

type ExecPlanID string

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

type CaseStepStatus string

const (
	CaseStepStatusUnknown   CaseStepStatus = "unknown"
	CaseStepStatusRunning   CaseStepStatus = "running"
	CaseStepStatusSucceeded CaseStepStatus = "succeeded"
	CaseStepStatusFailed    CaseStepStatus = "failed"
	CaseStepStatusCanceled  CaseStepStatus = "canceled"
	CaseStepStatusSkipped   CaseStepStatus = "skipped"
)

type CaseStepRuntime struct {
	Step       string
	StepName   string
	Status     CaseStepStatus
	StartedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt time.Time
	Attempt    int
	ErrorClass string
}

type ActiveTask struct {
	TaskID      TaskID
	ExecPlanID  ExecPlanID
	TaskType    string
	Phase       TaskPhase
	PhaseName   string
	CurrentStep CaseStepRuntime
	StartedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
}

type TaskReport struct {
	TaskID      TaskID
	ExecPlanID  ExecPlanID
	TaskType    string
	Phase       TaskPhase
	PhaseName   string
	CurrentStep CaseStepRuntime
	StartedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
}

type TaskHistoryRecord struct {
	TaskID        TaskID
	WorkerID      WorkerID
	ExecPlanID    ExecPlanID
	TaskType      string
	Phase         TaskPhase
	PhaseName     string
	CurrentStep   CaseStepRuntime
	Status        string
	StartedAt     time.Time
	EndedAt       time.Time
	LastUpdatedAt time.Time
	Metadata      map[string]string
}
