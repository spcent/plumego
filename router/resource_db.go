package router

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/store/db"
)

// ================================================
// Database Repository Pattern
// ================================================

// Repository defines the interface for database CRUD operations
type Repository[T any] interface {
	// FindAll retrieves all records with optional query parameters
	FindAll(ctx context.Context, params *QueryParams) ([]T, int64, error)

	// FindByID retrieves a single record by ID
	FindByID(ctx context.Context, id string) (*T, error)

	// Create inserts a new record
	Create(ctx context.Context, data *T) error

	// Update updates an existing record
	Update(ctx context.Context, id string, data *T) error

	// Delete removes a record by ID
	Delete(ctx context.Context, id string) error

	// Count returns the total number of records matching the query
	Count(ctx context.Context, params *QueryParams) (int64, error)

	// Exists checks if a record exists by ID
	Exists(ctx context.Context, id string) (bool, error)
}

// ================================================
// SQL Query Builder
// ================================================

// SQLBuilder builds SQL queries from QueryParams
type SQLBuilder struct {
	table      string
	columns    []string
	idColumn   string
	scanFunc   func(*sql.Rows) (any, error)
	insertFunc func(any) ([]any, error)
	updateFunc func(any) ([]any, error)
}

// NewSQLBuilder creates a new SQL query builder
func NewSQLBuilder(table string, idColumn string) *SQLBuilder {
	return &SQLBuilder{
		table:    table,
		idColumn: idColumn,
		columns:  []string{},
	}
}

// WithColumns sets the columns to select
func (b *SQLBuilder) WithColumns(columns ...string) *SQLBuilder {
	b.columns = columns
	return b
}

// WithScanFunc sets the function to scan rows into a struct
func (b *SQLBuilder) WithScanFunc(fn func(*sql.Rows) (any, error)) *SQLBuilder {
	b.scanFunc = fn
	return b
}

// WithInsertFunc sets the function to extract insert values from a struct
func (b *SQLBuilder) WithInsertFunc(fn func(any) ([]any, error)) *SQLBuilder {
	b.insertFunc = fn
	return b
}

// WithUpdateFunc sets the function to extract update values from a struct
func (b *SQLBuilder) WithUpdateFunc(fn func(any) ([]any, error)) *SQLBuilder {
	b.updateFunc = fn
	return b
}

// BuildSelectQuery builds a SELECT query with pagination, sorting, and filtering
func (b *SQLBuilder) BuildSelectQuery(params *QueryParams) (string, []any) {
	var query strings.Builder
	var args []any

	// SELECT clause
	query.WriteString("SELECT ")
	if len(b.columns) > 0 {
		query.WriteString(strings.Join(b.columns, ", "))
	} else {
		query.WriteString("*")
	}
	query.WriteString(" FROM ")
	query.WriteString(b.table)

	// WHERE clause
	if len(params.Filters) > 0 || params.Search != "" {
		query.WriteString(" WHERE ")
		conditions := []string{}

		// Add search condition
		if params.Search != "" && len(b.columns) > 0 {
			searchConditions := []string{}
			for _, col := range b.columns {
				searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE ?", col))
				args = append(args, "%"+params.Search+"%")
			}
			conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		}

		// Add filter conditions
		for key, value := range params.Filters {
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditions, " AND "))
	}

	// ORDER BY clause
	if len(params.Sort) > 0 {
		query.WriteString(" ORDER BY ")
		orders := []string{}
		for _, sort := range params.Sort {
			direction := "ASC"
			if sort.Desc {
				direction = "DESC"
			}
			orders = append(orders, fmt.Sprintf("%s %s", sort.Field, direction))
		}
		query.WriteString(strings.Join(orders, ", "))
	}

	// LIMIT and OFFSET
	if params.Limit > 0 {
		query.WriteString(" LIMIT ?")
		args = append(args, params.Limit)
	}
	if params.Offset > 0 {
		query.WriteString(" OFFSET ?")
		args = append(args, params.Offset)
	}

	return query.String(), args
}

// BuildCountQuery builds a COUNT query with filtering
func (b *SQLBuilder) BuildCountQuery(params *QueryParams) (string, []any) {
	var query strings.Builder
	var args []any

	query.WriteString("SELECT COUNT(*) FROM ")
	query.WriteString(b.table)

	// WHERE clause
	if len(params.Filters) > 0 || params.Search != "" {
		query.WriteString(" WHERE ")
		conditions := []string{}

		// Add search condition
		if params.Search != "" && len(b.columns) > 0 {
			searchConditions := []string{}
			for _, col := range b.columns {
				searchConditions = append(searchConditions, fmt.Sprintf("%s LIKE ?", col))
				args = append(args, "%"+params.Search+"%")
			}
			conditions = append(conditions, "("+strings.Join(searchConditions, " OR ")+")")
		}

		// Add filter conditions
		for key, value := range params.Filters {
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditions, " AND "))
	}

	return query.String(), args
}

