package sharding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"github.com/spcent/plumego/store/db/rw"
)

var (
	// ErrNoShards is returned when no shards are configured
	ErrNoShards = errors.New("sharding: no shards configured")

	// ErrShardNotFound is returned when the specified shard doesn't exist
	ErrShardNotFound = errors.New("sharding: shard not found")

	// ErrCrossShardQuery is returned when a query spans multiple shards and policy denies it
	ErrCrossShardQuery = errors.New("sharding: cross-shard query not allowed")

	// ErrAllShardsFailed is returned when all shards fail for a cross-shard query
	ErrAllShardsFailed = errors.New("sharding: all shards failed")
)

// CrossShardPolicy defines how to handle queries that span multiple shards
type CrossShardPolicy int

const (
	// CrossShardDeny rejects queries that cannot be routed to a single shard
	CrossShardDeny CrossShardPolicy = iota

	// CrossShardFirst executes the query on the first shard only
	CrossShardFirst

	// CrossShardAll executes the query on all shards concurrently
	CrossShardAll
)

// String returns the string representation of the policy
func (p CrossShardPolicy) String() string {
	switch p {
	case CrossShardDeny:
		return "deny"
	case CrossShardFirst:
		return "first"
	case CrossShardAll:
		return "all"
	default:
		return "unknown"
	}
}

// RouterConfig holds configuration for the router
type RouterConfig struct {
	// CrossShardPolicy determines how to handle cross-shard queries
	CrossShardPolicy CrossShardPolicy

	// EnableMetrics enables query metrics collection
	EnableMetrics bool

	// DefaultShardIndex is used when no shard can be determined (optional, -1 to disable)
	DefaultShardIndex int
}

// DefaultRouterConfig returns default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		CrossShardPolicy:  CrossShardDeny,
		EnableMetrics:     true,
		DefaultShardIndex: -1,
	}
}

// Router routes SQL queries to appropriate shards
type Router struct {
	// shards contains all shard clusters (each shard has read-write splitting)
	shards []*rw.Cluster

	// resolver extracts shard keys from queries
	resolver *ShardKeyResolver

	// config holds router configuration
	config RouterConfig

	// Metrics
	metrics RouterMetrics

	// Lifecycle
	mu sync.RWMutex
}

// RouterMetrics holds routing metrics
type RouterMetrics struct {
	TotalQueries       uint64
	SingleShardQueries uint64
	CrossShardQueries  uint64
	RoutingErrors      uint64
	ShardQueryCounts   []uint64 // Per-shard query counts
}

// RouterOption is a functional option for configuring Router
type RouterOption func(*Router)

// WithCrossShardPolicy sets the cross-shard query policy
func WithCrossShardPolicy(policy CrossShardPolicy) RouterOption {
	return func(r *Router) {
		r.config.CrossShardPolicy = policy
	}
}

// WithDefaultShard sets the default shard index
func WithDefaultShard(index int) RouterOption {
	return func(r *Router) {
		r.config.DefaultShardIndex = index
	}
}

// WithMetrics enables or disables metrics collection
func WithMetrics(enabled bool) RouterOption {
	return func(r *Router) {
		r.config.EnableMetrics = enabled
	}
}

// NewRouter creates a new sharding router
func NewRouter(shards []*rw.Cluster, registry *ShardingRuleRegistry, opts ...RouterOption) (*Router, error) {
	if len(shards) == 0 {
		return nil, ErrNoShards
	}

	if registry == nil {
		return nil, errors.New("sharding: registry cannot be nil")
	}

	// Validate registry
	if err := registry.ValidateAll(); err != nil {
		return nil, fmt.Errorf("sharding: invalid registry: %w", err)
	}

	router := &Router{
		shards:   shards,
		resolver: NewShardKeyResolver(registry),
		config:   DefaultRouterConfig(),
		metrics: RouterMetrics{
			ShardQueryCounts: make([]uint64, len(shards)),
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(router)
	}

	// Validate configuration
	if router.config.DefaultShardIndex >= len(shards) {
		return nil, fmt.Errorf("default shard index %d exceeds shard count %d",
			router.config.DefaultShardIndex, len(shards))
	}

	return router, nil
}

// ExecContext executes a write query (INSERT, UPDATE, DELETE)
// Routes to the primary of the appropriate shard
func (r *Router) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	r.recordQuery()

	// Resolve the shard for this query
	resolved, err := r.resolver.Resolve(query, args)
	if err != nil {
		r.recordRoutingError()
		return nil, fmt.Errorf("failed to resolve shard: %w", err)
	}

	// Validate shard index
	if resolved.ShardIndex < 0 || resolved.ShardIndex >= len(r.shards) {
		r.recordRoutingError()
		return nil, fmt.Errorf("%w: index %d", ErrShardNotFound, resolved.ShardIndex)
	}

	// Record metrics
	r.recordShardQuery(resolved.ShardIndex)

	// Execute on the resolved shard (will use primary)
	shard := r.shards[resolved.ShardIndex]
	return shard.ExecContext(ctx, query, args...)
}

