// Package postgres provides PostgreSQL introspection via information_schema and pg_catalog.
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"dbadmin/internal/dbmanager"
)

// Inspector implements dbmanager.Inspector for PostgreSQL.
type Inspector struct {
	db *sql.DB
}

// New creates a PostgreSQL inspector for the given open connection.
func New(db *sql.DB) *Inspector { return &Inspector{db: db} }

// defaultSchema returns "public" when schema is empty, matching PostgreSQL's
// default search_path behavior.
func defaultSchema(schema string) string {
	if schema == "" {
		return "public"
	}
	return schema
}

func (ins *Inspector) Databases(ctx context.Context) ([]string, error) {
	rows, err := ins.db.QueryContext(ctx,
		`SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname`)
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
	schema := defaultSchema(db)
	rows, err := ins.db.QueryContext(ctx,
		`SELECT t.table_name, t.table_type,
		        COALESCE(obj_description(c.oid, 'pg_class'), '') AS comment,
		        COALESCE(s.n_live_tup, 0) AS approx_rows
		 FROM information_schema.tables t
		 LEFT JOIN pg_catalog.pg_class c
		   ON c.relname = t.table_name
		  AND c.relnamespace = (SELECT oid FROM pg_catalog.pg_namespace WHERE nspname = t.table_schema)
		 LEFT JOIN pg_catalog.pg_stat_user_tables s
		   ON s.relname = t.table_name AND s.schemaname = t.table_schema
		 WHERE t.table_schema = $1
		 ORDER BY t.table_name`, schema)
	if err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	defer rows.Close()
	tables := make([]dbmanager.TableInfo, 0)
	for rows.Next() {
		var t dbmanager.TableInfo
		var tableType string
		if err := rows.Scan(&t.Name, &tableType, &t.Comment, &t.Rows); err != nil {
			return nil, err
		}
		if tableType == "VIEW" {
			t.Type = "VIEW"
		} else {
			t.Type = "TABLE"
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (ins *Inspector) Columns(ctx context.Context, db, table string) ([]dbmanager.ColumnInfo, error) {
	schema := defaultSchema(db)
	rows, err := ins.db.QueryContext(ctx,
		`SELECT c.column_name, c.ordinal_position, c.data_type,
		        CASE WHEN c.character_maximum_length IS NOT NULL
		             THEN c.data_type || '(' || c.character_maximum_length || ')'
		             WHEN c.numeric_precision IS NOT NULL AND c.numeric_scale IS NOT NULL AND c.data_type = 'numeric'
		             THEN c.data_type || '(' || c.numeric_precision || ',' || c.numeric_scale || ')'
		             ELSE c.data_type END AS full_type,
		        c.is_nullable, COALESCE(c.column_default, ''),
		        COALESCE(pk.is_pk, false), COALESCE(c.column_default LIKE 'nextval%', false)
		 FROM information_schema.columns c
		 LEFT JOIN (
		     SELECT kcu.column_name, true AS is_pk
		     FROM information_schema.table_constraints tc
		     JOIN information_schema.key_column_usage kcu
		       ON kcu.constraint_name = tc.constraint_name
		      AND kcu.table_schema = tc.table_schema
		      AND kcu.table_name = tc.table_name
		     WHERE tc.constraint_type = 'PRIMARY KEY'
		       AND tc.table_schema = $1 AND tc.table_name = $2
		 ) pk ON pk.column_name = c.column_name
		 WHERE c.table_schema = $1 AND c.table_name = $2
		 ORDER BY c.ordinal_position`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}
	defer rows.Close()
	cols := make([]dbmanager.ColumnInfo, 0)
	for rows.Next() {
		var c dbmanager.ColumnInfo
		var nullable string
		if err := rows.Scan(&c.Name, &c.Position, &c.DataType, &c.FullType,
			&nullable, &c.Default, &c.PrimaryKey, &c.AutoIncr); err != nil {
			return nil, err
		}
		c.Nullable = nullable == "YES"
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (ins *Inspector) Indexes(ctx context.Context, db, table string) ([]dbmanager.IndexInfo, error) {
	schema := defaultSchema(db)
	rows, err := ins.db.QueryContext(ctx,
		`SELECT ic.relname AS index_name, ix.indisunique, a.attname AS column_name,
		        am.amname AS index_type
		 FROM pg_index ix
		 JOIN pg_class t ON t.oid = ix.indrelid
		 JOIN pg_class ic ON ic.oid = ix.indexrelid
		 JOIN pg_namespace n ON n.oid = t.relnamespace
		 JOIN pg_am am ON am.oid = ic.relam
		 JOIN unnest(ix.indkey) WITH ORDINALITY AS k(attnum, ord) ON true
		 JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = k.attnum
		 WHERE n.nspname = $1 AND t.relname = $2
		 ORDER BY ic.relname, k.ord`, schema, table)
	if err != nil {
		return nil, fmt.Errorf("list indexes: %w", err)
	}
	defer rows.Close()

	byName := make(map[string]*dbmanager.IndexInfo)
	var order []string
	for rows.Next() {
		var name, col, idxType string
		var unique bool
		if err := rows.Scan(&name, &unique, &col, &idxType); err != nil {
			return nil, err
		}
		if _, ok := byName[name]; !ok {
			byName[name] = &dbmanager.IndexInfo{
				Name:   name,
				Unique: unique,
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
	schema := defaultSchema(db)
	rows, err := ins.db.QueryContext(ctx,
		`SELECT tc.constraint_name, kcu.column_name,
		        ccu.table_name AS ref_table, ccu.column_name AS ref_column,
		        COALESCE(rc.delete_rule, ''), COALESCE(rc.update_rule, '')
		 FROM information_schema.table_constraints tc
		 JOIN information_schema.key_column_usage kcu
		   ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema
		 JOIN information_schema.constraint_column_usage ccu
		   ON ccu.constraint_name = tc.constraint_name AND ccu.table_schema = tc.table_schema
		 JOIN information_schema.referential_constraints rc
		   ON rc.constraint_name = tc.constraint_name AND rc.constraint_schema = tc.table_schema
		 WHERE tc.constraint_type = 'FOREIGN KEY'
		   AND tc.table_schema = $1 AND tc.table_name = $2
		 ORDER BY tc.constraint_name`, schema, table)
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
	schema := defaultSchema(db)
	rows, err := ins.db.QueryContext(ctx,
		`SELECT kcu.column_name
		 FROM information_schema.table_constraints tc
		 JOIN information_schema.key_column_usage kcu
		   ON kcu.constraint_name = tc.constraint_name AND kcu.table_schema = tc.table_schema
		 WHERE tc.constraint_type = 'PRIMARY KEY'
		   AND tc.table_schema = $1 AND tc.table_name = $2
		 ORDER BY kcu.ordinal_position`, schema, table)
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
	cols, err := ins.Columns(ctx, db, table)
	if err != nil {
		return "", fmt.Errorf("create table ddl: %w", err)
	}
	schema := defaultSchema(db)

	var sb strings.Builder
	sb.WriteString("-- PostgreSQL DDL preview\n")
	sb.WriteString(fmt.Sprintf("-- Table: %s.%s\n", schema, table))
	sb.WriteString("-- Use pg_dump for complete DDL with constraints\n")
	sb.WriteString(fmt.Sprintf("CREATE TABLE %q.%q (\n", schema, table))
	for i, c := range cols {
		sb.WriteString("  ")
		sb.WriteString(fmt.Sprintf("%q", c.Name))
		sb.WriteString(" ")
		sb.WriteString(c.FullType)
		if !c.Nullable {
			sb.WriteString(" NOT NULL")
		}
		if c.Default != "" {
			sb.WriteString(" DEFAULT " + c.Default)
		}
		if i < len(cols)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(")")
	return sb.String(), nil
}