// BuildInsertQuery builds an INSERT query
func (b *SQLBuilder) BuildInsertQuery() string {
	placeholders := make([]string, len(b.columns))
	for i := range b.columns {
		placeholders[i] = "?"
	}

	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		b.table,
		strings.Join(b.columns, ", "),
		strings.Join(placeholders, ", "),
	)
}

// BuildUpdateQuery builds an UPDATE query
func (b *SQLBuilder) BuildUpdateQuery() string {
	setClauses := make([]string, len(b.columns))
	for i, col := range b.columns {
		setClauses[i] = fmt.Sprintf("%s = ?", col)
	}

	return fmt.Sprintf(
		"UPDATE %s SET %s WHERE %s = ?",
		b.table,
		strings.Join(setClauses, ", "),
		b.idColumn,
	)
}

// BuildDeleteQuery builds a DELETE query
func (b *SQLBuilder) BuildDeleteQuery() string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = ?", b.table, b.idColumn)
}

// BuildFindByIDQuery builds a SELECT query for a single record by ID
func (b *SQLBuilder) BuildFindByIDQuery() string {
	selectCols := "*"
	if len(b.columns) > 0 {
		selectCols = strings.Join(b.columns, ", ")
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", selectCols, b.table, b.idColumn)
}

// BuildExistsQuery builds an EXISTS query
func (b *SQLBuilder) BuildExistsQuery() string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", b.table, b.idColumn)
}

// ================================================
// Base Repository Implementation
// ================================================

// BaseRepository provides a base implementation for Repository interface
type BaseRepository[T any] struct {
	db      db.DB
	builder *SQLBuilder
}

// NewBaseRepository creates a new base repository
func NewBaseRepository[T any](database db.DB, builder *SQLBuilder) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:      database,
		builder: builder,
	}
}

// FindAll retrieves all records with pagination, filtering, and sorting
func (r *BaseRepository[T]) FindAll(ctx context.Context, params *QueryParams) ([]T, int64, error) {
	// Get total count first
	total, err := r.Count(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	// Build and execute query
	query, args := r.builder.BuildSelectQuery(params)
	rows, err := db.QueryContext(ctx, r.db, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	// Scan results
	var results []T
	for rows.Next() {
		if r.builder.scanFunc == nil {
			return nil, 0, fmt.Errorf("scan function not configured")
		}

		item, err := r.builder.scanFunc(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}

		// Type assertion
		if typedItem, ok := item.(T); ok {
			results = append(results, typedItem)
		} else if typedItem, ok := item.(*T); ok {
			results = append(results, *typedItem)
		} else {
			return nil, 0, fmt.Errorf("scan function returned wrong type")
		}
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return results, total, nil
}

// FindByID retrieves a single record by ID
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	query := r.builder.BuildFindByIDQuery()

	rows, err := db.QueryContext(ctx, r.db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query record: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("failed to query record: %w", err)
		}
		return nil, sql.ErrNoRows
	}

	if r.builder.scanFunc == nil {
		return nil, fmt.Errorf("scan function not configured")
	}

	item, err := r.builder.scanFunc(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	// Type assertion
	if typedItem, ok := item.(T); ok {
		return &typedItem, nil
	} else if typedItem, ok := item.(*T); ok {
		return typedItem, nil
	}

	return nil, fmt.Errorf("scan function returned wrong type")
}

// Create inserts a new record
func (r *BaseRepository[T]) Create(ctx context.Context, data *T) error {
	if r.builder.insertFunc == nil {
		return fmt.Errorf("insert function not configured")
	}

	values, err := r.builder.insertFunc(data)
	if err != nil {
		return fmt.Errorf("failed to extract insert values: %w", err)
	}

	query := r.builder.BuildInsertQuery()
	_, err = db.ExecContext(ctx, r.db, query, values...)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}

	return nil
}

// Update updates an existing record
func (r *BaseRepository[T]) Update(ctx context.Context, id string, data *T) error {
	if r.builder.updateFunc == nil {
		return fmt.Errorf("update function not configured")
	}

	values, err := r.builder.updateFunc(data)
	if err != nil {
		return fmt.Errorf("failed to extract update values: %w", err)
	}

	// Append ID for WHERE clause
	values = append(values, id)

	query := r.builder.BuildUpdateQuery()
	result, err := db.ExecContext(ctx, r.db, query, values...)
	if err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Delete removes a record by ID
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	query := r.builder.BuildDeleteQuery()
	result, err := db.ExecContext(ctx, r.db, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// Count returns the total number of records matching the query
func (r *BaseRepository[T]) Count(ctx context.Context, params *QueryParams) (int64, error) {
	query, args := r.builder.BuildCountQuery(params)
	row := db.QueryRowContext(ctx, r.db, query, args...)

	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}

	return count, nil
}

// Exists checks if a record exists by ID
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	query := r.builder.BuildExistsQuery()
	row := db.QueryRowContext(ctx, r.db, query, id)

	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists, nil
}

// ================================================
// Database Resource Controller
// ================================================

// DBResourceController combines BaseContextResourceController with Repository
type DBResourceController[T any] struct {
	*BaseContextResourceController
	repository Repository[T]
	validator  interface {
		Validate(any) error
	}
}

// NewDBResourceController creates a new database resource controller
func NewDBResourceController[T any](
	resourceName string,
	repository Repository[T],
) *DBResourceController[T] {
	return &DBResourceController[T]{
		BaseContextResourceController: NewBaseContextResourceController(resourceName),
		repository:                    repository,
	}
}

// WithValidator sets the validator for the controller
func (c *DBResourceController[T]) WithValidator(validator interface{ Validate(any) error }) *DBResourceController[T] {
	c.validator = validator
	return c
}

// IndexCtx handles GET /resource requests with database query
func (c *DBResourceController[T]) IndexCtx(ctx *contract.Ctx) {
	// Parse query parameters
	params := c.QueryBuilder.Parse(ctx.R)

	// Call before hook
	if err := c.Hooks.BeforeList(ctx.R.Context(), params); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Fetch from database
	results, total, err := c.repository.FindAll(ctx.R.Context(), params)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch records"})
		return
	}

	// Transform results
	transformedResults, err := c.Transformer.TransformCollection(ctx.R.Context(), results)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to transform results"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterList(ctx.R.Context(), params, transformedResults); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Create pagination metadata
	pagination := NewPaginationMeta(params.Page, params.PageSize, total)

	// Return paginated response
	ctx.JSON(http.StatusOK, PaginatedResponse{
		Data:       transformedResults,
		Pagination: pagination,
	})
}

