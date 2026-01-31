package sharding

import (
	"fmt"
	"strings"
	"sync"
)

// SQLRewriter rewrites SQL queries by replacing logical table names with physical table names.
// This is essential for physical sharding scenarios where each shard has different physical tables.
//
// Example:
//
//	rewriter := NewSQLRewriter(registry)
//	rewritten, err := rewriter.Rewrite("SELECT * FROM users WHERE id = ?", 0)
//	// Returns: "SELECT * FROM users_0 WHERE id = ?"
type SQLRewriter struct {
	registry *ShardingRuleRegistry
	parser   *SQLParser
	cache    *rewriteCache
}

// rewriteCache caches SQL rewrite results for performance
type rewriteCache struct {
	mu      sync.RWMutex
	cache   map[string]map[int]string // query -> shardIndex -> rewritten query
	maxSize int
}

// RewriteResult contains the result of SQL rewriting
type RewriteResult struct {
	// OriginalSQL is the original SQL query
	OriginalSQL string

	// RewrittenSQL is the rewritten SQL query with physical table names
	RewrittenSQL string

	// TableName is the logical table name that was rewritten
	TableName string

	// PhysicalTableName is the physical table name used in rewriting
	PhysicalTableName string

	// ShardIndex is the target shard index
	ShardIndex int

	// WasRewritten indicates if any rewriting occurred
	WasRewritten bool
}

// NewSQLRewriter creates a new SQL rewriter
func NewSQLRewriter(registry *ShardingRuleRegistry) *SQLRewriter {
	return &SQLRewriter{
		registry: registry,
		parser:   NewSQLParser(),
		cache:    newRewriteCache(1000), // Cache up to 1000 queries
	}
}

// NewSQLRewriterWithCacheSize creates a new SQL rewriter with custom cache size
func NewSQLRewriterWithCacheSize(registry *ShardingRuleRegistry, cacheSize int) *SQLRewriter {
	return &SQLRewriter{
		registry: registry,
		parser:   NewSQLParser(),
		cache:    newRewriteCache(cacheSize),
	}
}

// newRewriteCache creates a new rewrite cache
func newRewriteCache(maxSize int) *rewriteCache {
	return &rewriteCache{
		cache:   make(map[string]map[int]string),
		maxSize: maxSize,
	}
}

// Rewrite rewrites a SQL query by replacing the logical table name with the physical table name
// for the specified shard index.
//
// Example:
//
//	rewritten, err := rewriter.Rewrite("SELECT * FROM users WHERE id = ?", 0)
//	// Returns: "SELECT * FROM users_0 WHERE id = ?"
func (r *SQLRewriter) Rewrite(query string, shardIndex int) (string, error) {
	// Check cache first
	if cached := r.cache.get(query, shardIndex); cached != "" {
		return cached, nil
	}

	// Parse the SQL query
	parsed, err := r.parser.Parse(query)
	if err != nil {
		return "", fmt.Errorf("failed to parse SQL: %w", err)
	}

	// Get the sharding rule for this table
	rule, err := r.registry.Get(parsed.TableName)
	if err != nil {
		// No sharding rule found - return original query
		// This might be a non-sharded table
		return query, nil
	}

	// Validate shard index
	if shardIndex < 0 || shardIndex >= rule.ShardCount {
		return "", fmt.Errorf("invalid shard index %d for table %s (shard count: %d)",
			shardIndex, parsed.TableName, rule.ShardCount)
	}

	// Get the physical table name for this shard
	physicalTableName := rule.GetActualTableName(shardIndex)

	// If physical table name is same as logical, no rewrite needed
	if physicalTableName == parsed.TableName {
		r.cache.put(query, shardIndex, query)
		return query, nil
	}

	// Rewrite the query
	rewritten := r.replaceTableName(query, parsed.TableName, physicalTableName)

	// Cache the result
	r.cache.put(query, shardIndex, rewritten)

	return rewritten, nil
}

// RewriteWithDetails rewrites a SQL query and returns detailed information
func (r *SQLRewriter) RewriteWithDetails(query string, shardIndex int) (*RewriteResult, error) {
	// Parse the SQL query
	parsed, err := r.parser.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SQL: %w", err)
	}

	result := &RewriteResult{
		OriginalSQL:  query,
		RewrittenSQL: query,
		TableName:    parsed.TableName,
		ShardIndex:   shardIndex,
		WasRewritten: false,
	}

	// Get the sharding rule for this table
	rule, err := r.registry.Get(parsed.TableName)
	if err != nil {
		// No sharding rule found - return original query
		return result, nil
	}

	// Validate shard index
	if shardIndex < 0 || shardIndex >= rule.ShardCount {
		return nil, fmt.Errorf("invalid shard index %d for table %s (shard count: %d)",
			shardIndex, parsed.TableName, rule.ShardCount)
	}

	// Get the physical table name for this shard
	physicalTableName := rule.GetActualTableName(shardIndex)
	result.PhysicalTableName = physicalTableName

	// If physical table name is same as logical, no rewrite needed
	if physicalTableName == parsed.TableName {
		return result, nil
	}

	// Rewrite the query
	result.RewrittenSQL = r.replaceTableName(query, parsed.TableName, physicalTableName)
	result.WasRewritten = true

	return result, nil
}

