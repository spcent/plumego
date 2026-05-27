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

type NotificationJobStatus string

const (
	NotificationJobPending    NotificationJobStatus = "pending"
	NotificationJobProcessing NotificationJobStatus = "processing"
	NotificationJobDelivered  NotificationJobStatus = "delivered"
	NotificationJobFailed     NotificationJobStatus = "failed"
)

type NotificationSinkType string

const (
	NotificationSinkFeishu  NotificationSinkType = "feishu"
	NotificationSinkWebhook NotificationSinkType = "webhook"
)

type NotificationJob struct {
	JobID          string
	AlertID        string
	SinkType       NotificationSinkType
	Alert          AlertRecord
	Status         NotificationJobStatus
	Attempts       int
	NextAttemptAt  time.Time
	LockedUntil    time.Time
	LastErrorClass string
	LastError      string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeliveredAt    time.Time
}

type NotificationFailure struct {
	ErrorClass    string
	ErrorMessage  string
	Permanent     bool
	NextAttemptAt time.Time
}
