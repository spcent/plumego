package handler

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	mysqlinspect "dbadmin/internal/dbmanager/mysql"
	sqliteinspect "dbadmin/internal/dbmanager/sqlite"
	"dbadmin/internal/domain/connection"
)

// ExportHandler handles data export (CSV and SQL dump).
type ExportHandler struct {
	Connections *connection.Store
	Manager     *dbmanager.Manager
	Logger      plumelog.StructuredLogger
}

// Export streams a table as CSV or SQL INSERT statements.
// Query params: format=csv|sql, includeSchema=true|false, includeData=true|false
func (h ExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	connID := router.Param(r, "id")
	dbName := router.Param(r, "db")
	table := router.Param(r, "table")

	q := r.URL.Query()
	format := q.Get("format")
	if format == "" {
		format = "csv"
	}
	includeSchema := q.Get("includeSchema") != "false"
	includeData := q.Get("includeData") != "false"
	ts := time.Now().Format("20060102_150405")

	limit := parseExportLimit(q.Get("limit"))

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
	db, err := h.Manager.Open(r.Context(), conn)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to connect").
			Detail("error", err.Error()).Build()))
		return
	}

	var tableFQN string
	switch conn.Driver {
	case connection.DriverMySQL:
		tableFQN = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(table, conn.Driver))
	default:
		tableFQN = quoteIdent(table, conn.Driver)
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT %d", tableFQN, limit+1)
	rows, err := db.QueryContext(r.Context(), query)
	if err != nil {
		h.Logger.Error("export query", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("export query failed").Build()))
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to get columns").Build()))
		return
	}

	w.Header().Set("X-Export-Row-Limit", strconv.Itoa(limit))

	switch format {
	case "sql":
		h.exportSQL(w, r, rows, cols, conn, dbName, table, ts, includeSchema, includeData, limit)
	default:
		h.exportCSV(w, r, rows, cols, table, ts, includeData, limit)
	}
}

func (h ExportHandler) exportCSV(w http.ResponseWriter, r *http.Request, rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}, cols []string, table, ts string, includeData bool, limit int) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.csv"`, table, ts))
	cw := csv.NewWriter(w)
	_ = cw.Write(cols)

	if !includeData {
		cw.Flush()
		return
	}

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	n := 0
	for rows.Next() {
		if n >= limit {
			_ = cw.Write([]string{fmt.Sprintf("[export truncated at %d rows — refine the query or increase ?limit= up to %d]", limit, MaxExportRows)})
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			h.Logger.Warn("csv scan", plumelog.Fields{"error": err.Error()})
			continue
		}
		record := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case nil:
				record[i] = ""
			case []byte:
				if utf8.Valid(t) {
					record[i] = string(t)
				} else {
					record[i] = fmt.Sprintf("<BLOB %d bytes>", len(t))
				}
			default:
				record[i] = fmt.Sprintf("%v", t)
			}
		}
		_ = cw.Write(record)
		n++
	}
	cw.Flush()
}

func (h ExportHandler) exportSQL(w http.ResponseWriter, r *http.Request, rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}, cols []string, conn *connection.Connection, dbName, table, ts string, includeSchema, includeData bool, limit int) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s_%s.sql"`, table, ts))

	quotedCols := quoteIdents(cols, conn.Driver)
	var tableFQN string
	switch conn.Driver {
	case connection.DriverMySQL:
		tableFQN = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(table, conn.Driver))
	default:
		tableFQN = quoteIdent(table, conn.Driver)
	}

	var ins dbmanager.Inspector
	if conn.Driver == connection.DriverMySQL {
		db, _ := h.Manager.Open(r.Context(), conn)
		ins = mysqlinspect.New(db)
	} else {
		db, _ := h.Manager.Open(r.Context(), conn)
		ins = sqliteinspect.New(db)
	}

	if includeSchema {
		if ddl, err := ins.CreateTableDDL(r.Context(), dbName, table); err == nil {
			fmt.Fprintf(w, "-- Table: %s\n%s;\n\n", table, ddl)
		}
	}

	if !includeData {
		return
	}

	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}

	n := 0
	for rows.Next() {
		if n >= limit {
			fmt.Fprintf(w, "-- Export truncated at %d rows. Refine the query or increase ?limit= up to %d.\n", limit, MaxExportRows)
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		valStrs := make([]string, len(cols))
		for i, v := range vals {
			switch t := v.(type) {
			case nil:
				valStrs[i] = "NULL"
			case []byte:
				valStrs[i] = sqlQuoteString(string(t), conn.Driver)
			case string:
				valStrs[i] = sqlQuoteString(t, conn.Driver)
			case int64:
				valStrs[i] = fmt.Sprintf("%d", t)
			case float64:
				valStrs[i] = fmt.Sprintf("%g", t)
			default:
				valStrs[i] = sqlQuoteString(fmt.Sprintf("%v", t), conn.Driver)
			}
		}
		fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n",
			tableFQN,
			strings.Join(quotedCols, ", "),
			strings.Join(valStrs, ", "))
		n++
	}
}

// sqlQuoteString wraps s in single quotes, escaping embedded single quotes.
// For MySQL (default settings), backslashes are also escaped to prevent
// misinterpretation of sequences like \n, \t, \0 in exported INSERT statements.
func sqlQuoteString(s string, driver connection.DriverType) string {
	if driver == connection.DriverMySQL {
		s = strings.ReplaceAll(s, `\`, `\\`)
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
