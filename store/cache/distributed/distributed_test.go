package distributed

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/spcent/plumego/store/cache"
)

func TestConsistentHashRingBasicOperations(t *testing.T) {
	ring := NewConsistentHashRing(nil)

	// Create test nodes
	node1 := NewNode("node1", cache.NewMemoryCache())
	node2 := NewNode("node2", cache.NewMemoryCache())
	node3 := NewNode("node3", cache.NewMemoryCache())

	// Test Add
	if err := ring.Add(node1); err != nil {
		t.Errorf("failed to add node1: %v", err)
	}

	if err := ring.Add(node2); err != nil {
		t.Errorf("failed to add node2: %v", err)
	}

	if err := ring.Add(node3); err != nil {
		t.Errorf("failed to add node3: %v", err)
	}

	// Test Size
	if ring.Size() != 3 {
		t.Errorf("expected 3 nodes, got %d", ring.Size())
	}

	// Test duplicate add
	err := ring.Add(node1)
	if err != ErrNodeAlreadyExists {
		t.Errorf("expected ErrNodeAlreadyExists, got %v", err)
	}

	// Test Remove
	if err := ring.Remove("node2"); err != nil {
		t.Errorf("failed to remove node2: %v", err)
	}

	if ring.Size() != 2 {
		t.Errorf("expected 2 nodes after removal, got %d", ring.Size())
	}

	// Test remove non-existent
	err = ring.Remove("nonexistent")
	if err != ErrNodeNotFound {
		t.Errorf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestConsistentHashRingGet(t *testing.T) {
	ring := NewConsistentHashRing(nil)

	// Create and add nodes
	node1 := NewNode("node1", cache.NewMemoryCache())
	node2 := NewNode("node2", cache.NewMemoryCache())
	node3 := NewNode("node3", cache.NewMemoryCache())

	ring.Add(node1)
	ring.Add(node2)
	ring.Add(node3)

	// Test Get - same key should always go to same node
	key := "test-key"
	firstNode, err := ring.Get(key)
	if err != nil {
		t.Fatalf("failed to get node: %v", err)
	}

	for i := 0; i < 10; i++ {
		node, err := ring.Get(key)
		if err != nil {
			t.Errorf("failed to get node: %v", err)
		}

		if node.ID() != firstNode.ID() {
			t.Errorf("inconsistent routing: expected %s, got %s", firstNode.ID(), node.ID())
		}
	}
}

func TestConsistentHashRingGetN(t *testing.T) {
	ring := NewConsistentHashRing(nil)

	// Create and add nodes
	nodes := make([]CacheNode, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
		ring.Add(nodes[i])
	}

	// Test GetN
	key := "test-key"
	replicas, err := ring.GetN(key, 3)
	if err != nil {
		t.Fatalf("failed to get replicas: %v", err)
	}

	if len(replicas) != 3 {
		t.Errorf("expected 3 replicas, got %d", len(replicas))
	}

	// Check for unique nodes
	seen := make(map[string]bool)
	for _, node := range replicas {
		if seen[node.ID()] {
			t.Errorf("duplicate node in replicas: %s", node.ID())
		}
		seen[node.ID()] = true
	}
}

func TestConsistentHashRingDistribution(t *testing.T) {
	// Use more virtual nodes for better distribution
	config := &ConsistentHashRingConfig{
		VirtualNodes: 300,
		HashFunc:     nil,
	}
	ring := NewConsistentHashRing(config)

	// Add nodes
	numNodes := 3
	for i := 0; i < numNodes; i++ {
		node := NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
		ring.Add(node)
	}

	// Test distribution
	distribution := make(map[string]int)
	numKeys := 10000

	for i := 0; i < numKeys; i++ {
		key := fmt.Sprintf("key%d", i)
		node, err := ring.Get(key)
		if err != nil {
			t.Fatalf("failed to get node: %v", err)
		}

		distribution[node.ID()]++
	}

	// Check distribution variance
	expectedPerNode := numKeys / numNodes
	for nodeID, count := range distribution {
		variance := float64(count-expectedPerNode) / float64(expectedPerNode) * 100
		t.Logf("Node %s: %d keys (%.2f%% variance)", nodeID, count, variance)

		// Allow up to 30% variance for small node count
		// With more nodes and virtual nodes, variance decreases
		if variance > 30 || variance < -30 {
			t.Errorf("node %s has too much variance: %.2f%%", nodeID, variance)
		}
	}
}

func TestLocalCacheNode(t *testing.T) {
	cacheInstance := cache.NewMemoryCache()
	node := NewNode("test-node", cacheInstance, WithWeight(2))

	// Test ID
	if node.ID() != "test-node" {
		t.Errorf("expected ID 'test-node', got %s", node.ID())
	}

	// Test Weight
	if node.Weight() != 2 {
		t.Errorf("expected weight 2, got %d", node.Weight())
	}

	// Test initial health
	if !node.IsHealthy() {
		t.Error("expected node to be initially healthy")
	}

	if node.HealthStatus() != HealthStatusHealthy {
		t.Errorf("expected HealthStatusHealthy, got %v", node.HealthStatus())
	}

	// Test UpdateHealth
	node.UpdateHealth(HealthStatusUnhealthy)
	if node.IsHealthy() {
		t.Error("expected node to be unhealthy")
	}

	if node.HealthStatus() != HealthStatusUnhealthy {
		t.Errorf("expected HealthStatusUnhealthy, got %v", node.HealthStatus())
	}
}

func TestDistributedCacheBasicOperations(t *testing.T) {
	// Create nodes
	nodes := make([]CacheNode, 3)
	for i := 0; i < 3; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
	}

	// Create distributed cache
	config := DefaultConfig()
	config.ReplicationFactor = 1
	dc := New(nodes, config)
	defer dc.Close()

	ctx := context.Background()

	// Test Set
	err := dc.Set(ctx, "key1", []byte("value1"), time.Minute)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	// Test Get
	value, err := dc.Get(ctx, "key1")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	if string(value) != "value1" {
		t.Errorf("expected value1, got %s", string(value))
	}

	// Test Exists
	exists, err := dc.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}

	if !exists {
		t.Error("expected key to exist")
	}

	// Test Delete
	err = dc.Delete(ctx, "key1")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify deletion
	exists, err = dc.Exists(ctx, "key1")
	if err != nil {
		t.Errorf("Exists failed: %v", err)
	}

	if exists {
		t.Error("expected key to not exist after deletion")
	}
}

