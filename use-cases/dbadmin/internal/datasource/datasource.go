// Package datasource defines the unified data-source abstraction layer for
// the Data Workbench. Each driver (MySQL, SQLite, and future Redis / MongoDB /
// Elasticsearch) implements DataSourceDriver. Phase 1 covers SQL only.
package datasource

import "context"

// DataSourceType identifies the category of data source.
type DataSourceType string

const (
	TypeMySQL         DataSourceType = "mysql"
	TypeSQLite        DataSourceType = "sqlite"
	TypeRedis         DataSourceType = "redis"         // reserved — not yet implemented
	TypeMongoDB       DataSourceType = "mongodb"       // reserved — not yet implemented
	TypeElasticsearch DataSourceType = "elasticsearch" // reserved — not yet implemented
)

// ResourceNodeType classifies a node in the data-source resource tree.
type ResourceNodeType string

const (
	// SQL node types — available now.
	NodeSQLDatabase ResourceNodeType = "sql_database"
	NodeSQLTable    ResourceNodeType = "sql_table"
	NodeSQLView     ResourceNodeType = "sql_view"

	// Redis node types — config accepted; driver not yet fully implemented.
	NodeRedisDB  ResourceNodeType = "redis_db"
	NodeRedisKey ResourceNodeType = "redis_key"

	// MongoDB node types — driver not yet fully implemented, but config accepted.
	NodeMongoDatabase   ResourceNodeType = "mongo_database"
	NodeMongoCollection ResourceNodeType = "mongo_collection"

	// Elasticsearch node types — config accepted; driver not yet fully implemented.
	NodeESIndex      ResourceNodeType = "es_index"
	NodeESAlias      ResourceNodeType = "es_alias"
	NodeESDataStream ResourceNodeType = "es_data_stream"
)

// ResourceNode is a single node in the data-source resource tree.
//
// Tree structure per driver:
//   - SQL:           connection → sql_database → sql_table | sql_view
//   - Redis:         connection → redis_db → redis_key          (planned)
//   - MongoDB:       connection → mongo_database → mongo_collection (planned)
//   - Elasticsearch: connection → es_index | es_alias | es_data_stream (planned)
type ResourceNode struct {
	ID             string           `json:"id"`
	Name           string           `json:"name"`
	Type           ResourceNodeType `json:"type"`
	ParentID       string           `json:"parent_id,omitempty"`
	DataSourceType DataSourceType   `json:"datasource_type"`
	// Path is the slash-separated identifier within the connection.
	// SQL examples: "mydb" (database node) or "mydb/users" (table node).
	Path string         `json:"path"`
	Meta map[string]any `json:"meta,omitempty"`
}

// ResourceRef identifies a resource within a connection. Used in driver calls.
type ResourceRef struct {
	ConnectionID string
	// Path matches ResourceNode.Path; empty means the connection root.
	Path string
}

// ResourceDetail holds full metadata for a resource node.
// Columns are populated for table-level nodes.
type ResourceDetail struct {
	Node    ResourceNode `json:"node"`
	Columns []ColumnMeta `json:"columns,omitempty"`
}

// ColumnMeta is a minimal, driver-agnostic column descriptor.
type ColumnMeta struct {
	Name     string `json:"name"`
	DataType string `json:"data_type"`
	Nullable bool   `json:"nullable"`
}

// Session wraps a live driver-specific connection handle.
//
// Handle type per driver:
//   - SQL:           *sql.DB
//   - Redis:         (planned)
//   - MongoDB:       (planned)
//   - Elasticsearch: (planned)
type Session struct {
	ID        string
	ProfileID string // connection ID from connection.Store
	Type      DataSourceType
	Readonly  bool
	Handle    any // cast based on Type
}

// ConnectionConfig is the base interface for driver-specific configuration.
// Use the concrete types: SQLConfig (Phase 1), and future RedisConfig, etc.
type ConnectionConfig interface {
	DSType() DataSourceType
}

// DataSourceDriver is the extension point for adding new data source types.
//
// To add Redis (future):
//  1. Create internal/datasource/redis_adapter.go implementing this interface.
//  2. Add DriverRedis to connection.DriverType constants.
//  3. Register the adapter in app.go (ResourceHandler.drivers map).
//  4. Add a route group /api/conn/:id/redis/... for Redis-specific operations.
type DataSourceDriver interface {
	// Type returns the datasource type this driver handles.
	Type() DataSourceType
	// Test verifies connectivity without caching the connection.
	Test(ctx context.Context, cfg ConnectionConfig) error
	// Open returns a live Session. For SQL, this reuses the Manager's pool.
	Open(ctx context.Context, cfg ConnectionConfig) (*Session, error)
	// Close releases the session's underlying connection.
	Close(ctx context.Context, session *Session) error
	// ListResources returns child nodes for the given parent.
	// A nil parent returns top-level nodes (databases for SQL, key-spaces for Redis, etc.).
	ListResources(ctx context.Context, session *Session, parent *ResourceRef) ([]ResourceNode, error)
	// InspectResource returns detailed metadata for a specific node.
	InspectResource(ctx context.Context, session *Session, ref ResourceRef) (*ResourceDetail, error)
}
