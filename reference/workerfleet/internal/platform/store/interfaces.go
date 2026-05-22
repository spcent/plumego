package store

import (
	"context"

	"workerfleet/internal/domain"
)

type WorkerSnapshotStore interface {
	UpsertWorkerSnapshot(ctx context.Context, snapshot domain.WorkerSnapshot) error
	GetWorkerSnapshot(ctx context.Context, workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error)
	ListWorkerSnapshots(ctx context.Context, filter WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error)
	ListCurrentWorkerSnapshots(ctx context.Context) ([]domain.WorkerSnapshot, error)
	FleetCounts(ctx context.Context) (FleetCounts, error)
}

type ActiveTaskStore interface {
	ReplaceActiveTasks(ctx context.Context, workerID domain.WorkerID, tasks []domain.ActiveTask) error
	ActiveTasks(ctx context.Context, workerID domain.WorkerID) ([]domain.ActiveTask, bool, error)
	GetTask(ctx context.Context, taskID domain.TaskID) (CurrentTaskRecord, bool, error)
}

type WorkerEventStore interface {
	AppendWorkerEvent(ctx context.Context, event domain.DomainEvent) error
	ListWorkerEvents(ctx context.Context, workerID domain.WorkerID) ([]domain.DomainEvent, error)
}

type TaskHistoryStore interface {
	AppendTaskHistory(ctx context.Context, record TaskHistoryRecord) error
	TaskHistory(ctx context.Context, taskID domain.TaskID) ([]TaskHistoryRecord, error)
	LatestTask(ctx context.Context, taskID domain.TaskID) (TaskHistoryRecord, bool, error)
}

type CaseStepHistoryStore interface {
	AppendCaseStepHistory(ctx context.Context, record CaseStepHistoryRecord) error
	CaseStepHistory(ctx context.Context, taskID domain.TaskID) ([]CaseStepHistoryRecord, error)
	ListCaseStepHistory(ctx context.Context, filter CaseStepHistoryFilter) ([]CaseStepHistoryRecord, error)
}

type AlertStore interface {
	AppendAlert(ctx context.Context, record AlertRecord) error
	ListAlerts(ctx context.Context, filter AlertFilter) ([]AlertRecord, error)
	ListAlertRecords(ctx context.Context) ([]AlertRecord, error)
}

type QueryStore interface {
	WorkerSnapshotStore
	ActiveTaskStore
	TaskHistoryStore
	AlertStore
}
