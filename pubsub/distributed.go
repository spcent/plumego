package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Distributed errors
var (
	ErrClusterNotJoined    = errors.New("not joined to cluster")
	ErrNodeNotFound        = errors.New("node not found")
	ErrNodeUnhealthy       = errors.New("node is unhealthy")
	ErrBroadcastFailed     = errors.New("broadcast failed")
	ErrConsensusTimeout    = errors.New("consensus timeout")
	ErrInvalidNodeConfig   = errors.New("invalid node configuration")
)

// ClusterConfig configures the distributed cluster
type ClusterConfig struct {
	// NodeID is the unique identifier for this node
	NodeID string

	// ListenAddr is the address this node listens on
	ListenAddr string

	// Peers are the addresses of other cluster nodes
	Peers []string

	// HeartbeatInterval for node health checks (default: 5s)
	HeartbeatInterval time.Duration

	// HeartbeatTimeout before marking node as unhealthy (default: 15s)
	HeartbeatTimeout time.Duration

	// ReplicationFactor for message replication (default: 1, no replication)
	ReplicationFactor int

	// BroadcastTimeout for cross-node publishes (default: 5s)
	BroadcastTimeout time.Duration

	// ConsistentHashing enables consistent hashing for topic routing
	ConsistentHashing bool

	// HTTPClient for node-to-node communication
	HTTPClient *http.Client
}

