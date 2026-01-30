package sharding

import (
	"testing"
)

func TestSQLParser_Parse_Select(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name          string
		sql           string
		wantType      SQLType
		wantTable     string
		wantWhere     string
		wantCondCount int
		wantErr       bool
	}{
		{
			name:          "simple select",
			sql:           "SELECT * FROM users WHERE id = ?",
			wantType:      SQLTypeSelect,
			wantTable:     "users",
			wantWhere:     "id = ?",
			wantCondCount: 1,
			wantErr:       false,
		},
		{
			name:          "select with backticks",
			sql:           "SELECT * FROM `users` WHERE `user_id` = ?",
			wantType:      SQLTypeSelect,
			wantTable:     "users",
			wantWhere:     "`user_id` = ?",
			wantCondCount: 1,
			wantErr:       false,
		},
		{
			name:          "select with multiple conditions",
			sql:           "SELECT * FROM orders WHERE user_id = ? AND status = ?",
			wantType:      SQLTypeSelect,
			wantTable:     "orders",
			wantWhere:     "user_id = ? AND status = ?",
			wantCondCount: 2,
			wantErr:       false,
		},
		{
			name:          "select with IN clause",
			sql:           "SELECT * FROM users WHERE id IN (?, ?, ?)",
			wantType:      SQLTypeSelect,
			wantTable:     "users",
			wantWhere:     "id IN (?, ?, ?)",
			wantCondCount: 1,
			wantErr:       false,
		},
		{
			name:          "select with ORDER BY",
			sql:           "SELECT * FROM users WHERE status = ? ORDER BY created_at",
			wantType:      SQLTypeSelect,
			wantTable:     "users",
			wantWhere:     "status = ?",
			wantCondCount: 1,
			wantErr:       false,
		},
		{
			name:          "select with LIMIT",
			sql:           "SELECT * FROM users WHERE user_id = ? LIMIT 10",
			wantType:      SQLTypeSelect,
			wantTable:     "users",
			wantWhere:     "user_id = ?",
			wantCondCount: 1,
			wantErr:       false,
		},
		{
			name:      "select without where",
			sql:       "SELECT * FROM users",
			wantType:  SQLTypeSelect,
			wantTable: "users",
			wantWhere: "",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", parsed.Type, tt.wantType)
			}

			if parsed.TableName != tt.wantTable {
				t.Errorf("Parse() TableName = %v, want %v", parsed.TableName, tt.wantTable)
			}

			if parsed.WhereClause != tt.wantWhere {
				t.Errorf("Parse() WhereClause = %q, want %q", parsed.WhereClause, tt.wantWhere)
			}

			if tt.wantCondCount > 0 && len(parsed.Conditions) != tt.wantCondCount {
				t.Errorf("Parse() Conditions count = %d, want %d", len(parsed.Conditions), tt.wantCondCount)
			}
		})
	}
}

func TestSQLParser_Parse_Insert(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name        string
		sql         string
		wantType    SQLType
		wantTable   string
		wantColumns []string
		wantValues  []string
		wantErr     bool
	}{
		{
			name:        "simple insert",
			sql:         "INSERT INTO users (id, name) VALUES (?, ?)",
			wantType:    SQLTypeInsert,
			wantTable:   "users",
			wantColumns: []string{"id", "name"},
			wantValues:  []string{"?", "?"},
			wantErr:     false,
		},
		{
			name:        "insert with backticks",
			sql:         "INSERT INTO `users` (`user_id`, `email`) VALUES (?, ?)",
			wantType:    SQLTypeInsert,
			wantTable:   "users",
			wantColumns: []string{"user_id", "email"},
			wantValues:  []string{"?", "?"},
			wantErr:     false,
		},
		{
			name:        "insert with multiple columns",
			sql:         "INSERT INTO orders (user_id, product_id, quantity, price) VALUES (?, ?, ?, ?)",
			wantType:    SQLTypeInsert,
			wantTable:   "orders",
			wantColumns: []string{"user_id", "product_id", "quantity", "price"},
			wantValues:  []string{"?", "?", "?", "?"},
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", parsed.Type, tt.wantType)
			}

			if parsed.TableName != tt.wantTable {
				t.Errorf("Parse() TableName = %v, want %v", parsed.TableName, tt.wantTable)
			}

			if len(parsed.Columns) != len(tt.wantColumns) {
				t.Errorf("Parse() Columns count = %d, want %d", len(parsed.Columns), len(tt.wantColumns))
			}

			for i, col := range tt.wantColumns {
				if i >= len(parsed.Columns) {
					break
				}
				if parsed.Columns[i] != col {
					t.Errorf("Parse() Column[%d] = %q, want %q", i, parsed.Columns[i], col)
				}
			}
		})
	}
}

