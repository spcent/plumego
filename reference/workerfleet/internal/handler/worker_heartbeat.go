package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	"workerfleet/internal/domain"
)

type HeartbeatTask struct {
	TaskID      string            `json:"task_id" validate:"required"`
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
	WorkerID       string          `json:"worker_id" validate:"required"`
	ProcessAlive   bool            `json:"process_alive"`
	AcceptingTasks bool            `json:"accepting_tasks"`
	ObservedAt     time.Time       `json:"observed_at,omitempty"`
	LastError      string          `json:"last_error,omitempty"`
	ActiveTasks    []HeartbeatTask `json:"active_tasks,omitempty"`
}

type HeartbeatWorkerInput struct {
	WorkerID       domain.WorkerID
	ProcessAlive   bool
	AcceptingTasks bool
	ObservedAt     time.Time
	LastError      string
	ActiveTasks    []domain.TaskReport
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

	var req HeartbeatWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidJSON(w, r)
		return
	}
	if err := contract.ValidateStruct(&req); err != nil {
		_ = contract.WriteBindError(w, r, err)
		return
	}
	for _, task := range req.ActiveTasks {
		if err := contract.ValidateStruct(&task); err != nil {
			_ = contract.WriteBindError(w, r, err)
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

	result, err := h.service.HeartbeatWorker(r.Context(), HeartbeatWorkerInput{
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

	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}