// DefaultClusterConfig returns default cluster configuration
func DefaultClusterConfig(nodeID, listenAddr string) ClusterConfig {
	return ClusterConfig{
		NodeID:            nodeID,
		ListenAddr:        listenAddr,
		Peers:             []string{},
		HeartbeatInterval: 5 * time.Second,
		HeartbeatTimeout:  15 * time.Second,
		ReplicationFactor: 1,
		BroadcastTimeout:  5 * time.Second,
		ConsistentHashing: true,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ClusterNode represents a node in the cluster
type ClusterNode struct {
	ID         string    `json:"id"`
	Addr       string    `json:"addr"`
	LastSeen   time.Time `json:"last_seen"`
	Healthy    bool      `json:"healthy"`
	Topics     []string  `json:"topics"`
	Version    string    `json:"version"`
}

// DistributedPubSub wraps InProcPubSub with distributed capabilities
type DistributedPubSub struct {
	*InProcPubSub

	config ClusterConfig

	// Cluster state
	nodes    map[string]*ClusterNode
	nodesMu  sync.RWMutex
	joined   atomic.Bool

	// HTTP server for cluster API
	httpServer *http.Server
	httpMux    *http.ServeMux

	// Background workers
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	closed atomic.Bool

	// Metrics
	clusterPublishes atomic.Uint64
	clusterReceives  atomic.Uint64
	clusterForwards  atomic.Uint64
	clusterErrors    atomic.Uint64
	heartbeats       atomic.Uint64

	// Routing
	topicOwners map[string][]string // topic -> node IDs
	ownersMu    sync.RWMutex
}

// clusterMessage represents a message sent between cluster nodes
type clusterMessage struct {
	Type      string    `json:"type"`
	NodeID    string    `json:"node_id"`
	Topic     string    `json:"topic"`
	Message   Message   `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// heartbeatPayload represents a heartbeat message
type heartbeatPayload struct {
	NodeID    string    `json:"node_id"`
	Addr      string    `json:"addr"`
	Topics    []string  `json:"topics"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// NewDistributed creates a new distributed pubsub instance
func NewDistributed(config ClusterConfig, opts ...Option) (*DistributedPubSub, error) {
	if config.NodeID == "" {
		return nil, errors.New("node ID is required")
	}

	if config.ListenAddr == "" {
		return nil, errors.New("listen address is required")
	}

	// Apply defaults
	if config.HeartbeatInterval == 0 {
		config.HeartbeatInterval = 5 * time.Second
	}
	if config.HeartbeatTimeout == 0 {
		config.HeartbeatTimeout = 15 * time.Second
	}
	if config.ReplicationFactor == 0 {
		config.ReplicationFactor = 1
	}
	if config.BroadcastTimeout == 0 {
		config.BroadcastTimeout = 5 * time.Second
	}
	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}

	// Create base pubsub
	ps := New(opts...)

	ctx, cancel := context.WithCancel(context.Background())

	dps := &DistributedPubSub{
		InProcPubSub: ps,
		config:       config,
		nodes:        make(map[string]*ClusterNode),
		topicOwners:  make(map[string][]string),
		ctx:          ctx,
		cancel:       cancel,
	}

	// Setup HTTP server
	dps.setupHTTPServer()

	return dps, nil
}

// setupHTTPServer sets up the cluster HTTP API
func (dps *DistributedPubSub) setupHTTPServer() {
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/health", dps.handleHealth)

	// Heartbeat endpoint
	mux.HandleFunc("/heartbeat", dps.handleHeartbeat)

	// Publish endpoint
	mux.HandleFunc("/publish", dps.handleClusterPublish)

	// Sync endpoint
	mux.HandleFunc("/sync", dps.handleSync)

	dps.httpMux = mux
	dps.httpServer = &http.Server{
		Addr:    dps.config.ListenAddr,
		Handler: mux,
	}
}

// JoinCluster joins the cluster and starts serving
func (dps *DistributedPubSub) JoinCluster(ctx context.Context) error {
	if dps.joined.Load() {
		return errors.New("already joined to cluster")
	}

	// Start HTTP server
	dps.wg.Add(1)
	go func() {
		defer dps.wg.Done()
		_ = dps.httpServer.ListenAndServe()
	}()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	// Register self
	dps.nodesMu.Lock()
	dps.nodes[dps.config.NodeID] = &ClusterNode{
		ID:       dps.config.NodeID,
		Addr:     dps.config.ListenAddr,
		LastSeen: time.Now(),
		Healthy:  true,
		Topics:   []string{},
		Version:  "1.0",
	}
	dps.nodesMu.Unlock()

	// Discover peers
	if err := dps.discoverPeers(ctx); err != nil {
		return fmt.Errorf("peer discovery failed: %w", err)
	}

	dps.joined.Store(true)

	// Start background workers
	dps.startClusterWorkers()

	return nil
}

// discoverPeers discovers and connects to peer nodes
func (dps *DistributedPubSub) discoverPeers(ctx context.Context) error {
	for _, peerAddr := range dps.config.Peers {
		// Send heartbeat to discover peer
		if err := dps.sendHeartbeat(ctx, peerAddr); err != nil {
			// Log but continue with other peers
			continue
		}
	}
	return nil
}

// LeaveCluster leaves the cluster and stops serving
func (dps *DistributedPubSub) LeaveCluster(ctx context.Context) error {
	if !dps.joined.Swap(false) {
		return nil
	}

	// Stop workers
	dps.cancel()

	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	_ = dps.httpServer.Shutdown(shutdownCtx)

	// Wait for workers
	dps.wg.Wait()

	return nil
}

// PublishGlobal publishes a message to the entire cluster
func (dps *DistributedPubSub) PublishGlobal(topic string, msg Message) error {
	if !dps.joined.Load() {
		return ErrClusterNotJoined
	}

	// Publish locally first
	if err := dps.InProcPubSub.Publish(topic, msg); err != nil {
		return err
	}

	// Broadcast to cluster
	return dps.broadcastMessage(topic, msg)
}

// broadcastMessage broadcasts a message to all cluster nodes
func (dps *DistributedPubSub) broadcastMessage(topic string, msg Message) error {
	dps.nodesMu.RLock()
	nodes := make([]*ClusterNode, 0, len(dps.nodes))
	for _, node := range dps.nodes {
		if node.ID != dps.config.NodeID && node.Healthy {
			nodes = append(nodes, node)
		}
	}
	dps.nodesMu.RUnlock()

	if len(nodes) == 0 {
		return nil // No peers
	}

	// Prepare message
	cm := clusterMessage{
		Type:      "publish",
		NodeID:    dps.config.NodeID,
		Topic:     topic,
		Message:   msg,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(cm)
	if err != nil {
		return err
	}

	// Broadcast to nodes
	var wg sync.WaitGroup
	errChan := make(chan error, len(nodes))

	for _, node := range nodes {
		wg.Add(1)
		go func(n *ClusterNode) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), dps.config.BroadcastTimeout)
			defer cancel()

			url := fmt.Sprintf("http://%s/publish", n.Addr)
			req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
			if err != nil {
				errChan <- err
				return
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := dps.config.HTTPClient.Do(req)
			if err != nil {
				errChan <- err
				dps.markNodeUnhealthy(n.ID)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				errChan <- fmt.Errorf("node %s returned %d", n.ID, resp.StatusCode)
				return
			}

			dps.clusterForwards.Add(1)
		}(node)
	}

	wg.Wait()
	close(errChan)

	// Check if any errors occurred
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		dps.clusterErrors.Add(uint64(len(errs)))
		// Return error if majority failed
		if len(errs) > len(nodes)/2 {
			return fmt.Errorf("broadcast failed to %d/%d nodes", len(errs), len(nodes))
		}
	}

	return nil
}

