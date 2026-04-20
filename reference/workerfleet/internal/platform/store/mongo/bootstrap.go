package mongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

const SchemaVersion = 1

const (
	CollectionWorkerSnapshots = "worker_snapshots"
	CollectionWorkerActive    = "worker_active_tasks"
	CollectionTaskHistory     = "task_history"
	CollectionWorkerEvents    = "worker_events"
	CollectionAlertEvents     = "alert_events"
)

type Collections struct {
	WorkerSnapshots *mongo.Collection
	WorkerActive    *mongo.Collection
	TaskHistory     *mongo.Collection
	WorkerEvents    *mongo.Collection
	AlertEvents     *mongo.Collection
}

func NewCollections(db *mongo.Database) (Collections, error) {
	if db == nil {
		return Collections{}, errors.New("mongo database is required")
	}
	return Collections{
		WorkerSnapshots: db.Collection(CollectionWorkerSnapshots),
		WorkerActive:    db.Collection(CollectionWorkerActive),
		TaskHistory:     db.Collection(CollectionTaskHistory),
		WorkerEvents:    db.Collection(CollectionWorkerEvents),
		AlertEvents:     db.Collection(CollectionAlertEvents),
	}, nil
}

func EnsureIndexes(ctx context.Context, db *mongo.Database) error {
	collections, err := NewCollections(db)
	if err != nil {
		return err
	}

	ordered := []struct {
		name       string
		collection *mongo.Collection
		specs      []IndexSpec
	}{
		{name: CollectionWorkerSnapshots, collection: collections.WorkerSnapshots, specs: CollectionIndexSpecs()[CollectionWorkerSnapshots]},
		{name: CollectionWorkerActive, collection: collections.WorkerActive, specs: CollectionIndexSpecs()[CollectionWorkerActive]},
		{name: CollectionTaskHistory, collection: collections.TaskHistory, specs: CollectionIndexSpecs()[CollectionTaskHistory]},
		{name: CollectionWorkerEvents, collection: collections.WorkerEvents, specs: CollectionIndexSpecs()[CollectionWorkerEvents]},
		{name: CollectionAlertEvents, collection: collections.AlertEvents, specs: CollectionIndexSpecs()[CollectionAlertEvents]},
	}

	for _, entry := range ordered {
		if len(entry.specs) == 0 {
			continue
		}
		models := make([]mongo.IndexModel, 0, len(entry.specs))
		for _, spec := range entry.specs {
			models = append(models, spec.Model())
		}
		if _, err := entry.collection.Indexes().CreateMany(ctx, models); err != nil {
			return fmt.Errorf("ensure indexes for %s: %w", entry.name, err)
		}
	}

	return nil
}
