package distributed

import (
	"context"
	"errors"
	"fmt"
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

	// ReplicationSync means synchronous best-effort replica writes. It waits for
	// selected replicas and reports failures, but it does not roll back replicas
	// that accepted the mutation before another replica failed.
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

const (
	failoverRetryAttempts                 = 3
	failoverRetryBackoff                  = 10 * time.Millisecond
	defaultAsyncTimeout                   = 2 * time.Second
	defaultAsyncReplicationMaxConcurrency = 64
	defaultAsyncReplicationQueueSize      = 256
)

// ErrClosed is returned when operations are attempted after the cache is closed.
var ErrClosed = errors.New("distributed: cache closed")

// AsyncReplicationDropReason identifies why async secondary replication work was dropped.
type AsyncReplicationDropReason string

const (
	// AsyncReplicationDropQueueFull means the bounded async replication queue had no capacity.
	AsyncReplicationDropQueueFull AsyncReplicationDropReason = "queue_full"

	// AsyncReplicationDropClosed means the cache was closed before async replication could be scheduled.
	AsyncReplicationDropClosed AsyncReplicationDropReason = "closed"
)

// AsyncReplicationDrop describes an async secondary replication attempt that was not scheduled.
type AsyncReplicationDrop struct {
	Operation string
	Key       string
	NodeID    string
	Reason    AsyncReplicationDropReason
}

// DistributedCache implements a distributed cache using consistent hashing
type DistributedCache struct {
	ring              HashRing
	healthChecker     *HealthChecker
	replicationMode   ReplicationMode
	replicationFactor int
	failoverStrategy  FailoverStrategy
	failoverAttempts  int
	failoverBackoff   time.Duration
	asyncTimeout      time.Duration
	asyncConcurrency  int
	asyncQueue        chan asyncReplicationJob
	asyncDropHandler  func(AsyncReplicationDrop)
	asyncStop         chan struct{}
	asyncWG           sync.WaitGroup
	asyncClosed       atomic.Bool
	metrics           *DistributedMetrics
	mu                sync.RWMutex
	closeOnce         sync.Once
}

type asyncReplicationJob struct {
	operation string
	key       string
	nodeID    string
	run       func(context.Context) error
}

// DistributedMetrics tracks metrics for the distributed cache
type DistributedMetrics struct {
	mu sync.RWMutex

	TotalRequests       uint64
	FailoverCount       uint64
	ReplicationLag      time.Duration
	ReplicationFailures uint64
	HashCollisions      uint64
	HealthyNodes        int
	UnhealthyNodes      int
	RebalanceEvents     uint64
}

type collisionCounter interface {
	CollisionCount() uint64
}

// Config configures the distributed cache
type Config struct {
	VirtualNodes                   int
	ReplicationFactor              int
	ReplicationMode                ReplicationMode
	FailoverStrategy               FailoverStrategy
	HashFunc                       HashFunc
	HealthCheckInterval            time.Duration
	HealthCheckTimeout             time.Duration
	HealthProbe                    HealthProbe
	FailoverRetryAttempts          int
	FailoverRetryBackoff           time.Duration
	AsyncReplicationTimeout        time.Duration
	AsyncReplicationMaxConcurrency int
	AsyncReplicationQueueSize      int
	AsyncReplicationDropHandler    func(AsyncReplicationDrop)
	EnableMetrics                  bool
}

// DefaultConfig returns the default distributed cache configuration
func DefaultConfig() *Config {
	return &Config{
		VirtualNodes:                   150,
		ReplicationFactor:              1,
		ReplicationMode:                ReplicationAsync,
		FailoverStrategy:               FailoverNextNode,
		HashFunc:                       nil, // Will use default FNV-1a
		HealthCheckInterval:            10 * time.Second,
		HealthCheckTimeout:             2 * time.Second,
		FailoverRetryAttempts:          failoverRetryAttempts,
		FailoverRetryBackoff:           failoverRetryBackoff,
		AsyncReplicationTimeout:        defaultAsyncTimeout,
		AsyncReplicationMaxConcurrency: defaultAsyncReplicationMaxConcurrency,
		AsyncReplicationQueueSize:      defaultAsyncReplicationQueueSize,
		EnableMetrics:                  true,
	}
}

// Validate checks whether the distributed cache configuration is usable.
func (c *Config) Validate() error {
	if c == nil {
		return nil
	}
	if c.VirtualNodes < 0 {
		return errors.New("distributed: virtual nodes cannot be negative")
	}
	if c.ReplicationFactor <= 0 {
		return errors.New("distributed: replication factor must be greater than 0")
	}
	if c.ReplicationMode < ReplicationNone || c.ReplicationMode > ReplicationSync {
		return errors.New("distributed: invalid replication mode")
	}
	if c.FailoverStrategy < FailoverNextNode || c.FailoverStrategy > FailoverRetry {
		return errors.New("distributed: invalid failover strategy")
	}
	if c.HealthCheckInterval < 0 {
		return errors.New("distributed: health check interval cannot be negative")
	}
	if c.HealthCheckTimeout < 0 {
		return errors.New("distributed: health check timeout cannot be negative")
	}
	if c.FailoverRetryAttempts < 0 {
		return errors.New("distributed: failover retry attempts cannot be negative")
	}
	if c.FailoverRetryBackoff < 0 {
		return errors.New("distributed: failover retry backoff cannot be negative")
	}
	if c.AsyncReplicationTimeout < 0 {
		return errors.New("distributed: async replication timeout cannot be negative")
	}
	if c.AsyncReplicationMaxConcurrency < 0 {
		return errors.New("distributed: async replication max concurrency cannot be negative")
	}
	if c.AsyncReplicationQueueSize < 0 {
		return errors.New("distributed: async replication queue size cannot be negative")
	}
	return nil
}

// New creates a new distributed cache.
//
// Prefer NewWithConfig when callers need construction errors. New is retained as
// a compatibility helper and returns nil when validation fails.
func New(nodes []CacheNode, config *Config) *DistributedCache {
	dc, err := NewWithConfig(nodes, config)
	if err != nil {
		return nil
	}
	return dc
}

// NewWithConfig creates a new distributed cache and returns construction errors.
func NewWithConfig(nodes []CacheNode, config *Config) (*DistributedCache, error) {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}
	normalized := *config
	if normalized.VirtualNodes == 0 {
		normalized.VirtualNodes = DefaultConfig().VirtualNodes
	}
	if normalized.HealthCheckInterval == 0 {
		normalized.HealthCheckInterval = DefaultConfig().HealthCheckInterval
	}
	if normalized.HealthCheckTimeout == 0 {
		normalized.HealthCheckTimeout = DefaultConfig().HealthCheckTimeout
	}
	if normalized.FailoverRetryAttempts == 0 {
		normalized.FailoverRetryAttempts = DefaultConfig().FailoverRetryAttempts
	}
	if normalized.FailoverRetryBackoff == 0 {
		normalized.FailoverRetryBackoff = DefaultConfig().FailoverRetryBackoff
	}
	if normalized.AsyncReplicationTimeout == 0 {
		normalized.AsyncReplicationTimeout = DefaultConfig().AsyncReplicationTimeout
	}
	if normalized.AsyncReplicationMaxConcurrency == 0 {
		normalized.AsyncReplicationMaxConcurrency = DefaultConfig().AsyncReplicationMaxConcurrency
	}
	if normalized.AsyncReplicationQueueSize == 0 {
		normalized.AsyncReplicationQueueSize = DefaultConfig().AsyncReplicationQueueSize
	}

	// Create hash ring
	ringConfig := &ConsistentHashRingConfig{
		VirtualNodes: normalized.VirtualNodes,
		HashFunc:     normalized.HashFunc,
	}
	ring := NewConsistentHashRing(ringConfig)

	// Add nodes to ring
	for _, node := range nodes {
		if err := ring.Add(node); err != nil {
			return nil, err
		}
	}

	// Create health checker
	healthConfig := &HealthCheckerConfig{
		CheckInterval: normalized.HealthCheckInterval,
		CheckTimeout:  normalized.HealthCheckTimeout,
		Probe:         normalized.HealthProbe,
	}
	healthChecker := NewHealthChecker(healthConfig)

	// Add nodes to health checker
	for _, node := range nodes {
		if err := healthChecker.AddNode(node); err != nil {
			return nil, err
		}
	}

	// Start health checking
	healthChecker.Start()

	var metrics *DistributedMetrics
	if normalized.EnableMetrics {
		metrics = &DistributedMetrics{}
	}

	dc := &DistributedCache{
		ring:              ring,
		healthChecker:     healthChecker,
		replicationMode:   normalized.ReplicationMode,
		replicationFactor: normalized.ReplicationFactor,
		failoverStrategy:  normalized.FailoverStrategy,
		failoverAttempts:  normalized.FailoverRetryAttempts,
		failoverBackoff:   normalized.FailoverRetryBackoff,
		asyncTimeout:      normalized.AsyncReplicationTimeout,
		asyncConcurrency:  normalized.AsyncReplicationMaxConcurrency,
		asyncQueue:        make(chan asyncReplicationJob, normalized.AsyncReplicationQueueSize),
		asyncDropHandler:  normalized.AsyncReplicationDropHandler,
		asyncStop:         make(chan struct{}),
		metrics:           metrics,
	}
	dc.startAsyncReplicationWorkers(normalized.AsyncReplicationMaxConcurrency)

	return dc, nil
}