func TestSQLParser_Parse_Update(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name      string
		sql       string
		wantType  SQLType
		wantTable string
		wantWhere string
		wantErr   bool
	}{
		{
			name:      "simple update",
			sql:       "UPDATE users SET name = ? WHERE id = ?",
			wantType:  SQLTypeUpdate,
			wantTable: "users",
			wantWhere: "id = ?",
			wantErr:   false,
		},
		{
			name:      "update with backticks",
			sql:       "UPDATE `users` SET `status` = ? WHERE `user_id` = ?",
			wantType:  SQLTypeUpdate,
			wantTable: "users",
			wantWhere: "`user_id` = ?",
			wantErr:   false,
		},
		{
			name:      "update with multiple conditions",
			sql:       "UPDATE orders SET status = ? WHERE user_id = ? AND order_id = ?",
			wantType:  SQLTypeUpdate,
			wantTable: "orders",
			wantWhere: "user_id = ? AND order_id = ?",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", parsed.Type, tt.wantType)
			}

			if parsed.TableName != tt.wantTable {
				t.Errorf("Parse() TableName = %v, want %v", parsed.TableName, tt.wantTable)
			}

			if parsed.WhereClause != tt.wantWhere {
				t.Errorf("Parse() WhereClause = %q, want %q", parsed.WhereClause, tt.wantWhere)
			}
		})
	}
}

func TestSQLParser_Parse_Delete(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name      string
		sql       string
		wantType  SQLType
		wantTable string
		wantWhere string
		wantErr   bool
	}{
		{
			name:      "simple delete",
			sql:       "DELETE FROM users WHERE id = ?",
			wantType:  SQLTypeDelete,
			wantTable: "users",
			wantWhere: "id = ?",
			wantErr:   false,
		},
		{
			name:      "delete with backticks",
			sql:       "DELETE FROM `users` WHERE `user_id` = ?",
			wantType:  SQLTypeDelete,
			wantTable: "users",
			wantWhere: "`user_id` = ?",
			wantErr:   false,
		},
		{
			name:      "delete with multiple conditions",
			sql:       "DELETE FROM orders WHERE user_id = ? AND status = ?",
			wantType:  SQLTypeDelete,
			wantTable: "orders",
			wantWhere: "user_id = ? AND status = ?",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := parser.Parse(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Parse() unexpected error: %v", err)
			}

			if parsed.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", parsed.Type, tt.wantType)
			}

			if parsed.TableName != tt.wantTable {
				t.Errorf("Parse() TableName = %v, want %v", parsed.TableName, tt.wantTable)
			}

			if parsed.WhereClause != tt.wantWhere {
				t.Errorf("Parse() WhereClause = %q, want %q", parsed.WhereClause, tt.wantWhere)
			}
		})
	}
}

