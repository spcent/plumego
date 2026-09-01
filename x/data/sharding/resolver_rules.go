package sharding

import (
	"fmt"
	"sort"
	"strings"
)

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
	if strings.TrimSpace(parsed.WhereClause) == "" {
		return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	// For UPDATE statements, we need to offset by the number of SET placeholders
	argOffset := 0
	if parsed.Type == SQLTypeUpdate {
		argOffset = r.countSetPlaceholders(query)
	}

	op, argIndex, explicitIndex, ok := r.findShardKeyArgIndex(parsed.WhereClause, rule.ShardKeyColumn)
	if !ok {
		return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
	}

	// For now, we only support equality (=) conditions for shard key extraction
	// IN, >, <, etc. would require more complex handling
	if op != "=" {
		return nil, fmt.Errorf("%w: shard key must use = operator, got %s", ErrShardKeyNotFound, op)
	}

	// Add offset for UPDATE statements
	finalIndex := argIndex
	if !explicitIndex {
		finalIndex += argOffset
	}

	if finalIndex >= len(args) {
		return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, finalIndex+1, len(args))
	}

	return args[finalIndex], nil
}

// findShardKeyArgIndex finds the argument index for a given column in the WHERE clause
// by walking conditions in order and counting placeholders.
func (r *ShardKeyResolver) findShardKeyArgIndex(whereClause, shardKeyColumn string) (string, int, bool, bool) {
	argIndex := 0
	for _, part := range splitConditions(whereClause) {
		column, op, placeholders, explicitArgIndex, ok := parseCondition(part)
		if !ok {
			continue
		}
		if strings.EqualFold(column, shardKeyColumn) {
			if explicitArgIndex >= 0 {
				return op, explicitArgIndex, true, true
			}
			return op, argIndex, false, true
		}
		argIndex += placeholders
	}
	return "", -1, false, false
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

	count, _ := placeholderInfo(setClause)
	return count
}

// shardCondition captures a single WHERE predicate that references the shard key column.
type shardCondition struct {
	op               string
	placeholderCount int
	argStart         int
	explicitIndex    bool
}

// shardValueCollector accumulates shard-key values extracted from WHERE conditions.
type shardValueCollector struct {
	inValues   []any
	eqValue    any
	hasEqual   bool
	rangeStart any
	rangeEnd   any
	hasStart   bool
	hasEnd     bool
}

func (r *ShardKeyResolver) resolveMultipleFromWhere(query string, parsed *ParsedSQL, rule *ShardingRule, args []any) ([]*ResolvedShard, error) {
	if parsed.WhereClause == "" {
		return r.defaultShardOrError(parsed.TableName, rule)
	}

	argOffset := 0
	if parsed.Type == SQLTypeUpdate {
		argOffset = r.countSetPlaceholders(query)
	}

	conditions := r.collectShardConditions(parsed.WhereClause, rule.ShardKeyColumn)
	if len(conditions) == 0 {
		return r.defaultShardOrError(parsed.TableName, rule)
	}

	collector, err := r.extractShardValues(conditions, args, argOffset)
	if err != nil {
		return nil, err
	}

	return collector.buildShards(parsed.TableName, rule)
}

func (r *ShardKeyResolver) defaultShardOrError(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	if rule.HasDefaultShard() {
		return []*ResolvedShard{{
			TableName:  table,
			ShardIndex: rule.DefaultShard,
			ShardKey:   nil,
			Rule:       rule,
		}}, nil
	}
	return nil, fmt.Errorf("%w: column %s not found in WHERE clause", ErrShardKeyNotFound, rule.ShardKeyColumn)
}

func (r *ShardKeyResolver) collectShardConditions(whereClause, shardKeyColumn string) []shardCondition {
	var conditions []shardCondition
	argIndex := 0
	for _, part := range splitConditions(whereClause) {
		column, op, placeholders, explicitArgIndex, ok := parseCondition(part)
		if !ok {
			continue
		}

		conditionArgStart := argIndex
		explicitIndex := false
		if explicitArgIndex >= 0 {
			conditionArgStart = explicitArgIndex
			explicitIndex = true
		}

		if strings.EqualFold(column, shardKeyColumn) {
			conditions = append(conditions, shardCondition{
				op:               op,
				placeholderCount: placeholders,
				argStart:         conditionArgStart,
				explicitIndex:    explicitIndex,
			})
		}

		argIndex += placeholders
	}
	return conditions
}

func (r *ShardKeyResolver) extractShardValues(conditions []shardCondition, args []any, argOffset int) (*shardValueCollector, error) {
	collector := &shardValueCollector{}
	for _, cond := range conditions {
		switch strings.ToUpper(cond.op) {
		case "=":
			value, err := r.resolveSingleArg(cond, argOffset, args)
			if err != nil {
				return nil, err
			}
			collector.eqValue = value
			collector.hasEqual = true
		case "IN":
			values, err := r.resolveInArgs(cond, argOffset, args)
			if err != nil {
				return nil, err
			}
			collector.inValues = append(collector.inValues, values...)
		case ">", ">=":
			value, err := r.resolveSingleArg(cond, argOffset, args)
			if err != nil {
				return nil, err
			}
			collector.rangeStart = value
			collector.hasStart = true
		case "<", "<=":
			value, err := r.resolveSingleArg(cond, argOffset, args)
			if err != nil {
				return nil, err
			}
			collector.rangeEnd = value
			collector.hasEnd = true
		}
	}
	return collector, nil
}

