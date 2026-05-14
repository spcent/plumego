package sharding

import (
	"errors"
	"fmt"
)

var (
	// ErrShardKeyNotFound is returned when the shard key cannot be extracted
	ErrShardKeyNotFound = errors.New("shard key not found in query")

	// ErrInvalidArgumentCount is returned when argument count doesn't match placeholders
	ErrInvalidArgumentCount = errors.New("invalid argument count")

	// ErrShardKeyIndexOutOfBounds is returned when shard key index is out of bounds
	ErrShardKeyIndexOutOfBounds = errors.New("shard key index out of bounds")
)

// ResolvedShard contains the resolved shard information
type ResolvedShard struct {
	// TableName is the logical table name
	TableName string

	// ShardIndex is the resolved shard index
	ShardIndex int

	// ShardKey is the extracted shard key value
	ShardKey any

	// Rule is the sharding rule used for resolution
	Rule *ShardingRule
}

// ShardKeyResolver extracts shard keys from SQL queries
type ShardKeyResolver struct {
	parser   *SQLParser
	registry *ShardingRuleRegistry
}

// NewShardKeyResolver creates a new shard key resolver
func NewShardKeyResolver(registry *ShardingRuleRegistry) *ShardKeyResolver {
	return &ShardKeyResolver{
		parser:   NewSQLParser(),
		registry: registry,
	}
}

// Resolve extracts the shard key from a query and resolves the shard index
func (r *ShardKeyResolver) Resolve(query string, args []any) (*ResolvedShard, error) {
	// Parse the SQL query
	parsed, err := r.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	// Get sharding rule for the table
	rule, err := r.registry.Get(parsed.TableName)
	if err != nil {
		return nil, err
	}

	// Extract shard key value based on SQL type
	var shardKeyValue any

	switch parsed.Type {
	case SQLTypeInsert:
		shardKeyValue, err = r.extractFromInsert(parsed, rule, args)
	case SQLTypeSelect, SQLTypeUpdate, SQLTypeDelete:
		shardKeyValue, err = r.extractFromWhere(query, parsed, rule, args)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedSQL, parsed.Type)
	}

	if err != nil {
		// If no shard key found and there's a default shard, use it
		if errors.Is(err, ErrShardKeyNotFound) && rule.HasDefaultShard() {
			return &ResolvedShard{
				TableName:  parsed.TableName,
				ShardIndex: rule.DefaultShard,
				ShardKey:   nil,
				Rule:       rule,
			}, nil
		}
		return nil, err
	}

	// Calculate shard index using the strategy
	shardIndex, err := rule.Strategy.Shard(shardKeyValue, rule.ShardCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate shard: %w", err)
	}

	return &ResolvedShard{
		TableName:  parsed.TableName,
		ShardIndex: shardIndex,
		ShardKey:   shardKeyValue,
		Rule:       rule,
	}, nil
}

// ResolveMultiple resolves multiple possible shards (for range queries, IN clauses, etc.)
func (r *ShardKeyResolver) ResolveMultiple(query string, args []any) ([]*ResolvedShard, error) {
	parsed, err := r.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	rule, err := r.registry.Get(parsed.TableName)
	if err != nil {
		return nil, err
	}

	switch parsed.Type {
	case SQLTypeInsert:
		resolved, err := r.Resolve(query, args)
		if err != nil {
			return nil, err
		}
		return []*ResolvedShard{resolved}, nil
	case SQLTypeSelect, SQLTypeUpdate, SQLTypeDelete:
		return r.resolveMultipleFromWhere(query, parsed, rule, args)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedSQL, parsed.Type)
	}
}

// ResolveTableOnly extracts just the table name without resolving shard
// Useful for queries that don't need shard resolution
func (r *ShardKeyResolver) ResolveTableOnly(query string) (string, error) {
	return r.parser.ExtractTableName(query)
}

// CanResolve checks if a query can be resolved (has shard key in WHERE/VALUES)
func (r *ShardKeyResolver) CanResolve(query string) (bool, error) {
	parsed, err := r.parser.Parse(query)
	if err != nil {
		return false, err
	}

	rule, err := r.registry.Get(parsed.TableName)
	if err != nil {
		return false, err
	}

	switch parsed.Type {
	case SQLTypeInsert:
		// Check if shard key column is in the INSERT columns
		for _, col := range parsed.Columns {
			if col == rule.ShardKeyColumn {
				return true, nil
			}
		}
		return rule.HasDefaultShard(), nil

	case SQLTypeSelect, SQLTypeUpdate, SQLTypeDelete:
		// Check if shard key is in WHERE clause
		_, found := parsed.GetCondition(rule.ShardKeyColumn)
		if found {
			return true, nil
		}
		return rule.HasDefaultShard(), nil

	default:
		return false, ErrUnsupportedSQL
	}
}

// String returns a string representation of the resolved shard
func (r *ResolvedShard) String() string {
	return fmt.Sprintf("ResolvedShard{Table=%s, ShardIndex=%d, ShardKey=%v}",
		r.TableName, r.ShardIndex, r.ShardKey)
}

// GetActualTableName returns the actual physical table name for this shard
func (r *ResolvedShard) GetActualTableName() string {
	return r.Rule.GetActualTableName(r.ShardIndex)
}
