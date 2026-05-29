// Package sqlite provides SQLite introspection via PRAGMA statements.
package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dbadmin/internal/dbmanager"
)

// Inspector implements dbmanager.Inspector for SQLite.
type Inspector struct {
	db *sql.DB
}

// New creates a SQLite inspector.
func New(db *sql.DB) *Inspector { return &Inspector{db: db} }

// Databases returns a single entry for SQLite (it has no multi-database concept).
func (ins *Inspector) Databases(ctx context.Context) ([]string, error) {
	rows, err := ins.db.QueryContext(ctx, `PRAGMA database_list`)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	defer rows.Close()
	dbs := make([]string, 0)
	for rows.Next() {
		var seq int
		var name, file string
		if err := rows.Scan(&seq, &name, &file); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

func (ins *Inspector) Tables(ctx context.Context, _ string) ([]dbmanager.TableInfo, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT name, type FROM sqlite_master
		 WHERE type IN ('table','view') AND name NOT LIKE 'sqlite_%'
		 ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	tables := make([]dbmanager.TableInfo, 0)
	for rows.Next() {
		var t dbmanager.TableInfo
		var typ string
		if err := rows.Scan(&t.Name, &typ); err != nil {
			return nil, err
		}
		t.Type = strings.ToUpper(typ)
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

// pragmaTableInfo is a row from PRAGMA table_info.
type pragmaTableInfo struct {
	Cid     int
	Name    string
	Type    string
	NotNull int
	Default sql.NullString
	Pk      int
}

func (ins *Inspector) Columns(ctx context.Context, _, table string) ([]dbmanager.ColumnInfo, error) {
	rows, err := ins.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info("%s")`, table))
	if err != nil {
		return nil, fmt.Errorf("table_info: %w", err)
	}
	defer rows.Close()
	cols := make([]dbmanager.ColumnInfo, 0)
	for rows.Next() {
		var p pragmaTableInfo
		if err := rows.Scan(&p.Cid, &p.Name, &p.Type, &p.NotNull, &p.Default, &p.Pk); err != nil {
			return nil, err
		}
		c := dbmanager.ColumnInfo{
			Name:       p.Name,
			Position:   p.Cid + 1,
			DataType:   p.Type,
			FullType:   p.Type,
			Nullable:   p.NotNull == 0,
			PrimaryKey: p.Pk > 0,
		}
		if p.Default.Valid {
			c.Default = p.Default.String
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (ins *Inspector) Indexes(ctx context.Context, _, table string) ([]dbmanager.IndexInfo, error) {
	rows, err := ins.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_list("%s")`, table))
	if err != nil {
		return nil, fmt.Errorf("index_list: %w", err)
	}
	defer rows.Close()
	type indexRow struct {
		Seq     int
		Name    string
		Unique  int
		Origin  string
		Partial int
	}
	var indexRows []indexRow
	for rows.Next() {
		var ir indexRow
		if err := rows.Scan(&ir.Seq, &ir.Name, &ir.Unique, &ir.Origin, &ir.Partial); err != nil {
			return nil, err
		}
		indexRows = append(indexRows, ir)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	indexes := make([]dbmanager.IndexInfo, 0)
	for _, ir := range indexRows {
		cols, err := ins.indexColumns(ctx, ir.Name)
		if err != nil {
			return nil, err
		}
		indexes = append(indexes, dbmanager.IndexInfo{
			Name:    ir.Name,
			Unique:  ir.Unique == 1,
			Columns: cols,
		})
	}
	return indexes, nil
}

func (ins *Inspector) indexColumns(ctx context.Context, indexName string) ([]string, error) {
	rows, err := ins.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA index_info("%s")`, indexName))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols := make([]string, 0)
	for rows.Next() {
		var seqno, cid int
		var name string
		if err := rows.Scan(&seqno, &cid, &name); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func (ins *Inspector) ForeignKeys(ctx context.Context, _, table string) ([]dbmanager.ForeignKeyInfo, error) {
	rows, err := ins.db.QueryContext(ctx, fmt.Sprintf(`PRAGMA foreign_key_list("%s")`, table))
	if err != nil {
		return nil, fmt.Errorf("foreign_key_list: %w", err)
	}
	defer rows.Close()
	fks := make([]dbmanager.ForeignKeyInfo, 0)
	for rows.Next() {
		var id, seq int
		var refTable, from, to, onUpdate, onDelete, match string
		if err := rows.Scan(&id, &seq, &refTable, &from, &to, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		fks = append(fks, dbmanager.ForeignKeyInfo{
			Name:      fmt.Sprintf("fk_%d", id),
			Column:    from,
			RefTable:  refTable,
			RefColumn: to,
			OnDelete:  onDelete,
			OnUpdate:  onUpdate,
		})
	}
	return fks, rows.Err()
}

func (ins *Inspector) PrimaryKeys(ctx context.Context, _, table string) ([]string, error) {
	cols, err := ins.Columns(ctx, "", table)
	if err != nil {
		return nil, err
	}
	pks := make([]string, 0)
	for _, c := range cols {
		if c.PrimaryKey {
			pks = append(pks, c.Name)
		}
	}
	return pks, nil
}

func (ins *Inspector) CreateTableDDL(ctx context.Context, _, table string) (string, error) {
	var name, ddl string
	err := ins.db.QueryRowContext(ctx,
		`SELECT name, sql FROM sqlite_master WHERE type='table' AND name=?`, table).
		Scan(&name, &ddl)
	if err != nil {
		return "", fmt.Errorf("get DDL: %w", err)
	}
	return ddl, nil
}