func (r *ShardKeyResolver) resolveSingleArg(cond shardCondition, argOffset int, args []any) (any, error) {
	idx := cond.argStart
	if !cond.explicitIndex {
		idx += argOffset
	}
	if idx >= len(args) {
		return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, idx+1, len(args))
	}
	return args[idx], nil
}

func (r *ShardKeyResolver) resolveInArgs(cond shardCondition, argOffset int, args []any) ([]any, error) {
	if cond.placeholderCount == 0 {
		return nil, fmt.Errorf("%w: expected shard key values for IN clause", ErrInvalidArgumentCount)
	}
	start := cond.argStart
	if !cond.explicitIndex {
		start += argOffset
	}
	end := start + cond.placeholderCount
	if end > len(args) {
		return nil, fmt.Errorf("%w: expected at least %d args, got %d", ErrInvalidArgumentCount, end, len(args))
	}
	return args[start:end], nil
}

func (c *shardValueCollector) buildShards(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	if c.hasEqual {
		return c.resolveEqualShard(table, rule)
	}
	if len(c.inValues) > 0 {
		return c.resolveInShards(table, rule)
	}
	if c.hasStart && c.hasEnd {
		return c.resolveRangeShards(table, rule)
	}
	return c.resolveAllShards(table, rule)
}

func (c *shardValueCollector) resolveEqualShard(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	shardIndex, err := rule.Strategy.Shard(c.eqValue, rule.ShardCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate shard: %w", err)
	}
	return []*ResolvedShard{{
		TableName:  table,
		ShardIndex: shardIndex,
		ShardKey:   c.eqValue,
		Rule:       rule,
	}}, nil
}

func (c *shardValueCollector) resolveInShards(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	unique := make(map[int]any)
	for _, value := range c.inValues {
		shardIndex, err := rule.Strategy.Shard(value, rule.ShardCount)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate shard: %w", err)
		}
		if _, exists := unique[shardIndex]; !exists {
			unique[shardIndex] = value
		}
	}
	return buildResolvedShards(table, rule, unique), nil
}

func (c *shardValueCollector) resolveRangeShards(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	shards, err := rule.Strategy.ShardRange(c.rangeStart, c.rangeEnd, rule.ShardCount)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate shard range: %w", err)
	}
	unique := make(map[int]any)
	for _, shard := range shards {
		unique[shard] = map[string]any{"start": c.rangeStart, "end": c.rangeEnd}
	}
	return buildResolvedShards(table, rule, unique), nil
}

func (c *shardValueCollector) resolveAllShards(table string, rule *ShardingRule) ([]*ResolvedShard, error) {
	unique := make(map[int]any)
	for shard := 0; shard < rule.ShardCount; shard++ {
		unique[shard] = map[string]any{"start": c.rangeStart, "end": c.rangeEnd}
	}
	return buildResolvedShards(table, rule, unique), nil
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

func parseCondition(part string) (string, string, int, int, bool) {
	part = strings.ReplaceAll(part, "`", "")
	part = strings.TrimSpace(part)
	if part == "" {
		return "", "", 0, -1, false
	}

	placeholderCount, explicitArgIndex := placeholderInfo(part)

	upper := strings.ToUpper(part)
	if idx := strings.Index(upper, " IN "); idx > 0 {
		column := strings.TrimSpace(part[:idx])
		column = normalizeColumn(column)
		return column, "IN", placeholderCount, explicitArgIndex, true
	}

	for _, op := range []string{">=", "<=", "!=", "<>", "=", ">", "<"} {
		if idx := strings.Index(part, op); idx > 0 {
			column := strings.TrimSpace(part[:idx])
			column = normalizeColumn(column)
			return column, op, placeholderCount, explicitArgIndex, true
		}
	}

	return "", "", placeholderCount, explicitArgIndex, false
}

func placeholderInfo(s string) (int, int) {
	count := strings.Count(s, "?")
	firstExplicitIndex := -1

	for i := 0; i < len(s); i++ {
		if s[i] != '$' {
			continue
		}
		j := i + 1
		for j < len(s) && s[j] >= '0' && s[j] <= '9' {
			j++
		}
		if j == i+1 {
			continue
		}

		count++
		if firstExplicitIndex == -1 {
			var n int
			for _, ch := range s[i+1 : j] {
				n = n*10 + int(ch-'0')
			}
			if n > 0 {
				firstExplicitIndex = n - 1
			}
		}
		i = j - 1
	}

	return count, firstExplicitIndex
}

func normalizeColumn(column string) string {
	column = strings.TrimSpace(column)
	if dotIdx := strings.LastIndex(column, "."); dotIdx > 0 {
		column = column[dotIdx+1:]
	}
	return column
}
