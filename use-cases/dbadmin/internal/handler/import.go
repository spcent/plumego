package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
)

// ImportHandler handles SQL file imports.
type ImportHandler struct {
	Connections *connection.Store
	Manager     *dbmanager.Manager
	Logger      plumelog.StructuredLogger
}

type importRequest struct {
	SQL              string `json:"sql"`
	ConfirmDangerous bool   `json:"confirmDangerous"`
}

type importErrorDetail struct {
	Index   int    `json:"index"`
	Snippet string `json:"snippet"`
	Error   string `json:"error"`
}

type importResult struct {
	StatementsExecuted int                 `json:"statements_executed"`
	Errors             int                 `json:"errors"`
	ErrorsDetail       []importErrorDetail `json:"errors_detail"`
}

type dangerousStatement struct {
	Index   int    `json:"index"`
	Snippet string `json:"snippet"`
	Reason  string `json:"reason"`
}

// Import reads a JSON body {sql, confirmDangerous} and executes each statement.
func (h ImportHandler) Import(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := router.Param(r, "db")

	var req importRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if strings.TrimSpace(req.SQL) == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("sql is required").Build()))
		return
	}

	stmts := splitSQL(req.SQL)
	if len(stmts) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no SQL statements found in input").Build()))
		return
	}

	// Scan for dangerous statements before executing.
	var dangerous []dangerousStatement
	for i, stmt := range stmts {
		cls := classifySQL(stmt)
		if cls.IsDangerous {
			dangerous = append(dangerous, dangerousStatement{
				Index:   i + 1,
				Snippet: truncate(stmt, 100),
				Reason:  cls.Reason,
			})
		}
	}
	if len(dangerous) > 0 && !req.ConfirmDangerous {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("dangerous SQL requires confirmation").
			Detail("confirm_required", true).
			Detail("dangerous_statements", dangerous).
			Build()))
		return
	}

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
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	db, err := h.Manager.Open(conn)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to connect").
			Detail("error", err.Error()).Build()))
		return
	}

	// For MySQL, set the target database.
	if conn.Driver == connection.DriverMySQL && dbName != "" {
		if _, err := db.ExecContext(r.Context(), "USE "+quoteIdent(dbName, connection.DriverMySQL)); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to select database").Build()))
			return
		}
	}

	result := importResult{ErrorsDetail: []importErrorDetail{}}
	for i, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if _, err := db.ExecContext(r.Context(), stmt); err != nil {
			result.Errors++
			result.ErrorsDetail = append(result.ErrorsDetail, importErrorDetail{
				Index:   i + 1,
				Snippet: truncate(stmt, 100),
				Error:   err.Error(),
			})
			h.Logger.Warn("import statement failed", plumelog.Fields{
				"index": i + 1,
				"error": err.Error(),
				"stmt":  truncate(stmt, 100),
			})
			continue
		}
		result.StatementsExecuted++
	}

	status := http.StatusOK
	if result.Errors > 0 && result.StatementsExecuted == 0 {
		status = http.StatusUnprocessableEntity
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, status, result, nil))
}

// splitSQL splits a SQL script into individual statements using a character-level
// parser that correctly handles:
//   - Block comments /* ... */ (may span multiple lines or contain semicolons)
//   - Line comments -- ... and # ...
//   - Single-quoted and double-quoted string literals
func splitSQL(script string) []string {
	var stmts []string
	var current strings.Builder
	inSingle, inDouble, inBlock := false, false, false
	i := 0
	for i < len(script) {
		ch := script[i]
		switch {
		case !inSingle && !inDouble && !inBlock && i+1 < len(script) && ch == '/' && script[i+1] == '*':
			// Block comment start — skip entire comment, leave a space boundary.
			inBlock = true
			i += 2
		case inBlock && i+1 < len(script) && ch == '*' && script[i+1] == '/':
			inBlock = false
			i += 2
			current.WriteByte(' ')
		case inBlock:
			i++
		case !inSingle && !inDouble && i+1 < len(script) && ch == '-' && script[i+1] == '-':
			// Line comment: skip to end of line.
			for i < len(script) && script[i] != '\n' {
				i++
			}
		case !inSingle && !inDouble && ch == '#':
			// MySQL-style line comment: skip to end of line.
			for i < len(script) && script[i] != '\n' {
				i++
			}
		case ch == '\'' && !inDouble:
			inSingle = !inSingle
			current.WriteByte(ch)
			i++
		case ch == '"' && !inSingle:
			inDouble = !inDouble
			current.WriteByte(ch)
			i++
		case ch == ';' && !inSingle && !inDouble:
			if stmt := strings.TrimSpace(current.String()); stmt != "" {
				stmts = append(stmts, stmt)
			}
			current.Reset()
			i++
		default:
			current.WriteByte(ch)
			i++
		}
	}
	if stmt := strings.TrimSpace(current.String()); stmt != "" {
		stmts = append(stmts, stmt)
	}
	return stmts
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