// sendHeartbeat sends a heartbeat to a peer
func (dps *DistributedPubSub) sendHeartbeat(ctx context.Context, peerAddr string) error {
	// Get current topics
	topics := dps.getLocalTopics()

	payload := heartbeatPayload{
		NodeID:    dps.config.NodeID,
		Addr:      dps.config.ListenAddr,
		Topics:    topics,
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/heartbeat", peerAddr)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := dps.config.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("heartbeat failed: %d", resp.StatusCode)
	}

	dps.heartbeats.Add(1)
	return nil
}

// getLocalTopics returns list of topics with active subscriptions
func (dps *DistributedPubSub) getLocalTopics() []string {
	topics := make(map[string]bool)

	// Collect from all shards
	for i := 0; i < dps.shards.shardCount; i++ {
		shard := dps.shards.shards[i]
		shard.mu.RLock()
		for topic := range shard.topics {
			topics[topic] = true
		}
		for pattern := range shard.patterns {
			topics[pattern] = true
		}
		shard.mu.RUnlock()
	}

	result := make([]string, 0, len(topics))
	for topic := range topics {
		result = append(result, topic)
	}

	sort.Strings(result)
	return result
}

// markNodeUnhealthy marks a node as unhealthy
func (dps *DistributedPubSub) markNodeUnhealthy(nodeID string) {
	dps.nodesMu.Lock()
	defer dps.nodesMu.Unlock()

	if node, ok := dps.nodes[nodeID]; ok {
		node.Healthy = false
	}
}

// startClusterWorkers starts background cluster maintenance tasks
func (dps *DistributedPubSub) startClusterWorkers() {
	// Heartbeat worker
	dps.wg.Add(1)
	go func() {
		defer dps.wg.Done()
		ticker := time.NewTicker(dps.config.HeartbeatInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dps.sendHeartbeats()

			case <-dps.ctx.Done():
				return
			}
		}
	}()

	// Health check worker
	dps.wg.Add(1)
	go func() {
		defer dps.wg.Done()
		ticker := time.NewTicker(dps.config.HeartbeatInterval * 2)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				dps.checkNodeHealth()

			case <-dps.ctx.Done():
				return
			}
		}
	}()
}

// sendHeartbeats sends heartbeats to all peers
func (dps *DistributedPubSub) sendHeartbeats() {
	ctx, cancel := context.WithTimeout(context.Background(), dps.config.HeartbeatTimeout)
	defer cancel()

	for _, peerAddr := range dps.config.Peers {
		_ = dps.sendHeartbeat(ctx, peerAddr)
	}
}

