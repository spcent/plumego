package rw

import (
	"context"
	"testing"
)

func TestSQLTypePolicy(t *testing.T) {
	policy := NewSQLTypePolicy()

	tests := []struct {
		name  string
		query string
		want  bool // true = primary, false = replica
	}{
		// Write operations -> primary
		{"INSERT", "INSERT INTO users (name) VALUES ('alice')", true},
		{"UPDATE", "UPDATE users SET name = 'bob' WHERE id = 1", true},
		{"DELETE", "DELETE FROM users WHERE id = 1", true},
		{"CREATE", "CREATE TABLE test (id INT)", true},
		{"ALTER", "ALTER TABLE users ADD COLUMN age INT", true},
		{"DROP", "DROP TABLE test", true},
		{"TRUNCATE", "TRUNCATE TABLE users", true},
		{"REPLACE", "REPLACE INTO users VALUES (1, 'alice')", true},

		// Read operations -> replica
		{"SELECT", "SELECT * FROM users", false},
		{"SELECT with WHERE", "SELECT * FROM users WHERE id = 1", false},
		{"SELECT COUNT", "SELECT COUNT(*) FROM users", false},

		// Special cases -> primary
		{"SELECT FOR UPDATE", "SELECT * FROM users WHERE id = 1 FOR UPDATE", true},
		{"SELECT with FOR UPDATE uppercase", "SELECT * FROM USERS FOR UPDATE", true},

		// Other read operations -> replica
		{"SHOW", "SHOW TABLES", false},
		{"DESCRIBE", "DESCRIBE users", false},
		{"EXPLAIN", "EXPLAIN SELECT * FROM users", false},

		// Case insensitive
		{"lowercase insert", "insert into users (name) values ('test')", true},
		{"mixed case", "SeLeCt * FrOm users", false},

		// With leading whitespace
		{"whitespace", "  SELECT * FROM users", false},
		{"tab", "\tINSERT INTO users VALUES (1)", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			got := policy.ShouldUsePrimary(ctx, tt.query)
			if got != tt.want {
				t.Errorf("ShouldUsePrimary(%q) = %v, want %v", tt.query, got, tt.want)
			}
		})
	}
}

func TestContextHints(t *testing.T) {
	policy := NewSQLTypePolicy()

	t.Run("ForcePrimary", func(t *testing.T) {
		ctx := ForcePrimary(context.Background())

		// Even SELECT should go to primary
		if !policy.ShouldUsePrimary(ctx, "SELECT * FROM users") {
			t.Error("expected primary for SELECT with ForcePrimary context")
		}
	})

	t.Run("PreferReplica", func(t *testing.T) {
		ctx := PreferReplica(context.Background())

		// Even write should try replica (though this is unusual)
		if policy.ShouldUsePrimary(ctx, "SELECT * FROM users") {
			t.Error("expected replica for SELECT with PreferReplica context")
		}
	})

	t.Run("No hint", func(t *testing.T) {
		ctx := context.Background()

		// Normal behavior
		if policy.ShouldUsePrimary(ctx, "SELECT * FROM users") {
			t.Error("expected replica for normal SELECT")
		}

		if !policy.ShouldUsePrimary(ctx, "INSERT INTO users VALUES (1)") {
			t.Error("expected primary for normal INSERT")
		}
	})
}

func TestTransactionAwarePolicy(t *testing.T) {
	base := NewSQLTypePolicy()
	policy := NewTransactionAwarePolicy(base)

	t.Run("In transaction", func(t *testing.T) {
		ctx := MarkInTransaction(context.Background())

		// All queries should go to primary
		if !policy.ShouldUsePrimary(ctx, "SELECT * FROM users") {
			t.Error("expected primary for SELECT in transaction")
		}

		if !policy.ShouldUsePrimary(ctx, "INSERT INTO users VALUES (1)") {
			t.Error("expected primary for INSERT in transaction")
		}
	})

	t.Run("Not in transaction", func(t *testing.T) {
		ctx := context.Background()

		// Should use base policy
		if policy.ShouldUsePrimary(ctx, "SELECT * FROM users") {
			t.Error("expected replica for SELECT outside transaction")
		}

		if !policy.ShouldUsePrimary(ctx, "INSERT INTO users VALUES (1)") {
			t.Error("expected primary for INSERT outside transaction")
		}
	})
}

