package handler

import (
	"net/http"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/datasource"
	"dbadmin/internal/domain/connection"
)

// ResourceHandler serves the unified resource-tree endpoint.
//
//	GET /api/connections/:id/resources?parentId=<path>
//
// parentId is the ResourceNode.Path of the parent node.
// Omit parentId to list top-level nodes (databases for SQL, db-indices for Redis).
type ResourceHandler struct {
	Connections  *connection.Store
	SQLAdapter   *datasource.SQLAdapter
	RedisAdapter *datasource.RedisAdapter
	Logger       plumelog.StructuredLogger
}

// ListResources returns child ResourceNodes for the given parent within a connection.
func (h ResourceHandler) ListResources(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
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

	drv := h.driverFor(conn)
	if drv == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("unsupported data source type: "+string(conn.Driver)).Build()))
		return
	}

	session, err := drv.Open(r.Context(), buildConnectionConfig(conn))
	if err != nil {
		h.Logger.Error("open session for resources", plumelog.Fields{
			"conn_id": connID,
			"error":   err.Error(),
		})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to connect to data source").Build()))
		return
	}

	parentPath := r.URL.Query().Get("parentId")
	var parentRef *datasource.ResourceRef
	if parentPath != "" {
		parentRef = &datasource.ResourceRef{
			ConnectionID: connID,
			Path:         parentPath,
		}
	}

	nodes, err := drv.ListResources(r.Context(), session, parentRef)
	if err != nil {
		h.Logger.Error("list resources", plumelog.Fields{
			"conn_id": connID,
			"parent":  parentPath,
			"error":   err.Error(),
		})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list resources").Build()))
		return
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, nodes,
		map[string]any{"count": len(nodes)}))
}

// driverFor returns the appropriate DataSourceDriver for the connection's driver type.
func (h ResourceHandler) driverFor(conn *connection.Connection) datasource.DataSourceDriver {
	switch conn.Driver {
	case connection.DriverMySQL, connection.DriverSQLite:
		return h.SQLAdapter
	case connection.DriverRedis:
		return h.RedisAdapter
	default:
		return nil
	}
}

// buildConnectionConfig constructs the driver-specific ConnectionConfig for a connection.
func buildConnectionConfig(conn *connection.Connection) datasource.ConnectionConfig {
	switch conn.Driver {
	case connection.DriverRedis:
		return datasource.RedisConfig{Conn: conn}
	default:
		return datasource.SQLConfig{Conn: conn}
	}
}
