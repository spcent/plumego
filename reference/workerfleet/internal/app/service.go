package app

import (
	"context"
	"time"

	"workerfleet/internal/domain"
	"workerfleet/internal/handler"
	platformstore "workerfleet/internal/platform/store"
)

type Service struct {
	ingest *domain.IngestService
	store  platformstore.QueryStore
}

func NewService(ingest *domain.IngestService, store platformstore.QueryStore) *Service {
	return &Service{
		ingest: ingest,
		store:  store,
	}
}

func (s *Service) RegisterWorker(ctx context.Context, input handler.RegisterWorkerInput) (handler.RegisterWorkerResult, error) {
	if s.ingest == nil {
		return handler.RegisterWorkerResult{}, handler.ErrNotImplemented
	}
	snapshot, err := s.ingest.Register(domain.RegisterCommand{
		Identity:   input.Identity,
		ObservedAt: input.ObservedAt,
	})
	if err != nil {
		return handler.RegisterWorkerResult{}, err
	}
	return handler.RegisterWorkerResult{
		WorkerID:     string(snapshot.Identity.WorkerID),
		Status:       snapshot.Status,
		RegisteredAt: nonZeroTime(input.ObservedAt, snapshot.LastStatusChangedAt),
	}, nil
}

func (s *Service) HeartbeatWorker(ctx context.Context, input handler.HeartbeatWorkerInput) (handler.HeartbeatWorkerResult, error) {
	if s.ingest == nil {
		return handler.HeartbeatWorkerResult{}, handler.ErrNotImplemented
	}
	snapshot, err := s.ingest.Heartbeat(domain.WorkerReport{
		Identity: domain.WorkerIdentity{
			WorkerID: input.WorkerID,
		},
		ProcessAlive:   input.ProcessAlive,
		AcceptingTasks: input.AcceptingTasks,
		ObservedAt:     input.ObservedAt,
		LastError:      input.LastError,
		ActiveTasks:    input.ActiveTasks,
	})
	if err != nil {
		return handler.HeartbeatWorkerResult{}, err
	}
	return handler.HeartbeatWorkerResult{
		WorkerID:        string(snapshot.Identity.WorkerID),
		Status:          snapshot.Status,
		StatusReason:    snapshot.StatusReason,
		ObservedAt:      snapshot.Runtime.LastHeartbeatAt,
		ActiveTaskCount: snapshot.ActiveTaskCount,
	}, nil
}