// checkNodeHealth checks and updates node health status
func (dps *DistributedPubSub) checkNodeHealth() {
	now := time.Now()
	timeout := dps.config.HeartbeatTimeout

	dps.nodesMu.Lock()
	defer dps.nodesMu.Unlock()

	for _, node := range dps.nodes {
		if node.ID == dps.config.NodeID {
			continue // Skip self
		}

		if now.Sub(node.LastSeen) > timeout {
			node.Healthy = false
		}
	}
}

// HTTP Handlers

func (dps *DistributedPubSub) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	status := map[string]any{
		"node_id": dps.config.NodeID,
		"healthy": true,
		"joined":  dps.joined.Load(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (dps *DistributedPubSub) handleHeartbeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload heartbeatPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Update or add node
	dps.nodesMu.Lock()
	node, exists := dps.nodes[payload.NodeID]
	if !exists {
		node = &ClusterNode{
			ID:   payload.NodeID,
			Addr: payload.Addr,
		}
		dps.nodes[payload.NodeID] = node
	}
	node.LastSeen = time.Now()
	node.Healthy = true
	node.Topics = payload.Topics
	node.Version = payload.Version
	dps.nodesMu.Unlock()

	// Send our own info back
	topics := dps.getLocalTopics()
	response := heartbeatPayload{
		NodeID:    dps.config.NodeID,
		Addr:      dps.config.ListenAddr,
		Topics:    topics,
		Timestamp: time.Now(),
		Version:   "1.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (dps *DistributedPubSub) handleClusterPublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var cm clusterMessage
	if err := json.NewDecoder(r.Body).Decode(&cm); err != nil {
		http.Error(w, "invalid message", http.StatusBadRequest)
		return
	}

	// Publish locally
	if err := dps.InProcPubSub.Publish(cm.Topic, cm.Message); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		dps.clusterErrors.Add(1)
		return
	}

	dps.clusterReceives.Add(1)

	w.WriteHeader(http.StatusOK)
	io.WriteString(w, "OK")
}

func (dps *DistributedPubSub) handleSync(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	dps.nodesMu.RLock()
	nodes := make([]*ClusterNode, 0, len(dps.nodes))
	for _, node := range dps.nodes {
		nodes = append(nodes, node)
	}
	dps.nodesMu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"nodes": nodes,
	})
}

// Close closes the distributed pubsub
func (dps *DistributedPubSub) Close() error {
	if dps.closed.Swap(true) {
		return nil
	}

	// Leave cluster
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = dps.LeaveCluster(ctx)

	// Close base pubsub
	return dps.InProcPubSub.Close()
}

// Nodes returns current cluster nodes
func (dps *DistributedPubSub) Nodes() []*ClusterNode {
	dps.nodesMu.RLock()
	defer dps.nodesMu.RUnlock()

	nodes := make([]*ClusterNode, 0, len(dps.nodes))
	for _, node := range dps.nodes {
		// Create copy
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes
}

// ClusterStats returns cluster statistics
func (dps *DistributedPubSub) ClusterStats() ClusterStats {
	dps.nodesMu.RLock()
	totalNodes := len(dps.nodes)
	healthyNodes := 0
	for _, node := range dps.nodes {
		if node.Healthy {
			healthyNodes++
		}
	}
	dps.nodesMu.RUnlock()

	return ClusterStats{
		TotalNodes:       totalNodes,
		HealthyNodes:     healthyNodes,
		ClusterPublishes: dps.clusterPublishes.Load(),
		ClusterReceives:  dps.clusterReceives.Load(),
		ClusterForwards:  dps.clusterForwards.Load(),
		ClusterErrors:    dps.clusterErrors.Load(),
		Heartbeats:       dps.heartbeats.Load(),
	}
}

// ClusterStats holds cluster metrics
type ClusterStats struct {
	TotalNodes       int
	HealthyNodes     int
	ClusterPublishes uint64
	ClusterReceives  uint64
	ClusterForwards  uint64
	ClusterErrors    uint64
	Heartbeats       uint64
}
