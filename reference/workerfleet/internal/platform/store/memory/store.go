package memory

import (
	"sort"
	"sync"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

type Store struct {
	mu                sync.RWMutex
	snapshots         map[domain.WorkerID]domain.WorkerSnapshot
	activeTasksByWork map[domain.WorkerID][]domain.ActiveTask
	activeTaskIndex   map[domain.TaskID]domain.WorkerID
	taskHistory       map[domain.TaskID][]platformstore.TaskHistoryRecord
	workerEvents      []domain.DomainEvent
	alerts            []platformstore.AlertRecord
}

func NewStore() *Store {
	return &Store{
		snapshots:         make(map[domain.WorkerID]domain.WorkerSnapshot),
		activeTasksByWork: make(map[domain.WorkerID][]domain.ActiveTask),
		activeTaskIndex:   make(map[domain.TaskID]domain.WorkerID),
		taskHistory:       make(map[domain.TaskID][]platformstore.TaskHistoryRecord),
		workerEvents:      make([]domain.DomainEvent, 0, 64),
		alerts:            make([]platformstore.AlertRecord, 0, 64),
	}
}

func (s *Store) UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	workerID := snapshot.Identity.WorkerID
	if workerID == "" {
		return nil
	}

	cloned := cloneSnapshot(snapshot)
	s.replaceActiveTasksLocked(workerID, cloned.ActiveTasks)
	cloned.ActiveTasks = cloneTasks(s.activeTasksByWork[workerID])
	cloned.ActiveTaskCount = len(cloned.ActiveTasks)
	s.snapshots[workerID] = cloned
	return nil
}

func (s *Store) GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.snapshots[workerID]
	if !ok {
		return domain.WorkerSnapshot{}, false, nil
	}
	return cloneSnapshot(snapshot), true, nil
}

func (s *Store) ListWorkerSnapshots(filter platformstore.WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.WorkerSnapshot, 0, len(s.snapshots))
	for _, snapshot := range s.snapshots {
		if !matchesSnapshotFilter(snapshot, filter) {
			continue
		}
		out = append(out, cloneSnapshot(snapshot))
	}

	sort.Slice(out, func(i, j int) bool {
		return out[i].Identity.WorkerID < out[j].Identity.WorkerID
	})
	return out, nil
}

func (s *Store) ListCurrentWorkerSnapshots() ([]domain.WorkerSnapshot, error) {
	return s.ListWorkerSnapshots(platformstore.WorkerSnapshotFilter{})
}

func (s *Store) FleetCounts() platformstore.FleetCounts {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := platformstore.FleetCounts{TotalWorkers: len(s.snapshots)}
	for _, snapshot := range s.snapshots {
		switch snapshot.Status {
		case domain.WorkerStatusOnline:
			counts.OnlineWorkers++
		case domain.WorkerStatusDegraded:
			counts.DegradedWorkers++
		case domain.WorkerStatusOffline:
			counts.OfflineWorkers++
		default:
			counts.UnknownWorkers++
		}
		if snapshot.Runtime.AcceptingTasks {
			counts.AcceptingWorkers++
		}
		if len(snapshot.ActiveTasks) > 0 {
			counts.BusyWorkers++
		}
		counts.ActiveTaskCount += len(snapshot.ActiveTasks)
	}
	return counts
}

func (s *Store) ReplaceActiveTasks(workerID domain.WorkerID, tasks []domain.ActiveTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.replaceActiveTasksLocked(workerID, tasks)
	if snapshot, ok := s.snapshots[workerID]; ok {
		snapshot.ActiveTasks = cloneTasks(s.activeTasksByWork[workerID])
		snapshot.ActiveTaskCount = len(snapshot.ActiveTasks)
		s.snapshots[workerID] = snapshot
	}
	return nil
}

func (s *Store) ActiveTasks(workerID domain.WorkerID) ([]domain.ActiveTask, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, ok := s.activeTasksByWork[workerID]
	if !ok {
		return nil, false, nil
	}
	return cloneTasks(tasks), true, nil
}

