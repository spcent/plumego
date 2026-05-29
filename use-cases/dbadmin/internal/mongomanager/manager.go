package mongomanager

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"dbadmin/internal/domain/connection"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Manager manages MongoDB client connections.
type Manager struct {
	mu      sync.RWMutex
	clients map[string]*mongo.Client
}

// NewManager creates a new MongoDB connection manager.
func NewManager() *Manager {
	return &Manager{
		clients: make(map[string]*mongo.Client),
	}
}

// Open returns a cached client for the connection, creating one if necessary.
func (m *Manager) Open(conn *connection.Connection) (*mongo.Client, error) {
	m.mu.RLock()
	if client, ok := m.clients[conn.ID]; ok {
		m.mu.RUnlock()
		return client, nil
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if client, ok := m.clients[conn.ID]; ok {
		return client, nil
	}

	client, err := m.createClient(conn)
	if err != nil {
		return nil, err
	}

	m.clients[conn.ID] = client
	return client, nil
}

// Test verifies connectivity to the MongoDB server.
func (m *Manager) Test(ctx context.Context, conn *connection.Connection) error {
	client, err := m.createClient(conn)
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	return client.Ping(ctx, nil)
}

// Close closes and removes the client for the given connection ID.
func (m *Manager) Close(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if client, ok := m.clients[connID]; ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client.Disconnect(ctx)
		delete(m.clients, connID)
	}
}

// CloseAll closes all cached clients.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for id, client := range m.clients {
		client.Disconnect(ctx)
		delete(m.clients, id)
	}
}

func (m *Manager) createClient(conn *connection.Connection) (*mongo.Client, error) {
	uri := m.buildURI(conn)

	opts := options.Client().ApplyURI(uri)

	// Configure authentication
	if conn.Username != "" && conn.Password != "" {
		opts.SetAuth(options.Credential{
			Username: conn.Username,
			Password: conn.Password,
		})
		if conn.MongoAuthDB != "" {
			opts.Auth.AuthSource = conn.MongoAuthDB
		}
	}

	// Configure TLS
	if conn.MongoTLSEnabled {
		opts.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: false,
		})
	}

	// Configure replica set
	if conn.MongoReplicaSet != "" {
		opts.SetReplicaSet(conn.MongoReplicaSet)
	}

	// Set timeouts
	opts.SetConnectTimeout(10 * time.Second)
	opts.SetServerSelectionTimeout(10 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		client.Disconnect(ctx)
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

func (m *Manager) buildURI(conn *connection.Connection) string {
	// Use explicit URI if provided
	if conn.MongoURI != "" {
		return conn.MongoURI
	}

	// Build URI from components
	host := conn.Host
	if host == "" {
		host = "localhost"
	}

	port := conn.Port
	if port == 0 {
		port = 27017
	}

	scheme := "mongodb"
	if conn.MongoTLSEnabled {
		scheme = "mongodb+srv"
	}

	return fmt.Sprintf("%s://%s:%d", scheme, host, port)
}
