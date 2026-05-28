package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	plumelog "github.com/spcent/plumego/log"

	"dbadmin/internal/domain/connection"
)

// TestClassifySQLWriteOps verifies that write-type SQL statements are correctly
// classified as non-SELECT so the readonly guard blocks them.
func TestClassifySQLWriteOps(t *testing.T) {
	writes := []string{
		"INSERT INTO users (name) VALUES ('x')",
		"UPDATE users SET name = 'y' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
		"DROP TABLE users",
		"TRUNCATE TABLE users",
		"ALTER TABLE users ADD COLUMN age INT",
		"CREATE TABLE new_table (id INT)",
		"REPLACE INTO users VALUES (1, 'x')",
	}
	for _, sql := range writes {
		cls := classifySQL(sql)
		if cls.IsSelect {
			t.Errorf("classifySQL(%q) IsSelect=true, want false", sql)
		}
	}
}

// TestClassifySQLReadOps verifies that read-type SQL statements are classified
// as IsSelect so the readonly guard allows them.
func TestClassifySQLReadOps(t *testing.T) {
	reads := []string{
		"SELECT * FROM users",
		"SHOW TABLES",
		"EXPLAIN SELECT 1",
		"DESCRIBE users",
		"DESC users",
		"PRAGMA table_info(users)",
		"WITH cte AS (SELECT 1) SELECT * FROM cte",
	}
	for _, sql := range reads {
		cls := classifySQL(sql)
		if !cls.IsSelect {
			t.Errorf("classifySQL(%q) IsSelect=false, want true", sql)
		}
	}
}

// TestGuardReadonly verifies that guardReadonly returns true and writes HTTP 403
// for a readonly connection, and returns false for a non-readonly connection.
func TestGuardReadonly(t *testing.T) {
	log := plumelog.NewLogger()

	t.Run("readonly connection → 403", func(t *testing.T) {
		conn := &connection.Connection{Readonly: true}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		blocked := guardReadonly(conn, w, r, log)
		if !blocked {
			t.Fatal("guardReadonly returned false for readonly connection, want true")
		}
		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", w.Code, http.StatusForbidden)
		}
		var body struct {
			Error struct {
				Details map[string]any `json:"details"`
			} `json:"error"`
		}
		if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Error.Details["code"] != "READONLY_VIOLATION" {
			t.Errorf("details.code = %v, want READONLY_VIOLATION", body.Error.Details["code"])
		}
	})

	t.Run("non-readonly connection → allowed", func(t *testing.T) {
		conn := &connection.Connection{Readonly: false}
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", nil)

		blocked := guardReadonly(conn, w, r, log)
		if blocked {
			t.Fatal("guardReadonly returned true for non-readonly connection, want false")
		}
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d (no response should have been written)", w.Code, http.StatusOK)
		}
	})
}
