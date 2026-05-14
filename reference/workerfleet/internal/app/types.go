package app

import (
	"errors"
	"time"

	"workerfleet/internal/domain"
)

var (
	ErrNotFound       = errors.New("workerfleet resource not found")
	ErrNotImplemented = errors.New("workerfleet operation not implemented")
	ErrConflict       = errors.New("workerfleet conflict")
)

type RegisterWorkerInput struct {
	Identity   domain.WorkerIdentity
	ObservedAt time.Time
}

type RegisterWorkerResult struct {
	WorkerID     string
	Status       domain.WorkerStatus
	RegisteredAt time.Time
}

type HeartbeatWorkerInput struct {
	WorkerID       domain.WorkerID
	ProcessAlive   bool
	AcceptingTasks bool
	ObservedAt     time.Time
	LastError      string
	ActiveTasks    []domain.TaskReport
}

type HeartbeatWorkerResult struct {
	WorkerID        string
	Status          domain.WorkerStatus
	StatusReason    string
	ObservedAt      time.Time
	ActiveTaskCount int
}

type TaskView struct {
	TaskID      string
	ExecPlanID  string
	TaskType    string
	Phase       string
	PhaseName   string
	CurrentStep *StepView
	StartedAt   time.Time
	UpdatedAt   time.Time
	Metadata    map[string]string
}

type StepView struct {
	Step       string
	StepName   string
	Status     domain.CaseStepStatus
	StartedAt  time.Time
	UpdatedAt  time.Time
	FinishedAt time.Time
	Attempt    int
	ErrorClass string
}

type WorkerView struct {
	WorkerID        string
	Namespace       string
	PodName         string
	NodeName        string
	ContainerName   string
	Image           string
	Version         string
	Status          string
	StatusReason    string
	ProcessAlive    bool
	AcceptingTasks  bool
	LastSeenAt      time.Time
	LastReadyAt     time.Time
	ActiveTaskCount int
	ActiveTasks     []TaskView
}

type WorkerDetail = WorkerView

type WorkerListQuery struct {
	Status         domain.WorkerStatus
	Namespace      string
	NodeName       string
	TaskType       string
	AcceptingTasks *bool
	Page           int
	PageSize       int
}

type WorkerListResult struct {
	Items    []WorkerView
	Page     int
	PageSize int
	Total    int
}

type TaskDetail struct {
	TaskID      string
	WorkerID    string
	ExecPlanID  string
	TaskType    string
	Phase       string
	PhaseName   string
	CurrentStep *StepView
	Status      string
	StartedAt   time.Time
	UpdatedAt   time.Time
	EndedAt     time.Time
	Metadata    map[string]string
}

type CaseStepView struct {
	TaskID     string
	WorkerID   string
	ExecPlanID string
	Namespace  string
	PodName    string
	NodeName   string
	Step       string
	StepName   string
	Status     domain.CaseStepStatus
	Result     string
	ErrorClass string
	Attempt    int
	StartedAt  time.Time
	FinishedAt time.Time
	ObservedAt time.Time
	EventType  domain.EventType
}

type CaseTimelineResult struct {
	TaskID string
	Items  []CaseStepView
}

type ExecPlanCaseDrilldownQuery struct {
	ExecPlanID string
	NodeName   string
	PodName    string
	Step       string
	Page       int
	PageSize   int
}

type ExecPlanCaseDrilldownResult struct {
	ExecPlanID string
	Items      []CaseStepView
	Page       int
	PageSize   int
	Total      int
}

type FleetSummary struct {
	TotalWorkers     int
	OnlineWorkers    int
	DegradedWorkers  int
	OfflineWorkers   int
	UnknownWorkers   int
	AcceptingWorkers int
	BusyWorkers      int
	ActiveTaskCount  int
}

type AlertListQuery struct {
	WorkerID  string
	AlertType string
	Status    string
	Page      int
	PageSize  int
}

type AlertView struct {
	AlertID     string
	WorkerID    string
	TaskID      string
	AlertType   string
	Status      string
	Severity    string
	Message     string
	TriggeredAt time.Time
	ResolvedAt  time.Time
}

type AlertListResult struct {
	Items    []AlertView
	Page     int
	PageSize int
	Total    int
}
