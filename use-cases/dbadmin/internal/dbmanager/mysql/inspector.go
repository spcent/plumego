// Package mysql provides MySQL introspection via INFORMATION_SCHEMA.
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dbadmin/internal/dbmanager"
)

// Inspector implements dbmanager.Inspector for MySQL.
type Inspector struct {
	db *sql.DB
}

// New creates a MySQL inspector for the given open connection.
func New(db *sql.DB) *Inspector { return &Inspector{db: db} }

func (ins *Inspector) Databases(ctx context.Context) ([]string, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA ORDER BY SCHEMA_NAME`)
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	defer rows.Close()
	dbs := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

func (ins *Inspector) Tables(ctx context.Context, db string) ([]dbmanager.TableInfo, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT TABLE_NAME, TABLE_TYPE, IFNULL(TABLE_COMMENT,''), IFNULL(ENGINE,''), IFNULL(TABLE_ROWS,0)
		 FROM INFORMATION_SCHEMA.TABLES
		 WHERE TABLE_SCHEMA = ?
		 ORDER BY TABLE_NAME`, db)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	tables := make([]dbmanager.TableInfo, 0)
	for rows.Next() {
		var t dbmanager.TableInfo
		if err := rows.Scan(&t.Name, &t.Type, &t.Comment, &t.Engine, &t.Rows); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (ins *Inspector) Columns(ctx context.Context, db, table string) ([]dbmanager.ColumnInfo, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT COLUMN_NAME, ORDINAL_POSITION, DATA_TYPE, COLUMN_TYPE,
		        IS_NULLABLE, IFNULL(COLUMN_DEFAULT,''), IFNULL(COLUMN_COMMENT,''),
		        COLUMN_KEY, EXTRA
		 FROM INFORMATION_SCHEMA.COLUMNS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		 ORDER BY ORDINAL_POSITION`, db, table)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}
	defer rows.Close()
	cols := make([]dbmanager.ColumnInfo, 0)
	for rows.Next() {
		var c dbmanager.ColumnInfo
		var nullable, key, extra string
		if err := rows.Scan(&c.Name, &c.Position, &c.DataType, &c.FullType,
			&nullable, &c.Default, &c.Comment, &key, &extra); err != nil {
			return nil, err
		}
		c.Nullable = nullable == "YES"
		c.PrimaryKey = key == "PRI"
		c.AutoIncr = strings.Contains(extra, "auto_increment")
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (ins *Inspector) Indexes(ctx context.Context, db, table string) ([]dbmanager.IndexInfo, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT INDEX_NAME, NON_UNIQUE, COLUMN_NAME, IFNULL(INDEX_TYPE,'')
		 FROM INFORMATION_SCHEMA.STATISTICS
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		 ORDER BY INDEX_NAME, SEQ_IN_INDEX`, db, table)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	defer rows.Close()

	byName := make(map[string]*dbmanager.IndexInfo)
	var order []string
	for rows.Next() {
		var name, col, idxType string
		var nonUnique int
		if err := rows.Scan(&name, &nonUnique, &col, &idxType); err != nil {
			return nil, err
		}
		if _, ok := byName[name]; !ok {
			byName[name] = &dbmanager.IndexInfo{
				Name:   name,
				Unique: nonUnique == 0,
				Type:   idxType,
			}
			order = append(order, name)
		}
		byName[name].Columns = append(byName[name].Columns, col)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]dbmanager.IndexInfo, 0, len(order))
	for _, name := range order {
		out = append(out, *byName[name])
	}
	return out, nil
}

func (ins *Inspector) ForeignKeys(ctx context.Context, db, table string) ([]dbmanager.ForeignKeyInfo, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT kcu.CONSTRAINT_NAME, kcu.COLUMN_NAME,
		        kcu.REFERENCED_TABLE_NAME, kcu.REFERENCED_COLUMN_NAME,
		        IFNULL(rc.DELETE_RULE,''), IFNULL(rc.UPDATE_RULE,'')
		 FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE kcu
		 JOIN INFORMATION_SCHEMA.REFERENTIAL_CONSTRAINTS rc
		   ON rc.CONSTRAINT_SCHEMA = kcu.TABLE_SCHEMA
		  AND rc.CONSTRAINT_NAME   = kcu.CONSTRAINT_NAME
		 WHERE kcu.TABLE_SCHEMA = ? AND kcu.TABLE_NAME = ?
		   AND kcu.REFERENCED_TABLE_NAME IS NOT NULL
		 ORDER BY kcu.CONSTRAINT_NAME`, db, table)
	if err != nil {
		return nil, fmt.Errorf("list foreign keys: %w", err)
	}
	defer rows.Close()
	fks := make([]dbmanager.ForeignKeyInfo, 0)
	for rows.Next() {
		var fk dbmanager.ForeignKeyInfo
		if err := rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn,
			&fk.OnDelete, &fk.OnUpdate); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (ins *Inspector) PrimaryKeys(ctx context.Context, db, table string) ([]string, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		 WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY'
		 ORDER BY ORDINAL_POSITION`, db, table)
	if err != nil {
		return nil, fmt.Errorf("list primary keys: %w", err)
	}
	defer rows.Close()
	pks := make([]string, 0)
	for rows.Next() {
		var col string
		if err := rows.Scan(&col); err != nil {
			return nil, err
		}
		pks = append(pks, col)
	}
	return pks, rows.Err()
}

func (ins *Inspector) CreateTableDDL(ctx context.Context, db, table string) (string, error) {
	var name, ddl string
	err := ins.db.QueryRowContext(ctx,
		fmt.Sprintf("SHOW CREATE TABLE `%s`.`%s`", db, table)).Scan(&name, &ddl)
	if err != nil {
		return "", fmt.Errorf("show create table: %w", err)
	}
	return ddl, nil
}
