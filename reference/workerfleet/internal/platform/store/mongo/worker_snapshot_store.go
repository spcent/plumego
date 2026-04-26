package mongo

import (
	"errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) UpsertWorkerSnapshot(snapshot domain.WorkerSnapshot) error {
	workerID := snapshot.Identity.WorkerID
	if workerID == "" {
		return nil
	}

	ctx, cancel := s.operationContext()
	defer cancel()

	now := s.now()
	normalized := normalizeSnapshot(snapshot)
	doc := WorkerSnapshotDocFromDomain(normalized, now)
	if _, err := s.collections.WorkerSnapshots.ReplaceOne(
		ctx,
		workerSnapshotIDFilter(workerID),
		doc,
		options.Replace().SetUpsert(true),
	); err != nil {
		return err
	}
	return s.replaceActiveTasks(ctx, workerID, normalized.ActiveTasks)
}

func (s *Store) GetWorkerSnapshot(workerID domain.WorkerID) (domain.WorkerSnapshot, bool, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	var doc WorkerSnapshotDoc
	err := s.collections.WorkerSnapshots.FindOne(ctx, workerSnapshotIDFilter(workerID)).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return domain.WorkerSnapshot{}, false, nil
	}
	if err != nil {
		return domain.WorkerSnapshot{}, false, err
	}
	return doc.Domain(), true, nil
}

func (s *Store) ListWorkerSnapshots(filter platformstore.WorkerSnapshotFilter) ([]domain.WorkerSnapshot, error) {
	ctx, cancel := s.operationContext()
	defer cancel()

	cursor, err := s.collections.WorkerSnapshots.Find(ctx, workerSnapshotFilterDoc(filter), options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var docs []WorkerSnapshotDoc
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	out := make([]domain.WorkerSnapshot, 0, len(docs))
	for _, doc := range docs {
		out = append(out, doc.Domain())
	}
	return out, nil
}

func (s *Store) ListCurrentWorkerSnapshots() ([]domain.WorkerSnapshot, error) {
	return s.ListWorkerSnapshots(platformstore.WorkerSnapshotFilter{})
}

func (s *Store) FleetCounts() platformstore.FleetCounts {
	snapshots, err := s.ListWorkerSnapshots(platformstore.WorkerSnapshotFilter{})
	if err != nil {
		return platformstore.FleetCounts{}
	}
	counts := platformstore.FleetCounts{TotalWorkers: len(snapshots)}
	for _, snapshot := range snapshots {
		switch snapshot.Status {
		case domain.WorkerStatusOnline:
			counts.OnlineWorkers++
		case domain.WorkerStatusDegraded:
			counts.DegradedWorkers++
		case domain.WorkerStatusOffline:
			counts.OfflineWorkers++
		default:
			counts.UnknownWorkers++
		}
		if snapshot.Runtime.AcceptingTasks {
			counts.AcceptingWorkers++
		}
		if len(snapshot.ActiveTasks) > 0 {
			counts.BusyWorkers++
		}
		counts.ActiveTaskCount += len(snapshot.ActiveTasks)
	}
	return counts
}

func normalizeSnapshot(snapshot domain.WorkerSnapshot) domain.WorkerSnapshot {
	snapshot.ActiveTasks = cloneTasks(snapshot.ActiveTasks)
	snapshot.ActiveTaskCount = len(snapshot.ActiveTasks)
	return snapshot
}
