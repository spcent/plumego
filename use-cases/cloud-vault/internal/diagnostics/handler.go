package diagnostics

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
)

// Handler provides HTTP endpoints for diagnostic operations.
type Handler struct {
	service *Service
	logger  plumelog.StructuredLogger
}

// NewHandler creates a new diagnostics handler.
func NewHandler(service *Service, logger plumelog.StructuredLogger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// GenerateBundle handles POST /api/v1/diagnostics/generate
// It creates a new diagnostic bundle and returns metadata about the generated file.
func (h *Handler) GenerateBundle(w http.ResponseWriter, r *http.Request) {
	bundle, err := h.service.GenerateBundle(r.Context())
	if err != nil {
		h.logger.Error("failed to generate diagnostic bundle", plumelog.Fields{"error": err.Error()})
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("Failed to generate diagnostic bundle: "+err.Error()).
			Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusCreated, bundle, nil)
}

// DownloadBundle handles GET /api/v1/diagnostics/download/{filename}
// It streams the diagnostic bundle zip file to the client.
func (h *Handler) DownloadBundle(w http.ResponseWriter, r *http.Request) {
	filename := r.PathValue("filename")
	if filename == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInvalidRequest).
			Message("Missing filename parameter").
			Build())
		return
	}

	bundle, err := h.service.GetBundle(filename)
	if err != nil {
		h.logger.Error("diagnostic bundle not found", plumelog.Fields{
			"filename": filename,
			"error":    err.Error(),
		})
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("Diagnostic bundle not found").
			Build())
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+bundle.Filename+"\"")
	http.ServeFile(w, r, bundle.DownloadPath)
}

// ListBundles handles GET /api/v1/diagnostics
// It returns a list of all available diagnostic bundles.
func (h *Handler) ListBundles(w http.ResponseWriter, r *http.Request) {
	bundles, err := h.service.ListBundles()
	if err != nil {
		h.logger.Error("failed to list diagnostic bundles", plumelog.Fields{"error": err.Error()})
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("Failed to list diagnostic bundles: "+err.Error()).
			Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, bundles, nil)
}