// Close stops the distributed cache and health checker
func (dc *DistributedCache) Close() error {
	if dc == nil || dc.healthChecker == nil {
		return nil
	}
	dc.closeOnce.Do(func() {
		dc.asyncClosed.Store(true)
		if dc.asyncStop != nil {
			close(dc.asyncStop)
		}
		dc.asyncWG.Wait()
		dc.drainAsyncReplicationQueue()
		dc.healthChecker.Stop()
	})
	return nil
}

// Get retrieves a value from the distributed cache
func (dc *DistributedCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := dc.ensureOpen(); err != nil {
		return nil, err
	}
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
			return cloneBytes(value), nil
		}

		// If error is not "not found", try failover
		if !errors.Is(err, cache.ErrNotFound) {
			return dc.failoverGet(ctx, key, node.ID(), err)
		}

		return nil, err
	}

	// Primary node unhealthy, try failover
	return dc.failoverGet(ctx, key, node.ID(), ErrNodeUnhealthy)
}

// Set stores a value in the distributed cache
func (dc *DistributedCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	nodes, err := dc.replicationNodes(key)
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
			return nodes[0].Cache().Set(ctx, key, cloneBytes(value), ttl)
		}
		return ErrNodeUnhealthy
	}
}

// Delete removes a value from the distributed cache
func (dc *DistributedCache) Delete(ctx context.Context, key string) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	nodes, err := dc.replicationNodes(key)
	if err != nil {
		return err
	}

	// Delete from all replicas. This is best-effort and may leave partial
	// side effects when some replicas accept the delete before another fails.
	var firstErr error
	deleted := 0
	for _, node := range nodes {
		if !node.IsHealthy() {
			if firstErr == nil {
				firstErr = ErrNodeUnhealthy
			}
			continue
		}

		err := node.Cache().Delete(ctx, key)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if err == nil {
			deleted++
		}
	}

	if deleted == 0 && firstErr != nil {
		return firstErr
	}
	return firstErr
}