// ShowCtx handles GET /resource/:id requests with database query
func (c *DBResourceController[T]) ShowCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	// Fetch from database
	result, err := c.repository.FindByID(ctx.R.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, map[string]string{"error": "Record not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch record"})
		return
	}

	// Transform result
	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), result)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to transform result"})
		return
	}

	ctx.JSON(http.StatusOK, transformedResult)
}

// CreateCtx handles POST /resource requests with database insert
func (c *DBResourceController[T]) CreateCtx(ctx *contract.Ctx) {
	var data T

	// Bind request
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate
	if c.validator != nil {
		if err := c.validator.Validate(data); err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "Validation failed", "details": err})
			return
		}
	}

	// Call before hook
	if err := c.Hooks.BeforeCreate(ctx.R.Context(), &data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Insert into database
	if err := c.repository.Create(ctx.R.Context(), &data); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create record"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterCreate(ctx.R.Context(), &data); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Transform result
	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), &data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to transform result"})
		return
	}

	ctx.JSON(http.StatusCreated, transformedResult)
}

// UpdateCtx handles PUT /resource/:id requests with database update
func (c *DBResourceController[T]) UpdateCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	var data T
	if err := ctx.BindJSON(&data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate
	if c.validator != nil {
		if err := c.validator.Validate(data); err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "Validation failed", "details": err})
			return
		}
	}

	// Call before hook
	if err := c.Hooks.BeforeUpdate(ctx.R.Context(), id, &data); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Update in database
	if err := c.repository.Update(ctx.R.Context(), id, &data); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, map[string]string{"error": "Record not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update record"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterUpdate(ctx.R.Context(), id, &data); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Transform result
	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), &data)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to transform result"})
		return
	}

	ctx.JSON(http.StatusOK, transformedResult)
}

// DeleteCtx handles DELETE /resource/:id requests with database delete
func (c *DBResourceController[T]) DeleteCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": "ID is required"})
		return
	}

	// Call before hook
	if err := c.Hooks.BeforeDelete(ctx.R.Context(), id); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Delete from database
	if err := c.repository.Delete(ctx.R.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, map[string]string{"error": "Record not found"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete record"})
		return
	}

	// Call after hook
	if err := c.Hooks.AfterDelete(ctx.R.Context(), id); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// ================================================
// Common Model Fields
// ================================================

// BaseModel provides common fields for database models
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SoftDeleteModel extends BaseModel with soft delete support
type SoftDeleteModel struct {
	BaseModel
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
