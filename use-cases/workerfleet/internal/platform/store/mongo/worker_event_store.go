package mongo

import (
	"context"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"workerfleet/internal/domain"
	platformstore "workerfleet/internal/platform/store"
)

func (s *Store) AppendWorkerEvent(ctx context.Context, event domain.DomainEvent) error {
	ctx, cancel := s.operationContext(ctx)
	defer cancel()

	doc := WorkerEventDocFromDomain(event, s.now().Add(s.retention))
	_, err := s.collections.WorkerEvents.InsertOne(ctx, doc)
	if mongo.IsDuplicateKeyError(err) {
		err = nil
	}
	if err != nil {
		return err
	}
	record, ok, err := s.caseStepHistoryRecordFromEvent(ctx, event)
	if err != nil || !ok {
		return err
	}
	return s.appendCaseStepHistory(ctx, record)
}

func (s *Store) ListWorkerEvents(ctx context.Context, workerID domain.WorkerID) ([]domain.DomainEvent, error) {
	ctx, cancel := s.operationContext(ctx)
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

func (s *Store) caseStepHistoryRecordFromEvent(ctx context.Context, event domain.DomainEvent) (platformstore.CaseStepHistoryRecord, bool, error) {
	if event.Type != domain.EventTaskStepChanged && event.Type != domain.EventTaskStepFinished {
		return platformstore.CaseStepHistoryRecord{}, false, nil
	}
	if event.TaskID == "" {
		return platformstore.CaseStepHistoryRecord{}, false, nil
	}

	snapshot, _, err := s.getWorkerSnapshot(ctx, event.WorkerID)
	if err != nil {
		return platformstore.CaseStepHistoryRecord{}, false, err
	}
	record := platformstore.CaseStepHistoryRecord{
		TaskID:     event.TaskID,
		WorkerID:   event.WorkerID,
		Namespace:  snapshot.Identity.Namespace,
		PodName:    snapshot.Identity.PodName,
		NodeName:   snapshot.Identity.NodeName,
		ObservedAt: event.OccurredAt,
		EventType:  event.Type,
	}
	switch event.Type {
	case domain.EventTaskStepChanged:
		record.ExecPlanID = domain.ExecPlanID(event.Attributes["exec_plan_id"])
		record.Step = event.Attributes["to_step"]
		record.StepName = event.Attributes["to_step_name"]
		record.Status = domain.CaseStepStatus(event.Attributes["to_step_status"])
		record.Attempt = parsePositiveInt(event.Attributes["to_step_attempt"])
		record.StartedAt = parseTime(event.Attributes["to_step_started_at"])
	case domain.EventTaskStepFinished:
		record.ExecPlanID = domain.ExecPlanID(event.Attributes["exec_plan_id"])
		record.Step = event.Attributes["step"]
		record.StepName = event.Attributes["step_name"]
		record.Status = domain.CaseStepStatus(event.Attributes["step_status"])
		record.Result = event.Attributes["result"]
		record.ErrorClass = event.Attributes["error_class"]
		record.Attempt = parsePositiveInt(event.Attributes["step_attempt"])
		record.StartedAt = parseTime(event.Attributes["step_started"])
		record.FinishedAt = parseTime(event.Attributes["step_finished"])
	}
	if record.ExecPlanID == "" {
		record.ExecPlanID = execPlanIDForTask(snapshot, event.TaskID)
	}
	return record, record.Step != "", nil
}

func execPlanIDForTask(snapshot domain.WorkerSnapshot, taskID domain.TaskID) domain.ExecPlanID {
	for _, task := range snapshot.ActiveTasks {
		if task.TaskID == taskID {
			return task.ExecPlanID
		}
	}
	return ""
}

func parsePositiveInt(value string) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
