package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) AppendAlert(record platformstore.AlertRecord) error {
	ctx, cancel := s.operationContext()
	defer cancel()

	doc := AlertEventDocFromRecord(record, s.now().Add(s.retention))
	_, err := s.collections.AlertEvents.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (s *Store) ListAlerts(filter platformstore.AlertFilter) ([]platformstore.AlertRecord, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	cursor, err := s.collections.AlertEvents.Find(ctx, alertFilterDoc(filter), options.Find().SetSort(bson.D{{Key: "triggered_at", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []AlertEventDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	records := make([]platformstore.AlertRecord, 0, len(docs))
	for _, doc := range docs {
		records = append(records, doc.Record())
	}
	return records, nil
}

func (s *Store) ListAlertRecords() ([]platformstore.AlertRecord, error) {
	return s.ListAlerts(platformstore.AlertFilter{})
}

func alertFilterDoc(filter platformstore.AlertFilter) bson.M {
	doc := bson.M{}
	if filter.WorkerID != "" {
		doc["worker_id"] = string(filter.WorkerID)
	}
	if filter.AlertType != "" {
		doc["alert_type"] = string(filter.AlertType)
	}
	if filter.Status != "" {
		doc["status"] = string(filter.Status)
	}
	return doc
}
