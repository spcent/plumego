package handler

import (
	"strings"
	"testing"

	"dbadmin/internal/domain/connection"
)

// --- validateColumnType ---

func TestValidateColumnType_valid(t *testing.T) {
	cases := []string{
		"INT",
		"VARCHAR(255)",
		"DECIMAL(10,2)",
		"TEXT",
		"TIMESTAMP",
		"DOUBLE PRECISION",
		"INT UNSIGNED",
		"TINYINT(1)",
		"CHAR(36)",
	}
	for _, ct := range cases {
		if err := validateColumnType(ct); err != nil {
			t.Errorf("validateColumnType(%q) returned unexpected error: %v", ct, err)
		}
	}
}

func TestValidateColumnType_invalid(t *testing.T) {
	cases := []struct {
		in  string
		msg string
	}{
		{"", "required"},
		{"INT; DROP TABLE users", "unsafe"},
		{"VARCHAR(255) -- comment", "unsafe"},
		{"TEXT /* evil */", "unsafe"},
		{"INT*/evil", "unsafe"},
		{strings.Repeat("X", 65), "too long"},
		{"INT\x00zero", "unsafe"},
	}
	for _, c := range cases {
		err := validateColumnType(c.in)
		if err == nil {
			t.Errorf("validateColumnType(%q) = nil, want error containing %q", c.in, c.msg)
			continue
		}
		if !strings.Contains(err.Error(), c.msg) {
			t.Errorf("validateColumnType(%q) error = %q, want to contain %q", c.in, err.Error(), c.msg)
		}
	}
}

// --- validateDDLLiteral ---

func TestValidateDDLLiteral_valid(t *testing.T) {
	cases := []string{
		"0",
		"'default'",
		"CURRENT_TIMESTAMP",
		"NULL",
		"42",
	}
	for _, v := range cases {
		if err := validateDDLLiteral(v); err != nil {
			t.Errorf("validateDDLLiteral(%q) returned unexpected error: %v", v, err)
		}
	}
}

func TestValidateDDLLiteral_invalid(t *testing.T) {
	cases := []struct {
		in  string
		msg string
	}{
		{"x; DROP TABLE", "unsafe"},
		{"/* comment */", "unsafe"},
		{"value --", "unsafe"},
		{strings.Repeat("x", 201), "too long"},
		{"val\x00ue", "unsafe"},
	}
	for _, c := range cases {
		err := validateDDLLiteral(c.in)
		if err == nil {
			t.Errorf("validateDDLLiteral(%q) = nil, want error containing %q", c.in, c.msg)
			continue
		}
		if !strings.Contains(err.Error(), c.msg) {
			t.Errorf("validateDDLLiteral(%q) error = %q, want to contain %q", c.in, err.Error(), c.msg)
		}
	}
}

// --- validateEngine ---

func TestValidateEngine_valid(t *testing.T) {
	cases := []string{"", "innodb", "InnoDB", "INNODB", "myisam", "memory", "archive", "csv", "blackhole"}
	for _, e := range cases {
		if err := validateEngine(e); err != nil {
			t.Errorf("validateEngine(%q) returned unexpected error: %v", e, err)
		}
	}
}

func TestValidateEngine_invalid(t *testing.T) {
	cases := []string{"rocksdb", "tokudb", "'; DROP TABLE", "badengine"}
	for _, e := range cases {
		if err := validateEngine(e); err == nil {
			t.Errorf("validateEngine(%q) = nil, want error", e)
		}
	}
}

func TestValidateDDLIdentifier(t *testing.T) {
	for _, name := range []string{"users", "user_id", "_private", "My$Field", "T2"} {
		if err := validateDDLIdentifier("table name", name); err != nil {
			t.Errorf("validateDDLIdentifier(%q) = %v, want nil", name, err)
		}
	}
	for _, name := range []string{
		"", strings.Repeat("x", 65), "bad\x00name",
		"name-with-hyphen", "123start", "name with space", "a.b",
	} {
		if err := validateDDLIdentifier("table name", name); err == nil {
			t.Errorf("validateDDLIdentifier(%q) = nil, want error", name)
		}
	}
}

// --- validateIdentifierName ---

