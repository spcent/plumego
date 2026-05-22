package mongo

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) AppendCaseStepHistory(ctx context.Context, record platformstore.CaseStepHistoryRecord) error {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	return s.appendCaseStepHistory(ctx, record)
}

func (s *Store) CaseStepHistory(ctx context.Context, taskID domain.TaskID) ([]platformstore.CaseStepHistoryRecord, error) {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	cursor, err := s.collections.CaseStepHistory.Find(
		ctx,
		caseStepHistoryFilterDoc(platformstore.CaseStepHistoryFilter{TaskID: taskID}),
		options.Find().SetSort(bson.D{{Key: "observed_at", Value: 1}, {Key: "task_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return caseStepHistoryRecordsFromCursor(ctx, cursor)
}

func (s *Store) ListCaseStepHistory(ctx context.Context, filter platformstore.CaseStepHistoryFilter) ([]platformstore.CaseStepHistoryRecord, error) {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	cursor, err := s.collections.CaseStepHistory.Find(
		ctx,
		caseStepHistoryFilterDoc(filter),
		options.Find().SetSort(bson.D{{Key: "observed_at", Value: 1}, {Key: "task_id", Value: 1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	return caseStepHistoryRecordsFromCursor(ctx, cursor)
}

func (s *Store) appendCaseStepHistory(ctx context.Context, record platformstore.CaseStepHistoryRecord) error {
	if record.TaskID == "" || record.Step == "" {
		return nil
	}
	doc := CaseStepHistoryDocFromRecord(record, s.now().Add(s.retention))
	_, err := s.collections.CaseStepHistory.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		return nil
	}
	return err
}

func caseStepHistoryRecordsFromCursor(ctx context.Context, cursor *mongo.Cursor) ([]platformstore.CaseStepHistoryRecord, error) {
	var docs []CaseStepHistoryDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	records := make([]platformstore.CaseStepHistoryRecord, 0, len(docs))
	for _, doc := range docs {
		records = append(records, doc.Record())
	}
	return records, nil
}
