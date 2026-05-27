package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
)

type WorkerActiveTaskDoc struct {
	ID            string              `bson:"_id"`
	WorkerID      string              `bson:"worker_id"`
	TaskID        string              `bson:"task_id"`
	ExecPlanID    string              `bson:"exec_plan_id,omitempty"`
	TaskType      string              `bson:"task_type,omitempty"`
	Phase         string              `bson:"phase,omitempty"`
	PhaseName     string              `bson:"phase_name,omitempty"`
	CurrentStep   CaseStepEmbeddedDoc `bson:"current_step,omitempty"`
	StartedAt     time.Time           `bson:"started_at,omitempty"`
	UpdatedAt     time.Time           `bson:"updated_at,omitempty"`
	Metadata      map[string]string   `bson:"metadata,omitempty"`
	SchemaVersion int                 `bson:"schema_version"`
}

func WorkerActiveTaskDocFromDomain(workerID domain.WorkerID, task domain.ActiveTask) WorkerActiveTaskDoc {
	return WorkerActiveTaskDoc{
		ID:            ActiveTaskDocumentID(workerID, task.TaskID),
		WorkerID:      string(workerID),
		TaskID:        string(task.TaskID),
		ExecPlanID:    string(task.ExecPlanID),
		TaskType:      task.TaskType,
		Phase:         string(task.Phase),
		PhaseName:     task.PhaseName,
		CurrentStep:   CaseStepEmbeddedDocFromDomain(task.CurrentStep),
		StartedAt:     task.StartedAt,
		UpdatedAt:     task.UpdatedAt,
		Metadata:      cloneStringMap(task.Metadata),
		SchemaVersion: SchemaVersion,
	}
}

func ActiveTaskDocumentID(workerID domain.WorkerID, taskID domain.TaskID) string {
	return fmt.Sprintf("%s:%s", workerID, taskID)
}
