package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/esmanager"
	"dbadmin/internal/mongomanager"
	"dbadmin/internal/redismanager"
)

// ConnectionHandler handles CRUD operations for saved DB connections.
type ConnectionHandler struct {
	Connections  *connection.Store
	Manager      *dbmanager.Manager
	RedisManager *redismanager.Manager
	MongoManager *mongomanager.Manager
	ESManager    *esmanager.Manager
	Logger       plumelog.StructuredLogger
}

// List returns all saved connections (passwords redacted).
func (h ConnectionHandler) List(w http.ResponseWriter, r *http.Request) {
	conns, err := h.Connections.List()
	if err != nil {
		h.Logger.Error("list connections", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list connections").Build()))
		return
	}
	// Redact all sensitive credentials before returning.
	for _, c := range conns {
		c.Redact()
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, conns, map[string]any{"count": len(conns)}))
}

// Get returns a single connection by ID (password redacted).
func (h ConnectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	c, err := h.Connections.Get(id)
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
	c.Redact()
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, c, nil))
}

// Create saves a new connection.
func (h ConnectionHandler) Create(w http.ResponseWriter, r *http.Request) {
	var c connection.Connection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if err := validateConnection(&c); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	if err := h.Connections.Create(&c); err != nil {
		h.Logger.Error("create connection", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to save connection").Build()))
		return
	}
	c.Redact()
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, c, nil))
}

// Update replaces a connection.
func (h ConnectionHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	var c connection.Connection
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	c.ID = id
	if err := validateConnection(&c); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	if err := h.Connections.Update(&c); err != nil {
		if err == connection.ErrNotFound {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).Message("connection not found").Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to update connection").Build()))
		return
	}
	// Invalidate cached pool on update.
	h.Manager.Close(id)
	c.Redact()
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, c, nil))
}

// Delete removes a connection.
// Query param ?deleteFile=true deletes the associated uploaded SQLite temp file.
func (h ConnectionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	// Fetch before delete so we can optionally clean up the temp file.
	conn, fetchErr := h.Connections.Get(id)

	if err := h.Connections.Delete(id); err != nil {
		if err == connection.ErrNotFound {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).Message("connection not found").Build()))
			return
		}
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to delete connection").Build()))
		return
	}
	h.Manager.Close(id)

	// Optionally remove the server-managed temp file.
	if fetchErr == nil && conn.UploadedFile && conn.FilePath != "" &&
		r.URL.Query().Get("deleteFile") == "true" {
		os.Remove(conn.FilePath) //nolint:errcheck
	}

	w.WriteHeader(http.StatusNoContent)
}

// Test opens a temporary connection and pings it.
func (h ConnectionHandler) Test(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	c, err := h.Connections.Get(id)
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
	if c.Driver == connection.DriverRedis {
		if err := h.RedisManager.Test(r.Context(), c); err != nil {
			logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
				map[string]any{"ok": false, "error": err.Error()}, nil))
			return
		}
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true}, nil))
		return
	}
	if c.Driver == connection.DriverMongoDB {
		if err := h.MongoManager.Test(r.Context(), c); err != nil {
			logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
				map[string]any{"ok": false, "error": err.Error()}, nil))
			return
		}
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true}, nil))
		return
	}
	if c.Driver == connection.DriverElasticsearch {
		if err := h.ESManager.Test(r.Context(), c); err != nil {
			logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
				map[string]any{"ok": false, "error": err.Error()}, nil))
			return
		}
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{"ok": true}, nil))
		return
	}
	if err := h.Manager.Test(c); err != nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
			map[string]any{"ok": false, "error": err.Error()}, nil))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
		map[string]any{"ok": true}, nil))
}

func validateConnection(c *connection.Connection) error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}
	switch c.Driver {
	case connection.DriverMySQL:
		if c.Host == "" {
			return fmt.Errorf("host is required for mysql")
		}
		if c.Port == 0 {
			c.Port = 3306
		}
	case connection.DriverSQLite:
		if c.FilePath == "" {
			return fmt.Errorf("file_path is required for sqlite")
		}
	case connection.DriverRedis:
		if c.Host == "" {
			return fmt.Errorf("host is required for redis")
		}
		if c.Port == 0 {
			c.Port = 6379
		}
		if c.RedisDBIndex < 0 || c.RedisDBIndex > 15 {
			return fmt.Errorf("redis_db_index must be 0-15")
		}
	case connection.DriverMongoDB:
		if c.MongoURI == "" && c.Host == "" {
			return fmt.Errorf("mongo_uri or host is required for mongodb")
		}
		if c.MongoURI == "" && c.Port == 0 {
			c.Port = 27017
		}
	case connection.DriverElasticsearch:
		if len(c.ESNodes) == 0 && c.Host == "" {
			return fmt.Errorf("es_nodes or host is required for elasticsearch")
		}
		if c.Port == 0 && len(c.ESNodes) == 0 {
			c.Port = 9200
		}
		if len(c.ESNodes) == 0 {
			c.ESNodes = []string{fmt.Sprintf("http://%s:%d", c.Host, c.Port)}
		}
	default:
		return fmt.Errorf("driver must be 'mysql', 'sqlite', 'redis', 'mongodb', or 'elasticsearch'")
	}
	return nil
}
