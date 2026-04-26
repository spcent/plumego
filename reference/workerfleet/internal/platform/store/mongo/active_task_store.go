package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) ReplaceActiveTasks(workerID domain.WorkerID, tasks []domain.ActiveTask) error {
	ctx, cancel := s.operationContext()
	defer cancel()

	normalized := cloneTasks(tasks)
	if err := s.replaceActiveTasks(ctx, workerID, normalized); err != nil {
		return err
	}

	embedded := make([]ActiveTaskEmbeddedDoc, 0, len(normalized))
	now := s.now()
	for _, task := range normalized {
		embedded = append(embedded, ActiveTaskEmbeddedDocFromDomain(task, now))
	}
	_, err := s.collections.WorkerSnapshots.UpdateOne(ctx, workerSnapshotIDFilter(workerID), bson.M{
		"$set": bson.M{
			"active_tasks":      embedded,
			"active_task_count": len(embedded),
			"updated_at":        now,
		},
	})
	return err
}

func (s *Store) ActiveTasks(workerID domain.WorkerID) ([]domain.ActiveTask, bool, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	cursor, err := s.collections.WorkerActive.Find(ctx, workerActiveByWorkerIDFilter(workerID), options.Find().SetSort(bson.D{{Key: "task_id", Value: 1}}))
	if err != nil {
		return nil, false, err
	}
	defer cursor.Close(ctx)

	var docs []WorkerActiveTaskDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, false, err
	}
	if len(docs) == 0 {
		return nil, false, nil
	}

	tasks := make([]domain.ActiveTask, 0, len(docs))
	for _, doc := range docs {
		tasks = append(tasks, doc.Domain().Task)
	}
	return tasks, true, nil
}

func (s *Store) GetTask(taskID domain.TaskID) (platformstore.CurrentTaskRecord, bool, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	var doc WorkerActiveTaskDoc
	err := s.collections.WorkerActive.FindOne(ctx, workerActiveByTaskIDFilter(taskID)).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return platformstore.CurrentTaskRecord{}, false, nil
	}
	if err != nil {
		return platformstore.CurrentTaskRecord{}, false, err
	}
	return doc.Domain(), true, nil
}

func (s *Store) replaceActiveTasks(ctx context.Context, workerID domain.WorkerID, tasks []domain.ActiveTask) error {
	if _, err := s.collections.WorkerActive.DeleteMany(ctx, workerActiveByWorkerIDFilter(workerID)); err != nil {
		return err
	}

	docs := make([]any, 0, len(tasks))
	for _, task := range tasks {
		if task.TaskID == "" {
			continue
		}
		docs = append(docs, WorkerActiveTaskDocFromDomain(workerID, task))
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := s.collections.WorkerActive.InsertMany(ctx, docs)
	return err
}