// Exists checks if a key exists in the distributed cache
func (dc *DistributedCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := dc.ensureOpen(); err != nil {
		return false, err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	// Get primary node
	node, err := dc.ring.Get(key)
	if err != nil {
		return false, err
	}

	if !node.IsHealthy() {
		return dc.failoverExists(ctx, key, node.ID(), ErrNodeUnhealthy)
	}

	exists, err := node.Cache().Exists(ctx, key)
	if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return dc.failoverExists(ctx, key, node.ID(), err)
	}
	return exists, err
}

// Clear clears all data from all nodes (use with caution!). It is best-effort:
// callers receive an error when any node fails, but nodes already cleared are
// not rolled back.
func (dc *DistributedCache) Clear(ctx context.Context) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
	nodes := dc.ring.Nodes()

	var firstErr error
	cleared := 0
	unhealthy := 0
	for _, node := range nodes {
		if !node.IsHealthy() {
			unhealthy++
			if firstErr == nil {
				firstErr = ErrNodeUnhealthy
			}
			continue
		}

		err := node.Cache().Clear(ctx)
		if err != nil && firstErr == nil {
			firstErr = err
		}
		if err == nil {
			cleared++
		}
	}

	if len(nodes) == 0 {
		return ErrNoNodesAvailable
	}
	if cleared == 0 {
		return firstErr
	}
	if firstErr != nil {
		return fmt.Errorf("distributed: clear partially failed after clearing %d node(s), %d unhealthy: %w", cleared, unhealthy, firstErr)
	}

	return nil
}

