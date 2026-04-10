package rest

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
// Repository interface
// ================================================

// Repository defines the interface for database CRUD operations on a typed resource T.
type Repository[T any] interface {
	FindAll(ctx context.Context, params *QueryParams) ([]T, int64, error)
	FindByID(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, data *T) error
	Update(ctx context.Context, id string, data *T) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context, params *QueryParams) (int64, error)
	Exists(ctx context.Context, id string) (bool, error)
}

// ================================================
// SQLBuilder
// ================================================

// SQLBuilder builds SQL queries from QueryParams.
type SQLBuilder struct {
	table      string
	columns    []string
	idColumn   string
	scanFunc   func(*sql.Rows) (any, error)
	insertFunc func(any) ([]any, error)
	updateFunc func(any) ([]any, error)
}

// NewSQLBuilder creates a new SQL query builder for the given table and ID column.
func NewSQLBuilder(table string, idColumn string) *SQLBuilder {
	return &SQLBuilder{
		table:    table,
		idColumn: idColumn,
		columns:  []string{},
	}
}

// WithColumns sets the columns to select.
func (b *SQLBuilder) WithColumns(columns ...string) *SQLBuilder {
	b.columns = columns
	return b
}

// WithScanFunc sets the function to scan rows into a struct.
func (b *SQLBuilder) WithScanFunc(fn func(*sql.Rows) (any, error)) *SQLBuilder {
	b.scanFunc = fn
	return b
}

// WithInsertFunc sets the function to extract insert values from a struct.
func (b *SQLBuilder) WithInsertFunc(fn func(any) ([]any, error)) *SQLBuilder {
	b.insertFunc = fn
	return b
}

// WithUpdateFunc sets the function to extract update values from a struct.
func (b *SQLBuilder) WithUpdateFunc(fn func(any) ([]any, error)) *SQLBuilder {
	b.updateFunc = fn
	return b
}

// BuildSelectQuery builds a SELECT query with pagination, sorting, and filtering.
func (b *SQLBuilder) BuildSelectQuery(params *QueryParams) (string, []any) {
	var query strings.Builder
	var args []any

	query.WriteString("SELECT ")
	if len(b.columns) > 0 {
		query.WriteString(strings.Join(b.columns, ", "))
	} else {
		query.WriteString("*")
	}
	query.WriteString(" FROM ")
	query.WriteString(b.table)

	if len(params.Filters) > 0 || params.Search != "" {
		query.WriteString(" WHERE ")
		var conditions []string

		if params.Search != "" && len(b.columns) > 0 {
			var searchConds []string
			for _, col := range b.columns {
				searchConds = append(searchConds, fmt.Sprintf("%s LIKE ?", col))
				args = append(args, "%"+params.Search+"%")
			}
			conditions = append(conditions, "("+strings.Join(searchConds, " OR ")+")")
		}

		for key, value := range params.Filters {
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditions, " AND "))
	}

	if len(params.Sort) > 0 {
		query.WriteString(" ORDER BY ")
		var orders []string
		for _, sort := range params.Sort {
			direction := "ASC"
			if sort.Desc {
				direction = "DESC"
			}
			orders = append(orders, fmt.Sprintf("%s %s", sort.Field, direction))
		}
		query.WriteString(strings.Join(orders, ", "))
	}

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

// BuildCountQuery builds a COUNT query with filtering.
func (b *SQLBuilder) BuildCountQuery(params *QueryParams) (string, []any) {
	var query strings.Builder
	var args []any

	query.WriteString("SELECT COUNT(*) FROM ")
	query.WriteString(b.table)

	if len(params.Filters) > 0 || params.Search != "" {
		query.WriteString(" WHERE ")
		var conditions []string

		if params.Search != "" && len(b.columns) > 0 {
			var searchConds []string
			for _, col := range b.columns {
				searchConds = append(searchConds, fmt.Sprintf("%s LIKE ?", col))
				args = append(args, "%"+params.Search+"%")
			}
			conditions = append(conditions, "("+strings.Join(searchConds, " OR ")+")")
		}

		for key, value := range params.Filters {
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}

		query.WriteString(strings.Join(conditions, " AND "))
	}

	return query.String(), args
}

