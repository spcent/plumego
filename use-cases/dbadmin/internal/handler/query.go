package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/domain/history"
)

// SlowQueryThresholdMs defines the threshold for slow query logging.
// Queries taking longer than this will be logged at WARN level.
const SlowQueryThresholdMs = 1000

// QueryHandler handles SQL console execution and query history.
type QueryHandler struct {
	Connections         *connection.Store
	Manager             *dbmanager.Manager
	History             *history.Store
	Logger              plumelog.StructuredLogger
	QueryTimeoutSeconds int // Maximum query execution time in seconds
	Registry            *QueryRegistry
}

type queryRequest struct {
	SQL              string `json:"sql"`
	Database         string `json:"database"`
	Readonly         bool   `json:"readonly"`
	ConfirmDangerous bool   `json:"confirmDangerous"`
	QueryID          string `json:"queryId"`
}

type selectResult struct {
	Type            string           `json:"type"` // "result_set"
	QueryID         string           `json:"queryId,omitempty"`
	Columns         []string         `json:"columns"`
	Rows            []map[string]any `json:"rows"`
	ExecutionTimeMs int64            `json:"executionTimeMs"`
	Truncated       bool             `json:"truncated"`
}

type execResult struct {
	Type            string `json:"type"` // "exec_result"
	QueryID         string `json:"queryId,omitempty"`
	RowsAffected    int64  `json:"rowsAffected"`
	LastInsertId    int64  `json:"lastInsertId"`
	ExecutionTimeMs int64  `json:"executionTimeMs"`
}

type sqlClass struct {
	IsSelect    bool
	IsDangerous bool
	Reason      string
}

// classifySQL inspects the first keyword of a SQL statement.
// Returns whether it is a SELECT-type query and whether it is a dangerous operation.
func classifySQL(sql string) sqlClass {
	upper := strings.ToUpper(strings.TrimSpace(sql))

	if strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "EXPLAIN") ||
		strings.HasPrefix(upper, "DESCRIBE") ||
		strings.HasPrefix(upper, "DESC") ||
		strings.HasPrefix(upper, "PRAGMA") ||
		strings.HasPrefix(upper, "WITH") {
		return sqlClass{IsSelect: true}
	}

	if strings.HasPrefix(upper, "DROP") {
		return sqlClass{IsDangerous: true, Reason: "DROP destroys data irreversibly"}
	}
	if strings.HasPrefix(upper, "TRUNCATE") {
		return sqlClass{IsDangerous: true, Reason: "TRUNCATE removes all rows"}
	}
	if strings.HasPrefix(upper, "ALTER") {
		return sqlClass{IsDangerous: true, Reason: "ALTER TABLE modifies table structure"}
	}
	if strings.HasPrefix(upper, "DELETE") && !hasWhereClause(upper) {
		return sqlClass{IsDangerous: true, Reason: "DELETE without WHERE affects all rows"}
	}
	if strings.HasPrefix(upper, "UPDATE") && !hasWhereClause(upper) {
		return sqlClass{IsDangerous: true, Reason: "UPDATE without WHERE affects all rows"}
	}
	return sqlClass{}
}

// stripBlockComments removes /* ... */ block comments from SQL, replacing each
// comment with a single space so adjacent tokens remain separated.
func stripBlockComments(sql string) string {
	var out strings.Builder
	i := 0
	for i < len(sql) {
		if i+1 < len(sql) && sql[i] == '/' && sql[i+1] == '*' {
			out.WriteByte(' ')
			i += 2
			for i+1 < len(sql) && !(sql[i] == '*' && sql[i+1] == '/') {
				i++
			}
			i += 2 // skip closing */
			continue
		}
		out.WriteByte(sql[i])
		i++
	}
	return out.String()
}

func hasWhereClause(upperSQL string) bool {
	// Strip block comments first so /* WHERE */ inside a comment is ignored,
	// and DELETE /*comment*/ FROM t is not incorrectly classified as safe.
	clean := stripBlockComments(upperSQL)
	return strings.Contains(clean, " WHERE ") ||
		strings.Contains(clean, "\tWHERE ") ||
		strings.Contains(clean, "\nWHERE ") ||
		strings.Contains(clean, "\rWHERE ") ||
		strings.HasSuffix(strings.TrimRight(clean, " \t\n\r"), "WHERE")
}

