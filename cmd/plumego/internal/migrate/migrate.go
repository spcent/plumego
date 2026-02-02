package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	// ErrInvalidMigrationName is returned when a migration name is invalid.
	ErrInvalidMigrationName = errors.New("invalid migration name")

	// ErrMissingMigrationFile is returned when a required migration file is missing.
	ErrMissingMigrationFile = errors.New("missing migration file")
)

// Migration represents a single migration pair.
type Migration struct {
	Version  string
	Name     string
	UpPath   string
	DownPath string
}

// AppliedMigration represents an applied migration entry.
type AppliedMigration struct {
	Version   string
	Name      string
	AppliedAt time.Time
}

var migrationPattern = regexp.MustCompile(`^(\d+)_(.+)\.(up|down)\.sql$`)

// LoadMigrations loads migrations from a directory.
func LoadMigrations(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	migrations := make(map[string]*Migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		matches := migrationPattern.FindStringSubmatch(name)
		if len(matches) != 4 {
			continue
		}

		version := matches[1]
		baseName := matches[2]
		section := matches[3]

		path := filepath.Join(dir, name)
		entryRef := migrations[version]
		if entryRef == nil {
			entryRef = &Migration{Version: version, Name: baseName}
			migrations[version] = entryRef
		}

		if entryRef.Name == "" {
			entryRef.Name = baseName
		}

		if section == "up" {
			entryRef.UpPath = path
		} else {
			entryRef.DownPath = path
		}
	}

	result := make([]Migration, 0, len(migrations))
	for _, migration := range migrations {
		result = append(result, *migration)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Version < result[j].Version
	})

	return result, nil
}

// CreateMigrationFiles creates migration up/down files.
func CreateMigrationFiles(dir, name string, now time.Time) (Migration, error) {
	if strings.TrimSpace(name) == "" {
		return Migration{}, ErrInvalidMigrationName
	}

	sanitized := sanitizeName(name)
	if sanitized == "" {
		return Migration{}, ErrInvalidMigrationName
	}

	version := now.UTC().Format("20060102150405")
	base := fmt.Sprintf("%s_%s", version, sanitized)
	upPath := filepath.Join(dir, base+".up.sql")
	downPath := filepath.Join(dir, base+".down.sql")

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return Migration{}, err
	}

	if err := writeFileIfMissing(upPath, "-- Write your up migration here\n"); err != nil {
		return Migration{}, err
	}

	if err := writeFileIfMissing(downPath, "-- Write your down migration here\n"); err != nil {
		return Migration{}, err
	}

	return Migration{
		Version:  version,
		Name:     sanitized,
		UpPath:   upPath,
		DownPath: downPath,
	}, nil
}

// EnsureSchemaTable ensures the schema migrations table exists.
func EnsureSchemaTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		applied_at TIMESTAMP NOT NULL
	)`)
	return err
}

// FetchApplied returns applied migrations from the database.
func FetchApplied(ctx context.Context, db *sql.DB) ([]AppliedMigration, error) {
	rows, err := db.QueryContext(ctx, `SELECT version, name, applied_at FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applied []AppliedMigration
	for rows.Next() {
		var entry AppliedMigration
		if err := rows.Scan(&entry.Version, &entry.Name, &entry.AppliedAt); err != nil {
			return nil, err
		}
		applied = append(applied, entry)
	}

	return applied, rows.Err()
}

// ApplyUp applies an up migration.
func ApplyUp(ctx context.Context, db *sql.DB, driver string, migration Migration, appliedAt time.Time) (time.Duration, error) {
	if migration.UpPath == "" {
		return 0, fmt.Errorf("%w: missing up migration for %s", ErrMissingMigrationFile, migration.Version)
	}

	contents, err := os.ReadFile(migration.UpPath)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	insertQuery := insertMigrationQuery(driver)
	if _, err := tx.ExecContext(ctx, insertQuery, migration.Version, migration.Name, appliedAt); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

// ApplyDown applies a down migration.
func ApplyDown(ctx context.Context, db *sql.DB, driver string, migration Migration) (time.Duration, error) {
	if migration.DownPath == "" {
		return 0, fmt.Errorf("%w: missing down migration for %s", ErrMissingMigrationFile, migration.Version)
	}

	contents, err := os.ReadFile(migration.DownPath)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	deleteQuery := deleteMigrationQuery(driver)
	if _, err := tx.ExecContext(ctx, deleteQuery, migration.Version); err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return time.Since(start), nil
}

func sanitizeName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ToLower(name)

	cleaned := strings.Builder{}
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			cleaned.WriteRune(r)
		}
	}

	return strings.Trim(cleaned.String(), "_")
}

func insertMigrationQuery(driver string) string {
	placeholder := placeholderForDriver(driver, 1)
	second := placeholderForDriver(driver, 2)
	third := placeholderForDriver(driver, 3)
	return fmt.Sprintf("INSERT INTO schema_migrations (version, name, applied_at) VALUES (%s, %s, %s)", placeholder, second, third)
}

func deleteMigrationQuery(driver string) string {
	placeholder := placeholderForDriver(driver, 1)
	return fmt.Sprintf("DELETE FROM schema_migrations WHERE version = %s", placeholder)
}

func placeholderForDriver(driver string, index int) string {
	driver = strings.ToLower(driver)
	if strings.Contains(driver, "postgres") || strings.Contains(driver, "pgx") {
		return fmt.Sprintf("$%d", index)
	}
	return "?"
}

func writeFileIfMissing(path, content string) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file already exists: %s", path)
	}

	return os.WriteFile(path, []byte(content), 0o644)
}