// Incr increments an integer value
func (dc *DistributedCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := dc.ensureOpen(); err != nil {
		return 0, err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	nodes, err := dc.replicationNodes(key)
	if err != nil {
		return 0, err
	}

	return dc.incrReplicas(ctx, nodes, key, delta)
}

// Decr decrements an integer value
func (dc *DistributedCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := dc.ensureOpen(); err != nil {
		return 0, err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	nodes, err := dc.replicationNodes(key)
	if err != nil {
		return 0, err
	}

	return dc.incrReplicas(ctx, nodes, key, -delta)
}

// Append appends data to an existing value
func (dc *DistributedCache) Append(ctx context.Context, key string, data []byte) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.TotalRequests, 1)
	}

	nodes, err := dc.replicationNodes(key)
	if err != nil {
		return err
	}

	return dc.appendReplicas(ctx, nodes, key, data)
}

// AddNode adds a new node to the distributed cache
func (dc *DistributedCache) AddNode(node CacheNode) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
	dc.mu.Lock()
	defer dc.mu.Unlock()

	// Add to hash ring
	if err := dc.ring.Add(node); err != nil {
		return err
	}

	// Add to health checker
	if err := dc.healthChecker.AddNode(node); err != nil {
		_ = dc.ring.Remove(node.ID())
		return err
	}

	// Increment rebalance events
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.RebalanceEvents, 1)
	}

	return nil
}

// RemoveNode removes a node from the distributed cache
func (dc *DistributedCache) RemoveNode(nodeID string) error {
	if err := dc.ensureOpen(); err != nil {
		return err
	}
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
	if err := dc.ensureOpen(); err != nil {
		return HealthStatusUnhealthy, err
	}
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

	metrics := &DistributedMetrics{
		TotalRequests:       atomic.LoadUint64(&dc.metrics.TotalRequests),
		FailoverCount:       atomic.LoadUint64(&dc.metrics.FailoverCount),
		ReplicationLag:      dc.metrics.ReplicationLag,
		ReplicationFailures: atomic.LoadUint64(&dc.metrics.ReplicationFailures),
		HealthyNodes:        healthy,
		UnhealthyNodes:      unhealthy,
		RebalanceEvents:     atomic.LoadUint64(&dc.metrics.RebalanceEvents),
	}

	if counter, ok := dc.ring.(collisionCounter); ok {
		metrics.HashCollisions = counter.CollisionCount()
	} else {
		metrics.HashCollisions = atomic.LoadUint64(&dc.metrics.HashCollisions)
	}

	return metrics
}

// setSyncReplicas writes to all replicas synchronously
func (dc *DistributedCache) setSyncReplicas(ctx context.Context, nodes []CacheNode, key string, value []byte, ttl time.Duration) error {
	start := time.Now()
	errChan := make(chan error, len(nodes))
	healthyNodes := make([]CacheNode, 0, len(nodes))
	var firstErr error

	for _, node := range nodes {
		if !node.IsHealthy() {
			errChan <- ErrNodeUnhealthy
			continue
		}
		healthyNodes = append(healthyNodes, node)
	}

	if len(healthyNodes) == 0 {
		dc.recordReplicationFailure()
		dc.recordReplicationLag(start)
		return ErrNodeUnhealthy
	}

	var wg sync.WaitGroup
	jobs := make(chan CacheNode)
	workers := dc.replicaWriteConcurrency(len(healthyNodes))
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for node := range jobs {
				if err := node.Cache().Set(ctx, key, cloneBytes(value), ttl); err != nil {
					errChan <- err
				}
			}
		}()
	}

	for _, node := range healthyNodes {
		jobs <- node
	}
	close(jobs)
	wg.Wait()
	close(errChan)

	// Return first error if any
	for err := range errChan {
		if firstErr == nil {
			firstErr = err
		}
		dc.recordReplicationFailure()
	}

	dc.recordReplicationLag(start)
	return firstErr
}

