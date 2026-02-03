package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spcent/plumego/tenant"
)

func setupTenantTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			email TEXT
		);

		CREATE TABLE products (
			id INTEGER PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL,
			price REAL
		);

		INSERT INTO users (tenant_id, name, email) VALUES
			('tenant-1', 'Alice', 'alice@tenant1.com'),
			('tenant-1', 'Bob', 'bob@tenant1.com'),
			('tenant-2', 'Charlie', 'charlie@tenant2.com'),
			('tenant-2', 'Diana', 'diana@tenant2.com');

		INSERT INTO products (tenant_id, name, price) VALUES
			('tenant-1', 'Product A', 10.0),
			('tenant-1', 'Product B', 20.0),
			('tenant-2', 'Product C', 15.0);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestNewTenantDB(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	if tdb == nil {
		t.Fatal("expected non-nil TenantDB")
	}
	if tdb.tenantColumn != "tenant_id" {
		t.Errorf("expected default column 'tenant_id', got '%s'", tdb.tenantColumn)
	}
}

func TestNewTenantDB_CustomColumn(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db, WithTenantColumn("org_id"))
	if tdb.tenantColumn != "org_id" {
		t.Errorf("expected column 'org_id', got '%s'", tdb.tenantColumn)
	}
}

func TestTenantDB_QueryFromContext(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	rows, err := tdb.QueryFromContext(ctx, "SELECT name FROM users")
	if err != nil {
		t.Fatalf("QueryFromContext failed: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	// Should only get tenant-1 users
	if len(names) != 2 {
		t.Errorf("expected 2 users for tenant-1, got %d", len(names))
	}
	if !contains(names, "Alice") || !contains(names, "Bob") {
		t.Errorf("expected Alice and Bob, got %v", names)
	}
}

func TestTenantDB_QueryContext(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := context.Background()

	rows, err := tdb.QueryContext(ctx, "tenant-2", "SELECT name FROM users")
	if err != nil {
		t.Fatalf("QueryContext failed: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	// Should only get tenant-2 users
	if len(names) != 2 {
		t.Errorf("expected 2 users for tenant-2, got %d", len(names))
	}
	if !contains(names, "Charlie") || !contains(names, "Diana") {
		t.Errorf("expected Charlie and Diana, got %v", names)
	}
}

func TestTenantDB_QueryWithExistingWhere(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	rows, err := tdb.QueryFromContext(ctx, "SELECT name FROM users WHERE name LIKE 'A%'")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	if len(names) != 1 {
		t.Errorf("expected 1 user, got %d", len(names))
	}
	if len(names) > 0 && names[0] != "Alice" {
		t.Errorf("expected Alice, got %s", names[0])
	}
}

func TestTenantDB_QueryRowFromContext(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	var count int
	err := tdb.QueryRowFromContext(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("QueryRowFromContext failed: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 users for tenant-1, got %d", count)
	}
}

func TestTenantDB_ExecFromContext(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Update users for tenant-1
	result, err := tdb.ExecFromContext(ctx, "UPDATE users SET email = ? WHERE name = ?", "newemail@test.com", "Alice")
	if err != nil {
		t.Fatalf("ExecFromContext failed: %v", err)
	}

	rows, _ := result.RowsAffected()
	if rows != 1 {
		t.Errorf("expected 1 row affected, got %d", rows)
	}

	// Verify only tenant-1's Alice was updated
	var email string
	tdb.QueryRowContext(ctx, "tenant-1", "SELECT email FROM users WHERE name = ?", "Alice").Scan(&email)
	if email != "newemail@test.com" {
		t.Errorf("expected email updated to newemail@test.com, got %s", email)
	}

	// Verify tenant-2's data unchanged
	var tenant2Count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE tenant_id = 'tenant-2' AND email = 'newemail@test.com'").Scan(&tenant2Count)
	if tenant2Count != 0 {
		t.Error("tenant-2 data should not be affected")
	}
}

func TestTenantDB_ExecContext_Delete(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := context.Background()

	// Delete from tenant-1
	result, err := tdb.ExecContext(ctx, "tenant-1", "DELETE FROM users WHERE name = ?", "Alice")
	if err != nil {
		t.Fatalf("ExecContext DELETE failed: %v", err)
	}

	rows, _ := result.RowsAffected()
	if rows != 1 {
		t.Errorf("expected 1 row deleted, got %d", rows)
	}

	// Verify tenant-1 now has 1 user
	var count1 int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE tenant_id = 'tenant-1'").Scan(&count1)
	if count1 != 1 {
		t.Errorf("expected 1 user left in tenant-1, got %d", count1)
	}

	// Verify tenant-2 still has 2 users
	var count2 int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE tenant_id = 'tenant-2'").Scan(&count2)
	if count2 != 2 {
		t.Errorf("expected 2 users in tenant-2, got %d", count2)
	}
}

func TestTenantDB_AddTenantToInsert(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)

	query := "INSERT INTO users (name, email) VALUES (?, ?)"
	args := []interface{}{"Eve", "eve@test.com"}

	filteredQuery, filteredArgs := tdb.addTenantToInsert(query, "tenant-1", args)

	// Should add tenant_id column and value
	if !strings.Contains(filteredQuery, "tenant_id") {
		t.Error("expected tenant_id in filtered query")
	}
	if len(filteredArgs) != 3 {
		t.Errorf("expected 3 args (including tenant), got %d", len(filteredArgs))
	}
	if filteredArgs[2] != "tenant-1" {
		t.Errorf("expected tenant-1 as last arg, got %v", filteredArgs[2])
	}
}

func TestTenantDB_MissingTenantInContext(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := context.Background() // No tenant ID

	_, err := tdb.QueryFromContext(ctx, "SELECT * FROM users")
	if err == nil {
		t.Error("expected error when tenant ID missing")
	}
	if !strings.Contains(err.Error(), "tenant ID not found") {
		t.Errorf("expected 'tenant ID not found' error, got: %v", err)
	}
}

func TestTenantDB_RawDB(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)

	// RawDB should return underlying db
	rawDB := tdb.RawDB()
	if rawDB != db {
		t.Error("expected RawDB to return underlying db")
	}

	// Admin query - bypasses tenant filtering
	var totalUsers int
	err := rawDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&totalUsers)
	if err != nil {
		t.Fatalf("RawDB query failed: %v", err)
	}
	if totalUsers != 4 {
		t.Errorf("expected 4 total users (all tenants), got %d", totalUsers)
	}
}

func TestValidateQuery_DropTable(t *testing.T) {
	err := ValidateQuery("DROP TABLE users", "tenant_id")
	if err == nil {
		t.Error("expected error for DROP TABLE")
	}
	if !strings.Contains(err.Error(), "DROP TABLE not allowed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuery_Truncate(t *testing.T) {
	err := ValidateQuery("TRUNCATE TABLE users", "tenant_id")
	if err == nil {
		t.Error("expected error for TRUNCATE")
	}
	if !strings.Contains(err.Error(), "TRUNCATE not allowed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuery_MissingWhere(t *testing.T) {
	err := ValidateQuery("SELECT * FROM users", "tenant_id")
	if err == nil {
		t.Error("expected error for missing WHERE clause")
	}
	if !strings.Contains(err.Error(), "missing WHERE clause") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuery_MissingTenantFilter(t *testing.T) {
	err := ValidateQuery("SELECT * FROM users WHERE active = true", "tenant_id")
	if err == nil {
		t.Error("expected error for missing tenant_id filter")
	}
	if !strings.Contains(err.Error(), "missing tenant_id filter") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateQuery_Valid(t *testing.T) {
	err := ValidateQuery("SELECT * FROM users WHERE tenant_id = ? AND active = true", "tenant_id")
	if err != nil {
		t.Errorf("expected valid query to pass, got error: %v", err)
	}
}

func TestValidateQuery_Insert(t *testing.T) {
	// INSERT doesn't need WHERE clause
	err := ValidateQuery("INSERT INTO users (name) VALUES (?)", "tenant_id")
	if err != nil {
		t.Errorf("INSERT should be allowed without WHERE, got: %v", err)
	}
}

func TestTenantDB_NilDB(t *testing.T) {
	var tdb *TenantDB
	ctx := context.Background()

	_, err := tdb.QueryContext(ctx, "tenant-1", "SELECT * FROM users")
	if err == nil {
		t.Error("expected error for nil TenantDB")
	}
}

func TestTenantDB_ComplexQuery(t *testing.T) {
	db := setupTenantTestDB(t)
	defer db.Close()

	tdb := NewTenantDB(db)
	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Query with ORDER BY
	rows, err := tdb.QueryFromContext(ctx, "SELECT name FROM users ORDER BY name")
	if err != nil {
		t.Fatalf("Query with ORDER BY failed: %v", err)
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}

	if len(names) != 2 {
		t.Errorf("expected 2 users, got %d", len(names))
	}
	// Should be alphabetically sorted
	if len(names) == 2 && (names[0] != "Alice" || names[1] != "Bob") {
		t.Errorf("expected [Alice, Bob], got %v", names)
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
