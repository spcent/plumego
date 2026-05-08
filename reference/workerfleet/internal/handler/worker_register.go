package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"workerfleet/internal/domain"
)

var (
	ErrNotFound       = errors.New("workerfleet resource not found")
	ErrNotImplemented = errors.New("workerfleet operation not implemented")
	ErrConflict       = errors.New("workerfleet conflict")
)

type Service interface {
	RegisterWorker(ctx context.Context, input RegisterWorkerInput) (RegisterWorkerResult, error)
	HeartbeatWorker(ctx context.Context, input HeartbeatWorkerInput) (HeartbeatWorkerResult, error)
	ListWorkers(ctx context.Context, query WorkerListQuery) (WorkerListResult, error)
	GetWorker(ctx context.Context, workerID domain.WorkerID) (WorkerDetail, error)
	GetTask(ctx context.Context, taskID domain.TaskID) (TaskDetail, error)
	GetCaseTimeline(ctx context.Context, taskID domain.TaskID) (CaseTimelineResult, error)
	ListExecPlanCases(ctx context.Context, query ExecPlanCaseDrilldownQuery) (ExecPlanCaseDrilldownResult, error)
	FleetSummary(ctx context.Context) (FleetSummary, error)
	ListAlerts(ctx context.Context, query AlertListQuery) (AlertListResult, error)
}

type Handler struct {
	service Service
}

func New(service Service) *Handler {
	return &Handler{service: service}
}

type RegisterWorkerRequest struct {
	WorkerID      string    `json:"worker_id"`
	Namespace     string    `json:"namespace"`
	PodName       string    `json:"pod_name"`
	PodUID        string    `json:"pod_uid,omitempty"`
	NodeName      string    `json:"node_name,omitempty"`
	ContainerName string    `json:"container_name"`
	Image         string    `json:"image,omitempty"`
	Version       string    `json:"version,omitempty"`
	ObservedAt    time.Time `json:"observed_at,omitempty"`
}

type RegisterWorkerInput struct {
	Identity   domain.WorkerIdentity
	ObservedAt time.Time
}

type RegisterWorkerResult struct {
	WorkerID     string              `json:"worker_id"`
	Status       domain.WorkerStatus `json:"status"`
	RegisteredAt time.Time           `json:"registered_at"`
}

func (h *Handler) RegisterWorker(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		writeNotImplemented(w, r, "REGISTER_SERVICE_NOT_CONFIGURED", "register worker service not configured")
		return
	}

	var req RegisterWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvalidJSON(w, r)
		return
	}
	if strings.TrimSpace(req.WorkerID) == "" {
		writeRequiredJSONField(w, r, "worker_id")
		return
	}
	if strings.TrimSpace(req.Namespace) == "" {
		writeRequiredJSONField(w, r, "namespace")
		return
	}
	if strings.TrimSpace(req.PodName) == "" {
		writeRequiredJSONField(w, r, "pod_name")
		return
	}
	if strings.TrimSpace(req.ContainerName) == "" {
		writeRequiredJSONField(w, r, "container_name")
		return
	}

	result, err := h.service.RegisterWorker(r.Context(), RegisterWorkerInput{
		Identity: domain.WorkerIdentity{
			WorkerID:      domain.WorkerID(req.WorkerID),
			Namespace:     strings.TrimSpace(req.Namespace),
			PodName:       strings.TrimSpace(req.PodName),
			PodUID:        domain.PodUID(strings.TrimSpace(req.PodUID)),
			NodeName:      strings.TrimSpace(req.NodeName),
			ContainerName: strings.TrimSpace(req.ContainerName),
			Image:         strings.TrimSpace(req.Image),
			Version:       strings.TrimSpace(req.Version),
		},
		ObservedAt: req.ObservedAt,
	})
	if err != nil {
		writeServiceError(w, r, err)
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusCreated, result, nil)
}

func writeInvalidJSON(w http.ResponseWriter, r *http.Request) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeValidation).
		Code(contract.CodeInvalidJSON).
		Message("invalid request body").
		Build())
}

func writeRequiredJSONField(w http.ResponseWriter, r *http.Request, field string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(contract.CodeRequired).
		Message(field+" is required").
		Detail("field", field).
		Build())
}

func writeNotImplemented(w http.ResponseWriter, r *http.Request, code string, message string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotImplemented).
		Code(strings.ToUpper(code)).
		Message(message).
		Build())
}

func writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(contract.CodeResourceNotFound).
			Message("workerfleet resource not found").
			Build())
	case errors.Is(err, ErrConflict):
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeConflict).
			Code(contract.CodeConflict).
			Message("workerfleet conflict").
			Build())
	case errors.Is(err, ErrNotImplemented):
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotImplemented).
			Code(contract.CodeNotImplemented).
			Message("workerfleet operation not implemented").
			Build())
	default:
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(contract.CodeInternalError).
			Message("workerfleet service unavailable").
			Build())
	}
}