// BuildInsertQuery builds an INSERT query for the configured columns.
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

// BuildUpdateQuery builds an UPDATE query for all configured columns.
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

// BuildDeleteQuery builds a DELETE query.
func (b *SQLBuilder) BuildDeleteQuery() string {
	return fmt.Sprintf("DELETE FROM %s WHERE %s = ?", b.table, b.idColumn)
}

// BuildFindByIDQuery builds a SELECT query for a single record by ID.
func (b *SQLBuilder) BuildFindByIDQuery() string {
	selectCols := "*"
	if len(b.columns) > 0 {
		selectCols = strings.Join(b.columns, ", ")
	}
	return fmt.Sprintf("SELECT %s FROM %s WHERE %s = ?", selectCols, b.table, b.idColumn)
}

// BuildExistsQuery builds an EXISTS query.
func (b *SQLBuilder) BuildExistsQuery() string {
	return fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE %s = ?)", b.table, b.idColumn)
}

// ================================================
// BaseRepository
// ================================================

// BaseRepository provides a base implementation of Repository backed by a sql.DB.
type BaseRepository[T any] struct {
	db      db.DB
	builder *SQLBuilder
}

// NewBaseRepository creates a new BaseRepository with the given database and SQL builder.
func NewBaseRepository[T any](database db.DB, builder *SQLBuilder) *BaseRepository[T] {
	return &BaseRepository[T]{db: database, builder: builder}
}

// DB returns the underlying database connection. Use this when you need to run
// custom queries beyond what BaseRepository provides.
func (r *BaseRepository[T]) DB() db.DB { return r.db }

