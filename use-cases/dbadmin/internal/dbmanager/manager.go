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
	_ "github.com/lib/pq"
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
// ctx is forwarded to the Ping and retry logic so that a cancelled or
// deadline-exceeded request aborts the call immediately.
// The returned *sql.DB is owned by the Manager; callers must not close it.
func (m *Manager) Open(ctx context.Context, c *connection.Connection) (*sql.DB, error) {
	// Fast path: read-lock, check pool, release lock before any I/O.
	m.mu.RLock()
	existing, ok := m.pools[c.ID]
	m.mu.RUnlock()

	if ok {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		if err := existing.PingContext(pingCtx); err == nil {
			return existing, nil
		}
		// Stale — remove under write lock, but only if it hasn't already been
		// replaced by another goroutine.
		m.mu.Lock()
		if cur, still := m.pools[c.ID]; still && cur == existing {
			cur.Close()
			delete(m.pools, c.ID)
		}
		m.mu.Unlock()
	}

	// Slow path: create a new pool outside any lock so other goroutines are
	// not blocked during the network handshake / retry loop.
	dsn, err := buildDSN(c)
	if err != nil {
		return nil, err
	}
	driverName := string(c.Driver)
	switch c.Driver {
	case connection.DriverSQLite:
		driverName = "sqlite"
	case connection.DriverPostgres:
		driverName = "postgres"
	}

	cfg := retry.DefaultConfig()
	newDB, err := retry.WithResult[*sql.DB](ctx, cfg, func() (*sql.DB, error) {
		db, err := sql.Open(driverName, dsn)
		if err != nil {
			return nil, fmt.Errorf("open %s connection (%s): %w", c.Driver, redactDSN(dsn), err)
		}
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(5 * time.Minute)
		db.SetConnMaxIdleTime(10 * time.Minute)
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("ping %s: %w", c.Driver, err)
		}
		return db, nil
	})
	if err != nil {
		return nil, err
	}

	// Write lock: insert only if no other goroutine beat us to it.
	m.mu.Lock()
	defer m.mu.Unlock()
	if winner, ok := m.pools[c.ID]; ok {
		newDB.Close()
		return winner, nil
	}
	m.pools[c.ID] = newDB
	return newDB, nil
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
// ctx is forwarded to PingContext so that a cancelled or deadline-exceeded
// request aborts the call immediately.
func (m *Manager) Test(ctx context.Context, c *connection.Connection) error {
	dsn, err := buildDSN(c)
	if err != nil {
		return err
	}
	driverName := string(c.Driver)
	switch c.Driver {
	case connection.DriverSQLite:
		driverName = "sqlite"
	case connection.DriverPostgres:
		driverName = "postgres"
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("open: %w", err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}
	return nil
}

// redactDSN masks the password in a MySQL or PostgreSQL DSN for safe use in
// log/error messages.
// MySQL input:    "user:secret@tcp(host:3306)/db?..."
// MySQL output:   "user:***@tcp(host:3306)/db?..."
// Postgres input:  "postgres://user:secret@host:5432/db?sslmode=disable"
// Postgres output: "postgres://user:***@host:5432/db?sslmode=disable"
// Other DSNs (e.g. file paths) are returned unchanged.
func redactDSN(dsn string) string {
	if strings.HasPrefix(dsn, "postgres://") {
		rest := strings.TrimPrefix(dsn, "postgres://")
		at := strings.Index(rest, "@")
		if at <= 0 {
			return dsn
		}
		colon := strings.Index(rest[:at], ":")
		if colon <= 0 {
			return dsn
		}
		return "postgres://" + rest[:colon+1] + "***" + rest[at:]
	}
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
	case connection.DriverPostgres:
		sslmode := c.SQLTLSMode
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			c.Username, c.Password, c.Host, c.Port, c.Database, sslmode)
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
		return "", fmt.Errorf("%w: supported drivers are mysql, postgres, sqlite", ErrUnknownDriver)
	}
}
