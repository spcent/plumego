package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

type AlertEventDoc struct {
	ID            string            `bson:"_id"`
	DedupeKey     string            `bson:"dedupe_key"`
	AlertType     string            `bson:"alert_type"`
	WorkerID      string            `bson:"worker_id,omitempty"`
	TaskID        string            `bson:"task_id,omitempty"`
	Status        string            `bson:"status"`
	Severity      string            `bson:"severity,omitempty"`
	Message       string            `bson:"message,omitempty"`
	Details       map[string]string `bson:"details,omitempty"`
	TriggeredAt   time.Time         `bson:"triggered_at,omitempty"`
	ResolvedAt    time.Time         `bson:"resolved_at,omitempty"`
	ExpireAt      time.Time         `bson:"expire_at,omitempty"`
	SchemaVersion int               `bson:"schema_version"`
}

func AlertEventDocFromRecord(record platformstore.AlertRecord, expireAt time.Time) AlertEventDoc {
	return AlertEventDoc{
		ID:            fmt.Sprintf("%s:%s:%s", record.AlertType, record.WorkerID, latestAlertTime(record).UTC().Format(time.RFC3339Nano)),
		DedupeKey:     record.DedupeKey,
		AlertType:     string(record.AlertType),
		WorkerID:      string(record.WorkerID),
		TaskID:        string(record.TaskID),
		Status:        string(record.Status),
		Severity:      record.Severity,
		Message:       record.Message,
		Details:       cloneStringMap(record.Details),
		TriggeredAt:   record.TriggeredAt,
		ResolvedAt:    record.ResolvedAt,
		ExpireAt:      expireAt,
		SchemaVersion: SchemaVersion,
	}
}

func latestAlertTime(record platformstore.AlertRecord) time.Time {
	if !record.ResolvedAt.IsZero() {
		return record.ResolvedAt
	}
	return record.TriggeredAt
}

func (doc AlertEventDoc) Record() platformstore.AlertRecord {
	return platformstore.AlertRecord{
		AlertID:     doc.ID,
		WorkerID:    domain.WorkerID(doc.WorkerID),
		TaskID:      domain.TaskID(doc.TaskID),
		AlertType:   domain.AlertType(doc.AlertType),
		Status:      domain.AlertStatus(doc.Status),
		Severity:    doc.Severity,
		DedupeKey:   doc.DedupeKey,
		Message:     doc.Message,
		Details:     cloneStringMap(doc.Details),
		TriggeredAt: doc.TriggeredAt,
		ResolvedAt:  doc.ResolvedAt,
	}
}
