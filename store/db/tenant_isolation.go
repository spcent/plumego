package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/spcent/plumego/tenant"
)

// TenantDB wraps sql.DB with automatic tenant filtering for queries.
// It ensures that all queries are scoped to a specific tenant by automatically
// injecting WHERE tenant_id = ? clauses.
type TenantDB struct {
	db           *sql.DB
	tenantColumn string
}

// TenantDBOption configures TenantDB behavior.
type TenantDBOption func(*TenantDB)

// WithTenantColumn sets the column name for tenant filtering (default: "tenant_id").
func WithTenantColumn(column string) TenantDBOption {
	return func(tdb *TenantDB) {
		tdb.tenantColumn = column
	}
}

// NewTenantDB creates a tenant-aware database wrapper.
//
// Example:
//
//	tdb := NewTenantDB(db)
//	rows, err := tdb.QueryFromContext(ctx, "SELECT * FROM users WHERE active = ?", true)
//	// Automatically becomes: SELECT * FROM users WHERE tenant_id = 'current-tenant' AND active = true
func NewTenantDB(db *sql.DB, options ...TenantDBOption) *TenantDB {
	tdb := &TenantDB{
		db:           db,
		tenantColumn: "tenant_id",
	}

	for _, opt := range options {
		opt(tdb)
	}

	return tdb
}

// QueryFromContext extracts tenant from context and executes a query with automatic filtering.
func (tdb *TenantDB) QueryFromContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	tenantID := tenant.TenantIDFromContext(ctx)
	if tenantID == "" {
		return nil, errors.New("tenant ID not found in context")
	}

	return tdb.QueryContext(ctx, tenantID, query, args...)
}

// QueryContext executes a query with automatic tenant filtering.
func (tdb *TenantDB) QueryContext(ctx context.Context, tenantID string, query string, args ...interface{}) (*sql.Rows, error) {
	if tdb == nil || tdb.db == nil {
		return nil, errors.New("database not initialized")
	}

	filteredQuery, filteredArgs := tdb.addTenantFilter(query, tenantID, args)
	return tdb.db.QueryContext(ctx, filteredQuery, filteredArgs...)
}

// QueryRowFromContext extracts tenant from context and executes a single-row query.
func (tdb *TenantDB) QueryRowFromContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	tenantID := tenant.TenantIDFromContext(ctx)
	return tdb.QueryRowContext(ctx, tenantID, query, args...)
}

// QueryRowContext executes a single-row query with automatic tenant filtering.
func (tdb *TenantDB) QueryRowContext(ctx context.Context, tenantID string, query string, args ...interface{}) *sql.Row {
	if tdb == nil || tdb.db == nil {
		// Return a row that will error when scanned
		return &sql.Row{}
	}

	filteredQuery, filteredArgs := tdb.addTenantFilter(query, tenantID, args)
	return tdb.db.QueryRowContext(ctx, filteredQuery, filteredArgs...)
}

// ExecFromContext extracts tenant from context and executes a statement.
func (tdb *TenantDB) ExecFromContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	tenantID := tenant.TenantIDFromContext(ctx)
	if tenantID == "" {
		return nil, errors.New("tenant ID not found in context")
	}

	return tdb.ExecContext(ctx, tenantID, query, args...)
}

// ExecContext executes a statement with automatic tenant filtering.
func (tdb *TenantDB) ExecContext(ctx context.Context, tenantID string, query string, args ...interface{}) (sql.Result, error) {
	if tdb == nil || tdb.db == nil {
		return nil, errors.New("database not initialized")
	}

	filteredQuery, filteredArgs := tdb.addTenantFilter(query, tenantID, args)
	return tdb.db.ExecContext(ctx, filteredQuery, filteredArgs...)
}

// addTenantFilter injects tenant_id filter into SQL query.
func (tdb *TenantDB) addTenantFilter(query string, tenantID string, args []interface{}) (string, []interface{}) {
	queryLower := strings.ToLower(strings.TrimSpace(query))

	// For INSERT statements, add tenant_id to the column list
	if strings.HasPrefix(queryLower, "insert") {
		return tdb.addTenantToInsert(query, tenantID, args)
	}

	// For UPDATE statements, handle SET...WHERE separately
	if strings.HasPrefix(queryLower, "update") {
		return tdb.addTenantToUpdate(query, tenantID, args)
	}

	// For SELECT, DELETE - add WHERE clause filter (tenant_id at beginning)
	whereClause := fmt.Sprintf("%s = ?", tdb.tenantColumn)
	newArgs := append([]interface{}{tenantID}, args...)

	if strings.Contains(queryLower, " where ") {
		// Existing WHERE clause - add AND condition
		whereIdx := -1
		if idx := strings.Index(query, "WHERE"); idx != -1 {
			whereIdx = idx
		} else if idx := strings.Index(query, "where"); idx != -1 {
			whereIdx = idx
		}

		if whereIdx != -1 {
			filteredQuery := query[:whereIdx] + "WHERE " + whereClause + " AND " + query[whereIdx+5:]
			return filteredQuery, newArgs
		}
	}

	// No WHERE clause - add one
	// Find position to insert WHERE (before ORDER BY, LIMIT, GROUP BY, etc.)
	insertPos := len(query)
	for _, keyword := range []string{" ORDER BY", " LIMIT", " GROUP BY", " HAVING", " FOR UPDATE"} {
		if idx := strings.Index(strings.ToUpper(query), keyword); idx != -1 && idx < insertPos {
			insertPos = idx
		}
	}

	filteredQuery := query[:insertPos] + " WHERE " + whereClause + query[insertPos:]
	return filteredQuery, newArgs
}

