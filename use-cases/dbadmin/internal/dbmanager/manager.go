// Package dbmanager manages live database connections.
package dbmanager

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	// Register drivers via blank imports so they're available to database/sql.
	_ "github.com/go-sql-driver/mysql"
	_ "modernc.org/sqlite"

	"dbadmin/internal/domain/connection"
	"dbadmin/internal/retry"
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
		// Validate connection is still alive with a short timeout
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err == nil {
			return db, nil
		}
		// Connection is stale, close and remove it
		db.Close()
		delete(m.pools, c.ID)
	}

	dsn, err := buildDSN(c)
	if err != nil {
		return nil, err
	}

	driverName := string(c.Driver)
	if c.Driver == connection.DriverSQLite {
		driverName = "sqlite"
	}

	// Use retry logic for connection creation to handle transient failures
	ctx := context.Background()
	cfg := retry.DefaultConfig()

	db, err := retry.WithResult[*sql.DB](ctx, cfg, func() (*sql.DB, error) {
		db, err := sql.Open(driverName, dsn)
		if err != nil {
			// Use redactDSN so the password is never included in error messages.
			return nil, fmt.Errorf("open %s connection (%s): %w", c.Driver, redactDSN(dsn), err)
		}

		// Configure connection pool for production stability
		db.SetMaxOpenConns(25)                  // Max 25 concurrent connections
		db.SetMaxIdleConns(5)                   // Keep 5 idle connections ready
		db.SetConnMaxLifetime(5 * time.Minute)  // Recycle connections after 5 min
		db.SetConnMaxIdleTime(10 * time.Minute) // Close idle connections after 10 min

		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("ping %s: %w", c.Driver, err)
		}

		return db, nil
	})
	if err != nil {
		return nil, err
	}

	m.pools[c.ID] = db
	return db, nil
}

// Stats returns the pool statistics for a connection ID.
// Returns (stats, true) if the connection exists, (empty stats, false) otherwise.
func (m *Manager) Stats(connID string) (sql.DBStats, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if db, ok := m.pools[connID]; ok {
		return db.Stats(), true
	}
	return sql.DBStats{}, false
}

// AllStats returns pool statistics for all active SQL connections.
func (m *Manager) AllStats() map[string]sql.DBStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]sql.DBStats, len(m.pools))
	for connID, db := range m.pools {
		stats[connID] = db.Stats()
	}
	return stats
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

// redactDSN masks the password in a MySQL DSN for safe use in log/error messages.
// Input:  "user:secret@tcp(host:3306)/db?..."
// Output: "user:***@tcp(host:3306)/db?..."
// Non-MySQL DSNs (e.g. file paths) are returned unchanged.
func redactDSN(dsn string) string {
	at := strings.Index(dsn, "@")
	if at <= 0 {
		return dsn
	}
	colon := strings.Index(dsn[:at], ":")
	if colon <= 0 {
		return dsn
	}
	return dsn[:colon+1] + "***" + dsn[at:]
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
