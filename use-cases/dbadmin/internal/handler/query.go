package handler

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/domain/history"
)

// QueryHandler handles SQL console execution and query history.
type QueryHandler struct {
	Connections         *connection.Store
	Manager             *dbmanager.Manager
	History             *history.Store
	Logger              plumelog.StructuredLogger
	QueryTimeoutSeconds int // Maximum query execution time in seconds
}

type queryRequest struct {
	SQL              string `json:"sql"`
	Database         string `json:"database"`
	Readonly         bool   `json:"readonly"`
	ConfirmDangerous bool   `json:"confirmDangerous"`
}

type selectResult struct {
	Type            string           `json:"type"` // "result_set"
	Columns         []string         `json:"columns"`
	Rows            []map[string]any `json:"rows"`
	ExecutionTimeMs int64            `json:"executionTimeMs"`
	Truncated       bool             `json:"truncated"`
}

type execResult struct {
	Type            string `json:"type"` // "exec_result"
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
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

	// Apply timeout if configured
	ctx := r.Context()
	var cancel context.CancelFunc
	if h.QueryTimeoutSeconds > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Duration(h.QueryTimeoutSeconds)*time.Second)
		defer cancel()
	}

	// For MySQL, USE the selected database first.
	if conn.Driver == connection.DriverMySQL && req.Database != "" {
		if _, err := db.ExecContext(ctx, "USE "+quoteIdent(req.Database, connection.DriverMySQL)); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to select database").Build()))
			return
		}
	}

	start := time.Now()
	originalSQL := req.SQL // preserve original for history

	if cls.IsSelect {
		// Append LIMIT 1000 if no explicit LIMIT to prevent unbounded result sets.
		truncated := false
		if !strings.Contains(strings.ToUpper(sqlStr), "LIMIT") {
			sqlStr += " LIMIT 1000"
			truncated = true
		}

		rows, err := db.QueryContext(ctx, sqlStr)
		if err != nil {
			durationMs := time.Since(start).Milliseconds()
			_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, err.Error())

			// Check if timeout occurred
			if ctx.Err() == context.DeadlineExceeded {
				h.Logger.Error("query timeout", plumelog.Fields{"timeout": h.QueryTimeoutSeconds, "sql": sqlStr})
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeTimeout).
					Message("query execution timeout").
					Detail("timeout_seconds", h.QueryTimeoutSeconds).
					Build()))
			} else {
				h.Logger.Error("query failed", plumelog.Fields{"error": err.Error()})
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeInternal).Message(err.Error()).Build()))
			}
			return
		}
		defer rows.Close()

		cols, _ := rows.Columns()
		resultRows, err := scanRows(rows, cols)
		if err != nil {
			h.Logger.Error("scan rows failed", plumelog.Fields{"error": err.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to scan results").Build()))
			return
		}

		durationMs := time.Since(start).Milliseconds()
		_ = h.recordHistory(connID, req.Database, originalSQL, durationMs, "")
		logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, selectResult{
			Type:            "result_set",
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

		// Check if timeout occurred
		if ctx.Err() == context.DeadlineExceeded {
			h.Logger.Error("exec timeout", plumelog.Fields{"timeout": h.QueryTimeoutSeconds, "sql": sqlStr})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeTimeout).
				Message("query execution timeout").
				Detail("timeout_seconds", h.QueryTimeoutSeconds).
				Build()))
		} else {
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
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, execResult{
		Type:            "exec_result",
		RowsAffected:    rowsAffected,
		LastInsertId:    lastInsertId,
		ExecutionTimeMs: durationMs,
	}, nil))
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

func generateHistoryID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
