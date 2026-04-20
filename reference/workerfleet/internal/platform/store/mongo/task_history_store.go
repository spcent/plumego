package mongo

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) AppendTaskHistory(record platformstore.TaskHistoryRecord) error {
	ctx, cancel := s.operationContext()
	defer cancel()

	doc := TaskHistoryDocFromRecord(record, s.now().Add(s.retention))
	_, err := s.collections.TaskHistory.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func (s *Store) TaskHistory(taskID domain.TaskID) ([]platformstore.TaskHistoryRecord, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	cursor, err := s.collections.TaskHistory.Find(
		ctx,
		bson.M{"task_id": string(taskID)},
		options.Find().SetSort(bson.D{{Key: "last_updated_at", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []TaskHistoryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	records := make([]platformstore.TaskHistoryRecord, 0, len(docs))
	for _, doc := range docs {
		records = append(records, doc.Record())
	}
	return records, nil
}

func (s *Store) LatestTask(taskID domain.TaskID) (platformstore.TaskHistoryRecord, bool, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	var doc TaskHistoryDoc
	err := s.collections.TaskHistory.FindOne(
		ctx,
		bson.M{"task_id": string(taskID)},
		options.FindOne().SetSort(bson.D{{Key: "last_updated_at", Value: -1}}),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return platformstore.TaskHistoryRecord{}, false, nil
	}
	if err != nil {
		return platformstore.TaskHistoryRecord{}, false, err
	}
	return doc.Record(), true, nil
}