// hasMultipleStatements returns true if sql contains more than one statement
// (i.e., a non-whitespace token follows a semicolon outside of string literals).
func hasMultipleStatements(sql string) bool {
	inSingle, inDouble := false, false
	for i := 0; i < len(sql); i++ {
		ch := sql[i]
		switch ch {
		case '\\':
			if inSingle || inDouble {
				i++ // skip escaped character
			}
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case ';':
			if !inSingle && !inDouble {
				if strings.TrimSpace(sql[i+1:]) != "" {
					return true
				}
			}
		}
	}
	return false
}

// Execute runs a SQL statement and returns the results.
func (h QueryHandler) Execute(w http.ResponseWriter, r *http.Request) {
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

	var req queryRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	sqlStr := strings.TrimSpace(req.SQL)
	if sqlStr == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("sql is required").Build()))
		return
	}

	if hasMultipleStatements(sqlStr) {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("multi-statement SQL is not supported").Build()))
		return
	}

	cls := classifySQL(sqlStr)
	if cls.IsDangerous && !req.ConfirmDangerous {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("dangerous SQL requires confirmation").
			Detail("confirm_required", true).
			Detail("reason", cls.Reason).
			Build()))
		return
	}

	// Block write statements on readonly connections.
	if conn.Readonly && !cls.IsSelect {
		if guardReadonly(conn, w, r, h.Logger) {
			return
		}
	}

	db, err := h.Manager.Open(conn)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to connect to database").Build()))
		return
	}

	// Apply timeout if configured, wrapped in a cancellable context so the query
	// can also be cancelled manually via the Cancel endpoint.
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	if h.QueryTimeoutSeconds > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, time.Duration(h.QueryTimeoutSeconds)*time.Second)
		defer timeoutCancel()
	}

	start := time.Now()
	originalSQL := req.SQL // preserve original for history

	// Register query for cancellation if registry is available.
	queryID := strings.TrimSpace(req.QueryID)
	if queryID == "" {
		queryID, _ = generateHistoryID()
	}
	if h.Registry != nil {
		h.Registry.Register(queryID, connID, req.Database, originalSQL, cancel)
		defer h.Registry.Unregister(queryID)
	}

	// For MySQL, USE the selected database first.
	if conn.Driver == connection.DriverMySQL && req.Database != "" {
		if err := validateName(req.Database); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).Message("invalid database name: "+err.Error()).Build()))
			return
		}
		if _, err := db.ExecContext(ctx, "USE "+quoteIdent(req.Database, connection.DriverMySQL)); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to select database").Build()))
			return
		}
	}

	if cls.IsSelect {
		// Append LIMIT 1000 if no explicit LIMIT to prevent unbounded result sets.
		truncated := false
		if !strings.Contains(strings.ToUpper(sqlStr), "LIMIT") {
			sqlStr += " LIMIT " + strconv.Itoa(DefaultQueryRows)
			truncated = true
		}

		rows, err := db.QueryContext(ctx, sqlStr)
		if err != nil {
			durationMs := time.Since(start).Milliseconds()
			_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, err.Error())
			h.logSlowQuery(connID, req.Database, originalSQL, durationMs)

			// Check cancellation or timeout
			switch ctx.Err() {
			case context.Canceled:
				h.Logger.Info("query cancelled", plumelog.Fields{"queryId": queryID})
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeTimeout).
					Message("query was cancelled").
					Detail("query_id", queryID).
					Build()))
			case context.DeadlineExceeded:
				h.Logger.Error("query timeout", plumelog.Fields{"timeout": h.QueryTimeoutSeconds, "sql": sqlStr})
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeTimeout).
					Message("query execution timeout").
					Detail("timeout_seconds", h.QueryTimeoutSeconds).
					Detail("query_id", queryID).
					Build()))
			default:
				h.Logger.Error("query failed", plumelog.Fields{"error": err.Error()})
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeInternal).Message(err.Error()).Build()))
			}
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		resultRows, contentTruncated, err := scanRows(rows, cols)
		if contentTruncated {
			truncated = true
		}
		if err != nil {
			h.Logger.Error("scan rows failed", plumelog.Fields{"error": err.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to scan results").Build()))
			return
		}

		durationMs := time.Since(start).Milliseconds()
		_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, "")
		h.logSlowQuery(connID, req.Database, originalSQL, durationMs)
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, selectResult{
			Type:            "result_set",
			QueryID:         queryID,
			Columns:         cols,
			Rows:            resultRows,
			ExecutionTimeMs: durationMs,
			Truncated:       truncated,
		}, nil))
		return
	}

	// Non-SELECT: use Exec.
	res, err := db.ExecContext(ctx, sqlStr)
	if err != nil {
		durationMs := time.Since(start).Milliseconds()
		_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, err.Error())
		h.logSlowQuery(connID, req.Database, originalSQL, durationMs)

		// Check cancellation or timeout
		switch ctx.Err() {
		case context.Canceled:
			h.Logger.Info("exec cancelled", plumelog.Fields{"queryId": queryID})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeTimeout).
				Message("query was cancelled").
				Detail("query_id", queryID).
				Build()))
		case context.DeadlineExceeded:
			h.Logger.Error("exec timeout", plumelog.Fields{"timeout": h.QueryTimeoutSeconds, "sql": sqlStr})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeTimeout).
				Message("query execution timeout").
				Detail("timeout_seconds", h.QueryTimeoutSeconds).
				Detail("query_id", queryID).
				Build()))
		default:
			h.Logger.Error("exec failed", plumelog.Fields{"error": err.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message(err.Error()).Build()))
		}
		return
	}

	rowsAffected, _ := res.RowsAffected()
	lastInsertId, _ := res.LastInsertId()
	durationMs := time.Since(start).Milliseconds()
	_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, "")
	h.logSlowQuery(connID, req.Database, originalSQL, durationMs)
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, execResult{
		Type:            "exec_result",
		QueryID:         queryID,
		RowsAffected:    rowsAffected,
		LastInsertId:    lastInsertId,
		ExecutionTimeMs: durationMs,
	}, nil))
}

