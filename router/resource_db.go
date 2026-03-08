package router

// This file re-exports database repository helpers from the rest package as type aliases
// so that existing code continues to compile unchanged.
//
// All types here are deprecated for use via this package.
// The canonical location is github.com/spcent/plumego/rest.
//
// Migration: replace import "github.com/spcent/plumego/router" with
//            import "github.com/spcent/plumego/rest" and update type names.

import (
	"github.com/spcent/plumego/rest"
	"github.com/spcent/plumego/store/db"
)

// Repository defines the interface for database CRUD operations on a typed resource T.
//
// Deprecated: use rest.Repository instead.
type Repository[T any] = rest.Repository[T]

// SQLBuilder builds SQL queries from QueryParams.
//
// Deprecated: use rest.SQLBuilder instead.
type SQLBuilder = rest.SQLBuilder

// NewSQLBuilder creates a new SQL query builder for the given table and ID column.
//
// Deprecated: use rest.NewSQLBuilder instead.
func NewSQLBuilder(table string, idColumn string) *SQLBuilder {
	return rest.NewSQLBuilder(table, idColumn)
}

// BaseRepository provides a base implementation of Repository backed by a sql.DB.
//
// Deprecated: use rest.BaseRepository instead.
type BaseRepository[T any] = rest.BaseRepository[T]

// NewBaseRepository creates a new BaseRepository with the given database and SQL builder.
//
// Deprecated: use rest.NewBaseRepository instead.
func NewBaseRepository[T any](database db.DB, builder *SQLBuilder) *BaseRepository[T] {
	return rest.NewBaseRepository[T](database, builder)
}

// DBResourceController combines BaseContextResourceController with a Repository.
//
// Deprecated: use rest.DBResourceController instead.
type DBResourceController[T any] = rest.DBResourceController[T]

// NewDBResourceController creates a new database resource controller.
//
// Deprecated: use rest.NewDBResourceController instead.
func NewDBResourceController[T any](resourceName string, repository Repository[T]) *DBResourceController[T] {
	return rest.NewDBResourceController[T](resourceName, repository)
}

// BaseModel provides common ID and timestamp fields for database models.
//
// Deprecated: use rest.BaseModel instead.
type BaseModel = rest.BaseModel

// SoftDeleteModel extends BaseModel with a nullable soft-delete timestamp.
//
// Deprecated: use rest.SoftDeleteModel instead.
type SoftDeleteModel = rest.SoftDeleteModel
