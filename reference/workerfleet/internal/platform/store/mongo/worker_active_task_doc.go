package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
)

type WorkerActiveTaskDoc struct {
	ID            string            `bson:"_id"`
	WorkerID      string            `bson:"worker_id"`
	TaskID        string            `bson:"task_id"`
	TaskType      string            `bson:"task_type,omitempty"`
	Phase         string            `bson:"phase,omitempty"`
	PhaseName     string            `bson:"phase_name,omitempty"`
	StartedAt     time.Time         `bson:"started_at,omitempty"`
	UpdatedAt     time.Time         `bson:"updated_at,omitempty"`
	Metadata      map[string]string `bson:"metadata,omitempty"`
	SchemaVersion int               `bson:"schema_version"`
}

func WorkerActiveTaskDocFromDomain(workerID domain.WorkerID, task domain.ActiveTask) WorkerActiveTaskDoc {
	return WorkerActiveTaskDoc{
		ID:            ActiveTaskDocumentID(workerID, task.TaskID),
		WorkerID:      string(workerID),
		TaskID:        string(task.TaskID),
		TaskType:      task.TaskType,
		Phase:         string(task.Phase),
		PhaseName:     task.PhaseName,
		StartedAt:     task.StartedAt,
		UpdatedAt:     task.UpdatedAt,
		Metadata:      cloneStringMap(task.Metadata),
		SchemaVersion: SchemaVersion,
	}
}

func ActiveTaskDocumentID(workerID domain.WorkerID, taskID domain.TaskID) string {
	return fmt.Sprintf("%s:%s", workerID, taskID)
}
