package app

import (
	"context"
	"errors"
	"time"

	"workerfleet/internal/domain"
	workerfleetmetrics "workerfleet/internal/platform/metrics"
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