// addTenantToUpdate adds tenant_id filter to UPDATE statements.
// For UPDATE, the tenant_id parameter should go after SET parameters but before WHERE parameters.
func (tdb *TenantDB) addTenantToUpdate(query string, tenantID string, args []interface{}) (string, []interface{}) {
	whereClause := fmt.Sprintf("%s = ?", tdb.tenantColumn)

	// Find WHERE clause
	whereIdx := -1
	if idx := strings.Index(query, "WHERE"); idx != -1 {
		whereIdx = idx
	} else if idx := strings.Index(query, "where"); idx != -1 {
		whereIdx = idx
	}

	if whereIdx != -1 {
		// Has WHERE clause - insert tenant filter at the beginning
		// Query: UPDATE users SET email = ? WHERE name = ?
		// Args: [email, name]
		// Result: UPDATE users SET email = ? WHERE tenant_id = ? AND name = ?
		// New Args: [email, tenantID, name]

		// Count placeholders before WHERE to know where to insert tenantID arg
		beforeWhere := query[:whereIdx]
		setPlaceholders := strings.Count(beforeWhere, "?")

		// Insert tenant arg after SET parameters
		newArgs := make([]interface{}, 0, len(args)+1)
		newArgs = append(newArgs, args[:setPlaceholders]...)  // SET parameters
		newArgs = append(newArgs, tenantID)                   // tenant_id
		newArgs = append(newArgs, args[setPlaceholders:]...)  // WHERE parameters

		filteredQuery := query[:whereIdx] + "WHERE " + whereClause + " AND " + query[whereIdx+5:]
		return filteredQuery, newArgs
	}

	// No WHERE clause - add one
	// Query: UPDATE users SET email = ?
	// Result: UPDATE users SET email = ? WHERE tenant_id = ?
	newArgs := append(args, tenantID)
	filteredQuery := query + " WHERE " + whereClause
	return filteredQuery, newArgs
}

// addTenantToInsert adds tenant_id to INSERT statements.
func (tdb *TenantDB) addTenantToInsert(query string, tenantID string, args []interface{}) (string, []interface{}) {
	// Simple implementation: assumes INSERT INTO table (col1, col2) VALUES (?, ?)
	// In production, you might want a more robust SQL parser

	queryUpper := strings.ToUpper(query)
	valuesIdx := strings.Index(queryUpper, "VALUES")
	if valuesIdx == -1 {
		// Can't parse - return as-is (user must handle tenant_id manually)
		return query, args
	}

	beforeValues := query[:valuesIdx]
	afterValues := query[valuesIdx:]

	// Find column list
	colStart := strings.Index(beforeValues, "(")
	colEnd := strings.LastIndex(beforeValues, ")")
	if colStart == -1 || colEnd == -1 {
		return query, args
	}

	// Add tenant_id to column list
	modifiedBefore := beforeValues[:colEnd] + ", " + tdb.tenantColumn + beforeValues[colEnd:]

	// Add ? to VALUES
	valStart := strings.Index(afterValues, "(")
	valEnd := strings.LastIndex(afterValues, ")")
	if valStart == -1 || valEnd == -1 {
		return query, args
	}

	modifiedAfter := afterValues[:valEnd] + ", ?" + afterValues[valEnd:]

	filteredQuery := modifiedBefore + modifiedAfter
	newArgs := append(args, tenantID)

	return filteredQuery, newArgs
}

// ValidateQuery checks if a query is safe for multi-tenant use.
// It warns about queries that might leak data across tenants.
func ValidateQuery(query string, tenantColumn string) error {
	if tenantColumn == "" {
		tenantColumn = "tenant_id"
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Check for dangerous patterns
	if strings.Contains(queryLower, "drop table") {
		return errors.New("DROP TABLE not allowed in tenant-scoped queries")
	}
	if strings.Contains(queryLower, "truncate") {
		return errors.New("TRUNCATE not allowed in tenant-scoped queries")
	}

	// For SELECT/UPDATE/DELETE, warn if no WHERE clause
	if strings.HasPrefix(queryLower, "select") ||
		strings.HasPrefix(queryLower, "update") ||
		strings.HasPrefix(queryLower, "delete") {

		if !strings.Contains(queryLower, "where") {
			return fmt.Errorf("query missing WHERE clause - potential cross-tenant data leak")
		}

		// Check if tenant column is mentioned
		if !strings.Contains(queryLower, strings.ToLower(tenantColumn)) {
			return fmt.Errorf("query missing %s filter - potential cross-tenant data leak", tenantColumn)
		}
	}

	return nil
}

// RawDB returns the underlying *sql.DB for admin queries that bypass tenant filtering.
// Use with caution - this bypasses all tenant isolation!
//
// Example (admin operation):
//
//	// This query is NOT tenant-filtered
//	rows, err := tdb.RawDB().QueryContext(ctx, "SELECT COUNT(*) FROM users")
func (tdb *TenantDB) RawDB() *sql.DB {
	if tdb == nil {
		return nil
	}
	return tdb.db
}
