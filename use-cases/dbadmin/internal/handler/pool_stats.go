package handler

import (
	"net/http"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/esmanager"
	"dbadmin/internal/mongomanager"
	"dbadmin/internal/redismanager"
)

// PoolStatsHandler exposes connection pool metrics for monitoring.
type PoolStatsHandler struct {
	DBManager    *dbmanager.Manager
	RedisManager *redismanager.Manager
	MongoManager *mongomanager.Manager
	ESManager    *esmanager.Manager
	Logger       plumelog.StructuredLogger
}

// SQLPoolStats represents connection pool statistics for SQL databases.
type SQLPoolStats struct {
	ConnectionID string `json:"connection_id"`
	Driver       string `json:"driver"`
	MaxOpen      int    `json:"max_open"`
	Open         int    `json:"open"`
	InUse        int    `json:"in_use"`
	Idle         int    `json:"idle"`
	WaitCount    int64  `json:"wait_count"`
	WaitDuration string `json:"wait_duration"`
	MaxIdleClosed     int64 `json:"max_idle_closed"`
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"`
}

// GetSQLPoolStats returns pool statistics for a specific SQL connection.
func (h PoolStatsHandler) GetSQLPoolStats(w http.ResponseWriter, r *http.Request) {
	connID := r.URL.Query().Get("connection_id")
	if connID == "" {
		h.Logger.Warn("pool stats request missing connection_id")
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("connection_id query parameter is required").
			Build()))
		return
	}

	stats, ok := h.DBManager.Stats(connID)
	if !ok {
		h.Logger.Warn("pool stats request for unknown connection", plumelog.Fields{"connection_id": connID})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Message("connection not found or not yet opened").
			Build()))
		return
	}

	response := SQLPoolStats{
		ConnectionID:      connID,
		Driver:            "sql",
		MaxOpen:           stats.MaxOpenConnections,
		Open:              stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration.String(),
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, response, nil))
}

// AllPoolStats returns statistics for all active connections.
type AllPoolStats struct {
	SQLConnections      []SQLPoolStats          `json:"sql_connections,omitempty"`
	RedisConnections    int                     `json:"redis_connections"`
	MongoDBConnections  int                     `json:"mongodb_connections"`
	ESConnections       int                     `json:"es_connections"`
	Timestamp           string                  `json:"timestamp"`
}

// GetAllStats returns pool statistics for all connection types.
func (h PoolStatsHandler) GetAllStats(w http.ResponseWriter, r *http.Request) {
	stats := AllPoolStats{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// Get SQL connection stats
	sqlStats := h.DBManager.AllStats()
	stats.SQLConnections = make([]SQLPoolStats, 0, len(sqlStats))
	for connID, dbStats := range sqlStats {
		stats.SQLConnections = append(stats.SQLConnections, SQLPoolStats{
			ConnectionID:      connID,
			Driver:            "sql",
			MaxOpen:           dbStats.MaxOpenConnections,
			Open:              dbStats.OpenConnections,
			InUse:             dbStats.InUse,
			Idle:              dbStats.Idle,
			WaitCount:         dbStats.WaitCount,
			WaitDuration:      dbStats.WaitDuration.String(),
			MaxIdleClosed:     dbStats.MaxIdleClosed,
			MaxLifetimeClosed: dbStats.MaxLifetimeClosed,
		})
	}

	// Count active connections for other types
	stats.RedisConnections = h.RedisManager.Count()
	stats.MongoDBConnections = h.MongoManager.Count()
	stats.ESConnections = h.ESManager.Count()

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, stats, nil))
}
