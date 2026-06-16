package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	plumelog "github.com/spcent/plumego/log"
	"github.com/spcent/plumego/router"

	"dbadmin/internal/dbmanager"
	mysqlinspect "dbadmin/internal/dbmanager/mysql"
	sqliteinspect "dbadmin/internal/dbmanager/sqlite"
	"dbadmin/internal/domain/connection"
)

// RowHandler handles data row CRUD operations.
type RowHandler struct {
	Connections *connection.Store
	Manager     *dbmanager.Manager
	Logger      plumelog.StructuredLogger
}

type rowsResponse struct {
	Rows            []map[string]any `json:"rows"`
	Total           int64            `json:"total"`
	Page            int              `json:"page"`
	PageSize        int              `json:"pageSize"`
	Columns         []string         `json:"columns"`
	ExecutionTimeMs int64            `json:"executionTimeMs"`
}

type filterCondition struct {
	Column   string `json:"column"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type insertRequest struct {
	Values map[string]any `json:"values"`
}

type updateRequest struct {
	PrimaryKey map[string]any `json:"primaryKey"`
	Values     map[string]any `json:"values"`
	Confirm    bool           `json:"confirm"`
}

type deleteRequest struct {
	PrimaryKey map[string]any `json:"primaryKey"`
	Confirm    bool           `json:"confirm"`
}

type bulkDeleteRequest struct {
	PrimaryKeyColumn string `json:"primary_key_column"`
	Values           []any  `json:"values"`
	Confirm          bool   `json:"confirm"`
}

type bulkUpdateRequest struct {
	PrimaryKeyColumn string         `json:"primary_key_column"`
	Values           []any          `json:"values"`
	Set              map[string]any `json:"set"`
	Confirm          bool           `json:"confirm"`
}

type bulkRowsResponse struct {
	RowsAffected int64 `json:"rowsAffected"`
}

// List returns paginated rows from a table.
// Query params: page, pageSize (max 500), sortColumn, sortDirection (asc|desc),
// filters (JSON []filterCondition), selectedColumns (comma-separated).
func (h RowHandler) List(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	dbName := router.Param(r, "db")
	table := router.Param(r, "table")
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	pageSize, _ := strconv.Atoi(q.Get("pageSize"))
	page, pageSize = normalizePagination(page, pageSize)
	offset := (page - 1) * pageSize

	// Fetch column list once — reused for sort, filter, and selectedColumns validation.
	// An empty colSet means the inspector was unavailable; validation falls back to DB errors.
	sortColRaw := sanitizeIdentifier(q.Get("sortColumn"))
	colSet := make(map[string]bool)
	if sortColRaw != "" || q.Get("filters") != "" || q.Get("selectedColumns") != "" {
		inspector := h.newInspector(conn, db)
		if inspCols, inspErr := inspector.Columns(r.Context(), dbName, table); inspErr == nil {
			for _, c := range inspCols {
				colSet[c.Name] = true
			}
		} else if sortColRaw != "" {
			// Cannot validate sort column — fail safe rather than proceed unvalidated.
			h.Logger.Warn("inspector unavailable for sort validation", plumelog.Fields{"error": inspErr.Error()})
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeInternal).Message("failed to inspect table columns").Build()))
			return
		}
	}

	// Build ORDER BY — sortColumn must exist in the column set when colSet is populated.
	orderBy := ""
	if sortColRaw != "" {
		if len(colSet) > 0 && !colSet[sortColRaw] {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).Message("unknown sort column").Build()))
			return
		}
		dir := "ASC"
		if strings.EqualFold(q.Get("sortDirection"), "desc") {
			dir = "DESC"
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", quoteIdent(sortColRaw, conn.Driver), dir)
	}

	// Parse filters from JSON, validate column names, then build WHERE clause.
	var where string
	var whereArgs []any
	if filtersJSON := q.Get("filters"); filtersJSON != "" {
		var filters []filterCondition
		if err = json.Unmarshal([]byte(filtersJSON), &filters); err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).Message("invalid filters JSON").Build()))
			return
		}
		// Reject unknown or empty column names when we have column metadata.
		for _, f := range filters {
			col := sanitizeIdentifier(f.Column)
			if col == "" {
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeBadRequest).Message("invalid filter column name").Build()))
				return
			}
			if len(colSet) > 0 && !colSet[col] {
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeBadRequest).Message("unknown filter column: "+col).Build()))
				return
			}
		}
		where, whereArgs, err = buildFiltersWhere(filters, conn.Driver)
		if err != nil {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).Message(err.Error()).Build()))
			return
		}
	}

	// Resolve selected columns — validate against colSet when available; fall back to *.
	selectExpr := "*"
	if selCols := q.Get("selectedColumns"); selCols != "" {
		var quoted []string
		for _, p := range strings.Split(selCols, ",") {
			col := sanitizeIdentifier(strings.TrimSpace(p))
			if col == "" {
				continue
			}
			if len(colSet) > 0 && !colSet[col] {
				logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
					Type(contract.TypeBadRequest).Message("unknown selected column: "+col).Build()))
				return
			}
			quoted = append(quoted, quoteIdent(col, conn.Driver))
		}
		if len(quoted) > 0 {
			selectExpr = strings.Join(quoted, ", ")
		}
	}

	start := time.Now()

	// Count total matching rows.
	var total int64
	countQ := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", tableFQN, where)
	if err = db.QueryRowContext(r.Context(), countQ, whereArgs...).Scan(&total); err != nil {
		h.Logger.Error("count rows", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to count rows").Build()))
		return
	}

	// Fetch paginated rows.
	nArgs := len(whereArgs)
	selectQ := fmt.Sprintf("SELECT %s FROM %s%s%s LIMIT %s OFFSET %s",
		selectExpr, tableFQN, where, orderBy,
		nthPlaceholder(conn.Driver, nArgs+1),
		nthPlaceholder(conn.Driver, nArgs+2))
	args := append(whereArgs, pageSize, offset)
	rows, err := db.QueryContext(r.Context(), selectQ, args...)
	if err != nil {
		h.Logger.Error("select rows", plumelog.Fields{"error": err.Error(), "db": dbName, "table": table})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to fetch rows").Build()))
		return
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result, _, err := scanRows(rows, cols)
	if err != nil {
		h.Logger.Error("scan rows", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to scan rows").Build()))
		return
	}

	maskRowMaps(result, maskedColumnSet(conn.MaskedColumns))

	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, rowsResponse{
		Rows:            result,
		Total:           total,
		Page:            page,
		PageSize:        pageSize,
		Columns:         cols,
		ExecutionTimeMs: time.Since(start).Milliseconds(),
	}, nil))
}

// Create inserts a new row. Body: {"values": {"col": val, ...}}.
// JSON null values are inserted as SQL NULL.
func (h RowHandler) Create(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req insertRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if len(req.Values) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("values is required").Build()))
		return
	}

	cols, vals := sortedColumnsValues(req.Values)
	q := buildInsertSQL(conn.Driver, tableFQN, cols)

	if _, err := db.ExecContext(r.Context(), q, vals...); err != nil {
		h.Logger.Error("insert row", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to insert row").
			Detail("error", err.Error()).Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusCreated, req.Values, nil))
}

// Update updates an existing row. Body: {"primaryKey":{"id":1}, "values":{...}, "confirm":true}.
// Cannot update PK columns. Guaranteed non-empty WHERE (primaryKey required).
func (h RowHandler) Update(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req updateRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required: set confirm=true in request body").Build()))
		return
	}
	if len(req.PrimaryKey) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("primaryKey is required").Build()))
		return
	}
	if len(req.Values) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("values is required").Build()))
		return
	}

	// Remove PK columns from values — cannot update primary key.
	for k := range req.PrimaryKey {
		delete(req.Values, k)
	}
	if len(req.Values) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no non-PK fields to update").Build()))
		return
	}

	setCols, setVals := sortedColumnsValues(req.Values)
	whereSQL, whereArgs, err := buildWhereFromPK(req.PrimaryKey, conn.Driver, len(setCols)+1)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message(err.Error()).Build()))
		return
	}

	q := fmt.Sprintf("UPDATE %s SET %s%s",
		tableFQN, buildUpdateSetClause(conn.Driver, setCols), whereSQL)
	args := append(setVals, whereArgs...)

	res, err := db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.Logger.Error("update row", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to update row").
			Detail("error", err.Error()).Build()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("row not found").Build()))
		return
	}
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, req.Values, nil))
}

// Delete removes a row. Body: {"primaryKey":{"id":1}, "confirm":true}.
// Guaranteed non-empty WHERE (primaryKey required).
func (h RowHandler) Delete(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req deleteRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required: set confirm=true in request body").Build()))
		return
	}
	if len(req.PrimaryKey) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("primaryKey is required").Build()))
		return
	}

	whereSQL, whereArgs, err := buildWhereFromPK(req.PrimaryKey, conn.Driver, 1)
	if err != nil {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message(err.Error()).Build()))
		return
	}

	q := fmt.Sprintf("DELETE FROM %s%s", tableFQN, whereSQL)
	res, err := db.ExecContext(r.Context(), q, whereArgs...)
	if err != nil {
		h.Logger.Error("delete row", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to delete row").Build()))
		return
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("row not found").Build()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// BulkDelete deletes all rows matching primary_key_column IN (values).
// Body: {"primary_key_column":"id","values":[1,2,3],"confirm":true}.
// The number of values is capped at MaxBulkRowOperationRows.
func (h RowHandler) BulkDelete(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req bulkDeleteRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required: set confirm=true in request body").Build()))
		return
	}
	pkCol := sanitizeIdentifier(req.PrimaryKeyColumn)
	if pkCol == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("primary_key_column is required").Build()))
		return
	}
	if len(req.Values) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("values is required").Build()))
		return
	}
	if len(req.Values) > MaxBulkRowOperationRows {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(fmt.Sprintf("values is limited to %d rows", MaxBulkRowOperationRows)).Build()))
		return
	}

	whereSQL, whereArgs := buildWhereInClause(pkCol, req.Values, conn.Driver, 1)
	q := fmt.Sprintf("DELETE FROM %s%s", tableFQN, whereSQL)

	res, err := db.ExecContext(r.Context(), q, whereArgs...)
	if err != nil {
		h.Logger.Error("bulk delete rows", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to delete rows").
			Detail("error", err.Error()).Build()))
		return
	}
	n, _ := res.RowsAffected()
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, bulkRowsResponse{RowsAffected: n}, nil))
}

// BulkUpdate sets the same column values for every row matching
// primary_key_column IN (values).
// Body: {"primary_key_column":"id","values":[1,2,3],"set":{"col":val},"confirm":true}.
// The number of values is capped at MaxBulkRowOperationRows.
func (h RowHandler) BulkUpdate(w http.ResponseWriter, r *http.Request) {
	db, conn, tableFQN, err := h.openTable(r)
	if err != nil {
		h.writeErr(w, r, err)
		return
	}
	if guardReadonly(conn, w, r, h.Logger) {
		return
	}
	var req bulkUpdateRequest
	if !decodeJSONLimited(w, r, h.Logger, &req) {
		return
	}
	if !req.Confirm {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("confirm required: set confirm=true in request body").Build()))
		return
	}
	pkCol := sanitizeIdentifier(req.PrimaryKeyColumn)
	if pkCol == "" {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("primary_key_column is required").Build()))
		return
	}
	if len(req.Values) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("values is required").Build()))
		return
	}
	if len(req.Values) > MaxBulkRowOperationRows {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message(fmt.Sprintf("values is limited to %d rows", MaxBulkRowOperationRows)).Build()))
		return
	}
	if len(req.Set) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("set is required").Build()))
		return
	}

	// Cannot update the primary key column itself via bulk update.
	delete(req.Set, pkCol)
	if len(req.Set) == 0 {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).Message("no non-PK fields to update").Build()))
		return
	}

	setCols, setVals := sortedColumnsValues(req.Set)
	for _, c := range setCols {
		if sanitizeIdentifier(c) != c {
			logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
				Type(contract.TypeBadRequest).Message("invalid set column name: "+c).Build()))
			return
		}
	}

	whereSQL, whereArgs := buildWhereInClause(pkCol, req.Values, conn.Driver, len(setCols)+1)
	q := fmt.Sprintf("UPDATE %s SET %s%s",
		tableFQN, buildUpdateSetClause(conn.Driver, setCols), whereSQL)
	args := append(setVals, whereArgs...)

	res, err := db.ExecContext(r.Context(), q, args...)
	if err != nil {
		h.Logger.Error("bulk update rows", plumelog.Fields{"error": err.Error()})
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeInternal).Message("failed to update rows").
			Detail("error", err.Error()).Build()))
		return
	}
	n, _ := res.RowsAffected()
	logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, bulkRowsResponse{RowsAffected: n}, nil))
}

// openTable resolves and opens the DB connection and returns the fully-qualified table name.
func (h RowHandler) openTable(r *http.Request) (*sql.DB, *connection.Connection, string, error) {
	connID := router.Param(r, "id")
	conn, err := h.Connections.Get(connID)
	if err != nil {
		return nil, nil, "", err
	}
	db, err := h.Manager.Open(r.Context(), conn)
	if err != nil {
		return nil, conn, "", err
	}
	dbName := router.Param(r, "db")
	table := router.Param(r, "table")
	var tableFQN string
	switch conn.Driver {
	case connection.DriverMySQL:
		tableFQN = fmt.Sprintf("%s.%s", quoteIdent(dbName, conn.Driver), quoteIdent(table, conn.Driver))
	default:
		tableFQN = quoteIdent(table, conn.Driver)
	}
	return db, conn, tableFQN, nil
}

func (h RowHandler) newInspector(conn *connection.Connection, db *sql.DB) dbmanager.Inspector {
	switch conn.Driver {
	case connection.DriverMySQL:
		return mysqlinspect.New(db)
	default:
		return sqliteinspect.New(db)
	}
}

func (h RowHandler) writeErr(w http.ResponseWriter, r *http.Request, err error) {
	if err == connection.ErrNotFound {
		logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).Message("connection not found").Build()))
		return
	}
	h.Logger.Error("open connection", plumelog.Fields{"error": err.Error()})
	logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
		Type(contract.TypeInternal).Message("failed to connect to database").
		Detail("error", err.Error()).Build()))
}

// scanRows converts sql.Rows into a slice of maps.
// Large text and BLOB values are automatically truncated using DefaultPreviewLimits.
// The returned boolean indicates whether any value was truncated.
func scanRows(rows *sql.Rows, cols []string) ([]map[string]any, bool, error) {
	limits := DefaultPreviewLimits()
	result := make([]map[string]any, 0)
	anyTruncated := false
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, false, err
		}
		row := make(map[string]any, len(cols))
		for i, col := range cols {
			row[col] = vals[i]
		}
		row, truncated := limits.ApplyPreviewToRow(row)
		if truncated {
			anyTruncated = true
		}
		result = append(result, row)
	}
	return result, anyTruncated, rows.Err()
}

// sortedColumnsValues returns columns and values from a map in sorted key order.
// Sorted order is required for SQLite positional placeholders (?1, ?2, …) to be deterministic.
func sortedColumnsValues(m map[string]any) ([]string, []any) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	vals := make([]any, len(keys))
	for i, k := range keys {
		vals[i] = m[k]
	}
	return keys, vals
}

// mapToColumnsValues is kept for compatibility with existing callers.
func mapToColumnsValues(m map[string]any) ([]string, []any) {
	return sortedColumnsValues(m)
}

// buildWhereFromPK builds a WHERE clause from a primary-key map.
// Keys are sorted for deterministic placeholder numbering.
// startN is the 1-based index of the first placeholder (follows any SET params in UPDATE).
// Returns an error if pkMap is empty or any key sanitises to an empty string.
func buildWhereFromPK(pkMap map[string]any, driver connection.DriverType, startN int) (string, []any, error) {
	if len(pkMap) == 0 {
		return "", nil, fmt.Errorf("primaryKey must not be empty")
	}
	keys := make([]string, 0, len(pkMap))
	for k := range pkMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		col := sanitizeIdentifier(k)
		if col == "" {
			return "", nil, fmt.Errorf("invalid primary key column name: %q", k)
		}
		parts[i] = fmt.Sprintf("%s = %s", quoteIdent(col, driver), nthPlaceholder(driver, startN+i))
		args[i] = pkMap[k]
	}
	return " WHERE " + strings.Join(parts, " AND "), args, nil
}

// buildWhereInClause builds a parameterized "WHERE col IN (?, ?, ...)" clause
// for the given column and values, using the connection driver's placeholder
// style. startN is the 1-based index of the first placeholder (follows any
// SET params in an UPDATE). The caller must validate col is a safe identifier
// before calling this function.
func buildWhereInClause(col string, values []any, driver connection.DriverType, startN int) (string, []any) {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = nthPlaceholder(driver, startN+i)
	}
	where := fmt.Sprintf(" WHERE %s IN (%s)", quoteIdent(col, driver), strings.Join(placeholders, ", "))
	args := make([]any, len(values))
	copy(args, values)
	return where, args
}

func quoteIdent(name string, driver connection.DriverType) string {
	switch driver {
	case connection.DriverMySQL:
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	default:
		return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
	}
}

func quoteIdents(names []string, driver connection.DriverType) []string {
	out := make([]string, len(names))
	for i, n := range names {
		out[i] = quoteIdent(n, driver)
	}
	return out
}

func nthPlaceholder(driver connection.DriverType, n int) string {
	if driver == connection.DriverMySQL {
		return "?"
	}
	return fmt.Sprintf("?%d", n)
}

func buildPlaceholders(driver connection.DriverType, count int) []string {
	phs := make([]string, count)
	for i := range phs {
		phs[i] = nthPlaceholder(driver, i+1)
	}
	return phs
}

// sanitizeIdentifier removes characters that aren't alphanumeric or underscore.
func sanitizeIdentifier(s string) string {
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normalizePagination clamps page and pageSize to valid ranges.
// page < 1 -> 1; pageSize <= 0 or > 500 -> 50.
func normalizePagination(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > MaxPageSize {
		pageSize = 50
	}
	return page, pageSize
}

// buildInsertSQL returns an INSERT INTO ... VALUES (...) template without values.
func buildInsertSQL(driver connection.DriverType, tableFQN string, cols []string) string {
	ph := buildPlaceholders(driver, len(cols))
	return fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		tableFQN, strings.Join(quoteIdents(cols, driver), ", "), strings.Join(ph, ", "))
}

// buildUpdateSetClause returns the SET portion of an UPDATE statement.
// Placeholders start at ?1 (SQLite) or ? (MySQL) from position 1.
func buildUpdateSetClause(driver connection.DriverType, cols []string) string {
	parts := make([]string, len(cols))
	for i, c := range cols {
		parts[i] = fmt.Sprintf("%s = %s", quoteIdent(c, driver), nthPlaceholder(driver, i+1))
	}
	return strings.Join(parts, ", ")
}

// buildFiltersWhere builds a WHERE clause from a slice of filter conditions.
// Invalid column names and unknown operators are silently skipped.
// Returns ("", nil, nil) when no valid conditions are produced.
func buildFiltersWhere(filters []filterCondition, driver connection.DriverType) (string, []any, error) {
	var parts []string
	var args []any
	paramIdx := 1
	for _, f := range filters {
		col := sanitizeIdentifier(f.Column)
		if col == "" {
			continue
		}
		quoted := quoteIdent(col, driver)
		ph := func() string { return nthPlaceholder(driver, paramIdx) }
		switch f.Operator {
		case "eq":
			parts = append(parts, fmt.Sprintf("%s = %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "ne":
			parts = append(parts, fmt.Sprintf("%s != %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "gt":
			parts = append(parts, fmt.Sprintf("%s > %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "gte":
			parts = append(parts, fmt.Sprintf("%s >= %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "lt":
			parts = append(parts, fmt.Sprintf("%s < %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "lte":
			parts = append(parts, fmt.Sprintf("%s <= %s", quoted, ph()))
			args = append(args, f.Value)
			paramIdx++
		case "like":
			parts = append(parts, fmt.Sprintf("%s LIKE %s", quoted, ph()))
			args = append(args, "%"+f.Value+"%")
			paramIdx++
		case "not_like":
			parts = append(parts, fmt.Sprintf("%s NOT LIKE %s", quoted, ph()))
			args = append(args, "%"+f.Value+"%")
			paramIdx++
		case "is_null":
			parts = append(parts, fmt.Sprintf("%s IS NULL", quoted))
		case "is_not_null":
			parts = append(parts, fmt.Sprintf("%s IS NOT NULL", quoted))
		default:
			continue
		}
	}
	if len(parts) == 0 {
		return "", nil, nil
	}
	return " WHERE " + strings.Join(parts, " AND "), args, nil
}
