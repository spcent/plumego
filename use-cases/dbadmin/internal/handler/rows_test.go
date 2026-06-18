package handler

import (
	"strings"
	"testing"

	"dbadmin/internal/domain/connection"
)

// --- buildWhereFromPK ---

func TestBuildWhereFromPK_single_mysql(t *testing.T) {
	where, args, err := buildWhereFromPK(
		map[string]any{"id": 42},
		connection.DriverMySQL,
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := " WHERE `id` = ?"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != 42 {
		t.Errorf("args: got %v, want [42]", args)
	}
}

func TestBuildWhereFromPK_single_sqlite(t *testing.T) {
	where, args, err := buildWhereFromPK(
		map[string]any{"id": 7},
		connection.DriverSQLite,
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := ` WHERE "id" = ?1`
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != 7 {
		t.Errorf("args: got %v, want [7]", args)
	}
}

func TestBuildWhereFromPK_composite_mysql(t *testing.T) {
	// Keys are sorted: org_id, user_id
	where, args, err := buildWhereFromPK(
		map[string]any{"user_id": 2, "org_id": 10},
		connection.DriverMySQL,
		1,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := " WHERE `org_id` = ? AND `user_id` = ?"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 2 {
		t.Fatalf("args len: got %d, want 2", len(args))
	}
	// org_id=10 comes first (sorted), user_id=2 second
	if args[0] != 10 || args[1] != 2 {
		t.Errorf("args: got %v, want [10, 2]", args)
	}
}

func TestBuildWhereFromPK_composite_sqlite_startN(t *testing.T) {
	// Simulate UPDATE: SET has 3 columns, so WHERE starts at ?4
	where, args, err := buildWhereFromPK(
		map[string]any{"id": 5},
		connection.DriverSQLite,
		4, // startN = 4
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := ` WHERE "id" = ?4`
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != 5 {
		t.Errorf("args: got %v, want [5]", args)
	}
}

func TestBuildWhereFromPK_empty(t *testing.T) {
	_, _, err := buildWhereFromPK(map[string]any{}, connection.DriverMySQL, 1)
	if err == nil {
		t.Fatal("expected error for empty pkMap, got nil")
	}
}

func TestBuildWhereFromPK_invalidIdent(t *testing.T) {
	// Key that sanitises to "" (all special chars, no alphanumeric)
	_, _, err := buildWhereFromPK(
		map[string]any{"!!!": 1},
		connection.DriverMySQL,
		1,
	)
	if err == nil {
		t.Fatal("expected error for invalid identifier, got nil")
	}
}

// --- quoteIdent ---

func TestQuoteIdent_mysql(t *testing.T) {
	cases := []struct{ in, want string }{
		{"name", "`name`"},
		{"my`col", "`my``col`"}, // backtick escaped
		{"a``b", "`a````b`"},    // double backtick
	}
	for _, c := range cases {
		got := quoteIdent(c.in, connection.DriverMySQL)
		if got != c.want {
			t.Errorf("quoteIdent(%q, mysql) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestQuoteIdent_sqlite(t *testing.T) {
	cases := []struct{ in, want string }{
		{"name", `"name"`},
		{`my"col`, `"my""col"`}, // double-quote escaped
		{`a""b`, `"a""""b"`},
	}
	for _, c := range cases {
		got := quoteIdent(c.in, connection.DriverSQLite)
		if got != c.want {
			t.Errorf("quoteIdent(%q, sqlite) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- buildPlaceholders ---

func TestBuildPlaceholders_mysql(t *testing.T) {
	ph := buildPlaceholders(connection.DriverMySQL, 3)
	if len(ph) != 3 {
		t.Fatalf("len: got %d, want 3", len(ph))
	}
	for i, p := range ph {
		if p != "?" {
			t.Errorf("ph[%d] = %q, want ?", i, p)
		}
	}
}

func TestBuildPlaceholders_sqlite(t *testing.T) {
	ph := buildPlaceholders(connection.DriverSQLite, 3)
	want := []string{"?1", "?2", "?3"}
	for i, p := range ph {
		if p != want[i] {
			t.Errorf("ph[%d] = %q, want %q", i, p, want[i])
		}
	}
}

// --- sanitizeIdentifier ---

func TestSanitizeIdentifier(t *testing.T) {
	cases := []struct{ in, want string }{
		{"abc", "abc"},
		{"my_col", "my_col"},
		{"a b c", "abc"},
		{"'; DROP TABLE", "DROPTABLE"},
		{"123valid", "123valid"},
		{"!!!", ""},
	}
	for _, c := range cases {
		got := sanitizeIdentifier(c.in)
		if got != c.want {
			t.Errorf("sanitizeIdentifier(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// --- sortedColumnsValues ---

func TestSortedColumnsValues(t *testing.T) {
	m := map[string]any{"z": 3, "a": 1, "m": 2}
	cols, vals := sortedColumnsValues(m)
	if strings.Join(cols, ",") != "a,m,z" {
		t.Errorf("cols not sorted: %v", cols)
	}
	if vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("vals not in sorted key order: %v", vals)
	}
}

// --- buildFiltersWhere ---

func TestBuildFiltersWhere_eq_mysql(t *testing.T) {
	where, args, err := buildFiltersWhere(
		[]filterCondition{{Column: "status", Operator: "eq", Value: "active"}},
		connection.DriverMySQL,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := " WHERE `status` = ?"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("args: got %v, want [active]", args)
	}
}

func TestBuildFiltersWhere_like_sqlite(t *testing.T) {
	where, args, err := buildFiltersWhere(
		[]filterCondition{{Column: "name", Operator: "like", Value: "alice"}},
		connection.DriverSQLite,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := ` WHERE "name" LIKE ?1`
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != "%alice%" {
		t.Errorf("args: got %v, want [%%alice%%]", args)
	}
}

func TestBuildFiltersWhere_isNull(t *testing.T) {
	where, args, err := buildFiltersWhere(
		[]filterCondition{{Column: "deleted_at", Operator: "is_null"}},
		connection.DriverMySQL,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := " WHERE `deleted_at` IS NULL"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 0 {
		t.Errorf("args: got %v, want []", args)
	}
}

func TestBuildFiltersWhere_composite(t *testing.T) {
	filters := []filterCondition{
		{Column: "age", Operator: "gt", Value: "18"},
		{Column: "status", Operator: "eq", Value: "active"},
	}
	where, args, err := buildFiltersWhere(filters, connection.DriverSQLite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := ` WHERE "age" > ?1 AND "status" = ?2`
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 2 || args[0] != "18" || args[1] != "active" {
		t.Errorf("args: got %v, want [18 active]", args)
	}
}

func TestBuildFiltersWhere_empty(t *testing.T) {
	where, args, err := buildFiltersWhere([]filterCondition{}, connection.DriverMySQL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if where != "" {
		t.Errorf("where: got %q, want empty", where)
	}
	if len(args) != 0 {
		t.Errorf("args: got %v, want []", args)
	}
}

func TestBuildFiltersWhere_invalidCol(t *testing.T) {
	// Invalid column name is silently skipped; result should be empty.
	where, args, err := buildFiltersWhere(
		[]filterCondition{{Column: "!!!", Operator: "eq", Value: "x"}},
		connection.DriverMySQL,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if where != "" {
		t.Errorf("where: got %q, want empty (bad column skipped)", where)
	}
	if len(args) != 0 {
		t.Errorf("args: got %v, want []", args)
	}
}

func TestBuildFiltersWhere_unknownOp(t *testing.T) {
	// Unknown operator is silently skipped.
	where, args, err := buildFiltersWhere(
		[]filterCondition{{Column: "id", Operator: "between", Value: "1"}},
		connection.DriverMySQL,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if where != "" {
		t.Errorf("where: got %q, want empty (unknown op skipped)", where)
	}
	if len(args) != 0 {
		t.Errorf("args: got %v, want []", args)
	}
}

// --- normalizePagination ---

func TestNormalizePagination(t *testing.T) {
	cases := []struct {
		page, size         int
		wantPage, wantSize int
	}{
		{0, 0, 1, 50},    // both zero → defaults
		{-1, -1, 1, 50},  // negative → defaults
		{1, 500, 1, 500}, // max size allowed
		{1, 501, 1, 50},  // over max → default
		{3, 25, 3, 25},   // normal values pass through
		{1, 1, 1, 1},     // minimum positive pageSize
	}
	for _, c := range cases {
		p, ps := normalizePagination(c.page, c.size)
		if p != c.wantPage || ps != c.wantSize {
			t.Errorf("normalizePagination(%d,%d) = (%d,%d), want (%d,%d)",
				c.page, c.size, p, ps, c.wantPage, c.wantSize)
		}
	}
}

// --- buildInsertSQL ---

func TestBuildInsertSQL_mysql(t *testing.T) {
	got := buildInsertSQL(connection.DriverMySQL, "`mydb`.`users`", []string{"name", "age"})
	want := "INSERT INTO `mydb`.`users` (`name`, `age`) VALUES (?, ?)"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestBuildInsertSQL_sqlite(t *testing.T) {
	got := buildInsertSQL(connection.DriverSQLite, `"users"`, []string{"name", "age"})
	want := `INSERT INTO "users" ("name", "age") VALUES (?1, ?2)`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestBuildInsertSQL_single_col(t *testing.T) {
	got := buildInsertSQL(connection.DriverMySQL, "`t`", []string{"id"})
	want := "INSERT INTO `t` (`id`) VALUES (?)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- buildUpdateSetClause ---

func TestBuildUpdateSetClause_mysql(t *testing.T) {
	got := buildUpdateSetClause(connection.DriverMySQL, []string{"name", "age"})
	want := "`name` = ?, `age` = ?"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestBuildUpdateSetClause_sqlite(t *testing.T) {
	got := buildUpdateSetClause(connection.DriverSQLite, []string{"name", "age"})
	want := `"name" = ?1, "age" = ?2`
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}

func TestBuildUpdateSetClause_single(t *testing.T) {
	got := buildUpdateSetClause(connection.DriverMySQL, []string{"score"})
	want := "`score` = ?"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// --- buildWhereInClause ---

func TestBuildWhereInClause_mysql(t *testing.T) {
	where, args := buildWhereInClause("id", []any{1, 2, 3}, connection.DriverMySQL, 1)
	want := " WHERE `id` IN (?, ?, ?)"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 3 || args[0] != 1 || args[1] != 2 || args[2] != 3 {
		t.Errorf("args: got %v, want [1 2 3]", args)
	}
}

func TestBuildWhereInClause_sqlite_startN(t *testing.T) {
	// Simulate UPDATE: SET has 2 columns, so WHERE starts at ?3
	where, args := buildWhereInClause("id", []any{5, 7}, connection.DriverSQLite, 3)
	want := ` WHERE "id" IN (?3, ?4)`
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 2 || args[0] != 5 || args[1] != 7 {
		t.Errorf("args: got %v, want [5 7]", args)
	}
}

func TestBuildWhereInClause_single(t *testing.T) {
	where, args := buildWhereInClause("user_id", []any{42}, connection.DriverMySQL, 1)
	want := " WHERE `user_id` IN (?)"
	if where != want {
		t.Errorf("where: got %q, want %q", where, want)
	}
	if len(args) != 1 || args[0] != 42 {
		t.Errorf("args: got %v, want [42]", args)
	}
}
