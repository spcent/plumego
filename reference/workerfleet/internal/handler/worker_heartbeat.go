package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/domain"
)

type HeartbeatTask struct {
	TaskID      string            `json:"task_id"`
	ExecPlanID  string            `json:"exec_plan_id,omitempty"`
	TaskType    string            `json:"task_type,omitempty"`
	Phase       domain.TaskPhase  `json:"phase,omitempty"`
	PhaseName   string            `json:"phase_name,omitempty"`
	CurrentStep HeartbeatCaseStep `json:"current_step,omitempty"`
	StartedAt   time.Time         `json:"started_at,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type HeartbeatCaseStep struct {
	Step       string                `json:"step,omitempty"`
	StepName   string                `json:"step_name,omitempty"`
	Status     domain.CaseStepStatus `json:"status,omitempty"`
	StartedAt  time.Time             `json:"started_at,omitempty"`
	UpdatedAt  time.Time             `json:"updated_at,omitempty"`
	FinishedAt time.Time             `json:"finished_at,omitempty"`
	Attempt    int                   `json:"attempt,omitempty"`
	ErrorClass string                `json:"error_class,omitempty"`
}

type HeartbeatWorkerRequest struct {
	WorkerID       string          `json:"worker_id"`
	ProcessAlive   bool            `json:"process_alive"`
	AcceptingTasks bool            `json:"accepting_tasks"`
	ObservedAt     time.Time       `json:"observed_at,omitempty"`
	LastError      string          `json:"last_error,omitempty"`
	ActiveTasks    []HeartbeatTask `json:"active_tasks,omitempty"`
}

type HeartbeatWorkerResult struct {
	WorkerID        string              `json:"worker_id"`
	Status          domain.WorkerStatus `json:"status"`
	StatusReason    string              `json:"status_reason,omitempty"`
	ObservedAt      time.Time           `json:"observed_at"`
	ActiveTaskCount int                 `json:"active_task_count"`
}

func (h *Handler) HeartbeatWorker(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "heartbeat_service_not_configured", "worker heartbeat service not configured")
		return
	}
	if !h.requireWorkerIngressAuth(w, r) {
		return
	}

	var req HeartbeatWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidJSON(w, r)
		return
	}
	if strings.TrimSpace(req.WorkerID) == "" {
		writeRequiredJSONField(w, r, "worker_id")
		return
	}
	for _, task := range req.ActiveTasks {
		if strings.TrimSpace(task.TaskID) == "" {
			writeRequiredJSONField(w, r, "task_id")
			return
		}
	}

	activeTasks := make([]domain.TaskReport, 0, len(req.ActiveTasks))
	for _, task := range req.ActiveTasks {
		activeTasks = append(activeTasks, domain.TaskReport{
			TaskID:     domain.TaskID(task.TaskID),
			ExecPlanID: domain.ExecPlanID(task.ExecPlanID),
			TaskType:   task.TaskType,
			Phase:      task.Phase,
			PhaseName:  task.PhaseName,
			CurrentStep: domain.CaseStepRuntime{
				Step:       task.CurrentStep.Step,
				StepName:   task.CurrentStep.StepName,
				Status:     task.CurrentStep.Status,
				StartedAt:  task.CurrentStep.StartedAt,
				UpdatedAt:  task.CurrentStep.UpdatedAt,
				FinishedAt: task.CurrentStep.FinishedAt,
				Attempt:    task.CurrentStep.Attempt,
				ErrorClass: task.CurrentStep.ErrorClass,
			},
			StartedAt: task.StartedAt,
			UpdatedAt: task.UpdatedAt,
			Metadata:  task.Metadata,
		})
	}

	result, err := h.service.HeartbeatWorker(r.Context(), workerapp.HeartbeatWorkerInput{
		WorkerID:       domain.WorkerID(req.WorkerID),
		ProcessAlive:   req.ProcessAlive,
		AcceptingTasks: req.AcceptingTasks,
		ObservedAt:     req.ObservedAt,
		LastError:      req.LastError,
		ActiveTasks:    activeTasks,
	})
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, heartbeatWorkerResultFromApp(result), nil)
}

func heartbeatWorkerResultFromApp(result workerapp.HeartbeatWorkerResult) HeartbeatWorkerResult {
	return HeartbeatWorkerResult{
		WorkerID:        result.WorkerID,
		Status:          result.Status,
		StatusReason:    result.StatusReason,
		ObservedAt:      result.ObservedAt,
		ActiveTaskCount: result.ActiveTaskCount,
	}
}
