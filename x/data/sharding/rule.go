package sharding

import (
	"errors"
	"fmt"
)

var (
	// ErrNoShardingRule is returned when no sharding rule is configured for a table
	ErrNoShardingRule = errors.New("no sharding rule configured for table")

	// ErrInvalidShardingRule is returned when the sharding rule is invalid
	ErrInvalidShardingRule = errors.New("invalid sharding rule")

	// ErrInvalidShardCount is returned when the shard count is invalid
	ErrInvalidShardCount = errors.New("shard count must be greater than 0")
)

// ShardingRule defines the sharding configuration for a table
type ShardingRule struct {
	// TableName is the logical table name
	TableName string

	// ShardKeyColumn is the column name used for sharding
	ShardKeyColumn string

	// Strategy is the sharding strategy to use
	Strategy Strategy

	// ShardCount is the number of shards for this table
	ShardCount int

	// ActualTableNames maps shard index to actual physical table name
	// If empty, uses logical table name
	// Example: {0: "users_0", 1: "users_1"}
	ActualTableNames map[int]string

	// DefaultShard specifies which shard to use when shard key is not present
	// -1 means error if shard key is missing (default behavior)
	DefaultShard int
}

// NewShardingRule creates a new sharding rule with validation
func NewShardingRule(tableName, shardKeyColumn string, strategy Strategy, shardCount int) (*ShardingRule, error) {
	if tableName == "" {
		return nil, fmt.Errorf("%w: table name is required", ErrInvalidShardingRule)
	}

	if shardKeyColumn == "" {
		return nil, fmt.Errorf("%w: shard key column is required", ErrInvalidShardingRule)
	}

	if strategy == nil {
		return nil, fmt.Errorf("%w: strategy is required", ErrInvalidShardingRule)
	}

	if shardCount <= 0 {
		return nil, ErrInvalidShardCount
	}

	return &ShardingRule{
		TableName:        tableName,
		ShardKeyColumn:   shardKeyColumn,
		Strategy:         strategy,
		ShardCount:       shardCount,
		ActualTableNames: make(map[int]string),
		DefaultShard:     -1, // No default shard
	}, nil
}

// GetActualTableName returns the actual physical table name for a shard index
func (r *ShardingRule) GetActualTableName(shardIndex int) string {
	if actualName, ok := r.ActualTableNames[shardIndex]; ok {
		return actualName
	}
	// Default: return logical table name
	return r.TableName
}

// SetActualTableName sets the physical table name for a shard index
func (r *ShardingRule) SetActualTableName(shardIndex int, tableName string) {
	if r.ActualTableNames == nil {
		r.ActualTableNames = make(map[int]string)
	}
	r.ActualTableNames[shardIndex] = tableName
}

// SetDefaultShard sets the default shard index to use when shard key is missing
func (r *ShardingRule) SetDefaultShard(shardIndex int) error {
	if shardIndex < -1 || shardIndex >= r.ShardCount {
		return fmt.Errorf("invalid default shard index: %d (must be -1 or 0-%d)", shardIndex, r.ShardCount-1)
	}
	r.DefaultShard = shardIndex
	return nil
}

// HasDefaultShard returns true if a default shard is configured
func (r *ShardingRule) HasDefaultShard() bool {
	return r.DefaultShard >= 0
}

// Validate validates the sharding rule
func (r *ShardingRule) Validate() error {
	if r.TableName == "" {
		return fmt.Errorf("%w: table name is required", ErrInvalidShardingRule)
	}

	if r.ShardKeyColumn == "" {
		return fmt.Errorf("%w: shard key column is required", ErrInvalidShardingRule)
	}

	if r.Strategy == nil {
		return fmt.Errorf("%w: strategy is required", ErrInvalidShardingRule)
	}

	if r.ShardCount <= 0 {
		return ErrInvalidShardCount
	}

	// Validate default shard if set
	if r.DefaultShard >= r.ShardCount {
		return fmt.Errorf("default shard index %d exceeds shard count %d", r.DefaultShard, r.ShardCount)
	}

	return nil
}

// String returns a string representation of the sharding rule
func (r *ShardingRule) String() string {
	return fmt.Sprintf("ShardingRule{Table=%s, ShardKey=%s, Strategy=%s, ShardCount=%d, DefaultShard=%d}",
		r.TableName, r.ShardKeyColumn, r.Strategy.Name(), r.ShardCount, r.DefaultShard)
}

// ShardingRuleRegistry manages sharding rules for multiple tables
type ShardingRuleRegistry struct {
	rules map[string]*ShardingRule // table name -> rule
}

// NewShardingRuleRegistry creates a new sharding rule registry
func NewShardingRuleRegistry() *ShardingRuleRegistry {
	return &ShardingRuleRegistry{
		rules: make(map[string]*ShardingRule),
	}
}

// Register registers a sharding rule for a table
func (reg *ShardingRuleRegistry) Register(rule *ShardingRule) error {
	if err := rule.Validate(); err != nil {
		return err
	}

	reg.rules[rule.TableName] = rule
	return nil
}

// RegisterWithStrategy is a convenience method to register a rule with a strategy
func (reg *ShardingRuleRegistry) RegisterWithStrategy(
	tableName, shardKeyColumn string,
	strategy Strategy,
	shardCount int,
) error {
	rule, err := NewShardingRule(tableName, shardKeyColumn, strategy, shardCount)
	if err != nil {
		return err
	}
	return reg.Register(rule)
}

// Get retrieves the sharding rule for a table
func (reg *ShardingRuleRegistry) Get(tableName string) (*ShardingRule, error) {
	rule, ok := reg.rules[tableName]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrNoShardingRule, tableName)
	}
	return rule, nil
}

// Has checks if a sharding rule exists for a table
func (reg *ShardingRuleRegistry) Has(tableName string) bool {
	_, ok := reg.rules[tableName]
	return ok
}

// Remove removes the sharding rule for a table
func (reg *ShardingRuleRegistry) Remove(tableName string) {
	delete(reg.rules, tableName)
}

// GetAll returns all registered sharding rules
func (reg *ShardingRuleRegistry) GetAll() map[string]*ShardingRule {
	// Return a copy to prevent external modification
	result := make(map[string]*ShardingRule, len(reg.rules))
	for k, v := range reg.rules {
		result[k] = v
	}
	return result
}

// Count returns the number of registered rules
func (reg *ShardingRuleRegistry) Count() int {
	return len(reg.rules)
}

// Clear removes all sharding rules
func (reg *ShardingRuleRegistry) Clear() {
	reg.rules = make(map[string]*ShardingRule)
}

// ValidateAll validates all registered rules
func (reg *ShardingRuleRegistry) ValidateAll() error {
	for tableName, rule := range reg.rules {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("validation failed for table %s: %w", tableName, err)
		}
	}
	return nil
}
