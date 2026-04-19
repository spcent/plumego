package domain

import "time"

func BuildTaskHistoryRecords(previous WorkerSnapshot, current WorkerSnapshot, events []DomainEvent, now time.Time) []TaskHistoryRecord {
	currentByID := make(map[TaskID]ActiveTask, len(current.ActiveTasks))
	for _, task := range current.ActiveTasks {
		currentByID[task.TaskID] = task
	}

	previousByID := make(map[TaskID]ActiveTask, len(previous.ActiveTasks))
	for _, task := range previous.ActiveTasks {
		previousByID[task.TaskID] = task
	}

	records := make([]TaskHistoryRecord, 0, len(events))
	for _, event := range events {
		switch event.Type {
		case EventTaskStarted, EventTaskPhaseChanged:
			task, ok := currentByID[event.TaskID]
			if !ok {
				continue
			}
			records = append(records, TaskHistoryRecord{
				TaskID:        task.TaskID,
				WorkerID:      current.Identity.WorkerID,
				TaskType:      task.TaskType,
				Phase:         task.Phase,
				PhaseName:     task.PhaseName,
				Status:        "active",
				StartedAt:     task.StartedAt,
				LastUpdatedAt: nonZeroTime(task.UpdatedAt, now),
				Metadata:      cloneMetadata(task.Metadata),
			})
		case EventTaskFinished:
			task, ok := previousByID[event.TaskID]
			if !ok {
				continue
			}
			records = append(records, TaskHistoryRecord{
				TaskID:        task.TaskID,
				WorkerID:      previous.Identity.WorkerID,
				TaskType:      task.TaskType,
				Phase:         task.Phase,
				PhaseName:     task.PhaseName,
				Status:        "finished",
				StartedAt:     task.StartedAt,
				EndedAt:       now,
				LastUpdatedAt: now,
				Metadata:      cloneMetadata(task.Metadata),
			})
		}
	}
	return records
}
