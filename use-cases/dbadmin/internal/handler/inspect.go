package handler

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	mysqlinspect "dbadmin/internal/dbmanager/mysql"
	postgresinspect "dbadmin/internal/dbmanager/postgres"
	sqliteinspect "dbadmin/internal/dbmanager/sqlite"
	"dbadmin/internal/domain/connection"
)

// InspectHandler provides database introspection endpoints.
type InspectHandler struct {
	Connections *connection.Store
	Manager     *dbmanager.Manager
	Logger      plumelog.StructuredLogger
}

// openInspector opens the connection pool and returns the appropriate Inspector.
func (h InspectHandler) openInspector(r *http.Request) (dbmanager.Inspector, *connection.Connection, error) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, nil, err
	}
	db, err := h.Manager.Open(conn)
	if err != nil {
		return nil, conn, err
	}
	switch conn.Driver {
	case connection.DriverMySQL:
		return mysqlinspect.New(db), conn, nil
	case connection.DriverPostgres:
		return postgresinspect.New(db), conn, nil
	case connection.DriverSQLite:
		return sqliteinspect.New(db), conn, nil
	default:
		return nil, conn, fmt.Errorf("unsupported driver: %s", conn.Driver)
	}
}

// TableStructure returns columns, indexes, and foreign keys for a table.
func (h InspectHandler) TableStructure(w http.ResponseWriter, r *http.Request) {
	ins, conn, err := h.openInspector(r)
	if err != nil {
		h.writeInspectErr(w, r, err)
		return
	}
	db := router.Param(r, "db")
	table := router.Param(r, "table")
	if conn.Driver == connection.DriverSQLite {
		db = ""
	}

	cols, err := ins.Columns(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list columns", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to get columns").Build()))
		return
	}
	indexes, err := ins.Indexes(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list indexes", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to get indexes").Build()))
		return
	}
	fks, err := ins.ForeignKeys(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list foreign keys", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to get foreign keys").Build()))
		return
	}
	ddl, _ := ins.CreateTableDDL(r.Context(), db, table) // best-effort

	type structureResponse struct {
		Columns     []dbmanager.ColumnInfo     `json:"columns"`
		Indexes     []dbmanager.IndexInfo      `json:"indexes"`
		ForeignKeys []dbmanager.ForeignKeyInfo `json:"foreign_keys"`
		DDL         string                     `json:"ddl,omitempty"`
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, structureResponse{
		Columns:     cols,
		Indexes:     indexes,
		ForeignKeys: fks,
		DDL:         ddl,
	}, nil))
}

// Columns lists columns for a table.
func (h InspectHandler) Columns(w http.ResponseWriter, r *http.Request) {
	ins, conn, err := h.openInspector(r)
	if err != nil {
		h.writeInspectErr(w, r, err)
		return
	}
	db := router.Param(r, "db")
	table := router.Param(r, "table")
	if conn.Driver == connection.DriverSQLite {
		db = ""
	}
	cols, err := ins.Columns(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list columns", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list columns").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, cols, map[string]any{"count": len(cols)}))
}

// Indexes lists indexes for a table.
func (h InspectHandler) Indexes(w http.ResponseWriter, r *http.Request) {
	ins, conn, err := h.openInspector(r)
	if err != nil {
		h.writeInspectErr(w, r, err)
		return
	}
	db := router.Param(r, "db")
	table := router.Param(r, "table")
	if conn.Driver == connection.DriverSQLite {
		db = ""
	}
	indexes, err := ins.Indexes(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list indexes", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list indexes").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, indexes, map[string]any{"count": len(indexes)}))
}

// ForeignKeys lists foreign keys for a table.
func (h InspectHandler) ForeignKeys(w http.ResponseWriter, r *http.Request) {
	ins, conn, err := h.openInspector(r)
	if err != nil {
		h.writeInspectErr(w, r, err)
		return
	}
	db := router.Param(r, "db")
	table := router.Param(r, "table")
	if conn.Driver == connection.DriverSQLite {
		db = ""
	}
	fks, err := ins.ForeignKeys(r.Context(), db, table)
	if err != nil {
		h.Logger.Error("list foreign keys", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list foreign keys").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, fks, map[string]any{"count": len(fks)}))
}

// SchemaDoc returns a Markdown document of all table DDLs in the database.
func (h InspectHandler) SchemaDoc(w http.ResponseWriter, r *http.Request) {
	ins, conn, err := h.openInspector(r)
	if err != nil {
		h.writeInspectErr(w, r, err)
		return
	}
	db := router.Param(r, "db")
	if conn.Driver == connection.DriverSQLite {
		db = ""
	}

	tables, err := ins.Tables(r.Context(), db)
	if err != nil {
		h.Logger.Error("list tables for schema doc", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to list tables").Build()))
		return
	}

	var sb strings.Builder
	dbLabel := db
	if dbLabel == "" {
		dbLabel = router.Param(r, "db")
	}
	sb.WriteString(fmt.Sprintf("# Schema: %s\n\n", dbLabel))

	for _, tbl := range tables {
		sb.WriteString(fmt.Sprintf("## `%s`\n\n", tbl.Name))
		ddl, _ := ins.CreateTableDDL(r.Context(), db, tbl.Name)
		if ddl != "" {
			sb.WriteString("```sql\n")
			sb.WriteString(ddl)
			sb.WriteString("\n```\n\n")
		}
	}

	type schemaDocResp struct {
		Markdown string `json:"markdown"`
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, schemaDocResp{Markdown: sb.String()}, nil))
}

func (h InspectHandler) writeInspectErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == connection.ErrNotFound {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("connection not found").Build()))
		return
	}
	h.Logger.Error("open inspector", plumelog.Fields{"error": err.Error()})
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("failed to connect to database").
		Detail("error", err.Error()).Build()))
}
