package handler

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"dbadmin/internal/domain/connection"
)

// openTestSQLite opens an in-memory SQLite database, creates a test table,
// and returns the db handle. The caller must close it.
func openTestSQLite(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.ExecContext(context.Background(), `
		CREATE TABLE users (
			id   INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age  INTEGER,
			bio  TEXT
		)
	`)
	if err != nil {
		db.Close()
		t.Fatalf("create table: %v", err)
	}
	for _, row := range []struct {
		id   int
		name string
		age  int
		bio  string
	}{
		{1, "Alice", 30, "engineer"},
		{2, "Bob", 25, ""},
		{3, "Charlie", 35, "manager"},
	} {
		_, err = db.ExecContext(context.Background(),
			`INSERT INTO users VALUES (?1, ?2, ?3, ?4)`,
			row.id, row.name, row.age, row.bio)
		if err != nil {
			db.Close()
			t.Fatalf("insert: %v", err)
		}
	}
	return db
}

const sqliteDriver = connection.DriverSQLite

// TestIntegration_ListRows verifies that the query built by the row handler
// returns the correct count from a real SQLite database.
func TestIntegration_ListRows(t *testing.T) {
	db := openTestSQLite(t)
	defer db.Close()

	tableFQN := `"users"`
	page, pageSize := normalizePagination(1, 50)
	offset := (page - 1) * pageSize

	// No filters, no sort.
	where := ""
	var whereArgs []any
	orderBy := ""
	nArgs := len(whereArgs)
	selectQ := buildSelectQuery(`"name", "age"`, tableFQN, where, orderBy,
		nthPlaceholder(sqliteDriver, nArgs+1),
		nthPlaceholder(sqliteDriver, nArgs+2))
	args := append(whereArgs, pageSize, offset)

	rows, err := db.QueryContext(context.Background(), selectQ, args...)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	result, _, err := scanRows(rows, cols)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("len(result) = %d, want 3", len(result))
	}
}

// TestIntegration_FilteredCount verifies that the COUNT query respects filters.
func TestIntegration_FilteredCount(t *testing.T) {
	db := openTestSQLite(t)
	defer db.Close()

	filters := []filterCondition{{Column: "age", Operator: "gt", Value: "25"}}
	where, whereArgs, err := buildFiltersWhere(filters, sqliteDriver)
	if err != nil {
		t.Fatalf("buildFiltersWhere: %v", err)
	}

	var count int64
	countQ := `SELECT COUNT(*) FROM "users"` + where
	if err := db.QueryRowContext(context.Background(), countQ, whereArgs...).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	// Alice (30) and Charlie (35) are > 25.
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// TestIntegration_InsertUpdateDelete exercises the full CRUD path using the
// SQL builders against a real SQLite database.
func TestIntegration_InsertUpdateDelete(t *testing.T) {
	db := openTestSQLite(t)
	defer db.Close()

	// INSERT
	tableFQN := `"users"`
	cols, vals := sortedColumnsValues(map[string]any{"id": 99, "name": "Dave", "age": 40, "bio": "new"})
	insertSQL := buildInsertSQL(sqliteDriver, tableFQN, cols)
	if _, err := db.ExecContext(context.Background(), insertSQL, vals...); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Verify insert.
	var name string
	if err := db.QueryRowContext(context.Background(),
		`SELECT name FROM "users" WHERE id = 99`).Scan(&name); err != nil {
		t.Fatalf("select after insert: %v", err)
	}
	if name != "Dave" {
		t.Errorf("name = %q, want Dave", name)
	}

	// UPDATE
	setCols, setVals := sortedColumnsValues(map[string]any{"name": "David"})
	whereSQL, whereArgs, err := buildWhereFromPK(map[string]any{"id": 99}, sqliteDriver, len(setCols)+1)
	if err != nil {
		t.Fatalf("buildWhereFromPK: %v", err)
	}
	updateSQL := `UPDATE "users" SET ` + buildUpdateSetClause(sqliteDriver, setCols) + whereSQL
	args := append(setVals, whereArgs...)
	if _, err := db.ExecContext(context.Background(), updateSQL, args...); err != nil {
		t.Fatalf("update: %v", err)
	}

	if err := db.QueryRowContext(context.Background(),
		`SELECT name FROM "users" WHERE id = 99`).Scan(&name); err != nil {
		t.Fatalf("select after update: %v", err)
	}
	if name != "David" {
		t.Errorf("name after update = %q, want David", name)
	}

	// DELETE
	whereSQL2, whereArgs2, err := buildWhereFromPK(map[string]any{"id": 99}, sqliteDriver, 1)
	if err != nil {
		t.Fatalf("buildWhereFromPK: %v", err)
	}
	deleteSQL := `DELETE FROM "users"` + whereSQL2
	if _, err := db.ExecContext(context.Background(), deleteSQL, whereArgs2...); err != nil {
		t.Fatalf("delete: %v", err)
	}

	var cnt int
	db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM "users" WHERE id = 99`).Scan(&cnt)
	if cnt != 0 {
		t.Errorf("row still exists after delete")
	}
}

// TestIntegration_SortedResults verifies ORDER BY is correctly applied.
func TestIntegration_SortedResults(t *testing.T) {
	db := openTestSQLite(t)
	defer db.Close()

	orderBy := ` ORDER BY ` + quoteIdent("age", sqliteDriver) + ` DESC`
	q := buildSelectQuery("*", `"users"`, "", orderBy,
		nthPlaceholder(sqliteDriver, 1), nthPlaceholder(sqliteDriver, 2))

	rows, err := db.QueryContext(context.Background(), q, 10, 0)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	cols, _ := rows.Columns()
	result, _, err := scanRows(rows, cols)
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("len = %d, want 3", len(result))
	}
	// First row should be Charlie (age 35).
	if result[0]["name"] != "Charlie" {
		t.Errorf("first row name = %v, want Charlie", result[0]["name"])
	}
}

// buildSelectQuery is a small test helper that mirrors the SELECT query shape
// used in RowHandler.List without pulling in the full handler.
func buildSelectQuery(selectExpr, tableFQN, where, orderBy, limitPH, offsetPH string) string {
	return "SELECT " + selectExpr + " FROM " + tableFQN + where + orderBy +
		" LIMIT " + limitPH + " OFFSET " + offsetPH
}
