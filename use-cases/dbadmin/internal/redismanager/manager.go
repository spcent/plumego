// Package redismanager manages pooled Redis client connections.
package redismanager

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"sync"

	"dbadmin/internal/domain/connection"

	"github.com/redis/go-redis/v9"
)

var ErrNoClient = errors.New("redismanager: no client for connection")

// poolKey identifies a cached client: one pool per (connID, dbIndex).
type poolKey struct {
	connID  string
	dbIndex int
}

// Manager holds open *redis.Client instances keyed by (connID, dbIndex).
type Manager struct {
	mu      sync.RWMutex
	clients map[poolKey]*redis.Client
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{clients: make(map[poolKey]*redis.Client)}
}

// Open returns a cached client for the given connection and DB index,
// creating one on first call. The client is owned by the Manager; callers
// must not close it.
func (m *Manager) Open(c *connection.Connection, dbIndex int) (*redis.Client, error) {
	key := poolKey{connID: c.ID, dbIndex: dbIndex}

	m.mu.RLock()
	if cl, ok := m.clients[key]; ok {
		m.mu.RUnlock()
		return cl, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()
	// Double-check after acquiring write lock.
	if cl, ok := m.clients[key]; ok {
		return cl, nil
	}

	cl, err := m.build(c, dbIndex)
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
func (m *Manager) Get(connID string, dbIndex int) (*redis.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if cl, ok := m.clients[poolKey{connID: connID, dbIndex: dbIndex}]; ok {
		return cl, nil
	}
	return nil, ErrNoClient
}

// build constructs a new *redis.Client for the given connection config.
// The password is never written to logs or error strings.
func (m *Manager) build(c *connection.Connection, dbIndex int) (*redis.Client, error) {
	if c.Host == "" {
		return nil, fmt.Errorf("redis: host is required")
	}
	port := c.Port
	if port == 0 {
		port = 6379
	}
	addr := fmt.Sprintf("%s:%d", c.Host, port)

	opts := &redis.Options{
		Addr:     addr,
		Password: c.Password,
		DB:       dbIndex,
	}
	if c.TLSEnabled {
		opts.TLSConfig = &tls.Config{ServerName: c.Host}
	}
	cl := redis.NewClient(opts)
	return cl, nil
}
