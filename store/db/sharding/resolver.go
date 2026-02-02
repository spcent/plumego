package sharding

import (
	"errors"
	"fmt"
	"sort"
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

func (r *ShardKeyResolver) resolveMultipleFromWhere(query string, parsed *ParsedSQL, rule *ShardingRule, args []any) ([]*ResolvedShard, error) {
	if parsed.WhereClause == "" {
		if rule.HasDefaultShard() {
			return []*ResolvedShard{{
				TableName:  parsed.TableName,
				ShardIndex: rule.DefaultShard,
				ShardKey:   nil,
				Rule:       rule,
			}}, nil
		}
		return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	argOffset := 0
	if parsed.Type == SQLTypeUpdate {
		argOffset = r.countSetPlaceholders(query)
	}

	type shardCondition struct {
		op               string
		placeholderCount int
		argStart         int
	}

	var shardConditions []shardCondition
	argIndex := 0
	for _, part := range splitConditions(parsed.WhereClause) {
		column, op, placeholders, ok := parseCondition(part)
		if !ok {
			continue
		}

		if strings.EqualFold(column, rule.ShardKeyColumn) {
			shardConditions = append(shardConditions, shardCondition{
				op:               op,
				placeholderCount: placeholders,
				argStart:         argIndex,
			})
		}

		argIndex += placeholders
	}

	if len(shardConditions) == 0 {
		if rule.HasDefaultShard() {
			return []*ResolvedShard{{
				TableName:  parsed.TableName,
				ShardIndex: rule.DefaultShard,
				ShardKey:   nil,
				Rule:       rule,
			}}, nil
		}
		return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	var (
		inValues   []any
		eqValue    any
		hasEqual   bool
		rangeStart any
		rangeEnd   any
		hasStart   bool
		hasEnd     bool
	)

	for _, cond := range shardConditions {
		switch strings.ToUpper(cond.op) {
		case "=":
			idx := argOffset + cond.argStart
			if idx >= len(args) {
				return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, idx+1, len(args))
			}
			eqValue = args[idx]
			hasEqual = true
		case "IN":
			if cond.placeholderCount == 0 {
				return nil, fmt.Errorf("%w: expected shard key values for IN clause", ErrInvalidArgumentCount)
			}
			start := argOffset + cond.argStart
			end := start + cond.placeholderCount
			if end > len(args) {
				return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, end, len(args))
			}
			inValues = append(inValues, args[start:end]...)
		case ">", ">=":
			idx := argOffset + cond.argStart
			if idx >= len(args) {
				return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, idx+1, len(args))
			}
			rangeStart = args[idx]
			hasStart = true
		case "<", "<=":
			idx := argOffset + cond.argStart
			if idx >= len(args) {
				return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, idx+1, len(args))
			}
			rangeEnd = args[idx]
			hasEnd = true
		}
	}

	if hasEqual {
		shardIndex, err := rule.Strategy.Shard(eqValue, rule.ShardCount)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate shard: %w", err)
		}
		return []*ResolvedShard{{
			TableName:  parsed.TableName,
			ShardIndex: shardIndex,
			ShardKey:   eqValue,
			Rule:       rule,
		}}, nil
	}

	if len(inValues) > 0 {
		unique := make(map[int]any)
		for _, value := range inValues {
			shardIndex, err := rule.Strategy.Shard(value, rule.ShardCount)
			if err != nil {
				return nil, fmt.Errorf("failed to calculate shard: %w", err)
			}
			if _, exists := unique[shardIndex]; !exists {
				unique[shardIndex] = value
			}
		}
		return buildResolvedShards(parsed.TableName, rule, unique), nil
	}

	if hasStart && hasEnd {
		shards, err := rule.Strategy.ShardRange(rangeStart, rangeEnd, rule.ShardCount)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate shard range: %w", err)
		}
		unique := make(map[int]any)
		for _, shard := range shards {
			unique[shard] = map[string]any{"start": rangeStart, "end": rangeEnd}
		}
		return buildResolvedShards(parsed.TableName, rule, unique), nil
	}

	// Open-ended range or unsupported condition: fall back to all shards.
	unique := make(map[int]any)
	for shard := 0; shard < rule.ShardCount; shard++ {
		unique[shard] = map[string]any{"start": rangeStart, "end": rangeEnd}
	}
	return buildResolvedShards(parsed.TableName, rule, unique), nil
}

func buildResolvedShards(table string, rule *ShardingRule, shards map[int]any) []*ResolvedShard {
	indices := make([]int, 0, len(shards))
	for idx := range shards {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	results := make([]*ResolvedShard, 0, len(indices))
	for _, idx := range indices {
		results = append(results, &ResolvedShard{
			TableName:  table,
			ShardIndex: idx,
			ShardKey:   shards[idx],
			Rule:       rule,
		})
	}

	return results
}

func parseCondition(part string) (string, string, int, bool) {
	part = strings.ReplaceAll(part, "`", "")
	part = strings.TrimSpace(part)
	if part == "" {
		return "", "", 0, false
	}

	placeholderCount := strings.Count(part, "?")

	upper := strings.ToUpper(part)
	if idx := strings.Index(upper, " IN "); idx > 0 {
		column := strings.TrimSpace(part[:idx])
		column = normalizeColumn(column)
		return column, "IN", placeholderCount, true
	}

	for _, op := range []string{">=", "<=", "!=", "<>", "=", ">", "<"} {
		if idx := strings.Index(part, op); idx > 0 {
			column := strings.TrimSpace(part[:idx])
			column = normalizeColumn(column)
			return column, op, placeholderCount, true
		}
	}

	return "", "", placeholderCount, false
}

func normalizeColumn(column string) string {
	column = strings.TrimSpace(column)
	if dotIdx := strings.LastIndex(column, "."); dotIdx > 0 {
		column = column[dotIdx+1:]
	}
	return column
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
