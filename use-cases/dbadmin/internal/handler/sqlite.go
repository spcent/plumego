package handler

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
)

// sqliteMagic is the first 16 bytes of every valid SQLite3 database file.
const sqliteMagic = "SQLite format 3\x00"

// SQLiteHandler manages SQLite file upload and download.
type SQLiteHandler struct {
	Connections    *connection.Store
	Manager        *dbmanager.Manager
	UploadDir      string
	MaxUploadBytes int64
	Logger         plumelog.StructuredLogger
}

// Upload handles POST /api/sqlite/upload.
// Accepts multipart/form-data with field "file". Validates the SQLite magic
// header, writes the file to UploadDir under a cryptographically-random name,
// and returns the server-side path together with metadata.
func (h *SQLiteHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Reject oversized payloads before touching the body.
	if r.ContentLength > h.MaxUploadBytes {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(fmt.Sprintf("file too large: maximum is %d MiB", h.MaxUploadBytes>>20)).
			Detail("max_bytes", h.MaxUploadBytes).
			Build()))
		return
	}

	// ParseMultipartForm buffers up to 32 MiB in memory; the rest spills to disk.
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("failed to parse upload: "+err.Error()).
			Build()))
		return
	}
	defer r.MultipartForm.RemoveAll() //nolint:errcheck

	file, header, err := r.FormFile("file")
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("missing 'file' field in form").
			Build()))
		return
	}
	defer file.Close()

	// Secondary size guard using the declared part size.
	if header.Size > 0 && header.Size > h.MaxUploadBytes {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(fmt.Sprintf("file too large: maximum is %d MiB", h.MaxUploadBytes>>20)).
			Detail("max_bytes", h.MaxUploadBytes).
			Detail("file_size", header.Size).
			Build()))
		return
	}

	// Validate SQLite magic bytes.
	magic := make([]byte, 16)
	n, readErr := io.ReadFull(file, magic)
	if readErr != nil || n < 16 || string(magic) != sqliteMagic {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("invalid file: not a valid SQLite database (bad magic header)").
			Build()))
		return
	}
	// Seek back to beginning so we copy the full file.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to read uploaded file").
			Build()))
		return
	}

	// Generate a random filename — never derived from user input.
	randBytes := make([]byte, 16)
	if _, err := rand.Read(randBytes); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to generate a safe filename").
			Build()))
		return
	}
	destName := hex.EncodeToString(randBytes) + ".db"
	// filepath.Join is safe here: UploadDir is server-controlled, destName is hex.
	destPath := filepath.Join(h.UploadDir, destName)

	dst, err := os.Create(destPath)
	if err != nil {
		h.Logger.Error("create upload file", plumelog.Fields{"error": err.Error(), "path": destPath})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to save uploaded file").
			Build()))
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(destPath) //nolint:errcheck
		h.Logger.Error("write upload file", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Message("failed to write uploaded file").
			Build()))
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
		"file_path":     destPath,
		"size":          written,
		"original_name": header.Filename,
	}, nil))
}

// Download handles GET /api/conn/:id/sqlite/download.
// Streams the SQLite file for the given connection after re-verifying the magic
// header to prevent serving arbitrary server files even if file_path was
// manually configured.
func (h *SQLiteHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	conn, err := h.Connections.Get(id)
	if err != nil {
		if err == connection.ErrNotFound {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).Message("connection not found").Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to get connection").Build()))
		return
	}
	if conn.Driver != connection.DriverSQLite || conn.FilePath == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("connection is not a SQLite connection").Build()))
		return
	}

	f, err := os.Open(conn.FilePath)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("SQLite file not found on server").Build()))
		return
	}
	defer f.Close()

	// Re-verify magic before serving to guard against non-SQLite file_path values.
	magic := make([]byte, 16)
	if n, err := io.ReadFull(f, magic); err != nil || n < 16 || string(magic) != sqliteMagic {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("file is not a valid SQLite database").Build()))
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to read file").Build()))
		return
	}

	stat, _ := f.Stat()

	// Close the pool so any pending WAL is checkpointed before we stream.
	h.Manager.Close(id)

	downloadName := conn.OriginalFilename
	if downloadName == "" {
		downloadName = conn.Name + ".db"
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, downloadName))
	if stat != nil {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size()))
	}
	io.Copy(w, f) //nolint:errcheck
}