func (s *Store) GetTask(taskID domain.TaskID) (platformstore.CurrentTaskRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workerID, ok := s.activeTaskIndex[taskID]
	if !ok {
		return platformstore.CurrentTaskRecord{}, false, nil
	}
	for _, task := range s.activeTasksByWork[workerID] {
		if task.TaskID == taskID {
			return platformstore.CurrentTaskRecord{
				WorkerID: workerID,
				Task:     cloneTask(task),
			}, true, nil
		}
	}
	return platformstore.CurrentTaskRecord{}, false, nil
}

func (s *Store) AppendTaskHistory(record platformstore.TaskHistoryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := cloneTaskHistoryRecord(record)
	s.taskHistory[record.TaskID] = append(s.taskHistory[record.TaskID], cloned)
	sort.Slice(s.taskHistory[record.TaskID], func(i, j int) bool {
		return s.taskHistory[record.TaskID][i].LastUpdatedAt.Before(s.taskHistory[record.TaskID][j].LastUpdatedAt)
	})
	return nil
}

func (s *Store) TaskHistory(taskID domain.TaskID) ([]platformstore.TaskHistoryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.taskHistory[taskID]
	out := make([]platformstore.TaskHistoryRecord, 0, len(records))
	for _, record := range records {
		out = append(out, cloneTaskHistoryRecord(record))
	}
	return out, nil
}

func (s *Store) LatestTask(taskID domain.TaskID) (platformstore.TaskHistoryRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.taskHistory[taskID]
	if len(records) == 0 {
		return platformstore.TaskHistoryRecord{}, false, nil
	}
	return cloneTaskHistoryRecord(records[len(records)-1]), true, nil
}

func (s *Store) AppendWorkerEvent(event domain.DomainEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := event
	cloned.Attributes = cloneStringMap(event.Attributes)
	s.workerEvents = append(s.workerEvents, cloned)
	sort.Slice(s.workerEvents, func(i, j int) bool {
		return s.workerEvents[i].OccurredAt.Before(s.workerEvents[j].OccurredAt)
	})
	return nil
}

func (s *Store) ListWorkerEvents(workerID domain.WorkerID) ([]domain.DomainEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]domain.DomainEvent, 0, len(s.workerEvents))
	for _, event := range s.workerEvents {
		if workerID != "" && event.WorkerID != workerID {
			continue
		}
		cloned := event
		cloned.Attributes = cloneStringMap(event.Attributes)
		out = append(out, cloned)
	}
	return out, nil
}

func (s *Store) AppendAlert(record platformstore.AlertRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := cloneAlertRecord(record)
	s.alerts = append(s.alerts, cloned)
	sort.Slice(s.alerts, func(i, j int) bool {
		return s.alerts[i].TriggeredAt.Before(s.alerts[j].TriggeredAt)
	})
	return nil
}

func (s *Store) ListAlerts(filter platformstore.AlertFilter) ([]platformstore.AlertRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]platformstore.AlertRecord, 0, len(s.alerts))
	for _, alert := range s.alerts {
		if filter.WorkerID != "" && alert.WorkerID != filter.WorkerID {
			continue
		}
		if filter.AlertType != "" && alert.AlertType != filter.AlertType {
			continue
		}
		if filter.Status != "" && alert.Status != filter.Status {
			continue
		}
		out = append(out, cloneAlertRecord(alert))
	}
	return out, nil
}

func (s *Store) ListAlertRecords() ([]platformstore.AlertRecord, error) {
	return s.ListAlerts(platformstore.AlertFilter{})
}

