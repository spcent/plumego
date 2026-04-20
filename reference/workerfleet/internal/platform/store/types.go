package store

import "workerfleet/internal/domain"

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

type AlertFilter struct {
	WorkerID  domain.WorkerID
	AlertType domain.AlertType
	Status    domain.AlertStatus
}

type AlertRecord = domain.AlertRecord
