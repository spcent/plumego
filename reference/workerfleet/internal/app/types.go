package app

import (
	"context"
	"errors"
	"time"

	"workerfleet/internal/domain"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
	"workerfleet/internal/platform/notifier"
	platformstore "workerfleet/internal/platform/store"
)

var (
	ErrNotFound       = errors.New("workerfleet resource not found")
	ErrNotImplemented = errors.New("workerfleet operation not implemented")
	ErrConflict       = errors.New("workerfleet conflict")
)

type Runtime struct {
	Service *Service
	Metrics *workerfleetmetrics.Collector
	Close   func(context.Context) error
	Ready   func(context.Context) error
	shell   runtimeShell
}

type runtimeShell struct {
	ingest *domain.IngestService
	query  *Service
	loops  *LoopRunner
	alerts *AlertRunner
}

type RuntimeErrorObserver interface {
	ObserveRuntimeError(operation string, err error)
}

type runtimeStore interface {
	platformstore.QueryStore
	platformstore.NotificationOutboxStore
	platformstore.WorkerEventStore
	domain.SnapshotStore
	domain.TaskHistoryStore
	domain.WorkerEventStore
}

type LoopRunner struct {
	store             runtimeStore
	policy            domain.StatusPolicy
	metrics           *workerfleetmetrics.Observer
	errors            RuntimeErrorObserver
	lease             LoopLeaseCoordinator
	inventorySyncerFn func(Config) (inventorySyncer, error)
}

type AlertRunner struct {
	store         runtimeStore
	policy        domain.StatusPolicy
	alertPolicy   domain.AlertPolicy
	metrics       *workerfleetmetrics.Observer
	errors        RuntimeErrorObserver
	lease         LoopLeaseCoordinator
	dispatcherFn  func(Config) alertDispatcher
	engineFactory func() domainAlertEngine
}

type loopExecutionSettings struct {
	Name              string
	Interval          time.Duration
	Timeout           time.Duration
	FailureBackoff    time.Duration
	MaxFailureBackoff time.Duration
	Lease             LoopLeaseCoordinator
}

type loopExecutionResult struct {
	err      error
	timedOut bool
}

type LoopLeaseCoordinator interface {
	TryAcquire(ctx context.Context, loopName string) (release func(), acquired bool, err error)
}

type nopLoopLease struct{}

func (nopLoopLease) TryAcquire(context.Context, string) (func(), bool, error) {
	return nil, true, nil
}

type inventorySyncer interface {
	SyncOnce(ctx context.Context) (resourceVersion string, err error)
}

type alertDispatcher interface {
	Notify(ctx context.Context, alert domain.AlertRecord) error
	Bindings() []notifier.SinkBinding
}

type domainAlertEngine interface {
	Evaluate(ctx context.Context) ([]domain.AlertRecord, error)
}

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
	TaskID      string            `json:"task_id"`
	ExecPlanID  string            `json:"exec_plan_id,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	PhaseName   string            `json:"phase_name,omitempty"`
	CurrentStep *StepView         `json:"current_step,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type StepView struct {
	Step       string                `json:"step,omitempty"`
	StepName   string                `json:"step_name,omitempty"`
	Status     domain.CaseStepStatus `json:"status,omitempty"`
	StartedAt  time.Time             `json:"started_at,omitempty"`
	UpdatedAt  time.Time             `json:"updated_at,omitempty"`
	FinishedAt time.Time             `json:"finished_at,omitempty"`
	Attempt    int                   `json:"attempt,omitempty"`
	ErrorClass string                `json:"error_class,omitempty"`
}

type WorkerView struct {
	WorkerID        string     `json:"worker_id"`
	Namespace       string     `json:"namespace,omitempty"`
	PodName         string     `json:"pod_name,omitempty"`
	NodeName        string     `json:"node_name,omitempty"`
	ContainerName   string     `json:"container_name,omitempty"`
	Image           string     `json:"image,omitempty"`
	Version         string     `json:"version,omitempty"`
	Status          string     `json:"status"`
	StatusReason    string     `json:"status_reason,omitempty"`
	ProcessAlive    bool       `json:"process_alive"`
	AcceptingTasks  bool       `json:"accepting_tasks"`
	LastSeenAt      time.Time  `json:"last_seen_at,omitempty"`
	LastReadyAt     time.Time  `json:"last_ready_at,omitempty"`
	ActiveTaskCount int        `json:"active_task_count"`
	ActiveTasks     []TaskView `json:"active_tasks,omitempty"`
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
	Items    []WorkerView `json:"items"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
	Total    int          `json:"total"`
}

type TaskDetail struct {
	TaskID      string            `json:"task_id"`
	WorkerID    string            `json:"worker_id,omitempty"`
	ExecPlanID  string            `json:"exec_plan_id,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Phase       string            `json:"phase,omitempty"`
	PhaseName   string            `json:"phase_name,omitempty"`
	CurrentStep *StepView         `json:"current_step,omitempty"`
	Status      string            `json:"status,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	EndedAt     time.Time         `json:"ended_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type CaseStepView struct {
	TaskID     string                `json:"task_id"`
	WorkerID   string                `json:"worker_id,omitempty"`
	ExecPlanID string                `json:"exec_plan_id,omitempty"`
	Namespace  string                `json:"namespace,omitempty"`
	PodName    string                `json:"pod_name,omitempty"`
	NodeName   string                `json:"node_name,omitempty"`
	Step       string                `json:"step"`
	StepName   string                `json:"step_name,omitempty"`
	Status     domain.CaseStepStatus `json:"status,omitempty"`
	Result     string                `json:"result,omitempty"`
	ErrorClass string                `json:"error_class,omitempty"`
	Attempt    int                   `json:"attempt,omitempty"`
	StartedAt  time.Time             `json:"started_at,omitempty"`
	FinishedAt time.Time             `json:"finished_at,omitempty"`
	ObservedAt time.Time             `json:"observed_at,omitempty"`
	EventType  domain.EventType      `json:"event_type,omitempty"`
}

type CaseTimelineResult struct {
	TaskID string         `json:"task_id"`
	Items  []CaseStepView `json:"items"`
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
	ExecPlanID string         `json:"exec_plan_id"`
	Items      []CaseStepView `json:"items"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	Total      int            `json:"total"`
}

type FleetSummary struct {
	TotalWorkers     int `json:"total_workers"`
	OnlineWorkers    int `json:"online_workers"`
	DegradedWorkers  int `json:"degraded_workers"`
	OfflineWorkers   int `json:"offline_workers"`
	UnknownWorkers   int `json:"unknown_workers"`
	AcceptingWorkers int `json:"accepting_workers"`
	BusyWorkers      int `json:"busy_workers"`
	ActiveTaskCount  int `json:"active_task_count"`
}

type AlertListQuery struct {
	WorkerID  string
	AlertType string
	Status    string
	Page      int
	PageSize  int
}

type AlertView struct {
	AlertID     string    `json:"alert_id"`
	WorkerID    string    `json:"worker_id,omitempty"`
	TaskID      string    `json:"task_id,omitempty"`
	AlertType   string    `json:"alert_type"`
	Status      string    `json:"status"`
	Severity    string    `json:"severity,omitempty"`
	Message     string    `json:"message"`
	TriggeredAt time.Time `json:"triggered_at"`
	ResolvedAt  time.Time `json:"resolved_at,omitempty"`
}

type AlertListResult struct {
	Items    []AlertView `json:"items"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Total    int         `json:"total"`
}