func TestSQLParser_Parse_Errors(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name    string
		sql     string
		wantErr error
	}{
		{
			name:    "empty sql",
			sql:     "",
			wantErr: ErrUnsupportedSQL,
		},
		{
			name:    "unsupported statement",
			sql:     "CREATE TABLE users (id INT)",
			wantErr: ErrUnsupportedSQL,
		},
		{
			name:    "invalid select",
			sql:     "SELECT * FROM",
			wantErr: ErrUnsupportedSQL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse(tt.sql)

			if err == nil {
				t.Errorf("Parse() expected error, got nil")
				return
			}

			if tt.wantErr != nil && err != tt.wantErr {
				t.Errorf("Parse() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestParsedSQL_Methods(t *testing.T) {
	parser := NewSQLParser()

	t.Run("IsWrite", func(t *testing.T) {
		writeSQLs := []string{
			"INSERT INTO users (id) VALUES (?)",
			"UPDATE users SET name = ?",
			"DELETE FROM users WHERE id = ?",
		}

		for _, sql := range writeSQLs {
			parsed, err := parser.Parse(sql)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if !parsed.IsWrite() {
				t.Errorf("IsWrite() = false for %s, want true", sql)
			}
		}
	})

	t.Run("IsRead", func(t *testing.T) {
		parsed, err := parser.Parse("SELECT * FROM users")
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if !parsed.IsRead() {
			t.Errorf("IsRead() = false, want true")
		}
	})

	t.Run("HasWhere", func(t *testing.T) {
		withWhere, _ := parser.Parse("SELECT * FROM users WHERE id = ?")
		withoutWhere, _ := parser.Parse("SELECT * FROM users")

		if !withWhere.HasWhere() {
			t.Errorf("HasWhere() = false for query with WHERE, want true")
		}

		if withoutWhere.HasWhere() {
			t.Errorf("HasWhere() = true for query without WHERE, want false")
		}
	})

	t.Run("GetCondition", func(t *testing.T) {
		parsed, _ := parser.Parse("SELECT * FROM users WHERE user_id = ? AND status IN (?)")

		op, ok := parsed.GetCondition("user_id")
		if !ok {
			t.Errorf("GetCondition(user_id) not found")
		}
		if op != "=" {
			t.Errorf("GetCondition(user_id) = %q, want =", op)
		}

		op, ok = parsed.GetCondition("status")
		if !ok {
			t.Errorf("GetCondition(status) not found")
		}
		if op != "IN" {
			t.Errorf("GetCondition(status) = %q, want IN", op)
		}

		_, ok = parsed.GetCondition("nonexistent")
		if ok {
			t.Errorf("GetCondition(nonexistent) found, want not found")
		}
	})
}

func TestSQLParser_ExtractTableName(t *testing.T) {
	parser := NewSQLParser()

	tests := []struct {
		name      string
		sql       string
		wantTable string
		wantErr   bool
	}{
		{
			name:      "select",
			sql:       "SELECT * FROM users",
			wantTable: "users",
			wantErr:   false,
		},
		{
			name:      "insert",
			sql:       "INSERT INTO orders (id) VALUES (?)",
			wantTable: "orders",
			wantErr:   false,
		},
		{
			name:      "update",
			sql:       "UPDATE products SET price = ?",
			wantTable: "products",
			wantErr:   false,
		},
		{
			name:      "delete",
			sql:       "DELETE FROM logs WHERE id = ?",
			wantTable: "logs",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tableName, err := parser.ExtractTableName(tt.sql)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractTableName() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ExtractTableName() unexpected error: %v", err)
			}

			if tableName != tt.wantTable {
				t.Errorf("ExtractTableName() = %q, want %q", tableName, tt.wantTable)
			}
		})
	}
}

func TestSQLType_String(t *testing.T) {
	tests := []struct {
		sqlType SQLType
		want    string
	}{
		{SQLTypeSelect, "SELECT"},
		{SQLTypeInsert, "INSERT"},
		{SQLTypeUpdate, "UPDATE"},
		{SQLTypeDelete, "DELETE"},
		{SQLTypeUnknown, "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.sqlType.String(); got != tt.want {
				t.Errorf("SQLType.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
