// Package esmanager manages Elasticsearch client connections.
package esmanager

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/retry"

	"github.com/elastic/go-elasticsearch/v8"
)

// Manager manages Elasticsearch client connections.
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*elasticsearch.Client
}

// NewManager creates a new Elasticsearch connection manager.
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*elasticsearch.Client),
	}
}

// Open returns a cached client for the connection, creating one if necessary.
func (m *Manager) Open(conn *connection.Connection) (*elasticsearch.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cl, ok := m.clients[conn.ID]; ok {
		// Validate connection is still alive with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		res, err := cl.Ping(cl.Ping.WithContext(ctx))
		if err == nil && !res.IsError() {
			res.Body.Close()
			return cl, nil
		}
		if res != nil && res.Body != nil {
			res.Body.Close()
		}
		// Connection is stale, remove it
		delete(m.clients, conn.ID)
	}

	// Use retry logic for connection creation to handle transient failures
	cfg := retry.DefaultConfig()
	ctx := context.Background()

	cl, err := retry.WithResult[*elasticsearch.Client](ctx, cfg, func() (*elasticsearch.Client, error) {
		return m.createClient(conn)
	})
	if err != nil {
		return nil, err
	}

	m.clients[conn.ID] = cl
	return cl, nil
}

// Test verifies connectivity to the Elasticsearch cluster.
func (m *Manager) Test(ctx context.Context, conn *connection.Connection) error {
	cl, err := m.createClient(conn)
	if err != nil {
		return err
	}

	res, err := cl.Ping(cl.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping returned %s", res.Status())
	}

	return nil
}

// Get returns the cached client for the given connID, or nil if not open.
func (m *Manager) Get(connID string) *elasticsearch.Client {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.clients[connID]
}

// Count returns the number of active Elasticsearch connections.
func (m *Manager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.clients)
}

// Close closes and removes the client for the given connection ID.
func (m *Manager) Close(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// The elasticsearch.Client doesn't have an explicit Close method,
	// but removing it from the cache allows GC to clean up.
	delete(m.clients, connID)
}

// CloseAll closes all cached clients.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id := range m.clients {
		delete(m.clients, id)
	}
}

func (m *Manager) createClient(conn *connection.Connection) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: m.buildAddresses(conn),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: conn.ESInsecureSkipTLS,
			},
		},
	}

	// Configure authentication
	if conn.ESAPIKey != "" {
		cfg.APIKey = conn.ESAPIKey
	} else if conn.ESUsername != "" {
		cfg.Username = conn.ESUsername
		cfg.Password = conn.ESPassword
	} else if conn.Username != "" {
		// Fallback to generic username/password
		cfg.Username = conn.Username
		cfg.Password = conn.Password
	}

	cl, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create ES client: %w", err)
	}

	// Verify connection with ping
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := cl.Ping(cl.Ping.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to ping ES cluster: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("ping returned %s", res.Status())
	}

	return cl, nil
}

func (m *Manager) buildAddresses(conn *connection.Connection) []string {
	// Use explicit node list if provided
	if len(conn.ESNodes) > 0 {
		return conn.ESNodes
	}

	// Build single-node address from host:port
	host := conn.Host
	if host == "" {
		host = "localhost"
	}

	port := conn.Port
	if port == 0 {
		port = 9200
	}

	return []string{fmt.Sprintf("http://%s:%d", host, port)}
}
