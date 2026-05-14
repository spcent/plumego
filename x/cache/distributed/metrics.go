package distributed

import (
	"sync"
	"sync/atomic"
	"time"
)

// DistributedMetrics tracks metrics for the distributed cache
type DistributedMetrics struct {
	mu sync.RWMutex

	TotalRequests       uint64
	FailoverCount       uint64
	ReplicationLag      time.Duration
	ReplicationFailures uint64
	ReplicationFails    uint64
	HashCollisions      uint64
	HealthyNodes        int
	UnhealthyNodes      int
	RebalanceEvents     uint64
}

type collisionCounter interface {
	CollisionCount() uint64
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
		ReplicationFails:    atomic.LoadUint64(&dc.metrics.ReplicationFailures),
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