// FindAll retrieves all records with pagination, filtering, and sorting.
func (r *BaseRepository[T]) FindAll(ctx context.Context, params *QueryParams) ([]T, int64, error) {
	total, err := r.Count(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	query, args := r.builder.BuildSelectQuery(params)
	rows, err := db.QueryContext(ctx, r.db, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query records: %w", err)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		if r.builder.scanFunc == nil {
			return nil, 0, fmt.Errorf("scan function not configured")
		}
		item, err := r.builder.scanFunc(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan row: %w", err)
		}
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

// FindByID retrieves a single record by ID.
func (r *BaseRepository[T]) FindByID(ctx context.Context, id string) (*T, error) {
	rows, err := db.QueryContext(ctx, r.db, r.builder.BuildFindByIDQuery(), id)
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
	if typedItem, ok := item.(T); ok {
		return &typedItem, nil
	} else if typedItem, ok := item.(*T); ok {
		return typedItem, nil
	}
	return nil, fmt.Errorf("scan function returned wrong type")
}

// Create inserts a new record.
func (r *BaseRepository[T]) Create(ctx context.Context, data *T) error {
	if r.builder.insertFunc == nil {
		return fmt.Errorf("insert function not configured")
	}
	values, err := r.builder.insertFunc(data)
	if err != nil {
		return fmt.Errorf("failed to extract insert values: %w", err)
	}
	_, err = db.ExecContext(ctx, r.db, r.builder.BuildInsertQuery(), values...)
	if err != nil {
		return fmt.Errorf("failed to insert record: %w", err)
	}
	return nil
}

// Update updates an existing record.
func (r *BaseRepository[T]) Update(ctx context.Context, id string, data *T) error {
	if r.builder.updateFunc == nil {
		return fmt.Errorf("update function not configured")
	}
	values, err := r.builder.updateFunc(data)
	if err != nil {
		return fmt.Errorf("failed to extract update values: %w", err)
	}
	values = append(values, id)
	result, err := db.ExecContext(ctx, r.db, r.builder.BuildUpdateQuery(), values...)
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

// Delete removes a record by ID.
func (r *BaseRepository[T]) Delete(ctx context.Context, id string) error {
	result, err := db.ExecContext(ctx, r.db, r.builder.BuildDeleteQuery(), id)
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

// Count returns the number of records matching the query.
func (r *BaseRepository[T]) Count(ctx context.Context, params *QueryParams) (int64, error) {
	query, args := r.builder.BuildCountQuery(params)
	row := db.QueryRowContext(ctx, r.db, query, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count records: %w", err)
	}
	return count, nil
}

// Exists reports whether a record with the given ID exists.
func (r *BaseRepository[T]) Exists(ctx context.Context, id string) (bool, error) {
	row := db.QueryRowContext(ctx, r.db, r.builder.BuildExistsQuery(), id)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return exists, nil
}

// ================================================
// DBResourceController
// ================================================

// DBResourceController combines BaseContextResourceController with a Repository,
// providing database-backed CRUD handlers out of the box.
type DBResourceController[T any] struct {
	*BaseContextResourceController
	repository Repository[T]
	validator  interface {
		Validate(any) error
	}
}

// NewDBResourceController creates a new database resource controller.
func NewDBResourceController[T any](resourceName string, repository Repository[T]) *DBResourceController[T] {
	return &DBResourceController[T]{
		BaseContextResourceController: NewBaseContextResourceController(resourceName),
		repository:                    repository,
	}
}

// WithValidator sets the validator used by Create and Update handlers.
func (c *DBResourceController[T]) WithValidator(v interface{ Validate(any) error }) *DBResourceController[T] {
	c.validator = v
	return c
}

// ApplySpec applies the reusable resource specification to the controller.
func (c *DBResourceController[T]) ApplySpec(spec ResourceSpec) *DBResourceController[T] {
	if c == nil {
		return c
	}
	c.BaseContextResourceController.ApplySpec(spec)
	return c
}

// IndexCtx handles GET /resource with database query, pagination, and transformation.
func (c *DBResourceController[T]) IndexCtx(ctx *contract.Ctx) {
	params := c.ParseQueryParams(ctx.R)

	if err := c.Hooks.BeforeList(ctx.R.Context(), params); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code("invalid_list_request").
			Message(err.Error()).
			Build())
		return
	}

	results, total, err := c.repository.FindAll(ctx.R.Context(), params)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("list_failed").
			Message("failed to fetch records").
			Build())
		return
	}

	transformedResults, err := c.Transformer.TransformCollection(ctx.R.Context(), results)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("transform_collection_failed").
			Message("failed to transform results").
			Build())
		return
	}

	if err := c.Hooks.AfterList(ctx.R.Context(), params, transformedResults); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("after_list_failed").
			Message(err.Error()).
			Build())
		return
	}

	_ = ctx.Response(http.StatusOK, PaginatedResponse{
		Data:       transformedResults,
		Pagination: NewPaginationMeta(params.Page, params.PageSize, total),
	}, nil)
}

// ShowCtx handles GET /resource/:id.
func (c *DBResourceController[T]) ShowCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code("missing_id").
			Message("id is required").
			Build())
		return
	}

	result, err := c.repository.FindByID(ctx.R.Context(), id)
	if err != nil {
		if err == sql.ErrNoRows {
			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).
				Code("not_found").
				Message("record not found").
				Build())
			return
		}
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("show_failed").
			Message("failed to fetch record").
			Build())
		return
	}

	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), result)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("transform_failed").
			Message("failed to transform result").
			Build())
		return
	}
	_ = ctx.Response(http.StatusOK, transformedResult, nil)
}