func TestValidateIdentifierName_valid(t *testing.T) {
	cases := []string{"users", "my_view", "View2", "users_archive"}
	for _, name := range cases {
		if err := validateIdentifierName(name); err != nil {
			t.Errorf("validateIdentifierName(%q) returned unexpected error: %v", name, err)
		}
	}
}

func TestValidateIdentifierName_invalid(t *testing.T) {
	cases := []struct {
		in  string
		msg string
	}{
		{"", "required"},
		{"v; DROP TABLE users", "unsafe"},
		{"v -- comment", "unsafe"},
		{"v /* evil */", "unsafe"},
		{strings.Repeat("x", 65), "too long"},
		{"v\x00name", "unsafe"},
	}
	for _, c := range cases {
		err := validateIdentifierName(c.in)
		if err == nil {
			t.Errorf("validateIdentifierName(%q) = nil, want error containing %q", c.in, c.msg)
			continue
		}
		if !strings.Contains(err.Error(), c.msg) {
			t.Errorf("validateIdentifierName(%q) error = %q, want to contain %q", c.in, err.Error(), c.msg)
		}
	}
}

// --- buildAlterTable rename ---

func TestBuildAlterTable_renameMySQL(t *testing.T) {
	stmts := buildAlterTable("mydb", "users", alterTableRequest{RenameTable: "accounts"}, connection.DriverMySQL)
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	// Should use quoteIdent for all three identifiers.
	want := "RENAME TABLE `mydb`.`users` TO `mydb`.`accounts`"
	if stmts[0] != want {
		t.Errorf("got  %q\nwant %q", stmts[0], want)
	}
}

func TestBuildAlterTable_renameSQLite(t *testing.T) {
	stmts := buildAlterTable("", "orders", alterTableRequest{RenameTable: "archive_orders"}, connection.DriverSQLite)
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	want := `ALTER TABLE "orders" RENAME TO "archive_orders"`
	if stmts[0] != want {
		t.Errorf("got  %q\nwant %q", stmts[0], want)
	}
}

func TestBuildAlterTable_renameWithBacktick(t *testing.T) {
	// A new name with a backtick must be properly escaped (not injection).
	stmts := buildAlterTable("db", "t", alterTableRequest{RenameTable: "new`name"}, connection.DriverMySQL)
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	// quoteIdent doubles backticks inside, so the result is safe SQL.
	if strings.Contains(stmts[0], "new`name") {
		t.Errorf("unescaped backtick in rename stmt: %q", stmts[0])
	}
	if !strings.Contains(stmts[0], "new``name") {
		t.Errorf("expected escaped backtick in rename stmt: %q", stmts[0])
	}
}

func TestBuildAlterTable_quotesExistingTable(t *testing.T) {
	stmts := buildAlterTable("my`db", "user`data", alterTableRequest{
		AddColumns: []ColumnDef{{Name: "age", Type: "INT"}},
	}, connection.DriverMySQL)
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	want := "ALTER TABLE `my``db`.`user``data` ADD COLUMN `age` INT"
	if stmts[0] != want {
		t.Errorf("got  %q\nwant %q", stmts[0], want)
	}
}

func TestBuildCreateTable_nameQuoted(t *testing.T) {
	req := createTableRequest{
		Name:    "my`table",
		Columns: []ColumnDef{{Name: "id", Type: "INT"}},
	}
	ddl := buildCreateTable("mydb", req, connection.DriverMySQL)
	// The table name must be properly escaped.
	if strings.Contains(ddl, "my`table") && !strings.Contains(ddl, "my``table") {
		t.Errorf("unescaped backtick in CREATE TABLE: %q", ddl)
	}
}

func TestBuildAlterTable_addDropColumns(t *testing.T) {
	req := alterTableRequest{
		AddColumns:  []ColumnDef{{Name: "age", Type: "INT"}},
		DropColumns: []string{"legacy"},
	}
	stmts := buildAlterTable("mydb", "users", req, connection.DriverMySQL)
	if len(stmts) != 1 {
		t.Fatalf("len(stmts) = %d, want 1", len(stmts))
	}
	stmt := stmts[0]
	if !strings.Contains(stmt, "ADD COLUMN `age`") {
		t.Errorf("missing ADD COLUMN in %q", stmt)
	}
	if !strings.Contains(stmt, "DROP COLUMN `legacy`") {
		t.Errorf("missing DROP COLUMN in %q", stmt)
	}
}