func (s *Service) ListWorkers(ctx context.Context, query handler.WorkerListQuery) (handler.WorkerListResult, error) {
	if s.store == nil {
		return handler.WorkerListResult{}, handler.ErrNotImplemented
	}
	snapshots, err := s.store.ListWorkerSnapshots(platformstore.WorkerSnapshotFilter{
		Status:         query.Status,
		Namespace:      query.Namespace,
		NodeName:       query.NodeName,
		TaskType:       query.TaskType,
		AcceptingTasks: query.AcceptingTasks,
	})
	if err != nil {
		return handler.WorkerListResult{}, err
	}
	total := len(snapshots)
	start, end := paginate(total, query.Page, query.PageSize)
	items := make([]handler.WorkerView, 0, end-start)
	for _, snapshot := range snapshots[start:end] {
		items = append(items, workerViewFromSnapshot(snapshot))
	}
	return handler.WorkerListResult{
		Items:    items,
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) GetWorker(ctx context.Context, workerID domain.WorkerID) (handler.WorkerDetail, error) {
	if s.store == nil {
		return handler.WorkerDetail{}, handler.ErrNotImplemented
	}
	snapshot, ok, err := s.store.GetWorkerSnapshot(workerID)
	if err != nil {
		return handler.WorkerDetail{}, err
	}
	if !ok {
		return handler.WorkerDetail{}, handler.ErrNotFound
	}
	return workerViewFromSnapshot(snapshot), nil
}

func (s *Service) GetTask(ctx context.Context, taskID domain.TaskID) (handler.TaskDetail, error) {
	if s.store == nil {
		return handler.TaskDetail{}, handler.ErrNotImplemented
	}
	current, ok, err := s.store.GetTask(taskID)
	if err != nil {
		return handler.TaskDetail{}, err
	}
	if ok {
		return handler.TaskDetail{
			TaskID:      string(current.Task.TaskID),
			WorkerID:    string(current.WorkerID),
			ExecPlanID:  string(current.Task.ExecPlanID),
			TaskType:    current.Task.TaskType,
			Phase:       string(current.Task.Phase),
			PhaseName:   current.Task.PhaseName,
			CurrentStep: stepViewFromRuntime(current.Task.CurrentStep),
			Status:      "active",
			StartedAt:   current.Task.StartedAt,
			UpdatedAt:   current.Task.UpdatedAt,
			Metadata:    cloneStringMap(current.Task.Metadata),
		}, nil
	}
	latest, ok, err := s.store.LatestTask(taskID)
	if err != nil {
		return handler.TaskDetail{}, err
	}
	if !ok {
		return handler.TaskDetail{}, handler.ErrNotFound
	}
	return handler.TaskDetail{
		TaskID:      string(latest.TaskID),
		WorkerID:    string(latest.WorkerID),
		ExecPlanID:  string(latest.ExecPlanID),
		TaskType:    latest.TaskType,
		Phase:       string(latest.Phase),
		PhaseName:   latest.PhaseName,
		CurrentStep: stepViewFromRuntime(latest.CurrentStep),
		Status:      latest.Status,
		StartedAt:   latest.StartedAt,
		UpdatedAt:   latest.LastUpdatedAt,
		EndedAt:     latest.EndedAt,
		Metadata:    cloneStringMap(latest.Metadata),
	}, nil
}

func (s *Service) GetCaseTimeline(ctx context.Context, taskID domain.TaskID) (handler.CaseTimelineResult, error) {
	stepStore, ok := s.caseStepStore()
	if !ok {
		return handler.CaseTimelineResult{}, handler.ErrNotImplemented
	}
	records, err := stepStore.CaseStepHistory(taskID)
	if err != nil {
		return handler.CaseTimelineResult{}, err
	}
	if len(records) == 0 {
		return handler.CaseTimelineResult{}, handler.ErrNotFound
	}
	items := make([]handler.CaseStepView, 0, len(records))
	for _, record := range records {
		items = append(items, caseStepViewFromRecord(record))
	}
	return handler.CaseTimelineResult{
		TaskID: string(taskID),
		Items:  items,
	}, nil
}

func (s *Service) ListExecPlanCases(ctx context.Context, query handler.ExecPlanCaseDrilldownQuery) (handler.ExecPlanCaseDrilldownResult, error) {
	stepStore, ok := s.caseStepStore()
	if !ok {
		return handler.ExecPlanCaseDrilldownResult{}, handler.ErrNotImplemented
	}
	records, err := stepStore.ListCaseStepHistory(platformstore.CaseStepHistoryFilter{
		ExecPlanID: domain.ExecPlanID(query.ExecPlanID),
		NodeName:   query.NodeName,
		PodName:    query.PodName,
		Step:       query.Step,
	})
	if err != nil {
		return handler.ExecPlanCaseDrilldownResult{}, err
	}
	total := len(records)
	start, end := paginate(total, query.Page, query.PageSize)
	items := make([]handler.CaseStepView, 0, end-start)
	for _, record := range records[start:end] {
		items = append(items, caseStepViewFromRecord(record))
	}
	return handler.ExecPlanCaseDrilldownResult{
		ExecPlanID: query.ExecPlanID,
		Items:      items,
		Page:       query.Page,
		PageSize:   query.PageSize,
		Total:      total,
	}, nil
}

func (s *Service) FleetSummary(ctx context.Context) (handler.FleetSummary, error) {
	if s.store == nil {
		return handler.FleetSummary{}, handler.ErrNotImplemented
	}
	counts := s.store.FleetCounts()
	return handler.FleetSummary{
		TotalWorkers:     counts.TotalWorkers,
		OnlineWorkers:    counts.OnlineWorkers,
		DegradedWorkers:  counts.DegradedWorkers,
		OfflineWorkers:   counts.OfflineWorkers,
		UnknownWorkers:   counts.UnknownWorkers,
		AcceptingWorkers: counts.AcceptingWorkers,
		BusyWorkers:      counts.BusyWorkers,
		ActiveTaskCount:  counts.ActiveTaskCount,
	}, nil
}

func (s *Service) ListAlerts(ctx context.Context, query handler.AlertListQuery) (handler.AlertListResult, error) {
	if s.store == nil {
		return handler.AlertListResult{}, handler.ErrNotImplemented
	}
	alerts, err := s.store.ListAlerts(platformstore.AlertFilter{
		WorkerID:  domain.WorkerID(query.WorkerID),
		AlertType: domain.AlertType(query.AlertType),
		Status:    domain.AlertStatus(query.Status),
	})
	if err != nil {
		return handler.AlertListResult{}, err
	}
	total := len(alerts)
	start, end := paginate(total, query.Page, query.PageSize)
	items := make([]handler.AlertView, 0, end-start)
	for _, alert := range alerts[start:end] {
		items = append(items, handler.AlertView{
			AlertID:     alert.AlertID,
			WorkerID:    string(alert.WorkerID),
			TaskID:      string(alert.TaskID),
			AlertType:   string(alert.AlertType),
			Status:      string(alert.Status),
			Severity:    alert.Severity,
			Message:     alert.Message,
			TriggeredAt: alert.TriggeredAt,
			ResolvedAt:  alert.ResolvedAt,
		})
	}
	return handler.AlertListResult{
		Items:    items,
		Page:     query.Page,
		PageSize: query.PageSize,
		Total:    total,
	}, nil
}

func (s *Service) caseStepStore() (platformstore.CaseStepHistoryStore, bool) {
	if s.store == nil {
		return nil, false
	}
	stepStore, ok := s.store.(platformstore.CaseStepHistoryStore)
	return stepStore, ok
}

func caseStepViewFromRecord(record platformstore.CaseStepHistoryRecord) handler.CaseStepView {
	return handler.CaseStepView{
		TaskID:     string(record.TaskID),
		WorkerID:   string(record.WorkerID),
		ExecPlanID: string(record.ExecPlanID),
		Namespace:  record.Namespace,
		PodName:    record.PodName,
		NodeName:   record.NodeName,
		Step:       record.Step,
		StepName:   record.StepName,
		Status:     record.Status,
		Result:     record.Result,
		ErrorClass: record.ErrorClass,
		Attempt:    record.Attempt,
		StartedAt:  record.StartedAt,
		FinishedAt: record.FinishedAt,
		ObservedAt: record.ObservedAt,
		EventType:  record.EventType,
	}
}

func stepViewFromRuntime(step domain.CaseStepRuntime) *handler.StepView {
	if step.Step == "" {
		return nil
	}
	return &handler.StepView{
		Step:       step.Step,
		StepName:   step.StepName,
		Status:     step.Status,
		StartedAt:  step.StartedAt,
		UpdatedAt:  step.UpdatedAt,
		FinishedAt: step.FinishedAt,
		Attempt:    step.Attempt,
		ErrorClass: step.ErrorClass,
	}
}

func workerViewFromSnapshot(snapshot domain.WorkerSnapshot) handler.WorkerView {
	tasks := make([]handler.TaskView, 0, len(snapshot.ActiveTasks))
	for _, task := range snapshot.ActiveTasks {
		tasks = append(tasks, handler.TaskView{
			TaskID:      string(task.TaskID),
			ExecPlanID:  string(task.ExecPlanID),
			TaskType:    task.TaskType,
			Phase:       string(task.Phase),
			PhaseName:   task.PhaseName,
			CurrentStep: stepViewFromRuntime(task.CurrentStep),
			StartedAt:   task.StartedAt,
			UpdatedAt:   task.UpdatedAt,
			Metadata:    cloneStringMap(task.Metadata),
		})
	}
	return handler.WorkerView{
		WorkerID:        string(snapshot.Identity.WorkerID),
		Namespace:       snapshot.Identity.Namespace,
		PodName:         snapshot.Identity.PodName,
		NodeName:        snapshot.Identity.NodeName,
		ContainerName:   snapshot.Identity.ContainerName,
		Image:           snapshot.Identity.Image,
		Version:         snapshot.Identity.Version,
		Status:          string(snapshot.Status),
		StatusReason:    snapshot.StatusReason,
		ProcessAlive:    snapshot.Runtime.ProcessAlive,
		AcceptingTasks:  snapshot.Runtime.AcceptingTasks,
		LastSeenAt:      snapshot.Runtime.LastSeenAt,
		LastReadyAt:     snapshot.Runtime.LastReadyAt,
		ActiveTaskCount: snapshot.ActiveTaskCount,
		ActiveTasks:     tasks,
	}
}

func paginate(total int, page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func nonZeroTime(primary time.Time, fallback time.Time) time.Time {
	if !primary.IsZero() {
		return primary
	}
	return fallback
}

var _ handler.Service = (*Service)(nil)
