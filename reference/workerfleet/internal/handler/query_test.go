package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/spcent/plumego/contract"
	"workerfleet/internal/domain"
)

type stubService struct {
	registerFn    func(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error)
	heartbeatFn   func(ctx context.Context, input HeartbeatWorkerInput) (HeartbeatWorkerResult, error)
	listWorkersFn func(ctx context.Context, query WorkerListQuery) (WorkerListResult, error)
	getWorkerFn   func(ctx context.Context, workerID domain.WorkerID) (WorkerDetail, error)
	getTaskFn     func(ctx context.Context, taskID domain.TaskID) (TaskDetail, error)
	summaryFn     func(ctx context.Context) (FleetSummary, error)
	listAlertsFn  func(ctx context.Context, query AlertListQuery) (AlertListResult, error)
}

func (s stubService) RegisterWorker(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error) {
	return s.registerFn(ctx, input)
}

func (s stubService) HeartbeatWorker(ctx context.Context, input HeartbeatWorkerInput) (HeartbeatWorkerResult, error) {
	return s.heartbeatFn(ctx, input)
}

func (s stubService) ListWorkers(ctx context.Context, query WorkerListQuery) (WorkerListResult, error) {
	return s.listWorkersFn(ctx, query)
}

func (s stubService) GetWorker(ctx context.Context, workerID domain.WorkerID) (WorkerDetail, error) {
	return s.getWorkerFn(ctx, workerID)
}

func (s stubService) GetTask(ctx context.Context, taskID domain.TaskID) (TaskDetail, error) {
	return s.getTaskFn(ctx, taskID)
}

func (s stubService) FleetSummary(ctx context.Context) (FleetSummary, error) {
	return s.summaryFn(ctx)
}

func (s stubService) ListAlerts(ctx context.Context, query AlertListQuery) (AlertListResult, error) {
	return s.listAlertsFn(ctx, query)
}

func TestRegisterWorkerRejectsInvalidJSON(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return RegisterWorkerResult{}, nil
	}})

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeInvalidJSON)
}

func TestRegisterWorkerRejectsMissingWorkerID(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return RegisterWorkerResult{}, nil
	}})

	body := `{"namespace":"sim","pod_name":"worker-1","container_name":"worker"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHeartbeatWorkerAcceptsMultiTaskPayload(t *testing.T) {
	observedAt := time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)
	h := New(stubService{heartbeatFn: func(ctx context.Context, input HeartbeatWorkerInput) (HeartbeatWorkerResult, error) {
		if input.WorkerID != "worker-1" {
			t.Fatalf("worker_id = %q", input.WorkerID)
		}
		if len(input.ActiveTasks) != 2 {
			t.Fatalf("len(active_tasks) = %d, want 2", len(input.ActiveTasks))
		}
		if input.ActiveTasks[0].ExecPlanID != "plan-1" {
			t.Fatalf("exec_plan_id = %q, want plan-1", input.ActiveTasks[0].ExecPlanID)
		}
		if input.ActiveTasks[0].CurrentStep.Step != "simulate" {
			t.Fatalf("current_step.step = %q, want simulate", input.ActiveTasks[0].CurrentStep.Step)
		}
		if input.ActiveTasks[0].CurrentStep.Status != domain.CaseStepStatusRunning {
			t.Fatalf("current_step.status = %q, want running", input.ActiveTasks[0].CurrentStep.Status)
		}
		return HeartbeatWorkerResult{
			WorkerID:        "worker-1",
			Status:          domain.WorkerStatusOnline,
			StatusReason:    "busy",
			ObservedAt:      observedAt,
			ActiveTaskCount: 2,
		}, nil
	}})

	body := `{
		"worker_id":"worker-1",
		"process_alive":true,
		"accepting_tasks":false,
		"active_tasks":[
			{"task_id":"task-1","exec_plan_id":"plan-1","task_type":"simulation","phase":"running","phase_name":"running","current_step":{"step":"simulate","step_name":"simulation","status":"running","attempt":1}},
			{"task_id":"task-2","task_type":"simulation","phase":"preparing","phase_name":"warming"}
		]
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/workers/heartbeat", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.HeartbeatWorker(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var envelope contract.Response
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	payload, ok := envelope.Data.(map[string]any)
	if !ok {
		t.Fatalf("unexpected response payload %#v", envelope.Data)
	}
	if int(payload["active_task_count"].(float64)) != 2 {
		t.Fatalf("active_task_count = %#v, want 2", payload["active_task_count"])
	}
}

func TestListWorkersRejectsInvalidAcceptingTasksQuery(t *testing.T) {
	h := New(stubService{listWorkersFn: func(ctx context.Context, query WorkerListQuery) (WorkerListResult, error) {
		t.Fatalf("list workers should not be called")
		return WorkerListResult{}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/workers?accepting_tasks=maybe", nil)
	rec := httptest.NewRecorder()

	h.ListWorkers(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeInvalidQuery)
}

func TestGetWorkerReturnsNotFound(t *testing.T) {
	h := New(stubService{getWorkerFn: func(ctx context.Context, workerID domain.WorkerID) (WorkerDetail, error) {
		return WorkerDetail{}, ErrNotFound
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/workers/worker-404", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		Params: map[string]string{"worker_id": "worker-404"},
	}))
	rec := httptest.NewRecorder()

	h.GetWorker(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestFleetSummarySuccess(t *testing.T) {
	h := New(stubService{summaryFn: func(ctx context.Context) (FleetSummary, error) {
		return FleetSummary{
			TotalWorkers:     8,
			OnlineWorkers:    6,
			DegradedWorkers:  1,
			OfflineWorkers:   1,
			AcceptingWorkers: 4,
			BusyWorkers:      3,
			ActiveTaskCount:  21,
		}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/fleet/summary", nil)
	rec := httptest.NewRecorder()

	h.FleetSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func assertErrorCode(t *testing.T, body []byte, code string) {
	t.Helper()

	var payload contract.ErrorResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if payload.Error.Code != code {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, code)
	}
}

func TestServiceErrorConflictMapsToConflict(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error) {
		return RegisterWorkerResult{}, fmt.Errorf("wrapped: %w", ErrConflict)
	}})

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
