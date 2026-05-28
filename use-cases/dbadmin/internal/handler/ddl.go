package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
)

// DDLHandler handles schema (DDL) operations: create/alter/drop table.
type DDLHandler struct {
	Connections *connection.Store
	Manager     *dbmanager.Manager
	Logger      plumelog.StructuredLogger
}

// ColumnDef describes a column in a CREATE TABLE request.
type ColumnDef struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default,omitempty"`
	Comment  string `json:"comment,omitempty"`
}

type createTableRequest struct {
	Name    string      `json:"name"`
	Columns []ColumnDef `json:"columns"`
	Engine  string      `json:"engine,omitempty"` // MySQL only
	Charset string      `json:"charset,omitempty"`
}

type alterTableRequest struct {
	AddColumns  []ColumnDef `json:"add_columns,omitempty"`
	DropColumns []string    `json:"drop_columns,omitempty"`
	RenameTable string      `json:"rename_table,omitempty"`
}

// CreateTable executes CREATE TABLE.
func (h DDLHandler) CreateTable(w http.ResponseWriter, r *http.Request) {
	db, conn, dbName, err := h.openDB(r)
	if err != nil {
		h.writeConnErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) { return }
	var req createTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	if req.Name == "" || len(req.Columns) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("name and columns are required").Build()))
		return
	}
	for _, col := range req.Columns {
		if col.Default != "" {
			if err := validateDDLLiteral(col.Default); err != nil {
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeValidation).
					Message("invalid default value for column "+col.Name+": "+err.Error()).Build()))
				return
			}
		}
	}
	if err := validateEngine(req.Engine); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	ddl := buildCreateTable(dbName, req, conn.Driver)
	if _, err := db.ExecContext(r.Context(), ddl); err != nil {
		h.Logger.Error("create table", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to create table").
			Detail("error", err.Error()).Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated,
		map[string]string{"table": req.Name, "status": "created"}, nil))
}

// AlterTable executes ALTER TABLE (add/drop columns, rename).
func (h DDLHandler) AlterTable(w http.ResponseWriter, r *http.Request) {
	db, conn, dbName, err := h.openDB(r)
	if err != nil {
		h.writeConnErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) { return }
	table := router.Param(r, "table")
	var req alterTableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("invalid request body").Build()))
		return
	}
	stmts := buildAlterTable(dbName, table, req, conn.Driver)
	if len(stmts) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no alter operations specified").Build()))
		return
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(r.Context(), stmt); err != nil {
			h.Logger.Error("alter table", plumelog.Fields{"error": err.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("alter table failed").
				Detail("error", err.Error()).Build()))
			return
		}
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK,
		map[string]string{"table": table, "status": "altered"}, nil))
}

// DropTable executes DROP TABLE. Requires ?confirm=true to prevent accidental drops.
func (h DDLHandler) DropTable(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("confirm") != "true" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required: add ?confirm=true").Build()))
		return
	}
	db, conn, dbName, err := h.openDB(r)
	if err != nil {
		h.writeConnErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) { return }
	table := router.Param(r, "table")
	var fqn string
	switch conn.Driver {
	case connection.DriverMySQL:
		fqn = fmt.Sprintf("`%s`.`%s`", dbName, table)
	default:
		fqn = fmt.Sprintf(`"%s"`, table)
	}
	if _, err := db.ExecContext(r.Context(), "DROP TABLE IF EXISTS "+fqn); err != nil {
		h.Logger.Error("drop table", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to drop table").
			Detail("error", err.Error()).Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h DDLHandler) openDB(r *http.Request) (*sql.DB, *connection.Connection, string, error) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, nil, "", err
	}
	db, err := h.Manager.Open(conn)
	if err != nil {
		return nil, conn, "", err
	}
	return db, conn, router.Param(r, "db"), nil
}

func (h DDLHandler) writeConnErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == connection.ErrNotFound {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("connection not found").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("failed to connect").
		Detail("error", err.Error()).Build()))
}

func buildCreateTable(dbName string, req createTableRequest, driver connection.DriverType) string {
	var sb strings.Builder
	switch driver {
	case connection.DriverMySQL:
		sb.WriteString(fmt.Sprintf("CREATE TABLE `%s`.`%s` (\n", dbName, req.Name))
	default:
		sb.WriteString(fmt.Sprintf("CREATE TABLE \"%s\" (\n", req.Name))
	}
	for i, col := range req.Columns {
		sb.WriteString("  ")
		sb.WriteString(quoteIdent(col.Name, driver))
		sb.WriteString(" ")
		sb.WriteString(col.Type)
		if !col.Nullable {
			sb.WriteString(" NOT NULL")
		}
		if col.Default != "" {
			sb.WriteString(" DEFAULT " + col.Default)
		}
		if driver == connection.DriverMySQL && col.Comment != "" {
			sb.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.Comment, "'", "''")))
		}
		if i < len(req.Columns)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(")")
	if driver == connection.DriverMySQL {
		engine := sanitizeIdentifier(req.Engine)
		if engine == "" {
			engine = "InnoDB"
		}
		sb.WriteString(" ENGINE=" + engine)
		charset := sanitizeIdentifier(req.Charset)
		if charset == "" {
			charset = "utf8mb4"
		}
		sb.WriteString(" DEFAULT CHARSET=" + charset)
	}
	return sb.String()
}

// validateDDLLiteral rejects DEFAULT values that contain SQL statement terminators
// or comment markers to prevent DDL injection via user-supplied column defaults.
func validateDDLLiteral(s string) error {
	if len(s) > 200 {
		return fmt.Errorf("value too long (max 200 chars)")
	}
	for _, bad := range []string{";", "/*", "*/", "--", "\x00"} {
		if strings.Contains(s, bad) {
			return fmt.Errorf("value contains unsafe characters")
		}
	}
	return nil
}

// validateEngine allowlists MySQL storage engine names.
func validateEngine(s string) error {
	if s == "" {
		return nil
	}
	allowed := map[string]bool{
		"innodb": true, "myisam": true, "memory": true,
		"archive": true, "csv": true, "blackhole": true,
	}
	if !allowed[strings.ToLower(s)] {
		return fmt.Errorf("unsupported engine %q", s)
	}
	return nil
}

func buildAlterTable(dbName, table string, req alterTableRequest, driver connection.DriverType) []string {
	var stmts []string
	switch driver {
	case connection.DriverMySQL:
		fqn := fmt.Sprintf("`%s`.`%s`", dbName, table)
		var parts []string
		for _, col := range req.AddColumns {
			parts = append(parts, fmt.Sprintf("ADD COLUMN %s %s", quoteIdent(col.Name, driver), col.Type))
		}
		for _, col := range req.DropColumns {
			parts = append(parts, fmt.Sprintf("DROP COLUMN %s", quoteIdent(col, driver)))
		}
		if len(parts) > 0 {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s %s", fqn, strings.Join(parts, ", ")))
		}
		if req.RenameTable != "" {
			stmts = append(stmts, fmt.Sprintf("RENAME TABLE %s TO `%s`.`%s`", fqn, dbName, req.RenameTable))
		}
	default:
		// SQLite: only ADD COLUMN and RENAME TABLE are supported.
		fqn := fmt.Sprintf(`"%s"`, table)
		for _, col := range req.AddColumns {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
				fqn, quoteIdent(col.Name, driver), col.Type))
		}
		if req.RenameTable != "" {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME TO \"%s\"", fqn, req.RenameTable))
		}
	}
	return stmts
}
