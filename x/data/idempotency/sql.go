package idempotency

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Dialect string

const (
	DialectPostgres Dialect = "postgres"
	DialectMySQL    Dialect = "mysql"
)

type SQLConfig struct {
	Dialect          Dialect
	Table            string
	Now              func() time.Time
	IsDuplicateError func(error) bool
}

func DefaultSQLConfig() SQLConfig {
	return SQLConfig{
		Dialect: DialectPostgres,
		Table:   "idempotency_keys",
		Now:     time.Now,
	}
}

type SQLStore struct {
	db  *sql.DB
	cfg SQLConfig
	now func() time.Time
}

func NewSQLStore(db *sql.DB, cfg SQLConfig) *SQLStore {
	if cfg.Table == "" {
		cfg.Table = "idempotency_keys"
	}
	if cfg.Dialect == "" {
		cfg.Dialect = DialectPostgres
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	if cfg.IsDuplicateError == nil {
		cfg.IsDuplicateError = defaultDuplicateError
	}
	return &SQLStore{db: db, cfg: cfg, now: cfg.Now}
}

func (s *SQLStore) Get(ctx context.Context, key string) (Record, bool, error) {
	if s == nil || s.db == nil {
		return Record{}, false, ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return Record{}, false, ErrInvalidKey
	}
	if err := s.validateConfig(); err != nil {
		return Record{}, false, err
	}

	rec, found, err := s.getRecord(ctx, key)
	if err != nil || !found {
		return Record{}, found, err
	}

	now := s.now()
	if s.isExpiredAt(rec, now) {
		_, _ = s.deleteExpired(ctx, key, now)
		return Record{}, false, nil
	}

	return rec, true, nil
}

func (s *SQLStore) getRecord(ctx context.Context, key string) (Record, bool, error) {
	query := fmt.Sprintf("SELECT key, request_hash, status, response, created_at, updated_at, expires_at FROM %s WHERE key = %s", s.cfg.Table, s.placeholder(1))
	row := s.db.QueryRowContext(ctx, query, key)

	var rec Record
	var status string
	var expiresAt sql.NullTime
	if err := row.Scan(&rec.Key, &rec.RequestHash, &status, &rec.Response, &rec.CreatedAt, &rec.UpdatedAt, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, false, nil
		}
		return Record{}, false, err
	}
	if expiresAt.Valid {
		rec.ExpiresAt = expiresAt.Time
	}
	rec.Status = Status(status)

	return rec, true, nil
}

func (s *SQLStore) PutIfAbsent(ctx context.Context, record Record) (bool, error) {
	if s == nil || s.db == nil {
		return false, ErrNotFound
	}
	if strings.TrimSpace(record.Key) == "" {
		return false, ErrInvalidKey
	}
	if !record.ExpiresAt.IsZero() && !record.ExpiresAt.After(s.now()) {
		return false, ErrExpired
	}
	if err := s.validateConfig(); err != nil {
		return false, err
	}

	now := s.now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = now
	}
	record.UpdatedAt = now
	if record.Status == "" {
		record.Status = StatusInProgress
	}
	record.Response = append([]byte(nil), record.Response...)

	query, args := s.buildInsert(record)
	for attempt := 0; attempt < 2; attempt++ {
		_, err := s.db.ExecContext(ctx, query, args...)
		if err == nil {
			return true, nil
		}
		if !s.isDuplicateError(err) {
			return false, err
		}

		existing, found, getErr := s.getRecord(ctx, record.Key)
		if getErr != nil {
			return false, getErr
		}
		if found && !s.isExpired(existing) {
			return false, nil
		}
		if found {
			deleted, deleteErr := s.deleteExpired(ctx, record.Key, s.now())
			if deleteErr != nil {
				return false, deleteErr
			}
			if !deleted {
				return false, nil
			}
			continue
		}
	}
	return false, nil
}

