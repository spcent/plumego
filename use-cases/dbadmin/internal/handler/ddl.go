package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
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
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req createTableRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if req.Name == "" || len(req.Columns) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("name and columns are required").Build()))
		return
	}
	if err := validateDDLIdentifier("table name", req.Name); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	if conn.Driver == connection.DriverMySQL {
		if err := validateDDLIdentifier("database name", dbName); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
	}
	for _, col := range req.Columns {
		if err := validateDDLIdentifier("column name", col.Name); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
		if err := validateColumnType(col.Type); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).
				Message("invalid type for column "+col.Name+": "+err.Error()).Build()))
			return
		}
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
	// codeql[go/sql-injection]: DDL identifiers and literals are validated above,
	// then identifiers are quoted with driver-specific escaping before execution.
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
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	table := router.Param(r, "table")
	var req alterTableRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if err := validateDDLIdentifier("table name", table); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	if conn.Driver == connection.DriverMySQL {
		if err := validateDDLIdentifier("database name", dbName); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
	}
	for _, col := range req.AddColumns {
		if err := validateDDLIdentifier("column name", col.Name); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
		if err := validateColumnType(col.Type); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).
				Message("invalid type for column "+col.Name+": "+err.Error()).Build()))
			return
		}
	}
	for _, col := range req.DropColumns {
		if err := validateDDLIdentifier("column name", col); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
	}
	if req.RenameTable != "" {
		if err := validateDDLIdentifier("table name", req.RenameTable); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
	}
	stmts := buildAlterTable(dbName, table, req, conn.Driver)
	if len(stmts) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no alter operations specified").Build()))
		return
	}
	for _, stmt := range stmts {
		// codeql[go/sql-injection]: ALTER statements are assembled only after
		// identifier/type validation, with all identifiers quoted by buildAlterTable.
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
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	table := router.Param(r, "table")
	if err := validateDDLIdentifier("table name", table); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(err.Error()).Build()))
		return
	}
	if conn.Driver == connection.DriverMySQL {
		if err := validateDDLIdentifier("database name", dbName); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeValidation).Message(err.Error()).Build()))
			return
		}
	}
	var fqn string
	switch conn.Driver {
	case connection.DriverMySQL:
		fqn = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(table, conn.Driver))
	case connection.DriverPostgres:
		fqn = postgresTableFQN(dbName, table)
	default:
		fqn = quoteIdent(table, conn.Driver)
	}
	// codeql[go/sql-injection]: DROP TABLE uses validated identifiers quoted via quoteIdent.
	if _, err := db.ExecContext(r.Context(), "DROP TABLE IF EXISTS "+fqn); err != nil {
		h.Logger.Error("drop table", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to drop table").
			Detail("error", err.Error()).Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createViewRequest struct {
	Name  string `json:"name"`
	Query string `json:"query"`
}

// maxViewQueryLen bounds the size of a CREATE VIEW query body. The body is
// not parsed or validated beyond length/non-empty checks — it is trusted
// user-supplied SQL, the same trust model applied to CreateTable's column DDL.
const maxViewQueryLen = 20000

// CreateView executes CREATE VIEW.
func (h DDLHandler) CreateView(w http.ResponseWriter, r *http.Request) {
	db, conn, dbName, err := h.openDB(r)
	if err != nil {
		h.writeConnErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req createViewRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if req.Name == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("name is required").Build()))
		return
	}
	if err := validateIdentifierName(req.Name); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("invalid view name: "+err.Error()).Build()))
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("query is required").Build()))
		return
	}
	if len(query) > maxViewQueryLen {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message(fmt.Sprintf("query too long (max %d chars)", maxViewQueryLen)).Build()))
		return
	}
	var fqn string
	switch conn.Driver {
	case connection.DriverMySQL:
		fqn = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(req.Name, conn.Driver))
	default:
		fqn = quoteIdent(req.Name, conn.Driver)
	}
	ddl := "CREATE VIEW " + fqn + " AS " + query
	if _, err := db.ExecContext(r.Context(), ddl); err != nil { // codeql[go/sql-injection-builtin]
		h.Logger.Error("create view", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to create view").
			Detail("error", err.Error()).Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated,
		map[string]string{"view": req.Name, "status": "created"}, nil))
}

