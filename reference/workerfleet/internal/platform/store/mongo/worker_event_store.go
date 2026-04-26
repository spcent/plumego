package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
)

func (s *Store) AppendWorkerEvent(event domain.DomainEvent) error {
	ctx, cancel := s.operationContext()
	defer cancel()

	doc := WorkerEventDocFromDomain(event, s.now().Add(s.retention))
	_, err := s.collections.WorkerEvents.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (s *Store) ListWorkerEvents(workerID domain.WorkerID) ([]domain.DomainEvent, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	cursor, err := s.collections.WorkerEvents.Find(ctx, workerEventFilterDoc(workerID), options.Find().SetSort(bson.D{{Key: "occurred_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []WorkerEventDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	events := make([]domain.DomainEvent, 0, len(docs))
	for _, doc := range docs {
		events = append(events, doc.Domain())
	}
	return events, nil
}