// QueryContext executes a read query (SELECT)
// Routes to a replica (or primary) of the appropriate shard
func (r *Router) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	r.recordQuery()

	// Check if we can resolve to a single shard
	canResolve, err := r.resolver.CanResolve(query)
	if err != nil {
		r.recordRoutingError()
		return nil, fmt.Errorf("failed to check resolvability: %w", err)
	}

	if !canResolve {
		// Cannot resolve to single shard - handle cross-shard query
		return r.handleCrossShardQuery(ctx, query, args)
	}

	// Resolve the shard for this query
	resolved, err := r.resolver.Resolve(query, args)
	if err != nil {
		r.recordRoutingError()
		return nil, fmt.Errorf("failed to resolve shard: %w", err)
	}

	// Validate shard index
	if resolved.ShardIndex < 0 || resolved.ShardIndex >= len(r.shards) {
		r.recordRoutingError()
		return nil, fmt.Errorf("%w: index %d", ErrShardNotFound, resolved.ShardIndex)
	}

	// Record metrics
	r.recordShardQuery(resolved.ShardIndex)

	// Execute on the resolved shard (may use replica)
	shard := r.shards[resolved.ShardIndex]
	return shard.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query that returns at most one row
// Routes to a replica (or primary) of the appropriate shard
func (r *Router) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	r.recordQuery()

	// Resolve the shard for this query
	resolved, err := r.resolver.Resolve(query, args)
	if err != nil {
		r.recordRoutingError()
		// Return a Row with error - the error will surface when Scan is called
		// We cannot return error directly as QueryRowContext doesn't return error
		// Use a fallback approach: try default shard if configured
		if r.config.DefaultShardIndex >= 0 {
			shard := r.shards[r.config.DefaultShardIndex]
			return shard.QueryRowContext(ctx, query, args...)
		}
		// No default, return from first shard (will likely error on Scan)
		return r.shards[0].QueryRowContext(ctx, query, args...)
	}

	// Validate shard index
	if resolved.ShardIndex < 0 || resolved.ShardIndex >= len(r.shards) {
		r.recordRoutingError()
		// Fallback to default or first shard
		if r.config.DefaultShardIndex >= 0 {
			shard := r.shards[r.config.DefaultShardIndex]
			return shard.QueryRowContext(ctx, query, args...)
		}
		return r.shards[0].QueryRowContext(ctx, query, args...)
	}

	// Record metrics
	r.recordShardQuery(resolved.ShardIndex)

	// Execute on the resolved shard
	shard := r.shards[resolved.ShardIndex]
	return shard.QueryRowContext(ctx, query, args...)
}

// BeginTx begins a transaction on a specific shard
// The shard must be determined before starting the transaction
// For sharded transactions, use BeginTxOnShard with explicit shard index
func (r *Router) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	// For now, transactions require explicit shard selection
	// Use default shard if configured
	if r.config.DefaultShardIndex >= 0 {
		return r.BeginTxOnShard(ctx, r.config.DefaultShardIndex, opts)
	}

	return nil, errors.New("sharding: cannot begin transaction without explicit shard selection")
}

// BeginTxOnShard begins a transaction on a specific shard
func (r *Router) BeginTxOnShard(ctx context.Context, shardIndex int, opts *sql.TxOptions) (*sql.Tx, error) {
	if shardIndex < 0 || shardIndex >= len(r.shards) {
		return nil, fmt.Errorf("%w: index %d", ErrShardNotFound, shardIndex)
	}

	shard := r.shards[shardIndex]
	return shard.BeginTx(ctx, opts)
}

