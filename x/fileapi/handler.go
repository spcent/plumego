// Package fileapi provides an HTTP handler for tenant-aware file operations.
// It composes x/data/file storage and metadata implementations with the
// contract error model.
//
// Tenant identity must be attached via x/tenant/core.WithTenantID before
// reaching any handler method — typically by middleware in the calling
// application. User identity is optional and may be attached via WithUserID.
package fileapi

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/spcent/plumego/contract"
	storefile "github.com/spcent/plumego/store/file"
	datafile "github.com/spcent/plumego/x/data/file"
	tenantcore "github.com/spcent/plumego/x/tenant/core"
)

// Handler provides HTTP endpoints for file operations.
type Handler struct {
	storage  datafile.Storage
	metadata datafile.MetadataManager
	maxSize  int64
}

// NewHandler creates a new file handler with a default maximum upload size of
// 100 MiB.
func NewHandler(storage datafile.Storage, metadata datafile.MetadataManager) *Handler {
	return &Handler{
		storage:  storage,
		metadata: metadata,
		maxSize:  100 << 20,
	}
}

// WithMaxSize sets the maximum allowed file size for uploads.
func (h *Handler) WithMaxSize(size int64) *Handler {
	h.maxSize = size
	return h
}

// Upload handles file upload via multipart form.
// POST /files
// Form fields: file (required), generate_thumb, thumb_width, thumb_height.
// Requires tenant identity in request context.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing tenant id in context")
		return
	}

	userID := UserIDFromContext(ctx)

	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		h.writeError(w, r, http.StatusBadRequest, fmt.Sprintf("failed to parse form: %v", err))
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, r, http.StatusBadRequest, fmt.Sprintf("missing file: %v", err))
		return
	}
	defer f.Close()

	generateThumb := r.FormValue("generate_thumb") == "true"
	thumbWidth, _ := strconv.Atoi(r.FormValue("thumb_width"))
	thumbHeight, _ := strconv.Atoi(r.FormValue("thumb_height"))

	opts := datafile.PutOptions{
		TenantID:      tenantID,
		Reader:        f,
		FileName:      header.Filename,
		ContentType:   header.Header.Get("Content-Type"),
		GenerateThumb: generateThumb,
		ThumbWidth:    thumbWidth,
		ThumbHeight:   thumbHeight,
		UploadedBy:    userID,
		Metadata:      make(map[string]any),
	}

	result, err := h.storage.Put(ctx, opts)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("upload failed: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Download streams file content.
// GET /files/{id}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing file id")
		return
	}

	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		h.writeMetadataError(w, r, err)
		return
	}

	go h.metadata.UpdateAccessTime(context.Background(), fileID)

	reader, err := h.storage.Get(ctx, fileMeta.Path)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to read file: %v", err))
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", fileMeta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileMeta.Size))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileMeta.Name))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

// GetInfo returns file metadata.
// GET /files/{id}/info
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing file id")
		return
	}

	fileMeta, err := h.metadata.Get(r.Context(), fileID)
	if err != nil {
		h.writeMetadataError(w, r, err)
		return
	}

	h.writeJSON(w, http.StatusOK, fileMeta)
}

// Delete soft-deletes a file.
// DELETE /files/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing file id")
		return
	}

	if err := h.metadata.Delete(r.Context(), fileID); err != nil {
		h.writeMetadataError(w, r, err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "file deleted"})
}

// List returns a paginated list of files.
// GET /files?page=1&page_size=20&tenant_id=xxx&mime_type=image/jpeg&uploaded_by=user1
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := datafile.Query{Page: 1, PageSize: 20}

	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			query.Page = v
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			query.PageSize = v
		}
	}

	query.TenantID = r.URL.Query().Get("tenant_id")
	query.UploadedBy = r.URL.Query().Get("uploaded_by")
	query.MimeType = r.URL.Query().Get("mime_type")
	query.OrderBy = r.URL.Query().Get("order_by")

	if s := r.URL.Query().Get("start_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			query.StartTime = t
		}
	}

	if s := r.URL.Query().Get("end_time"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			query.EndTime = t
		}
	}

	files, total, err := h.metadata.List(ctx, query)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("list failed: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]any{
		"items":      files,
		"total":      total,
		"page":       query.Page,
		"page_size":  query.PageSize,
		"total_page": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	})
}

// GetURL returns a temporary access URL for the file.
// GET /files/{id}/url?expiry=3600
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, r, http.StatusBadRequest, "missing file id")
		return
	}

	expiry := 15 * time.Minute
	if s := r.URL.Query().Get("expiry"); s != "" {
		if seconds, err := strconv.Atoi(s); err == nil && seconds > 0 {
			expiry = time.Duration(seconds) * time.Second
		}
	}

	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		h.writeMetadataError(w, r, err)
		return
	}

	fileURL, err := h.storage.GetURL(ctx, fileMeta.Path, expiry)
	if err != nil {
		h.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("failed to generate url: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"url":        html.EscapeString(fileURL),
		"expires_in": strconv.Itoa(int(expiry.Seconds())),
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	_ = contract.WriteJSON(w, status, data)
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, status int, message string) {
	contract.WriteError(w, r, contract.NewErrorBuilder().
		Status(status).
		Code(http.StatusText(status)).
		Message(message).
		Category(contract.CategoryForStatus(status)).
		Build())
}

func (h *Handler) writeMetadataError(w http.ResponseWriter, r *http.Request, err error) {
	if err == storefile.ErrNotFound {
		h.writeError(w, r, http.StatusNotFound, "file not found")
	} else {
		h.writeError(w, r, http.StatusInternalServerError, fmt.Sprintf("metadata error: %v", err))
	}
}
