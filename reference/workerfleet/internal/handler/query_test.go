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
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/domain"
)

type stubService struct {
	registerFn    func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error)
	heartbeatFn   func(ctx context.Context, input workerapp.HeartbeatWorkerInput) (workerapp.HeartbeatWorkerResult, error)
	listWorkersFn func(ctx context.Context, query workerapp.WorkerListQuery) (workerapp.WorkerListResult, error)
	getWorkerFn   func(ctx context.Context, workerID domain.WorkerID) (workerapp.WorkerDetail, error)
	getTaskFn     func(ctx context.Context, taskID domain.TaskID) (workerapp.TaskDetail, error)
	timelineFn    func(ctx context.Context, taskID domain.TaskID) (workerapp.CaseTimelineResult, error)
	drilldownFn   func(ctx context.Context, query workerapp.ExecPlanCaseDrilldownQuery) (workerapp.ExecPlanCaseDrilldownResult, error)
	summaryFn     func(ctx context.Context) (workerapp.FleetSummary, error)
	listAlertsFn  func(ctx context.Context, query workerapp.AlertListQuery) (workerapp.AlertListResult, error)
}

type testResponseEnvelope[T any] struct {
	Data T `json:"data"`
}

func (s stubService) RegisterWorker(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
	return s.registerFn(ctx, input)
}

func (s stubService) HeartbeatWorker(ctx context.Context, input workerapp.HeartbeatWorkerInput) (workerapp.HeartbeatWorkerResult, error) {
	return s.heartbeatFn(ctx, input)
}

func (s stubService) ListWorkers(ctx context.Context, query workerapp.WorkerListQuery) (workerapp.WorkerListResult, error) {
	return s.listWorkersFn(ctx, query)
}

func (s stubService) GetWorker(ctx context.Context, workerID domain.WorkerID) (workerapp.WorkerDetail, error) {
	return s.getWorkerFn(ctx, workerID)
}

func (s stubService) GetTask(ctx context.Context, taskID domain.TaskID) (workerapp.TaskDetail, error) {
	return s.getTaskFn(ctx, taskID)
}

func (s stubService) GetCaseTimeline(ctx context.Context, taskID domain.TaskID) (workerapp.CaseTimelineResult, error) {
	return s.timelineFn(ctx, taskID)
}

func (s stubService) ListExecPlanCases(ctx context.Context, query workerapp.ExecPlanCaseDrilldownQuery) (workerapp.ExecPlanCaseDrilldownResult, error) {
	return s.drilldownFn(ctx, query)
}

func (s stubService) FleetSummary(ctx context.Context) (workerapp.FleetSummary, error) {
	return s.summaryFn(ctx)
}

func (s stubService) ListAlerts(ctx context.Context, query workerapp.AlertListQuery) (workerapp.AlertListResult, error) {
	return s.listAlertsFn(ctx, query)
}

func TestRegisterWorkerRejectsInvalidJSON(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return workerapp.RegisterWorkerResult{}, nil
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
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return workerapp.RegisterWorkerResult{}, nil
	}})

	body := `{"namespace":"sim","pod_name":"worker-1","container_name":"worker"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestRegisterWorkerRequiresConfiguredAuth(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return workerapp.RegisterWorkerResult{}, nil
	}}, WithWorkerIngressAuth(workerapp.WorkerIngressAuthConfig{Token: "secret"}))

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeUnauthorized)
	assertErrorMessage(t, rec.Body.Bytes(), "worker ingress authentication required")
}

func TestRegisterWorkerRejectsMalformedAuth(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return workerapp.RegisterWorkerResult{}, nil
	}}, WithWorkerIngressAuth(workerapp.WorkerIngressAuthConfig{Token: "secret"}))

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	req.Header.Set("Authorization", "Basic secret")
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeUnauthorized)
}

func TestRegisterWorkerRejectsInvalidAuth(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		t.Fatalf("register should not be called")
		return workerapp.RegisterWorkerResult{}, nil
	}}, WithWorkerIngressAuth(workerapp.WorkerIngressAuthConfig{Token: "secret"}))

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	req.Header.Set("Authorization", "Bearer wrong")
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeUnauthorized)
}

func TestRegisterWorkerAcceptsValidAuth(t *testing.T) {
	observedAt := time.Date(2026, 4, 19, 10, 0, 0, 0, time.UTC)
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		if input.Identity.WorkerID != "worker-1" {
			t.Fatalf("worker_id = %q, want worker-1", input.Identity.WorkerID)
		}
		return workerapp.RegisterWorkerResult{
			WorkerID:     "worker-1",
			Status:       domain.WorkerStatusUnknown,
			RegisteredAt: observedAt,
		}, nil
	}}, WithWorkerIngressAuth(workerapp.WorkerIngressAuthConfig{Token: "secret"}))

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	req.Header.Set("Authorization", "Bearer secret")
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestHeartbeatWorkerAcceptsMultiTaskPayload(t *testing.T) {
	observedAt := time.Date(2026, 4, 19, 11, 0, 0, 0, time.UTC)
	h := New(stubService{heartbeatFn: func(ctx context.Context, input workerapp.HeartbeatWorkerInput) (workerapp.HeartbeatWorkerResult, error) {
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
		return workerapp.HeartbeatWorkerResult{
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
	var envelope testResponseEnvelope[struct {
		ActiveTaskCount int `json:"active_task_count"`
	}]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Data.ActiveTaskCount != 2 {
		t.Fatalf("active_task_count = %#v, want 2", envelope.Data.ActiveTaskCount)
	}
}

func TestHeartbeatWorkerRequiresConfiguredAuth(t *testing.T) {
	h := New(stubService{heartbeatFn: func(ctx context.Context, input workerapp.HeartbeatWorkerInput) (workerapp.HeartbeatWorkerResult, error) {
		t.Fatalf("heartbeat should not be called")
		return workerapp.HeartbeatWorkerResult{}, nil
	}}, WithWorkerIngressAuth(workerapp.WorkerIngressAuthConfig{Token: "secret"}))

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/heartbeat", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"process_alive":true,
		"accepting_tasks":true
	}`))
	rec := httptest.NewRecorder()

	h.HeartbeatWorker(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeUnauthorized)
}

