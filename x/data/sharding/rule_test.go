package sharding

import (
	"errors"
	"testing"
)

func TestNewShardingRule(t *testing.T) {
	strategy := NewModStrategy()

	tests := []struct {
		name           string
		tableName      string
		shardKeyColumn string
		strategy       Strategy
		shardCount     int
		wantErr        bool
		expectedErr    error
	}{
		{
			name:           "valid rule",
			tableName:      "users",
			shardKeyColumn: "user_id",
			strategy:       strategy,
			shardCount:     4,
			wantErr:        false,
		},
		{
			name:           "empty table name",
			tableName:      "",
			shardKeyColumn: "user_id",
			strategy:       strategy,
			shardCount:     4,
			wantErr:        true,
			expectedErr:    ErrInvalidShardingRule,
		},
		{
			name:           "empty shard key column",
			tableName:      "users",
			shardKeyColumn: "",
			strategy:       strategy,
			shardCount:     4,
			wantErr:        true,
			expectedErr:    ErrInvalidShardingRule,
		},
		{
			name:           "nil strategy",
			tableName:      "users",
			shardKeyColumn: "user_id",
			strategy:       nil,
			shardCount:     4,
			wantErr:        true,
			expectedErr:    ErrInvalidShardingRule,
		},
		{
			name:           "zero shard count",
			tableName:      "users",
			shardKeyColumn: "user_id",
			strategy:       strategy,
			shardCount:     0,
			wantErr:        true,
			expectedErr:    ErrInvalidShardCount,
		},
		{
			name:           "negative shard count",
			tableName:      "users",
			shardKeyColumn: "user_id",
			strategy:       strategy,
			shardCount:     -1,
			wantErr:        true,
			expectedErr:    ErrInvalidShardCount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule, err := NewShardingRule(tt.tableName, tt.shardKeyColumn, tt.strategy, tt.shardCount)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewShardingRule() expected error, got nil")
					return
				}
				if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
					t.Errorf("NewShardingRule() error = %v, want %v", err, tt.expectedErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewShardingRule() unexpected error: %v", err)
			}

			if rule.TableName != tt.tableName {
				t.Errorf("TableName = %q, want %q", rule.TableName, tt.tableName)
			}

			if rule.ShardKeyColumn != tt.shardKeyColumn {
				t.Errorf("ShardKeyColumn = %q, want %q", rule.ShardKeyColumn, tt.shardKeyColumn)
			}

			if rule.ShardCount != tt.shardCount {
				t.Errorf("ShardCount = %d, want %d", rule.ShardCount, tt.shardCount)
			}

			if rule.DefaultShard != -1 {
				t.Errorf("DefaultShard = %d, want -1", rule.DefaultShard)
			}
		})
	}
}

func TestShardingRule_GetActualTableName(t *testing.T) {
	rule, _ := NewShardingRule("users", "user_id", NewModStrategy(), 4)

	t.Run("without actual table names", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			actualName := rule.GetActualTableName(i)
			if actualName != "users" {
				t.Errorf("GetActualTableName(%d) = %q, want %q", i, actualName, "users")
			}
		}
	})

	t.Run("with actual table names", func(t *testing.T) {
		rule.SetActualTableName(0, "users_0")
		rule.SetActualTableName(1, "users_1")
		rule.SetActualTableName(2, "users_2")
		rule.SetActualTableName(3, "users_3")

		for i := 0; i < 4; i++ {
			expected := "users_" + string(rune('0'+i))
			actualName := rule.GetActualTableName(i)
			if actualName != expected {
				t.Errorf("GetActualTableName(%d) = %q, want %q", i, actualName, expected)
			}
		}
	})
}