func (dc *DistributedCache) replicaWriteConcurrency(replicaCount int) int {
	if replicaCount <= 1 {
		return 1
	}
	limit := dc.asyncConcurrency
	if limit <= 0 {
		limit = defaultAsyncReplicationMaxConcurrency
	}
	if limit > replicaCount {
		return replicaCount
	}
	return limit
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

	err := primary.Cache().Set(ctx, key, cloneBytes(value), ttl)
	if err != nil {
		return err
	}

	// Write to replicas asynchronously
	for i := 1; i < len(nodes); i++ {
		node := nodes[i]
		if !node.IsHealthy() {
			dc.recordReplicationFailure()
			continue
		}

		replicaValue := append([]byte(nil), value...)
		replicaNode := node
		dc.scheduleAsyncReplication(asyncReplicationJob{
			operation: "set",
			key:       key,
			nodeID:    replicaNode.ID(),
			run: func(replicaCtx context.Context) error {
				return replicaNode.Cache().Set(replicaCtx, key, replicaValue, ttl)
			},
		})
	}

	return nil
}

func (dc *DistributedCache) incrReplicas(ctx context.Context, nodes []CacheNode, key string, delta int64) (int64, error) {
	if len(nodes) == 0 {
		return 0, ErrNoNodesAvailable
	}

	primary := nodes[0]
	if !primary.IsHealthy() {
		return 0, ErrNodeUnhealthy
	}

	value, err := primary.Cache().Incr(ctx, key, delta)
	if err != nil {
		return 0, err
	}

	switch dc.replicationMode {
	case ReplicationSync:
		if err := dc.incrSecondaryReplicas(ctx, nodes[1:], key, delta); err != nil {
			return value, err
		}
	case ReplicationAsync:
		for _, node := range nodes[1:] {
			if !node.IsHealthy() {
				dc.recordReplicationFailure()
				continue
			}
			replicaNode := node
			dc.scheduleAsyncReplication(asyncReplicationJob{
				operation: "incr",
				key:       key,
				nodeID:    replicaNode.ID(),
				run: func(replicaCtx context.Context) error {
					_, err := replicaNode.Cache().Incr(replicaCtx, key, delta)
					return err
				},
			})
		}
	}

	return value, nil
}

func (dc *DistributedCache) incrSecondaryReplicas(ctx context.Context, nodes []CacheNode, key string, delta int64) error {
	start := time.Now()
	defer dc.recordReplicationLag(start)

	for _, node := range nodes {
		if !node.IsHealthy() {
			dc.recordReplicationFailure()
			return ErrNodeUnhealthy
		}
		if _, err := node.Cache().Incr(ctx, key, delta); err != nil {
			dc.recordReplicationFailure()
			return err
		}
	}
	return nil
}

func (dc *DistributedCache) appendReplicas(ctx context.Context, nodes []CacheNode, key string, data []byte) error {
	if len(nodes) == 0 {
		return ErrNoNodesAvailable
	}

	primary := nodes[0]
	if !primary.IsHealthy() {
		return ErrNodeUnhealthy
	}

	if err := primary.Cache().Append(ctx, key, cloneBytes(data)); err != nil {
		return err
	}

	switch dc.replicationMode {
	case ReplicationSync:
		start := time.Now()
		for _, node := range nodes[1:] {
			if !node.IsHealthy() {
				dc.recordReplicationFailure()
				dc.recordReplicationLag(start)
				return ErrNodeUnhealthy
			}
			if err := node.Cache().Append(ctx, key, cloneBytes(data)); err != nil {
				dc.recordReplicationFailure()
				dc.recordReplicationLag(start)
				return err
			}
		}
		dc.recordReplicationLag(start)
	case ReplicationAsync:
		for _, node := range nodes[1:] {
			if !node.IsHealthy() {
				dc.recordReplicationFailure()
				continue
			}
			replicaData := append([]byte(nil), data...)
			replicaNode := node
			dc.scheduleAsyncReplication(asyncReplicationJob{
				operation: "append",
				key:       key,
				nodeID:    replicaNode.ID(),
				run: func(replicaCtx context.Context) error {
					return replicaNode.Cache().Append(replicaCtx, key, replicaData)
				},
			})
		}
	}

	return nil
}

