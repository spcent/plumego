package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// ErrInvalidConfig is returned when configuration is invalid
	ErrInvalidConfig = errors.New("db: invalid config")

	// ErrConnectionFailed is returned when connection fails
	ErrConnectionFailed = errors.New("db: connection failed")

	// ErrPingFailed is returned when ping fails
	ErrPingFailed = errors.New("db: ping failed")

	// ErrTransactionFailed is returned when transaction operations fail
	ErrTransactionFailed = errors.New("db: transaction failed")

	// ErrQueryFailed is returned when query execution fails
	ErrQueryFailed = errors.New("db: query failed")

	// ErrNoRows is returned when query returns no rows
	ErrNoRows = errors.New("db: no rows returned")

	// ErrMultipleRows is returned when query returns multiple rows
	ErrMultipleRows = errors.New("db: multiple rows returned")
)

// Config defines the parameters for opening and tuning a sql.DB.
type Config struct {
	// Driver is the database driver name (e.g., "mysql", "postgres", "sqlite3")
	Driver string

	// DSN is the data source name
	DSN string

	// MaxOpenConns is the maximum number of open connections to the database
	MaxOpenConns int

	// MaxIdleConns is the maximum number of idle connections in the connection pool
	MaxIdleConns int

	// ConnMaxLifetime is the maximum amount of time a connection may be reused
	ConnMaxLifetime time.Duration

	// ConnMaxIdleTime is the maximum amount of time a connection may be idle
	ConnMaxIdleTime time.Duration

	// PingTimeout is the timeout for the initial ping test
	PingTimeout time.Duration

	// QueryTimeout is the default timeout for queries (0 = no timeout)
	QueryTimeout time.Duration

	// TransactionTimeout is the default timeout for transactions (0 = no timeout)
	TransactionTimeout time.Duration
}

// DB defines the minimal behavior required by SQL consumers.
type DB interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	PingContext(ctx context.Context) error
	Close() error
}

// OpenFunc allows swapping sql.Open for tests.
type OpenFunc func(driver, dsn string) (*sql.DB, error)

// Validate ensures the configuration is usable.
func (c Config) Validate() error {
	if strings.TrimSpace(c.Driver) == "" {
		return fmt.Errorf("%w: driver is required", ErrInvalidConfig)
	}
	if strings.TrimSpace(c.DSN) == "" {
		return fmt.Errorf("%w: dsn is required", ErrInvalidConfig)
	}
	if c.MaxOpenConns < 0 {
		return fmt.Errorf("%w: MaxOpenConns must be non-negative", ErrInvalidConfig)
	}
	if c.MaxIdleConns < 0 {
		return fmt.Errorf("%w: MaxIdleConns must be non-negative", ErrInvalidConfig)
	}
	if c.ConnMaxLifetime < 0 {
		return fmt.Errorf("%w: ConnMaxLifetime must be non-negative", ErrInvalidConfig)
	}
	if c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("%w: ConnMaxIdleTime must be non-negative", ErrInvalidConfig)
	}
	if c.PingTimeout < 0 {
		return fmt.Errorf("%w: PingTimeout must be non-negative", ErrInvalidConfig)
	}
	if c.QueryTimeout < 0 {
		return fmt.Errorf("%w: QueryTimeout must be non-negative", ErrInvalidConfig)
	}
	if c.TransactionTimeout < 0 {
		return fmt.Errorf("%w: TransactionTimeout must be non-negative", ErrInvalidConfig)
	}
	return nil
}

// DefaultConfig returns a default configuration for common use cases.
func DefaultConfig(driver, dsn string) Config {
	return Config{
		Driver:             driver,
		DSN:                dsn,
		MaxOpenConns:       10,
		MaxIdleConns:       5,
		ConnMaxLifetime:    30 * time.Minute,
		ConnMaxIdleTime:    5 * time.Minute,
		PingTimeout:        5 * time.Second,
		QueryTimeout:       30 * time.Second,
		TransactionTimeout: 60 * time.Second,
	}
}