func TestDistributedCacheReplication(t *testing.T) {
	// Create nodes
	nodes := make([]CacheNode, 3)
	for i := 0; i < 3; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
	}

	// Create distributed cache with replication
	config := DefaultConfig()
	config.ReplicationFactor = 2
	config.ReplicationMode = ReplicationSync
	dc := New(nodes, config)
	defer dc.Close()

	ctx := context.Background()

	// Set a key
	err := dc.Set(ctx, "replicated-key", []byte("replicated-value"), time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify the key is on multiple nodes
	// Note: We can't directly verify this without exposing internals,
	// but we can test failover
}

func TestDistributedCacheFailover(t *testing.T) {
	// Create nodes
	nodes := make([]CacheNode, 3)
	for i := 0; i < 3; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
	}

	// Create distributed cache with replication
	config := DefaultConfig()
	config.ReplicationFactor = 2
	config.ReplicationMode = ReplicationSync
	dc := New(nodes, config)
	defer dc.Close()

	ctx := context.Background()

	// Set a key
	key := "failover-key"
	value := []byte("failover-value")
	err := dc.Set(ctx, key, value, time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get the primary node for this key
	primaryNode, err := dc.ring.Get(key)
	if err != nil {
		t.Fatalf("failed to get primary node: %v", err)
	}

	// Mark primary node as unhealthy
	primaryNode.UpdateHealth(HealthStatusUnhealthy)

	// Try to get the value (should failover to replica)
	retrievedValue, err := dc.Get(ctx, key)
	if err != nil {
		t.Errorf("Get with failover failed: %v", err)
	}

	if string(retrievedValue) != string(value) {
		t.Errorf("expected %s, got %s", string(value), string(retrievedValue))
	}

	// Verify failover metric was incremented
	metrics := dc.GetMetrics()
	if metrics.FailoverCount == 0 {
		t.Error("expected failover count > 0")
	}
}

func TestDistributedCacheNodeManagement(t *testing.T) {
	// Create initial nodes
	nodes := make([]CacheNode, 2)
	for i := 0; i < 2; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	// Verify initial node count
	if len(dc.Nodes()) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(dc.Nodes()))
	}

	// Add a new node
	newNode := NewNode("node2", cache.NewMemoryCache())
	err := dc.AddNode(newNode)
	if err != nil {
		t.Errorf("AddNode failed: %v", err)
	}

	if len(dc.Nodes()) != 3 {
		t.Errorf("expected 3 nodes after add, got %d", len(dc.Nodes()))
	}

	// Remove a node
	err = dc.RemoveNode("node1")
	if err != nil {
		t.Errorf("RemoveNode failed: %v", err)
	}

	if len(dc.Nodes()) != 2 {
		t.Errorf("expected 2 nodes after removal, got %d", len(dc.Nodes()))
	}

	// Verify rebalance events
	metrics := dc.GetMetrics()
	if metrics.RebalanceEvents != 2 {
		t.Errorf("expected 2 rebalance events, got %d", metrics.RebalanceEvents)
	}
}

