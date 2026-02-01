package distributed

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/store/cache"
)

// ReplicationMode defines how data is replicated across nodes
type ReplicationMode int

const (
	// ReplicationNone means no replication (single copy)
	ReplicationNone ReplicationMode = iota

	// ReplicationAsync means asynchronous replication (fast writes, eventual consistency)
	ReplicationAsync

	// ReplicationSync means synchronous replication (slower writes, strong consistency)
	ReplicationSync
)

// FailoverStrategy defines how to handle node failures
type FailoverStrategy int

const (
	// FailoverNextNode tries the next node in the hash ring
	FailoverNextNode FailoverStrategy = iota

	// FailoverAllNodes broadcasts to all healthy nodes
	FailoverAllNodes

	// FailoverRetry retries the same node with backoff
	FailoverRetry
)

// DistributedCache implements a distributed cache using consistent hashing
type DistributedCache struct {
	ring              HashRing
	healthChecker     *HealthChecker
	replicationMode   ReplicationMode
	replicationFactor int
	failoverStrategy  FailoverStrategy
	metrics           *DistributedMetrics
	mu                sync.RWMutex
}

// DistributedMetrics tracks metrics for the distributed cache
type DistributedMetrics struct {
	mu sync.RWMutex

	TotalRequests    uint64
	FailoverCount    uint64
	ReplicationLag   time.Duration
	HashCollisions   uint64
	HealthyNodes     int
	UnhealthyNodes   int
	RebalanceEvents  uint64
}

// Config configures the distributed cache
type Config struct {
	VirtualNodes        int
	ReplicationFactor   int
	ReplicationMode     ReplicationMode
	FailoverStrategy    FailoverStrategy
	HashFunc            HashFunc
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	EnableMetrics       bool
}

// DefaultConfig returns the default distributed cache configuration
func DefaultConfig() *Config {
	return &Config{
		VirtualNodes:        150,
		ReplicationFactor:   1,
		ReplicationMode:     ReplicationAsync,
		FailoverStrategy:    FailoverNextNode,
		HashFunc:            nil, // Will use default FNV-1a
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  2 * time.Second,
		EnableMetrics:       true,
	}
}

// New creates a new distributed cache
func New(nodes []CacheNode, config *Config) *DistributedCache {
	if config == nil {
		config = DefaultConfig()
	}

	// Create hash ring
	ringConfig := &ConsistentHashRingConfig{
		VirtualNodes: config.VirtualNodes,
		HashFunc:     config.HashFunc,
	}
	ring := NewConsistentHashRing(ringConfig)

	// Add nodes to ring
	for _, node := range nodes {
		ring.Add(node)
	}

	// Create health checker
	healthConfig := &HealthCheckerConfig{
		CheckInterval: config.HealthCheckInterval,
		CheckTimeout:  config.HealthCheckTimeout,
	}
	healthChecker := NewHealthChecker(healthConfig)

	// Add nodes to health checker
	for _, node := range nodes {
		healthChecker.AddNode(node)
	}

	// Start health checking
	healthChecker.Start()

	dc := &DistributedCache{
		ring:              ring,
		healthChecker:     healthChecker,
		replicationMode:   config.ReplicationMode,
		replicationFactor: config.ReplicationFactor,
		failoverStrategy:  config.FailoverStrategy,
		metrics:           &DistributedMetrics{},
	}

	return dc
}

// Close stops the distributed cache and health checker
func (dc *DistributedCache) Close() error {
	dc.healthChecker.Stop()
	return nil
}

// Get retrieves a value from the distributed cache
func (dc *DistributedCache) Get(ctx context.Context, key string) ([]byte, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return nil, err
	}

	// Try primary node
	if node.IsHealthy() {
		value, err := node.Cache().Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// If error is not "not found", try failover
		if !errors.Is(err, cache.ErrNotFound) {
			return dc.failoverGet(ctx, key, node.ID())
		}

		return nil, err
	}

	// Primary node unhealthy, try failover
	return dc.failoverGet(ctx, key, node.ID())
}

// Set stores a value in the distributed cache
func (dc *DistributedCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get nodes for replication
	nodes, err := dc.ring.GetN(key, dc.replicationFactor)
	if err != nil {
		return err
	}

	switch dc.replicationMode {
	case ReplicationSync:
		return dc.setSyncReplicas(ctx, nodes, key, value, ttl)
	case ReplicationAsync:
		return dc.setAsyncReplicas(ctx, nodes, key, value, ttl)
	default:
		// No replication, just write to primary
		if len(nodes) > 0 && nodes[0].IsHealthy() {
			return nodes[0].Cache().Set(ctx, key, value, ttl)
		}
		return ErrNodeUnhealthy
	}
}

// Delete removes a value from the distributed cache
func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get nodes for replication
	nodes, err := dc.ring.GetN(key, dc.replicationFactor)
	if err != nil {
		return err
	}

	// Delete from all replicas
	var firstErr error
	for _, node := range nodes {
		if !node.IsHealthy() {
			continue
		}

		err := node.Cache().Delete(ctx, key)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Exists checks if a key exists in the distributed cache
func (dc *DistributedCache) Exists(ctx context.Context, key string) (bool, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return false, err
	}

	if !node.IsHealthy() {
		return false, ErrNodeUnhealthy
	}

	return node.Cache().Exists(ctx, key)
}

