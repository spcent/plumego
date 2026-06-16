// Package redismanager manages pooled Redis client connections.
package redismanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"
	"time"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/retry"

	"github.com/redis/go-redis/v9"
)

var ErrNoClient = errors.New("redismanager: no client for connection")

// poolKey identifies a cached client: one pool per (connID, dbIndex).
type poolKey struct {
	connID  string
	dbIndex int
}

// Manager holds open redis.UniversalClient instances keyed by (connID, dbIndex).
// UniversalClient is satisfied by *redis.Client (standalone and sentinel
// failover clients) and *redis.ClusterClient (cluster mode).
type Manager struct {
	mu      sync.RWMutex
	clients map[poolKey]redis.UniversalClient
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{clients: make(map[poolKey]redis.UniversalClient)}
}

// Open returns a cached client for the given connection and DB index,
// creating one on first call. The client is owned by the Manager; callers
// must not close it.
func (m *Manager) Open(c *connection.Connection, dbIndex int) (redis.UniversalClient, error) {
	key := poolKey{connID: c.ID, dbIndex: dbIndex}

	m.mu.Lock()
	defer m.mu.Unlock()

	if cl, ok := m.clients[key]; ok {
		// Validate connection is still alive with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := cl.Ping(ctx).Err(); err == nil {
			return cl, nil
		}
		// Connection is stale, close and remove it
		cl.Close()
		delete(m.clients, key)
	}

	// Use retry logic for connection creation to handle transient failures
	ctx := context.Background()
	cfg := retry.DefaultConfig()

	cl, err := retry.WithResult[redis.UniversalClient](ctx, cfg, func() (redis.UniversalClient, error) {
		client, err := m.build(c, dbIndex)
		if err != nil {
			return nil, err
		}

		// Validate connection with a ping
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := client.Ping(pingCtx).Err(); err != nil {
			client.Close()
			return nil, fmt.Errorf("ping: %w", err)
		}

		return client, nil
	})
	if err != nil {
		return nil, err
	}

	m.clients[key] = cl
	return cl, nil
}

// Test opens a temporary client, pings it, and closes it without caching.
func (m *Manager) Test(ctx context.Context, c *connection.Connection) error {
	cl, err := m.build(c, c.RedisDBIndex)
	if err != nil {
		return err
	}
	defer cl.Close()
	if err := cl.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

// Close closes and removes all clients for connID.
func (m *Manager) Close(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, cl := range m.clients {
		if k.connID == connID {
			cl.Close()
			delete(m.clients, k)
		}
	}
}

// CloseAll closes every open client.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, cl := range m.clients {
		cl.Close()
		delete(m.clients, k)
	}
}

// Get returns the cached client for the given connID and dbIndex, or
// ErrNoClient if none exists.
func (m *Manager) Get(connID string, dbIndex int) (redis.UniversalClient, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if cl, ok := m.clients[poolKey{connID: connID, dbIndex: dbIndex}]; ok {
		return cl, nil
	}
	return nil, ErrNoClient
}

// Count returns the number of active Redis connections.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// buildTLSConfig returns a *tls.Config for the connection's target host when
// TLS is enabled, or nil otherwise. Shared by every mode's client builder so
// behavior stays identical across standalone, cluster, and sentinel modes.
func buildTLSConfig(c *connection.Connection) *tls.Config {
	if !c.TLSEnabled {
		return nil
	}
	return &tls.Config{ServerName: c.Host}
}

// build constructs a new redis.UniversalClient for the given connection
// config, dispatching on c.RedisMode. The password is never written to logs
// or error strings.
func (m *Manager) build(c *connection.Connection, dbIndex int) (redis.UniversalClient, error) {
	switch c.RedisMode {
	case "cluster":
		return buildClusterClient(c)
	case "sentinel":
		return buildSentinelClient(c, dbIndex)
	default:
		return buildStandaloneClient(c, dbIndex)
	}
}

// buildStandaloneClient builds a *redis.Client for standalone mode
// ("" or "standalone"). Unchanged from prior single-mode behavior.
func buildStandaloneClient(c *connection.Connection, dbIndex int) (redis.UniversalClient, error) {
	if c.Host == "" {
		return nil, fmt.Errorf("redis: host is required")
	}
	port := c.Port
	if port == 0 {
		port = 6379
	}
	addr := fmt.Sprintf("%s:%d", c.Host, port)

	opts := &redis.Options{
		Addr:      addr,
		Password:  c.Password,
		DB:        dbIndex,
		TLSConfig: buildTLSConfig(c),
	}
	return redis.NewClient(opts), nil
}

// buildClusterClient builds a *redis.ClusterClient for cluster mode.
// Redis Cluster does not support SELECT / the standalone DB-index concept,
// so dbIndex is intentionally ignored here — every key lives in the single
// cluster-wide keyspace (DB 0).
func buildClusterClient(c *connection.Connection) (redis.UniversalClient, error) {
	if len(c.RedisClusterAddrs) == 0 {
		return nil, fmt.Errorf("redis: cluster addrs are required")
	}
	opts := &redis.ClusterOptions{
		Addrs:     c.RedisClusterAddrs,
		Password:  c.Password,
		TLSConfig: buildTLSConfig(c),
	}
	return redis.NewClusterClient(opts), nil
}

// buildSentinelClient builds a failover *redis.Client for sentinel mode,
// connecting via the Sentinel-monitored master group identified by
// RedisSentinelMasterName.
func buildSentinelClient(c *connection.Connection, dbIndex int) (redis.UniversalClient, error) {
	if len(c.RedisSentinelAddrs) == 0 {
		return nil, fmt.Errorf("redis: sentinel addrs are required")
	}
	if c.RedisSentinelMasterName == "" {
		return nil, fmt.Errorf("redis: sentinel master name is required")
	}
	opts := &redis.FailoverOptions{
		MasterName:    c.RedisSentinelMasterName,
		SentinelAddrs: c.RedisSentinelAddrs,
		Password:      c.Password,
		DB:            dbIndex,
		TLSConfig:     buildTLSConfig(c),
	}
	return redis.NewFailoverClient(opts), nil
}
