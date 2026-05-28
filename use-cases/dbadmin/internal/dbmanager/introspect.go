package dbmanager

import "context"

// TableInfo describes a table or view.
type TableInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`    // TABLE or VIEW
	Comment string `json:"comment,omitempty"`
	Engine  string `json:"engine,omitempty"` // MySQL only
	Rows    int64  `json:"rows,omitempty"`
}

// ColumnInfo describes a column.
type ColumnInfo struct {
	Name       string `json:"name"`
	Position   int    `json:"position"`
	DataType   string `json:"data_type"`
	FullType   string `json:"full_type"`
	Nullable   bool   `json:"nullable"`
	Default    string `json:"default,omitempty"`
	Comment    string `json:"comment,omitempty"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
	AutoIncr   bool   `json:"auto_increment,omitempty"`
}

// IndexInfo describes an index.
type IndexInfo struct {
	Name    string   `json:"name"`
	Unique  bool     `json:"unique"`
	Columns []string `json:"columns"`
	Type    string   `json:"type,omitempty"` // BTREE, HASH, etc.
}

// ForeignKeyInfo describes a foreign key constraint.
type ForeignKeyInfo struct {
	Name       string `json:"name"`
	Column     string `json:"column"`
	RefTable   string `json:"ref_table"`
	RefColumn  string `json:"ref_column"`
	OnDelete   string `json:"on_delete,omitempty"`
	OnUpdate   string `json:"on_update,omitempty"`
}

// Inspector provides database introspection for a specific connection.
type Inspector interface {
	Databases(ctx context.Context) ([]string, error)
	Tables(ctx context.Context, db string) ([]TableInfo, error)
	Columns(ctx context.Context, db, table string) ([]ColumnInfo, error)
	Indexes(ctx context.Context, db, table string) ([]IndexInfo, error)
	ForeignKeys(ctx context.Context, db, table string) ([]ForeignKeyInfo, error)
	PrimaryKeys(ctx context.Context, db, table string) ([]string, error)
	CreateTableDDL(ctx context.Context, db, table string) (string, error)
}