// Cancel cancels an active query by its ID.
// POST /api/queries/cancel - Body: {"queryId": "abc123"}
func (h QueryHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	if h.Registry == nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("query registry not available").Build()))
		return
	}

	var req struct {
		QueryID string `json:"queryId"`
	}
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}

	if req.QueryID == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("queryId is required").Build()))
		return
	}

	if h.Registry.Cancel(req.QueryID) {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, map[string]any{
			"status":  "cancelled",
			"queryId": req.QueryID,
		}, nil))
	} else {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("query not found or already completed").Build()))
	}
}

// ListActive returns all currently running queries.
// GET /api/queries/active
func (h QueryHandler) ListActive(w http.ResponseWriter, r *http.Request) {
	if h.Registry == nil {
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, []any{}, nil))
		return
	}

	queries := h.Registry.ListActive()
	result := make([]map[string]any, 0, len(queries))
	for _, q := range queries {
		result = append(result, map[string]any{
			"queryId":   q.QueryID,
			"connId":    q.ConnID,
			"database":  q.Database,
			"sql":       q.SQL,
			"startTime": q.StartTime,
			"duration":  time.Since(q.StartTime).String(),
		})
	}

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, result, map[string]any{"count": len(result)}))
}

// ListHistory returns query history for a connection.
func (h QueryHandler) ListHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	entries, err := h.History.List(connID)
	if err != nil {
		h.Logger.Error("list history", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list history").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, entries, map[string]any{"count": len(entries)}))
}

// DeleteHistory removes a single history entry.
func (h QueryHandler) DeleteHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	entryID := router.Param(r, "entryId")
	if err := h.History.Delete(connID, entryID); err != nil {
		h.Logger.Error("delete history", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to delete history entry").Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ClearHistory removes all history entries for a connection.
func (h QueryHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	if err := h.History.Clear(connID); err != nil {
		h.Logger.Error("clear history", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to clear history").Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h QueryHandler) recordHistory(connID, dbName, sql string, durationMS int64, errStr string) error {
	id, _ := generateHistoryID()
	return h.History.Add(&history.Entry{
		ID:        id,
		ConnID:    connID,
		Database:  dbName,
		SQL:       sql,
		Duration:  durationMS,
		Error:     errStr,
		CreatedAt: time.Now().UTC(),
	})
}

// logSlowQuery logs queries that exceed the slow query threshold.
func (h QueryHandler) logSlowQuery(connID, dbName, sql string, durationMs int64) {
	if durationMs > SlowQueryThresholdMs {
		h.Logger.Warn("slow query detected", plumelog.Fields{
			"connection_id": connID,
			"database":      dbName,
			"duration_ms":   durationMs,
			"threshold_ms":  SlowQueryThresholdMs,
			"sql":           truncateSQLForLog(sql, 200),
		})
	}
}

// truncateSQLForLog truncates SQL for logging to prevent excessively long log entries.
func truncateSQLForLog(sql string, maxLen int) string {
	if len(sql) <= maxLen {
		return sql
	}
	return sql[:maxLen] + "... (truncated)"
}

func generateHistoryID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