func TestListWorkersRejectsInvalidAcceptingTasksQuery(t *testing.T) {
	h := New(stubService{listWorkersFn: func(ctx context.Context, query workerapp.WorkerListQuery) (workerapp.WorkerListResult, error) {
		t.Fatalf("list workers should not be called")
		return workerapp.WorkerListResult{}, nil
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
	h := New(stubService{getWorkerFn: func(ctx context.Context, workerID domain.WorkerID) (workerapp.WorkerDetail, error) {
		return workerapp.WorkerDetail{}, workerapp.ErrNotFound
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
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeResourceNotFound)
	assertErrorMessage(t, rec.Body.Bytes(), "workerfleet resource not found")
}

func TestFleetSummarySuccess(t *testing.T) {
	h := New(stubService{summaryFn: func(ctx context.Context) (workerapp.FleetSummary, error) {
		return workerapp.FleetSummary{
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

func TestGetCaseTimelineSuccess(t *testing.T) {
	now := time.Date(2026, 4, 21, 10, 0, 0, 0, time.UTC)
	h := New(stubService{timelineFn: func(ctx context.Context, taskID domain.TaskID) (workerapp.CaseTimelineResult, error) {
		if taskID != "case-1" {
			t.Fatalf("task_id = %q, want case-1", taskID)
		}
		return workerapp.CaseTimelineResult{
			TaskID: "case-1",
			Items: []workerapp.CaseStepView{{
				TaskID:     "case-1",
				ExecPlanID: "plan-1",
				Step:       "simulate",
				Status:     domain.CaseStepStatusSucceeded,
				ObservedAt: now,
			}},
		}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/case-1/timeline", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		Params: map[string]string{"task_id": "case-1"},
	}))
	rec := httptest.NewRecorder()

	h.GetCaseTimeline(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var envelope testResponseEnvelope[workerapp.CaseTimelineResult]
	if err := json.Unmarshal(rec.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if envelope.Data.TaskID != "case-1" {
		t.Fatalf("task_id = %#v, want case-1", envelope.Data.TaskID)
	}
}

func TestGetTaskOmitsEmptyCurrentStep(t *testing.T) {
	h := New(stubService{getTaskFn: func(ctx context.Context, taskID domain.TaskID) (workerapp.TaskDetail, error) {
		if taskID != "task-1" {
			t.Fatalf("task_id = %q, want task-1", taskID)
		}
		return workerapp.TaskDetail{
			TaskID:     "task-1",
			WorkerID:   "worker-1",
			ExecPlanID: "plan-1",
			TaskType:   "simulation",
			Phase:      "running",
			PhaseName:  "running",
			Status:     "active",
		}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task-1", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		Params: map[string]string{"task_id": "task-1"},
	}))
	rec := httptest.NewRecorder()

	h.GetTask(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if bytes.Contains(rec.Body.Bytes(), []byte(`"current_step"`)) {
		t.Fatalf("response should omit empty current_step: %s", rec.Body.String())
	}
}

func TestListExecPlanCasesParsesFilters(t *testing.T) {
	h := New(stubService{drilldownFn: func(ctx context.Context, query workerapp.ExecPlanCaseDrilldownQuery) (workerapp.ExecPlanCaseDrilldownResult, error) {
		if query.ExecPlanID != "plan-1" {
			t.Fatalf("exec_plan_id = %q, want plan-1", query.ExecPlanID)
		}
		if query.NodeName != "node-a" || query.PodName != "pod-a" || query.Step != "simulate" {
			t.Fatalf("unexpected query %#v", query)
		}
		if query.Page != 2 || query.PageSize != 10 {
			t.Fatalf("pagination = %d/%d, want 2/10", query.Page, query.PageSize)
		}
		return workerapp.ExecPlanCaseDrilldownResult{
			ExecPlanID: "plan-1",
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      0,
		}, nil
	}})

	req := httptest.NewRequest(http.MethodGet, "/v1/exec-plans/plan-1/cases?node_name=node-a&pod_name=pod-a&step=simulate&page=2&page_size=10", nil)
	req = req.WithContext(contract.WithRequestContext(req.Context(), contract.RequestContext{
		Params: map[string]string{"exec_plan_id": "plan-1"},
	}))
	rec := httptest.NewRecorder()

	h.ListExecPlanCases(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func assertErrorCode(t *testing.T, body []byte, code string) {
	t.Helper()

	var payload struct {
		Error struct {
			Code     string                 `json:"code"`
			Message  string                 `json:"message"`
			Category contract.ErrorCategory `json:"category"`
			Type     contract.ErrorType     `json:"type,omitempty"`
			Details  map[string]any         `json:"details,omitempty"`
		} `json:"error"`
		RequestID string `json:"request_id,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if payload.Error.Code != code {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, code)
	}
}

func TestServiceErrorConflictMapsToConflict(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		return workerapp.RegisterWorkerResult{}, fmt.Errorf("wrapped: %w", workerapp.ErrConflict)
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
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeConflict)
	assertErrorMessage(t, rec.Body.Bytes(), "workerfleet conflict")
}

func TestServiceErrorInternalUsesSafeMessage(t *testing.T) {
	h := New(stubService{registerFn: func(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error) {
		return workerapp.RegisterWorkerResult{}, fmt.Errorf("database password leaked")
	}})

	req := httptest.NewRequest(http.MethodPost, "/v1/workers/register", bytes.NewBufferString(`{
		"worker_id":"worker-1",
		"namespace":"sim",
		"pod_name":"worker-1",
		"container_name":"worker"
	}`))
	rec := httptest.NewRecorder()

	h.RegisterWorker(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	assertErrorCode(t, rec.Body.Bytes(), contract.CodeInternalError)
	assertErrorMessage(t, rec.Body.Bytes(), "workerfleet service unavailable")
}

func TestNotConfiguredErrorsUseUppercaseCodes(t *testing.T) {
	tests := []struct {
		name     string
		call     func(*Handler, http.ResponseWriter, *http.Request)
		path     string
		wantCode string
	}{
		{
			name:     "register",
			call:     (*Handler).RegisterWorker,
			path:     "/v1/workers/register",
			wantCode: "REGISTER_SERVICE_NOT_CONFIGURED",
		},
		{
			name:     "list workers",
			call:     (*Handler).ListWorkers,
			path:     "/v1/workers",
			wantCode: "LIST_WORKERS_NOT_CONFIGURED",
		},
		{
			name:     "fleet summary",
			call:     (*Handler).FleetSummary,
			path:     "/v1/fleet/summary",
			wantCode: "FLEET_SUMMARY_NOT_CONFIGURED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := New(nil)
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()

			tt.call(h, rec, req)

			if rec.Code != http.StatusNotImplemented {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotImplemented)
			}
			assertErrorCode(t, rec.Body.Bytes(), tt.wantCode)
		})
	}
}

func assertErrorMessage(t *testing.T, body []byte, message string) {
	t.Helper()

	var payload struct {
		Error struct {
			Code     string                 `json:"code"`
			Message  string                 `json:"message"`
			Category contract.ErrorCategory `json:"category"`
			Type     contract.ErrorType     `json:"type,omitempty"`
			Details  map[string]any         `json:"details,omitempty"`
		} `json:"error"`
		RequestID string `json:"request_id,omitempty"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal error response: %v", err)
	}
	if payload.Error.Message != message {
		t.Fatalf("error message = %q, want %q", payload.Error.Message, message)
	}
}