func TestDistributedCacheIncr(t *testing.T) {
	nodes := []CacheNode{
		NewNode("node1", cache.NewMemoryCache()),
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	ctx := context.Background()

	// Test Incr
	val, err := dc.Incr(ctx, "counter", 5)
	if err != nil {
		t.Errorf("Incr failed: %v", err)
	}

	if val != 5 {
		t.Errorf("expected 5, got %d", val)
	}

	// Incr again
	val, err = dc.Incr(ctx, "counter", 3)
	if err != nil {
		t.Errorf("Incr failed: %v", err)
	}

	if val != 8 {
		t.Errorf("expected 8, got %d", val)
	}
}

func TestDistributedCacheDecr(t *testing.T) {
	nodes := []CacheNode{
		NewNode("node1", cache.NewMemoryCache()),
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	ctx := context.Background()

	// Test Decr
	val, err := dc.Decr(ctx, "counter", 5)
	if err != nil {
		t.Errorf("Decr failed: %v", err)
	}

	if val != -5 {
		t.Errorf("expected -5, got %d", val)
	}
}

func TestDistributedCacheConcurrency(t *testing.T) {
	nodes := make([]CacheNode, 3)
	for i := 0; i < 3; i++ {
		nodes[i] = NewNode(fmt.Sprintf("node%d", i), cache.NewMemoryCache())
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	ctx := context.Background()
	const numGoroutines = 100
	const numOpsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < numOpsPerGoroutine; j++ {
				key := fmt.Sprintf("key%d-%d", id, j)
				value := []byte(fmt.Sprintf("value%d-%d", id, j))

				// Set
				_ = dc.Set(ctx, key, value, time.Minute)

				// Get
				_, _ = dc.Get(ctx, key)

				// Delete
				_ = dc.Delete(ctx, key)
			}
		}(i)
	}

	wg.Wait()

	// Verify metrics
	metrics := dc.GetMetrics()
	expectedRequests := uint64(numGoroutines * numOpsPerGoroutine * 3) // Set + Get + Delete
	if metrics.TotalRequests != expectedRequests {
		t.Errorf("expected %d requests, got %d", expectedRequests, metrics.TotalRequests)
	}
}

func TestDistributedCacheMetrics(t *testing.T) {
	nodes := []CacheNode{
		NewNode("node1", cache.NewMemoryCache()),
		NewNode("node2", cache.NewMemoryCache()),
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	ctx := context.Background()

	// Perform operations
	_ = dc.Set(ctx, "key1", []byte("value1"), time.Minute)
	_, _ = dc.Get(ctx, "key1")
	_ = dc.Delete(ctx, "key1")

	// Get metrics
	metrics := dc.GetMetrics()

	if metrics.TotalRequests != 3 {
		t.Errorf("expected 3 requests, got %d", metrics.TotalRequests)
	}

	if metrics.HealthyNodes != 2 {
		t.Errorf("expected 2 healthy nodes, got %d", metrics.HealthyNodes)
	}

	if metrics.UnhealthyNodes != 0 {
		t.Errorf("expected 0 unhealthy nodes, got %d", metrics.UnhealthyNodes)
	}
}

func TestHealthChecker(t *testing.T) {
	config := DefaultHealthCheckerConfig()
	config.CheckInterval = 100 * time.Millisecond

	hc := NewHealthChecker(config)

	// Create test node
	node := NewNode("test-node", cache.NewMemoryCache())
	hc.AddNode(node)

	// Start health checking
	hc.Start()
	defer hc.Stop()

	// Wait for a health check
	time.Sleep(200 * time.Millisecond)

	// Node should still be healthy
	status, err := hc.GetNodeStatus("test-node")
	if err != nil {
		t.Errorf("failed to get node status: %v", err)
	}

	if status != HealthStatusHealthy {
		t.Errorf("expected healthy status, got %v", status)
	}
}

func TestDistributedCacheAppend(t *testing.T) {
	nodes := []CacheNode{
		NewNode("node1", cache.NewMemoryCache()),
	}

	dc := New(nodes, DefaultConfig())
	defer dc.Close()

	ctx := context.Background()

	// Append to non-existent key
	err := dc.Append(ctx, "append-key", []byte("hello"))
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}

	// Append again
	err = dc.Append(ctx, "append-key", []byte(" world"))
	if err != nil {
		t.Errorf("Append failed: %v", err)
	}

	// Verify result
	value, err := dc.Get(ctx, "append-key")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}

	expected := "hello world"
	if string(value) != expected {
		t.Errorf("expected %s, got %s", expected, string(value))
	}
}