func (dc *DistributedCache) replicationNodes(key string) ([]CacheNode, error) {
	if dc.replicationMode == ReplicationNone {
		node, err := dc.ring.Get(key)
		if err != nil {
			return nil, err
		}
		return []CacheNode{node}, nil
	}
	return dc.ring.GetN(key, dc.replicationFactor)
}

func cloneBytes(data []byte) []byte {
	if data == nil {
		return nil
	}
	return append([]byte(nil), data...)
}

func (dc *DistributedCache) startAsyncReplicationWorkers(count int) {
	for i := 0; i < count; i++ {
		dc.asyncWG.Add(1)
		go dc.asyncReplicationWorker()
	}
}

func (dc *DistributedCache) asyncReplicationWorker() {
	defer dc.asyncWG.Done()
	for {
		select {
		case <-dc.asyncStop:
			return
		case job := <-dc.asyncQueue:
			dc.runAsyncReplication(job)
		}
	}
}

func (dc *DistributedCache) ensureOpen() error {
	if dc != nil && dc.asyncClosed.Load() {
		return ErrClosed
	}
	return nil
}

func (dc *DistributedCache) scheduleAsyncReplication(job asyncReplicationJob) {
	if dc.asyncClosed.Load() {
		dc.dropAsyncReplication(job, AsyncReplicationDropClosed)
		return
	}
	select {
	case dc.asyncQueue <- job:
	case <-dc.asyncStop:
		dc.dropAsyncReplication(job, AsyncReplicationDropClosed)
	default:
		dc.dropAsyncReplication(job, AsyncReplicationDropQueueFull)
	}
}

func (dc *DistributedCache) drainAsyncReplicationQueue() {
	for {
		select {
		case job := <-dc.asyncQueue:
			dc.dropAsyncReplication(job, AsyncReplicationDropClosed)
		default:
			return
		}
	}
}

func (dc *DistributedCache) runAsyncReplication(job asyncReplicationJob) {
	start := time.Now()
	replicaCtx, cancel := dc.asyncReplicationContext()
	defer cancel()
	if err := job.run(replicaCtx); err != nil {
		dc.recordReplicationFailure()
	}
	dc.recordReplicationLag(start)
}

func (dc *DistributedCache) dropAsyncReplication(job asyncReplicationJob, reason AsyncReplicationDropReason) {
	dc.recordReplicationFailure()
	if dc.asyncDropHandler != nil {
		dc.callAsyncDropHandler(AsyncReplicationDrop{
			Operation: job.operation,
			Key:       job.key,
			NodeID:    job.nodeID,
			Reason:    reason,
		})
	}
}

func (dc *DistributedCache) callAsyncDropHandler(drop AsyncReplicationDrop) {
	defer func() {
		if recover() != nil {
			dc.recordReplicationFailure()
		}
	}()
	dc.asyncDropHandler(drop)
}

func (dc *DistributedCache) asyncReplicationContext() (context.Context, context.CancelFunc) {
	if dc.asyncTimeout <= 0 {
		return context.WithTimeout(context.Background(), defaultAsyncTimeout)
	}
	return context.WithTimeout(context.Background(), dc.asyncTimeout)
}

func (dc *DistributedCache) recordReplicationFailure() {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.ReplicationFailures, 1)
	}
}

func (dc *DistributedCache) recordReplicationLag(start time.Time) {
	if dc.metrics == nil {
		return
	}
	dc.metrics.mu.Lock()
	dc.metrics.ReplicationLag = time.Since(start)
	dc.metrics.mu.Unlock()
}

// failoverGet attempts to get a value from replica nodes
func (dc *DistributedCache) failoverGet(ctx context.Context, key string, failedNodeID string, cause error) ([]byte, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.FailoverCount, 1)
	}

	switch dc.failoverStrategy {
	case FailoverAllNodes:
		return dc.getFromNodes(ctx, key, failedNodeID, dc.ring.Nodes(), cause)
	case FailoverRetry:
		return dc.retryFailedNode(ctx, key, failedNodeID, cause)
	default:
		nodes, err := dc.replicationNodes(key)
		if err != nil {
			return nil, err
		}
		return dc.getFromNodes(ctx, key, failedNodeID, nodes, cause)
	}
}

