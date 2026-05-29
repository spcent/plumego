package sql

import (
	"database/sql"
	"fmt"
	"strings"
)

// dialect abstracts SQL differences between database backends.
type dialect interface {
	// placeholder returns the parameter placeholder for the nth argument (1-based).
	//   sqlite: "$1", "$2", ...
	//   mysql:  "?", "?", ...
	placeholder(n int) string

	// placeholders returns count comma-separated placeholders starting from position from.
	//   sqlite: placeholders(1, 3) -> "$1, $2, $3"
	//   mysql:  placeholders(1, 3) -> "?, ?, ?"
	placeholders(from, count int) string

	// returningID generates the clause to retrieve the auto-generated primary key after INSERT.
	//   sqlite: " RETURNING endpoint_id"
	//   mysql:  "" (caller uses LastInsertId)
	returningID(column string) string

	// upsertClause generates the full conflict resolution clause for upserts.
	// cols are the columns to update.
	//   sqlite: "ON CONFLICT(col1, col2) DO UPDATE SET col3 = excluded.col3, col4 = excluded.col4"
	//   mysql:  "ON DUPLICATE KEY UPDATE col3 = VALUES(col3), col4 = VALUES(col4)"
	upsertClause(conflictCols string, cols []string) string

	// initPragmas runs driver-specific initialization after connection open.
	//   sqlite: PRAGMA foreign_keys=ON; PRAGMA journal_mode=WAL; ...
	//   mysql:  no-op
	initPragmas(db *sql.DB) error
}

// sqliteDialect implements dialect for SQLite.
type sqliteDialect struct{}

func (d sqliteDialect) placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (d sqliteDialect) placeholders(from, count int) string {
	parts := make([]string, count)
	for i := range count {
		parts[i] = fmt.Sprintf("$%d", from+i)
	}
	return strings.Join(parts, ", ")
}

func (d sqliteDialect) returningID(column string) string {
	return fmt.Sprintf(" RETURNING %s", column)
}

func (d sqliteDialect) upsertClause(conflictCols string, cols []string) string {
	exprs := make([]string, len(cols))
	for i, col := range cols {
		exprs[i] = fmt.Sprintf("%s = excluded.%s", col, col)
	}
	return fmt.Sprintf("ON CONFLICT(%s) DO UPDATE SET %s", conflictCols, strings.Join(exprs, ", "))
}

func (d sqliteDialect) initPragmas(db *sql.DB) error {
	pragmas := []string{
		"PRAGMA foreign_keys=ON",
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=10000",
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			return fmt.Errorf("exec %q: %w", p, err)
		}
	}
	return nil
}

// mysqlDialect implements dialect for MySQL.
type mysqlDialect struct{}

func (d mysqlDialect) placeholder(n int) string {
	return "?"
}

func (d mysqlDialect) placeholders(from, count int) string {
	parts := make([]string, count)
	for i := range count {
		parts[i] = "?"
	}
	return strings.Join(parts, ", ")
}

func (d mysqlDialect) returningID(column string) string {
	return ""
}

func (d mysqlDialect) upsertClause(conflictCols string, cols []string) string {
	// MySQL uses VALUES(col) to reference the value that would have been inserted.
	exprs := make([]string, len(cols))
	for i, col := range cols {
		exprs[i] = fmt.Sprintf("%s = VALUES(%s)", col, col)
	}
	return fmt.Sprintf("ON DUPLICATE KEY UPDATE %s", strings.Join(exprs, ", "))
}

func (d mysqlDialect) initPragmas(db *sql.DB) error {
	// MySQL doesn't need PRAGMA-like initialization.
	return nil
}
