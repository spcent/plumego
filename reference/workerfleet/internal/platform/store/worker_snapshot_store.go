package store

import (
	"sort"
	"sync"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type WorkerSnapshotFilter struct {
	Status         domain.WorkerStatus
	Namespace      string
	NodeName       string
	TaskType       string
	AcceptingTasks *bool
}

type WorkerSnapshotStore interface {
	UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error
	GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error)
	ListWorkerSnapshots(filter WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error)
	ListCurrentWorkerSnapshots() ([]domain.WorkerSnapshot, error)
	FleetCounts() FleetCounts
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

type MemoryStore struct {
	mu                sync.RWMutex
	snapshots         map[domain.WorkerID]domain.WorkerSnapshot
	activeTasksByWork map[domain.WorkerID][]domain.ActiveTask
	activeTaskIndex   map[domain.TaskID]domain.WorkerID
	taskHistory       map[domain.TaskID][]TaskHistoryRecord
	workerEvents      []domain.DomainEvent
	alerts            []AlertRecord
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		snapshots:         make(map[domain.WorkerID]domain.WorkerSnapshot),
		activeTasksByWork: make(map[domain.WorkerID][]domain.ActiveTask),
		activeTaskIndex:   make(map[domain.TaskID]domain.WorkerID),
		taskHistory:       make(map[domain.TaskID][]TaskHistoryRecord),
		workerEvents:      make([]domain.DomainEvent, 0, 64),
		alerts:            make([]AlertRecord, 0, 64),
	}
}

func (s *MemoryStore) UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error {
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

func (s *MemoryStore) GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot, ok := s.snapshots[workerID]
	if !ok {
		return domain.WorkerSnapshot{}, false, nil
	}
	return cloneSnapshot(snapshot), true, nil
}

func (s *MemoryStore) ListWorkerSnapshots(filter WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error) {
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

func (s *MemoryStore) FleetCounts() FleetCounts {
	s.mu.RLock()
	defer s.mu.RUnlock()

	counts := FleetCounts{TotalWorkers: len(s.snapshots)}
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

func (s *MemoryStore) ListCurrentWorkerSnapshots() ([]domain.WorkerSnapshot, error) {
	return s.ListWorkerSnapshots(WorkerSnapshotFilter{})
}

func matchesSnapshotFilter(snapshot domain.WorkerSnapshot, filter WorkerSnapshotFilter) bool {
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

var _ WorkerSnapshotStore = (*MemoryStore)(nil)