func (s *Store) ApplyRetention(now time.Time, retention time.Duration) platformstore.RetentionResult {
	if retention <= 0 {
		retention = platformstore.DefaultRetention
	}
	cutoff := now.Add(-retention)

	s.mu.Lock()
	defer s.mu.Unlock()

	result := platformstore.RetentionResult{}
	for taskID, records := range s.taskHistory {
		filtered := records[:0]
		for _, record := range records {
			if latestTaskRecordTime(record).Before(cutoff) {
				result.TaskHistoryPruned++
				continue
			}
			filtered = append(filtered, record)
		}
		if len(filtered) == 0 {
			delete(s.taskHistory, taskID)
			continue
		}
		s.taskHistory[taskID] = filtered
	}

	filteredEvents := s.workerEvents[:0]
	for _, event := range s.workerEvents {
		if event.OccurredAt.Before(cutoff) {
			result.WorkerEventsPruned++
			continue
		}
		filteredEvents = append(filteredEvents, event)
	}
	s.workerEvents = filteredEvents

	filteredAlerts := s.alerts[:0]
	for _, alert := range s.alerts {
		if latestAlertTime(alert).Before(cutoff) {
			result.AlertsPruned++
			continue
		}
		filteredAlerts = append(filteredAlerts, alert)
	}
	s.alerts = filteredAlerts

	return result
}

func (s *Store) replaceActiveTasksLocked(workerID domain.WorkerID, tasks []domain.ActiveTask) {
	for taskID, owner := range s.activeTaskIndex {
		if owner == workerID {
			delete(s.activeTaskIndex, taskID)
		}
	}

	cloned := cloneTasks(tasks)
	sort.Slice(cloned, func(i, j int) bool { return cloned[i].TaskID < cloned[j].TaskID })
	s.activeTasksByWork[workerID] = cloned
	for _, task := range cloned {
		if task.TaskID == "" {
			continue
		}
		s.activeTaskIndex[task.TaskID] = workerID
	}
}

func matchesSnapshotFilter(snapshot domain.WorkerSnapshot, filter platformstore.WorkerSnapshotFilter) bool {
	if filter.Status != "" && snapshot.Status != filter.Status {
		return false
	}
	if filter.Namespace != "" && snapshot.Identity.Namespace != filter.Namespace {
		return false
	}
	if filter.NodeName != "" && snapshot.Identity.NodeName != filter.NodeName {
		return false
	}
	if filter.AcceptingTasks != nil && snapshot.Runtime.AcceptingTasks != *filter.AcceptingTasks {
		return false
	}
	if filter.TaskType != "" && !hasTaskType(snapshot.ActiveTasks, filter.TaskType) {
		return false
	}
	return true
}

func hasTaskType(tasks []domain.ActiveTask, taskType string) bool {
	for _, task := range tasks {
		if task.TaskType == taskType {
			return true
		}
	}
	return false
}

func cloneSnapshot(snapshot domain.WorkerSnapshot) domain.WorkerSnapshot {
	snapshot.ActiveTasks = cloneTasks(snapshot.ActiveTasks)
	return snapshot
}

func cloneTasks(tasks []domain.ActiveTask) []domain.ActiveTask {
	if len(tasks) == 0 {
		return nil
	}
	out := make([]domain.ActiveTask, 0, len(tasks))
	for _, task := range tasks {
		out = append(out, cloneTask(task))
	}
	return out
}

func cloneTask(task domain.ActiveTask) domain.ActiveTask {
	task.Metadata = cloneStringMap(task.Metadata)
	return task
}

func cloneTaskHistoryRecord(record platformstore.TaskHistoryRecord) platformstore.TaskHistoryRecord {
	record.Metadata = cloneStringMap(record.Metadata)
	return record
}

func cloneAlertRecord(record platformstore.AlertRecord) platformstore.AlertRecord {
	record.Details = cloneStringMap(record.Details)
	return record
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func latestTaskRecordTime(record platformstore.TaskHistoryRecord) time.Time {
	if !record.EndedAt.IsZero() {
		return record.EndedAt
	}
	if !record.LastUpdatedAt.IsZero() {
		return record.LastUpdatedAt
	}
	return record.StartedAt
}

func latestAlertTime(alert platformstore.AlertRecord) time.Time {
	if !alert.ResolvedAt.IsZero() {
		return alert.ResolvedAt
	}
	return alert.TriggeredAt
}

var _ platformstore.QueryStore = (*Store)(nil)
var _ platformstore.WorkerEventStore = (*Store)(nil)
var _ platformstore.RetentionStore = (*Store)(nil)
var _ domain.SnapshotStore = (*Store)(nil)
var _ domain.TaskHistoryStore = (*Store)(nil)
var _ domain.WorkerEventStore = (*Store)(nil)
