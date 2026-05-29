package handler

import (
	"strings"
	"testing"
)

// --- classifySQL ---

func TestClassifySQL_selectTypes(t *testing.T) {
	cases := []struct {
		sql      string
		isSelect bool
	}{
		{"SELECT * FROM users", true},
		{"select id from t", true},
		{"SHOW TABLES", true},
		{"SHOW DATABASES", true},
		{"EXPLAIN SELECT 1", true},
		{"DESCRIBE users", true},
		{"DESC users", true},
		{"PRAGMA table_info(users)", true},
		{"WITH cte AS (SELECT 1) SELECT * FROM cte", true},
		{"  SELECT 1  ", true},
	}
	for _, c := range cases {
		cls := classifySQL(c.sql)
		if cls.IsSelect != c.isSelect {
			t.Errorf("classifySQL(%q).IsSelect = %v, want %v", c.sql, cls.IsSelect, c.isSelect)
		}
	}
}

func TestClassifySQL_dangerousTypes(t *testing.T) {
	cases := []struct {
		sql    string
		danger bool
		reason string
	}{
		{"DROP TABLE users", true, "DROP"},
		{"TRUNCATE TABLE users", true, "TRUNCATE"},
		{"ALTER TABLE users ADD COLUMN x INT", true, "ALTER"},
		{"DELETE FROM users", true, "without WHERE"},
		{"UPDATE users SET name = 'x'", true, "without WHERE"},
	}
	for _, c := range cases {
		cls := classifySQL(c.sql)
		if !cls.IsDangerous {
			t.Errorf("classifySQL(%q).IsDangerous = false, want true", c.sql)
		}
		if !strings.Contains(cls.Reason, c.reason) {
			t.Errorf("classifySQL(%q).Reason = %q, want to contain %q", c.sql, cls.Reason, c.reason)
		}
	}
}

func TestClassifySQL_safeWriteWithWhere(t *testing.T) {
	// DELETE/UPDATE with WHERE are not flagged as dangerous.
	cases := []string{
		"DELETE FROM users WHERE id = 1",
		"UPDATE users SET name = 'x' WHERE id = 1",
		"delete from orders where status = 'cancelled'",
	}
	for _, sql := range cases {
		cls := classifySQL(sql)
		if cls.IsDangerous {
			t.Errorf("classifySQL(%q).IsDangerous = true, want false", sql)
		}
	}
}

func TestClassifySQL_notDangerousNotSelect(t *testing.T) {
	// INSERT, REPLACE, CREATE — not dangerous, not select.
	cases := []string{
		"INSERT INTO users (name) VALUES ('x')",
		"REPLACE INTO t VALUES (1)",
		"CREATE TABLE t (id INT)",
	}
	for _, sql := range cases {
		cls := classifySQL(sql)
		if cls.IsSelect {
			t.Errorf("classifySQL(%q).IsSelect = true, want false", sql)
		}
		if cls.IsDangerous {
			t.Errorf("classifySQL(%q).IsDangerous = true, want false", sql)
		}
	}
}

// --- stripBlockComments ---

func TestStripBlockComments(t *testing.T) {
	cases := []struct{ in, want string }{
		{"SELECT 1", "SELECT 1"},
		{"SELECT /* comment */ 1", "SELECT   1"},
		{"/* leading */SELECT 1", " SELECT 1"},
		{"DELETE /* WHERE */ FROM t", "DELETE   FROM t"},
		{"a/**/b", "a b"},
		{"no comment here", "no comment here"},
		// Non-nested: first */ ends the comment; remaining */ and text are literal.
		// The comment span is /* outer /* inner */ — replaced by a single space.
		// The ` still comment */ end` follows as literal text.
		{"/* outer /* inner */ still comment */ end", "  still comment */ end"},
	}
	for _, c := range cases {
		got := stripBlockComments(c.in)
		if got != c.want {
			t.Errorf("stripBlockComments(%q)\n  got  %q\n  want %q", c.in, got, c.want)
		}
	}
}

// --- hasWhereClause ---

func TestHasWhereClause_present(t *testing.T) {
	cases := []string{
		"DELETE FROM t WHERE id = 1",
		"UPDATE t SET x=1 WHERE id=1",
		"DELETE FROM T\tWHERE id=1",
		"DELETE FROM T\nWHERE id=1",
		"DELETE FROM T\rWHERE id=1",   // Windows CR
		"DELETE FROM T\r\nWHERE id=1", // Windows CRLF
		"DELETE FROM T WHERE",         // trailing WHERE
		"SELECT * FROM t WHERE 1=1",
	}
	for _, sql := range cases {
		upper := strings.ToUpper(sql)
		if !hasWhereClause(upper) {
			t.Errorf("hasWhereClause(%q) = false, want true", sql)
		}
	}
}

func TestHasWhereClause_absent(t *testing.T) {
	cases := []string{
		"DELETE FROM t",
		"UPDATE t SET x=1",
		"DELETE FROM t /* WHERE id=1 */", // WHERE is in a comment
		"SELECT * FROM t",
	}
	for _, sql := range cases {
		upper := strings.ToUpper(sql)
		if hasWhereClause(upper) {
			t.Errorf("hasWhereClause(%q) = true, want false", sql)
		}
	}
}

// --- hasMultipleStatements ---

func TestHasMultipleStatements(t *testing.T) {
	cases := []struct {
		sql  string
		want bool
	}{
		{"SELECT 1", false},
		{"SELECT 1;", false},         // trailing semicolon only
		{"SELECT 1; SELECT 2", true}, // two statements
		{"SELECT 'a;b'", false},      // semicolon inside string
		{`SELECT "a;b"`, false},      // semicolon inside double-quoted string
		{"INSERT INTO t VALUES ('x'); DROP TABLE t", true},
		{"SELECT 1;\n", false}, // trailing whitespace only
	}
	for _, c := range cases {
		got := hasMultipleStatements(c.sql)
		if got != c.want {
			t.Errorf("hasMultipleStatements(%q) = %v, want %v", c.sql, got, c.want)
		}
	}
}