func TestCompositePolicy(t *testing.T) {
	policy1 := NewSQLTypePolicy()
	policy2 := NewAlwaysPrimaryPolicy()

	composite := NewCompositePolicy(policy1, policy2)

	// Should return true if ANY policy returns true
	ctx := context.Background()

	// AlwaysPrimaryPolicy always returns true
	if !composite.ShouldUsePrimary(ctx, "SELECT * FROM users") {
		t.Error("expected primary because AlwaysPrimaryPolicy is included")
	}
}

func TestAlwaysPrimaryPolicy(t *testing.T) {
	policy := NewAlwaysPrimaryPolicy()

	tests := []string{
		"SELECT * FROM users",
		"INSERT INTO users VALUES (1)",
		"UPDATE users SET name = 'test'",
		"DELETE FROM users",
		"SHOW TABLES",
	}

	ctx := context.Background()
	for _, query := range tests {
		if !policy.ShouldUsePrimary(ctx, query) {
			t.Errorf("AlwaysPrimaryPolicy should return true for %q", query)
		}
	}
}

func TestIsWriteOperation(t *testing.T) {
	tests := []struct {
		query string
		want  bool
	}{
		{"", false},
		{"   ", false},
		{"SELECT * FROM users", false},
		{"INSERT INTO users VALUES (1)", true},
		{"UPDATE users SET name = 'test'", true},
		{"DELETE FROM users WHERE id = 1", true},
		{"CREATE TABLE test (id INT)", true},
		{"ALTER TABLE users ADD COLUMN age INT", true},
		{"DROP TABLE test", true},
		{"TRUNCATE TABLE users", true},
		{"REPLACE INTO users VALUES (1)", true},
		{"SELECT * FROM users FOR UPDATE", true},
		{"  select * from users  ", false},
		{"  INSERT  INTO test", true},
	}

	for _, tt := range tests {
		got := isWriteOperation(tt.query)
		if got != tt.want {
			t.Errorf("isWriteOperation(%q) = %v, want %v", tt.query, got, tt.want)
		}
	}
}

func TestPolicyNames(t *testing.T) {
	tests := []struct {
		name   string
		policy RoutingPolicy
		want   string
	}{
		{"SQLType", NewSQLTypePolicy(), "sql_type"},
		{"AlwaysPrimary", NewAlwaysPrimaryPolicy(), "always_primary"},
		{"TransactionAware", NewTransactionAwarePolicy(NewSQLTypePolicy()), "transaction_aware(sql_type)"},
		{"Composite", NewCompositePolicy(NewSQLTypePolicy(), NewAlwaysPrimaryPolicy()), "composite(sql_type,always_primary)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.policy.Name()
			if got != tt.want {
				t.Errorf("Name() = %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkSQLTypePolicy(b *testing.B) {
	policy := NewSQLTypePolicy()
	ctx := context.Background()
	queries := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO users (name) VALUES ('test')",
		"UPDATE users SET name = 'updated' WHERE id = 1",
		"DELETE FROM users WHERE id = 1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		policy.ShouldUsePrimary(ctx, query)
	}
}

func BenchmarkIsWriteOperation(b *testing.B) {
	queries := []string{
		"SELECT * FROM users WHERE id = 1",
		"INSERT INTO users (name) VALUES ('test')",
		"UPDATE users SET name = 'updated'",
		"DELETE FROM users WHERE id = 1",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := queries[i%len(queries)]
		isWriteOperation(query)
	}
}
