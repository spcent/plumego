package store

import (
	"sort"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type TaskHistoryRecord = domain.TaskHistoryRecord

type WorkerEventStore interface {
	AppendWorkerEvent(event domain.DomainEvent) error
	ListWorkerEvents(workerID domain.WorkerID) ([]domain.DomainEvent, error)
}

type TaskHistoryStore interface {
	AppendTaskHistory(record TaskHistoryRecord) error
	TaskHistory(taskID domain.TaskID) ([]TaskHistoryRecord, error)
	LatestTask(taskID domain.TaskID) (TaskHistoryRecord, bool, error)
}

func (s *MemoryStore) AppendTaskHistory(record TaskHistoryRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cloned := cloneTaskHistoryRecord(record)
	s.taskHistory[record.TaskID] = append(s.taskHistory[record.TaskID], cloned)
	sort.Slice(s.taskHistory[record.TaskID], func(i, j int) bool {
		return s.taskHistory[record.TaskID][i].LastUpdatedAt.Before(s.taskHistory[record.TaskID][j].LastUpdatedAt)
	})
	return nil
}

func (s *MemoryStore) TaskHistory(taskID domain.TaskID) ([]TaskHistoryRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.taskHistory[taskID]
	out := make([]TaskHistoryRecord, 0, len(records))
	for _, record := range records {
		out = append(out, cloneTaskHistoryRecord(record))
	}
	return out, nil
}

func (s *MemoryStore) LatestTask(taskID domain.TaskID) (TaskHistoryRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	records := s.taskHistory[taskID]
	if len(records) == 0 {
		return TaskHistoryRecord{}, false, nil
	}
	return cloneTaskHistoryRecord(records[len(records)-1]), true, nil
}

func (s *MemoryStore) AppendWorkerEvent(event domain.DomainEvent) error {
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

func (s *MemoryStore) ListWorkerEvents(workerID domain.WorkerID) ([]domain.DomainEvent, error) {
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

func cloneTaskHistoryRecord(record TaskHistoryRecord) TaskHistoryRecord {
	record.Metadata = cloneStringMap(record.Metadata)
	return record
}

var _ TaskHistoryStore = (*MemoryStore)(nil)
var _ WorkerEventStore = (*MemoryStore)(nil)
