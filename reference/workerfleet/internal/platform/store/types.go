package store

import (
	"time"

	"workerfleet/internal/domain"
)

type WorkerSnapshotFilter struct {
	Status         domain.WorkerStatus
	Namespace      string
	NodeName       string
	TaskType       string
	AcceptingTasks *bool
}

type FleetCounts struct {
	TotalWorkers     int
	OnlineWorkers    int
	DegradedWorkers  int
	OfflineWorkers   int
	UnknownWorkers   int
	AcceptingWorkers int
	BusyWorkers      int
	ActiveTaskCount  int
}

type CurrentTaskRecord struct {
	WorkerID domain.WorkerID
	Task     domain.ActiveTask
}

type TaskHistoryRecord = domain.TaskHistoryRecord

type CaseStepHistoryFilter struct {
	TaskID     domain.TaskID
	WorkerID   domain.WorkerID
	ExecPlanID domain.ExecPlanID
	NodeName   string
	PodName    string
	Step       string
}

type CaseStepHistoryRecord struct {
	TaskID     domain.TaskID
	WorkerID   domain.WorkerID
	ExecPlanID domain.ExecPlanID
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

type AlertFilter struct {
	WorkerID  domain.WorkerID
	AlertType domain.AlertType
	Status    domain.AlertStatus
}

type AlertRecord = domain.AlertRecord
