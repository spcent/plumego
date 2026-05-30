package update

import (
	"encoding/json"
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler provides HTTP handlers for update endpoints.
type Handler struct {
	service *Service
	logger  plumelog.StructuredLogger
}

// NewHandler creates a new update handler.
func NewHandler(service *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GetStatus returns the current update status.
// GET /api/v1/system/update/status
func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.service.GetStatus(r.Context())

	if err := contract.WriteResponse(w, r, http.StatusOK, status, nil); err != nil {
		h.logger.Error("failed to write update status response", plumelog.Fields{
			"error": err.Error(),
		})
	}
}

// CheckNow performs an immediate update check.
// POST /api/v1/system/update/check
func (h *Handler) CheckNow(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.CheckNow(r.Context())
	if err != nil {
		if werr := contract.WriteError(w, r,
			contract.NewErrorBuilder().
				Type(contract.TypeInternal).
				Message("update check failed: "+err.Error()).
				Build()); werr != nil {
			h.logger.Error("failed to write error response", plumelog.Fields{
				"error": werr.Error(),
			})
		}
		return
	}

	if err := contract.WriteResponse(w, r, http.StatusOK, status, nil); err != nil {
		h.logger.Error("failed to write update check response", plumelog.Fields{
			"error": err.Error(),
		})
	}
}

// GetVersion returns the current application version.
// GET /api/v1/system/version
func (h *Handler) GetVersion(w http.ResponseWriter, r *http.Request) {
	if err := contract.WriteResponse(w, r, http.StatusOK, h.service.checker.currentVersion, nil); err != nil {
		h.logger.Error("failed to write version response", plumelog.Fields{
			"error": err.Error(),
		})
	}
}

// SetEnabled enables or disables update checking.
// POST /api/v1/system/update/config
func (h *Handler) SetEnabled(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		if werr := contract.WriteError(w, r,
			contract.NewErrorBuilder().
				Type(contract.TypeInvalidRequest).
				Message("invalid request body").
				Build()); werr != nil {
			h.logger.Error("failed to write error response", plumelog.Fields{
				"error": werr.Error(),
			})
		}
		return
	}

	h.service.config.Enabled = req.Enabled

	h.logger.Info("update check configuration updated", plumelog.Fields{
		"enabled": req.Enabled,
	})

	if err := contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"enabled": req.Enabled,
	}, nil); err != nil {
		h.logger.Error("failed to write config response", plumelog.Fields{
			"error": err.Error(),
		})
	}
}