// Open opens a database connection using the standard sql.Open.
func Open(config Config) (*sql.DB, error) {
	return OpenWith(config, sql.Open)
}

// OpenWith opens a database connection with an injected opener.
func OpenWith(config Config, open OpenFunc) (*sql.DB, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	if open == nil {
		return nil, fmt.Errorf("%w: open function is required", ErrInvalidConfig)
	}

	db, err := open(config.Driver, config.DSN)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	ApplyConfig(db, config)

	if config.PingTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), config.PingTimeout)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("%w: %v", ErrPingFailed, err)
		}
	}

	return db, nil
}

// OpenWithRetry opens a database connection with retry logic.
func OpenWithRetry(config Config, open OpenFunc, maxRetries int, retryDelay time.Duration) (*sql.DB, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		db, err := OpenWith(config, open)
		if err == nil {
			return db, nil
		}
		if errors.Is(err, ErrInvalidConfig) {
			return nil, err
		}

		lastErr = err
	}

	return nil, fmt.Errorf("%w: failed after %d attempts: %v", ErrConnectionFailed, maxRetries+1, lastErr)
}

// ApplyConfig updates pooling and lifetime settings on an existing sql.DB.
func ApplyConfig(db *sql.DB, config Config) {
	if db == nil {
		return
	}

	maxOpen := config.MaxOpenConns
	maxIdle := config.MaxIdleConns
	if maxOpen > 0 && maxIdle > maxOpen {
		maxIdle = maxOpen
	}

	if maxOpen > 0 {
		db.SetMaxOpenConns(maxOpen)
	}
	if maxIdle > 0 {
		db.SetMaxIdleConns(maxIdle)
	}
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	}
	if config.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.ConnMaxIdleTime)
	}
}

// ExecContext executes a query with timeout support.
func ExecContext(ctx context.Context, db DB, query string, args ...any) (sql.Result, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrQueryFailed)
	}

	ctx, cancel := withQueryTimeout(ctx, db)
	defer cancel()

	return db.ExecContext(ctx, query, args...)
}

// QueryContext executes a query with timeout support.
func QueryContext(ctx context.Context, db DB, query string, args ...any) (*sql.Rows, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrQueryFailed)
	}

	ctx, cancel := withQueryTimeout(ctx, db)
	defer cancel()

	return db.QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query with timeout support.
func QueryRowContext(ctx context.Context, db DB, query string, args ...any) *sql.Row {
	if db == nil {
		return nil
	}

	ctx, cancel := withQueryTimeout(ctx, db)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...)
}

// WithTimeout creates a context with timeout for database operations.
func WithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout > 0 {
		return context.WithTimeout(ctx, timeout)
	}
	return ctx, func() {}
}

// WithTransaction executes a function within a transaction.
// If the function returns an error, the transaction is rolled back.
// Otherwise, the transaction is committed.
func WithTransaction(ctx context.Context, db DB, txOpts *sql.TxOptions, fn func(*sql.Tx) error) error {
	if db == nil {
		return fmt.Errorf("%w: database is nil", ErrTransactionFailed)
	}

	ctx, cancel := withTransactionTimeout(ctx, db)
	defer cancel()

	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("%w: begin transaction failed: %v", ErrTransactionFailed, err)
	}

	// Defer rollback in case of panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	err = fn(tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("%w: %v", ErrTransactionFailed, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: commit failed: %v", ErrTransactionFailed, err)
	}

	return nil
}

// QueryRow executes a query and returns the first row.
// Use QueryRowStrict when you need single-row enforcement.
func QueryRow(ctx context.Context, db DB, query string, args ...any) (*sql.Row, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrQueryFailed)
	}

	ctx, cancel := withQueryTimeout(ctx, db)
	defer cancel()

	return db.QueryRowContext(ctx, query, args...), nil
}

