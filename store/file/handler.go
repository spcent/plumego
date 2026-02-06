package file

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Handler provides HTTP endpoints for file operations.
type Handler struct {
	storage  Storage
	metadata MetadataManager
	maxSize  int64 // Max file size in bytes
}

// NewHandler creates a new file handler.
func NewHandler(storage Storage, metadata MetadataManager) *Handler {
	return &Handler{
		storage:  storage,
		metadata: metadata,
		maxSize:  100 << 20, // Default 100 MiB
	}
}

// WithMaxSize sets the maximum file size.
func (h *Handler) WithMaxSize(size int64) *Handler {
	h.maxSize = size
	return h
}

// RegisterRoutes registers file handling routes to a router.
// This method is compatible with plumego's router interface.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /files", h.Upload)
	mux.HandleFunc("GET /files/{id}", h.Download)
	mux.HandleFunc("GET /files/{id}/info", h.GetInfo)
	mux.HandleFunc("DELETE /files/{id}", h.Delete)
	mux.HandleFunc("GET /files", h.List)
	mux.HandleFunc("GET /files/{id}/url", h.GetURL)
}

// Upload handles file upload via multipart form.
// POST /files
// Form fields:
//   - file: file data (required)
//   - generate_thumb: whether to generate thumbnail (optional, default false)
//   - thumb_width: thumbnail width (optional, default 200)
//   - thumb_height: thumbnail height (optional, default 200)
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract tenant ID from context
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		h.writeError(w, http.StatusBadRequest, "missing tenant_id in context")
		return
	}

	// Extract user ID from context
	userID, _ := ctx.Value("user_id").(string)

	// Parse multipart form
	if err := r.ParseMultipartForm(h.maxSize); err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("failed to parse form: %v", err))
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeError(w, http.StatusBadRequest, fmt.Sprintf("missing file: %v", err))
		return
	}
	defer file.Close()

	// Parse options
	generateThumb := r.FormValue("generate_thumb") == "true"
	thumbWidth, _ := strconv.Atoi(r.FormValue("thumb_width"))
	thumbHeight, _ := strconv.Atoi(r.FormValue("thumb_height"))

	// Build put options
	opts := PutOptions{
		TenantID:      tenantID,
		Reader:        file,
		FileName:      header.Filename,
		ContentType:   header.Header.Get("Content-Type"),
		GenerateThumb: generateThumb,
		ThumbWidth:    thumbWidth,
		ThumbHeight:   thumbHeight,
		UploadedBy:    userID,
		Metadata:      make(map[string]any),
	}

	// Upload file
	result, err := h.storage.Put(ctx, opts)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("upload failed: %v", err))
		return
	}

	// Return file metadata
	h.writeJSON(w, http.StatusOK, result)
}

// Download streams file content.
// GET /files/{id}
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, http.StatusBadRequest, "missing file id")
		return
	}

	// Get file metadata
	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		if err == ErrNotFound {
			h.writeError(w, http.StatusNotFound, "file not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get metadata: %v", err))
		}
		return
	}

	// Update access time (async, ignore errors)
	go h.metadata.UpdateAccessTime(context.Background(), fileID)

	// Get file content
	reader, err := h.storage.Get(ctx, fileMeta.Path)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to read file: %v", err))
		return
	}
	defer reader.Close()

	// Set response headers
	w.Header().Set("Content-Type", fileMeta.MimeType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileMeta.Size))
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileMeta.Name))
	w.Header().Set("Cache-Control", "public, max-age=31536000")

	// Stream file content
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

// GetInfo returns file metadata.
// GET /files/{id}/info
func (h *Handler) GetInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, http.StatusBadRequest, "missing file id")
		return
	}

	// Get file metadata
	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		if err == ErrNotFound {
			h.writeError(w, http.StatusNotFound, "file not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get metadata: %v", err))
		}
		return
	}

	h.writeJSON(w, http.StatusOK, fileMeta)
}

// Delete soft-deletes a file.
// DELETE /files/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, http.StatusBadRequest, "missing file id")
		return
	}

	// Soft delete in metadata
	if err := h.metadata.Delete(ctx, fileID); err != nil {
		if err == ErrNotFound {
			h.writeError(w, http.StatusNotFound, "file not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("delete failed: %v", err))
		}
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{"message": "file deleted"})
}

// List returns paginated list of files.
// GET /files?page=1&page_size=20&tenant_id=xxx&mime_type=image/jpeg&uploaded_by=user1
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	query := Query{
		Page:     1,
		PageSize: 20,
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			query.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}

	query.TenantID = r.URL.Query().Get("tenant_id")
	query.UploadedBy = r.URL.Query().Get("uploaded_by")
	query.MimeType = r.URL.Query().Get("mime_type")
	query.OrderBy = r.URL.Query().Get("order_by")

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			query.StartTime = t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			query.EndTime = t
		}
	}

	// Query files
	files, total, err := h.metadata.List(ctx, query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("list failed: %v", err))
		return
	}

	// Build response
	resp := map[string]any{
		"items":      files,
		"total":      total,
		"page":       query.Page,
		"page_size":  query.PageSize,
		"total_page": (total + int64(query.PageSize) - 1) / int64(query.PageSize),
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// GetURL returns a temporary access URL for the file.
// GET /files/{id}/url?expiry=3600
func (h *Handler) GetURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	fileID := r.PathValue("id")

	if fileID == "" {
		h.writeError(w, http.StatusBadRequest, "missing file id")
		return
	}

	// Parse expiry duration
	expiry := 15 * time.Minute // Default 15 minutes
	if expiryStr := r.URL.Query().Get("expiry"); expiryStr != "" {
		if seconds, err := strconv.Atoi(expiryStr); err == nil && seconds > 0 {
			expiry = time.Duration(seconds) * time.Second
		}
	}

	// Get file metadata
	fileMeta, err := h.metadata.Get(ctx, fileID)
	if err != nil {
		if err == ErrNotFound {
			h.writeError(w, http.StatusNotFound, "file not found")
		} else {
			h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to get metadata: %v", err))
		}
		return
	}

	// Get URL from storage
	url, err := h.storage.GetURL(ctx, fileMeta.Path, expiry)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to generate url: %v", err))
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"url":        url,
		"expires_in": fmt.Sprintf("%d", int(expiry.Seconds())),
	})
}

// writeJSON writes a JSON response.
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeError writes an error response.
func (h *Handler) writeError(w http.ResponseWriter, status int, message string) {
	h.writeJSON(w, status, map[string]any{
		"error":   http.StatusText(status),
		"message": message,
	})
}

// ServeHTTP implements http.Handler for middleware chaining.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Route based on method and path
	path := r.URL.Path
	method := r.Method

	switch {
	case method == http.MethodPost && path == "/files":
		h.Upload(w, r)
	case method == http.MethodGet && strings.HasPrefix(path, "/files/"):
		parts := strings.Split(strings.TrimPrefix(path, "/files/"), "/")
		if len(parts) == 1 {
			h.Download(w, r)
		} else if len(parts) == 2 && parts[1] == "info" {
			h.GetInfo(w, r)
		} else if len(parts) == 2 && parts[1] == "url" {
			h.GetURL(w, r)
		} else {
			h.writeError(w, http.StatusNotFound, "not found")
		}
	case method == http.MethodDelete && strings.HasPrefix(path, "/files/"):
		h.Delete(w, r)
	case method == http.MethodGet && path == "/files":
		h.List(w, r)
	default:
		h.writeError(w, http.StatusNotFound, "not found")
	}
}
