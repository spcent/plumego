// Package fileapi provides an HTTP handler for tenant-aware file operations.
// It composes x/data/file storage and metadata implementations with the
// contract error model.
//
// Tenant identity must be attached via x/tenant/core.WithTenantID before
// reaching any handler method — typically by middleware in the calling
// application. User identity is optional and may be attached via WithUserID.
package fileapi

import (
	"errors"
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
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing tenant id in context").
			Category(contract.CategoryValidation).
			Build())
		return
	}

	userID := UserIDFromContext(ctx)

	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code(contract.CodeBadRequest).
			Message("invalid multipart form").
			Build())
		return
	}

	f, header, err := r.FormFile("file")
	if err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code(contract.CodeRequired).
			Message("file field is required").
			Build())
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
		writeFileInternalError(w, r, "upload failed")
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, result, nil)
}

// Download streams file content.
// GET /files/{id}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing file id").
			Category(contract.CategoryValidation).
			Build())
		return
	}

	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		writeFileMetadataError(w, r, err)
		return
	}

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" || fileMeta.TenantID != tenantID {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeForbidden).
			Message("access denied").
			Build())
		return
	}

	_ = h.metadata.UpdateAccessTime(ctx, fileID)

	reader, err := h.storage.Get(ctx, fileMeta.Path)
	if err != nil {
		writeFileInternalError(w, r, "failed to read file")
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", fileMeta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileMeta.Size))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", fileMeta.Name))
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, reader)
}

// GetInfo returns file metadata.
// GET /files/{id}/info
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing file id").
			Category(contract.CategoryValidation).
			Build())
		return
	}

	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		writeFileMetadataError(w, r, err)
		return
	}

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" || fileMeta.TenantID != tenantID {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeForbidden).
			Message("access denied").
			Build())
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, fileMeta, nil)
}

// Delete soft-deletes a file.
// DELETE /files/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing file id").
			Category(contract.CategoryValidation).
			Build())
		return
	}

	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		writeFileMetadataError(w, r, err)
		return
	}

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" || fileMeta.TenantID != tenantID {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeForbidden).
			Message("access denied").
			Build())
		return
	}

	if err := h.metadata.Delete(ctx, fileID); err != nil {
		writeFileMetadataError(w, r, err)
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"message": "file deleted"}, nil)
}

// List returns a paginated list of files.
// GET /files?page=1&page_size=20&mime_type=image/jpeg&uploaded_by=user1
// Requires tenant identity in request context.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	query := datafile.Query{Page: 1, PageSize: 20}

	if p := r.URL.Query().Get("page"); p != "" {
		v, err := parsePositiveIntParam("page", p, 0)
		if err != nil {
			writeInvalidQueryError(w, r, "page")
			return
		}
		query.Page = v
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		v, err := parsePositiveIntParam("page_size", ps, 100)
		if err != nil {
			writeInvalidQueryError(w, r, "page_size")
			return
		}
		query.PageSize = v
	}

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing tenant id in context").
			Category(contract.CategoryValidation).
			Build())
		return
	}
	query.TenantID = tenantID
	query.UploadedBy = r.URL.Query().Get("uploaded_by")
	query.MimeType = r.URL.Query().Get("mime_type")
	query.OrderBy = r.URL.Query().Get("order_by")

	if s := r.URL.Query().Get("start_time"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeInvalidQueryError(w, r, "start_time")
			return
		}
		query.StartTime = t
	}

	if s := r.URL.Query().Get("end_time"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeInvalidQueryError(w, r, "end_time")
			return
		}
		query.EndTime = t
	}

	files, total, err := h.metadata.List(ctx, query)
	if err != nil {
		writeFileInternalError(w, r, "list failed")
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"items":      files,
		"total":      total,
		"page":       query.Page,
		"page_size":  query.PageSize,
		"total_page": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	}, nil)
}

// GetURL returns a temporary access URL for the file.
// GET /files/{id}/url?expiry=3600
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code(contract.CodeBadRequest).
			Message("missing file id").
			Category(contract.CategoryValidation).
			Build())
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
		writeFileMetadataError(w, r, err)
		return
	}

	tenantID := tenantcore.TenantIDFromContext(ctx)
	if tenantID == "" || fileMeta.TenantID != tenantID {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeForbidden).
			Message("access denied").
			Build())
		return
	}

	fileURL, err := h.storage.GetURL(ctx, fileMeta.Path, expiry)
	if err != nil {
		writeFileInternalError(w, r, "failed to generate url")
		return
	}

	_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
		"url":        html.EscapeString(fileURL),
		"expires_in": strconv.Itoa(int(expiry.Seconds())),
	}, nil)
}

func writeFileMetadataError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, storefile.ErrNotFound) {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("file not found").
			Build())
	} else {
		writeFileInternalError(w, r, "metadata error")
	}
}

func parsePositiveIntParam(field, value string, max int) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return 0, fmt.Errorf("invalid %s", field)
	}
	if max > 0 && n > max {
		return 0, fmt.Errorf("%s out of range", field)
	}
	return n, nil
}

func writeInvalidQueryError(w http.ResponseWriter, r *http.Request, field string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeValidation).
		Code(contract.CodeInvalidQuery).
		Message("invalid query parameter").
		Detail("field", field).
		Build())
}

func writeFileInternalError(w http.ResponseWriter, r *http.Request, message string) {
	_ = contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).
		Code(contract.CodeInternalError).
		Message(message).
		Build())
}