// Clear clears all data from all nodes (use with caution!)
func (dc *DistributedCache) Clear(ctx context.Context) error {
	nodes := dc.ring.Nodes()

	var firstErr error
	for _, node := range nodes {
		if !node.IsHealthy() {
			continue
		}

		err := node.Cache().Clear(ctx)
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	return firstErr
}

// Incr increments an integer value
func (dc *DistributedCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return 0, err
	}

	if !node.IsHealthy() {
		return 0, ErrNodeUnhealthy
	}

	return node.Cache().Incr(ctx, key, delta)
}

// Decr decrements an integer value
func (dc *DistributedCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return 0, err
	}

	if !node.IsHealthy() {
		return 0, ErrNodeUnhealthy
	}

	return node.Cache().Decr(ctx, key, delta)
}

// Append appends data to an existing value
func (dc *DistributedCache) Append(ctx context.Context, key string, data []byte) error {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return err
	}

	if !node.IsHealthy() {
		return ErrNodeUnhealthy
	}

	return node.Cache().Append(ctx, key, data)
}

// AddNode adds a new node to the distributed cache
func (dc *DistributedCache) AddNode(node CacheNode) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Add to hash ring
	if err := dc.ring.Add(node); err != nil {
		return err
	}

	// Add to health checker
	dc.healthChecker.AddNode(node)

	// Increment rebalance events
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.RebalanceEvents, 1)
	}

	return nil
}

// RemoveNode removes a node from the distributed cache
func (dc *DistributedCache) RemoveNode(nodeID string) error {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Remove from health checker
	dc.healthChecker.RemoveNode(nodeID)

	// Remove from hash ring
	if err := dc.ring.Remove(nodeID); err != nil {
		return err
	}

	// Increment rebalance events
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.RebalanceEvents, 1)
	}

	return nil
}

// Nodes returns all nodes in the distributed cache
func (dc *DistributedCache) Nodes() []CacheNode {
	return dc.ring.Nodes()
}

// NodeHealth returns the health status of a node
func (dc *DistributedCache) NodeHealth(nodeID string) (HealthStatus, error) {
	return dc.healthChecker.GetNodeStatus(nodeID)
}

// GetMetrics returns a snapshot of distributed cache metrics
func (dc *DistributedCache) GetMetrics() *DistributedMetrics {
	if dc.metrics == nil {
		return nil
	}

	dc.metrics.mu.RLock()
	defer dc.metrics.mu.RUnlock()

	// Count healthy/unhealthy nodes
	healthy := 0
	unhealthy := 0
	for _, node := range dc.ring.Nodes() {
		if node.IsHealthy() {
			healthy++
		} else {
			unhealthy++
		}
	}

	return &DistributedMetrics{
		TotalRequests:   atomic.LoadUint64(&dc.metrics.TotalRequests),
		FailoverCount:   atomic.LoadUint64(&dc.metrics.FailoverCount),
		ReplicationLag:  dc.metrics.ReplicationLag,
		HashCollisions:  atomic.LoadUint64(&dc.metrics.HashCollisions),
		HealthyNodes:    healthy,
		UnhealthyNodes:  unhealthy,
		RebalanceEvents: atomic.LoadUint64(&dc.metrics.RebalanceEvents),
	}
}

// setSyncReplicas writes to all replicas synchronously
func (dc *DistributedCache) setSyncReplicas(ctx context.Context, nodes []CacheNode, key string, value []byte, ttl time.Duration) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for _, node := range nodes {
		if !node.IsHealthy() {
			continue
		}

		wg.Add(1)
		go func(n CacheNode) {
			defer wg.Done()
			err := n.Cache().Set(ctx, key, value, ttl)
			if err != nil {
				errChan <- err
			}
		}(node)
	}

	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		return err
	}

	return nil
}

// setAsyncReplicas writes to primary synchronously, others asynchronously
func (dc *DistributedCache) setAsyncReplicas(ctx context.Context, nodes []CacheNode, key string, value []byte, ttl time.Duration) error {
	if len(nodes) == 0 {
		return ErrNoNodesAvailable
	}

	// Write to primary node synchronously
	primary := nodes[0]
	if !primary.IsHealthy() {
		return ErrNodeUnhealthy
	}

	err := primary.Cache().Set(ctx, key, value, ttl)
	if err != nil {
		return err
	}

	// Write to replicas asynchronously
	for i := 1; i < len(nodes); i++ {
		node := nodes[i]
		if !node.IsHealthy() {
			continue
		}

		go func(n CacheNode) {
			// Use background context for async replication
			bgCtx := context.Background()
			n.Cache().Set(bgCtx, key, value, ttl)
		}(node)
	}

	return nil
}

// failoverGet attempts to get a value from replica nodes
func (dc *DistributedCache) failoverGet(ctx context.Context, key string, failedNodeID string) ([]byte, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.FailoverCount, 1)
	}

	// Get replica nodes
	nodes, err := dc.ring.GetN(key, dc.replicationFactor)
	if err != nil {
		return nil, err
	}

	// Try each replica (skip the failed node)
	for _, node := range nodes {
		if node.ID() == failedNodeID || !node.IsHealthy() {
			continue
		}

		value, err := node.Cache().Get(ctx, key)
		if err == nil {
			return value, nil
		}

		// If not found, continue to next replica
		if errors.Is(err, cache.ErrNotFound) {
			continue
		}
	}

	return nil, cache.ErrNotFound
}