func (s *SQLStore) Complete(ctx context.Context, key string, response []byte) error {
	if s == nil || s.db == nil {
		return ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if err := s.validateConfig(); err != nil {
		return err
	}

	now := s.now()
	query := fmt.Sprintf(
		"UPDATE %s SET status = %s, response = %s, updated_at = %s WHERE key = %s AND (expires_at IS NULL OR expires_at > %s)",
		s.cfg.Table,
		s.placeholder(1),
		s.placeholder(2),
		s.placeholder(3),
		s.placeholder(4),
		s.placeholder(5),
	)
	res, err := s.db.ExecContext(ctx, query, StatusCompleted, append([]byte(nil), response...), now, key, now)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		_, _ = s.deleteExpired(ctx, key, now)
		return ErrNotFound
	}
	return nil
}

func (s *SQLStore) deleteExpired(ctx context.Context, key string, now time.Time) (bool, error) {
	query := fmt.Sprintf(
		"DELETE FROM %s WHERE key = %s AND expires_at IS NOT NULL AND expires_at <= %s",
		s.cfg.Table,
		s.placeholder(1),
		s.placeholder(2),
	)
	res, err := s.db.ExecContext(ctx, query, key, now)
	if err != nil {
		return false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func (s *SQLStore) Delete(ctx context.Context, key string) error {
	if s == nil || s.db == nil {
		return ErrNotFound
	}
	key = strings.TrimSpace(key)
	if key == "" {
		return ErrInvalidKey
	}
	if err := s.validateConfig(); err != nil {
		return err
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE key = %s", s.cfg.Table, s.placeholder(1))
	res, err := s.db.ExecContext(ctx, query, key)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *SQLStore) isExpired(record Record) bool {
	return s.isExpiredAt(record, s.now())
}

func (s *SQLStore) isExpiredAt(record Record, now time.Time) bool {
	return !record.ExpiresAt.IsZero() && !record.ExpiresAt.After(now)
}

func (s *SQLStore) validateConfig() error {
	if s.cfg.Dialect != DialectPostgres && s.cfg.Dialect != DialectMySQL {
		return fmt.Errorf("%w: unsupported dialect %q", ErrInvalidConfig, s.cfg.Dialect)
	}
	if !isSQLIdentifier(s.cfg.Table) {
		return fmt.Errorf("%w: unsafe table identifier %q", ErrInvalidConfig, s.cfg.Table)
	}
	return nil
}

func (s *SQLStore) buildInsert(record Record) (string, []any) {
	cols := []string{"key", "request_hash", "status", "response", "created_at", "updated_at", "expires_at"}
	placeholders := make([]string, 0, len(cols))
	args := make([]any, 0, len(cols))
	for i := range cols {
		placeholders = append(placeholders, s.placeholder(i+1))
	}
	args = append(args, record.Key, record.RequestHash, string(record.Status), record.Response, record.CreatedAt, record.UpdatedAt, nullTime(record.ExpiresAt))
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", s.cfg.Table, strings.Join(cols, ", "), strings.Join(placeholders, ", "))
	return query, args
}

func (s *SQLStore) placeholder(idx int) string {
	if s.cfg.Dialect == DialectPostgres {
		return fmt.Sprintf("$%d", idx)
	}
	return "?"
}

func nullTime(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t
}

func (s *SQLStore) isDuplicateError(err error) bool {
	if s == nil || s.cfg.IsDuplicateError == nil {
		return defaultDuplicateError(err)
	}
	return s.cfg.IsDuplicateError(err)
}

func defaultDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return errors.Is(err, sql.ErrNoRows) == false && (strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique"))
}

func isSQLIdentifier(identifier string) bool {
	if identifier == "" {
		return false
	}
	for _, part := range strings.Split(identifier, ".") {
		if !isSQLIdentifierPart(part) {
			return false
		}
	}
	return true
}

func isSQLIdentifierPart(part string) bool {
	if part == "" {
		return false
	}
	for i := 0; i < len(part); i++ {
		c := part[i]
		if i == 0 {
			if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_' {
				continue
			}
			return false
		}
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_' {
			continue
		}
		return false
	}
	return true
}
