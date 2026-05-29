package backup

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"
)

// Handler provides HTTP handlers for backup operations.
type Handler struct {
	service *Service
	dataDir string
	logger  plumelog.StructuredLogger
}

// NewHandler creates a new backup handler.
func NewHandler(service *Service, dataDir string, logger plumelog.StructuredLogger) *Handler {
	return &Handler{
		service: service,
		dataDir: dataDir,
		logger:  logger,
	}
}

func (h *Handler) logWriteErr(err error) {
	if err != nil && h.logger != nil {
		h.logger.Error("backup handler: write response", plumelog.Fields{"err": err.Error()})
	}
}

// CreateBackup handles POST /api/v1/system/backup
func (h *Handler) CreateBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateBackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}

	backup, err := h.service.CreateBackup(ctx, req.IncludeConfig)
	if err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeInternal).Message(fmt.Sprintf("create backup: %v", err)).Build()))
		return
	}

	h.logWriteErr(contract.WriteResponse(w, r, http.StatusCreated, CreateBackupResponse{
		Backup: *backup,
	}, nil))
}

// ListBackups handles GET /api/v1/system/backups
func (h *Handler) ListBackups(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	backups, err := h.service.ListBackups(ctx)
	if err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeInternal).Message(fmt.Sprintf("list backups: %v", err)).Build()))
		return
	}

	h.logWriteErr(contract.WriteResponse(w, r, http.StatusOK, ListBackupsResponse{
		Backups: backups,
	}, nil))
}

// DownloadBackup handles GET /api/v1/system/backups/:name/download
func (h *Handler) DownloadBackup(w http.ResponseWriter, r *http.Request) {
	name := router.Param(r, "name")

	if err := ValidateName(name); err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message(fmt.Sprintf("invalid backup name: %s", name)).Build()))
		return
	}

	backupPath, err := h.service.GetBackupPath(name)
	if err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeNotFound).Message(fmt.Sprintf("backup not found: %s", name)).Build()))
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
	http.ServeFile(w, r, backupPath)
}

// DeleteBackup handles DELETE /api/v1/system/backups/:name
func (h *Handler) DeleteBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := router.Param(r, "name")

	if err := ValidateName(name); err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message(fmt.Sprintf("invalid backup name: %s", name)).Build()))
		return
	}

	if err := h.service.DeleteBackup(ctx, name); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logWriteErr(contract.WriteError(w, r,
				contract.NewErrorBuilder().Type(contract.TypeNotFound).Message(fmt.Sprintf("backup not found: %s", name)).Build()))
			return
		}
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeInternal).Message(fmt.Sprintf("delete backup: %v", err)).Build()))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RestoreBackup handles POST /api/v1/system/restore
func (h *Handler) RestoreBackup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}

	if req.Confirm != "RESTORE" {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message("confirmation required: set confirm to 'RESTORE'").Build()))
		return
	}

	if err := ValidateName(req.BackupName); err != nil {
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeBadRequest).Message(fmt.Sprintf("invalid backup name: %s", req.BackupName)).Build()))
		return
	}

	// Restore is allowed after explicit confirmation; callers coordinate session drain before invoking it.

	if err := h.service.RestoreBackup(ctx, req.BackupName, h.dataDir); err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.logWriteErr(contract.WriteError(w, r,
				contract.NewErrorBuilder().Type(contract.TypeNotFound).Message(fmt.Sprintf("backup not found: %s", req.BackupName)).Build()))
			return
		}
		h.logWriteErr(contract.WriteError(w, r,
			contract.NewErrorBuilder().Type(contract.TypeInternal).Message(fmt.Sprintf("restore backup: %v", err)).Build()))
		return
	}

	h.logWriteErr(contract.WriteResponse(w, r, http.StatusOK, RestoreResponse{
		Message: "restore completed successfully",
	}, nil))
}
