package datasource

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dbadmin/internal/dbmanager"
	mysqlinspect "dbadmin/internal/dbmanager/mysql"
	sqliteinspect "dbadmin/internal/dbmanager/sqlite"
	"dbadmin/internal/domain/connection"
)

// SQLConfig adapts *connection.Connection to the ConnectionConfig interface,
// bridging the existing connection domain model to DataSourceDriver.
type SQLConfig struct {
	Conn *connection.Connection
}

// DSType implements ConnectionConfig.
func (c SQLConfig) DSType() DataSourceType {
	if c.Conn.Driver == connection.DriverSQLite {
		return TypeSQLite
	}
	return TypeMySQL
}

// SQLAdapter implements DataSourceDriver for MySQL and SQLite by delegating to
// the existing dbmanager.Manager and driver-specific inspectors. No existing
// SQL business logic is rewritten; this is a pure adapter.
type SQLAdapter struct {
	manager *dbmanager.Manager
}

// NewSQLAdapter creates a SQLAdapter backed by the provided Manager.
func NewSQLAdapter(manager *dbmanager.Manager) *SQLAdapter {
	return &SQLAdapter{manager: manager}
}

// Type returns TypeMySQL as the base SQL type. The actual per-session type is
// determined from the connection's driver field and stored in Session.Type.
func (a *SQLAdapter) Type() DataSourceType { return TypeMySQL }

// Test verifies connectivity without caching the pool.
func (a *SQLAdapter) Test(ctx context.Context, cfg ConnectionConfig) error {
	sc, ok := cfg.(SQLConfig)
	if !ok {
		return fmt.Errorf("sql_adapter: expected SQLConfig, got %T", cfg)
	}
	return a.manager.Test(ctx, sc.Conn)
}

// Open opens (or reuses) a *sql.DB pool and wraps it in a Session.
func (a *SQLAdapter) Open(ctx context.Context, cfg ConnectionConfig) (*Session, error) {
	sc, ok := cfg.(SQLConfig)
	if !ok {
		return nil, fmt.Errorf("sql_adapter: expected SQLConfig, got %T", cfg)
	}
	db, err := a.manager.Open(ctx, sc.Conn)
	if err != nil {
		return nil, err
	}
	return &Session{
		ID:        sc.Conn.ID,
		ProfileID: sc.Conn.ID,
		Type:      sc.DSType(),
		Readonly:  sc.Conn.Readonly,
		Handle:    db,
	}, nil
}

// Close closes the session's underlying pool.
func (a *SQLAdapter) Close(_ context.Context, session *Session) error {
	a.manager.Close(session.ProfileID)
	return nil
}

// ListResources returns ResourceNodes for the given parent.
//
//   - parent == nil or parent.Path == "": returns sql_database nodes
//   - parent.Path == "<dbName>":          returns sql_table / sql_view nodes
func (a *SQLAdapter) ListResources(ctx context.Context, session *Session, parent *ResourceRef) ([]ResourceNode, error) {
	db, ok := session.Handle.(*sql.DB)
	if !ok {
		return nil, fmt.Errorf("sql_adapter: session handle is not *sql.DB")
	}
	inspector := a.newInspector(db, session.Type)

	if parent == nil || parent.Path == "" {
		return a.listDatabases(ctx, inspector, session.Type)
	}
	return a.listTables(ctx, inspector, session.Type, parent.Path)
}

func (a *SQLAdapter) listDatabases(ctx context.Context, inspector dbmanager.Inspector, dsType DataSourceType) ([]ResourceNode, error) {
	dbs, err := inspector.Databases(ctx)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	nodes := make([]ResourceNode, 0, len(dbs))
	for _, name := range dbs {
		nodes = append(nodes, ResourceNode{
			ID:             name,
			Name:           name,
			Type:           NodeSQLDatabase,
			DataSourceType: dsType,
			Path:           name,
		})
	}
	return nodes, nil
}

func (a *SQLAdapter) listTables(ctx context.Context, inspector dbmanager.Inspector, dsType DataSourceType, dbName string) ([]ResourceNode, error) {
	// SQLite ignores the database-name argument in table queries.
	queryDB := dbName
	if dsType == TypeSQLite {
		queryDB = ""
	}
	tables, err := inspector.Tables(ctx, queryDB)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	nodes := make([]ResourceNode, 0, len(tables))
	for _, t := range tables {
		nodeType := NodeSQLTable
		if strings.EqualFold(t.Type, "VIEW") {
			nodeType = NodeSQLView
		}
		path := dbName + "/" + t.Name
		meta := map[string]any{"table_type": t.Type}
		if t.Comment != "" {
			meta["comment"] = t.Comment
		}
		if t.Engine != "" {
			meta["engine"] = t.Engine
		}
		if t.Rows > 0 {
			meta["rows"] = t.Rows
		}
		nodes = append(nodes, ResourceNode{
			ID:             path,
			Name:           t.Name,
			Type:           nodeType,
			ParentID:       dbName,
			DataSourceType: dsType,
			Path:           path,
			Meta:           meta,
		})
	}
	return nodes, nil
}

// InspectResource returns basic node identity. Column details are deferred to a future pass.
func (a *SQLAdapter) InspectResource(_ context.Context, session *Session, ref ResourceRef) (*ResourceDetail, error) {
	parts := strings.SplitN(ref.Path, "/", 2)
	name := ref.Path
	if len(parts) == 2 {
		name = parts[1]
	}
	return &ResourceDetail{
		Node: ResourceNode{
			ID:             ref.Path,
			Name:           name,
			DataSourceType: session.Type,
			Path:           ref.Path,
		},
	}, nil
}

func (a *SQLAdapter) newInspector(db *sql.DB, dsType DataSourceType) dbmanager.Inspector {
	if dsType == TypeSQLite {
		return sqliteinspect.New(db)
	}
	return mysqlinspect.New(db)
}