// replaceTableName replaces all occurrences of the logical table name with the physical table name
// while preserving SQL semantics (quotes, case sensitivity, etc.)
func (r *SQLRewriter) replaceTableName(query, logicalTable, physicalTable string) string {
	// Handle different quoting styles
	replacements := []struct {
		from string
		to   string
	}{
		// Backtick quoted (MySQL style)
		{fmt.Sprintf("`%s`", logicalTable), fmt.Sprintf("`%s`", physicalTable)},
		// Double quote quoted (PostgreSQL style)
		{fmt.Sprintf(`"%s"`, logicalTable), fmt.Sprintf(`"%s"`, physicalTable)},
		// Square bracket quoted (SQL Server style)
		{fmt.Sprintf("[%s]", logicalTable), fmt.Sprintf("[%s]", physicalTable)},
	}

	result := query

	// First, replace quoted table names
	for _, repl := range replacements {
		result = strings.ReplaceAll(result, repl.from, repl.to)
	}

	// Then, replace unquoted table names
	// We need to be careful to only replace whole words
	result = r.replaceWholeWord(result, logicalTable, physicalTable)

	return result
}

// replaceWholeWord replaces whole word occurrences only (not substrings)
// This prevents replacing "users" in "users_backup" for example
func (r *SQLRewriter) replaceWholeWord(text, from, to string) string {
	// Word boundaries for SQL identifiers
	boundaries := " \t\n\r,()=<>!;."

	// Convert to lowercase for case-insensitive matching
	lowerText := strings.ToLower(text)
	lowerFrom := strings.ToLower(from)

	var result strings.Builder
	result.Grow(len(text))

	i := 0
	for i < len(text) {
		// Try to find the next occurrence
		idx := strings.Index(lowerText[i:], lowerFrom)
		if idx == -1 {
			// No more occurrences, append rest of text
			result.WriteString(text[i:])
			break
		}

		// Adjust index to absolute position
		idx += i

		// Check if it's a whole word (preceded and followed by word boundaries)
		precedingOK := idx == 0 || strings.ContainsRune(boundaries, rune(text[idx-1]))
		followingIdx := idx + len(from)
		followingOK := followingIdx >= len(text) || strings.ContainsRune(boundaries, rune(text[followingIdx]))

		if precedingOK && followingOK {
			// It's a whole word - replace it
			result.WriteString(text[i:idx])
			result.WriteString(to)
			i = followingIdx
		} else {
			// Not a whole word - skip this occurrence
			result.WriteString(text[i : idx+1])
			i = idx + 1
		}
	}

	return result.String()
}

// ClearCache clears the rewrite cache
func (r *SQLRewriter) ClearCache() {
	r.cache.clear()
}

// CacheStats returns cache statistics
func (r *SQLRewriter) CacheStats() CacheStats {
	return r.cache.stats()
}

// CacheStats holds cache statistics
type CacheStats struct {
	Entries  int
	MaxSize  int
	HitRate  float64
	Requests uint64
	Hits     uint64
	Misses   uint64
}

// get retrieves a cached rewritten query
func (c *rewriteCache) get(query string, shardIndex int) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if shardMap, ok := c.cache[query]; ok {
		if rewritten, ok := shardMap[shardIndex]; ok {
			return rewritten
		}
	}

	return ""
}

// put stores a rewritten query in the cache
func (c *rewriteCache) put(query string, shardIndex int, rewritten string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if we need to evict entries (simple size-based eviction)
	if len(c.cache) >= c.maxSize {
		// Evict a random entry (could be improved with LRU)
		for k := range c.cache {
			delete(c.cache, k)
			break
		}
	}

	// Store the rewritten query
	if _, ok := c.cache[query]; !ok {
		c.cache[query] = make(map[int]string)
	}
	c.cache[query][shardIndex] = rewritten
}

// clear clears the cache
func (c *rewriteCache) clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]map[int]string)
}

// stats returns cache statistics
func (c *rewriteCache) stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := 0
	for _, shardMap := range c.cache {
		entries += len(shardMap)
	}

	return CacheStats{
		Entries: entries,
		MaxSize: c.maxSize,
	}
}

// BatchRewrite rewrites multiple queries for different shard indices
// This is useful for cross-shard queries
func (r *SQLRewriter) BatchRewrite(query string, shardIndices []int) (map[int]string, error) {
	results := make(map[int]string, len(shardIndices))

	for _, shardIdx := range shardIndices {
		rewritten, err := r.Rewrite(query, shardIdx)
		if err != nil {
			return nil, fmt.Errorf("failed to rewrite for shard %d: %w", shardIdx, err)
		}
		results[shardIdx] = rewritten
	}

	return results, nil
}
