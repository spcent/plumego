package system

import (
	"encoding/json"
	"net/http"

	"cloud-vault/internal/version"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler exposes system observability endpoints.
type Handler struct {
	svc    *Service
	logger plumelog.StructuredLogger
}

func NewHandler(svc *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GET /api/v1/system/health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	result := h.svc.Health(r.Context())
	if result.Status == StatusError {
		if err := contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeUnavailable).
			Message("system health check failed").
			Build()); err != nil && h.logger != nil {
			h.logger.Error("system health: write error", plumelog.Fields{"err": err.Error()})
		}
		return
	}
	if err := contract.WriteResponse(w, r, http.StatusOK, result, nil); err != nil && h.logger != nil {
		h.logger.Error("system health: write", plumelog.Fields{"err": err.Error()})
	}
}

// GET /api/v1/system/stats
func (h *Handler) Stats(w http.ResponseWriter, r *http.Request) {
	result, err := h.svc.Stats(r.Context())
	if err != nil {
		if werr := contract.WriteError(w, r,
			contract.NewErrorBuilder().Type("INTERNAL").Message(err.Error()).Build()); werr != nil && h.logger != nil {
			h.logger.Error("system stats: write error", plumelog.Fields{"err": werr.Error()})
		}
		return
	}
	if werr := contract.WriteResponse(w, r, http.StatusOK, result, nil); werr != nil && h.logger != nil {
		h.logger.Error("system stats: write", plumelog.Fields{"err": werr.Error()})
	}
}

// POST /api/v1/system/doctor
func (h *Handler) Doctor(w http.ResponseWriter, r *http.Request) {
	var req DoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is fine — means run all checks.
		req = DoctorRequest{}
	}

	result := h.svc.Doctor(r.Context(), req)
	if err := contract.WriteResponse(w, r, http.StatusOK, result, nil); err != nil && h.logger != nil {
		h.logger.Error("system doctor: write", plumelog.Fields{"err": err.Error()})
	}
}

// GET /api/v1/system/version
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	info := version.GetBuildInfo()
	if err := contract.WriteResponse(w, r, http.StatusOK, info, nil); err != nil && h.logger != nil {
		h.logger.Error("system version: write", plumelog.Fields{"err": err.Error()})
	}
}