func TestShardingRule_SetDefaultShard(t *testing.T) {
	rule, _ := NewShardingRule("users", "user_id", NewModStrategy(), 4)

	tests := []struct {
		name           string
		shardIndex     int
		wantErr        bool
		wantHasDefault bool
	}{
		{
			name:           "valid default shard",
			shardIndex:     0,
			wantErr:        false,
			wantHasDefault: true,
		},
		{
			name:           "valid last shard",
			shardIndex:     3,
			wantErr:        false,
			wantHasDefault: true,
		},
		{
			name:           "disable default shard",
			shardIndex:     -1,
			wantErr:        false,
			wantHasDefault: false,
		},
		{
			name:           "invalid negative",
			shardIndex:     -2,
			wantErr:        true,
			wantHasDefault: false,
		},
		{
			name:           "out of bounds",
			shardIndex:     4,
			wantErr:        true,
			wantHasDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.SetDefaultShard(tt.shardIndex)

			if tt.wantErr {
				if err == nil {
					t.Errorf("SetDefaultShard(%d) expected error, got nil", tt.shardIndex)
				}
				return
			}

			if err != nil {
				t.Fatalf("SetDefaultShard(%d) unexpected error: %v", tt.shardIndex, err)
			}

			if rule.HasDefaultShard() != tt.wantHasDefault {
				t.Errorf("HasDefaultShard() = %v, want %v", rule.HasDefaultShard(), tt.wantHasDefault)
			}

			if !tt.wantErr && tt.wantHasDefault && rule.DefaultShard != tt.shardIndex {
				t.Errorf("DefaultShard = %d, want %d", rule.DefaultShard, tt.shardIndex)
			}
		})
	}
}