// CreateCtx handles POST /resource.
func (c *DBResourceController[T]) CreateCtx(ctx *contract.Ctx) {
	var data T
	if err := ctx.BindJSON(&data, nil); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code("invalid_request").
			Message("invalid request body").
			Build())
		return
	}

	if c.validator != nil {
		if err := c.validator.Validate(data); err != nil {
			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Status(http.StatusUnprocessableEntity).
				Category(contract.CategoryValidation).
				Code("validation_failed").
				Message("validation failed").
				Detail("details", err).
				Build())
			return
		}
	}

	if err := c.Hooks.BeforeCreate(ctx.R.Context(), &data); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code("before_create_failed").
			Message(err.Error()).
			Build())
		return
	}

	if err := c.repository.Create(ctx.R.Context(), &data); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("create_failed").
			Message("failed to create record").
			Build())
		return
	}

	if err := c.Hooks.AfterCreate(ctx.R.Context(), &data); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("after_create_failed").
			Message(err.Error()).
			Build())
		return
	}

	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), &data)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("transform_failed").
			Message("failed to transform result").
			Build())
		return
	}
	_ = ctx.Response(http.StatusCreated, transformedResult, nil)
}

// UpdateCtx handles PUT /resource/:id.
func (c *DBResourceController[T]) UpdateCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code("missing_id").
			Message("id is required").
			Build())
		return
	}

	var data T
	if err := ctx.BindJSON(&data, nil); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeValidation).
			Code("invalid_request").
			Message("invalid request body").
			Build())
		return
	}

	if c.validator != nil {
		if err := c.validator.Validate(data); err != nil {
			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Status(http.StatusUnprocessableEntity).
				Category(contract.CategoryValidation).
				Code("validation_failed").
				Message("validation failed").
				Detail("details", err).
				Build())
			return
		}
	}

	if err := c.Hooks.BeforeUpdate(ctx.R.Context(), id, &data); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code("before_update_failed").
			Message(err.Error()).
			Build())
		return
	}

	if err := c.repository.Update(ctx.R.Context(), id, &data); err != nil {
		if err == sql.ErrNoRows {
			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).
				Code("not_found").
				Message("record not found").
				Build())
			return
		}
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("update_failed").
			Message("failed to update record").
			Build())
		return
	}

	if err := c.Hooks.AfterUpdate(ctx.R.Context(), id, &data); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("after_update_failed").
			Message(err.Error()).
			Build())
		return
	}

	transformedResult, err := c.Transformer.Transform(ctx.R.Context(), &data)
	if err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("transform_failed").
			Message("failed to transform result").
			Build())
		return
	}
	_ = ctx.Response(http.StatusOK, transformedResult, nil)
}

// DeleteCtx handles DELETE /resource/:id.
func (c *DBResourceController[T]) DeleteCtx(ctx *contract.Ctx) {
	id := c.ParamExtractor.GetID(ctx.R)
	if id == "" {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Code("missing_id").
			Message("id is required").
			Build())
		return
	}

	if err := c.Hooks.BeforeDelete(ctx.R.Context(), id); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Status(http.StatusBadRequest).
			Code("before_delete_failed").
			Message(err.Error()).
			Build())
		return
	}

	if err := c.repository.Delete(ctx.R.Context(), id); err != nil {
		if err == sql.ErrNoRows {
			_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
				Type(contract.TypeNotFound).
				Code("not_found").
				Message("record not found").
				Build())
			return
		}
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("delete_failed").
			Message("failed to delete record").
			Build())
		return
	}

	if err := c.Hooks.AfterDelete(ctx.R.Context(), id); err != nil {
		_ = contract.WriteError(ctx.W, ctx.R, contract.NewErrorBuilder().
			Type(contract.TypeInternal).
			Code("after_delete_failed").
			Message(err.Error()).
			Build())
		return
	}

	_ = ctx.Response(http.StatusNoContent, nil, nil)
}

// ================================================
// Common model field sets
// ================================================

// BaseModel provides common ID and timestamp fields for database models.
type BaseModel struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// SoftDeleteModel extends BaseModel with a nullable soft-delete timestamp.
type SoftDeleteModel struct {
	BaseModel
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}