// DropView executes DROP VIEW. Requires ?confirm=true to prevent accidental drops.
func (h DDLHandler) DropView(w http.ResponseWriter, r *http.Request) {
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
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	view := router.Param(r, "view")
	if err := validateIdentifierName(view); err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeValidation).Message("invalid view name: "+err.Error()).Build()))
		return
	}
	var fqn string
	switch conn.Driver {
	case connection.DriverMySQL:
		fqn = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(view, conn.Driver))
	default:
		fqn = quoteIdent(view, conn.Driver)
	}
	if _, err := db.ExecContext(r.Context(), "DROP VIEW IF EXISTS "+fqn); err != nil { // codeql[go/sql-injection]
		h.Logger.Error("drop view", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to drop view").
			Detail("error", err.Error()).Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// validateIdentifierName rejects names that contain SQL statement terminators
// or comment markers, mirroring validateColumnType's blocklist approach. The
// name is later passed through quoteIdent for driver-specific escaping.
func validateIdentifierName(s string) error {
	if s == "" {
		return fmt.Errorf("name is required")
	}
	if len(s) > 64 {
		return fmt.Errorf("name too long (max 64 chars)")
	}
	for _, bad := range []string{";", "--", "/*", "*/", "\x00"} {
		if strings.Contains(s, bad) {
			return fmt.Errorf("name contains unsafe characters")
		}
	}
	return nil
}

func (h DDLHandler) openDB(r *http.Request) (*sql.DB, *connection.Connection, string, error) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, nil, "", err
	}
	db, err := h.Manager.Open(r.Context(), conn)
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

// postgresTableFQN builds a schema-qualified, quoted PostgreSQL table
// reference. An empty schema is omitted, relying on the connection's
// search_path (defaults to "public").
func postgresTableFQN(schema, table string) string {
	if schema == "" {
		return fmt.Sprintf(`"%s"`, table)
	}
	return fmt.Sprintf(`"%s"."%s"`, schema, table)
}

func buildCreateTable(dbName string, req createTableRequest, driver connection.DriverType) string {
	var sb strings.Builder
	switch driver {
	case connection.DriverMySQL:
		sb.WriteString(fmt.Sprintf("CREATE TABLE %s.%s (\n", quoteIdent(dbName, driver), quoteIdent(req.Name, driver)))
	case connection.DriverPostgres:
		sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", postgresTableFQN(dbName, req.Name)))
	default:
		sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoteIdent(req.Name, driver)))
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

// ddlIdentifierRe allows only SQL-safe identifier characters: start with a
// letter, underscore, or dollar sign; remainder letters, digits, underscore,
// dollar sign.  This is an allowlist — anything outside it is rejected even
// though quoteIdent would technically make it safe, giving defence-in-depth.
var ddlIdentifierRe = regexp.MustCompile(`^[a-zA-Z_$][a-zA-Z0-9_$]*$`)

// columnTypeRe allows standard SQL data-type syntax: one or more
// space-separated words (e.g. DOUBLE PRECISION, INT UNSIGNED) optionally
// followed by a single (n) or (n,m) precision clause.  Everything else —
// semicolons, commas outside parens, dashes, comment markers — is rejected.
var columnTypeRe = regexp.MustCompile(`(?i)^[a-z][a-z0-9]*(?:\s+[a-z][a-z0-9]*)*(?:\(\d+(?:,\d+)?\))?$`)

func validateDDLIdentifier(kind, name string) error {
	if name == "" {
		return fmt.Errorf("%s is required", kind)
	}
	if len(name) > 64 {
		return fmt.Errorf("%s too long (max 64 chars)", kind)
	}
	if !ddlIdentifierRe.MatchString(name) {
		return fmt.Errorf("%s contains unsafe characters", kind)
	}
	return nil
}

// validateColumnType rejects column type strings outside the expected
// SQL-type grammar, using an allowlist regex rather than a blocklist.
func validateColumnType(s string) error {
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("column type is required")
	}
	if len(s) > 64 {
		return fmt.Errorf("column type too long (max 64 chars)")
	}
	if !columnTypeRe.MatchString(s) {
		return fmt.Errorf("column type contains unsafe characters")
	}
	return nil
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
		fqn := fmt.Sprintf("%s.%s", quoteIdent(dbName, driver), quoteIdent(table, driver))
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
			stmts = append(stmts, fmt.Sprintf("RENAME TABLE %s TO %s.%s",
				fqn, quoteIdent(dbName, driver), quoteIdent(req.RenameTable, driver)))
		}
	case connection.DriverPostgres:
		// PostgreSQL supports ADD COLUMN, DROP COLUMN, and RENAME TO in standard SQL.
		fqn := postgresTableFQN(dbName, table)
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
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME TO %s",
				fqn, quoteIdent(req.RenameTable, driver)))
		}
	default:
		// SQLite: only ADD COLUMN and RENAME TABLE are supported.
		fqn := quoteIdent(table, driver)
		for _, col := range req.AddColumns {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s",
				fqn, quoteIdent(col.Name, driver), col.Type))
		}
		if req.RenameTable != "" {
			stmts = append(stmts, fmt.Sprintf("ALTER TABLE %s RENAME TO %s",
				fqn, quoteIdent(req.RenameTable, driver)))
		}
	}
	return stmts
}
