package store

import (
	"sort"

	"github.com/spcent/plumego/reference/workerfleet/internal/domain"
)

type ActiveTaskStore interface {
	ReplaceActiveTasks(workerID domain.WorkerID, tasks []domain.ActiveTask) error
	ActiveTasks(workerID domain.WorkerID) ([]domain.ActiveTask, bool, error)
	GetTask(taskID domain.TaskID) (CurrentTaskRecord, bool, error)
}

type CurrentTaskRecord struct {
	WorkerID domain.WorkerID
	Task     domain.ActiveTask
}

func (s *MemoryStore) ReplaceActiveTasks(workerID domain.WorkerID, tasks []domain.ActiveTask) error {
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

func (s *MemoryStore) ActiveTasks(workerID domain.WorkerID) ([]domain.ActiveTask, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks, ok := s.activeTasksByWork[workerID]
	if !ok {
		return nil, false, nil
	}
	return cloneTasks(tasks), true, nil
}

func (s *MemoryStore) GetTask(taskID domain.TaskID) (CurrentTaskRecord, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workerID, ok := s.activeTaskIndex[taskID]
	if !ok {
		return CurrentTaskRecord{}, false, nil
	}
	for _, task := range s.activeTasksByWork[workerID] {
		if task.TaskID == taskID {
			return CurrentTaskRecord{
				WorkerID: workerID,
				Task:     cloneTask(task),
			}, true, nil
		}
	}
	return CurrentTaskRecord{}, false, nil
}

func (s *MemoryStore) replaceActiveTasksLocked(workerID domain.WorkerID, tasks []domain.ActiveTask) {
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

var _ ActiveTaskStore = (*MemoryStore)(nil)