// handleCrossShardQuery handles queries that span multiple shards
func (r *Router) handleCrossShardQuery(ctx context.Context, query string, args []any) (*sql.Rows, error) {
	r.recordCrossShardQuery()

	switch r.config.CrossShardPolicy {
	case CrossShardDeny:
		return nil, ErrCrossShardQuery

	case CrossShardFirst:
		// Query the first shard only
		if len(r.shards) == 0 {
			return nil, ErrNoShards
		}
		return r.shards[0].QueryContext(ctx, query, args...)

	case CrossShardAll:
		// Query all shards and merge results
		return r.queryAllShards(ctx, query, args)

	default:
		return nil, fmt.Errorf("unknown cross-shard policy: %v", r.config.CrossShardPolicy)
	}
}

// queryAllShards executes a query on all shards concurrently
// Returns the first successful result or an error if all fail
func (r *Router) queryAllShards(ctx context.Context, query string, args []any) (*sql.Rows, error) {
	type result struct {
		rows  *sql.Rows
		err   error
		shard int
	}

	results := make(chan result, len(r.shards))
	var wg sync.WaitGroup

	// Launch queries on all shards concurrently
	for i, shard := range r.shards {
		wg.Add(1)
		go func(idx int, s *rw.Cluster) {
			defer wg.Done()
			rows, err := s.QueryContext(ctx, query, args...)
			results <- result{rows: rows, err: err, shard: idx}
		}(i, shard)
	}

	// Close results channel when all goroutines complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var firstSuccess *sql.Rows
	var errors []error

	for res := range results {
		if res.err == nil {
			if firstSuccess == nil {
				firstSuccess = res.rows
			} else {
				// Close additional successful results
				res.rows.Close()
			}
		} else {
			errors = append(errors, fmt.Errorf("shard %d: %w", res.shard, res.err))
		}
	}

	if firstSuccess != nil {
		return firstSuccess, nil
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("%w: %v", ErrAllShardsFailed, errors)
	}

	return nil, ErrAllShardsFailed
}

// PingContext checks connectivity to all shards
func (r *Router) PingContext(ctx context.Context) error {
	var errs []error

	for i, shard := range r.shards {
		if err := shard.PingContext(ctx); err != nil {
			errs = append(errs, fmt.Errorf("shard %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("ping failed: %v", errs)
	}

	return nil
}

// Close closes all shard connections
func (r *Router) Close() error {
	var errs []error

	for i, shard := range r.shards {
		if err := shard.Close(); err != nil {
			errs = append(errs, fmt.Errorf("shard %d: %w", i, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("close failed: %v", errs)
	}

	return nil
}

// GetShard returns the shard cluster at the given index
func (r *Router) GetShard(index int) (*rw.Cluster, error) {
	if index < 0 || index >= len(r.shards) {
		return nil, fmt.Errorf("%w: index %d", ErrShardNotFound, index)
	}
	return r.shards[index], nil
}

// ShardCount returns the number of shards
func (r *Router) ShardCount() int {
	return len(r.shards)
}

// Metrics returns a copy of router metrics
func (r *Router) Metrics() RouterMetrics {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Deep copy shard counts
	shardCounts := make([]uint64, len(r.metrics.ShardQueryCounts))
	copy(shardCounts, r.metrics.ShardQueryCounts)

	return RouterMetrics{
		TotalQueries:       r.metrics.TotalQueries,
		SingleShardQueries: r.metrics.SingleShardQueries,
		CrossShardQueries:  r.metrics.CrossShardQueries,
		RoutingErrors:      r.metrics.RoutingErrors,
		ShardQueryCounts:   shardCounts,
	}
}

// recordQuery increments total query count
func (r *Router) recordQuery() {
	if !r.config.EnableMetrics {
		return
	}
	r.mu.Lock()
	r.metrics.TotalQueries++
	r.metrics.SingleShardQueries++
	r.mu.Unlock()
}

// recordCrossShardQuery increments cross-shard query count
func (r *Router) recordCrossShardQuery() {
	if !r.config.EnableMetrics {
		return
	}
	r.mu.Lock()
	r.metrics.CrossShardQueries++
	r.mu.Unlock()
}

// recordRoutingError increments routing error count
func (r *Router) recordRoutingError() {
	if !r.config.EnableMetrics {
		return
	}
	r.mu.Lock()
	r.metrics.RoutingErrors++
	r.mu.Unlock()
}

// recordShardQuery increments query count for a specific shard
func (r *Router) recordShardQuery(shardIndex int) {
	if !r.config.EnableMetrics {
		return
	}
	r.mu.Lock()
	if shardIndex >= 0 && shardIndex < len(r.metrics.ShardQueryCounts) {
		r.metrics.ShardQueryCounts[shardIndex]++
	}
	r.mu.Unlock()
}