// QueryRowStrict executes a query and enforces single-row semantics.
// Returns ErrNoRows if no rows are returned.
// Returns ErrMultipleRows if multiple rows are returned.
func QueryRowStrict(ctx context.Context, db DB, query string, scan func(*sql.Rows) error, args ...any) error {
	if db == nil {
		return fmt.Errorf("%w: database is nil", ErrQueryFailed)
	}
	if scan == nil {
		return fmt.Errorf("%w: scan function is nil", ErrQueryFailed)
	}

	rows, err := QueryContext(ctx, db, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return fmt.Errorf("%w: %v", ErrQueryFailed, err)
		}
		return fmt.Errorf("%w: %v", ErrNoRows, sql.ErrNoRows)
	}

	if err := scan(rows); err != nil {
		return err
	}

	if rows.Next() {
		return ErrMultipleRows
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return nil
}

// ScanRow is a helper function to scan a single row into a destination.
func ScanRow(row *sql.Row, dest ...any) error {
	if row == nil {
		return fmt.Errorf("%w: row is nil", ErrQueryFailed)
	}

	err := row.Scan(dest...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: %v", ErrNoRows, err)
		}
		return fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return nil
}

// ScanRows is a helper function to scan multiple rows into a slice.
func ScanRows[T any](rows *sql.Rows, scanFunc func(*sql.Rows) (T, error)) ([]T, error) {
	if rows == nil {
		return nil, fmt.Errorf("%w: rows is nil", ErrQueryFailed)
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		result, err := scanFunc(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrQueryFailed, err)
	}

	return results, nil
}

// Ping tests the database connection with timeout.
func Ping(ctx context.Context, db DB, timeout time.Duration) error {
	if db == nil {
		return fmt.Errorf("%w: database is nil", ErrPingFailed)
	}

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrPingFailed, err)
	}

	return nil
}

// HealthCheck performs a comprehensive health check of the database.
func HealthCheck(ctx context.Context, db DB, timeout time.Duration) (HealthStatus, error) {
	if db == nil {
		return HealthStatus{Status: "unhealthy"}, fmt.Errorf("%w: database is nil", ErrPingFailed)
	}

	start := time.Now()

	// Ping the database
	pingErr := Ping(ctx, db, timeout)

	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now(),
		Latency:   time.Since(start),
	}

	if pingErr != nil {
		health.Status = "unhealthy"
		health.Error = pingErr.Error()
		return health, pingErr
	}

	// Get connection stats if available
	if sqlDB, ok := db.(*sql.DB); ok {
		stats := sqlDB.Stats()
		health.OpenConnections = stats.OpenConnections
		health.InUse = stats.InUse
		health.Idle = stats.Idle
		health.WaitCount = stats.WaitCount
		health.WaitDuration = stats.WaitDuration
	}

	return health, nil
}

// HealthStatus represents the health status of the database.
type HealthStatus struct {
	Status          string        `json:"status"`
	Timestamp       time.Time     `json:"timestamp"`
	Latency         time.Duration `json:"latency"`
	Error           string        `json:"error,omitempty"`
	OpenConnections int           `json:"open_connections,omitempty"`
	InUse           int           `json:"in_use,omitempty"`
	Idle            int           `json:"idle,omitempty"`
	WaitCount       int64         `json:"wait_count,omitempty"`
	WaitDuration    time.Duration `json:"wait_duration,omitempty"`
}

// GetConfig is a helper interface for databases that can expose their configuration.
type GetConfig interface {
	GetConfig() Config
}

func withQueryTimeout(ctx context.Context, db DB) (context.Context, context.CancelFunc) {
	config, ok := getConfig(db)
	if !ok || config.QueryTimeout <= 0 {
		return ctx, func() {}
	}
	return WithTimeout(ctx, config.QueryTimeout)
}

func withTransactionTimeout(ctx context.Context, db DB) (context.Context, context.CancelFunc) {
	config, ok := getConfig(db)
	if !ok || config.TransactionTimeout <= 0 {
		return ctx, func() {}
	}
	return WithTimeout(ctx, config.TransactionTimeout)
}

func getConfig(db DB) (Config, bool) {
	if db == nil {
		return Config{}, false
	}
	cfg, ok := db.(GetConfig)
	if !ok {
		return Config{}, false
	}
	return cfg.GetConfig(), true
}
