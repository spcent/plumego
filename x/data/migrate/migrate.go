// Package migrate wraps goose migration execution for independently versioned
// x/data migration runners.
package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pressly/goose/v3"
)

var (
	// ErrInvalidConfig is returned when migrator construction input is invalid.
	ErrInvalidConfig = errors.New("migrate: invalid config")

	// ErrMigrationFailed is returned when a goose operation fails.
	ErrMigrationFailed = errors.New("migrate: operation failed")
)

// Migrator executes SQL migrations through goose.
type Migrator struct {
	provider *goose.Provider
	empty    bool
}

// MigrationStatus describes one migration and whether it is pending or applied.
type MigrationStatus struct {
	Version   int64      `json:"version"`
	Name      string     `json:"name"`
	State     string     `json:"state"`
	AppliedAt *time.Time `json:"applied_at,omitempty"`
}

// New creates a migrator using SQLite3 dialect. Use NewWithDialect for other databases.
func New(db *sql.DB, migrationsDir string) (*Migrator, error) {
	return NewWithDialect(goose.DialectSQLite3, db, migrationsDir)
}

// NewWithDialect creates a migrator for an explicit goose dialect.
func NewWithDialect(dialect goose.Dialect, db *sql.DB, migrationsDir string) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrInvalidConfig)
	}
	info, err := os.Stat(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("%w: migrations directory: %w", ErrInvalidConfig, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: migrations path is not a directory", ErrInvalidConfig)
	}

	provider, err := goose.NewProvider(dialect, db, os.DirFS(migrationsDir), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		if isNoMigrationsError(err) {
			return &Migrator{empty: true}, nil
		}
		return nil, fmt.Errorf("%w: create goose provider: %w", ErrInvalidConfig, err)
	}
	return &Migrator{provider: provider}, nil
}

// NewWithFS creates a migrator from a caller-provided filesystem.
func NewWithFS(dialect goose.Dialect, db *sql.DB, fsys fs.FS) (*Migrator, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: database is nil", ErrInvalidConfig)
	}
	if fsys == nil {
		return nil, fmt.Errorf("%w: filesystem is nil", ErrInvalidConfig)
	}
	provider, err := goose.NewProvider(dialect, db, fsys, goose.WithDisableGlobalRegistry(true))
	if err != nil {
		if isNoMigrationsError(err) {
			return &Migrator{empty: true}, nil
		}
		return nil, fmt.Errorf("%w: create goose provider: %w", ErrInvalidConfig, err)
	}
	return &Migrator{provider: provider}, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
	if m == nil || m.provider == nil {
		if m != nil && m.empty {
			return nil
		}
		return fmt.Errorf("%w: migrator is nil", ErrMigrationFailed)
	}
	if _, err := m.provider.Up(ctx); err != nil {
		return fmt.Errorf("%w: up: %w", ErrMigrationFailed, err)
	}
	return nil
}

// Down rolls back one applied migration.
func (m *Migrator) Down(ctx context.Context) error {
	if m == nil || m.provider == nil {
		if m != nil && m.empty {
			return nil
		}
		return fmt.Errorf("%w: migrator is nil", ErrMigrationFailed)
	}
	if _, err := m.provider.Down(ctx); err != nil {
		return fmt.Errorf("%w: down: %w", ErrMigrationFailed, err)
	}
	return nil
}

// Status returns pending and applied migration state.
func (m *Migrator) Status(ctx context.Context) ([]MigrationStatus, error) {
	if m == nil || m.provider == nil {
		if m != nil && m.empty {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: migrator is nil", ErrMigrationFailed)
	}
	statuses, err := m.provider.Status(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: status: %w", ErrMigrationFailed, err)
	}
	out := make([]MigrationStatus, 0, len(statuses))
	for _, status := range statuses {
		if status == nil || status.Source == nil {
			continue
		}
		entry := MigrationStatus{
			Version: status.Source.Version,
			Name:    migrationName(status.Source.Path, status.Source.Version),
			State:   string(status.State),
		}
		if !status.AppliedAt.IsZero() {
			appliedAt := status.AppliedAt
			entry.AppliedAt = &appliedAt
		}
		out = append(out, entry)
	}
	return out, nil
}

// Close releases provider resources.
func (m *Migrator) Close() error {
	if m == nil || m.provider == nil {
		return nil
	}
	return m.provider.Close()
}

func migrationName(path string, version int64) string {
	base := filepath.Base(path)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	prefix := strconv.FormatInt(version, 10) + "_"
	return strings.TrimPrefix(base, prefix)
}

func isNoMigrationsError(err error) bool {
	return strings.Contains(err.Error(), "no migrations found")
}
