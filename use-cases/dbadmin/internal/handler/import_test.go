package handler

import (
	"strings"
	"testing"
)

func TestSplitSQL_basic(t *testing.T) {
	stmts := splitSQL("SELECT 1; SELECT 2; SELECT 3")
	if len(stmts) != 3 {
		t.Fatalf("len = %d, want 3: %v", len(stmts), stmts)
	}
}

func TestSplitSQL_trailingSemicolon(t *testing.T) {
	stmts := splitSQL("SELECT 1;")
	if len(stmts) != 1 || stmts[0] != "SELECT 1" {
		t.Fatalf("got %v, want [SELECT 1]", stmts)
	}
}

func TestSplitSQL_noSemicolon(t *testing.T) {
	stmts := splitSQL("SELECT 1")
	if len(stmts) != 1 || stmts[0] != "SELECT 1" {
		t.Fatalf("got %v, want [SELECT 1]", stmts)
	}
}

func TestSplitSQL_empty(t *testing.T) {
	stmts := splitSQL("   ")
	if len(stmts) != 0 {
		t.Fatalf("got %v, want []", stmts)
	}
}

func TestSplitSQL_blockCommentWithSemicolon(t *testing.T) {
	// Semicolon inside block comment must not split statements.
	input := "SELECT /* has ; semicolon */ 1; SELECT 2"
	stmts := splitSQL(input)
	if len(stmts) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(stmts), stmts)
	}
	if !strings.Contains(stmts[0], "SELECT") {
		t.Errorf("first stmt wrong: %q", stmts[0])
	}
}

func TestSplitSQL_lineCommentWithSemicolon(t *testing.T) {
	// The semicolon inside the -- comment must not cause a premature split.
	// Both stmts end with real semicolons; there are 2 statements.
	input := "SELECT 1; -- this has a; semicolon\nSELECT 2"
	stmts := splitSQL(input)
	if len(stmts) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(stmts), stmts)
	}
	if stmts[0] != "SELECT 1" {
		t.Errorf("stmt[0] = %q, want %q", stmts[0], "SELECT 1")
	}
}

func TestSplitSQL_hashCommentWithSemicolon(t *testing.T) {
	// MySQL # comment style: semicolon inside # comment must not split.
	input := "SELECT 1; # semicolon; here\nSELECT 2"
	stmts := splitSQL(input)
	if len(stmts) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(stmts), stmts)
	}
	if stmts[0] != "SELECT 1" {
		t.Errorf("stmt[0] = %q, want %q", stmts[0], "SELECT 1")
	}
}

func TestSplitSQL_singleQuotedSemicolon(t *testing.T) {
	// Semicolon inside single-quoted string must not split.
	input := "INSERT INTO t VALUES ('a;b'); SELECT 1"
	stmts := splitSQL(input)
	if len(stmts) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(stmts), stmts)
	}
	if !strings.Contains(stmts[0], "a;b") {
		t.Errorf("semicolon inside string was not preserved: %q", stmts[0])
	}
}

func TestSplitSQL_doubleQuotedSemicolon(t *testing.T) {
	// Semicolon inside double-quoted identifier must not split.
	input := `SELECT "col;name" FROM t; SELECT 1`
	stmts := splitSQL(input)
	if len(stmts) != 2 {
		t.Fatalf("len = %d, want 2: %v", len(stmts), stmts)
	}
}

func TestSplitSQL_multilineBlockComment(t *testing.T) {
	input := "/* line1\nline2\nline3 */\nSELECT 1"
	stmts := splitSQL(input)
	if len(stmts) != 1 {
		t.Fatalf("len = %d, want 1: %v", len(stmts), stmts)
	}
}

func TestSplitSQL_typicalMySQLDump(t *testing.T) {
	input := `
-- MySQL dump
CREATE TABLE users (
  id INT PRIMARY KEY,
  name VARCHAR(255)
);

INSERT INTO users VALUES (1, 'Alice');
INSERT INTO users VALUES (2, 'Bob''s');
`
	stmts := splitSQL(input)
	if len(stmts) != 3 {
		t.Fatalf("len = %d, want 3: %v", len(stmts), stmts)
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in   string
		n    int
		want string
	}{
		{"hello", 10, "hello"},
		{"hello world", 5, "hello..."},
		{"", 5, ""},
	}
	for _, c := range cases {
		got := truncate(c.in, c.n)
		if got != c.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", c.in, c.n, got, c.want)
		}
	}
}
