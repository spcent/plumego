// Package dbmanager manages live database connections.
package dbmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	// Register drivers via blank imports so they're available to database/sql.
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	"dbadmin/internal/domain/connection"
)

var ErrUnknownDriver = errors.New("dbmanager: unknown driver")

// Manager holds open *sql.DB pools keyed by connection ID.
type Manager struct {
	mu    sync.RWMutex
	pools map[string]*sql.DB
}

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{pools: make(map[string]*sql.DB)}
}

// Open opens (or reuses) a connection pool for the given config.
// The returned *sql.DB is owned by the Manager; callers must not close it.
func (m *Manager) Open(c *connection.Connection) (*sql.DB, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if db, ok := m.pools[c.ID]; ok {
		return db, nil
	}

	dsn, err := buildDSN(c)
	if err != nil {
		return nil, err
	}

	driverName := string(c.Driver)
	if c.Driver == connection.DriverSQLite {
		driverName = "sqlite"
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s connection: %w", c.Driver, err)
	}
	if err := db.PingContext(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping %s: %w", c.Driver, err)
	}
	m.pools[c.ID] = db
	return db, nil
}

// Get returns the live pool for connID, or nil if not open.
func (m *Manager) Get(connID string) *sql.DB {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pools[connID]
}

// Close closes and removes the pool for connID.
func (m *Manager) Close(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if db, ok := m.pools[connID]; ok {
		db.Close()
		delete(m.pools, connID)
	}
}

// CloseAll closes all open pools.
func (m *Manager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, db := range m.pools {
		db.Close()
		delete(m.pools, id)
	}
}

// Test opens a temporary connection and pings it without caching the pool.
func (m *Manager) Test(c *connection.Connection) error {
	dsn, err := buildDSN(c)
	if err != nil {
		return err
	}
	driverName := string(c.Driver)
	if c.Driver == connection.DriverSQLite {
		driverName = "sqlite"
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(context.Background()); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}


func buildDSN(c *connection.Connection) (string, error) {
	switch c.Driver {
	case connection.DriverMySQL:
		// user:pass@tcp(host:port)/database?charset=utf8mb4&parseTime=true&loc=UTC
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=UTC",
			c.Username, c.Password, c.Host, c.Port, c.Database)
		if c.Options != "" {
			dsn += "&" + c.Options
		}
		return dsn, nil
	case connection.DriverSQLite:
		if c.FilePath == "" {
			return "", fmt.Errorf("sqlite requires file_path")
		}
		return c.FilePath, nil
	default:
		return "", ErrUnknownDriver
	}
}
