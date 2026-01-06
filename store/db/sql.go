package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

var ErrInvalidConfig = errors.New("db: invalid config")

// Config defines the parameters for opening and tuning a sql.DB.
type Config struct {
	Driver          string
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	PingTimeout     time.Duration
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
	if c.MaxOpenConns < 0 || c.MaxIdleConns < 0 {
		return fmt.Errorf("%w: connection limits must be non-negative", ErrInvalidConfig)
	}
	if c.ConnMaxLifetime < 0 || c.ConnMaxIdleTime < 0 {
		return fmt.Errorf("%w: timeouts must be non-negative", ErrInvalidConfig)
	}
	if c.PingTimeout < 0 {
		return fmt.Errorf("%w: ping timeout must be non-negative", ErrInvalidConfig)
	}
	return nil
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
		return nil, err
	}

	ApplyConfig(db, config)

	if config.PingTimeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), config.PingTimeout)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return db, nil
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
