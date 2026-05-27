package mongo

import (
	"fmt"
	"time"

	"workerfleet/internal/domain"
)

type WorkerEventDoc struct {
	ID            string            `bson:"_id"`
	WorkerID      string            `bson:"worker_id,omitempty"`
	TaskID        string            `bson:"task_id,omitempty"`
	EventType     string            `bson:"event_type"`
	Reason        string            `bson:"reason,omitempty"`
	Attributes    map[string]string `bson:"attributes,omitempty"`
	OccurredAt    time.Time         `bson:"occurred_at,omitempty"`
	ExpireAt      time.Time         `bson:"expire_at,omitempty"`
	SchemaVersion int               `bson:"schema_version"`
}

func WorkerEventDocFromDomain(event domain.DomainEvent, expireAt time.Time) WorkerEventDoc {
	return WorkerEventDoc{
		ID:            fmt.Sprintf("%s:%s:%s", event.Type, event.WorkerID, event.OccurredAt.UTC().Format(time.RFC3339Nano)),
		WorkerID:      string(event.WorkerID),
		TaskID:        string(event.TaskID),
		EventType:     string(event.Type),
		Reason:        event.Reason,
		Attributes:    cloneStringMap(event.Attributes),
		OccurredAt:    event.OccurredAt,
		ExpireAt:      expireAt,
		SchemaVersion: SchemaVersion,
	}
}

func (doc WorkerEventDoc) Domain() domain.DomainEvent {
	return domain.DomainEvent{
		Type:       domain.EventType(doc.EventType),
		OccurredAt: doc.OccurredAt,
		WorkerID:   domain.WorkerID(doc.WorkerID),
		TaskID:     domain.TaskID(doc.TaskID),
		Reason:     doc.Reason,
		Attributes: cloneStringMap(doc.Attributes),
	}
}