func (dc *DistributedCache) failoverExists(ctx context.Context, key string, failedNodeID string, cause error) (bool, error) {
	if dc.metrics != nil {
		atomic.AddUint64(&dc.metrics.FailoverCount, 1)
	}

	switch dc.failoverStrategy {
	case FailoverAllNodes:
		return dc.existsInNodes(ctx, key, failedNodeID, dc.ring.Nodes(), cause)
	case FailoverRetry:
		var failedNode CacheNode
		for _, node := range dc.ring.Nodes() {
			if node.ID() == failedNodeID {
				failedNode = node
				break
			}
		}
		if failedNode == nil {
			return false, ErrNodeNotFound
		}
		if !failedNode.IsHealthy() {
			return false, ErrNodeUnhealthy
		}
		var lastErr error
		for i := 0; i < dc.failoverAttempts; i++ {
			exists, err := failedNode.Cache().Exists(ctx, key)
			if err == nil {
				return exists, nil
			}
			if errors.Is(err, cache.ErrNotFound) {
				return false, cache.ErrNotFound
			}
			lastErr = err
			if i == dc.failoverAttempts-1 {
				break
			}
			timer := time.NewTimer(dc.failoverBackoff)
			select {
			case <-ctx.Done():
				timer.Stop()
				return false, ctx.Err()
			case <-timer.C:
			}
		}
		if lastErr != nil {
			return false, lastErr
		}
		return false, cause
	default:
		nodes, err := dc.replicationNodes(key)
		if err != nil {
			return false, err
		}
		return dc.existsInNodes(ctx, key, failedNodeID, nodes, cause)
	}
}

func (dc *DistributedCache) existsInNodes(ctx context.Context, key string, failedNodeID string, nodes []CacheNode, cause error) (bool, error) {
	var firstErr error
	for _, node := range nodes {
		if node.ID() == failedNodeID || !node.IsHealthy() {
			continue
		}
		exists, err := node.Cache().Exists(ctx, key)
		if err == nil && exists {
			return true, nil
		}
		if err != nil && !errors.Is(err, cache.ErrNotFound) && firstErr == nil {
			firstErr = err
		}
	}
	if firstErr != nil {
		return false, firstErr
	}
	return false, cause
}

func (dc *DistributedCache) getFromNodes(ctx context.Context, key string, failedNodeID string, nodes []CacheNode, cause error) ([]byte, error) {
	var firstErr error
	for _, node := range nodes {
		if node.ID() == failedNodeID || !node.IsHealthy() {
			continue
		}

		value, err := node.Cache().Get(ctx, key)
		if err == nil {
			return cloneBytes(value), nil
		}

		if !errors.Is(err, cache.ErrNotFound) && firstErr == nil {
			firstErr = err
		}
	}

	if firstErr != nil {
		return nil, firstErr
	}
	if cause != nil && !errors.Is(cause, cache.ErrNotFound) {
		return nil, cause
	}
	return nil, cache.ErrNotFound
}

func (dc *DistributedCache) retryFailedNode(ctx context.Context, key string, failedNodeID string, cause error) ([]byte, error) {
	var failedNode CacheNode
	for _, node := range dc.ring.Nodes() {
		if node.ID() == failedNodeID {
			failedNode = node
			break
		}
	}
	if failedNode == nil {
		return nil, ErrNodeNotFound
	}
	if !failedNode.IsHealthy() {
		return nil, ErrNodeUnhealthy
	}

	var lastErr error
	for i := 0; i < dc.failoverAttempts; i++ {
		value, err := failedNode.Cache().Get(ctx, key)
		if err == nil {
			return cloneBytes(value), nil
		}
		if errors.Is(err, cache.ErrNotFound) {
			return nil, cache.ErrNotFound
		}
		lastErr = err
		if i == dc.failoverAttempts-1 {
			break
		}
		timer := time.NewTimer(dc.failoverBackoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, cause
}