func TestShardingRule_Validate(t *testing.T) {
	strategy := NewModStrategy()

	tests := []struct {
		name    string
		rule    *ShardingRule
		wantErr bool
	}{
		{
			name: "valid rule",
			rule: &ShardingRule{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       strategy,
				ShardCount:     4,
				DefaultShard:   -1,
			},
			wantErr: false,
		},
		{
			name: "empty table name",
			rule: &ShardingRule{
				TableName:      "",
				ShardKeyColumn: "user_id",
				Strategy:       strategy,
				ShardCount:     4,
			},
			wantErr: true,
		},
		{
			name: "empty shard key",
			rule: &ShardingRule{
				TableName:      "users",
				ShardKeyColumn: "",
				Strategy:       strategy,
				ShardCount:     4,
			},
			wantErr: true,
		},
		{
			name: "nil strategy",
			rule: &ShardingRule{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       nil,
				ShardCount:     4,
			},
			wantErr: true,
		},
		{
			name: "invalid shard count",
			rule: &ShardingRule{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       strategy,
				ShardCount:     0,
			},
			wantErr: true,
		},
		{
			name: "invalid default shard",
			rule: &ShardingRule{
				TableName:      "users",
				ShardKeyColumn: "user_id",
				Strategy:       strategy,
				ShardCount:     4,
				DefaultShard:   10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.rule.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestShardingRuleRegistry_Register(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	t.Run("register valid rule", func(t *testing.T) {
		rule, _ := NewShardingRule("users", "user_id", strategy, 4)
		err := registry.Register(rule)
		if err != nil {
			t.Errorf("Register() unexpected error: %v", err)
		}

		if registry.Count() != 1 {
			t.Errorf("Count() = %d, want 1", registry.Count())
		}
	})

	t.Run("register invalid rule", func(t *testing.T) {
		invalidRule := &ShardingRule{
			TableName:      "",
			ShardKeyColumn: "user_id",
			Strategy:       strategy,
			ShardCount:     4,
		}

		err := registry.Register(invalidRule)
		if err == nil {
			t.Errorf("Register() expected error for invalid rule, got nil")
		}
	})

	t.Run("overwrite existing rule", func(t *testing.T) {
		rule1, _ := NewShardingRule("orders", "order_id", strategy, 4)
		rule2, _ := NewShardingRule("orders", "user_id", strategy, 8)

		registry.Register(rule1)
		initialCount := registry.Count()

		registry.Register(rule2)
		if registry.Count() != initialCount {
			t.Errorf("Count() = %d, want %d", registry.Count(), initialCount)
		}

		retrieved, _ := registry.Get("orders")
		if retrieved.ShardKeyColumn != "user_id" {
			t.Errorf("ShardKeyColumn = %q, want user_id", retrieved.ShardKeyColumn)
		}
	})
}

func TestShardingRuleRegistry_Get(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	rule, _ := NewShardingRule("users", "user_id", strategy, 4)
	registry.Register(rule)

	t.Run("get existing rule", func(t *testing.T) {
		retrieved, err := registry.Get("users")
		if err != nil {
			t.Errorf("Get() unexpected error: %v", err)
		}

		if retrieved.TableName != "users" {
			t.Errorf("TableName = %q, want users", retrieved.TableName)
		}
	})

	t.Run("get non-existing rule", func(t *testing.T) {
		_, err := registry.Get("nonexistent")
		if err == nil {
			t.Errorf("Get() expected error for non-existing table, got nil")
		}

		if !errors.Is(err, ErrNoShardingRule) {
			t.Errorf("Get() error = %v, want %v", err, ErrNoShardingRule)
		}
	})
}

func TestShardingRuleRegistry_Has(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	rule, _ := NewShardingRule("users", "user_id", strategy, 4)
	registry.Register(rule)

	if !registry.Has("users") {
		t.Errorf("Has(users) = false, want true")
	}

	if registry.Has("nonexistent") {
		t.Errorf("Has(nonexistent) = true, want false")
	}
}

func TestShardingRuleRegistry_Remove(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	rule, _ := NewShardingRule("users", "user_id", strategy, 4)
	registry.Register(rule)

	initialCount := registry.Count()
	if initialCount != 1 {
		t.Fatalf("Initial count = %d, want 1", initialCount)
	}

	registry.Remove("users")

	if registry.Count() != 0 {
		t.Errorf("Count() after Remove = %d, want 0", registry.Count())
	}

	if registry.Has("users") {
		t.Errorf("Has(users) = true after Remove, want false")
	}
}

func TestShardingRuleRegistry_GetAll(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	rule1, _ := NewShardingRule("users", "user_id", strategy, 4)
	rule2, _ := NewShardingRule("orders", "order_id", strategy, 8)

	registry.Register(rule1)
	registry.Register(rule2)

	all := registry.GetAll()

	if len(all) != 2 {
		t.Errorf("GetAll() count = %d, want 2", len(all))
	}

	if _, ok := all["users"]; !ok {
		t.Errorf("GetAll() missing users")
	}

	if _, ok := all["orders"]; !ok {
		t.Errorf("GetAll() missing orders")
	}

	// Verify it's a copy by modifying it
	delete(all, "users")
	if !registry.Has("users") {
		t.Errorf("Modifying GetAll() result affected registry")
	}
}

func TestShardingRuleRegistry_Clear(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	rule1, _ := NewShardingRule("users", "user_id", strategy, 4)
	rule2, _ := NewShardingRule("orders", "order_id", strategy, 8)

	registry.Register(rule1)
	registry.Register(rule2)

	if registry.Count() != 2 {
		t.Fatalf("Count() before Clear = %d, want 2", registry.Count())
	}

	registry.Clear()

	if registry.Count() != 0 {
		t.Errorf("Count() after Clear = %d, want 0", registry.Count())
	}
}

func TestShardingRuleRegistry_ValidateAll(t *testing.T) {
	strategy := NewModStrategy()

	t.Run("all valid", func(t *testing.T) {
		registry := NewShardingRuleRegistry()

		rule1, _ := NewShardingRule("users", "user_id", strategy, 4)
		rule2, _ := NewShardingRule("orders", "order_id", strategy, 8)

		registry.Register(rule1)
		registry.Register(rule2)

		err := registry.ValidateAll()
		if err != nil {
			t.Errorf("ValidateAll() unexpected error: %v", err)
		}
	})

	t.Run("with invalid rule", func(t *testing.T) {
		registry := NewShardingRuleRegistry()

		// Manually add invalid rule to bypass Register validation
		invalidRule := &ShardingRule{
			TableName:      "invalid",
			ShardKeyColumn: "",
			Strategy:       strategy,
			ShardCount:     4,
		}
		registry.rules["invalid"] = invalidRule

		err := registry.ValidateAll()
		if err == nil {
			t.Errorf("ValidateAll() expected error for invalid rule, got nil")
		}
	})
}

func TestShardingRuleRegistry_RegisterWithStrategy(t *testing.T) {
	registry := NewShardingRuleRegistry()
	strategy := NewModStrategy()

	err := registry.RegisterWithStrategy("users", "user_id", strategy, 4)
	if err != nil {
		t.Errorf("RegisterWithStrategy() unexpected error: %v", err)
	}

	rule, err := registry.Get("users")
	if err != nil {
		t.Fatalf("Get() unexpected error: %v", err)
	}

	if rule.TableName != "users" {
		t.Errorf("TableName = %q, want users", rule.TableName)
	}

	if rule.ShardKeyColumn != "user_id" {
		t.Errorf("ShardKeyColumn = %q, want user_id", rule.ShardKeyColumn)
	}

	if rule.ShardCount != 4 {
		t.Errorf("ShardCount = %d, want 4", rule.ShardCount)
	}
}
