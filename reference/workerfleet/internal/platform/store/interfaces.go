package store

import "workerfleet/internal/domain"

type WorkerSnapshotStore interface {
	UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error
	GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error)
	ListWorkerSnapshots(filter WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error)
	ListCurrentWorkerSnapshots() ([]domain.WorkerSnapshot, error)
	FleetCounts() FleetCounts
}

type ActiveTaskStore interface {
	ReplaceActiveTasks(workerID domain.WorkerID, tasks []domain.ActiveTask) error
	ActiveTasks(workerID domain.WorkerID) ([]domain.ActiveTask, bool, error)
	GetTask(taskID domain.TaskID) (CurrentTaskRecord, bool, error)
}

type WorkerEventStore interface {
	AppendWorkerEvent(event domain.DomainEvent) error
	ListWorkerEvents(workerID domain.WorkerID) ([]domain.DomainEvent, error)
}

type TaskHistoryStore interface {
	AppendTaskHistory(record TaskHistoryRecord) error
	TaskHistory(taskID domain.TaskID) ([]TaskHistoryRecord, error)
	LatestTask(taskID domain.TaskID) (TaskHistoryRecord, bool, error)
}

type AlertStore interface {
	AppendAlert(record AlertRecord) error
	ListAlerts(filter AlertFilter) ([]AlertRecord, error)
	ListAlertRecords() ([]AlertRecord, error)
}

type QueryStore interface {
	WorkerSnapshotStore
	ActiveTaskStore
	TaskHistoryStore
	AlertStore
}
