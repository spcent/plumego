package distributed

import (
	"errors"
	"fmt"
	"hash/fnv"
	"sort"
	"sync"
)

var (
	// ErrNoNodesAvailable is returned when no nodes are available in the hash ring
	ErrNoNodesAvailable = errors.New("distributed: no nodes available")

	// ErrNodeNotFound is returned when a node is not found in the hash ring
	ErrNodeNotFound = errors.New("distributed: node not found")

	// ErrNodeAlreadyExists is returned when trying to add a duplicate node
	ErrNodeAlreadyExists = errors.New("distributed: node already exists")
)

// HashFunc is a function that hashes a key to a uint32
type HashFunc func(data []byte) uint32

// HashRing provides consistent hashing for distributing keys across nodes
type HashRing interface {
	// Add adds a node to the hash ring
	Add(node CacheNode) error

	// Remove removes a node from the hash ring
	Remove(nodeID string) error

	// Get returns the node responsible for the given key
	Get(key string) (CacheNode, error)

	// GetN returns the N nodes responsible for the given key (for replication)
	GetN(key string, n int) ([]CacheNode, error)

	// Nodes returns all nodes in the hash ring
	Nodes() []CacheNode

	// Size returns the number of physical nodes
	Size() int
}

// ConsistentHashRing implements HashRing using consistent hashing with virtual nodes
type ConsistentHashRing struct {
	mu           sync.RWMutex
	ring         map[uint32]CacheNode // Hash -> Node mapping
	sortedHashes []uint32             // Sorted hash values for binary search
	nodes        map[string]CacheNode // Node ID -> Node mapping
	virtualNodes int                  // Number of virtual nodes per physical node
	hashFunc     HashFunc             // Hash function to use
}

// ConsistentHashRingConfig configures the hash ring
type ConsistentHashRingConfig struct {
	VirtualNodes int      // Number of virtual nodes per physical node (default: 150)
	HashFunc     HashFunc // Hash function to use (default: FNV-1a)
}

// DefaultHashRingConfig returns the default configuration
func DefaultHashRingConfig() *ConsistentHashRingConfig {
	return &ConsistentHashRingConfig{
		VirtualNodes: 150,
		HashFunc:     fnv1aHash,
	}
}

// NewConsistentHashRing creates a new consistent hash ring
func NewConsistentHashRing(config *ConsistentHashRingConfig) *ConsistentHashRing {
	if config == nil {
		config = DefaultHashRingConfig()
	}

	if config.VirtualNodes <= 0 {
		config.VirtualNodes = 150
	}

	if config.HashFunc == nil {
		config.HashFunc = fnv1aHash
	}

	return &ConsistentHashRing{
		ring:         make(map[uint32]CacheNode),
		sortedHashes: make([]uint32, 0),
		nodes:        make(map[string]CacheNode),
		virtualNodes: config.VirtualNodes,
		hashFunc:     config.HashFunc,
	}
}

// Add adds a node to the hash ring
func (r *ConsistentHashRing) Add(node CacheNode) error {
	if node == nil {
		return errors.New("distributed: node cannot be nil")
	}

	nodeID := node.ID()
	if nodeID == "" {
		return errors.New("distributed: node ID cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if node already exists
	if _, exists := r.nodes[nodeID]; exists {
		return ErrNodeAlreadyExists
	}

	// Add virtual nodes
	for i := 0; i < r.virtualNodes; i++ {
		hash := r.hashVirtualNode(nodeID, i)
		r.ring[hash] = node
		r.sortedHashes = append(r.sortedHashes, hash)
	}

	// Sort hashes for binary search
	sort.Slice(r.sortedHashes, func(i, j int) bool {
		return r.sortedHashes[i] < r.sortedHashes[j]
	})

	// Store node
	r.nodes[nodeID] = node

	return nil
}

// Remove removes a node from the hash ring
func (r *ConsistentHashRing) Remove(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if node exists
	if _, exists := r.nodes[nodeID]; !exists {
		return ErrNodeNotFound
	}

	// Remove virtual nodes
	for i := 0; i < r.virtualNodes; i++ {
		hash := r.hashVirtualNode(nodeID, i)
		delete(r.ring, hash)
	}

	// Rebuild sorted hashes
	r.sortedHashes = make([]uint32, 0, len(r.ring))
	for hash := range r.ring {
		r.sortedHashes = append(r.sortedHashes, hash)
	}

	sort.Slice(r.sortedHashes, func(i, j int) bool {
		return r.sortedHashes[i] < r.sortedHashes[j]
	})

	// Remove node
	delete(r.nodes, nodeID)

	return nil
}

// Get returns the node responsible for the given key
func (r *ConsistentHashRing) Get(key string) (CacheNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.sortedHashes) == 0 {
		return nil, ErrNoNodesAvailable
	}

	// Hash the key
	hash := r.hashFunc([]byte(key))

	// Binary search for the node
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})

	// Wrap around if needed
	if idx >= len(r.sortedHashes) {
		idx = 0
	}

	return r.ring[r.sortedHashes[idx]], nil
}

// GetN returns the N nodes responsible for the given key (for replication)
func (r *ConsistentHashRing) GetN(key string, n int) ([]CacheNode, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.sortedHashes) == 0 {
		return nil, ErrNoNodesAvailable
	}

	if n <= 0 {
		return nil, errors.New("distributed: n must be greater than 0")
	}

	// Hash the key
	hash := r.hashFunc([]byte(key))

	// Binary search for the starting position
	idx := sort.Search(len(r.sortedHashes), func(i int) bool {
		return r.sortedHashes[i] >= hash
	})

	// Wrap around if needed
	if idx >= len(r.sortedHashes) {
		idx = 0
	}

	// Collect unique nodes
	seen := make(map[string]bool)
	nodes := make([]CacheNode, 0, n)

	for len(nodes) < n && len(seen) < len(r.nodes) {
		node := r.ring[r.sortedHashes[idx]]
		nodeID := node.ID()

		if !seen[nodeID] {
			nodes = append(nodes, node)
			seen[nodeID] = true
		}

		// Move to next position in ring
		idx = (idx + 1) % len(r.sortedHashes)
	}

	return nodes, nil
}

// Nodes returns all nodes in the hash ring
func (r *ConsistentHashRing) Nodes() []CacheNode {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]CacheNode, 0, len(r.nodes))
	for _, node := range r.nodes {
		nodes = append(nodes, node)
	}

	return nodes
}

// Size returns the number of physical nodes
func (r *ConsistentHashRing) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.nodes)
}

// hashVirtualNode hashes a virtual node identifier
func (r *ConsistentHashRing) hashVirtualNode(nodeID string, vnode int) uint32 {
	key := fmt.Sprintf("%s#%d", nodeID, vnode)
	return r.hashFunc([]byte(key))
}

// fnv1aHash is the default hash function using FNV-1a
func fnv1aHash(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}
