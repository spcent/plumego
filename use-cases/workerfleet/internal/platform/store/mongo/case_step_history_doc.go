package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

type CaseStepHistoryDoc struct {
	ID            string    `bson:"_id"`
	TaskID        string    `bson:"task_id"`
	WorkerID      string    `bson:"worker_id,omitempty"`
	ExecPlanID    string    `bson:"exec_plan_id,omitempty"`
	Namespace     string    `bson:"namespace,omitempty"`
	PodName       string    `bson:"pod_name,omitempty"`
	NodeName      string    `bson:"node_name,omitempty"`
	Step          string    `bson:"step"`
	StepName      string    `bson:"step_name,omitempty"`
	Status        string    `bson:"status,omitempty"`
	Result        string    `bson:"result,omitempty"`
	ErrorClass    string    `bson:"error_class,omitempty"`
	Attempt       int       `bson:"attempt,omitempty"`
	StartedAt     time.Time `bson:"started_at,omitempty"`
	FinishedAt    time.Time `bson:"finished_at,omitempty"`
	ObservedAt    time.Time `bson:"observed_at,omitempty"`
	EventType     string    `bson:"event_type,omitempty"`
	ExpireAt      time.Time `bson:"expire_at,omitempty"`
	SchemaVersion int       `bson:"schema_version"`
}

func CaseStepHistoryDocFromRecord(record platformstore.CaseStepHistoryRecord, expireAt time.Time) CaseStepHistoryDoc {
	return CaseStepHistoryDoc{
		ID:            fmt.Sprintf("%s:%s:%s:%s", record.TaskID, record.Step, record.EventType, latestCaseStepHistoryTime(record).UTC().Format(time.RFC3339Nano)),
		TaskID:        string(record.TaskID),
		WorkerID:      string(record.WorkerID),
		ExecPlanID:    string(record.ExecPlanID),
		Namespace:     record.Namespace,
		PodName:       record.PodName,
		NodeName:      record.NodeName,
		Step:          record.Step,
		StepName:      record.StepName,
		Status:        string(record.Status),
		Result:        record.Result,
		ErrorClass:    record.ErrorClass,
		Attempt:       record.Attempt,
		StartedAt:     record.StartedAt,
		FinishedAt:    record.FinishedAt,
		ObservedAt:    nonZeroTime(record.ObservedAt, latestCaseStepHistoryTime(record)),
		EventType:     string(record.EventType),
		ExpireAt:      expireAt,
		SchemaVersion: SchemaVersion,
	}
}

func latestCaseStepHistoryTime(record platformstore.CaseStepHistoryRecord) time.Time {
	if !record.FinishedAt.IsZero() {
		return record.FinishedAt
	}
	if !record.ObservedAt.IsZero() {
		return record.ObservedAt
	}
	return record.StartedAt
}

func (doc CaseStepHistoryDoc) Record() platformstore.CaseStepHistoryRecord {
	return platformstore.CaseStepHistoryRecord{
		TaskID:     domain.TaskID(doc.TaskID),
		WorkerID:   domain.WorkerID(doc.WorkerID),
		ExecPlanID: domain.ExecPlanID(doc.ExecPlanID),
		Namespace:  doc.Namespace,
		PodName:    doc.PodName,
		NodeName:   doc.NodeName,
		Step:       doc.Step,
		StepName:   doc.StepName,
		Status:     domain.CaseStepStatus(doc.Status),
		Result:     doc.Result,
		ErrorClass: doc.ErrorClass,
		Attempt:    doc.Attempt,
		StartedAt:  doc.StartedAt,
		FinishedAt: doc.FinishedAt,
		ObservedAt: doc.ObservedAt,
		EventType:  domain.EventType(doc.EventType),
	}
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}
