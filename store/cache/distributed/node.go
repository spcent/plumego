package distributed

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spcent/plumego/store/cache"
)

var (
	// ErrNodeUnhealthy is returned when a node is unhealthy
	ErrNodeUnhealthy = errors.New("distributed: node is unhealthy")
)

// HealthStatus represents the health status of a node
type HealthStatus int

const (
	// HealthStatusHealthy indicates the node is healthy
	HealthStatusHealthy HealthStatus = iota

	// HealthStatusDegraded indicates the node is experiencing issues
	HealthStatusDegraded

	// HealthStatusUnhealthy indicates the node is unhealthy
	HealthStatusUnhealthy
)

func (s HealthStatus) String() string {
	switch s {
	case HealthStatusHealthy:
		return "healthy"
	case HealthStatusDegraded:
		return "degraded"
	case HealthStatusUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

// CacheNode represents a cache instance in a distributed cache cluster
type CacheNode interface {
	// ID returns the unique identifier for this node
	ID() string

	// Cache returns the underlying cache instance
	Cache() cache.Cache

	// Weight returns the weight of this node for load balancing (1 = normal)
	Weight() int

	// IsHealthy returns whether this node is currently healthy
	IsHealthy() bool

	// HealthStatus returns the current health status
	HealthStatus() HealthStatus

	// UpdateHealth updates the health status of this node
	UpdateHealth(status HealthStatus)
}

// LocalCacheNode implements CacheNode for a local cache instance
type LocalCacheNode struct {
	id           string
	cache        cache.Cache
	weight       int
	healthStatus atomic.Int32 // Stores HealthStatus as int32
	mu           sync.RWMutex
}

// NodeConfig configures a cache node
type NodeConfig struct {
	ID     string       // Unique node identifier
	Cache  cache.Cache  // Cache instance
	Weight int          // Node weight for load balancing (default: 1)
}

// NewNode creates a new local cache node
func NewNode(id string, cacheInstance cache.Cache, opts ...NodeOption) *LocalCacheNode {
	node := &LocalCacheNode{
		id:     id,
		cache:  cacheInstance,
		weight: 1,
	}

	node.healthStatus.Store(int32(HealthStatusHealthy))

	// Apply options
	for _, opt := range opts {
		opt(node)
	}

	return node
}

// NodeOption is a functional option for configuring a cache node
type NodeOption func(*LocalCacheNode)

// WithWeight sets the weight of the node
func WithWeight(weight int) NodeOption {
	return func(n *LocalCacheNode) {
		if weight > 0 {
			n.weight = weight
		}
	}
}

// ID returns the unique identifier for this node
func (n *LocalCacheNode) ID() string {
	return n.id
}

// Cache returns the underlying cache instance
func (n *LocalCacheNode) Cache() cache.Cache {
	return n.cache
}

// Weight returns the weight of this node
func (n *LocalCacheNode) Weight() int {
	return n.weight
}

// IsHealthy returns whether this node is currently healthy
func (n *LocalCacheNode) IsHealthy() bool {
	status := HealthStatus(n.healthStatus.Load())
	return status == HealthStatusHealthy || status == HealthStatusDegraded
}

// HealthStatus returns the current health status
func (n *LocalCacheNode) HealthStatus() HealthStatus {
	return HealthStatus(n.healthStatus.Load())
}

// UpdateHealth updates the health status of this node
func (n *LocalCacheNode) UpdateHealth(status HealthStatus) {
	n.healthStatus.Store(int32(status))
}

// HealthChecker performs health checks on cache nodes
type HealthChecker struct {
	mu              sync.RWMutex
	nodes           map[string]CacheNode
	checkInterval   time.Duration
	checkTimeout    time.Duration
	stopChan        chan struct{}
	wg              sync.WaitGroup
	failureCallback func(nodeID string, err error)
}

// HealthCheckerConfig configures the health checker
type HealthCheckerConfig struct {
	CheckInterval   time.Duration // How often to check (default: 10s)
	CheckTimeout    time.Duration // Timeout for each check (default: 2s)
	FailureCallback func(nodeID string, err error)
}

// DefaultHealthCheckerConfig returns the default health checker configuration
func DefaultHealthCheckerConfig() *HealthCheckerConfig {
	return &HealthCheckerConfig{
		CheckInterval: 10 * time.Second,
		CheckTimeout:  2 * time.Second,
	}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *HealthCheckerConfig) *HealthChecker {
	if config == nil {
		config = DefaultHealthCheckerConfig()
	}

	if config.CheckInterval <= 0 {
		config.CheckInterval = 10 * time.Second
	}

	if config.CheckTimeout <= 0 {
		config.CheckTimeout = 2 * time.Second
	}

	return &HealthChecker{
		nodes:           make(map[string]CacheNode),
		checkInterval:   config.CheckInterval,
		checkTimeout:    config.CheckTimeout,
		stopChan:        make(chan struct{}),
		failureCallback: config.FailureCallback,
	}
}

// AddNode adds a node to be monitored
func (hc *HealthChecker) AddNode(node CacheNode) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.nodes[node.ID()] = node
}

// RemoveNode removes a node from monitoring
func (hc *HealthChecker) RemoveNode(nodeID string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	delete(hc.nodes, nodeID)
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	hc.wg.Add(1)
	go hc.checkLoop()
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stopChan)
	hc.wg.Wait()
}

// checkLoop periodically checks node health
func (hc *HealthChecker) checkLoop() {
	defer hc.wg.Done()

	ticker := time.NewTicker(hc.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAllNodes()
		case <-hc.stopChan:
			return
		}
	}
}

// checkAllNodes checks the health of all nodes
func (hc *HealthChecker) checkAllNodes() {
	hc.mu.RLock()
	nodes := make([]CacheNode, 0, len(hc.nodes))
	for _, node := range hc.nodes {
		nodes = append(nodes, node)
	}
	hc.mu.RUnlock()

	for _, node := range nodes {
		hc.checkNode(node)
	}
}

// checkNode checks the health of a single node
func (hc *HealthChecker) checkNode(node CacheNode) {
	ctx, cancel := context.WithTimeout(context.Background(), hc.checkTimeout)
	defer cancel()

	// Try to check if a test key exists (lightweight operation)
	testKey := "__health_check__"
	_, err := node.Cache().Exists(ctx, testKey)

	if err != nil {
		// Node is unhealthy
		node.UpdateHealth(HealthStatusUnhealthy)

		// Call failure callback if set
		if hc.failureCallback != nil {
			hc.failureCallback(node.ID(), err)
		}
	} else {
		// Node is healthy
		node.UpdateHealth(HealthStatusHealthy)
	}
}

// GetNodeStatus returns the health status of a node
func (hc *HealthChecker) GetNodeStatus(nodeID string) (HealthStatus, error) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	node, exists := hc.nodes[nodeID]
	if !exists {
		return HealthStatusUnhealthy, ErrNodeNotFound
	}

	return node.HealthStatus(), nil
}
