package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

type TaskHistoryDoc struct {
	ID            string              `bson:"_id"`
	TaskID        string              `bson:"task_id"`
	WorkerID      string              `bson:"worker_id"`
	ExecPlanID    string              `bson:"exec_plan_id,omitempty"`
	TaskType      string              `bson:"task_type,omitempty"`
	Phase         string              `bson:"phase,omitempty"`
	PhaseName     string              `bson:"phase_name,omitempty"`
	CurrentStep   CaseStepEmbeddedDoc `bson:"current_step,omitempty"`
	Status        string              `bson:"status,omitempty"`
	StartedAt     time.Time           `bson:"started_at,omitempty"`
	EndedAt       time.Time           `bson:"ended_at,omitempty"`
	LastUpdatedAt time.Time           `bson:"last_updated_at,omitempty"`
	Metadata      map[string]string   `bson:"metadata,omitempty"`
	ExpireAt      time.Time           `bson:"expire_at,omitempty"`
	SchemaVersion int                 `bson:"schema_version"`
}

func TaskHistoryDocFromRecord(record platformstore.TaskHistoryRecord, expireAt time.Time) TaskHistoryDoc {
	return TaskHistoryDoc{
		ID:            fmt.Sprintf("%s:%s", record.TaskID, latestTaskHistoryTime(record).UTC().Format(time.RFC3339Nano)),
		TaskID:        string(record.TaskID),
		WorkerID:      string(record.WorkerID),
		ExecPlanID:    string(record.ExecPlanID),
		TaskType:      record.TaskType,
		Phase:         string(record.Phase),
		PhaseName:     record.PhaseName,
		CurrentStep:   CaseStepEmbeddedDocFromDomain(record.CurrentStep),
		Status:        record.Status,
		StartedAt:     record.StartedAt,
		EndedAt:       record.EndedAt,
		LastUpdatedAt: record.LastUpdatedAt,
		Metadata:      cloneStringMap(record.Metadata),
		ExpireAt:      expireAt,
		SchemaVersion: SchemaVersion,
	}
}

func latestTaskHistoryTime(record platformstore.TaskHistoryRecord) time.Time {
	if !record.EndedAt.IsZero() {
		return record.EndedAt
	}
	if !record.LastUpdatedAt.IsZero() {
		return record.LastUpdatedAt
	}
	return record.StartedAt
}

func (doc TaskHistoryDoc) Record() platformstore.TaskHistoryRecord {
	return platformstore.TaskHistoryRecord{
		TaskID:        domain.TaskID(doc.TaskID),
		WorkerID:      domain.WorkerID(doc.WorkerID),
		ExecPlanID:    domain.ExecPlanID(doc.ExecPlanID),
		TaskType:      doc.TaskType,
		Phase:         domain.TaskPhase(doc.Phase),
		PhaseName:     doc.PhaseName,
		CurrentStep:   doc.CurrentStep.Domain(),
		Status:        doc.Status,
		StartedAt:     doc.StartedAt,
		EndedAt:       doc.EndedAt,
		LastUpdatedAt: doc.LastUpdatedAt,
		Metadata:      cloneStringMap(doc.Metadata),
	}
}
