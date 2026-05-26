package handler

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	workerapp "workerfleet/internal/app"
	"workerfleet/internal/domain"
)

type Service interface {
	RegisterWorker(ctx context.Context, input workerapp.RegisterWorkerInput) (workerapp.RegisterWorkerResult, error)
	HeartbeatWorker(ctx context.Context, input workerapp.HeartbeatWorkerInput) (workerapp.HeartbeatWorkerResult, error)
	ListWorkers(ctx context.Context, query workerapp.WorkerListQuery) (workerapp.WorkerListResult, error)
	GetWorker(ctx context.Context, workerID domain.WorkerID) (workerapp.WorkerView, error)
	GetTask(ctx context.Context, taskID domain.TaskID) (workerapp.TaskDetail, error)
	GetCaseTimeline(ctx context.Context, taskID domain.TaskID) (workerapp.CaseTimelineResult, error)
	ListExecPlanCases(ctx context.Context, query workerapp.ExecPlanCaseDrilldownQuery) (workerapp.ExecPlanCaseDrilldownResult, error)
	FleetSummary(ctx context.Context) (workerapp.FleetSummary, error)
	ListAlerts(ctx context.Context, query workerapp.AlertListQuery) (workerapp.AlertListResult, error)
}

// Handler serves all workerfleet HTTP endpoints.
// Logger should be set via WithLogger; when nil, write errors are silently dropped.
type Handler struct {
	service    Service
	workerAuth workerapp.WorkerIngressAuthConfig
	adminAuth  workerapp.AdminAuthConfig
	Logger     plumelog.StructuredLogger
}

type Option func(*Handler)

func WithWorkerIngressAuth(auth workerapp.WorkerIngressAuthConfig) Option {
	return func(h *Handler) {
		h.workerAuth = auth
	}
}

func WithAdminAuth(auth workerapp.AdminAuthConfig) Option {
	return func(h *Handler) {
		h.adminAuth = auth
	}
}

func WithLogger(logger plumelog.StructuredLogger) Option {
	return func(h *Handler) {
		h.Logger = logger
	}
}

func New(service Service, opts ...Option) *Handler {
	h := &Handler{service: service}
	for _, opt := range opts {
		if opt != nil {
			opt(h)
		}
	}
	return h
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

type RegisterWorkerResult struct {
	WorkerID     string              `json:"worker_id"`
	Status       domain.WorkerStatus `json:"status"`
	RegisteredAt time.Time           `json:"registered_at"`
}

func (h *Handler) RegisterWorker(w http.ResponseWriter, r *http.Request) {
	if h.service == nil {
		h.writeNotImplemented(w, r, "REGISTER_SERVICE_NOT_CONFIGURED", "register worker service not configured")
		return
	}
	if !h.requireWorkerIngressAuth(w, r) {
		return
	}

	var req RegisterWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeInvalidJSON(w, r)
		return
	}
	if strings.TrimSpace(req.WorkerID) == "" {
		h.writeRequiredJSONField(w, r, "worker_id")
		return
	}
	if strings.TrimSpace(req.Namespace) == "" {
		h.writeRequiredJSONField(w, r, "namespace")
		return
	}
	if strings.TrimSpace(req.PodName) == "" {
		h.writeRequiredJSONField(w, r, "pod_name")
		return
	}
	if strings.TrimSpace(req.ContainerName) == "" {
		h.writeRequiredJSONField(w, r, "container_name")
		return
	}

	result, err := h.service.RegisterWorker(r.Context(), workerapp.RegisterWorkerInput{
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
		h.writeServiceError(w, r, err)
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, registerWorkerResultFromApp(result), nil))
}

func (h *Handler) requireWorkerIngressAuth(w http.ResponseWriter, r *http.Request) bool {
	token := strings.TrimSpace(h.workerAuth.Token)
	if token == "" {
		return true
	}

	got, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok || !constantTimeTokenEqual(got, token) {
		h.writeWorkerAuthError(w, r)
		return false
	}
	return true
}

func (h *Handler) requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	token := strings.TrimSpace(h.adminAuth.Token)
	if !h.adminAuth.Required && token == "" {
		return true
	}
	if token == "" {
		h.writeAdminAuthError(w, r)
		return false
	}

	got, ok := bearerToken(r.Header.Get("Authorization"))
	if !ok || !constantTimeTokenEqual(got, token) {
		h.writeAdminAuthError(w, r)
		return false
	}
	return true
}

func bearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	if token == "" || strings.Contains(token, " ") {
		return "", false
	}
	return token, true
}

func constantTimeTokenEqual(got string, want string) bool {
	gotHash := sha256.Sum256([]byte(got))
	wantHash := sha256.Sum256([]byte(want))
	return subtle.ConstantTimeCompare(gotHash[:], wantHash[:]) == 1
}

func (h *Handler) writeWorkerAuthError(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeUnauthorized).
		Code(contract.CodeUnauthorized).
		Message("worker ingress authentication required").
		Build()))
}

func (h *Handler) writeAdminAuthError(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeUnauthorized).
		Code(contract.CodeUnauthorized).
		Message("workerfleet admin authentication required").
		Build()))
}

func registerWorkerResultFromApp(result workerapp.RegisterWorkerResult) RegisterWorkerResult {
	return RegisterWorkerResult{
		WorkerID:     result.WorkerID,
		Status:       result.Status,
		RegisteredAt: result.RegisteredAt,
	}
}

func (h *Handler) writeInvalidJSON(w http.ResponseWriter, r *http.Request) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeValidation).
		Code(contract.CodeInvalidJSON).
		Message("invalid request body").
		Build()))
}

func (h *Handler) writeRequiredJSONField(w http.ResponseWriter, r *http.Request, field string) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeRequired).
		Code(contract.CodeRequired).
		Message(field+" is required").
		Detail("field", field).
		Build()))
}

func (h *Handler) writeNotImplemented(w http.ResponseWriter, r *http.Request, code string, message string) {
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeNotImplemented).
		Code(strings.ToUpper(code)).
		Message(message).
		Build()))
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, workerapp.ErrNotFound):
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Code(contract.CodeResourceNotFound).
			Message("workerfleet resource not found").
			Build()))
	case errors.Is(err, workerapp.ErrConflict):
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeConflict).
			Code(contract.CodeConflict).
			Message("workerfleet conflict").
			Build()))
	case errors.Is(err, workerapp.ErrNotImplemented):
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotImplemented).
			Code(contract.CodeNotImplemented).
			Message("workerfleet operation not implemented").
			Build()))
	default:
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code(contract.CodeInternalError).
			Message("workerfleet service unavailable").
			Build()))
	}
}
