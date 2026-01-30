package sharding

import (
	"errors"
	"fmt"
	"strings"
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

// extractFromInsert extracts shard key from INSERT statement
func (r *ShardKeyResolver) extractFromInsert(parsed *ParsedSQL, rule *ShardingRule, args []any) (any, error) {
	// Validate that number of args matches number of columns
	if len(args) != len(parsed.Columns) {
		return nil, fmt.Errorf("%w: expected %d args to match %d columns, got %d",
			ErrInvalidArgumentCount, len(parsed.Columns), len(parsed.Columns), len(args))
	}

	// Find the column index for the shard key
	shardKeyIdx := -1
	for i, col := range parsed.Columns {
		if col == rule.ShardKeyColumn {
			shardKeyIdx = i
			break
		}
	}

	if shardKeyIdx == -1 {
		return nil, fmt.Errorf("%w: column %s not found in INSERT", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	// Check if we have enough arguments (should always be true after the check above)
	if shardKeyIdx >= len(args) {
		return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrShardKeyIndexOutOfBounds, shardKeyIdx+1, len(args))
	}

	return args[shardKeyIdx], nil
}

// extractFromWhere extracts shard key from WHERE clause
func (r *ShardKeyResolver) extractFromWhere(query string, parsed *ParsedSQL, rule *ShardingRule, args []any) (any, error) {
	// Check if the shard key column is in the WHERE conditions
	op, found := parsed.GetCondition(rule.ShardKeyColumn)
	if !found {
		return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	// For now, we only support equality (=) conditions for shard key extraction
	// IN, >, <, etc. would require more complex handling
	if op != "=" {
		return nil, fmt.Errorf("%w: shard key must use = operator, got %s", ErrShardKeyNotFound, op)
	}

	// For UPDATE statements, we need to offset by the number of SET placeholders
	argOffset := 0
	if parsed.Type == SQLTypeUpdate {
		argOffset = r.countSetPlaceholders(query)
	}

	// Find the position of the shard key in the WHERE clause
	// Count placeholders before the shard key column
	argIndex := r.findArgumentIndex(parsed, rule.ShardKeyColumn)
	if argIndex == -1 {
		return nil, fmt.Errorf("%w: could not determine argument position for %s", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	// Add offset for UPDATE statements
	finalIndex := argOffset + argIndex

	if finalIndex >= len(args) {
		return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, finalIndex+1, len(args))
	}

	return args[finalIndex], nil
}

// findArgumentIndex finds the argument index for a given column in the WHERE clause
// This is a simplified version that counts the position of the column in Conditions map
func (r *ShardKeyResolver) findArgumentIndex(parsed *ParsedSQL, shardKeyColumn string) int {
	// In a simplified approach, we assume arguments appear in the order they're parsed
	// This works for simple queries like: WHERE user_id = ? AND status = ?
	idx := 0
	for col := range parsed.Conditions {
		if col == shardKeyColumn {
			return idx
		}
		idx++
	}
	return -1
}

// countSetPlaceholders counts the number of placeholders in UPDATE SET clause
func (r *ShardKeyResolver) countSetPlaceholders(query string) int {
	// Find the SET and WHERE keywords
	queryUpper := strings.ToUpper(query)
	setIdx := strings.Index(queryUpper, " SET ")
	whereIdx := strings.Index(queryUpper, " WHERE ")

	if setIdx == -1 {
		return 0
	}

	// Extract the SET clause
	var setClause string
	if whereIdx > setIdx {
		setClause = query[setIdx+5 : whereIdx]
	} else {
		// No WHERE clause, count to end of query
		setClause = query[setIdx+5:]
	}

	// Count '?' placeholders in SET clause
	return strings.Count(setClause, "?")
}

// ResolveMultiple resolves multiple possible shards (for range queries, IN clauses, etc.)
func (r *ShardKeyResolver) ResolveMultiple(query string, args []any) ([]*ResolvedShard, error) {
	// For now, just return single resolution
	// TODO: Implement proper multi-shard resolution for IN clauses and range queries
	resolved, err := r.Resolve(query, args)
	if err != nil {
		return nil, err
	}
	return []*ResolvedShard{resolved}, nil
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
